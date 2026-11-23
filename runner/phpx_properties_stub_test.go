//go:build !phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxPropertiesConferNothingUntagged pins upstream's answer: a private
// property is readable from anywhere.
func TestPhpxPropertiesConferNothingUntagged(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	obj := &model.Object{Class: &model.Class{Name: "A"}}
	if hidden, err := rt.phpxPropVerdict(obj, "p", nil); hidden || err != nil {
		t.Errorf("phpxPropVerdict = %v, %v, want an open access untagged", hidden, err)
	}
	if err := rt.phpxPropAccess(obj, "p", nil); err != nil {
		t.Errorf("phpxPropAccess = %v, want nil untagged", err)
	}
}

// TestPhpxCoercePropertyUntagged pins the untagged answer: a value is stored as
// given, because no property carries a type to convert for.
func TestPhpxCoercePropertyUntagged(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	value := &model.Object{Class: &model.Class{Name: "S"}}
	if got := rt.phpxCoerceProperty(&model.Object{Class: &model.Class{Name: "T"}}, "name", value); got != value {
		t.Errorf("phpxCoerceProperty = %v, want the value unchanged untagged", got)
	}
}
