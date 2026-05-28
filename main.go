package main

import (
	"log"

	tea "charm.land/bubbletea/v2"

	"github.com/marceloramalhoc/tmux-tui/internal/ui"
)

func main() {
	p := tea.NewProgram(ui.NewModel())
	finalModel, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
	if m, ok := finalModel.(ui.Model); ok && m.PendingAttach != nil {
		m.PendingAttach()
	}
}
