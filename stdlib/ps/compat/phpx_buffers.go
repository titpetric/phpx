//go:build phpx

package compat

import "github.com/titpetric/phpx/runner"

// phpxRegisterBuffers adds ob_clean, ob_flush, ob_get_length and
// ob_list_handlers to the buffer stack the other ob_* functions share.
func phpxRegisterBuffers(rt *runner.Runtime, buffers *buffers) {
	rt.RegisterFunc("ob_clean", func() bool {
		if _, ok := buffers.pop(false); !ok {
			return false
		}
		buffers.push()
		return true
	})
	rt.RegisterFunc("ob_flush", func() bool {
		if _, ok := buffers.pop(true); !ok {
			return false
		}
		buffers.push()
		return true
	})
	rt.RegisterFunc("ob_get_length", func() any {
		contents, ok := buffers.contents()
		if !ok {
			return false
		}
		return int64(len(contents))
	})
	rt.RegisterFunc("ob_list_handlers", func() []string {
		out := make([]string, buffers.level())
		for i := range out {
			out[i] = "default output handler"
		}
		return out
	})
}
