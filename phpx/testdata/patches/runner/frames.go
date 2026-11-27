package runner

// This file holds the frame stack. A returning frame notes what it held, so an
// instance with a `__destruct` whose last reference went with it is destroyed
// before the caller's next statement; see phpx_destruct.go.

// pushFrame makes s the frame statements run in, noting the call site
// debug_backtrace() reports for it: the line the call is on and the file the
// caller runs.
func (rt *Runtime) pushFrame(s *Scope) {
	rt.frameSeq++
	s.phpxFrameID = rt.frameSeq
	s.file = ""
	s.callLine = rt.currentSourceLine()
	s.callFile = rt.entrypoint
	if n := len(rt.frames); n > 0 {
		if file := rt.frames[n-1].file; file != "" {
			s.callFile = file
		}
	}
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
	popped.phpxFrameID = 0
	rt.phpxFramePopped(popped)
	// A statement inside a frame only marks the bookkeeping dirty; the frame's
	// return is where the drops are resolved, once, instead of a mark of every
	// root after each statement. See phpxFlush.
	_ = rt.phpxFlush(extra)
}
