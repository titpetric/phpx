// Package astcodec encodes a parsed model.Program to bytes and reads it back,
// so a build cache can skip the lexer and the parser for an unchanged file.
//
// The payload is a versioned private format, not an interchange format. A
// reader that meets a tag, a value type or a field order it does not know
// answers an error, and the caller parses the source instead. Every field of
// every node is written, and a statement keeps its source span.
//
// Encoding pays once for a file a run parses. Decoding is the hot path: it
// allocates nodes in chunks the way the parser does.
package astcodec

import (
	"errors"
	"fmt"
)

// ErrFormat reports bytes that do not decode as an encoded program.
var ErrFormat = errors.New("astcodec: malformed program")

// Format is the payload layout; a change to a field order changes it.
const Format = 1

// tag names one node type of the payload.
type tag uint8

// valueTag names one dynamic type Lit.Value holds.
type valueTag uint8

// The node tags. A statement and an expression share one space.
const (
	tagNil tag = iota
	tagInlineHTML
	tagEcho
	tagExprStmt
	tagAssign
	tagIf
	tagForeach
	tagFor
	tagDoWhile
	tagReturn
	tagInclude
	tagFuncDecl
	tagClassDecl
	tagTraitDecl
	tagInterfaceDecl
	tagUse
	tagDeclare
	tagUnset
	tagStaticVar
	tagGlobal
	tagConstDecl
	tagThrow
	tagTry
	tagSwitch
	tagBreak
	tagContinue
	tagLabel
	tagGoto
	tagLit
	tagInterp
	tagVar
	tagArrayLit
	tagGoLit
	tagAddrOf
	tagDeref
	tagIndex
	tagPropAccess
	tagCall
	tagMethodCall
	tagNew
	tagRef
	tagUnary
	tagParenthesized
	tagBinary
	tagTernary
	tagCoalesce
	tagClassConst
	tagCast
	tagClosure
	tagAssignExpr
	tagListExpr
	tagStaticCall
	tagStaticProp
	tagInvoke
)

// tagAnonRef names an anonymous class by its index in AnonClasses.
const tagAnonRef tag = 255

// spanTag reports whether a node of this tag carries a source span.
var spanTag = func() [256]bool {
	var out [256]bool
	for _, t := range []tag{
		tagInlineHTML, tagEcho, tagExprStmt, tagAssign, tagIf, tagForeach, tagFor,
		tagDoWhile, tagReturn, tagInclude, tagFuncDecl, tagClassDecl, tagTraitDecl,
		tagInterfaceDecl, tagUse, tagDeclare, tagUnset, tagStaticVar, tagGlobal,
		tagConstDecl, tagThrow, tagTry, tagSwitch, tagBreak, tagContinue, tagLabel,
		tagGoto,
	} {
		out[t] = true
	}
	return out
}()

// The literal value tags, one per dynamic type Lit.Value holds.
const (
	valNil valueTag = iota
	valBool
	valInt
	valInt64
	valFloat
	valString
)

// encodeError reports a node the encoder does not know.
type encodeError struct{ node string }

func (e encodeError) Error() string { return fmt.Sprintf("astcodec: unknown node %s", e.node) }
