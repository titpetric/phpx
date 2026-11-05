//go:build phpx

package runner

import (
	"io"
	"testing"
)

func TestPhpxLateRunsInOrder(t *testing.T) {
	var order []int
	PhpxRegisterLate(func(*Runtime) { order = append(order, 1) })
	PhpxRegisterLate(func(*Runtime) { order = append(order, 2) })
	PhpxRunLate(New(io.Discard, Options{}))
	if len(order) < 2 || order[len(order)-2] != 1 || order[len(order)-1] != 2 {
		t.Fatalf("order %v", order)
	}
}
