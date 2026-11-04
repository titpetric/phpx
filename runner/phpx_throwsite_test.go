//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxStampThrowableSkipsPlainObjects pins that only an object with the
// throwable root's members is stamped.
func TestPhpxStampThrowableSkipsPlainObjects(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	plain := model.NewObject(&model.Class{Name: "P"})
	plain.SetProp("line", int64(7))
	rt.phpxStampThrowable(plain, nil)
	if got, _ := plain.Prop("line"); got != int64(7) {
		t.Errorf("plain line = %v, want 7", got)
	}
	exc := model.NewObject(&model.Class{Name: "E"})
	exc.SetProp("line", int64(0))
	exc.SetProp("previous", nil)
	rt.phpxStampThrowable(exc, nil)
	if got, _ := exc.Prop("line"); got != int64(0) {
		t.Errorf("line = %v, want 0 with no statement running", got)
	}
}
