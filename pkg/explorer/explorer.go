package explorer

import (
	"context"
	"fmt"

	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/crbednarz/df-explorer/pkg/history"
	"github.com/docker/docker/client"
)

type Explorer struct {
	cli        *client.Client
	server     *Server
	dockerfile string
	history    *history.History
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
		history:    history.New(),
	}

	return e, nil
}

func (e *Explorer) SpawnContainer(ctx context.Context) (*docker.Container, error) {
	// TODO: Use provided dockerfile instead of hardcoded image
	return e.server.SpawnContainer(ctx, e.cli, "ubuntu:latest")
}

func (e *Explorer) Run() error {
	return e.server.Listen(func(command Command) error {
		e.history.Add(command.Command)
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
	return e.server.Close()
}
