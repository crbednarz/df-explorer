package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/moby/buildkit/client/llb"
)

type dockerfileView struct {
	dockerfile *docker.Dockerfile
	blockList  list.Model
	viewport   viewport.Model
	chunkMap   map[string]*sourceChunkItem
}

type BuildStatus int

const (
	StatusPending BuildStatus = iota
	StatusInProgress
	StatusCompleted
)

type sourceChunkItem struct {
	Text     string
	Metadata *llb.OpMetadata
	Vertex   string
	Status   BuildStatus
}

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(1)
	selectedItemStyle = itemStyle.Foreground(lipgloss.Color("170"))
	pendingStyle      = itemStyle.Foreground(lipgloss.Color("244"))
	inProgressStyle   = itemStyle.Foreground(lipgloss.Color("33"))
	completedStyle    = itemStyle.Foreground(lipgloss.Color("34"))
	noStageStyle      = itemStyle.Foreground(lipgloss.Color("240"))
)

func newDockerfileView() *dockerfileView {
	d := itemDelegate{}
	return &dockerfileView{
		viewport:  viewport.New(80, 40),
		blockList: list.New(nil, d, 80, 40),
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
	case explorer.BuildStartEvent:
		for _, chunk := range df.chunkMap {
			chunk.Status = StatusPending
		}
	case explorer.BuildProgressEvent:
		df.handleProgress(msg)
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
	chunkMap := make(map[string]*sourceChunkItem)

	for i, chunk := range chunks {
		chunk := &sourceChunkItem{
			Text:     chunk.Text,
			Metadata: chunk.Metadata,
			Vertex:   chunk.VertexHash,
		}
		sourceBlocks[i] = chunk
		if chunk.Metadata != nil {
			chunkMap[chunk.Vertex] = chunk
		}
	}

	df.blockList.SetItems(sourceBlocks)
	df.chunkMap = chunkMap
	df.dockerfile = dockerfile
}

func (df *dockerfileView) handleProgress(event explorer.BuildProgressEvent) {
	status := event.Status
	for _, vertex := range status.Vertexes {
		chunk, ok := df.chunkMap[string(vertex.Digest)]
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
}

func (s sourceChunkItem) FilterValue() string { return s.Text }
func (s sourceChunkItem) Title() string       { return s.Text }
func (s sourceChunkItem) Description() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	block, ok := listItem.(*sourceChunkItem)
	if !ok {
		return
	}

	str := block.Text

	prefix := " "
	style := itemStyle

	if index == m.Index() {
		prefix = ">"
		style = selectedItemStyle
	} else if block.Metadata == nil {
		style = noStageStyle
	} else {
		prefix = "#"
		switch block.Status {
		case StatusPending:
			style = pendingStyle
		case StatusInProgress:
			style = inProgressStyle
		case StatusCompleted:
			style = completedStyle
		}
	}

	fmt.Fprint(w, style.Render(fmt.Sprintf("%s %s", prefix, str)))
}
