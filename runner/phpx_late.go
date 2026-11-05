//go:build phpx

package runner

import "sync"

// A late binding installs after the filesystem functions are rooted: RegisterFS
// registers them again against the script's directory and would replace a
// binding that overrode one of them earlier.
var phpxLate struct {
	sync.Mutex
	installers []func(*Runtime)
}

// PhpxRegisterLate contributes an installer that runs after RegisterFS.
func PhpxRegisterLate(installer func(*Runtime)) {
	phpxLate.Lock()
	defer phpxLate.Unlock()
	phpxLate.installers = append(phpxLate.installers, installer)
}

// PhpxRunLate runs the late installers on rt, in registration order.
func PhpxRunLate(rt *Runtime) {
	phpxLate.Lock()
	installers := append([](func(*Runtime)){}, phpxLate.installers...)
	phpxLate.Unlock()
	for _, install := range installers {
		install(rt)
	}
}
