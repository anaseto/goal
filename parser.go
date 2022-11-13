package goal

import (
	"fmt"
	"strconv"
)

// parser builds an Expr AST from a ppExpr.
type parser struct {
	ctx        *Context      // main execution and compilation context
	pp         *pparser      // pre-parsing into text-based non-resolved AST
	argc       int           // stack length for current sub-expression
	slen       int           // virtual stack length
	arglist    bool          // whether current expression has an argument list
	scopeStack []*LambdaCode // scope information
	pos        int           // last token position
	it         ppIter        // ppExprs iterator
	drop       bool          // whether to add a drop at the end
}

func newParser(ctx *Context) *parser {
	p := &parser{
		ctx: ctx,
		pp:  newPParser(ctx.scanner),
	}
	return p
}

// Parse builds on the context AST using input from the current scanner until
// EOF.
func (p *parser) Parse() error {
	for {
		err := p.ParseNext()
		if err != nil {
			_, eof := err.(ErrEOF)
			if !eof {
				return err
			}
			return nil
		}
	}
}

// Parse builds on the context AST using input from the current scanner until
// the end of a whole expression is found. It returns ErrEOF on EOF.
func (p *parser) ParseNext() error {
	ctx := p.ctx
	if p.drop {
		p.push(opDrop)
	}
	var eof bool
	pps, err := p.pp.Next()
	if err != nil {
		_, eof = err.(ErrEOF)
		if !eof {
			ctx.parser = newParser(ctx)
			return err
		}
	}
	slen := p.slen
	err = p.ppExprs(pps)
	p.drop = p.slen > slen
	if err != nil {
		ctx.parser = newParser(ctx)
		return err
	}
	if eof {
		return ErrEOF{}
	}
	return nil
}

func (p *parser) push(opc opcode) {
	lc := p.scope()
	if lc != nil {
		lc.Body = append(lc.Body, opc)
		lc.Pos = append(lc.Pos, p.pos)
	} else {
		p.ctx.prog.Body = append(p.ctx.prog.Body, opc)
		p.ctx.prog.Pos = append(p.ctx.prog.Pos, p.pos)
		p.ctx.prog.last = len(p.ctx.prog.Body) - 1
	}
	switch opc {
	case opApply:
		// v v -> v
		p.slen--
		p.argc--
	case opApply2:
		// v v v -> v
		p.slen -= 2
		p.argc -= 2
	case opDrop:
		// v ->
		p.slen--
		p.argc--
	case opAssignLocal, opAssignGlobal:
	default:
		// -> v
		p.slen++
		p.argc++
	}
}

func (p *parser) push2(op, arg opcode) {
	lc := p.scope()
	if lc != nil {
		lc.Body = append(lc.Body, op, arg)
		lc.Pos = append(lc.Pos, p.pos, 0)
	} else {
		p.ctx.prog.Body = append(p.ctx.prog.Body, op, arg)
		p.ctx.prog.Pos = append(p.ctx.prog.Pos, p.pos, 0)
		p.ctx.prog.last = len(p.ctx.prog.Body) - 2
	}
	switch op {
	case opApplyN:
		// v ... v v -> v
		p.slen -= int(arg)
		p.argc -= int(arg)
	default:
		// -> v
		p.slen++
		p.argc++
	}
}

func (p *parser) apply() {
	switch {
	case p.argc == 2:
		p.push(opApply)
	case p.argc == 3:
		p.push(opApply2)
	case p.argc > 3:
		p.push2(opApplyN, opcode(p.argc-1))
	}
}

func (p *parser) errorf(format string, a ...interface{}) error {
	// TODO: in case of error, read the file again to get from pos the line
	// and print the line that produced the error with some column marker.
	return fmt.Errorf("error:%d:"+format, append([]interface{}{p.pos}, a...))
}

func (p *parser) scope() *LambdaCode {
	if len(p.scopeStack) == 0 {
		return nil
	}
	return p.scopeStack[len(p.scopeStack)-1]
}

func (p *parser) ppExprs(pps ppExprs) error {
	slen := p.slen
	p.argc = 0
	it := p.it
	p.it = newppIter(pps)
	for p.it.Next() {
		ppe := p.it.Expr()
		err := p.ppExpr(ppe)
		if err != nil {
			return err
		}
	}
	p.it = it
	if p.slen == slen {
		p.push(opNil)
	}
	return nil
}

func (p *parser) ppExpr(ppe ppExpr) error {
	switch ppe := ppe.(type) {
	case ppToken:
		err := p.ppToken(ppe)
		if err != nil {
			return err
		}
	case ppAdverbs:
		err := p.ppAdverbs(ppe)
		if err != nil {
			return err
		}
	case ppStrand:
		err := p.ppStrand(ppe)
		if err != nil {
			return err
		}
	case ppParenExpr:
		argc := p.argc
		p.argc = 0
		oarglist := p.arglist
		p.arglist = false
		err := p.ppParenExpr(ppe)
		p.arglist = oarglist
		p.argc = argc + 1
		p.apply()
		if err != nil {
			return err
		}
	case ppBlock:
		oarglist := p.arglist
		p.arglist = false
		err := p.ppBlock(ppe)
		p.arglist = oarglist
		if err != nil {
			return err
		}
	default:
		panic(p.errorf("unknown ppExpr type"))
	}
	return nil
}

func (p *parser) ppToken(tok ppToken) error {
	p.pos = tok.Pos
	switch tok.Type {
	case ppNUMBER:
		v, err := parseNumber(tok.Text)
		if err != nil {
			return err
		}
		if p.argc > 0 {
			return p.errorf("number atoms cannot be applied")
		}
		id := p.ctx.storeConst(v)
		p.push2(opConst, opcode(id))
		return nil
	case ppSTRING:
		s, err := strconv.Unquote(tok.Text)
		if err != nil {
			return err
		}
		if p.argc > 0 {
			return p.errorf("string atoms cannot be applied")
		}
		id := p.ctx.storeConst(S(s))
		p.push2(opConst, opcode(id))
		return nil
	case ppIDENT:
		// read or apply, not assign
		if p.scope() == nil {
			// global scope: global variable
			p.ppGlobal(tok)
			return nil
		}
		// local scope: argument, local or global variable
		p.ppLocal(tok)
		return nil
	case ppVERB:
		return p.ppVerb(tok)
	default:
		// should not happen
		return p.errorf("unexpected token type: %v", tok.Type)
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

func (p *parser) ppGlobal(tok ppToken) {
	id := p.ctx.global(tok.Text)
	p.push2(opGlobal, opcode(id))
	p.apply()
}

func (p *parser) ppLocal(tok ppToken) {
	lc := p.scope()
	local, ok := lc.local(tok.Text)
	if ok {
		p.push2(opLocal, opArg)
		lc.locals[len(lc.Body)-1] = local
		return
	}
	p.ppGlobal(tok)
}

func (p *parser) ppVariadic(tok ppToken) error {
	v := parseBuiltin(tok.Rune)
	p.push2(opVariadic, opcode(v))
	p.apply()
	return nil
}

func (p *parser) ppVerb(tok ppToken) error {
	ppe := p.it.Peek()
	argc := p.argc
	if ppe == nil || p.arglist {
		return p.ppVariadic(tok)
	}
	if !isLeftArg(ppe) {
		return p.ppVariadic(tok)
	}
	if identTok, ok := getIdent(ppe); ok {
		if p.ppAssign(tok, identTok) {
			p.it.Next()
			return nil
		}
	}
	if argc == 0 {
		p.push(opNil)
	}
	p.it.Next()
	p.argc = 0
	err := p.ppExpr(ppe)
	if err != nil {
		return err
	}
	p.argc = 2
	return p.ppVariadic(tok)
}

func getIdent(ppe ppExpr) (ppToken, bool) {
	tok, ok := ppe.(ppToken)
	return tok, ok && tok.Type == ppIDENT

}

func isLeftArg(ppe ppExpr) bool {
	switch ppe := ppe.(type) {
	case ppToken:
		switch ppe.Type {
		case ppVERB:
			return false
		}
	case ppAdverbs:
		return false
	}
	return true
}

func (p *parser) ppAssign(verbTok, identTok ppToken) bool {
	if verbTok.Rune != ':' || p.argc != 1 {
		return false
	}
	lc := p.scope()
	if lc == nil {
		id := p.ctx.global(identTok.Text)
		p.push2(opAssignGlobal, opcode(id))
		return true
	}
	local, ok := lc.local(identTok.Text)
	if ok {
		p.push2(opAssignLocal, opArg)
		lc.locals[len(lc.Body)-1] = local
		return true
	}
	local = Local{Type: LocalVar, ID: lc.nVars}
	lc.Locals[identTok.Text] = local
	p.push2(opAssignLocal, opArg)
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

func getVerb(ppe ppExpr) (ppToken, bool) {
	tok, ok := ppe.(ppToken)
	return tok, ok && tok.Type == ppVERB

}

func (p *parser) ppAdverbs(adverbs ppAdverbs) error {
	tok := adverbs[len(adverbs)-1]
	adverbs = adverbs[:len(adverbs)-1]
	argc := p.argc
	ppe := p.it.Peek()
	if ppe == nil {
		if len(adverbs) > 0 {
			return errf("adverb train should modify a value")
		}
		p.push(opNil)
		return p.ppVariadic(tok)
	}
	if argc == 0 {
		p.push(opNil)
	}
	p.it.Next()
	var err error
	p.argc = 0
	if vTok, ok := getVerb(ppe); ok {
		err = p.ppVariadic(vTok)
	} else {
		err = p.ppExpr(ppe)
	}
	if err != nil {
		return err
	}
	for _, atok := range adverbs {
		p.argc = 1
		err := p.ppVariadic(atok)
		if err != nil {
			return err
		}
	}
	nppe := p.it.Peek()
	if nppe == nil || p.arglist || !isLeftArg(nppe) {
		p.argc = 2
		return p.ppVariadic(tok)
	}
	p.it.Next()
	p.argc = 0
	err = p.ppExpr(nppe)
	if err != nil {
		return err
	}
	p.argc = 3
	return p.ppVariadic(tok)
}

func (p *parser) ppStrand(pps ppStrand) error {
	a := make(AV, 0, len(pps))
	for _, tok := range pps {
		switch tok.Type {
		case ppNUMBER:
			v, err := parseNumber(tok.Text)
			if err != nil {
				return p.errorf("number syntax: %v", err)
			}
			a = append(a, v)
		case ppSTRING:
			s, err := strconv.Unquote(tok.Text)
			if err != nil {
				return p.errorf("string syntax: %v", err)
			}
			a = append(a, S(s))
		}
	}
	id := p.ctx.storeConst(canonical(a))
	// len(pps) > 0
	p.push2(opConst, opcode(id))
	return nil
}

func (p *parser) ppParenExpr(ppp ppParenExpr) error {
	err := p.ppExprs(ppExprs(ppp))
	return err
}

func (p *parser) ppBlock(ppb ppBlock) error {
	switch ppb.Type {
	case ppLAMBDA:
		return p.ppLambda(ppb.Body, ppb.Args)
	case ppARGS:
		return p.ppArgs(ppb.Body)
	case ppSEQ:
		return p.ppSeq(ppb.Body)
	case ppLIST:
		return p.ppList(ppb.Body)
	default:
		panic(fmt.Sprintf("unknown block type: %d", ppb.Type))
	}
}

func (p *parser) ppLambda(body []ppExprs, args []string) error {
	argc := p.argc
	slen := p.slen
	p.slen = 0
	p.argc = 0
	lc := &LambdaCode{
		Locals: map[string]Local{},
		locals: map[int]Local{},
	}
	p.scopeStack = append(p.scopeStack, lc)
	if len(args) != 0 {
		err := p.ppLambdaArgs(args)
		if err != nil {
			return err
		}
	}
	for i, exprs := range body {
		slen := p.slen
		err := p.ppExprs(exprs)
		if err != nil {
			return err
		}
		if i < len(body)-1 && p.slen > slen {
			p.push(opDrop)
		}
	}
	p.scopeStack = p.scopeStack[:len(p.scopeStack)-1]
	id := len(p.ctx.prog.Lambdas)
	p.ctx.prog.Lambdas = append(p.ctx.prog.Lambdas, lc)
	p.argc = argc
	p.slen = slen
	p.push2(opLambda, opcode(id))
	p.apply()
	return nil
}

func (p *parser) ppLambdaArgs(args []string) error {
	lc := p.scope()
	lc.NamedArgs = true
	for i, arg := range args {
		_, ok := lc.Locals[arg]
		if ok {
			return p.errorf("name %s appears twice in argument list", arg)
		}
		lc.Locals[arg] = Local{
			Type: LocalArg,
			ID:   i,
		}
	}
	return nil
}

func (p *parser) ppArgs(body []ppExprs) error {
	if len(body) >= 3 {
		expr := p.it.Peek()
		switch expr := expr.(type) {
		case ppToken:
			if expr.Type == ppVERB && expr.Rune == '$' {
				return p.parseCond(body)
			}
		}
	}
	argc := p.argc
	for i := len(body) - 1; i >= 0; i-- {
		exprs := body[i]
		err := p.ppExprs(exprs)
		if err != nil {
			return err
		}
	}
	if !p.it.Next() {
		// should not happpen: it would be a sequence
		panic(p.errorf("used as a sequence, but args").Error())
	}
	ppe := p.it.Expr()
	p.argc = len(body)
	p.arglist = true
	err := p.ppExpr(ppe)
	p.arglist = false
	if err != nil {
		return err
	}
	p.argc = argc + 1
	p.apply()
	return nil
}

func (p *parser) parseCond(body []ppExprs) error {
	panic("TODO: parseCond")
	// TODO
	return nil
}

func (p *parser) ppSeq(body []ppExprs) error {
	argc := p.argc
	for i, exprs := range body {
		slen := p.slen
		err := p.ppExprs(exprs)
		if err != nil {
			return err
		}
		if i < len(body)-1 && p.slen > slen {
			p.push(opDrop)
		}
	}
	p.argc = argc + 1
	p.apply()
	return nil
}

func (p *parser) ppList(body []ppExprs) error {
	argc := p.argc
	for i := len(body) - 1; i >= 0; i-- {
		exprs := body[i]
		p.argc = 0
		err := p.ppExprs(exprs)
		if err != nil {
			return err
		}
	}
	p.push2(opVariadic, opcode(vList))
	p.push2(opApplyN, opcode(len(body)))
	p.argc = argc + 1
	p.apply()
	return nil
}

// pparser builds a ppExpr pre-AST.
type pparser struct {
	s      *Scanner
	token  Token // current token
	pToken Token // peeked token
	oToken Token // old (previous) token
	depth  []Token
	peeked bool
}

func newPParser(s *Scanner) *pparser {
	pp := &pparser{s: s}
	return pp
}

type ErrEOF struct{}

func (e ErrEOF) Error() string {
	return "EOF"
}

// Next returns a whole expression, in stack-based order.
func (p *pparser) Next() (ppExprs, error) {
	pps := ppExprs{}
	for {
		ppe, err := p.ppExpr()
		if err != nil {
			ppRev([]ppExpr(pps))
			return pps, err
		}
		tok, ok := ppe.(ppToken)
		if ok && (tok.Type == ppSEP || tok.Type == ppEOF) {
			ppRev([]ppExpr(pps))
			if tok.Type == ppEOF {
				return pps, ErrEOF{}
			}
			return pps, nil
		}
		pps = append(pps, ppe)
	}
}

func (p *pparser) errorf(format string, a ...interface{}) error {
	// TODO: in case of error, read the file again to get from pos the line
	// and print the line that produced the error with some column marker.
	return fmt.Errorf("error:%d:"+format, append([]interface{}{p.token.Pos}, a...))
}

func (p *pparser) peek() Token {
	if p.peeked {
		return p.pToken
	}
	p.pToken = p.s.Next()
	p.peeked = true
	return p.pToken
}

func (p *pparser) next() Token {
	p.oToken = p.token
	if p.peeked {
		p.token = p.pToken
		p.peeked = false
		return p.token
	}
	p.token = p.s.Next()
	return p.token
}

func closeToken(opTok TokenType) TokenType {
	switch opTok {
	case LEFTBRACE:
		return RIGHTBRACE
	case LEFTBRACKET:
		return RIGHTBRACKET
	case LEFTPAREN:
		return RIGHTPAREN
	default:
		panic(fmt.Sprintf("not an opening token:%s", opTok.String()))
	}
}

func (p *pparser) ppExpr() (ppExpr, error) {
	switch tok := p.next(); tok.Type {
	case EOF:
		return ppToken{Type: ppEOF, Pos: tok.Pos}, nil
	case NEWLINE, SEMICOLON:
		return ppToken{Type: ppSEP, Pos: tok.Pos}, nil
	case ERROR:
		return nil, p.errorf("error token: %s", tok)
	case ADVERB:
		return p.ppAdverbs()
		//return nil, p.errorf("adverb %s at start of expression", tok)
	case IDENT:
		return ppToken{Type: ppIDENT, Pos: tok.Pos, Text: tok.Text}, nil
	case LEFTBRACE, LEFTBRACKET, LEFTPAREN:
		return p.ppExprBlock()
	case NUMBER, STRING:
		ntok := p.peek()
		switch ntok.Type {
		case NUMBER, STRING:
			return p.ppExprStrand()
		default:
			pptok := ppToken{Pos: p.token.Pos, Text: p.token.Text}
			switch p.token.Type {
			case NUMBER:
				pptok.Type = ppNUMBER
			case STRING:
				pptok.Type = ppSTRING
			}
			return pptok, nil
		}
	case RIGHTBRACE, RIGHTBRACKET, RIGHTPAREN:
		if len(p.depth) == 0 {
			return nil, p.errorf("unexpected %s without opening matching pair", tok)
		}
		opTok := p.depth[len(p.depth)-1]
		clTokt := closeToken(opTok.Type)
		if clTokt != tok.Type {
			return nil, p.errorf("unexpected %s without closing previous %s at %d", tok, opTok, opTok.Pos)
		}
		p.depth = p.depth[:len(p.depth)-1]
		return ppToken{Type: ppCLOSE, Pos: tok.Pos}, nil
	case VERB:
		return ppToken{Type: ppVERB, Pos: tok.Pos, Rune: tok.Rune}, nil
	default:
		// should not happen
		return nil, p.errorf("invalid token: %v", tok)
	}
}

func (p *pparser) ppExprBlock() (ppExpr, error) {
	var bt ppBlockType
	p.depth = append(p.depth, p.token)
	ppb := ppBlock{}
	switch p.token.Type {
	case LEFTBRACE:
		bt = ppLAMBDA
		ntok := p.peek()
		if ntok.Type == LEFTBRACKET {
			p.next()
			args, err := p.ppLambdaArgs()
			if err != nil {
				return ppb, err
			}
			if len(args) == 0 {
				return ppb, p.errorf("empty argument list")
			}
			ppb.Args = args
		}
	case LEFTBRACKET:
		switch p.oToken.Type {
		case NEWLINE, SEMICOLON, LEFTBRACKET, LEFTPAREN, NONE:
			bt = ppSEQ
		default:
			// arguments being applied to something
			bt = ppARGS
		}
	case LEFTPAREN:
		bt = ppLIST
	}
	ppb.Type = bt
	ppb.Body = []ppExprs{}
	ppb.Body = append(ppb.Body, ppExprs{})
	for {
		ppe, err := p.ppExpr()
		if err != nil {
			return ppb, err
		}
		tok, ok := ppe.(ppToken)
		if !ok {
			ppb.push(ppe)
			continue
		}
		switch tok.Type {
		case ppCLOSE:
			ppRev(ppb.Body[len(ppb.Body)-1])
			if ppb.Type == ppLIST && len(ppb.Body) == 1 &&
				len(ppb.Body[0]) > 0 {
				// not a list, but a parenthesized
				// expression.
				return ppParenExpr(ppb.Body[0]), nil
			}
			return ppb, nil
		case ppEOF:
			ppRev(ppb.Body[len(ppb.Body)-1])
			opTok := p.depth[len(p.depth)-1]
			return ppb, p.errorf("unexpected EOF without closing previous %s at %d", opTok, opTok.Pos)
		case ppSEP:
			ppRev(ppb.Body[len(ppb.Body)-1])
			ppb.Body = append(ppb.Body, ppExprs{})
		default:
			ppb.push(ppe)
		}
	}
}

func (p *pparser) ppLambdaArgs() ([]string, error) {
	// p.token.Type is LEFTBRACKET
	args := []string{}
	for {
		p.next()
		switch p.token.Type {
		case IDENT:
			args = append(args, p.token.Text)
		case RIGHTBRACKET:
			return args, nil
		default:
			return args, p.errorf("expected identifier or ] in argument list, but got %s", p.token)
		}
		p.next()
		switch p.token.Type {
		case RIGHTBRACKET:
			return args, nil
		case SEMICOLON:
		default:
			return args, p.errorf("expected ; or ] in argument list but got %s", p.token)
		}
	}
}

func (p *pparser) ppAdverbs() (ppExpr, error) {
	// p.token.Type is NUMBER or STRING for current and peek
	ppb := ppAdverbs{}
	for {
		switch p.token.Type {
		case ADVERB:
			ppb = append(ppb, ppToken{Type: ppADVERB, Pos: p.token.Pos, Rune: p.token.Rune})
		}
		ntok := p.peek()
		switch ntok.Type {
		case ADVERB:
			p.next()
		default:
			return ppb, nil
		}
	}
}

func (p *pparser) ppExprStrand() (ppExpr, error) {
	// p.token.Type is NUMBER or STRING for current and peek
	ppb := ppStrand{}
	for {
		switch p.token.Type {
		case NUMBER:
			ppb = append(ppb, ppToken{Type: ppNUMBER, Pos: p.token.Pos, Text: p.token.Text})
		case STRING:
			ppb = append(ppb, ppToken{Type: ppSTRING, Pos: p.token.Pos, Text: p.token.Text})
		}
		ntok := p.peek()
		switch ntok.Type {
		case NUMBER, STRING:
			p.next()
		default:
			return ppb, nil
		}
	}
}

func ppRev(pps []ppExpr) {
	for i := 0; i < len(pps)/2; i++ {
		pps[i], pps[len(pps)-i-1] = pps[len(pps)-i-1], pps[i]
	}
}

func bodyRev(body []ppExprs) {
	for i := 0; i < len(body)/2; i++ {
		body[i], body[len(body)-i-1] = body[len(body)-i-1], body[i]
	}
}
