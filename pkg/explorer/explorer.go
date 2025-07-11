package explorer

import (
	"context"
	"fmt"

	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/docker/docker/client"
)

type Explorer struct {
	cli        *client.Client
	server     *Server
	dockerfile string
	status     string
}

func New(ctx context.Context, cli *client.Client, dockerfile string) (*Explorer, error) {
	server, err := newServer()
	if err != nil {
		return nil, fmt.Errorf("unable to create server: %w", err)
	}
	e := &Explorer{
		cli:        cli,
		server:     server,
		dockerfile: dockerfile,
	}

	return e, nil
}

func (e *Explorer) SpawnContainer(ctx context.Context) (*docker.Container, error) {
	// TODO: Use provided dockerfile instead of hardcoded image
	return e.server.SpawnContainer(ctx, e.cli, "ubuntu:latest")
}

func (e *Explorer) Listen() error {
	return e.server.Listen(func(command Command) error {
		e.status = command.Command
		return nil
	})
}

func (e *Explorer) Status() string {
	return e.status
}

func (e *Explorer) Close() error {
	return e.server.Close()
}
