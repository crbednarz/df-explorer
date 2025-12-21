package dockerfile

import (
	"github.com/moby/buildkit/client/llb"
)

type BuildStatus int

const (
	StatusPending BuildStatus = iota
	StatusInProgress
	StatusCompleted
)

type sourceOp struct {
	Text     string
	Metadata *llb.OpMetadata
	Vertex   string
	Status   BuildStatus
}

func (s sourceOp) FilterValue() string { return s.Text }
func (s sourceOp) Title() string       { return s.Text }
func (s sourceOp) Description() string { return "" }
