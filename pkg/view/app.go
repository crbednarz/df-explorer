package view

import (
	"io"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	program *tea.Program
	model   *model
}

type model struct {
	vterm *VTermModel
}

func NewApp(containerAttachment io.ReadWriter) *App {
	vterm := NewVTerm()
	model := &model{
		vterm: vterm,
	}
	teaInputReader, teaInputWriter := io.Pipe()
	go io.Copy(io.MultiWriter(containerAttachment, teaInputWriter), os.Stdin)
	go io.Copy(vterm, containerAttachment)

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithInput(teaInputReader))

	return &App{
		program: p,
		model:   model,
	}
}

func (app *App) Run() error {
	_, err := app.program.Run()
	return err
}

func (app *App) Close() error {
	return app.model.vterm.Close()
}

type frameMsg struct{}

func animate() tea.Cmd {
	return tea.Tick(time.Second/30.0, func(_ time.Time) tea.Msg {
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
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.vterm.SetSize(msg.Width, msg.Height)
		return m, nil
	case frameMsg:
		return m, animate()
	}

	return m, nil
}

func (m model) View() string {
	return m.vterm.Contents()
}
