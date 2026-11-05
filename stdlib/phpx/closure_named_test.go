//go:build phpx

package phpx

import (
	"io"
	"testing"

	"github.com/titpetric/phpx/runner"
)

func TestNamedClosureRejectsNonCallable(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	registerNamedClosure(rt)
	fn, ok := rt.LookupFunc("Closure::fromCallable")
	if !ok {
		t.Fatal("not registered")
	}
	if _, err := fn.(func(any) (any, error))(42); err == nil {
		t.Fatal("42 is not callable")
	}
}
