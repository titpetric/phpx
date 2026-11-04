package parser

import "testing"

func TestPhpxMatchParses(t *testing.T) {
	if _, err := ParseFile("t.php", `<?php $x = match ($a) { 1, 2 => "a", default => throw new E("x") };`); err != nil {
		t.Fatal(err)
	}
}
