package repo

import (
	"context"

	"github.com/rohankmr414/grove/internal/config"
)

func Discover(ctx context.Context, cfg config.Config) ([]RepoCandidate, error) {
	cachedRepos, err := DiscoverCached(ctx, cfg.RepoCacheRoot)
	if err != nil {
		return nil, err
	}

	all := make([]RepoCandidate, 0, len(cachedRepos))
	all = append(all, cachedRepos...)

	ghRepos, err := DiscoverGitHub(ctx, cfg)
	if err != nil {
		return nil, err
	}
	all = append(all, ghRepos...)

	return dedupe(all), nil
}

func dedupe(input []RepoCandidate) []RepoCandidate {
	seen := make(map[string]RepoCandidate, len(input))
	order := make([]string, 0, len(input))

	for _, candidate := range input {
		key := candidate.FullName
		if key == "" {
			key = candidate.CloneURL
		}
		if key == "" {
			key = candidate.LocalPath
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = candidate
		order = append(order, key)
	}

	result := make([]RepoCandidate, 0, len(order))
	for _, key := range order {
		result = append(result, seen[key])
	}
	return result
}
