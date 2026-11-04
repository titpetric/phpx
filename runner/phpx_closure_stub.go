//go:build !phpx

package runner

// phpxClosureMethod answers the untagged build, where a closure has no
// bindTo() or call(), as upstream has it.
func (rt *Runtime) phpxClosureMethod(base any, method string, args []any) (any, bool, error) {
	return nil, false, nil
}
