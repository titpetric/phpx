package runner

import "testing"

// TestNamespaceLateSitsUnderPX pins that the late installers select with the PX area.
func TestNamespaceLateSitsUnderPX(t *testing.T) {
	t.Parallel()
	if !InNamespace(NamespaceLate, NamespacePX) {
		t.Fatalf("%s is not under %s", NamespaceLate, NamespacePX)
	}
}
