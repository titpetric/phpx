//go:build !phpx

package runner

import (
	"testing"
)

func TestPhpxMagicCallStub(t *testing.T) {
	t.Parallel()
	if _, handled, _ := New(nil, Options{}).phpxMagicCall(nil, "m", nil, nil); handled {
		t.Fatal("untagged build has no __call")
	}
}
