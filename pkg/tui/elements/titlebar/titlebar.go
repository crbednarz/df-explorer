package titlebar

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/crbednarz/df-explorer/pkg/tui/style"
)

type Model struct {
	fileName string
	path     string
	style    lipgloss.Style
	width    int
}

func New(theme *style.Theme) *Model {
	return &Model{
		style: lipgloss.NewStyle().
			Background(theme.PanelColor).
			Padding(0, 1),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case explorer.DockerfileEvent:
		m.fileName = msg.Dockerfile.FileName()
		m.path = msg.Dockerfile.Dir()
	}

	return nil
}

func (m *Model) View() string {
	pathText := "Path: " + m.path
	pathWidth := lipgloss.Width(pathText)
	nameView := m.style.Width(m.width - pathWidth - 2).Render(m.fileName)
	pathView := m.style.Render(pathText)

	return lipgloss.JoinHorizontal(lipgloss.Top, nameView, pathView)
}

func (m *Model) SetWidth(width int) {
	m.width = width
}
