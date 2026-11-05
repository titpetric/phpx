//go:build phpx

package runner

import (
	"io"
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxBindRefArgsPlainWithoutRefs(t *testing.T) {
	in := []any{1, 2}
	out := phpxBindRefArgs(&model.FuncDecl{}, in, &Scope{})
	if len(out) != 2 || out[0] != 1 {
		t.Fatalf("got %v", out)
	}
}

func TestPhpxInvokeArgsStripsForBuiltinName(t *testing.T) {
	old := model.PhpxAnyByRef.Load()
	model.PhpxAnyByRef.Store(true)
	defer model.PhpxAnyByRef.Store(old)
	rt := New(io.Discard, Options{})
	in := []any{&model.RefArg{Value: 7}}
	if got := rt.phpxInvokeArgs("is_array", in); got[0] != 7 {
		t.Fatalf("builtin name kept the wrapper: %v", got)
	}
	if got := rt.phpxInvokeArgs(&Callable{}, in); got[0] == 7 {
		t.Fatal("a closure lost the wrapper")
	}
}
