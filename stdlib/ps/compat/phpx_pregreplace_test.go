//go:build phpx

package compat

import (
	"strings"
	"testing"
)

func TestPhpxExpandReplacement(t *testing.T) {
	subject := "ab"
	idx := []int{0, 2, 0, 1, 1, 2}
	cases := map[string]string{
		`\1-\2`: "a-b", "$2$1": "ba", "${1}x": "ax", `\\1`: `\1`, "$": "$", "a$b": "a$b",
		`\0`: "ab", "$x": "$x", `\9`: "", "${9}": "",
	}
	for repl, want := range cases {
		var b strings.Builder
		phpxExpandReplacement(&b, repl, subject, idx)
		if b.String() != want {
			t.Errorf("replacement %q gave %q, want %q", repl, b.String(), want)
		}
	}
}

func TestPhpxReplaceOneLimit(t *testing.T) {
	c := newRegexpCache()
	got, n, ok := c.phpxReplaceOne("/a/", "X", "aaaa", 2)
	if !ok || got != "XXaa" || n != 2 {
		t.Fatalf("got %q %d %v", got, n, ok)
	}
	if _, _, ok := c.phpxReplaceOne("/(/", "X", "a", -1); ok {
		t.Fatal("a pattern that does not compile should fail")
	}
}
