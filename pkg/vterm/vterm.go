package vterm

/*
#cgo CFLAGS: -I${SRCDIR}/../../libvterm/include
#cgo LDFLAGS: ${SRCDIR}/../../libvterm/.libs/libvterm.a
#include <vterm.h>

VTermValue vterm_value_from_int(int int_value) {
  VTermValue value = { .number = int_value };
	return value;
}

int vterm_get_cells(const VTermScreen* screen, VTermRect area, VTermScreenCell* cells, size_t capacity) {
  size_t writeIndex = 0;
  for (size_t y = area.start_row; y < area.end_row; y++) {
		for (size_t x = area.start_col; x < area.end_col; x++) {
			VTermPos pos = { .row = y, .col = x };
			if (vterm_screen_get_cell(screen, pos, &cells[writeIndex]) == 1) {
				writeIndex++;
				if (writeIndex == capacity) {
					return capacity;
				}
			}
		}
  }
	return writeIndex;
}
*/
import "C"

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

// libvterm lacks documentation, so the following references were used:
// - https://github.com/mattn/go-libvterm
// - https://github.com/neovim/neovim/blob/master/src/nvim/terminal.c

type VTerm struct {
	term   *C.VTerm
	screen *C.VTermScreen
	state  *C.VTermState
	mutex  sync.RWMutex
	buffer []C.VTermScreenCell
}

func New(width int, height int) *VTerm {
	term := C.vterm_new(C.int(height), C.int(width))
	C.vterm_set_utf8(term, 1)
	screen := C.vterm_obtain_screen(term)
	state := C.vterm_obtain_state(term)

	vt := &VTerm{
		term:   term,
		screen: screen,
		state:  state,
	}

	vt.setup()
	return vt
}

func (vt *VTerm) setup() {
	C.vterm_screen_enable_altscreen(vt.screen, C.int(1))
	C.vterm_screen_enable_reflow(vt.screen, C.bool(true))
	C.vterm_screen_reset(vt.screen, C.int(1))

	cursorShape := C.vterm_value_from_int(C.VTERM_PROP_CURSORSHAPE_BLOCK)
	C.vterm_state_set_termprop(vt.state, C.VTERM_PROP_CURSORSHAPE, &cursorShape)

	cursorBlink := C.vterm_value_from_int(1)
	C.vterm_state_set_termprop(vt.state, C.VTERM_PROP_CURSORBLINK, &cursorBlink)

	cursorVisible := C.vterm_value_from_int(1)
	C.vterm_state_set_termprop(vt.state, C.VTERM_PROP_CURSORVISIBLE, &cursorVisible)
}

func (vt *VTerm) Write(data []byte) (int, error) {
	if len(data) == 0 || data == nil {
		return 0, nil
	}

	vt.mutex.Lock()
	defer vt.mutex.Unlock()

	cData := (*C.char)(unsafe.SliceData(data))
	cLength := C.size_t(len(data))
	return int(C.vterm_input_write(vt.term, cData, cLength)), nil
}

func (vt *VTerm) SetSize(width int, height int) {
	vt.mutex.Lock()
	defer vt.mutex.Unlock()
	C.vterm_set_size(vt.term, C.int(height), C.int(width))
	C.vterm_screen_flush_damage(vt.screen)
	C.vterm_screen_reset(vt.screen, C.int(1))
}

func (vt *VTerm) GetSize() (int, int) {
	vt.mutex.RLock()
	defer vt.mutex.RUnlock()
	var width, height C.int
	C.vterm_get_size(vt.term, &height, &width)

	return int(width), int(height)
}

func (vt *VTerm) ensureBufferCapacity(width int, height int) {
	if vt.buffer == nil || len(vt.buffer) < width*height {
		vt.buffer = make([]C.VTermScreenCell, width*height)
	}
}

func (vt *VTerm) Contents() (string, error) {
	vt.mutex.RLock()
	defer vt.mutex.RUnlock()

	width, height := vt.GetSize()
	vt.ensureBufferCapacity(width, height)
	var cursorPos C.VTermPos
	C.vterm_state_get_cursorpos(vt.state, &cursorPos)

	area := C.VTermRect{
		start_row: 0,
		start_col: 0,
		end_row:   C.int(height),
		end_col:   C.int(width),
	}
	C.vterm_get_cells(vt.screen, area, unsafe.SliceData(vt.buffer), C.size_t(len(vt.buffer)))
	var output strings.Builder
	var lastFg, lastBg [4]uint8
	output.WriteString("\x1b[0m")

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := vt.buffer[y*width+x]
			fg := cell.fg
			bg := cell.bg
			isCursorHere := (x == int(cursorPos.col) && y == int(cursorPos.row))

			if fg != lastFg || bg != lastBg {
				fmt.Fprint(&output, "\x1b[")
				if fg[0]&0x2 != 0 {
					fmt.Fprintf(&output, "39;")
				} else if fg[0]&0x1 == 1 {
					fmt.Fprintf(&output, "38;5;%d;", fg[1])
				} else {
					fmt.Fprintf(&output, "38;2;%d;%d;%d;", fg[1], fg[2], fg[3])
				}

				if bg[0]&0x4 != 0 {
					fmt.Fprintf(&output, "49")
				} else if bg[0]&0x1 == 1 {
					fmt.Fprintf(&output, "48;5;%d", bg[1])
				} else {
					fmt.Fprintf(&output, "48;2;%d;%d;%d", bg[1], bg[2], bg[3])
				}
				lastFg, lastBg = fg, bg
				fmt.Fprint(&output, "m")
			}
			if isCursorHere {
				// Swap FG and BG for the cursor position
				fmt.Fprint(&output, "\x1b[7m")
			}

			for i := 0; i < int(cell.width); i++ {
				if cell.chars[i] == 0 && isCursorHere {
					output.WriteString(" ") // Write a space for empty cells at cursor
				} else {
					output.WriteRune(rune(cell.chars[i]))
				}
			}
			if isCursorHere {
				// Swap FG and BG back after the cursor position
				fmt.Fprint(&output, "\x1b[27m")
			}
		}
		if y == height-1 {
			output.WriteString("\x1b[0m")
		} else {
			output.WriteString("\x1b[0m\n")
		}
	}
	return output.String(), nil
}

func (vt *VTerm) Close() error {
	vt.mutex.Lock()
	defer vt.mutex.Unlock()
	C.vterm_free(vt.term)
	return nil
}
