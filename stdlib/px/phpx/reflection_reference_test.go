//go:build phpx

package phpx

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestReflectionReferenceFromArrayElement(t *testing.T) {
	values := model.NewArray()
	values.Set("k", "v")
	values.Append("first")
	if got := normalizeArrayKey(true); got != int64(1) {
		t.Fatalf("bool key = %v", got)
	}
	if got := normalizeArrayKey(2.9); got != int64(2) {
		t.Fatalf("float key = %v", got)
	}
	if _, found := values.Get(normalizeArrayKey(2.9)); found {
		t.Fatal("a float key reached an absent element")
	}
	if _, found := values.Get(normalizeArrayKey("k")); !found {
		t.Fatal("a string key missed the element")
	}
	if got := (&reflectionError{"x"}).ThrowableClass(); got != "ReflectionException" {
		t.Fatalf("class = %q", got)
	}
}
