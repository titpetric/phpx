package files

import (
	"io"
	"testing"

	"github.com/titpetric/phpx/runner"
)

func TestPhpxWrapper(t *testing.T) {
	for _, scheme := range []string{"stdout", "stderr", "memory", "temp"} {
		if h, ok := phpxWrapper(runner.New(io.Discard, runner.Options{}), scheme); !ok || h == nil {
			t.Errorf("php://%s did not open", scheme)
		}
	}
	if _, ok := phpxWrapper(runner.New(io.Discard, runner.Options{}), "filter"); ok {
		t.Error("php://filter must stay unopened")
	}
}
