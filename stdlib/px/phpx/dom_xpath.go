//go:build phpx

package phpx

import (
	"strconv"
	"strings"
)

// DOMXPath runs the query subset the XML readers of a test runner issue: a
// path of child and descendant steps, an attribute predicate, a position and
// the node test functions. Everything else answers an empty list, which is
// what a reader sees for a part of its document that is absent.

// DOMXPath is `new DOMXPath($document)`.
type DOMXPath struct {
	Document *DOMDocument
	// Namespaces holds the prefixes registerNamespace records. The query
	// subset matches element names as written, so a prefix is a name a
	// caller can already spell; the map keeps what a reader asked for.
	Namespaces map[string]string
}

// RegisterNamespace records a prefix and answers true, as php does.
func (x *DOMXPath) RegisterNamespace(prefix, uri string) bool {
	if x.Namespaces == nil {
		x.Namespaces = map[string]string{}
	}
	x.Namespaces[prefix] = uri
	return true
}

// RegisterPhpFunctions accepts php's function registration. The query subset
// this runtime answers calls no php function, so the registration is kept only
// as an accepted call.
func (x *DOMXPath) RegisterPhpFunctions(_ ...any) {}

func init() {
	// The XML parse reads documents; the class is the one call site.
}

// Query answers the nodes the expression selects.
func (x *DOMXPath) Query(expression string, context ...any) *DOMNodeList {
	out := &DOMNodeList{}
	doc := x.Document
	if doc == nil || doc.root == nil {
		return out
	}
	var start []*domNode
	if len(context) > 0 && context[0] != nil {
		if n := nodeOf(context[0]); n != nil {
			start = []*domNode{n}
		}
	} else if doc.DocumentElement != nil {
		if n := nodeOf(doc.DocumentElement); n != nil {
			start = []*domNode{n}
		}
	}
	seen := map[*domNode]bool{}
	for _, n := range evalXPath(expression, start, doc.root) {
		if seen[n] || n.self == nil {
			continue
		}
		seen[n] = true
		out.items = append(out.items, n.self)
	}
	out.Length = int64(len(out.items))
	return out
}

// Rewind starts a foreach over the list.
func (l *DOMNodeList) Rewind() error { l.pos = 0; return nil }

// Valid reports whether the cursor has a node.
func (l *DOMNodeList) Valid() (bool, error) { return l != nil && l.pos < len(l.items), nil }

// Current answers the node at the cursor.
func (l *DOMNodeList) Current() (any, error) { return l.items[l.pos], nil }

// Key answers the cursor position, as php numbers a list.
func (l *DOMNodeList) Key() (any, error) { return int64(l.pos), nil }

// Next advances the cursor.
func (l *DOMNodeList) Next() error { l.pos++; return nil }

// evalXPath evaluates one expression, union included.
func evalXPath(expr string, start []*domNode, root *domNode) []*domNode {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}
	if parts := splitXPath(expr, '|'); len(parts) > 1 {
		var out []*domNode
		for _, part := range parts {
			out = append(out, evalXPath(part, start, root)...)
		}
		return out
	}
	if strings.HasPrefix(expr, "//") {
		start = descendantsOrSelf([]*domNode{root})
		expr = expr[2:]
	} else if strings.HasPrefix(expr, "/") {
		start = []*domNode{root}
		expr = expr[1:]
	}
	current := start
	for _, step := range splitXPath(expr, '/') {
		step = strings.TrimSpace(step)
		if step == "" {
			current = descendantsOrSelf(current)
			continue
		}
		current = applyXPathStep(current, step)
	}
	return current
}

// applyXPathStep applies one step to every context node.
func applyXPathStep(nodes []*domNode, step string) []*domNode {
	path, predicates := splitXPathPredicates(step)
	var out []*domNode
	for _, n := range nodes {
		out = append(out, stepNodes(n, path)...)
	}
	for _, predicate := range predicates {
		var kept []*domNode
		for i, n := range out {
			if predicateHolds(n, predicate, i+1, len(out)) {
				kept = append(kept, n)
			}
		}
		out = kept
	}
	return out
}

// stepNodes answers the nodes one step selects below a context node.
func stepNodes(n *domNode, step string) []*domNode {
	switch step {
	case ".", "":
		return []*domNode{n}
	case "..":
		if n.parent != nil {
			return []*domNode{n.parent}
		}
		return nil
	case "*":
		var out []*domNode
		for _, child := range n.ChildNodes {
			if c := nodeOf(child); c != nil && c.NodeType == 1 {
				out = append(out, c)
			}
		}
		return out
	case "node()":
		return childNodesOf(n, func(*domNode) bool { return true })
	case "text()":
		return childNodesOf(n, func(c *domNode) bool { return c.NodeType == 3 })
	case "comment()":
		return childNodesOf(n, func(c *domNode) bool { return c.NodeType == 8 })
	}
	if strings.HasPrefix(step, "@") {
		// An attribute is not a node here: no reader iterates one.
		return nil
	}
	return childNodesOf(n, func(c *domNode) bool { return c.NodeType == 1 && c.NodeName == step })
}

func childNodesOf(n *domNode, match func(*domNode) bool) []*domNode {
	var out []*domNode
	for _, child := range n.ChildNodes {
		if c := nodeOf(child); c != nil && match(c) {
			out = append(out, c)
		}
	}
	return out
}

// descendantsOrSelf walks the trees under nodes, each node once.
func descendantsOrSelf(nodes []*domNode) []*domNode {
	var out []*domNode
	var walk func(*domNode)
	walk = func(n *domNode) {
		out = append(out, n)
		for _, child := range n.ChildNodes {
			if c := nodeOf(child); c != nil {
				walk(c)
			}
		}
	}
	for _, n := range nodes {
		walk(n)
	}
	return out
}

// splitXPath splits an expression at sep, ignoring a sep inside brackets or
// quotes.
func splitXPath(expr string, sep byte) []string {
	var parts []string
	depth := 0
	quote := byte(0)
	start := 0
	for i := 0; i < len(expr); i++ {
		c := expr[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '\'' || c == '"':
			quote = c
		case c == '[':
			depth++
		case c == ']':
			if depth > 0 {
				depth--
			}
		case c == sep && depth == 0:
			parts = append(parts, expr[start:i])
			start = i + 1
		}
	}
	parts = append(parts, expr[start:])
	return parts
}

// splitXPathPredicates separates the node test from its predicates.
func splitXPathPredicates(step string) (string, []string) {
	idx := strings.IndexByte(step, '[')
	if idx < 0 {
		return step, nil
	}
	path := strings.TrimSpace(step[:idx])
	var predicates []string
	for i := idx; i < len(step); {
		if step[i] != '[' {
			i++
			continue
		}
		depth := 1
		j := i + 1
		for j < len(step) && depth > 0 {
			switch step[j] {
			case '[':
				depth++
			case ']':
				depth--
			}
			j++
		}
		predicates = append(predicates, strings.TrimSpace(step[i+1:j-1]))
		i = j
	}
	return path, predicates
}

// predicateHolds tests one predicate against a candidate.
func predicateHolds(n *domNode, predicate string, position, size int) bool {
	predicate = strings.TrimSpace(predicate)
	if predicate == "" {
		return true
	}
	if predicate == "last()" {
		return position == size
	}
	if !strings.ContainsAny(predicate, "=<>!()'\"") {
		if i, err := strconv.Atoi(predicate); err == nil {
			return position == i
		}
		return false
	}
	if strings.HasPrefix(predicate, "not(") && strings.HasSuffix(predicate, ")") {
		return !predicateHolds(n, predicate[4:len(predicate)-1], position, size)
	}
	left, op, right := splitXPathComparison(predicate)
	if op == "" {
		return false
	}
	want := xpathLiteral(right)
	var have string
	switch {
	case left == "name()":
		have = n.NodeName
	case left == "text()" || left == ".":
		have = n.TextContent
	case strings.HasPrefix(left, "@"):
		have = n.GetAttribute(left[1:])
	case strings.HasPrefix(left, "php:functionString("):
		// `php:functionString("strtolower", @rel)` is what a reader that
		// selects an attribute case-insensitively writes; embed's document
		// selects its oembed link that way.
		have = xpathFunctionString(left, n)
	default:
		return false
	}
	switch op {
	case "=":
		return have == want
	case "!=":
		return have != want
	}
	return false
}

// xpathFunctionString evaluates `php:functionString("name", argument)` on a
// candidate node: the argument is read the way a comparison reads its left
// side, and the string functions a case-insensitive attribute test uses are
// answered here. Any other function answers the empty string.
func xpathFunctionString(call string, n *domNode) string {
	inner := call[len("php:functionString("):]
	inner = strings.TrimSuffix(strings.TrimSpace(inner), ")")
	name, arg, ok := splitXPathArgs(inner)
	if !ok {
		return ""
	}
	var value string
	switch {
	case arg == "text()" || arg == ".":
		value = n.TextContent
	case strings.HasPrefix(arg, "@"):
		value = n.GetAttribute(arg[1:])
	default:
		value = xpathLiteral(arg)
	}
	switch xpathLiteral(name) {
	case "strtolower", "mb_strtolower":
		return strings.ToLower(value)
	case "strtoupper", "mb_strtoupper":
		return strings.ToUpper(value)
	case "trim":
		return strings.TrimSpace(value)
	case "ltrim":
		return strings.TrimLeft(value, " \t\n\r\x00\x0B")
	case "rtrim":
		return strings.TrimRight(value, " \t\n\r\x00\x0B")
	}
	return ""
}

// splitXPathArgs splits a call's arguments at the comma outside quotes.
func splitXPathArgs(inner string) (string, string, bool) {
	quote := byte(0)
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '\'' || c == '"':
			quote = c
		case c == ',':
			return strings.TrimSpace(inner[:i]), strings.TrimSpace(inner[i+1:]), true
		}
	}
	return "", "", false
}

// splitXPathComparison splits `a op b` outside quotes.
func splitXPathComparison(expr string) (string, string, string) {
	quote := byte(0)
	for i := 0; i < len(expr); i++ {
		c := expr[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			continue
		}
		switch c {
		case '=':
			return strings.TrimSpace(expr[:i]), "=", strings.TrimSpace(expr[i+1:])
		case '!':
			if i+1 < len(expr) && expr[i+1] == '=' {
				return strings.TrimSpace(expr[:i]), "!=", strings.TrimSpace(expr[i+2:])
			}
		}
	}
	return "", "", ""
}

// xpathLiteral answers the string a comparison names, unquoting it.
func xpathLiteral(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && (s[0] == '\'' || s[0] == '"') && s[len(s)-1] == s[0] {
		return s[1 : len(s)-1]
	}
	return s
}
