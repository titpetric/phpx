package parser

import "testing"

func TestPhpxNamedParses(t *testing.T) {
	if _, err := ParseFile("t.php", `<?php f(a: 1, b: [...$x, 2]); new Foo(previous: $e);`); err != nil {
		t.Fatal(err)
	}
}
