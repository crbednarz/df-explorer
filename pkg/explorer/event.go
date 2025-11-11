package explorer

import (
	"github.com/crbednarz/df-explorer/pkg/docker"
	buildkit "github.com/moby/buildkit/client"
)

type Event interface{}

type CommandEvent struct {
	Command    string
	Operation  OperationType
	State      CommandState
	ReturnCode int
}

type DockerfileEvent struct {
	Dockerfile *docker.Dockerfile
}

type BuildProgressEvent struct {
	Status *buildkit.SolveStatus
}

type BuildStartEvent struct{}
