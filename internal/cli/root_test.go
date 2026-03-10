package cli

import (
	"strings"
	"testing"
)

func TestExecuteRootHelpFlag(t *testing.T) {
	err := Execute([]string{"--help"})
	if err == nil {
		t.Fatal("expected help output")
	}
	if code := ExitCode(err); code != 0 {
		t.Fatalf("expected help exit code 0, got %d", code)
	}
	if !strings.Contains(err.Error(), "grove <command> [flags]") {
		t.Fatalf("expected root usage, got %q", err.Error())
	}
}

func TestExecuteUnknownCommand(t *testing.T) {
	err := Execute([]string{"wat"})
	if err == nil {
		t.Fatal("expected usage error")
	}
	if code := ExitCode(err); code != 2 {
		t.Fatalf("expected usage exit code 2, got %d", code)
	}
	if !strings.Contains(err.Error(), `unknown command "wat"`) {
		t.Fatalf("expected unknown command message, got %q", err.Error())
	}
}

func TestExecuteSubcommandUnexpectedArgs(t *testing.T) {
	err := Execute([]string{"version", "extra"})
	if err == nil {
		t.Fatal("expected usage error")
	}
	if code := ExitCode(err); code != 2 {
		t.Fatalf("expected usage exit code 2, got %d", code)
	}
	if !strings.Contains(err.Error(), `unexpected arguments: "extra"`) {
		t.Fatalf("expected unexpected arguments message, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "Usage:\n  grove version") {
		t.Fatalf("expected command usage, got %q", err.Error())
	}
}

func TestExecuteCommandHelp(t *testing.T) {
	err := Execute([]string{"help", "init"})
	if err == nil {
		t.Fatal("expected help output")
	}
	if code := ExitCode(err); code != 0 {
		t.Fatalf("expected help exit code 0, got %d", code)
	}
	if !strings.Contains(err.Error(), "Usage:\n  grove init <workspace>") {
		t.Fatalf("expected init help, got %q", err.Error())
	}
}
