package view

import (
	"fmt"
	"os"
	"strings"

	"github.com/hinshun/vt10x"
)

type VTermModel struct {
	term vt10x.Terminal
}

func NewVTerm() *VTermModel {
	vt := vt10x.New(vt10x.WithWriter(os.Stdin), vt10x.WithSize(80, 20))
	vt.Cursor()
	return &VTermModel{
		term: vt,
	}
}

func (vt *VTermModel) Write(p []byte) (n int, err error) {
	return vt.term.Write(p)
}

func (vt *VTermModel) Contents() string {
	// Useful reference on ANSI escape codes:
	// https://stackoverflow.com/a/33206814
	var b strings.Builder
	lastFG, lastBG := vt10x.DefaultFG, vt10x.DefaultBG
	width, height := vt.term.Size()
	cursor := vt.term.Cursor()
	b.WriteString("\x1b[0m")

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := vt.term.Cell(x, y)
			fg := cell.FG
			bg := cell.BG

			if x == cursor.X && y == cursor.Y {
				// Swap FG and BG for the cursor position
				fmt.Fprint(&b, "\x1b[7m")
			}

			if fg != lastFG || bg != lastBG {
				fmt.Fprint(&b, "\x1b[")
				//
				if fg == vt10x.DefaultFG {
					fmt.Fprint(&b, "39;")
				} else {
					fmt.Fprintf(&b, "38;5;%d", fg)
				}
				if bg == vt10x.DefaultBG {
					fmt.Fprintf(&b, "49;")
				} else {
					fmt.Fprintf(&b, "48;5;%d", bg)
				}
				fmt.Fprint(&b, "m")
				lastFG, lastBG = fg, bg
			}
			b.WriteRune(cell.Char)
			if x == cursor.X && y == cursor.Y {
				// Swap FG and BG back after the cursor position
				fmt.Fprint(&b, "\x1b[27m")
			}
		}
		b.WriteString("\x1b[0m\n")
		lastFG, lastBG = 0, 0
	}
	return b.String()
}

func (vt *VTermModel) SetSize(width int, height int) {
	vt.term.Resize(width, height)
}

func (vt *VTermModel) Close() error {
	return nil
}
