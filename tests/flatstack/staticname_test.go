//go:build phpx

package flatstack_test

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/flatstack"
	"github.com/titpetric/phpx/parser"
)

// TestFlatstackRejectsStaticPropName pins the static property named by an
// expression as a form the compiler declines: the bytecode carries the
// property name, and a value names it here. The program still runs on the
// interpreter; this asserts the rejection, which a fixture cannot tell from a
// compiled run that agreed with it.
func TestFlatstackRejectsStaticPropName(t *testing.T) {
	t.Parallel()
	program, err := parser.Parse(`<?php class A { public static $x = 1; } $p = "x"; echo A::${$p};`)
	if err != nil {
		t.Fatal(err)
	}
	err = flatstack.Supports(program)
	if err == nil {
		t.Fatal("expected a static property named by an expression to be rejected")
	}
	if !strings.Contains(err.Error(), "dynamic static property name") {
		t.Fatalf("error = %v, want one naming the static property", err)
	}
}
