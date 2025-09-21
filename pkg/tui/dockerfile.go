package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/crbednarz/df-explorer/pkg/docker"
)

type dockerfileView struct {
	dockerfile *docker.Dockerfile
}

func (df *dockerfileView) Init() tea.Cmd {
	return nil
}

func (df *dockerfileView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return df, nil
}

func (df *dockerfileView) View() string {
	return ""
}
