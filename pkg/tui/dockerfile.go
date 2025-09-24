package tui

import (
	"fmt"
	"io"
	"strings"

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
}

type sourceChunkItem struct {
	Text     string
	Metadata *llb.OpMetadata
}

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(1)
	debugStyle        = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("241"))
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("170"))
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
		sourceBlocks[i] = &sourceChunkItem{
			Text:     chunk.Text,
			Metadata: chunk.Metadata,
		}
	}

	df.blockList.SetItems(sourceBlocks)
	df.dockerfile = dockerfile
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

	fn := func(s ...string) string {
		return itemStyle.Render("  " + strings.Join(s, " "))
	}
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	} else if block.Metadata != nil {
		fn = func(s ...string) string {
			return debugStyle.Render("# " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
