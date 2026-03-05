package window

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/crbednarz/df-explorer/pkg/tui/controller"
	"github.com/crbednarz/df-explorer/pkg/tui/elements/sourceview"
	"github.com/crbednarz/df-explorer/pkg/tui/elements/terminal"
	"github.com/crbednarz/df-explorer/pkg/tui/message"
	"github.com/crbednarz/df-explorer/pkg/tui/style"
)

var (
	panelActiveStyle   = lipgloss.NewStyle().Margin(0, 0).Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("4"))
	panelInactiveStyle = lipgloss.NewStyle().Margin(0, 0).Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("8"))
)

type Model struct {
	file         *sourceview.Model
	term         *terminal.Model
	controller   *controller.Model
	vtermFocused bool
}

func New(explorer *explorer.Explorer) *Model {
	theme := style.DefaultTheme()
	return &Model{
		file:         sourceview.New(theme),
		term:         terminal.New(explorer),
		controller:   controller.New(explorer),
		vtermFocused: true,
	}
}

type (
	frameMsg struct{}
)

func animate() tea.Cmd {
	return tea.Tick(time.Second/60.0, func(_ time.Time) tea.Msg {
		return frameMsg{}
	})
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.file.Init(),
		m.term.Init(),
		m.controller.Init(),
		animate(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var mainCmd tea.Cmd
	var terminalCmd tea.Cmd
	var controllerCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+j":
			m.vtermFocused = !m.vtermFocused
			return m, nil
		case "ctrl+k":
			m.vtermFocused = !m.vtermFocused
			return m, nil
		default:
			window, windowCmd := m.updateSelf(msg)
			if m.vtermFocused {
				m.term, terminalCmd = m.term.Update(msg)
				return window, tea.Batch(windowCmd, terminalCmd)
			} else {
				m.file, mainCmd = m.file.Update(msg)
				return window, tea.Batch(windowCmd, mainCmd)
			}
		}
	}
	window, windowCmd := m.updateSelf(msg)
	m.file, mainCmd = m.file.Update(msg)
	m.term, terminalCmd = m.term.Update(msg)
	m.controller, controllerCmd = m.controller.Update(msg)

	return window, tea.Batch(windowCmd, mainCmd, terminalCmd, controllerCmd)
}

func (m *Model) updateSelf(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.term.SetSize(msg.Width-1, 10)
		m.file.SetSize(msg.Width-1, msg.Height-10)
	case frameMsg:
		return m, animate()
	case message.FatalError:
		fmt.Println("Fatal error:", msg.Err)
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) View() string {
	vtermView := m.term.View()
	fileView := m.file.View()

	if m.vtermFocused {
		vtermView = panelActiveStyle.Render(vtermView)
		fileView = panelInactiveStyle.Render(fileView)
	} else {
		vtermView = panelInactiveStyle.Render(vtermView)
		fileView = panelActiveStyle.Render(fileView)
	}
	return lipgloss.JoinVertical(lipgloss.Left, fileView, vtermView)
}
