//go:build phpx

package coverage

// The fork marks a line rather than counting executions by default: the
// profile asks which lines ran, and a mark leaves every later pass a load.
// --covermode=count asks for the counter. An untagged build keeps phpscript's
// default, which is why this is an override and not a second copy of the
// package.
func init() { DefaultMode = ModeAtomic }
