//go:build phpx

package compat

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/runner"
)

// TestPhpxBuffersPresent pins that the extra ob_* functions are registered.
func TestPhpxBuffersPresent(t *testing.T) {
	t.Parallel()
	rt := runner.New(&strings.Builder{}, runner.Options{})
	Register(rt)
	for _, name := range []string{"ob_clean", "ob_flush", "ob_get_length", "ob_list_handlers"} {
		if _, ok := rt.LookupFunc(name); !ok {
			t.Errorf("%s not registered", name)
		}
	}
}
