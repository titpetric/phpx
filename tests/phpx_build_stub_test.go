//go:build !phpx

package tests

import (
	"testing"
)

// TestPhpxBuildStub pins that an untagged build reports no phpx runtime.
func TestPhpxBuildStub(t *testing.T) {
	t.Parallel()
	if PhpxBuild {
		t.Fatal("an untagged build reports phpx")
	}
}
