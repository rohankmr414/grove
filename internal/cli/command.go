package cli

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/spf13/pflag"
)

type command struct {
	Name        string
	Aliases     []string
	Usage       string
	Short       string
	Hidden      bool
	MinArgs     int
	MaxArgs     int
	ArgsName    string
	Description string
	Run         func([]string) error
}

type helpError struct {
	text string
}

func (e *helpError) Error() string {
	return e.text
}

type usageError struct {
	text string
}

func (e *usageError) Error() string {
	return e.text
}

func ExitCode(err error) int {
	switch err.(type) {
	case nil:
		return 0
	case *helpError:
		return 0
	case *usageError:
		return 2
	default:
		return 1
	}
}

func (c command) execute(args []string) error {
	flagSet := pflag.NewFlagSet(c.Name, pflag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	var showHelp bool
	flagSet.BoolVarP(&showHelp, "help", "h", false, "Show help")

	if err := flagSet.Parse(args); err != nil {
		return &usageError{text: fmt.Sprintf("%s\n\n%s", err.Error(), c.helpText())}
	}
	if showHelp {
		return &helpError{text: c.helpText()}
	}

	rest := flagSet.Args()
	if err := c.validateArgs(rest); err != nil {
		return &usageError{text: fmt.Sprintf("%s\n\n%s", err.Error(), c.helpText())}
	}

	return c.Run(rest)
}

func (c command) validateArgs(args []string) error {
	switch {
	case c.MinArgs == c.MaxArgs && len(args) != c.MinArgs:
		if len(args) < c.MinArgs {
			return fmt.Errorf("accepts %s", exactArgText(c.MinArgs, c.ArgsName))
		}
		return unexpectedArgsError(args[c.MaxArgs:])
	case len(args) < c.MinArgs:
		return fmt.Errorf("accepts %s", minArgText(c.MinArgs, c.ArgsName))
	case c.MaxArgs >= 0 && len(args) > c.MaxArgs:
		return unexpectedArgsError(args[c.MaxArgs:])
	default:
		return nil
	}
}

func (c command) helpText() string {
	var builder strings.Builder
	builder.WriteString("Usage:\n")
	builder.WriteString("  ")
	builder.WriteString(c.Usage)
	builder.WriteString("\n")

	if c.Description != "" {
		builder.WriteString("\n")
		builder.WriteString(c.Description)
		builder.WriteString("\n")
	}
	if len(c.Aliases) > 0 {
		builder.WriteString("\nAliases:\n")
		builder.WriteString("  ")
		builder.WriteString(strings.Join(c.Aliases, ", "))
		builder.WriteString("\n")
	}

	builder.WriteString("\nFlags:\n")
	builder.WriteString("  -h, --help   Show help\n")
	return builder.String()
}

func exactArgText(count int, argsName string) string {
	switch count {
	case 0:
		return "no arguments"
	case 1:
		if argsName != "" {
			return fmt.Sprintf("exactly 1 argument (%s)", argsName)
		}
		return "exactly 1 argument"
	default:
		if argsName != "" {
			return fmt.Sprintf("exactly %d arguments (%s)", count, argsName)
		}
		return fmt.Sprintf("exactly %d arguments", count)
	}
}

func minArgText(count int, argsName string) string {
	if count == 1 {
		if argsName != "" {
			return fmt.Sprintf("at least 1 argument (%s)", argsName)
		}
		return "at least 1 argument"
	}
	if argsName != "" {
		return fmt.Sprintf("at least %d arguments (%s)", count, argsName)
	}
	return fmt.Sprintf("at least %d arguments", count)
}

func unexpectedArgsError(args []string) error {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, fmt.Sprintf("%q", arg))
	}
	return fmt.Errorf("unexpected arguments: %s", strings.Join(quoted, ", "))
}

func renderRootHelp() string {
	var builder strings.Builder
	builder.WriteString("grove manages multi-repository workspaces using git worktrees.\n\n")
	builder.WriteString("Usage:\n")
	builder.WriteString("  grove <command> [flags]\n\n")
	builder.WriteString("Commands:\n")

	for _, cmd := range visibleCommands() {
		name := cmd.Name
		if len(cmd.Aliases) > 0 {
			name = fmt.Sprintf("%s (%s)", cmd.Name, strings.Join(cmd.Aliases, ", "))
		}
		builder.WriteString(fmt.Sprintf("  %-18s %s\n", name, cmd.Short))
	}

	builder.WriteString("\nUse \"grove help <command>\" for more information about a command.\n")
	return builder.String()
}

func visibleCommands() []command {
	commands := allCommands()
	visible := make([]command, 0, len(commands))
	for _, cmd := range commands {
		if !cmd.Hidden {
			visible = append(visible, cmd)
		}
	}
	return visible
}

func findCommand(name string) (command, bool) {
	commands := allCommands()
	index := slices.IndexFunc(commands, func(cmd command) bool {
		return cmd.Name == name || slices.Contains(cmd.Aliases, name)
	})
	if index == -1 {
		return command{}, false
	}
	return commands[index], true
}
