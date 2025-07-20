package tui

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

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
	vterm    *VTermModel
	explorer *explorer.Explorer
}

func NewApp(e *explorer.Explorer) *App {
	vterm := NewVTerm()
	model := &model{
		vterm:    vterm,
		explorer: e,
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
		err := app.model.explorer.Listen()
		if err != nil {
			log.Fatalf("error while listening to explorer commands: %v", err)
		}
	}()

	_, err = app.program.Run()
	return err
}

func (app *App) Close() error {
	return app.model.vterm.Close()
}

type frameMsg struct{}

func animate() tea.Cmd {
	return tea.Tick(time.Second/60.0, func(_ time.Time) tea.Msg {
		return frameMsg{}
	})
}

func (m model) Init() tea.Cmd {
	return animate()
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.vterm.SetSize(msg.Width, msg.Height-2)
		return m, nil
	case frameMsg:
		return m, animate()
	}

	return m, nil
}

func (m model) View() string {
	contents, err := m.vterm.Contents()
	if err != nil {
		log.Fatalf("error during rendering: %v", err)
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.explorer.Status(), contents)
}
