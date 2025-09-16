package tui

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/docker/docker/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/term"
)

type App struct {
	program  *tea.Program
	model    *model
	inputTee io.Reader
}

type model struct {
	vterm           *VTermPanel
	history         *HistoryPanel
	explorer        *explorer.Explorer
	historyViewport viewport.Model
	commandChannel  chan commandMsg
}

func NewApp(ctx context.Context, cli *client.Client) *App {
	vterm := NewVTerm()

	inputTeeReader, inputTeeWriter := io.Pipe()
	inputReader := io.TeeReader(os.Stdin, inputTeeWriter)

	explorer, err := explorer.New(ctx, cli)
	if err != nil {
		log.Fatalf("unable to create dockerfile explorer: %v", err)
	}

	model := &model{
		vterm:           vterm,
		history:         newHistoryPanel(),
		explorer:        explorer,
		historyViewport: viewport.New(80, 40),
		commandChannel:  make(chan commandMsg, 100),
	}

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithInput(inputReader))

	return &App{
		program: p,
		model:   model,
		// TODO: stdin should be selective routed based on focus, not duplicated.
		inputTee: inputTeeReader,
	}
}

func (app *App) Run(ctx context.Context) error {
	// Switch stdin to raw mode so that individual inputs can be processed by the container.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("unable to set terminal to raw mode: %w", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	eg, ctx := errgroup.WithContext(ctx)
	// TODO: This copy needs to be canacellable (there's also no error handling)
	eg.Go(func() error {
		_, err := io.Copy(app.model.explorer.Attachment(), app.inputTee)
		if err != nil {
			return fmt.Errorf("error during stdin write to container: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		_, err := io.Copy(app.model.vterm, app.model.explorer.Attachment())
		if err != nil {
			return fmt.Errorf("error during stdout read from container: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		err := app.model.explorer.Run(ctx, func(event explorer.ServerEvent) error {
			// TODO: Explorer really shouldn't be emitting ServerEvent
			app.model.commandChannel <- commandMsg{}
			return nil
		})
		if err != nil {
			return fmt.Errorf("runtime explorer error: %w", err)
		}
		return nil
	})

	// TODO: eg.Wait should be uesd here with some sort of cancel mechanism
	_, err = app.program.Run()
	if err != nil {
		return fmt.Errorf("runtime tea error: %w", err)
	}
	return nil
}

func (app *App) Close() error {
	if err := app.model.explorer.Close(); err != nil {
		return err
	}

	return app.model.vterm.Close()
}

type (
	frameMsg   struct{}
	commandMsg struct{}
)

func animate() tea.Cmd {
	return tea.Tick(time.Second/60.0, func(_ time.Time) tea.Msg {
		return frameMsg{}
	})
}

func waitForCommand(commandChannel chan commandMsg) tea.Cmd {
	return func() tea.Msg {
		return <-commandChannel
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.historyViewport.Init(), animate(), waitForCommand(m.commandChannel))
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			m.historyViewport, cmd = m.historyViewport.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.vterm.SetSize(msg.Width, 10)
		m.historyViewport.Width = msg.Width
		m.historyViewport.Height = msg.Height - m.vterm.Height()
		var cmd tea.Cmd
		m.historyViewport, cmd = m.historyViewport.Update(msg)
		return m, cmd

	case commandMsg:
		m.history.Set(m.explorer.History())
		return m, waitForCommand(m.commandChannel)
	case frameMsg:
		return m, animate()
	}

	return m, nil
}

func (m model) View() string {
	contents := m.vterm.View()
	m.historyViewport.SetContent(m.history.View())
	return lipgloss.JoinVertical(lipgloss.Left, m.historyViewport.View(), contents)
}
