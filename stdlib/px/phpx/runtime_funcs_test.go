//go:build phpx

package phpx

import "testing"

func TestSplitFileLines(t *testing.T) {
	cases := []struct {
		in    string
		flags int64
		want  []string
	}{
		{"a\nb\n\nc\n", 0, []string{"a\n", "b\n", "\n", "c\n"}},
		{"a\nb\n\nc\n", 2, []string{"a", "b", "", "c"}},
		{"a\nb\n\nc\n", 6, []string{"a", "b", "c"}},
		{"a\nb\n\nc\n", 4, []string{"a\n", "b\n", "\n", "c\n"}},
		{"a\r\nb", 2, []string{"a", "b"}},
		{"", 0, nil},
	}
	for _, c := range cases {
		got := splitFileLines(c.in, c.flags)
		if got.Len() != len(c.want) {
			t.Errorf("%q flags %d: %d lines, want %d", c.in, c.flags, got.Len(), len(c.want))
			continue
		}
		for i, w := range c.want {
			if v, _ := got.Get(int64(i)); v != w {
				t.Errorf("%q flags %d line %d = %q, want %q", c.in, c.flags, i, v, w)
			}
		}
	}
}
