# Workspace Templates & Resurrect Integration

tmuxer supports two complementary features for saving and restoring your tmux workspace:

- **Templates** — named layouts you define (sessions, windows, directories, startup commands) that tmuxer can apply on demand.
- **Resurrect saves** — snapshots created by [tmux-resurrect](https://github.com/tmux-plugins/tmux-resurrect) that you can browse and restore from inside tmuxer.

---

## Templates

### Storage

Templates follow the [XDG Base Directory](https://specifications.freedesktop.org/basedir-spec/latest/) spec:

| Location | Purpose |
|---|---|
| `$XDG_DATA_HOME/tmuxer/templates/` | User templates (default: `~/.local/share/tmuxer/templates/`) |
| `/usr/share/tmuxer/templates/` | System-wide defaults (read-only) |

User templates shadow system templates with the same filename. `Save` and `Delete` always target the user directory.

### File Format

Each template is a JSON file named `<template-name>.json`. Spaces, slashes, and dots in the name are replaced with `_`.

```json
{
  "name": "my-workspace",
  "sessions": [
    {
      "name": "app",
      "windows": [
        {
          "name": "editor",
          "dir": "~/Projects/myapp",
          "command": "nvim"
        },
        {
          "name": "server",
          "dir": "~/Projects/myapp",
          "command": "go run ."
        },
        {
          "name": "shell",
          "dir": "~/Projects/myapp"
        }
      ]
    },
    {
      "name": "infra",
      "windows": [
        {
          "name": "logs",
          "dir": "/var/log",
          "command": "tail -f syslog"
        }
      ]
    }
  ]
}
```

**Fields:**

| Field | Required | Description |
|---|---|---|
| `name` | yes | Template display name |
| `sessions[].name` | yes | tmux session name to create |
| `sessions[].windows[].name` | yes | Window name |
| `sessions[].windows[].dir` | no | Starting directory (`~` expanded by tmux) |
| `sessions[].windows[].command` | no | Command sent to the pane on creation |

`command` is sent via `tmux send-keys` after the window opens. Use it for long-running processes (`nvim`, `go run .`, `lazygit`). Leave it empty for a plain shell.

### Editing Templates Directly

Templates are plain JSON — you can create or edit them in any text editor:

```bash
# Open user templates directory
$EDITOR ~/.local/share/tmuxer/templates/my-workspace.json
```

Changes take effect immediately the next time you open the Templates screen in tmuxer.

---

## Creating Templates in tmuxer

### Method 1 — Capture current state (`C`)

Press `C` from the Sessions screen. tmuxer snapshots all running sessions, their windows, and each window's working directory, then opens the wizard pre-filled with that data.

1. You only need to type a name for the template.
2. Review the captured sessions and windows.
3. Press `enter` to save.

This is the fastest way to save your current workspace as a reusable template.

### Method 2 — Wizard from scratch (`T` → `n`)

Press `T` to open the Templates screen, then `n` to launch the wizard.

The wizard has five steps:

```
Step 1 — Name
  Type a template name and press enter.

Step 2 — Add sessions
  Type a session name and press enter to add it.
  Repeat for each session you want.
  Press tab when done adding sessions.

Step 3 — Select session
  Use j/k to select which session to add windows to.
  Press enter to start adding windows.
  Press tab to skip to review.

Step 4 — Add windows
  For each window, fill in three fields (enter to advance each):
    • Window name
    • Working directory  (enter to skip, leaves it blank)
    • Startup command   (enter to skip)
  Press tab when done adding windows to the current session.
  You return to step 3 to select another session.

Step 5 — Review
  Inspect the full template tree.
  Press enter to save, esc to go back and edit.
```

**Navigation summary:**

| Key | Action |
|---|---|
| `enter` | Confirm / advance to next field or step |
| `tab` | Finish current section, advance to next step |
| `esc` | Go back one step (at step 1: cancel wizard) |
| `j` / `k` | Navigate session list (step 3) |
| `ctrl+c` | Cancel and discard |

---

## Applying a Template

1. Press `T` from the Sessions screen.
2. Use `j`/`k` to select a template.
3. Press `enter` — a confirmation prompt shows the number of sessions that will be created.
4. Press `y` to apply.

tmuxer runs the following for each entry in the template:

1. `tmux new-session -d -s <session>` — creates the session detached.
2. `tmux new-window -d -t <session> -n <name> -c <dir>` — creates each window with its directory.
3. `tmux send-keys -t <session>:<window> <command> Enter` — sends the startup command if set.

The sessions appear in the Sessions panel once tmuxer returns from the Templates screen.

### Deleting a Template

Select a template and press `d`. The file is removed from `~/.local/share/tmuxer/templates/`. System templates at `/usr/share/tmuxer/templates/` cannot be deleted from the TUI.

---

## Resurrect Saves

### What it is

[tmux-resurrect](https://github.com/tmux-plugins/tmux-resurrect) (installed via TPM) periodically saves your tmux environment to `~/.local/share/tmux/resurrect/`. tmuxer lets you browse those saves and restore any of them without leaving the TUI.

### Opening the browser (`R`)

Press `R` from the Sessions screen. tmuxer lists all saves newest-first, showing the timestamp and number of sessions in each file.

```
[ Resurrect Saves ]

> 2026-06-05 20:23:41  (3 sessions)
  2026-06-05 16:48:42  (2 sessions)
  2026-06-05 15:05:19  (3 sessions)
  2026-06-01 18:56:00  (1 session)
```

### Restoring a save

1. Use `j`/`k` to select a save.
2. Press `enter` — a confirmation prompt appears.
3. Press `y` to restore.

tmuxer temporarily repoints the `last` symlink to the chosen file, runs the TPM restore script (`~/.tmux/plugins/tmux-resurrect/scripts/restore.sh`), then restores the symlink to its previous target.

> **Note:** tmux-resurrect must be installed via TPM for restore to work. If the restore script is not found at `~/.tmux/plugins/tmux-resurrect/scripts/restore.sh`, tmuxer shows an error.

### Saves directory

Saves are stored at `~/.local/share/tmux/resurrect/`. The `last` symlink points to the most recently created save and is what tmux-resurrect uses by default.

---

## Key Reference

| Key | Screen | Action |
|---|---|---|
| `R` | Sessions | Open resurrect saves browser |
| `T` | Sessions | Open templates browser |
| `C` | Sessions | Capture current tmux state → wizard |
| `enter` | Resurrect / Templates | Restore / apply selected item |
| `n` | Templates | New template (wizard) |
| `d` | Templates | Delete selected template |
| `esc` / `q` | Any sub-screen | Return to Sessions |
