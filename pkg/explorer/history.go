package explorer

import "fmt"

type History struct {
	Entries []HistoryEntry
}

type HistoryEntry struct {
	Command string
}

func (h *History) Add(event ServerEvent) {
	if event.State == CommandStateSuccess {
		h.Entries = append(h.Entries, HistoryEntry{
			Command: fmt.Sprintf("%s: %s", string(event.Operation), event.Command),
		})
	}
}
