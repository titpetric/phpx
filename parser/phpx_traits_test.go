package parser

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxTraitParses holds the syntax half of the trait reversal down to the
// AST: a trait body carries the members a class body does, a nested `use`
// line lands in TraitDecl.Traits, and a class records the traits it names.
// What the members mean at runtime is the fork's hoist; phpx/testdata/traits
// settles that against php.
func TestPhpxTraitParses(t *testing.T) {
	t.Parallel()
	prog, err := Parse("<?php\ntrait Named {\n use Loud;\n public $name = \"x\";\n public function labels() { return 1; }\n}\nclass Box {\n use Named, Other;\n}\n")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	var trait *model.TraitDecl
	var class *model.ClassDecl
	for _, s := range prog.Stmts {
		switch n := s.(type) {
		case *model.TraitDecl:
			trait = n
		case *model.ClassDecl:
			class = n
		}
	}
	if trait == nil {
		t.Fatal("the trait declaration was not recorded")
	}
	if len(trait.Traits) != 1 || trait.Traits[0] != "Loud" {
		t.Errorf("nested traits = %v, want [Loud]", trait.Traits)
	}
	if len(trait.Fields) != 1 || trait.Fields[0].Name != "name" {
		t.Errorf("fields = %v, want the one property", trait.Fields)
	}
	if len(trait.Methods) != 1 || trait.Methods[0].Name != "labels" {
		t.Errorf("methods = %v, want labels", trait.Methods)
	}
	if class == nil {
		t.Fatal("the class declaration was not recorded")
	}
	if len(class.Traits) != 2 || class.Traits[0] != "Named" || class.Traits[1] != "Other" {
		t.Errorf("class traits = %v, want [Named Other]", class.Traits)
	}
}

// TestPhpxTraitRulesParse holds the adaptation block to the AST: an insteadof
// rule names its winner and losers, and an `as` rule carries a visibility, an
// alias, or both.
func TestPhpxTraitRulesParse(t *testing.T) {
	t.Parallel()
	prog, err := Parse("<?php\nclass C {\n use A, B {\n A::f insteadof B;\n B::f as protected g;\n h as private;\n }\n}\n")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	rules := prog.Stmts[0].(*model.ClassDecl).TraitRules
	if len(rules) != 3 {
		t.Fatalf("rules = %+v, want 3", rules)
	}
	if !rules[0].IsInsteadof() || rules[0].Trait != "A" || rules[0].Insteadof[0] != "B" {
		t.Errorf("insteadof rule = %+v", rules[0])
	}
	if rules[1].Trait != "B" || rules[1].Alias != "g" || rules[1].Visibility != "protected" {
		t.Errorf("alias rule = %+v", rules[1])
	}
	if rules[2].Trait != "" || rules[2].Method != "h" || rules[2].Visibility != "private" || rules[2].Alias != "" {
		t.Errorf("visibility rule = %+v", rules[2])
	}
	if _, err := Parse("<?php\nclass C { use A { f insteadof B; } }\n"); err == nil {
		t.Error("insteadof without a trait name parsed")
	}
}
