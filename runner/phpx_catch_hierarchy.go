//go:build phpx

package runner

// phpxCatchHierarchy is whether a catch clause naming a parent takes a child
// built-in throwable: it does in phpx.
const phpxCatchHierarchy = true
