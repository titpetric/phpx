//go:build phpx

package phpx

import (
	"context"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
	"github.com/titpetric/phpx/telemetry"
)

func init() {
	runner.RegisterBinding(registerPhpxExceptions)
}

// Throwable is a built-in exception with the third constructor argument,
// `new Exception($message, $code, $previous)`, and the accessors php answers
// for a throwable, with the file and line it was created at.
type Throwable struct {
	Message  string `json:"message"`
	Code     int    `json:"code"`
	Class    string `json:"class"`
	previous any
	file     string
	line     int
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

// GetFile returns the file the throwable was created in.
func (t *Throwable) GetFile() string { return t.file }

// GetLine returns the line the throwable was created on.
func (t *Throwable) GetLine() int { return t.line }

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
		rt.RegisterConstructor(class, func(ctx context.Context, message string, code int, previous any) (*Throwable, error) {
			file, line := telemetry.SpanSource(ctx)
			return &Throwable{Message: message, Code: code, Class: class, previous: previous, file: file, line: line}, nil
		})
	}
}
