// Package components provides reusable UI components.
package components

import (
	"strings"

	"github.com/marceloramalhoc/tmux-tui/internal/ui/components/common"
)

// HelpBar renders a one-line key hint bar.
type HelpBar struct{}

// NewHelpBar creates a HelpBar.
func NewHelpBar() HelpBar { return HelpBar{} }

// Hint is a key/description pair for the help bar.
type Hint struct {
	Key  string
	Desc string
}

// View renders hints as a row of key/description pairs in order.
func (h HelpBar) View(hints []Hint) string {
	parts := make([]string, 0, len(hints))
	for _, hint := range hints {
		parts = append(parts, common.HelpBarStyle.Render(hint.Key+" "+hint.Desc))
	}
	return strings.Join(parts, common.HelpBarStyle.Render("  ·  "))
}
