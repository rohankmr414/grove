package cli

import "fmt"

func allCommands() []command {
	return []command{
		{
			Name:        "init",
			Usage:       "grove init <workspace>",
			Short:       "Initialize a workspace from discovered repositories",
			MinArgs:     1,
			MaxArgs:     1,
			ArgsName:    "workspace",
			Description: "Creates a workspace, lets you pick repositories, and provisions git worktrees for each selection.",
			Run:         runInit,
		},
		{
			Name:        "add",
			Usage:       "grove add",
			Short:       "Add repositories to the current workspace",
			MinArgs:     0,
			MaxArgs:     0,
			Description: "Discovers repositories and adds the selected ones to the workspace for the current directory.",
			Run:         runAdd,
		},
		{
			Name:        "cd",
			Usage:       "grove cd <workspace>",
			Short:       "Change into a workspace via shell integration",
			MinArgs:     1,
			MaxArgs:     1,
			ArgsName:    "workspace",
			Description: "This command is implemented by the generated shell function. Running the binary directly only explains how to enable shell integration.",
			Run:         runCD,
		},
		{
			Name:        "list",
			Usage:       "grove list",
			Short:       "List grove workspaces",
			MinArgs:     0,
			MaxArgs:     0,
			Description: "Prints the names of all known grove workspaces.",
			Run:         runList,
		},
		{
			Name:        "path",
			Usage:       "grove path [workspace]",
			Short:       "Print the filesystem path for a workspace",
			MinArgs:     0,
			MaxArgs:     1,
			ArgsName:    "workspace",
			Description: "Prints the path for the named workspace, or the current workspace when run from inside one.",
			Run:         runPath,
		},
		{
			Name:        "status",
			Usage:       "grove status [workspace]",
			Short:       "Show branch and dirty state across repositories",
			MinArgs:     0,
			MaxArgs:     1,
			ArgsName:    "workspace",
			Description: "Shows repository status for the named workspace, or the current workspace when run from inside one.",
			Run:         runStatus,
		},
		{
			Name:        "remove",
			Usage:       "grove remove <workspace>",
			Short:       "Remove a workspace and its worktrees",
			MinArgs:     1,
			MaxArgs:     1,
			ArgsName:    "workspace",
			Description: "Deletes the workspace worktrees without removing the canonical cached clones.",
			Run:         runRemove,
		},
		{
			Name:        "version",
			Usage:       "grove version",
			Short:       "Print build version information",
			MinArgs:     0,
			MaxArgs:     0,
			Description: "Shows the application version, commit, and build timestamp.",
			Run:         runVersion,
		},
		{
			Name:        "shell-init",
			Usage:       "grove shell-init [zsh|bash]",
			Short:       "Print shell integration for grove cd and completion",
			MinArgs:     0,
			MaxArgs:     1,
			ArgsName:    "shell",
			Description: "Generates shell functions that enable `grove cd`, post-init auto-jump, and basic command completion.",
			Run:         runShellInit,
		},
		{
			Name:        "help",
			Usage:       "grove help [command]",
			Short:       "Show help for grove or a specific command",
			MinArgs:     0,
			MaxArgs:     1,
			ArgsName:    "command",
			Description: "Prints top-level help or command-specific usage text.",
			Run:         runHelp,
		},
		{
			Name:     "__path",
			Usage:    "grove __path [workspace]",
			MinArgs:  0,
			MaxArgs:  1,
			ArgsName: "workspace",
			Hidden:   true,
			Run:      runHiddenPath,
		},
		{
			Name:    "__workspaces",
			Usage:   "grove __workspaces",
			MinArgs: 0,
			MaxArgs: 0,
			Hidden:  true,
			Run:     runListWorkspaces,
		},
	}
}

func Execute(args []string) error {
	if len(args) == 0 {
		return &usageError{text: renderRootHelp()}
	}

	switch args[0] {
	case "help":
		return runHelp(args[1:])
	case "-h", "--help":
		return &helpError{text: renderRootHelp()}
	}

	cmd, ok := findCommand(args[0])
	if !ok {
		return &usageError{text: fmt.Sprintf("unknown command %q\n\n%s", args[0], renderRootHelp())}
	}

	return cmd.execute(args[1:])
}
