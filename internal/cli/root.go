package cli

import (
	"io"
	"os"

	"github.com/rohankmr414/grove/internal/config"
	"github.com/rohankmr414/grove/internal/workspace"
	"github.com/spf13/cobra"
)

func Execute(args []string) error {
	return execute(args, os.Stdout, os.Stderr)
}

func execute(args []string, stdout, stderr io.Writer) error {
	cmd := newRootCommand(stdout, stderr)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func newRootCommand(stdout, stderr io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "grove",
		Short:         "Manage multi-repository workspaces using git worktrees",
		Long:          "grove manages multi-repository workspaces using git worktrees.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	root.SetOut(stdout)
	root.SetErr(stderr)

	repoCommand := &cobra.Command{
		Use:   "repo",
		Short: "Manage repositories in the current workspace",
		Long:  "Add or remove repositories in the current workspace.",
	}
	repoCommand.AddCommand(
		newCommand(
			"add",
			"Add repositories to the current workspace",
			"Discovers repositories and adds the selected ones to the workspace for the current directory.",
			nil,
			cobra.NoArgs,
			runAdd,
		),
		newCommand(
			"remove",
			"Remove repositories from the current workspace",
			"Shows repositories already present in the current workspace and removes the selected worktrees.",
			[]string{"rm"},
			cobra.NoArgs,
			runRemoveRepo,
		),
	)

	root.AddCommand(
		newCommand(
			"init <workspace>",
			"Initialize a workspace from discovered repositories",
			"Creates a workspace, lets you pick repositories, and provisions git worktrees for each selection.",
			nil,
			cobra.ExactArgs(1),
			runInit,
		),
		repoCommand,
		newCommand(
			"cd <workspace>",
			"Change into a workspace via shell integration",
			"This command is implemented by the generated shell function. Running the binary directly only explains how to enable shell integration.",
			nil,
			cobra.ExactArgs(1),
			runCD,
		),
		newCommand(
			"list",
			"List grove workspaces",
			"Prints the names of all known grove workspaces.",
			[]string{"ls"},
			cobra.NoArgs,
			runList,
		),
		newCommand(
			"path [workspace]",
			"Print the filesystem path for a workspace",
			"Prints the path for the named workspace, or the current workspace when run from inside one.",
			nil,
			cobra.MaximumNArgs(1),
			runPath,
		),
		newCommand(
			"status [workspace]",
			"Show branch and dirty state across repositories",
			"Shows repository status for the named workspace, or the current workspace when run from inside one.",
			nil,
			cobra.MaximumNArgs(1),
			runStatus,
		),
		newCommand(
			"remove <workspace>",
			"Remove a workspace and its worktrees",
			"Deletes the workspace worktrees without removing the canonical cached clones.",
			[]string{"rm"},
			cobra.ExactArgs(1),
			runRemove,
		),
		newCommand(
			"version",
			"Print build version information",
			"Shows the application version, commit, and build timestamp.",
			nil,
			cobra.NoArgs,
			runVersion,
		),
		newCommand(
			"shell-init [zsh|bash]",
			"Print shell integration for grove cd",
			"Generates shell functions that enable `grove cd` and post-init auto-jump.",
			nil,
			cobra.MaximumNArgs(1),
			runShellInit,
		),
		newCompletionCommand(),
		newHiddenCommand("__path [workspace]", cobra.MaximumNArgs(1), runHiddenPath),
	)

	attachWorkspaceCompletion(root)

	return root
}

func newCommand(use, short, long string, aliases []string, args cobra.PositionalArgs, run func([]string) error) *cobra.Command {
	return &cobra.Command{
		Use:     use,
		Short:   short,
		Long:    long,
		Aliases: aliases,
		Args:    args,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(args)
		},
	}
}

func newHiddenCommand(use string, args cobra.PositionalArgs, run func([]string) error) *cobra.Command {
	return &cobra.Command{
		Use:    use,
		Hidden: true,
		Args:   args,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(args)
		},
	}
}

func ExecuteForTest(args []string, stdout, stderr io.Writer) error {
	return execute(args, stdout, stderr)
}

func attachWorkspaceCompletion(root *cobra.Command) {
	workspaceCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cfg, err := config.Load()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		manager := workspace.NewManager(cfg)
		workspaces, err := manager.List()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		names := make([]string, 0, len(workspaces))
		for _, ws := range workspaces {
			if toComplete == "" || len(ws.Name) >= len(toComplete) && ws.Name[:len(toComplete)] == toComplete {
				names = append(names, ws.Name)
			}
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	}

	for _, name := range []string{"cd", "path", "status", "remove", "rm"} {
		cmd, _, err := root.Find([]string{name})
		if err == nil && cmd != nil {
			cmd.ValidArgsFunction = workspaceCompletion
		}
	}
}
