package formatter_test

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/formatter"
	"github.com/titpetric/phpx/parser"
)

// TestPhpxTraitPrintsBack pins the formatter's half of the trait reversal: a
// trait declaration and the `use` line a class carries survive a rewrite, so a
// file with a trait is not deleted by the printer that did not know the node.
func TestPhpxTraitPrintsBack(t *testing.T) {
	t.Parallel()
	in := `<?php

trait Named {
	public $name = "x";

	public function labels() {
		return $this->name;
	}
}

class Box {
	use Named, Other;
}
`
	prog, err := parser.Parse(in)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got := formatter.Print(prog, formatter.Options{})
	for _, want := range []string{"trait Named {", "use Named, Other;", "public function labels() {"} {
		if !strings.Contains(got, want) {
			t.Errorf("output does not carry %q:\n%s", want, got)
		}
	}
}
