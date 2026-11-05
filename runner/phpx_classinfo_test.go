//go:build phpx

package runner

import (
	"io"
	"testing"
)

func TestPhpxClassAliasOfInterface(t *testing.T) {
	rt := New(io.Discard, Options{})
	if rt.PhpxClassAlias("No\\Such\\Thing", "Alias", false) {
		t.Fatal("an unknown original was aliased")
	}
}

func TestPhpxAliasMatches(t *testing.T) {
	phpxInterfaceAliases.Store("legacy_x", "x")
	if !phpxAliasMatches("Legacy_X", "X") || !phpxAliasMatches("x", "legacy_x") {
		t.Fatal("an alias and its interface are one")
	}
	if phpxAliasMatches("Y", "X") {
		t.Fatal("two interfaces were taken for one")
	}
}
