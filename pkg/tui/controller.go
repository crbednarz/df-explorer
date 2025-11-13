package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/crbednarz/df-explorer/pkg/explorer"
)

type controller struct {
	explorer *explorer.Explorer
}

func newController(explorer *explorer.Explorer) *controller {
	return &controller{
		explorer: explorer,
	}
}

func (c *controller) Init() tea.Cmd {
	return nil
}

func (c *controller) Update(message tea.Msg) (*controller, tea.Cmd) {
	switch message.(type) {
	case RebuildRequestMsg:
		return c, func() tea.Msg {
			c.explorer.Rebuild(context.TODO())
			return nil
		}
	}
	return c, nil
}

func (c *controller) View() string {
	return ""
}
