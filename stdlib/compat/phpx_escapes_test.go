package compat

import (
	"testing"
)

func TestPhpxEscapes(t *testing.T) {
	for in, want := range map[string]string{
		`(.*\R)`:    `(.*(?:\r\n|\n|\r))`,
		`a\hb`:      `a[ \t]b`,
		`[\h,]`:     `[ \t,]`,
		`\d+\\R`:    `\d+\\R`,
		`plain`:     `plain`,
		`\w*+b`:     `(?>\w*)b`,
		`[a-z]++`:   `(?>[a-z]+)`,
		`(?:ab)?+c`: `(?>(?:ab)?)c`,
		`x{2}+y`:    `(?>x{2})y`,
		`\++`:       `\++`,
		`[+]+`:      `[+]+`,
		`(?+1)`:     `(?+1)`,
	} {
		if got := phpxEscapes(in, false); got != want {
			t.Errorf("phpxEscapes(%q) = %q, want %q", in, got, want)
		}
	}
}
