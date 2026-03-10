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
	if len(args) != 0 {
		return fmt.Errorf("usage: grove add")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	manager := workspace.NewManager(cfg)
	ws, err := manager.DetectCurrent()
	if err != nil {
		return err
	}

	source, err := repo.NewSource(context.Background(), cfg)
	if err != nil {
		return err
	}

	selected, err := ui.PickRepositories(source.Initial, source.Search)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return fmt.Errorf("no repositories selected")
	}

	return manager.AddRepositories(context.Background(), ws, selected)
}
