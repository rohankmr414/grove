package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rohankmr414/grove/internal/config"
	"github.com/rohankmr414/grove/internal/util"
)

func TestRemoveWorkspaceWithPartiallyRemovedWorktree(t *testing.T) {
	ctx := context.Background()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	canonical := filepath.Join(root, "cache", "repo")
	workspaceRoot := filepath.Join(root, "workspaces")
	ws := Workspace{Name: "feature", Path: filepath.Join(workspaceRoot, "feature")}
	if err := os.MkdirAll(ws.Path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeMarker(ws); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		t.Helper()
		if err := util.Run(ctx, "git", args...); err != nil {
			t.Fatal(err)
		}
	}
	run("init", canonical)
	run("-C", canonical, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", "initial")
	worktree := filepath.Join(ws.Path, "repo")
	run("-C", canonical, "worktree", "add", "-b", "feature", worktree)
	gitFile, err := os.ReadFile(filepath.Join(worktree, ".git"))
	if err != nil {
		t.Fatal(err)
	}
	run("-C", canonical, "worktree", "remove", "--force", worktree)
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string][]byte{".git": gitFile, "leftover.txt": []byte("leftover")} {
		if err := os.WriteFile(filepath.Join(worktree, name), content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	manager := NewManager(config.Config{WorkspaceRoot: workspaceRoot})
	if err := manager.Remove(ctx, ws.Name); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ws.Path); !os.IsNotExist(err) {
		t.Fatalf("workspace still exists: %v", err)
	}
	run("-C", canonical, "rev-parse", "--verify", "refs/heads/feature")
}

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
