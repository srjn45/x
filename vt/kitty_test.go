package vt

import (
	"slices"
	"strings"
	"testing"
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
