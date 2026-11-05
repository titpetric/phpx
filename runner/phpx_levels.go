//go:build phpx

package runner

// phpxOuterLevels records how many loops beyond the innermost a `break N` or
// `continue N` leaves. Zero and one both name the innermost alone.
func (rt *Runtime) phpxOuterLevels(levels int) {
	rt.phpxOuter = 0
	if levels > 1 {
		rt.phpxOuter = levels - 1
	}
}

// phpxLeave reports whether the loop that got fl from its body has to pass it
// on to the loop around it: a break or continue with levels left to go. It
// uses one level up.
func (rt *Runtime) phpxLeave(fl flow) bool {
	if rt.phpxOuter > 0 && (fl == flowBreak || fl == flowContinue) {
		rt.phpxOuter--
		return true
	}
	return false
}
