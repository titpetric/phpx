package phpval

// IsNumericString reports whether s is a numeric string as php's is_numeric()
// reads one: an integer or a float, with whitespace allowed on both sides.
func IsNumericString(s string) bool {
	_, ok := numericString(s)
	return ok
}
