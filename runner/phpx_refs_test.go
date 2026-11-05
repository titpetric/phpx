//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxIsRefAssign pins which assignments bind by reference.
func TestPhpxIsRefAssign(t *testing.T) {
	t.Parallel()
	ref := &model.Ref{X: &model.Var{Name: "a"}}
	if _, ok := phpxIsRefAssign("=", ref); !ok {
		t.Error("=& not read")
	}
	if _, ok := phpxIsRefAssign(".=", ref); ok {
		t.Error(".=& read as a reference")
	}
	if _, ok := phpxIsRefAssign("=", &model.Var{Name: "a"}); ok {
		t.Error("a plain assignment read as a reference")
	}
}
