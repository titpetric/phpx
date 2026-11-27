package core

import (
	"testing"
)

// TestASCIICaseFolding covers the two folds this package writes itself in place
// of the strings package: php's strtolower and strtoupper fold the ASCII range
// alone. A byte over 0x7f travels unchanged. A UTF-8 sequence survives a round
// trip through either fold.
func TestASCIICaseFolding(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		in    string
		lower string
		upper string
	}{
		{"Hello", "hello", "HELLO"},
		{"ÄÖÜ", "ÄÖÜ", "ÄÖÜ"},
		{"a1_Z", "a1_z", "A1_Z"},
		{"", "", ""},
	} {
		if got := asciiLower(test.in); got != test.lower {
			t.Errorf("asciiLower(%q) = %q, want %q", test.in, got, test.lower)
		}
		if got := asciiUpper(test.in); got != test.upper {
			t.Errorf("asciiUpper(%q) = %q, want %q", test.in, got, test.upper)
		}
	}
}

// caseSink keeps a fold's result reachable, so its buffer cannot be stacked.
var caseSink string

// TestASCIICaseFoldingAllocations pins the allocation rule the profile reads:
// a fold that changes no byte hands the input string back and allocates
// nothing, while a fold that changes one allocates the result string alone.
func TestASCIICaseFoldingAllocations(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func()
		want float64
	}{
		{"asciiLower on lower-case text", func() { caseSink = asciiLower("already lower") }, 0},
		{"asciiUpper on upper-case text", func() { caseSink = asciiUpper("ALREADY UPPER") }, 0},
		{"asciiLower on mixed-case text", func() { caseSink = asciiLower("Hello World") }, 1},
		{"asciiUpper on mixed-case text", func() { caseSink = asciiUpper("Hello World") }, 1},
	} {
		if got := testing.AllocsPerRun(100, tc.run); got != tc.want {
			t.Errorf("%s allocated %v times, want %v", tc.name, got, tc.want)
		}
	}
}

// TestUcwordsSeparators covers ucwords(): the default separators are five
// whitespace bytes. A call naming its separators replaces the defaults. A
// hyphen capitalises only when the call names it.
func TestUcwordsSeparators(t *testing.T) {
	t.Parallel()
	if got := phpUcwords("hello big-world"); got != "Hello Big-world" {
		t.Errorf("ucwords = %q, want the default separators", got)
	}
	if got := phpUcwords("hello big-world", "-"); got != "Hello big-World" {
		t.Errorf("ucwords with a named separator = %q", got)
	}
	if got := phpUcfirst("étude"); got != "étude" {
		t.Errorf("ucfirst over a multi-byte first rune = %q, want it unchanged", got)
	}
	if got := phpLcfirst("Hello"); got != "hello" {
		t.Errorf("lcfirst = %q, want hello", got)
	}
}
