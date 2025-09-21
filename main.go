package main

import (
	"context"
	"log"

	"github.com/crbednarz/df-explorer/pkg/tui"
	"github.com/docker/docker/client"
)

func main() {
	log.Println("Starting df-explorer...")
	ctx := context.Background()
	log.Println("Initializing Docker client...")
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("unable to initialize docker client: %v", err)
	}
	defer cli.Close()

	log.Println("Creating TUI application...")
	app, err := tui.NewApp(ctx, cli)
	if err != nil {
		log.Fatalf("unable to create UI: %v", err)
	}
	defer app.Close()

	log.Println("Running TUI application...")
	err = app.Run(ctx)
	if err != nil {
		log.Fatalf("failed to run UI: %v", err)
	}
}
