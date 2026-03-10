package cli

import (
	"context"

	"github.com/rohankmr414/grove/internal/config"
	"github.com/rohankmr414/grove/internal/workspace"
)

func runRemove(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	manager := workspace.NewManager(cfg)
	return manager.Remove(context.Background(), args[0])
}
