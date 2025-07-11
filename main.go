package main

import (
	"context"
	"log"

	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/crbednarz/df-explorer/pkg/view"
	"github.com/docker/docker/client"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("unable to initialize docker client: %v", err)
	}
	defer cli.Close()

	explorer, err := explorer.New(ctx, cli, "./Dockerfile")
	if err != nil {
		log.Fatalf("unable to create dockerfile explorer: %v", err)
	}
	defer explorer.Close()

	app := view.NewApp(explorer)
	defer app.Close()

	err = app.Run(ctx)
	if err != nil {
		log.Fatalf("failed to run UI: %v", err)
	}
}
