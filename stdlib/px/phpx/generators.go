//go:build phpx

package phpx

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/modelutil"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Generators`, registerPhpxGenerators)
}

// Generator is what a function containing yield returns. Its body runs on a
// goroutine of its own and the two sides hand control back and forth, so only
// one of them runs at a time.
type Generator struct {
	run  *genRun
	step func(*Generator) error
}

// genRun is the state a Generator shares with the goroutine running its body.
// The goroutine reaches this and never the Generator, so a Generator a foreach
// abandons is collectable and the cleanup that stops the body can run.
type genRun struct {
	body    func() (any, error)
	started bool
	done    bool
	key     any
	val     any
	ret     any
	err     error
	auto    int64
	sent    any
	frames  []*runner.Scope
	mark    int
	resume  chan struct{}
	paused  chan struct{}
	stop    chan struct{}
}

// current is the generator whose body is running, per runtime.
type genStack struct{ stack []*genRun }

func registerPhpxGenerators(rt *runner.Runtime) {
	st := &genStack{}
	rt.RegisterFunc("__phpx_generator", func(body any) (*Generator, error) {
		call, ok := rt.Callable(body)
		if !ok {
			return nil, fmt.Errorf("generator body is not callable")
		}
		r := &genRun{
			body:   func() (any, error) { return call() },
			resume: make(chan struct{}),
			paused: make(chan struct{}),
			stop:   make(chan struct{}),
		}
		g := &Generator{run: r}
		g.step = func(x *Generator) error { return stepFor(rt)(x) }
		// The body goroutine holds r, never g, so a dropped Generator becomes
		// unreachable and its cleanup runs. Closing stop is what ends a body
		// no consumer will resume; the cleanup runs on a goroutine of its own.
		runtime.AddCleanup(g, func(stop chan struct{}) { close(stop) }, r.stop)
		return g, nil
	})
	yield := func(key, val any, hasKey bool) any {
		r := st.stack[len(st.stack)-1]
		if hasKey {
			r.key = key
			if n, ok := key.(int64); ok && n >= r.auto {
				r.auto = n + 1
			}
		} else {
			r.key = r.auto
			r.auto++
		}
		r.val = val
		r.sent = nil
		r.frames = rt.PhpxFrameDetach(r.mark)
		r.paused <- struct{}{}
		select {
		case <-r.resume:
			return r.sent
		case <-r.stop:
			// The Generator was collected while its body was suspended. The
			// frames stay detached: their popFrame calls skip a detached
			// scope, so unwinding touches no frame the runtime is using.
			// Goexit ends the goroutine and drops what the body held.
			runtime.Goexit()
		}
		return nil
	}
	rt.RegisterFunc("__phpx_yield", func(key, val any, hasKey bool) any { return yield(key, val, hasKey) })
	rt.RegisterFunc("__phpx_yield_from", func(src any) (any, error) {
		switch t := src.(type) {
		case *Generator:
			if err := t.Rewind(); err != nil {
				return nil, err
			}
			for {
				ok, err := t.Valid()
				if err != nil {
					return nil, err
				}
				if !ok {
					break
				}
				k, _ := t.Key()
				v, _ := t.Current()
				yield(k, v, true)
				if err := t.Next(); err != nil {
					return nil, err
				}
			}
			return t.run.ret, nil
		}
		var err error
		walkAny(rt, src, func(k, v any) bool { yield(k, v, true); return true }, &err)
		return nil, err
	})
	// step resumes the body until its next yield or its end.
	step := func(g *Generator) error {
		r := g.run
		if r.done {
			return nil
		}
		st.stack = append(st.stack, r)
		r.mark = rt.PhpxFrameMark()
		rt.PhpxFrameAttach(r.frames)
		r.frames = nil
		if !r.started {
			r.started = true
			go func() {
				<-r.resume
				r.ret, r.err = r.body()
				r.done = true
				r.key, r.val = nil, nil
				r.paused <- struct{}{}
			}()
		}
		r.resume <- struct{}{}
		<-r.paused
		st.stack = st.stack[:len(st.stack)-1]
		if r.done {
			rt.PhpxFrameDetach(r.mark)
		}
		err := r.err
		r.err = nil
		return err
	}
	steppers.Store(rt, step)
}

var steppers sync.Map // *runner.Runtime -> func(*Generator) error

func stepFor(rt *runner.Runtime) func(*Generator) error {
	v, _ := steppers.Load(rt)
	return v.(func(*Generator) error)
}

// walkAny visits an array or a Traversable object.
func walkAny(rt *runner.Runtime, src any, visit func(k, v any) bool, errp *error) {
	switch t := src.(type) {
	case *model.Array:
		t.Range(visit)
	case *model.Object:
		_, *errp = rt.PhpxIterate(t, visit)
	case runner.PhpxHostIterator:
		*errp = runner.PhpxWalkHost(t, visit)
	default:
		// A Go slice or map a binding returned is an array to a script.
		modelutil.RangeValues(src, visit)
	}
}

func (g *Generator) prime() error {
	if !g.run.started {
		return g.step(g)
	}
	return nil
}

// Rewind starts the generator; php refuses a rewind after the first yield.
func (g *Generator) Rewind() error { return g.prime() }

// Valid reports whether the generator holds a value.
func (g *Generator) Valid() (bool, error) {
	if err := g.prime(); err != nil {
		return false, err
	}
	return !g.run.done, nil
}

// Current is the value at the generator's position.
func (g *Generator) Current() (any, error) {
	if err := g.prime(); err != nil {
		return nil, err
	}
	return g.run.val, nil
}

// Key is the key at the generator's position.
func (g *Generator) Key() (any, error) {
	if err := g.prime(); err != nil {
		return nil, err
	}
	return g.run.key, nil
}

// Next resumes the generator.
func (g *Generator) Next() error {
	if err := g.prime(); err != nil {
		return err
	}
	return g.step(g)
}

// Send resumes the generator with a value, which the paused yield answers, and
// returns the next yielded value.
func (g *Generator) Send(v any) (any, error) {
	if err := g.prime(); err != nil {
		return nil, err
	}
	g.run.sent = v
	g.resumeWith(v)
	if err := g.step(g); err != nil {
		return nil, err
	}
	return g.run.val, nil
}

func (g *Generator) resumeWith(v any) { g.run.sent = v }

// GetReturn is the value the body returned.
func (g *Generator) GetReturn() any { return g.run.ret }
