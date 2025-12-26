package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/crbednarz/df-explorer/pkg/tui/dockerfile"
	"github.com/crbednarz/df-explorer/pkg/tui/message"
)

var (
	panelActiveStyle   = lipgloss.NewStyle().Margin(0, 0).Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("4"))
	panelInactiveStyle = lipgloss.NewStyle().Margin(0, 0).Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("8"))
)

type windowModel struct {
	file         *dockerfile.Model
	terminal     *vtermPanel
	controller   *controller
	vtermFocused bool
}

func newWindow(explorer *explorer.Explorer) *windowModel {
	return &windowModel{
		file:         dockerfile.New(),
		terminal:     newVTermPanel(explorer),
		controller:   newController(explorer),
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

func (m *windowModel) Init() tea.Cmd {
	return tea.Batch(
		m.file.Init(),
		m.terminal.Init(),
		m.controller.Init(),
		animate(),
	)
}

func (m *windowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.terminal, terminalCmd = m.terminal.Update(msg)
				return window, tea.Batch(windowCmd, terminalCmd)
			} else {
				m.file, mainCmd = m.file.Update(msg)
				return window, tea.Batch(windowCmd, mainCmd)
			}
		}
	}
	window, windowCmd := m.updateSelf(msg)
	m.file, mainCmd = m.file.Update(msg)
	m.terminal, terminalCmd = m.terminal.Update(msg)
	m.controller, controllerCmd = m.controller.Update(msg)

	return window, tea.Batch(windowCmd, mainCmd, terminalCmd, controllerCmd)
}

func (m *windowModel) updateSelf(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.terminal.SetSize(msg.Width-1, 10)
		m.file.SetSize(msg.Width-1, msg.Height-10)
	case frameMsg:
		return m, animate()
	case message.FatalError:
		fmt.Println("Fatal error:", msg.Err)
		return m, tea.Quit
	}
	return m, nil
}

func (m *windowModel) View() string {
	vtermView := m.terminal.View()
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
