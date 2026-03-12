package docker

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

type Container interface {
	Attachment() io.ReadWriter
	SetSize(width uint, height uint) error
	Close() error
	ID() string
}

type DockerContainer struct {
	cli           *client.Client
	imageName     string
	containerID   string
	attachment    io.ReadWriteCloser
	removeOnClean bool
}

type ContainerTerminal struct {
	IO io.ReadWriteCloser
}

type containerOptions struct {
	mounts          []mount.Mount
	shouldAttach    bool
	name            string
	securityOptions []string
	entryPoint      []string
	command         []string
	shouldPull      bool
	shouldRemove    bool
	shouldReuse     bool
}

func newContainerOptions() *containerOptions {
	return &containerOptions{
		shouldRemove: true,
	}
}

type ContainerOption func(*containerOptions)

func WithMount(localPath string, containerPath string) ContainerOption {
	return func(options *containerOptions) {
		options.mounts = append(options.mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   localPath,
			Target:   containerPath,
			ReadOnly: false,
		})
	}
}

func WithAttach(isAttached bool) ContainerOption {
	return func(options *containerOptions) {
		options.shouldAttach = isAttached
	}
}

func WithName(name string) ContainerOption {
	return func(options *containerOptions) {
		options.name = name
	}
}

func WithSecurityOption(option string) ContainerOption {
	return func(options *containerOptions) {
		options.securityOptions = append(options.securityOptions, option)
	}
}

func WithEntryPoint(entryPoint []string) ContainerOption {
	return func(options *containerOptions) {
		options.entryPoint = entryPoint
	}
}

func WithCommand(command []string) ContainerOption {
	return func(options *containerOptions) {
		options.command = command
	}
}

func WithPull() ContainerOption {
	return func(options *containerOptions) {
		options.shouldPull = true
	}
}

func WithRemoveOnClean(shouldRemove bool) ContainerOption {
	return func(options *containerOptions) {
		options.shouldRemove = shouldRemove
	}
}

func WithReuse(shouldReuse bool) ContainerOption {
	return func(options *containerOptions) {
		options.shouldReuse = shouldReuse
	}
}

func NewContainer(ctx context.Context, cli *client.Client, image string, optionFuncs ...ContainerOption) (*DockerContainer, error) {
	options := newContainerOptions()
	for _, fn := range optionFuncs {
		fn(options)
	}

	if options.shouldPull {
		err := pullImage(ctx, cli, image)
		if err != nil {
			return nil, fmt.Errorf("failed to pull image: %w", err)
		}
	}

	container := &DockerContainer{
		cli:           cli,
		imageName:     image,
		containerID:   "",
		removeOnClean: options.shouldRemove,
	}
	err := container.run(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to run container: %w", err)
	}

	if options.shouldAttach {
		attachment, err := container.attach(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to attach to container: %w", err)
		}
		container.attachment = attachment
	}

	return container, nil
}

func (c *DockerContainer) run(ctx context.Context, options *containerOptions) error {
	if c.containerID != "" {
		return fmt.Errorf("container is already running with ID: %s", c.containerID)
	}

	if options.name != "" {
		// We can ignore the error here, as we'll just create a new container if needed
		existing, _ := c.findContainerIDByName(ctx, string(options.name))
		if existing != "" {
			if options.shouldReuse {
				c.containerID = existing
			} else {
				err := c.cli.ContainerRemove(ctx, existing, container.RemoveOptions{Force: true})
				if err != nil {
					return fmt.Errorf("failed to remove existing container with name %s: %w", options.name, err)
				}
			}
		}
	}

	if c.containerID == "" {
		resp, err := c.cli.ContainerCreate(
			ctx,
			&container.Config{
				Image:      c.imageName,
				Cmd:        options.command,
				Entrypoint: options.entryPoint,
				Tty:        true,
				OpenStdin:  true,
			},
			&container.HostConfig{
				Mounts:      options.mounts,
				SecurityOpt: options.securityOptions,
			},
			nil, nil, string(options.name))
		if err != nil {
			return err
		}
		c.containerID = resp.ID
	}

	err := c.cli.ContainerStart(ctx, c.containerID, container.StartOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *DockerContainer) Attachment() io.ReadWriter {
	return c.attachment
}

func (c *DockerContainer) findContainerIDByName(ctx context.Context, name string) (string, error) {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	searchName := "/" + name
	for _, existingContainer := range containers {
		if slices.Contains(existingContainer.Names, searchName) {
			return existingContainer.ID, nil
		}
	}
	return "", nil
}

func (c *DockerContainer) attach(ctx context.Context) (io.ReadWriteCloser, error) {
	if c.containerID == "" {
		return nil, fmt.Errorf("container is not running")
	}

	attachment, err := c.cli.ContainerAttach(ctx, c.containerID, container.AttachOptions{
		Stderr: true,
		Stdout: true,
		Stdin:  true,
		Stream: true,
	})
	if err != nil {
		return nil, err
	}

	return attachment.Conn, nil
}

func (c *DockerContainer) SetSize(width uint, height uint) error {
	return c.cli.ContainerResize(context.TODO(), c.containerID, container.ResizeOptions{
		Height: height,
		Width:  width,
	})
}

func (c *DockerContainer) ID() string {
	return c.containerID
}

func (c *DockerContainer) Close() error {
	if c.containerID == "" {
		return nil
	}
	if c.attachment != nil {
		err := c.attachment.Close()
		if err != nil {
			return err
		}
	}

	// TODO: Investigate why graceful shutdown doesn't work
	timeout := 0
	err := c.cli.ContainerStop(context.TODO(), c.containerID, container.StopOptions{Timeout: &timeout})
	if !c.removeOnClean {
		return err
	}
	return c.cli.ContainerRemove(context.TODO(), c.containerID, container.RemoveOptions{Force: true})
}

func pullImage(ctx context.Context, cli *client.Client, imageName string) error {
	reader, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	io.Copy(io.Discard, reader)
	reader.Close()

	return nil
}
