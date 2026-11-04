package files

import "testing"

func TestPhpxWrapper(t *testing.T) {
	for _, scheme := range []string{"stdout", "stderr"} {
		if h, ok := phpxWrapper(scheme); !ok || h == nil {
			t.Errorf("php://%s did not open", scheme)
		}
	}
	if _, ok := phpxWrapper("memory"); ok {
		t.Error("php://filter must stay unopened")
	}
}
