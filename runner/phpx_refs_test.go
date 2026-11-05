//go:build phpx

package runner

import (
	"io"
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxIsRefAssign pins which assignments bind by reference.
func TestPhpxIsRefAssign(t *testing.T) {
	t.Parallel()
	ref := &model.Ref{X: &model.Var{Name: "a"}}
	if _, ok := phpxIsRefAssign("=", ref); !ok {
		t.Error("=& not read")
	}
	if _, ok := phpxIsRefAssign(".=", ref); ok {
		t.Error(".=& read as a reference")
	}
	if _, ok := phpxIsRefAssign("=", &model.Var{Name: "a"}); ok {
		t.Error("a plain assignment read as a reference")
	}
}

func TestPhpxElementCellAppendsAndWritesBack(t *testing.T) {
	rt := New(io.Discard, Options{})
	scope := rt.newScope()
	arr := model.NewArray()
	scope.Set("a", arr)
	cell, err := rt.phpxElementCell(&model.Index{Base: &model.Var{Name: "a"}}, scope)
	if err != nil || cell == nil {
		t.Fatalf("cell %v, err %v", cell, err)
	}
	cell.Set("x")
	if v, _ := arr.Get(int64(0)); v != "x" || arr.Len() != 1 {
		t.Fatalf("element not written back: %v len %d", v, arr.Len())
	}
}

func TestPhpxInstallTraverseAddsHelper(t *testing.T) {
	rt := New(io.Discard, Options{})
	env := map[string]any{}
	rt.phpxInstallTraverse(env)
	if env["__phpx_traverse"] == nil {
		t.Fatal("helper missing")
	}
}
