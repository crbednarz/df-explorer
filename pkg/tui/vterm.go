package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	vterm "github.com/crbednarz/df-explorer/pkg/vterm"
)

type vtermPanel struct {
	term *vterm.VTerm
}

func (vt *vtermPanel) Write(data []byte) (int, error) {
	return vt.term.Write(data)
}

func (vt *vtermPanel) Init() tea.Cmd {
	return nil
}

func (vt *vtermPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
