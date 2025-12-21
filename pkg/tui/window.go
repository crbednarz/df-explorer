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

type windowModel struct {
	main         *dockerfile.Model
	terminal     *vtermPanel
	controller   *controller
	vtermFocused bool
}

func newWindow(explorer *explorer.Explorer) *windowModel {
	return &windowModel{
		main:         dockerfile.New(),
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
		m.main.Init(),
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
				m.main, mainCmd = m.main.Update(msg)
				return window, tea.Batch(windowCmd, mainCmd)
			}
		}
	}
	window, windowCmd := m.updateSelf(msg)
	m.main, mainCmd = m.main.Update(msg)
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
		m.terminal.SetSize(msg.Width, 10)
		m.main.SetSize(msg.Width, msg.Height-10)
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
	mainView := m.main.View()
	return lipgloss.JoinVertical(lipgloss.Left, mainView, vtermView)
}
