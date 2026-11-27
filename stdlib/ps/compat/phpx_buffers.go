//go:build phpx

package compat

import "github.com/titpetric/phpx/runner"

// phpxRegisterBuffers adds ob_clean, ob_flush, ob_get_length and
// ob_list_handlers to the buffer stack the other ob_* functions share.
func phpxRegisterBuffers(rt *runner.Runtime, buffers *buffers) {
	// ob_clean drops the active buffer's contents without closing the level.
	rt.RegisterFunc("ob_clean", buffers.clean)
	// ob_flush writes the active buffer's contents to the output below it
	// without closing the level.
	rt.RegisterFunc("ob_flush", buffers.flush)
	// ob_get_length returns the byte count of the active buffer.
	rt.RegisterFunc("ob_get_length", func() any {
		n, ok := buffers.length()
		if !ok {
			return false
		}
		return int64(n)
	})
	// ob_list_handlers names the handler of each active level. A level this
	// runtime opens has one handler, whose name php spells this way.
	rt.RegisterFunc("ob_list_handlers", func() []string {
		out := make([]string, buffers.level())
		for i := range out {
			out[i] = "default output handler"
		}
		return out
	})
}
