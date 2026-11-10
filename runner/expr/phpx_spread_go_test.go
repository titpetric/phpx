package expr

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestGoPairs(t *testing.T) {
	a, ok := goPairs(map[string]any{"b": 2, "a": 1})
	if !ok || a.Len() != 2 {
		t.Fatalf("map: %v %v", a, ok)
	}
	var keys []any
	a.Range(func(k, _ any) bool { keys = append(keys, k); return true })
	if keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("order: %v", keys)
	}
	s, ok := goPairs([]any{1, 2, 3})
	if !ok || s.Len() != 3 {
		t.Fatalf("slice: %v %v", s, ok)
	}
	if _, ok := goPairs(model.NewArray()); ok {
		t.Fatal("array is not a go collection")
	}
	if _, ok := goPairs(map[int]any{1: 1}); ok {
		t.Fatal("int-keyed map")
	}
}
