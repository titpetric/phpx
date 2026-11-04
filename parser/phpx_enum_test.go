package parser

import "testing"

func TestPhpxEnumParses(t *testing.T) {
	if _, err := ParseFile("t.php", `<?php enum S: string implements JsonSerializable { case A = "a"; case B; const X = self::A; public function f() { return 1; } } $x = S::A->value;`); err != nil {
		t.Fatal(err)
	}
}
