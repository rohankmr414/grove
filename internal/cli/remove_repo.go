package cli

import (
	"context"
	"fmt"

	"github.com/rohankmr414/grove/internal/config"
	"github.com/rohankmr414/grove/internal/ui"
	"github.com/rohankmr414/grove/internal/workspace"
)

func runRemoveRepo(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	manager := workspace.NewManager(cfg)
	ws, err := manager.DetectCurrent()
	if err != nil {
		return err
	}

	repositories, err := manager.WorkspaceRepositories(ws)
	if err != nil {
		return err
	}

	selected, err := ui.PickRepositories(repositories)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return fmt.Errorf("no repositories selected")
	}

	names := make([]string, 0, len(selected))
	for _, repository := range selected {
		names = append(names, repository.Name)
	}

	return manager.RemoveRepositories(context.Background(), ws, names...)
}
