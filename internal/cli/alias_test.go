package cli

import "testing"

func TestFindCommandAlias(t *testing.T) {
	tests := map[string]string{
		"ls": "list",
		"rm": "remove",
	}

	for alias, expected := range tests {
		cmd, ok := findCommand(alias)
		if !ok {
			t.Fatalf("expected alias %q to resolve", alias)
		}
		if cmd.Name != expected {
			t.Fatalf("alias %q resolved to %q, want %q", alias, cmd.Name, expected)
		}
	}
}
