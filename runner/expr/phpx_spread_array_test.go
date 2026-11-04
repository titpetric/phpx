package expr

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestSpreadArrayDeclinesALiteralWithoutOne(t *testing.T) {
	t.Parallel()
	dc := &directCompiler{}
	_, ok, err := dc.phpxSpreadArray(&model.ArrayLit{Items: []model.ArrayItem{{Val: &model.Lit{Value: int64(1)}}}})
	if ok || err != nil {
		t.Fatalf("got ok=%v err=%v; a literal with no spread is upstream's", ok, err)
	}
}
