package git

import (
	"context"

	"github.com/rohankmr414/grove/internal/util"
)

func BranchExists(ctx context.Context, repoPath, branch string) bool {
	if refExists(ctx, repoPath, "refs/heads/"+branch) {
		return true
	}
	return refExists(ctx, repoPath, "refs/remotes/origin/"+branch)
}

func refExists(ctx context.Context, repoPath, ref string) bool {
	return util.Run(ctx, "git", "-C", repoPath, "show-ref", "--verify", "--quiet", ref) == nil
}

func CurrentBranch(ctx context.Context, repoPath string) (string, error) {
	return util.Output(ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
}
