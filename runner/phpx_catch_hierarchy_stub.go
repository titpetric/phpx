//go:build !phpx

package runner

// phpxCatchHierarchy is whether a catch clause naming a parent takes a child
// built-in throwable: upstream keeps no hierarchy.
const phpxCatchHierarchy = false
