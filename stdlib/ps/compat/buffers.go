package compat

import (
	"bytes"

	"github.com/titpetric/phpx/runner"
)

// buffers holds one runtime's output buffering levels, typed for ob_get_contents.
//
// A level is a byte buffer with a free list behind it: ob_start takes one and
// ob_end_clean, ob_end_flush, ob_clean and ob_flush give it back with its
// capacity, so a suite that opens a level per test stops allocating a fresh
// byte slice per test. A string read out of a level is a copy, as php's is, so
// the buffer under it is reusable immediately.
type buffers struct {
	rt    *runner.Runtime
	stack []*bytes.Buffer
	free  []*bytes.Buffer
}

const (
	// obFreeLevels caps the buffers kept for reuse. The active levels and the
	// free list together are the whole cost of buffering.
	obFreeLevels = 16
	// obFreeBytes caps the capacity a reused buffer keeps. A level that grew
	// past it is dropped: holding a page for the next level costs more than
	// the allocation it saves.
	obFreeBytes = 1 << 20
)

func newBuffers(rt *runner.Runtime) *buffers {
	return &buffers{rt: rt}
}

// take answers an empty buffer, reusing one a closed level returned.
func (b *buffers) take() *bytes.Buffer {
	if n := len(b.free); n > 0 {
		buf := b.free[n-1]
		b.free = b.free[:n-1]
		return buf
	}
	return &bytes.Buffer{}
}

// put empties a buffer and keeps it for the next level. A buffer that grew
// past obFreeBytes, or a full free list, is left for the collector.
func (b *buffers) put(buf *bytes.Buffer) {
	if buf == nil || buf.Cap() > obFreeBytes || len(b.free) >= obFreeLevels {
		return
	}
	buf.Reset()
	b.free = append(b.free, buf)
}

// push starts a buffering level, backing ob_start.
func (b *buffers) push() {
	buf := b.take()
	b.stack = append(b.stack, buf)
	b.rt.PushOutput(buf)
}

// level reports how many levels are active, backing ob_get_level.
func (b *buffers) level() int { return len(b.stack) }

// reset drops every buffering level a program left open. The runtime released
// the redirections before calling this, so the buffers are the only state left.
func (b *buffers) reset() {
	for _, buf := range b.stack {
		b.put(buf)
	}
	clear(b.stack)
	b.stack = b.stack[:0]
}

// contents returns the innermost buffer without ending it. The string is a
// copy, so a later clean or flush cannot change what the script holds. It
// answers false when no buffer is active.
func (b *buffers) contents() (string, bool) {
	if len(b.stack) == 0 {
		return "", false
	}
	return b.stack[len(b.stack)-1].String(), true
}

// length reports the innermost buffer's byte count without copying it.
func (b *buffers) length() (int, bool) {
	if len(b.stack) == 0 {
		return 0, false
	}
	return b.stack[len(b.stack)-1].Len(), true
}

// pop ends the innermost level and answers its contents. Writing them on when
// flush is true targets the enclosing level, so nested buffers compose.
func (b *buffers) pop(flush bool) (string, bool) {
	if len(b.stack) == 0 {
		return "", false
	}
	last := len(b.stack) - 1
	buf := b.stack[last]
	b.stack = b.stack[:last]
	b.rt.PopOutput()
	if flush {
		// Write, not WriteString: the enclosing writer may be another level's
		// byte buffer, and the bytes are already here.
		_, _ = b.rt.Output().Write(buf.Bytes())
	}
	contents := buf.String()
	b.put(buf)
	return contents, true
}

// end ends the innermost level without answering its contents, so the script
// never sees the text and the buffer is reused. Writing them on when flush is
// true targets the enclosing level.
func (b *buffers) end(flush bool) bool {
	if len(b.stack) == 0 {
		return false
	}
	last := len(b.stack) - 1
	buf := b.stack[last]
	b.stack = b.stack[:last]
	b.rt.PopOutput()
	if flush {
		_, _ = b.rt.Output().Write(buf.Bytes())
	}
	b.put(buf)
	return true
}

// finish flushes every level a program left open, innermost first, as php does
// at request shutdown. Each flush writes into the level below it, so a message
// written under an open ob_start reaches the output it was aimed at.
func (b *buffers) finish() {
	for len(b.stack) > 0 {
		b.end(true)
	}
}

// clean drops the innermost level's text without ending it, backing ob_clean.
func (b *buffers) clean() bool {
	if len(b.stack) == 0 {
		return false
	}
	b.stack[len(b.stack)-1].Reset()
	return true
}

// flush writes the innermost level's text to the enclosing level and keeps the
// level open, backing ob_flush.
func (b *buffers) flush() bool {
	if len(b.stack) == 0 {
		return false
	}
	buf := b.stack[len(b.stack)-1]
	// Write to the writer under this level, not to the level itself.
	b.rt.PopOutput()
	_, _ = b.rt.Output().Write(buf.Bytes())
	buf.Reset()
	b.rt.PushOutput(buf)
	return true
}
