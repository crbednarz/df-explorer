package tui

import (
	"fmt"

	vterm "github.com/crbednarz/df-explorer/pkg/vterm"
)

type VTermPanel struct {
	term *vterm.VTerm
}

func NewVTerm() *VTermPanel {
	vt := vterm.New(20, 90)
	terminal := &VTermPanel{
		term: vt,
	}

	return terminal
}

func (vt *VTermPanel) Write(data []byte) (int, error) {
	return vt.term.Write(data)
}

func (vt *VTermPanel) View() string {
	contents, err := vt.term.Contents()
	if err != nil {
		return fmt.Sprintf("Error retrieving contents: %v", err)
	}
	return contents
}

func (vt *VTermPanel) SetSize(width int, height int) {
	vt.term.SetSize(width, height)
}

func (vt *VTermPanel) Close() error {
	return vt.term.Close()
}

func (vt *VTermPanel) Width() int {
	width, _ := vt.term.GetSize()
	return width
}

func (vt *VTermPanel) Height() int {
	_, height := vt.term.GetSize()
	return height
}
