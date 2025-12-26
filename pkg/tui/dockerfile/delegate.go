package dockerfile

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type itemDelegate struct{}

func (d itemDelegate) Height() int {
	return 1
}

func (d itemDelegate) Spacing() int {
	return 0
}

func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	block, ok := listItem.(*sourceOp)
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
