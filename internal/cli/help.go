package cli

import "fmt"

func runHelp(args []string) error {
	if len(args) == 0 {
		return &helpError{text: renderRootHelp()}
	}

	cmd, ok := findCommand(args[0])
	if !ok || cmd.Hidden {
		return &usageError{text: fmt.Sprintf("unknown help topic %q\n\n%s", args[0], renderRootHelp())}
	}

	return &helpError{text: cmd.helpText()}
}
