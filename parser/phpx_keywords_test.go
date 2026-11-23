package parser

import (
	"testing"
)

// TestPhpxKeywordCase pins that php's keywords are read without regard to
// case: a statement, a class body's `use`, `new`, `::class` and the three
// class words all parse in php's spelling. akaunting writes `Use Documents;`
// and the parser read it as an expression.
func TestPhpxKeywordCase(t *testing.T) {
	for _, src := range []string{
		`<?php CLASS A { USE T; }`,
		`<?php $x = NEW A();`,
		`<?php $n = $o::CLASS;`,
		`<?php $n = A::CLASS;`,
		`<?php IF ($x) { ECHO "y"; } ELSE { ECHO "n"; }`,
		`<?php FOREACH ([1] AS $v) { RETURN; }`,
		`<?php ABSTRACT CLASS B EXTENDS A IMPLEMENTS I { PUBLIC STATIC FUNCTION f() { RETURN STATIC::CLASS; } }`,
		`<?php ENUM Suit: string { CASE Hearts = 'H'; }`,
		`<?php $a = TRUE; $b = NULL; $c = ARRAY(1);`,
	} {
		if _, err := ParseFile("t.php", src); err != nil {
			t.Errorf("%s: %v", src, err)
		}
	}
}

// TestPhpxKeywordCaseKeepsGoWords pins that phpscript's own keywords still
// compare exactly: a php program may declare and call a function named `Func`.
func TestPhpxKeywordCaseKeepsGoWords(t *testing.T) {
	if _, err := ParseFile("t.php", `<?php function func() { return 1; } echo Func();`); err != nil {
		t.Fatal(err)
	}
}
