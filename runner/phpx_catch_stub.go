//go:build !phpx

package runner

// phpxHostErrorCatches answers the untagged build: an error with no php class
// reaches any clause, as upstream has it.
func phpxHostErrorCatches(_ string) bool { return true }
