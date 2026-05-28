package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/marcelorc13/tmuxer/internal/ui/components"
	"github.com/marcelorc13/tmuxer/internal/ui/components/common"
)

// View renders the active screen.
func (m Model) View() tea.View {
	var content string
	if m.state == stateConfirmingKill || m.state == stateConfirmingKillWindow {
		content = common.Title.Render("Sessions") + "\n\n" + m.confirm.View()
	} else {
		// Each panel: half terminal width, full terminal height. Border = 2 each side.
		leftOuter := m.width / 2
		rightOuter := m.width - leftOuter
		innerH := max(0, m.height-3)
		leftInner := max(0, leftOuter-2)
		rightInner := max(0, rightOuter-2)

		leftStyle := common.PanelInactive.Width(leftInner).Height(innerH)
		rightStyle := common.PanelInactive.Width(rightInner).Height(innerH)
		if m.focus == focusSession {
			leftStyle = common.PanelActive.Width(leftInner).Height(innerH)
		} else {
			rightStyle = common.PanelActive.Width(rightInner).Height(innerH)
		}

		left := m.renderSessionPanel(innerH)
		right := m.renderWindowPanel(innerH)
		content = lipgloss.JoinHorizontal(lipgloss.Top, leftStyle.Render(left), rightStyle.Render(right))
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m Model) renderSessionPanel(height int) string {
	var body strings.Builder

	body.WriteString(common.Title.Render("Sessions") + "\n\n")

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
		body.WriteString(row + "\n")
	}

	if len(m.sessions) == 0 && m.state == stateNormal {
		body.WriteString(common.SessionNormal.Render("no sessions") + "\n")
	}

	if m.state == stateCreatingNew {
		body.WriteString("\nNew session: " + m.textInput.View() + "\n")
	}

	if m.err != nil {
		body.WriteString("\n" + common.ErrorText.Render(m.err.Error()) + "\n")
	}

	hints := []components.Hint{
		{Key: "↑/k ↓/j", Desc: "navigate"},
		{Key: "→/l", Desc: "windows"},
		{Key: "a", Desc: "attach"},
		{Key: "n", Desc: "new"},
		{Key: "r", Desc: "rename"},
		{Key: "d", Desc: "kill"},
		{Key: "q", Desc: "quit"},
	}
	hintsStr := components.NewHelpBar().View(hints)

	if m.focus != focusSession {
		// no hints; just pad to fill height
		bodyStr := body.String()
		bodyLines := strings.Count(bodyStr, "\n")
		pad := height - bodyLines - 1
		if pad > 0 {
			bodyStr += strings.Repeat("\n", pad)
		}
		return bodyStr
	}

	hintLines := strings.Count(hintsStr, "\n") + 1
	bodyStr := body.String()
	bodyLines := strings.Count(bodyStr, "\n")
	pad := height - bodyLines - hintLines
	if pad > 0 {
		bodyStr += strings.Repeat("\n", pad)
	}

	return bodyStr + hintsStr
}

func (m Model) renderWindowPanel(height int) string {
	var body strings.Builder

	body.WriteString(common.Title.Render("Windows") + "\n\n")

	if len(m.sessions) == 0 {
		body.WriteString(common.SessionNormal.Render("no session selected") + "\n")
	} else if len(m.windows) == 0 && m.state != stateCreatingNewWindow {
		body.WriteString(common.SessionNormal.Render("no windows") + "\n")
	} else {
		for i, w := range m.windows {
			indicator := " "
			if w.Active {
				indicator = common.StatusAttached.Render("*")
			}

			label := fmt.Sprintf("%d: %s", w.Index, w.Name)
			var row string
			if i == m.windowCursor && m.focus == focusWindow {
				if m.state == stateRenamingWindow {
					cursor := common.SessionSelected.Render("> ")
					row = fmt.Sprintf("%s%s %d: %s", cursor, indicator, w.Index, m.textInput.View())
				} else {
					cursor := common.SessionSelected.Render("> ")
					name := common.SessionSelected.Render(label)
					row = fmt.Sprintf("%s%s %s", cursor, indicator, name)
				}
			} else if i == m.windowCursor {
				cursor := common.SessionSelected.Render("> ")
				row = fmt.Sprintf("%s%s %s", cursor, indicator, label)
			} else {
				row = fmt.Sprintf("  %s %s", indicator, label)
			}
			body.WriteString(row + "\n")
		}
	}

	if m.state == stateCreatingNewWindow {
		body.WriteString("\nNew window: " + m.textInput.View() + "\n")
	}

	hints := []components.Hint{
		{Key: "↑/k ↓/j", Desc: "navigate"},
		{Key: "←/h", Desc: "sessions"},
		{Key: "a", Desc: "attach"},
		{Key: "n", Desc: "new"},
		{Key: "r", Desc: "rename"},
		{Key: "d", Desc: "kill"},
		{Key: "q", Desc: "quit"},
	}
	hintsStr := components.NewHelpBar().View(hints)

	if m.focus != focusWindow {
		// no hints; just pad to fill height
		bodyStr := body.String()
		bodyLines := strings.Count(bodyStr, "\n")
		pad := height - bodyLines - 1
		if pad > 0 {
			bodyStr += strings.Repeat("\n", pad)
		}
		return bodyStr
	}

	hintLines := strings.Count(hintsStr, "\n") + 1
	bodyStr := body.String()
	bodyLines := strings.Count(bodyStr, "\n")
	pad := height - bodyLines - hintLines
	if pad > 0 {
		bodyStr += strings.Repeat("\n", pad)
	}

	return bodyStr + hintsStr
}
