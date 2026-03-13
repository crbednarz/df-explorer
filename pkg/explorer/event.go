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

type ContainerChangeEvent struct {
	Container docker.Container
}

type BuildProgressEvent struct {
	Status *buildkit.SolveStatus
}

type BuildStartEvent struct{}

type BuildEndEvent struct {
	Error error
}

// ContainerID returns ID of new container, if present, otherwise returning an empty string
func (c *ContainerChangeEvent) ContainerID() string {
	if c.Container != nil {
		return c.Container.ID()
	}
	return ""
}
