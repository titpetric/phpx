//go:build phpx

package phpx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Serialize`, registerPhpxSerialize)
}

// registerPhpxSerialize installs serialize() and unserialize() for scalars,
// arrays and the public properties of objects, in php's text format.
func registerPhpxSerialize(rt *runner.Runtime) {
	rt.RegisterFunc("serialize", func(v any) string {
		var b strings.Builder
		phpxSerialize(&b, v)
		return b.String()
	})
	rt.RegisterFunc("unserialize", func(data string, _ ...any) any {
		v, _, err := phpxUnserialize(data)
		if err != nil {
			return false
		}
		return v
	})
}

func phpxSerialize(b *strings.Builder, v any) {
	switch t := v.(type) {
	case nil:
		b.WriteString("N;")
	case bool:
		if t {
			b.WriteString("b:1;")
		} else {
			b.WriteString("b:0;")
		}
	case int64:
		fmt.Fprintf(b, "i:%d;", t)
	case int:
		fmt.Fprintf(b, "i:%d;", t)
	case float64:
		b.WriteString("d:" + strconv.FormatFloat(t, 'g', 17, 64) + ";")
	case string:
		fmt.Fprintf(b, "s:%d:\"%s\";", len(t), t)
	case *model.Array:
		fmt.Fprintf(b, "a:%d:{", t.Len())
		t.Range(func(k, val any) bool {
			phpxSerialize(b, k)
			phpxSerialize(b, val)
			return true
		})
		b.WriteString("}")
	case *model.Object:
		name := ""
		if t.Class != nil {
			name = t.Class.Name
		}
		fmt.Fprintf(b, "O:%d:\"%s\":%d:{", len(name), name, t.Len())
		t.Range(func(prop string, val any) bool {
			phpxSerialize(b, prop)
			phpxSerialize(b, val)
			return true
		})
		b.WriteString("}")
	default:
		b.WriteString("N;")
	}
}

func phpxUnserialize(s string) (any, string, error) {
	if s == "" {
		return nil, "", fmt.Errorf("empty")
	}
	switch s[0] {
	case 'N':
		return nil, strings.TrimPrefix(s[1:], ";"), nil
	case 'b', 'i', 'd':
		end := strings.IndexByte(s, ';')
		if end < 0 || len(s) < 3 {
			return nil, "", fmt.Errorf("bad")
		}
		body := s[2:end]
		rest := s[end+1:]
		switch s[0] {
		case 'b':
			return body == "1", rest, nil
		case 'i':
			n, err := strconv.ParseInt(body, 10, 64)
			return n, rest, err
		default:
			f, err := strconv.ParseFloat(body, 64)
			return f, rest, err
		}
	case 's':
		colon := strings.IndexByte(s[2:], ':')
		if colon < 0 {
			return nil, "", fmt.Errorf("bad")
		}
		n, err := strconv.Atoi(s[2 : 2+colon])
		start := 2 + colon + 2
		if err != nil || start+n+2 > len(s) {
			return nil, "", fmt.Errorf("bad")
		}
		return s[start : start+n], s[start+n+2:], nil
	case 'a':
		colon := strings.IndexByte(s[2:], ':')
		n, err := strconv.Atoi(s[2 : 2+colon])
		if colon < 0 || err != nil {
			return nil, "", fmt.Errorf("bad")
		}
		rest := s[2+colon+2:]
		out := model.NewArraySize(n)
		for i := 0; i < n; i++ {
			var k, v any
			if k, rest, err = phpxUnserialize(rest); err != nil {
				return nil, "", err
			}
			if v, rest, err = phpxUnserialize(rest); err != nil {
				return nil, "", err
			}
			out.Set(k, v)
		}
		return out, strings.TrimPrefix(rest, "}"), nil
	}
	return nil, "", fmt.Errorf("unsupported")
}
