//go:build phpx

package phpx

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
)

// PDOStatement is what prepare() and query() answer: a query, the bindings
// set on it and the rows the last execute read. The rows are materialised at
// execute, so every fetch mode reads one shape and a cursor lives no longer
// than the statement.
type PDOStatement struct {
	pdo   *PDO
	query string

	binds      map[string]any
	positional map[int64]any
	fetchMode  int64

	executed bool
	pos      int
	columns  []string
	rows     [][]any
	affected int64

	errInfo *model.Array
}

// BindValue binds a value to a named or 1-based positional placeholder.
func (s *PDOStatement) BindValue(param any, value any, typ ...any) bool {
	kind := int64(paramStr)
	if len(typ) > 0 {
		kind = phpval.Int(typ[0])
	}
	s.bind(param, pdoBindValue(value, kind))
	return true
}

// BindParam binds a value the same way bindValue does. php binds by reference
// and reads the variable at execute; the eager statement keeps the value it
// was given, which is the same for a caller that binds once and executes.
func (s *PDOStatement) BindParam(param any, value any, typ ...any) bool {
	return s.BindValue(param, value, typ...)
}

// bind stores one binding under its placeholder spelling.
func (s *PDOStatement) bind(param any, value any) {
	switch p := param.(type) {
	case int64:
		if s.positional == nil {
			s.positional = map[int64]any{}
		}
		s.positional[p] = value
	case int:
		s.bind(int64(p), value)
	case string:
		if s.binds == nil {
			s.binds = map[string]any{}
		}
		s.binds[strings.TrimPrefix(p, ":")] = value
	}
}

// Execute runs the statement with its bindings and any given here.
func (s *PDOStatement) Execute(ctx context.Context, params ...any) (any, error) {
	if len(params) > 0 && params[0] != nil {
		s.merge(params[0])
	}
	if err := s.run(ctx, nil); err != nil {
		if exc := s.pdo.fail(err); exc != nil {
			return nil, exc
		}
		return false, nil
	}
	return true, nil
}

// merge reads the bindings execute() was given: positional or named.
func (s *PDOStatement) merge(params any) {
	switch values := params.(type) {
	case *model.Array:
		values.Range(func(key, value any) bool {
			switch k := key.(type) {
			case int64:
				s.bind(k+1, pdoBindValue(value, paramStr))
			case string:
				s.bind(k, pdoBindValue(value, paramStr))
			}
			return true
		})
	case []any:
		for i, value := range values {
			s.bind(int64(i+1), pdoBindValue(value, paramStr))
		}
	}
}

// run executes the query and keeps its rows or its affected count.
func (s *PDOStatement) run(ctx context.Context, _ any) error {
	query, args, err := s.prepareArgs()
	if err != nil {
		return err
	}
	query = s.pdo.rebind(query)
	s.executed, s.pos, s.errInfo = true, 0, nil
	s.columns, s.rows = nil, nil
	executor := s.pdo.executor()
	if isReadStatement(query) {
		rows, err := executor.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		return s.scan(rows)
	}
	result, err := executor.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	s.affected = s.pdo.recordAffected(result)
	return nil
}

// scan reads every row and the column names beside them.
func (s *PDOStatement) scan(rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	s.columns = columns
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return err
		}
		for i, value := range values {
			values[i] = pdoValue(value)
		}
		s.rows = append(s.rows, values)
	}
	return rows.Err()
}

// prepareArgs writes the placeholders in the driver's form and lines the
// bound values up with them.
func (s *PDOStatement) prepareArgs() (string, []any, error) {
	var out strings.Builder
	args := make([]any, 0, len(s.positional)+len(s.binds))
	next := int64(1)
	for i := 0; i < len(s.query); i++ {
		c := s.query[i]
		switch c {
		case '\'', '"', '`':
			end := quotedEnd(s.query, i)
			out.WriteString(s.query[i:end])
			i = end - 1
		case ':':
			name, end := placeholderName(s.query, i+1)
			if name == "" {
				out.WriteByte(c)
				continue
			}
			// pdo_sqlite binds a parameter no value was given as NULL, and
			// executes; php's other drivers refuse it.
			value := s.binds[name]
			args = append(args, value)
			out.WriteByte('?')
			i = end - 1
		case '?':
			value := s.positional[next]
			args = append(args, value)
			next++
			out.WriteByte('?')
		default:
			out.WriteByte(c)
		}
	}
	return out.String(), args, nil
}

// quotedEnd finds the index just past a quoted run that starts at i.
func quotedEnd(query string, i int) int {
	quote := query[i]
	for j := i + 1; j < len(query); j++ {
		if query[j] != quote {
			continue
		}
		if j+1 < len(query) && query[j+1] == quote {
			j++
			continue
		}
		return j + 1
	}
	return len(query)
}

// placeholderName reads the identifier after a colon, nil at its end.
func placeholderName(query string, i int) (string, int) {
	start := i
	for i < len(query) && (query[i] == '_' || query[i] >= 'a' && query[i] <= 'z' || query[i] >= 'A' && query[i] <= 'Z' || i > start && query[i] >= '0' && query[i] <= '9') {
		i++
	}
	return query[start:i], i
}

// isReadStatement reports whether a statement answers rows: every other kind
// runs through Exec and answers a row count.
func isReadStatement(query string) bool {
	fields := strings.Fields(strings.ToUpper(query))
	if len(fields) == 0 {
		return false
	}
	switch fields[0] {
	case "SELECT", "WITH", "PRAGMA", "EXPLAIN", "VALUES", "SHOW", "DESCRIBE", "DESC", "TABLE":
		return true
	}
	return false
}

// Fetch answers the next row in the mode asked for, false at the end.
func (s *PDOStatement) Fetch(mode ...any) (any, error) {
	row, err := s.next()
	if err != nil {
		return nil, err
	}
	if row == nil {
		return false, nil
	}
	return s.rowValue(row, s.mode(mode), rest(mode))
}

// FetchAll answers every remaining row in the mode asked for.
func (s *PDOStatement) FetchAll(mode ...any) (any, error) {
	value := s.mode(mode)
	switch value {
	case fetchColumn:
		column := int64(0)
		if len(mode) > 1 {
			column = phpval.Int(mode[1])
		}
		out := model.NewArray()
		for s.pos < len(s.rows) {
			out.Append(s.column(s.rows[s.pos], column))
			s.pos++
		}
		return out, nil
	case fetchKeyPair:
		out := model.NewArray()
		for s.pos < len(s.rows) {
			row := s.rows[s.pos]
			if len(row) > 1 {
				out.Set(row[0], row[1])
			}
			s.pos++
		}
		return out, nil
	}
	out := model.NewArray()
	for s.pos < len(s.rows) {
		row, err := s.rowValue(s.rows[s.pos], value, rest(mode))
		if err != nil {
			return nil, err
		}
		out.Append(row)
		s.pos++
	}
	return out, nil
}

// FetchColumn answers one column of the next row.
func (s *PDOStatement) FetchColumn(column ...any) (any, error) {
	index := int64(0)
	if len(column) > 0 {
		index = phpval.Int(column[0])
	}
	if s.pos >= len(s.rows) {
		return false, nil
	}
	value := s.column(s.rows[s.pos], index)
	s.pos++
	return value, nil
}

// FetchObject answers the next row as an object of the named class, stdClass
// when the name is omitted.
func (s *PDOStatement) FetchObject(class ...any) (any, error) {
	row, err := s.next()
	if err != nil {
		return nil, err
	}
	if row == nil {
		return false, nil
	}
	name := "stdClass"
	if len(class) > 0 && class[0] != nil {
		name = phpval.String(class[0])
	}
	return s.object(row, name)
}

// RowCount answers the affected row count of the last execute.
func (s *PDOStatement) RowCount() int64 { return s.affected }

// ColumnCount answers the number of columns of the result.
func (s *PDOStatement) ColumnCount() int64 { return int64(len(s.columns)) }

// CloseCursor drops the rows and lets a later execute refill them.
func (s *PDOStatement) CloseCursor() bool {
	s.rows, s.pos = nil, 0
	return true
}

// SetFetchMode sets the mode a later fetch without an argument uses.
func (s *PDOStatement) SetFetchMode(mode int64, _ ...any) bool {
	s.fetchMode = mode
	return true
}

// GetColumnMeta answers the column's name and its php type.
func (s *PDOStatement) GetColumnMeta(index int64) (any, error) {
	if index < 0 || int(index) >= len(s.columns) {
		return false, nil
	}
	out := model.NewArray()
	out.Set("name", s.columns[index])
	out.Set("native_type", "string")
	out.Set("pdo_type", int64(paramStr))
	out.Set("flags", model.NewArray())
	out.Set("len", int64(-1))
	out.Set("precision", int64(0))
	return out, nil
}

// ErrorCode answers the SQLSTATE of the statement's last error.
func (s *PDOStatement) ErrorCode() string {
	if s.errInfo == nil {
		return "00000"
	}
	if value, _ := s.errInfo.Get(int64(0)); value != nil {
		return phpval.String(value)
	}
	return "00000"
}

// ErrorInfo answers the statement's last error triple.
func (s *PDOStatement) ErrorInfo() *model.Array {
	if s.errInfo == nil {
		return s.pdo.ErrorInfo()
	}
	return s.errInfo
}

// NextRowset answers false: no statement here answers a second result set.
func (s *PDOStatement) NextRowset() bool { return false }

// DebugDumpParams is php's debug hook; the statement has nothing to dump.
func (s *PDOStatement) DebugDumpParams() {}

// next advances past one row, nil at the end of the result.
func (s *PDOStatement) next() ([]any, error) {
	if !s.executed {
		return nil, newPDOException("HY000", 0, "PDOStatement::fetch(): The statement must be executed before fetching")
	}
	if s.pos >= len(s.rows) {
		return nil, nil
	}
	row := s.rows[s.pos]
	s.pos++
	return row, nil
}

// mode resolves fetch's mode argument against the statement's default.
func (s *PDOStatement) mode(mode []any) int64 {
	if len(mode) == 0 || mode[0] == nil {
		mode = []any{s.fetchMode}
	}
	value := phpval.Int(mode[0])
	if value == fetchDefault {
		value = s.fetchMode
	}
	return value &^ (fetchGroup | fetchUnique)
}

// rowValue renders one row in mode. The args after the mode name the class
// FETCH_CLASS builds, and name the column FETCH_COLUMN reads.
func (s *PDOStatement) rowValue(row []any, mode int64, args []any) (any, error) {
	switch mode {
	case fetchNum:
		return s.list(row), nil
	case fetchObj, fetchLazy, fetchBound, fetchClass, fetchInto:
		name := "stdClass"
		if len(args) > 0 && args[0] != nil {
			name = phpval.String(args[0])
		}
		return s.object(row, name)
	case fetchColumn:
		index := int64(0)
		if len(args) > 0 {
			index = phpval.Int(args[0])
		}
		return s.column(row, index), nil
	}
	return s.assoc(row, mode == fetchBoth), nil
}

// rest is the arguments after the fetch mode, empty when there are none.
func rest(mode []any) []any {
	if len(mode) < 2 {
		return nil
	}
	return mode[1:]
}

// list renders a row as a list.
func (s *PDOStatement) list(row []any) *model.Array {
	out := model.NewArraySize(len(row))
	for _, value := range row {
		out.Append(s.convert(value))
	}
	return out
}

// assoc renders a row keyed by column name, with the positions beside it for
// FETCH_BOTH.
func (s *PDOStatement) assoc(row []any, both bool) *model.Array {
	out := model.NewArraySize(len(row) * 2)
	for i, value := range row {
		name := s.columnName(i)
		out.Set(name, s.convert(value))
		if both {
			out.Set(int64(i), s.convert(value))
		}
	}
	return out
}

// object renders a row as an instance of name.
func (s *PDOStatement) object(row []any, name string) (any, error) {
	obj := model.NewStdClass()
	if !strings.EqualFold(name, "stdClass") {
		value, err := s.pdo.rt.PhpxNewWithoutConstructor(name)
		if err != nil {
			return nil, err
		}
		target, ok := value.(*model.Object)
		if !ok {
			return nil, fmt.Errorf("PDO: class %s does not answer properties", name)
		}
		obj = target
	}
	for i, value := range row {
		obj.SetProp(s.columnName(i), s.convert(value))
	}
	return obj, nil
}

// column reads one value of a row, with php's false for a missing column.
func (s *PDOStatement) column(row []any, index int64) any {
	if index < 0 || int(index) >= len(row) {
		return false
	}
	return s.convert(row[index])
}

// columnName applies ATTR_CASE to a column name.
func (s *PDOStatement) columnName(i int) string {
	if i >= len(s.columns) {
		return strconv.Itoa(i)
	}
	name := s.columns[i]
	if mode, ok := s.pdo.attrs[attrCase]; ok {
		switch phpval.Int(mode) {
		case caseUpper:
			return strings.ToUpper(name)
		case caseLower:
			return strings.ToLower(name)
		}
	}
	return name
}

// convert applies ATTR_ORACLE_NULLS and ATTR_STRINGIFY_FETCHES to a value.
func (s *PDOStatement) convert(value any) any {
	if nulls, ok := s.pdo.attrs[attrOracleNulls]; ok {
		switch phpval.Int(nulls) {
		case nullToString:
			if value == nil {
				return ""
			}
		case nullEmptyString:
			if text, ok := value.(string); ok && text == "" {
				return nil
			}
		}
	}
	if flag, ok := s.pdo.attrs[attrStringifyFetches]; ok && phpval.Truthy(flag) {
		switch value.(type) {
		case int64, float64:
			return phpval.String(value)
		}
	}
	return value
}

// pdoValue normalises a scanned driver value to php's storage type.
func pdoValue(value any) any {
	switch v := value.(type) {
	case []byte:
		return string(v)
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	case int32:
		return int64(v)
	}
	return value
}

// pdoBindValue coerces a bound value to the PDO::PARAM_ type it was bound
// with, as php's driver does before it hands it to the server.
func pdoBindValue(value any, typ int64) any {
	switch typ {
	case paramNull:
		return nil
	case paramBool:
		if b, ok := value.(bool); ok {
			if b {
				return int64(1)
			}
			return int64(0)
		}
		return phpval.Int(value)
	case paramInt:
		return phpval.Int(value)
	case paramLob:
		if value == nil {
			return nil
		}
		return phpval.String(value)
	}
	if value == nil {
		return nil
	}
	return phpval.String(value)
}
