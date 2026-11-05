//go:build phpx

package phpx

import (
	"io"
	"os"
	"testing"

	"github.com/titpetric/phpx/runner"
)

func TestTmpfileIsAnonymous(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	registerTmpfile(rt)
	fn, _ := rt.Callable("tmpfile")
	h, err := fn()
	if err != nil {
		t.Fatal(err)
	}
	f, ok := h.(*os.File)
	if !ok {
		t.Fatalf("tmpfile gave %T", h)
	}
	defer f.Close()
	if _, err := f.WriteString("abc"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(f.Name()); err == nil {
		t.Fatal("the file should be unlinked")
	}
}

func TestUmaskRestores(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	registerTmpfile(rt)
	fn, _ := rt.Callable("umask")
	first, _ := fn()
	second, _ := fn()
	if first != second {
		t.Fatalf("reading the mask changed it: %v then %v", first, second)
	}
}
