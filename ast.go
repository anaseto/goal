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
	ppNode()
}

// exprs is used to represent a whole expression. It is not a ppExpr.
type exprs []expr

// pToken represents a simplified token after processing into expr.
type pToken struct {
	Type pTokenType
	Rune rune
	Pos  int
	Text string
}

func (pt pToken) String() string {
	return fmt.Sprintf("{%v %d %s}", pt.Type, pt.Pos, pt.Text)
}

// pTokenType represents tokens in a ppExpr.
type pTokenType int32

// These constants represent the possible tokens in a ppExpr. The SEP,
// EOF and CLOSE types are not emitted in the final result.
const (
	pSEP pTokenType = iota
	pEOF
	pCLOSE

	pNUMBER
	pSTRING
	pIDENT
	pVERB
	pADVERB
)

type pStrand []pToken // for stranding, like 1 23 456

type pAdverbs []pToken // for an adverb sequence

type pParenExpr exprs // for parenthesized sub-expressions

type pBlock struct {
	Type pBlockType
	Body []exprs
	Args []string
}

func (pb pBlock) String() (s string) {
	switch pb.Type {
	case pLAMBDA:
		args := "[" + strings.Join([]string(pb.Args), ";") + "]"
		s = fmt.Sprintf("{%s %v %v}", args, pb.Type, pb.Body)
	case pARGS:
		s = fmt.Sprintf("[%v %v]", pb.Type, pb.Body)
	case pLIST:
		s = fmt.Sprintf("(%v %v)", pb.Type, pb.Body)
	}
	return s
}

func (pb pBlock) push(pe expr) {
	pb.Body[len(pb.Body)-1] = append(pb.Body[len(pb.Body)-1], pe)
}

type pBlockType int

const (
	pLAMBDA pBlockType = iota
	pARGS
	pSEQ
	pLIST
)

func (pt pToken) ppNode()     {}
func (ps pStrand) ppNode()    {}
func (pa pAdverbs) ppNode()   {}
func (pp pParenExpr) ppNode() {}
func (pb pBlock) ppNode()     {}

// pIter is an iterator for exprs slices, with peek functionality.
type pIter struct {
	pps exprs
	i   int
}

func newpIter(pps exprs) pIter {
	return pIter{pps: pps, i: -1}
}

func (it *pIter) Next() bool {
	it.i++
	return it.i < len(it.pps)
}

func (it *pIter) Expr() expr {
	return it.pps[it.i]
}

func (it *pIter) Index() int {
	return it.i
}

func (it *pIter) Set(i int) {
	it.i = i
}

func (it *pIter) Peek() expr {
	if it.i+1 >= len(it.pps) {
		return nil
	}
	return it.pps[it.i+1]
}

func (it *pIter) PeekN(n int) expr {
	if it.i+n >= len(it.pps) {
		return nil
	}
	return it.pps[it.i+n]
}
