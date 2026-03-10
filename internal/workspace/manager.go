package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rohankmr414/grove/internal/config"
	"github.com/rohankmr414/grove/internal/git"
	"github.com/rohankmr414/grove/internal/repo"
	"github.com/rohankmr414/grove/internal/util"
)

type Manager struct {
	cfg config.Config
}

func NewManager(cfg config.Config) Manager {
	return Manager{cfg: cfg}
}

func (m Manager) AssertDoesNotExist(name string) error {
	workspacePath := filepath.Join(m.cfg.WorkspaceRoot, name)
	if _, err := os.Stat(workspacePath); err == nil {
		return fmt.Errorf("workspace %q already exists at %s; use `grove add` to modify it", name, workspacePath)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (m Manager) Init(ctx context.Context, name string, candidates []repo.RepoCandidate) error {
	workspacePath := filepath.Join(m.cfg.WorkspaceRoot, name)
	if err := m.AssertDoesNotExist(name); err != nil {
		return err
	}

	fmt.Printf("Initializing workspace %q\n", name)
	fmt.Printf("Workspace path: %s\n", workspacePath)
	fmt.Printf("Repositories selected: %d\n\n", len(candidates))

	if err := util.EnsureDir(workspacePath); err != nil {
		return err
	}

	ws := Workspace{Name: name, Path: workspacePath}
	if err := writeMarker(ws); err != nil {
		return err
	}

	if err := m.AddRepositories(ctx, ws, candidates); err != nil {
		return err
	}

	fmt.Printf("\nWorkspace ready: %s\n", ws.Path)
	return nil
}

func (m Manager) AddRepositories(ctx context.Context, ws Workspace, candidates []repo.RepoCandidate) error {
	for i, candidate := range candidates {
		fmt.Printf("[%d/%d] %s\n", i+1, len(candidates), candidate.DisplayName())

		canonicalPath := candidate.CachePath(m.cfg.RepoCacheRoot)
		if _, err := os.Stat(canonicalPath); err == nil {
			fmt.Printf("  using cached clone: %s\n", canonicalPath)
		} else {
			fmt.Printf("  cloning into cache: %s\n", canonicalPath)
		}

		canonicalPath, err := git.EnsureCanonicalClone(ctx, candidate, m.cfg.RepoCacheRoot)
		if err != nil {
			return fmt.Errorf("prepare clone for %s: %w", candidate.DisplayName(), err)
		}

		fmt.Printf("  fetching refs\n")
		if err := git.Fetch(ctx, canonicalPath); err != nil {
			return fmt.Errorf("fetch %s: %w", candidate.DisplayName(), err)
		}

		defaultBranch := candidate.DefaultBranch
		if defaultBranch == "" {
			defaultBranch = "main"
		}

		branch := workspaceBranch(ws.Name)
		worktreePath := filepath.Join(ws.Path, candidate.Name)
		if _, err := os.Stat(worktreePath); err == nil {
			fmt.Printf("  worktree already exists: %s\n\n", worktreePath)
			continue
		}

		fmt.Printf("  creating worktree: %s (%s)\n", worktreePath, branch)
		if err := git.AddWorktree(ctx, canonicalPath, worktreePath, branch, defaultBranch); err != nil {
			return fmt.Errorf("create worktree for %s: %w", candidate.DisplayName(), err)
		}
		fmt.Printf("  done\n\n")
	}
	return nil
}

func (m Manager) DetectCurrent() (Workspace, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Workspace{}, err
	}

	for dir := cwd; dir != "/" && dir != "."; dir = filepath.Dir(dir) {
		if _, err := os.Stat(markerPath(dir)); err == nil {
			return readMarker(dir)
		}
		next := filepath.Dir(dir)
		if next == dir {
			break
		}
	}

	return Workspace{}, fmt.Errorf("current directory is not inside a grove workspace")
}

func (m Manager) ResolveWorkspace(args []string) (Workspace, error) {
	if len(args) == 1 {
		workspacePath := filepath.Join(m.cfg.WorkspaceRoot, args[0])
		return readMarker(workspacePath)
	}
	return m.DetectCurrent()
}

func (m Manager) Status(ctx context.Context, args []string) (StatusReport, error) {
	ws, err := m.ResolveWorkspace(args)
	if err != nil {
		return StatusReport{}, err
	}

	entries, err := os.ReadDir(ws.Path)
	if err != nil {
		return StatusReport{}, err
	}

	var report StatusReport
	report.Name = ws.Name

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		repoPath := filepath.Join(ws.Path, entry.Name())
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); err != nil {
			continue
		}

		branch, err := git.CurrentBranch(ctx, repoPath)
		if err != nil {
			return StatusReport{}, err
		}
		state, err := repoState(ctx, repoPath)
		if err != nil {
			return StatusReport{}, err
		}

		report.Repositories = append(report.Repositories, StatusEntry{
			Name:   entry.Name(),
			Branch: branch,
			State:  state,
		})
	}

	sort.Slice(report.Repositories, func(i, j int) bool {
		return report.Repositories[i].Name < report.Repositories[j].Name
	})

	return report, nil
}

func (m Manager) Remove(ctx context.Context, name string) error {
	ws, err := readMarker(filepath.Join(m.cfg.WorkspaceRoot, name))
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(ws.Path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		repoPath := filepath.Join(ws.Path, entry.Name())
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); err != nil {
			continue
		}
		canonical, err := git.CanonicalRoot(ctx, repoPath)
		if err != nil {
			return err
		}
		if err := git.RemoveWorktree(ctx, canonical, repoPath); err != nil {
			return err
		}
	}

	return os.RemoveAll(ws.Path)
}

func (m Manager) List() ([]Workspace, error) {
	entries, err := os.ReadDir(m.cfg.WorkspaceRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	workspaces := make([]Workspace, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		wsPath := filepath.Join(m.cfg.WorkspaceRoot, entry.Name())
		ws, err := readMarker(wsPath)
		if err != nil {
			continue
		}
		workspaces = append(workspaces, ws)
	}

	sort.Slice(workspaces, func(i, j int) bool {
		return workspaces[i].Name < workspaces[j].Name
	})

	return workspaces, nil
}

func workspaceBranch(name string) string {
	if strings.Contains(name, "/") {
		return name
	}
	return "feature/" + name
}

func repoState(ctx context.Context, repoPath string) (string, error) {
	output, err := util.Output(ctx, "git", "-C", repoPath, "status", "--porcelain")
	if err != nil {
		return "", err
	}
	if output == "" {
		return "clean", nil
	}
	return "modified", nil
}
