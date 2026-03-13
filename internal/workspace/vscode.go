package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/rohankmr414/grove/internal/util"
)

var vscodeGitSettings = map[string]any{
	"git.autoRepositoryDetection": "subFolders",
	"git.repositoryScanMaxDepth":  2,
}

func (m Manager) syncEditorWorkspace(ws Workspace) error {
	if err := m.writeVSCodeSettings(ws); err != nil {
		return err
	}
	return nil
}

func (m Manager) writeVSCodeSettings(ws Workspace) error {
	settingsPath := filepath.Join(ws.Path, ".vscode", "settings.json")
	settings := map[string]any{}

	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	// Opening the workspace root with `code .` should make nested worktrees show up as repositories.
	for key, value := range vscodeGitSettings {
		settings[key] = value
	}

	repositories, err := workspaceRepositories(ws.Path)
	if err != nil {
		return err
	}
	settings["git.scanRepositories"] = repositories

	if err := util.EnsureDir(filepath.Dir(settingsPath)); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(settingsPath, data, 0o644)
}

func workspaceRepositories(workspacePath string) ([]string, error) {
	entries, err := os.ReadDir(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("read workspace directory: %w", err)
	}

	repositories := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".vscode" {
			continue
		}

		gitPath := filepath.Join(workspacePath, entry.Name(), ".git")
		if _, err := os.Stat(gitPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("check git metadata for %q: %w", entry.Name(), err)
		}

		repositories = append(repositories, entry.Name())
	}

	sort.Strings(repositories)
	return repositories, nil
}
