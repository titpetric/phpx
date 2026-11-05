//go:build !phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxSnapshotSourceStubIsOff(t *testing.T) {
	if phpxSnapshotSource(&model.Foreach{}) {
		t.Fatal("the untagged build walks the live array")
	}
}
