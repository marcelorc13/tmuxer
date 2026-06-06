// Package resurrect reads tmux-resurrect save files.
package resurrect

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Save represents a single tmux-resurrect save file.
type Save struct {
	Path      string
	Timestamp time.Time
	Sessions  int // number of distinct session names in the file
}

// PaneRecord holds a parsed pane line from a resurrect file.
type PaneRecord struct {
	Session    string
	WindowIdx  string
	PaneIdx    string
	WorkingDir string
	Program    string
	Command    string
}

// WindowRecord holds a parsed window line from a resurrect file.
type WindowRecord struct {
	Session   string
	WindowIdx string
	Name      string
	Active    bool
}

// SavesDir returns the default tmux-resurrect saves directory.
func SavesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "tmux", "resurrect")
}

// ListSaves returns all save files sorted by timestamp descending (newest first).
func ListSaves() ([]Save, error) {
	dir := SavesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("resurrect: readdir %q: %w", dir, err)
	}

	var saves []Save
	for _, e := range entries {
		if e.IsDir() || e.Name() == "last" {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "tmux_resurrect_") || !strings.HasSuffix(name, ".txt") {
			continue
		}
		// filename: tmux_resurrect_20260605T202341.txt
		raw := strings.TrimPrefix(strings.TrimSuffix(name, ".txt"), "tmux_resurrect_")
		ts, err := time.ParseInLocation("20060102T150405", raw, time.Local)
		if err != nil {
			continue
		}
		path := filepath.Join(dir, name)
		sessions := countSessions(path)
		saves = append(saves, Save{
			Path:      path,
			Timestamp: ts,
			Sessions:  sessions,
		})
	}

	sort.Slice(saves, func(i, j int) bool {
		return saves[i].Timestamp.After(saves[j].Timestamp)
	})
	return saves, nil
}

// ParseSave parses a resurrect file into pane and window records.
func ParseSave(path string) ([]PaneRecord, []WindowRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("resurrect: open %q: %w", path, err)
	}
	defer f.Close()

	var panes []PaneRecord
	var windows []WindowRecord

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "pane":
			if len(fields) < 12 {
				continue
			}
			panes = append(panes, PaneRecord{
				Session:    fields[1],
				WindowIdx:  fields[2],
				PaneIdx:    fields[3],
				WorkingDir: fields[8],
				Program:    fields[10],
				Command:    fields[11],
			})
		case "window":
			if len(fields) < 6 {
				continue
			}
			windows = append(windows, WindowRecord{
				Session:   fields[1],
				WindowIdx: fields[2],
				Name:      strings.TrimPrefix(fields[3], ":"),
				Active:    len(fields) > 5 && fields[5] == ":*",
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("resurrect: scan %q: %w", path, err)
	}
	return panes, windows, nil
}

// countSessions counts distinct session names in a save file.
func countSessions(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	seen := map[string]struct{}{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) >= 2 && (fields[0] == "pane" || fields[0] == "window") {
			seen[fields[1]] = struct{}{}
		}
	}
	return len(seen)
}
