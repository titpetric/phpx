//go:build !phpx

package runner

import (
	"testing"
)

func TestPhpxPreludeStubBuilds(t *testing.T) {
	t.Parallel()
	// The untagged build is upstream's: the stub answers nothing and the seam
	// in the upstream file reads as a no-op.
	rt := New(nil, Options{})
	if rt == nil {
		t.Fatal("no runtime")
	}
}
