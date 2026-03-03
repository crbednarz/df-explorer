package elements

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Element interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}
