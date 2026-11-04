package runner

import (
	"errors"
	"io/fs"
)

// layeredFS serves the primary tree first and falls back to the vendor tree
// for a path the primary does not hold.
type layeredFS struct {
	primary, vendor fs.FS
}

// Open reads from the primary tree, then the vendor tree.
func (l layeredFS) Open(name string) (fs.File, error) {
	if l.primary != nil {
		f, err := l.primary.Open(name)
		if err == nil || !errors.Is(err, fs.ErrNotExist) {
			return f, err
		}
	}
	return l.vendor.Open(name)
}

// Vendor mounts a source tree, typically a go:embed of a composer vendor
// directory, behind the root filesystem. A path the root holds wins, so an
// application overrides a vendored file by shipping its own.
//
// Mounting is opt in, and a runtime that never calls it reads the root alone.
// When Options.Include is empty, Vendor sets it to vendor/autoload.php so
// composer's loader registers before the entrypoint runs.
func (rt *Runtime) Vendor(tree fs.FS) {
	if tree == nil {
		return
	}
	rt.opts.RootFS = layeredFS{primary: rt.opts.RootFS, vendor: tree}
	if rt.opts.Include == "" {
		rt.opts.Include = "vendor/autoload.php"
	}
}
