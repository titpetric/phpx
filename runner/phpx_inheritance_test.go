//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxClassFlattensTheParent holds the merge down to its parts: the child
// acquires the parent's method, its own declaration of a name wins, and each
// method keeps the class that wrote it. The end-to-end behaviour is settled
// against php by phpx/testdata/inheritance; this test names the seam.
func TestPhpxClassFlattensTheParent(t *testing.T) {
	t.Parallel()
	parent := &model.ClassDecl{
		Name:   "Animal",
		Fields: []model.Field{{Name: "legs"}},
		Consts: []model.Field{{Name: "KIND"}},
		Methods: []*model.FuncDecl{
			{Name: "speak"},
			{Name: "describe"},
		},
	}
	child := &model.ClassDecl{
		Name:   "Dog",
		Parent: "Animal",
		Fields: []model.Field{{Name: "legs"}},
		Methods: []*model.FuncDecl{
			{Name: "speak"},
			{Name: "fetch"},
		},
	}
	decls := map[string]*model.ClassDecl{"Animal": parent, "Dog": child}

	rt := &Runtime{}
	class, err := rt.phpxClass(child, decls, nil)
	if err != nil {
		t.Fatalf("phpxClass: %v", err)
	}
	if got := class.Ancestors; len(got) != 1 || got[0] != "Animal" {
		t.Errorf("Ancestors = %v, want [Animal]", got)
	}
	for _, name := range []string{"speak", "describe", "fetch"} {
		if _, ok := class.Methods[name]; !ok {
			t.Errorf("method %s did not arrive", name)
		}
	}
	if got := class.MethodOwner["describe"]; got != "Animal" {
		t.Errorf("owner of describe = %q, want Animal", got)
	}
	if got := class.MethodOwner["speak"]; got != "Dog" {
		t.Errorf("owner of the override = %q, want Dog", got)
	}
	if len(class.Consts) != 1 || class.Consts[0].Name != "KIND" {
		t.Errorf("consts = %v, want the inherited KIND", class.Consts)
	}
	if len(class.Fields) != 1 {
		t.Errorf("fields = %v, want the one redeclared field", class.Fields)
	}
}

// TestPhpxClassChainIsNearestFirst covers a grandchild: the chain reads the
// direct parent first, which is what `parent::` and get_parent_class read.
func TestPhpxClassChainIsNearestFirst(t *testing.T) {
	t.Parallel()
	decls := map[string]*model.ClassDecl{
		"A": {Name: "A"},
		"B": {Name: "B", Parent: "A"},
		"C": {Name: "C", Parent: "B"},
	}

	rt := &Runtime{}
	class, err := rt.phpxClass(decls["C"], decls, nil)
	if err != nil {
		t.Fatalf("phpxClass: %v", err)
	}
	if len(class.Ancestors) != 2 || class.Ancestors[0] != "B" || class.Ancestors[1] != "A" {
		t.Errorf("Ancestors = %v, want [B A]", class.Ancestors)
	}
	if !phpxCarries(class, &model.Class{Name: "A"}) {
		t.Error("a grandchild does not carry an ancestor")
	}
	if phpxCarries(class, &model.Class{Name: "D"}) {
		t.Error("a class carries one it does not derive from")
	}
}

// TestPhpxClassRejectsAnUnknownParent keeps php's fatal: a parent no
// declaration and no host constructor answers is an error, not an empty class.
func TestPhpxClassRejectsAnUnknownParent(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	_, err := rt.phpxClass(&model.ClassDecl{Name: "Dog", Parent: "Missing"}, nil, nil)
	if err == nil {
		t.Fatal("a missing parent was accepted")
	}
}
