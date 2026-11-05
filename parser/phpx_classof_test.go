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

// TestAssignStaticPropOnValue pins that `$cls::$prop = v` parses.
func TestAssignStaticPropOnValue(t *testing.T) {
	t.Parallel()
	if _, err := Parse("<?php $d::$v = 5; $d::$s .= 'x';"); err != nil {
		t.Fatal(err)
	}
}
