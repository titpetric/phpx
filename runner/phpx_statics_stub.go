//go:build !phpx

package runner

// staticStorageFor answers the untagged build, where each class keeps its own
// copy of an inherited static, as upstream does.
func (rt *Runtime) staticStorageFor(class, _ string, scope *Scope) (map[string]any, error) {
	return rt.staticStorage(class, scope)
}
