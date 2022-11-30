package goal

import (
	"fmt"
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
	Body      []opcode
	Pos       []int
	Names     []string
	Rank      int
	NamedArgs bool
	Locals    map[string]lambdaLocal // arguments and variables
	Source    string
	Filename  string
	StartPos  int
	EndPos    int

	opIdxLocal map[int]lambdaLocal // opcode index -> local variable
	nVars      int
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
	param, ok := l.Locals[s]
	if ok {
		return param, true
	}
	if !l.NamedArgs && len(s) == 1 {
		switch r := rune(s[0]); r {
		case 'x', 'y', 'z':
			id := r - 'x'
			arg := lambdaLocal{Type: localArg, ID: int(id)}
			l.Locals[s] = arg
			return arg, true
		}
	}
	return lambdaLocal{}, false
}

// programString returns a string representation of the compiled program and
// relevant data.
func (ctx *Context) programString() string {
	sb := &strings.Builder{}
	fmt.Fprintln(sb, "---- Compiled program -----")
	fmt.Fprintln(sb, "Instructions:")
	fmt.Fprint(sb, ctx.opcodesString(ctx.gCode.Body, nil))
	fmt.Fprintln(sb, "Globals:")
	for id, name := range ctx.gNames {
		fmt.Fprintf(sb, "\t%s\t%d\n", name, id)
	}
	fmt.Fprintln(sb, "Constants:")
	for id, ci := range ctx.constants {
		fmt.Fprintf(sb, "\t%d\t%s\n", id, ci.Sprint(ctx))
	}
	for id, lc := range ctx.lambdas {
		fmt.Fprintf(sb, "---- Lambda %d (Rank: %d) -----\n", id, lc.Rank)
		fmt.Fprintf(sb, "%s", ctx.lambdaString(lc))
	}
	return sb.String()
}

func (ctx *Context) lambdaString(lc *lambdaCode) string {
	sb := &strings.Builder{}
	fmt.Fprintln(sb, "Instructions:")
	fmt.Fprint(sb, ctx.opcodesString(lc.Body, lc))
	fmt.Fprintln(sb, "Locals:")
	for i, name := range lc.Names {
		fmt.Fprintf(sb, "\t%s\t%d\n", name, i)
	}
	return sb.String()
}

// compiler incrementally builds a semi-resolved program from a parsed expr.
type compiler struct {
	ctx        *Context      // main execution and compilation context
	p          *parser       // parsing into text-based non-resolved AST
	arglist    bool          // whether current expression has an argument list
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
			if _, ok := err.(errEOF); ok {
				//c = nil
				return nil
			}
			return err
		}
	}
}

// Parse builds on the context program using input from the current scanner
// until the end of a whole expression is found. It returns ErrEOF on EOF.
func (c *compiler) ParseCompileNext() error {
	ctx := c.ctx
	if c.drop {
		c.push(opDrop)
	}
	var eof bool
	expr, err := c.p.Next()
	//fmt.Printf("expr: %v\n", expr)
	if err != nil {
		_, eof = err.(errEOF)
		if !eof {
			c.ctx.errPos = append(c.ctx.errPos,
				position{Filename: c.ctx.fname, Pos: c.p.token.Pos})
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
		return errEOF{}
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
		id := c.ctx.storeConst(NewV(x))
		c.push2(opConst, opcode(id))
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
		if n == 1 && tok.Text == ":" {
			c.push(opReturn)
			return nil
		}
		return c.doVariadic(tok, n)
	case astMONAD:
		return c.doVariadic(tok, n)
	case astEMPTYLIST:
		id := c.ctx.storeConst(NewV(AV{}))
		c.push2(opConst, opcode(id))
		c.applyN(n)
		return nil
	default:
		// should not happen
		return c.errorf("unexpected token type: %v", tok.Type)
	}
}

func parseNumber(s string) (Value, error) {
	switch s {
	case "0w":
		s = "Inf"
	case "-0w":
		s = "-Inf"
	}
	i, errI := strconv.ParseInt(s, 0, 0)
	if errI == nil {
		return I(i), nil
	}
	f, errF := strconv.ParseFloat(s, 64)
	if errF == nil {
		return F(f), nil
	}
	err := errF.(*strconv.NumError)
	return nil, err.Err
}

func (c *compiler) doGlobal(tok *astToken, n int) {
	id := c.ctx.global(tok.Text)
	c.push2(opGlobal, opcode(id))
	c.applyN(n)
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

func (c *compiler) doVariadic(tok *astToken, n int) error {
	// tok.Type either MONAD, DYAD or ADVERB
	v := c.parseBuiltin(tok.Text)
	opos := c.pos
	c.pos = tok.Pos
	c.pushVariadic(v, n)
	c.pos = opos
	return nil
}

func (c *compiler) pushVariadic(v Variadic, n int) {
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

func getIdent(e expr) (*astToken, bool) {
	tok, ok := e.(*astToken)
	return tok, ok && tok.Type == astIDENT

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

func (c *compiler) doAssign(verbTok *astToken, left, right expr, n int) (bool, error) {
	var identTok *astToken
	switch left := left.(type) {
	case *astToken:
		if left.Type != astIDENT {
			return false, nil
		}
		identTok = left
	default:
		return false, nil
	}
	err := c.doExpr(right, 0)
	if err != nil {
		return false, err
	}
	lc := c.scope()
	if lc == nil || verbTok.Text == "::" {
		id := c.ctx.global(identTok.Text)
		c.push2(opAssignGlobal, opcode(id))
		return true, nil
	}
	local, ok := lc.local(identTok.Text)
	if ok {
		c.push2(opAssignLocal, opArg)
		lc.opIdxLocal[len(lc.Body)-1] = local
		return true, nil
	}
	local = lambdaLocal{Type: localVar, ID: lc.nVars}
	lc.Locals[identTok.Text] = local
	c.push2(opAssignLocal, opArg)
	lc.opIdxLocal[len(lc.Body)-1] = local
	lc.nVars++
	c.applyN(n)
	return true, nil
}

func (c *compiler) parseBuiltin(s string) Variadic {
	v, ok := c.ctx.vNames[s]
	if !ok {
		panic("unknown variadic op: " + s)
	}
	return v
}

func getVerb(e expr) (*astToken, bool) {
	tok, ok := e.(*astToken)
	return tok, ok && (tok.Type == astDYAD || tok.Type == astMONAD)

}

func (c *compiler) doDerivedVerb(dv *astDerivedVerb, n int) error {
	if dv.Verb == nil {
		return c.doVariadic(dv.Adverb, n)
	}
	err := c.doExpr(dv.Verb, 0)
	if err != nil {
		return err
	}
	err = c.doVariadic(dv.Adverb, 1)
	if err != nil {
		return err
	}
	c.applyN(n)
	return nil
}

func (c *compiler) doStrand(st *astStrand, n int) error {
	a := make(AV, 0, len(st.Lits))
	for _, tok := range st.Lits {
		switch tok.Type {
		case astNUMBER:
			x, err := parseNumber(tok.Text)
			if err != nil {
				c.pos = tok.Pos
				return c.errorf("number: %v", err)
			}
			a = append(a, NewV(x))
		case astSTRING:
			s, err := strconv.Unquote(tok.Text)
			if err != nil {
				c.pos = tok.Pos
				return c.errorf("string: %v", err)
			}
			a = append(a, NewS(s))
		}
	}
	id := c.ctx.storeConst(NewV(canonical(a)))
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
		Locals:     map[string]lambdaLocal{},
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
	lc.EndPos = b.EndPos
	lc.Source = c.ctx.sources[c.ctx.fname][lc.StartPos:lc.EndPos]
	lc.Filename = c.ctx.fname
	c.ctx.resolveLambda(lc)
	c.push2(opLambda, opcode(id))
	c.applyAtN(b.EndPos, n)
	return nil
}

func (c *compiler) doLambdaArgs(args []string) error {
	lc := c.scope()
	lc.NamedArgs = true
	for i, arg := range args {
		_, ok := lc.Locals[arg]
		if ok {
			return c.errorf("name %s appears twice in argument list", arg)
		}
		lc.Locals[arg] = lambdaLocal{
			Type: localArg,
			ID:   i,
		}
	}
	return nil
}

func (ctx *Context) resolveLambda(lc *lambdaCode) {
	nargs := 0
	nlocals := 0
	for _, local := range lc.Locals {
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
	nvars := nlocals - nargs
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
	for k, local := range lc.Locals {
		names[getID(local)] = k
	}
	lc.Names = names
	for ip := 0; ip < len(lc.Body); {
		op := lc.Body[ip]
		ip++
		switch op {
		case opLocal:
			lc.Body[ip] = opcode(getID(lc.opIdxLocal[ip]))
		case opAssignLocal:
			lc.Body[ip] = opcode(getID(lc.opIdxLocal[ip]))
		}
		ip += op.argc()
	}
	lc.opIdxLocal = nil
}

func (c *compiler) doApply2(a *astApply2, n int) error {
	switch v := a.Verb.(type) {
	case *astToken:
		if v.Type != astDYAD {
			break
		}
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
		case ":", "::":
			done, err := c.doAssign(v, a.Left, a.Right, n)
			if err != nil || done {
				return err
			}
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
	case *astDerivedVerb:
		err = c.doExpr(e.Verb, 0)
		if err != nil {
			return err
		}
		err = c.doExpr(a.Left, 0)
		if err != nil {
			return err
		}
		c.doVariadic(e.Adverb, 3)
	default:
		panic(fmt.Sprintf("bad verb type for apply2: %v", e))
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
				err := c.doCond(a, n)
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
		case ":", "::":
			switch len(a.Args) {
			case 1:
				// TODO: :[arg] when n > 0 ?
				if n == 0 && v.Text == ":" {
					err := c.doExpr(a.Args[0], 0)
					if err != nil {
						return err
					}
					c.push(opReturn)
					return nil
				}
			case 2:
				done, err := c.doAssign(v, a.Args[0], a.Args[1], n)
				if err != nil || done {
					return err
				}
			}
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

func (c *compiler) doCond(a *astApplyN, n int) error {
	body := a.Args
	if len(body)%2 != 1 {
		return c.errorf("conditional ?[if;then;else] with even number of statements")
	}
	cond := body[0]
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
		err := c.doExpr(then, 0)
		if err != nil {
			return err
		}
		c.push2(opJump, opArg)
		jumpsEnd = append(jumpsEnd, len(c.body())-1)
		jumpsElse = append(jumpsElse, len(c.body()))
		c.push(opDrop)
		elseCond := body[i+1]
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
		err := c.doExpr(ei, 0)
		if err != nil {
			return err
		}
	}
	c.push2(opVariadic, opcode(vList))
	c.push2(opApplyN, opcode(len(body)))
	c.applyN(n)
	return nil
}
