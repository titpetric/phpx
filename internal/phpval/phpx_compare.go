package phpval

// hostComparer is a Go value that orders itself against another value of its kind.
type hostComparer interface {
	PhpxCompare(other any) (int, bool)
}

// hostCompare asks either side of a comparison to order the pair.
func hostCompare(x, y any) (int, bool) {
	if c, ok := x.(hostComparer); ok {
		if r, ok := c.PhpxCompare(y); ok {
			return r, true
		}
	}
	if c, ok := y.(hostComparer); ok {
		if r, ok := c.PhpxCompare(x); ok {
			return -r, true
		}
	}
	return 0, false
}
