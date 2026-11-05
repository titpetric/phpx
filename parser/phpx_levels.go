//go:build phpx

package parser

import "strconv"

// phpxLoopLevels reads the number after `break` or `continue`, or 0 when the
// statement ends there.
func (p *parser) phpxLoopLevels() int {
	if p.cur().kind != tInt {
		return 0
	}
	n, err := strconv.Atoi(p.next().val)
	if err != nil {
		return 0
	}
	return n
}
