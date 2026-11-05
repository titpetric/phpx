//go:build !phpx

package runner

// phpxAliasMatches answers the untagged build, which has no interface aliases.
func phpxAliasMatches(_, _ string) bool { return false }
