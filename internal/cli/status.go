package cli

import (
	"context"
	"fmt"

	"github.com/rohankmr414/grove/internal/config"
	"github.com/rohankmr414/grove/internal/workspace"
)

func runStatus(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	manager := workspace.NewManager(cfg)
	report, err := manager.Status(context.Background(), args)
	if err != nil {
		return err
	}

	fmt.Printf("Workspace: %s\n\n", report.Name)
	for _, entry := range report.Repositories {
		fmt.Printf("%-12s %-24s %s\n", entry.Name, entry.Branch, entry.State)
	}

	return nil
}
