package cli

import (
	"context"
	"fmt"
	"unicode/utf8"

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
	nameWidth := utf8.RuneCountInString("repository")
	branchWidth := utf8.RuneCountInString("branch")
	for _, entry := range report.Repositories {
		if width := utf8.RuneCountInString(entry.Name); width > nameWidth {
			nameWidth = width
		}
		if width := utf8.RuneCountInString(entry.Branch); width > branchWidth {
			branchWidth = width
		}
	}

	for _, entry := range report.Repositories {
		fmt.Printf("%-*s  %-*s  %s\n", nameWidth, entry.Name, branchWidth, entry.Branch, entry.State)
	}

	return nil
}
