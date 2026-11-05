package parser

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestVarVarParses(t *testing.T) {
	for _, src := range []string{
		`<?php $x = $$name;`, `<?php $$name = 5;`, `<?php $${a . 'b'} .= 'c';`, `<?php echo $$$n;`,
	} {
		prog, err := Parse(src)
		if err != nil {
			t.Fatalf("%s: %v", src, err)
		}
		if len(prog.Stmts) == 0 {
			t.Fatalf("%s: no statements", src)
		}
	}
	_ = model.Call{}
}
