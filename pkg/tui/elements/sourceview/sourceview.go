package sourceview

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/crbednarz/df-explorer/pkg/tui/message"
	"github.com/crbednarz/df-explorer/pkg/tui/style"
)

type Model struct {
	dockerfile  *docker.Dockerfile
	sectionList list.Model
	viewport    viewport.Model
	sectionMap  map[string]*sectionItem
	keys        sourceViewKeyMap
}

type sourceViewKeyMap struct {
	Rebuild key.Binding
}

func New(theme *style.Theme) *Model {
	d := newSectionDelegate(theme)
	m := &Model{
		sectionList: list.New(nil, d, 80, 40),
		viewport:    viewport.New(80, 40),
		keys: sourceViewKeyMap{
			Rebuild: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "rebuild"),
			),
		},
	}
	m.sectionList.SetShowStatusBar(false)
	m.sectionList.SetFilteringEnabled(false)
	m.sectionList.SetShowHelp(false)
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
		for _, section := range m.sectionMap {
			section.Status = StatusPending
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
	m.sectionList, listCmd = m.sectionList.Update(msg)
	return m, tea.Batch(viewportCmd, listCmd)
}

func (m *Model) View() string {
	if m.dockerfile == nil {
		return ""
	}

	m.viewport.SetContent(m.sectionList.View())
	output := m.viewport.View()
	return output
}

func (m *Model) SetSize(width int, height int) {
	x, y := style.PanelBorder.GetFrameSize()
	m.viewport.Width = width - x
	m.viewport.Height = height - y

	m.sectionList.SetSize(m.viewport.Width, m.viewport.Height)
	m.sectionList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keys.Rebuild,
		}
	}
}

func (m *Model) setDockerfile(dockerfile *docker.Dockerfile) {
	sections := dockerfile.Source().Sections

	sectionItemMap := make(map[string]*sectionItem)
	sectionItems := make([]list.Item, len(sections))

	for i, section := range sections {
		item := &sectionItem{
			Text:     section.Text,
			Metadata: section.Metadata,
			Vertex:   section.VertexHash,
		}

		sectionItems[i] = item
		if item.Metadata != nil {
			sectionItemMap[item.Vertex] = item
		}
	}

	m.sectionList.SetItems(sectionItems)
	m.sectionList.Title = titleFromDockerfile(dockerfile)
	m.sectionMap = sectionItemMap
	m.dockerfile = dockerfile
}

func (m *Model) handleProgress(event explorer.BuildProgressEvent) {
	status := event.Status
	for _, vertex := range status.Vertexes {
		section, ok := m.sectionMap[string(vertex.Digest)]
		if ok {
			if vertex.Completed != nil {
				section.Status = StatusCompleted
			} else if vertex.Started != nil {
				section.Status = StatusInProgress
			} else {
				section.Status = StatusPending
			}
		}
	}

	for _, s := range status.Statuses {
		section, ok := m.sectionMap[string(s.Vertex)]
		if ok {
			if s.Completed != nil {
				section.Status = StatusCompleted
			} else if s.Started != nil {
				section.Status = StatusInProgress
			} else {
				section.Status = StatusPending
			}
		}
	}
}

func titleFromDockerfile(dockerfile *docker.Dockerfile) string {
	return dockerfile.FileName()
}
