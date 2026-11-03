//go:build !phpx

package runner

// phpxStringOf answers the untagged build, where an object with no string
// form echoes as the empty string, as upstream has it.
func phpxStringOf(v any) (string, error) { return phpString(v), nil }
