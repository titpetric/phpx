//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/runner"
)

// pdoConsts are PDO's class constants with the values php 8.4 gives them: the
// integer ones from a reflection dump, plus the one string, ERR_NONE. The
// MYSQL_ and PGSQL_ and SQLITE_ names are the legacy spellings a package still
// writes for Pdo\Mysql and Pdo\Pgsql and Pdo\Sqlite.
var pdoConsts = map[string]any{
	"ATTR_AUTOCOMMIT":                   int64(0),
	"ATTR_CASE":                         int64(8),
	"ATTR_CLIENT_VERSION":               int64(5),
	"ATTR_CONNECTION_STATUS":            int64(7),
	"ATTR_CURSOR":                       int64(10),
	"ATTR_CURSOR_NAME":                  int64(9),
	"ATTR_DEFAULT_FETCH_MODE":           int64(19),
	"ATTR_DEFAULT_STR_PARAM":            int64(21),
	"ATTR_DRIVER_NAME":                  int64(16),
	"ATTR_EMULATE_PREPARES":             int64(20),
	"ATTR_ERRMODE":                      int64(3),
	"ATTR_FETCH_CATALOG_NAMES":          int64(15),
	"ATTR_FETCH_TABLE_NAMES":            int64(14),
	"ATTR_MAX_COLUMN_LEN":               int64(18),
	"ATTR_ORACLE_NULLS":                 int64(11),
	"ATTR_PERSISTENT":                   int64(12),
	"ATTR_PREFETCH":                     int64(1),
	"ATTR_SERVER_INFO":                  int64(6),
	"ATTR_SERVER_VERSION":               int64(4),
	"ATTR_STATEMENT_CLASS":              int64(13),
	"ATTR_STRINGIFY_FETCHES":            int64(17),
	"ATTR_TIMEOUT":                      int64(2),
	"CASE_LOWER":                        int64(2),
	"CASE_NATURAL":                      int64(0),
	"CASE_UPPER":                        int64(1),
	"CURSOR_FWDONLY":                    int64(0),
	"CURSOR_SCROLL":                     int64(1),
	"ERRMODE_EXCEPTION":                 int64(2),
	"ERRMODE_SILENT":                    int64(0),
	"ERRMODE_WARNING":                   int64(1),
	"ERR_NONE":                          "00000",
	"FETCH_ASSOC":                       int64(2),
	"FETCH_BOTH":                        int64(4),
	"FETCH_BOUND":                       int64(6),
	"FETCH_CLASS":                       int64(8),
	"FETCH_CLASSTYPE":                   int64(262144),
	"FETCH_COLUMN":                      int64(7),
	"FETCH_DEFAULT":                     int64(0),
	"FETCH_FUNC":                        int64(10),
	"FETCH_GROUP":                       int64(65536),
	"FETCH_INTO":                        int64(9),
	"FETCH_KEY_PAIR":                    int64(12),
	"FETCH_LAZY":                        int64(1),
	"FETCH_NAMED":                       int64(11),
	"FETCH_NUM":                         int64(3),
	"FETCH_OBJ":                         int64(5),
	"FETCH_ORI_ABS":                     int64(4),
	"FETCH_ORI_FIRST":                   int64(2),
	"FETCH_ORI_LAST":                    int64(3),
	"FETCH_ORI_NEXT":                    int64(0),
	"FETCH_ORI_PRIOR":                   int64(1),
	"FETCH_ORI_REL":                     int64(5),
	"FETCH_PROPS_LATE":                  int64(1048576),
	"FETCH_SERIALIZE":                   int64(524288),
	"FETCH_UNIQUE":                      int64(196608),
	"MYSQL_ATTR_COMPRESS":               int64(1003),
	"MYSQL_ATTR_DIRECT_QUERY":           int64(1004),
	"MYSQL_ATTR_FOUND_ROWS":             int64(1005),
	"MYSQL_ATTR_IGNORE_SPACE":           int64(1006),
	"MYSQL_ATTR_INIT_COMMAND":           int64(1002),
	"MYSQL_ATTR_LOCAL_INFILE":           int64(1001),
	"MYSQL_ATTR_LOCAL_INFILE_DIRECTORY": int64(1015),
	"MYSQL_ATTR_MULTI_STATEMENTS":       int64(1013),
	"MYSQL_ATTR_SERVER_PUBLIC_KEY":      int64(1012),
	"MYSQL_ATTR_SSL_CA":                 int64(1009),
	"MYSQL_ATTR_SSL_CAPATH":             int64(1010),
	"MYSQL_ATTR_SSL_CERT":               int64(1008),
	"MYSQL_ATTR_SSL_CIPHER":             int64(1011),
	"MYSQL_ATTR_SSL_KEY":                int64(1007),
	"MYSQL_ATTR_SSL_VERIFY_SERVER_CERT": int64(1014),
	"MYSQL_ATTR_USE_BUFFERED_QUERY":     int64(1000),
	"NULL_EMPTY_STRING":                 int64(1),
	"NULL_NATURAL":                      int64(0),
	"NULL_TO_STRING":                    int64(2),
	"PARAM_BOOL":                        int64(5),
	"PARAM_EVT_ALLOC":                   int64(0),
	"PARAM_EVT_EXEC_POST":               int64(3),
	"PARAM_EVT_EXEC_PRE":                int64(2),
	"PARAM_EVT_FETCH_POST":              int64(5),
	"PARAM_EVT_FETCH_PRE":               int64(4),
	"PARAM_EVT_FREE":                    int64(1),
	"PARAM_EVT_NORMALIZE":               int64(6),
	"PARAM_INPUT_OUTPUT":                int64(2147483648),
	"PARAM_INT":                         int64(1),
	"PARAM_LOB":                         int64(3),
	"PARAM_NULL":                        int64(0),
	"PARAM_STMT":                        int64(4),
	"PARAM_STR":                         int64(2),
	"PARAM_STR_CHAR":                    int64(536870912),
	"PARAM_STR_NATL":                    int64(1073741824),
	"PGSQL_ATTR_DISABLE_PREPARES":       int64(1000),
	"PGSQL_TRANSACTION_ACTIVE":          int64(1),
	"PGSQL_TRANSACTION_IDLE":            int64(0),
	"PGSQL_TRANSACTION_INERROR":         int64(3),
	"PGSQL_TRANSACTION_INTRANS":         int64(2),
	"PGSQL_TRANSACTION_UNKNOWN":         int64(4),
	"SQLITE_ATTR_EXTENDED_RESULT_CODES": int64(1002),
	"SQLITE_ATTR_OPEN_FLAGS":            int64(1000),
	"SQLITE_ATTR_READONLY_STATEMENT":    int64(1001),
	"SQLITE_DETERMINISTIC":              int64(2048),
	"SQLITE_OPEN_CREATE":                int64(4),
	"SQLITE_OPEN_READONLY":              int64(1),
	"SQLITE_OPEN_READWRITE":             int64(2),
}

// installPDOClass makes PDO's constants answer without a class in the table:
// `new PDO` is the host constructor, and the table is cleared between
// programs, so a registered class or class source would leave the constants
// unresolved. The host constant registry is the runtime's own and survives.
func installPDOClass() {
	for name, value := range pdoConsts {
		runner.PhpxRegisterHostConst("PDO", name, value)
	}
}

// The values of the constants the binding reads back by name. They are the
// numbers php 8.4 reports for each name, which is what the map above stores.
const (
	attrCase             = int64(8)
	attrDefaultFetchMode = int64(19)
	attrDriverName       = int64(16)
	attrErrMode          = int64(3)
	attrOracleNulls      = int64(11)
	attrServerVersion    = int64(4)
	attrClientVersion    = int64(5)
	attrConnectionStatus = int64(7)
	attrStringifyFetches = int64(17)

	errmodeSilent    = int64(0)
	errmodeWarning   = int64(1)
	errmodeException = int64(2)

	fetchDefault = int64(0)
	fetchLazy    = int64(1)
	fetchAssoc   = int64(2)
	fetchNum     = int64(3)
	fetchBoth    = int64(4)
	fetchObj     = int64(5)
	fetchBound   = int64(6)
	fetchColumn  = int64(7)
	fetchClass   = int64(8)
	fetchInto    = int64(9)
	fetchKeyPair = int64(12)
	fetchGroup   = int64(65536)
	fetchUnique  = int64(196608)

	paramNull = int64(0)
	paramInt  = int64(1)
	paramStr  = int64(2)
	paramLob  = int64(3)
	paramBool = int64(5)

	caseNatural = int64(0)
	caseUpper   = int64(1)
	caseLower   = int64(2)

	nullNatural     = int64(0)
	nullEmptyString = int64(1)
	nullToString    = int64(2)
)
