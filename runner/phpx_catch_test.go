//go:build phpx

package runner

import "testing"

func TestPhpxHostErrorCatches(t *testing.T) {
	for _, name := range []string{"Exception", "\\RuntimeException", "valueerror", "JsonException"} {
		if !phpxHostErrorCatches(name) {
			t.Errorf("%s should take a host error", name)
		}
	}
	for _, name := range []string{"PHPUnit\\Framework\\IncompleteTest", "Nette\\Schema\\ValidationException"} {
		if phpxHostErrorCatches(name) {
			t.Errorf("%s should not take a host error", name)
		}
	}
}
