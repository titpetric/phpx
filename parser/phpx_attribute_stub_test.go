//go:build !phpx

package parser

import "testing"

// TestPhpxSkipAttributeUntagged pins that an untagged lexer skips nothing here.
func TestPhpxSkipAttributeUntagged(t *testing.T) {
	t.Parallel()
	if (&lexer{src: "#[A]"}).phpxSkipAttribute() {
		t.Error("untagged lexer skipped an attribute")
	}
}
