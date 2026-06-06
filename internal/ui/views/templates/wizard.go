package templates

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/marcelorc13/tmuxer/internal/templates"
	"github.com/marcelorc13/tmuxer/internal/ui/components"
	"github.com/marcelorc13/tmuxer/internal/ui/components/common"
	"github.com/marcelorc13/tmuxer/internal/ui/uimsgs"
)

type wizardStep int

const (
	stepName         wizardStep = iota // template name
	stepAddSession                     // type session name, enter to add
	stepSelectSession                  // pick session to add windows to
	stepAddWindow                      // sub-steps: window name → dir → command
	stepReview                         // show full draft, enter to save
)

type windowField int

const (
	fieldWindowName windowField = iota
	fieldWindowDir
	fieldWindowCmd
)

// WizardSavedMsg signals a template was saved successfully.
type WizardSavedMsg struct{ Name string }

// Wizard is the multi-step template creation form.
type Wizard struct {
	step         wizardStep
	windowField  windowField
	draft        templates.Template
	sessionCursor int   // which session is selected in stepSelectSession
	draftWindow  templates.TemplateWindow // window being built in stepAddWindow
	captureMode  bool  // true when pre-filled from CaptureCurrentState
	err          error
	input        textinput.Model
	width        int
	height       int
}

// NewWizard creates an empty wizard ready for scratch creation.
func NewWizard(width, height int) Wizard {
	return newWizard(templates.Template{}, false, width, height)
}

// NewWizardFromCapture creates a wizard pre-filled with a captured template.
// The user only needs to provide a name; the rest is shown in review.
func NewWizardFromCapture(draft templates.Template, width, height int) Wizard {
	return newWizard(draft, true, width, height)
}

func newWizard(draft templates.Template, captureMode bool, width, height int) Wizard {
	ti := textinput.New()
	ti.CharLimit = 64
	ti.Placeholder = "template name"
	ti.Focus()
	return Wizard{
		step:        stepName,
		draft:       draft,
		captureMode: captureMode,
		input:       ti,
		width:       width,
		height:      height,
	}
}

// Update handles messages for the wizard.
func (w Wizard) Update(msg tea.Msg) (Wizard, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.width, w.height = msg.Width, msg.Height

	case tea.KeyPressMsg:
		return w.handleKey(msg)
	}
	return w, nil
}

func (w Wizard) handleKey(msg tea.KeyPressMsg) (Wizard, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return w, func() tea.Msg { return uimsgs.BackMsg{} }

	case "esc":
		// Go back one step; at first step, cancel the wizard entirely.
		switch w.step {
		case stepName:
			return w, func() tea.Msg { return uimsgs.BackMsg{} }
		case stepAddSession:
			w.step = stepName
			w.focusInput("template name", w.draft.Name)
		case stepSelectSession:
			w.step = stepAddSession
			w.focusInput("session name (enter to add, tab when done)", "")
		case stepAddWindow:
			w.step = stepSelectSession
			w.input.Blur()
		case stepReview:
			if w.captureMode {
				w.step = stepName
				w.focusInput("template name", w.draft.Name)
			} else {
				w.step = stepSelectSession
				w.input.Blur()
			}
		}
		return w, nil

	case "enter":
		return w.handleEnter()

	case "tab":
		return w.handleTab()
	}

	// Delegate typing to input when it accepts text.
	if w.inputActive() {
		var cmd tea.Cmd
		updated, cmd := w.input.Update(msg)
		w.input = updated
		return w, cmd
	}

	// stepSelectSession: j/k navigation.
	switch msg.String() {
	case "j", "down":
		if w.step == stepSelectSession && w.sessionCursor < len(w.draft.Sessions)-1 {
			w.sessionCursor++
		}
	case "k", "up":
		if w.step == stepSelectSession && w.sessionCursor > 0 {
			w.sessionCursor--
		}
	}
	return w, nil
}

func (w Wizard) handleEnter() (Wizard, tea.Cmd) {
	val := strings.TrimSpace(w.input.Value())

	switch w.step {
	case stepName:
		if val == "" {
			return w, nil
		}
		w.draft.Name = val
		if w.captureMode {
			// Skip session/window building — go straight to review.
			w.step = stepReview
			w.input.Blur()
		} else {
			w.step = stepAddSession
			w.focusInput("session name (enter to add, tab when done)", "")
		}

	case stepAddSession:
		if val != "" {
			w.draft.Sessions = append(w.draft.Sessions, templates.TemplateSession{Name: val})
			w.focusInput("session name (enter to add, tab when done)", "")
		}

	case stepSelectSession:
		if len(w.draft.Sessions) > 0 {
			w.step = stepAddWindow
			w.windowField = fieldWindowName
			w.draftWindow = templates.TemplateWindow{}
			w.focusInput("window name (enter to add, tab when done)", "")
		}

	case stepAddWindow:
		return w.handleWindowField(val)

	case stepReview:
		return w.save()
	}
	return w, nil
}

func (w Wizard) handleWindowField(val string) (Wizard, tea.Cmd) {
	switch w.windowField {
	case fieldWindowName:
		if val == "" {
			return w, nil
		}
		w.draftWindow.Name = val
		w.windowField = fieldWindowDir
		w.focusInput("working directory (enter to skip)", "")

	case fieldWindowDir:
		w.draftWindow.Dir = val
		w.windowField = fieldWindowCmd
		w.focusInput("startup command (enter to skip)", "")

	case fieldWindowCmd:
		w.draftWindow.Command = val
		// Append window to currently selected session.
		w.draft.Sessions[w.sessionCursor].Windows = append(
			w.draft.Sessions[w.sessionCursor].Windows, w.draftWindow,
		)
		// Loop back to window name for the next window.
		w.windowField = fieldWindowName
		w.draftWindow = templates.TemplateWindow{}
		w.focusInput("window name (enter to add, tab when done)", "")
	}
	return w, nil
}

func (w Wizard) handleTab() (Wizard, tea.Cmd) {
	switch w.step {
	case stepAddSession:
		if len(w.draft.Sessions) == 0 {
			return w, nil // need at least one session
		}
		w.step = stepSelectSession
		w.sessionCursor = 0
		w.input.Blur()

	case stepAddWindow:
		// Tab in window name field = done adding windows to this session.
		w.step = stepSelectSession
		w.input.Blur()

	case stepSelectSession:
		// Tab from session list = go to review.
		w.step = stepReview
		w.input.Blur()
	}
	return w, nil
}

func (w Wizard) save() (Wizard, tea.Cmd) {
	if err := templates.Save(w.draft); err != nil {
		w.err = err
		return w, nil
	}
	name := w.draft.Name
	return w, func() tea.Msg { return WizardSavedMsg{Name: name} }
}

// View renders the wizard.
func (w Wizard) View() string {
	var b strings.Builder

	title := fmt.Sprintf("New Template  •  %s", w.stepLabel())
	b.WriteString(common.Title.Render(title) + "\n\n")

	if w.err != nil {
		b.WriteString(common.ErrorText.Render(w.err.Error()) + "\n\n")
	}

	switch w.step {
	case stepName:
		b.WriteString("Template name:\n")
		b.WriteString("  " + w.input.View() + "\n")

	case stepAddSession:
		b.WriteString("Sessions added:\n")
		for _, s := range w.draft.Sessions {
			b.WriteString(common.StatusAttached.Render("  + ") + s.Name + "\n")
		}
		if len(w.draft.Sessions) == 0 {
			b.WriteString(common.StatusDetached.Render("  (none yet)") + "\n")
		}
		b.WriteString("\nAdd session name:\n")
		b.WriteString("  " + w.input.View() + "\n")

	case stepSelectSession:
		b.WriteString("Select session to add windows to:\n\n")
		for i, s := range w.draft.Sessions {
			winCount := fmt.Sprintf("(%d windows)", len(s.Windows))
			line := fmt.Sprintf("%-20s  %s", s.Name, winCount)
			if i == w.sessionCursor {
				b.WriteString(common.SessionSelected.Render("> "+line) + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
		}

	case stepAddWindow:
		sess := w.draft.Sessions[w.sessionCursor]
		b.WriteString(fmt.Sprintf("Adding windows to session %q:\n\n",
			common.SessionSelected.Render(sess.Name)))
		for _, win := range sess.Windows {
			b.WriteString(common.StatusAttached.Render("  + ") + win.Name)
			if win.Dir != "" {
				b.WriteString("  " + common.StatusDetached.Render(win.Dir))
			}
			if win.Command != "" {
				b.WriteString("  " + common.StatusDetached.Render("$ "+win.Command))
			}
			b.WriteString("\n")
		}
		if len(sess.Windows) == 0 {
			b.WriteString(common.StatusDetached.Render("  (none yet)") + "\n")
		}
		b.WriteString("\n" + w.windowFieldLabel() + ":\n")
		b.WriteString("  " + w.input.View() + "\n")

	case stepReview:
		b.WriteString(common.SessionNormal.Render("Review your template:\n\n"))
		b.WriteString(fmt.Sprintf("  Name:  %s\n\n", common.SessionSelected.Render(w.draft.Name)))
		for _, sess := range w.draft.Sessions {
			b.WriteString(fmt.Sprintf("  Session: %s\n", common.SessionSelected.Render(sess.Name)))
			for _, win := range sess.Windows {
				dir := win.Dir
				if dir == "" {
					dir = "~"
				}
				b.WriteString(fmt.Sprintf("    • %-18s  %s", win.Name, common.StatusDetached.Render(dir)))
				if win.Command != "" {
					b.WriteString("  " + common.StatusDetached.Render("$ "+win.Command))
				}
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
		b.WriteString("Press enter to save, esc to go back.\n")
	}

	innerH := max(0, w.height-3)
	bodyStr := b.String()
	bodyLines := strings.Count(bodyStr, "\n")

	hints := w.hints()
	hintsStr := components.NewHelpBar().View(hints)
	hintLines := strings.Count(hintsStr, "\n") + 1

	pad := innerH - bodyLines - hintLines
	if pad > 0 {
		bodyStr += strings.Repeat("\n", pad)
	}
	return bodyStr + hintsStr
}

func (w Wizard) stepLabel() string {
	steps := []string{"Name", "Sessions", "Select Session", "Windows", "Review"}
	idx := int(w.step)
	if idx >= len(steps) {
		idx = len(steps) - 1
	}
	return fmt.Sprintf("Step %d/%d — %s", idx+1, len(steps), steps[idx])
}

func (w Wizard) windowFieldLabel() string {
	switch w.windowField {
	case fieldWindowName:
		return "Window name"
	case fieldWindowDir:
		return "Working directory"
	case fieldWindowCmd:
		return "Startup command"
	}
	return ""
}

func (w Wizard) inputActive() bool {
	return w.step == stepName ||
		w.step == stepAddSession ||
		w.step == stepAddWindow
}

func (w *Wizard) focusInput(placeholder, value string) {
	w.input.Placeholder = placeholder
	w.input.Reset()
	w.input.SetValue(value)
	w.input.CursorEnd()
	w.input.Focus()
}

func (w Wizard) hints() []components.Hint {
	switch w.step {
	case stepName:
		return []components.Hint{
			{Key: "enter", Desc: "next"},
			{Key: "esc", Desc: "cancel"},
		}
	case stepAddSession:
		return []components.Hint{
			{Key: "enter", Desc: "add session"},
			{Key: "tab", Desc: "next step"},
			{Key: "esc", Desc: "back"},
		}
	case stepSelectSession:
		return []components.Hint{
			{Key: "↑/k ↓/j", Desc: "navigate"},
			{Key: "enter", Desc: "add windows"},
			{Key: "tab", Desc: "review"},
			{Key: "esc", Desc: "back"},
		}
	case stepAddWindow:
		return []components.Hint{
			{Key: "enter", Desc: "next field"},
			{Key: "tab", Desc: "done with session"},
			{Key: "esc", Desc: "back"},
		}
	case stepReview:
		return []components.Hint{
			{Key: "enter", Desc: "save"},
			{Key: "esc", Desc: "back"},
		}
	}
	return nil
}
