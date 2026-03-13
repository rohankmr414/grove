package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigPathDefaultsToHomeConfigDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	got, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}

	want := filepath.Join(home, ".config", "grove", "config.yaml")
	if got != want {
		t.Fatalf("ConfigPath() = %q, want %q", got, want)
	}
}

func TestConfigPathUsesXDGConfigHomeWhenSet(t *testing.T) {
	xdg := filepath.Join(t.TempDir(), "xdg-config")
	t.Setenv("XDG_CONFIG_HOME", xdg)

	got, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}

	want := filepath.Join(xdg, "grove", "config.yaml")
	if got != want {
		t.Fatalf("ConfigPath() = %q, want %q", got, want)
	}
}

func TestWorkspaceInitDirExpandsHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	got, err := WorkspaceInitDir()
	if err != nil {
		t.Fatalf("WorkspaceInitDir: %v", err)
	}

	want := filepath.Join(home, ".config", "grove", "workspace-init")
	if got != want {
		t.Fatalf("WorkspaceInitDir() = %q, want %q", got, want)
	}
}

func TestWorkspaceInitDirUsesXDGConfigHomeWhenSet(t *testing.T) {
	xdg := filepath.Join(t.TempDir(), "xdg-config")
	t.Setenv("XDG_CONFIG_HOME", xdg)

	got, err := WorkspaceInitDir()
	if err != nil {
		t.Fatalf("WorkspaceInitDir: %v", err)
	}

	want := filepath.Join(xdg, "grove", "workspace-init")
	if got != want {
		t.Fatalf("WorkspaceInitDir() = %q, want %q", got, want)
	}
}

func TestLoadParsesYAMLConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	configPath, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	configData := []byte(`
# Workspace defaults
workspace_root: ~/work/groves
repo_cache_root: ~/.cache/grove/repos

github:
  enabled: false
  orgs:
    - acme
    - grove
`)
	if err := os.WriteFile(configPath, configData, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got, want := cfg.WorkspaceRoot, filepath.Join(home, "work", "groves"); got != want {
		t.Fatalf("WorkspaceRoot = %q, want %q", got, want)
	}
	if got, want := cfg.RepoCacheRoot, filepath.Join(home, ".cache", "grove", "repos"); got != want {
		t.Fatalf("RepoCacheRoot = %q, want %q", got, want)
	}
	if cfg.GitHub.Enabled {
		t.Fatalf("GitHub.Enabled = true, want false")
	}
	wantOrgs := []string{"acme", "grove"}
	if len(cfg.GitHub.Orgs) != len(wantOrgs) {
		t.Fatalf("GitHub.Orgs length = %d, want %d", len(cfg.GitHub.Orgs), len(wantOrgs))
	}
	for i, want := range wantOrgs {
		if got := cfg.GitHub.Orgs[i]; got != want {
			t.Fatalf("GitHub.Orgs[%d] = %q, want %q", i, got, want)
		}
	}
}

func TestLoadUsesDefaultsForMissingConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got, want := cfg.WorkspaceRoot, filepath.Join(home, "groves"); got != want {
		t.Fatalf("WorkspaceRoot = %q, want %q", got, want)
	}
	if got, want := cfg.RepoCacheRoot, filepath.Join(home, ".grove", "repos"); got != want {
		t.Fatalf("RepoCacheRoot = %q, want %q", got, want)
	}
	if !cfg.GitHub.Enabled {
		t.Fatalf("GitHub.Enabled = false, want true")
	}
}
