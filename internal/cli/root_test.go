package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecuteRootHelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := ExecuteForTest([]string{"--help"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "grove [command]") {
		t.Fatalf("expected root usage, got %q", stdout.String())
	}
}

func TestExecuteUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := ExecuteForTest([]string{"wat"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `unknown command "wat"`) {
		t.Fatalf("expected unknown command message, got %q", err.Error())
	}
}

func TestExecuteSubcommandUnexpectedArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := ExecuteForTest([]string{"version", "extra"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `unknown command "extra" for "grove version"`) {
		t.Fatalf("expected cobra unknown-command error, got %q", err.Error())
	}
}

func TestExecuteCommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := ExecuteForTest([]string{"help", "init"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(stdout.String(), "grove init <workspace>") {
		t.Fatalf("expected init help, got %q", stdout.String())
	}
}

func TestExecuteRepoSubcommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := ExecuteForTest([]string{"help", "repo"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(stdout.String(), "grove repo") {
		t.Fatalf("expected repo help, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "add") || !strings.Contains(stdout.String(), "remove") {
		t.Fatalf("expected repo subcommands in help, got %q", stdout.String())
	}
}
