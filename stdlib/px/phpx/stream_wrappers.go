//go:build phpx

package phpx

import (
	"fmt"
	"sync"

	"github.com/titpetric/phpx/runner"
)

// A stream wrapper registry is what hoa/stream reaches for at load: a package
// registers a class for a protocol and stream_get_wrappers() reports it. The
// scheme is recorded and the class name kept; opening a registered scheme is
// not bound, so a read through one fails like any unknown wrapper.
var streamWrappers struct {
	sync.Mutex
	byRuntime map[*runner.Runtime]*wrapperRegistry
}

type wrapperRegistry struct {
	sync.Mutex
	schemes []string
}

// wrappersFor answers the wrapper registry of one runtime, making it on first use.
func wrappersFor(rt *runner.Runtime) *wrapperRegistry {
	streamWrappers.Lock()
	defer streamWrappers.Unlock()
	if streamWrappers.byRuntime == nil {
		streamWrappers.byRuntime = map[*runner.Runtime]*wrapperRegistry{}
	}
	registry, ok := streamWrappers.byRuntime[rt]
	if !ok {
		registry = &wrapperRegistry{}
		streamWrappers.byRuntime[rt] = registry
	}
	return registry
}

func init() {
	runner.PhpxRegisterLate(registerStreamWrappers)
}

// registerStreamWrappers installs stream_wrapper_register(), unregister() and
// restore().
func registerStreamWrappers(rt *runner.Runtime) {
	// stream_wrapper_register records the class that answers a scheme. php
	// refuses a class it cannot load and warns about a scheme it already
	// carries, and this binding does both.
	rt.RegisterFunc("stream_wrapper_register", func(protocol, class string, _ ...any) (bool, error) {
		if ok, err := rt.ClassExists(class, true); !ok {
			if err != nil {
				return false, err
			}
			return false, &streamTypeError{fmt.Sprintf("stream_wrapper_register(): Argument #2 ($class) must be a valid class name, %s given", class)}
		}
		registry := wrappersFor(rt)
		registry.Lock()
		defer registry.Unlock()
		for _, scheme := range registry.schemes {
			if scheme == protocol {
				rt.PhpxWarning(2, fmt.Sprintf("stream_wrapper_register(): Protocol %s:// is already defined.", protocol))
				return false, nil
			}
		}
		registry.schemes = append(registry.schemes, protocol)
		return true, nil
	})
	// stream_wrapper_unregister drops a scheme this binding registered, and
	// warns about one it does not carry.
	rt.RegisterFunc("stream_wrapper_unregister", func(protocol string) bool {
		registry := wrappersFor(rt)
		registry.Lock()
		defer registry.Unlock()
		for i, scheme := range registry.schemes {
			if scheme == protocol {
				registry.schemes = append(registry.schemes[:i], registry.schemes[i+1:]...)
				return true
			}
		}
		rt.PhpxWarning(2, fmt.Sprintf("stream_wrapper_unregister(): Unable to unregister protocol %s://", protocol))
		return false
	})
	// stream_wrapper_restore answers false for every scheme: the binding
	// carries no built-in scheme it could restore.
	rt.RegisterFunc("stream_wrapper_restore", func(protocol string) bool {
		rt.PhpxWarning(2, fmt.Sprintf("stream_wrapper_restore(): Unable to restore protocol %s://", protocol))
		return false
	})
}

// registeredStreamWrappers answers the schemes a script registered.
func registeredStreamWrappers(rt *runner.Runtime) []string {
	registry := wrappersFor(rt)
	registry.Lock()
	defer registry.Unlock()
	return append([]string(nil), registry.schemes...)
}

// streamTypeError is a TypeError with php's wording for an argument this binding
// refuses before the call reaches a typed Go parameter.
type streamTypeError struct{ message string }

func (e *streamTypeError) Error() string          { return e.message }
func (e *streamTypeError) ThrowableClass() string { return "TypeError" }
