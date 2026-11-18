//go:build phpx

package phpx

import (
	"io"
	"testing"

	"github.com/titpetric/phpx/runner"
)

func TestStreamWrapperRegistry(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	registry := wrappersFor(rt)
	registry.Lock()
	registry.schemes = append(registry.schemes, "hoa")
	registry.Unlock()
	if got := registeredStreamWrappers(rt); len(got) != 1 || got[0] != "hoa" {
		t.Fatalf("schemes = %v", got)
	}
	if got := registeredStreamWrappers(runner.New(io.Discard, runner.Options{})); len(got) != 0 {
		t.Fatalf("a second runtime sees %v", got)
	}
}

func TestStreamWrapperUnregister(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	registerStreamWrappers(rt)
	unregister, ok := rt.Callable("stream_wrapper_unregister")
	if !ok {
		t.Fatal("stream_wrapper_unregister is not registered")
	}
	if got, err := unregister("hoa"); got != false || err != nil {
		t.Fatalf("unregister unknown = %v, %v", got, err)
	}
	registry := wrappersFor(rt)
	registry.Lock()
	registry.schemes = append(registry.schemes, "hoa")
	registry.Unlock()
	if got, err := unregister("hoa"); got != true || err != nil {
		t.Fatalf("unregister known = %v, %v", got, err)
	}
	if got := registeredStreamWrappers(rt); len(got) != 0 {
		t.Fatalf("schemes after unregister = %v", got)
	}
}
