//go:build phpx

package phpx

import (
	"io"
	"testing"

	"github.com/titpetric/phpx/runner"
)

func TestStreamFilterRegistry(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	registerStreamFilters(rt)
	register, ok := rt.Callable("stream_filter_register")
	if !ok {
		t.Fatal("stream_filter_register is not registered")
	}
	if got, err := register("chunk", "Chunk"); got != true || err != nil {
		t.Fatalf("first register = %v, %v", got, err)
	}
	if got, err := register("chunk", "Other"); got != false || err != nil {
		t.Fatalf("second register = %v, %v", got, err)
	}
	if _, err := register("", "Chunk"); err == nil {
		t.Fatal("an empty name did not raise")
	}
	if names := registeredStreamFilters(rt); len(names) != 1 || names[0] != "chunk" {
		t.Fatalf("names = %v", names)
	}
	registry := filtersFor(rt)
	if registry != filtersFor(rt) {
		t.Fatal("the registry is not stable per runtime")
	}
	if names := registeredStreamFilters(runner.New(io.Discard, runner.Options{})); len(names) != 0 {
		t.Fatalf("a second runtime sees %v", names)
	}
}
