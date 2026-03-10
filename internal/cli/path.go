package cli

import (
	"fmt"

	"github.com/rohankmr414/grove/internal/config"
	"github.com/rohankmr414/grove/internal/workspace"
)

func runPath(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: grove path [workspace]")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	manager := workspace.NewManager(cfg)
	ws, err := manager.ResolveWorkspace(args)
	if err != nil {
		return err
	}

	fmt.Println(ws.Path)
	return nil
}

func runHiddenPath(args []string) error {
	return runPath(args)
}

func runListWorkspaces(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: grove __workspaces")
	}

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

func runCD(args []string) error {
	return fmt.Errorf("`grove cd` requires shell integration; add `eval \"$(grove shell-init %s)\"` to your shell config", detectShell())
}

func detectShell() string {
	return "zsh"
}
