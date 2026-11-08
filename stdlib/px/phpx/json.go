//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/modelutil"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\JSON`, registerPhpxJSON)
}

// json_encode flags, and the error codes and messages json_last_error reports.
const (
	jsonHexTag        = 1
	jsonHexAmp        = 2
	jsonHexApos       = 4
	jsonHexQuot       = 8
	jsonForceObject   = 16
	jsonNumericCheck  = 32
	jsonUnescSlashes  = 64
	jsonPretty        = 128
	jsonUnescUnicode  = 256
	jsonPartial       = 512
	jsonPreserveZero  = 1024
	jsonUnescLineTerm = 2048
	jsonIgnoreUTF8    = 1048576
	jsonSubstUTF8     = 2097152
	jsonThrow         = 4194304
	jsonObjectAsArray = 1
	jsonBigintString  = 2
)

var jsonMessages = map[int64]string{
	0: "No error", 1: "Maximum stack depth exceeded", 2: "State mismatch (invalid or malformed JSON)",
	3: "Control character error, possibly incorrectly encoded", 4: "Syntax error",
	5: "Malformed UTF-8 characters, possibly incorrectly encoded", 6: "Recursion detected",
	7: "Inf and NaN cannot be JSON encoded", 8: "Type is not supported",
}

// jsonFailure is an encode or decode error with the code php reports for it.
type jsonFailure struct{ code int64 }

func (f *jsonFailure) Error() string { return jsonMessages[f.code] }

// jsonState carries one json_encode call: the flags, the depth reached and the
// error that stops it or, under JSON_PARTIAL_OUTPUT_ON_ERROR, is only noted.
type jsonState struct {
	rt     *runner.Runtime
	prev   func(any, ...any) (any, error)
	flags  int64
	maxDep int
	failed int64
	inKey  bool
}

// registerPhpxJSON installs json_encode, json_decode, json_last_error and
// json_last_error_msg with php's flags, errors and number and string forms.
func registerPhpxJSON(rt *runner.Runtime) {
	var prev func(any, ...any) (any, error)
	if fn, ok := rt.LookupFunc("json_encode"); ok {
		prev, _ = fn.(func(any, ...any) (any, error))
	}
	var last int64
	fail := func(flags int64, code int64) (any, error) {
		last = code
		if flags&jsonThrow != 0 {
			last = 0
			return nil, &Throwable{Message: jsonMessages[code], Code: int(code), Class: "JsonException"}
		}
		return nil, nil
	}
	rt.RegisterFunc("json_encode", func(value any, rest ...any) (any, error) {
		var flags int64
		depth := int64(512)
		if len(rest) > 0 {
			flags = phpval.Int(rest[0])
		}
		if len(rest) > 1 {
			depth = phpval.Int(rest[1])
		}
		st := &jsonState{rt: rt, prev: prev, flags: flags, maxDep: int(depth)}
		var b strings.Builder
		if err := st.encode(&b, value, 0); err != nil {
			if f, ok := err.(*jsonFailure); ok {
				if flags&jsonPartial == 0 {
					last = f.code
					if flags&jsonThrow != 0 {
						last = 0
						return nil, &Throwable{Message: f.Error(), Code: int(f.code), Class: "JsonException"}
					}
					return false, nil
				}
			} else {
				return nil, err
			}
		}
		if flags&jsonThrow == 0 {
			last = st.failed
		}
		return b.String(), nil
	})
	rt.RegisterFunc("json_decode", func(text string, rest ...any) (any, error) {
		assoc := false
		depth := int64(512)
		var flags int64
		if len(rest) > 0 && rest[0] != nil {
			assoc = phpval.Truthy(rest[0])
		}
		if len(rest) > 1 {
			depth = phpval.Int(rest[1])
		}
		if len(rest) > 2 {
			flags = phpval.Int(rest[2])
			if len(rest) > 0 && rest[0] == nil && flags&jsonObjectAsArray != 0 {
				assoc = true
			}
		}
		d := &jsonDecoder{src: text, assoc: assoc, bigString: flags&jsonBigintString != 0, maxDep: int(depth)}
		value, code := d.decode()
		if code != 0 {
			return fail(flags, code)
		}
		if flags&jsonThrow == 0 {
			last = 0
		}
		return value, nil
	})
	rt.RegisterFunc("json_last_error", func() int64 { return last })
	rt.RegisterFunc("json_last_error_msg", func() string { return jsonMessages[last] })
}

// encode writes v. Strings, numbers and containers follow php's forms; an
// object is its jsonSerialize() or its public properties.
func (st *jsonState) encode(b *strings.Builder, v any, depth int) error {
	if depth > st.maxDep {
		return st.note(&jsonFailure{1}, b, "null")
	}
	switch t := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if t {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case int:
		b.WriteString(strconv.FormatInt(int64(t), 10))
	case int64:
		b.WriteString(strconv.FormatInt(t, 10))
	case float64:
		return st.float(b, t)
	case string:
		return st.str(b, t)
	case *model.Object:
		return st.object(b, t, depth)
	default:
		if modelutil.IsCollection(v) {
			return st.array(b, v, depth)
		}
		if st.prev != nil {
			out, err := st.prev(v)
			if err != nil {
				return err
			}
			if s, ok := out.(string); ok {
				b.WriteString(s)
				return nil
			}
		}
		return st.note(&jsonFailure{8}, b, "null")
	}
	return nil
}

// note records an error: it stops the encode unless partial output is asked
// for, which writes the stand-in and goes on.
func (st *jsonState) note(f *jsonFailure, b *strings.Builder, standIn string) error {
	if st.flags&jsonPartial != 0 {
		st.failed = f.code
		b.WriteString(standIn)
		return nil
	}
	return f
}

func (st *jsonState) float(b *strings.Builder, f float64) error {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return st.note(&jsonFailure{7}, b, "0")
	}
	b.WriteString(phpxJSONFloat(f, st.flags&jsonPreserveZero != 0))
	return nil
}

// phpxJSONFloat spells f as php does with serialize_precision -1: the shortest
// digits that read back, an exponent form past 1e15 or below 1e-4.
func phpxJSONFloat(f float64, keepZero bool) string {
	if f == 0 {
		sign := ""
		if math.Signbit(f) {
			sign = "-"
		}
		if keepZero {
			return sign + "0.0"
		}
		return sign + "0"
	}
	exp := strconv.FormatFloat(f, 'e', -1, 64)
	mant, e, _ := strings.Cut(exp, "e")
	n, _ := strconv.Atoi(e)
	if n < -4 || n >= 15 {
		if !strings.Contains(mant, ".") {
			mant += ".0"
		}
		sign := "+"
		if n < 0 {
			sign = "-"
			n = -n
		}
		return mant + "e" + sign + strconv.Itoa(n)
	}
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if keepZero && !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

func (st *jsonState) str(b *strings.Builder, s string) error {
	if st.flags&jsonNumericCheck != 0 && !st.inKey {
		if n, ok := phpval.StringNumber(s); ok {
			return st.encode(b, n, 0)
		}
	}
	if !utf8.ValidString(s) {
		switch {
		case st.flags&jsonIgnoreUTF8 != 0:
			s = strings.ToValidUTF8(s, "")
		case st.flags&jsonSubstUTF8 != 0:
			s = strings.ToValidUTF8(s, "\xef\xbf\xbd")
		default:
			return st.note(&jsonFailure{5}, b, "null")
		}
	}
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			if st.flags&jsonHexQuot != 0 {
				b.WriteString("\\u" + "0022")
			} else {
				b.WriteString(`\"`)
			}
		case '\\':
			b.WriteString(`\\`)
		case '/':
			if st.flags&jsonUnescSlashes != 0 {
				b.WriteByte('/')
			} else {
				b.WriteString(`\/`)
			}
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '<', '>':
			if st.flags&jsonHexTag != 0 {
				if r == '<' {
					b.WriteString("\\u" + "003C")
				} else {
					b.WriteString("\\u" + "003E")
				}
			} else {
				b.WriteRune(r)
			}
		case '&':
			if st.flags&jsonHexAmp != 0 {
				b.WriteString("\\u" + "0026")
			} else {
				b.WriteByte('&')
			}
		case '\'':
			if st.flags&jsonHexApos != 0 {
				b.WriteString("\\u" + "0027")
			} else {
				b.WriteByte('\'')
			}
		default:
			switch {
			case r < 0x20:
				b.WriteString(`\u00`)
				b.WriteByte("0123456789abcdef"[r>>4])
				b.WriteByte("0123456789abcdef"[r&15])
			case r < 0x80:
				b.WriteRune(r)
			case st.flags&jsonUnescUnicode != 0 && ((r != 0x2028 && r != 0x2029) || st.flags&jsonUnescLineTerm != 0):
				b.WriteRune(r)
			case r >= 0x10000:
				r -= 0x10000
				writeU(b, 0xd800+(r>>10))
				writeU(b, 0xdc00+(r&0x3ff))
			default:
				writeU(b, r)
			}
		}
	}
	b.WriteByte('"')
	return nil
}

func writeU(b *strings.Builder, r rune) {
	b.WriteString(`\u`)
	for shift := 12; shift >= 0; shift -= 4 {
		b.WriteByte("0123456789abcdef"[(r>>uint(shift))&15])
	}
}

// array writes a list as [..] and anything else as {..}.
func (st *jsonState) array(b *strings.Builder, a any, depth int) error {
	// A Go map has no order; its keys are written sorted, so the same map
	// encodes the same way every time.
	if reflect.ValueOf(a).Kind() == reflect.Map {
		sorted := model.NewArray()
		modelutil.RangeValues(a, func(k, v any) bool { sorted.Set(phpval.String(k), v); return true })
		keys := sorted.Keys()
		sort.Slice(keys, func(i, j int) bool { return phpval.String(keys[i]) < phpval.String(keys[j]) })
		ordered := model.NewArray()
		for _, k := range keys {
			v, _ := sorted.Get(k)
			ordered.Set(k, v)
		}
		a = ordered
	}
	list := st.flags&jsonForceObject == 0
	if list {
		i := int64(0)
		modelutil.RangeValues(a, func(k, _ any) bool {
			if n, ok := k.(int64); !ok || n != i {
				list = false
				return false
			}
			i++
			return true
		})
	}
	return st.container(b, depth, list, func(each func(key string, v any) error) error {
		var err error
		i := 0
		modelutil.RangeValues(a, func(k, v any) bool {
			key := ""
			switch t := k.(type) {
			case string:
				key = t
			case int64:
				key = strconv.FormatInt(t, 10)
			default:
				key = phpval.String(k)
			}
			if err = each(key, v); err != nil {
				return false
			}
			i++
			return true
		})
		return err
	})
}

func (st *jsonState) object(b *strings.Builder, o *model.Object, depth int) error {
	if o.Class != nil {
		if res, found, err := st.rt.PhpxCallMethod(o, "jsonSerialize"); found || err != nil {
			if err != nil {
				return err
			}
			if res == any(o) {
				return st.objectProps(b, o, depth)
			}
			return st.encode(b, res, depth)
		}
		if _, hasFrom := o.Class.Methods["from"]; hasFrom {
			if val, ok := o.Prop("value"); ok {
				return st.encode(b, val, depth)
			}
		}
	}
	return st.objectProps(b, o, depth)
}

func (st *jsonState) objectProps(b *strings.Builder, o *model.Object, depth int) error {
	return st.container(b, depth, false, func(each func(key string, v any) error) error {
		var err error
		st.rt.PhpxPublicProps(o, func(name string, v any) bool {
			err = each(name, v)
			return err == nil
		})
		return err
	})
}

// container writes the brackets, separators and indentation around the
// members walk hands it.
func (st *jsonState) container(b *strings.Builder, depth int, list bool, walk func(each func(key string, v any) error) error) error {
	open, closer := byte('{'), byte('}')
	if list {
		open, closer = '[', ']'
	}
	b.WriteByte(open)
	pretty := st.flags&jsonPretty != 0
	count := 0
	err := walk(func(key string, v any) error {
		if count > 0 {
			b.WriteByte(',')
		}
		count++
		if pretty {
			b.WriteByte('\n')
			b.WriteString(strings.Repeat("    ", depth+1))
		}
		if !list {
			st.inKey = true
			err := st.str(b, key)
			st.inKey = false
			if err != nil {
				return err
			}
			b.WriteByte(':')
			if pretty {
				b.WriteByte(' ')
			}
		}
		return st.encode(b, v, depth+1)
	})
	if err != nil {
		return err
	}
	if pretty && count > 0 {
		b.WriteByte('\n')
		b.WriteString(strings.Repeat("    ", depth))
	}
	b.WriteByte(closer)
	return nil
}

// jsonDecoder is a json_decode parse of one document.
type jsonDecoder struct {
	src       string
	pos       int
	assoc     bool
	bigString bool
	maxDep    int
}

func (d *jsonDecoder) decode() (any, int64) {
	if !utf8.ValidString(d.src) {
		return nil, 5
	}
	d.skip()
	v, code := d.value(1)
	if code != 0 {
		return nil, code
	}
	d.skip()
	if d.pos != len(d.src) {
		return nil, 4
	}
	return v, 0
}

func (d *jsonDecoder) skip() {
	for d.pos < len(d.src) {
		switch d.src[d.pos] {
		case ' ', '\t', '\n', '\r':
			d.pos++
		default:
			return
		}
	}
}

func (d *jsonDecoder) value(depth int) (any, int64) {
	if d.pos >= len(d.src) {
		return nil, 4
	}
	switch c := d.src[d.pos]; {
	case c == '{':
		return d.object(depth)
	case c == '[':
		return d.list(depth)
	case c == '"':
		return d.str()
	case c == '-' || (c >= '0' && c <= '9'):
		return d.number()
	case strings.HasPrefix(d.src[d.pos:], "true"):
		d.pos += 4
		return true, 0
	case strings.HasPrefix(d.src[d.pos:], "false"):
		d.pos += 5
		return false, 0
	case strings.HasPrefix(d.src[d.pos:], "null"):
		d.pos += 4
		return nil, 0
	}
	return nil, 4
}

func (d *jsonDecoder) number() (any, int64) {
	start := d.pos
	if d.src[d.pos] == '-' {
		d.pos++
	}
	digits := func() int {
		n := 0
		for d.pos < len(d.src) && d.src[d.pos] >= '0' && d.src[d.pos] <= '9' {
			d.pos++
			n++
		}
		return n
	}
	if d.pos < len(d.src) && d.src[d.pos] == '0' {
		d.pos++
	} else if digits() == 0 {
		return nil, 4
	}
	float := false
	if d.pos < len(d.src) && d.src[d.pos] == '.' {
		d.pos++
		float = true
		if digits() == 0 {
			return nil, 4
		}
	}
	if d.pos < len(d.src) && (d.src[d.pos] == 'e' || d.src[d.pos] == 'E') {
		d.pos++
		float = true
		if d.pos < len(d.src) && (d.src[d.pos] == '+' || d.src[d.pos] == '-') {
			d.pos++
		}
		if digits() == 0 {
			return nil, 4
		}
	}
	text := d.src[start:d.pos]
	if !float {
		if n, err := strconv.ParseInt(text, 10, 64); err == nil {
			return n, 0
		}
		if d.bigString {
			return text, 0
		}
	}
	f, _ := strconv.ParseFloat(text, 64)
	return f, 0
}

func (d *jsonDecoder) str() (any, int64) {
	s, code := d.rawString()
	if code != 0 {
		return nil, code
	}
	return s, 0
}

func (d *jsonDecoder) rawString() (string, int64) {
	d.pos++ // the opening quote
	var b strings.Builder
	for d.pos < len(d.src) {
		c := d.src[d.pos]
		switch {
		case c == '"':
			d.pos++
			return b.String(), 0
		case c < 0x20:
			return "", 3
		case c == '\\':
			d.pos++
			if d.pos >= len(d.src) {
				return "", 4
			}
			esc := d.src[d.pos]
			d.pos++
			switch esc {
			case '"', '\\', '/':
				b.WriteByte(esc)
			case 'b':
				b.WriteByte('\b')
			case 'f':
				b.WriteByte('\f')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case 'u':
				r, ok := d.hex4()
				if !ok {
					return "", 4
				}
				if r >= 0xd800 && r < 0xdc00 {
					if strings.HasPrefix(d.src[d.pos:], `\u`) {
						d.pos += 2
						lo, ok := d.hex4()
						if !ok || lo < 0xdc00 || lo > 0xdfff {
							return "", 10
						}
						r = 0x10000 + (r-0xd800)<<10 + (lo - 0xdc00)
					} else {
						return "", 10
					}
				} else if r >= 0xdc00 && r < 0xe000 {
					return "", 10
				}
				b.WriteRune(r)
			default:
				return "", 4
			}
		default:
			b.WriteByte(c)
			d.pos++
		}
	}
	return "", 4
}

func (d *jsonDecoder) hex4() (rune, bool) {
	if d.pos+4 > len(d.src) {
		return 0, false
	}
	n, err := strconv.ParseUint(d.src[d.pos:d.pos+4], 16, 32)
	if err != nil {
		return 0, false
	}
	d.pos += 4
	return rune(n), true
}

func (d *jsonDecoder) list(depth int) (any, int64) {
	if depth > d.maxDep {
		return nil, 1
	}
	d.pos++
	out := model.NewArray()
	d.skip()
	if d.pos < len(d.src) && d.src[d.pos] == ']' {
		d.pos++
		return out, 0
	}
	for {
		d.skip()
		v, code := d.value(depth + 1)
		if code != 0 {
			return nil, code
		}
		out.Append(v)
		d.skip()
		if d.pos >= len(d.src) {
			return nil, 4
		}
		switch d.src[d.pos] {
		case ',':
			d.pos++
		case ']':
			d.pos++
			return out, 0
		default:
			return nil, 4
		}
	}
}

func (d *jsonDecoder) object(depth int) (any, int64) {
	if depth > d.maxDep {
		return nil, 1
	}
	d.pos++
	arr := model.NewArray()
	obj := model.NewStdClass()
	d.skip()
	if d.pos < len(d.src) && d.src[d.pos] == '}' {
		d.pos++
		if d.assoc {
			return arr, 0
		}
		return obj, 0
	}
	for {
		d.skip()
		if d.pos >= len(d.src) || d.src[d.pos] != '"' {
			return nil, 4
		}
		key, code := d.rawString()
		if code != 0 {
			return nil, code
		}
		d.skip()
		if d.pos >= len(d.src) || d.src[d.pos] != ':' {
			return nil, 4
		}
		d.pos++
		d.skip()
		v, code := d.value(depth + 1)
		if code != 0 {
			return nil, code
		}
		if d.assoc {
			arr.Set(phpval.Key(key), v)
		} else {
			obj.SetProp(key, v)
		}
		d.skip()
		if d.pos >= len(d.src) {
			return nil, 4
		}
		switch d.src[d.pos] {
		case ',':
			d.pos++
		case '}':
			d.pos++
			if d.assoc {
				return arr, 0
			}
			return obj, 0
		default:
			return nil, 4
		}
	}
}
