package parser

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxArrowCapturesFreeVariables(t *testing.T) {
	prog, err := ParseFile("t.php", `<?php $f = fn($x) => $x + $n + $this->p + $n;`)
	if err != nil {
		t.Fatal(err)
	}
	cl := prog.Stmts[0].(*model.Assign).Value.(*model.Closure)
	if len(cl.Uses) != 1 || cl.Uses[0].Name != "n" {
		t.Fatalf("uses = %+v, want [n]", cl.Uses)
	}
	if _, ok := cl.Body[0].(*model.Return); !ok {
		t.Fatalf("body = %T, want *model.Return", cl.Body[0])
	}
}

func TestPhpxArrowKeepsFnAsAFunctionName(t *testing.T) {
	for _, src := range []string{`<?php $a = fn();`} {
		if _, err := ParseFile("t.php", src); err != nil {
			t.Errorf("%s: %v", src, err)
		}
	}
}
