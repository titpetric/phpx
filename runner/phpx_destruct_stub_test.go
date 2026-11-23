//go:build !phpx

package runner

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/parser"
)

// TestPhpxDestructRunsNothingUntagged pins upstream's answer in the tree an
// untagged build compiles: `__destruct` is an ordinary method nothing calls,
// so a program that drops every reference prints no drop line. The fork's
// build is phpx_destruct_test.go.
func TestPhpxDestructRunsNothingUntagged(t *testing.T) {
	t.Parallel()
	src := `<?php
class P {
	public $n;
	function __construct($n) { $this->n = $n; echo " new ", $n; }
	function __destruct() { echo " drop ", $this->n; }
}
$a = new P("a");
unset($a);
(new P("temp"));
`
	prog, err := parser.ParseFileOptions("main.php", src, parser.Options{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out strings.Builder
	rt := New(&out, Options{})
	rt.UpdateFilename("main.php")
	if err := rt.Run(prog); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got, want := out.String(), " new a new temp"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
