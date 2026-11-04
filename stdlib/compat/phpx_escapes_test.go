package compat

import (
	"testing"
)

func TestPhpxEscapes(t *testing.T) {
	for in, want := range map[string]string{
		`(.*\R)`: `(.*(?:\r\n|\n|\r))`,
		`a\hb`:   `a[ \t]b`,
		`[\h,]`:  `[ \t,]`,
		`\d+\\R`: `\d+\\R`,
		`plain`:  `plain`,
	} {
		if got := phpxEscapes(in); got != want {
			t.Errorf("phpxEscapes(%q) = %q, want %q", in, got, want)
		}
	}
}
