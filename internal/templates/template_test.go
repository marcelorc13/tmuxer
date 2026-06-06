package templates_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/marcelorc13/tmuxer/internal/templates"
)

// setXDG points XDG_DATA_HOME at a temp dir so tests don't touch real data.
func setXDG(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)
	return tmp
}

func TestSaveLoadDelete(t *testing.T) {
	setXDG(t)

	want := templates.Template{
		Name: "test-workspace",
		Sessions: []templates.TemplateSession{
			{
				Name: "app",
				Windows: []templates.TemplateWindow{
					{Name: "editor", Dir: "~/Projects/app", Command: "nvim"},
					{Name: "server", Dir: "~/Projects/app"},
				},
			},
		},
	}

	if err := templates.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := templates.Load("test-workspace")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("name: got %q want %q", got.Name, want.Name)
	}
	if len(got.Sessions) != 1 || len(got.Sessions[0].Windows) != 2 {
		t.Errorf("structure mismatch: %+v", got)
	}
	if got.Sessions[0].Windows[0].Command != "nvim" {
		t.Errorf("command: got %q want %q", got.Sessions[0].Windows[0].Command, "nvim")
	}

	all, err := templates.ListAll()
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("ListAll len: got %d want 1", len(all))
	}

	if err := templates.Delete("test-workspace"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	all2, _ := templates.ListAll()
	if len(all2) != 0 {
		t.Errorf("after Delete: expected 0 templates, got %d", len(all2))
	}
}

func TestSanitizeName(t *testing.T) {
	tmp := setXDG(t)

	tpl := templates.Template{Name: "my workspace/v2", Sessions: nil}
	if err := templates.Save(tpl); err != nil {
		t.Fatalf("Save: %v", err)
	}

	dir := filepath.Join(tmp, "tmuxer", "templates")
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}
	if entries[0].Name() != "my_workspace_v2.json" {
		t.Errorf("filename: got %q", entries[0].Name())
	}
}

func TestUserTemplatesDir_XDG(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)

	dir, err := templates.UserTemplatesDir()
	if err != nil {
		t.Fatalf("UserTemplatesDir: %v", err)
	}
	want := filepath.Join(tmp, "tmuxer", "templates")
	if dir != want {
		t.Errorf("dir: got %q want %q", dir, want)
	}
}

func TestUserTemplatesDir_FallbackHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", tmp)

	dir, err := templates.UserTemplatesDir()
	if err != nil {
		t.Fatalf("UserTemplatesDir: %v", err)
	}
	want := filepath.Join(tmp, ".local", "share", "tmuxer", "templates")
	if dir != want {
		t.Errorf("dir: got %q want %q", dir, want)
	}
}
