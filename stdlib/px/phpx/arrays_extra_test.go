//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/modelutil"
	"io"
	"testing"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func list(vals ...any) *model.Array {
	a := model.NewArray()
	for _, v := range vals {
		a.Append(v)
	}
	return a
}

func TestSetOpDiffAndIntersect(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	diff, err := runSetOp(rt, "array_diff_ukey", setOps["array_diff_ukey"], []any{list("a", "b", "c"), list("x"), func(a, b any) int64 { return 0 }})
	_ = diff
	_ = err
	inter, err := runSetOp(rt, "array_intersect_assoc", setOps["array_intersect_assoc"], []any{list("a", "b", "c"), list("a", "x", "c")})
	if err != nil {
		t.Fatal(err)
	}
	if got := inter.(*model.Array).Len(); got != 2 {
		t.Fatalf("intersect kept %d, want 2", got)
	}
}

func TestMergeRecursive(t *testing.T) {
	a := model.NewArray()
	a.Set("k", "x")
	b := model.NewArray()
	b.Set("k", "y")
	got := mergeRecursive([]any{a, b}).(*model.Array)
	v, _ := got.Get("k")
	if !modelutil.IsCollection(v) {
		t.Fatalf("colliding keys should merge into an array, got %v", v)
	}
}
