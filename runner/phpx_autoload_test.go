//go:build phpx

package runner

import (
	"io"
	"strings"
	"testing"
	"testing/fstest"
)

// TestPhpxAutoloadResolvesTheParent covers php's rule that `extends` asks the
// autoload queue for a name the file does not declare. composer's PSR-4 loader
// is how a framework hands a parent to the class that extends it, and
// `--include vendor/autoload.php` registers the loader before the program is
// hoisted.
func TestPhpxAutoloadResolvesTheParent(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	rt := New(&out, Options{RootFS: fstest.MapFS{
		"lib/acme/base.php": &fstest.MapFile{Data: []byte(`<?php
namespace Acme;

class Base
{
	public $name = "base";

	public function greet()
	{
		return "hello " . $this->name;
	}
}
`)},
	}})
	rt.SetIncludePath("lib")
	rt.RegisterAutoloader(rt.SPLAutoload, false)

	program, err := rt.Load(`<?php
class Child extends Acme\Base
{
}
echo (new Child())->greet(), "\n";
`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := rt.Run(program); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got, want := out.String(), "hello base\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

// TestPhpxAutoloadOfAMissingParentFails keeps the fatal for a name no loader
// answers: consulting the queue changes nothing for a parent nothing declares.
func TestPhpxAutoloadOfAMissingParentFails(t *testing.T) {
	t.Parallel()
	rt := New(io.Discard, Options{})
	program, err := rt.Load(`<?php class Child extends Absent {}`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	err = rt.Run(program)
	if err == nil || !strings.Contains(err.Error(), "parent class Absent not found") {
		t.Errorf("error = %v, want the missing-parent fatal", err)
	}
}
