package parser

import (
	"testing"
)

func TestPhpxNullsafeParses(t *testing.T) {
	for _, src := range []string{
		`<?php $n = $o?->a;`,
		`<?php $n = $o?->m(1)->p?->q();`,
		`<?php $n = $o?->a ?? 'd';`,
		`<?php $n = $a ?-1 : 2;`,
	} {
		if _, err := ParseFile("t.php", src); err != nil {
			t.Errorf("%s: %v", src, err)
		}
	}
}
