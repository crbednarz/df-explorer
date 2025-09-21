package tui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	vterm "github.com/crbednarz/df-explorer/pkg/vterm"
	"github.com/docker/docker/client"
	"github.com/muesli/cancelreader"
	"golang.org/x/sync/errgroup"
	"golang.org/x/term"
)

type App struct {
	explorer *explorer.Explorer
	vterm    *vterm.VTerm
	window   *windowModel
	program  *tea.Program
}

func NewApp(ctx context.Context, cli *client.Client) (*App, error) {
	explorer, err := explorer.New(ctx, cli)
	if err != nil {
		return nil, fmt.Errorf("unable to create dockerfile explorer: %v", err)
	}

	vterm := vterm.New(80, 20)
	return &App{
		explorer: explorer,
		vterm:    vterm,
		window: &windowModel{
			main: &dockerfileView{},
			terminal: &vtermPanel{
				term: vterm,
			},
		},
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

	attachmentReader, err := cancelreader.NewReader(a.explorer.Attachment())
	if err != nil {
		return fmt.Errorf("bubbletea: error creating cancel reader for container attachment: %w", err)
	}

	inputReader := io.TeeReader(stdinReader, a.explorer.Attachment())
	p := tea.NewProgram(
		a.window,
		tea.WithAltScreen(),
		tea.WithInput(inputReader),
		tea.WithFPS(60),
	)

	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)

	eg.Go(func() error {
		_, err := io.Copy(a.vterm, attachmentReader)
		if err != nil {
			return fmt.Errorf("error during stdout read from container: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		err := a.explorer.Run(ctx, func(event explorer.ServerEvent) error {
			return nil
		})
		if err != nil {
			return fmt.Errorf("runtime explorer error: %w", err)
		}
		return nil
	})

	_, err = p.Run()
	if err != nil {
		return fmt.Errorf("runtime tea error: %w", err)
	}
	stdinReader.Cancel()
	attachmentReader.Cancel()
	cancel()

	err = eg.Wait()
	if errors.Is(err, context.Canceled) || errors.Is(err, cancelreader.ErrCanceled) {
		return nil
	}
	return err
}

func (a *App) Close() error {
	err := a.vterm.Close()
	if err != nil {
		return err
	}
	return a.explorer.Close()
}
