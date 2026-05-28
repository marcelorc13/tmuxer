// Package tmux provides functions for interacting with tmux sessions, windows, and panes.
package tmux

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	tea "charm.land/bubbletea/v2"
)

// Session represents a tmux session.
type Session struct {
	Name     string
	Attached bool
}

// Window represents a tmux window.
type Window struct {
	Index  int
	Name   string
	Active bool
}

// Pane represents a tmux pane.
type Pane struct {
	Index  int
	Active bool
	Width  int
	Height int
}

// SessionsLoadedMsg carries the result of ListSessions.
type SessionsLoadedMsg struct {
	Sessions []Session
	Err      error
}

// SessionCreatedMsg carries the result of NewSession.
type SessionCreatedMsg struct{ Err error }

// SessionRenamedMsg carries the result of RenameSession.
type SessionRenamedMsg struct{ Err error }

// SessionKilledMsg carries the result of KillSession.
type SessionKilledMsg struct{ Err error }

// WindowsLoadedMsg carries the result of ListWindows.
type WindowsLoadedMsg struct {
	Session string
	Windows []Window
	Err     error
}

// WindowCreatedMsg carries the result of NewWindow.
type WindowCreatedMsg struct{ Err error }

// WindowRenamedMsg carries the result of RenameWindow.
type WindowRenamedMsg struct{ Err error }

// WindowKilledMsg carries the result of KillWindow.
type WindowKilledMsg struct{ Err error }

// PanesLoadedMsg carries the result of ListPanes.
type PanesLoadedMsg struct {
	Session string
	Window  int
	Panes   []Pane
	Err     error
}

// PaneSplitMsg carries the result of SplitPane.
type PaneSplitMsg struct{ Err error }

// PaneKilledMsg carries the result of KillPane.
type PaneKilledMsg struct{ Err error }

// ListSessions returns a Cmd that lists all tmux sessions.
// Docs: https://man.openbsd.org/tmux#list-sessions
func ListSessions() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("tmux", "list-sessions", "-F",
			"#{session_name}:#{session_attached}").Output()
		if err != nil {
			// tmux exits non-zero when no server is running — treat as empty list.
			return SessionsLoadedMsg{Sessions: nil, Err: nil}
		}
		return SessionsLoadedMsg{Sessions: parseSessions(string(out))}
	}
}

// NewSession returns a Cmd that creates a new detached session.
// If name is empty, tmux assigns an auto-generated name.
// Docs: https://man.openbsd.org/tmux#new-session
func NewSession(name string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"new-session", "-d"}
		if name != "" {
			args = append(args, "-s", name)
		}
		err := exec.Command("tmux", args...).Run()
		return SessionCreatedMsg{Err: err}
	}
}

// RenameSession returns a Cmd that renames a session.
// Docs: https://man.openbsd.org/tmux#rename-session
func RenameSession(oldName, newName string) tea.Cmd {
	return func() tea.Msg {
		err := exec.Command("tmux", "rename-session", "-t", oldName, newName).Run()
		return SessionRenamedMsg{Err: err}
	}
}

// KillSession returns a Cmd that kills a session.
// Docs: https://man.openbsd.org/tmux#kill-session
func KillSession(name string) tea.Cmd {
	return func() tea.Msg {
		err := exec.Command("tmux", "kill-session", "-t", name).Run()
		return SessionKilledMsg{Err: err}
	}
}

// AttachFunc returns a closure that attaches to the named session.
// When already inside tmux ($TMUX is set), uses switch-client to avoid nested
// session errors. Otherwise replaces the current process via syscall.Exec.
// Call after tea.Program.Run() returns so BubbleTea cleans up the terminal first.
// Docs: https://man.openbsd.org/tmux#attach-session, https://man.openbsd.org/tmux#switch-client
func AttachFunc(name string) func() {
	return func() {
		tmuxPath, err := exec.LookPath("tmux")
		if err != nil {
			return
		}
		if os.Getenv("TMUX") != "" {
			// Inside tmux: switch the current client to the target session.
			_ = exec.Command(tmuxPath, "switch-client", "-t", name).Run()
			return
		}
		args := []string{"tmux", "attach-session", "-t", name}
		_ = syscall.Exec(tmuxPath, args, os.Environ())
	}
}

// ListWindows returns a Cmd that lists all windows in the given session.
// Docs: https://man.openbsd.org/tmux#list-windows
func ListWindows(session string) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("tmux", "list-windows", "-t", session, "-F",
			"#{window_index}:#{window_name}:#{window_active}").Output()
		if err != nil {
			return WindowsLoadedMsg{Session: session, Err: err}
		}
		return WindowsLoadedMsg{Session: session, Windows: parseWindows(string(out))}
	}
}

// NewWindow returns a Cmd that creates a new window in the given session.
// If name is empty, tmux uses the default name.
// Docs: https://man.openbsd.org/tmux#new-window
func NewWindow(session, name string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"new-window", "-d",  "-t", session}
		if name != "" {
			args = append(args, "-n", name)
		}
		err := exec.Command("tmux", args...).Run()
		return WindowCreatedMsg{Err: err}
	}
}

// RenameWindow returns a Cmd that renames a window by index.
// Docs: https://man.openbsd.org/tmux#rename-window
func RenameWindow(session string, windowIndex int, newName string) tea.Cmd {
	return func() tea.Msg {
		target := session + ":" + strconv.Itoa(windowIndex)
		err := exec.Command("tmux", "rename-window", "-t", target, newName).Run()
		return WindowRenamedMsg{Err: err}
	}
}

// KillWindow returns a Cmd that kills a window by index.
// Docs: https://man.openbsd.org/tmux#kill-window
func KillWindow(session string, windowIndex int) tea.Cmd {
	return func() tea.Msg {
		target := session + ":" + strconv.Itoa(windowIndex)
		err := exec.Command("tmux", "kill-window", "-t", target).Run()
		return WindowKilledMsg{Err: err}
	}
}

// ListPanes returns a Cmd that lists all panes in session:window.
// Docs: https://man.openbsd.org/tmux#list-panes
func ListPanes(session string, windowIndex int) tea.Cmd {
	return func() tea.Msg {
		target := session + ":" + strconv.Itoa(windowIndex)
		out, err := exec.Command("tmux", "list-panes", "-t", target, "-F",
			"#{pane_index}:#{pane_active}:#{pane_width}:#{pane_height}").Output()
		if err != nil {
			return PanesLoadedMsg{Session: session, Window: windowIndex, Err: err}
		}
		return PanesLoadedMsg{
			Session: session,
			Window:  windowIndex,
			Panes:   parsePanes(string(out)),
		}
	}
}

// SplitPane returns a Cmd that splits a pane in session:window.
// vertical=true splits top/bottom (-v), false splits left/right (-h).
// Docs: https://man.openbsd.org/tmux#split-window
func SplitPane(session string, windowIndex int, vertical bool) tea.Cmd {
	return func() tea.Msg {
		target := session + ":" + strconv.Itoa(windowIndex)
		dir := "-h"
		if vertical {
			dir = "-v"
		}
		err := exec.Command("tmux", "split-window", dir, "-t", target).Run()
		return PaneSplitMsg{Err: err}
	}
}

// KillPane returns a Cmd that kills a specific pane.
// Docs: https://man.openbsd.org/tmux#kill-pane
func KillPane(session string, windowIndex, paneIndex int) tea.Cmd {
	return func() tea.Msg {
		target := session + ":" + strconv.Itoa(windowIndex) + "." + strconv.Itoa(paneIndex)
		err := exec.Command("tmux", "kill-pane", "-t", target).Run()
		return PaneKilledMsg{Err: err}
	}
}

func parseSessions(out string) []Session {
	var sessions []Session
	for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		sessions = append(sessions, Session{
			Name:     parts[0],
			Attached: parts[1] == "1",
		})
	}
	return sessions
}

func parseWindows(out string) []Window {
	var windows []Window
	for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}
		idx, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		windows = append(windows, Window{
			Index:  idx,
			Name:   parts[1],
			Active: parts[2] == "1",
		})
	}
	return windows
}

func parsePanes(out string) []Pane {
	var panes []Pane
	for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 4)
		if len(parts) != 4 {
			continue
		}
		idx, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		w, _ := strconv.Atoi(parts[2])
		h, _ := strconv.Atoi(parts[3])
		panes = append(panes, Pane{
			Index:  idx,
			Active: parts[1] == "1",
			Width:  w,
			Height: h,
		})
	}
	return panes
}
