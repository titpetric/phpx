//go:build phpx

package pexec_test

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/parser"
	"github.com/titpetric/phpx/runner"
	"github.com/titpetric/phpx/stdlib"
)

// TestProcOpen pins a child started with pipes: its stdout is read through the
// pipe and proc_close answers its exit code.
func TestProcOpen(t *testing.T) {
	t.Parallel()
	program, err := parser.Parse(`<?php
$p = proc_open(["sh", "-c", "echo out; exit 3"], [1 => ["pipe", "w"], 2 => ["file", "/dev/null", "w"]], $pipes);
echo trim(stream_get_contents($pipes[1])), "|", proc_close($p), "\n";
$p = proc_open("cat", [0 => ["pipe", "r"], 1 => ["pipe", "w"]], $pipes);
fwrite($pipes[0], "piped"); fclose($pipes[0]);
echo stream_get_contents($pipes[1]), "|", proc_close($p), "\n";`)
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	rt := runner.New(&out, runner.Options{})
	stdlib.Register(rt)
	if err := rt.Run(program); err != nil {
		t.Fatal(err)
	}
	if want := "out|3\npiped|0\n"; out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}
