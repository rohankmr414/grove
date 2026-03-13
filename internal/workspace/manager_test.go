package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{name: "auth-feature", want: "workspace/auth-feature"},
		{name: "bugfix/login-redirect", want: "bugfix/login-redirect"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := workspaceBranch(tt.name); got != tt.want {
				t.Fatalf("workspaceBranch(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestSeedWorkspaceRootCopiesFilesFromWorkspaceInitDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	initDir := filepath.Join(home, ".config", "grove", "workspace-init")
	if err := os.MkdirAll(initDir, 0o755); err != nil {
		t.Fatalf("create workspace init dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(initDir, "CLAUDE.md"), []byte("custom instructions\n"), 0o640); err != nil {
		t.Fatalf("write CLAUDE.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(initDir, ".tool-versions"), []byte("go 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write .tool-versions: %v", err)
	}

	ws := Workspace{
		Name: "feature-x",
		Path: filepath.Join(t.TempDir(), "feature-x"),
	}
	if err := os.MkdirAll(ws.Path, 0o755); err != nil {
		t.Fatalf("create workspace dir: %v", err)
	}

	manager := Manager{}
	if err := manager.seedWorkspaceRoot(ws); err != nil {
		t.Fatalf("seed workspace root: %v", err)
	}

	claudeData, err := os.ReadFile(filepath.Join(ws.Path, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read copied CLAUDE.md: %v", err)
	}
	if string(claudeData) != "custom instructions\n" {
		t.Fatalf("CLAUDE.md contents = %q", string(claudeData))
	}

	toolVersionsData, err := os.ReadFile(filepath.Join(ws.Path, ".tool-versions"))
	if err != nil {
		t.Fatalf("read copied .tool-versions: %v", err)
	}
	if string(toolVersionsData) != "go 1.25.0\n" {
		t.Fatalf(".tool-versions contents = %q", string(toolVersionsData))
	}
}

func TestSeedWorkspaceRootIgnoresMissingWorkspaceInitDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ws := Workspace{
		Name: "feature-x",
		Path: filepath.Join(t.TempDir(), "feature-x"),
	}
	if err := os.MkdirAll(ws.Path, 0o755); err != nil {
		t.Fatalf("create workspace dir: %v", err)
	}

	manager := Manager{}
	if err := manager.seedWorkspaceRoot(ws); err != nil {
		t.Fatalf("seed workspace root: %v", err)
	}

	entries, err := os.ReadDir(ws.Path)
	if err != nil {
		t.Fatalf("read workspace dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty workspace dir, found %d entries", len(entries))
	}
}

func TestSeedWorkspaceRootRejectsReservedMarkerFilename(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	initDir := filepath.Join(home, ".config", "grove", "workspace-init")
	if err := os.MkdirAll(initDir, 0o755); err != nil {
		t.Fatalf("create workspace init dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(initDir, markerFile), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write reserved file: %v", err)
	}

	ws := Workspace{
		Name: "feature-x",
		Path: filepath.Join(t.TempDir(), "feature-x"),
	}
	if err := os.MkdirAll(ws.Path, 0o755); err != nil {
		t.Fatalf("create workspace dir: %v", err)
	}

	manager := Manager{}
	err := manager.seedWorkspaceRoot(ws)
	if err == nil {
		t.Fatal("expected error for reserved marker filename")
	}
	if !strings.Contains(err.Error(), markerFile) {
		t.Fatalf("error %q does not mention %s", err, markerFile)
	}
}
