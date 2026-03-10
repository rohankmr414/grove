package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	Size          int    `json:"size"`
	UpdatedAt     string `json:"updated_at"`
}

type githubSearchResponse struct {
	Items []githubAPIRepo `json:"items"`
}

func SearchGitHub(ctx context.Context, cfg config.Config, query string) ([]RepoCandidate, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}

	if !cfg.GitHub.Enabled {
		return searchCachedGitHubCandidates(query)
	}

	token := githubToken(ctx)
	if token == "" {
		return searchCachedGitHubCandidates(query)
	}

	var apiRepos []githubAPIRepo
	if len(cfg.GitHub.Orgs) == 0 {
		repos, err := searchUserRepos(ctx, token, query)
		if err != nil {
			return searchCachedGitHubCandidates(query)
		}
		apiRepos = append(apiRepos, repos...)
	} else {
		for _, org := range cfg.GitHub.Orgs {
			repos, err := searchOrgRepos(ctx, token, org, query)
			if err != nil {
				return searchCachedGitHubCandidates(query)
			}
			apiRepos = append(apiRepos, repos...)
		}
	}

	candidates := make([]RepoCandidate, 0, len(apiRepos))
	for _, entry := range apiRepos {
		if !githubRepoUsable(entry) {
			continue
		}
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

	return dedupe(candidates), nil
}

func githubRepoUsable(entry githubAPIRepo) bool {
	return entry.DefaultBranch != "" && entry.Size > 0
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

func searchOrgRepos(ctx context.Context, token, org, query string) ([]githubAPIRepo, error) {
	searchQuery := fmt.Sprintf("%s in:name org:%s archived:false", query, org)
	return searchGitHubRepos(ctx, token, searchQuery)
}

func searchUserRepos(ctx context.Context, token, query string) ([]githubAPIRepo, error) {
	login, err := fetchViewerLogin(ctx, token)
	if err != nil {
		return nil, err
	}
	searchQuery := fmt.Sprintf("%s in:name user:%s archived:false", query, login)
	return searchGitHubRepos(ctx, token, searchQuery)
}

func searchGitHubRepos(ctx context.Context, token, query string) ([]githubAPIRepo, error) {
	client := &http.Client{}
	endpoint := "https://api.github.com/search/repositories?q=" + url.QueryEscape(query) + "&per_page=50"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
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

	var payload githubSearchResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode github search response: %w", err)
	}
	return payload.Items, nil
}

func fetchViewerLogin(ctx context.Context, token string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("github api returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var payload struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	if payload.Login == "" {
		return "", fmt.Errorf("github user login not found")
	}
	return payload.Login, nil
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

func searchCachedGitHubCandidates(query string) ([]RepoCandidate, error) {
	repos, err := loadCachedGitHubCandidates()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return repos, nil
	}

	filtered := make([]RepoCandidate, 0, len(repos))
	for _, candidate := range repos {
		name := strings.ToLower(candidate.DisplayName())
		full := strings.ToLower(candidate.FullName)
		if strings.Contains(name, query) || strings.Contains(full, query) {
			filtered = append(filtered, candidate)
		}
	}
	return filtered, nil
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
