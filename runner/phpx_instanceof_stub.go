//go:build !phpx

package runner

// phpxHostThrowableIs adds no answer untagged: upstream's instanceof stands.
func phpxHostThrowableIs(_ any, _ string) bool { return false }

// PhpxClassOfError adds no answer untagged: get_class reports the Go type name.
func PhpxClassOfError(_ any) (string, bool) { return "", false }
