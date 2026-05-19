// Package pane provides the pane detail view (stub).
package pane

import tea "charm.land/bubbletea/v2"

// Model is the pane detail view.
type Model struct{}

// NewModel creates a pane detail Model.
func NewModel() Model { return Model{} }

// Init is a no-op stub.
func (m Model) Init() tea.Cmd { return nil }

// Update is a no-op stub.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) { return m, nil }

// View renders a placeholder.
func (m Model) View() string { return "panes view — not yet implemented\n" }
