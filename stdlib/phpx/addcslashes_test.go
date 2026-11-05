//go:build phpx

package phpx

import "testing"

func TestAddCSlashes(t *testing.T) {
	cases := []struct{ in, chars, want string }{
		{"foo[bar]", "A..Z", "foo[bar]"},
		{"zoo['.']", "z..A", "\\zoo['\\.']"},
		{"a\x01b\x7f", "\x00..\x1f\x7f", "a\\001b\\177"},
		{"line\n", "\n", "line\\n"},
		{"x.y", ".", "x\\.y"},
	}
	for _, c := range cases {
		if got := addCSlashes(c.in, c.chars); got != c.want {
			t.Errorf("addCSlashes(%q, %q) = %q, want %q", c.in, c.chars, got, c.want)
		}
	}
}
