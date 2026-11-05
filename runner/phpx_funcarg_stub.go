//go:build !phpx

package runner

// phpxFuncGetArg answers the untagged build, which has no func_get_arg to find.
func phpxFuncGetArg(_, _ string, _ []any, _ *Scope) (any, bool) { return nil, false }
