package parser

import (
	"testing"
)

func TestPhpxCoalesceParses(t *testing.T) {
	for _, src := range []string{
		`<?php $a = $b ?? 1;`,
		`<?php $a = $b['k'] ?? $c ?? 2;`,
		`<?php $a ??= 5; $o->p ??= []; $m['x'] ??= 1;`,
		`<?php $a = $x ?? 1 ? 'y' : 'n';`,
		`<?php $a = f() ?? 1; $b = $o->m()['k'] ?? 2; $c = f()->p->q ?? 3;`,
	} {
		if _, err := ParseFile("t.php", src); err != nil {
			t.Errorf("%s: %v", src, err)
		}
	}
}
