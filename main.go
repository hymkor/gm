package main

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/completion"
	"github.com/nyaosorg/go-readline-ny/keys"
	"github.com/nyaosorg/go-readline-skk"
	"github.com/nyaosorg/go-ttyadapter/tty10"

	"github.com/hymkor/go-multiline-ny"
	"github.com/hymkor/go-windows1x-virtualterminal"
)

//go:embed SKK-JISYO.L.bz2
var skkJisyoLbz2 []byte

func load(filename string) ([]string, error) {
	fd, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	lines := []string{}
	for sc := bufio.NewScanner(fd); sc.Scan(); {
		lines = append(lines, sc.Text())
	}
	fd.Close()
	return lines, nil
}

func save(fn string, lines []string, flag int) error {
	fd, err := os.OpenFile(fn, flag, 0644)
	if err != nil {
		return err
	}
	for _, line := range lines {
		fmt.Fprintln(fd, line)
	}
	return fd.Close()
}

var (
	flagMoveEnd = flag.Bool("move-end", false, "Move cursor to end of file")
)

type miniBuffer struct {
	ed     *multiline.Editor
	rewind func()
}

func (q *miniBuffer) Enter(w io.Writer, prompt string) (int, error) {
	q.rewind = q.ed.GotoEndLine()
	return io.WriteString(w, prompt)
}

func (q *miniBuffer) Leave(w io.Writer) (int, error) {
	io.WriteString(w, "\x1B[2K")
	q.rewind()
	return 0, nil
}

func (q *miniBuffer) Recurse() skk.MiniBuffer {
	return skk.MiniBufferOnCurrentLine{}
}

type cmdSave struct {
	ed       *multiline.Editor
	filename string
}

func (c *cmdSave) String() string {
	return "save to " + c.filename
}

func askKey(B *readline.Buffer, m *multiline.Editor, s string) (string, error) {
	rewind := m.GotoEndLine()
	io.WriteString(B.Out, s)
	key, err := B.GetKey()
	io.WriteString(B.Out, "\x1B[2K")
	rewind()
	return key, err
}

func alert(ctx context.Context, B *readline.Buffer, m *multiline.Editor, s string) readline.Result {
	rewind := m.GotoEndLine()
	io.WriteString(B.Out, s)
	key, err := B.GetKey()
	io.WriteString(B.Out, "\x1B[2K")
	rewind()
	B.RepaintAll()
	if err == nil {
		return B.LookupCommand(key).Call(ctx, B)
	}
	return readline.CONTINUE
}

func ask(ctx context.Context, me *multiline.Editor, defaultText string) (string, error) {
	miniBuffer1 := &miniBuffer{
		ed: me,
	}
	ed1 := &readline.Editor{
		Out:    me.LineEditor.Out,
		Writer: me.Writer(),
		PromptWriter: func(w io.Writer) (int, error) {
			return miniBuffer1.Enter(w, "Save filename: ")
		},
		LineFeedWriter: func(_ readline.Result, w io.Writer) (int, error) {
			io.WriteString(w, "\x1B[2K")
			return miniBuffer1.Leave(w)
		},
		Default: defaultText,
	}
	ed1.BindKey(keys.CtrlI, &completion.CmdCompletion2{
		Candidates: completion.PathComplete,
	})
	return ed1.ReadLine(ctx)
}

func (c *cmdSave) Call(ctx context.Context, B *readline.Buffer) readline.Result {
	lines := c.ed.Lines()
	current := c.ed.CursorLine()
	if current < len(lines) {
		lines[current] = B.String()
	} else {
		lines = append(lines, B.String())
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	fname := c.filename
	if fname == "" {
		flag = os.O_WRONLY | os.O_CREATE | os.O_EXCL | os.O_TRUNC
		text, err := ask(ctx, c.ed, c.filename)
		B.RepaintAll()
		if err != nil || text == "" {
			return readline.CONTINUE
		}
		fname = text
	}
	if err := save(fname, lines, flag); err != nil {
		return alert(ctx, B, c.ed, err.Error())
	}
	c.ed.Dirty = false
	c.filename = fname
	return alert(ctx, B, c.ed, "saved as "+c.filename)
}

type noOperation struct{}

func (noOperation) String() string {
	return "NO_OPERATION"
}

func (noOperation) Call(context.Context, *readline.Buffer) readline.Result {
	return readline.CONTINUE
}

func mains(args []string) error {
	var filename string
	var lines []string
	if len(args) > 0 {
		var err error
		lines, err = load(args[0])
		if err != nil {
			return err
		}
		filename = args[0]
	}
	if closer, err := virtualterminal.EnableStdout(); err != nil {
		return err
	} else {
		defer closer()
	}

	var ed multiline.Editor
	ed.SetPrompt(func(w io.Writer, lnum int) (int, error) {
		return fmt.Fprintf(w, "\x1B[0;32;1m%2d\x1B[0;37;1m ", lnum+1)
	})
	ed.SetTty(&tty10.Tty{})
	ed.SetDefault(lines)
	ed.SetMoveEnd(*flagMoveEnd)
	const resetColor = "\x1B[0m"
	ed.Highlight = []readline.Highlight{
		skk.BlackMarkerHighlight,
		skk.WhiteMarkerHighlight,
	}
	ed.ResetColor = resetColor
	ed.DefaultColor = resetColor
	ed.BindKey(keys.CtrlC, noOperation{})
	ed.BindKey(keys.Escape, noOperation{})

	ctrlX := ed.NewPrefixCommand("C-x ")
	ctrlX.BindKey(keys.CtrlC, &readline.GoCommand{
		Name: "Quit",
		Func: func(ctx context.Context, B *readline.Buffer) readline.Result {
			ed.Sync(B.String())
			if ed.Dirty {
				answer, err := askKey(B, &ed, "the file is not saved. Quit sure ? (y/n) ")
				if err != nil {
					alert(ctx, B, &ed, err.Error())
					B.RepaintAll()
					return readline.CONTINUE
				}
				if answer != "y" && answer != "Y" {
					B.RepaintAll()
					return readline.CONTINUE
				}
			}
			return ed.Submit(ctx, B)
		},
	})
	ctrlX.BindKey(keys.CtrlS, &cmdSave{ed: &ed, filename: filename})
	ed.BindKey(keys.CtrlX, ctrlX)

	skk1, err := skk.Config{
		BindTo:         &ed.LineEditor,
		KeepModeOnExit: true,
	}.Setup()
	if err != nil {
		return err
	}

	skk1.MiniBuffer = &miniBuffer{ed: &ed}
	err = skk1.System.Read(bzip2.NewReader(bytes.NewReader(skkJisyoLbz2)))
	if err != nil {
		return err
	}
	ctx := context.Background()
	_, err = ed.Read(ctx)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

}
