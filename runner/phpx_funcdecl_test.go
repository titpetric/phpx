//go:build phpx

package runner

import (
	"strings"
	"testing"
)

// TestPhpxFuncDeclDeclaresWhenReached covers the polyfill guard: a function
// below the top level exists once control reaches its statement, and not before.
func TestPhpxFuncDeclDeclaresWhenReached(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	rt := New(&out, Options{})
	program, err := rt.Load(`<?php
namespace Lib;
$run = true;
if ($run) { function f() { return "f"; } }
echo f(), "\n";
`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := rt.Run(program); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got, want := out.String(), "f\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}
