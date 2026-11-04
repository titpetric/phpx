//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxExceptions)
}

// Throwable is a built-in exception with the third constructor argument,
// `new Exception($message, $code, $previous)`, and the accessors php answers
// for a throwable. The file and line of the throw are not recorded.
type Throwable struct {
	Message  string `json:"message"`
	Code     int    `json:"code"`
	Class    string `json:"class"`
	previous any
}

// ThrowableClass names the class a catch clause filters on.
func (t *Throwable) ThrowableClass() string { return t.Class }

// GetMessage returns the message.
func (t *Throwable) GetMessage() string { return t.Message }

// GetCode returns the code.
func (t *Throwable) GetCode() int { return t.Code }

// Error is the message, which makes the value a Go error.
func (t *Throwable) Error() string { return t.Message }

// GetPrevious returns the exception this one wraps, or null.
func (t *Throwable) GetPrevious() any { return t.previous }

// GetFile returns the file of the throw, empty here.
func (t *Throwable) GetFile() string { return "" }

// GetLine returns the line of the throw, 0 here.
func (t *Throwable) GetLine() int { return 0 }

// GetTrace returns the call trace, always empty here.
func (t *Throwable) GetTrace() *model.Array { return model.NewArray() }

// GetTraceAsString returns the trace in php's text form, which is the bare
// entry of an empty one.
func (t *Throwable) GetTraceAsString() string { return "#0 {main}" }

// phpxThrowables are the built-in classes whose constructor takes a previous.
var phpxThrowables = []string{
	"Exception", "ErrorException", "RuntimeException", "LogicException",
	"InvalidArgumentException", "DomainException", "LengthException",
	"OutOfRangeException", "OutOfBoundsException", "RangeException",
	"OverflowException", "UnderflowException", "UnexpectedValueException",
	"BadFunctionCallException", "BadMethodCallException", "JsonException",
	"Error", "TypeError", "ValueError", "ArithmeticError", "DivisionByZeroError",
	"ArgumentCountError",
}

// registerPhpxExceptions replaces upstream's two-argument constructors.
func registerPhpxExceptions(rt *runner.Runtime) {
	for _, class := range phpxThrowables {
		rt.RegisterConstructor(class, func(message string, code int, previous any) (*Throwable, error) {
			return &Throwable{Message: message, Code: code, Class: class, previous: previous}, nil
		})
	}
}
