//go:build phpx

package phpx

import (
	"io"
	"runtime"
	"testing"
	"time"

	"github.com/titpetric/phpx/runner"
)

// TestAbandonedGeneratorStopsItsGoroutine pins the lifetime of the goroutine a
// generator starts. A foreach that breaks leaves the body parked on its resume
// channel, and the generator is unreachable from the script then, so the
// runtime closes its stop channel and the body ends. The goroutine count
// returns to its baseline once the abandoned generators are collected.
//
// The php column is not applicable: php bounds no Go goroutine, so the fixture
// matrix cannot express this. The test drives the binding directly.
func TestAbandonedGeneratorStopsItsGoroutine(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	registerPhpxGenerators(rt)
	rawGen, ok := rt.LookupFunc("__phpx_generator")
	if !ok {
		t.Fatal("__phpx_generator is not registered")
	}
	rawYield, ok := rt.LookupFunc("__phpx_yield")
	if !ok {
		t.Fatal("__phpx_yield is not registered")
	}
	yield := rawYield.(func(any, any, bool) any)
	newGenerator := rawGen.(func(any) (*Generator, error))

	runtime.GC()
	baseline := runtime.NumGoroutine()

	// Each call drops its generator on return. Ten thousand bodies park on
	// their resume channel; without a stop signal they stay parked for the
	// life of the process.
	abandon := func() {
		g, err := newGenerator(func() (any, error) {
			yield(nil, int64(1), false)
			yield(nil, int64(2), false)
			return nil, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		v, err := g.Current()
		if err != nil {
			t.Fatal(err)
		}
		if v != int64(1) {
			t.Fatalf("current = %v, want 1", v)
		}
	}
	for i := 0; i < 10000; i++ {
		abandon()
	}

	deadline := time.Now().Add(10 * time.Second)
	for {
		runtime.GC()
		if n := runtime.NumGoroutine(); n <= baseline+2 {
			return
		} else if time.Now().After(deadline) {
			t.Fatalf("goroutines = %d, baseline = %d: abandoned generators still parked", n, baseline)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
