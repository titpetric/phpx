//go:build !phpx

package runner

// phpxInstallRefExpr answers the untagged build, where only a variable is an
// output parameter.
func (rt *Runtime) phpxInstallRefExpr(_ map[string]any, _ *scopeRef) {}

// phpxHandleID answers the untagged build, which casts a handle as it always did.
func phpxHandleID(_ any) (int64, bool) { return 0, false }
