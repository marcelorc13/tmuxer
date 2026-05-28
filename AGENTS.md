# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

TUI app for managing tmux sessions, windows, and panes — including tmux-resurrect integration. Built with Go + BubbleTea v2 (`charm.land/bubbletea/v2`).

## Commands

```bash
# Build
go build ./...

# Run
go run .

# Test
go test ./...

# Single test
go test ./internal/... -run TestFunctionName -v

# Lint (requires golangci-lint)
golangci-lint run
```

Use `make` for common dev tasks:

```bash
make build
make run
make test
make lint
```

## Architecture

Follows The Elm Architecture (MVU) via BubbleTea:

- **Model** — immutable state struct; no side effects
- **Update(msg) → (Model, Cmd)** — pure state transitions; Cmds produce async Msgs
- **View() → string** — pure render
- **Init() → tea.Cmd** — startup command (nil if none)

### Directory layout

```
tmux-tui/
├── main.go                  # Entry point — creates and starts tea.Program
├── go.mod / go.sum
├── Makefile                 # build, run, test, lint
└── internal/
    ├── tmux/                # tmux interaction layer (exec tmux commands, parse output)
    └── ui/
        ├── root.go          # root model, activeView enum, top-level Update/View
        ├── views/
        │   ├── session/     # session list screen
        │   ├── window/      # window list screen
        │   └── pane/        # pane detail screen
        └── components/
            ├── common/      # shared styles (lipgloss), vim keymap
            ├── spinner.go
            └── *.go         # reusable sub-models (helpbar, confirm modal, etc.)
```

### Key patterns

**Component model**: each view and component is its own `Model` struct implementing `tea.Model`. Parent composes children by embedding or holding them as fields and delegating `Update`/`View`.

**Component Update signature**: reusable components return their own concrete type (not `tea.Model`) so callers retain type safety:

```go
func (s Spinner) Update(msg tea.Msg) (Spinner, tea.Cmd) { ... }
```

**Tmux commands**: run via `exec.Command("tmux", ...)` returning `tea.Cmd` (wraps in goroutine). Result comes back as a custom `Msg` type — never block in `Update`.

**Tmux documentation**: authoritative reference is https://man.openbsd.org/tmux — use it when adding new tmux commands, checking flag semantics, or verifying format specifiers (`#{...}`). Each function in `internal/tmux/tmux.go` links to its relevant section. Target syntax: `session:window.pane` (e.g., `main:0.1`).

**Msgs vs Cmds**: `Msg` = data arriving (event/result); `Cmd` = function that produces a future `Msg`. Return `nil` Cmd when no async work needed. Use `tea.Batch` to fan out multiple commands.

**Navigation**: root model holds `activeView` enum; views delegate up via `Msg` types to trigger screen transitions.

**Key bindings**: vim-motion-compatible navigation across all views. Define with `key.NewBinding` from `github.com/charmbracelet/bubbles/key`; shared nav keymap lives in `internal/ui/components/common/keymap.go`.

```go
var keys = keyMap{
    Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
    Up:   key.NewBinding(key.WithKeys("k", "up"),     key.WithHelp("k/↑", "up")),
    Down: key.NewBinding(key.WithKeys("j", "down"),   key.WithHelp("j/↓", "down")),
    Top:  key.NewBinding(key.WithKeys("g"),            key.WithHelp("g", "top")),
    Bot:  key.NewBinding(key.WithKeys("G"),            key.WithHelp("G", "bottom")),
}
```

**Styling**: use `github.com/charmbracelet/lipgloss` for all colors/layout. Keep styles in `internal/ui/components/common/styles.go`.

**Shared state**: pass a config/store struct into each model at construction; no global mutable state.

### Entry point pattern

```go
// main.go
func main() {
    p := tea.NewProgram(ui.NewRootModel(), tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## Dependencies

| Package | Purpose |
|---|---|
| `charm.land/bubbletea/v2` | Core MVU framework |
| `github.com/charmbracelet/bubbles` | Prebuilt components (spinner, list, textinput, viewport) |
| `github.com/charmbracelet/lipgloss` | Terminal styling |

## Go conventions enforced here

- Exported types/funcs have doc comments
- Constructors named `New<Type>` returning the concrete type (not interface)
- Errors wrapped with `fmt.Errorf("context: %w", err)`
- No global mutable state — pass deps via constructor
