package cli

import (
	"context"
	"fmt"

	"github.com/rohankmr414/grove/internal/config"
	"github.com/rohankmr414/grove/internal/workspace"
)

func runRemove(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: grove remove <workspace>")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	manager := workspace.NewManager(cfg)
	return manager.Remove(context.Background(), args[0])
}
