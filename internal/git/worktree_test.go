package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rohankmr414/grove/internal/util"
)

func TestAddWorktree_StartsNewBranchAtFetchedDefaultBranch(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	remotePath := filepath.Join(root, "remote.git")
	seedPath := filepath.Join(root, "seed")
	canonicalPath := filepath.Join(root, "canonical")
	worktreePath := filepath.Join(root, "worktrees", "feature")

	mustRun(t, ctx, "git", "init", "--bare", remotePath)
	mustRun(t, ctx, "git", "clone", remotePath, seedPath)
	mustRun(t, ctx, "git", "-C", seedPath, "config", "user.name", "Test User")
	mustRun(t, ctx, "git", "-C", seedPath, "config", "user.email", "test@example.com")
	mustWriteFile(t, filepath.Join(seedPath, "README.md"), "initial\n")
	mustRun(t, ctx, "git", "-C", seedPath, "add", "README.md")
	mustRun(t, ctx, "git", "-C", seedPath, "commit", "-m", "initial")
	mustRun(t, ctx, "git", "-C", seedPath, "branch", "-M", "main")
	mustRun(t, ctx, "git", "-C", seedPath, "push", "-u", "origin", "main")
	mustRun(t, ctx, "git", "-C", remotePath, "symbolic-ref", "HEAD", "refs/heads/main")
	mustRun(t, ctx, "git", "clone", remotePath, canonicalPath)

	staleHead, err := util.Output(ctx, "git", "-C", canonicalPath, "rev-parse", "refs/heads/main")
	if err != nil {
		t.Fatalf("read cached main revision: %v", err)
	}

	mustWriteFile(t, filepath.Join(seedPath, "README.md"), "latest\n")
	mustRun(t, ctx, "git", "-C", seedPath, "add", "README.md")
	mustRun(t, ctx, "git", "-C", seedPath, "commit", "-m", "latest")
	mustRun(t, ctx, "git", "-C", seedPath, "push", "origin", "main")

	if err := Fetch(ctx, canonicalPath); err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}
	latestHead, err := util.Output(ctx, "git", "-C", canonicalPath, "rev-parse", "refs/remotes/origin/main")
	if err != nil {
		t.Fatalf("read fetched main revision: %v", err)
	}
	if staleHead == latestHead {
		t.Fatal("test setup did not leave the cached local branch stale")
	}

	if err := AddWorktree(ctx, canonicalPath, worktreePath, "workspace/feature", "main"); err != nil {
		t.Fatalf("AddWorktree returned error: %v", err)
	}

	worktreeHead, err := util.Output(ctx, "git", "-C", worktreePath, "rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("read worktree revision: %v", err)
	}
	if worktreeHead != latestHead {
		t.Fatalf("worktree revision = %s, want latest fetched revision %s", worktreeHead, latestHead)
	}
}

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

func mustWriteFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
