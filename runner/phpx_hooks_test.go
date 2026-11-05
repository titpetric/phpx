//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxHooksSkipAClassWithNone(t *testing.T) {
	rt := New(nil, Options{})
	if _, handled, _ := rt.phpxHookGet(&model.Object{}, "x", nil); handled {
		t.Error("a nil object has no hook")
	}
}
