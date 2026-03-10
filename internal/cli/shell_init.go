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
		"_grove_cd_completions() {",
		"  local -a workspaces",
		"  workspaces=(${(f)$(command grove __workspaces 2>/dev/null)})",
		"  _describe 'workspace' workspaces",
		"}",
		"_grove_complete() {",
		"  if (( CURRENT == 2 )); then",
		"    _arguments '1:command:(init add cd list path status remove version shell-init help)'",
		"    return",
		"  fi",
		"  if [[ ${words[2]} == cd ]]; then",
		"    _grove_cd_completions",
		"  fi",
		"}",
		"compdef _grove_complete grove",
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
		"_grove_complete() {",
		"  local cur prev",
		"  COMPREPLY=()",
		"  cur=\"${COMP_WORDS[COMP_CWORD]}\"",
		"  prev=\"${COMP_WORDS[COMP_CWORD-1]}\"",
		"  if [ ${COMP_CWORD} -eq 1 ]; then",
		"    COMPREPLY=( $(compgen -W 'init add cd list path status remove version shell-init help' -- \"$cur\") )",
		"    return 0",
		"  fi",
		"  if [ \"$prev\" = \"cd\" ]; then",
		"    COMPREPLY=( $(compgen -W \"$(command grove __workspaces 2>/dev/null)\" -- \"$cur\") )",
		"    return 0",
		"  fi",
		"}",
		"complete -F _grove_complete grove",
	}
	return strings.Join(lines, "\n") + "\n"
}
