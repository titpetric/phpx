package astcodec

import (
	"fmt"
)

// chunk hands out nodes of one type, every node at a distinct address.
type chunk[T any] struct {
	free []T
	size int
}

const (
	chunkMin = 16
	chunkMax = 128
)

func (c *chunk[T]) new() *T {
	if len(c.free) == 0 {
		switch {
		case c.size == 0:
			c.size = chunkMin
		case c.size < chunkMax:
			c.size *= 2
		}
		c.free = make([]T, c.size)
	}
	n := &c.free[0]
	c.free = c.free[1:]
	return n
}

// typeName names the dynamic type of a value the encoder does not know.
func typeName(v any) string { return fmt.Sprintf("%T", v) }
