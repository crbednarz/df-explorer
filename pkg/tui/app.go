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
	"golang.org/x/term"
)

type App struct {
	program     *tea.Program
	model       *model
	inputWriter io.Writer
}

type model struct {
	vterm           *VTermPanel
	history         *HistoryPanel
	explorer        *explorer.Explorer
	historyViewport viewport.Model
	commandChannel  chan commandMsg
}

func NewApp(e *explorer.Explorer) *App {
	vterm := NewVTerm()
	model := &model{
		vterm:           vterm,
		history:         newHistoryPanel(),
		explorer:        e,
		historyViewport: viewport.New(80, 40),
		commandChannel:  make(chan commandMsg, 100),
	}

	teaInputReader, teaInputWriter := io.Pipe()
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithInput(teaInputReader))

	return &App{
		program:     p,
		model:       model,
		inputWriter: teaInputWriter,
	}
}

func (app *App) Run(ctx context.Context) error {
	container, err := app.model.explorer.SpawnContainer(ctx)
	if err != nil {
		return fmt.Errorf("unable to spawn container: %w", err)
	}
	defer container.Close()

	// Switch stdin to raw mode so that individual inputs can be processed by the container.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("unable to set terminal to raw mode: %w", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()
	go io.Copy(io.MultiWriter(container.Attachment(), app.inputWriter), os.Stdin)
	go io.Copy(app.model.vterm, container.Attachment())

	go func() {
		err := app.model.explorer.Run(func(event explorer.ServerEvent) error {
			app.model.commandChannel <- commandMsg{}
			return nil
		})
		if err != nil {
			log.Fatalf("error while running explorer: %v", err)
		}
	}()

	_, err = app.program.Run()
	return err
}

func (app *App) Close() error {
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
		case "esc", "ctrl+c":
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
