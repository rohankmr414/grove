package repo

import (
	"context"
	"strings"

	"github.com/rohankmr414/grove/internal/util"
)

func parseGitHubFullName(remote string) string {
	remote = strings.TrimSuffix(remote, ".git")
	remote = strings.TrimSpace(remote)
	switch {
	case strings.HasPrefix(remote, "git@github.com:"):
		return strings.TrimPrefix(remote, "git@github.com:")
	case strings.HasPrefix(remote, "https://github.com/"):
		return strings.TrimPrefix(remote, "https://github.com/")
	case strings.HasPrefix(remote, "ssh://git@github.com/"):
		return strings.TrimPrefix(remote, "ssh://git@github.com/")
	default:
		return ""
	}
}

func defaultBranch(ctx context.Context, repoRoot string) (string, error) {
	head, err := util.Output(ctx, "git", "-C", repoRoot, "symbolic-ref", "refs/remotes/origin/HEAD", "--short")
	if err == nil && strings.Contains(head, "/") {
		return head[strings.LastIndex(head, "/")+1:], nil
	}
	return util.Output(ctx, "git", "-C", repoRoot, "rev-parse", "--abbrev-ref", "HEAD")
}

func repoHasCommits(ctx context.Context, repoRoot string) (bool, error) {
	if err := util.Run(ctx, "git", "-C", repoRoot, "rev-parse", "--verify", "HEAD"); err != nil {
		return false, nil
	}
	return true, nil
}
