package vt

import (
	"slices"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestKittyFlags(t *testing.T) {
	full := strings.Repeat("\x1b[>1u", kittyStackSize)
	for _, tc := range []struct {
		name, input string
		want        []int
	}{
		{"none before any push", "", nil},
		{"push", "\x1b[>1u\x1b[>8u", []int{1, 8}},
		{"push without flags", "\x1b[>u", []int{0}},
		{"only the supported levels are kept", "\x1b[>31u", []int{9}},
		{"pop one", "\x1b[>1u\x1b[>9u\x1b[<u", []int{1}},
		{"pop n", "\x1b[>1u\x1b[>8u\x1b[>9u\x1b[<2u", []int{1}},
		{"pop past the bottom empties", "\x1b[>1u\x1b[<5u", []int{}},
		{"pop an empty stack", "\x1b[<u", nil},
		{"set replaces the top", "\x1b[>1u\x1b[>1u\x1b[=8u", []int{1, 8}},
		{"set in mode 1", "\x1b[>1u\x1b[=8;1u", []int{8}},
		{"set in mode 2 adds", "\x1b[>1u\x1b[=8;2u", []int{9}},
		{"set in mode 3 removes", "\x1b[>9u\x1b[=1;3u", []int{8}},
		{"set keeps only the supported levels", "\x1b[>1u\x1b[=6;2u", []int{1}},
		{"set on an empty stack", "\x1b[=1u", []int{1}},
		{"set in an unknown mode is ignored", "\x1b[>1u\x1b[=8;4u", []int{1}},
		{"a full stack drops the oldest", "\x1b[>8u" + full, slices.Repeat([]int{1}, kittyStackSize)},
		{"a full reset empties it", "\x1b[>1u\x1bc", nil},
		{"a soft reset leaves it", "\x1b[>1u\x1b[!p", []int{1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			term := newTestTerminal(t, 10, 1)
			_, _ = term.WriteString(tc.input)
			main, alt := term.KittyFlags()
			if !slices.Equal(main, tc.want) || len(alt) != 0 {
				t.Fatalf("want main %v and alternate [], got %v and %v", tc.want, main, alt)
			}
		})
	}
}

func TestKittyFlagsPerScreen(t *testing.T) {
	term := newTestTerminal(t, 10, 1)
	_, _ = term.WriteString("\x1b[>1u\x1b[?1049h\x1b[>8u\x1b[=1;2u")
	if main, alt := term.KittyFlags(); !slices.Equal(main, []int{1}) || !slices.Equal(alt, []int{9}) {
		t.Fatalf("on the alternate screen: want main [1] and alternate [9], got %v and %v", main, alt)
	}
	_, _ = term.WriteString("\x1b[?1049l\x1b[<u")
	if main, alt := term.KittyFlags(); len(main) != 0 || !slices.Equal(alt, []int{9}) {
		t.Fatalf("back on the main screen: want main [] and alternate [9], got %v and %v", main, alt)
	}
	_, _ = term.WriteString("\x1b[?1049h\x1b[>1u")
	if main, alt := term.KittyFlags(); len(main) != 0 || !slices.Equal(alt, []int{9, 1}) {
		t.Fatalf("on the alternate screen again: want main [] and alternate [9 1], got %v and %v", main, alt)
	}
	_, _ = term.WriteString("\x1bc")
	if main, alt := term.KittyFlags(); main != nil || alt != nil {
		t.Fatalf("after a full reset: want both empty, got %v and %v", main, alt)
	}
}

func TestKittyFlagsAreACopy(t *testing.T) {
	term := newTestTerminal(t, 10, 1)
	_, _ = term.WriteString("\x1b[>1u")
	main, _ := term.KittyFlags()
	main[0] = 8
	if main, _ := term.KittyFlags(); !slices.Equal(main, []int{1}) {
		t.Fatalf("want a copy of the stack, got %v", main)
	}
}

func TestKittyKey(t *testing.T) {
	type key = KeyPressEvent
	for _, tc := range []struct {
		name  string
		key   key
		flags int
		want  string // "legacy" when the key is sent as without the protocol
	}{
		{"text", key{Code: 'a', Text: "a"}, 1, "legacy"},
		{"shifted text", key{Code: 'a', Text: "A", Mod: ModShift}, 1, "legacy"},
		{"Esc", key{Code: KeyEscape}, 1, "\x1b[27u"},
		{"Ctrl+C", key{Code: 'c', Mod: ModCtrl}, 1, "\x1b[99;5u"},
		{"Ctrl+Shift+A", key{Code: 'a', Mod: ModCtrl | ModShift}, 1, "\x1b[97;6u"},
		{"Ctrl+Shift+A with the shifted code", key{Code: 'A', Mod: ModCtrl | ModShift}, 1, "\x1b[97;6u"},
		{"Alt+x", key{Code: 'x', Text: "x", Mod: ModAlt}, 1, "\x1b[120;3u"},
		{"Ctrl+Space", key{Code: KeySpace, Mod: ModCtrl}, 1, "\x1b[32;5u"},
		{"Enter", key{Code: KeyEnter}, 1, "legacy"},
		{"Shift+Enter", key{Code: KeyEnter, Mod: ModShift}, 1, "\x1b[13;2u"},
		{"Ctrl+Enter", key{Code: KeyEnter, Mod: ModCtrl}, 1, "\x1b[13;5u"},
		{"Tab", key{Code: KeyTab}, 1, "legacy"},
		{"Shift+Tab", key{Code: KeyTab, Mod: ModShift}, 1, "\x1b[9;2u"},
		{"Backspace", key{Code: KeyBackspace}, 1, "legacy"},
		{"Alt+Backspace", key{Code: KeyBackspace, Mod: ModAlt}, 1, "\x1b[127;3u"},
		{"Up", key{Code: KeyUp}, 1, "legacy"},
		{"Shift+Up", key{Code: KeyUp, Mod: ModShift}, 1, "\x1b[1;2A"},
		{"Ctrl+Home", key{Code: KeyHome, Mod: ModCtrl}, 1, "\x1b[1;5H"},
		{"Shift+F1", key{Code: KeyF1, Mod: ModShift}, 1, "\x1b[1;2P"},
		{"Shift+F3", key{Code: KeyF3, Mod: ModShift}, 1, "\x1b[13;2~"},
		{"Ctrl+Delete", key{Code: KeyDelete, Mod: ModCtrl}, 1, "\x1b[3;5~"},
		{"Shift+F12", key{Code: KeyF12, Mod: ModShift}, 1, "\x1b[24;2~"},
		{"F13", key{Code: KeyF13}, 1, "\x1b[57376u"},
		{"F35", key{Code: KeyF35}, 1, "\x1b[57398u"},
		{"keypad 5", key{Code: KeyKp5}, 1, "legacy"},
		{"Ctrl+keypad 5", key{Code: KeyKp5, Mod: ModCtrl}, 1, "\x1b[57404;5u"},
		{"keypad Enter", key{Code: KeyKpEnter}, 1, "\x1b[57414u"},
		{"keypad Up", key{Code: KeyKpUp}, 1, "\x1b[57419u"},
		{"Shift alone", key{Code: uv.KeyLeftShift, Mod: ModShift}, 1, ""},
		{"Caps Lock", key{Code: uv.KeyCapsLock}, 1, ""},
		{"Caps Lock is no modifier at level 1", key{Code: 'c', Mod: ModCtrl | uv.ModCapsLock}, 1, "\x1b[99;5u"},
		{"Meta is the protocol's 32", key{Code: 'a', Mod: ModMeta}, 1, "\x1b[97;33u"},
		{"Super is the protocol's 8", key{Code: 'a', Mod: uv.ModSuper}, 1, "\x1b[97;9u"},
		{"input method text", key{Code: KeyExtended, Text: "日本"}, 1, "legacy"},

		{"text at level 8", key{Code: 'a', Text: "a"}, 9, "\x1b[97u"},
		{"shifted text at level 8", key{Code: 'a', Text: "A", Mod: ModShift}, 9, "\x1b[97;2u"},
		{"Enter at level 8", key{Code: KeyEnter}, 9, "\x1b[13u"},
		{"Tab at level 8", key{Code: KeyTab}, 9, "\x1b[9u"},
		{"Backspace at level 8", key{Code: KeyBackspace}, 9, "\x1b[127u"},
		{"Up at level 8", key{Code: KeyUp}, 9, "\x1b[A"},
		{"Delete at level 8", key{Code: KeyDelete}, 9, "\x1b[3~"},
		{"keypad 5 at level 8", key{Code: KeyKp5}, 9, "\x1b[57404u"},
		{"Shift alone at level 8", key{Code: uv.KeyLeftShift, Mod: ModShift}, 9, "\x1b[57441;2u"},
		{"Caps Lock is a modifier at level 8", key{Code: 'c', Mod: uv.ModCapsLock}, 8, "\x1b[99;65u"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := kittyKey(tc.key, tc.flags)
			if !ok {
				got = "legacy"
			}
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

// TestSendKeyKittyFlags sends keys through the emulator, at the flags the
// output set: a key level 1 leaves alone is sent as without the protocol,
// which sends Shift+Enter as nothing.
func TestSendKeyKittyFlags(t *testing.T) {
	for _, tc := range []struct {
		name, output string
		keys         []KeyPressEvent
		want         string
	}{
		{"no flags", "", []KeyPressEvent{{Code: KeyEnter, Mod: ModShift}, {Code: 'c', Mod: ModCtrl}}, "\x03"},
		{"level 1", "\x1b[>1u", []KeyPressEvent{{Code: KeyEnter, Mod: ModShift}, {Code: KeyEnter}, {Code: 'a', Text: "A", Mod: ModShift}}, "\x1b[13;2u\rA"},
		{"level 1 popped", "\x1b[>1u\x1b[<u", []KeyPressEvent{{Code: KeyEnter, Mod: ModShift}}, ""},
		{"level 1 on the other screen", "\x1b[>1u\x1b[?1049h", []KeyPressEvent{{Code: KeyEnter, Mod: ModShift}}, ""},
		{"level 8", "\x1b[>8u", []KeyPressEvent{{Code: 'a', Text: "a"}, {Code: KeyEnter}}, "\x1b[97u\x1b[13u"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			term := newTestTerminal(t, 10, 1)
			_, _ = term.WriteString(tc.output)
			got := sent(term, func() {
				for _, k := range tc.keys {
					term.SendKey(k)
				}
			})
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestKittyFlagsAnswered(t *testing.T) {
	for _, tc := range []struct{ output, want string }{
		{"", "\x1b[?0u"},
		{"\x1b[>1u", "\x1b[?1u"},
		{"\x1b[>31u", "\x1b[?9u"},
		{"\x1b[>1u\x1b[?1049h", "\x1b[?0u"},
	} {
		term := newTestTerminal(t, 10, 1)
		_, _ = term.WriteString(tc.output)
		got := sent(term, func() { _, _ = term.WriteString("\x1b[?u") })
		if got != tc.want {
			t.Fatalf("after %q: want %q, got %q", tc.output, tc.want, got)
		}
	}
}

// sent returns what the terminal sent the program while do ran: a marker
// sent after it ends the reading, so a key sent as nothing is seen as such.
func sent(term *Emulator, do func()) string {
	const marker = "<end>"
	done := make(chan string, 1)
	go func() {
		var out []byte
		buf := make([]byte, 64)
		for !strings.HasSuffix(string(out), marker) {
			n, err := term.Read(buf)
			if err != nil {
				break
			}
			out = append(out, buf[:n]...)
		}
		done <- strings.TrimSuffix(string(out), marker)
	}()
	do()
	term.SendText(marker)
	return <-done
}
