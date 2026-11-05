//go:build phpx

package parser

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestParseBreakLevels pins that the number after break and continue is read.
func TestParseBreakLevels(t *testing.T) {
	t.Parallel()
	prog, err := Parse("<?php for (;;) { for (;;) { break 2; continue 3; break; } }")
	if err != nil {
		t.Fatal(err)
	}
	inner := prog.Stmts[0].(*model.For).Body[0].(*model.For).Body
	if b := inner[0].(*model.Break); b.Levels != 2 {
		t.Errorf("break 2 levels = %d", b.Levels)
	}
	if c := inner[1].(*model.Continue); c.Levels != 3 {
		t.Errorf("continue 3 levels = %d", c.Levels)
	}
	if b := inner[2].(*model.Break); b.Levels != 0 {
		t.Errorf("break levels = %d", b.Levels)
	}
}
