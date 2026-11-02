//go:build phpx

package lint

import (
	"testing"

	"github.com/titpetric/phpx/parser"
)

// TestPhpxInterfaceViolationsReadsTheFinishedClass covers the linter's fork
// answer: a method the parent declares satisfies the contract, so the class is
// not reported, and a method nothing declares still is.
func TestPhpxInterfaceViolationsReadsTheFinishedClass(t *testing.T) {
	prog, err := parser.Parse(`<?php
interface Speaker { public function speak(); }
class Base { public function speak() {} }
class Child extends Base implements Speaker {}
class Other implements Speaker {}
`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	violations := phpxInterfaceViolations(prog)
	if len(violations) != 1 || violations[0].Class != "Other" {
		t.Errorf("violations = %v, want the one class nothing supplies", violations)
	}
}

// TestPhpxInterfaceViolationsSkipsAnUnknownParent covers the parent the file
// does not declare: the linter cannot tell a missing method from the parent's,
// and the runtime autoloads the parent and checks the finished class.
func TestPhpxInterfaceViolationsSkipsAnUnknownParent(t *testing.T) {
	prog, err := parser.Parse(`<?php
interface Speaker { public function speak(); }
class Child extends Vendor implements Speaker {}
`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if violations := phpxInterfaceViolations(prog); len(violations) != 0 {
		t.Errorf("violations = %v, want an unknown parent left to the runtime", violations)
	}
}
