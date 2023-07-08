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
	"path/filepath"

	"github.com/hymkor/go-multiline-ny"
	"github.com/hymkor/go-readline-skk"
	"github.com/mattn/go-colorable"
	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/keys"
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

func save(fn string, lines []string) error {
	fd, err := os.Create(fn)
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
	return fmt.Fprintf(w, "New Candidate for \"%s\": ", prompt)
}

func (q *miniBuffer) Leave(w io.Writer) (int, error) {
	q.rewind()
	return 0, nil
}

func (q *miniBuffer) Recurse(originalPrompt string) skk.MiniBuffer {
	return &skk.MiniBufferOnCurrentLine{OriginalPrompt: originalPrompt}
}

func progName(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}

type cmdSave struct {
	ed       *multiline.Editor
	filename string
}

func (c *cmdSave) String() string {
	return "save to " + c.filename
}

func alert(ctx context.Context, B *readline.Buffer, m *multiline.Editor, s string) readline.Result {
	rewind := m.GotoEndLine()
	io.WriteString(B.Out, s)
	key, err := B.GetKey()
	rewind()
	B.RepaintAll()
	if err == nil {
		return B.LookupCommand(key).Call(ctx, B)
	}
	return readline.CONTINUE
}

func (c *cmdSave) Call(ctx context.Context, B *readline.Buffer) readline.Result {
	lines := c.ed.Lines()
	current := c.ed.CursorLine()
	if current < len(lines) {
		lines[current] = B.String()
	} else {
		lines = append(lines, B.String())
	}
	if err := save(c.filename, lines); err != nil {
		return alert(ctx, B, c.ed, err.Error())
	}
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
	if len(args) <= 0 {
		return fmt.Errorf("usage: %s FILENAME", progName(os.Args[0]))
	}
	lines, err := load(args[0])
	if err != nil {
		return err
	}

	f := colorable.EnableColorsStdout(nil)
	defer f()

	var ed multiline.Editor
	ed.SetPrompt(func(w io.Writer, lnum int) (int, error) {
		return fmt.Fprintf(w, "\x1B[0;32;1m%2d\x1B[0;37;1m ", lnum+1)
	})
	ed.SetWriter(colorable.NewColorableStdout())
	ed.SetDefault(lines)
	ed.SetMoveEnd(*flagMoveEnd)

	ed.BindKey(keys.CtrlC, noOperation{})

	ctrlX := &multiline.PrefixCommand{}
	ctrlX.BindKey(keys.CtrlC, readline.AnonymousCommand(ed.Submit))
	ctrlX.BindKey(keys.CtrlS, &cmdSave{ed: &ed, filename: args[0]})
	ed.BindKey(keys.CtrlX, ctrlX)

	skk1 := skk.New()
	skk1.MiniBuffer = &miniBuffer{ed: &ed}
	skk1.System.ReadEucJp(bzip2.NewReader(bytes.NewReader(skkJisyoLbz2)))
	ed.LineEditor.BindKey(keys.CtrlJ, skk1)

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
