package runner

import (
	"strings"
	"testing"
	"testing/fstest"
)

// TestPhpxVendorAutoloadsFromAnEmbeddedTree mounts a tree with a composer-style
// loader behind an empty root and runs a script that names a vendored class.
func TestPhpxVendorAutoloadsFromAnEmbeddedTree(t *testing.T) {
	t.Parallel()
	vendor := fstest.MapFS{
		"vendor/autoload.php": &fstest.MapFile{Data: []byte(`<?php
require __DIR__ . '/src/Lib/Greeter.php';
`)},
		"vendor/src/Lib/Greeter.php": &fstest.MapFile{Data: []byte(`<?php
namespace Lib;
class Greeter { public function hi() { return "hi"; } }
`)},
	}
	var out strings.Builder
	rt := New(&out, Options{RootFS: fstest.MapFS{}})
	rt.Vendor(vendor)
	program, err := rt.Load(`<?php echo (new Lib\Greeter())->hi(), "\n";`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := rt.Run(program); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := out.String(); got != "hi\n" {
		t.Errorf("output = %q, want %q", got, "hi\n")
	}
}
