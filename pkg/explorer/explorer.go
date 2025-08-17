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
	history    *History
	dockerfile string
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
		history:    newHistory(),
	}

	return e, nil
}

func (e *Explorer) History() []HistoryEntry {
	return e.history.Entries
}

func (e *Explorer) SpawnContainer(ctx context.Context) (*docker.Container, error) {
	// TODO: Use provided dockerfile instead of hardcoded image
	return e.server.SpawnContainer(ctx, e.cli, "ubuntu:latest")
}

func (e *Explorer) Run(callback CommandCallback) error {
	return e.server.Listen(func(cmd Command) error {
		e.history.Add(cmd.Command)
		callback(cmd)
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
