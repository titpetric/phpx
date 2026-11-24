package parser

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestAssignAsPrefixOperand pins the parse: a prefix operator folds onto the
// assignment beneath it, because php reads the operand as the assignment and
// not as the operator's result. symfony's CurlResponse reads `(int) $ch = $h`.
func TestAssignAsPrefixOperand(t *testing.T) {
	t.Parallel()
	prog, err := Parse("<?php $a = (int) $ch = $h; $b = -$n = 3; $c = @$s = 'b'; $d = clone $o = new C;")
	if err != nil {
		t.Fatal(err)
	}
	values := make([]model.Expr, 0, 4)
	for _, stmt := range prog.Stmts {
		values = append(values, stmt.(*model.Assign).Value)
	}
	cast, ok := values[0].(*model.Cast)
	if !ok {
		t.Fatalf("`(int) $ch = $h` = %T", values[0])
	}
	if assign, ok := cast.X.(*model.AssignExpr); !ok || assign.Op != "=" {
		t.Errorf("cast operand = %#v, want the assignment", cast.X)
	}
	neg, ok := values[1].(*model.Unary)
	if !ok || neg.Op != "-" {
		t.Fatalf("`-$n = 3` = %#v", values[1])
	}
	if assign, ok := neg.X.(*model.AssignExpr); !ok || assign.Op != "=" {
		t.Errorf("unary operand = %#v, want the assignment", neg.X)
	}
	if assign, ok := values[2].(*model.AssignExpr); !ok || assign.Op != "=" {
		t.Errorf("`@$s = 'b'` = %#v, want the assignment", values[2])
	}
	clone, ok := values[3].(*model.Call)
	if !ok || clone.Name != "__phpx_clone" || len(clone.Args) != 1 {
		t.Fatalf("`clone $o = new C` = %#v", values[3])
	}
	if assign, ok := clone.Args[0].(*model.AssignExpr); !ok || assign.Op != "=" {
		t.Errorf("clone operand = %#v, want the assignment", clone.Args[0])
	}
}
