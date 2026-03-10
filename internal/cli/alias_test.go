package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecuteAliasHelp(t *testing.T) {
	tests := map[string]string{
		"ls": "grove list",
		"rm": "grove remove <workspace>",
	}

	for alias, expected := range tests {
		var stdout, stderr bytes.Buffer
		err := ExecuteForTest([]string{alias, "--help"}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("alias %q returned error: %v", alias, err)
		}
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("alias %q help missing %q in %q", alias, expected, stdout.String())
		}
	}
}
