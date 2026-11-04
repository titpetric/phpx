package parser

import (
	"testing"
)

func TestPhpxLogicalParses(t *testing.T) {
	for _, src := range []string{
		`<?php $x = f() or die();`,
		`<?php if ($a xor $b) { echo 1; }`,
		`<?php $x = $a and $b or $c;`,
		`<?php for ($i = 0, $c = 3; $i < $c; $i++, $j--) {}`,
		`<?php for (;;) { break; }`,
		`<?php return self::$i ?? self::$i = new self;`,
	} {
		if _, err := ParseFile("t.php", src); err != nil {
			t.Errorf("%s: %v", src, err)
		}
	}
}
