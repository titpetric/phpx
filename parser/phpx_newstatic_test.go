//go:build phpx

package parser

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestNewStaticProperty pins that `new static::$cls()` builds the class its static property names.
func TestNewStaticProperty(t *testing.T) {
	t.Parallel()
	prog, err := Parse("<?php $o = new static::$cls(1); $p = new A::$b;")
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []string{"cls", "b"} {
		n := prog.Stmts[i].(*model.Assign).Value.(*model.New)
		prop, ok := n.ClassExpr.(*model.StaticProp)
		if !ok || prop.Name != want {
			t.Errorf("statement %d: ClassExpr = %#v", i, n.ClassExpr)
		}
	}
}
