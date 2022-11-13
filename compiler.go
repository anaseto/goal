package goal

import (
	"fmt"
	"strconv"
	"strings"
)

// Program represents a compiled program.
type Program struct {
	Body    []opcode
	Pos     []int
	Lambdas []*LambdaCode

	cLambdas int // index next of last compiled lambda
	cBody    int // number of already processed body ops
	last     int // index of last non-argument opcode
}

// LambdaCode represents a compiled user defined function.
type LambdaCode struct {
	Body      []opcode
	Pos       []int
	Names     []string
	Rank      int
	NamedArgs bool
	Locals    map[string]Local // arguments and variables

	locals map[int]Local // opcode index -> local variable
	nVars  int
}

// Local represents either an argument or a local variable. IDs are
// unique for a given type only.
type Local struct {
	Type LocalType
	ID   int
}

// LocalType represents different kinds of locals.
type LocalType int

// These constants describe the supported kinds of locals.
const (
	LocalArg LocalType = iota
	LocalVar
)

func (l *LambdaCode) local(s string) (Local, bool) {
	param, ok := l.Locals[s]
	if ok {
		return param, true
	}
	if !l.NamedArgs && len(s) == 1 {
		switch r := rune(s[0]); r {
		case 'x', 'y', 'z':
			id := r - 'x'
			arg := Local{Type: LocalArg, ID: int(id)}
			l.Locals[s] = arg
			return arg, true
		}
	}
	return Local{}, false
}

// ProgramString returns a string representation of the compiled program and
// relevant data.
func (ctx *Context) ProgramString() string {
	sb := &strings.Builder{}
	fmt.Fprintln(sb, "---- Compiled program -----")
	fmt.Fprintln(sb, "Instructions:")
	fmt.Fprint(sb, ctx.opcodesString(ctx.prog.Body, nil))
	fmt.Fprintln(sb, "Globals:")
	for id, name := range ctx.gNames {
		fmt.Fprintf(sb, "\t%s\t%d\n", name, id)
	}
	fmt.Fprintln(sb, "Constants:")
	for id, v := range ctx.constants {
		fmt.Fprintf(sb, "\t%d\t%v\n", id, v)
	}
	for id, lc := range ctx.prog.Lambdas {
		fmt.Fprintf(sb, "---- Lambda %d (Rank: %d) -----\n", id, lc.Rank)
		fmt.Fprintf(sb, "%s", ctx.lambdaString(lc))
	}
	return sb.String()
}

func (ctx *Context) lambdaString(lc *LambdaCode) string {
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
	scopeStack []*LambdaCode // scope information
	pos        int           // last token position
	it         pIter         // ppExprs iterator
	drop       bool          // whether to add a drop at the end
}

func newCompiler(ctx *Context) *compiler {
	c := &compiler{
		ctx: ctx,
		p:   newParser(ctx.scanner),
	}
	return c
}

// ParseCompile builds on the context AST using input from the current scanner until
// EOF.
func (c *compiler) ParseCompile() error {
	for {
		err := c.ParseCompileNext()
		if err != nil {
			_, eof := err.(ErrEOF)
			if !eof {
				return err
			}
			return nil
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
	pps, err := c.p.Next()
	if err != nil {
		_, eof = err.(ErrEOF)
		if !eof {
			ctx.compiler = newCompiler(ctx)
			return err
		}
	}
	slen := c.slen
	err = c.ppExprs(pps)
	c.drop = c.slen > slen
	if err != nil {
		ctx.compiler = newCompiler(ctx)
		return err
	}
	if eof {
		return ErrEOF{}
	}
	return nil
}

func (c *compiler) push(opc opcode) {
	lc := c.scope()
	if lc != nil {
		lc.Body = append(lc.Body, opc)
		lc.Pos = append(lc.Pos, c.pos)
	} else {
		c.ctx.prog.Body = append(c.ctx.prog.Body, opc)
		c.ctx.prog.Pos = append(c.ctx.prog.Pos, c.pos)
		c.ctx.prog.last = len(c.ctx.prog.Body) - 1
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
	case opAssignLocal, opAssignGlobal:
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
		lc.Pos = append(lc.Pos, c.pos, 0)
	} else {
		c.ctx.prog.Body = append(c.ctx.prog.Body, op, arg)
		c.ctx.prog.Pos = append(c.ctx.prog.Pos, c.pos, 0)
		c.ctx.prog.last = len(c.ctx.prog.Body) - 2
	}
	switch op {
	case opApplyN:
		// v ... v v -> v
		c.slen -= int(arg)
		c.argc -= int(arg)
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

func (c *compiler) errorf(format string, a ...interface{}) error {
	// TODO: in case of error, read the file again to get from pos the line
	// and print the line that produced the error with some column marker.
	return fmt.Errorf("error:%d:"+format, append([]interface{}{c.pos}, a...))
}

func (c *compiler) scope() *LambdaCode {
	if len(c.scopeStack) == 0 {
		return nil
	}
	return c.scopeStack[len(c.scopeStack)-1]
}

func (c *compiler) ppExprs(pps exprs) error {
	slen := c.slen
	c.argc = 0
	it := c.it
	c.it = newpIter(pps)
	for c.it.Next() {
		ppe := c.it.Expr()
		err := c.ppExpr(ppe)
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

func (c *compiler) ppExpr(ppe expr) error {
	switch ppe := ppe.(type) {
	case pToken:
		err := c.ppToken(ppe)
		if err != nil {
			return err
		}
	case pAdverbs:
		err := c.ppAdverbs(ppe)
		if err != nil {
			return err
		}
	case pStrand:
		err := c.ppStrand(ppe)
		if err != nil {
			return err
		}
	case pParenExpr:
		argc := c.argc
		c.argc = 0
		oarglist := c.arglist
		c.arglist = false
		err := c.ppParenExpr(ppe)
		c.arglist = oarglist
		c.argc = argc + 1
		c.apply()
		if err != nil {
			return err
		}
	case pBlock:
		oarglist := c.arglist
		c.arglist = false
		err := c.ppBlock(ppe)
		c.arglist = oarglist
		if err != nil {
			return err
		}
	default:
		panic(c.errorf("unknown ppExpr type"))
	}
	return nil
}

func (c *compiler) ppToken(tok pToken) error {
	c.pos = tok.Pos
	switch tok.Type {
	case pNUMBER:
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
	case pSTRING:
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
	case pIDENT:
		// read or apply, not assign
		if c.scope() == nil {
			// global scope: global variable
			c.ppGlobal(tok)
			return nil
		}
		// local scope: argument, local or global variable
		c.ppLocal(tok)
		return nil
	case pVERB:
		return c.ppVerb(tok)
	default:
		// should not happen
		return c.errorf("unexpected token type: %v", tok.Type)
	}
}

func parseNumber(s string) (V, error) {
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

func (c *compiler) ppGlobal(tok pToken) {
	id := c.ctx.global(tok.Text)
	c.push2(opGlobal, opcode(id))
	c.apply()
}

func (c *compiler) ppLocal(tok pToken) {
	lc := c.scope()
	local, ok := lc.local(tok.Text)
	if ok {
		c.push2(opLocal, opArg)
		lc.locals[len(lc.Body)-1] = local
		return
	}
	c.ppGlobal(tok)
}

func (c *compiler) ppVariadic(tok pToken) error {
	v := parseBuiltin(tok.Rune)
	c.push2(opVariadic, opcode(v))
	c.apply()
	return nil
}

func (c *compiler) ppVerb(tok pToken) error {
	ppe := c.it.Peek()
	argc := c.argc
	if ppe == nil || c.arglist {
		return c.ppVariadic(tok)
	}
	if !isLeftArg(ppe) {
		return c.ppVariadic(tok)
	}
	if identTok, ok := getIdent(ppe); ok {
		if c.ppAssign(tok, identTok) {
			c.it.Next()
			return nil
		}
	}
	if argc == 0 {
		c.push(opNil)
	}
	c.it.Next()
	c.argc = 0
	err := c.ppExpr(ppe)
	if err != nil {
		return err
	}
	c.argc = 2
	return c.ppVariadic(tok)
}

func getIdent(ppe expr) (pToken, bool) {
	tok, ok := ppe.(pToken)
	return tok, ok && tok.Type == pIDENT

}

func isLeftArg(ppe expr) bool {
	switch ppe := ppe.(type) {
	case pToken:
		switch ppe.Type {
		case pVERB:
			return false
		}
	case pAdverbs:
		return false
	}
	return true
}

func (c *compiler) ppAssign(verbTok, identTok pToken) bool {
	if verbTok.Rune != ':' || c.argc != 1 {
		return false
	}
	lc := c.scope()
	if lc == nil {
		id := c.ctx.global(identTok.Text)
		c.push2(opAssignGlobal, opcode(id))
		return true
	}
	local, ok := lc.local(identTok.Text)
	if ok {
		c.push2(opAssignLocal, opArg)
		lc.locals[len(lc.Body)-1] = local
		return true
	}
	local = Local{Type: LocalVar, ID: lc.nVars}
	lc.Locals[identTok.Text] = local
	c.push2(opAssignLocal, opArg)
	lc.locals[len(lc.Body)-1] = local
	lc.nVars++
	return true
}

func parseBuiltin(r rune) (verb Variadic) {
	switch r {
	case ':':
		verb = vRight
	case '+':
		verb = vAdd
	case '-':
		verb = vSubtract
	case '*':
		verb = vMultiply
	case '%':
		verb = vDivide
	case '!':
		verb = vMod
	case '&':
		verb = vMin
	case '|':
		verb = vMax
	case '<':
		verb = vLess
	case '>':
		verb = vMore
	case '=':
		verb = vEqual
	case '~':
		verb = vMatch
	case ',':
		verb = vJoin
	case '^':
		verb = vCut
	case '#':
		verb = vTake
	case '_':
		verb = vDrop
	case '$':
		verb = vCast
	case '?':
		verb = vFind
	case '@':
		verb = vApply
	case '.':
		verb = vApplyN
	case '\'':
		verb = vEach
	case '/':
		verb = vFold
	case '\\':
		verb = vScan
	}
	return verb
}

func getVerb(ppe expr) (pToken, bool) {
	tok, ok := ppe.(pToken)
	return tok, ok && tok.Type == pVERB

}

func (c *compiler) ppAdverbs(adverbs pAdverbs) error {
	tok := adverbs[len(adverbs)-1]
	adverbs = adverbs[:len(adverbs)-1]
	argc := c.argc
	ppe := c.it.Peek()
	if ppe == nil {
		if len(adverbs) > 0 {
			return errf("adverb train should modify a value")
		}
		c.push(opNil)
		return c.ppVariadic(tok)
	}
	if argc == 0 {
		c.push(opNil)
	}
	c.it.Next()
	var err error
	c.argc = 0
	if vTok, ok := getVerb(ppe); ok {
		err = c.ppVariadic(vTok)
	} else {
		err = c.ppExpr(ppe)
	}
	if err != nil {
		return err
	}
	for _, atok := range adverbs {
		c.argc = 1
		err := c.ppVariadic(atok)
		if err != nil {
			return err
		}
	}
	nppe := c.it.Peek()
	if nppe == nil || c.arglist || !isLeftArg(nppe) {
		c.argc = 2
		return c.ppVariadic(tok)
	}
	c.it.Next()
	c.argc = 0
	err = c.ppExpr(nppe)
	if err != nil {
		return err
	}
	c.argc = 3
	return c.ppVariadic(tok)
}

func (c *compiler) ppStrand(pps pStrand) error {
	a := make(AV, 0, len(pps))
	for _, tok := range pps {
		switch tok.Type {
		case pNUMBER:
			v, err := parseNumber(tok.Text)
			if err != nil {
				return c.errorf("number syntax: %v", err)
			}
			a = append(a, v)
		case pSTRING:
			s, err := strconv.Unquote(tok.Text)
			if err != nil {
				return c.errorf("string syntax: %v", err)
			}
			a = append(a, S(s))
		}
	}
	id := c.ctx.storeConst(canonical(a))
	// len(pps) > 0
	c.push2(opConst, opcode(id))
	return nil
}

func (c *compiler) ppParenExpr(ppp pParenExpr) error {
	err := c.ppExprs(exprs(ppp))
	return err
}

func (c *compiler) ppBlock(ppb pBlock) error {
	switch ppb.Type {
	case pLAMBDA:
		return c.ppLambda(ppb.Body, ppb.Args)
	case pARGS:
		return c.ppArgs(ppb.Body)
	case pSEQ:
		return c.ppSeq(ppb.Body)
	case pLIST:
		return c.ppList(ppb.Body)
	default:
		panic(fmt.Sprintf("unknown block type: %d", ppb.Type))
	}
}

func (c *compiler) ppLambda(body []exprs, args []string) error {
	argc := c.argc
	slen := c.slen
	c.slen = 0
	c.argc = 0
	lc := &LambdaCode{
		Locals: map[string]Local{},
		locals: map[int]Local{},
	}
	c.scopeStack = append(c.scopeStack, lc)
	if len(args) != 0 {
		err := c.ppLambdaArgs(args)
		if err != nil {
			return err
		}
	}
	for i, exprs := range body {
		slen := c.slen
		err := c.ppExprs(exprs)
		if err != nil {
			return err
		}
		if i < len(body)-1 && c.slen > slen {
			c.push(opDrop)
		}
	}
	c.scopeStack = c.scopeStack[:len(c.scopeStack)-1]
	id := len(c.ctx.prog.Lambdas)
	c.ctx.prog.Lambdas = append(c.ctx.prog.Lambdas, lc)
	c.argc = argc
	c.slen = slen
	c.push2(opLambda, opcode(id))
	c.apply()
	return nil
}

func (c *compiler) ppLambdaArgs(args []string) error {
	lc := c.scope()
	lc.NamedArgs = true
	for i, arg := range args {
		_, ok := lc.Locals[arg]
		if ok {
			return c.errorf("name %s appears twice in argument list", arg)
		}
		lc.Locals[arg] = Local{
			Type: LocalArg,
			ID:   i,
		}
	}
	return nil
}

func (c *compiler) ppArgs(body []exprs) error {
	if len(body) >= 3 {
		expr := c.it.Peek()
		switch expr := expr.(type) {
		case pToken:
			if expr.Type == pVERB && expr.Rune == '$' {
				return c.parseCond(body)
			}
		}
	}
	argc := c.argc
	for i := len(body) - 1; i >= 0; i-- {
		exprs := body[i]
		err := c.ppExprs(exprs)
		if err != nil {
			return err
		}
	}
	if !c.it.Next() {
		// should not happpen: it would be a sequence
		panic(c.errorf("used as a sequence, but args").Error())
	}
	ppe := c.it.Expr()
	c.argc = len(body)
	c.arglist = true
	err := c.ppExpr(ppe)
	c.arglist = false
	if err != nil {
		return err
	}
	c.argc = argc + 1
	c.apply()
	return nil
}

func (c *compiler) parseCond(body []exprs) error {
	panic("TODO: parseCond")
	// TODO
	return nil
}

func (c *compiler) ppSeq(body []exprs) error {
	argc := c.argc
	for i, exprs := range body {
		slen := c.slen
		err := c.ppExprs(exprs)
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

func (c *compiler) ppList(body []exprs) error {
	argc := c.argc
	for i := len(body) - 1; i >= 0; i-- {
		exprs := body[i]
		c.argc = 0
		err := c.ppExprs(exprs)
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
