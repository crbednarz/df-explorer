package sourceview

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/crbednarz/df-explorer/pkg/tui/elements/statusbar"
	"github.com/crbednarz/df-explorer/pkg/tui/message"
	"github.com/crbednarz/df-explorer/pkg/tui/style"
)

type Model struct {
	dockerfile  *docker.Dockerfile
	sectionList list.Model
	status      *statusbar.Model
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
		keys: sourceViewKeyMap{
			Rebuild: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "rebuild"),
			),
		},
		status: statusbar.New(theme),
	}
	m.sectionList.SetShowStatusBar(false)
	m.sectionList.SetFilteringEnabled(false)
	m.sectionList.SetShowHelp(false)
	m.sectionList.SetShowTitle(false)
	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
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
			return func() tea.Msg {
				return message.RebuildRequest{}
			}
		}
	}
	var listCmd tea.Cmd
	m.sectionList, listCmd = m.sectionList.Update(msg)
	statusCmd := m.status.Update(msg)
	return tea.Batch(listCmd, statusCmd)
}

func (m *Model) View() string {
	if m.dockerfile == nil {
		return ""
	}

	listView := m.sectionList.View()
	statusView := m.status.View()

	view := lipgloss.JoinVertical(lipgloss.Left, statusView, listView)
	return view
}

func (m *Model) SetSize(width int, height int) {
	m.sectionList.SetSize(width, height-1)
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
