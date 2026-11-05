//go:build phpx

package runner

import (
	"github.com/titpetric/phpx/model"
	"io"
	"strings"
	"testing"
)

// TestPhpxContractAcceptsInheritedMethods covers php's rule that `implements`
// is answered by the finished class: a method the parent carries satisfies the
// interface, and the program runs. The contract used to be read from the
// class's own AST alone, which refused this program.
func TestPhpxContractAcceptsInheritedMethods(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	rt := New(&out, Options{})
	program, err := rt.Load(`<?php
interface Speaker { public function speak(); }
class Base { public function speak() { return "base"; } }
class Child extends Base implements Speaker {}
echo (new Child)->speak(), "\n";
`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := rt.Run(program); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got, want := out.String(), "base\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

// TestPhpxContractAcceptsTraitMethods covers the other contributor: a trait a
// class uses supplies the interface's method, and the class is not missing it.
func TestPhpxContractAcceptsTraitMethods(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	rt := New(&out, Options{})
	program, err := rt.Load(`<?php
interface Speaker { public function speak(); }
trait Talk { public function speak() { return "trait"; } }
class Child implements Speaker { use Talk; }
echo (new Child)->speak(), "\n";
`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := rt.Run(program); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got, want := out.String(), "trait\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

// TestPhpxContractRefusesAMissingMethod keeps the fatal for the case the
// contract exists for: a method nothing contributes is still named.
func TestPhpxContractRefusesAMissingMethod(t *testing.T) {
	t.Parallel()
	rt := New(io.Discard, Options{})
	program, err := rt.Load(`<?php
interface Speaker { public function speak(); }
class Child implements Speaker {}
`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	err = rt.Run(program)
	if err == nil || !strings.Contains(err.Error(), "class Child does not declare method speak() required by interface Speaker") {
		t.Errorf("error = %v, want the contract message", err)
	}
}

// TestPhpxInterfaceConstantsFromAnotherFile covers the framework layout: the
// interface is loaded before the class file, one interface per file, and the
// class reads the constant across the file boundary. The runtime's registered
// declarations answer a name the class's own file does not declare.
func TestPhpxInterfaceConstantsFromAnotherFile(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	rt := New(&out, Options{})

	contract, err := rt.Load(`<?php interface Quota { const LIMIT = 5; }`)
	if err != nil {
		t.Fatalf("Load interface: %v", err)
	}
	if err := rt.Run(contract); err != nil {
		t.Fatalf("Run interface: %v", err)
	}

	class, err := rt.Load(`<?php
class Store implements Quota {}
echo Store::LIMIT, "\n";
`)
	if err != nil {
		t.Fatalf("Load class: %v", err)
	}
	if err := rt.Run(class); err != nil {
		t.Fatalf("Run class: %v", err)
	}
	if got, want := out.String(), "5\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

// TestPhpxInterfaceConstantsReachTheClass covers php's answer for
// `Class::CONST` when the constant is an interface's: the class that names the
// interface reads it, an interface that extends another reads the inherited
// one through its own name, and a class constant of the same name wins.
func TestPhpxInterfaceConstantsReachTheClass(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	rt := New(&out, Options{})
	program, err := rt.Load(`<?php
interface Base { const X = 1; }
interface Wide extends Base { const Y = 2; }
class Child implements Wide { const X = 3; }
echo Wide::X, " ", Child::X, " ", Child::Y, "\n";
`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := rt.Run(program); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got, want := out.String(), "1 3 2\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestPhpxInterfaceConstTreeWithoutExtends(t *testing.T) {
	rt := New(io.Discard, Options{})
	id := &model.InterfaceDecl{Name: "I", Consts: []model.Field{{Name: "A"}}}
	if got := rt.phpxInterfaceConstTree(id, nil); len(got) != 1 {
		t.Fatalf("got %d constants", len(got))
	}
}
