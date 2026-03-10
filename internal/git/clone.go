package git

import (
	"context"
	"os"
	"path/filepath"

	"github.com/rohankmr414/grove/internal/repo"
	"github.com/rohankmr414/grove/internal/util"
)

func EnsureCanonicalClone(ctx context.Context, candidate repo.RepoCandidate, cacheRoot string) (string, error) {
	canonicalPath := candidate.CachePath(cacheRoot)
	if _, err := os.Stat(canonicalPath); err == nil {
		return canonicalPath, nil
	}

	if err := util.EnsureDir(filepath.Dir(canonicalPath)); err != nil {
		return "", err
	}
	if err := util.Run(ctx, "git", "clone", candidate.CloneURL, canonicalPath); err != nil {
		return "", err
	}
	return canonicalPath, nil
}
