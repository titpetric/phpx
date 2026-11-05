//go:build phpx

package runner

import "strings"

// phpxFuncGetArg answers func_get_arg(i) when a namespaced file reaches it
// through the global fallback, where scopeBuiltin cannot pass the index.
func phpxFuncGetArg(name, fallback string, args []any, scope *Scope) (any, bool) {
	if !strings.EqualFold(name, "func_get_arg") && !strings.EqualFold(fallback, "func_get_arg") {
		return nil, false
	}
	if len(args) == 0 {
		return nil, true
	}
	all := funcGetArgs(scope)
	i := toInt(args[0])
	if i < 0 || int(i) >= len(all) {
		return nil, true
	}
	return all[i], true
}
