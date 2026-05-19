package components

import (
	tea "charm.land/bubbletea/v2"
	"github.com/marceloramalhoc/tmux-tui/internal/ui/components/common"
)

// ConfirmedMsg is sent when the user confirms the action.
type ConfirmedMsg struct{}

// CancelledMsg is sent when the user cancels the action.
type CancelledMsg struct{}

// ConfirmModal is a modal overlay that asks the user to confirm an action.
type ConfirmModal struct {
	visible bool
	message string
}

// NewConfirmModal creates a ConfirmModal.
func NewConfirmModal() ConfirmModal { return ConfirmModal{} }

// Show makes the modal visible with the given confirmation message.
func (c ConfirmModal) Show(message string) ConfirmModal {
	c.visible = true
	c.message = message
	return c
}

// Hide hides the modal.
func (c ConfirmModal) Hide() ConfirmModal {
	c.visible = false
	return c
}

// IsVisible reports whether the modal is currently shown.
func (c ConfirmModal) IsVisible() bool { return c.visible }

// Update handles key input when the modal is visible.
func (c ConfirmModal) Update(msg tea.Msg) (ConfirmModal, tea.Cmd) {
	kp, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return c, nil
	}
	switch kp.String() {
	case "y", "Y":
		c = c.Hide()
		return c, func() tea.Msg { return ConfirmedMsg{} }
	case "n", "N", "esc":
		c = c.Hide()
		return c, func() tea.Msg { return CancelledMsg{} }
	}
	return c, nil
}

// View renders the modal box or an empty string if not visible.
func (c ConfirmModal) View() string {
	if !c.visible {
		return ""
	}
	return common.ModalBox.Render(c.message + " [y/N]")
}
