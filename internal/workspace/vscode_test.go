package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSyncEditorWorkspaceWritesVSCodeMetadata(t *testing.T) {
	root := t.TempDir()
	ws := Workspace{
		Name: "feature-x",
		Path: filepath.Join(root, "feature-x"),
	}

	if err := os.MkdirAll(filepath.Join(ws.Path, ".vscode"), 0o755); err != nil {
		t.Fatalf("create vscode dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ws.Path, ".vscode", "settings.json"), []byte("{\n  \"editor.tabSize\": 2\n}\n"), 0o644); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	for _, name := range []string{"api", "web"} {
		repoPath := filepath.Join(ws.Path, name)
		if err := os.MkdirAll(repoPath, 0o755); err != nil {
			t.Fatalf("create repo dir %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(repoPath, ".git"), []byte("gitdir: /tmp/example\n"), 0o644); err != nil {
			t.Fatalf("seed .git for %s: %v", name, err)
		}
	}

	manager := Manager{}
	if err := manager.syncEditorWorkspace(ws); err != nil {
		t.Fatalf("sync editor workspace: %v", err)
	}

	settings := readVSCodeSettings(t, ws.Path)
	if settings["editor.tabSize"] != float64(2) {
		t.Fatalf("expected existing setting to survive, got %#v", settings["editor.tabSize"])
	}
	if settings["git.autoRepositoryDetection"] != "subFolders" {
		t.Fatalf("expected git.autoRepositoryDetection=subFolders, got %#v", settings["git.autoRepositoryDetection"])
	}
	if settings["git.repositoryScanMaxDepth"] != float64(2) {
		t.Fatalf("expected git.repositoryScanMaxDepth=2, got %#v", settings["git.repositoryScanMaxDepth"])
	}

	assertScanRepositories(t, settings, []string{"api", "web"})
}

func TestSyncEditorWorkspaceRefreshesScanRepositoriesAfterRepoAdd(t *testing.T) {
	root := t.TempDir()
	ws := Workspace{
		Name: "feature-x",
		Path: filepath.Join(root, "feature-x"),
	}

	if err := os.MkdirAll(filepath.Join(ws.Path, "api"), 0o755); err != nil {
		t.Fatalf("create api repo dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ws.Path, "api", ".git"), []byte("gitdir: /tmp/api\n"), 0o644); err != nil {
		t.Fatalf("seed .git for api: %v", err)
	}

	manager := Manager{}
	if err := manager.syncEditorWorkspace(ws); err != nil {
		t.Fatalf("first sync editor workspace: %v", err)
	}
	assertScanRepositories(t, readVSCodeSettings(t, ws.Path), []string{"api"})

	if err := os.MkdirAll(filepath.Join(ws.Path, "web"), 0o755); err != nil {
		t.Fatalf("create web repo dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ws.Path, "web", ".git"), []byte("gitdir: /tmp/web\n"), 0o644); err != nil {
		t.Fatalf("seed .git for web: %v", err)
	}

	if err := manager.syncEditorWorkspace(ws); err != nil {
		t.Fatalf("second sync editor workspace: %v", err)
	}
	assertScanRepositories(t, readVSCodeSettings(t, ws.Path), []string{"api", "web"})
}

func readVSCodeSettings(t *testing.T, workspacePath string) map[string]any {
	t.Helper()

	settingsData, err := os.ReadFile(filepath.Join(workspacePath, ".vscode", "settings.json"))
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(settingsData, &settings); err != nil {
		t.Fatalf("unmarshal settings: %v", err)
	}
	return settings
}

func assertScanRepositories(t *testing.T, settings map[string]any, expected []string) {
	t.Helper()

	scanRepositories, ok := settings["git.scanRepositories"].([]any)
	if !ok {
		t.Fatalf("expected git.scanRepositories to be an array, got %#v", settings["git.scanRepositories"])
	}
	if len(scanRepositories) != len(expected) {
		t.Fatalf("expected git.scanRepositories length %d, got %d", len(expected), len(scanRepositories))
	}
	for i := range expected {
		if scanRepositories[i] != expected[i] {
			t.Fatalf("expected git.scanRepositories=%v, got %#v", expected, scanRepositories)
		}
	}
}
