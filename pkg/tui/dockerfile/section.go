package dockerfile

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/moby/buildkit/client/llb"
)

type BuildStatus int

const (
	StatusPending BuildStatus = iota
	StatusInProgress
	StatusCompleted
)

type sectionItem struct {
	Text     string
	Metadata *llb.OpMetadata
	Vertex   string
	Status   BuildStatus
}

func (s *sectionItem) FilterValue() string {
	return s.Text
}

func (s *sectionItem) Title() string {
	return s.Text
}

func (s *sectionItem) Description() string {
	return ""
}

type sectionDelegate struct{}

func (d sectionDelegate) Height() int {
	return 1
}

func (d sectionDelegate) Spacing() int {
	return 0
}

func (d sectionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d sectionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	block, ok := listItem.(*sectionItem)
	if !ok {
		return
	}

	str := block.Text

	prefix := " "
	style := itemStyle

	if index == m.Index() {
		style = selectedItemStyle
	} else if block.Metadata == nil {
		style = noStageStyle
	} else {
		switch block.Status {
		case StatusPending:
			style = pendingStyle
		case StatusInProgress:
			animationTime := (time.Now().UnixMilli() / 250) % 4
			animations := []string{"-", "\\", "|", "/"}
			prefix = animations[animationTime]
			style = inProgressStyle
		case StatusCompleted:
			style = completedStyle
		}
	}

	_, _ = fmt.Fprint(w, style.Render(fmt.Sprintf("%s %s", prefix, str)))
}
