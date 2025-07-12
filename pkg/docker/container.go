package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

type Container struct {
	cli         *client.Client
	imageName   string
	containerId string
}

type ContainerTerminal struct {
	IO io.ReadWriteCloser
}

type containerOptions struct {
	mountOption
}

type ContainerOption interface {
	apply(*containerOptions)
}

type mountOption struct {
	mounts []mount.Mount
}

func (m *mountOption) apply(options *containerOptions) {
	options.mounts = append(options.mounts, m.mounts...)
}

func WithMount(localPath string, containerPath string) ContainerOption {
	return &mountOption{
		mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   localPath,
				Target:   containerPath,
				ReadOnly: false,
			},
		},
	}
}

func NewContainer(ctx context.Context, cli *client.Client, image string, optionFuncs ...ContainerOption) (*Container, error) {
	options := containerOptions{}
	for _, fn := range optionFuncs {
		fn.apply(&options)
	}

	err := pullImage(ctx, cli, image)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	container := &Container{
		cli:         cli,
		imageName:   image,
		containerId: "",
	}
	err = container.run(ctx, &options)
	if err != nil {
		return nil, fmt.Errorf("failed to run container: %w", err)
	}

	return container, nil
}

func (d *Container) run(ctx context.Context, options *containerOptions) error {
	if d.containerId != "" {
		return fmt.Errorf("container is already running with ID: %s", d.containerId)
	}
	resp, err := d.cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:     d.imageName,
			Cmd:       []string{"/bin/bash"},
			Tty:       true,
			OpenStdin: true,
		},
		&container.HostConfig{
			Mounts: options.mounts,
		},
		nil, nil, "")
	if err != nil {
		return err
	}
	d.containerId = resp.ID

	err = d.cli.ContainerStart(ctx, d.containerId, container.StartOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (d *Container) Attach(ctx context.Context) (io.ReadWriteCloser, error) {
	if d.containerId == "" {
		return nil, fmt.Errorf("container is not running")
	}

	attachment, err := d.cli.ContainerAttach(ctx, d.containerId, container.AttachOptions{
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

func (d *Container) WaitForExit(ctx context.Context, clean bool) error {
	statusCh, errCh := d.cli.ContainerWait(ctx, d.containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return err
	case <-statusCh:
	}

	if clean {
		err := d.cli.ContainerRemove(ctx, d.containerId, container.RemoveOptions{Force: true})
		if err != nil {
			return fmt.Errorf("failed to remove container: %w", err)
		}
	}
	return nil
}

func (d *Container) Close() error {
	if d.containerId == "" {
		return nil
	}
	return d.cli.ContainerRemove(context.TODO(), d.containerId, container.RemoveOptions{Force: true})
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

/*
func waitForContainerExit(ctx context.Context, cli *client.Client, containerId string) error {
	statusCh, errCh := cli.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return err
	case <-statusCh:
	}
	return nil
}*/
