package cli

import (
	"context"
	"fmt"

	"github.com/rohankmr414/grove/internal/config"
	"github.com/rohankmr414/grove/internal/repo"
	"github.com/rohankmr414/grove/internal/ui"
	"github.com/rohankmr414/grove/internal/workspace"
)

func runInit(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	manager := workspace.NewManager(cfg)
	if err := manager.AssertDoesNotExist(args[0]); err != nil {
		return err
	}

	candidates, err := repo.Discover(context.Background(), cfg)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no repositories discovered; ensure ~/.grove/repos has cached clones or authenticate GitHub via gh/GITHUB_TOKEN")
	}

	selected, err := ui.PickRepositories(candidates)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return &usageError{text: "no repositories selected"}
	}

	return manager.Init(context.Background(), args[0], selected)
}
