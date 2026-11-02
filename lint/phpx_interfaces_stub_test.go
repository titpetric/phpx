//go:build !phpx

package lint

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/parser"
)

// TestPhpxInterfaceViolationsUntagged pins the linter's answer in an untagged
// build: a class has the methods it declares in its own body, so a method a
// parent carries is reported missing. The fork's build is phpx_interfaces.go.
func TestPhpxInterfaceViolationsUntagged(t *testing.T) {
	prog, err := parser.Parse(`<?php
interface Speaker { public function speak(); }
class Base { public function speak() {} }
class Child extends Base implements Speaker {}
`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	violations := phpxInterfaceViolations(prog)
	if len(violations) != 1 || !strings.Contains(violations[0].String(), "class Child does not declare method speak()") {
		t.Errorf("violations = %v, want the inherited method reported missing", violations)
	}
}
