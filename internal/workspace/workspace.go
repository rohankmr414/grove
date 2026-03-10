package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const markerFile = ".grove-workspace.json"

type Workspace struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type StatusEntry struct {
	Name   string
	Branch string
	State  string
}

type StatusReport struct {
	Name         string
	Repositories []StatusEntry
}

func markerPath(workspacePath string) string {
	return filepath.Join(workspacePath, markerFile)
}

func writeMarker(ws Workspace) error {
	data, err := json.MarshalIndent(ws, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(markerPath(ws.Path), data, 0o644)
}

func readMarker(workspacePath string) (Workspace, error) {
	data, err := os.ReadFile(markerPath(workspacePath))
	if err != nil {
		return Workspace{}, err
	}
	var ws Workspace
	if err := json.Unmarshal(data, &ws); err != nil {
		return Workspace{}, err
	}
	if ws.Name == "" || ws.Path == "" {
		return Workspace{}, fmt.Errorf("invalid workspace metadata in %s", markerPath(workspacePath))
	}
	return ws, nil
}
