package cli

import "fmt"

func Execute(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "add":
		return runAdd(args[1:])
	case "cd":
		return runCD(args[1:])
	case "path":
		return runPath(args[1:])
	case "status":
		return runStatus(args[1:])
	case "remove":
		return runRemove(args[1:])
	case "version":
		return runVersion(args[1:])
	case "shell-init":
		return runShellInit(args[1:])
	case "__path":
		return runHiddenPath(args[1:])
	case "__workspaces":
		return runListWorkspaces(args[1:])
	case "help", "-h", "--help":
		return usageError()
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage())
	}
}

func usageError() error {
	return fmt.Errorf("%s", usage())
}

func usage() string {
	return `grove manages multi-repository workspaces using git worktrees.

Usage:
  grove init <workspace>
  grove add
  grove cd <workspace>
  grove path [workspace]
  grove status [workspace]
  grove remove <workspace>
  grove version
  grove shell-init [zsh|bash]`
}
