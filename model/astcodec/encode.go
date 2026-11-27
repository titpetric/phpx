package astcodec

import (
	"math"

	"github.com/titpetric/phpx/model"
)

// encoder writes the string table and the payload in one pass, and a name
// written a hundred times costs one entry.
type encoder struct {
	buf   []byte
	table map[string]uint32
	anon  map[*model.ClassDecl]uint32
	spans map[model.Stmt]model.SourceSpan
	err   error
}

// Encode writes p as a payload Decode reads back.
func Encode(p *model.Program) ([]byte, error) {
	if p == nil {
		return nil, ErrFormat
	}
	e := &encoder{table: make(map[string]uint32), spans: p.SourceSpans}
	e.anon = make(map[*model.ClassDecl]uint32, len(p.AnonClasses))
	for i, decl := range p.AnonClasses {
		if decl != nil {
			e.anon[decl] = uint32(i)
		}
	}
	e.classes(p.AnonClasses)
	e.stmtList(p.Stmts)
	e.str(p.Namespace)
	e.int(p.NamespaceLine)
	if e.err != nil {
		return nil, e.err
	}
	return e.buf, nil
}

func (e *encoder) byte(v byte) { e.buf = append(e.buf, v) }

// tag writes one node tag.
func (e *encoder) tag(t tag) { e.buf = append(e.buf, byte(t)) }

func (e *encoder) uvarint(v uint64) {
	for v >= 0x80 {
		e.buf = append(e.buf, byte(v)|0x80)
		v >>= 7
	}
	e.buf = append(e.buf, byte(v))
}

// int writes a signed value as a zigzag varint, keeping small negatives short.
func (e *encoder) int(v int) { e.uvarint(uint64(v<<1) ^ uint64(v>>63)) }

func (e *encoder) str(s string) {
	if id, ok := e.table[s]; ok {
		e.uvarint(uint64(id)<<1 | 1)
		return
	}
	e.table[s] = uint32(len(e.table))
	e.uvarint(uint64(len(s)) << 1)
	e.buf = append(e.buf, s...)
}

func (e *encoder) boolean(v bool) {
	if v {
		e.byte(1)
		return
	}
	e.byte(0)
}

func (e *encoder) value(v any) {
	switch x := v.(type) {
	case nil:
		e.byte(byte(valNil))
	case bool:
		e.byte(byte(valBool))
		e.boolean(x)
	case int:
		e.byte(byte(valInt))
		e.int(x)
	case int64:
		e.byte(byte(valInt64))
		e.uvarint(uint64(x<<1) ^ uint64(x>>63))
	case float64:
		e.byte(byte(valFloat))
		e.uvarint(math.Float64bits(x))
	case string:
		e.byte(byte(valString))
		e.str(x)
	default:
		e.err = &encodeError{node: "literal " + typeName(v)}
	}
}

// node writes one node and every node below it.
func (e *encoder) node(n any) {
	if e.err != nil {
		return
	}
	start := len(e.buf)
	e.nodeBody(n)
	if e.err != nil {
		return
	}
	e.span(n, tag(e.buf[start]))
}

// span writes the source span of a statement node, or that it holds none.
func (e *encoder) span(n any, t tag) {
	stmt, ok := n.(model.Stmt)
	if !ok || !spanTag[t] {
		return
	}
	span, ok := e.spans[stmt]
	if !ok {
		e.byte(0)
		return
	}
	e.byte(1)
	e.int(span.Start)
	e.int(span.End)
}

func (e *encoder) nodeBody(n any) {
	switch v := n.(type) {
	case *model.InlineHTML:
		e.tag(tagInlineHTML)
		e.str(v.Text)
	case *model.Echo:
		e.tag(tagEcho)
		e.exprList(v.Args)
	case *model.ExprStmt:
		e.tag(tagExprStmt)
		e.expr(v.X)
	case *model.Assign:
		e.tag(tagAssign)
		e.expr(v.Target)
		e.str(v.Op)
		e.expr(v.Value)
	case *model.If:
		e.tag(tagIf)
		e.expr(v.Cond)
		e.stmtList(v.Then)
		e.stmtList(v.Else)
		e.int(v.ElseLine)
	case *model.Foreach:
		e.tag(tagForeach)
		e.expr(v.Source)
		e.expr(v.KeyTarget)
		e.expr(v.ValTarget)
		e.boolean(v.ByRef)
		e.str(v.KeyVar)
		e.str(v.ValVar)
		e.stmtList(v.Body)
	case *model.For:
		e.tag(tagFor)
		e.stmt(v.Init)
		e.expr(v.Cond)
		e.stmt(v.Post)
		e.stmtList(v.Body)
	case *model.DoWhile:
		e.tag(tagDoWhile)
		e.stmtList(v.Body)
		e.expr(v.Cond)
	case *model.Return:
		e.tag(tagReturn)
		e.expr(v.Value)
	case *model.Include:
		e.tag(tagInclude)
		e.expr(v.Path)
		e.str(v.Keyword)
		e.boolean(v.Once)
		e.boolean(v.Parenthesized)
	case *model.FuncDecl:
		e.tag(tagFuncDecl)
		e.funcDecl(v)
	case *model.ClassDecl:
		e.tag(tagClassDecl)
		e.classDecl(v)
	case *model.TraitDecl:
		e.tag(tagTraitDecl)
		e.str(v.Name)
		e.strList(v.Traits)
		e.traitRules(v.TraitRules)
		e.fields(v.Fields)
		e.fields(v.Statics)
		e.fields(v.Consts)
		e.funcDecls(v.Methods)
	case *model.InterfaceDecl:
		e.tag(tagInterfaceDecl)
		e.str(v.Name)
		e.strList(v.Extends)
		e.fields(v.Consts)
		e.funcDecls(v.Methods)
	case *model.Use:
		e.tag(tagUse)
		e.str(v.Kind)
		e.count(len(v.Imports), v.Imports == nil)
		for _, imp := range v.Imports {
			e.str(imp.Path)
			e.str(imp.Alias)
			e.str(imp.Package)
		}
	case *model.Declare:
		e.tag(tagDeclare)
		e.count(len(v.Directives), v.Directives == nil)
		for _, d := range v.Directives {
			e.str(d.Name)
			e.expr(d.Value)
		}
		e.stmtList(v.Body)
		e.boolean(v.Block)
	case *model.Unset:
		e.tag(tagUnset)
		e.exprList(v.Targets)
	case *model.StaticVar:
		e.tag(tagStaticVar)
		e.count(len(v.Vars), v.Vars == nil)
		for _, d := range v.Vars {
			e.str(d.Name)
			e.expr(d.Default)
		}
	case *model.Global:
		e.tag(tagGlobal)
		e.strList(v.Names)
	case *model.ConstDecl:
		e.tag(tagConstDecl)
		e.fields(v.Consts)
	case *model.Throw:
		e.tag(tagThrow)
		e.expr(v.X)
	case *model.Try:
		e.tag(tagTry)
		e.stmtList(v.Body)
		e.count(len(v.Catches), v.Catches == nil)
		for _, c := range v.Catches {
			e.str(c.Type)
			e.str(c.Var)
			e.stmtList(c.Body)
			e.int(c.Line)
		}
		e.stmtList(v.Finally)
		e.int(v.FinallyLine)
	case *model.Switch:
		e.tag(tagSwitch)
		e.expr(v.Cond)
		e.count(len(v.Cases), v.Cases == nil)
		for _, c := range v.Cases {
			e.expr(c.Value)
			e.stmtList(c.Body)
			e.int(c.Line)
		}
		e.stmtList(v.Default)
	case *model.Break:
		e.tag(tagBreak)
		e.int(v.Line)
		e.int(v.Levels)
	case *model.Continue:
		e.tag(tagContinue)
		e.int(v.Line)
		e.int(v.Levels)
	case *model.Label:
		e.tag(tagLabel)
		e.str(v.Name)
		e.int(v.Line)
	case *model.Goto:
		e.tag(tagGoto)
		e.str(v.Label)
		e.int(v.Line)
	case *model.Lit:
		e.tag(tagLit)
		e.value(v.Value)
		e.str(v.Raw)
	case *model.Interp:
		e.tag(tagInterp)
		e.exprList(v.Parts)
		e.str(v.Raw)
	case *model.Var:
		e.tag(tagVar)
		e.str(v.Name)
		e.boolean(v.Const)
	case *model.ArrayLit:
		e.tag(tagArrayLit)
		e.items(v.Items)
	case *model.GoLit:
		e.tag(tagGoLit)
		e.boolean(v.Map)
		e.str(v.KeyType)
		e.str(v.ValType)
		e.items(v.Items)
	case *model.AddrOf:
		e.tag(tagAddrOf)
		e.expr(v.X)
	case *model.Deref:
		e.tag(tagDeref)
		e.expr(v.X)
	case *model.Index:
		e.tag(tagIndex)
		e.expr(v.Base)
		e.expr(v.Index)
	case *model.PropAccess:
		e.tag(tagPropAccess)
		e.expr(v.Base)
		e.str(v.Name)
		e.expr(v.NameExpr)
	case *model.Call:
		e.tag(tagCall)
		e.str(v.Name)
		e.str(v.Fallback)
		e.exprList(v.Args)
		e.boolean(v.Bare)
		e.boolean(v.FirstClass)
	case *model.MethodCall:
		e.tag(tagMethodCall)
		e.expr(v.Base)
		e.str(v.Method)
		e.expr(v.MethodExpr)
		e.exprList(v.Args)
		e.boolean(v.FirstClass)
	case *model.New:
		e.tag(tagNew)
		e.str(v.Class)
		e.expr(v.ClassExpr)
		e.exprList(v.Args)
		if v.Decl == nil {
			e.tag(tagNil)
		} else if idx, ok := e.anon[v.Decl]; ok {
			e.tag(tagAnonRef)
			e.uvarint(uint64(idx))
		} else {
			e.node(v.Decl)
		}
	case *model.Ref:
		e.tag(tagRef)
		e.expr(v.X)
	case *model.Unary:
		e.tag(tagUnary)
		e.str(v.Op)
		e.expr(v.X)
		e.boolean(v.Postfix)
	case *model.Parenthesized:
		e.tag(tagParenthesized)
		e.expr(v.X)
	case *model.Binary:
		e.tag(tagBinary)
		e.str(v.Op)
		e.expr(v.Left)
		e.expr(v.Right)
	case *model.Ternary:
		e.tag(tagTernary)
		e.expr(v.Cond)
		e.expr(v.Then)
		e.expr(v.Else)
	case *model.Coalesce:
		e.tag(tagCoalesce)
		e.expr(v.Left)
		e.expr(v.Right)
		e.boolean(v.Assign)
	case *model.ClassConst:
		e.tag(tagClassConst)
		e.str(v.Class)
		e.str(v.Name)
	case *model.Cast:
		e.tag(tagCast)
		e.str(v.Type)
		e.expr(v.X)
	case *model.Closure:
		e.tag(tagClosure)
		e.params(v.Params)
		e.count(len(v.Uses), v.Uses == nil)
		for _, u := range v.Uses {
			e.str(u.Name)
			e.boolean(u.ByRef)
		}
		e.stmtList(v.Body)
		e.str(v.ReturnType)
		e.boolean(v.Static)
		e.boolean(v.ByRef)
	case *model.AssignExpr:
		e.tag(tagAssignExpr)
		e.expr(v.Target)
		e.str(v.Op)
		e.expr(v.Value)
		e.int(v.Line)
	case *model.ListExpr:
		e.tag(tagListExpr)
		e.exprList(v.Elems)
	case *model.StaticCall:
		e.tag(tagStaticCall)
		e.str(v.Class)
		e.str(v.Method)
		e.expr(v.MethodExpr)
		e.exprList(v.Args)
		e.boolean(v.FirstClass)
	case *model.StaticProp:
		e.tag(tagStaticProp)
		e.str(v.Class)
		e.str(v.Name)
		e.expr(v.NameExpr)
	case *model.Invoke:
		e.tag(tagInvoke)
		e.expr(v.Callee)
		e.exprList(v.Args)
		e.boolean(v.FirstClass)
	default:
		e.err = &encodeError{node: typeName(n)}
	}
}

func (e *encoder) expr(x model.Expr) {
	if x == nil {
		e.tag(tagNil)
		return
	}
	e.node(x)
}

func (e *encoder) stmt(s model.Stmt) {
	if s == nil {
		e.tag(tagNil)
		return
	}
	e.node(s)
}

// count writes a slice length, and whether the slice is nil.
func (e *encoder) count(n int, isNil bool) {
	if isNil {
		e.uvarint(0)
		return
	}
	e.uvarint(uint64(n) + 1)
}

func (e *encoder) exprList(list []model.Expr) {
	e.count(len(list), list == nil)
	for _, x := range list {
		e.expr(x)
	}
}

func (e *encoder) stmtList(list []model.Stmt) {
	e.count(len(list), list == nil)
	for _, s := range list {
		e.stmt(s)
	}
}

func (e *encoder) strList(list []string) {
	e.count(len(list), list == nil)
	for _, s := range list {
		e.str(s)
	}
}

func (e *encoder) items(list []model.ArrayItem) {
	e.count(len(list), list == nil)
	for _, item := range list {
		e.expr(item.Key)
		e.expr(item.Val)
		e.boolean(item.Ref)
	}
}

func (e *encoder) params(list []model.Param) {
	e.count(len(list), list == nil)
	for _, p := range list {
		e.str(p.Name)
		e.expr(p.Default)
		e.str(p.Modifiers)
		e.str(p.Type)
		e.str(p.Resolved)
		e.boolean(p.ByRef)
		e.boolean(p.Variadic)
	}
}

func (e *encoder) fields(list []model.Field) {
	e.count(len(list), list == nil)
	for _, f := range list {
		e.str(f.Name)
		e.expr(f.Default)
		e.str(f.Visibility)
		e.boolean(f.Readonly)
		e.str(f.Type)
		e.str(f.Resolved)
		e.int(f.Span.Start)
		e.int(f.Span.End)
	}
}

func (e *encoder) traitRules(list []model.TraitRule) {
	e.count(len(list), list == nil)
	for _, r := range list {
		e.str(r.Trait)
		e.str(r.Method)
		e.strList(r.Insteadof)
		e.str(r.Alias)
		e.str(r.Visibility)
	}
}

func (e *encoder) funcDecl(f *model.FuncDecl) {
	e.str(f.Class)
	e.str(f.Name)
	e.str(f.Filename)
	e.params(f.Params)
	e.stmtList(f.Body)
	e.str(f.Visibility)
	e.str(f.ReturnType)
	e.str(f.ResolvedReturn)
	e.boolean(f.Static)
	e.boolean(f.Abstract)
	e.boolean(f.ByRef)
}

func (e *encoder) funcDecls(list []*model.FuncDecl) {
	e.count(len(list), list == nil)
	for _, f := range list {
		if f == nil {
			e.tag(tagNil)
			continue
		}
		e.node(f)
	}
}

func (e *encoder) classDecl(c *model.ClassDecl) {
	e.str(c.Name)
	e.str(c.Filename)
	e.str(c.Parent)
	e.strList(c.Implements)
	e.strList(c.Traits)
	e.traitRules(c.TraitRules)
	e.boolean(c.Abstract)
	e.boolean(c.Final)
	e.boolean(c.Readonly)
	e.fields(c.Fields)
	e.fields(c.Statics)
	e.fields(c.Consts)
	e.funcDecls(c.Methods)
}

func (e *encoder) classes(list []*model.ClassDecl) {
	e.count(len(list), list == nil)
	for _, c := range list {
		if c == nil {
			e.tag(tagNil)
			continue
		}
		e.node(c)
	}
}
