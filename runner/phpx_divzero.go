//go:build phpx

package runner

import "github.com/titpetric/phpx/internal/phpval"

// phpxDivZero answers php's DivisionByZeroError for `/` and `%` with a zero
// divisor. The message differs by operator, and `%` casts to int first, so
// 0.5 as a divisor is zero there and a plain float for `/`.
func phpxDivZero(op string, a, b any) error {
	if op != "/" && op != "%" {
		return nil
	}
	if s, ok := b.(string); ok {
		if n, ok := phpval.StringNumber(s); ok {
			b = n
		}
	}
	if op == "%" {
		if toInt(b) == 0 {
			return &DivisionByZeroError{Message: "Modulo by zero"}
		}
		return nil
	}
	if isFloat(b) {
		if toFloat(b) == 0 {
			return &DivisionByZeroError{Message: "Division by zero"}
		}
		return nil
	}
	if toInt(b) == 0 {
		return &DivisionByZeroError{Message: "Division by zero"}
	}
	return nil
}
