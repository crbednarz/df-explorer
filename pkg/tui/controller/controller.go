package controller

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/crbednarz/df-explorer/pkg/tui/message"
)

type Model struct {
	explorer *explorer.Explorer
}

func New(explorer *explorer.Explorer) *Model {
	return &Model{
		explorer: explorer,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case message.RebuildRequest:
		return func() tea.Msg {
			m.explorer.Rebuild(context.TODO())
			return nil
		}
	}
	return nil
}

func (m *Model) View() string {
	return ""
}
