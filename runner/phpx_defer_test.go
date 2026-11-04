//go:build phpx

package runner

import (
	"errors"
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxDeferOnlyMissingParents pins that only the missing-parent fatal sets
// a class aside, and that declaring a class never set aside is a no-op.
func TestPhpxDeferOnlyMissingParents(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	cd := &model.ClassDecl{Name: "A"}
	if rt.phpxDeferClass(cd, errors.New("other"), nil, nil, nil, "") {
		t.Error("deferred on an unrelated error")
	}
	if !rt.phpxDeferClass(cd, &phpxParentMissing{Child: "A", Parent: "B"}, nil, nil, nil, "f.php") {
		t.Error("did not defer on a missing parent")
	}
	if err := rt.phpxDeclareDeferred(&model.ClassDecl{Name: "Z"}); err != nil {
		t.Errorf("declaring an unknown class = %v", err)
	}
}
