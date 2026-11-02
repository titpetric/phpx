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
	class, err := rt.phpxClass(child, decls, nil, nil)
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

// TestPhpxClassConfersNoTraitUntagged pins upstream's answer for a trait too:
// the parser records `use` and the untagged build copies no member from it.
func TestPhpxClassConfersNoTraitUntagged(t *testing.T) {
	t.Parallel()
	trait := &model.TraitDecl{Name: "Greets", Methods: []*model.FuncDecl{{Name: "greet"}}}
	class := &model.ClassDecl{Name: "Base", Traits: []string{"Greets"}}
	traits := map[string]*model.TraitDecl{"Greets": trait}

	rt := &Runtime{}
	c, err := rt.phpxClass(class, nil, traits, nil)
	if err != nil {
		t.Fatalf("phpxClass: %v", err)
	}
	if _, ok := c.Methods["greet"]; ok {
		t.Error("a trait's method arrived in an untagged build")
	}
}
