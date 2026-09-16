package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/rohankmr414/grove/internal/util"
)

func AddWorktree(ctx context.Context, canonicalPath, worktreePath, branch, defaultBranch string) error {
	if err := util.EnsureDir(filepath.Dir(worktreePath)); err != nil {
		return err
	}

	if err := DetachHead(ctx, canonicalPath); err != nil {
		return err
	}

	if BranchExists(ctx, canonicalPath, branch) {
		return util.Run(ctx, "git", "-C", canonicalPath, "worktree", "add", worktreePath, branch)
	}

	// Fetch updates the remote-tracking branch, while the cache's local default
	// branch can remain at the revision that was current when it was cloned.
	// Start new workspace branches from the fetched ref so new worktrees contain
	// the latest remote revision.
	startPoint := defaultBranch
	remoteDefaultBranch := "refs/remotes/origin/" + defaultBranch
	if refExists(ctx, canonicalPath, remoteDefaultBranch) {
		startPoint = remoteDefaultBranch
	}
	return util.Run(ctx, "git", "-C", canonicalPath, "worktree", "add", "-b", branch, worktreePath, startPoint)
}

func RemoveWorktree(ctx context.Context, canonicalPath, worktreePath string) error {
	registered, err := worktreeRegistered(ctx, canonicalPath, worktreePath)
	if err != nil {
		return err
	}
	if !registered {
		// A previous removal may have deleted the registration but left .git
		// and other files behind. Only accept a verifiable stale worktree.
		root, err := staleWorktreeRoot(ctx, worktreePath)
		if err != nil || root != filepath.Clean(canonicalPath) {
			return fmt.Errorf("refusing to remove unregistered worktree %s: cannot verify stale worktree metadata", worktreePath)
		}
	} else if err := util.Run(ctx, "git", "-C", canonicalPath, "worktree", "remove", worktreePath, "--force"); err != nil {
		// Git removes the registration even when recursive deletion fails.
		// Never bypass a refusal that leaves the worktree registered (e.g. a lock).
		registered, checkErr := worktreeRegistered(ctx, canonicalPath, worktreePath)
		if checkErr != nil || registered || ctx.Err() != nil {
			return err
		}
	}
	return removeWorktreeFiles(ctx, worktreePath)
}

func worktreeRegistered(ctx context.Context, canonicalPath, worktreePath string) (bool, error) {
	output, err := util.Output(ctx, "git", "-C", canonicalPath, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return false, err
	}
	for _, field := range strings.Split(output, "\x00") {
		if path, ok := strings.CutPrefix(field, "worktree "); ok && filepath.Clean(path) == filepath.Clean(worktreePath) {
			return true, nil
		}
	}
	return false, nil
}

func removeWorktreeFiles(ctx context.Context, path string) error {
	for attempt := 0; attempt < 3; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := os.RemoveAll(path)
		if err == nil {
			return nil
		}
		if !errors.Is(err, syscall.ENOTEMPTY) || attempt == 2 {
			return fmt.Errorf("remove leftover worktree files at %s (stop any processes writing here, then retry): %w", path, err)
		}
		timer := time.NewTimer(100 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return nil
}

func CanonicalRoot(ctx context.Context, repoPath string) (string, error) {
	commonDir, err := util.Output(ctx, "git", "-C", repoPath, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		if root, staleErr := staleWorktreeRoot(ctx, repoPath); staleErr == nil {
			return root, nil
		}
		return "", err
	}
	return filepath.Dir(commonDir), nil
}

// staleWorktreeRoot recognizes only missing metadata under a canonical clone's
// .git/worktrees directory. Malformed pointers and existing metadata are errors.
func staleWorktreeRoot(ctx context.Context, repoPath string) (string, error) {
	data, err := os.ReadFile(filepath.Join(repoPath, ".git"))
	if err != nil {
		return "", err
	}
	gitDir, ok := strings.CutPrefix(strings.TrimSpace(string(data)), "gitdir: ")
	if !ok || gitDir == "" {
		return "", fmt.Errorf("invalid worktree gitdir at %s", repoPath)
	}
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(repoPath, gitDir)
	}
	gitDir = filepath.Clean(gitDir)
	worktreesDir := filepath.Dir(gitDir)
	commonDir := filepath.Dir(worktreesDir)
	if filepath.Base(worktreesDir) != "worktrees" || filepath.Base(commonDir) != ".git" {
		return "", fmt.Errorf("unexpected worktree gitdir %s", gitDir)
	}
	if _, err := os.Lstat(gitDir); !os.IsNotExist(err) {
		return "", fmt.Errorf("worktree metadata is not missing at %s", gitDir)
	}
	root := filepath.Dir(commonDir)
	actual, err := util.Output(ctx, "git", "-C", root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	if filepath.Clean(actual) != commonDir {
		return "", fmt.Errorf("canonical git directory does not match %s", commonDir)
	}
	return root, nil
}

func DetachHead(ctx context.Context, canonicalPath string) error {
	head, err := util.Output(ctx, "git", "-C", canonicalPath, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return nil
	}

	if head == "" {
		return nil
	}

	if err := util.Run(ctx, "git", "-C", canonicalPath, "switch", "--detach", "--quiet"); err != nil {
		return fmt.Errorf("detach cache HEAD at %s: %w", canonicalPath, err)
	}

	return nil
}
