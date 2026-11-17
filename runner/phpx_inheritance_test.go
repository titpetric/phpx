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
	class, err := rt.phpxClass(child, decls, nil, nil)
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
	class, err := rt.phpxClass(decls["C"], decls, nil, nil)
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
	_, err := rt.phpxClass(&model.ClassDecl{Name: "Dog", Parent: "Missing"}, nil, nil, nil)
	if err == nil {
		t.Fatal("a missing parent was accepted")
	}
}

// TestPhpxClassRefusesRedeclaration covers the fatal php raises for a class
// name declared twice, and for the name stdClass.
func TestPhpxClassRefusesRedeclaration(t *testing.T) {
	t.Parallel()
	first, second := &model.ClassDecl{Name: "A"}, &model.ClassDecl{Name: "A"}
	rt := &Runtime{}
	if _, err := rt.phpxClass(first, map[string]*model.ClassDecl{"A": second}, nil, nil); err == nil || err.Error() != "Cannot redeclare class A" {
		t.Errorf("duplicate = %v", err)
	}
	std := &model.ClassDecl{Name: "stdClass"}
	if _, err := rt.phpxClass(std, map[string]*model.ClassDecl{"stdClass": std}, nil, nil); err == nil || err.Error() != "Cannot redeclare class stdClass" {
		t.Errorf("stdClass = %v", err)
	}
}

func TestPhpxBuiltinDescends(t *testing.T) {
	cases := []struct {
		class, want string
		ok          bool
	}{
		{"UnexpectedValueException", "RuntimeException", true},
		{"UnexpectedValueException", "\\Exception", true},
		{"BadMethodCallException", "LogicException", true},
		{"DivisionByZeroError", "Error", true},
		{"RuntimeException", "LogicException", false},
		{"Exception", "Exception", false},
	}
	for _, c := range cases {
		if got := phpxBuiltinDescends(c.class, c.want); got != c.ok {
			t.Errorf("%s descends %s = %v, want %v", c.class, c.want, got, c.ok)
		}
	}
}

// TestPhpxCheckRedeclared walks the property redeclaration table php refuses:
// a narrowed visibility, a name that turns static or readonly, and a changed
// type, each with php's message. The fixture beside phpx/testdata/classes
// settles two of them against php itself.
func TestPhpxCheckRedeclared(t *testing.T) {
	parent := &model.Class{
		Name: "P",
		Fields: []model.Field{
			{Name: "pub"},
			{Name: "prot", Visibility: "protected"},
			{Name: "arr", Type: "array"},
			{Name: "ro", Readonly: true},
		},
		Statics: []model.Field{{Name: "st", Visibility: "protected"}},
	}
	for _, tc := range []struct {
		name    string
		own     []model.Field
		statics []model.Field
		want    string
	}{
		{"public narrowed", []model.Field{{Name: "pub", Visibility: "protected"}}, nil,
			"Access level to C::$pub must be public (as in class P)"},
		{"protected narrowed", []model.Field{{Name: "prot", Visibility: "private"}}, nil,
			"Access level to C::$prot must be protected (as in class P) or weaker"},
		{"static to non static", []model.Field{{Name: "st"}}, nil,
			"Cannot redeclare static P::$st as non static C::$st"},
		{"non static to static", nil, []model.Field{{Name: "pub"}},
			"Cannot redeclare non static P::$pub as static C::$pub"},
		{"readonly", []model.Field{{Name: "ro"}}, nil,
			"Cannot redeclare readonly property P::$ro as non-readonly C::$ro"},
		{"type change", []model.Field{{Name: "arr", Type: "int"}}, nil,
			"Type of C::$arr must be array (as in class P)"},
		{"type added", []model.Field{{Name: "pub", Type: "int"}}, nil,
			"Type of C::$pub must not be defined (as in class P)"},
	} {
		err := phpxCheckRedeclared("C", parent, tc.own, tc.statics)
		if err == nil || err.Error() != tc.want {
			t.Errorf("%s: got %v, want %q", tc.name, err, tc.want)
		}
	}
	if err := phpxCheckRedeclared("C", parent, []model.Field{{Name: "arr", Type: "\\array"}}, nil); err != nil {
		t.Errorf("an unchanged type was refused: %v", err)
	}
}
