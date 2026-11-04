//go:build !phpx

package parser

// phpxResolveTypeHint answers the untagged build, which keeps a type as written.
func (p *parser) phpxResolveTypeHint(hint string) string { return hint }
