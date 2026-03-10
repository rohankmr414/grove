package cli

import (
	"strings"
	"testing"
)

func TestShellInitUsesCobraCompletion(t *testing.T) {
	zsh := zshShellInitScript()
	if want := "source <(command grove completion zsh)"; !strings.Contains(zsh, want) {
		t.Fatalf("zsh shell-init missing %q", want)
	}

	bash := bashShellInitScript()
	if want := "source <(command grove completion bash)"; !strings.Contains(bash, want) {
		t.Fatalf("bash shell-init missing %q", want)
	}
}
