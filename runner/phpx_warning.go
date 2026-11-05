package runner

// PhpxSetWarningSink installs the receiver of PhpxWarning for this runtime: the
// handler stack of set_error_handler. With none set a warning is dropped, as a
// script run quietly drops it.
func (rt *Runtime) PhpxSetWarningSink(sink func(level int64, message string)) {
	rt.phpxSink = sink
}

// PhpxWarning raises a warning of the given E_* level with message, for a
// failure the runtime reports on a script's behalf, a failed fopen among them.
func (rt *Runtime) PhpxWarning(level int64, message string) {
	if rt.phpxSink != nil {
		rt.phpxSink(level, message)
	}
}
