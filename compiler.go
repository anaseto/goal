package goal

import (
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// globalCode represents the last code compiled in global context, outside any
// lambda.
type globalCode struct {
	Body []opcode // compiled code
	Pos  []int    // positions in the source

	last int // index of last non-argument opcode
}

// lambdaCode represents a compiled user defined function.
type lambdaCode struct {
	Body        []opcode  // object code of the function
	Pos         []int     // position associated to opcode of same index
	Names       []string  // local arguments and variables names
	Rank        int       // number of arguments
	Source      string    // source code of the function
	Filename    string    // filename of the file containing the source (if any)
	StartPos    int       // starting position in the source
	UnusedArgs  []int32   // reversed indices of unused arguments
	UsedArgs    []int32   // reversed indices of used arguments
	AssignLists [][]int32 // assignement lists (resolved)

	namedArgs   bool                   // uses named parameters like {[a;b;c]....}
	lastUses    []lastUse              // opcode index and block number of variable last use
	joinPoints  []int32                // number of jumps ending at a given opcode index
	assignLists [][]lambdaLocal        // assignement lists (locals)
	locals      map[string]lambdaLocal // arguments and variables
	opIdxLocal  map[int]lambdaLocal    // opcode index -> local variable
	nVars       int                    // number of non-argument variables
}

// lambdaLocal represents either an argument or a local variable. IDs are
// unique for a given type only.
type lambdaLocal struct {
	Type localType
	ID   int
}

// localType represents different kinds of locals.
type localType int

// These constants describe the supported kinds of locals.
const (
	localArg localType = iota
	localVar
)

func (l *lambdaCode) local(s string) (lambdaLocal, bool) {
	if strings.ContainsRune(s, '.') {
		return lambdaLocal{}, false
	}
	param, ok := l.locals[s]
	if ok {
		return param, true
	}
	if !l.namedArgs && len(s) == 1 {
		switch r := rune(s[0]); r {
		case 'x', 'y', 'z':
			for rr := 'x'; rr <= r; rr++ {
				// If z is used, then arity is 3, even if y and
				// x are not used.
				rs := string(rr)
				_, ok := l.locals[rs]
				if ok {
					continue
				}
				id := rr - 'x'
				arg := lambdaLocal{Type: localArg, ID: int(id)}
				l.locals[rs] = arg
				if rr == r {
					return arg, true
				}
			}
		}
	}
	return lambdaLocal{}, false
}

// programString returns a string representation of the compiled program and
// relevant data.
func (ctx *Context) programString() string {
	sb := strings.Builder{}
	fmt.Fprintln(&sb, "---- Compiled program -----")
	fmt.Fprintln(&sb, "Instructions:")
	fmt.Fprint(&sb, ctx.opcodesString(ctx.gCode.Body, nil))
	fmt.Fprintln(&sb, "Globals:")
	for id, name := range ctx.gNames {
		fmt.Fprintf(&sb, "\t%s\t%d\n", name, id)
	}
	fmt.Fprintln(&sb, "Constants:")
	for id, ci := range ctx.constants {
		fmt.Fprintf(&sb, "\t%d\t%s\n", id, ci.Sprint(ctx))
	}
	for id, lc := range ctx.lambdas {
		fmt.Fprintf(&sb, "---- Lambda %d (Rank: %d) -----\n", id, lc.Rank)
		fmt.Fprintf(&sb, "%s", ctx.lambdaString(lc))
	}
	return sb.String()
}

func (ctx *Context) lambdaString(lc *lambdaCode) string {
	sb := strings.Builder{}
	fmt.Fprintln(&sb, "Instructions:")
	fmt.Fprint(&sb, ctx.opcodesString(lc.Body, lc))
	fmt.Fprintln(&sb, "Locals:")
	for i, name := range lc.Names {
		fmt.Fprintf(&sb, "\t%s\t%d\n", name, i)
	}
	return sb.String()
}

// compiler incrementally builds a semi-resolved program from a parsed expr.
type compiler struct {
	ctx        *Context      // main execution and compilation context
	p          *parser       // parsing into text-based non-resolved AST
	scopeStack []*lambdaCode // scope information
	pos        int           // last token position
	drop       bool          // whether to add a drop at the end
}

func newCompiler(ctx *Context) *compiler {
	c := &compiler{
		ctx: ctx,
		p:   newParser(ctx),
	}
	return c
}

// ParseCompile builds on the context AST using input from the current scanner until
// EOF.
func (c *compiler) ParseCompile() error {
	for {
		err := c.ParseCompileNext()
		if err != nil {
			if err == io.EOF {
				//c = nil
				return nil
			}
			return err
		}
	}
}

// Parse builds on the context program using input from the current scanner
// until the end of a whole expression is found. It returns io.EOF on EOF.
func (c *compiler) ParseCompileNext() error {
	ctx := c.ctx
	if c.drop {
		c.push(opDrop)
	}
	var eof bool
	expr, err := c.p.Next()
	//fmt.Printf("expr: %v\n", expr)
	if err != nil {
		eof = err == io.EOF
		if !eof {
			ctx.compiler = newCompiler(ctx)
			return err
		}
	}
	err = c.doExpr(expr, 0)
	if err != nil {
		ctx.compiler = newCompiler(ctx)
		return err
	}
	c.drop = nonEmpty(expr)
	if eof {
		return io.EOF
	}
	return nil
}

// push pushes a zero-argument opcode to the current's scope code.
func (c *compiler) push(opc opcode) {
	lc := c.scope()
	if lc != nil {
		lc.Body = append(lc.Body, opc)
		lc.Pos = append(lc.Pos, c.pos)
	} else {
		c.ctx.gCode.Body = append(c.ctx.gCode.Body, opc)
		c.ctx.gCode.Pos = append(c.ctx.gCode.Pos, c.pos)
		c.ctx.gCode.last = len(c.ctx.gCode.Body) - 1
	}
}

// push pushes a one-argument opcode to the current's scope code.
func (c *compiler) push2(op, arg opcode) {
	lc := c.scope()
	if lc != nil {
		lc.Body = append(lc.Body, op, arg)
		lc.Pos = append(lc.Pos, c.pos, c.pos)
	} else {
		c.ctx.gCode.Body = append(c.ctx.gCode.Body, op, arg)
		c.ctx.gCode.Pos = append(c.ctx.gCode.Pos, c.pos, c.pos)
		c.ctx.gCode.last = len(c.ctx.gCode.Body) - 2
	}
}

// push pushes a two-argument opcode to the current's scope code.
func (c *compiler) push3(op, arg1, arg2 opcode) {
	lc := c.scope()
	if lc != nil {
		lc.Body = append(lc.Body, op, arg1, arg2)
		lc.Pos = append(lc.Pos, c.pos, c.pos, c.pos)
	} else {
		c.ctx.gCode.Body = append(c.ctx.gCode.Body, op, arg1, arg2)
		c.ctx.gCode.Pos = append(c.ctx.gCode.Pos, c.pos, c.pos, c.pos)
		c.ctx.gCode.last = len(c.ctx.gCode.Body) - 2
	}
}

// applyN pushes an apply opcode for the given number of arguments. It does
// nothing for zero.
func (c *compiler) applyN(n int) {
	switch {
	case n == 1:
		c.push(opApply)
	case n == 2:
		c.push(opApply2)
	case n > 2:
		c.push2(opApplyN, opcode(n))
	}
}

// applyAtN calls applyN, but recording custom position information.
func (c *compiler) applyAtN(pos int, n int) {
	opos := c.pos
	c.pos = pos
	c.applyN(n)
	c.pos = opos
}

// errorf returns a formatted error.
func (c *compiler) errorf(format string, a ...interface{}) error {
	c.ctx.errPos = append(c.ctx.errPos, position{Filename: c.ctx.fname, Pos: c.pos})
	return fmt.Errorf(format, a...)
}

// perrorf returns a formatted error with custom position information.
func (c *compiler) perrorf(pos int, format string, a ...interface{}) error {
	c.ctx.errPos = append(c.ctx.errPos, position{Filename: c.ctx.fname, Pos: pos})
	return fmt.Errorf(format, a...)
}

// scope returns the current lambda's scope, or nil.
func (c *compiler) scope() *lambdaCode {
	if len(c.scopeStack) == 0 {
		return nil
	}
	return c.scopeStack[len(c.scopeStack)-1]
}

// body returns the current scope's code.
func (c *compiler) body() []opcode {
	lc := c.scope()
	if lc != nil {
		return lc.Body
	}
	return c.ctx.gCode.Body
}

func bool2int(b bool) (i int) {
	if b {
		i = 1
	}
	return
}

func (c *compiler) doExpr(e expr, n int) error {
	switch e := e.(type) {
	case exprs:
		return c.doExprs(e, n)
	case *astToken:
		err := c.doToken(e, n)
		if err != nil {
			return err
		}
	case *astReturn:
		// n == 0 is normally ensured by construction for returns.
		if n > 0 {
			panic(c.errorf("doExpr: astReturn: n > 0 (%d)", n))
		}
		err := c.doExpr(e.Expr, 0)
		if err != nil {
			return err
		}
		if e.OnError {
			c.push(opTry)
		} else {
			c.push(opReturn)
		}
		return nil
	case *astAssign:
		err := c.doAssign(e, n)
		if err != nil {
			return err
		}
	case *astListAssign:
		err := c.doListAssign(e, n)
		if err != nil {
			return err
		}
	case *astAssignOp:
		err := c.doAssignOp(e, n)
		if err != nil {
			return err
		}
	case *astAssignAmendOp:
		err := c.doAssignAmendOp(e, n)
		if err != nil {
			return err
		}
	case *astAssignDeepAmendOp:
		err := c.doAssignDeepAmendOp(e, n)
		if err != nil {
			return err
		}
	case *astDerivedVerb:
		err := c.doDerivedVerb(e, n)
		if err != nil {
			return err
		}
	case *astStrand:
		c.pos = e.Pos
		err := c.doStrand(e, n)
		if err != nil {
			return err
		}
	case *astParen:
		err := c.doParen(e, n)
		if err != nil {
			return err
		}
	case *astApply2:
		return c.doApply2(e, n)
	case *astApply2Adverb:
		return c.doApply2Adverb(e, n)
	case *astApplyN:
		return c.doApplyN(e, n)
	case *astList:
		return c.doList(e, n)
	case *astSeq:
		return c.doSeq(e, n)
	case *astLambda:
		err := c.doLambda(e, n)
		if err != nil {
			return err
		}
	default:
		panic(c.errorf("unknown expr type"))
	}
	return nil
}

func (c *compiler) doExprs(es exprs, n int) error {
	for i, e := range es {
		err := c.doExpr(e, bool2int(i > 0))
		if err != nil {
			return err
		}
	}
	if len(es) == 0 {
		c.push(opNil)
		return nil
	}
	c.applyN(n)
	return nil
}

func (c *compiler) doToken(tok *astToken, n int) error {
	c.pos = tok.Pos
	switch tok.Type {
	case astNUMBER:
		x, err := parseNumber(tok.Text)
		if err != nil {
			return c.errorf("number: %v", err)
		}
		if n > 0 {
			return c.errorf("type n cannot be applied")
		}
		if x.kind == valInt && x.n <= math.MaxInt32 && x.n >= math.MinInt32 {
			c.push2(opInt, opcode(int32(x.n)))
		} else {
			id := c.ctx.storeConst(x)
			c.push2(opConst, opcode(id))
		}
		return nil
	case astSTRING:
		s, err := strconv.Unquote(tok.Text)
		if err != nil {
			return c.errorf("string: %v", err)
		}
		id := c.ctx.storeConst(NewS(s))
		c.push2(opConst, opcode(id))
		c.applyN(n)
		return nil
	case astIDENT:
		// read or apply, not assign
		if c.scope() == nil {
			// global scope: global variable
			c.doGlobal(tok, n)
			return nil
		}
		// local scope: argument, local or global variable
		c.doLocal(tok, n)
		return nil
	case astDYAD:
		c.doVariadic(tok, n)
		return nil
	case astMONAD:
		c.doVariadic(tok, n)
		return nil
	case astREGEXP:
		r, err := regexp.Compile(tok.Text)
		if err != nil {
			return c.errorf("rx// : %v", err)
		}
		id := c.ctx.storeConst(NewV(&rx{Regexp: r}))
		c.push2(opConst, opcode(id))
		c.applyN(n)
		return nil
	case astEMPTYLIST:
		c.push2(opConst, opcode(constAV))
		c.applyN(n)
		return nil
	default:
		// should not happen
		return c.errorf("unexpected token type: %v", tok.Type)
	}
}

func parseNumber(s string) (V, error) {
	switch s {
	case "0n":
		s = "NaN"
	case "0w":
		s = "Inf"
	case "-0w":
		s = "-Inf"
	}
	i, errI := strconv.ParseInt(s, 0, 0)
	if errI == nil {
		return NewI(i), nil
	}
	f, errF := strconv.ParseFloat(s, 64)
	if errF == nil {
		return NewF(f), nil
	}
	err := errF.(*strconv.NumError)
	return V{}, err.Err
}

func (c *compiler) doGlobal(tok *astToken, n int) {
	id := c.ctx.global(tok.Text)
	switch n {
	case 0:
		c.push2(opGlobal, opcode(id))
	case 1:
		c.push2(opApplyGlobal, opcode(id))
	default:
		c.push3(opApplyNGlobal, opcode(id), opcode(n))
	}
}

func (c *compiler) doLocal(tok *astToken, n int) {
	lc := c.scope()
	local, ok := lc.local(tok.Text)
	if ok {
		c.push2(opLocal, opArg)
		lc.opIdxLocal[len(lc.Body)-1] = local
		c.applyN(n)
		return
	}
	c.doGlobal(tok, n)
}

func (c *compiler) doAdverb(tok *astToken) {
	v := c.parseVariadic(tok.Text)
	opos := c.pos
	c.pos = tok.Pos
	c.push2(opDerive, opcode(v))
	c.pos = opos
}

func (c *compiler) doVariadic(tok *astToken, n int) {
	c.doVariadicAt(tok.Text, tok.Pos, n)
}

func (c *compiler) doVariadicAt(s string, pos, n int) {
	// tok.Type either MONAD, DYAD or ADVERB
	v := c.parseVariadic(s)
	opos := c.pos
	c.pos = pos
	c.pushVariadic(v, n)
	c.pos = opos
}

func (c *compiler) pushVariadic(v variadic, n int) {
	switch n {
	case 0:
		c.push2(opVariadic, opcode(v))
	case 1:
		c.push2(opApplyV, opcode(v))
	case 2:
		c.push2(opApply2V, opcode(v))
	default:
		c.push3(opApplyNV, opcode(v), opcode(n))
	}
}

func isLeftArg(e expr) bool {
	switch e := e.(type) {
	case *astToken:
		switch e.Type {
		case astDYAD:
			return false
		case astMONAD:
			return false
		}
	case *astDerivedVerb:
		return false
	}
	return true
}

func (c *compiler) doAssign(e *astAssign, n int) error {
	err := c.doExpr(e.Right, 0)
	if err != nil {
		return err
	}
	lc := c.scope()
	if lc == nil || e.Global {
		id := c.ctx.global(e.Name)
		c.push2(opAssignGlobal, opcode(id))
		c.applyN(n)
		return nil
	}
	local, ok := lc.local(e.Name)
	if !ok {
		local = lambdaLocal{Type: localVar, ID: lc.nVars}
		lc.locals[e.Name] = local
		lc.nVars++
	}
	c.push2(opAssignLocal, opArg)
	lc.opIdxLocal[len(lc.Body)-1] = local
	c.applyN(n)
	return nil
}

func (c *compiler) doListAssign(e *astListAssign, n int) error {
	err := c.doExpr(e.Right, 0)
	if err != nil {
		return err
	}
	lc := c.scope()
	if lc == nil || e.Global {
		lid := len(c.ctx.gAssignLists)
		idList := make([]int, len(e.Names))
		for i, name := range e.Names {
			id := c.ctx.global(name)
			idList[i] = id
		}
		c.ctx.gAssignLists = append(c.ctx.gAssignLists, idList)
		c.push2(opListAssignGlobal, opcode(lid))
		c.applyN(n)
		return nil
	}
	lid := len(lc.assignLists)
	localList := make([]lambdaLocal, len(e.Names))
	for i, name := range e.Names {
		local, ok := lc.local(name)
		if !ok {
			local = lambdaLocal{Type: localVar, ID: lc.nVars}
			lc.locals[name] = local
			lc.nVars++
		}
		localList[i] = local
	}
	lc.assignLists = append(lc.assignLists, localList)
	c.push2(opListAssignLocal, opcode(lid))
	c.applyN(n)
	return nil
}

func (c *compiler) doAssignOp(e *astAssignOp, n int) error {
	err := c.doExpr(e.Right, 0)
	if err != nil {
		return err
	}
	lc := c.scope()
	if lc == nil || e.Global {
		id, ok := c.ctx.gIDs[e.Name]
		if !ok {
			if lc == nil {
				return c.perrorf(e.Pos,
					"undefined global in assignement operation: %s", e.Name)
			}
			id = c.ctx.global(e.Name)
		}
		c.push2(opGlobalLast, opcode(id))
		c.doVariadicAt(e.Dyad, e.Pos-1, 2)
		c.push2(opAssignGlobal, opcode(id))
		c.applyN(n)
		return nil
	}
	local, ok := lc.local(e.Name)
	if !ok {
		return c.perrorf(e.Pos,
			"undefined local in assignement operation: %s", e.Name)
	}
	c.push2(opLocalLast, opArg)
	lc.opIdxLocal[len(lc.Body)-1] = local
	c.doVariadicAt(e.Dyad, e.Pos-1, 2)
	c.push2(opAssignLocal, opArg)
	lc.opIdxLocal[len(lc.Body)-1] = local
	c.applyN(n)
	return nil
}

func (c *compiler) doAssignAmendOp(e *astAssignAmendOp, n int) error {
	err := c.doExpr(e.Right, 0)
	if err != nil {
		return err
	}
	c.doVariadicAt(e.Dyad, e.Pos-1, 0)
	if !nonEmpty(e.Indices) {
		return c.perrorf(e.Pos, "no indices in assignement amend operation")
	}
	err = c.doExpr(e.Indices, 0)
	if err != nil {
		return err
	}
	lc := c.scope()
	if lc == nil || e.Global {
		id, ok := c.ctx.gIDs[e.Name]
		if !ok {
			if lc == nil {
				return c.perrorf(e.Pos,
					"undefined global in assignement amend operation: %s", e.Name)
			}
			id = c.ctx.global(e.Name)
		}
		c.push2(opGlobalLast, opcode(id))
		c.doVariadicAt("@", e.Pos-1, 4)
		c.push2(opAssignGlobal, opcode(id))
		c.applyN(n)
		return nil
	}
	local, ok := lc.local(e.Name)
	if !ok {
		return c.perrorf(e.Pos,
			"undefined local in assignement amend operation: %s", e.Name)
	}
	c.push2(opLocalLast, opArg)
	lc.opIdxLocal[len(lc.Body)-1] = local
	c.doVariadicAt("@", e.Pos-1, 4)
	c.push2(opAssignLocal, opArg)
	lc.opIdxLocal[len(lc.Body)-1] = local
	c.applyN(n)
	return nil
}

func (c *compiler) doAssignDeepAmendOp(e *astAssignDeepAmendOp, n int) error {
	err := c.doExpr(e.Right, 0)
	if err != nil {
		return err
	}
	c.doVariadicAt(e.Dyad, e.Pos-1, 0)
	err = c.doList(e.Indices, 0)
	if err != nil {
		return err
	}
	lc := c.scope()
	if lc == nil || e.Global {
		id, ok := c.ctx.gIDs[e.Name]
		if !ok {
			if lc == nil {
				return c.perrorf(e.Pos,
					"undefined global in assignement amend operation: %s", e.Name)
			}
			id = c.ctx.global(e.Name)
		}
		c.push2(opGlobalLast, opcode(id))
		c.doVariadicAt(".", e.Pos-1, 4)
		c.push2(opAssignGlobal, opcode(id))
		c.applyN(n)
		return nil
	}
	local, ok := lc.local(e.Name)
	if !ok {
		return c.perrorf(e.Pos,
			"undefined local in assignement amend operation: %s", e.Name)
	}
	c.push2(opLocalLast, opArg)
	lc.opIdxLocal[len(lc.Body)-1] = local
	c.doVariadicAt(".", e.Pos-1, 4)
	c.push2(opAssignLocal, opArg)
	lc.opIdxLocal[len(lc.Body)-1] = local
	c.applyN(n)
	return nil
}

func (c *compiler) parseVariadic(s string) variadic {
	v, ok := c.ctx.vNames[s]
	if !ok {
		panic("unknown variadic op: " + s)
	}
	return v
}

func (c *compiler) doDerivedVerb(dv *astDerivedVerb, n int) error {
	if dv.Verb == nil {
		c.doVariadic(dv.Adverb, n)
		return nil
	}
	err := c.doExpr(dv.Verb, 0)
	if err != nil {
		return err
	}
	c.doAdverb(dv.Adverb)
	c.applyAtN(dv.Adverb.Pos, n)
	return nil
}

func (c *compiler) doStrand(st *astStrand, n int) error {
	a := make([]V, 0, len(st.Lits))
	for _, tok := range st.Lits {
		switch tok.Type {
		case astNUMBER:
			x, err := parseNumber(tok.Text)
			if err != nil {
				c.pos = tok.Pos
				return c.errorf("number: %v", err)
			}
			a = append(a, x)
		case astSTRING:
			s, err := strconv.Unquote(tok.Text)
			if err != nil {
				c.pos = tok.Pos
				return c.errorf("string: %v", err)
			}
			a = append(a, NewS(s))
		}
	}
	r := Canonical(NewAV(a))
	r.InitRC()
	id := c.ctx.storeConst(r)
	c.pos = st.Pos
	c.push2(opConst, opcode(id))
	c.applyN(n)
	return nil
}

func (c *compiler) doParen(p *astParen, n int) error {
	err := c.doExpr(p.Expr, 0)
	if err != nil {
		return err
	}
	c.applyAtN(p.EndPos, n)
	return err
}

func (c *compiler) doLambda(b *astLambda, n int) error {
	body := b.Body
	args := b.Args
	lc := &lambdaCode{
		locals:     map[string]lambdaLocal{},
		opIdxLocal: map[int]lambdaLocal{},
	}
	c.scopeStack = append(c.scopeStack, lc)
	if len(args) != 0 {
		err := c.doLambdaArgs(args)
		if err != nil {
			return err
		}
	}
	for i, expr := range body {
		err := c.doExpr(expr, 0)
		if err != nil {
			return err
		}
		if i < len(body)-1 && nonEmpty(expr) {
			c.push(opDrop)
		}
	}
	c.scopeStack = c.scopeStack[:len(c.scopeStack)-1]
	id := len(c.ctx.lambdas)
	c.ctx.lambdas = append(c.ctx.lambdas, lc)
	lc.StartPos = b.StartPos
	lc.Source = c.ctx.sources[c.ctx.fname][lc.StartPos:b.EndPos]
	lc.Filename = c.ctx.fname
	c.ctx.resolveLambda(lc)
	c.ctx.analyzeLambdaLiveness(lc)
	c.push2(opLambda, opcode(id))
	c.applyAtN(b.EndPos, n)
	return nil
}

func (c *compiler) doLambdaArgs(args []string) error {
	lc := c.scope()
	lc.namedArgs = true
	for i, arg := range args {
		_, ok := lc.locals[arg]
		if ok {
			return c.errorf("name %s appears twice in argument list", arg)
		}
		lc.locals[arg] = lambdaLocal{
			Type: localArg,
			ID:   i,
		}
	}
	return nil
}

func (ctx *Context) resolveLambda(lc *lambdaCode) {
	nargs := 0
	nlocals := 0
	for _, local := range lc.locals {
		nlocals++
		if local.Type == localArg {
			nargs++
		}
	}
	if nargs == 0 {
		// All lambdas have at least one argument, even if not used.
		nlocals++
		nargs = 1
	}
	nvars := lc.nVars
	lc.Rank = nargs
	names := make([]string, nlocals)
	getID := func(local lambdaLocal) int {
		switch local.Type {
		case localArg:
			return local.ID + nvars
		case localVar:
			return local.ID
		default:
			panic(fmt.Sprintf("unknown local type: %d", local.Type))
		}
	}
	for k, local := range lc.locals {
		names[getID(local)] = k
	}
	lc.Names = names
	if len(lc.assignLists) > 0 {
		lc.AssignLists = make([][]int32, len(lc.assignLists))
	}
	for ip := 0; ip < len(lc.Body); {
		op := lc.Body[ip]
		ip++
		switch op {
		case opLocal, opLocalLast:
			lc.Body[ip] = opcode(getID(lc.opIdxLocal[ip]))
		case opAssignLocal:
			lc.Body[ip] = opcode(getID(lc.opIdxLocal[ip]))
		case opListAssignLocal:
			i := lc.Body[ip]
			locals := lc.assignLists[i]
			ids := make([]int32, len(locals))
			for j, local := range locals {
				ids[j] = int32(getID(local))
			}
			lc.AssignLists[i] = ids
		}
		ip += op.argc()
	}
	// free unused data after this resolving pass
	lc.locals = nil
	lc.opIdxLocal = nil
	lc.assignLists = nil
}

type lastUse struct {
	branch int32 // branch number
	bn     int32 // block number of last use
	opIdx  int32 // opcode index of last use
}

func (ctx *Context) analyzeLambdaLiveness(lc *lambdaCode) {
	// We do a simple and fast one-pass analysis for now, to optimize
	// common cases, handling only def-use in the same basic block.
	// Branches with uneven use of variables might lead to some refcounts
	// not being decreased as much as possible in all paths, leading to
	// some extra cloning. The analysis still gives quite some good results
	// for little complexity.
	//
	// Branching in goal is limited. There are five kinds of cases, all
	// going forward (no loops):
	//
	// ?[if;then;else] gives ...jumpFalse #then; ...jump #else;
	// and[x;y;z] gives ...jumpFalse #y+#z; jumpFalse #z
	// or[x;y;z] gives ...jumpTrue #y+#z; jumpTrue #z
	// :x gives opReturn
	//
	// Errors act kinda like return and both are ignored. This means some
	// refcounts might not be decreased if the early path is taken. Note
	// that this only means some more cloning might happen, it does not
	// leak memory, as memory management is handled by Go's GC
	// independently. Refcount in goal is only used as an optimization to
	// reduce cloning.
	lc.lastUses = make([]lastUse, len(lc.Names))
	// bn is the basic-block number, starting from 1.
	var bn int32 = 1
	// The branch number is similar to the basic-block number, but starts
	// from zero and it gets reduced on join points by the number of jumps
	// pointing to them.  In particular, this means that the branch number
	// is zero when we are not in a branch, and positive otherwise.
	var branch int32
	for ip := 0; ip < len(lc.Body); {
		op := lc.Body[ip]
		if lc.joinPoints != nil && lc.joinPoints[ip] > 0 {
			branch -= lc.joinPoints[ip]
			bn++
		}
		ip++
		switch op {
		case opJumpFalse, opJumpTrue, opJump:
			if lc.joinPoints == nil {
				lc.joinPoints = make([]int32, len(lc.Body)+1)
			}
			branch++
			bn++
			lc.joinPoints[ip+int(lc.Body[ip])]++
		case opReturn:
			for _, lu := range lc.lastUses {
				if branch > 0 && lu.bn != bn || lu.bn == 0 {
					continue
				}
				lc.Body[lu.opIdx] = opLocalLast
			}
		case opLocal, opLocalLast:
			i := lc.Body[ip]
			lc.lastUses[i].opIdx = int32(ip) - 1
			lc.lastUses[i].bn = bn
			lc.lastUses[i].branch = branch
		case opAssignLocal:
			i := lc.Body[ip]
			lu := lc.lastUses[i]
			if branch > 0 && lu.bn != bn || lu.bn == 0 {
				break
			}
			lc.Body[lu.opIdx] = opLocalLast
		case opListAssignLocal:
			ids := lc.AssignLists[lc.Body[ip]]
			for _, i := range ids {
				lu := lc.lastUses[i]
				if branch > 0 && lu.bn != bn || lu.bn == 0 {
					continue
				}
				lc.Body[lu.opIdx] = opLocalLast
			}
		}
		ip += op.argc()
	}
	for i, lu := range lc.lastUses {
		if lu.bn == 0 {
			if i >= lc.nVars {
				lc.UnusedArgs = append(lc.UnusedArgs, int32(len(lc.Names)-i-1))
			}
			continue
		}
		if i >= lc.nVars {
			lc.UsedArgs = append(lc.UsedArgs, int32(len(lc.Names)-i-1))
		}
		lc.Body[lu.opIdx] = opLocalLast
	}
	// free unused data after this pass
	lc.joinPoints = nil
	lc.lastUses = nil
}

func (c *compiler) doApply2(a *astApply2, n int) error {
	switch v := a.Verb.(type) {
	case *astToken:
		// e.Type == astDYAD
		switch v.Text {
		case "and":
			aN := &astApplyN{
				Verb: a.Verb,
				Args: []expr{a.Left, a.Right},
			}
			err := c.doAnd(aN, n, v.Pos)
			if err != nil {
				return err
			}
			return nil
		case "or":
			aN := &astApplyN{
				Verb: a.Verb,
				Args: []expr{a.Left, a.Right},
			}
			err := c.doOr(aN, n, v.Pos)
			if err != nil {
				return err
			}
			return nil
		}
	}
	err := c.doExpr(a.Right, 0)
	if err != nil {
		return err
	}
	switch e := a.Verb.(type) {
	case *astToken:
		// e.Type == astDYAD
		err = c.doExpr(a.Left, 0)
		if err != nil {
			return err
		}
		c.doVariadic(e, 2)
	default:
		panic(fmt.Sprintf("bad verb type for apply2: %v", e))
	}
	c.applyN(n)
	return nil
}

func (c *compiler) doApply2Adverb(a *astApply2Adverb, n int) error {
	err := c.doExpr(a.Right, 0)
	if err != nil {
		return err
	}
	switch e := a.Verb.(type) {
	case *astDerivedVerb:
		err = c.doExpr(a.Left, 0)
		if err != nil {
			return err
		}
		err = c.doExpr(e.Verb, 0)
		if err != nil {
			return err
		}
		c.doVariadic(e.Adverb, 3)
	default:
		panic(fmt.Sprintf("bad verb type for apply2adverb: %v", e))
	}
	c.applyN(n)
	return nil
}

func (c *compiler) doApplyN(a *astApplyN, n int) error {
	switch v := a.Verb.(type) {
	case *astToken:
		if v.Type != astDYAD {
			break
		}
		switch v.Text {
		case "?":
			if len(a.Args) >= 3 {
				err := c.doCond(a, n, v.Pos)
				if err != nil {
					return err
				}
				return nil
			}
		case "and":
			err := c.doAnd(a, n, v.Pos)
			if err != nil {
				return err
			}
			return nil
		case "or":
			err := c.doOr(a, n, v.Pos)
			if err != nil {
				return err
			}
			return nil
		}
	}
	for i := len(a.Args) - 1; i >= 0; i-- {
		ei := a.Args[i]
		err := c.doExpr(ei, 0)
		if err != nil {
			return err
		}
	}
	err := c.doExpr(a.Verb, len(a.Args))
	if err != nil {
		return err
	}
	c.applyAtN(a.EndPos, n)
	return nil
}

func (c *compiler) doCond(a *astApplyN, n, pos int) error {
	body := a.Args
	if len(body)%2 != 1 {
		return c.errorf("conditional ?[if;then;else] with even number of statements")
	}
	cond := body[0]
	if !nonEmpty(cond) {
		return c.perrorf(pos, "?[if;then;else] : empty condition")
	}
	err := c.doExpr(cond, 0)
	if err != nil {
		return err
	}
	c.push2(opJumpFalse, opArg)
	jmpCond := len(c.body()) - 1
	jumpsEnd := []int{}
	jumpsElse := []int{}
	jumpsCond := []int{}
	for i := 1; i < len(body); i += 2 {
		c.push(opDrop)
		then := body[i]
		if !nonEmpty(then) {
			return c.perrorf(pos, "?[if;then;else] : empty then (%d-th)", i+1)
		}
		err := c.doExpr(then, 0)
		if err != nil {
			return err
		}
		c.push2(opJump, opArg)
		jumpsEnd = append(jumpsEnd, len(c.body())-1)
		jumpsElse = append(jumpsElse, len(c.body()))
		c.push(opDrop)
		elseCond := body[i+1]
		if !nonEmpty(elseCond) {
			return c.perrorf(pos, "?[if;then;else] : empty cond (%d-th)", i+2)
		}
		err = c.doExpr(elseCond, 0)
		if err != nil {
			return err
		}
		if i+1 < len(body)-1 {
			c.push2(opJumpFalse, opArg)
			jumpsCond = append(jumpsCond, len(c.body())-1)
		}
	}
	c.body()[jmpCond] = opcode(jumpsElse[0] - jmpCond)
	for i, offset := range jumpsCond {
		c.body()[offset] = opcode(jumpsElse[i+1] - offset)
	}
	end := len(c.body())
	for _, offset := range jumpsEnd {
		c.body()[offset] = opcode(end - offset)
	}
	c.applyN(n)
	return nil
}

func (c *compiler) doAnd(a *astApplyN, n int, pos int) error {
	body := a.Args
	jumpsEnd := []int{}
	for i, ei := range body {
		if i > 0 {
			c.push(opDrop)
		}
		if !nonEmpty(ei) {
			return c.perrorf(pos, "and[...] : empty argument (%d-th)", i+1)
		}
		err := c.doExpr(ei, 0)
		if err != nil {
			return err
		}
		if i < len(body)-1 {
			c.push2(opJumpFalse, opArg)
			jumpsEnd = append(jumpsEnd, len(c.body())-1)
		}
	}
	end := len(c.body())
	for _, offset := range jumpsEnd {
		c.body()[offset] = opcode(end - offset)
	}
	c.applyN(n)
	return nil
}

func (c *compiler) doOr(b *astApplyN, n int, pos int) error {
	body := b.Args
	jumpsEnd := []int{}
	for i, ei := range body {
		if i > 0 {
			c.push(opDrop)
		}
		if !nonEmpty(ei) {
			return c.perrorf(pos, "or[...] : empty argument (%d-th)", i+1)
		}
		err := c.doExpr(ei, 0)
		if err != nil {
			return err
		}
		if i < len(body)-1 {
			c.push2(opJumpTrue, opArg)
			jumpsEnd = append(jumpsEnd, len(c.body())-1)
		}
	}
	end := len(c.body())
	for _, offset := range jumpsEnd {
		c.body()[offset] = opcode(end - offset)
	}
	c.applyN(n)
	return nil
}

func (c *compiler) doSeq(b *astSeq, n int) error {
	body := b.Body
	for i, ei := range body {
		err := c.doExpr(ei, 0)
		if err != nil {
			return err
		}
		if i < len(body)-1 && nonEmpty(ei) {
			c.push(opDrop)
		}
	}
	c.applyN(n)
	return nil
}

func (c *compiler) doList(l *astList, n int) error {
	body := l.Args
	for i := len(body) - 1; i >= 0; i-- {
		ei := body[i]
		if nonEmpty(ei) {
			err := c.doExpr(ei, 0)
			if err != nil {
				return err
			}
		} else {
			c.push2(opVariadic, opcode(vMultiply))
		}
	}
	c.pushVariadic(vList, len(body))
	c.applyN(n)
	return nil
}
