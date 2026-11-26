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
// arrays and objects, in php's text format. A property's key carries the
// visibility php writes into it: a public name as it is, a protected one as
// "\0*\0name" and a private one as "\0Class\0name".
func registerPhpxSerialize(rt *runner.Runtime) {
	rt.RegisterFunc("serialize", func(v any) string {
		var b strings.Builder
		phpxSerialize(rt, &b, v)
		return b.String()
	})
	rt.RegisterFunc("unserialize", func(data string, options ...any) any {
		allowed := phpxAllowedClasses(options)
		v, _, err := phpxUnserialize(rt, data, allowed)
		if err != nil {
			return false
		}
		return v
	})
}

// phpxAllowedClasses reads the allowed_classes option of unserialize(): nil
// means every class, false means none, and an array names the classes allowed.
func phpxAllowedClasses(options []any) map[string]bool {
	if len(options) == 0 {
		return nil
	}
	opts, ok := options[0].(*model.Array)
	if !ok {
		return nil
	}
	v, ok := opts.Get("allowed_classes")
	if !ok {
		return nil
	}
	if b, isBool := v.(bool); isBool {
		if b {
			return nil
		}
		return map[string]bool{}
	}
	set := map[string]bool{}
	if arr, isArr := v.(*model.Array); isArr {
		arr.Range(func(_, name any) bool {
			if s, isStr := name.(string); isStr {
				set[strings.ToLower(strings.TrimPrefix(s, "\\"))] = true
			}
			return true
		})
	}
	return set
}

// phpxAllowed reports whether a class may be built: the option absent allows
// every class, and a leading backslash names the same class as the bare name.
func phpxAllowed(allowed map[string]bool, class string) bool {
	if allowed == nil {
		return true
	}
	return allowed[strings.ToLower(strings.TrimPrefix(class, "\\"))]
}

func phpxSerialize(rt *runner.Runtime, b *strings.Builder, v any) {
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
			phpxSerialize(rt, b, k)
			phpxSerialize(rt, b, val)
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
			phpxSerialize(rt, b, phpxSerializedKey(rt, name, prop))
			phpxSerialize(rt, b, val)
			return true
		})
		b.WriteString("}")
	default:
		b.WriteString("N;")
	}
}

// phpxSerializedKey spells a property name the way php's object format does.
// A hidden shadow slot already holds the mangled key; a declared non-public
// property is mangled now.
func phpxSerializedKey(rt *runner.Runtime, class, name string) string {
	if len(name) > 0 && name[0] == 0 {
		return name
	}
	field, ok := rt.PhpxClassField(class, name)
	if !ok {
		return name
	}
	switch field.Visibility {
	case "private":
		return "\x00" + class + "\x00" + name
	case "protected":
		return "\x00*\x00" + name
	}
	return name
}

func phpxUnserialize(rt *runner.Runtime, s string, allowed map[string]bool) (any, string, error) {
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
		value, rest, err := phpxSerializedString(s)
		return value, rest, err
	case 'a':
		colon := strings.IndexByte(s[2:], ':')
		if colon < 0 {
			return nil, "", fmt.Errorf("bad")
		}
		n, err := strconv.Atoi(s[2 : 2+colon])
		if err != nil {
			return nil, "", fmt.Errorf("bad")
		}
		rest := s[2+colon+2:]
		out := model.NewArraySize(n)
		for i := 0; i < n; i++ {
			var k, v any
			if k, rest, err = phpxUnserialize(rt, rest, allowed); err != nil {
				return nil, "", err
			}
			if v, rest, err = phpxUnserialize(rt, rest, allowed); err != nil {
				return nil, "", err
			}
			out.Set(k, v)
		}
		return out, strings.TrimPrefix(rest, "}"), nil
	case 'O':
		return phpxUnserializeObject(rt, s, allowed)
	}
	return nil, "", fmt.Errorf("unsupported")
}

// phpxSerializedString reads `s:<len>:"<bytes>";`, where len counts bytes and
// the value may hold any byte, a mangled property name's NUL included.
func phpxSerializedString(s string) (any, string, error) {
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
}

// phpxUnserializeObject reads `O:<len>:"<class>":<count>:{...}` and builds the
// instance without running a constructor, the way php does. A class outside
// allowed_classes becomes __PHP_Incomplete_Class.
func phpxUnserializeObject(rt *runner.Runtime, s string, allowed map[string]bool) (any, string, error) {
	colon := strings.IndexByte(s[2:], ':')
	if colon < 0 {
		return nil, "", fmt.Errorf("bad")
	}
	nameLen, err := strconv.Atoi(s[2 : 2+colon])
	if err != nil {
		return nil, "", fmt.Errorf("bad")
	}
	// O:<len>:"<class>":<count>:{...}: the colon IndexByte found ends the
	// length, so the quote that opens the class name is the next byte.
	rest := s[2+colon+1:]
	if len(rest) < nameLen+4 || rest[0] != '"' {
		return nil, "", fmt.Errorf("bad")
	}
	class := rest[1 : 1+nameLen]
	rest = rest[1+nameLen:]
	if !strings.HasPrefix(rest, "\":") {
		return nil, "", fmt.Errorf("bad")
	}
	rest = rest[2:]
	countEnd := strings.IndexByte(rest, ':')
	if countEnd < 0 {
		return nil, "", fmt.Errorf("bad")
	}
	count, err := strconv.Atoi(rest[:countEnd])
	if err != nil || count < 0 {
		return nil, "", fmt.Errorf("bad")
	}
	rest = rest[countEnd+1:]
	if !strings.HasPrefix(rest, "{") {
		return nil, "", fmt.Errorf("bad")
	}
	rest = rest[1:]

	var obj *model.Object
	if phpxAllowed(allowed, class) {
		v, err := rt.PhpxNewWithoutConstructor(class)
		if err != nil {
			return nil, "", err
		}
		obj, _ = v.(*model.Object)
	}
	if obj == nil {
		obj = model.NewObject(&model.Class{Name: "__PHP_Incomplete_Class"})
		obj.SetProp("__PHP_Incomplete_Class_Name", class)
	}
	for i := 0; i < count; i++ {
		var key, value any
		if key, rest, err = phpxUnserialize(rt, rest, allowed); err != nil {
			return nil, "", err
		}
		if value, rest, err = phpxUnserialize(rt, rest, allowed); err != nil {
			return nil, "", err
		}
		name, isStr := key.(string)
		if !isStr || obj.Class.Name == "__PHP_Incomplete_Class" {
			continue
		}
		obj.SetProp(phpxPropertyName(class, name), value)
	}
	return obj, strings.TrimPrefix(rest, "}"), nil
}

// phpxPropertyName strips php's visibility mangling, so a private property is
// stored under the name the runtime reads it by. An ancestor's shadow slot
// keeps its mangled key, which is the slot the runtime keeps it in.
func phpxPropertyName(class, name string) string {
	if len(name) == 0 || name[0] != 0 {
		return name
	}
	parts := strings.SplitN(name, "\x00", 3)
	if len(parts) != 3 {
		return name
	}
	if parts[1] == class || parts[1] == "*" {
		return parts[2]
	}
	return name
}
