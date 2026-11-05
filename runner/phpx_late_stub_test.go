//go:build !phpx

package runner

import (
	"testing"
)

func TestPhpxLateStubIsInert(t *testing.T) {
	PhpxRegisterLate(func(*Runtime) { t.Fatal("untagged build ran a late installer") })
	PhpxRunLate(nil)
}
