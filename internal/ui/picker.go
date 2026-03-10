package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rohankmr414/grove/internal/repo"
	"github.com/sahilm/fuzzy"
)

const maxVisibleResults = 12

type pickerModel struct {
	input      textinput.Model
	candidates []repo.RepoCandidate
	displays   []string
	matches    []int
	cursor     int
	selected   map[int]struct{}
	cancelled  bool
	confirmed  bool
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230"))
	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	activeRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("25")).
			Padding(0, 1)
	rowStyle = lipgloss.NewStyle().
			Padding(0, 1)
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)
	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))
)

func PickRepositories(candidates []repo.RepoCandidate) ([]repo.RepoCandidate, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	model := newPickerModel(candidates)
	result, err := tea.NewProgram(model, tea.WithAltScreen()).Run()
	if err != nil {
		return nil, err
	}

	finalModel, ok := result.(pickerModel)
	if !ok {
		return nil, fmt.Errorf("unexpected picker model type %T", result)
	}
	if finalModel.cancelled {
		return nil, nil
	}
	if !finalModel.confirmed {
		return nil, fmt.Errorf("repository selection did not complete")
	}

	return collectSelected(finalModel.candidates, finalModel.selected), nil
}

func newPickerModel(candidates []repo.RepoCandidate) pickerModel {
	input := textinput.New()
	input.Prompt = "Search> "
	input.Placeholder = "type to fuzzy filter repositories"
	input.Focus()
	input.CharLimit = 256

	displays := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		displays = append(displays, candidate.DisplayName())
	}

	model := pickerModel{
		input:      input,
		candidates: candidates,
		displays:   displays,
		selected:   make(map[int]struct{}),
	}
	model.refreshMatches()
	return model
}

func (m pickerModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "ctrl+n":
			if m.cursor < len(m.matches)-1 {
				m.cursor++
			}
			return m, nil
		case "tab", " ":
			m.toggleCurrent()
			return m, nil
		case "enter":
			if len(m.selected) == 0 && len(m.matches) > 0 {
				m.toggleCurrent()
			}
			if len(m.selected) == 0 {
				return m, nil
			}
			m.confirmed = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.refreshMatches()
	return m, cmd
}

func (m pickerModel) View() string {
	var builder strings.Builder

	builder.WriteString(titleStyle.Render("Repository Picker"))
	builder.WriteString("\n")
	builder.WriteString(subtitleStyle.Render("Multi-select repositories for this workspace"))
	builder.WriteString("\n\n")

	if len(m.matches) == 0 {
		builder.WriteString(mutedStyle.Render("  No matching repositories."))
		builder.WriteString("\n\n")
		builder.WriteString(statusLine(m))
		builder.WriteString("\n")
		builder.WriteString(searchLine(m))
		builder.WriteString("\n")
		builder.WriteString(helpLine())
		return builder.String()
	}

	limit := len(m.matches)
	if limit > maxVisibleResults {
		limit = maxVisibleResults
	}

	start := 0
	if m.cursor >= limit {
		start = m.cursor - limit + 1
	}
	end := start + limit
	if end > len(m.matches) {
		end = len(m.matches)
	}

	for visibleIndex := start; visibleIndex < end; visibleIndex++ {
		matchIndex := m.matches[visibleIndex]
		pointer := " "
		if visibleIndex == m.cursor {
			pointer = "▌"
		}

		marker := "○"
		if _, ok := m.selected[matchIndex]; ok {
			marker = selectedStyle.Render("●")
		}

		line := fmt.Sprintf("%s %s  %s", pointer, marker, m.candidates[matchIndex].DisplayName())
		if visibleIndex == m.cursor {
			builder.WriteString(activeRowStyle.Render(line))
		} else {
			builder.WriteString(rowStyle.Render(line))
		}
		builder.WriteString("\n")
	}

	if len(m.matches) > end {
		builder.WriteString("\n")
		builder.WriteString(mutedStyle.Render(fmt.Sprintf("... %d more matches", len(m.matches)-end)))
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	builder.WriteString(statusLine(m))
	builder.WriteString("\n")
	builder.WriteString(searchLine(m))
	builder.WriteString("\n")
	builder.WriteString(helpLine())

	return builder.String()
}

func statusLine(m pickerModel) string {
	return subtitleStyle.Render(fmt.Sprintf("%d matches  •  %d selected", len(m.matches), len(m.selected)))
}

func searchLine(m pickerModel) string {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Render(m.input.View())
}

func helpLine() string {
	return helpStyle.Render("↑/↓ move  tab/space toggle  enter confirm  esc cancel")
}

func (m *pickerModel) refreshMatches() {
	query := strings.TrimSpace(m.input.Value())
	m.matches = m.matches[:0]

	if query == "" {
		for index := range m.candidates {
			m.matches = append(m.matches, index)
		}
	} else {
		results := fuzzy.Find(query, m.displays)
		for _, result := range results {
			m.matches = append(m.matches, result.Index)
		}
	}

	if len(m.matches) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.matches) {
		m.cursor = len(m.matches) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *pickerModel) toggleCurrent() {
	if len(m.matches) == 0 || m.cursor < 0 || m.cursor >= len(m.matches) {
		return
	}

	current := m.matches[m.cursor]
	if _, ok := m.selected[current]; ok {
		delete(m.selected, current)
		return
	}
	m.selected[current] = struct{}{}
}

func collectSelected(candidates []repo.RepoCandidate, selected map[int]struct{}) []repo.RepoCandidate {
	indexes := make([]int, 0, len(selected))
	for index := range selected {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)

	result := make([]repo.RepoCandidate, 0, len(indexes))
	for _, index := range indexes {
		result = append(result, candidates[index])
	}
	return result
}
