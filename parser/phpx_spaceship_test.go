package parser

import (
	"testing"
)

func TestPhpxSpaceshipParses(t *testing.T) {
	for _, src := range []string{
		`<?php $a = 1 <=> 2;`,
		`<?php usort($r, fn($x, $y) => $x[0] <=> $y[0]);`,
		`<?php $a = 1 + 1 <=> 2 + 3;`,
	} {
		if _, err := ParseFile("t.php", src); err != nil {
			t.Errorf("%s: %v", src, err)
		}
	}
}
