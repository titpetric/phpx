//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Constants`, registerPhpxConstants)
}

// phpxConstants are the predefined constants php code reads and upstream leaves
// out. The values are php's.
var phpxConstants = map[string]int64{
	"E_ERROR": 1, "E_WARNING": 2, "E_PARSE": 4, "E_NOTICE": 8, "E_CORE_ERROR": 16, "E_CORE_WARNING": 32,
	"E_COMPILE_ERROR": 64, "E_COMPILE_WARNING": 128, "E_USER_ERROR": 256, "E_USER_WARNING": 512,
	"E_USER_NOTICE": 1024, "E_STRICT": 2048, "E_RECOVERABLE_ERROR": 4096, "E_DEPRECATED": 8192,
	"E_USER_DEPRECATED": 16384, "E_ALL": 32767,
	"COUNT_NORMAL": 0, "COUNT_RECURSIVE": 1,
	"DEBUG_BACKTRACE_PROVIDE_OBJECT": 1, "DEBUG_BACKTRACE_IGNORE_ARGS": 2,
	"ENT_DISALLOWED": 128, "ENT_XML1": 16,
	"FILE_USE_INCLUDE_PATH": 1, "FILE_IGNORE_NEW_LINES": 2, "FILE_SKIP_EMPTY_LINES": 4,
	"FILTER_VALIDATE_URL": 273, "FILTER_VALIDATE_EMAIL": 274,
	"JSON_ERROR_NONE": 0, "JSON_ERROR_DEPTH": 1, "JSON_ERROR_STATE_MISMATCH": 2, "JSON_ERROR_CTRL_CHAR": 3,
	"JSON_ERROR_SYNTAX": 4, "JSON_ERROR_UTF8": 5,
	"JSON_PRETTY_PRINT": 128, "JSON_UNESCAPED_SLASHES": 64, "JSON_UNESCAPED_UNICODE": 256,
	"JSON_THROW_ON_ERROR": 4194304, "JSON_PRESERVE_ZERO_FRACTION": 1024, "JSON_NUMERIC_CHECK": 32,
	"JSON_HEX_TAG": 1, "JSON_HEX_AMP": 2, "JSON_HEX_APOS": 4, "JSON_HEX_QUOT": 8, "JSON_FORCE_OBJECT": 16,
	"JSON_PARTIAL_OUTPUT_ON_ERROR": 512, "JSON_INVALID_UTF8_SUBSTITUTE": 2097152, "JSON_INVALID_UTF8_IGNORE": 1048576,
	"JSON_OBJECT_AS_ARRAY": 1, "JSON_BIGINT_AS_STRING": 2,
	"LC_CTYPE": 0, "LC_NUMERIC": 1, "LC_TIME": 2, "LC_COLLATE": 3, "LC_MONETARY": 4, "LC_MESSAGES": 5, "LC_ALL": 6,
	"PHP_URL_SCHEME": 0, "PHP_URL_HOST": 1, "PHP_URL_PORT": 2, "PHP_URL_USER": 3, "PHP_URL_PASS": 4,
	"PHP_URL_PATH": 5, "PHP_URL_QUERY": 6, "PHP_URL_FRAGMENT": 7,
	"PATHINFO_DIRNAME": 1, "PATHINFO_BASENAME": 2, "PATHINFO_EXTENSION": 4, "PATHINFO_FILENAME": 8,
	"PHP_QUERY_RFC1738": 1, "PHP_QUERY_RFC3986": 2, "PHP_MAXPATHLEN": 4096,
	"SORT_REGULAR": 0, "SORT_NUMERIC": 1, "SORT_STRING": 2, "SORT_DESC": 3, "SORT_ASC": 4,
	"SORT_LOCALE_STRING": 5, "SORT_NATURAL": 6, "SORT_FLAG_CASE": 8,
}

// phpxStringConstants are the DATE_* formats.
var phpxStringConstants = map[string]string{
	"DATE_ATOM": "Y-m-d\\TH:i:sP", "DATE_COOKIE": "l, d-M-Y H:i:s T", "DATE_ISO8601": "Y-m-d\\TH:i:sO",
	"DATE_RFC822": "D, d M y H:i:s O", "DATE_RFC850": "l, d-M-y H:i:s T", "DATE_RFC1036": "D, d M y H:i:s O",
	"DATE_RFC1123": "D, d M Y H:i:s O", "DATE_RFC7231": "D, d M Y H:i:s \\G\\M\\T", "DATE_RFC2822": "D, d M Y H:i:s O",
	"DATE_RFC3339": "Y-m-d\\TH:i:sP", "DATE_RFC3339_EXTENDED": "Y-m-d\\TH:i:s.vP", "DATE_RSS": "D, d M Y H:i:s O", "DATE_W3C": "Y-m-d\\TH:i:sP",
}

// registerPhpxConstants defines each of them that the runtime does not.
func registerPhpxConstants(rt *runner.Runtime) {
	for name, value := range phpxConstants {
		if _, exists := rt.Const(name); !exists {
			rt.SetConst(name, value)
		}
	}
	for name, value := range phpxStringConstants {
		if _, exists := rt.Const(name); !exists {
			rt.SetConst(name, value)
		}
	}
}
