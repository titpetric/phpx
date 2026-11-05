//go:build phpx

package phpx

import (
	"bytes"
	"io"
	"testing"

	"github.com/titpetric/phpx/runner"
)

func TestResourcesClosedState(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	rt.RegisterFunc("fclose", func(any) bool { return true })
	rt.RegisterFunc("gettype", func(any) string { return "other" })
	rt.RegisterFunc("get_debug_type", func(any) string { return "other" })
	registerResources(rt)
	h := &bytes.Buffer{}
	gettype, _ := rt.Callable("gettype")
	if v, _ := gettype(h); v != "resource" {
		t.Fatalf("open = %v", v)
	}
	fclose, _ := rt.Callable("fclose")
	_, _ = fclose(h)
	if v, _ := gettype(h); v != "resource (closed)" {
		t.Fatalf("closed = %v", v)
	}
	isResource, _ := rt.Callable("is_resource")
	if v, _ := isResource(h); v != false {
		t.Fatalf("is_resource of a closed handle = %v", v)
	}
}
