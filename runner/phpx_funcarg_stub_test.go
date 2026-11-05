//go:build !phpx

package runner

import (
	"testing"
)

// TestPhpxFuncGetArgUntagged pins that the untagged build answers nothing.
func TestPhpxFuncGetArgUntagged(t *testing.T) {
	t.Parallel()
	if _, ok := phpxFuncGetArg("func_get_arg", "", nil, nil); ok {
		t.Error("untagged build answered func_get_arg")
	}
}
