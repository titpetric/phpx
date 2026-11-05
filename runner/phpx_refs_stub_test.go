//go:build !phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxRefAssignUntagged pins that the untagged build binds nothing by reference.
func TestPhpxRefAssignUntagged(t *testing.T) {
	t.Parallel()
	if _, ok := phpxIsRefAssign("=", &model.Ref{X: &model.Var{Name: "a"}}); ok {
		t.Error("untagged build read a reference assignment")
	}
}
