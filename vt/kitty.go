package vt

import (
	"slices"

	"github.com/charmbracelet/x/ansi"
)

// kittySupported are the levels of the kitty keyboard protocol the emulator
// speaks; any other a program asks for is dropped when the flags are stored.
const kittySupported = ansi.KittyDisambiguateEscapeCodes | ansi.KittyReportAllKeysAsEscapeCodes

// kittyStackSize is how many entries each screen's stack of kitty keyboard
// flags holds: a push onto a full one drops the oldest, as kitty does, so a
// program pushing in a loop cannot grow it without end.
const kittyStackSize = 16

// kittyStack returns the current screen's stack of kitty keyboard flags: the
// main and the alternate screen each have their own, as the protocol says.
func (e *Emulator) kittyStack() *[]int {
	if e.IsAltScreen() {
		return &e.kitty[1]
	}
	return &e.kitty[0]
}

// pushKittyFlags pushes flags onto the current screen's stack [ansi.PushKittyKeyboard].
func (e *Emulator) pushKittyFlags(flags int) {
	s := e.kittyStack()
	if len(*s) == kittyStackSize {
		*s = slices.Delete(*s, 0, 1)
	}
	*s = append(*s, flags&kittySupported)
}

// popKittyFlags pops n entries off the current screen's stack, emptying it
// when it holds fewer [ansi.PopKittyKeyboard].
func (e *Emulator) popKittyFlags(n int) {
	s := e.kittyStack()
	*s = (*s)[:max(len(*s)-n, 0)]
}

// setKittyFlags changes the top of the current screen's stack
// [ansi.KittyKeyboard]: mode 1 sets it to flags, 2 adds them and 3 removes
// them. On an empty stack it changes the flags in force, 0, as if they were
// the top. It reports false for any other mode.
func (e *Emulator) setKittyFlags(flags, mode int) bool {
	s := e.kittyStack()
	if len(*s) == 0 {
		*s = append(*s, 0)
	}
	top := &(*s)[len(*s)-1]
	switch mode {
	case 1:
		*top = flags
	case 2:
		*top |= flags
	case 3:
		*top &^= flags
	default:
		return false
	}
	*top &= kittySupported
	return true
}

// KittyFlags returns the stacks of kitty keyboard flags of the main and the
// alternate screen, oldest first. The flags in force are the top of the
// current screen's stack, 0 when it is empty.
func (e *Emulator) KittyFlags() (main, alt []int) {
	return slices.Clone(e.kitty[0]), slices.Clone(e.kitty[1])
}
