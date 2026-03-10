package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rohankmr414/grove/internal/config"
	"github.com/rohankmr414/grove/internal/util"
)

type githubAPIRepo struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	SSHURL        string `json:"ssh_url"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	UpdatedAt     string `json:"updated_at"`
}

func DiscoverGitHub(ctx context.Context, cfg config.Config) ([]RepoCandidate, error) {
	if !cfg.GitHub.Enabled {
		return loadCachedGitHubCandidates()
	}

	token := githubToken(ctx)
	if token == "" {
		return loadCachedGitHubCandidates()
	}

	var apiRepos []githubAPIRepo
	if len(cfg.GitHub.Orgs) == 0 {
		repos, err := fetchUserRepos(ctx, token)
		if err != nil {
			cached, cacheErr := loadGitHubCache()
			if cacheErr == nil {
				return cached, nil
			}
			return nil, err
		}
		apiRepos = append(apiRepos, repos...)
	} else {
		for _, org := range cfg.GitHub.Orgs {
			repos, err := fetchOrgRepos(ctx, token, org)
			if err != nil {
				cached, cacheErr := loadGitHubCache()
				if cacheErr == nil {
					return cached, nil
				}
				return nil, err
			}
			apiRepos = append(apiRepos, repos...)
		}
	}

	candidates := make([]RepoCandidate, 0, len(apiRepos))
	for _, entry := range apiRepos {
		candidates = append(candidates, RepoCandidate{
			Name:          entry.Name,
			FullName:      entry.FullName,
			CloneURL:      entry.SSHURL,
			DefaultBranch: entry.DefaultBranch,
			Source:        "github",
		})
	}

	if err := saveGitHubCache(candidates); err != nil {
		return nil, err
	}

	return candidates, nil
}

func githubToken(ctx context.Context) string {
	if util.LookPath("gh") {
		token, err := util.Output(ctx, "gh", "auth", "token")
		if err == nil && token != "" {
			return token
		}
	}
	return strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
}

func fetchOrgRepos(ctx context.Context, token, org string) ([]githubAPIRepo, error) {
	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos?per_page=100&page=%%d", org)
	return fetchGitHubRepoPages(ctx, token, url)
}

func fetchUserRepos(ctx context.Context, token string) ([]githubAPIRepo, error) {
	return fetchGitHubRepoPages(ctx, token, "https://api.github.com/user/repos?per_page=100&affiliation=owner,collaborator,organization_member&page=%d")
}

func fetchGitHubRepoPages(ctx context.Context, token, urlFormat string) ([]githubAPIRepo, error) {
	client := &http.Client{}
	var result []githubAPIRepo

	for page := 1; ; page++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(urlFormat, page), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("github api returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
		}

		var pageRepos []githubAPIRepo
		if err := json.Unmarshal(body, &pageRepos); err != nil {
			return nil, fmt.Errorf("decode github response: %w", err)
		}
		result = append(result, pageRepos...)
		if len(pageRepos) < 100 {
			break
		}
	}

	return result, nil
}

func cachePath() (string, error) {
	return util.ExpandPath("~/.grove/cache/repos.json")
}

func loadGitHubCache() ([]RepoCandidate, error) {
	path, err := cachePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var repos []RepoCandidate
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func loadCachedGitHubCandidates() ([]RepoCandidate, error) {
	repos, err := loadGitHubCache()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return repos, nil
}

func saveGitHubCache(repos []RepoCandidate) error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	if err := util.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
