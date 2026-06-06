// Package templates manages workspace template definitions stored as JSON.
// Storage follows the XDG Base Directory spec:
//   - User templates: $XDG_DATA_HOME/tmuxer/templates/ (default ~/.local/share/tmuxer/templates/)
//   - System templates: /usr/share/tmuxer/templates/
//
// User templates shadow system templates of the same name.
package templates

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const systemTemplatesDir = "/usr/share/tmuxer/templates"

// TemplateWindow defines a single tmux window within a template session.
type TemplateWindow struct {
	Name    string `json:"name"`
	Dir     string `json:"dir"`
	Command string `json:"command,omitempty"`
}

// TemplateSession defines a tmux session within a template.
type TemplateSession struct {
	Name    string           `json:"name"`
	Windows []TemplateWindow `json:"windows"`
}

// Template is a named workspace layout.
type Template struct {
	Name     string            `json:"name"`
	Sessions []TemplateSession `json:"sessions"`
}

// UserTemplatesDir returns the XDG-compliant user templates directory and
// creates it if it does not exist.
func UserTemplatesDir() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("templates: home dir: %w", err)
		}
		base = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(base, "tmuxer", "templates")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("templates: mkdir: %w", err)
	}
	return dir, nil
}

// Save writes a template to the user templates directory, overwriting any
// existing file with the same name.
func Save(t Template) error {
	dir, err := UserTemplatesDir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("templates: marshal: %w", err)
	}
	path := filepath.Join(dir, sanitizeName(t.Name)+".json")
	if err := os.WriteFile(path, data, 0o640); err != nil {
		return fmt.Errorf("templates: write: %w", err)
	}
	return nil
}

// Load reads a template by name. User dir is checked first; system dir is the
// fallback.
func Load(name string) (Template, error) {
	filename := sanitizeName(name) + ".json"

	userDir, err := UserTemplatesDir()
	if err != nil {
		return Template{}, err
	}

	for _, dir := range []string{userDir, systemTemplatesDir} {
		data, err := os.ReadFile(filepath.Join(dir, filename))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return Template{}, fmt.Errorf("templates: read %q: %w", name, err)
		}
		var t Template
		if err := json.Unmarshal(data, &t); err != nil {
			return Template{}, fmt.Errorf("templates: parse %q: %w", name, err)
		}
		return t, nil
	}
	return Template{}, fmt.Errorf("templates: %q not found", name)
}

// ListAll returns all templates from both the user and system dirs. User
// templates shadow system templates with the same sanitized filename.
func ListAll() ([]Template, error) {
	userDir, err := UserTemplatesDir()
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{} // tracks filenames already added
	var out []Template

	for _, dir := range []string{userDir, systemTemplatesDir} {
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("templates: readdir %q: %w", dir, err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			if _, ok := seen[e.Name()]; ok {
				continue // user template already added; skip system duplicate
			}
			seen[e.Name()] = struct{}{}
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			var t Template
			if err := json.Unmarshal(data, &t); err != nil {
				continue
			}
			out = append(out, t)
		}
	}
	return out, nil
}

// Rename loads a template by oldName, saves it under newName, then deletes the
// old file. Both names must be non-empty and distinct.
func Rename(oldName, newName string) error {
	if oldName == newName || newName == "" {
		return nil
	}
	t, err := Load(oldName)
	if err != nil {
		return err
	}
	t.Name = newName
	if err := Save(t); err != nil {
		return err
	}
	return Delete(oldName)
}

// Delete removes a template from the user templates directory.
func Delete(name string) error {
	dir, err := UserTemplatesDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, sanitizeName(name)+".json")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("templates: delete %q: %w", name, err)
	}
	return nil
}

// sanitizeName converts a template name to a safe filename.
func sanitizeName(name string) string {
	r := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		" ", "_",
		".", "_",
	)
	return r.Replace(name)
}
