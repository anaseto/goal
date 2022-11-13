package goal

import (
	"fmt"
	"strings"
)

// ppExpr are built by the first left to right pass, resulting in a tree
// of blocks producing a whole expression, with simplified token
// information, and stack-like order. It is a pre-processing IR. The
// representation of specific semantics of the language is left to a
// second IR built around type Expr.
type ppExpr interface {
	ppNode()
}

// ppExprs is used to represent a whole expression. It is not a ppExpr.
type ppExprs []ppExpr

// ppIter is an iterator for ppExprs slices, with peek functionality.
type ppIter struct {
	pps ppExprs
	i   int
}

func newppIter(pps ppExprs) ppIter {
	return ppIter{pps: pps, i: -1}
}

func (it *ppIter) Next() bool {
	it.i++
	return it.i < len(it.pps)
}

func (it *ppIter) Expr() ppExpr {
	return it.pps[it.i]
}

func (it *ppIter) Index() int {
	return it.i
}

func (it *ppIter) Set(i int) {
	it.i = i
}

func (it *ppIter) Peek() ppExpr {
	if it.i+1 >= len(it.pps) {
		return nil
	}
	return it.pps[it.i+1]
}

func (it *ppIter) PeekN(n int) ppExpr {
	if it.i+n >= len(it.pps) {
		return nil
	}
	return it.pps[it.i+n]
}

// ppToken represents a simplified token after processing into ppExpr.
type ppToken struct {
	Type ppTokenType
	Rune rune
	Pos  int
	Text string
}

func (ppt ppToken) String() string {
	return fmt.Sprintf("{%v %d %s}", ppt.Type, ppt.Pos, ppt.Text)
}

// ppTokenType represents tokens in a ppExpr.
type ppTokenType int32

// These constants represent the possible tokens in a ppExpr. The SEP,
// EOF and CLOSE types are not emitted in the final result.
const (
	ppSEP ppTokenType = iota
	ppEOF
	ppCLOSE

	ppNUMBER
	ppSTRING
	ppIDENT
	ppVERB
	ppADVERB
)

type ppStrand []ppToken // for stranding, like 1 23 456

type ppAdverbs []ppToken // for an adverb sequence

type ppParenExpr ppExprs // for parenthesized sub-expressions

type ppBlock struct {
	Type ppBlockType
	Body []ppExprs
	Args []string
}

func (ppb ppBlock) String() (s string) {
	switch ppb.Type {
	case ppLAMBDA:
		args := "[" + strings.Join([]string(ppb.Args), ";") + "]"
		s = fmt.Sprintf("{%s %v %v}", args, ppb.Type, ppb.Body)
	case ppARGS:
		s = fmt.Sprintf("[%v %v]", ppb.Type, ppb.Body)
	case ppLIST:
		s = fmt.Sprintf("(%v %v)", ppb.Type, ppb.Body)
	}
	return s
}

func (ppb ppBlock) push(ppe ppExpr) {
	ppb.Body[len(ppb.Body)-1] = append(ppb.Body[len(ppb.Body)-1], ppe)
}

type ppBlockType int

const (
	ppLAMBDA ppBlockType = iota
	ppARGS
	ppSEQ
	ppLIST
)

func (tok ppToken) ppNode()     {}
func (pps ppStrand) ppNode()    {}
func (ppa ppAdverbs) ppNode()   {}
func (ppp ppParenExpr) ppNode() {}
func (ppb ppBlock) ppNode()     {}
