package explorer

import "github.com/crbednarz/df-explorer/pkg/docker"

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
