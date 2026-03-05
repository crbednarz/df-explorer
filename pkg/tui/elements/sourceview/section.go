package sourceview

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/crbednarz/df-explorer/pkg/tui/style"
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

type sectionDelegate struct {
	itemStyle         lipgloss.Style
	selectedItemStyle lipgloss.Style
	pendingStyle      lipgloss.Style
	inProgressStyle   lipgloss.Style
	completedStyle    lipgloss.Style
	noStageStyle      lipgloss.Style
}

func newSectionDelegate(theme *style.Theme) *sectionDelegate {
	baseStyle := lipgloss.NewStyle().PaddingLeft(1).Background(theme.BackgroundColor)
	return &sectionDelegate{
		itemStyle:         baseStyle,
		selectedItemStyle: baseStyle.Foreground(theme.AccentColor),
		pendingStyle:      baseStyle,
		inProgressStyle:   baseStyle,
		completedStyle:    baseStyle,
		noStageStyle:      baseStyle,
	}
}

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
	style := d.itemStyle

	if index == m.Index() {
		style = d.selectedItemStyle
	} else if block.Metadata == nil {
		style = d.noStageStyle
	} else {
		switch block.Status {
		case StatusPending:
			style = d.pendingStyle
		case StatusInProgress:
			animationTime := (time.Now().UnixMilli() / 250) % 4
			animations := []string{"-", "\\", "|", "/"}
			prefix = animations[animationTime]
			style = d.inProgressStyle
		case StatusCompleted:
			style = d.completedStyle
		}
	}

	_, _ = fmt.Fprint(w, style.Render(fmt.Sprintf("%s %s", prefix, str)))
}
