//go:build !phpx

package runner

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxClassConfersNothingUntagged pins upstream's answer in the tree an
// untagged build compiles: `extends` records a name and the child arrives with
// the members it declared. The fork's build is phpx_inheritance_test.go.
func TestPhpxClassConfersNothingUntagged(t *testing.T) {
	t.Parallel()
	parent := &model.ClassDecl{Name: "Animal", Methods: []*model.FuncDecl{{Name: "speak"}}}
	child := &model.ClassDecl{Name: "Dog", Parent: "Animal", Methods: []*model.FuncDecl{{Name: "fetch"}}}
	decls := map[string]*model.ClassDecl{"Animal": parent, "Dog": child}

	rt := &Runtime{}
	class, err := rt.phpxClass(child, decls, nil)
	if err != nil {
		t.Fatalf("phpxClass: %v", err)
	}
	if len(class.Ancestors) != 0 {
		t.Errorf("Ancestors = %v, want none", class.Ancestors)
	}
	if _, ok := class.Methods["speak"]; ok {
		t.Error("the parent's method arrived in an untagged build")
	}
	if phpxCarries(class, &model.Class{Name: "Animal"}) {
		t.Error("an untagged build forwarded $this to a parent")
	}
	if err := phpxNoParentError(); !strings.Contains(err.Error(), "no inheritance at runtime") {
		t.Errorf("parent error = %v, want upstream's wording", err)
	}
}
