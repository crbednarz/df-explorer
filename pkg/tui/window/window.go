package window

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	source       *sourceview.Model
	term         *terminal.Model
	controller   *controller.Model
	vtermFocused bool
}

func New(explorer *explorer.Explorer) *Model {
	theme := style.DefaultTheme()
	return &Model{
		source:       sourceview.New(theme),
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
		m.source.Init(),
		m.term.Init(),
		m.controller.Init(),
		animate(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				terminalCmd := m.term.Update(msg)
				return window, tea.Batch(windowCmd, terminalCmd)
			} else {
				sourceViewCmd := m.source.Update(msg)
				return window, tea.Batch(windowCmd, sourceViewCmd)
			}
		}
	}
	window, windowCmd := m.updateSelf(msg)
	sourceViewCmd := m.source.Update(msg)
	terminalCmd := m.term.Update(msg)
	controllerCmd := m.controller.Update(msg)

	return window, tea.Batch(windowCmd, sourceViewCmd, terminalCmd, controllerCmd)
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
		m.source.SetSize(msg.Width-1, msg.Height-10)
	case frameMsg:
		return m, animate()
	case message.FatalError:
		fmt.Println("Fatal error:", msg.Err)
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) View() tea.View {
	vtermView := m.term.View()
	sourceView := m.source.View()

	if m.vtermFocused {
		vtermView = panelActiveStyle.Render(vtermView)
		sourceView = panelInactiveStyle.Render(sourceView)
	} else {
		vtermView = panelInactiveStyle.Render(vtermView)
		sourceView = panelActiveStyle.Render(sourceView)
	}
	view := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, sourceView, vtermView))
	view.AltScreen = true
	return view
}
