package main

import (
	"fmt"
	"strconv"
)

// Parser builds an Expr AST from a ppExpr.
type Parser struct {
	pp         *parser
	prog       *AstProgram
	argc       int // stack length for current sub-expression
	scopeStack []*AstLambdaCode
	pos        int
	it         ppIter
}

func (p *Parser) Init(s *Scanner) {
	pp := &parser{}
	pp.Init(s)
	p.pp = pp
	p.prog = &AstProgram{
		Globals: map[string]int{},
	}
	p.argc = 0
}

func (p *Parser) pushExpr(e Expr) {
	lc := p.scope()
	if lc != nil {
		lc.Body = append(lc.Body, e)
	} else {
		p.prog.Body = append(p.prog.Body, e)
	}
	switch e := e.(type) {
	case AstApply:
		// v v -> v
		p.argc--
	case AstApplyN:
		// v ... v v -> v
		p.argc -= e.N
	case AstDrop:
		// v ->
		p.argc--
	case AstAssignLocal, AstAssignGlobal:
	default:
		// -> v
		p.argc++
	}
}

func (p *Parser) apply() {
	switch {
	case p.argc == 2:
		p.pushExpr(AstApply{})
	case p.argc > 2:
		p.pushExpr(AstApplyN{N: p.argc - 1})
	}
}

func (p *Parser) errorf(format string, a ...interface{}) error {
	// TODO: in case of error, read the file again to get from pos the line
	// and print the line that produced the error with some column marker.
	return fmt.Errorf("error:%d:"+format, append([]interface{}{p.pos}, a...))
}

func (p *Parser) scope() *AstLambdaCode {
	if len(p.scopeStack) == 0 {
		return nil
	}
	return p.scopeStack[len(p.scopeStack)-1]
}

func (p *Parser) Parse() error {
	for {
		pps, eof, err := p.pp.Next()
		if err != nil {
			return err
		}
		if eof {
			return nil
		}
		err = p.ppExprs(pps)
		if err != nil {
			return err
		}
	}
}

func (p *Parser) ppExprs(pps ppExprs) error {
	argc := p.argc
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
	p.argc = argc
	return nil
}

func (p *Parser) ppExpr(ppe ppExpr) error {
	switch ppe := ppe.(type) {
	case ppToken:
		err := p.ppToken(ppe)
		if err != nil {
			return err
		}
	case ppStrand:
		err := p.ppStrand(ppe)
		if err != nil {
			return err
		}
	case ppParenExpr:
		err := p.ppParenExpr(ppe)
		if err != nil {
			return err
		}
	case ppBlock:
		err := p.ppBlock(ppe)
		if err != nil {
			return err
		}
	case ppLambdaArgs:
		err := p.ppLambdaArgs(ppe)
		if err != nil {
			return err
		}
	default:
		panic(p.errorf("unknown ppExpr type"))
	}
	return nil
}

func (p *Parser) ppToken(tok ppToken) error {
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
		id := p.prog.storeConst(v)
		p.pushExpr(AstConst{ID: id, Pos: tok.Pos})
		return nil
	case ppSTRING:
		s, err := strconv.Unquote(tok.Text)
		if err != nil {
			return err
		}
		if p.argc > 0 {
			return p.errorf("string atoms cannot be applied")
		}
		id := p.prog.storeConst(S(s))
		p.pushExpr(AstConst{ID: id, Pos: tok.Pos})
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
	case ppADVERB:
		return p.ppAdVerb(tok)
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

func (p *Parser) ppGlobal(tok ppToken) {
	id := p.prog.global(tok.Text)
	p.pushExpr(AstGlobal{
		Name: tok.Text,
		ID:   id,
		Pos:  tok.Pos,
	})
	p.apply()
}

func (p *Parser) ppLocal(tok ppToken) {
	local, ok := p.scope().local(tok.Text)
	if ok {
		p.pushExpr(AstLocal{
			Name:  tok.Text,
			Local: local,
			Pos:   tok.Pos,
		})
		p.apply()
		return
	}
	p.ppGlobal(tok)
}

func (p *Parser) ppVerb(tok ppToken) error {
	ppe := p.it.Peek()
	argc := p.argc
	fmt.Printf("argc: %d\ttok: %v\n", argc, tok)
	if ppe != nil && argc == 1 {
		switch ppe := ppe.(type) {
		case ppToken:
			switch ppe.Type {
			case ppIDENT:
				if p.ppAssign(tok, ppe) {
					p.it.Next()
					return nil
				}
				fallthrough
			case ppNUMBER, ppSTRING:
				p.it.Next()
				p.argc = 0
				err := p.ppToken(ppe)
				if err != nil {
					return err
				}
				p.argc += argc
			}
		case ppStrand:
			p.it.Next()
			p.argc = 0
			err := p.ppStrand(ppe)
			if err != nil {
				return err
			}
			p.argc += argc
		case ppParenExpr:
			p.it.Next()
			err := p.ppParenExpr(ppe)
			if err != nil {
				return err
			}
			p.argc += argc
		case ppBlock:
			p.it.Next()
			err := p.ppBlock(ppe)
			if err != nil {
				return err
			}
			p.argc += argc
		}
	}
	switch p.argc {
	case 1:
		monad := parseMonad(tok.Text)
		p.pushExpr(AstMonad{
			Monad: monad,
			Pos:   tok.Pos,
		})
		p.apply()
	default:
		dyad := parseDyad(tok.Text)
		p.pushExpr(AstDyad{
			Dyad: dyad,
			Pos:  tok.Pos,
		})
		p.apply()
	}
	return nil
}

func (p *Parser) ppAssign(verbTok, identTok ppToken) bool {
	if verbTok.Text != ":" || p.argc != 1 {
		return false
	}
	lc := p.scope()
	if lc == nil {
		id := p.prog.global(identTok.Text)
		p.pushExpr(AstAssignGlobal{
			Name: identTok.Text,
			ID:   id,
			Pos:  identTok.Pos,
		})
		return true
	}
	local, ok := lc.local(identTok.Text)
	if ok {
		p.pushExpr(AstAssignLocal{
			Name:  identTok.Text,
			Local: local,
			Pos:   identTok.Pos,
		})
		return true
	}
	p.pushExpr(AstAssignLocal{
		Name:  identTok.Text,
		Local: Local{Type: LocalVar, ID: lc.nVars},
		Pos:   identTok.Pos,
	})
	lc.nVars++
	return true
}

func parseDyad(s string) (verb Dyad) {
	switch s {
	case ":":
		verb = VRight
	case "+":
		verb = VAdd
	case "-":
		verb = VSubtract
	case "*":
		verb = VMultiply
	case "%":
		verb = VDivide
	case "!":
		verb = VMod
	case "&":
		verb = VMin
	case "|":
		verb = VMax
	case "<":
		verb = VLess
	case ">":
		verb = VMore
	case "=":
		verb = VEqual
	case "~":
		verb = VMatch
	case ",":
		verb = VConcat
	case "^":
		verb = VCut
	case "#":
		verb = VTake
	case "_":
		verb = VDrop
	case "$":
		verb = VCast
	case "?":
		verb = VFind
	case "@":
		verb = VApply
	case ".":
		verb = VApplyN
	}
	return verb
}

func parseMonad(s string) (verb Monad) {
	switch s {
	case ":":
		verb = VReturn
	case "+":
		verb = VFlip
	case "-":
		verb = VNegate
	case "*":
		verb = VFirst
	case "%":
		verb = VClassify
	case "!":
		verb = VEnum
	case "&":
		verb = VWhere
	case "|":
		verb = VReverse
	case "<":
		verb = VAscend
	case ">":
		verb = VDescend
	case "=":
		verb = VGroup
	case "~":
		verb = VNot
	case ",":
		verb = VEnlist
	case "^":
		verb = VSort
	case "#":
		verb = VLen
	case "_":
		verb = VFloor
	case "$":
		verb = VString
	case "?":
		verb = VNub
	case "@":
		verb = VType
	case ".":
		verb = VEval
	}
	return verb
}

func (p *Parser) ppAdVerb(tok ppToken) error {
	// TODO: parse adverbs
	return nil
}

func (p *Parser) ppStrand(pps ppStrand) error {
	a := make(AV, 0, len(pps))
	for _, tok := range pps {
		switch tok.Type {
		case ppNUMBER:
			v, err := parseNumber(tok.Text)
			if err != nil {
				return p.errorf("string syntax: %v", err)
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
	id := p.prog.storeConst(canonical(a))
	// len(pps) > 0
	p.pushExpr(AstConst{ID: id, Pos: pps[0].Pos})
	return nil
}

func (p *Parser) ppParenExpr(ppp ppParenExpr) error {
	err := p.ppExprs(ppExprs(ppp))
	return err
}

func (p *Parser) ppBlock(ppb ppBlock) error {
	switch ppb.Type {
	case ppLAMBDA:
		return p.ppLambda(ppb.Body)
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

func (p *Parser) ppLambda(body []ppExprs) error {
	argc := p.argc
	p.argc = 0
	lc := &AstLambdaCode{
		Locals: map[string]Local{},
	}
	p.scopeStack = append(p.scopeStack, lc)
	for _, exprs := range body {
		err := p.ppExprs(exprs)
		if err != nil {
			return err
		}
	}
	p.scopeStack = p.scopeStack[:len(p.scopeStack)-1]
	id := len(p.prog.Lambdas)
	p.prog.Lambdas = append(p.prog.Lambdas, lc)
	p.argc = argc
	p.pushExpr(AstLambda{Lambda: Lambda(id)})
	p.apply()
	return nil
}

func (p *Parser) ppLambdaArgs(args ppLambdaArgs) error {
	lc := p.scope()
	lc.NamedArgs = true
	for i, arg := range args {
		_, ok := lc.Locals[arg]
		if ok {
			return p.errorf("name %s appears twice in signature", arg)
		}
		lc.Locals[arg] = Local{
			Type: LocalArg,
			ID:   i,
		}
	}
	return nil
}

func (p *Parser) ppArgs(body []ppExprs) error {
	if len(body) >= 3 {
		expr := p.it.Peek()
		switch expr := expr.(type) {
		case ppToken:
			if expr.Type == ppVERB && expr.Text == "$" {
				return p.parseCond(body)
			}
		}
	}
	argc := p.argc
	bodyRev(body)
	for _, exprs := range body {
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
	err := p.ppExpr(ppe)
	if err != nil {
		return err
	}
	if argc > 0 {
		p.argc = argc + 1
		p.apply()
	}
	return nil
}

func (p *Parser) parseCond(body []ppExprs) error {
	panic("TODO: parseCond")
	// TODO
	return nil
}

func (p *Parser) ppSeq(body []ppExprs) error {
	argc := p.argc
	for i, exprs := range body {
		err := p.ppExprs(exprs)
		if err != nil {
			return err
		}
		if i < len(body)-1 {
			p.pushExpr(AstDrop{})
		}
	}
	p.argc = argc
	return nil
}

func (p *Parser) ppList(body []ppExprs) error {
	argc := p.argc
	bodyRev(body)
	for _, exprs := range body {
		err := p.ppExprs(exprs)
		if err != nil {
			return err
		}
	}
	p.pushExpr(AstVariadic{Variadic: VList})
	p.pushExpr(AstApplyN{N: len(body)})
	p.argc = argc
	return nil
}

// parser builds a ppExpr pre-AST.
type parser struct {
	ctx    *Context // unused (for now)
	s      *Scanner
	token  Token // current token
	pToken Token // peeked token
	oToken Token // old (previous) token
	depth  []Token
	peeked bool
}

func (p *parser) errorf(format string, a ...interface{}) error {
	// TODO: in case of error, read the file again to get from pos the line
	// and print the line that produced the error with some column marker.
	return fmt.Errorf("error:%d:"+format, append([]interface{}{p.token.Pos}, a...))
}

// Init initializes the parser.
func (p *parser) Init(s *Scanner) {
	s.Init()
	p.s = s
}

func (p *parser) peek() Token {
	if p.peeked {
		return p.pToken
	}
	p.pToken = p.s.Next()
	p.peeked = true
	return p.pToken
}

func (p *parser) next() Token {
	p.oToken = p.token
	if p.peeked {
		p.token = p.pToken
		p.peeked = false
		return p.token
	}
	p.token = p.s.Next()
	return p.token
}

// Next returns a whole expression, in stack-based order.
func (p *parser) Next() (ppExprs, bool, error) {
	pps := ppExprs{}
	for {
		ppe, err := p.ppExpr()
		if err != nil {
			ppRev([]ppExpr(pps))
			return pps, false, err
		}
		tok, ok := ppe.(ppToken)
		if ok && (tok.Type == ppSEP || tok.Type == ppEOF) {
			ppRev([]ppExpr(pps))
			return pps, tok.Type == ppEOF, nil
		}
		pps = append(pps, ppe)
	}
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

func (p *parser) ppExpr() (ppExpr, error) {
	switch tok := p.next(); tok.Type {
	case EOF:
		return ppToken{Type: ppEOF, Pos: tok.Pos}, nil
	case NEWLINE, SEMICOLON:
		return ppToken{Type: ppSEP, Pos: tok.Pos}, nil
	case ERROR:
		return nil, p.errorf("error token: %s", tok)
	case ADVERB:
		return nil, p.errorf("adverb %s at start of expression", tok)
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
		return ppToken{Type: ppVERB, Pos: tok.Pos, Text: tok.Text}, nil
	default:
		// should not happen
		return nil, p.errorf("invalid token '%s' of type %d", tok.Text, tok.Type)
	}
}

func (p *parser) ppExprBlock() (ppExpr, error) {
	var bt ppBlockType
	switch p.token.Type {
	case LEFTBRACE:
		bt = ppLAMBDA
	case LEFTBRACKET:
		switch p.oToken.Type {
		case NEWLINE, SEMICOLON, LEFTBRACKET, LEFTPAREN:
			bt = ppSEQ
		case LEFTBRACE:
			return p.ppLambdaArgs()
		default:
			// arguments being applied to something
			bt = ppARGS
		}
	case LEFTPAREN:
		bt = ppLIST
	}
	p.depth = append(p.depth, p.token)
	ppb := ppBlock{}
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

func (p *parser) ppLambdaArgs() (ppExpr, error) {
	// p.token.Type is LEFTBRACKET
	args := ppLambdaArgs{}
	for {
		p.next()
		switch p.token.Type {
		case IDENT:
			args = append(args, p.token.Text)
		case RIGHTBRACKET:
			return args, nil
		default:
			return args, p.errorf("expected identifier or ] but got %s", p.token)
		}
		p.next()
		switch p.token.Type {
		case RIGHTBRACKET:
			return args, nil
		case SEMICOLON:
		default:
			return args, p.errorf("expected ; or ] but got %s", p.token)
		}
	}
}

func (p *parser) ppExprStrand() (ppExpr, error) {
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
