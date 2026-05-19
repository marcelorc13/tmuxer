// Package tmux provides functions for interacting with tmux sessions.
package tmux

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	tea "charm.land/bubbletea/v2"
)

// Session represents a tmux session.
type Session struct {
	Name     string
	Attached bool
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

// ListSessions returns a Cmd that lists all tmux sessions.
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
func RenameSession(oldName, newName string) tea.Cmd {
	return func() tea.Msg {
		err := exec.Command("tmux", "rename-session", "-t", oldName, newName).Run()
		return SessionRenamedMsg{Err: err}
	}
}

// KillSession returns a Cmd that kills a session.
func KillSession(name string) tea.Cmd {
	return func() tea.Msg {
		err := exec.Command("tmux", "kill-session", "-t", name).Run()
		return SessionKilledMsg{Err: err}
	}
}

// AttachFunc returns a closure that replaces the current process with
// `tmux attach-session -t name` via syscall.Exec.
// Call this after tea.Program.Run() returns so BubbleTea can clean up the terminal first.
func AttachFunc(name string) func() {
	return func() {
		tmuxPath, err := exec.LookPath("tmux")
		if err != nil {
			return
		}
		args := []string{"tmux", "attach-session", "-t", name}
		_ = syscall.Exec(tmuxPath, args, os.Environ())
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
