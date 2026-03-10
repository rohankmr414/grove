package repo

import (
	"context"

	"github.com/rohankmr414/grove/internal/config"
)

type SearchFunc func(context.Context, string) ([]RepoCandidate, error)

type Source struct {
	Initial []RepoCandidate
	Search  SearchFunc
}

func NewSource(ctx context.Context, cfg config.Config) (Source, error) {
	cachedRepos, err := DiscoverCached(ctx, cfg.RepoCacheRoot)
	if err != nil {
		return Source{}, err
	}

	cachedGitHubRepos, err := loadCachedGitHubCandidates()
	if err != nil {
		cachedGitHubRepos = nil
	}

	all := make([]RepoCandidate, 0, len(cachedRepos)+len(cachedGitHubRepos))
	all = append(all, cachedRepos...)
	all = append(all, cachedGitHubRepos...)

	return Source{
		Initial: dedupe(all),
		Search: func(ctx context.Context, query string) ([]RepoCandidate, error) {
			return SearchGitHub(ctx, cfg, query)
		},
	}, nil
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
