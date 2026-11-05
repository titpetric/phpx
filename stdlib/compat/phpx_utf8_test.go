//go:build phpx

package compat

import "testing"

func TestPhpxUCP(t *testing.T) {
	got, boundary := phpxUCP(`\w+\d[\s\w]\b`)
	want := `[\p{L}\p{N}_]+\p{Nd}[\s\x{85}\x{a0}\p{Z}\p{L}\p{N}_]\b`
	if got != want || !boundary {
		t.Fatalf("got %q boundary %v, want %q", got, boundary, want)
	}
}

func TestPhpxBadUTF8(t *testing.T) {
	c := newRegexpCache()
	if !c.phpxBadUTF8("/./u", "\xe2") || c.lastError.Load() != 4 {
		t.Fatal("an invalid subject under /u should be refused with error 4")
	}
	if c.phpxBadUTF8("/./", "\xe2") || c.lastError.Load() != 0 {
		t.Fatal("without /u any bytes are a subject")
	}
}
