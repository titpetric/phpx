// Package internals holds PHP's internal memory and runtime introspection bindings.
package internals

import (
	"github.com/titpetric/phpx/runner"
)

// init contributes the internals bindings to stdlib.Register.
func init() {
	runner.RegisterArea(`Runtime\PS\Internals`, Register)
}

// Register installs internals symbols into the runtime.
func Register(rt *runner.Runtime) {
	// memory_get_usage reports the live memory estimate of the running script
	// in bytes. With $real_usage it reports the bytes the Go heap has taken
	// from the system, and without it the estimate a walk of the live values
	// answers.
	rt.RegisterFunc("memory_get_usage", func(realUsage ...bool) int64 {
		if len(realUsage) > 0 && realUsage[0] {
			return rt.RealMemoryUsage()
		}
		return rt.MemoryUsage()
	})
	// memory_get_peak_usage reports the high-water mark of memory_get_usage,
	// reading $real_usage the same way.
	rt.RegisterFunc("memory_get_peak_usage", func(realUsage ...bool) int64 {
		if len(realUsage) > 0 && realUsage[0] {
			return rt.RealMemoryPeak()
		}
		return rt.MemoryPeak()
	})
}
