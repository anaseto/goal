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
	astREGEXP
	astEMPTYLIST
)

// astReturn represents an early return :x or 'x.
type astReturn struct {
	Expr    expr
	OnError bool
}

// astAssign represents an assignment x:y.
type astAssign struct {
	Name   string // x
	Global bool   // whether :: or not
	Right  expr   // y
	Pos    int
}

// astListAssign represents an assignment (x0;...):y.
type astListAssign struct {
	Names  []string // (x0;...)
	Global bool     // whether :: or not
	Right  expr     // y
	Pos    int
}

// astAssignOp represents a variable assignment with a built-in operator, of
// the form x op: y, semantically equivalent to x: x op y.
type astAssignOp struct {
	Name   string // x
	Global bool   // wether :: or not
	Dyad   string // op
	Right  expr   // y
	Pos    int
}

// astAssinAmendOp represents an assign-amend call with a built-in operator, of
// the form x[y]op: z, semantically equivalent to x: @[x;y;op;z].
type astAssignAmendOp struct {
	Name    string // x
	Global  bool   // whether :: or not
	Dyad    string // op
	Indices expr   // y
	Right   expr   // z
	Pos     int
}

// astAssinDeepAmendOp represents an assign-amend call with a built-in operator, of
// the form x[y;...]op: z, semantically equivalent to x: .[x;y;op;z].
type astAssignDeepAmendOp struct {
	Name    string   // x
	Global  bool     // whether :: or not
	Dyad    string   // op
	Indices *astList // y
	Right   expr     // z
	Pos     int
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
	Verb  expr // dyad
	Left  expr
	Right expr
}

type astApply2Adverb struct {
	Verb  expr // derived verb
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

func (es exprs) node()                {}
func (t *astToken) node()             {}
func (a *astReturn) node()            {}
func (a *astAssign) node()            {}
func (a *astListAssign) node()        {}
func (a *astAssignOp) node()          {}
func (a *astAssignAmendOp) node()     {}
func (a *astAssignDeepAmendOp) node() {}
func (st *astStrand) node()           {}
func (dv *astDerivedVerb) node()      {}
func (p *astParen) node()             {}
func (a *astApply2) node()            {}
func (a *astApply2Adverb) node()      {}
func (a *astApplyN) node()            {}
func (l *astList) node()              {}
func (b *astSeq) node()               {}
func (b *astLambda) node()            {}

func nonEmpty(e expr) bool {
	switch e := e.(type) {
	case exprs:
		return len(e) > 0
	default:
		return true
	}
}

type parseCLOSE struct {
	Pos int
}

func (p parseCLOSE) Error() string { return "CLOSE" }
