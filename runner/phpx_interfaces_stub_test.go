//go:build !phpx

package runner

import (
	"io"
	"strings"
	"testing"
)

// TestPhpxContractReadsOnlyTheClassUntagged pins upstream's answer in the tree
// an untagged build compiles: `implements` is checked against the methods the
// class writes, and a parent's method is no candidate because `extends`
// confers nothing there. The fork's build is phpx_interfaces_test.go.
func TestPhpxContractReadsOnlyTheClassUntagged(t *testing.T) {
	t.Parallel()
	rt := New(io.Discard, Options{})
	program, err := rt.Load(`<?php
interface Speaker { public function speak(); }
class Base { public function speak() { return "base"; } }
class Child extends Base implements Speaker {}
`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	err = rt.Run(program)
	if err == nil || !strings.Contains(err.Error(), "class Child does not declare method speak() required by interface Speaker") {
		t.Errorf("error = %v, want upstream's contract message", err)
	}
}
