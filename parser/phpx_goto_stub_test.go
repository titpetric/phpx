//go:build !phpx

package parser

import (
	"testing"
)

// TestPhpxGotoStmtUntagged pins that the untagged parser reads no goto.
func TestPhpxGotoStmtUntagged(t *testing.T) {
	t.Parallel()
	if _, ok := (&parser{}).phpxGotoStmt(token{}); ok {
		t.Error("untagged parser read a goto")
	}
}
