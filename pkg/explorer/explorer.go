package explorer

import (
	"context"
	"fmt"
	"io"

	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/docker/docker/client"
)

type Explorer struct {
	cli        *client.Client
	server     *Server
	history    History
	dockerfile *docker.Dockerfile
	builder    *docker.Builder
	attachment dynamicIO
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
		attachment: dynamicIO{},
	}

	return e, nil
}

func (e *Explorer) History() []HistoryEntry {
	return e.history.Entries
}

func (e *Explorer) Attachment() io.ReadWriter {
	return &e.attachment
}

func (e *Explorer) Run(ctx context.Context, callback CommandCallback) error {
	image, err := e.dockerfile.Build(ctx, e.builder)
	if err != nil {
		return fmt.Errorf("unable to build docker image: %w", err)
	}
	fmt.Println("Image built successfully:", image)

	container, err := e.server.SpawnContainer(ctx, e.cli, image)
	if err != nil {
		return fmt.Errorf("unable to spawn container: %w", err)
	}
	defer container.Close()
	e.attachment.SetReaderWriter(container.Attachment(), container.Attachment())

	return e.server.Listen(func(event ServerEvent) error {
		e.history.Add(event)
		if string(event.Operation) != "" {
			e.dockerfile.Append(fmt.Sprintf("%s %s", event.Operation, event.Command))
		}
		callback(event)
		return nil
	})
}

func (e *Explorer) Close() error {
	if err := e.builder.Close(); err != nil {
		return err
	}
	return e.server.Close()
}
