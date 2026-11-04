package parser

import "testing"

func TestPhpxDestructureParses(t *testing.T) {
	for _, src := range []string{
		`<?php [$a, $b] = [1, 2];`,
		`<?php [$a, [$b, $c]] = f();`,
		`<?php ['x' => $a, 'y' => [$b]] = f();`,
		`<?php [$a, , $c] = f();`,
		`<?php foreach ($r as [$a, $b]) { echo $a; }`,
		`<?php foreach ($r as $k => ['id' => $id, 'p' => [$x]]) { echo $id; }`,
		`<?php $r = [$a, [$b]] = f();`,
	} {
		if _, err := ParseFile("t.php", src); err != nil {
			t.Errorf("%s: %v", src, err)
		}
	}
}
