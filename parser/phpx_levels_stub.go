//go:build !phpx

package parser

// phpxLoopLevels answers the untagged build, which reads no number after
// `break` and leaves it to the statement terminator.
func (p *parser) phpxLoopLevels() int { return 0 }
