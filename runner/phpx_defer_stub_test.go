//go:build !phpx

package runner

import (
	"errors"
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxDeferUntagged pins that the untagged build defers nothing.
func TestPhpxDeferUntagged(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	cd := &model.ClassDecl{Name: "A"}
	if rt.phpxDeferClass(cd, errors.New("x"), nil, nil, nil, "") {
		t.Error("untagged build deferred a class")
	}
	if err := rt.phpxDeclareDeferred(cd); err != nil {
		t.Error(err)
	}
}
