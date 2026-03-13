package config

import (
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
