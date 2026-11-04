package parser

// phpxClassOf consumes `::class` after an expression and reports whether it
// did. `$obj::class` lowers to `get_class($obj)`, which is what php answers
// for an object; a literal `Name::class` never reaches it because the name
// parser reads that form first.
func (p *parser) phpxClassOf() bool {
	if p.isOp("::") && p.peek(1).val == "class" {
		p.next()
		p.next()
		return true
	}
	return false
}
