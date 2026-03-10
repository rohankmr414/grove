package git

import (
	"context"

	"github.com/rohankmr414/grove/internal/util"
)

func BranchExists(ctx context.Context, repoPath, branch string) bool {
	if err := util.Run(ctx, "git", "-C", repoPath, "show-ref", "--verify", "--quiet", "refs/heads/"+branch); err == nil {
		return true
	}
	return util.Run(ctx, "git", "-C", repoPath, "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch) == nil
}

func CurrentBranch(ctx context.Context, repoPath string) (string, error) {
	return util.Output(ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
}
