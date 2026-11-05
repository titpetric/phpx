package expr

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxCopyElementCopiesStorageReads(t *testing.T) {
	arr := model.NewArray()
	arr.Append(1)
	read := func(*Env) (any, error) { return arr, nil }

	got, _ := phpxCopyElement(&model.Var{Name: "a"}, read)(nil)
	if got.(*model.Array) == arr {
		t.Fatal("a variable's array was shared into the literal")
	}
	got, _ = phpxCopyElement(&model.Lit{Value: 1}, read)(nil)
	if got.(*model.Array) != arr {
		t.Fatal("a literal element was copied")
	}
}
