//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxOwnedKeepsSlicesForTheGoBuiltinsOnly covers which import keeps a Go
// slice a slice: `use Runtime\Go` does, and an import of an area below it binds
// functions and leaves the values they answer to become php arrays.
func TestPhpxOwnedKeepsSlicesForTheGoBuiltinsOnly(t *testing.T) {
	RegisterNamespace(model.GoNamespace, func(*Runtime) {})
	RegisterNamespace(model.GoNamespace+`\Test\Owned`, func(*Runtime) {})

	plain := New(nil, Options{})
	if _, ok := plain.phpxOwned([]string{"a"}).(*model.Array); !ok {
		t.Error("a runtime with no import keeps a slice")
	}

	below := New(nil, Options{})
	below.UseNamespace(model.GoNamespace + `\Test\Owned`)
	if _, ok := below.phpxOwned([]string{"a"}).(*model.Array); !ok {
		t.Error("an import below Runtime\\Go keeps a slice")
	}

	builtins := New(nil, Options{})
	builtins.UseNamespace(model.GoNamespace)
	if _, ok := builtins.phpxOwned([]string{"a"}).([]string); !ok {
		t.Error("Runtime\\Go turned a slice into an array")
	}
}
