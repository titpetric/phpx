//go:build phpx

package phpx

import (
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxVendorFuncs)
}

// registerPhpxVendorFuncs installs what the composer libraries call: url
// parsing, stream metadata, the small string and filesystem helpers.
func registerPhpxVendorFuncs(rt *runner.Runtime) {
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
	rt.RegisterFunc("trigger_error", func(_ string, _ ...int64) bool { return true })
	rt.RegisterFunc("user_error", func(_ string, _ ...int64) bool { return true })
	rt.RegisterFunc("error_get_last", func() any { return nil })
	rt.RegisterFunc("flock", func(_ any, _ int64, _ ...any) bool { return true })
	rt.RegisterFunc("stream_set_chunk_size", func(_ any, n int64) int64 { return n })
	rt.RegisterFunc("stream_set_timeout", func(_ any, _ ...int64) bool { return true })
	rt.RegisterFunc("stream_set_blocking", func(_ any, _ bool) bool { return true })
	rt.RegisterFunc("is_link", func(name string) bool {
		fi, err := os.Lstat(name)
		return err == nil && fi.Mode()&os.ModeSymlink != 0
	})
	rt.RegisterFunc("rmdir", func(name string) bool { return os.Remove(name) == nil })
	rt.RegisterFunc("sys_get_temp_dir", func() string { return os.TempDir() })
	rt.RegisterFunc("strspn", func(s, mask string, rest ...int64) int64 {
		n := 0
		for n < len(s) && strings.IndexByte(mask, s[n]) >= 0 {
			n++
		}
		return int64(n)
	})
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
	rt.RegisterFunc("preg_grep", func(pattern string, input *model.Array) (any, error) {
		out := model.NewArray()
		var err error
		input.Range(func(k, v any) bool {
			var m any
			m, err = rt.InvokeNamed("preg_match", pattern, toString(v))
			if err != nil {
				return false
			}
			if n, _ := m.(int64); n > 0 {
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
	u, err := url.Parse(raw)
	if err != nil {
		return nil
	}
	out := model.NewArray()
	if u.Scheme != "" {
		out.Set("scheme", u.Scheme)
	}
	if h := u.Hostname(); h != "" {
		out.Set("host", h)
	}
	if p := u.Port(); p != "" {
		n, _ := strconv.ParseInt(p, 10, 64)
		out.Set("port", n)
	}
	if u.User != nil {
		out.Set("user", u.User.Username())
		if pw, ok := u.User.Password(); ok {
			out.Set("pass", pw)
		}
	}
	if u.Path != "" {
		out.Set("path", u.Path)
	}
	if u.RawQuery != "" {
		out.Set("query", u.RawQuery)
	}
	if u.Fragment != "" {
		out.Set("fragment", u.Fragment)
	}
	return out
}
