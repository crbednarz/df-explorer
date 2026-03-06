package elements

import (
	tea "charm.land/bubbletea/v2"
)

type Element interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}
