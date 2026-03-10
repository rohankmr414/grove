package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/rohankmr414/grove/internal/config"
	"github.com/rohankmr414/grove/internal/git"
	"github.com/rohankmr414/grove/internal/repo"
	"github.com/rohankmr414/grove/internal/ui"
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
		return fmt.Errorf("workspace %q already exists at %s; use `grove repo add` to modify it", name, workspacePath)
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
	prepared, err := m.prepareRepositories(ctx, candidates)
	if err != nil {
		return err
	}

	for i, candidate := range candidates {
		canonicalPath := prepared[i].canonicalPath

		defaultBranch := candidate.DefaultBranch
		if defaultBranch == "" {
			defaultBranch = "main"
		}

		branch := workspaceBranch(ws.Name)
		worktreePath := filepath.Join(ws.Path, candidate.Name)
		if _, err := os.Stat(worktreePath); err == nil {
			continue
		}

		if err := git.AddWorktree(ctx, canonicalPath, worktreePath, branch, defaultBranch); err != nil {
			return fmt.Errorf("create worktree for %s: %w", candidate.DisplayName(), err)
		}
	}
	return m.syncEditorWorkspace(ws)
}

type preparedRepository struct {
	canonicalPath string
}

func (m Manager) prepareRepositories(ctx context.Context, candidates []repo.RepoCandidate) ([]preparedRepository, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	progress := ui.NewRepoProgress(os.Stdout, candidates)
	defer progress.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make([]preparedRepository, len(candidates))
	workerCount := min(len(candidates), max(2, min(runtime.GOMAXPROCS(0), 6)))
	jobs := make(chan int)

	var (
		wg       sync.WaitGroup
		firstErr error
		errMu    sync.Mutex
	)

	setErr := func(err error) {
		errMu.Lock()
		defer errMu.Unlock()
		if firstErr != nil {
			return
		}
		firstErr = err
		cancel()
	}

	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				if ctx.Err() != nil {
					return
				}
				if err := m.prepareRepository(ctx, candidates[index], index, progress, &results[index]); err != nil {
					setErr(err)
					progress.Update(index, "failed", -1)
					return
				}
			}
		}()
	}

	for index := range candidates {
		if ctx.Err() != nil {
			break
		}
		jobs <- index
	}
	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}

func (m Manager) prepareRepository(ctx context.Context, candidate repo.RepoCandidate, index int, progress *ui.RepoProgress, result *preparedRepository) error {
	canonicalPath := candidate.CachePath(m.cfg.RepoCacheRoot)
	if _, err := os.Stat(canonicalPath); err == nil {
		progress.Update(index, "using cache", 100)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check cache for %s: %w", candidate.DisplayName(), err)
	} else {
		progress.Update(index, "starting clone", 0)
		var cloneErr error
		canonicalPath, cloneErr = git.EnsureCanonicalClone(ctx, candidate, m.cfg.RepoCacheRoot, func(update git.CloneProgress) {
			progress.Update(index, update.Message, update.Percent)
		})
		if cloneErr != nil {
			return fmt.Errorf("prepare clone for %s: %w", candidate.DisplayName(), cloneErr)
		}
	}

	progress.Update(index, "fetching refs", -1)
	if err := git.Fetch(ctx, canonicalPath); err != nil {
		return fmt.Errorf("fetch %s: %w", candidate.DisplayName(), err)
	}

	result.canonicalPath = canonicalPath
	progress.Update(index, "ready", 100)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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

	repositories, err := m.WorkspaceRepositories(ws)
	if err != nil {
		return err
	}

	names := make([]string, 0, len(repositories))
	for _, repository := range repositories {
		names = append(names, repository.Name)
	}

	if err := m.RemoveRepositories(ctx, ws, names...); err != nil {
		return err
	}

	return os.RemoveAll(ws.Path)
}

func (m Manager) WorkspaceRepositories(ws Workspace) ([]repo.RepoCandidate, error) {
	entries, err := os.ReadDir(ws.Path)
	if err != nil {
		return nil, err
	}

	repositories := make([]repo.RepoCandidate, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoPath := filepath.Join(ws.Path, entry.Name())
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); err != nil {
			continue
		}

		repositories = append(repositories, repo.RepoCandidate{
			Name:      entry.Name(),
			FullName:  entry.Name(),
			LocalPath: repoPath,
			Source:    "workspace",
		})
	}

	sort.Slice(repositories, func(i, j int) bool {
		return repositories[i].Name < repositories[j].Name
	})

	return repositories, nil
}

func (m Manager) RemoveRepositories(ctx context.Context, ws Workspace, names ...string) error {
	for _, name := range names {
		repoPath := filepath.Join(ws.Path, name)
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		canonical, err := git.CanonicalRoot(ctx, repoPath)
		if err != nil {
			return err
		}
		if err := git.RemoveWorktree(ctx, canonical, repoPath); err != nil {
			return err
		}
	}

	return m.syncEditorWorkspace(ws)
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
