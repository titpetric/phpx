package parser

import (
	"testing"
)

func TestPhpxClassOfParses(t *testing.T) {
	for _, src := range []string{
		`<?php $n = $o::class;`,
		`<?php $n = $this->o::class . "x";`,
		`<?php $n = f()::class;`,
		`<?php $n = A::class;`,
	} {
		if _, err := ParseFile("t.php", src); err != nil {
			t.Errorf("%s: %v", src, err)
		}
	}
}
