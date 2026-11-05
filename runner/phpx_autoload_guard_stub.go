//go:build !phpx

package runner

// phpxEnterAutoload never refuses untagged.
func phpxEnterAutoload(_ *Runtime, _ string) (func(), bool) { return func() {}, true }
