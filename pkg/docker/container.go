package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type DockerContainer struct {
	cli         *client.Client
	imageName   string
	containerId string
}

type ContainerTerminal struct {
	IO io.ReadWriteCloser
}

func NewContainer(ctx context.Context, cli *client.Client, image string) (*DockerContainer, error) {
	container := &DockerContainer{
		cli:         cli,
		imageName:   image,
		containerId: "",
	}
	err := container.run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run container: %w", err)
	}

	return container, nil
}

func (d *DockerContainer) run(ctx context.Context) error {
	if d.containerId != "" {
		return fmt.Errorf("container is already running with ID: %s", d.containerId)
	}
	resp, err := d.cli.ContainerCreate(ctx, &container.Config{
		Image:     d.imageName,
		Cmd:       []string{"/bin/sh"},
		Tty:       true,
		OpenStdin: true,
	}, nil, nil, nil, "")
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

func (d *DockerContainer) Attach(ctx context.Context) (ContainerTerminal, error) {
	if d.containerId == "" {
		return ContainerTerminal{}, fmt.Errorf("container is not running")
	}

	attachment, err := d.cli.ContainerAttach(ctx, d.containerId, container.AttachOptions{
		Stderr: true,
		Stdout: true,
		Stdin:  true,
		Stream: true,
	})
	if err != nil {
		return ContainerTerminal{}, err
	}

	return ContainerTerminal{
		IO: attachment.Conn,
	}, nil
}

func (d *DockerContainer) WaitForExit(ctx context.Context) error {
	statusCh, errCh := d.cli.ContainerWait(ctx, d.containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return err
	case <-statusCh:
	}
	return nil
}

func (d *DockerContainer) Close() error {
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
