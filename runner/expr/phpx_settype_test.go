package expr

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxSettypeArgsLeavesOtherCalls(t *testing.T) {
	dc := &directCompiler{}
	args := []closure{nil, nil}
	got, err := dc.phpxSettypeArgs(&model.Call{Name: "strlen"}, args)
	if err != nil || len(got) != 2 {
		t.Fatalf("got %d args, err %v", len(got), err)
	}
}
