//go:build phpx

package phpx

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
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

// TestPDOMemoryKeepsDataAcrossReopen pins what sqliteCreateFunction relies on:
// reopening the pool for a registration that came after a statement leaves an
// in-memory database's rows in place, because the keeper connection holds the
// named database.
func TestPDOMemoryKeepsDataAcrossReopen(t *testing.T) {
	ctx := context.Background()
	rt := runner.New(io.Discard, runner.Options{})
	p, err := newPDO(rt, ctx, "sqlite::memory:")
	if err != nil {
		t.Fatal(err)
	}
	if !p.memory || p.keeper == nil {
		t.Fatalf("memory = %v, keeper = %v", p.memory, p.keeper)
	}
	if _, err := p.Exec(ctx, "create table t (x integer)"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Exec(ctx, "insert into t values (1)"); err != nil {
		t.Fatal(err)
	}
	p.reopen()
	var n int64
	if err := p.db.QueryRowContext(ctx, "select count(*) from t").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("rows after reopen = %d, want 1", n)
	}
}

// TestPDOMemoryHandlesAreSeparate pins that the name a memory database takes
// is per handle, so two PDOs keep two databases as php's do.
func TestPDOMemoryHandlesAreSeparate(t *testing.T) {
	ctx := context.Background()
	rt := runner.New(io.Discard, runner.Options{})
	first, err := newPDO(rt, ctx, "sqlite::memory:")
	if err != nil {
		t.Fatal(err)
	}
	second, err := newPDO(rt, ctx, "sqlite::memory:")
	if err != nil {
		t.Fatal(err)
	}
	if first.dsn == second.dsn {
		t.Fatalf("both handles use %q", first.dsn)
	}
	if _, err := first.Exec(ctx, "create table t (x integer)"); err != nil {
		t.Fatal(err)
	}
	if _, err := second.db.QueryContext(ctx, "select * from t"); err == nil {
		t.Fatal("the second handle sees the first handle's table")
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
