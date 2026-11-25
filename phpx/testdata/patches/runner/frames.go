package runner

// This file holds the frame stack. A returning frame notes what it held, so an
// instance with a `__destruct` whose last reference went with it is destroyed
// before the caller's next statement; see phpx_destruct.go.

// pushFrame makes s the frame statements run in.
func (rt *Runtime) pushFrame(s *Scope) {
	rt.frames = append(rt.frames, s)
}

// popFrame drops the innermost frame and notes what it held.
func (rt *Runtime) popFrame() {
	popped := rt.frames[len(rt.frames)-1]
	rt.frames = rt.frames[:len(rt.frames)-1]
	rt.phpxFramePopped(popped)
}
