package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// ResurrectRestoredMsg carries the result of RestoreResurrect.
type ResurrectRestoredMsg struct{ Err error }

// StateCapturedMsg carries a snapshot of the current tmux state.
type StateCapturedMsg struct {
	// Sessions is the ordered list of session names.
	Sessions []string
	// Windows maps session name → windows in that session.
	Windows map[string][]Window
	// Dirs maps "session:windowIndex" → pane working directory.
	Dirs map[string]string
	Err  error
}

// RestoreResurrect runs the tmux-resurrect restore script for the given save file.
// It temporarily points the 'last' symlink at savePath, runs restore.sh, then
// restores the previous symlink target.
// Docs: https://github.com/tmux-plugins/tmux-resurrect
func RestoreResurrect(savePath string) tea.Cmd {
	return func() tea.Msg {
		restoreScript := filepath.Join(os.Getenv("HOME"), ".tmux", "plugins",
			"tmux-resurrect", "scripts", "restore.sh")
		if _, err := os.Stat(restoreScript); err != nil {
			return ResurrectRestoredMsg{Err: fmt.Errorf("restore script not found: %s", restoreScript)}
		}

		savesDir := filepath.Dir(savePath)
		lastLink := filepath.Join(savesDir, "last")

		// Read current symlink target so we can restore it afterward.
		prevTarget, _ := os.Readlink(lastLink)

		// Point 'last' at the chosen save.
		_ = os.Remove(lastLink)
		if err := os.Symlink(savePath, lastLink); err != nil {
			return ResurrectRestoredMsg{Err: fmt.Errorf("resurrect: symlink: %w", err)}
		}

		err := exec.Command("bash", restoreScript).Run()

		// Restore original symlink regardless of restore outcome.
		if prevTarget != "" {
			_ = os.Remove(lastLink)
			_ = os.Symlink(prevTarget, lastLink)
		}

		return ResurrectRestoredMsg{Err: err}
	}
}

// CaptureCurrentState snapshots all running sessions, their windows, and each
// window's active pane working directory. All tmux queries run synchronously
// inside a single goroutine and are returned in one message.
func CaptureCurrentState() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
		if err != nil {
			return StateCapturedMsg{Err: fmt.Errorf("capture: list-sessions: %w", err)}
		}

		var sessionNames []string
		for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
			if line != "" {
				sessionNames = append(sessionNames, line)
			}
		}

		windowsMap := make(map[string][]Window)
		dirsMap := make(map[string]string)

		for _, sess := range sessionNames {
			wOut, err := exec.Command("tmux", "list-windows", "-t", sess, "-F",
				"#{window_index}:#{window_name}:#{window_active}").Output()
			if err != nil {
				continue
			}
			windows := parseWindows(string(wOut))
			windowsMap[sess] = windows

			for _, w := range windows {
				target := fmt.Sprintf("%s:%d", sess, w.Index)
				dOut, err := exec.Command("tmux", "display-message", "-p", "-t",
					target, "#{pane_current_path}").Output()
				if err != nil {
					continue
				}
				dir := strings.TrimSpace(string(dOut))
				if dir != "" {
					key := fmt.Sprintf("%s:%d", sess, w.Index)
					dirsMap[key] = dir
				}
			}
		}

		return StateCapturedMsg{
			Sessions: sessionNames,
			Windows:  windowsMap,
			Dirs:     dirsMap,
		}
	}
}
