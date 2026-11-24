//go:build phpx

package parser

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestStaticPropNamedByExpression pins the parse: `Class::${expr}` holds the
// name expression, and `Class::$$name` holds the variable it reads, where
// `Class::$name` stays the literal name.
func TestStaticPropNamedByExpression(t *testing.T) {
	t.Parallel()
	prog, err := Parse("<?php $a = A::${$p}; $b = A::$$n; $c = A::$lit;")
	if err != nil {
		t.Fatal(err)
	}
	props := make([]*model.StaticProp, 0, 3)
	for _, stmt := range prog.Stmts {
		props = append(props, stmt.(*model.Assign).Value.(*model.StaticProp))
	}
	if name, ok := props[0].NameExpr.(*model.Var); !ok || name.Name != "p" {
		t.Errorf("`A::${$p}` NameExpr = %#v", props[0].NameExpr)
	}
	if name, ok := props[1].NameExpr.(*model.Var); !ok || name.Name != "n" {
		t.Errorf("`A::$$n` NameExpr = %#v", props[1].NameExpr)
	}
	if props[2].NameExpr != nil || props[2].Name != "lit" {
		t.Errorf("`A::$lit` = %#v", props[2])
	}
}
