package view

import (
	vterm "github.com/crbednarz/df-explorer/pkg/vterm"
)

type VTermModel struct {
	term *vterm.VTerm
}

func NewVTerm() *VTermModel {
	vt := vterm.New(20, 90)
	terminal := &VTermModel{
		term: vt,
	}

	return terminal
}

func (vt *VTermModel) Write(data []byte) (int, error) {
	return vt.term.Write(data)
}

func (vt *VTermModel) Contents() (string, error) {
	return vt.term.Contents()
}

func (vt *VTermModel) SetSize(width int, height int) {
	vt.term.SetSize(width, height)
}

func (vt *VTermModel) Close() error {
	return vt.term.Close()
}
