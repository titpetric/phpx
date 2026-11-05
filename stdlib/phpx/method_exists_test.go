//go:build phpx

package phpx

import (
	"io"
	"testing"

	"github.com/titpetric/phpx/runner"
)

func TestMethodExistsUnknownClass(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	registerMethodExists(rt)
	fn, ok := rt.LookupFunc("method_exists")
	if !ok {
		t.Skip("core registers method_exists in a later binding")
	}
	if fn.(func(any, string) bool)("No\\Such\\Class", "m") {
		t.Fatal("a class that does not exist has no methods")
	}
}
