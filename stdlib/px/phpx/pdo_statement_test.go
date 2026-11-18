//go:build phpx

package phpx

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPDOPlaceholders(t *testing.T) {
	s := &PDOStatement{query: "select * from t where a = ? and b = :b and c = '? :c'"}
	s.bind(int64(1), "one")
	s.bind(":b", "two")
	query, args, err := s.prepareArgs()
	if err != nil {
		t.Fatal(err)
	}
	if want := "select * from t where a = ? and b = ? and c = '? :c'"; query != want {
		t.Fatalf("query = %q, want %q", query, want)
	}
	if len(args) != 2 || args[0] != "one" || args[1] != "two" {
		t.Fatalf("args = %v", args)
	}
}

func TestPDOPlaceholderMissing(t *testing.T) {
	s := &PDOStatement{query: "select * from t where a = ? and b = :b"}
	query, args, err := s.prepareArgs()
	if err != nil {
		t.Fatal(err)
	}
	if query != "select * from t where a = ? and b = ?" || len(args) != 2 || args[0] != nil || args[1] != nil {
		t.Fatalf("query = %q, args = %v", query, args)
	}
}

func TestPDOReadStatement(t *testing.T) {
	for _, query := range []string{"select 1", "  WITH x AS (select 1) select * from x", "PRAGMA table_info(t)"} {
		if !isReadStatement(query) {
			t.Errorf("%q reads", query)
		}
	}
	for _, query := range []string{"insert into t values (1)", "update t set a = 1", ""} {
		if isReadStatement(query) {
			t.Errorf("%q writes", query)
		}
	}
}

func TestPDOFetchModes(t *testing.T) {
	s := &PDOStatement{
		pdo:     &PDO{attrs: map[int64]any{}},
		columns: []string{"id", "name"},
		rows:    [][]any{{int64(7), "seven"}},
	}
	assoc, _ := s.rowValue(s.rows[0], fetchAssoc, nil)
	if v, _ := assoc.(*model.Array).Get("name"); v != "seven" {
		t.Fatalf("assoc name = %v", v)
	}
	both, _ := s.rowValue(s.rows[0], fetchBoth, nil)
	if v, _ := both.(*model.Array).Get(int64(1)); v != "seven" {
		t.Fatalf("both position 1 = %v", v)
	}
	object, err := s.rowValue(s.rows[0], fetchObj, nil)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := object.(*model.Object).Prop("id"); v != int64(7) {
		t.Fatalf("object id = %v", v)
	}
}

// mustGet reads an array entry a test expects to be there.
func mustGet(a *model.Array, key any) any {
	v, _ := a.Get(key)
	return v
}

func TestPDOFetchColumn(t *testing.T) {
	s := &PDOStatement{
		pdo:     &PDO{attrs: map[int64]any{}},
		columns: []string{"id", "name"},
		rows:    [][]any{{int64(7), "seven"}, {int64(8), "eight"}},
	}
	all, err := s.FetchAll(int64(fetchColumn), int64(1))
	if err != nil {
		t.Fatal(err)
	}
	list := all.(*model.Array)
	if list.Len() != 2 || mustGet(list, int64(0)) != "seven" {
		t.Fatalf("column = %v", list.Map())
	}
}
