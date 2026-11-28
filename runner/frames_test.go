package runner

import (
	"testing"
)

// TestFramesPushPop pins the frame stack. popFrame drops the innermost frame
// and leaves the one below current; a returning frame also notes what it held,
// which the fork's destructor reads (runner/phpx_destruct.go).
func TestFramesPushPop(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	outer := NewScope()
	inner := NewScope()
	rt.pushFrame(outer)
	rt.pushFrame(inner)
	if len(rt.frames) != 2 || rt.frames[1] != inner {
		t.Fatalf("frames = %v, want the two pushed", rt.frames)
	}
	rt.popFrame(outer)
	if len(rt.frames) != 1 || rt.frames[0] != outer {
		t.Fatalf("after pop frames = %v, want the outer frame", rt.frames)
	}
}
