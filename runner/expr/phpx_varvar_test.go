package expr

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxVarVarIgnoresOtherCalls(t *testing.T) {
	dc := &directCompiler{}
	if _, ok, _ := dc.phpxVarVar(&model.Call{Name: "strlen"}); ok {
		t.Fatal("an ordinary call was taken for a variable variable")
	}
}
