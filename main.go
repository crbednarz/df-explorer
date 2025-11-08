package main

import (
	"context"
	"fmt"
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
	defer func() {
		err := cli.Close()
		if err != nil {
			_ = fmt.Errorf("unable to close docker client: %v", err)
		}
	}()

	log.Println("Creating TUI application...")
	app, err := tui.NewApp(ctx, cli)
	if err != nil {
		log.Fatalf("unable to create UI: %v", err)
	}
	defer func() {
		err := app.Close()
		if err != nil {
			_ = fmt.Errorf("unable to close UI: %v", err)
		}
	}()

	log.Println("Running TUI application...")
	err = app.Run(ctx)
	if err != nil {
		log.Fatalf("failed to run UI: %v", err)
	}
}
