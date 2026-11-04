package parser

import "testing"

func TestPhpxHeredocParses(t *testing.T) {
	for _, src := range []string{
		"<?php $a = <<<EOT\nhi $x\nEOT;\n",
		"<?php $a = <<<'EOT'\nraw $x\nEOT;\n",
		"<?php $a = <<<\"EOT\"\n  indented {$x['k']}\n  EOT;\n",
		"<?php f(<<<A\none\nA, <<<B\ntwo\nB);\n",
		"<?php class C { const T = <<<'X'\nbody\nX; }\n",
	} {
		if _, err := ParseFile("t.php", src); err != nil {
			t.Errorf("%q: %v", src, err)
		}
	}
}
