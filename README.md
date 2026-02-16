# TuxGo

A tmux session manager that automatically creates and configures tmux sessions based on YAML configuration files. Define your window layouts, panel splits, and startup commands once, and TuxGo sets everything up for you.

## Features

- **YAML-based configuration** with local (per-project) and global support
- **Hierarchical panel layouts** for complex split arrangements (horizontal + vertical mixed)
- **Flat panel layouts** for simple multi-pane windows
- **Project pattern matching** using glob patterns in global config
- **Nested tmux detection** -- uses `switch-client` when already inside tmux
- **Interactive session picker** with a TUI powered by Bubble Tea
- **Shell completions** for bash, zsh, fish, and powershell
- **UTF-8 support** via `tmux -u` on all commands

## Installation

```bash
go install github.com/josesaulburgos/tuxgo/cmd/tuxgo@latest
```

Make sure `$HOME/go/bin` is in your PATH:

```bash
# fish
fish_add_path $HOME/go/bin

# bash/zsh
export PATH="$HOME/go/bin:$PATH"
```

Or build from source:

```bash
git clone https://github.com/josesaulburgos/tuxgo.git
cd tuxgo
make install
```

## Quick Start

```bash
# Create a local config in your project directory
tuxgo init

# Edit the generated .tuxgo.yaml to your liking, then run:
tuxgo
```

## Usage

```
tuxgo                # Create/attach to tmux session using config
tuxgo init           # Create example .tuxgo.yaml in current directory
tuxgo list           # List active tmux sessions and select one (alias: ls)
tuxgo completion     # Generate shell completion scripts
tuxgo help           # Show help
```

### Shell Completions

```bash
# fish
tuxgo completion fish | source

# bash
source <(tuxgo completion bash)

# zsh
tuxgo completion zsh > "${fpath[1]}/_tuxgo"
```

## Configuration

TuxGo looks for configuration in this order (first found wins):

1. **Local config**: `.tuxgo.yaml` or `.tuxgo.yml` in the current directory
2. **Global config**: `~/.config/tuxgo/config.yaml`

### Local Config (`.tuxgo.yaml`)

Place this file in your project root. It applies only to that directory.

```yaml
windows:
  - name: editor
    command: "nvim ."

  - name: dev
    layout: horizontal
    panels:
      - "go run ."
      - "tail -f logs/app.log"
```

### Global Config (`~/.config/tuxgo/config.yaml`)

Supports a `default` section and project-specific configs matched by glob patterns.

```yaml
default:
  windows:
    - name: editor
      command: "nvim ."

projects:
  - pattern: "*/my-go-project"
    windows:
      - name: editor
        command: "nvim ."
      - name: server
        command: "go run ."
```

### Window Types

#### Simple window (single pane)

```yaml
- name: editor
  command: "nvim ."
```

#### Flat layout (equal panels)

```yaml
- name: dev
  layout: horizontal   # or "vertical"
  panels:
    - "go run ."
    - "npm run dev"
    - "tail -f app.log"
```

#### Hierarchical layout (mixed splits)

Use `root` with nested `children` for complex arrangements. Each node is either a **container** (has `split` + `children`) or a **leaf** (has `command`). Containers can have up to 2 children.

```yaml
- name: workspace
  root:
    split: "horizontal"
    children:
      - command: "nvim ."           # left pane
      - split: "vertical"          # right side, split vertically
        children:
          - command: "htop"         # top-right pane
          - command: "watch date"   # bottom-right pane
```

This produces:

```
+---------------+----------+
|               |   htop   |
|     nvim      +----------+
|               |  watch   |
+---------------+----------+
```

You can nest further for more complex layouts:

```yaml
- name: complex
  root:
    split: "horizontal"
    children:
      - split: "vertical"
        children:
          - command: "nvim ."
          - command: "make watch"
      - split: "vertical"
        children:
          - command: "htop"
          - command: "tail -f app.log"
```

```
+----------+----------+
|   nvim   |   htop   |
+----------+----------+
|   make   |   tail   |
+----------+----------+
```

## Project Structure

```
tuxgo/
  cmd/
    tuxgo/
      main.go                 # Entry point
  internal/
    cli/
      root.go                 # Root command (create/attach session)
      init.go                 # 'tuxgo init' subcommand
      list.go                 # 'tuxgo list' subcommand
    config/
      config.go               # Config structs and YAML parsing
      loader.go               # Config file discovery and loading
      templates.go             # Default config templates
      config_test.go
    tmux/
      session.go              # Session create, attach, detect
      window.go               # Window and hierarchical layout creation
      pane.go                 # Pane split, select, send keys
      util.go                 # Session name, list sessions, helpers
      validate.go             # Config validation
      tmux_test.go
    matcher/
      matcher.go              # Glob pattern matching for projects
      matcher_test.go
    tui/
      session_picker.go       # Interactive session selector (Bubble Tea)
  Makefile
  go.mod
  go.sum
  README.md
  .gitignore
```

## Development

```bash
make build      # Compile binary
make install    # Install to $GOPATH/bin
make test       # Run all tests
make fmt        # Format code
make vet        # Run go vet
make lint       # Run golangci-lint
make clean      # Remove build artifacts
make help       # Show all targets
```

## How It Works

1. TuxGo reads the config for the current directory (local `.tuxgo.yaml` first, then global)
2. For global config, it matches the current path against project patterns (first match wins)
3. If no project matches, it falls back to the `default` section
4. If a tmux session with the directory name already exists, it attaches to it
5. Otherwise, it creates a new session with all configured windows and panels
6. If running inside tmux, it uses `switch-client`; outside, it uses `attach-session`

## Dependencies

- [tmux](https://github.com/tmux/tmux) must be installed
- [Cobra](https://github.com/spf13/cobra) -- CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) -- TUI framework for session picker
- [yaml.v3](https://gopkg.in/yaml.v3) -- YAML parsing

## License

MIT
