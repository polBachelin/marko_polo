package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
)

const version = "0.1.0"

const usage = `marko — a terminal markdown reader

Usage:
  marko <file.md>    Render a markdown file
  marko -            Read from stdin
  cat file | marko   Pipe markdown to stdin

Options:
  --help      Show this help
  --version   Show version

Environment:
  GLAMOUR_STYLE   Set rendering style (dark, light, notty, dracula, ascii)
  PAGER           Set pager command (default: less -r)`

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "marko: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	md, err := getInput()
	if err != nil {
		return err
	}

	width := terminalWidth()

	rendered, err := render(md, width)
	if err != nil {
		return fmt.Errorf("render failed: %w", err)
	}

	output(rendered)
	return nil
}

func getInput() ([]byte, error) {
	args := os.Args[1:]

	if len(args) == 0 {
		if stdinIsPiped() {
			return io.ReadAll(os.Stdin)
		}
		fmt.Println(usage)
		os.Exit(0)
	}

	switch args[0] {
	case "--help", "-h":
		fmt.Println(usage)
		os.Exit(0)
	case "--version", "-v":
		fmt.Printf("marko %s\n", version)
		os.Exit(0)
	case "-":
		return io.ReadAll(os.Stdin)
	}

	if len(args) > 1 {
		return nil, fmt.Errorf("too many arguments (expected 1 file)")
	}

	path := args[0]
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("%s: file is empty", path)
	}

	return data, nil
}

func render(md []byte, width int) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
		glamour.WithEmoji(),
	)
	if err != nil {
		return "", err
	}
	return r.Render(string(md))
}

func output(rendered string) {
	if !stdoutIsTTY() {
		fmt.Print(rendered)
		return
	}

	height := terminalHeight()
	lines := strings.Count(rendered, "\n")

	if lines <= height {
		fmt.Print(rendered)
		return
	}

	if err := pager(rendered); err != nil {
		fmt.Print(rendered)
	}
}

func pager(content string) error {
	pagerCmd := os.Getenv("PAGER")
	if pagerCmd == "" {
		pagerCmd = "less -r"
	}

	parts := strings.Fields(pagerCmd)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func stdinIsPiped() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}

func stdoutIsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func terminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	if w > 120 {
		return 120
	}
	return w
}

func terminalHeight() int {
	_, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || h <= 0 {
		return 24
	}
	return h
}
