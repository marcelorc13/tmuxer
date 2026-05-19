// Package session provides the session list view.
package session

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/marceloramalhoc/tmux-tui/internal/tmux"
	"github.com/marceloramalhoc/tmux-tui/internal/ui/components"
	"github.com/marceloramalhoc/tmux-tui/internal/ui/components/common"
)

type viewState int

const (
	stateNormal viewState = iota
	stateRenaming
	stateCreatingNew
	stateConfirmingKill
	stateLoading
)

// AttachRequestMsg is sent to the root model when the user wants to attach to a session.
type AttachRequestMsg struct{ Name string }

// Model is the session list view.
type Model struct {
	sessions  []tmux.Session
	cursor    int
	state     viewState
	textInput *textinput.Model
	confirm   components.ConfirmModal
	err       error
	width     int
	height    int
}

// NewModel creates a session list Model.
func NewModel() Model {
	ti := textinput.New()
	ti.CharLimit = 64
	return Model{
		state:     stateNormal,
		textInput: &ti,
		confirm:   components.NewConfirmModal(),
	}
}

// Init triggers the initial session list load.
func (m Model) Init() tea.Cmd {
	return tmux.ListSessions()
}

// Update handles messages and returns the updated model and any commands.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tmux.SessionsLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.sessions = msg.Sessions
		m.state = stateNormal

		if m.cursor >= len(m.sessions) {
			m.cursor = max(0, len(m.sessions)-1)
		}

		return m, nil

	case tmux.SessionCreatedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}
		return m, tmux.ListSessions()

	case tmux.SessionRenamedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}
		return m, tmux.ListSessions()

	case tmux.SessionKilledMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}
		return m, tmux.ListSessions()

	case components.ConfirmedMsg:
		if m.state == stateConfirmingKill && m.cursor < len(m.sessions) {
			name := m.sessions[m.cursor].Name
			m.state = stateLoading
			return m, tmux.KillSession(name)
		}
		return m, nil

	case components.CancelledMsg:
		m.state = stateNormal
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if m.state == stateConfirmingKill {
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.Update(msg)
		return m, cmd
	}

	if m.state == stateRenaming || m.state == stateCreatingNew {
		return m.handleTextInput(msg)
	}

	switch {
	// navigation
	case key.Matches(msg, common.Keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, common.Keys.Down):
		if m.cursor < len(m.sessions)-1 {
			m.cursor++
		}

	case key.Matches(msg, common.Keys.Top):
		m.cursor = 0

	case key.Matches(msg, common.Keys.Bottom):
		m.cursor = max(0, len(m.sessions)-1)

	// actions
	case key.Matches(msg, common.Keys.Enter):
		if len(m.sessions) > 0 {
			name := m.sessions[m.cursor].Name
			return m, func() tea.Msg { return AttachRequestMsg{Name: name} }
		}

	case key.Matches(msg, common.Keys.New):
		m.state = stateCreatingNew
		m.textInput.Reset()
		m.textInput.Placeholder = "session name (empty = auto)"
		m.textInput.Focus()

	case key.Matches(msg, common.Keys.Rename):
		if len(m.sessions) > 0 {
			m.state = stateRenaming
			m.textInput.Reset()
			m.textInput.SetValue(m.sessions[m.cursor].Name)
			m.textInput.CursorEnd()
			m.textInput.Focus()
		}

	case key.Matches(msg, common.Keys.Kill):
		if len(m.sessions) > 0 {
			m.state = stateConfirmingKill
			msg := fmt.Sprintf("Kill '%s'?", m.sessions[m.cursor].Name)
			m.confirm = m.confirm.Show(msg)
		}

	case key.Matches(msg, common.Keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleTextInput(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		val := strings.TrimSpace(m.textInput.Value())
		m.textInput.Blur()
		if m.state == stateCreatingNew {
			m.state = stateLoading
			return m, tmux.NewSession(val)
		}
		if m.state == stateRenaming && len(m.sessions) > 0 {
			old := m.sessions[m.cursor].Name
			m.state = stateLoading
			if val != "" && val != old {
				return m, tmux.RenameSession(old, val)
			}
			return m, tmux.ListSessions()
		}

	case "esc":
		m.state = stateNormal
		m.textInput.Blur()
		return m, nil
	default:
		var cmd tea.Cmd
		updated, cmd := m.textInput.Update(msg)
		*m.textInput = updated
		return m, cmd
	}

	return m, nil
}

// View renders the session list screen.
func (m Model) View() string {
	var sb strings.Builder

	sb.WriteString(common.Title.Render("Sessions") + "\n\n")

	if m.state == stateConfirmingKill {
		sb.WriteString(m.confirm.View())
		return sb.String()
	}

	for i, s := range m.sessions {
		indicator := common.StatusDetached.Render("")
		if s.Attached {
			indicator = common.StatusAttached.Render("*")
		}

		var row string
		if m.state == stateRenaming && i == m.cursor {
			cursor := common.SessionSelected.Render("> ")
			row = fmt.Sprintf("%s%s %s", cursor, indicator, m.textInput.View())
		} else if i == m.cursor {
			cursor := common.SessionSelected.Render("> ")
			name := common.SessionSelected.Render(s.Name)
			row = fmt.Sprintf("%s%s %s", cursor, indicator, name)
		} else {
			row = fmt.Sprintf("  %s %s", indicator, s.Name)
		}
		sb.WriteString(row + "\n")
	}

	if len(m.sessions) == 0 && m.state == stateNormal {
		sb.WriteString(common.SessionNormal.Render("no sessions") + "\n")
	}

	if m.state == stateCreatingNew {
		sb.WriteString("\nNew session: " + m.textInput.View() + "\n")
	}

	if m.err != nil {
		sb.WriteString("\n" + common.ErrorText.Render(m.err.Error()) + "\n")
	}

	hints := []components.Hint{
		{Key: "↑/k ↓/j", Desc: "navigate"},
		{Key: "enter", Desc: "attach"},
		{Key: "n", Desc: "new"},
		{Key: "r", Desc: "rename"},
		{Key: "d", Desc: "kill"},
		{Key: "q", Desc: "quit"},
	}
	sb.WriteString("\n")
	sb.WriteString(components.NewHelpBar().View(hints))

	return sb.String()
}
