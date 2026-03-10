package repo

import "path/filepath"

type RepoCandidate struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
	LocalPath     string `json:"local_path"`
	Source        string `json:"source"`
}

func (c RepoCandidate) DisplayName() string {
	switch c.Source {
	case "cache":
		if c.FullName != "" {
			return "[cache] " + c.FullName
		}
		return "[cache] " + c.LocalPath
	case "github":
		if c.FullName != "" {
			return "[gh] " + c.FullName
		}
	}
	if c.FullName != "" {
		return "[" + c.Source + "] " + c.FullName
	}
	return "[" + c.Source + "] " + c.Name
}

func (c RepoCandidate) CachePath(root string) string {
	owner, name := c.OwnerRepo()
	return filepath.Join(root, owner, name)
}

func (c RepoCandidate) OwnerRepo() (string, string) {
	if c.FullName != "" {
		for i := 0; i < len(c.FullName); i++ {
			if c.FullName[i] == '/' {
				return c.FullName[:i], c.FullName[i+1:]
			}
		}
	}
	return "local", c.Name
}
