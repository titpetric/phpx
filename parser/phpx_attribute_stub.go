//go:build !phpx

package parser

// phpxSkipAttribute answers the untagged build, where `#[` starts a line comment.
func (l *lexer) phpxSkipAttribute() bool { return false }
