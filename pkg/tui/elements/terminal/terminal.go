package terminal

import (
	"fmt"
	"io"

	tea "charm.land/bubbletea/v2"
	"github.com/crbednarz/df-explorer/pkg/explorer"
	"github.com/crbednarz/df-explorer/pkg/tui/message"
	vterm "github.com/crbednarz/df-explorer/pkg/vterm"
	"github.com/muesli/cancelreader"
)

type Model struct {
	vterm            *vterm.VTerm
	explorer         *explorer.Explorer
	attachmentReader cancelreader.CancelReader
}

func New(explorer *explorer.Explorer) *Model {
	vterm := vterm.New(80, 20)
	return &Model{
		vterm:    vterm,
		explorer: explorer,
	}
}

func (m *Model) Write(data []byte) (int, error) {
	return m.vterm.Write(data)
}

func (m *Model) Init() tea.Cmd {
	attachmentReader, err := cancelreader.NewReader(m.explorer.ContainerProxy())
	if err != nil {
		return func() tea.Msg {
			return message.FatalError{Err: fmt.Errorf("error creating cancelable reader for container attachment: %w", err)}
		}
	}
	m.attachmentReader = attachmentReader
	m.vterm.SetWriteCallback(func(data []byte) {
		m.explorer.ContainerProxy().Write(data)
	})
	return func() tea.Msg {
		_, err := io.Copy(m.vterm, attachmentReader)
		if err != nil {
			return message.FatalError{Err: fmt.Errorf("error reading from container attachment: %w", err)}
		}
		return nil
	}
}

func (m *Model) Update(message tea.Msg) (*Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyPressMsg:
		m.vterm.WriteKey(teaKeyToVtermKey(msg.Key()))
	}
	return m, nil
}

func (m *Model) View() string {
	contents, err := m.vterm.Contents()
	if err != nil {
		return fmt.Sprintf("Error retrieving contents: %v", err)
	}
	return contents
}

func (m *Model) SetSize(width int, height int) {
	m.vterm.SetSize(width, height)
	fmt.Fprintf(m.explorer.ContainerProxy(), "\x1b[8;%d;%dt", height, width)
}

func (m *Model) Close() error {
	m.attachmentReader.Cancel()
	return m.vterm.Close()
}

func (m *Model) Width() int {
	width, _ := m.vterm.GetSize()
	return width
}

func (m *Model) Height() int {
	_, height := m.vterm.GetSize()
	return height
}

func teaKeyToVtermKey(teaKey tea.Key) vterm.Key {
	var vtermMod vterm.KeyModifier
	switch teaKey.Mod {
	case tea.ModShift:
		vtermMod = vterm.KeyModShift
	case tea.ModAlt:
		vtermMod = vterm.KeyModAlt
	case tea.ModCtrl:
		vtermMod = vterm.KeyModCtrl
	}

	code := uint(teaKey.Code)
	if teaKey.Text != "" {
		code = uint(teaKey.Text[0])
	}

	return vterm.Key{
		IsUnichar: true,
		Modifier:  vtermMod,
		Code:      code,
	}
}
