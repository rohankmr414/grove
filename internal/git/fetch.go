package git

import (
	"context"

	"github.com/rohankmr414/grove/internal/util"
)

func Fetch(ctx context.Context, repoPath string) error {
	return util.Run(ctx, "git", "-C", repoPath, "fetch", "--all", "--prune")
}
