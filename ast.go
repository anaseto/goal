package goal

import (
	"fmt"
	"strings"
)

// expr are built by the first left to right pass, resulting in a tree
// of blocks producing a whole expression, with simplified token
// information, and stack-like order. It is a non-resolved AST intermediary
// representation.
type expr interface {
	node()
}

// exprs represents a sequence of expressions applied in sequence monadically.
type exprs []expr

func (es exprs) String() string {
	sb := &strings.Builder{}
	for i, e := range es {
		fmt.Fprintf(sb, "%v", e)
		if i < len(es)-1 {
			fmt.Fprint(sb, " ")
		}
	}
	return sb.String()
}

// astToken represents a simplified token after processing into expr.
type astToken struct {
	Type astTokenType
	Pos  int
	Text string
}

func (t *astToken) String() string {
	return fmt.Sprintf("%v[%s]", t.Type, t.Text)
}

// astTokenType represents tokens in a ppExpr.
type astTokenType int

// These constants represent the possible tokens in a ppExpr.
const (
	astNUMBER astTokenType = iota
	astSTRING
	astIDENT
	astMONAD
	astDYAD
	astADVERB // only within astDerivedVerb
	astEMPTYLIST
)

// astAssign represents an assignment x:y.
type astAssign struct {
	Name   string
	Global bool
	Right  expr
	Pos    int
}

// astAssignOp represents a variable assignment with a built-in operator, of
// the form x op: y, semantically equivalent to x: x op y.
type astAssignOp struct {
	Name   string
	Global bool
	Dyad   string
	Right  expr
	Pos    int
}

// astStrand represents a stranding of literals, like 1 23 456
type astStrand struct {
	Lits []astToken
	Pos  int
}

func astTokensString(toks []astToken) string {
	sb := &strings.Builder{}
	for i, e := range toks {
		fmt.Fprintf(sb, "%v[%s]", e.Type, e.Text)
		if i < len(toks)-1 {
			fmt.Fprint(sb, ";")
		}
	}
	return sb.String()
}

func (t *astStrand) String() string {
	return fmt.Sprintf("STRAND[%s]", astTokensString(t.Lits))
}

// astDerivedVerb represents a derived verb.
type astDerivedVerb struct {
	Adverb *astToken
	Verb   expr
}

func (t *astDerivedVerb) String() string {
	return fmt.Sprintf("DERIVED[%s;%v]", t.Adverb.Text, t.Verb)
}

type astParen struct {
	Expr     expr // parenthesized sub-expressions
	StartPos int  // remove?
	EndPos   int  // remove?
}

type astApply2 struct {
	Verb  expr // dyad or derived verb
	Left  expr
	Right expr
}

func (a *astApply2) String() (s string) {
	s = fmt.Sprintf("%v[%v;%v]", a.Verb, a.Left, a.Right)
	return s
}

func argsString(es []expr) string {
	sb := &strings.Builder{}
	for i, e := range es {
		fmt.Fprintf(sb, "%v", e)
		if i < len(es)-1 {
			fmt.Fprint(sb, ";")
		}
	}
	return sb.String()
}

type astApplyN struct {
	Verb     expr
	Args     []expr
	StartPos int // remove?
	EndPos   int // remove?
}

func (a *astApplyN) String() (s string) {
	s = fmt.Sprintf("%v[%s]", a.Verb, argsString(a.Args))
	return s
}

type astList struct {
	Args     []expr
	StartPos int // remove?
	EndPos   int // remove?
}

func (l *astList) String() (s string) {
	s = fmt.Sprintf("list[%d;%s]", len(l.Args), argsString(l.Args))
	return s
}

type astSeq struct {
	Body     []expr
	StartPos int // remove?
	EndPos   int // remove?
}

func (b *astSeq) String() (s string) {
	s = fmt.Sprintf("[%v]", argsString(b.Body))
	return s
}

type astLambda struct {
	Body     []expr
	Args     []string
	StartPos int
	EndPos   int
}

func (b *astLambda) String() (s string) {
	args := "[" + strings.Join([]string(b.Args), ";") + "]"
	return fmt.Sprintf("{%s %v}", args, argsString(b.Body))
}

func (es exprs) node()           {}
func (t *astToken) node()        {}
func (a *astAssign) node()       {}
func (a *astAssignOp) node()     {}
func (st *astStrand) node()      {}
func (dv *astDerivedVerb) node() {}
func (p *astParen) node()        {}
func (a *astApply2) node()       {}
func (a *astApplyN) node()       {}
func (l *astList) node()         {}
func (b *astSeq) node()          {}
func (b *astLambda) node()       {}

func nonEmpty(e expr) bool {
	switch e := e.(type) {
	case exprs:
		return len(e) > 0
	case *astParen:
		return nonEmpty(e.Expr)
	default:
		return true
	}
}

type parseEOF struct {
	Pos int
}

type parseCLOSE struct {
	Pos int
}

func (p parseEOF) Error() string   { return "EOF" }
func (p parseCLOSE) Error() string { return "CLOSE" }
