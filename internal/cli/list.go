package cli

import (
	"fmt"

	"github.com/rohankmr414/grove/internal/config"
	"github.com/rohankmr414/grove/internal/workspace"
)

func runList(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	manager := workspace.NewManager(cfg)
	workspaces, err := manager.List()
	if err != nil {
		return err
	}

	for _, ws := range workspaces {
		fmt.Println(ws.Name)
	}
	return nil
}
