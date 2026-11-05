//go:build !phpx

package runner

import (
	"testing"
)

func TestPhpxHostErrorCatchesStubTakesAll(t *testing.T) {
	if !phpxHostErrorCatches("Anything") {
		t.Fatal("the untagged build keeps upstream's rule")
	}
}
