package explorer

type History struct {
	Entries []HistoryEntry
}

type EntryType int

const (
	EphemeralCommand EntryType = iota
	RunCommand
	EnvCommand
)

type EntryState int

const (
	StateRunning EntryState = iota
	StateComplete
)

type HistoryEntry struct {
	Command string
}

func newHistory() *History {
	return &History{
		Entries: make([]HistoryEntry, 0),
	}
}

func (h *History) Add(raw string) {
	h.Entries = append(h.Entries, HistoryEntry{
		Command: raw,
	})
}
