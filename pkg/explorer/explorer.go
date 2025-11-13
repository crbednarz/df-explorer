package explorer

import (
	"context"
	"fmt"

	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/docker/docker/client"
	buildkit "github.com/moby/buildkit/client"
	"golang.org/x/sync/errgroup"
)

type EventCallback func(event Event) error

type Status int

const (
	StatusIdle Status = iota
	StatusBuilding
)

type Explorer struct {
	cli           *client.Client
	server        *Server
	history       History
	dockerfile    *docker.Dockerfile
	builder       *docker.Builder
	container     ContainerProxy
	eventCallback EventCallback
	status        Status
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
		status:     StatusIdle,
	}

	return e, nil
}

func (e *Explorer) Run(ctx context.Context, callback EventCallback) error {
	e.eventCallback = callback
	err := e.eventCallback(DockerfileEvent{
		Dockerfile: e.dockerfile,
	})
	if err != nil {
		return err
	}

	e.Rebuild(ctx)

	// TODO: this should really check if ImageID is ""
	container, err := e.server.SpawnContainer(ctx, e.cli, e.dockerfile.ImageID())
	if err != nil {
		return fmt.Errorf("unable to spawn container: %w", err)
	}
	defer e.container.Close()
	e.container.SetContainer(container)

	return e.server.Listen(ctx, func(event ServerEvent) error {
		e.history.Add(event)
		if string(event.Operation) != "" {
			e.dockerfile.Append(fmt.Sprintf("%s %s", event.Operation, event.Command))
			err := callback(DockerfileEvent{
				Dockerfile: e.dockerfile,
			})
			if err != nil {
				return err
			}
		}
		return callback(&CommandEvent{
			Command:    event.Command,
			Operation:  event.Operation,
			State:      event.State,
			ReturnCode: event.ReturnCode,
		})
	})
}

func (e *Explorer) Rebuild(ctx context.Context) error {
	if e.status == StatusBuilding {
		return fmt.Errorf("build already in progress")
	}
	e.status = StatusBuilding
	defer func() { e.status = StatusIdle }()

	err := e.eventCallback(BuildStartEvent{})
	if err != nil {
		return err
	}
	progress := make(chan *buildkit.SolveStatus)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		for status := range progress {
			err := e.eventCallback(BuildProgressEvent{
				Status: status,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	eg.Go(func() error {
		_, err := e.dockerfile.Build(ctx, e.builder, progress)
		return err
	})

	return eg.Wait()
}

func (e *Explorer) Close() error {
	if err := e.builder.Close(); err != nil {
		return err
	}
	return e.server.Close()
}

func (e *Explorer) ContainerProxy() *ContainerProxy {
	return &e.container
}
