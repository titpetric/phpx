package parser

import "testing"

func TestHaltCompilerStopsParsing(t *testing.T) {
	src := "<?php\necho 1;\n__halt_compiler();\n@@ data {{ not php\n"
	prog, err := Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(prog.Stmts) != 1 {
		t.Fatalf("want the echo alone, got %d statements", len(prog.Stmts))
	}
}
