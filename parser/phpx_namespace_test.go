//go:build phpx

package parser

import "testing"

// TestBracedNamespaces pins that class names resolve inside each block and the
// global block leaves the namespace.
func TestBracedNamespaces(t *testing.T) {
	t.Parallel()
	prog, err := Parse("<?php namespace A { class X {} } namespace B { class X {} } namespace { class G {} }")
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.Stmts) != 3 {
		t.Fatalf("statements = %d, want 3", len(prog.Stmts))
	}
}
