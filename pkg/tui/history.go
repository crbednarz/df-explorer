package tui

import (
	"strings"

	"github.com/crbednarz/df-explorer/pkg/explorer"
)

type HistoryPanel struct {
	history []explorer.HistoryEntry
}

func newHistoryPanel() *HistoryPanel {
	return &HistoryPanel{}
}

func (h *HistoryPanel) Set(history []explorer.HistoryEntry) {
	h.history = history
}

func (h *HistoryPanel) View() string {
	if len(h.history) == 0 {
		return ""
	}

	start := max(0, len(h.history)-10)
	end := len(h.history)

	var output strings.Builder
	for i := start; i < end; i++ {
		output.WriteString(h.history[i].Command + "\n")
	}
	return output.String()
}
