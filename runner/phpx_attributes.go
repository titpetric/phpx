//go:build phpx

package runner

import (
	"io/fs"
	"strings"
	"sync"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/parser"
)

// phpxDeclMeta is what sits above a declaration in the source: its doc comment
// and the attribute groups, as written.
type phpxDeclMeta struct {
	Doc    string
	Groups []string // the text between #[ and ], one entry per group
}

// phpxSourceMeta is a file read once for the metadata Reflection exposes. The
// lexer drops comments and attributes, so this scans the text itself. Keys are
// "Class", "Class::method" and "Class::$property".
type phpxSourceMeta struct {
	Namespace string
	Uses      map[string]string
	Decls     map[string]phpxDeclMeta
}

var phpxMetaCache sync.Map // file -> *phpxSourceMeta

// PhpxSourceMeta reads and scans file, which is a path as a declaration
// recorded it. It answers nil for a file the runtime cannot read.
func (rt *Runtime) PhpxSourceMeta(file string) *phpxSourceMeta {
	if file == "" || strings.HasPrefix(file, "phpx:") || rt.opts.RootFS == nil {
		return nil
	}
	if v, ok := phpxMetaCache.Load(file); ok {
		return v.(*phpxSourceMeta)
	}
	b, err := fs.ReadFile(rt.opts.RootFS, rt.resolveFSPath(file))
	if err != nil {
		return nil
	}
	meta := phpxScanSource(string(b))
	phpxMetaCache.Store(file, meta)
	return meta
}

// phpxScanSource finds each class-like, method and property declaration and
// the doc comment and attributes written ahead of it.
func phpxScanSource(src string) *phpxSourceMeta {
	meta := &phpxSourceMeta{Uses: map[string]string{}, Decls: map[string]phpxDeclMeta{}}
	var (
		doc, class string
		groups     []string
		depth      int
		classDepth = -1
		i          = 0
	)
	n := len(src)
	skipString := func(q byte) {
		i++
		for i < n && src[i] != q {
			if src[i] == '\\' {
				i++
			}
			i++
		}
		i++
	}
	for i < n {
		c := src[i]
		switch {
		case c == '/' && i+1 < n && src[i+1] == '/' || c == '#' && !(i+1 < n && src[i+1] == '['):
			for i < n && src[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < n && src[i+1] == '*':
			end := strings.Index(src[i+2:], "*/")
			if end < 0 {
				i = n
				break
			}
			text := src[i : i+2+end+2]
			if strings.HasPrefix(text, "/**") {
				doc = text
			}
			i += 2 + end + 2
		case c == '#' && i+1 < n && src[i+1] == '[':
			start := i + 2
			level := 1
			i += 2
			for i < n && level > 0 {
				switch src[i] {
				case '[':
					level++
				case ']':
					level--
				case '\'', '"':
					skipString(src[i])
					continue
				}
				i++
			}
			groups = append(groups, src[start:i-1])
		case c == '\'' || c == '"':
			skipString(c)
		case c == '<' && strings.HasPrefix(src[i:], "<<<"):
			i = phpxSkipHeredoc(src, i)
		case c == '{':
			depth++
			i++
			groups, doc = nil, ""
		case c == '}':
			depth--
			if classDepth >= 0 && depth <= classDepth {
				class, classDepth = "", -1
			}
			i++
			groups, doc = nil, ""
		case c == ';':
			i++
			groups, doc = nil, ""
		case c == '$' && class != "" && depth == classDepth+1 && i+1 < n && phpxIdentByte(src[i+1]):
			j := i + 1
			for j < n && phpxIdentByte(src[j]) {
				j++
			}
			if len(groups) > 0 || doc != "" {
				meta.Decls[class+"::$"+src[i+1:j]] = phpxDeclMeta{Doc: doc, Groups: groups}
			}
			groups, doc = nil, ""
			i = j
		case phpxIdentByte(c):
			j := i
			for j < n && (phpxIdentByte(src[j]) || src[j] == '\\') {
				j++
			}
			word := src[i:j]
			i = j
			switch strings.ToLower(word) {
			case "namespace":
				if depth == 0 {
					meta.Namespace = phpxNextName(src, &i)
					groups, doc = nil, ""
				}
			case "use":
				if depth == 0 {
					phpxScanUse(src, &i, meta)
					groups, doc = nil, ""
				}
			case "class", "interface", "trait", "enum":
				if prev := strings.TrimSpace(src[max(0, i-len(word)-3) : i-len(word)]); strings.HasSuffix(prev, "::") || strings.HasSuffix(prev, "->") {
					break
				}
				name := phpxNextName(src, &i)
				if name == "" || depth != 0 {
					break
				}
				class, classDepth = name, depth
				if len(groups) > 0 || doc != "" {
					meta.Decls[class] = phpxDeclMeta{Doc: doc, Groups: groups}
				}
				groups, doc = nil, ""
			case "function":
				if class == "" || depth != classDepth+1 {
					break
				}
				for i < n && (src[i] == ' ' || src[i] == '\t' || src[i] == '&' || src[i] == '\n') {
					i++
				}
				name := phpxNextName(src, &i)
				if name != "" && (len(groups) > 0 || doc != "") {
					meta.Decls[class+"::"+name] = phpxDeclMeta{Doc: doc, Groups: groups}
				}
				groups, doc = nil, ""
			}
		default:
			i++
		}
	}
	return meta
}

func phpxIdentByte(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c >= 0x80
}

// phpxNextName reads the next name after whitespace and advances past it.
func phpxNextName(src string, i *int) string {
	for *i < len(src) && (src[*i] == ' ' || src[*i] == '\t' || src[*i] == '\n' || src[*i] == '\r') {
		*i++
	}
	j := *i
	for j < len(src) && (phpxIdentByte(src[j]) || src[j] == '\\') {
		j++
	}
	name := src[*i:j]
	*i = j
	return name
}

// phpxScanUse records the class imports of one `use` statement.
func phpxScanUse(src string, i *int, meta *phpxSourceMeta) {
	end := strings.IndexByte(src[*i:], ';')
	if end < 0 {
		return
	}
	stmt := strings.TrimSpace(src[*i : *i+end])
	*i += end + 1
	if strings.HasPrefix(stmt, "function ") || strings.HasPrefix(stmt, "const ") {
		return
	}
	prefix, group, isGroup := strings.Cut(stmt, "{")
	if isGroup {
		for _, part := range strings.Split(strings.TrimSuffix(strings.TrimSpace(group), "}"), ",") {
			phpxAddUse(meta, strings.TrimSpace(prefix)+strings.TrimSpace(part))
		}
		return
	}
	for _, part := range strings.Split(stmt, ",") {
		phpxAddUse(meta, strings.TrimSpace(part))
	}
}

func phpxAddUse(meta *phpxSourceMeta, entry string) {
	entry = strings.TrimPrefix(strings.TrimSpace(entry), "\\")
	alias := entry
	if lower := strings.ToLower(entry); strings.Contains(lower, " as ") {
		at := strings.Index(lower, " as ")
		alias = strings.TrimSpace(entry[at+4:])
		entry = strings.TrimSpace(entry[:at])
	} else if i := strings.LastIndexByte(entry, '\\'); i >= 0 {
		alias = entry[i+1:]
	}
	meta.Uses[alias] = entry
}

// resolve turns a name written in the file into the class it names.
func (m *phpxSourceMeta) resolve(name string) string {
	if strings.HasPrefix(name, "\\") {
		return name[1:]
	}
	head, rest, _ := strings.Cut(name, "\\")
	if full, ok := m.Uses[head]; ok {
		if rest == "" {
			return full
		}
		return full + "\\" + rest
	}
	if m.Namespace != "" {
		return m.Namespace + "\\" + name
	}
	return name
}

// PhpxAttributes lists the attributes and the doc comment above the member key
// ("Class", "Class::method", "Class::$property") of a class declared in file,
// as [[name, "args source"], ...] and the doc text or false.
func (rt *Runtime) PhpxAttributes(file, key string) (attrs *model.Array, doc any) {
	attrs, doc = model.NewArray(), false
	meta := rt.PhpxSourceMeta(file)
	if meta == nil {
		return attrs, doc
	}
	short := key
	if i := strings.LastIndexByte(key, '\\'); i >= 0 {
		short = key[i+1:]
	}
	decl, ok := meta.Decls[short]
	if !ok {
		return attrs, doc
	}
	if decl.Doc != "" {
		doc = decl.Doc
	}
	for _, group := range decl.Groups {
		for _, one := range phpxSplitTop(group) {
			one = strings.TrimSpace(one)
			if one == "" {
				continue
			}
			name, args := one, ""
			if p := strings.IndexByte(one, '('); p >= 0 {
				name, args = strings.TrimSpace(one[:p]), strings.TrimSuffix(strings.TrimSpace(one[p+1:]), ")")
			}
			pair := model.NewArray()
			pair.Append(meta.resolve(name))
			pair.Append(args)
			attrs.Append(pair)
		}
	}
	return attrs, doc
}

// phpxSplitTop splits s at the commas that sit outside brackets and quotes.
func phpxSplitTop(s string) []string {
	var parts []string
	level, start := 0, 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(', '[', '{':
			level++
		case ')', ']', '}':
			level--
		case '\'', '"':
			q := s[i]
			i++
			for i < len(s) && s[i] != q {
				if s[i] == '\\' {
					i++
				}
				i++
			}
		case ',':
			if level == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	return append(parts, s[start:])
}

// PhpxAttributeArguments evaluates the argument list of an attribute, written
// in file and in class, as the array of its positional values.
func (rt *Runtime) PhpxAttributeArguments(file, class, args string) (any, error) {
	out := model.NewArray()
	if strings.TrimSpace(args) == "" {
		return out, nil
	}
	meta := rt.PhpxSourceMeta(file)
	var head strings.Builder
	head.WriteString("<?php ")
	if meta != nil && meta.Namespace != "" {
		head.WriteString("namespace " + meta.Namespace + "; ")
	}
	if meta != nil {
		for alias, full := range meta.Uses {
			head.WriteString("use " + full + " as " + alias + "; ")
		}
	}
	prog, err := parser.ParseFile("attribute", head.String()+"return ["+phpxNamedToKeys(args)+"];")
	if err != nil {
		return out, nil
	}
	scope := rt.newScope()
	if class != "" {
		scope.Set("__class__", rt.boxed(class))
	}
	for _, st := range prog.Stmts {
		if ret, ok := st.(*model.Return); ok && ret.Value != nil {
			return rt.Eval(ret.Value, scope)
		}
	}
	return out, nil
}

// phpxNamedToKeys rewrites an argument list into array items: `a, name: b`
// becomes `a, 'name' => b`, so the named arguments of an attribute keep their
// names in the array getArguments() answers.
func phpxNamedToKeys(args string) string {
	parts := phpxSplitTop(args)
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		j := 0
		for j < len(trimmed) && phpxIdentByte(trimmed[j]) {
			j++
		}
		if j > 0 && j < len(trimmed) && !(trimmed[0] >= '0' && trimmed[0] <= '9') {
			rest := strings.TrimLeft(trimmed[j:], " \t")
			if strings.HasPrefix(rest, ":") && !strings.HasPrefix(rest, "::") {
				parts[i] = "'" + trimmed[:j] + "' => " + strings.TrimSpace(rest[1:])
			}
		}
	}
	return strings.Join(parts, ",")
}

// phpxSkipHeredoc answers the offset just past the heredoc or nowdoc that opens
// at i, so the braces and hashes written inside it are text.
func phpxSkipHeredoc(src string, i int) int {
	j := i + 3
	for j < len(src) && (src[j] == ' ' || src[j] == '\t') {
		j++
	}
	if j < len(src) && (src[j] == '\'' || src[j] == '"') {
		j++
	}
	start := j
	for j < len(src) && phpxIdentByte(src[j]) {
		j++
	}
	label := src[start:j]
	if label == "" {
		return i + 3
	}
	for {
		nl := strings.IndexByte(src[j:], '\n')
		if nl < 0 {
			return len(src)
		}
		j += nl + 1
		line := strings.TrimLeft(src[j:], " \t")
		if strings.HasPrefix(line, label) && (len(line) == len(label) || !phpxIdentByte(line[len(label)])) {
			return j + (len(src[j:]) - len(line)) + len(label)
		}
	}
}
