package git

import (
	"context"
	"path/filepath"

	"github.com/rohankmr414/grove/internal/util"
)

func AddWorktree(ctx context.Context, canonicalPath, worktreePath, branch, defaultBranch string) error {
	if err := util.EnsureDir(filepath.Dir(worktreePath)); err != nil {
		return err
	}

	if BranchExists(ctx, canonicalPath, branch) {
		return util.Run(ctx, "git", "-C", canonicalPath, "worktree", "add", worktreePath, branch)
	}

	return util.Run(ctx, "git", "-C", canonicalPath, "worktree", "add", "-b", branch, worktreePath, defaultBranch)
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
