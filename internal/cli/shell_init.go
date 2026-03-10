package cli

import (
	"fmt"
	"strings"
)

func runShellInit(args []string) error {
	shell := "zsh"
	if len(args) == 1 {
		shell = args[0]
	}

	switch shell {
	case "zsh":
		fmt.Print(zshShellInitScript())
		return nil
	case "bash":
		fmt.Print(bashShellInitScript())
		return nil
	default:
		return fmt.Errorf("unsupported shell %q", shell)
	}
}

func zshShellInitScript() string {
	lines := []string{
		"grove() {",
		"  if [ \"$1\" = \"cd\" ]; then",
		"    shift",
		"    local target",
		"    target=\"$(command grove __path \"$@\")\" || return $?",
		"    builtin cd \"$target\"",
		"    return 0",
		"  fi",
		"  if [ \"$1\" = \"init\" ]; then",
		"    local workspace",
		"    workspace=\"$2\"",
		"    command grove \"$@\" || return $?",
		"    [ -n \"$workspace\" ] || return 0",
		"    local target",
		"    target=\"$(command grove __path \"$workspace\")\" || return $?",
		"    builtin cd \"$target\"",
		"    return 0",
		"  fi",
		"  command grove \"$@\"",
		"}",
		"source <(command grove completion zsh)",
	}
	return strings.Join(lines, "\n") + "\n"
}

func bashShellInitScript() string {
	lines := []string{
		"grove() {",
		"  if [ \"$1\" = \"cd\" ]; then",
		"    shift",
		"    local target",
		"    target=\"$(command grove __path \"$@\")\" || return $?",
		"    builtin cd \"$target\"",
		"    return 0",
		"  fi",
		"  if [ \"$1\" = \"init\" ]; then",
		"    local workspace",
		"    workspace=\"$2\"",
		"    command grove \"$@\" || return $?",
		"    [ -n \"$workspace\" ] || return 0",
		"    local target",
		"    target=\"$(command grove __path \"$workspace\")\" || return $?",
		"    builtin cd \"$target\"",
		"    return 0",
		"  fi",
		"  command grove \"$@\"",
		"}",
		"source <(command grove completion bash)",
	}
	return strings.Join(lines, "\n") + "\n"
}
