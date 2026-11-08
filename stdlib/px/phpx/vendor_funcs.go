//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/modelutil"
	"github.com/titpetric/phpx/internal/phpval"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\VendorFuncs`, registerPhpxVendorFuncs)
}

// registerPhpxVendorFuncs installs what the composer libraries call: url
// parsing, stream metadata, the small string and filesystem helpers.
func registerPhpxVendorFuncs(rt *runner.Runtime) {
	rt.RegisterFunc("fileinode", func(path string) any {
		info, err := os.Stat(path)
		if err != nil {
			return false
		}
		if st, ok := info.Sys().(*syscall.Stat_t); ok {
			return int64(st.Ino)
		}
		return int64(0)
	})
	rt.RegisterFunc("parse_url", func(raw string, component ...int64) any {
		parts := phpxParseURL(raw)
		if parts == nil {
			return false
		}
		if len(component) > 0 && component[0] >= 0 {
			keys := []string{"scheme", "host", "port", "user", "pass", "path", "query", "fragment"}
			if int(component[0]) >= len(keys) {
				return false
			}
			v, ok := parts.Get(keys[component[0]])
			if !ok {
				return nil
			}
			return v
		}
		return parts
	})
	rt.RegisterFunc("stream_get_meta_data", func(_ any) *model.Array {
		out := model.NewArray()
		out.Set("timed_out", false)
		out.Set("blocked", true)
		out.Set("eof", false)
		out.Set("wrapper_type", "plainfile")
		out.Set("stream_type", "STDIO")
		out.Set("mode", "r+")
		out.Set("unread_bytes", int64(0))
		out.Set("seekable", true)
		out.Set("uri", "")
		return out
	})
	rt.RegisterFunc("fstat", func(f any) any {
		file, ok := f.(*os.File)
		if !ok {
			return false
		}
		fi, err := file.Stat()
		if err != nil {
			return false
		}
		out := model.NewArray()
		out.Set("size", fi.Size())
		out.Set("mtime", fi.ModTime().Unix())
		out.Set("mode", int64(fi.Mode()))
		return out
	})
	rt.RegisterFunc("http_build_query", func(data any, rest ...any) string {
		sep, enc := "&", int64(1)
		if len(rest) > 1 {
			if v, ok := rest[1].(string); ok && v != "" {
				sep = v
			}
		}
		if len(rest) > 2 {
			if v, ok := rest[2].(int64); ok {
				enc = v
			}
		}
		prefix := ""
		if len(rest) > 0 {
			prefix, _ = rest[0].(string)
		}
		var parts []string
		escape := func(s string) string {
			if enc == 2 {
				return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
			}
			return strings.ReplaceAll(url.QueryEscape(s), "~", "%7E")
		}
		var walk func(key string, v any)
		walk = func(key string, v any) {
			switch t := v.(type) {
			case *model.Array:
				t.Range(func(k, val any) bool {
					walk(key+"["+toString(k)+"]", val)
					return true
				})
			case nil:
			case bool:
				if t {
					parts = append(parts, escape(key)+"=1")
				} else {
					parts = append(parts, escape(key)+"=0")
				}
			default:
				parts = append(parts, escape(key)+"="+escape(toString(v)))
			}
		}
		if arr, ok := data.(*model.Array); ok {
			arr.Range(func(k, val any) bool {
				key := toString(k)
				if _, isInt := k.(int64); isInt {
					key = prefix + key
				}
				if _, isArr := val.(*model.Array); isArr {
					// php writes nested keys as name[sub]; the brackets are escaped.
					walk(key, val)
				} else {
					walk(key, val)
				}
				return true
			})
		}
		return strings.Join(parts, sep)
	})
	rt.RegisterFunc("fflush", func(_ any) bool { return true })
	rt.RegisterFunc("vsprintf", func(format string, args *model.Array) (any, error) {
		vals := []any{format}
		if args != nil {
			args.Range(func(_, v any) bool { vals = append(vals, v); return true })
		}
		return rt.InvokeNamed("sprintf", vals...)
	})
	rt.RegisterFunc("error_get_last", func() any { return nil })
	rt.RegisterFunc("flock", func(_ any, _ int64, _ ...any) bool { return true })
	rt.RegisterFunc("stream_set_chunk_size", func(_ any, n int64) int64 { return n })
	rt.RegisterFunc("stream_set_timeout", func(_ any, _ ...int64) bool { return true })
	rt.RegisterFunc("stream_set_blocking", func(_ any, _ bool) bool { return true })
	rt.RegisterFunc("is_link", func(name string) bool {
		fi, err := os.Lstat(name)
		return err == nil && fi.Mode()&os.ModeSymlink != 0
	})
	rt.RegisterFunc("sys_get_temp_dir", func() string { return os.TempDir() })
	rt.RegisterFunc("strspn", func(s, mask string, rest ...int64) int64 { return phpxSpan(s, mask, true, rest) })
	rt.RegisterFunc("strip_tags", func(s string, _ ...any) string {
		return regexp.MustCompile(`<[^>]*>`).ReplaceAllString(s, "")
	})
	rt.RegisterFunc("stripcslashes", func(s string) string {
		var b strings.Builder
		for i := 0; i < len(s); i++ {
			if s[i] == '\\' && i+1 < len(s) {
				i++
				switch s[i] {
				case 'n':
					b.WriteByte('\n')
				case 't':
					b.WriteByte('\t')
				case 'r':
					b.WriteByte('\r')
				default:
					b.WriteByte(s[i])
				}
				continue
			}
			b.WriteByte(s[i])
		}
		return b.String()
	})
	rt.RegisterFunc("preg_grep", func(pattern string, input any, flags ...any) (any, error) {
		invert := len(flags) > 0 && phpval.Int(flags[0])&1 != 0
		out := model.NewArray()
		var err error
		modelutil.RangeValues(input, func(k, v any) bool {
			var m any
			m, err = rt.InvokeNamed("preg_match", pattern, toString(v))
			if err != nil {
				return false
			}
			n, _ := m.(int64)
			if (n > 0) != invert {
				out.Set(k, v)
			}
			return true
		})
		return out, err
	})
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if f, ok := v.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return strconv.FormatInt(toInt(v), 10)
}

func toInt(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	}
	return 0
}

// phpxParseURL splits a URL as php's parse_url does, in php's key order.
func phpxParseURL(raw string) *model.Array {
	out := model.NewArray()
	rest := raw
	var fragment, query *string
	if i := strings.IndexByte(rest, '#'); i >= 0 {
		f := rest[i+1:]
		fragment = &f
		rest = rest[:i]
	}
	if i := strings.IndexByte(rest, '?'); i >= 0 {
		q := rest[i+1:]
		query = &q
		rest = rest[:i]
	}
	scheme := ""
	if i := strings.IndexByte(rest, ':'); i > 0 && phpxIsScheme(rest[:i]) {
		after := rest[i+1:]
		// "host:80" and "host:80/path" are an authority and a port, not a
		// scheme: what follows the colon is all digits up to a slash.
		portOnly := after != ""
		for j := 0; j < len(after) && after[j] != '/'; j++ {
			if after[j] < '0' || after[j] > '9' {
				portOnly = false
				break
			}
		}
		if after != "" && after[0] == '/' {
			portOnly = false
		}
		if !portOnly {
			scheme = rest[:i]
			rest = after
			out.Set("scheme", scheme)
		}
	}
	hasAuthority := false
	if strings.HasPrefix(rest, "//") {
		hasAuthority = true
		rest = rest[2:]
	} else if scheme == "" && rest != "" && rest[0] != '/' && out.Len() == 0 {
		// "host:80" with no scheme reads as an authority too.
		if i := strings.IndexByte(rest, ':'); i > 0 {
			hasAuthority = true
		}
	}
	if hasAuthority {
		authority := rest
		path := ""
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			authority, path = rest[:i], rest[i:]
		}
		rest = path
		var user, pass *string
		if i := strings.LastIndexByte(authority, '@'); i >= 0 {
			info := authority[:i]
			authority = authority[i+1:]
			if j := strings.IndexByte(info, ':'); j >= 0 {
				u, pw := info[:j], info[j+1:]
				user, pass = &u, &pw
			} else {
				user = &info
			}
		}
		host, port := authority, ""
		if strings.HasPrefix(authority, "[") {
			if j := strings.IndexByte(authority, ']'); j >= 0 {
				host = authority[:j+1]
				if strings.HasPrefix(authority[j+1:], ":") {
					port = authority[j+2:]
				}
			}
		} else if j := strings.LastIndexByte(authority, ':'); j >= 0 {
			host, port = authority[:j], authority[j+1:]
		}
		if host == "" && scheme != "file" {
			return nil
		}
		if host != "" {
			out.Set("host", host)
		}
		if port != "" {
			n, err := strconv.ParseInt(port, 10, 64)
			if err != nil || n > 65535 {
				return nil
			}
			out.Set("port", n)
		}
		if user != nil {
			out.Set("user", *user)
		}
		if pass != nil {
			out.Set("pass", *pass)
		}
	}
	if rest != "" {
		out.Set("path", rest)
	}
	if query != nil {
		out.Set("query", *query)
	}
	if fragment != nil {
		out.Set("fragment", *fragment)
	}
	return out
}

// phpxIsScheme reports whether s is a URL scheme: a letter, then letters,
// digits, plus, minus or dot.
func phpxIsScheme(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		letter := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
		if i == 0 && !letter {
			return false
		}
		if !letter && !(c >= '0' && c <= '9') && c != '+' && c != '-' && c != '.' {
			return false
		}
	}
	return s != ""
}
