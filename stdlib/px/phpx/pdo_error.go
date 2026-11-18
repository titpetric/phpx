//go:build phpx

package phpx

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/titpetric/phpx/model"
)

// PDOException is the throwable a failed statement raises. php's PDOException
// extends RuntimeException and carries the SQLSTATE as its code and the three
// part errorInfo array as a property, which is what a wrapping exception reads.
type PDOException struct {
	Msg string
	// ErrorInfo is php's triple: the SQLSTATE, the driver code and the driver
	// message. It is exported so `$e->errorInfo` reads it.
	ErrorInfo *model.Array

	sqlState string
}

// Error renders the message php prints.
func (e *PDOException) Error() string { return e.Msg }

// ThrowableClass names the class a catch clause sees.
func (e *PDOException) ThrowableClass() string { return "PDOException" }

// GetCode answers the SQLSTATE, as php's PDOException does.
func (e *PDOException) GetCode() any { return e.SQLState() }

// GetMessage answers the rendered message.
func (e *PDOException) GetMessage() string { return e.Msg }

// SQLState returns the SQLSTATE code.
func (e *PDOException) SQLState() string { return e.sqlState }

// newPDOException builds the exception PHP renders as
// SQLSTATE[<state>]: <class>: <code> <message>.
func newPDOException(sqlState string, code int64, message string) *PDOException {
	msgs := model.NewArray()
	msgs.Append(sqlState)
	if code > 0 {
		msgs.Append(code)
	} else {
		msgs.Append(nil)
	}
	msgs.Append(message)
	detail := message
	if code > 0 {
		detail = fmt.Sprintf("%d %s", code, message)
	}
	return &PDOException{
		Msg:       fmt.Sprintf("SQLSTATE[%s]: %s: %s", sqlState, sqlStateClass(sqlState), detail),
		ErrorInfo: msgs,
		sqlState:  sqlState,
	}
}

// newPDOExceptionFromError maps a driver error to the PDOException php raises
// for it: sqlite's result code is read off the message the driver appends.
func newPDOExceptionFromError(driver string, err error) *PDOException {
	var exc *PDOException
	if errors.As(err, &exc) {
		return exc
	}
	text := err.Error()
	if driver == "sqlite" {
		return sqliteException(text)
	}
	return newPDOException("HY000", 0, text)
}

// sqliteException reads modernc's "<message> (<extended code>)" and renders
// php's SQLSTATE for the primary result code.
func sqliteException(text string) *PDOException {
	code, message := splitSQLiteCode(text)
	primary := code
	if primary > 255 {
		primary &= 0xff
	}
	state := "HY000"
	if primary == 19 {
		state = "23000"
	}
	return newPDOException(state, primary, message)
}

// splitSQLiteCode separates the trailing "(code)" and the driver's kind prefix
// from the message.
func splitSQLiteCode(text string) (int64, string) {
	code := int64(1)
	if i := strings.LastIndexByte(text, '('); i >= 0 && strings.HasSuffix(text, ")") {
		if n, err := strconv.ParseInt(text[i+1:len(text)-1], 10, 64); err == nil {
			code = n
			text = strings.TrimSpace(text[:i])
		}
	}
	for _, prefix := range []string{"SQL logic error: ", "constraint failed: "} {
		text = strings.TrimPrefix(text, prefix)
	}
	return code, text
}

// sqlStateClass is the class text php puts in the message for a SQLSTATE.
func sqlStateClass(state string) string {
	switch state {
	case "23000":
		return "Integrity constraint violation"
	case "IM001":
		return "Driver does not support this function"
	}
	return "General error"
}
