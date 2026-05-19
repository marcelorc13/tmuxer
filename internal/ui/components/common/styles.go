// Package common provides shared styles and keybindings for all views.
package common

import "charm.land/lipgloss/v2"

var (
	SessionNormal = lipgloss.NewStyle()

	SessionSelected = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	StatusAttached = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))

	StatusDetached = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	ModalBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#EF4444")).
			Padding(1, 2)

	HelpBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	ErrorText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444"))

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#A78BFA"))
)
