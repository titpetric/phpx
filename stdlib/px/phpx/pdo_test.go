//go:build phpx

package phpx

import (
	"strings"
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPDODSN(t *testing.T) {
	driver, goDriver, goDSN, err := pdoDSN("sqlite:/tmp/app.db", "", "")
	if err != nil || driver != "sqlite" || goDriver != "sqlite" || goDSN != "/tmp/app.db" {
		t.Fatalf("sqlite = %q %q %q %v", driver, goDriver, goDSN, err)
	}
	if _, _, goDSN, _ := pdoDSN("sqlite::memory:", "", ""); goDSN != ":memory:" {
		t.Fatalf("memory dsn = %q", goDSN)
	}
	driver, goDriver, goDSN, err = pdoDSN("mysql:host=db;port=3307;dbname=app;charset=utf8mb4", "user", "pass")
	if err != nil || driver != "mysql" || goDriver != "mysql" {
		t.Fatalf("mysql = %q %q %v", driver, goDriver, err)
	}
	if want := "user:pass@tcp(db:3307)/app?"; !strings.HasPrefix(goDSN, want) {
		t.Fatalf("mysql go dsn = %q", goDSN)
	}
	if !strings.Contains(goDSN, "charset=utf8mb4") {
		t.Fatalf("mysql charset missing from %q", goDSN)
	}
	driver, goDriver, goDSN, err = pdoDSN("pgsql:host=db;port=5433;dbname=app", "user", "pass")
	if err != nil || driver != "pgsql" || goDriver != "pgx" {
		t.Fatalf("pgsql = %q %q %v", driver, goDriver, err)
	}
	if want := "postgres://user:pass@db:5433/app"; !strings.HasPrefix(goDSN, want) {
		t.Fatalf("pgsql go dsn = %q", goDSN)
	}
	if _, _, _, err := pdoDSN("oracle:db", "", ""); err == nil || !strings.Contains(err.Error(), "could not find driver") {
		t.Fatalf("oracle err = %v", err)
	}
}

func TestPDOOptionMap(t *testing.T) {
	options := model.NewArray()
	options.Set(int64(attrErrMode), int64(errmodeSilent))
	options.Set(int64(attrDefaultFetchMode), int64(fetchObj))
	got := pdoOptionMap(options)
	if len(got) != 2 || got[attrErrMode] != int64(errmodeSilent) || got[attrDefaultFetchMode] != int64(fetchObj) {
		t.Fatalf("options = %v", got)
	}
}

func TestPDOMemoryDSN(t *testing.T) {
	for _, dsn := range []string{":memory:", "file::memory:", "/tmp/app.db?mode=memory"} {
		if !isMemoryDSN(dsn) {
			t.Errorf("%q is an in-memory dsn", dsn)
		}
	}
	for _, dsn := range []string{"/tmp/app.db", "file:/tmp/app.db?cache=shared"} {
		if isMemoryDSN(dsn) {
			t.Errorf("%q is a file dsn", dsn)
		}
	}
}

func TestPDOBindValue(t *testing.T) {
	if got := pdoBindValue(true, paramBool); got != int64(1) {
		t.Errorf("bool true = %v", got)
	}
	if got := pdoBindValue(false, paramBool); got != int64(0) {
		t.Errorf("bool false = %v", got)
	}
	if got := pdoBindValue("42", paramInt); got != int64(42) {
		t.Errorf("int = %v", got)
	}
	if got := pdoBindValue(int64(42), paramStr); got != "42" {
		t.Errorf("str = %v", got)
	}
	if got := pdoBindValue(false, paramStr); got != "" {
		t.Errorf("false as str = %q", got)
	}
	if got := pdoBindValue("x", paramNull); got != nil {
		t.Errorf("null = %v", got)
	}
}
