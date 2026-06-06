// Package resurrect provides the UI for browsing and restoring tmux-resurrect saves.
package resurrect

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/marcelorc13/tmuxer/internal/resurrect"
	tmuxcmd "github.com/marcelorc13/tmuxer/internal/tmux"
	"github.com/marcelorc13/tmuxer/internal/ui/components"
	"github.com/marcelorc13/tmuxer/internal/ui/components/common"
	"github.com/marcelorc13/tmuxer/internal/ui/uimsgs"
)

type state int

const (
	stateList state = iota
	stateConfirming
	stateRestoring
	stateError
)

// SavesLoadedMsg carries the result of the initial list load.
type SavesLoadedMsg struct {
	Saves []resurrect.Save
	Err   error
}

// Model is the resurrect saves browser.
type Model struct {
	saves   []resurrect.Save
	cursor  int
	state   state
	confirm components.ConfirmModal
	err     error
	width   int
	height  int
}

// New creates a Model ready to use.
func New(width, height int) Model {
	return Model{
		width:   width,
		height:  height,
		confirm: components.NewConfirmModal(),
	}
}

// Init loads the list of saves.
func (m Model) Init() tea.Cmd {
	return loadSaves()
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case SavesLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = stateError
			return m, nil
		}
		m.saves = msg.Saves
		m.state = stateList

	case tmuxcmd.ResurrectRestoredMsg:
		m.state = stateList
		if msg.Err != nil {
			m.err = msg.Err
			m.state = stateError
		}
		return m, nil

	case components.ConfirmedMsg:
		if m.state == stateConfirming && m.cursor < len(m.saves) {
			m.state = stateRestoring
			return m, tmuxcmd.RestoreResurrect(m.saves[m.cursor].Path)
		}

	case components.CancelledMsg:
		m.state = stateList

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if m.state == stateConfirming {
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.Update(msg)
		return m, cmd
	}

	switch {
	case key.Matches(msg, common.Keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, common.Keys.Down):
		if m.cursor < len(m.saves)-1 {
			m.cursor++
		}

	case key.Matches(msg, common.Keys.Top):
		m.cursor = 0

	case key.Matches(msg, common.Keys.Bottom):
		if len(m.saves) > 0 {
			m.cursor = len(m.saves) - 1
		}

	case key.Matches(msg, common.Keys.Enter):
		if len(m.saves) > 0 && m.state == stateList {
			save := m.saves[m.cursor]
			prompt := fmt.Sprintf("Restore save from %s?",
				save.Timestamp.Format("2006-01-02 15:04:05"))
			m.confirm = m.confirm.Show(prompt)
			m.state = stateConfirming
		}

	case key.Matches(msg, common.Keys.Esc), key.Matches(msg, common.Keys.Quit):
		return m, back()
	}
	return m, nil
}

// View renders the resurrect saves browser.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(common.Title.Render("Resurrect Saves") + "\n\n")

	if m.state == stateConfirming {
		b.WriteString(m.confirm.View())
		return b.String()
	}

	if m.state == stateRestoring {
		b.WriteString(common.SessionNormal.Render("Restoring...") + "\n")
		return b.String()
	}

	if m.state == stateError && m.err != nil {
		b.WriteString(common.ErrorText.Render(m.err.Error()) + "\n")
	}

	if len(m.saves) == 0 {
		b.WriteString(common.SessionNormal.Render("no saves found") + "\n")
	} else {
		for i, s := range m.saves {
			ts := s.Timestamp.Format("2006-01-02 15:04:05")
			sessions := fmt.Sprintf("%d session", s.Sessions)
			if s.Sessions != 1 {
				sessions += "s"
			}
			line := fmt.Sprintf("%s  (%s)", ts, sessions)
			if i == m.cursor {
				b.WriteString(common.SessionSelected.Render("> "+line) + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
		}
	}

	innerH := max(0, m.height-3)
	bodyStr := b.String()
	bodyLines := strings.Count(bodyStr, "\n")

	hints := []components.Hint{
		{Key: "↑/k ↓/j", Desc: "navigate"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "enter", Desc: "restore"},
		{Key: "esc/q", Desc: "back"},
	}
	hintsStr := components.NewHelpBar().View(hints)
	hintLines := strings.Count(hintsStr, "\n") + 1

	pad := innerH - bodyLines - hintLines
	if pad > 0 {
		bodyStr += strings.Repeat("\n", pad)
	}
	return bodyStr + hintsStr
}

// loadSaves is the tea.Cmd that fetches saves from disk.
func loadSaves() tea.Cmd {
	return func() tea.Msg {
		saves, err := resurrect.ListSaves()
		return SavesLoadedMsg{Saves: saves, Err: err}
	}
}

// back emits a BackMsg to the root model.
func back() tea.Cmd {
	return func() tea.Msg { return uimsgs.BackMsg{} }
}
