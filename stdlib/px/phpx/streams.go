//go:build phpx

package phpx

import (
	"sync"

	"github.com/titpetric/phpx/runner"
)

// The stream filter registry is what composer's autoload files reach for: a
// package registers a filter class at load time and the application never
// appends it. The name is recorded and stream_get_filters() reports it, as
// php does; appending a filter is not bound, so nothing filters a stream yet.
// The registry is per runtime, because a fixture or a request starts with php's
// empty one.
var streamFilters struct {
	sync.Mutex
	byRuntime map[*runner.Runtime]*filterRegistry
}

type filterRegistry struct {
	sync.Mutex
	names []string
}

// filtersFor answers the registry of one runtime, making it on first use.
func filtersFor(rt *runner.Runtime) *filterRegistry {
	streamFilters.Lock()
	defer streamFilters.Unlock()
	if streamFilters.byRuntime == nil {
		streamFilters.byRuntime = map[*runner.Runtime]*filterRegistry{}
	}
	registry, ok := streamFilters.byRuntime[rt]
	if !ok {
		registry = &filterRegistry{}
		streamFilters.byRuntime[rt] = registry
	}
	return registry
}

func init() {
	runner.PhpxRegisterLate(registerStreamFilters)
}

// registerStreamFilters installs stream_filter_register().
func registerStreamFilters(rt *runner.Runtime) {
	// stream_filter_register records the class that answers a filter name and
	// answers false for a name that already has one, as php does. An empty
	// name or class is php's ValueError, numbered as php numbers it.
	rt.RegisterFunc("stream_filter_register", func(name string, class string) (bool, error) {
		if name == "" {
			return false, &valueError{"stream_filter_register(): Argument #1 ($filter_name) must be a non-empty string"}
		}
		if class == "" {
			return false, &valueError{"stream_filter_register(): Argument #2 ($class) must be a non-empty string"}
		}
		registry := filtersFor(rt)
		registry.Lock()
		defer registry.Unlock()
		for _, existing := range registry.names {
			if existing == name {
				return false, nil
			}
		}
		registry.names = append(registry.names, name)
		return true, nil
	})
}

// registeredStreamFilters answers the names a script registered, in the order
// php's registry reports them.
func registeredStreamFilters(rt *runner.Runtime) []string {
	registry := filtersFor(rt)
	registry.Lock()
	defer registry.Unlock()
	return append([]string(nil), registry.names...)
}
