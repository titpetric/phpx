package expr

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxRefArgDeclinesWithoutFlag(t *testing.T) {
	old := model.PhpxAnyByRef.Load()
	model.PhpxAnyByRef.Store(false)
	defer model.PhpxAnyByRef.Store(old)
	dc := &directCompiler{}
	if _, ok := dc.phpxRefArg(&model.Var{Name: "x"}); ok {
		t.Fatal("wrapped an argument with no by-reference parameter declared")
	}
}
