package astcodec

import (
	"math"

	"github.com/titpetric/phpx/model"
)

// decoder reads a payload into the string table and a program. A read past
// the end, an unknown tag and an index out of range end in ErrFormat.
type decoder struct {
	src   string
	pos   int
	table []string
	err   error
	anon  []*model.ClassDecl
	spans map[model.Stmt]model.SourceSpan
	lits  chunk[model.Lit]
	cmps  chunk[model.Echo]
	conds chunk[model.If]
	loops chunk[model.Foreach]
	trows chunk[model.Throw]
	unary chunk[model.Unary]
	refs  bool
	tern  chunk[model.Ternary]
	coal  chunk[model.Coalesce]
	casts chunk[model.Cast]
	news  chunk[model.New]
	paren chunk[model.Parenthesized]
	inter chunk[model.Interp]
	ccons chunk[model.ClassConst]
	scall chunk[model.StaticCall]
	sprop chunk[model.StaticProp]
	vars  chunk[model.Var]
	calls chunk[model.Call]
	bins  chunk[model.Binary]
	idxs  chunk[model.Index]
	props chunk[model.PropAccess]
	meths chunk[model.MethodCall]
	sets  chunk[model.Assign]
	short chunk[model.ExprStmt]
	rets  chunk[model.Return]
	arrs  chunk[model.ArrayLit]
	invs  chunk[model.Invoke]
}

// Decode reads a payload Encode wrote as a program.
func Decode(data []byte) (*model.Program, error) {
	return DecodeString(string(data))
}

// DecodeString reads a payload held as a string. Every name the payload
// carries is a slice of that string, so the decode copies none of them; the
// caller keeps the string alive as long as the program.
func DecodeString(src string) (*model.Program, error) {
	d := &decoder{src: src}
	// A new name takes about forty payload bytes. Sizing the table once keeps
	// it from doubling its way to a program; the estimate is a hint, and the
	// append grows past it when a file spells more names than that.
	d.table = make([]string, 0, len(src)/40+8)
	// The parser records a span for about one statement in twenty-five bytes,
	// and a map that starts there skips the rehash cycle.
	d.spans = make(map[model.Stmt]model.SourceSpan, len(src)/25+8)
	p := &model.Program{}
	// The anonymous classes come first: a New in a statement refers to one by
	// its index in that list.
	p.AnonClasses = d.classes()
	d.anon = p.AnonClasses
	p.Stmts = d.stmtList()
	p.Namespace = d.str()
	p.NamespaceLine = d.int()
	p.SourceSpans = d.spans
	if d.err != nil {
		return nil, d.err
	}
	// A payload that declares a by-reference parameter turns the runtime's
	// reference paths on, as the parser does when it reads one.
	if d.refs {
		model.PhpxAnyByRef.Store(true)
	}
	return p, nil
}

func (d *decoder) fail() {
	if d.err == nil {
		d.err = ErrFormat
	}
}

func (d *decoder) byte() byte {
	if d.err != nil {
		return 0
	}
	if d.pos >= len(d.src) {
		d.fail()
		return 0
	}
	v := d.src[d.pos]
	d.pos++
	return v
}

func (d *decoder) uvarint() uint64 {
	if d.err != nil {
		return 0
	}
	var v uint64
	for shift := uint(0); ; shift += 7 {
		if d.pos >= len(d.src) || shift > 63 {
			d.fail()
			return 0
		}
		b := d.src[d.pos]
		d.pos++
		if b < 0x80 {
			return v | uint64(b)<<shift
		}
		v |= uint64(b&0x7f) << shift
	}
}

func (d *decoder) int() int {
	v := d.uvarint()
	return int(v>>1) ^ -int(v&1)
}

func (d *decoder) str() string {
	if d.err != nil {
		return ""
	}
	v := d.uvarint()
	if v&1 == 1 {
		id := v >> 1
		if id >= uint64(len(d.table)) {
			d.fail()
			return ""
		}
		return d.table[id]
	}
	n := v >> 1
	if n > uint64(len(d.src)-d.pos) {
		d.fail()
		return ""
	}
	s := d.src[d.pos : d.pos+int(n)]
	d.pos += int(n)
	d.table = append(d.table, s)
	return s
}

func (d *decoder) boolean() bool { return d.byte() != 0 }

func (d *decoder) value() any {
	switch valueTag(d.byte()) {
	case valNil:
		return nil
	case valBool:
		return d.boolean()
	case valInt:
		return d.int()
	case valInt64:
		v := d.uvarint()
		return int64(v>>1) ^ -int64(v&1)
	case valFloat:
		return math.Float64frombits(d.uvarint())
	case valString:
		return d.str()
	default:
		d.fail()
		return nil
	}
}

// span reads the source span a statement node carries, if any.
func (d *decoder) span(n any, t tag) {
	if !spanTag[t] {
		return
	}
	if d.byte() == 0 {
		return
	}
	start, end := d.int(), d.int()
	if stmt, ok := n.(model.Stmt); ok {
		if d.spans == nil {
			d.spans = make(map[model.Stmt]model.SourceSpan)
		}
		d.spans[stmt] = model.SourceSpan{Start: start, End: end}
	}
}

func (d *decoder) node() any {
	t := d.byte()
	if d.err != nil {
		return nil
	}
	n := d.body(tag(t))
	d.span(n, tag(t))
	return n
}

func (d *decoder) body(t tag) any {
	if t == tagAnonRef {
		idx := d.uvarint()
		if d.err != nil || idx >= uint64(len(d.anon)) {
			d.fail()
			return nil
		}
		return d.anon[idx]
	}
	switch t {
	case tagNil:
		return nil
	case tagInlineHTML:
		x := &model.InlineHTML{}
		x.Text = d.str()
		return x
	case tagEcho:
		x := d.cmps.new()
		x.Args = d.exprList()
		return x
	case tagExprStmt:
		x := d.short.new()
		x.X = d.expr()
		return x
	case tagAssign:
		x := d.sets.new()
		x.Target = d.expr()
		x.Op = d.str()
		x.Value = d.expr()
		return x
	case tagIf:
		x := d.conds.new()
		x.Cond = d.expr()
		x.Then = d.stmtList()
		x.Else = d.stmtList()
		x.ElseLine = d.int()
		return x
	case tagForeach:
		x := d.loops.new()
		x.Source = d.expr()
		x.KeyTarget = d.expr()
		x.ValTarget = d.expr()
		x.ByRef = d.boolean()
		x.KeyVar = d.str()
		x.ValVar = d.str()
		x.Body = d.stmtList()
		return x
	case tagFor:
		x := &model.For{}
		x.Init = d.stmt()
		x.Cond = d.expr()
		x.Post = d.stmt()
		x.Body = d.stmtList()
		return x
	case tagDoWhile:
		x := &model.DoWhile{}
		x.Body = d.stmtList()
		x.Cond = d.expr()
		return x
	case tagReturn:
		x := d.rets.new()
		x.Value = d.expr()
		return x
	case tagInclude:
		x := &model.Include{}
		x.Path = d.expr()
		x.Keyword = d.str()
		x.Once = d.boolean()
		x.Parenthesized = d.boolean()
		return x
	case tagFuncDecl:
		x := &model.FuncDecl{}
		d.funcDecl(x)
		return x
	case tagClassDecl:
		x := &model.ClassDecl{}
		d.classDecl(x)
		return x
	case tagTraitDecl:
		x := &model.TraitDecl{}
		x.Name = d.str()
		x.Traits = d.strList()
		x.TraitRules = d.traitRules()
		x.Fields = d.fields()
		x.Statics = d.fields()
		x.Consts = d.fields()
		x.Methods = d.funcDecls()
		return x
	case tagInterfaceDecl:
		x := &model.InterfaceDecl{}
		x.Name = d.str()
		x.Extends = d.strList()
		x.Consts = d.fields()
		x.Methods = d.funcDecls()
		return x
	case tagUse:
		x := &model.Use{}
		x.Kind = d.str()
		n, isNil := d.count()
		if !isNil {
			x.Imports = make([]model.UseImport, n)
		}
		for i := range x.Imports {
			x.Imports[i] = model.UseImport{Path: d.str(), Alias: d.str(), Package: d.str()}
		}
		return x
	case tagDeclare:
		x := &model.Declare{}
		n, isNil := d.count()
		if !isNil {
			x.Directives = make([]model.DeclareDirective, n)
		}
		for i := range x.Directives {
			x.Directives[i] = model.DeclareDirective{Name: d.str(), Value: d.expr()}
		}
		x.Body = d.stmtList()
		x.Block = d.boolean()
		return x
	case tagUnset:
		x := &model.Unset{}
		x.Targets = d.exprList()
		return x
	case tagStaticVar:
		x := &model.StaticVar{}
		n, isNil := d.count()
		if !isNil {
			x.Vars = make([]model.StaticVarDecl, n)
		}
		for i := range x.Vars {
			x.Vars[i] = model.StaticVarDecl{Name: d.str(), Default: d.expr()}
		}
		return x
	case tagGlobal:
		x := &model.Global{}
		x.Names = d.strList()
		return x
	case tagConstDecl:
		x := &model.ConstDecl{}
		x.Consts = d.fields()
		return x
	case tagThrow:
		x := d.trows.new()
		x.X = d.expr()
		return x
	case tagTry:
		x := &model.Try{}
		x.Body = d.stmtList()
		n, isNil := d.count()
		if !isNil {
			x.Catches = make([]model.Catch, n)
		}
		for i := range x.Catches {
			x.Catches[i] = model.Catch{Type: d.str(), Var: d.str(), Body: d.stmtList(), Line: d.int()}
		}
		x.Finally = d.stmtList()
		x.FinallyLine = d.int()
		return x
	case tagSwitch:
		x := &model.Switch{}
		x.Cond = d.expr()
		n, isNil := d.count()
		if !isNil {
			x.Cases = make([]model.SwitchCase, n)
		}
		for i := range x.Cases {
			x.Cases[i] = model.SwitchCase{Value: d.expr(), Body: d.stmtList(), Line: d.int()}
		}
		x.Default = d.stmtList()
		return x
	case tagBreak:
		x := &model.Break{}
		x.Line = d.int()
		x.Levels = d.int()
		return x
	case tagContinue:
		x := &model.Continue{}
		x.Line = d.int()
		x.Levels = d.int()
		return x
	case tagLabel:
		x := &model.Label{}
		x.Name = d.str()
		x.Line = d.int()
		return x
	case tagGoto:
		x := &model.Goto{}
		x.Label = d.str()
		x.Line = d.int()
		return x
	case tagLit:
		x := d.lits.new()
		x.Value = d.value()
		x.Raw = d.str()
		return x
	case tagInterp:
		x := d.inter.new()
		x.Parts = d.exprList()
		x.Raw = d.str()
		return x
	case tagVar:
		x := d.vars.new()
		x.Name = d.str()
		x.Const = d.boolean()
		return x
	case tagArrayLit:
		x := d.arrs.new()
		x.Items = d.items()
		return x
	case tagGoLit:
		x := &model.GoLit{}
		x.Map = d.boolean()
		x.KeyType = d.str()
		x.ValType = d.str()
		x.Items = d.items()
		return x
	case tagAddrOf:
		x := &model.AddrOf{}
		x.X = d.expr()
		return x
	case tagDeref:
		x := &model.Deref{}
		x.X = d.expr()
		return x
	case tagIndex:
		x := d.idxs.new()
		x.Base = d.expr()
		x.Index = d.expr()
		return x
	case tagPropAccess:
		x := d.props.new()
		x.Base = d.expr()
		x.Name = d.str()
		x.NameExpr = d.expr()
		return x
	case tagCall:
		x := d.calls.new()
		x.Name = d.str()
		x.Fallback = d.str()
		x.Args = d.exprList()
		x.Bare = d.boolean()
		x.FirstClass = d.boolean()
		return x
	case tagMethodCall:
		x := d.meths.new()
		x.Base = d.expr()
		x.Method = d.str()
		x.MethodExpr = d.expr()
		x.Args = d.exprList()
		x.FirstClass = d.boolean()
		return x
	case tagNew:
		x := d.news.new()
		x.Class = d.str()
		x.ClassExpr = d.expr()
		x.Args = d.exprList()
		if decl := d.node(); decl != nil {
			x.Decl, _ = decl.(*model.ClassDecl)
		}
		return x
	case tagRef:
		x := &model.Ref{}
		x.X = d.expr()
		return x
	case tagUnary:
		x := d.unary.new()
		x.Op = d.str()
		x.X = d.expr()
		x.Postfix = d.boolean()
		return x
	case tagParenthesized:
		x := d.paren.new()
		x.X = d.expr()
		return x
	case tagBinary:
		x := d.bins.new()
		x.Op = d.str()
		x.Left = d.expr()
		x.Right = d.expr()
		return x
	case tagTernary:
		x := d.tern.new()
		x.Cond = d.expr()
		x.Then = d.expr()
		x.Else = d.expr()
		return x
	case tagCoalesce:
		x := d.coal.new()
		x.Left = d.expr()
		x.Right = d.expr()
		x.Assign = d.boolean()
		return x
	case tagClassConst:
		x := d.ccons.new()
		x.Class = d.str()
		x.Name = d.str()
		return x
	case tagCast:
		x := d.casts.new()
		x.Type = d.str()
		x.X = d.expr()
		return x
	case tagClosure:
		x := &model.Closure{}
		x.Params = d.params()
		n, isNil := d.count()
		if !isNil {
			x.Uses = make([]model.ClosureUse, n)
		}
		for i := range x.Uses {
			x.Uses[i] = model.ClosureUse{Name: d.str(), ByRef: d.boolean()}
		}
		x.Body = d.stmtList()
		x.ReturnType = d.str()
		x.Static = d.boolean()
		x.ByRef = d.boolean()
		return x
	case tagAssignExpr:
		x := &model.AssignExpr{}
		x.Target = d.expr()
		x.Op = d.str()
		x.Value = d.expr()
		x.Line = d.int()
		return x
	case tagListExpr:
		x := &model.ListExpr{}
		x.Elems = d.exprList()
		return x
	case tagStaticCall:
		x := d.scall.new()
		x.Class = d.str()
		x.Method = d.str()
		x.MethodExpr = d.expr()
		x.Args = d.exprList()
		x.FirstClass = d.boolean()
		return x
	case tagStaticProp:
		x := d.sprop.new()
		x.Class = d.str()
		x.Name = d.str()
		x.NameExpr = d.expr()
		return x
	case tagInvoke:
		x := d.invs.new()
		x.Callee = d.expr()
		x.Args = d.exprList()
		x.FirstClass = d.boolean()
		return x
	default:
		d.fail()
		return nil
	}
}

func (d *decoder) expr() model.Expr {
	x, _ := d.node().(model.Expr)
	return x
}

func (d *decoder) stmt() model.Stmt {
	s, _ := d.node().(model.Stmt)
	return s
}

// count reads a slice length, and whether the payload wrote a nil slice.
func (d *decoder) count() (int, bool) {
	v := d.uvarint()
	if d.err != nil || v > uint64(len(d.src))+1 {
		d.fail()
		return 0, true
	}
	if v == 0 {
		return 0, true
	}
	return int(v - 1), false
}

func (d *decoder) exprList() []model.Expr {
	n, isNil := d.count()
	if isNil {
		return nil
	}
	out := make([]model.Expr, n)
	for i := range out {
		out[i] = d.expr()
	}
	return out
}

func (d *decoder) stmtList() []model.Stmt {
	n, isNil := d.count()
	if isNil {
		return nil
	}
	out := make([]model.Stmt, n)
	for i := range out {
		out[i] = d.stmt()
	}
	return out
}

func (d *decoder) strList() []string {
	n, isNil := d.count()
	if isNil {
		return nil
	}
	out := make([]string, n)
	for i := range out {
		out[i] = d.str()
	}
	return out
}

func (d *decoder) items() []model.ArrayItem {
	n, isNil := d.count()
	if isNil {
		return nil
	}
	out := make([]model.ArrayItem, n)
	for i := range out {
		out[i] = model.ArrayItem{Key: d.expr(), Val: d.expr(), Ref: d.boolean()}
	}
	return out
}

func (d *decoder) params() []model.Param {
	n, isNil := d.count()
	if isNil {
		return nil
	}
	out := make([]model.Param, n)
	for i := range out {
		out[i] = model.Param{
			Name:      d.str(),
			Default:   d.expr(),
			Modifiers: d.str(),
			Type:      d.str(),
			Resolved:  d.str(),
			ByRef:     d.boolean(),
			Variadic:  d.boolean(),
		}
		d.refs = d.refs || out[i].ByRef
	}
	return out
}

func (d *decoder) fields() []model.Field {
	n, isNil := d.count()
	if isNil {
		return nil
	}
	out := make([]model.Field, n)
	for i := range out {
		out[i] = model.Field{
			Name:       d.str(),
			Default:    d.expr(),
			Visibility: d.str(),
			Readonly:   d.boolean(),
			Type:       d.str(),
			Resolved:   d.str(),
		}
		out[i].Span = model.SourceSpan{Start: d.int(), End: d.int()}
	}
	return out
}

func (d *decoder) traitRules() []model.TraitRule {
	n, isNil := d.count()
	if isNil {
		return nil
	}
	out := make([]model.TraitRule, n)
	for i := range out {
		out[i] = model.TraitRule{
			Trait:      d.str(),
			Method:     d.str(),
			Insteadof:  d.strList(),
			Alias:      d.str(),
			Visibility: d.str(),
		}
	}
	return out
}

func (d *decoder) funcDecl(f *model.FuncDecl) {
	f.Class = d.str()
	f.Name = d.str()
	f.Filename = d.str()
	f.Params = d.params()
	f.Body = d.stmtList()
	f.Visibility = d.str()
	f.ReturnType = d.str()
	f.ResolvedReturn = d.str()
	f.Static = d.boolean()
	f.Abstract = d.boolean()
	f.ByRef = d.boolean()
}

func (d *decoder) funcDecls() []*model.FuncDecl {
	n, isNil := d.count()
	if isNil {
		return nil
	}
	out := make([]*model.FuncDecl, n)
	for i := range out {
		if decl, ok := d.node().(*model.FuncDecl); ok {
			out[i] = decl
		}
	}
	return out
}

func (d *decoder) classDecl(c *model.ClassDecl) {
	c.Name = d.str()
	c.Filename = d.str()
	c.Parent = d.str()
	c.Implements = d.strList()
	c.Traits = d.strList()
	c.TraitRules = d.traitRules()
	c.Abstract = d.boolean()
	c.Final = d.boolean()
	c.Readonly = d.boolean()
	c.Fields = d.fields()
	c.Statics = d.fields()
	c.Consts = d.fields()
	c.Methods = d.funcDecls()
}

func (d *decoder) classes() []*model.ClassDecl {
	n, isNil := d.count()
	if isNil {
		return nil
	}
	out := make([]*model.ClassDecl, n)
	for i := range out {
		if decl, ok := d.node().(*model.ClassDecl); ok {
			out[i] = decl
		}
	}
	return out
}
