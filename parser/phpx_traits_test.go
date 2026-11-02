package parser

import (
	"strings"
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

// TestPhpxTraitAdaptationRefused pins the limit the fork documents: the
// `insteadof`/`as` block is not read, and a source carrying one says so
// instead of losing it silently.
func TestPhpxTraitAdaptationRefused(t *testing.T) {
	t.Parallel()
	_, err := Parse("<?php\ntrait T {}\nclass C {\n use T { T::f insteadof T; }\n}\n")
	if err == nil || !strings.Contains(err.Error(), "trait adaptations") {
		t.Errorf("error = %v, want the adaptation refusal", err)
	}
}
