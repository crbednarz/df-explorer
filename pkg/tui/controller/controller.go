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

func (c *Model) Init() tea.Cmd {
	return nil
}

func (c *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg.(type) {
	case message.RebuildRequest:
		return c, func() tea.Msg {
			c.explorer.Rebuild(context.TODO())
			return nil
		}
	}
	return c, nil
}

func (c *Model) View() string {
	return ""
}
