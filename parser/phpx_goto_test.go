//go:build phpx

package parser

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestParseGotoAndLabel pins the two statements and what a label is not.
func TestParseGotoAndLabel(t *testing.T) {
	t.Parallel()
	prog, err := Parse("<?php start: $a = 1; goto start; switch ($a) { default: break; } class C { const X = self::Y; }")
	if err != nil {
		t.Fatal(err)
	}
	if l, ok := prog.Stmts[0].(*model.Label); !ok || l.Name != "start" {
		t.Errorf("first statement = %#v", prog.Stmts[0])
	}
	if g, ok := prog.Stmts[2].(*model.Goto); !ok || g.Label != "start" {
		t.Errorf("third statement = %#v", prog.Stmts[2])
	}
}
