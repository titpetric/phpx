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

func TestPhpxParentClassOfBuiltins(t *testing.T) {
	rt := New(io.Discard, Options{})
	if got := rt.PhpxParentClass("UnexpectedValueException"); got != "RuntimeException" {
		t.Errorf("parent = %v", got)
	}
	if got := rt.PhpxParentClass("Exception"); got != false {
		t.Errorf("Exception has no parent, got %v", got)
	}
	if !rt.PhpxIsSubclassOf("BadMethodCallException", "LogicException", true) {
		t.Error("BadMethodCallException is a LogicException")
	}
	if rt.PhpxIsSubclassOf("RuntimeException", "RuntimeException", true) {
		t.Error("a class is not its own subclass")
	}
}
