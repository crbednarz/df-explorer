package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/crbednarz/df-explorer/pkg/explorer"
)

type windowModel struct {
	main         *dockerfileView
	terminal     *vtermPanel
	vtermFocused bool
}

func newWindow(explorer *explorer.Explorer) *windowModel {
	return &windowModel{
		main:     newDockerfileView(),
		terminal: newVTermPanel(explorer.Attachment()),
	}
}

type (
	frameMsg struct{}
	focusMsg struct {
		VTermFocused bool
	}
)

func animate() tea.Cmd {
	return tea.Tick(time.Second/60.0, func(_ time.Time) tea.Msg {
		return frameMsg{}
	})
}

func (m *windowModel) Init() tea.Cmd {
	return tea.Batch(m.main.Init(), m.terminal.Init(), animate())
}

func (m *windowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var mainCmd tea.Cmd
	var terminalCmd tea.Cmd

	window, windowCmd := m.updateSelf(msg)
	m.main, mainCmd = m.main.Update(msg)
	m.terminal, terminalCmd = m.terminal.Update(msg)

	return window, tea.Batch(windowCmd, mainCmd, terminalCmd)
}

func (m *windowModel) updateSelf(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+j":
			return m, func() tea.Msg { return focusMsg{VTermFocused: true} }
		case "ctrl+k":
			return m, func() tea.Msg { return focusMsg{VTermFocused: false} }
		}
	case tea.WindowSizeMsg:
		m.terminal.SetSize(msg.Width, 10)
		m.main.SetSize(msg.Width, msg.Height-10)
	case frameMsg:
		return m, animate()
	case FatalErrorMsg:
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
