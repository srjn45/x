package vt

import (
	"fmt"
	"slices"
	"unicode"

	uv "github.com/charmbracelet/ultraviolet"
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

// kittyFlagsInForce returns the top of the current screen's stack, 0 when it
// is empty.
func (e *Emulator) kittyFlagsInForce() int {
	s := *e.kittyStack()
	if len(s) == 0 {
		return 0
	}
	return s[len(s)-1]
}

// kittyFlagsReport answers a program asking for the flags in force.
func kittyFlagsReport(flags int) string {
	return fmt.Sprintf("\x1b[?%du", flags)
}

// kittyKeyCode is how the kitty keyboard protocol sends a key that types no
// text: CSI number;modifiers final.
type kittyKeyCode struct {
	number int
	final  byte
}

// kittyKeyCodes are the keys that type no text, by the protocol's table of
// functional keys.
var kittyKeyCodes = map[rune]kittyKeyCode{
	KeyEscape: {27, 'u'}, KeyEnter: {13, 'u'}, KeyTab: {9, 'u'}, KeyBackspace: {127, 'u'},
	KeyInsert: {2, '~'}, KeyDelete: {3, '~'}, KeyPgUp: {5, '~'}, KeyPgDown: {6, '~'},
	KeyLeft: {1, 'D'}, KeyRight: {1, 'C'}, KeyUp: {1, 'A'}, KeyDown: {1, 'B'},
	KeyHome: {1, 'H'}, KeyEnd: {1, 'F'}, KeyBegin: {1, 'E'},
	KeyF1: {1, 'P'}, KeyF2: {1, 'Q'}, KeyF3: {13, '~'}, KeyF4: {1, 'S'},
	KeyF5: {15, '~'}, KeyF6: {17, '~'}, KeyF7: {18, '~'}, KeyF8: {19, '~'},
	KeyF9: {20, '~'}, KeyF10: {21, '~'}, KeyF11: {23, '~'}, KeyF12: {24, '~'},

	uv.KeyCapsLock: {57358, 'u'}, uv.KeyScrollLock: {57359, 'u'}, uv.KeyNumLock: {57360, 'u'},
	uv.KeyPrintScreen: {57361, 'u'}, uv.KeyPause: {57362, 'u'}, uv.KeyMenu: {57363, 'u'},

	KeyKp0: {57399, 'u'}, KeyKp1: {57400, 'u'}, KeyKp2: {57401, 'u'}, KeyKp3: {57402, 'u'},
	KeyKp4: {57403, 'u'}, KeyKp5: {57404, 'u'}, KeyKp6: {57405, 'u'}, KeyKp7: {57406, 'u'},
	KeyKp8: {57407, 'u'}, KeyKp9: {57408, 'u'},
	KeyKpDecimal: {57409, 'u'}, KeyKpDivide: {57410, 'u'}, KeyKpMultiply: {57411, 'u'},
	KeyKpMinus: {57412, 'u'}, KeyKpPlus: {57413, 'u'}, KeyKpEnter: {57414, 'u'},
	KeyKpEqual: {57415, 'u'}, KeyKpSep: {57416, 'u'},
	KeyKpLeft: {57417, 'u'}, KeyKpRight: {57418, 'u'}, KeyKpUp: {57419, 'u'}, KeyKpDown: {57420, 'u'},
	KeyKpPgUp: {57421, 'u'}, KeyKpPgDown: {57422, 'u'}, KeyKpHome: {57423, 'u'}, KeyKpEnd: {57424, 'u'},
	KeyKpInsert: {57425, 'u'}, KeyKpDelete: {57426, 'u'}, KeyKpBegin: {57427, 'u'},

	uv.KeyMediaPlay: {57428, 'u'}, uv.KeyMediaPause: {57429, 'u'}, uv.KeyMediaPlayPause: {57430, 'u'},
	uv.KeyMediaReverse: {57431, 'u'}, uv.KeyMediaStop: {57432, 'u'}, uv.KeyMediaFastForward: {57433, 'u'},
	uv.KeyMediaRewind: {57434, 'u'}, uv.KeyMediaNext: {57435, 'u'}, uv.KeyMediaPrev: {57436, 'u'},
	uv.KeyMediaRecord: {57437, 'u'}, uv.KeyLowerVol: {57438, 'u'}, uv.KeyRaiseVol: {57439, 'u'},
	uv.KeyMute: {57440, 'u'},

	uv.KeyLeftShift: {57441, 'u'}, uv.KeyLeftCtrl: {57442, 'u'}, uv.KeyLeftAlt: {57443, 'u'},
	uv.KeyLeftSuper: {57444, 'u'}, uv.KeyLeftHyper: {57445, 'u'}, uv.KeyLeftMeta: {57446, 'u'},
	uv.KeyRightShift: {57447, 'u'}, uv.KeyRightCtrl: {57448, 'u'}, uv.KeyRightAlt: {57449, 'u'},
	uv.KeyRightSuper: {57450, 'u'}, uv.KeyRightHyper: {57451, 'u'}, uv.KeyRightMeta: {57452, 'u'},
	uv.KeyIsoLevel3Shift: {57453, 'u'}, uv.KeyIsoLevel5Shift: {57454, 'u'},
}

func init() {
	for i := range 23 {
		kittyKeyCodes[KeyF13+rune(i)] = kittyKeyCode{57376 + i, 'u'}
	}
}

// kittyKeypadText are the keypad's keys that type a character: at level 1,
// as in kitty, they still type it when pressed alone or with Shift.
var kittyKeypadText = map[rune]bool{
	KeyKp0: true, KeyKp1: true, KeyKp2: true, KeyKp3: true, KeyKp4: true,
	KeyKp5: true, KeyKp6: true, KeyKp7: true, KeyKp8: true, KeyKp9: true,
	KeyKpDecimal: true, KeyKpDivide: true, KeyKpMultiply: true, KeyKpMinus: true,
	KeyKpPlus: true, KeyKpEqual: true, KeyKpSep: true,
}

// kittyOnlyAll are the keys the protocol reports only when all keys are sent
// as codes: the lock keys and the modifiers pressed alone.
var kittyOnlyAll = map[rune]bool{
	uv.KeyCapsLock: true, uv.KeyScrollLock: true, uv.KeyNumLock: true,
	uv.KeyLeftShift: true, uv.KeyLeftCtrl: true, uv.KeyLeftAlt: true,
	uv.KeyLeftSuper: true, uv.KeyLeftHyper: true, uv.KeyLeftMeta: true,
	uv.KeyRightShift: true, uv.KeyRightCtrl: true, uv.KeyRightAlt: true,
	uv.KeyRightSuper: true, uv.KeyRightHyper: true, uv.KeyRightMeta: true,
	uv.KeyIsoLevel3Shift: true, uv.KeyIsoLevel5Shift: true,
}

// The protocol's modifier bits.
const (
	kittyShift    = 1
	kittyAlt      = 2
	kittyCtrl     = 4
	kittySuper    = 8
	kittyHyper    = 16
	kittyMeta     = 32
	kittyCapsLock = 64
	kittyNumLock  = 128
)

// kittyKey encodes k as the kitty keyboard protocol does at flags, and
// reports false for a key sent as without the protocol: at level 1, a key
// typing text alone or with Shift, and Enter, Tab, Backspace and the other
// keys with a legacy form, pressed alone. A key the protocol does not report
// at flags is encoded as nothing.
func kittyKey(k KeyPressEvent, flags int) (string, bool) {
	all := flags&ansi.KittyReportAllKeysAsEscapeCodes != 0
	mods := kittyMods(k.Mod, all)
	code, functional := kittyKeyCodes[k.Code]
	if !functional {
		if k.Code == KeyExtended || k.Code == 0 {
			return "", false // text no single key names, as from an input method
		}
		code = kittyKeyCode{int(unicode.ToLower(k.Code)), 'u'}
	}
	if !all {
		switch {
		case kittyOnlyAll[k.Code]:
			return "", true
		case !functional || kittyKeypadText[k.Code]:
			if mods&^kittyShift == 0 {
				return "", false
			}
		case k.Code == KeyEscape:
			// Escape alone is told apart too: as ESC it starts every
			// sequence.
		case code.number >= 57344:
			// The protocol's own keys have no legacy form.
		case mods == 0:
			return "", false
		}
	}
	switch {
	case mods != 0:
		return fmt.Sprintf("\x1b[%d;%d%c", code.number, mods+1, code.final), true
	case code.number == 1:
		return fmt.Sprintf("\x1b[%c", code.final), true
	default:
		return fmt.Sprintf("\x1b[%d%c", code.number, code.final), true
	}
}

// kittyMods returns the protocol's modifier bits for m, the lock keys' only
// when all keys are sent as codes. Bubble Tea keeps xterm's bits, in which
// Meta and Super are the other way round.
func kittyMods(m KeyMod, all bool) int {
	var bits int
	for mod, bit := range map[KeyMod]int{
		ModShift: kittyShift, ModAlt: kittyAlt, ModCtrl: kittyCtrl,
		uv.ModSuper: kittySuper, uv.ModHyper: kittyHyper, ModMeta: kittyMeta,
	} {
		if m&mod != 0 {
			bits |= bit
		}
	}
	if all {
		if m&uv.ModCapsLock != 0 {
			bits |= kittyCapsLock
		}
		if m&uv.ModNumLock != 0 {
			bits |= kittyNumLock
		}
	}
	return bits
}
