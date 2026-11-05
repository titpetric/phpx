package parser

import "testing"

func TestPhpxQualifiedLit(t *testing.T) {
	for _, src := range []string{
		"<?php if (\\false) { echo 1; } $a = \\true; $b = \\null; $c = \\FALSE;",
		"<?php namespace A; $a = \\true; $b = \\strlen('x'); $c = \\Foo::BAR;",
	} {
		if _, err := ParseFile("t.php", src); err != nil {
			t.Errorf("%q: %v", src, err)
		}
	}
}
