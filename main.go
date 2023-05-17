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
	"github.com/mattn/go-colorable"
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

var flagMoveEnd = flag.Bool("move-end", false, "Move cursor to end of file")

func progName(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
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

	ctx := context.Background()
	lines, err = ed.Read(ctx)
	if err != nil {
		return err
	}
	return save(args[0], lines)
}

func main() {
	flag.Parse()
	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

}
