package explorer

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/docker/docker/client"
	buildkit "github.com/moby/buildkit/client"
	"golang.org/x/sync/errgroup"
)

// FIXME: Event callback should not return an error
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
	mu            sync.Mutex
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

	if err := e.Rebuild(ctx); err != nil {
		// FIXME: This shouldn't return an error. Containers are allowed to fail
		return fmt.Errorf("initial container build failed: %w", err)
	}

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
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.status == StatusBuilding {
		return fmt.Errorf("build already in progress")
	}
	e.status = StatusBuilding
	defer func() { e.status = StatusIdle }()

	e.eventCallback(BuildStartEvent{})
	progress := make(chan *buildkit.SolveStatus)
	eg, errGroupCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		for status := range progress {
			e.eventCallback(BuildProgressEvent{
				Status: status,
			})
		}
		e.eventCallback(BuildEndEvent{})
		return nil
	})
	eg.Go(func() error {
		_, err := e.dockerfile.Build(errGroupCtx, e.builder, progress)
		return err
	})

	err := eg.Wait()
	if err == nil {
		return e.setContainerFromImageID(ctx, e.dockerfile.ImageID())
	} else {
		log.Fatalf("err: %v", err)
	}
	return err
}

func (e *Explorer) setContainerFromImageID(ctx context.Context, imageID string) error {
	container, err := e.server.SpawnContainer(ctx, e.cli, imageID)
	if err != nil {
		return fmt.Errorf("unable to spawn container: %w", err)
	}
	e.container.SetContainer(container)
	e.eventCallback(ContainerChangeEvent{
		Container: container,
	})
	return nil
}

func (e *Explorer) Close() error {
	if err := e.builder.Close(); err != nil {
		return err
	}
	if err := e.container.Close(); err != nil {
		return err
	}
	return e.server.Close()
}

func (e *Explorer) ContainerProxy() *ContainerProxy {
	return &e.container
}
