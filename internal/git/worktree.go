package git

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/rohankmr414/grove/internal/util"
)

func AddWorktree(ctx context.Context, canonicalPath, worktreePath, branch, defaultBranch string) error {
	if err := util.EnsureDir(filepath.Dir(worktreePath)); err != nil {
		return err
	}

	if err := DetachHead(ctx, canonicalPath); err != nil {
		return err
	}

	if BranchExists(ctx, canonicalPath, branch) {
		return util.Run(ctx, "git", "-C", canonicalPath, "worktree", "add", worktreePath, branch)
	}

	// Fetch updates the remote-tracking branch, while the cache's local default
	// branch can remain at the revision that was current when it was cloned.
	// Start new workspace branches from the fetched ref so new worktrees contain
	// the latest remote revision.
	startPoint := defaultBranch
	remoteDefaultBranch := "refs/remotes/origin/" + defaultBranch
	if refExists(ctx, canonicalPath, remoteDefaultBranch) {
		startPoint = remoteDefaultBranch
	}
	return util.Run(ctx, "git", "-C", canonicalPath, "worktree", "add", "-b", branch, worktreePath, startPoint)
}

func RemoveWorktree(ctx context.Context, canonicalPath, worktreePath string) error {
	return util.Run(ctx, "git", "-C", canonicalPath, "worktree", "remove", worktreePath, "--force")
}

func CanonicalRoot(ctx context.Context, repoPath string) (string, error) {
	commonDir, err := util.Output(ctx, "git", "-C", repoPath, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return filepath.Dir(commonDir), nil
}

func DetachHead(ctx context.Context, canonicalPath string) error {
	head, err := util.Output(ctx, "git", "-C", canonicalPath, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return nil
	}

	if head == "" {
		return nil
	}

	if err := util.Run(ctx, "git", "-C", canonicalPath, "switch", "--detach", "--quiet"); err != nil {
		return fmt.Errorf("detach cache HEAD at %s: %w", canonicalPath, err)
	}

	return nil
}
