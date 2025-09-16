package main

import (
	"context"
	"log"

	"github.com/crbednarz/df-explorer/pkg/tui"
	"github.com/docker/docker/client"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("unable to initialize docker client: %v", err)
	}
	defer cli.Close()

	app := tui.NewApp(ctx, cli)
	defer app.Close()

	err = app.Run(ctx)
	if err != nil {
		log.Fatalf("failed to run UI: %v", err)
	}
}
