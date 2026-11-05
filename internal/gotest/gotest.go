// Package gotest runs the scripts the Runtime\Go namespace tests share: a script
// per case on a runtime carrying the default surface, so an import is what
// installs the namespace under test.
package gotest

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/runner"
	"github.com/titpetric/phpx/stdlib"
)

// Case is a script and the output it echoes.
type Case struct {
	Src  string
	Want string
}

// Run executes each case on a fresh runtime and compares the output.
func Run(t *testing.T, namespace string, cases []Case) {
	t.Helper()
	for _, tc := range cases {
		var out strings.Builder
		rt := runner.New(&out, runner.Options{})
		stdlib.Register(rt)
		program, err := rt.Load("<?php " + tc.Src)
		if err != nil {
			t.Errorf("%s: parse %q: %v", namespace, tc.Src, err)
			continue
		}
		if err := rt.Run(program); err != nil {
			t.Errorf("%s: run %q: %v", namespace, tc.Src, err)
			continue
		}
		if out.String() != tc.Want {
			t.Errorf("%s: %q\n got  %q\n want %q", namespace, tc.Src, out.String(), tc.Want)
		}
	}
}
