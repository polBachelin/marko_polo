package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"golang.org/x/term"
)

const version = "0.2.0"

const usage = `marko — a terminal markdown reader

Usage:
  marko <file.md>       Open in visual reader (default)
  marko -t <file.md>    Render markdown in terminal
  marko -               Read from stdin
  cat file | marko      Pipe markdown to stdin

Options:
  -t, --term    Render in terminal instead of visual reader
  --help        Show this help
  --version     Show version

Environment:
  GLAMOUR_STYLE   Set terminal rendering style (dark, light, notty, dracula, ascii)
  PAGER           Set pager command (default: less -r)`

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "marko: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	termMode, args := parseFlags(os.Args[1:])

	md, err := getInput(args)
	if err != nil {
		return err
	}

	if termMode {
		width := terminalWidth()
		rendered, err := render(md, width)
		if err != nil {
			return fmt.Errorf("render failed: %w", err)
		}
		output(rendered)
		return nil
	}

	return openReader(md)
}

func parseFlags(args []string) (termMode bool, remaining []string) {
	for _, arg := range args {
		switch arg {
		case "-t", "--term":
			termMode = true
		default:
			remaining = append(remaining, arg)
		}
	}
	return
}

func getInput(args []string) ([]byte, error) {
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

// --- Terminal rendering ---

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

// --- Visual reader ---

func openReader(md []byte) error {
	body := renderHTML(md)
	title := extractTitle(md)
	page := readerPage(title, body)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	url := "http://" + ln.Addr().String()
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, page)
		}),
	}

	go srv.Serve(ln)

	fmt.Printf("Reader opened at %s — Press Ctrl+C to close\n", url)
	openBrowser(url)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	fmt.Println("\nClosing reader...")
	return srv.Shutdown(context.Background())
}

func renderHTML(md []byte) string {
	mdParser := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	mdParser.Convert(md, &buf)
	return buf.String()
}

func extractTitle(md []byte) string {
	for _, line := range strings.Split(string(md), "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return "marko reader"
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}

func readerPage(title, content string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + title + `</title>
<style>
:root {
  --bg: #ffffff;
  --fg: #24292e;
  --secondary: #586069;
  --border: #e1e4e8;
  --code-bg: #f6f8fa;
  --link: #0366d6;
  --quote-border: #dfe2e5;
  --table-border: #dfe2e5;
}
@media (prefers-color-scheme: dark) {
  :root {
    --bg: #0d1117;
    --fg: #c9d1d9;
    --secondary: #8b949e;
    --border: #30363d;
    --code-bg: #161b22;
    --link: #58a6ff;
    --quote-border: #3b434b;
    --table-border: #30363d;
  }
}
* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans", Helvetica, Arial, sans-serif;
  font-size: 17px;
  line-height: 1.7;
  color: var(--fg);
  background: var(--bg);
  padding: 3rem 1.5rem;
}
article { max-width: 720px; margin: 0 auto; }
h1, h2, h3, h4, h5, h6 {
  margin-top: 1.5em;
  margin-bottom: 0.5em;
  font-weight: 600;
  line-height: 1.3;
}
h1 { font-size: 2em; border-bottom: 1px solid var(--border); padding-bottom: 0.3em; }
h2 { font-size: 1.5em; border-bottom: 1px solid var(--border); padding-bottom: 0.3em; }
h3 { font-size: 1.25em; }
h1:first-child { margin-top: 0; }
p { margin-bottom: 1em; }
a { color: var(--link); text-decoration: none; }
a:hover { text-decoration: underline; }
code {
  font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
  font-size: 0.875em;
  background: var(--code-bg);
  padding: 0.2em 0.4em;
  border-radius: 4px;
}
pre {
  margin-bottom: 1em;
  padding: 1em;
  overflow-x: auto;
  border-radius: 8px;
  line-height: 1.5;
  background: var(--code-bg);
}
pre code { background: none; padding: 0; }
blockquote {
  margin-bottom: 1em;
  padding: 0.5em 1em;
  border-left: 4px solid var(--quote-border);
  color: var(--secondary);
}
ul, ol { margin-bottom: 1em; padding-left: 2em; }
li { margin-bottom: 0.25em; }
table { width: 100%; margin-bottom: 1em; border-collapse: collapse; }
th, td { padding: 0.5em 1em; border: 1px solid var(--table-border); text-align: left; }
th { font-weight: 600; background: var(--code-bg); }
img { max-width: 100%; height: auto; }
hr { margin: 1.5em 0; border: none; border-top: 1px solid var(--border); }
input[type="checkbox"] { margin-right: 0.5em; }
</style>
</head>
<body>
<article>` + content + `</article>
</body>
</html>`
}

// --- Utilities ---

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
