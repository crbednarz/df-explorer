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
	history    History
	dockerfile *docker.Dockerfile
	builder    *docker.Builder
}

func New(ctx context.Context, cli *client.Client) (*Explorer, error) {
	server, err := newServer()
	if err != nil {
		return nil, fmt.Errorf("unable to create server: %w", err)
	}

	builder, err := docker.NewBuilder(ctx, cli)
	if err != nil {
		return nil, fmt.Errorf("unable to create docker builder: %w", err)
	}

	// TODO: Load dockerfile path from arg
	dockerfile, err := docker.NewDockerfile(".", "Dockerfile")
	if err != nil {
		return nil, fmt.Errorf("unable to construct dockerfile: %w", err)
	}
	e := &Explorer{
		cli:        cli,
		server:     server,
		dockerfile: dockerfile,
		builder:    builder,
	}

	return e, nil
}

func (e *Explorer) History() []HistoryEntry {
	return e.history.Entries
}

func (e *Explorer) SpawnContainer(ctx context.Context) (*docker.Container, error) {
	image, err := e.dockerfile.Build(ctx, e.builder)
	if err != nil {
		return nil, fmt.Errorf("unable to build docker image: %w", err)
	}
	fmt.Println("Image built successfully:", image)
	return e.server.SpawnContainer(ctx, e.cli, image)
}

func (e *Explorer) Run(callback CommandCallback) error {
	return e.server.Listen(func(event ServerEvent) error {
		e.history.Add(event)
		if string(event.Operation) != "" {
			e.dockerfile.Append(fmt.Sprintf("%s %s", event.Operation, event.Command))
		}
		callback(event)
		return nil
	})
}

func (e *Explorer) Rebuild(commands []Command) error {
	// TODO: Impelement this
	return fmt.Errorf("rebuild not implemented")
}

func (e *Explorer) Redeploy() error {
	// TODO: Impelement this
	return fmt.Errorf("redeploy not implemented")
}

func (e *Explorer) Snapshot() error {
	// TODO: Impelement this
	return fmt.Errorf("snapshot not implemented")
}

func (e *Explorer) Close() error {
	// TODO: Report multiple errors
	e.builder.Close()
	return e.server.Close()
}
