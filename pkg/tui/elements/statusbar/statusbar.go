package statusbar

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/crbednarz/df-explorer/pkg/tui/style"
)

type BuildStatus int

const (
	StatusPending BuildStatus = iota
	StatusInProgress
	StatusCompleted
)

type Model struct {
	name        string
	status      BuildStatus
	containerID string
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
	case explorer.BuildStartEvent:
		m.status = StatusPending
	case explorer.BuildProgressEvent:
		m.handleProgress(msg)
	case explorer.BuildEndEvent:
		m.status = StatusCompleted
	case explorer.ContainerChangeEvent:
		m.containerID = msg.ContainerID()

	}

	return nil
}

func (m *Model) View() string {
	var status string
	switch m.status {
	case StatusCompleted:
		status = "Completed"
	case StatusInProgress:
		status = "In Progress"
	case StatusPending:
		status = "Pending"
	}
	return fmt.Sprintf("File: %s - Status: %s - Container: %s", m.name, status, m.containerID)
}

func (m *Model) handleProgress(event explorer.BuildProgressEvent) {
	status := event.Status
	total := 0
	completed := 0

	for _, s := range status.Statuses {
		total++
		if s.Completed != nil {
			completed++
		}
	}

	if total == completed {
		m.status = StatusCompleted
	} else {
		m.status = StatusInProgress
	}
}
