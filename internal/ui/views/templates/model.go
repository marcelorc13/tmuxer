// Package templates provides the UI for browsing and applying workspace templates.
package templates

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	tmuxcmd "github.com/marcelorc13/tmuxer/internal/tmux"
	"github.com/marcelorc13/tmuxer/internal/templates"
	"github.com/marcelorc13/tmuxer/internal/ui/components"
	"github.com/marcelorc13/tmuxer/internal/ui/components/common"
	"github.com/marcelorc13/tmuxer/internal/ui/uimsgs"
)

type state int

const (
	stateList state = iota
	stateConfirming
	stateApplying
	stateError
)

// TemplatesLoadedMsg carries the result of listing templates from disk.
type TemplatesLoadedMsg struct {
	Templates []templates.Template
	Err       error
}

// TemplateAppliedMsg signals that all apply commands have been dispatched.
type TemplateAppliedMsg struct{ Err error }

// OpenWizardMsg signals the root to open the wizard (Phase 6).
type OpenWizardMsg struct{ Draft templates.Template }

// Model is the templates browser.
type Model struct {
	items   []templates.Template
	cursor  int
	state   state
	confirm components.ConfirmModal
	err     error
	width   int
	height  int

	// apply progress tracking
	pending int // commands still in flight
	applyErr error
}

// New creates a Model ready to use.
func New(width, height int) Model {
	return Model{
		width:   width,
		height:  height,
		confirm: components.NewConfirmModal(),
	}
}

// Init loads templates from disk.
func (m Model) Init() tea.Cmd {
	return loadTemplates()
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case TemplatesLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = stateError
			return m, nil
		}
		m.items = msg.Templates
		m.state = stateList

	case tmuxcmd.SessionCreatedMsg:
		if msg.Err != nil {
			m.applyErr = msg.Err
		}
		m.pending--
		if m.pending <= 0 {
			return m, applied(m.applyErr)
		}

	case tmuxcmd.WindowCreatedMsg:
		if msg.Err != nil {
			m.applyErr = msg.Err
		}
		m.pending--
		if m.pending <= 0 {
			return m, applied(m.applyErr)
		}

	case tmuxcmd.KeysSentMsg:
		if msg.Err != nil {
			m.applyErr = msg.Err
		}
		m.pending--
		if m.pending <= 0 {
			return m, applied(m.applyErr)
		}

	case TemplateAppliedMsg:
		return m, back()

	case components.ConfirmedMsg:
		if m.state == stateConfirming && m.cursor < len(m.items) {
			m.state = stateApplying
			m.applyErr = nil
			cmds, count := applyTemplate(m.items[m.cursor])
			m.pending = count
			if count == 0 {
				return m, back()
			}
			return m, tea.Batch(cmds...)
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
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}

	case key.Matches(msg, common.Keys.Top):
		m.cursor = 0

	case key.Matches(msg, common.Keys.Bottom):
		if len(m.items) > 0 {
			m.cursor = len(m.items) - 1
		}

	case key.Matches(msg, common.Keys.Enter):
		if len(m.items) > 0 && m.state == stateList {
			t := m.items[m.cursor]
			sessions := len(t.Sessions)
			noun := "session"
			if sessions != 1 {
				noun = "sessions"
			}
			prompt := fmt.Sprintf("Apply '%s'? Creates %d %s", t.Name, sessions, noun)
			m.confirm = m.confirm.Show(prompt)
			m.state = stateConfirming
		}

	case key.Matches(msg, common.Keys.New):
		if m.state == stateList {
			return m, openWizard(templates.Template{})
		}

	case key.Matches(msg, common.Keys.Kill):
		if len(m.items) > 0 && m.state == stateList {
			name := m.items[m.cursor].Name
			if err := templates.Delete(name); err != nil {
				m.err = err
				m.state = stateError
			} else {
				return m, loadTemplates()
			}
		}

	case key.Matches(msg, common.Keys.Esc), key.Matches(msg, common.Keys.Quit):
		return m, back()
	}
	return m, nil
}

// View renders the templates browser.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(common.Title.Render("Templates") + "\n\n")

	if m.state == stateConfirming {
		b.WriteString(m.confirm.View())
		return b.String()
	}

	if m.state == stateApplying {
		b.WriteString(common.SessionNormal.Render("Applying template...") + "\n")
		return b.String()
	}

	if m.state == stateError && m.err != nil {
		b.WriteString(common.ErrorText.Render(m.err.Error()) + "\n")
	}

	if len(m.items) == 0 {
		b.WriteString(common.SessionNormal.Render("no templates — press n to create one") + "\n")
	} else {
		for i, t := range m.items {
			sessions := fmt.Sprintf("%d session", len(t.Sessions))
			if len(t.Sessions) != 1 {
				sessions += "s"
			}
			line := fmt.Sprintf("%-24s  %s", t.Name, sessions)
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
		{Key: "enter", Desc: "apply"},
		{Key: "n", Desc: "new"},
		{Key: "d", Desc: "delete"},
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

// applyTemplate returns a batch of tmux commands to create all sessions and
// windows defined in the template. Returns the commands and total count.
func applyTemplate(t templates.Template) ([]tea.Cmd, int) {
	var cmds []tea.Cmd
	for _, sess := range t.Sessions {
		cmds = append(cmds, tmuxcmd.NewSession(sess.Name))
		for _, w := range sess.Windows {
			cmds = append(cmds, tmuxcmd.NewWindowWithDir(sess.Name, w.Name, w.Dir))
			if w.Command != "" {
				target := sess.Name + ":" + w.Name
				cmds = append(cmds, tmuxcmd.SendKeys(target, w.Command))
			}
		}
	}
	return cmds, len(cmds)
}

func loadTemplates() tea.Cmd {
	return func() tea.Msg {
		items, err := templates.ListAll()
		return TemplatesLoadedMsg{Templates: items, Err: err}
	}
}

func back() tea.Cmd {
	return func() tea.Msg { return uimsgs.BackMsg{} }
}

func applied(err error) tea.Cmd {
	return func() tea.Msg { return TemplateAppliedMsg{Err: err} }
}

func openWizard(draft templates.Template) tea.Cmd {
	return func() tea.Msg { return OpenWizardMsg{Draft: draft} }
}
