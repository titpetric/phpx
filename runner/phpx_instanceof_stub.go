//go:build !phpx

package runner

// phpxHostThrowableIs adds no answer untagged: upstream's instanceof stands.
func phpxHostThrowableIs(_ any, _ string) bool { return false }
