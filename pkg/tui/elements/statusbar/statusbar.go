package statusbar

import (
	tea "charm.land/bubbletea/v2"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/crbednarz/df-explorer/pkg/tui/style"
)

type Model struct {
	name string
}

func New(theme *style.Theme) *Model {
	return &Model{}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case explorer.DockerfileEvent:
		m.name = msg.Dockerfile.FileName()
	}

	return nil
}

func (m *Model) View() string {
	return m.name
}
