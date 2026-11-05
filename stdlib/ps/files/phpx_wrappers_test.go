package files

import (
	"io"
	"testing"

	"github.com/titpetric/phpx/runner"
)

func TestPhpxWrappersAreRegistered(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	for _, scheme := range []string{"stdout", "stderr", "memory", "temp"} {
		if h, ok, err := openWrapper(rt, scheme, "r"); !ok || err != nil || h == nil {
			t.Errorf("php://%s did not open", scheme)
		}
	}
	if _, ok, _ := openWrapper(rt, "filter", "r"); ok {
		t.Error("php://filter must stay unopened")
	}
}
