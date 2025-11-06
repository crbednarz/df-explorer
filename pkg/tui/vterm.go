package tui

import (
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	vterm "github.com/crbednarz/df-explorer/pkg/vterm"
	"github.com/muesli/cancelreader"
)

type vtermPanel struct {
	term             *vterm.VTerm
	attachment       io.ReadWriter
	attachmentReader cancelreader.CancelReader
}

func newVTermPanel(attachment io.ReadWriter) *vtermPanel {
	vterm := vterm.New(80, 20)
	return &vtermPanel{
		term:       vterm,
		attachment: attachment,
	}
}

func (vt *vtermPanel) Write(data []byte) (int, error) {
	return vt.term.Write(data)
}

func (vt *vtermPanel) Init() tea.Cmd {
	attachmentReader, err := cancelreader.NewReader(vt.attachment)
	if err != nil {
		return func() tea.Msg {
			return FatalErrorMsg{Err: fmt.Errorf("error creating cancelable reader for container attachment: %w", err)}
		}
	}
	vt.attachmentReader = attachmentReader
	vt.term.SetWriteCallback(func(data []byte) {
		vt.attachment.Write(data)
	})
	return func() tea.Msg {
		_, err := io.Copy(vt.term, attachmentReader)
		if err != nil {
			return FatalErrorMsg{Err: fmt.Errorf("error reading from container attachment: %w", err)}
		}
		return nil
	}
}

func (vt *vtermPanel) Update(message tea.Msg) (*vtermPanel, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		vt.handleKeyMsg(msg)
	}
	return vt, nil
}

func (vt *vtermPanel) View() string {
	contents, err := vt.term.Contents()
	if err != nil {
		return fmt.Sprintf("Error retrieving contents: %v", err)
	}
	return contents
}

func (vt *vtermPanel) SetSize(width int, height int) {
	vt.term.SetSize(width, height)
}

func (vt *vtermPanel) Close() error {
	vt.attachmentReader.Cancel()
	return vt.term.Close()
}

func (vt *vtermPanel) Width() int {
	width, _ := vt.term.GetSize()
	return width
}

func (vt *vtermPanel) Height() int {
	_, height := vt.term.GetSize()
	return height
}

func (vt *vtermPanel) handleKeyMsg(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyRunes:
		for _, rune := range msg.Runes {
			vt.term.WriteKey(int(rune))
		}
	}
}
