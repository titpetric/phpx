//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/modelutil"
	"fmt"
	"sync"

	"github.com/titpetric/phpx/model"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Generators`, registerPhpxGenerators)
}

// Generator is what a function containing yield returns. Its body runs on a
// goroutine of its own and the two sides hand control back and forth, so only
// one of them runs at a time.
type Generator struct {
	rt      *runner.Runtime
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
	step    func(*Generator) error
}

// current is the generator whose body is running, per runtime.
type genStack struct{ stack []*Generator }

func registerPhpxGenerators(rt *runner.Runtime) {
	st := &genStack{}
	rt.RegisterFunc("__phpx_generator", func(body any) (*Generator, error) {
		call, ok := rt.Callable(body)
		if !ok {
			return nil, fmt.Errorf("generator body is not callable")
		}
		g := &Generator{rt: rt, resume: make(chan struct{}), paused: make(chan struct{}), step: nil}
		g.body = func() (any, error) { return call() }
		g.step = func(x *Generator) error { return stepFor(rt)(x) }
		return g, nil
	})
	yield := func(key, val any, hasKey bool) any {
		g := st.stack[len(st.stack)-1]
		if hasKey {
			g.key = key
			if n, ok := key.(int64); ok && n >= g.auto {
				g.auto = n + 1
			}
		} else {
			g.key = g.auto
			g.auto++
		}
		g.val = val
		g.sent = nil
		g.frames = rt.PhpxFrameDetach(g.mark)
		g.paused <- struct{}{}
		<-g.resume
		return g.sent
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
			return t.ret, nil
		}
		var err error
		walkAny(rt, src, func(k, v any) bool { yield(k, v, true); return true }, &err)
		return nil, err
	})
	// step resumes the body until its next yield or its end.
	step := func(g *Generator) error {
		if g.done {
			return nil
		}
		st.stack = append(st.stack, g)
		g.mark = rt.PhpxFrameMark()
		rt.PhpxFrameAttach(g.frames)
		g.frames = nil
		if !g.started {
			g.started = true
			go func() {
				<-g.resume
				g.ret, g.err = g.body()
				g.done = true
				g.key, g.val = nil, nil
				g.paused <- struct{}{}
			}()
		}
		g.resume <- struct{}{}
		<-g.paused
		st.stack = st.stack[:len(st.stack)-1]
		if g.done {
			rt.PhpxFrameDetach(g.mark)
		}
		err := g.err
		g.err = nil
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
	if !g.started {
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
	return !g.done, nil
}

// Current is the value at the generator's position.
func (g *Generator) Current() (any, error) {
	if err := g.prime(); err != nil {
		return nil, err
	}
	return g.val, nil
}

// Key is the key at the generator's position.
func (g *Generator) Key() (any, error) {
	if err := g.prime(); err != nil {
		return nil, err
	}
	return g.key, nil
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
	g.sent = v
	g.resumeWith(v)
	if err := g.step(g); err != nil {
		return nil, err
	}
	return g.val, nil
}

func (g *Generator) resumeWith(v any) { g.sent = v }

// GetReturn is the value the body returned.
func (g *Generator) GetReturn() any { return g.ret }
