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
	for id, v := range ctx.constants {
		fmt.Fprintf(sb, "\t%d\t%s\n", id, v.Sprint(ctx))
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
	argc       int           // stack length for current sub-expression
	slen       int           // virtual stack length
	arglist    bool          // whether current expression has an argument list
	scopeStack []*lambdaCode // scope information
	pos        int           // last token position
	it         astIter       // exprs iterator
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
	exprs, err := c.p.Next()
	if err != nil {
		_, eof = err.(errEOF)
		if !eof {
			c.ctx.errPos = append(c.ctx.errPos,
				position{Filename: c.ctx.fname, Pos: c.p.token.Pos})
			ctx.compiler = newCompiler(ctx)
			return err
		}
	}
	slen := c.slen
	err = c.doExprs(exprs)
	c.drop = c.slen > slen
	if err != nil {
		ctx.compiler = newCompiler(ctx)
		return err
	}
	if eof {
		return errEOF{}
	}
	return nil
}

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
	switch opc {
	case opApply:
		// v v -> v
		c.slen--
		c.argc--
	case opApply2:
		// v v v -> v
		c.slen -= 2
		c.argc -= 2
	case opDrop:
		// v ->
		c.slen--
		c.argc--
	case opAssignLocal, opAssignGlobal, opReturn:
	default:
		// -> v
		c.slen++
		c.argc++
	}
}

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
	switch op {
	case opApplyN:
		// v ... v v -> v
		c.slen -= int(arg)
		c.argc -= int(arg)
	case opApplyV, opJump:
	case opApply2V, opJumpFalse:
		c.slen--
		c.argc--
	default:
		// -> v
		c.slen++
		c.argc++
	}
}

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
	switch op {
	case opApplyNV:
		// v ... v v -> v
		c.slen -= int(arg2 - 1)
		c.argc -= int(arg2 - 1)
	default:
		// -> v
		c.slen++
		c.argc++
	}
}

func (c *compiler) apply() {
	switch {
	case c.argc == 2:
		c.push(opApply)
	case c.argc == 3:
		c.push(opApply2)
	case c.argc > 3:
		c.push2(opApplyN, opcode(c.argc-1))
	}
}

func (c *compiler) applyAt(pos int) {
	opos := c.pos
	c.pos = pos
	c.apply()
	c.pos = opos
}

func (c *compiler) errorf(format string, a ...interface{}) error {
	// TODO: in case of error, read the file again to get from pos the line
	// and print the line that produced the error with some column marker.
	return fmt.Errorf("compile error: "+format, a...)
}

func (c *compiler) scope() *lambdaCode {
	if len(c.scopeStack) == 0 {
		return nil
	}
	return c.scopeStack[len(c.scopeStack)-1]
}

func (c *compiler) body() []opcode {
	lc := c.scope()
	if lc != nil {
		return lc.Body
	}
	return c.ctx.gCode.Body
}

func (c *compiler) doExprs(es exprs) error {
	slen := c.slen
	c.argc = 0
	it := c.it
	c.it = newAstIter(es)
	for c.it.Next() {
		e := c.it.Expr()
		err := c.doExpr(e)
		if err != nil {
			return err
		}
	}
	c.it = it
	if c.slen == slen {
		c.push(opNil)
	}
	return nil
}

func (c *compiler) doExpr(e expr) error {
	switch e := e.(type) {
	case *astToken:
		err := c.doToken(e)
		if err != nil {
			return err
		}
	case *astReturn:
		c.pos = e.Pos
		c.push(opReturn)
	case *astAdverbs:
		err := c.doAdverbs(e)
		if err != nil {
			return err
		}
	case *astStrand:
		c.pos = e.Pos
		err := c.doStrand(e)
		if err != nil {
			return err
		}
	case *astParenExpr:
		err := c.doParenExpr(e)
		if err != nil {
			return err
		}
	case *astBlock:
		oarglist := c.arglist
		c.arglist = false
		err := c.doBlock(e)
		c.arglist = oarglist
		if err != nil {
			return err
		}
	default:
		panic(c.errorf("unknown expr type"))
	}
	return nil
}

func (c *compiler) doToken(tok *astToken) error {
	c.pos = tok.Pos
	switch tok.Type {
	case astNUMBER:
		v, err := parseNumber(tok.Text)
		if err != nil {
			return err
		}
		if c.argc > 0 {
			return c.errorf("number atoms cannot be applied")
		}
		id := c.ctx.storeConst(v)
		c.push2(opConst, opcode(id))
		return nil
	case astSTRING:
		s, err := strconv.Unquote(tok.Text)
		if err != nil {
			return err
		}
		if c.argc > 0 {
			return c.errorf("string atoms cannot be applied")
		}
		id := c.ctx.storeConst(S(s))
		c.push2(opConst, opcode(id))
		return nil
	case astIDENT:
		// read or apply, not assign
		if c.scope() == nil {
			// global scope: global variable
			c.doGlobal(tok)
			return nil
		}
		// local scope: argument, local or global variable
		c.doLocal(tok)
		return nil
	case astVERB:
		return c.doVerb(tok)
	default:
		// should not happen
		return c.errorf("unexpected token type: %v", tok.Type)
	}
}

func parseNumber(s string) (V, error) {
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
	return nil, errF
}

func (c *compiler) doGlobal(tok *astToken) {
	id := c.ctx.global(tok.Text)
	c.push2(opGlobal, opcode(id))
	c.apply()
}

func (c *compiler) doLocal(tok *astToken) {
	lc := c.scope()
	local, ok := lc.local(tok.Text)
	if ok {
		c.push2(opLocal, opArg)
		lc.opIdxLocal[len(lc.Body)-1] = local
		c.apply()
		return
	}
	c.doGlobal(tok)
}

func (c *compiler) doVariadic(tok *astToken) error {
	v := parseBuiltin(tok.Text)
	opos := c.pos
	c.pos = tok.Pos
	c.pushVariadic(v)
	c.pos = opos
	return nil
}

func (c *compiler) pushVariadic(v Variadic) {
	switch c.argc {
	case 0:
		c.push2(opVariadic, opcode(v))
	case 1:
		c.push2(opApplyV, opcode(v))
	case 2:
		c.push2(opApply2V, opcode(v))
	default:
		c.push3(opApplyNV, opcode(v), opcode(c.argc))
	}
}

func (c *compiler) doVerb(tok *astToken) error {
	e := c.it.Peek()
	argc := c.argc
	if e == nil || c.arglist {
		return c.doVariadic(tok)
	}
	if !isLeftArg(e) {
		if argc == 0 {
			c.push(opNil)
			c.push(opNil)
		}
		return c.doVariadic(tok)
	}
	if identTok, ok := getIdent(e); ok {
		if c.doAssign(tok, identTok) {
			c.it.Next()
			return nil
		}
	}
	if argc == 0 {
		c.push(opNil)
	}
	c.it.Next()
	c.argc = 0
	err := c.doExpr(e)
	if err != nil {
		return err
	}
	c.argc = 2
	return c.doVariadic(tok)
}

func getIdent(e expr) (*astToken, bool) {
	tok, ok := e.(*astToken)
	return tok, ok && tok.Type == astIDENT

}

func isLeftArg(e expr) bool {
	switch e := e.(type) {
	case *astToken:
		switch e.Type {
		case astVERB:
			return false
		}
	case *astAdverbs:
		return false
	}
	return true
}

func (c *compiler) doAssign(verbTok, identTok *astToken) bool {
	if verbTok.Text != ":" && verbTok.Text != "::" || c.argc != 1 {
		return false
	}
	lc := c.scope()
	if lc == nil || verbTok.Text == "::" {
		id := c.ctx.global(identTok.Text)
		c.push2(opAssignGlobal, opcode(id))
		return true
	}
	local, ok := lc.local(identTok.Text)
	if ok {
		c.push2(opAssignLocal, opArg)
		lc.opIdxLocal[len(lc.Body)-1] = local
		return true
	}
	local = lambdaLocal{Type: localVar, ID: lc.nVars}
	lc.Locals[identTok.Text] = local
	c.push2(opAssignLocal, opArg)
	lc.opIdxLocal[len(lc.Body)-1] = local
	lc.nVars++
	return true
}

func parseBuiltin(s string) (verb Variadic) {
	switch s {
	case ":", "::":
		verb = vRight
	case "+":
		verb = vAdd
	case "-":
		verb = vSubtract
	case "*":
		verb = vMultiply
	case "%":
		verb = vDivide
	case "!":
		verb = vMod
	case "&":
		verb = vMin
	case "|":
		verb = vMax
	case "<":
		verb = vLess
	case ">":
		verb = vMore
	case "=":
		verb = vEqual
	case "~":
		verb = vMatch
	case ",":
		verb = vJoin
	case "^":
		verb = vWithout
	case "#":
		verb = vTake
	case "_":
		verb = vDrop
	case "$":
		verb = vCast
	case "?":
		verb = vFind
	case "@":
		verb = vApply
	case ".":
		verb = vApplyN
	case "in":
		verb = vIn
	case "sign":
		verb = vSign
	case "ocount":
		verb = vOCount
	case "icount":
		verb = vICount
	case "'":
		verb = vEach
	case "/":
		verb = vFold
	case "\\":
		verb = vScan
	default:
		panic("unknown variadic op: " + s)
	}
	return verb
}

func getVerb(e expr) (*astToken, bool) {
	tok, ok := e.(*astToken)
	return tok, ok && tok.Type == astVERB

}

func (c *compiler) doAdverbs(adverbs *astAdverbs) error {
	tok := &adverbs.Train[len(adverbs.Train)-1]
	ads := adverbs.Train[:len(adverbs.Train)-1]
	argc := c.argc
	e := c.it.Peek()
	if e == nil {
		if len(ads) > 0 {
			return errf("adverb train should modify a value")
		}
		c.push(opNil)
		return c.doVariadic(tok)
	}
	if argc == 0 {
		c.push(opNil)
	}
	c.it.Next()
	var err error
	c.argc = 0
	if vTok, ok := getVerb(e); ok {
		err = c.doVariadic(vTok)
	} else {
		err = c.doExpr(e)
	}
	if err != nil {
		return err
	}
	for i := range ads {
		atok := &ads[i]
		c.argc = 1
		err := c.doVariadic(atok)
		if err != nil {
			return err
		}
	}
	nppe := c.it.Peek()
	if nppe == nil || c.arglist || !isLeftArg(nppe) {
		c.argc = 2
		return c.doVariadic(tok)
	}
	c.it.Next()
	c.argc = 0
	err = c.doExpr(nppe)
	if err != nil {
		return err
	}
	c.argc = 3
	return c.doVariadic(tok)
}

func (c *compiler) doStrand(st *astStrand) error {
	a := make(AV, 0, len(st.Lits))
	for _, tok := range st.Lits {
		switch tok.Type {
		case astNUMBER:
			v, err := parseNumber(tok.Text)
			if err != nil {
				c.pos = tok.Pos
				return c.errorf("number syntax: %v", err)
			}
			a = append(a, v)
		case astSTRING:
			s, err := strconv.Unquote(tok.Text)
			if err != nil {
				c.pos = tok.Pos
				return c.errorf("string syntax: %v", err)
			}
			a = append(a, S(s))
		}
	}
	id := c.ctx.storeConst(canonical(a))
	c.pos = st.Pos
	c.push2(opConst, opcode(id))
	c.apply()
	return nil
}

func (c *compiler) doParenExpr(pe *astParenExpr) error {
	argc := c.argc
	c.argc = 0
	oarglist := c.arglist
	c.arglist = false
	err := c.doExprs(pe.Exprs)
	if err != nil {
		return err
	}
	c.arglist = oarglist
	c.argc = argc + 1
	c.applyAt(pe.EndPos)
	return err
}

func (c *compiler) doBlock(b *astBlock) error {
	switch b.Type {
	case astLAMBDA:
		return c.doLambda(b)
	case astARGS:
		return c.doArgs(b)
	case astSEQ:
		return c.doSeq(b.Body)
	case astLIST:
		return c.doList(b.Body)
	default:
		panic(fmt.Sprintf("unknown block type: %d", b.Type))
	}
}

func (c *compiler) doLambda(b *astBlock) error {
	body := b.Body
	args := b.Args
	argc := c.argc
	slen := c.slen
	c.slen = 0
	c.argc = 0
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
	for i, exprs := range body {
		slen := c.slen
		err := c.doExprs(exprs)
		if err != nil {
			return err
		}
		if i < len(body)-1 && c.slen > slen {
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
	c.argc = argc
	c.slen = slen
	c.push2(opLambda, opcode(id))
	c.applyAt(b.EndPos)
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

func (c *compiler) doArgs(b *astBlock) error {
	body := b.Body
	if len(body) >= 3 {
		expr := c.it.Peek()
		switch expr := expr.(type) {
		case *astToken:
			if expr.Type == astVERB && expr.Text == "?" {
				err := c.doCond(b)
				if err != nil {
					return err
				}
				c.it.Next()
				return nil
			}
		}
	}
	argc := c.argc
	for i := len(body) - 1; i >= 0; i-- {
		exprs := body[i]
		err := c.doExprs(exprs)
		if err != nil {
			return err
		}
	}
	if !c.it.Next() {
		// should not happpen: it would be a sequence
		panic(c.errorf("used as a sequence, but args").Error())
	}
	e := c.it.Expr()
	c.argc = len(body)
	c.arglist = true
	err := c.doExpr(e)
	c.arglist = false
	if err != nil {
		return err
	}
	c.argc = argc + 1
	c.apply()
	return nil
}

func (c *compiler) doCond(b *astBlock) error {
	body := b.Body
	if len(body)%2 != 1 {
		return c.errorf("conditional ?[if;then;else] with even number of statements")
	}
	argc := c.argc
	cond := body[0]
	//slen := c.slen
	err := c.doExprs(cond)
	if err != nil {
		return err
	}
	c.push2(opJumpFalse, opArg)
	jmpCond := len(c.body()) - 1
	jumpsEnd := []int{}
	jumpsElse := []int{}
	jumpsCond := []int{}
	for i := 1; i < len(body); i += 2 {
		then := body[i]
		err := c.doExprs(then)
		if err != nil {
			return err
		}
		c.push2(opJump, opArg)
		jumpsEnd = append(jumpsEnd, len(c.body())-1)
		jumpsElse = append(jumpsElse, len(c.body()))
		elseCond := body[i+1]
		err = c.doExprs(elseCond)
		if err != nil {
			return err
		}
		if i+1 < len(body)-1 {
			c.push2(opJumpFalse, opArg)
			jumpsCond = append(jumpsCond, len(c.body())-1)
		}
	}
	c.body()[jmpCond] = opcode(jumpsElse[0] - jmpCond)
	for i, v := range jumpsCond {
		c.body()[v] = opcode(jumpsElse[i+1] - v)
	}
	end := len(c.body())
	for _, v := range jumpsEnd {
		c.body()[v] = opcode(end - v)
	}
	c.argc = argc + 1
	c.apply()
	return nil
}

func (c *compiler) doSeq(body []exprs) error {
	argc := c.argc
	for i, exprs := range body {
		slen := c.slen
		err := c.doExprs(exprs)
		if err != nil {
			return err
		}
		if i < len(body)-1 && c.slen > slen {
			c.push(opDrop)
		}
	}
	c.argc = argc + 1
	c.apply()
	return nil
}

func (c *compiler) doList(body []exprs) error {
	argc := c.argc
	for i := len(body) - 1; i >= 0; i-- {
		exprs := body[i]
		c.argc = 0
		err := c.doExprs(exprs)
		if err != nil {
			return err
		}
	}
	c.push2(opVariadic, opcode(vList))
	c.push2(opApplyN, opcode(len(body)))
	c.argc = argc + 1
	c.apply()
	return nil
}
