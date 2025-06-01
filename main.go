package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	reader, err := cli.ImagePull(ctx, "docker.io/library/alpine", image.PullOptions{})
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	io.Copy(io.Discard, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:     "alpine",
		Cmd:       []string{"/bin/sh"},
		Tty:       true,
		OpenStdin: true,
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}
	defer cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		panic(err)
	}

	println("Attaching...")

	attachment, err := cli.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stderr: true,
		Stdout: true,
		Stdin:  true,
		Stream: true,
	})
	if err != nil {
		panic(err)
	}

	go io.Copy(os.Stdout, attachment.Reader)
	go io.Copy(os.Stderr, attachment.Reader)

	inout := make(chan []byte)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inout <- []byte(scanner.Text())
		}
	}()

	// Write to docker container
	go func(w io.WriteCloser) {
		for {
			data, ok := <-inout
			// log.Println("Received to send to docker", string(data))
			if !ok {
				fmt.Println("!ok")
				w.Close()
				return
			}

			w.Write(append(data, '\n'))
		}
	}(attachment.Conn)

	err = waitForContainerExit(ctx, cli, resp.ID)
	if err != nil {
		panic(err)
	}
}

func waitForContainerExit(ctx context.Context, cli *client.Client, containerId string) error {
	statusCh, errCh := cli.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return err
	case <-statusCh:
	}
	return nil
}
