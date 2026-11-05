//go:build phpx

package phpx

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestMultisortOrdersByFirstThenSecond(t *testing.T) {
	a, b := model.NewArray(), model.NewArray()
	for _, v := range []int64{3, 1, 3, 2} {
		a.Append(v)
	}
	for _, v := range []string{"x", "y", "a", "z"} {
		b.Append(v)
	}
	if ok, err := multisort([]any{a, int64(sortAsc), b}); !ok || err != nil {
		t.Fatalf("multisort = %v, %v", ok, err)
	}
	if v, _ := a.Get(int64(0)); v != int64(1) {
		t.Fatalf("first = %v", v)
	}
	if v, _ := b.Get(int64(2)); v != "a" {
		t.Fatalf("the second array follows the first: %v", v)
	}
}
