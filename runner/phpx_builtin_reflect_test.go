//go:build phpx

package runner

import (
	"io"
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxGoParamInfo(t *testing.T) {
	info := phpxGoParamInfo(func(a string, rest ...int) {})
	if info.Len() != 2 {
		t.Fatalf("got %d parameters, want 2", info.Len())
	}
	last, _ := info.Get(int64(1))
	if v, _ := last.(*model.Array).Get("variadic"); v != true {
		t.Fatal("the variadic Go parameter is not variadic")
	}
}

func TestPhpxNamedCallableUnknownName(t *testing.T) {
	rt := New(io.Discard, Options{})
	if _, ok := rt.PhpxNamedCallable("no_such_function_here"); ok {
		t.Fatal("an unknown name produced a callable")
	}
}

func TestPhpxIsClosure(t *testing.T) {
	if !PhpxIsClosure(&Callable{}) || PhpxIsClosure("strlen") {
		t.Fatal("only a *Callable is a closure")
	}
}
