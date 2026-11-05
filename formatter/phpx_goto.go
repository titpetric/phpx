package formatter

import (
	"strconv"
)

// phpxLoopJump spells break or continue with the level it was written with.
func phpxLoopJump(keyword string, levels int) string {
	if levels > 1 {
		return keyword + " " + strconv.Itoa(levels) + ";"
	}
	return keyword + ";"
}
