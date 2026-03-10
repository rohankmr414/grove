package ui

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rohankmr414/grove/internal/repo"
	"github.com/rohankmr414/grove/internal/util"
)

func PickRepositories(candidates []repo.RepoCandidate) ([]repo.RepoCandidate, error) {
	if util.LookPath("fzf") {
		return pickWithFZF(candidates)
	}
	return pickFromPrompt(candidates)
}

func pickWithFZF(candidates []repo.RepoCandidate) ([]repo.RepoCandidate, error) {
	cmd := exec.Command("fzf", "--multi", "--prompt", "Select repositories> ")
	var input bytes.Buffer
	for i, candidate := range candidates {
		fmt.Fprintf(&input, "%d\t%s\n", i, candidate.DisplayName())
	}
	cmd.Stdin = &input
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 130 {
			return nil, nil
		}
		return nil, fmt.Errorf("fzf selection failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	selected := make([]repo.RepoCandidate, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		idxText, _, found := strings.Cut(line, "\t")
		if !found {
			continue
		}
		idx, err := strconv.Atoi(idxText)
		if err != nil || idx < 0 || idx >= len(candidates) {
			continue
		}
		selected = append(selected, candidates[idx])
	}
	return selected, nil
}

func pickFromPrompt(candidates []repo.RepoCandidate) ([]repo.RepoCandidate, error) {
	for i, candidate := range candidates {
		fmt.Printf("%d. %s\n", i+1, candidate.DisplayName())
	}
	fmt.Print("Select repositories by number (comma separated): ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	fields := strings.Split(strings.TrimSpace(line), ",")
	selected := make([]repo.RepoCandidate, 0, len(fields))
	seen := map[int]struct{}{}

	for _, field := range fields {
		if strings.TrimSpace(field) == "" {
			continue
		}
		idx, err := strconv.Atoi(strings.TrimSpace(field))
		if err != nil {
			return nil, fmt.Errorf("invalid selection %q", field)
		}
		idx--
		if idx < 0 || idx >= len(candidates) {
			return nil, fmt.Errorf("selection %d out of range", idx+1)
		}
		if _, ok := seen[idx]; ok {
			continue
		}
		seen[idx] = struct{}{}
		selected = append(selected, candidates[idx])
	}

	return selected, nil
}
