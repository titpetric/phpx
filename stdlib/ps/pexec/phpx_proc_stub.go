//go:build !phpx

package pexec

import "github.com/titpetric/phpx/runner"

// registerPhpxProc answers the untagged build, which starts no child process
// of its own: proc_open arrives with the fork's area.
func registerPhpxProc(*runner.Runtime, root) {}
