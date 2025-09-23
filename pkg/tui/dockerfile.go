package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/crbednarz/df-explorer/pkg/explorer"
)

type dockerfileView struct {
	dockerfile *docker.Dockerfile
	blockList  list.Model
	viewport   viewport.Model
}

type sourceBlock struct {
	Text string
}

func newDockerfileView() *dockerfileView {
	return &dockerfileView{
		viewport:  viewport.New(80, 40),
		blockList: list.New(nil, list.NewDefaultDelegate(), 80, 40),
	}
}

func (df *dockerfileView) SetSize(width int, height int) {
	df.viewport.Width = width
	df.viewport.Height = height
	df.blockList.SetSize(width, height)
}

func (df *dockerfileView) Init() tea.Cmd {
	return df.viewport.Init()
}

func (df *dockerfileView) Update(msg tea.Msg) (*dockerfileView, tea.Cmd) {
	switch msg := msg.(type) {
	case explorer.DockerfileEvent:
		df.setDockerfile(msg.Dockerfile)
	}
	var viewportCmd, listCmd tea.Cmd
	df.viewport, viewportCmd = df.viewport.Update(msg)
	df.blockList, listCmd = df.blockList.Update(msg)
	return df, tea.Batch(viewportCmd, listCmd)
}

func (df *dockerfileView) View() string {
	if df.dockerfile == nil {
		return ""
	}

	df.viewport.SetContent(df.blockList.View())
	return df.viewport.View()
}

func (df *dockerfileView) setDockerfile(dockerfile *docker.Dockerfile) {
	chunks := dockerfile.Source().Chunks

	sourceBlocks := make([]list.Item, len(chunks))

	for i, chunk := range chunks {
		sourceBlocks[i] = &sourceBlock{Text: chunk.Text}
	}

	df.blockList.SetItems(sourceBlocks)
	df.dockerfile = dockerfile
}

func (s sourceBlock) FilterValue() string { return s.Text }
func (s sourceBlock) Title() string       { return s.Text }
func (s sourceBlock) Description() string { return "" }
