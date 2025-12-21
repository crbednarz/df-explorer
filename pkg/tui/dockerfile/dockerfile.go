package dockerfile

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/crbednarz/df-explorer/pkg/tui/message"
)

type Model struct {
	dockerfile *docker.Dockerfile
	blockList  list.Model
	viewport   viewport.Model
	chunkMap   map[string]*sourceOp
	keys       dockerfileViewKeyMap
}

type dockerfileViewKeyMap struct {
	Rebuild key.Binding
}

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(1)
	selectedItemStyle = itemStyle.Foreground(lipgloss.Color("170"))
	pendingStyle      = itemStyle.Foreground(lipgloss.Color("#ffff00"))
	inProgressStyle   = itemStyle.Foreground(lipgloss.Color("#ff0000"))
	completedStyle    = itemStyle.Foreground(lipgloss.Color("#00ff00"))
	noStageStyle      = itemStyle.Foreground(lipgloss.Color("240"))
)

func New() *Model {
	d := itemDelegate{}
	m := &Model{
		blockList: list.New(nil, d, 80, 40),
		viewport:  viewport.New(80, 40),
		keys: dockerfileViewKeyMap{
			Rebuild: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "rebuild"),
			),
		},
	}
	m.blockList.SetShowTitle(false)
	m.blockList.SetShowStatusBar(false)
	m.blockList.SetFilteringEnabled(false)
	m.blockList.SetShowHelp(false)
	return m
}

func (m *Model) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case explorer.DockerfileEvent:
		m.setDockerfile(msg.Dockerfile)
	case explorer.BuildStartEvent:
		for _, chunk := range m.chunkMap {
			chunk.Status = StatusPending
		}
	case explorer.BuildProgressEvent:
		m.handleProgress(msg)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Rebuild):
			return m, func() tea.Msg {
				return message.RebuildRequest{}
			}
		}
	}
	var viewportCmd, listCmd tea.Cmd
	m.viewport, viewportCmd = m.viewport.Update(msg)
	m.blockList, listCmd = m.blockList.Update(msg)
	return m, tea.Batch(viewportCmd, listCmd)
}

func (m *Model) View() string {
	if m.dockerfile == nil {
		return ""
	}

	m.viewport.SetContent(m.blockList.View())
	return m.viewport.View()
}

func (m *Model) SetSize(width int, height int) {
	m.viewport.Width = width
	m.viewport.Height = height

	m.blockList.SetSize(width, height)
	m.blockList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keys.Rebuild,
		}
	}
}

func (m *Model) setDockerfile(dockerfile *docker.Dockerfile) {
	chunks := dockerfile.Source().Chunks

	sourceBlocks := make([]list.Item, len(chunks))
	chunkMap := make(map[string]*sourceOp)

	for i, chunk := range chunks {
		chunk := &sourceOp{
			Text:     chunk.Text,
			Metadata: chunk.Metadata,
			Vertex:   chunk.VertexHash,
		}
		sourceBlocks[i] = chunk
		if chunk.Metadata != nil {
			chunkMap[chunk.Vertex] = chunk
		}
	}

	m.blockList.SetItems(sourceBlocks)
	m.chunkMap = chunkMap
	m.dockerfile = dockerfile
}

func (m *Model) handleProgress(event explorer.BuildProgressEvent) {
	status := event.Status
	for _, vertex := range status.Vertexes {
		chunk, ok := m.chunkMap[string(vertex.Digest)]
		if ok {
			if vertex.Completed != nil {
				chunk.Status = StatusCompleted
			} else if vertex.Started != nil {
				chunk.Status = StatusInProgress
			} else {
				chunk.Status = StatusPending
			}
		}
	}

	for _, s := range status.Statuses {
		chunk, ok := m.chunkMap[string(s.Vertex)]
		if ok {
			if s.Completed != nil {
				chunk.Status = StatusCompleted
			} else if s.Started != nil {
				chunk.Status = StatusInProgress
			} else {
				chunk.Status = StatusPending
			}
		}
	}
}
