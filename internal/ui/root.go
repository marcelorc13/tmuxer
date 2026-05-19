// Package ui provides the root BubbleTea model.
package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/marceloramalhoc/tmux-tui/internal/tmux"
	"github.com/marceloramalhoc/tmux-tui/internal/ui/views/pane"
	"github.com/marceloramalhoc/tmux-tui/internal/ui/views/session"
	"github.com/marceloramalhoc/tmux-tui/internal/ui/views/window"
)

// ViewID identifies which view is currently active.
type ViewID int

const (
	SessionView ViewID = iota
	WindowView
	PaneView
)

// NavigateMsg triggers a screen transition.
type NavigateMsg struct{ To ViewID }

// RootModel is the top-level BubbleTea model.
type RootModel struct {
	activeView  ViewID
	sessionView session.Model
	windowView  window.Model
	paneView    pane.Model
	width       int
	height      int
	// PendingAttach holds a func to exec tmux after p.Run() returns.
	PendingAttach func()
}

// NewRootModel creates a RootModel ready to run.
func NewRootModel() RootModel {
	return RootModel{
		activeView:  SessionView,
		sessionView: session.NewModel(),
		windowView:  window.NewModel(),
		paneView:    pane.NewModel(),
	}
}

// Init starts the session list loading.
func (m RootModel) Init() tea.Cmd {
	return m.sessionView.Init()
}

// Update routes messages to the active view and handles cross-cutting concerns.
func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		var cmd tea.Cmd
		m.sessionView, cmd = m.sessionView.Update(msg)
		return m, cmd

	case session.AttachRequestMsg:
		m.PendingAttach = tmux.AttachFunc(msg.Name)
		return m, tea.Quit

	case NavigateMsg:
		m.activeView = msg.To
		return m, nil
	}

	switch m.activeView {
	case SessionView:
		var cmd tea.Cmd
		m.sessionView, cmd = m.sessionView.Update(msg)
		return m, cmd

	case WindowView:
		var cmd tea.Cmd
		m.windowView, cmd = m.windowView.Update(msg)
		return m, cmd

	case PaneView:
		var cmd tea.Cmd
		m.paneView, cmd = m.paneView.Update(msg)
		return m, cmd
	}
	return m, nil
}

// View renders the active view inside an alt-screen.
func (m RootModel) View() tea.View {
	var content string
	switch m.activeView {
	case SessionView:
		content = m.sessionView.View()

	case WindowView:
		content = m.windowView.View()

	case PaneView:
		content = m.paneView.View()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
