package vterm

/*
#include <vterm.h>
*/
import "C"

type KeyModifier byte

const (
	KeyModNone  KeyModifier = 0x00
	KeyModShift KeyModifier = 0x01
	KeyModAlt   KeyModifier = 0x02
	KeyModCtrl  KeyModifier = 0x04
)

type Key struct {
	IsUnichar bool
	Code      uint
	Modifier  KeyModifier
}

var (
	KeyNone         = Key{Code: C.VTERM_KEY_NONE}
	KeyEnter        = Key{Code: C.VTERM_KEY_ENTER}
	KeyTab          = Key{Code: C.VTERM_KEY_TAB}
	KeyBackspace    = Key{Code: C.VTERM_KEY_BACKSPACE}
	KeyEscape       = Key{Code: C.VTERM_KEY_ESCAPE}
	KeyUp           = Key{Code: C.VTERM_KEY_UP}
	KeyDown         = Key{Code: C.VTERM_KEY_DOWN}
	KeyLeft         = Key{Code: C.VTERM_KEY_LEFT}
	KeyRight        = Key{Code: C.VTERM_KEY_RIGHT}
	KeyIns          = Key{Code: C.VTERM_KEY_INS}
	KeyDel          = Key{Code: C.VTERM_KEY_DEL}
	KeyHome         = Key{Code: C.VTERM_KEY_HOME}
	KeyEnd          = Key{Code: C.VTERM_KEY_END}
	KeyPageup       = Key{Code: C.VTERM_KEY_PAGEUP}
	KeyPagedown     = Key{Code: C.VTERM_KEY_PAGEDOWN}
	KeyFunction0    = Key{Code: C.VTERM_KEY_FUNCTION_0}
	KeyFunctionMax  = Key{Code: C.VTERM_KEY_FUNCTION_MAX}
	KeyKeypad0      = Key{Code: C.VTERM_KEY_KP_0}
	KeyKeypad1      = Key{Code: C.VTERM_KEY_KP_1}
	KeyKeypad2      = Key{Code: C.VTERM_KEY_KP_2}
	KeyKeypad3      = Key{Code: C.VTERM_KEY_KP_3}
	KeyKeypad4      = Key{Code: C.VTERM_KEY_KP_4}
	KeyKeypad5      = Key{Code: C.VTERM_KEY_KP_5}
	KeyKeypad6      = Key{Code: C.VTERM_KEY_KP_6}
	KeyKeypad7      = Key{Code: C.VTERM_KEY_KP_7}
	KeyKeypad8      = Key{Code: C.VTERM_KEY_KP_8}
	KeyKeypad9      = Key{Code: C.VTERM_KEY_KP_9}
	KeyKeypadMult   = Key{Code: C.VTERM_KEY_KP_MULT}
	KeyKeypadPlus   = Key{Code: C.VTERM_KEY_KP_PLUS}
	KeyKeypadComma  = Key{Code: C.VTERM_KEY_KP_COMMA}
	KeyKeypadMinus  = Key{Code: C.VTERM_KEY_KP_MINUS}
	KeyKeypadPeriod = Key{Code: C.VTERM_KEY_KP_PERIOD}
	KeyKeypadDivide = Key{Code: C.VTERM_KEY_KP_DIVIDE}
	KeyKeypadEnter  = Key{Code: C.VTERM_KEY_KP_ENTER}
	KeyKeypadEqual  = Key{Code: C.VTERM_KEY_KP_EQUAL}
	KeyKeypadMax    = Key{Code: C.VTERM_KEY_MAX}
)
