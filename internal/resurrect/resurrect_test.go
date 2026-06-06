package resurrect_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/marcelorc13/tmuxer/internal/resurrect"
)

const sampleFile = `pane	app	1	0	:	1	user@host:~/Projects/app	~/Projects/app	/home/user/Projects/app	1	nvim	:nvim
pane	app	2	0	:	1	user@host:~/Projects/app	~/Projects/app	/home/user/Projects/app	1	bash	:bash
window	app	1	:editor	0	:	abc,190x45,0,0,1	off
window	app	2	:server	1	:*	def,190x45,0,0,2	off
state	app
`

func makeSaveFile(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(sampleFile), 0o644); err != nil {
		t.Fatalf("write sample: %v", err)
	}
	return path
}

func TestListSaves(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	savesDir := filepath.Join(tmp, ".local", "share", "tmux", "resurrect")
	if err := os.MkdirAll(savesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	makeSaveFile(t, savesDir, "tmux_resurrect_20260601T120000.txt")
	makeSaveFile(t, savesDir, "tmux_resurrect_20260605T080000.txt")

	saves, err := resurrect.ListSaves()
	if err != nil {
		t.Fatalf("ListSaves: %v", err)
	}
	if len(saves) != 2 {
		t.Fatalf("expected 2 saves, got %d", len(saves))
	}
	// Newest first.
	if saves[0].Timestamp.Before(saves[1].Timestamp) {
		t.Error("saves not sorted newest-first")
	}
	// Sessions count.
	if saves[0].Sessions != 1 {
		t.Errorf("sessions: got %d want 1", saves[0].Sessions)
	}
}

func TestParseSave(t *testing.T) {
	tmp := t.TempDir()
	path := makeSaveFile(t, tmp, "tmux_resurrect_20260605T120000.txt")

	panes, windows, err := resurrect.ParseSave(path)
	if err != nil {
		t.Fatalf("ParseSave: %v", err)
	}
	if len(panes) != 2 {
		t.Errorf("panes: got %d want 2", len(panes))
	}
	if panes[0].Program != "nvim" {
		t.Errorf("pane[0] program: got %q want %q", panes[0].Program, "nvim")
	}
	if panes[0].WorkingDir != "/home/user/Projects/app" {
		t.Errorf("pane[0] dir: got %q", panes[0].WorkingDir)
	}
	if len(windows) != 2 {
		t.Errorf("windows: got %d want 2", len(windows))
	}
	if windows[1].Name != "server" {
		t.Errorf("window[1] name: got %q want %q", windows[1].Name, "server")
	}

	_ = time.Now() // ensure time import used
}
