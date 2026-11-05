//go:build phpx

package phpx

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/runner"
)

// TestSapiStubsRegistered pins that the CLI answers are all bound.
func TestSapiStubsRegistered(t *testing.T) {
	t.Parallel()
	rt := runner.New(&strings.Builder{}, runner.Options{})
	registerSapiStubs(rt)
	for _, name := range []string{"header_remove", "headers_list", "flush", "getmyuid", "get_current_user", "disk_free_space"} {
		if _, ok := rt.LookupFunc(name); !ok {
			t.Errorf("%s missing", name)
		}
	}
}
