//go:build phpx

package runner

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxClassResolvesAbstract holds the abstract bookkeeping down to its
// parts: an abstract class keeps the methods it leaves to a child, in
// declaration order, and a child that implements every one is concrete. The
// end-to-end behaviour is settled against php by phpx/testdata/abstract; this
// test names the seam.
func TestPhpxClassResolvesAbstract(t *testing.T) {
	t.Parallel()
	base := &model.ClassDecl{
		Name:     "Shape",
		Abstract: true,
		Methods: []*model.FuncDecl{
			{Name: "area", Abstract: true},
			{Name: "name", Abstract: true},
			{Name: "describe"},
		},
	}
	child := &model.ClassDecl{
		Name:   "Square",
		Parent: "Shape",
		Methods: []*model.FuncDecl{
			{Name: "area"},
			{Name: "name"},
		},
	}
	decls := map[string]*model.ClassDecl{"Shape": base, "Square": child}

	rt := &Runtime{}
	class, err := rt.phpxClass(base, decls, nil, nil)
	if err != nil {
		t.Fatalf("phpxClass(Shape): %v", err)
	}
	if !class.Abstract {
		t.Error("Shape is not abstract")
	}
	if got := class.AbstractMethods; len(got) != 2 || got[0] != "Shape::area" || got[1] != "Shape::name" {
		t.Errorf("AbstractMethods = %v, want [Shape::area Shape::name]", got)
	}

	concrete, err := rt.phpxClass(child, decls, nil, nil)
	if err != nil {
		t.Fatalf("phpxClass(Square): %v", err)
	}
	if concrete.Abstract || len(concrete.AbstractMethods) != 0 {
		t.Errorf("Square = abstract %v with %v, want concrete", concrete.Abstract, concrete.AbstractMethods)
	}
}

// TestPhpxClassRejectsUnimplemented holds php's fatal for a concrete class
// that leaves a method without a body, and the order the message lists them
// in: the class's own declarations first, then the parent's.
func TestPhpxClassRejectsUnimplemented(t *testing.T) {
	t.Parallel()
	decls := map[string]*model.ClassDecl{
		"A": {Name: "A", Abstract: true, Methods: []*model.FuncDecl{
			{Name: "a1", Abstract: true},
			{Name: "a2", Abstract: true},
		}},
		"B": {Name: "B", Parent: "A", Abstract: true, Methods: []*model.FuncDecl{
			{Name: "b1", Abstract: true},
			{Name: "b2", Abstract: true},
		}},
		"C": {Name: "C", Parent: "B", Methods: []*model.FuncDecl{
			{Name: "a2"},
			{Name: "b1"},
		}},
	}

	rt := &Runtime{}
	_, err := rt.phpxClass(decls["C"], decls, nil, nil)
	if err == nil {
		t.Fatal("phpxClass(C) returned no error")
	}
	want := "Class C contains 2 abstract methods and must therefore be declared abstract or implement the remaining methods (B::b2, A::a1)"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
}

// TestPhpxClassRejectsAbstractRedeclaration covers php's refusal to turn a
// method the parent implements back into a declaration without a body.
func TestPhpxClassRejectsAbstractRedeclaration(t *testing.T) {
	t.Parallel()
	decls := map[string]*model.ClassDecl{
		"P": {Name: "P", Methods: []*model.FuncDecl{{Name: "f"}}},
		"C": {Name: "C", Parent: "P", Abstract: true, Methods: []*model.FuncDecl{{Name: "f", Abstract: true}}},
	}

	rt := &Runtime{}
	_, err := rt.phpxClass(decls["C"], decls, nil, nil)
	if err == nil {
		t.Fatal("phpxClass(C) returned no error")
	}
	want := "Cannot make non abstract method P::f() abstract in class C"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
}

// TestPhpxClassTraitAbstractSatisfiedByParent covers the other direction: a
// trait's abstract method is satisfied by the concrete method a parent
// implements, and the parent's body is the one that runs.
func TestPhpxClassTraitAbstractSatisfiedByParent(t *testing.T) {
	t.Parallel()
	decls := map[string]*model.ClassDecl{
		"P": {Name: "P", Methods: []*model.FuncDecl{{Name: "f"}}},
		"C": {Name: "C", Parent: "P", Traits: []string{"T"}},
	}
	traits := map[string]*model.TraitDecl{
		"T": {Name: "T", Methods: []*model.FuncDecl{{Name: "f", Abstract: true}}},
	}

	rt := &Runtime{}
	class, err := rt.phpxClass(decls["C"], decls, traits, nil)
	if err != nil {
		t.Fatalf("phpxClass(C): %v", err)
	}
	if class.Abstract || len(class.AbstractMethods) != 0 {
		t.Errorf("C = abstract %v with %v, want concrete", class.Abstract, class.AbstractMethods)
	}
	if got := class.MethodOwner["f"]; got != "P" {
		t.Errorf("owner of f = %q, want P", got)
	}
}

// TestPhpxAbstractRefusals covers the two messages php raises at run time: a
// call to a body that does not exist, and `new` on a class that leaves one.
func TestPhpxAbstractRefusals(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	decl := &model.FuncDecl{Name: "f", Abstract: true}
	if err := rt.phpxCallAbstract(&model.Class{Name: "A"}, decl); err == nil || err.Error() != "Cannot call abstract method A::f()" {
		t.Errorf("phpxCallAbstract = %v, want the php message", err)
	}
	if err := rt.phpxCallAbstract(&model.Class{Name: "A"}, &model.FuncDecl{Name: "f"}); err != nil {
		t.Errorf("phpxCallAbstract on a body = %v, want nil", err)
	}
	if err := phpxInstantiable(&model.Class{Name: "A", Abstract: true}); err == nil || !strings.Contains(err.Error(), "Cannot instantiate abstract class A") {
		t.Errorf("phpxInstantiable = %v, want the php message", err)
	}
	if err := phpxInstantiable(&model.Class{Name: "B"}); err != nil {
		t.Errorf("phpxInstantiable on a concrete class = %v, want nil", err)
	}
}

// TestPhpxClassRejectsPrivateAbstract covers the declaration php refuses
// before it counts the class: a method with no body may not be private. A
// trait may declare one, because the class that uses it supplies the body.
func TestPhpxClassRejectsPrivateAbstract(t *testing.T) {
	t.Parallel()
	class := &model.ClassDecl{
		Name:     "A",
		Abstract: true,
		Methods:  []*model.FuncDecl{{Name: "f", Abstract: true, Visibility: "private"}},
	}

	rt := &Runtime{}
	_, err := rt.phpxClass(class, nil, nil, nil)
	if err == nil {
		t.Fatal("phpxClass(A) returned no error")
	}
	want := "Abstract function A::f() cannot be declared private"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}

	trait := &model.TraitDecl{Name: "T", Methods: []*model.FuncDecl{{Name: "f", Abstract: true, Visibility: "private"}}}
	traits := map[string]*model.TraitDecl{"T": trait}
	user := &model.ClassDecl{Name: "C", Traits: []string{"T"}, Methods: []*model.FuncDecl{{Name: "f", Visibility: "private"}}}
	if _, err := rt.phpxClass(user, nil, traits, nil); err != nil {
		t.Errorf("a trait's private abstract method was refused: %v", err)
	}
}
