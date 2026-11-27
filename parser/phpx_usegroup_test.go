package parser

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxUsegroupParses(t *testing.T) {
	if _, err := ParseFile("t.php", `<?php namespace N; use A\B\{C, D as E, F\G}; use function A\{h};`); err != nil {
		t.Fatal(err)
	}
}

// TestPhpxUseFunctionKeepsClass pins the split between the function and class
// import tables: `use function Demo\version;` beside `final class Version`
// left the class qualifying to `Demo\version`, so it was declared under the
// function's name, and Pest stopped on a plugin class that did not exist.
func TestPhpxUseFunctionKeepsClass(t *testing.T) {
	program, err := ParseFile("t.php", `<?php
namespace Demo;
function version() { return 'helper'; }
use function Demo\version;
final class Version { }
version();`)
	if err != nil {
		t.Fatal(err)
	}
	var class *model.ClassDecl
	var call *model.Call
	for _, stmt := range program.Stmts {
		switch node := stmt.(type) {
		case *model.ClassDecl:
			class = node
		case *model.ExprStmt:
			call, _ = node.X.(*model.Call)
		}
	}
	if class == nil || class.Name != `Demo\Version` {
		t.Fatalf("class = %+v, want Demo\\Version", class)
	}
	if call == nil || call.Name != `Demo\version` {
		t.Fatalf("call = %+v, want Demo\\version", call)
	}
}
