package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/crbednarz/df-explorer/pkg/docker"
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

	container, err := docker.NewContainer(ctx, cli, "alpine:latest")
	if err != nil {
		log.Fatalf("unable to create docker container: %v", err)
	}
	defer container.Close()

	terminal, err := container.Attach(ctx)
	if err != nil {
		log.Fatalf("unable to attach to container: %v", err)
	}
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	go io.Copy(terminal, os.Stdin)
	go io.Copy(os.Stdout, terminal)
	go io.Copy(os.Stderr, terminal)

	err = container.WaitForExit(ctx, true)
	if err != nil {
		log.Fatalf("error waiting for container exit: %v", err)
	}
}
