//go:build phpx

package phpx

import "testing"

func TestSlashesRoundTrip(t *testing.T) {
	in := "a'b\"c\\d\x00e"
	esc := addSlashes(in)
	if esc != "a\\'b\\\"c\\\\d\\0e" {
		t.Fatalf("addSlashes = %q", esc)
	}
	if got := stripSlashes(esc); got != in {
		t.Fatalf("stripSlashes = %q", got)
	}
	if got := stripSlashes("x\\"); got != "x" {
		t.Fatalf("trailing backslash = %q", got)
	}
}
