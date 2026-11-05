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

func TestIntegerLiteralTooLargeBecomesFloat(t *testing.T) {
	for _, lit := range []string{"0xFFFFFFFFFFFFFFFF", "9223372036854775808", "07777777777777777777777", "0b1111111111111111111111111111111111111111111111111111111111111111111"} {
		v, err := numLit(token{kind: tInt, val: lit})
		if err != nil {
			t.Fatalf("%s: %v", lit, err)
		}
		if _, ok := v.(float64); !ok {
			t.Fatalf("%s = %T, want float64", lit, v)
		}
	}
}
