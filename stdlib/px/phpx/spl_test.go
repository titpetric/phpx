//go:build phpx

package phpx

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestCountRecursive pins php's COUNT_RECURSIVE: every element at every depth,
// which is the difference laravel's MessageBag::count() reads.
func TestCountRecursive(t *testing.T) {
	inner := model.NewArray()
	inner.Append("required")
	inner.Append("email")
	bag := model.NewArray()
	bag.Set("name", inner)
	bag.Set("password", "x")
	if got := countRecursive(bag); got != 4 {
		t.Errorf("countRecursive = %d, want 4", got)
	}
	if got := countRecursive("scalar"); got != 0 {
		t.Errorf("countRecursive(scalar) = %d, want 0", got)
	}
	if got := countRecursive(model.NewArray()); got != 0 {
		t.Errorf("countRecursive(empty) = %d, want 0", got)
	}
	if !countModeRecursive(int64(1)) || countModeRecursive(int64(0)) || countModeRecursive(false) {
		t.Error("countModeRecursive does not read COUNT_RECURSIVE")
	}
}
