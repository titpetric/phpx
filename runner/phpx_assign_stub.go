//go:build !phpx

package runner

// phpxInstallEnv adds nothing untagged.
func (rt *Runtime) phpxInstallEnv(_ map[string]any, _ *scopeRef) {}
