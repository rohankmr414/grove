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

	settingsData, err := os.ReadFile(filepath.Join(ws.Path, ".vscode", "settings.json"))
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(settingsData, &settings); err != nil {
		t.Fatalf("unmarshal settings: %v", err)
	}
	if settings["editor.tabSize"] != float64(2) {
		t.Fatalf("expected existing setting to survive, got %#v", settings["editor.tabSize"])
	}
	if settings["git.autoRepositoryDetection"] != "subFolders" {
		t.Fatalf("expected git.autoRepositoryDetection=subFolders, got %#v", settings["git.autoRepositoryDetection"])
	}
	if settings["git.repositoryScanMaxDepth"] != float64(2) {
		t.Fatalf("expected git.repositoryScanMaxDepth=2, got %#v", settings["git.repositoryScanMaxDepth"])
	}
}
