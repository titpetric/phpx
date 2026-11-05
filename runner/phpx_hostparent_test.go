//go:build phpx

package runner

import (
	"bytes"
	"io"
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxHostInheritWithoutHostAncestor(t *testing.T) {
	rt := New(io.Discard, Options{})
	c := &model.Class{Name: "Plain", Ancestors: []string{"NotAHostClass"}}
	obj := model.NewObject(c)
	if err := rt.phpxHostInherit(obj, c, nil, rt.newScope()); err != nil {
		t.Fatal(err)
	}
	if _, ok := phpxHostValue(obj); ok {
		t.Fatal("a class with no host ancestor got a host value")
	}
}

func TestPhpxExtendsHost(t *testing.T) {
	c := &model.Class{Name: "Child", Ancestors: []string{"DateTime"}}
	if !phpxExtendsHost(c, "datetime") || phpxExtendsHost(c, "ArrayObject") {
		t.Fatal("ancestors are compared without regard to case")
	}
}

func TestPhpxHostHasMethod(t *testing.T) {
	if !phpxHostHasMethod(&bytes.Buffer{}, "writestring") || phpxHostHasMethod(&bytes.Buffer{}, "nope") {
		t.Fatal("methods are found without regard to case")
	}
}

func TestPhpxHostArgOfPlainObject(t *testing.T) {
	if _, ok := phpxHostArg(model.NewObject(&model.Class{Name: "X"})); ok {
		t.Fatal("an object with no host value has no host argument")
	}
	if _, ok := phpxHostArg("str"); ok {
		t.Fatal("a string is no object")
	}
}

func TestPhpxHostClassConstOfOtherClass(t *testing.T) {
	rt := New(io.Discard, Options{})
	if _, ok := rt.phpxHostClassConst("ArrayObject", "ATOM", &scopeRef{scope: rt.newScope()}); ok {
		t.Fatal("ArrayObject has no date constants")
	}
}

func TestPhpxUnwrapHostLeavesOthers(t *testing.T) {
	if PhpxUnwrapHost("x") != "x" {
		t.Fatal("a string is not unwrapped")
	}
}
