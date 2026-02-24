# marko_polo

A lightweight terminal markdown reader. Renders `.md` files with colors, syntax highlighting, and proper formatting — right in your terminal.

Built with [Glamour](https://github.com/charmbracelet/glamour) for beautiful markdown rendering.

## Install

```bash
# Build from source
git clone git@github.com:polBachelin/marko_polo.git
cd marko_polo
make install
```

This builds the `marko` binary and copies it to `/usr/local/bin`.

## Usage

```bash
# Render a file
marko README.md

# Pipe from stdin
cat notes.md | marko

# Explicit stdin
marko -

# Help & version
marko --help
marko --version
```

## Configuration

| Variable | Description | Default |
|---|---|---|
| `GLAMOUR_STYLE` | Rendering style (`dark`, `light`, `notty`, `dracula`, `ascii`) | Auto-detected |
| `PAGER` | Pager for long output | `less -r` |

## Shell Alias

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
alias marko='/usr/local/bin/marko'
```

## Uninstall

```bash
make uninstall
```
