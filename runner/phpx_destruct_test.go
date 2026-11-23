//go:build phpx

package runner

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/parser"
)

// phpxDestructSource is the table phpx/testdata/classes/destruct.phpt settles
// against php, cut to the cases the Go seam is responsible for. The fixture is
// the claim; this is what names it.
const phpxDestructSource = `<?php
class P {
	public $n;
	function __construct($n) { $this->n = $n; echo " new ", $n; }
	function __destruct() { echo " drop ", $this->n; }
}
class C extends P {}
function makes() { $local = new P("local"); return new P("returned"); }
$kept = makes();
$b = new P("b"); $b = new P("c");
unset($b);
(new P("temp"));
`

// TestPhpxDestructRuns holds the interpreter's answer to the fixture: a frame
// that returns destroys what it held, a write over a variable destroys the old
// value, and a statement whose value nobody keeps destroys it.
func TestPhpxDestructRuns(t *testing.T) {
	t.Parallel()
	prog, err := parser.ParseFileOptions("main.php", phpxDestructSource, parser.Options{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out strings.Builder
	rt := New(&out, Options{})
	rt.UpdateFilename("main.php")
	if err := rt.Run(prog); err != nil {
		t.Fatalf("run: %v", err)
	}
	want := " new local new returned drop local new b new c drop b drop c new temp drop temp" +
		" drop returned"
	if got := out.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

// TestPhpxDestructDeclinesTheEngine pins that the bytecode engine leaves a
// program declaring a destructor to the interpreter, which is where the
// destructor runs.
func TestPhpxDestructDeclinesTheEngine(t *testing.T) {
	t.Parallel()
	prog, err := parser.ParseFileOptions("main.php", phpxDestructSource, parser.Options{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out strings.Builder
	rt := NewFlatStack(&out, Options{})
	rt.UpdateFilename("main.php")
	if err := rt.Run(prog); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := out.String(); !strings.Contains(got, "drop temp") {
		t.Fatalf("output = %q, want the destructor's drops", got)
	}
}
