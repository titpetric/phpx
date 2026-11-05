//go:build phpx

package phpx

import (
	"io"
	"testing"

	"github.com/titpetric/phpx/runner"
)

func TestGeneratedConstantsDoNotOverride(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	rt.SetConst("PREG_NO_ERROR", int64(99))
	registerGeneratedConstants(rt)
	if v, _ := rt.Const("PREG_NO_ERROR"); v != int64(99) {
		t.Fatalf("a defined constant was replaced: %v", v)
	}
	if v, ok := rt.Const("MB_CASE_TITLE"); !ok || v != int64(2) {
		t.Fatalf("MB_CASE_TITLE = %v, %v", v, ok)
	}
}
