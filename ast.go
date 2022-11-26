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

// exprs is used to represent a whole expression. It is not a ppExpr. It
// represents a sequence of expressions applied in sequence.
type exprs []expr

// astToken represents a simplified token after processing into expr.
type astToken struct {
	Type astTokenType
	Pos  int
	Text string
}

func (t *astToken) String() string {
	return fmt.Sprintf("{%v %d %s}", t.Type, t.Pos, t.Text)
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
	astADVERB // only within astAdverbs
)

type astReturn struct {
	Pos int
}

type astStrand struct {
	Lits []astToken // stranding of literals, like 1 23 456
	Pos  int
}

type astAdverbs struct {
	Train []astToken // an adverb sequence
}

type astParen struct {
	Exprs    exprs // parenthesized sub-expressions
	StartPos int   // remove?
	EndPos   int   // remove?
}

type astApplyN struct {
	Expr     expr
	Args     []exprs
	StartPos int // remove?
	EndPos   int // remove?
}

func (a *astApplyN) String() (s string) {
	s = fmt.Sprintf("%v[%v]", a.Expr, a.Args)
	return s
}

func (a *astApplyN) push(e expr) {
	a.Args[len(a.Args)-1] = append(a.Args[len(a.Args)-1], e)
}

type astList struct {
	Args     []exprs
	StartPos int // remove?
	EndPos   int // remove?
}

func (l *astList) String() (s string) {
	s = fmt.Sprintf("(%v)", l.Args)
	return s
}

func (l *astList) push(e expr) {
	l.Args[len(l.Args)-1] = append(l.Args[len(l.Args)-1], e)
}

type astSeq struct {
	Body     []exprs
	StartPos int // remove?
	EndPos   int // remove?
}

func (b *astSeq) String() (s string) {
	s = fmt.Sprintf("[%v]", b.Body)
	return s
}

func (b *astSeq) push(e expr) {
	b.Body[len(b.Body)-1] = append(b.Body[len(b.Body)-1], e)
}

type astLambda struct {
	Body     []exprs
	Args     []string
	StartPos int
	EndPos   int
}

func (b *astLambda) String() (s string) {
	args := "[" + strings.Join([]string(b.Args), ";") + "]"
	return fmt.Sprintf("{%s %v}", args, b.Body)
}

func (b *astLambda) push(e expr) {
	b.Body[len(b.Body)-1] = append(b.Body[len(b.Body)-1], e)
}

func (t *astToken) node()     {}
func (t *astReturn) node()    {}
func (st *astStrand) node()   {}
func (ads *astAdverbs) node() {}
func (p *astParen) node()     {}
func (a *astApplyN) node()    {}
func (l *astList) node()      {}
func (b *astSeq) node()       {}
func (b *astLambda) node()    {}

type parseEOF struct{}
type parseSEP struct{}
type parseCLOSE struct {
	Pos int
}

func (p parseEOF) Error() string   { return "EOF" }
func (p parseSEP) Error() string   { return "SEP" }
func (p parseCLOSE) Error() string { return "CLOSE" }

// astIter is an iterator for exprs slices, with peek functionality.
type astIter struct {
	exprs exprs
	i     int
}

func newAstIter(es exprs) astIter {
	return astIter{exprs: es, i: -1}
}

func (it *astIter) Next() bool {
	it.i++
	return it.i < len(it.exprs)
}

func (it *astIter) Expr() expr {
	return it.exprs[it.i]
}

func (it *astIter) Index() int {
	return it.i
}

func (it *astIter) Set(i int) {
	it.i = i
}

func (it *astIter) Peek() expr {
	if it.i+1 >= len(it.exprs) {
		return nil
	}
	return it.exprs[it.i+1]
}

func (it *astIter) PeekN(n int) expr {
	if it.i+n >= len(it.exprs) {
		return nil
	}
	return it.exprs[it.i+n]
}
