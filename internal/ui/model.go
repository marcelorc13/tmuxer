// Package ui provides the top-level BubbleTea model.
package ui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/marcelorc13/tmuxer/internal/tmux"
	"github.com/marcelorc13/tmuxer/internal/ui/components"
	"github.com/marcelorc13/tmuxer/internal/ui/components/common"
)

type viewState int

const (
	stateNormal viewState = iota
	stateRenaming
	stateCreatingNew
	stateConfirmingKill
	stateLoading
	stateRenamingWindow
	stateConfirmingKillWindow
	stateCreatingNewWindow
)

type focusPanel int

const (
	focusSession focusPanel = iota
	focusWindow
)

// Model is the top-level BubbleTea model.
type Model struct {
	sessions      []tmux.Session
	cursor        int
	windows       []tmux.Window
	windowCursor  int
	focus         focusPanel
	state         viewState
	textInput     *textinput.Model
	confirm       components.ConfirmModal
	err           error
	width         int
	height        int
	PendingAttach func()
}

// NewModel creates a Model ready to run.
func NewModel() Model {
	ti := textinput.New()
	ti.CharLimit = 64
	return Model{
		state:     stateNormal,
		focus:     focusSession,
		textInput: &ti,
		confirm:   components.NewConfirmModal(),
	}
}

// Init triggers the initial session list load.
func (m Model) Init() tea.Cmd {
	return tmux.ListSessions()
}

// Update handles messages and returns the updated model and any commands.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		if len(m.sessions) > 0 {
			return m, tmux.ListWindows(m.sessions[m.cursor].Name)
		}
		return m, nil

	case tmux.WindowsLoadedMsg:
		if msg.Err == nil && len(m.sessions) > 0 && msg.Session == m.sessions[m.cursor].Name {
			m.windows = msg.Windows
			if m.windowCursor >= len(m.windows) {
				m.windowCursor = max(0, len(m.windows)-1)
			}
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

	case tmux.WindowRenamedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}
		if len(m.sessions) > 0 {
			return m, tmux.ListWindows(m.sessions[m.cursor].Name)
		}
		return m, nil

	case tmux.WindowCreatedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}
		if len(m.sessions) > 0 {
			return m, tmux.ListWindows(m.sessions[m.cursor].Name)
		}
		return m, nil

	case tmux.WindowKilledMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}
		if len(m.sessions) > 0 {
			return m, tmux.ListWindows(m.sessions[m.cursor].Name)
		}
		return m, nil

	case components.ConfirmedMsg:
		switch m.state {
		case stateConfirmingKill:
			if m.cursor < len(m.sessions) {
				name := m.sessions[m.cursor].Name
				m.state = stateLoading
				return m, tmux.KillSession(name)
			}
		case stateConfirmingKillWindow:
			if len(m.sessions) > 0 && m.windowCursor < len(m.windows) {
				sess := m.sessions[m.cursor].Name
				idx := m.windows[m.windowCursor].Index
				m.state = stateLoading
				return m, tmux.KillWindow(sess, idx)
			}
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
	if m.state == stateConfirmingKill || m.state == stateConfirmingKillWindow {
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.Update(msg)
		return m, cmd
	}

	if m.state == stateRenaming || m.state == stateCreatingNew {
		return m.handleTextInput(msg)
	}

	if m.state == stateRenamingWindow || m.state == stateCreatingNewWindow {
		return m.handleWindowRenameInput(msg)
	}

	if m.focus == focusWindow {
		return m.handleWindowKey(msg)
	}

	return m.handleSessionKey(msg)
}

func (m Model) handleSessionKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, common.Keys.Up):
		if m.cursor > 0 {
			m.cursor--
			m.windowCursor = 0
			if len(m.sessions) > 0 {
				return m, tmux.ListWindows(m.sessions[m.cursor].Name)
			}
		}

	case key.Matches(msg, common.Keys.Down):
		if m.cursor < len(m.sessions)-1 {
			m.cursor++
			m.windowCursor = 0
			return m, tmux.ListWindows(m.sessions[m.cursor].Name)
		}

	case key.Matches(msg, common.Keys.Top):
		m.cursor = 0
		m.windowCursor = 0
		if len(m.sessions) > 0 {
			return m, tmux.ListWindows(m.sessions[m.cursor].Name)
		}

	case key.Matches(msg, common.Keys.Bottom):
		m.cursor = max(0, len(m.sessions)-1)
		m.windowCursor = 0
		if len(m.sessions) > 0 {
			return m, tmux.ListWindows(m.sessions[m.cursor].Name)
		}

	case key.Matches(msg, common.Keys.FocusNext):
		if len(m.sessions) > 0 && len(m.windows) > 0 {
			m.focus = focusWindow
		}

	case key.Matches(msg, common.Keys.Attach):
		if len(m.sessions) > 0 {
			name := m.sessions[m.cursor].Name
			m.PendingAttach = tmux.AttachFunc(name)
			return m, tea.Quit
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
			prompt := fmt.Sprintf("Kill '%s'?", m.sessions[m.cursor].Name)
			m.confirm = m.confirm.Show(prompt)
		}

	case key.Matches(msg, common.Keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleWindowKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, common.Keys.Up):
		if m.windowCursor > 0 {
			m.windowCursor--
		}

	case key.Matches(msg, common.Keys.Down):
		if m.windowCursor < len(m.windows)-1 {
			m.windowCursor++
		}

	case key.Matches(msg, common.Keys.Top):
		m.windowCursor = 0

	case key.Matches(msg, common.Keys.Bottom):
		m.windowCursor = max(0, len(m.windows)-1)

	case key.Matches(msg, common.Keys.Attach):
		if len(m.sessions) > 0 && len(m.windows) > 0 {
			target := m.sessions[m.cursor].Name + ":" + strconv.Itoa(m.windows[m.windowCursor].Index)
			m.PendingAttach = tmux.AttachFunc(target)
			return m, tea.Quit
		}

	case key.Matches(msg, common.Keys.New):
		if len(m.sessions) > 0 {
			m.state = stateCreatingNewWindow
			m.textInput.Reset()
			m.textInput.Placeholder = "window name (empty = auto)"
			m.textInput.Focus()
		}

	case key.Matches(msg, common.Keys.Rename):
		if len(m.windows) > 0 {
			m.state = stateRenamingWindow
			m.textInput.Reset()
			m.textInput.SetValue(m.windows[m.windowCursor].Name)
			m.textInput.CursorEnd()
			m.textInput.Focus()
		}

	case key.Matches(msg, common.Keys.Kill):
		if len(m.windows) > 0 {
			m.state = stateConfirmingKillWindow
			prompt := fmt.Sprintf("Kill window '%s'?", m.windows[m.windowCursor].Name)
			m.confirm = m.confirm.Show(prompt)
		}

	case key.Matches(msg, common.Keys.FocusPrev):
		m.focus = focusSession

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

func (m Model) handleWindowRenameInput(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		val := strings.TrimSpace(m.textInput.Value())
		m.textInput.Blur()
		prev := m.state
		m.state = stateNormal
		if len(m.sessions) == 0 {
			return m, nil
		}
		sess := m.sessions[m.cursor].Name
		if prev == stateCreatingNewWindow {
			return m, tmux.NewWindow(sess, val)
		}
		if val != "" && len(m.windows) > 0 && val != m.windows[m.windowCursor].Name {
			return m, tmux.RenameWindow(sess, m.windows[m.windowCursor].Index, val)
		}
		return m, tmux.ListWindows(sess)

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
}
