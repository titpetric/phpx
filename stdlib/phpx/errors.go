//go:build phpx

package phpx

import (
	"context"

	"github.com/titpetric/phpx/runner"
	"github.com/titpetric/phpx/telemetry"
)

func init() {
	runner.RegisterBinding(registerPhpxErrorHandlers)
}

// phpxErrorHandler is one set_error_handler() registration.
type phpxErrorHandler struct {
	callback any
	levels   int64
}

// registerPhpxErrorHandlers installs the handler stack. trigger_error calls
// the handler on top of it when its level is in the mask; with none, or when
// the handler answers false, the message goes unreported, as a script run with
// display_errors off.
func registerPhpxErrorHandlers(rt *runner.Runtime) {
	var stack []phpxErrorHandler
	rt.RegisterFunc("set_error_handler", func(callback any, levels ...int64) any {
		mask := int64(32767)
		if len(levels) > 0 {
			mask = levels[0]
		}
		var previous any
		if len(stack) > 0 {
			previous = stack[len(stack)-1].callback
		}
		stack = append(stack, phpxErrorHandler{callback: callback, levels: mask})
		return previous
	})
	rt.RegisterFunc("restore_error_handler", func() bool {
		if len(stack) > 0 {
			stack = stack[:len(stack)-1]
		}
		return true
	})
	trigger := func(ctx context.Context, message string, level ...int64) (bool, error) {
		kind := int64(1024)
		if len(level) > 0 {
			kind = level[0]
		}
		if len(stack) == 0 {
			return true, nil
		}
		top := stack[len(stack)-1]
		if top.callback == nil || top.levels&kind == 0 {
			return true, nil
		}
		call, ok := rt.Callable(top.callback)
		if !ok {
			return true, nil
		}
		file, line := telemetry.SpanSource(ctx)
		_, err := call(kind, message, file, int64(line))
		return true, err
	}
	rt.RegisterFunc("trigger_error", trigger)
	rt.RegisterFunc("user_error", trigger)
}
