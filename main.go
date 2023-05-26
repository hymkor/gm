package main

import (
	"bufio"
	"context"
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
	flagSKK     = flag.String("skk", "", "Enable SKK and Specify JISYO-Path")
)

type queryPrompter struct {
	ed *multiline.Editor
}

func (q *queryPrompter) Prompt(w io.Writer, prompt string) (int, error) {
	return fmt.Fprintf(w, "\rNew Candidate for \"%s\": ", prompt)
}

func (q *queryPrompter) LineFeed(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "\r\x1B[0;32;1m%2d\x1B[0;37;1m ", q.ed.CursorLine()+1)
}

func (q *queryPrompter) Recurse(originalPrompt string) skk.QueryPrompter {
	return &skk.QueryOnCurrentLine{OriginalPrompt: originalPrompt}
}

func progName(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}

type CtrlX struct {
	readline.KeyMap
}

func (cx *CtrlX) String() string {
	return ""
}

func (cx *CtrlX) Call(ctx context.Context, B *readline.Buffer) readline.Result {
	// B.InsertAndRepaint(string(keys.CtrlX))
	key, err := B.GetKey()
	if err != nil {
		return readline.CONTINUE
	}
	f, ok := cx.KeyMap.Lookup(keys.Code(key))
	if !ok {
		return readline.CONTINUE
	}
	return f.Call(ctx, B)
}

func noOperation(_ context.Context, _ *readline.Buffer) readline.Result {
	return readline.CONTINUE
}

type cmdSave struct {
	ed       *multiline.Editor
	filename string
}

func (c *cmdSave) String() string {
	return "save to " + c.filename
}

func alert(ctx context.Context, B *readline.Buffer, s string) readline.Result {
	fmt.Fprintf(B.Out, "\r%s\r", s)
	key, err := B.GetKey()
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
		return alert(ctx, B, err.Error())
	}
	return alert(ctx, B, "saved as "+c.filename)
}

func mains(args []string) error {
	if len(args) <= 0 {
		return fmt.Errorf("Usage: %s FILENAME", progName(os.Args[0]))
	}
	lines, err := load(args[0])
	if err != nil {
		return err
	}

	var ed multiline.Editor
	ed.SetPrompt(func(w io.Writer, lnum int) (int, error) {
		return fmt.Fprintf(w, "\x1B[0;32;1m%2d\x1B[0;37;1m ", lnum+1)
	})
	ed.SetWriter(colorable.NewColorableStdout())
	ed.SetDefault(lines)
	ed.SetMoveEnd(*flagMoveEnd)

	ctrlX := &CtrlX{}
	ctrlX.BindKey(keys.CtrlC, readline.AnonymousCommand(ed.Submit))
	ctrlX.BindKey(keys.CtrlS, &cmdSave{ed: &ed, filename: args[0]})
	ed.BindKey(keys.CtrlX, ctrlX)
	ed.BindKey(keys.CtrlC, readline.AnonymousCommand(noOperation))

	if *flagSKK != "" {
		skk1, err := skk.Load("", *flagSKK)
		if err == nil {
			ed.LineEditor.BindKey(keys.CtrlJ, skk1)
			skk1.QueryPrompter = &queryPrompter{ed: &ed}
		}
	}

	ctx := context.Background()
	lines, err = ed.Read(ctx)
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
