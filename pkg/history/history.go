package history

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

func New() *History {
	return &History{}
}

func (h *History) Add(raw string) {
	h.Entries = append(h.Entries, HistoryEntry{
		Command: raw,
	})
}
