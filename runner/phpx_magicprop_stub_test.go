//go:build !phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxMagicPropUntagged pins that the untagged build handles no magic property.
func TestPhpxMagicPropUntagged(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	obj := model.NewObject(&model.Class{Name: "A"})
	if _, handled, _ := rt.phpxMagicGet(obj, "x", false, nil); handled {
		t.Error("untagged build handled __get")
	}
	if handled, _ := rt.phpxMagicSet(obj, "x", 1, nil); handled {
		t.Error("untagged build handled __set")
	}
	if handled, _ := rt.phpxMagicUnset(obj, "x", nil); handled {
		t.Error("untagged build handled __unset")
	}
}
