//go:build !phpx

package runner

import (
	"testing"
)

// TestCatchHierarchyStub pins that an untagged build keeps upstream's flat
// throwable names.
func TestCatchHierarchyStub(t *testing.T) {
	t.Parallel()
	if phpxCatchHierarchy {
		t.Fatal("an untagged build reports a throwable hierarchy")
	}
}
