package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rohankmr414/grove/internal/util"
)

func TestDetachHead_DetachesWhenOnBranch(t *testing.T) {
	ctx := context.Background()
	repoPath := t.TempDir()

	mustRun(t, ctx, "git", "init", repoPath)
	mustRun(t, ctx, "git", "-C", repoPath, "config", "user.name", "Test User")
	mustRun(t, ctx, "git", "-C", repoPath, "config", "user.email", "test@example.com")

	filePath := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(filePath, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	mustRun(t, ctx, "git", "-C", repoPath, "add", "README.md")
	mustRun(t, ctx, "git", "-C", repoPath, "commit", "-m", "initial")

	if err := DetachHead(ctx, repoPath); err != nil {
		t.Fatalf("DetachHead returned error: %v", err)
	}

	headRef, err := util.Output(ctx, "git", "-C", repoPath, "symbolic-ref", "--quiet", "HEAD")
	if err == nil {
		t.Fatalf("expected detached HEAD, still attached to %q", headRef)
	}
}

func mustRun(t *testing.T, ctx context.Context, name string, args ...string) {
	t.Helper()
	if err := util.Run(ctx, name, args...); err != nil {
		t.Fatalf("%s %v: %v", name, args, err)
	}
}
