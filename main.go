package main

import (
	"context"
	"log"
	"os"

	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/crbednarz/df-explorer/pkg/view"
	"github.com/docker/docker/client"
	"golang.org/x/term"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("unable to initialize docker client: %v", err)
	}
	defer cli.Close()

	container, err := docker.NewContainer(ctx, cli, "ubuntu:latest")
	if err != nil {
		log.Fatalf("unable to create docker container: %v", err)
	}
	defer container.Close()

	attachment, err := container.Attach(ctx)
	if err != nil {
		log.Fatalf("unable to attach to container: %v", err)
	}
	defer attachment.Close()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	app := view.NewApp(attachment)
	defer app.Close()

	err = app.Run()
	if err != nil {
		log.Fatalf("failed to run UI: %v", err)
	}
}
