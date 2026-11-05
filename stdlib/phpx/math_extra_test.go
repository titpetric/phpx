//go:build phpx

package phpx

import (
	"io"
	"math"
	"testing"

	"github.com/titpetric/phpx/runner"
)

func TestMathExtra(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	registerMathExtra(rt)
	call := func(name string, args ...any) any {
		fn, ok := rt.Callable(name)
		if !ok {
			t.Fatalf("%s is not registered", name)
		}
		v, err := fn(args...)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	if got := call("atan2", 1.0, 1.0).(float64); math.Abs(got-math.Pi/4) > 1e-12 {
		t.Errorf("atan2 = %v", got)
	}
	if got := call("rad2deg", math.Pi).(float64); math.Abs(got-180) > 1e-12 {
		t.Errorf("rad2deg = %v", got)
	}
	if got := call("hypot", 3.0, 4.0); got != 5.0 {
		t.Errorf("hypot = %v", got)
	}
	if got := call("getrandmax"); got != int64(2147483647) {
		t.Errorf("getrandmax = %v", got)
	}
}
