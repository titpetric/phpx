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

// TestPhpxTraitRulesPrintBack pins that the adaptation block of a `use` line
// survives a rewrite.
func TestPhpxTraitRulesPrintBack(t *testing.T) {
	t.Parallel()
	prog, err := parser.Parse("<?php\nclass C {\n use A, B {\n A::f insteadof B;\n B::f as protected g;\n h as private;\n }\n}\n")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got := formatter.Print(prog, formatter.Options{})
	for _, want := range []string{"use A, B {", "A::f insteadof B;", "B::f as protected g;", "h as private;"} {
		if !strings.Contains(got, want) {
			t.Errorf("output does not carry %q:\n%s", want, got)
		}
	}
}
