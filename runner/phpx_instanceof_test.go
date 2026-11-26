//go:build phpx

package runner

import (
	"errors"
	"fmt"
	"testing"
)

// TestPhpxClassOfError pins that a host error with no class of its own names
// php's Error, which reflection over get_class() then finds.
func TestPhpxClassOfError(t *testing.T) {
	t.Parallel()
	plain := errors.New("x")
	wrapped := fmt.Errorf("load: %w", plain)
	for _, err := range []error{plain, wrapped} {
		if got := instanceOfClass(err); got != "Error" {
			t.Errorf("instanceOfClass(%v) = %q, want Error", err, got)
		}
		if got := phpClassName(err); got != "Error" {
			t.Errorf("phpClassName(%v) = %q, want Error", err, got)
		}
		if class, ok := PhpxClassOfError(err); !ok || class != "Error" {
			t.Errorf("PhpxClassOfError(%v) = %q, %v", err, class, ok)
		}
	}
	if class, ok := PhpxClassOfError("text"); ok || class != "" {
		t.Errorf("PhpxClassOfError(text) = %q, %v, want no answer", class, ok)
	}
}
