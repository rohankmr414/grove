package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rohankmr414/grove/internal/util"
	"gopkg.in/yaml.v3"
)

type GitHubConfig struct {
	Enabled bool     `yaml:"enabled"`
	Orgs    []string `yaml:"orgs"`
}

type Config struct {
	WorkspaceRoot string       `yaml:"workspace_root"`
	RepoCacheRoot string       `yaml:"repo_cache_root"`
	GitHub        GitHubConfig `yaml:"github"`
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

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.normalize(); err != nil {
		return Config{}, err
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

func (c *Config) normalize() error {
	c.WorkspaceRoot = mustExpand(c.WorkspaceRoot)
	c.RepoCacheRoot = mustExpand(c.RepoCacheRoot)

	if c.WorkspaceRoot == "" || c.RepoCacheRoot == "" {
		return fmt.Errorf("config must define workspace_root and repo_cache_root")
	}
	return nil
}

func mustExpand(path string) string {
	expanded, err := util.ExpandPath(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return expanded
}
