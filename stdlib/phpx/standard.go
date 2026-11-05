//go:build phpx

package phpx

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"io"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxStandard)
}

// phpxStandardConstants are the filter, stream and idna constants php defines
// and a library reads by name.
var phpxStandardConstants = map[string]int64{
	"EXTR_OVERWRITE": 0, "EXTR_SKIP": 1, "EXTR_PREFIX_SAME": 2, "EXTR_PREFIX_ALL": 3, "EXTR_PREFIX_INVALID": 4,
	"EXTR_PREFIX_IF_EXISTS": 5, "EXTR_IF_EXISTS": 6, "EXTR_REFS": 256,
	"IDNA_DEFAULT": 48, "IDNA_ALLOW_UNASSIGNED": 1, "IDNA_USE_STD3_RULES": 2, "IDNA_CHECK_BIDI": 4,
	"IDNA_CHECK_CONTEXTJ": 8, "IDNA_NONTRANSITIONAL_TO_ASCII": 16, "IDNA_NONTRANSITIONAL_TO_UNICODE": 32,
	"INTL_IDNA_VARIANT_UTS46": 1, "INTL_IDNA_VARIANT_2003": 0,
	"INPUT_POST":                          0,
	"INPUT_GET":                           1,
	"INPUT_COOKIE":                        2,
	"INPUT_ENV":                           4,
	"INPUT_SERVER":                        5,
	"FILTER_FLAG_NONE":                    0,
	"FILTER_REQUIRE_SCALAR":               33554432,
	"FILTER_REQUIRE_ARRAY":                16777216,
	"FILTER_FORCE_ARRAY":                  67108864,
	"FILTER_NULL_ON_FAILURE":              134217728,
	"FILTER_VALIDATE_INT":                 257,
	"FILTER_VALIDATE_BOOLEAN":             258,
	"FILTER_VALIDATE_BOOL":                258,
	"FILTER_VALIDATE_FLOAT":               259,
	"FILTER_VALIDATE_REGEXP":              272,
	"FILTER_VALIDATE_DOMAIN":              277,
	"FILTER_VALIDATE_URL":                 273,
	"FILTER_VALIDATE_EMAIL":               274,
	"FILTER_VALIDATE_IP":                  275,
	"FILTER_VALIDATE_MAC":                 276,
	"FILTER_DEFAULT":                      516,
	"FILTER_UNSAFE_RAW":                   516,
	"FILTER_SANITIZE_STRING":              513,
	"FILTER_SANITIZE_STRIPPED":            513,
	"FILTER_SANITIZE_ENCODED":             514,
	"FILTER_SANITIZE_SPECIAL_CHARS":       515,
	"FILTER_SANITIZE_FULL_SPECIAL_CHARS":  522,
	"FILTER_SANITIZE_EMAIL":               517,
	"FILTER_SANITIZE_URL":                 518,
	"FILTER_SANITIZE_NUMBER_INT":          519,
	"FILTER_SANITIZE_NUMBER_FLOAT":        520,
	"FILTER_SANITIZE_ADD_SLASHES":         523,
	"FILTER_CALLBACK":                     1024,
	"FILTER_FLAG_ALLOW_OCTAL":             1,
	"FILTER_FLAG_ALLOW_HEX":               2,
	"FILTER_FLAG_STRIP_LOW":               4,
	"FILTER_FLAG_STRIP_HIGH":              8,
	"FILTER_FLAG_STRIP_BACKTICK":          512,
	"FILTER_FLAG_ENCODE_LOW":              16,
	"FILTER_FLAG_ENCODE_HIGH":             32,
	"FILTER_FLAG_ENCODE_AMP":              64,
	"FILTER_FLAG_NO_ENCODE_QUOTES":        128,
	"FILTER_FLAG_EMPTY_STRING_NULL":       256,
	"FILTER_FLAG_ALLOW_FRACTION":          4096,
	"FILTER_FLAG_ALLOW_THOUSAND":          8192,
	"FILTER_FLAG_ALLOW_SCIENTIFIC":        16384,
	"FILTER_FLAG_PATH_REQUIRED":           262144,
	"FILTER_FLAG_QUERY_REQUIRED":          524288,
	"FILTER_FLAG_IPV4":                    1048576,
	"FILTER_FLAG_IPV6":                    2097152,
	"FILTER_FLAG_NO_RES_RANGE":            4194304,
	"FILTER_FLAG_NO_PRIV_RANGE":           8388608,
	"FILTER_FLAG_GLOBAL_RANGE":            268435456,
	"FILTER_FLAG_HOSTNAME":                1048576,
	"FILTER_FLAG_EMAIL_UNICODE":           1048576,
	"STREAM_NOTIFY_CONNECT":               2,
	"STREAM_NOTIFY_AUTH_REQUIRED":         3,
	"STREAM_NOTIFY_AUTH_RESULT":           10,
	"STREAM_NOTIFY_MIME_TYPE_IS":          4,
	"STREAM_NOTIFY_FILE_SIZE_IS":          5,
	"STREAM_NOTIFY_REDIRECTED":            6,
	"STREAM_NOTIFY_PROGRESS":              7,
	"STREAM_NOTIFY_FAILURE":               9,
	"STREAM_NOTIFY_COMPLETED":             8,
	"STREAM_NOTIFY_RESOLVE":               1,
	"STREAM_NOTIFY_SEVERITY_INFO":         0,
	"STREAM_NOTIFY_SEVERITY_WARN":         1,
	"STREAM_NOTIFY_SEVERITY_ERR":          2,
	"STREAM_FILTER_READ":                  1,
	"STREAM_FILTER_WRITE":                 2,
	"STREAM_FILTER_ALL":                   3,
	"STREAM_CLIENT_PERSISTENT":            1,
	"STREAM_CLIENT_ASYNC_CONNECT":         2,
	"STREAM_CLIENT_CONNECT":               4,
	"STREAM_CRYPTO_METHOD_ANY_CLIENT":     127,
	"STREAM_CRYPTO_METHOD_SSLv2_CLIENT":   3,
	"STREAM_CRYPTO_METHOD_SSLv3_CLIENT":   5,
	"STREAM_CRYPTO_METHOD_SSLv23_CLIENT":  57,
	"STREAM_CRYPTO_METHOD_TLS_CLIENT":     121,
	"STREAM_CRYPTO_METHOD_TLSv1_0_CLIENT": 9,
	"STREAM_CRYPTO_METHOD_TLSv1_1_CLIENT": 17,
	"STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT": 33,
	"STREAM_CRYPTO_METHOD_TLSv1_3_CLIENT": 65,
	"STREAM_CRYPTO_METHOD_ANY_SERVER":     126,
	"STREAM_CRYPTO_METHOD_SSLv2_SERVER":   2,
	"STREAM_CRYPTO_METHOD_SSLv3_SERVER":   4,
	"STREAM_CRYPTO_METHOD_SSLv23_SERVER":  120,
	"STREAM_CRYPTO_METHOD_TLS_SERVER":     120,
	"STREAM_CRYPTO_METHOD_TLSv1_0_SERVER": 8,
	"STREAM_CRYPTO_METHOD_TLSv1_1_SERVER": 16,
	"STREAM_CRYPTO_METHOD_TLSv1_2_SERVER": 32,
	"STREAM_CRYPTO_METHOD_TLSv1_3_SERVER": 64,
	"STREAM_CRYPTO_PROTO_SSLv3":           4,
	"STREAM_CRYPTO_PROTO_TLSv1_0":         8,
	"STREAM_CRYPTO_PROTO_TLSv1_1":         16,
	"STREAM_CRYPTO_PROTO_TLSv1_2":         32,
	"STREAM_CRYPTO_PROTO_TLSv1_3":         64,
	"STREAM_SHUT_RD":                      0,
	"STREAM_SHUT_WR":                      1,
	"STREAM_SHUT_RDWR":                    2,
	"STREAM_PF_INET":                      2,
	"STREAM_PF_INET6":                     10,
	"STREAM_PF_UNIX":                      1,
	"STREAM_IPPROTO_IP":                   0,
	"STREAM_IPPROTO_TCP":                  6,
	"STREAM_IPPROTO_UDP":                  17,
	"STREAM_IPPROTO_ICMP":                 1,
	"STREAM_IPPROTO_RAW":                  255,
	"STREAM_SOCK_STREAM":                  1,
	"STREAM_SOCK_DGRAM":                   2,
	"STREAM_SOCK_RAW":                     3,
	"STREAM_SOCK_SEQPACKET":               5,
	"STREAM_SOCK_RDM":                     4,
	"STREAM_PEEK":                         2,
	"STREAM_OOB":                          1,
	"STREAM_SERVER_BIND":                  4,
	"STREAM_SERVER_LISTEN":                8,
	"STREAM_USE_PATH":                     1,
	"STREAM_IGNORE_URL":                   2,
	"STREAM_REPORT_ERRORS":                8,
	"STREAM_MUST_SEEK":                    16,
	"STREAM_URL_STAT_LINK":                1,
	"STREAM_URL_STAT_QUIET":               2,
	"STREAM_MKDIR_RECURSIVE":              1,
	"STREAM_IS_URL":                       1,
	"STREAM_OPTION_BLOCKING":              1,
	"STREAM_OPTION_READ_TIMEOUT":          4,
	"STREAM_OPTION_READ_BUFFER":           2,
	"STREAM_OPTION_WRITE_BUFFER":          3,
	"STREAM_BUFFER_NONE":                  0,
	"STREAM_BUFFER_LINE":                  1,
	"STREAM_BUFFER_FULL":                  2,
	"STREAM_CAST_AS_STREAM":               0,
	"STREAM_CAST_FOR_SELECT":              3,
	"STREAM_META_TOUCH":                   1,
	"STREAM_META_OWNER":                   3,
	"STREAM_META_OWNER_NAME":              2,
	"STREAM_META_GROUP":                   5,
	"STREAM_META_GROUP_NAME":              4,
	"STREAM_META_ACCESS":                  6,
}

// registerPhpxStandard installs the functions of php's standard, filter, zlib
// and network address families that libraries call and the base runtime lacks.
func registerPhpxStandard(rt *runner.Runtime) {
	for name, value := range phpxStandardConstants {
		rt.SetConst(name, value)
	}
	rt.RegisterFunc("getenv", func(name string, _ ...any) any {
		if value, ok := rt.Env[name]; ok {
			return value
		}
		return false
	})
	ini := map[string]string{
		"allow_url_fopen": "1", "memory_limit": "128M", "precision": "14", "serialize_precision": "-1",
		"default_charset": "UTF-8", "display_errors": "", "error_reporting": "", "max_execution_time": "0",
		"date.timezone": "UTC", "zend.assertions": "-1", "user_agent": "", "default_socket_timeout": "60",
	}
	rt.RegisterFunc("ini_get", func(option string) any {
		if value, ok := ini[option]; ok {
			return value
		}
		return false
	})
	rt.RegisterFunc("ini_set", func(option string, value any) any {
		old, ok := ini[option]
		ini[option] = phpval.String(value)
		if !ok {
			return false
		}
		return old
	})
	rt.RegisterFunc("filter_var", func(value any, rest ...any) any { return phpxFilterVar(rt, value, rest) })
	rt.RegisterFunc("inet_ntop", func(packed string) any {
		switch len(packed) {
		case 4, 16:
			return net.IP([]byte(packed)).String()
		}
		return false
	})
	rt.RegisterFunc("inet_pton", func(address string) any {
		ip := net.ParseIP(address)
		if ip == nil {
			return false
		}
		if v4 := ip.To4(); v4 != nil && !strings.Contains(address, ":") {
			return string(v4)
		}
		return string(ip.To16())
	})
	rt.RegisterFunc("ip2long", func(address string) any {
		ip := net.ParseIP(address).To4()
		if ip == nil || strings.Contains(address, ":") {
			return false
		}
		return int64(ip[0])<<24 | int64(ip[1])<<16 | int64(ip[2])<<8 | int64(ip[3])
	})
	rt.RegisterFunc("long2ip", func(n int64) string {
		return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n)).String()
	})
	rt.RegisterFunc("gzencode", func(data string, rest ...int64) any { return phpxCompress(data, "gzip", rest) })
	rt.RegisterFunc("gzcompress", func(data string, rest ...int64) any { return phpxCompress(data, "zlib", rest) })
	rt.RegisterFunc("gzdeflate", func(data string, rest ...int64) any { return phpxCompress(data, "raw", rest) })
	rt.RegisterFunc("gzdecode", func(data string, _ ...int64) any { return phpxDecompress(data, "gzip") })
	rt.RegisterFunc("gzuncompress", func(data string, _ ...int64) any { return phpxDecompress(data, "zlib") })
	rt.RegisterFunc("gzinflate", func(data string, _ ...int64) any { return phpxDecompress(data, "raw") })
	rt.RegisterFunc("stream_get_wrappers", func() *model.Array {
		out := model.NewArray()
		for _, name := range []string{"https", "ftps", "compress.zlib", "php", "file", "glob", "data", "http", "ftp"} {
			out.Append(name)
		}
		return out
	})
	rt.RegisterFunc("extract", func(ctx context.Context, array any, rest ...any) int64 {
		scope, ok := runner.ScopeFromContext(ctx)
		if !ok || !model.IsCollection(array) {
			return 0
		}
		flags, prefix := int64(0), ""
		if len(rest) > 0 {
			flags = phpval.Int(rest[0]) &^ 256
		}
		if len(rest) > 1 {
			prefix = phpval.String(rest[1])
		}
		valid := func(name string) bool {
			if name == "" || name == "this" {
				return false
			}
			for i, r := range name {
				if !(r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= 0x80 || (i > 0 && r >= '0' && r <= '9')) {
					return false
				}
			}
			return true
		}
		var count int64
		model.RangeValues(array, func(k, v any) bool {
			name := phpval.String(k)
			_, exists := scope.Get(name)
			switch flags {
			case 1: // EXTR_SKIP
				if exists {
					return true
				}
			case 2: // EXTR_PREFIX_SAME
				if exists {
					name = prefix + "_" + name
				}
			case 3: // EXTR_PREFIX_ALL
				name = prefix + "_" + name
			case 4: // EXTR_PREFIX_INVALID
				if !valid(name) {
					name = prefix + "_" + name
				}
			case 5: // EXTR_PREFIX_IF_EXISTS
				if !exists {
					return true
				}
				name = prefix + "_" + name
			case 6: // EXTR_IF_EXISTS
				if !exists {
					return true
				}
			}
			if !valid(name) {
				return true
			}
			scope.Set(name, v)
			count++
			return true
		})
		return count
	})
	rt.RegisterFunc("wordwrap", func(str string, rest ...any) (string, error) {
		width, brk, cut := int64(75), "\n", false
		if len(rest) > 0 {
			width = phpval.Int(rest[0])
		}
		if len(rest) > 1 {
			brk = phpval.String(rest[1])
		}
		if len(rest) > 2 {
			cut = phpval.Truthy(rest[2])
		}
		if brk == "" {
			return "", &valueError{message: "wordwrap(): Argument #3 ($break) cannot be empty"}
		}
		if width == 0 && cut {
			return "", &valueError{message: "wordwrap(): Argument #4 ($cut_long_words) cannot be true when argument #2 ($width) is 0"}
		}
		return phpxWordwrap(str, int(width), brk, cut), nil
	})
	rt.RegisterFunc("stream_copy_to_stream", func(from io.Reader, to io.Writer, rest ...any) any {
		if from == nil || to == nil {
			return false
		}
		if len(rest) > 1 {
			if seeker, ok := from.(io.Seeker); ok {
				_, _ = seeker.Seek(phpval.Int(rest[1]), io.SeekStart)
			}
		}
		src := from
		if len(rest) > 0 && rest[0] != nil {
			src = io.LimitReader(from, phpval.Int(rest[0]))
		}
		n, err := io.Copy(to, src)
		if err != nil {
			return false
		}
		return n
	})
	rt.RegisterFunc("array_replace_recursive", func(base any, others ...any) any {
		out := model.CopyArray(model.ToArray(base))
		for _, other := range others {
			phpxReplaceRecursive(out, other)
		}
		return out
	})
	rt.RegisterFunc("grapheme_strlen", func(s string) any { return int64(len(phpxGraphemes(s))) })
	rt.RegisterFunc("grapheme_substr", func(s string, start int64, rest ...int64) any {
		g := phpxGraphemes(s)
		n := int64(len(g))
		if start < 0 {
			start += n
			if start < 0 {
				start = 0
			}
		}
		if start > n {
			return false
		}
		end := n
		if len(rest) > 0 {
			if rest[0] < 0 {
				end = n + rest[0]
			} else if start+rest[0] < n {
				end = start + rest[0]
			}
		}
		if end < start {
			return ""
		}
		return strings.Join(g[start:end], "")
	})
	rt.RegisterFunc("grapheme_strpos", func(haystack, needle string, rest ...int64) any {
		g := phpxGraphemes(haystack)
		offset := int64(0)
		if len(rest) > 0 {
			offset = rest[0]
		}
		if offset < 0 {
			offset += int64(len(g))
		}
		if offset < 0 || offset > int64(len(g)) {
			return false
		}
		prefix := strings.Join(g[:offset], "")
		i := strings.Index(haystack[len(prefix):], needle)
		if i < 0 {
			return false
		}
		return int64(len(phpxGraphemes(haystack[:len(prefix)+i])))
	})
	rt.RegisterFunc("strnatcmp", func(a, b string) int64 { return phpxNatCompare(a, b, false) })
	rt.RegisterFunc("strnatcasecmp", func(a, b string) int64 { return phpxNatCompare(a, b, true) })
	rt.RegisterFunc("get_class_methods", func(class any) (any, error) {
		names, err := rt.PhpxMethodNames(class)
		if err != nil {
			return nil, err
		}
		out := model.NewArray()
		for _, n := range names {
			out.Append(n)
		}
		return out, nil
	})
	rt.RegisterFunc("mb_strimwidth", func(s string, start, width int64, rest ...string) string {
		trim := ""
		if len(rest) > 0 {
			trim = rest[0]
		}
		runes := []rune(s)
		if start < 0 {
			start += int64(len(runes))
		}
		if start > int64(len(runes)) {
			start = int64(len(runes))
		}
		runes = runes[start:]
		if int64(len(runes)) <= width {
			return string(runes)
		}
		keep := width - int64(len([]rune(trim)))
		if keep < 0 {
			keep = 0
		}
		return string(runes[:keep]) + trim
	})
}

// phpxGraphemes splits s into user-perceived characters: a rune with the
// combining marks and joiners that follow it.
func phpxGraphemes(s string) []string {
	var out []string
	joined := false
	for _, r := range s {
		if len(out) > 0 && (joined || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Mc, r) || r == 0x200d || (r >= 0xfe00 && r <= 0xfe0f) || (r >= 0x1f3fb && r <= 0x1f3ff)) {
			out[len(out)-1] += string(r)
			joined = r == 0x200d
			continue
		}
		joined = false
		out = append(out, string(r))
	}
	return out
}

// phpxReplaceRecursive merges src into dst, descending where both hold arrays.
func phpxReplaceRecursive(dst *model.Array, src any) {
	model.RangeValues(src, func(k, v any) bool {
		if existing, ok := dst.Get(k); ok && model.IsCollection(existing) && model.IsCollection(v) {
			child := model.CopyArray(model.ToArray(existing))
			phpxReplaceRecursive(child, v)
			dst.Set(k, child)
			return true
		}
		dst.Set(k, model.CopyValue(v))
		return true
	})
}

// phpxWordwrap is php's wordwrap(), ported from its general algorithm: lines
// break at the last space inside the width and, with cut, inside a longer word.
func phpxWordwrap(text string, width int, brk string, cut bool) string {
	if text == "" {
		return ""
	}
	var b strings.Builder
	lastStart, lastSpace := 0, 0
	for cur := 0; cur < len(text); cur++ {
		switch {
		case text[cur] == brk[0] && cur+len(brk) < len(text) && strings.HasPrefix(text[cur:], brk):
			b.WriteString(text[lastStart : cur+len(brk)])
			cur += len(brk) - 1
			lastStart = cur + 1
			lastSpace = lastStart
		case text[cur] == ' ':
			if cur-lastStart >= width {
				b.WriteString(text[lastStart:cur])
				b.WriteString(brk)
				lastStart = cur + 1
			}
			lastSpace = cur
		case cur-lastStart >= width && cut && lastStart >= lastSpace:
			b.WriteString(text[lastStart:cur])
			b.WriteString(brk)
			lastStart = cur
			lastSpace = cur
		case cur-lastStart >= width && lastStart < lastSpace:
			b.WriteString(text[lastStart:lastSpace])
			b.WriteString(brk)
			lastSpace++
			lastStart = lastSpace
		}
	}
	if lastStart != len(text) {
		b.WriteString(text[lastStart:])
	}
	return b.String()
}

func phpxCompress(data, kind string, level []int64) any {
	lv := -1
	if len(level) > 0 {
		lv = int(level[0])
	}
	var buf bytes.Buffer
	var w io.WriteCloser
	var err error
	switch kind {
	case "gzip":
		w, err = gzip.NewWriterLevel(&buf, lv)
	case "zlib":
		w, err = zlib.NewWriterLevel(&buf, lv)
	default:
		w, err = flate.NewWriter(&buf, lv)
	}
	if err != nil {
		return false
	}
	if _, err := w.Write([]byte(data)); err != nil {
		return false
	}
	if err := w.Close(); err != nil {
		return false
	}
	return buf.String()
}

func phpxDecompress(data, kind string) any {
	var r io.Reader
	var err error
	switch kind {
	case "gzip":
		r, err = gzip.NewReader(strings.NewReader(data))
	case "zlib":
		r, err = zlib.NewReader(strings.NewReader(data))
	default:
		r = flate.NewReader(strings.NewReader(data))
	}
	if err != nil {
		return false
	}
	out, err := io.ReadAll(r)
	if err != nil {
		return false
	}
	return string(out)
}

// phpxNatCompare is strnatcmp: digit runs compare as numbers, the rest byte by byte.
func phpxNatCompare(a, b string, fold bool) int64 {
	if fold {
		a, b = strings.ToLower(a), strings.ToLower(b)
	}
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if unicode.IsDigit(rune(a[i])) && unicode.IsDigit(rune(b[j])) {
			si := i
			for i < len(a) && unicode.IsDigit(rune(a[i])) {
				i++
			}
			sj := j
			for j < len(b) && unicode.IsDigit(rune(b[j])) {
				j++
			}
			na, nb := strings.TrimLeft(a[si:i], "0"), strings.TrimLeft(b[sj:j], "0")
			if len(na) != len(nb) {
				if len(na) < len(nb) {
					return -1
				}
				return 1
			}
			if na != nb {
				if na < nb {
					return -1
				}
				return 1
			}
			continue
		}
		if a[i] != b[j] {
			if a[i] < b[j] {
				return -1
			}
			return 1
		}
		i++
		j++
	}
	switch {
	case i < len(a):
		return 1
	case j < len(b):
		return -1
	}
	return 0
}

var phpxEmail = regexp.MustCompile(`^[A-Za-z0-9.!#$%&'*+/=?^_{|}~-]+@[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)+$`)

// phpxFilterVar is filter_var() for the validate and the common sanitize filters.
func phpxFilterVar(rt *runner.Runtime, value any, rest []any) any {
	filter := int64(516)
	var flags int64
	var options *model.Array
	if len(rest) > 0 {
		filter = phpval.Int(rest[0])
	}
	if len(rest) > 1 {
		if arr := model.ToArray(rest[1]); arr != nil && model.IsCollection(rest[1]) {
			if f, ok := arr.Get("flags"); ok {
				flags = phpval.Int(f)
			}
			if o, ok := arr.Get("options"); ok {
				options = model.ToArray(o)
				if filter == 1024 {
					return phpxFilterCallback(rt, value, o)
				}
			}
		} else {
			flags = phpval.Int(rest[1])
		}
	}
	fail := func() any {
		if flags&134217728 != 0 {
			return nil
		}
		if options != nil {
			if d, ok := options.Get("default"); ok {
				return d
			}
		}
		return false
	}
	if model.IsCollection(value) {
		return fail()
	}
	text := phpval.String(value)
	switch filter {
	case 257: // int
		t := strings.TrimSpace(text)
		var n int64
		var err error
		switch {
		case flags&2 != 0 && (strings.HasPrefix(t, "0x") || strings.HasPrefix(t, "0X")):
			n, err = strconv.ParseInt(t[2:], 16, 64)
		case flags&1 != 0 && len(t) > 1 && t[0] == '0':
			n, err = strconv.ParseInt(t[1:], 8, 64)
		default:
			if len(t) > 1 && (t[0] == '0' || strings.HasPrefix(t, "-0") || strings.HasPrefix(t, "+0")) {
				return fail()
			}
			n, err = strconv.ParseInt(strings.TrimPrefix(t, "+"), 10, 64)
		}
		if err != nil {
			return fail()
		}
		if options != nil {
			if min, ok := options.Get("min_range"); ok && n < phpval.Int(min) {
				return fail()
			}
			if max, ok := options.Get("max_range"); ok && n > phpval.Int(max) {
				return fail()
			}
		}
		return n
	case 258: // bool
		switch strings.ToLower(strings.TrimSpace(text)) {
		case "1", "true", "on", "yes":
			return true
		case "0", "false", "off", "no", "":
			if text == "" && flags&134217728 != 0 {
				return false
			}
			return false
		}
		if flags&134217728 != 0 {
			return nil
		}
		return false
	case 259: // float
		t := strings.TrimSpace(text)
		if flags&8192 != 0 {
			t = strings.ReplaceAll(t, ",", "")
		}
		f, err := strconv.ParseFloat(t, 64)
		if err != nil || t == "" || strings.ContainsAny(t, "xXpP_") || strings.EqualFold(t, "inf") || strings.EqualFold(t, "nan") {
			return fail()
		}
		return f
	case 272: // regexp
		if options != nil {
			if re, ok := options.Get("regexp"); ok {
				if call, found := rt.Callable("preg_match"); found {
					if n, err := call(re, text); err == nil && phpval.Int(n) == 1 {
						return text
					}
				}
			}
		}
		return fail()
	case 273: // url
		u, err := url.Parse(text)
		if err != nil || u.Scheme == "" || (u.Host == "" && u.Scheme != "mailto" && u.Scheme != "file" && u.Scheme != "news" && u.Scheme != "urn") {
			return fail()
		}
		if flags&262144 != 0 && u.Path == "" {
			return fail()
		}
		if flags&524288 != 0 && u.RawQuery == "" {
			return fail()
		}
		return text
	case 274: // email
		if len(text) > 320 || !phpxEmail.MatchString(text) || strings.Contains(text, "..") {
			return fail()
		}
		return text
	case 275: // ip
		return phpxFilterIP(text, flags, fail)
	case 276: // mac
		if _, err := net.ParseMAC(text); err != nil {
			return fail()
		}
		return text
	case 277: // domain
		if text == "" || len(text) > 253 {
			return fail()
		}
		return text
	case 519: // number int
		return strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' || r == '+' || r == '-' {
				return r
			}
			return -1
		}, text)
	case 520: // number float
		return strings.Map(func(r rune) rune {
			switch {
			case r >= '0' && r <= '9', r == '+', r == '-':
				return r
			case r == '.' && flags&4096 != 0, r == ',' && flags&8192 != 0, (r == 'e' || r == 'E') && flags&16384 != 0:
				return r
			}
			return -1
		}, text)
	case 517: // sanitize email
		return strings.Map(func(r rune) rune {
			if r < 128 && (unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("!#$%&'*+-=?^_`{|}~@.[]", r)) {
				return r
			}
			return -1
		}, text)
	case 515, 522: // special chars
		return strings.NewReplacer("&", "&#38;", `"`, "&#34;", "'", "&#39;", "<", "&#60;", ">", "&#62;").Replace(text)
	case 523:
		return strings.NewReplacer(`\`, `\\`, `'`, `\'`, `"`, `\"`, "\x00", `\0`).Replace(text)
	case 514:
		return url.QueryEscape(text)
	case 516:
		return value
	}
	return value
}

func phpxFilterCallback(rt *runner.Runtime, value, callback any) any {
	call, ok := rt.Callable(callback)
	if !ok {
		return false
	}
	out, err := call(value)
	if err != nil {
		return false
	}
	return out
}

// phpxFilterIP validates an address against FILTER_FLAG_IPV4/IPV6 and the
// private and reserved range flags.
func phpxFilterIP(text string, flags int64, fail func() any) any {
	ip := net.ParseIP(text)
	if ip == nil {
		return fail()
	}
	isV6 := strings.Contains(text, ":")
	v4, v6 := flags&1048576 != 0, flags&2097152 != 0
	if (v4 || v6) && !(v4 && !isV6) && !(v6 && isV6) {
		return fail()
	}
	if !isV6 {
		// A dotted quad has four plain decimal parts, no leading zeros.
		parts := strings.Split(text, ".")
		if len(parts) != 4 {
			return fail()
		}
		for _, p := range parts {
			if len(p) > 1 && p[0] == '0' {
				return fail()
			}
		}
	}
	if flags&8388608 != 0 && ip.IsPrivate() {
		return fail()
	}
	if flags&4194304 != 0 && (ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsMulticast()) {
		return fail()
	}
	return text
}
