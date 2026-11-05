//go:build phpx

package parser

import "testing"

// TestPhpxSyntax pins the three spellings latte's source uses: a by-reference
// method name, `Class::{"name"}()` and a float written `.1`.
func TestPhpxSyntax(t *testing.T) {
	t.Parallel()
	for _, src := range []string{
		"<?php abstract class A { abstract public function &getIterator(): \\Generator; public function &r() { return $this->x; } }",
		"<?php $x = Foo::{\"format{$type}Attribute\"}($a, $b);",
		"<?php $x = (1 / 4 + 1) * 10 + .1; $y = .5e3 + .25;",
	} {
		if _, err := Parse(src); err != nil {
			t.Errorf("%q: %v", src, err)
		}
	}
	if _, err := Parse("<?php $a . 1;"); err != nil {
		t.Errorf("concatenation with a spaced point: %v", err)
	}
}
