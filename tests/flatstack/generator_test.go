//go:build phpx

package flatstack_test

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/flatstack"
	"github.com/titpetric/phpx/parser"
)

// TestFlatstackRejectsGenerator pins a generator as a form the compiler
// declines: its body runs on a goroutine the interpreter's frame machinery
// hands control to, and the compiled VM has no seam to detach a frame on. The
// program still runs on the interpreter; this asserts the rejection, which a
// fixture cannot tell from a compiled run that agreed with it.
func TestFlatstackRejectsGenerator(t *testing.T) {
	t.Parallel()
	program, err := parser.Parse(`<?php function g() { yield 1; } foreach (g() as $v) { echo $v; }`)
	if err != nil {
		t.Fatal(err)
	}
	err = flatstack.Supports(program)
	if err == nil {
		t.Fatal("expected a generator to be rejected")
	}
	if !strings.Contains(err.Error(), "generator") {
		t.Fatalf("error = %v, want one naming the generator", err)
	}
}
