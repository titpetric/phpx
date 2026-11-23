package runner

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/parser"
)

// TestExecUnsetTargets runs unset() over each lvalue shape it accepts: a name,
// an array element, a property and a name nothing set. The runtime in this
// package carries no standard library, so the programs print through the
// language alone; the behaviour itself is settled against php in
// phpx/testdata/classes/destruct.phpt.
func TestExecUnsetTargets(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		src     string
		wantErr bool
	}{
		{name: "variable", src: `<?php $a = 1; unset($a);`},
		{name: "index", src: `<?php $m = ['k' => 2]; unset($m['k']);`},
		{name: "property", src: `<?php class C { public $p = 3; } $c = new C; unset($c->p);`},
		{name: "missing", src: `<?php unset($nothing);`},
		{name: "list element", src: `<?php $list = [7, 8]; unset($list[1]); echo $list[0], "\n";`},
		{name: "empty index", src: `<?php unset($m[]);`, wantErr: true},
		{name: "not an lvalue", src: `<?php unset(1 + 2);`, wantErr: true},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			prog, err := parser.ParseFileOptions("main.php", tc.src, parser.Options{})
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			var out strings.Builder
			rt := New(&out, Options{})
			rt.UpdateFilename("main.php")
			err = rt.Run(prog)
			if (err != nil) != tc.wantErr {
				t.Fatalf("run = %v, want error %v", err, tc.wantErr)
			}
		})
	}
}

// TestExecUnsetLeavesTheRest pins that an unset element leaves its neighbours
// in order, which is what the array store hands back to foreach.
func TestExecUnsetLeavesTheRest(t *testing.T) {
	t.Parallel()
	src := `<?php
$m = ['a' => 1, 'b' => 2, 'c' => 3];
unset($m['b']);
foreach ($m as $k => $v) { echo $k, "=", $v, "\n"; }
`
	prog, err := parser.ParseFileOptions("main.php", src, parser.Options{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out strings.Builder
	rt := New(&out, Options{})
	rt.UpdateFilename("main.php")
	if err := rt.Run(prog); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got, want := out.String(), "a=1\nc=3\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
