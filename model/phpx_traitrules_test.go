package model

import "testing"

func TestTraitRuleIsInsteadof(t *testing.T) {
	t.Parallel()
	if !(TraitRule{Trait: "A", Method: "f", Insteadof: []string{"B"}}).IsInsteadof() {
		t.Error("a rule naming a loser is an insteadof rule")
	}
	if (TraitRule{Method: "f", Alias: "g"}).IsInsteadof() {
		t.Error("an alias rule is not an insteadof rule")
	}
}
