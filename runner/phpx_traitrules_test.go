//go:build phpx

package runner

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxAdaptTraits holds the adaptation block to its parts: insteadof drops
// the loser, an alias adds a second name, a visibility alone changes the
// method in place, and a rule naming a method or trait that is not there
// answers php's fatal. The behaviour against php is phpx/testdata/traits.
func TestPhpxAdaptTraits(t *testing.T) {
	t.Parallel()
	a := newPhpxTraitBody()
	a.overlay(&model.TraitDecl{Name: "A", Methods: []*model.FuncDecl{{Name: "f", Visibility: "public"}}})
	b := newPhpxTraitBody()
	b.overlay(&model.TraitDecl{Name: "B", Methods: []*model.FuncDecl{{Name: "f"}}})
	rules := []model.TraitRule{
		{Trait: "A", Method: "f", Insteadof: []string{"B"}},
		{Trait: "B", Method: "f", Alias: "g"},
		{Method: "f", Visibility: "protected"},
	}
	if err := phpxAdaptTraits(rules, []string{"A", "B"}, []*phpxTraitBody{a, b}); err != nil {
		t.Fatalf("phpxAdaptTraits: %v", err)
	}
	if _, ok := b.methods["f"]; ok {
		t.Error("B::f survived insteadof")
	}
	if _, ok := b.methods["g"]; !ok {
		t.Error("alias g was not added to B")
	}
	if got := a.methods["f"].Visibility; got != "protected" {
		t.Errorf("visibility of A::f = %q, want protected", got)
	}

	err := phpxAdaptTraits([]model.TraitRule{{Method: "nope", Alias: "x"}}, []string{"A"}, []*phpxTraitBody{a})
	if err == nil || !strings.Contains(err.Error(), "method does not exist") && !strings.Contains(err.Error(), "this method does not exist") {
		t.Errorf("missing method error = %v", err)
	}
	err = phpxAdaptTraits([]model.TraitRule{{Trait: "Z", Method: "f", Alias: "x"}}, []string{"A"}, []*phpxTraitBody{a})
	if err == nil || !strings.Contains(err.Error(), "Required Trait Z") {
		t.Errorf("missing trait error = %v", err)
	}
}
