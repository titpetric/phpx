//go:build phpx

package runner

import (
	"errors"
	"io"
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxThrowableHierarchy(t *testing.T) {
	decls := phpxThrowableDeclarations()
	if len(decls) != len(phpxThrowableParents) {
		t.Fatalf("declared %d of %d built-in throwables", len(decls), len(phpxThrowableParents))
	}
	rt := New(io.Discard, Options{})
	c, ok, err := rt.phpxThrowableParent("\\BadMethodCallException", nil, nil, map[string]bool{})
	if !ok || err != nil {
		t.Fatalf("BadMethodCallException: ok=%v err=%v", ok, err)
	}
	want := []string{"BadFunctionCallException", "LogicException", "Exception"}
	if len(c.Ancestors) != len(want) {
		t.Fatalf("ancestors %v, want %v", c.Ancestors, want)
	}
	for i, name := range want {
		if c.Ancestors[i] != name {
			t.Errorf("ancestor %d is %s, want %s", i, c.Ancestors[i], name)
		}
	}
	if _, ok := rt.classes["LogicException"]; !ok {
		t.Error("an ancestor is registered for parent:: to resolve")
	}
	if _, ok, _ := rt.phpxThrowableParent("PDO", nil, nil, map[string]bool{}); ok {
		t.Error("a host class that is no throwable keeps the host-class answer")
	}
}

func TestPhpxIsErrorClass(t *testing.T) {
	fault := newObjectError(model.NewObject(&model.Class{Name: "Fault", Ancestors: []string{"Error"}}))
	if !phpxIsErrorClass("Fault", fault) {
		t.Error("a class extending Error is an Error")
	}
	myErr := newObjectError(model.NewObject(&model.Class{Name: "MyError", Ancestors: []string{"Exception"}}))
	if phpxIsErrorClass("MyError", myErr) {
		t.Error("a class extending Exception is no Error, whatever its name ends in")
	}
	if !phpxIsErrorClass("TypeError", errors.New("x")) {
		t.Error("a host throwable keeps the name suffix rule")
	}
}
