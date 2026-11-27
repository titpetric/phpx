package compat

import (
	"io"
	"strings"
	"testing"

	"github.com/titpetric/phpx/runner"
)

// TestRegisterBuffers covers the functions the installer binds.
func TestRegisterBuffers(t *testing.T) {
	t.Parallel()
	rt := runner.New(io.Discard, runner.Options{})
	RegisterBuffers(rt)
	for _, name := range []string{"ob_start", "ob_get_level", "ob_get_contents"} {
		if _, ok := rt.LookupFunc(name); !ok {
			t.Errorf("RegisterBuffers installed no %s", name)
		}
	}
}

// TestBuffersNest covers the stack discipline a template engine renders
// through: ob_start pushes a level, a flush writes into the level underneath,
// a clean drops its text. Writing a flushed level to the real output
// instead would make an outer buffer miss everything an inner one rendered.
func TestBuffersNest(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	rt := runner.New(&out, runner.Options{})
	b := newBuffers(rt)

	b.push()
	rt.Output().Write([]byte("outer-"))
	b.push()
	rt.Output().Write([]byte("inner"))

	if got := b.level(); got != 2 {
		t.Fatalf("level() = %d, want 2", got)
	}

	// The inner level flushes into the outer one and the real writer is
	// still empty and the outer buffer now holds both halves.
	if contents, ok := b.pop(true); !ok || contents != "inner" {
		t.Fatalf("pop(flush) = %q, %t", contents, ok)
	}
	if out.String() != "" {
		t.Errorf("the writer holds %q while a level is still open", out.String())
	}
	if contents, _ := b.contents(); contents != "outer-inner" {
		t.Errorf("the outer level holds %q, want both halves", contents)
	}

	if contents, ok := b.pop(true); !ok || contents != "outer-inner" {
		t.Fatalf("pop(flush) = %q, %t", contents, ok)
	}
	if got, want := out.String(), "outer-inner"; got != want {
		t.Errorf("the writer holds %q, want %q", got, want)
	}
	if got := b.level(); got != 0 {
		t.Errorf("level() = %d after both levels closed", got)
	}
}

// TestBuffersDiscardOnClean covers ob_end_clean: the level closes, and pop
// drops its text. A template that rendered something a page did not want
// leaves no text in the output underneath.
func TestBuffersDiscardOnClean(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	rt := runner.New(&out, runner.Options{})
	b := newBuffers(rt)

	rt.Output().Write([]byte("before;"))
	b.push()
	rt.Output().Write([]byte("captured;"))
	if contents, ok := b.pop(false); !ok || contents != "captured;" {
		t.Fatalf("pop(discard) = %q, %t", contents, ok)
	}
	rt.Output().Write([]byte("after"))

	if got, want := out.String(), "before;after"; got != want {
		t.Errorf("the writer holds %q, want %q", got, want)
	}
}

// TestBuffersReuseAfterDiscard covers the free list: a level closed without
// handing its text to the script returns its buffer, and the next ob_start
// takes it instead of allocating. A suite that opens one level per test is
// the caller this exists for.
func TestBuffersReuseAfterDiscard(t *testing.T) {
	t.Parallel()
	rt := runner.New(io.Discard, runner.Options{})
	b := newBuffers(rt)

	b.push()
	rt.Output().Write([]byte("captured"))
	used := b.stack[0]
	if !b.end(false) {
		t.Fatal("end(false) = false with a level open")
	}
	if len(b.free) != 1 {
		t.Fatalf("the free list holds %d buffers, want 1", len(b.free))
	}

	b.push()
	if b.stack[0] != used {
		t.Error("the next push allocated a fresh buffer instead of reusing one")
	}
	if got := b.level(); got != 1 {
		t.Errorf("level() = %d, want 1", got)
	}
}

// TestBuffersCleanKeepsTheLevel covers ob_clean and ob_flush: both empty the
// active buffer and leave the level open, and a flush writes what it held to
// the output underneath without closing the level.
func TestBuffersCleanKeepsTheLevel(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	rt := runner.New(&out, runner.Options{})
	b := newBuffers(rt)

	b.push()
	rt.Output().Write([]byte("first"))
	if !b.clean() {
		t.Error("clean() = false with a level open")
	}
	if got := b.level(); got != 1 {
		t.Errorf("level() = %d after clean, want 1", got)
	}
	if contents, _ := b.contents(); contents != "" {
		t.Errorf("the level holds %q after clean, want nothing", contents)
	}

	rt.Output().Write([]byte("second"))
	if !b.flush() {
		t.Error("flush() = false with a level open")
	}
	if got, want := out.String(), "second"; got != want {
		t.Errorf("the writer holds %q after flush, want %q", got, want)
	}
	if got := b.level(); got != 1 {
		t.Errorf("level() = %d after flush, want 1", got)
	}
	if contents, _ := b.contents(); contents != "" {
		t.Errorf("the level holds %q after flush, want nothing", contents)
	}
}

// TestBuffersContentsIsACopy covers the string a script reads back: a later
// clean or write cannot change text already handed out. A returned string
// aliasing the buffer would make ob_get_contents depend on what happened next.
func TestBuffersContentsIsACopy(t *testing.T) {
	t.Parallel()
	rt := runner.New(io.Discard, runner.Options{})
	b := newBuffers(rt)

	b.push()
	rt.Output().Write([]byte("abc"))
	held, _ := b.contents()
	b.clean()
	rt.Output().Write([]byte("xyz"))

	if got, want := held, "abc"; got != want {
		t.Errorf("the string read earlier holds %q, want %q", got, want)
	}
	if got, want := b.stack[0].Len(), 3; got != want {
		t.Errorf("length() = %d, want %d", got, want)
	}
}

// TestBuffersFinishFlushesOpenLevels covers the end of a program: every level
// still open is flushed into the one below it, innermost first, and the base
// writer ends up with the whole text in order.
func TestBuffersFinishFlushesOpenLevels(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	rt := runner.New(&out, runner.Options{})
	b := newBuffers(rt)

	b.push()
	rt.Output().Write([]byte("outer-"))
	b.push()
	rt.Output().Write([]byte("inner"))
	b.finish()

	if got := b.level(); got != 0 {
		t.Errorf("level() = %d after finish, want 0", got)
	}
	if got, want := out.String(), "outer-inner"; got != want {
		t.Errorf("the writer holds %q after finish, want %q", got, want)
	}
}

// TestBuffersReportAbsence covers the result PHP distinguishes with ===:
// when no level is open, contents and pop report false, never an empty string.
// A script can tell an empty buffer from an absent buffer.
func TestBuffersReportAbsence(t *testing.T) {
	t.Parallel()
	b := newBuffers(runner.New(&strings.Builder{}, runner.Options{}))

	if contents, ok := b.contents(); ok || contents != "" {
		t.Errorf("contents() = %q, %t with no level open", contents, ok)
	}
	for _, flush := range []bool{false, true} {
		if contents, ok := b.pop(flush); ok || contents != "" {
			t.Errorf("pop(%t) = %q, %t with no level open", flush, contents, ok)
		}
	}
	if got := b.level(); got != 0 {
		t.Errorf("level() = %d on a fresh stack", got)
	}
}
