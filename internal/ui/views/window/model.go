// Package window provides the window list view (stub).
package window

import tea "charm.land/bubbletea/v2"

// Model is the window list view.
type Model struct{}

// NewModel creates a window list Model.
func NewModel() Model { return Model{} }

// Init is a no-op stub.
func (m Model) Init() tea.Cmd { return nil }

// Update is a no-op stub.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) { return m, nil }

// View renders a placeholder.
func (m Model) View() string { return "windows view — not yet implemented\n" }
