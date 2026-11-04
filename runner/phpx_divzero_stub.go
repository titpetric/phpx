//go:build !phpx

package runner

// phpxDivZero answers the untagged build, where a division by zero is 0.
func phpxDivZero(_ string, _, _ any) error { return nil }
