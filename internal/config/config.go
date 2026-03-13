package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rohankmr414/grove/internal/util"
)

type GitHubConfig struct {
	Enabled bool
	Orgs    []string
}

type Config struct {
	WorkspaceRoot string
	RepoCacheRoot string
	GitHub        GitHubConfig
}

func ConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Clean(filepath.Join(xdg, "grove")), nil
	}
	return util.ExpandPath("~/.config/grove")
}

func ConfigPath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}

func WorkspaceInitDir() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "workspace-init"), nil
}

func Load() (Config, error) {
	cfg := defaults()

	configPath, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	section := ""
	listKey := ""
	cfg.GitHub.Orgs = nil

	for scanner.Scan() {
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		if strings.HasPrefix(trimmed, "- ") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			switch {
			case section == "github" && listKey == "orgs":
				cfg.GitHub.Orgs = append(cfg.GitHub.Orgs, value)
			}
			continue
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			return Config{}, fmt.Errorf("invalid config line: %q", raw)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		listKey = ""

		if indent == 0 {
			section = ""
			switch key {
			case "workspace_root":
				cfg.WorkspaceRoot = mustExpand(value)
			case "repo_cache_root":
				cfg.RepoCacheRoot = mustExpand(value)
			case "github":
				section = "github"
			case "local_roots":
				listKey = key
			}
			continue
		}

		if section == "github" {
			switch key {
			case "enabled":
				cfg.GitHub.Enabled = strings.EqualFold(value, "true")
			case "orgs":
				listKey = key
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	if cfg.WorkspaceRoot == "" || cfg.RepoCacheRoot == "" {
		return Config{}, fmt.Errorf("config must define workspace_root and repo_cache_root")
	}

	return cfg, nil
}

func defaults() Config {
	return Config{
		WorkspaceRoot: mustExpand("~/groves"),
		RepoCacheRoot: mustExpand("~/.grove/repos"),
		GitHub: GitHubConfig{
			Enabled: true,
		},
	}
}

func mustExpand(path string) string {
	expanded, err := util.ExpandPath(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return expanded
}
