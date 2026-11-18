//go:build phpx

package phpx

import (
	"errors"
	"testing"
)

func TestSQLiteException(t *testing.T) {
	exc := sqliteException("constraint failed: UNIQUE constraint failed: item.id (1555)")
	if exc.SQLState() != "23000" {
		t.Errorf("state = %q", exc.SQLState())
	}
	if want := "SQLSTATE[23000]: Integrity constraint violation: 19 UNIQUE constraint failed: item.id"; exc.Error() != want {
		t.Errorf("message = %q, want %q", exc.Error(), want)
	}
	if v, _ := exc.ErrorInfo.Get(int64(1)); v != int64(19) {
		t.Errorf("driver code = %v", v)
	}
	if v, _ := exc.ErrorInfo.Get(int64(2)); v != "UNIQUE constraint failed: item.id" {
		t.Errorf("driver message = %v", v)
	}
}

func TestSQLiteSyntaxError(t *testing.T) {
	exc := sqliteException(`SQL logic error: near "bad": syntax error (1)`)
	if exc.SQLState() != "HY000" || exc.Error() != `SQLSTATE[HY000]: General error: 1 near "bad": syntax error` {
		t.Fatalf("exception = %q", exc.Error())
	}
}

func TestPDOExceptionPassthrough(t *testing.T) {
	original := newPDOException("HY000", 1, "no such table: x")
	if got := newPDOExceptionFromError("sqlite", original); got != original {
		t.Fatal("a PDOException was wrapped again")
	}
	if got := newPDOExceptionFromError("mysql", errors.New("dial tcp")); got.SQLState() != "HY000" {
		t.Fatalf("state = %q", got.SQLState())
	}
}
