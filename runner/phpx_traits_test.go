//go:build phpx

package runner

import (
	"io"
	"strings"
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxClassAppliesTraits holds the trait composition down to its parts:
// the members arrive, the class's own method wins, a nested trait composes,
// and `self::` in a trait method resolves to the class that used it. The
// end-to-end behaviour is settled against php by phpx/testdata/traits; this
// test names the seam.
func TestPhpxClassAppliesTraits(t *testing.T) {
	t.Parallel()
	loud := &model.TraitDecl{
		Name:    "Loud",
		Methods: []*model.FuncDecl{{Name: "shout"}},
	}
	named := &model.TraitDecl{
		Name:    "Named",
		Traits:  []string{"Loud"},
		Fields:  []model.Field{{Name: "label"}},
		Statics: []model.Field{{Name: "seen"}},
		Consts:  []model.Field{{Name: "KIND"}},
		Methods: []*model.FuncDecl{{Name: "shout"}, {Name: "labels"}},
	}
	owner := &model.ClassDecl{
		Name:   "Owner",
		Traits: []string{"Named"},
		Methods: []*model.FuncDecl{
			{Name: "labels"},
		},
	}
	traits := map[string]*model.TraitDecl{"Loud": loud, "Named": named}

	rt := &Runtime{}
	class, err := rt.phpxClass(owner, nil, traits, nil)
	if err != nil {
		t.Fatalf("phpxClass: %v", err)
	}
	for _, name := range []string{"shout", "labels"} {
		if _, ok := class.Methods[name]; !ok {
			t.Errorf("method %s did not arrive", name)
		}
	}
	if got := class.MethodOwner["shout"]; got != "Owner" {
		t.Errorf("owner of shout = %q, want the using class", got)
	}
	if got := class.MethodOwner["labels"]; got != "Owner" {
		t.Errorf("owner of the override = %q, want Owner", got)
	}
	if len(class.Fields) != 1 || len(class.Statics) != 1 || len(class.Consts) != 1 {
		t.Errorf("members = %v/%v/%v, want one of each", class.Fields, class.Statics, class.Consts)
	}
	if len(class.Ancestors) != 0 {
		t.Errorf("Ancestors = %v, want none: a trait is not a type", class.Ancestors)
	}
}

// TestPhpxClassRejectsTraitCollisions pins php's three composition fatals:
// two traits declaring one method, two traits declaring one property with
// different defaults, and a missing trait.
func TestPhpxClassRejectsTraitCollisions(t *testing.T) {
	t.Parallel()
	first := &model.TraitDecl{Name: "A", Methods: []*model.FuncDecl{{Name: "foo"}}}
	second := &model.TraitDecl{Name: "B", Methods: []*model.FuncDecl{{Name: "foo"}}}
	traits := map[string]*model.TraitDecl{"A": first, "B": second}
	rt := &Runtime{}

	_, err := rt.phpxClass(&model.ClassDecl{Name: "C", Traits: []string{"A", "B"}}, nil, traits, nil)
	if err == nil || !strings.Contains(err.Error(), "Trait method B::foo has not been applied as C::foo, because of collision with A::foo") {
		t.Errorf("method collision = %v, want php's message", err)
	}

	one := &model.TraitDecl{Name: "One", Fields: []model.Field{{Name: "x", Default: &model.Lit{Value: 1}}}}
	two := &model.TraitDecl{Name: "Two", Fields: []model.Field{{Name: "x", Default: &model.Lit{Value: 2}}}}
	fields := map[string]*model.TraitDecl{"One": one, "Two": two}
	_, err = rt.phpxClass(&model.ClassDecl{Name: "Both", Traits: []string{"One", "Two"}}, nil, fields, nil)
	if err == nil || !strings.Contains(err.Error(), "One and Two define the same property ($x) in the composition of Both") {
		t.Errorf("property collision = %v, want php's message", err)
	}

	_, err = rt.phpxClass(&model.ClassDecl{Name: "C", Traits: []string{"Missing"}}, nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), `Trait "Missing" not found`) {
		t.Errorf("missing trait = %v, want php's message", err)
	}
}

func TestPhpxTraitExistsUnknown(t *testing.T) {
	rt := New(io.Discard, Options{})
	if rt.PhpxTraitExists("No\\Such\\Trait", true) {
		t.Fatal("an unknown trait does not exist")
	}
	if rt.PhpxClassUses("No\\Such", true) != false {
		t.Fatal("class_uses of an unknown class is false")
	}
}
