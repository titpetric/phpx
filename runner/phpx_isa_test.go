//go:build phpx

package runner

import (
	"io"
	"testing"
)

func TestPhpxIsAClassNameNeedsAllowString(t *testing.T) {
	rt := New(io.Discard, Options{})
	if rt.PhpxIsA("stdClass", "stdClass", false) {
		t.Fatal("a class name string is no object without allow_string")
	}
	if !rt.PhpxIsA("stdClass", "stdClass", true) {
		t.Fatal("allow_string should accept a declared class name")
	}
	if rt.PhpxIsA("NoSuchClass", "NoSuchClass", true) {
		t.Fatal("an unknown class name is false")
	}
}
