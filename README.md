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

### From GitHub Releases (recommended)

Download the latest binary for your platform from the [Releases page](https://github.com/Osmait/tuxgo/releases).

```bash
# Example for macOS arm64
tar -xzf tuxgo_*_darwin_arm64.tar.gz
sudo mv tuxgo /usr/local/bin/
```

### From source with `go install`

```bash
go install github.com/Osmait/tuxgo/cmd/tuxgo@latest
```

Make sure `$HOME/go/bin` is in your PATH:

```bash
# fish
fish_add_path $HOME/go/bin

# bash/zsh
export PATH="$HOME/go/bin:$PATH"
```

### Build from source

```bash
git clone https://github.com/Osmait/tuxgo.git
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

### Hierarchical Layout Guide

For complex multi-pane arrangements, use `root` with nested `children`. This is a **binary tree** model where you describe how to recursively split the screen.

#### Key concepts

There are only two types of nodes:

| Node type | Has | Does |
|-----------|-----|------|
| **Container** | `split` + `children` | Splits the space in two. Does NOT run a command. |
| **Leaf** | `command` | Runs a command. Does NOT split. |

**Rules:**
- `split: "horizontal"` divides the space into **left** and **right**
- `split: "vertical"` divides the space into **top** and **bottom**
- Each container has exactly **2 children** (binary split)
- The first child gets the **left/top** portion
- The second child gets the **right/bottom** portion
- Children can be either leaves (run a command) or containers (split further)

#### How to think about it

Think of it as **recursively cutting a rectangle**:

1. Start with the full window as one rectangle
2. Decide: do I want to split this rectangle horizontally (left|right) or vertically (top/bottom)?
3. For each half: is it a final pane (leaf with command) or do I need to split it again (container)?

#### Example 1: Editor with side panel

Split horizontally: editor on the left, terminal on the right.

```yaml
- name: dev
  root:
    split: "horizontal"           # cut left | right
    children:
      - command: "nvim ."         # left: editor
      - command: "npm run dev"    # right: terminal
```

```
+-------------+-----------+
|             |           |
|    nvim     |  npm run  |
|             |           |
+-------------+-----------+
```

#### Example 2: Editor with stacked side panels

Split horizontally first, then split the right side vertically.

```yaml
- name: workspace
  root:
    split: "horizontal"              # step 1: cut left | right
    children:
      - command: "nvim ."            # left: editor (leaf)
      - split: "vertical"            # right: split again top / bottom
        children:
          - command: "htop"          #   top-right: htop
          - command: "watch date"    #   bottom-right: watch
```

How to read this:

```
Step 1: horizontal split        Step 2: vertical split on right
+-------------+-----------+     +-------------+-----------+
|             |           |     |             |   htop    |
|    nvim     |   (right) | --> |    nvim     +-----------+
|             |           |     |             |   watch   |
+-------------+-----------+     +-------------+-----------+
```

#### Example 3: 2x2 grid

Split horizontally, then split each side vertically.

```yaml
- name: grid
  root:
    split: "horizontal"              # step 1: left | right
    children:
      - split: "vertical"            # step 2: split left into top / bottom
        children:
          - command: "nvim ."        #   top-left
          - command: "make watch"    #   bottom-left
      - split: "vertical"            # step 3: split right into top / bottom
        children:
          - command: "htop"          #   top-right
          - command: "tail -f app.log" # bottom-right
```

```
Step 1: horizontal     Step 2: split left    Step 3: split right
+--------+--------+    +--------+--------+   +--------+--------+
|        |        |    |  nvim  |        |   |  nvim  |  htop  |
| (left) |(right) | -> +--------+ (right)| ->+--------+--------+
|        |        |    |  make  |        |   |  make  |  tail  |
+--------+--------+    +--------+--------+   +--------+--------+
```

#### Example 4: Main editor with three side panels

Split horizontally, then the right side vertically into three (requires nesting two vertical splits).

```yaml
- name: ide
  root:
    split: "horizontal"
    children:
      - command: "nvim ."                  # left: large editor
      - split: "vertical"                  # right: split into top / bottom
        children:
          - command: "go run ."            #   top-right: server
          - split: "vertical"              #   bottom-right: split again
            children:
              - command: "go test ./..."   #     middle-right: tests
              - command: "lazygit"         #     bottom-right: git
```

```
+----------------+-----------+
|                |  go run   |
|                +-----------+
|      nvim      |  go test  |
|                +-----------+
|                |  lazygit  |
+----------------+-----------+
```

#### Example 5: Top bar with bottom panels

Split vertically first (top/bottom), then the bottom horizontally (left/right).

```yaml
- name: monitor
  root:
    split: "vertical"                    # step 1: top / bottom
    children:
      - command: "htop"                  # top: full-width htop
      - split: "horizontal"              # bottom: split left | right
        children:
          - command: "watch df -h"       #   bottom-left: disk
          - command: "tail -f /var/log/syslog" # bottom-right: logs
```

```
+---------------------------+
|           htop            |
+-------------+-------------+
|   watch df  |  tail logs  |
+-------------+-------------+
```

#### Example 6: Three columns

Split horizontally, then split one side horizontally again.

```yaml
- name: columns
  root:
    split: "horizontal"
    children:
      - command: "nvim ."                # left column
      - split: "horizontal"              # right side: split into 2 columns
        children:
          - command: "go run ."          #   middle column
          - command: "lazygit"           #   right column
```

```
+---------+---------+---------+
|         |         |         |
|  nvim   | go run  | lazygit |
|         |         |         |
+---------+---------+---------+
```

#### Quick reference

| Want this | Root split | Children |
|-----------|-----------|----------|
| Left \| Right | `horizontal` | 2 leaves |
| Top / Bottom | `vertical` | 2 leaves |
| Left \| (Top-Right / Bottom-Right) | `horizontal` | leaf + vertical container |
| (Top-Left / Bottom-Left) \| Right | `horizontal` | vertical container + leaf |
| 2x2 grid | `horizontal` | 2 vertical containers |
| 3 columns | `horizontal` | leaf + horizontal container |
| 3 rows | `vertical` | leaf + vertical container |

## Project Structure

```
tuxgo/
  .github/
    workflows/
      release.yml               # CI: auto-release on tag push
  cmd/
    tuxgo/
      main.go                 # Entry point
  internal/
    cli/
      root.go                 # Root command (create/attach session)
      init.go                 # 'tuxgo init' subcommand
      list.go                 # 'tuxgo list' subcommand
      version.go              # 'tuxgo version' subcommand
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
  .goreleaser.yaml
  Makefile
  go.mod
  go.sum
  README.md
  .gitignore
```

## Development

```bash
make build          # Compile binary
make install        # Install to $GOPATH/bin
make test           # Run all tests
make fmt            # Format code
make vet            # Run go vet
make lint           # Run golangci-lint
make clean          # Remove build artifacts
make next-version   # Show next version based on conventional commits
make release-dry    # Test goreleaser locally (no publish)
make release        # Tag + push to trigger GitHub Actions release
make help           # Show all targets
```

## Releases

Releases are fully automated using [Conventional Commits](https://www.conventionalcommits.org/), [GoReleaser](https://goreleaser.com/), and GitHub Actions.

### How it works

1. Write commits using conventional commit format (see below)
2. Run `make release` -- this analyzes commits since the last tag, determines the version bump, creates a git tag, and pushes it
3. GitHub Actions triggers on the tag push, runs tests, and uses GoReleaser to build cross-platform binaries and publish a GitHub Release

### Conventional Commits

Format: `<type>(<optional scope>): <description>`

| Prefix | Version bump | Example |
|--------|-------------|---------|
| `fix:` | Patch (0.0.x) | `fix: handle empty config gracefully` |
| `feat:` | Minor (0.x.0) | `feat: add session renaming support` |
| `feat!:` or `BREAKING CHANGE` | Major (x.0.0) | `feat!: change config format` |
| `docs:` | -- | `docs: update installation guide` |
| `chore:` | -- | `chore: update dependencies` |
| `refactor:` | -- | `refactor: extract session logic` |
| `test:` | -- | `test: add matcher edge cases` |

Only `feat`, `fix`, and breaking changes affect the version. Other types are valid conventional commits but do not trigger a version bump on their own -- a `fix:` or `feat:` commit must also be present.

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
