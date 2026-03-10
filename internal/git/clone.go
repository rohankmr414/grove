package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rohankmr414/grove/internal/repo"
	"github.com/rohankmr414/grove/internal/util"
)

type CloneProgress struct {
	Stage   string
	Message string
	Percent int
}

func EnsureCanonicalClone(ctx context.Context, candidate repo.RepoCandidate, cacheRoot string, progress func(CloneProgress)) (string, error) {
	canonicalPath := candidate.CachePath(cacheRoot)
	if _, err := os.Stat(canonicalPath); err == nil {
		return canonicalPath, nil
	}

	if err := util.EnsureDir(filepath.Dir(canonicalPath)); err != nil {
		return "", err
	}
	if err := runClone(ctx, candidate.CloneURL, canonicalPath, progress); err != nil {
		return "", err
	}
	return canonicalPath, nil
}

func runClone(ctx context.Context, cloneURL, canonicalPath string, progress func(CloneProgress)) error {
	cmd := exec.CommandContext(ctx, "git", "clone", "--progress", cloneURL, canonicalPath)
	cmd.Stdout = io.Discard

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	var errorOutput bytes.Buffer
	done := make(chan struct{})
	go func() {
		defer close(done)
		streamCloneProgress(stderr, func(line string) {
			if update, ok := parseCloneProgress(line); ok {
				if progress != nil {
					progress(update)
				}
				return
			}

			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				return
			}
			if errorOutput.Len() > 0 {
				errorOutput.WriteByte('\n')
			}
			errorOutput.WriteString(trimmed)
		})
	}()

	if err := cmd.Start(); err != nil {
		return err
	}

	waitErr := cmd.Wait()
	<-done
	if waitErr != nil {
		msg := strings.TrimSpace(errorOutput.String())
		if msg == "" {
			return fmt.Errorf("git clone %s %s failed: %w", cloneURL, canonicalPath, waitErr)
		}
		return fmt.Errorf("git clone %s %s failed: %w\n%s", cloneURL, canonicalPath, waitErr, msg)
	}

	if progress != nil {
		progress(CloneProgress{Stage: "done", Message: "clone complete", Percent: 100})
	}
	return nil
}

func streamCloneProgress(reader io.Reader, handle func(string)) {
	buffered := bufio.NewReader(reader)
	var current strings.Builder

	flush := func() {
		if current.Len() == 0 {
			return
		}
		handle(current.String())
		current.Reset()
	}

	for {
		b, err := buffered.ReadByte()
		if err != nil {
			if err == io.EOF {
				flush()
			}
			return
		}

		switch b {
		case '\r', '\n':
			flush()
		default:
			current.WriteByte(b)
		}
	}
}

func parseCloneProgress(line string) (CloneProgress, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return CloneProgress{}, false
	}

	if strings.HasPrefix(trimmed, "Cloning into ") {
		return CloneProgress{Stage: "clone", Message: "starting clone", Percent: 0}, true
	}

	stageWeights := []struct {
		Prefix string
		Stage  string
		Base   int
		Span   int
		Label  string
	}{
		{Prefix: "remote: Counting objects:", Stage: "clone", Base: 0, Span: 5, Label: "counting objects"},
		{Prefix: "remote: Compressing objects:", Stage: "clone", Base: 5, Span: 10, Label: "compressing objects"},
		{Prefix: "Receiving objects:", Stage: "clone", Base: 15, Span: 75, Label: "receiving objects"},
		{Prefix: "Resolving deltas:", Stage: "clone", Base: 90, Span: 10, Label: "resolving deltas"},
	}

	for _, stage := range stageWeights {
		if !strings.HasPrefix(trimmed, stage.Prefix) {
			continue
		}

		rawPercent, ok := extractPercent(trimmed)
		if !ok {
			return CloneProgress{Stage: stage.Stage, Message: stage.Label, Percent: stage.Base}, true
		}

		percent := stage.Base + (rawPercent * stage.Span / 100)
		if percent > 100 {
			percent = 100
		}
		return CloneProgress{
			Stage:   stage.Stage,
			Message: stage.Label,
			Percent: percent,
		}, true
	}

	if strings.HasPrefix(trimmed, "Updating files:") {
		rawPercent, ok := extractPercent(trimmed)
		if !ok {
			rawPercent = 100
		}
		return CloneProgress{
			Stage:   "clone",
			Message: "updating files",
			Percent: rawPercent,
		}, true
	}

	return CloneProgress{}, false
}

func extractPercent(line string) (int, bool) {
	percentIndex := strings.IndexByte(line, '%')
	if percentIndex == -1 {
		return 0, false
	}

	start := percentIndex - 1
	for start >= 0 && line[start] >= '0' && line[start] <= '9' {
		start--
	}

	value, err := strconv.Atoi(strings.TrimSpace(line[start+1 : percentIndex]))
	if err != nil {
		return 0, false
	}
	return value, true
}
