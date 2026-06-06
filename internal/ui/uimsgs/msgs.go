// Package uimsgs holds message types shared between the root ui package and sub-views.
// It has no dependencies on ui or sub-view packages to avoid import cycles.
package uimsgs

// BackMsg is sent by sub-views to signal a return to the sessions screen.
type BackMsg struct{}
