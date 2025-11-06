package tui

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/docker/docker/client"
	"github.com/muesli/cancelreader"
	"golang.org/x/sync/errgroup"
	"golang.org/x/term"
)

type App struct {
	explorer *explorer.Explorer
	window   *windowModel
}

func NewApp(ctx context.Context, cli *client.Client) (*App, error) {
	e, err := explorer.New(ctx, cli)
	if err != nil {
		return nil, fmt.Errorf("unable to create dockerfile explorer: %v", err)
	}

	return &App{
		explorer: e,
		window:   newWindow(e),
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("unable to set terminal to raw mode: %w", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	stdinReader, err := cancelreader.NewReader(os.Stdin)
	if err != nil {
		return fmt.Errorf("bubbletea: error creating cancel reader for stdin: %w", err)
	}

	p := tea.NewProgram(
		a.window,
		tea.WithAltScreen(),
		tea.WithInput(stdinReader),
		tea.WithFPS(60),
	)

	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)

	eg.Go(func() error {
		err := a.explorer.Run(ctx, func(event explorer.Event) error {
			p.Send(event)
			return nil
		})
		if err != nil {
			return fmt.Errorf("runtime explorer error: %w", err)
		}
		return nil
	})

	_, err = p.Run()
	stdinReader.Cancel()
	cancel()
	if err != nil {
		return fmt.Errorf("runtime tea error: %w", err)
	}

	err = eg.Wait()
	if errors.Is(err, context.Canceled) || errors.Is(err, cancelreader.ErrCanceled) {
		return nil
	}
	return err
}

func (a *App) Close() error {
	return a.explorer.Close()
}
