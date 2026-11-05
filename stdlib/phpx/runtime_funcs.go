//go:build phpx

package phpx

import (
	"fmt"
	"github.com/titpetric/phpx/internal/phpval"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxRuntimeFuncs)
}

// registerPhpxRuntimeFuncs installs the engine and environment functions a
// vendor tree calls and a script rarely does: assertions, the garbage
// collector, handlers, the loaded extension list and printf.
func registerPhpxRuntimeFuncs(rt *runner.Runtime) {
	// assert() is php.ini's zend.assertions=-1 in production: the expression is
	// not evaluated for effect and the call answers true.
	rt.RegisterFunc("assert", func(_ ...any) bool { return true })
	rt.RegisterFunc("printf", func(format string, args ...any) (int64, error) {
		out, err := rt.InvokeNamed("sprintf", append([]any{format}, args...)...)
		if err != nil {
			return 0, err
		}
		s := fmt.Sprint(out)
		_, err = rt.Output().Write([]byte(s))
		return int64(len(s)), err
	})
	// is_callable($value, $syntaxOnly, &$callableName): upstream takes the value
	// alone. The extra arguments are accepted and ignored.
	if prev, ok := rt.LookupFunc("is_callable"); ok {
		if check, ok := prev.(func(any) bool); ok {
			rt.RegisterFunc("is_callable", func(v any, _ ...any) bool { return check(v) })
		}
	}
	// libxml's error buffer. The extension stays out of scope and
	// extension_loaded('libxml') stays false; phpunit's TestCase::runBare calls
	// libxml_clear_errors() with no guard, and with nothing parsed the buffer
	// these three read is empty.
	rt.RegisterFunc("libxml_clear_errors", func() {})
	rt.RegisterFunc("libxml_use_internal_errors", func(_ ...any) bool { return false })
	rt.RegisterFunc("libxml_get_errors", func() *model.Array { return model.NewArray() })
	rt.RegisterFunc("stream_isatty", func(_ any) bool { return false })
	rt.RegisterFunc("sapi_windows_vt100_support", func(_ ...any) bool { return false })
	rt.RegisterFunc("setlocale", func(_ int64, _ ...any) any { return "C" })
	rt.RegisterFunc("get_declared_traits", func() *model.Array { return model.NewArray() })
	rt.RegisterFunc("get_declared_interfaces", func() *model.Array { return model.NewArray() })
	rt.RegisterFunc("get_resource_type", func(_ any) string { return "stream" })
	rt.RegisterFunc("addcslashes", func(s, chars string) string {
		var b strings.Builder
		for i := 0; i < len(s); i++ {
			if strings.IndexByte(chars, s[i]) >= 0 {
				b.WriteByte('\\')
			}
			b.WriteByte(s[i])
		}
		return b.String()
	})
	rt.RegisterFunc("array_diff_assoc", func(a any, others ...any) *model.Array {
		out := model.NewArray()
		src, ok := a.(*model.Array)
		if !ok {
			return out
		}
		src.Range(func(k, v any) bool {
			for _, o := range others {
				if oa, ok := o.(*model.Array); ok {
					if ov, has := oa.Get(k); has && fmt.Sprint(ov) == fmt.Sprint(v) {
						return true
					}
				}
			}
			out.Set(k, v)
			return true
		})
		return out
	})
	rt.RegisterFunc("gc_enable", func() {})
	rt.RegisterFunc("gc_disable", func() {})
	rt.RegisterFunc("gc_enabled", func() bool { return true })
	rt.RegisterFunc("gc_status", func() *model.Array {
		out := model.NewArray()
		for _, k := range []string{"runs", "collected", "threshold", "roots"} {
			out.Set(k, int64(0))
		}
		return out
	})
	rt.RegisterFunc("memory_reset_peak_usage", func() {})
	rt.RegisterFunc("php_ini_loaded_file", func() any { return false })
	rt.RegisterFunc("php_ini_scanned_files", func() any { return false })
	rt.RegisterFunc("ini_get_all", func(_ ...any) *model.Array { return model.NewArray() })
	rt.RegisterFunc("gc_collect_cycles", func() int64 { return 0 })
	rt.RegisterFunc("clearstatcache", func(_ ...any) {})
	rt.RegisterFunc("error_clear_last", func() {})
	rt.RegisterFunc("debug_backtrace", func(args ...any) *model.Array {
		flags, limit := int64(1), 0
		if len(args) > 0 {
			if n, ok := args[0].(int64); ok {
				flags = n
			}
		}
		if len(args) > 1 {
			if n, ok := args[1].(int64); ok {
				limit = int(n)
			}
		}
		return rt.PhpxBacktrace(flags&1 != 0, limit)
	})
	rt.RegisterFunc("get_loaded_extensions", func(_ ...any) *model.Array {
		out := model.NewArray()
		for _, name := range []string{"Core", "date", "json", "pcre", "standard", "ctype", "SPL"} {
			out.Append(name)
		}
		return out
	})
	rt.RegisterFunc("enum_exists", func(_ string, _ ...bool) bool { return false })
	rt.RegisterFunc("set_exception_handler", func(handler any) any { return nil })
	rt.RegisterFunc("restore_exception_handler", func() bool { return true })
	rt.RegisterFunc("is_infinite", func(v float64) bool { return math.IsInf(v, 0) })
	rt.RegisterFunc("bindec", func(s string) any { return parseBase(s, 2) })
	rt.RegisterFunc("dechex", func(n int64) string { return strconv.FormatUint(uint64(n), 16) })
	rt.RegisterFunc("decbin", func(n int64) string { return strconv.FormatUint(uint64(n), 2) })
	rt.RegisterFunc("decoct", func(n int64) string { return strconv.FormatUint(uint64(n), 8) })
	rt.RegisterFunc("base_convert", func(num string, from, to int64) string {
		n, err := strconv.ParseUint(strings.ToLower(num), int(from), 64)
		if err != nil {
			return "0"
		}
		return strconv.FormatUint(n, int(to))
	})
	rt.RegisterFunc("strpbrk", func(s, chars string) any {
		if i := strings.IndexAny(s, chars); i >= 0 {
			return s[i:]
		}
		return false
	})
	rt.RegisterFunc("tempnam", func(dir, prefix string) any {
		f, err := os.CreateTemp(dir, prefix)
		if err != nil {
			return false
		}
		defer f.Close()
		return f.Name()
	})
	rt.RegisterFunc("file", func(name string, rest ...any) any {
		var content string
		if strings.Contains(name, "://") || strings.HasPrefix(name, "data:") {
			get, ok := rt.Callable("file_get_contents")
			if !ok {
				return false
			}
			res, err := get(name)
			s, isString := res.(string)
			if err != nil || !isString {
				return false
			}
			content = s
		} else {
			b, err := os.ReadFile(name)
			if err != nil {
				return false
			}
			content = string(b)
		}
		flags := int64(0)
		if len(rest) > 0 {
			flags = phpval.Int(rest[0])
		}
		return splitFileLines(content, flags)
	})
}

// splitFileLines cuts content into lines the way file() does: each keeps its
// newline unless FILE_IGNORE_NEW_LINES (2) is set, and FILE_SKIP_EMPTY_LINES
// (4) drops the empty ones, which only exist once the newline is gone.
func splitFileLines(content string, flags int64) *model.Array {
	out := model.NewArray()
	for len(content) > 0 {
		line := content
		if i := strings.IndexByte(content, '\n'); i >= 0 {
			line, content = content[:i+1], content[i+1:]
		} else {
			content = ""
		}
		if flags&2 != 0 {
			line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			if line == "" && flags&4 != 0 {
				continue
			}
		}
		out.Append(line)
	}
	return out
}

func parseBase(s string, base int) any {
	n, err := strconv.ParseUint(s, base, 64)
	if err != nil {
		return int64(0)
	}
	if n > math.MaxInt64 {
		return float64(n)
	}
	return int64(n)
}
