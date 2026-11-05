//go:build phpx

package runner

import "testing"

// TestPhpxFuncGetArgOnlyItsName pins that other names fall through.
func TestPhpxFuncGetArgOnlyItsName(t *testing.T) {
	t.Parallel()
	if _, ok := phpxFuncGetArg("strlen", "", nil, nil); ok {
		t.Error("answered a name that is not func_get_arg")
	}
}
