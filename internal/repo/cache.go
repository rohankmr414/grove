package repo

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rohankmr414/grove/internal/util"
)

func DiscoverCached(ctx context.Context, cacheRoot string) ([]RepoCandidate, error) {
	info, err := os.Stat(cacheRoot)
	if err != nil || !info.IsDir() {
		return nil, nil
	}

	var repos []RepoCandidate
	seen := map[string]struct{}{}

	err = filepath.WalkDir(cacheRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.Name() != ".git" {
			return nil
		}

		repoRoot := filepath.Dir(path)
		if _, ok := seen[repoRoot]; ok {
			return filepath.SkipDir
		}
		seen[repoRoot] = struct{}{}

		candidate, err := cachedCandidate(ctx, repoRoot, cacheRoot)
		if err == nil {
			repos = append(repos, candidate)
		}
		return filepath.SkipDir
	})
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func cachedCandidate(ctx context.Context, repoRoot, cacheRoot string) (RepoCandidate, error) {
	candidate := RepoCandidate{
		Name:      filepath.Base(repoRoot),
		LocalPath: repoRoot,
		Source:    "cache",
	}

	relative, err := filepath.Rel(cacheRoot, repoRoot)
	if err == nil {
		parts := splitPath(relative)
		if len(parts) >= 2 {
			candidate.FullName = parts[len(parts)-2] + "/" + parts[len(parts)-1]
			candidate.Name = parts[len(parts)-1]
		}
	}

	origin, err := util.Output(ctx, "git", "-C", repoRoot, "remote", "get-url", "origin")
	if err == nil {
		candidate.CloneURL = origin
		if candidate.FullName == "" {
			if ownerRepo := parseGitHubFullName(origin); ownerRepo != "" {
				candidate.FullName = ownerRepo
			}
		}
	} else {
		candidate.CloneURL = repoRoot
	}

	defaultBranch, err := defaultBranch(ctx, repoRoot)
	if err == nil {
		candidate.DefaultBranch = defaultBranch
	}

	return candidate, nil
}

func splitPath(path string) []string {
	clean := filepath.Clean(path)
	if clean == "." || clean == string(filepath.Separator) {
		return nil
	}
	return strings.Split(filepath.ToSlash(clean), "/")
}
