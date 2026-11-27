//go:build phpx

package runner_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/titpetric/phpx/parser"
	"github.com/titpetric/phpx/runner"
	"github.com/titpetric/phpx/stdlib"
)

// TestPhpxEvalNamesCaller covers the file php gives code eval() runs:
// "<caller>(<line>) : eval()'d code", which __FILE__ inside the eval and a
// class it declares both report. Pest's TestSuiteLoader accepts the test-case
// class it evals only when that name holds TestCaseFactory.php, and
// ReflectionMethod::getAttributes() reads the #[Test] the generated method
// carries. Both answers are php 8.4's.
func TestPhpxEvalNamesCaller(t *testing.T) {
	t.Parallel()
	const src = `<?php
eval('class EvalName { #[Test] public function t() {} }');
$file = (new ReflectionClass('EvalName'))->getFileName();
echo $file === __FILE__ . '(2) : eval()\'d code' ? "name" : "name:" . $file, "\n";
$m = (new ReflectionClass('EvalName'))->getMethod('t');
$attrs = $m->getAttributes();
echo count($attrs), " ", $attrs ? $attrs[0]->getName() : "-", "\n";
`
	var out strings.Builder
	rt := runner.New(&out, runner.Options{RootFS: fstest.MapFS{
		"script.php": &fstest.MapFile{Data: []byte(src)},
	}})
	stdlib.Register(rt)
	program, err := rt.LoadFile("script.php")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := rt.Run(program); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got, want := out.String(), "name\n1 Test\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

// TestPhpxEvalWithoutCaller keeps the fallback: code with no caller frame takes
// php's bare "eval()'d code" name rather than an empty one.
func TestPhpxEvalWithoutCaller(t *testing.T) {
	t.Parallel()
	program, err := parser.Parse(`<?php eval('echo __FILE__, "\n";');`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out strings.Builder
	rt := runner.New(&out, runner.Options{})
	stdlib.Register(rt)
	if err := rt.Run(program); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got, want := out.String(), "eval()'d code\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}
