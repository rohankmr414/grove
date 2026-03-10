package cli

import (
	"context"
	"fmt"

	"github.com/rohankmr414/grove/internal/config"
	"github.com/rohankmr414/grove/internal/repo"
	"github.com/rohankmr414/grove/internal/ui"
	"github.com/rohankmr414/grove/internal/workspace"
)

func runAdd(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	manager := workspace.NewManager(cfg)
	ws, err := manager.DetectCurrent()
	if err != nil {
		return err
	}

	candidates, err := repo.Discover(context.Background(), cfg)
	if err != nil {
		return err
	}

	selected, err := ui.PickRepositories(candidates)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return fmt.Errorf("no repositories selected")
	}

	return manager.AddRepositories(context.Background(), ws, selected)
}
