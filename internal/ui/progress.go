package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	bubbleprogress "github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/rohankmr414/grove/internal/repo"
)

type RepoProgress struct {
	out           io.Writer
	isTerminal    bool
	renderedLines int
	bar           bubbleprogress.Model
	tasks         []progressTask
	mu            sync.Mutex
}

type progressTask struct {
	name          string
	status        string
	percent       int
	lastLogged    string
	lastPercent   int
	indeterminate bool
}

func NewRepoProgress(out io.Writer, candidates []repo.RepoCandidate) *RepoProgress {
	tasks := make([]progressTask, 0, len(candidates))
	for _, candidate := range candidates {
		tasks = append(tasks, progressTask{
			name:        candidate.DisplayName(),
			status:      "queued",
			percent:     0,
			lastPercent: -1,
		})
	}

	progress := &RepoProgress{
		out:        out,
		isTerminal: isTerminal(out),
		bar: bubbleprogress.New(
			bubbleprogress.WithWidth(16),
			bubbleprogress.WithoutPercentage(),
			bubbleprogress.WithFillCharacters('━', '─'),
			bubbleprogress.WithSolidFill("242"),
		),
		tasks: tasks,
	}
	progress.bar.EmptyColor = "238"
	if progress.isTerminal {
		progress.renderLocked()
	}
	return progress
}

func (p *RepoProgress) Update(index int, status string, percent int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if index < 0 || index >= len(p.tasks) {
		return
	}

	task := &p.tasks[index]
	task.status = status
	task.indeterminate = percent < 0
	if percent >= 0 {
		if percent > 100 {
			percent = 100
		}
		task.percent = percent
	}

	if p.isTerminal {
		p.renderLocked()
		return
	}

	p.logTaskUpdateLocked(task)
}

func (p *RepoProgress) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isTerminal && p.renderedLines > 0 {
		p.renderLocked()
		fmt.Fprintln(p.out)
	}
}

func (p *RepoProgress) logTaskUpdateLocked(task *progressTask) {
	if task.indeterminate {
		if task.status == task.lastLogged {
			return
		}
		task.lastLogged = task.status
		fmt.Fprintf(p.out, "%s: %s\n", task.name, renderStatus(task.status))
		return
	}

	step := task.percent / 10
	if task.status == task.lastLogged && step == task.lastPercent {
		return
	}
	task.lastLogged = task.status
	task.lastPercent = step
	fmt.Fprintf(p.out, "%s: %s (%d%%)\n", task.name, renderStatus(task.status), task.percent)
}

func (p *RepoProgress) renderLocked() {
	lines := p.linesLocked()
	if len(lines) == 0 {
		return
	}
	if p.renderedLines > 0 {
		fmt.Fprintf(p.out, "\x1b[%dA\x1b[J", p.renderedLines)
	}
	fmt.Fprint(p.out, strings.Join(lines, "\n"))
	fmt.Fprint(p.out, "\n")
	p.renderedLines = len(lines)
}

func (p *RepoProgress) linesLocked() []string {
	ready := 0
	for _, task := range p.tasks {
		if task.status == "ready" || task.status == "cached" {
			ready++
		}
	}

	lines := []string{
		metaStyle.Render(fmt.Sprintf("Preparing repositories %d/%d", ready, len(p.tasks))),
	}
	for _, task := range p.tasks {
		lines = append(lines, fmt.Sprintf("  %-34s %s", faintStyle.Render(truncate(task.name, 34)), p.renderTask(task)))
	}
	return lines
}

func (p *RepoProgress) renderTask(task progressTask) string {
	if task.indeterminate {
		return renderStatus(task.status)
	}
	bar := p.bar.ViewAs(float64(task.percent) / 100)
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		bar,
		faintStyle.Render(fmt.Sprintf(" %3d%%", task.percent)),
		" ",
		renderStatus(task.status),
	)
}

func renderStatus(status string) string {
	switch status {
	case "ready", "using cache":
		return metaStyle.Render(status)
	case "failed":
		return cursorStyle.Render(status)
	default:
		return faintStyle.Render(status)
	}
}

func truncate(value string, width int) string {
	if len(value) <= width {
		return value
	}
	if width <= 3 {
		return value[:width]
	}
	return value[:width-3] + "..."
}

func isTerminal(out io.Writer) bool {
	file, ok := out.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
