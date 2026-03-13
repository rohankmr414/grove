package workspace

import "testing"

func TestWorkspaceBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{name: "auth-feature", want: "workspace/auth-feature"},
		{name: "bugfix/login-redirect", want: "bugfix/login-redirect"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := workspaceBranch(tt.name); got != tt.want {
				t.Fatalf("workspaceBranch(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
