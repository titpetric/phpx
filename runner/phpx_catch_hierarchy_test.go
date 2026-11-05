//go:build phpx

package runner

import "testing"

// TestCatchHierarchyTagged pins that a phpx build lets a clause naming a parent
// take a child built-in throwable.
func TestCatchHierarchyTagged(t *testing.T) {
	t.Parallel()
	if !phpxCatchHierarchy {
		t.Fatal("a phpx build keeps no throwable hierarchy")
	}
}
