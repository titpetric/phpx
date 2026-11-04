package phpval

// IsNumericString reports whether s is numeric as php's is_numeric() reads it.
func IsNumericString(s string) bool {
	_, ok := numericString(s)
	return ok
}
