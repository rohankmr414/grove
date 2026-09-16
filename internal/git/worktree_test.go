package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rohankmr414/grove/internal/util"
)

func removalFixture(t *testing.T) (context.Context, string, string) {
	t.Helper()
	ctx := context.Background()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	canonical := filepath.Join(root, "canonical")
	worktree := filepath.Join(root, "feature")
	mustRun(t, ctx, "git", "init", canonical)
	mustRun(t, ctx, "git", "-C", canonical, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", "initial")
	mustRun(t, ctx, "git", "-C", canonical, "worktree", "add", "-b", "feature", worktree)
	mustWriteFile(t, filepath.Join(worktree, "untracked.txt"), "leftover\n")
	return ctx, canonical, worktree
}

func TestRemoveWorktree(t *testing.T) {
	ctx, canonical, worktree := removalFixture(t)
	if err := RemoveWorktree(ctx, canonical, worktree); err != nil {
		t.Fatal(err)
	}
	assertWorktreeRemoved(t, ctx, canonical, worktree)
}

func TestRemoveWorktreeRecoversAfterGitDeletionFailure(t *testing.T) {
	ctx, canonical, worktree := removalFixture(t)
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	// Keep the files while letting real Git remove the registration, exactly
	// the state left behind when Git's recursive deletion fails.
	script := `#!/bin/sh
if [ "$3" = worktree ] && [ "$4" = remove ]; then
  mv "$5" "$5.leftover" || exit 1
  "$REAL_GIT" "$@"
  result=$?
  mv "$5.leftover" "$5" || exit 1
  [ "$result" -eq 0 ] || exit "$result"
  echo 'error: failed to delete worktree: Directory not empty' >&2
  exit 255
fi
exec "$REAL_GIT" "$@"
`
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REAL_GIT", realGit)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if err := RemoveWorktree(ctx, canonical, worktree); err != nil {
		t.Fatal(err)
	}
	assertWorktreeRemoved(t, ctx, canonical, worktree)
}

func TestRemoveWorktreeRetriesStaleWorktree(t *testing.T) {
	ctx, canonical, worktree := removalFixture(t)
	gitFile, err := os.ReadFile(filepath.Join(worktree, ".git"))
	if err != nil {
		t.Fatal(err)
	}
	mustRun(t, ctx, "git", "-C", canonical, "worktree", "remove", "--force", worktree)
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(worktree, ".git"), string(gitFile))
	mustWriteFile(t, filepath.Join(worktree, "leftover.txt"), "leftover\n")
	root, err := CanonicalRoot(ctx, worktree)
	if err != nil || root != canonical {
		t.Fatalf("CanonicalRoot = %q, %v; want %q", root, err, canonical)
	}
	if err := RemoveWorktree(ctx, root, worktree); err != nil {
		t.Fatal(err)
	}
	assertWorktreeRemoved(t, ctx, canonical, worktree)
}

func TestRemoveWorktreePreservesLockedWorktree(t *testing.T) {
	ctx, canonical, worktree := removalFixture(t)
	mustRun(t, ctx, "git", "-C", canonical, "worktree", "lock", worktree)
	if err := RemoveWorktree(ctx, canonical, worktree); err == nil || !strings.Contains(err.Error(), "locked") {
		t.Fatalf("expected locked worktree error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(worktree, "untracked.txt")); err != nil {
		t.Fatalf("locked worktree files lost: %v", err)
	}
}

func TestRemoveWorktreeRejectsUnrelatedDirectory(t *testing.T) {
	ctx, canonical, _ := removalFixture(t)
	other := t.TempDir()
	mustWriteFile(t, filepath.Join(other, "keep.txt"), "keep\n")
	if err := RemoveWorktree(ctx, canonical, other); err == nil {
		t.Fatal("expected error for unrelated directory")
	}
	if _, err := os.Stat(filepath.Join(other, "keep.txt")); err != nil {
		t.Fatal(err)
	}
}

func assertWorktreeRemoved(t *testing.T, ctx context.Context, canonical, worktree string) {
	t.Helper()
	if _, err := os.Stat(worktree); !os.IsNotExist(err) {
		t.Fatalf("worktree still exists: %v", err)
	}
	registered, err := worktreeRegistered(ctx, canonical, worktree)
	if err != nil || registered {
		t.Fatalf("worktree registered = %v, error = %v", registered, err)
	}
	if _, err := os.Stat(filepath.Join(canonical, ".git")); err != nil {
		t.Fatalf("canonical repository lost: %v", err)
	}
}

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
