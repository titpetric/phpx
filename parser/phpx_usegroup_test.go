package parser

import "testing"

func TestPhpxUsegroupParses(t *testing.T) {
	if _, err := ParseFile("t.php", `<?php namespace N; use A\B\{C, D as E, F\G}; use function A\{h};`); err != nil {
		t.Fatal(err)
	}
}
