package runner

// This file holds the frame stack. A returning frame notes what it held, so an
// instance with a `__destruct` whose last reference went with it is destroyed
// before the caller's next statement; see phpx_destruct.go.

// pushFrame makes s the frame statements run in.
func (rt *Runtime) pushFrame(s *Scope) {
	rt.frames = append(rt.frames, s)
}

// popFrame drops the innermost frame and flushes what it held.
func (rt *Runtime) popFrame() { rt.popFrameHolding(nil) }

// popFrameHolding drops the innermost frame and flushes its drops while extra,
// the value the frame returns, still counts as held: no root names it yet, and
// a flush that did not know it would destroy it before the caller stores it.
func (rt *Runtime) popFrameHolding(extra any) {
	popped := rt.frames[len(rt.frames)-1]
	rt.frames = rt.frames[:len(rt.frames)-1]
	rt.phpxFramePopped(popped)
	// A statement inside a frame only marks the bookkeeping dirty; the frame's
	// return is where the drops are resolved, once, instead of a mark of every
	// root after each statement. See phpxFlush.
	_ = rt.phpxFlush(extra)
}
