//go:build !phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxHostParentStub(t *testing.T) {
	t.Parallel()
	if _, ok := phpxHostValue(&model.Object{}); ok {
		t.Fatal("untagged build has no host parent")
	}
}
