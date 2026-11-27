//go:build !phpx

package runner

import (
	"testing"
)

// TestPhpxMethodMemoUnusedUntagged pins that upstream's build carries no memo:
// the field exists so runtime.go compiles in both, and nothing fills it.
func TestPhpxMethodMemoUnusedUntagged(t *testing.T) {
	t.Parallel()
	rt := New(nil, Options{})
	if rt.phpxMethods != nil {
		t.Fatal("an untagged runtime fills no method memo")
	}
}
