//go:build !phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxAbstractConfersNothingUntagged pins upstream's answer in the tree an
// untagged build compiles: an abstract class instantiates and a method with no
// body runs as an empty body. The fork's build is phpx_abstract_test.go.
func TestPhpxAbstractConfersNothingUntagged(t *testing.T) {
	t.Parallel()
	class := &model.Class{Name: "A", Abstract: true}
	if err := phpxInstantiable(class); err != nil {
		t.Errorf("phpxInstantiable = %v, want nil in an untagged build", err)
	}
	decl := &model.FuncDecl{Name: "f", Abstract: true}
	rt := &Runtime{}
	if err := rt.phpxCallAbstract(class, decl); err != nil {
		t.Errorf("phpxCallAbstract = %v, want nil in an untagged build", err)
	}
	if got := rt.phpxDeclOwner(class, decl); got != "A" {
		t.Errorf("phpxDeclOwner = %q, want the instance's class", got)
	}
}
