package goal

// expr are built by the first left to right pass, resulting in a tree
// of blocks producing a whole expression, with simplified token
// information, and stack-like order. It is a non-resolved AST intermediary
// representation.
type expr interface {
	node()
}

// exprs represents a sequence of expressions applied in sequence monadically.
type exprs []expr

// astToken represents a simplified token after processing into expr.
type astToken struct {
	Type astTokenType
	Pos  int
	Text string
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

// astReturn represents an early return :x or 'x.
type astReturn struct {
	Expr    expr
	OnError bool
}

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

// astDerivedVerb represents a derived verb.
type astDerivedVerb struct {
	Adverb *astToken
	Verb   expr
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

type astApplyN struct {
	Verb     expr
	Args     []expr
	StartPos int // remove?
	EndPos   int // remove?
}

type astList struct {
	Args     []expr
	StartPos int // remove?
	EndPos   int // remove?
}

type astSeq struct {
	Body     []expr
	StartPos int // remove?
	EndPos   int // remove?
}

type astLambda struct {
	Body     []expr
	Args     []string
	StartPos int
	EndPos   int
}

func (es exprs) node()           {}
func (t *astToken) node()        {}
func (a *astReturn) node()       {}
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
