package main

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	vterm "github.com/mattn/go-libvterm"
	"golang.org/x/term"
)

func loadImage(ctx context.Context, cli *client.Client, imageName string) error {
	reader, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	io.Copy(io.Discard, reader)
	reader.Close()

	return nil
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

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	err = loadImage(ctx, cli, "docker.io/library/alpine")
	if err != nil {
		panic(err)
	}

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

	vt := vterm.New(25, 80)
	defer vt.Close()

	vt.SetUTF8(true)

	screen := vt.ObtainScreen()
	screen.Reset(true)

	go io.Copy(vt, attachment.Reader)
	screen.OnDamage = func(rect *vterm.Rect) int {
		runes := make([]rune, 10000)
		updated := screen.GetChars(&runes, rect)
		os.Stdout.Write([]byte(string(runes[0:updated])))
		return 1
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	go func() { _, _ = io.Copy(attachment.Conn, os.Stdin) }()
	err = waitForContainerExit(ctx, cli, resp.ID)
	if err != nil {
		panic(err)
	}
}
