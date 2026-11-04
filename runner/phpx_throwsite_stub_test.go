//go:build !phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxStampThrowableUntagged pins upstream's answer: no site is recorded.
func TestPhpxStampThrowableUntagged(t *testing.T) {
	t.Parallel()
	obj := model.NewObject(&model.Class{Name: "E"})
	obj.SetProp("line", int64(0))
	(&Runtime{}).phpxStampThrowable(obj, nil)
	if got, _ := obj.Prop("line"); got != int64(0) {
		t.Errorf("line = %v, want untouched", got)
	}
}
