package ui

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

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
	matches    []int
	cursor     int
	width      int
	height     int
	selected   map[int]struct{}
	cancelled  bool
	confirmed  bool
}

var (
	promptStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	metaStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("186"))
	faintStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	selectedGutterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	selectedCursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	activeStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	selectedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
)

func PickRepositories(candidates []repo.RepoCandidate) ([]repo.RepoCandidate, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	model := newPickerModel(candidates)
	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return nil, err
	}

	finalModel, ok := result.(pickerModel)
	if !ok {
		return nil, fmt.Errorf("unexpected picker model type %T", result)
	}
	clearRenderedPicker(finalModel.View())
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
	input.Prompt = promptStyle.Render("> ")
	input.Placeholder = "type to search"
	input.PlaceholderStyle = faintStyle
	input.Focus()
	input.CharLimit = 256

	model := pickerModel{
		input:      input,
		candidates: candidates,
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
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
	previous := m.input.Value()
	m.input, cmd = m.input.Update(msg)
	if previous != m.input.Value() {
		m.refreshMatches()
		return m, cmd
	}
	m.refreshMatches()
	return m, cmd
}

func (m pickerModel) View() string {
	var builder strings.Builder

	builder.WriteString(searchLine(m))
	builder.WriteString("\n")
	builder.WriteString(statusLine(m))
	builder.WriteString("\n")

	if len(m.matches) == 0 {
		builder.WriteString(faintStyle.Render("  no matches"))
		return builder.String()
	}

	limit := m.visibleRows()

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
		_, isSelected := m.selected[matchIndex]
		isActive := visibleIndex == m.cursor

		gutter := " "
		lineStyle := lipgloss.NewStyle()

		if isSelected {
			gutter = selectedGutterStyle.Render("▪")
			lineStyle = selectedStyle
		}
		if isActive {
			gutter = cursorStyle.Render("▌")
			lineStyle = activeStyle
		}
		if isSelected && isActive {
			gutter = selectedCursorStyle.Render("▌")
			lineStyle = activeStyle.Bold(true)
		}

		row := fmt.Sprintf("%s %s", gutter, m.candidates[matchIndex].DisplayName())
		builder.WriteString(lineStyle.Render(truncateToWidth(row, m.width)))
		builder.WriteString("\n")
	}

	if len(m.matches) > end {
		builder.WriteString(faintStyle.Render(truncateToWidth(fmt.Sprintf("... %d more", len(m.matches)-end), m.width)))
		builder.WriteString("\n")
	}

	return strings.TrimRight(builder.String(), "\n")
}

func statusLine(m pickerModel) string {
	summary := metaStyle.Render(fmt.Sprintf("%d/%d", len(m.matches), len(m.candidates)))
	if len(m.selected) > 0 {
		summary += " " + metaStyle.Render(fmt.Sprintf("(%d)", len(m.selected)))
	}
	separatorWidth := 36
	if m.width > 0 {
		separatorWidth = m.width - lipgloss.Width(summary) - 1
	}
	if separatorWidth < 8 {
		separatorWidth = 8
	}
	return truncateToWidth(summary+" "+faintStyle.Render(strings.Repeat("─", separatorWidth)), m.width)
}

func searchLine(m pickerModel) string {
	return m.input.View()
}

func (m *pickerModel) refreshMatches() {
	query := strings.TrimSpace(m.input.Value())
	m.matches = m.matches[:0]

	if query == "" {
		for index := range m.candidates {
			m.matches = append(m.matches, index)
		}
	} else {
		searchKeys := make([]string, 0, len(m.candidates))
		for _, candidate := range m.candidates {
			searchKeys = append(searchKeys, candidateSearchKey(candidate))
		}
		for _, result := range fuzzy.Find(query, searchKeys) {
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

func candidateSearchKey(candidate repo.RepoCandidate) string {
	if candidate.FullName != "" {
		return candidate.FullName
	}
	if candidate.Name != "" {
		return candidate.Name
	}
	return candidate.DisplayName()
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

func (m pickerModel) visibleRows() int {
	if m.height > 0 {
		available := m.height - 3
		if available < 1 {
			return 1
		}
		if available < maxVisibleResults && available < len(m.matches) {
			return available
		}
	}
	if len(m.matches) < maxVisibleResults {
		return len(m.matches)
	}
	return maxVisibleResults
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

func clearRenderedPicker(view string) {
	lines := strings.Count(strings.TrimRight(view, "\n"), "\n") + 1
	if strings.TrimSpace(view) == "" {
		return
	}

	for i := 0; i < lines; i++ {
		fmt.Print("\r\033[2K")
		if i < lines-1 {
			fmt.Print("\033[1A")
		}
	}
	fmt.Print("\r")
}

func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return s
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width <= 1 {
		return "…"
	}

	var builder strings.Builder
	current := 0
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError && size == 1 {
			break
		}
		if current+1 >= width {
			break
		}
		builder.WriteRune(r)
		current++
		s = s[size:]
	}
	builder.WriteRune('…')
	return builder.String()
}
