//go:build !phpx

package runner

// PhpxRegisterLate drops the installer untagged: upstream has no late phase.
func PhpxRegisterLate(_ func(*Runtime)) {}

// PhpxRunLate does nothing untagged.
func PhpxRunLate(_ *Runtime) {}
