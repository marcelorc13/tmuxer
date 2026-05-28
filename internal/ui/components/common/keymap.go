package common

import "charm.land/bubbles/v2/key"

// NavKeyMap defines vim-motion compatible keybindings shared across views.
type NavKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Top       key.Binding
	Bottom    key.Binding
	Quit      key.Binding
	Enter     key.Binding
	Attach    key.Binding
	FocusNext key.Binding
	FocusPrev key.Binding
	New       key.Binding
	Rename    key.Binding
	Kill      key.Binding
	Esc       key.Binding
}

// Keys is the default navigation keymap.
var Keys = NavKeyMap{
	Up:        key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
	Down:      key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
	Top:       key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
	Bottom:    key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Enter:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Attach:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "attach")),
	FocusNext: key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("l/→", "windows")),
	FocusPrev: key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("h/←", "sessions")),
	New:       key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
	Rename:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename")),
	Kill:      key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "kill")),
	Esc:       key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
}
