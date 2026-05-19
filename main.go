package main

import (
	"log"

	tea "charm.land/bubbletea/v2"

	"github.com/marceloramalhoc/tmux-tui/internal/ui"
)

func main() {
	root := ui.NewRootModel()
	p := tea.NewProgram(root)
	finalModel, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
	if rm, ok := finalModel.(ui.RootModel); ok && rm.PendingAttach != nil {
		rm.PendingAttach()
	}
}
