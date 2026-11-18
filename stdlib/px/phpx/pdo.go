//go:build phpx

package phpx

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
	"modernc.org/sqlite"
)

// PHP's PDO is the connection object a database layer is written against: a
// DSN opens it, prepare() and query() hand back a PDOStatement, and the
// attributes name the error and fetch behaviour. The binding opens sqlite,
// mysql and pgsql; any other scheme answers php's "could not find driver".
//
// The rows a statement answers are read eagerly and kept on the statement: a
// cursor held open across a page's lifetime is not what the callers here do,
// and eager rows keep every fetch mode over one shape.
func init() {
	runner.RegisterArea(`Runtime\PX\PDO`, registerPDO)
}

// registerPDO installs the PDO constructor, the constants its class answers and
// the two statics a script calls.
func registerPDO(rt *runner.Runtime) {
	rt.RegisterConstructor("PDO", func(ctx context.Context, dsn string, rest ...any) (*PDO, error) {
		return newPDO(rt, ctx, dsn, rest...)
	})
	// PDO::getAvailableDrivers answers the drivers this binding opens.
	rt.RegisterFunc("PDO::getAvailableDrivers", func() *model.Array {
		out := model.NewArray()
		for _, name := range []string{"mysql", "pgsql", "sqlite"} {
			out.Append(name)
		}
		return out
	})
	// PDO::connect is the static spelling php 8.4 added for `new PDO`.
	rt.RegisterFunc("PDO::connect", func(ctx context.Context, dsn string, rest ...any) (*PDO, error) {
		return newPDO(rt, ctx, dsn, rest...)
	})
	installPDOClass()
}

// PDO is the connection a `new PDO(...)` expression builds.
type PDO struct {
	rt *runner.Runtime
	db *sqlx.DB
	// driver is the name the PDO driver answers to php (sqlite, mysql, pgsql),
	// which ATTR_DRIVER_NAME reads; goDriver is the name database/sql registers.
	driver   string
	goDriver string

	errMode   int64
	fetchMode int64
	attrs     map[int64]any

	tx         *sqlx.Tx
	errInfo    *model.Array
	insertID   int64
	serverInfo string

	// sqlite is the private driver a sqlite connection registers functions on,
	// so a function one PDO defines reaches its own connections alone.
	sqlite *sqlite.Driver
	memory bool
	dsn    string
	used   bool
}

// sqliteConnector hands database/sql the connections of one private driver.
type sqliteConnector struct {
	driver *sqlite.Driver
	dsn    string
}

// Connect opens one connection on the private driver.
func (c sqliteConnector) Connect(context.Context) (driver.Conn, error) {
	return c.driver.Open(c.dsn)
}

// Driver answers the driver the connector opens through.
func (c sqliteConnector) Driver() driver.Driver { return c.driver }

// newPDO opens the connection a PDO DSN names. The username and password are
// read when the DSN itself does not carry them, as php does.
func newPDO(rt *runner.Runtime, ctx context.Context, dsn string, rest ...any) (*PDO, error) {
	user, pass, options := pdoArgs(rest)
	driver, goDriver, goDSN, err := pdoDSN(dsn, user, pass)
	if err != nil {
		return nil, newPDOException("HY000", 0, err.Error())
	}
	var pool *sql.DB
	var private *sqlite.Driver
	if goDriver == "sqlite" {
		private = &sqlite.Driver{}
		pool = sql.OpenDB(sqliteConnector{driver: private, dsn: goDSN})
	} else {
		pool, err = sql.Open(goDriver, goDSN)
	}
	if err != nil {
		return nil, newPDOException("HY000", 0, err.Error())
	}
	db := sqlx.NewDb(pool, goDriver)
	// A sqlite connection opens on its first statement, which leaves the
	// window a script registers a function in before the driver hands out a
	// connection that predates it.
	if goDriver != "sqlite" {
		if err := db.PingContext(ctx); err != nil {
			pool.Close()
			return nil, newPDOException("HY000", 0, err.Error())
		}
	}
	p := &PDO{
		rt:        rt,
		db:        db,
		driver:    driver,
		goDriver:  goDriver,
		errMode:   errmodeException,
		fetchMode: fetchBoth,
		attrs:     map[int64]any{},
		sqlite:    private,
		memory:    isMemoryDSN(goDSN),
		dsn:       goDSN,
	}
	for attr, value := range options {
		p.applyAttribute(attr, value)
	}
	return p, nil
}

// pdoArgs reads the three arguments after the DSN: a username, a password and
// the option array, each optional.
func pdoArgs(rest []any) (user, pass string, options map[int64]any) {
	if len(rest) > 0 && rest[0] != nil {
		user = phpval.String(rest[0])
	}
	if len(rest) > 1 && rest[1] != nil {
		pass = phpval.String(rest[1])
	}
	if len(rest) > 2 {
		options = pdoOptionMap(rest[2])
	}
	return user, pass, options
}

// pdoOptionMap reads a php array keyed by an ATTR_ constant.
func pdoOptionMap(v any) map[int64]any {
	out := map[int64]any{}
	switch options := v.(type) {
	case *model.Array:
		options.Range(func(key, value any) bool {
			switch k := key.(type) {
			case int64:
				out[k] = value
			case string:
				if n, err := strconv.ParseInt(k, 10, 64); err == nil {
					out[n] = value
				}
			}
			return true
		})
	case map[string]any:
		for k, value := range options {
			if n, err := strconv.ParseInt(k, 10, 64); err == nil {
				out[n] = value
			}
		}
	case map[int64]any:
		return options
	}
	return out
}

// pdoDSN turns a PDO DSN into a driver database/sql knows, its Go spelling and
// the php name ATTR_DRIVER_NAME answers. The keyword form of mysql and pgsql
// is written as PDO writes it, `key=value;key=value`.
func pdoDSN(dsn, user, pass string) (driver, goDriver, goDSN string, err error) {
	scheme, rest, ok := strings.Cut(dsn, ":")
	if !ok {
		return "", "", "", fmt.Errorf("invalid data source name")
	}
	switch strings.ToLower(scheme) {
	case "sqlite":
		if rest == "" {
			rest = ":memory:"
		}
		return "sqlite", "sqlite", rest, nil
	case "mysql":
		return "mysql", "mysql", mysqlDSN(rest, user, pass), nil
	case "pgsql", "postgres", "postgresql":
		source, buildErr := pgsqlDSN(rest, user, pass)
		return "pgsql", "pgx", source, buildErr
	case "uri":
		return "", "", "", fmt.Errorf("could not find driver")
	default:
		return "", "", "", fmt.Errorf("could not find driver")
	}
}

// mysqlDSN writes PDO's semicolon form as go-sql-driver's DSN.
func mysqlDSN(dsn, user, pass string) string {
	host, port, dbname, socket := "127.0.0.1", "3306", "", ""
	query := url.Values{}
	for _, pair := range strings.Split(dsn, ";") {
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		switch strings.ToLower(key) {
		case "host":
			host = value
		case "port":
			port = value
		case "dbname":
			dbname = value
		case "unix_socket":
			socket = value
		case "charset":
			query.Set("charset", value)
		}
	}
	query.Set("parseTime", "true")
	query.Set("loc", "Local")
	address := "tcp(" + host + ":" + port + ")"
	if socket != "" {
		address = "unix(" + socket + ")"
	}
	return user + ":" + pass + "@" + address + "/" + dbname + "?" + query.Encode()
}

// pgsqlDSN writes PDO's semicolon form as a pgx connection URL.
func pgsqlDSN(dsn, user, pass string) (string, error) {
	values := url.Values{}
	for _, pair := range strings.Split(dsn, ";") {
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		switch strings.ToLower(key) {
		case "user":
			user = value
		case "password":
			pass = value
		default:
			values.Set(key, value)
		}
	}
	host := values.Get("host")
	if host == "" {
		host = "127.0.0.1"
	}
	port := values.Get("port")
	if port == "" {
		port = "5432"
	}
	values.Del("host")
	values.Del("port")
	name := values.Get("dbname")
	values.Del("dbname")
	u := url.URL{Scheme: "postgres", User: url.UserPassword(user, pass), Host: host + ":" + port, Path: "/" + name, RawQuery: values.Encode()}
	return u.String(), nil
}

// executor is the pool or the transaction a statement runs against.
func (p *PDO) executor() interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
} {
	p.used = true
	if p.tx != nil {
		return p.tx
	}
	return p.db
}

// rebind writes a statement's placeholders the way the driver expects them.
func (p *PDO) rebind(query string) string {
	return sqlx.Rebind(sqlx.BindType(p.goDriver), query)
}

// applyAttribute records an attribute, and ignores the ones that carry no
// difference to this binding: every name a package sets still has to exist.
func (p *PDO) applyAttribute(attr int64, value any) {
	switch attr {
	case attrErrMode:
		p.errMode = phpval.Int(value)
	case attrDefaultFetchMode:
		p.fetchMode = phpval.Int(value)
	default:
		p.attrs[attr] = value
	}
}

// SetAttribute sets a connection attribute and answers true.
func (p *PDO) SetAttribute(attr int64, value any) any {
	p.applyAttribute(attr, value)
	return true
}

// GetAttribute answers the attributes php's PDO answers for a live connection.
func (p *PDO) GetAttribute(attr int64) (any, error) {
	switch attr {
	case attrErrMode:
		return p.errMode, nil
	case attrDefaultFetchMode:
		return p.fetchMode, nil
	case attrDriverName:
		return p.driver, nil
	case attrServerVersion:
		return p.serverVersion(), nil
	case attrClientVersion:
		return p.serverVersion(), nil
	case attrConnectionStatus:
		return "", nil
	}
	if value, ok := p.attrs[attr]; ok {
		return value, nil
	}
	// php reports an attribute the driver does not carry as a warning and
	// answers false, whatever the error mode is.
	p.rt.PhpxWarning(2, "PDO::getAttribute(): SQLSTATE[IM001]: Driver does not support this function: driver does not support that attribute")
	return false, nil
}

// serverVersion reads the server's version string once.
func (p *PDO) serverVersion() string {
	if p.serverInfo != "" {
		return p.serverInfo
	}
	query := "select sqlite_version()"
	switch p.driver {
	case "mysql":
		query = "select version()"
	case "pgsql":
		query = "show server_version"
	}
	var version string
	if err := p.db.Get(&version, p.rebind(query)); err != nil {
		return ""
	}
	p.serverInfo = version
	return version
}

// Exec runs a statement and answers the affected row count.
func (p *PDO) Exec(ctx context.Context, statement string) (any, error) {
	result, err := p.executor().ExecContext(ctx, p.rebind(statement))
	if err != nil {
		return p.failValue(err)
	}
	return p.recordAffected(result), nil
}

// Query runs a statement and answers the PDOStatement its rows are read from.
func (p *PDO) Query(ctx context.Context, query string, fetchMode ...any) (any, error) {
	statement := p.statement(query)
	if len(fetchMode) > 0 {
		statement.fetchMode = phpval.Int(fetchMode[0])
	}
	if err := statement.run(ctx, nil); err != nil {
		return p.failValue(err)
	}
	return statement, nil
}

// Prepare answers a PDOStatement for a query whose bindings the caller sets.
func (p *PDO) Prepare(query string, options ...any) (*PDOStatement, error) {
	return p.statement(query), nil
}

// Quote escapes a string for the driver and wraps it in quotes.
func (p *PDO) Quote(value string, _ ...any) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

// LastInsertId answers the id the last insert made.
func (p *PDO) LastInsertId(name ...any) string {
	return strconv.FormatInt(p.insertID, 10)
}

// BeginTransaction opens a transaction every later statement runs inside.
func (p *PDO) BeginTransaction(ctx context.Context) (any, error) {
	if p.tx != nil {
		return p.failValue(newPDOException("HY000", 0, "There is already an active transaction"))
	}
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return p.failValue(err)
	}
	p.tx = tx
	return true, nil
}

// Commit ends the active transaction.
func (p *PDO) Commit(ctx context.Context) (any, error) {
	if p.tx == nil {
		return p.failValue(newPDOException("HY000", 0, "There is no active transaction"))
	}
	tx := p.tx
	p.tx = nil
	if err := tx.Commit(); err != nil {
		return p.failValue(err)
	}
	return true, nil
}

// RollBack discards the active transaction.
func (p *PDO) RollBack(ctx context.Context) (any, error) {
	if p.tx == nil {
		return p.failValue(newPDOException("HY000", 0, "There is no active transaction"))
	}
	tx := p.tx
	p.tx = nil
	if err := tx.Rollback(); err != nil {
		return p.failValue(err)
	}
	return true, nil
}

// InTransaction reports whether a transaction is open.
func (p *PDO) InTransaction() bool {
	return p.tx != nil
}

// ErrorCode answers the SQLSTATE of the last error, "00000" after a success.
func (p *PDO) ErrorCode() string {
	if p.errInfo == nil {
		return "00000"
	}
	if value, _ := p.errInfo.Get(int64(0)); value != nil {
		return phpval.String(value)
	}
	return "00000"
}

// ErrorInfo answers the SQLSTATE, driver code and driver message of the last error.
func (p *PDO) ErrorInfo() *model.Array {
	if p.errInfo == nil {
		out := model.NewArray()
		out.Append("00000")
		out.Append(nil)
		out.Append(nil)
		return out
	}
	return p.errInfo
}

// recordAffected reads an exec result into the row count and the insert id php
// keeps beside the connection.
func (p *PDO) recordAffected(result sql.Result) int64 {
	affected, _ := result.RowsAffected()
	if id, err := result.LastInsertId(); err == nil && id > 0 {
		p.insertID = id
	}
	return affected
}

// failValue records err and answers php's false unless the error mode throws.
func (p *PDO) failValue(err error) (any, error) {
	if exc := p.fail(err); exc != nil {
		return nil, exc
	}
	return false, nil
}

// fail records err and answers what the error mode makes of it: nil when the
// mode reports it without throwing, the PDOException when it throws.
func (p *PDO) fail(err error) error {
	exc := newPDOExceptionFromError(p.driver, err)
	p.errInfo = exc.ErrorInfo
	switch p.errMode {
	case errmodeException:
		return exc
	case errmodeWarning:
		p.rt.PhpxWarning(2, exc.Error())
		return nil
	}
	return nil
}

// statement builds a statement bound to this connection.
func (p *PDO) statement(query string) *PDOStatement {
	return &PDOStatement{pdo: p, query: query, fetchMode: p.fetchMode, binds: map[string]any{}}
}

// SqliteCreateFunction registers a php callback as a scalar sql function on
// this connection's sqlite driver. The driver applies a function to the
// connections it opens after the registration, so the idle connection the
// constructor's ping made is dropped for a file database. An in-memory
// database lives in that connection and keeps no function added this way.
func (p *PDO) SqliteCreateFunction(name string, callable any, nargs ...any) (any, error) {
	if p.sqlite == nil {
		return nil, newPDOException("IM001", 0, "SQLSTATE[IM001]: Driver does not support this function: This driver does not support sqliteCreateFunction")
	}
	if p.tx != nil {
		return nil, newPDOException("HY000", 0, "PDO::sqliteCreateFunction(): cannot register a function inside a transaction")
	}
	fn, ok := p.rt.Callable(callable)
	if !ok {
		return nil, fmt.Errorf("PDO::sqliteCreateFunction(): Argument #2 ($callback) must be a valid callback")
	}
	count := int32(-1)
	if len(nargs) > 0 && nargs[0] != nil {
		count = int32(phpval.Int(nargs[0]))
	}
	err := p.sqlite.RegisterScalarFunction(name, count, func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		values := make([]any, len(args))
		for i, arg := range args {
			values[i] = pdoValue(arg)
		}
		out, err := fn(values...)
		if err != nil {
			return nil, err
		}
		return pdoFunctionValue(out), nil
	})
	if err != nil {
		return nil, newPDOException("HY000", 0, err.Error())
	}
	if !p.memory && p.used {
		p.reopen()
	}
	return true, nil
}

// reopen drops a sqlite pool so the next statement opens a connection the
// function just registered reaches. An in-memory database lives in the pool
// and is not reopened.
func (p *PDO) reopen() {
	p.db.Close()
	p.db = sqlx.NewDb(sql.OpenDB(sqliteConnector{driver: p.sqlite, dsn: p.dsn}), "sqlite")
}

// pdoFunctionValue narrows a php callback's answer to the value sqlite stores.
func pdoFunctionValue(value any) any {
	switch v := value.(type) {
	case bool:
		if v {
			return int64(1)
		}
		return int64(0)
	case int:
		return int64(v)
	case nil, int64, float64, string, []byte:
		return v
	}
	return phpval.String(value)
}

// isMemoryDSN reports whether a sqlite DSN keeps its database in memory.
func isMemoryDSN(dsn string) bool {
	name, query, _ := strings.Cut(dsn, "?")
	if strings.EqualFold(name, ":memory:") || strings.EqualFold(name, "file::memory:") {
		return true
	}
	for _, pair := range strings.Split(query, "&") {
		key, value, ok := strings.Cut(pair, "=")
		if ok && strings.EqualFold(key, "mode") && strings.EqualFold(value, "memory") {
			return true
		}
	}
	return false
}
