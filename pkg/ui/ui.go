package ui

import (
	"git-graph/pkg/commit"
	"log"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	libgloss "github.com/charmbracelet/lipgloss"
)

var highlight_style libgloss.Style = libgloss.NewStyle().Foreground(libgloss.Color("229")).Background(libgloss.Color("57")).Bold(true)

type model struct {
	lines         []map[string]string
	current_hash  string
	jump          int
	cursor        int
	graph_width   int
	details_width int
	height        int
	details_view  viewport.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.details_width = msg.Width - m.graph_width - 10
		m.height = msg.Height - 1
		m.details_view = viewport.New(m.details_width, m.height)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "down", "j":
			if m.jump*(m.cursor+1) < len(m.lines) {
				m.cursor++
			} else {
				m.cursor = (len(m.lines) - 1) / m.jump
			}
			m.current_hash = m.lines[m.jump*m.cursor]["hash"]
		case "up", "k":
			if m.jump*(m.cursor-1) >= 0 {
				m.cursor--
			} else {
				m.cursor = 0
			}
			m.current_hash = m.lines[m.jump*m.cursor]["hash"]
		}
	}
	return m, nil
}

func updateGraphView(m *model) string {
	start_index := m.cursor * m.jump
	if start_index+m.height > len(m.lines) {
		start_index = max(len(m.lines)-m.height, 0)
	}
	view_height := m.height
	if view_height > len(m.lines) {
		view_height = len(m.lines)
	}

	var graph strings.Builder
	for i := range view_height {
		line := m.lines[start_index+i]
		if line["hash"] == m.current_hash {
			highlighted := highlight_style.Render(line["hash"])
			graph.WriteString(line["graph"] + " " + highlighted + " " + line["body"] + "\n")
		} else {
			graph.WriteString(line["graph"] + " " + line["hash"] + " " + line["body"] + "\n")
		}
	}
	return graph.String()
}

func (m model) View() string {
	graph_style := libgloss.NewStyle().
		Width(m.graph_width).
		BorderRight(true).
		BorderStyle(libgloss.ThickBorder()).
		BorderForeground(libgloss.Color("242"))

	details_style := libgloss.NewStyle().Width(m.details_width).MarginLeft(1)
	m.details_view.SetContent(getDetails(m.current_hash))

	return libgloss.JoinHorizontal(
		libgloss.Top,
		graph_style.Render(updateGraphView(&m)),
		details_style.Render(m.details_view.View()),
	)
}

func getDetails(hash string) string {
	return commit.GetCommitStats(hash)
}

func strLen(str string) int {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	s := ansiRegex.ReplaceAllString(str, "")
	return len(s)
}

func initModel(lines []map[string]string, jump int) model {
	width := strLen(lines[0]["graph"]) + len(lines[0]["hash"]) + len(lines[0]["body"]) + 1
	return model{
		lines:        lines,
		jump:         jump,
		current_hash: lines[0]["hash"],
		cursor:       0,
		graph_width:  width,
	}
}

func Run(lines []map[string]string, jump int) {
	m := initModel(lines, jump)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
