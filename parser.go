package goal

import (
	"fmt"
	"io"
	"strings"
)

// parser builds an expr non-resolved AST.
type parser struct {
	ctx    *Context
	token  Token // current token
	pToken Token // peeked token
	depth  []Token
	peeked bool
}

func newParser(ctx *Context) *parser {
	p := &parser{ctx: ctx}
	return p
}

// Next returns a whole expression, in stack-based order.
func (p *parser) Next() (expr, error) {
	es, err := p.expr(exprs{})
	if err != nil {
		pDoExprs(es)
		return es, err
	}
	pDoExprs(es)
	if p.token.Type == EOF {
		return es, io.EOF
	}
	return es, nil
}

func (p *parser) errorf(format string, a ...interface{}) error {
	p.ctx.errPos = append(p.ctx.errPos, position{Filename: p.ctx.fname, Pos: p.token.Pos})
	return fmt.Errorf("syntax: "+format, a...)
}

func (p *parser) peek() Token {
	if p.peeked {
		return p.pToken
	}
	p.pToken = p.ctx.scanner.Next()
	p.peeked = true
	return p.pToken
}

func (p *parser) next() Token {
	if p.peeked {
		p.token = p.pToken
		p.peeked = false
		return p.token
	}
	p.token = p.ctx.scanner.Next()
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

// expr parses an expression and returns it, or a parse error.
func (p *parser) expr(es exprs) (exprs, error) {
	var err error
	var e expr
	switch tok := p.next(); tok.Type {
	case EOF:
		if len(p.depth) > 0 {
			opTok := p.depth[len(p.depth)-1]
			err = p.errorf("unexpected EOF without closing previous %s", opTok)
			return es, err
		}
		return es, nil
	case NEWLINE:
		switch p.peek().Type {
		case RIGHTPAREN, RIGHTBRACKET, RIGHTBRACE:
			// Ignore a trailing newline if there is nothing before
			// closing delimiter.
			return p.expr(es)
		}
		return es, nil
	case SEMICOLON:
		return es, nil
	case ERROR:
		err = p.errorf("%s", tok.Text)
		return es, err
	case ADVERB:
		switch tok.Text {
		case "'":
			switch p.peek().Type {
			case EOF, NEWLINE, SEMICOLON,
				RIGHTBRACE, RIGHTBRACKET, RIGHTPAREN,
				LEFTBRACKET, ADVERB:
			default:
				e, err = p.earlyReturn(tok.Text)
				return append(es, e), err
			}
		}
		// We handle adverb at start of expression.
		e = p.derivedVerb(nil)
	case IDENT:
		switch ntok := p.peek(); ntok.Type {
		case DYAD:
			if ntok.Text != ":" && ntok.Text != "::" {
				e = &astToken{Type: astIDENT, Pos: tok.Pos, Text: tok.Text}
				break
			}
			p.next()
			e, err = p.assign(tok, strings.HasSuffix(ntok.Text, "::"))
			return append(es, e), err
		case DYADASSIGN:
			p.next()
			e, err = p.assignOp(tok,
				strings.TrimRight(ntok.Text, ":"),
				strings.HasSuffix(ntok.Text, "::"))
			return append(es, e), err
		default:
			e = &astToken{Type: astIDENT, Pos: tok.Pos, Text: tok.Text}
		}
	case LEFTBRACE:
		e, err = p.lambda()
	case LEFTBRACKET:
		// We have a sequence, because index-like application is
		// handled after each expr() below.
		e, err = p.sequence()
	case LEFTPAREN:
		ntok := p.peek()
		if ntok.Type == RIGHTPAREN {
			p.next()
			e = &astToken{Type: astEMPTYLIST, Pos: tok.Pos, Text: tok.Text}
		} else {
			e, err = p.list()
			if err != nil {
				return es, err
			}
			switch ntok := p.peek(); ntok.Type {
			case DYAD:
				if ntok.Text != ":" && ntok.Text != "::" {
					break
				}
				if !isAssignList(e) {
					break
				}
				p.next()
				e, err = p.listAssign(ntok.Pos, getAssignList(e), strings.HasSuffix(ntok.Text, "::"))
				return append(es, e), err
			}
		}
	case NUMBER, STRING:
		switch p.peek().Type {
		case NUMBER, STRING:
			e, err = p.strand()
		default:
			ptok := &astToken{Pos: p.token.Pos, Text: p.token.Text}
			switch p.token.Type {
			case NUMBER:
				ptok.Type = astNUMBER
			case STRING:
				ptok.Type = astSTRING
			}
			e = ptok
		}
	case REGEXP:
		e = &astToken{Type: astREGEXP, Pos: tok.Pos, Text: tok.Text}
	case QQSTART:
		e, err = p.interpolation()
	case RIGHTBRACE, RIGHTBRACKET, RIGHTPAREN:
		if len(p.depth) == 0 {
			err = p.errorf("unexpected %s without opening matching pair", tok)
			return es, err
		}
		opTok := p.depth[len(p.depth)-1]
		clTokt := closeToken(opTok.Type)
		if clTokt != tok.Type {
			err = p.errorf("unexpected %s without closing last %s", tok, opTok)
			p.ctx.errPos = append(p.ctx.errPos, position{Filename: p.ctx.fname, Pos: opTok.Pos})
			return es, err
		}
		p.depth = p.depth[:len(p.depth)-1]
		err = parseCLOSE{tok.Pos}
		return es, err
	case DYAD:
		switch tok.Text {
		case ":":
			if len(es) > 0 && isLeftArg(es[len(es)-1]) {
				break
			}
			switch p.peek().Type {
			case EOF, NEWLINE, SEMICOLON,
				RIGHTBRACE, RIGHTBRACKET, RIGHTPAREN,
				LEFTBRACKET, ADVERB:
			default:
				e, err = p.earlyReturn(tok.Text)
				return append(es, e), err
			}
		}
		e = &astToken{Type: astDYAD, Pos: tok.Pos, Text: tok.Text}
	case DYADASSIGN:
		if tok.Text != "::" {
			return es, p.errorf("assignment operation without identifier left")
		}
		e = &astToken{Type: astDYAD, Pos: tok.Pos, Text: tok.Text}
	case MONAD:
		e = &astToken{Type: astMONAD, Pos: tok.Pos, Text: tok.Text}
	default:
		// should not happen
		panic(fmt.Sprintf("invalid token: %v", tok))
	}
	if err != nil {
		return es, err
	}
	// At this point e is an expression that may be followed by adverbs.
	// All is left is to account for adverbs, and find out if the value is
	// applied with or without bracket indexing, or used as an argument
	// of the next expression (dyad or derived verb).
loop:
	for tok := p.peek(); ; tok = p.peek() {
		switch tok.Type {
		case ADVERB:
			p.next()
			e = p.derivedVerb(e)
		case LEFTBRACKET:
			p.next()
			e, err = p.applyN(e)
			if err != nil || isAmend(e) {
				return append(es, e), err
			}
		default:
			break loop
		}
	}
	// Expression e is now finished, so we handle the case of a dyad or
	// derived verb.
	if len(es) == 0 || !isLeftArg(es[len(es)-1]) {
		// No left argument.
		return p.expr(append(es, e))
	}
	switch ee := e.(type) {
	case *astToken:
		if ee.Type == astDYAD {
			e, err = p.apply2(e, es[len(es)-1])
			es[len(es)-1] = e
			if err != nil {
				return es, err
			}
			return es, nil
		}
	case *astDerivedVerb:
		e, err = p.apply2Adverb(e, es[len(es)-1])
		es[len(es)-1] = e
		if err != nil {
			return es, err
		}
		return es, nil
	}
	return p.expr(append(es, e))
}

func (p *parser) lambda() (expr, error) {
	p.depth = append(p.depth, p.token)
	b := &astLambda{StartPos: p.token.Pos}
	ntok := p.peek()
	if ntok.Type == LEFTBRACKET {
		p.next()
		args, err := p.lambdaArgs()
		if err != nil {
			return b, err
		}
		if len(args) == 0 {
			return b, p.errorf("empty argument list")
		}
		b.Args = args
	}
	b.Body = []expr{}
	for {
		es, err := p.expr(exprs{})
		if err == nil {
			if len(es) > 0 {
				pDoExprs(es)
				b.Body = append(b.Body, es)
			}
			continue
		}
		switch err := err.(type) {
		case parseCLOSE:
			if len(es) > 0 {
				pDoExprs(es)
				b.Body = append(b.Body, es)
			}
			b.EndPos = err.Pos + 1
			if len(b.Body) == 0 {
				return b, p.errorf("empty lambda")
			}
			return b, nil
		default:
			return b, err
		}
	}
}

func (p *parser) lambdaArgs() ([]string, error) {
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
			return args, p.errorf("expected ; or ] in argument list, but got %s", p.token)
		}
	}
}

func (p *parser) sequence() (expr, error) {
	p.depth = append(p.depth, p.token)
	b := &astSeq{StartPos: p.token.Pos}
	b.Body = []expr{}
	for {
		es, err := p.expr(exprs{})
		if err == nil {
			if len(es) > 0 {
				pDoExprs(es)
				b.Body = append(b.Body, es)
			}
			continue
		}
		switch err := err.(type) {
		case parseCLOSE:
			if len(es) > 0 {
				pDoExprs(es)
				b.Body = append(b.Body, es)
			}
			b.EndPos = err.Pos + 1
			if len(b.Body) == 0 {
				return b, p.errorf("empty sequence")
			}
			return b, nil
		default:
			return b, err
		}
	}
}

func (p *parser) applyN(verb expr) (expr, error) {
	p.depth = append(p.depth, p.token)
	a := &astApplyN{
		Verb:     verb,
		Args:     []expr{},
		StartPos: p.token.Pos,
	}
loop:
	for {
		es, err := p.expr(exprs{})
		if err == nil {
			pDoExprs(es)
			a.Args = append(a.Args, es)
			continue
		}
		switch err := err.(type) {
		case parseCLOSE:
			pDoExprs(es)
			a.Args = append(a.Args, es)
			a.EndPos = err.Pos + 1
			break loop
		default:
			return a, err
		}
	}
	identok, ok := getIdent(verb)
	if !ok {
		return a, nil
	}
	switch ntok := p.peek(); ntok.Type {
	case DYAD:
		if ntok.Text != ":" && ntok.Text != "::" {
			return a, nil
		}
		p.next()
		global := strings.HasSuffix(ntok.Text, "::")
		if len(a.Args) > 1 {
			return p.assignDeepAmendOp(identok, a, ":", global)
		}
		return p.assignAmendOp(identok, a.Args, ":", global)
	case DYADASSIGN:
		p.next()
		dyad := strings.TrimRight(ntok.Text, ":")
		global := strings.HasSuffix(ntok.Text, "::")
		if len(a.Args) > 1 {
			return p.assignDeepAmendOp(identok, a, dyad, global)
		}
		return p.assignAmendOp(identok, a.Args, dyad, global)
	default:
		return a, nil
	}
}

func getIdent(e expr) (*astToken, bool) {
	atok, ok := e.(*astToken)
	if !ok || atok.Type != astIDENT {
		return nil, false
	}
	return atok, true
}

func isAmend(e expr) bool {
	_, ok := e.(*astAssignAmendOp)
	return ok
}

func (p *parser) assignAmendOp(identok *astToken, args []expr,
	dyad string, global bool) (expr, error) {
	a := &astAssignAmendOp{
		Name:    identok.Text,
		Global:  global,
		Indices: args[0], // len(args) == 1
		Dyad:    dyad,
		Pos:     identok.Pos,
	}
	es, err := p.subExpr()
	a.Right = es
	if err != nil {
		return a, err
	}
	if len(es) == 0 {
		return a, p.errorf("assignment operation without expression right")
	}
	return a, nil
}

func (p *parser) assignDeepAmendOp(identok *astToken, an *astApplyN,
	dyad string, global bool) (expr, error) {
	a := &astAssignDeepAmendOp{
		Name:    identok.Text,
		Global:  global,
		Indices: &astList{StartPos: an.StartPos, Args: an.Args, EndPos: an.EndPos},
		Dyad:    dyad,
		Pos:     identok.Pos,
	}
	es, err := p.subExpr()
	a.Right = es
	if err != nil {
		return a, err
	}
	if len(es) == 0 {
		return a, p.errorf("assignment operation without expression right")
	}
	return a, nil
}

func (p *parser) apply2(verb, left expr) (expr, error) {
	a := &astApply2{
		Verb: verb,
		Left: left,
	}
	es, err := p.subExpr()
	a.Right = es
	return a, err
}

func (p *parser) apply2Adverb(verb, left expr) (expr, error) {
	a := &astApply2Adverb{
		Verb: verb,
		Left: left,
	}
	es, err := p.subExpr()
	a.Right = es
	return a, err
}

func (p *parser) earlyReturn(s string) (expr, error) {
	a := &astReturn{OnError: s == "'"}
	es, err := p.subExpr()
	a.Expr = es
	if err != nil {
		return a, err
	}
	if len(es) == 0 {
		// should not happen
		return a, p.errorf("return without expression right")
	}
	return a, nil
}

func (p *parser) assign(tok Token, global bool) (expr, error) {
	a := &astAssign{
		Name:   tok.Text,
		Global: global,
		Pos:    tok.Pos,
	}
	es, err := p.subExpr()
	a.Right = es
	if err != nil {
		return a, err
	}
	if len(es) == 0 {
		return a, p.errorf("assignment without expression right")
	}
	return a, nil
}

func (p *parser) listAssign(pos int, names []string, global bool) (expr, error) {
	a := &astListAssign{
		Names:  names,
		Global: global,
		Pos:    pos,
	}
	es, err := p.subExpr()
	a.Right = es
	if err != nil {
		return a, err
	}
	if len(es) == 0 {
		return a, p.errorf("assignment without expression right")
	}
	return a, nil
}

func (p *parser) assignOp(tok Token, dyad string, global bool) (expr, error) {
	a := &astAssignOp{
		Name:   tok.Text,
		Global: global,
		Dyad:   dyad,
		Pos:    tok.Pos,
	}
	es, err := p.subExpr()
	a.Right = es
	if err != nil {
		return a, err
	}
	if len(es) == 0 {
		return a, p.errorf("assignment operation without expression right")
	}
	return a, nil
}

func (p *parser) subExpr() (exprs, error) {
	es, err := p.expr(exprs{})
	pDoExprs(es)
	return es, err
}

func (p *parser) list() (expr, error) {
	p.depth = append(p.depth, p.token)
	l := &astList{StartPos: p.token.Pos}
	l.Args = []expr{}
	for {
		es, err := p.expr(exprs{})
		if err == nil {
			if len(es) == 0 {
				return l, p.errorf("empty slot in list")
			}
			pDoExprs(es)
			l.Args = append(l.Args, es)
			continue
		}
		switch err := err.(type) {
		case parseCLOSE:
			if len(es) == 0 {
				return l, p.errorf("empty slot in list")
			}
			pDoExprs(es)
			l.Args = append(l.Args, es)
			if len(l.Args) == 1 {
				// not a list, but a parenthesized
				// expression.
				return &astParen{
					Expr:     es,
					StartPos: l.StartPos,
					EndPos:   err.Pos + 1,
				}, nil
			}
			l.EndPos = err.Pos + 1
			return l, nil
		default:
			return l, err
		}
	}
}

func isAssignList(e expr) bool {
	le, ok := e.(*astList)
	if !ok || len(le.Args) == 0 {
		return false
	}
	for _, arg := range le.Args {
		es, ok := arg.(exprs)
		if !ok || len(es) != 1 {
			return false
		}
		tok, ok := es[0].(*astToken)
		if !ok || tok.Type != astIDENT {
			return false
		}
	}
	return true
}

func getAssignList(e expr) []string {
	le := e.(*astList)
	names := make([]string, len(le.Args))
	for i, arg := range le.Args {
		tok := arg.(exprs)[0].(*astToken)
		names[i] = tok.Text
	}
	return names
}

func (p *parser) derivedVerb(e expr) *astDerivedVerb {
	// p.token.Type is ADVERB
	atok := &astToken{Type: astADVERB, Pos: p.token.Pos, Text: p.token.Text}
	dv := &astDerivedVerb{Adverb: atok, Verb: e}
	return dv
}

func (p *parser) strand() (expr, error) {
	// p.token.Type is NUMBER or STRING for current and peek
	st := &astStrand{Pos: p.token.Pos}
	for {
		switch p.token.Type {
		case NUMBER:
			st.Lits = append(st.Lits, astToken{Type: astNUMBER, Pos: p.token.Pos, Text: p.token.Text})
		case STRING:
			st.Lits = append(st.Lits, astToken{Type: astSTRING, Pos: p.token.Pos, Text: p.token.Text})
		}
		ntok := p.peek()
		switch ntok.Type {
		case NUMBER, STRING:
			p.next()
		default:
			return st, nil
		}
	}
}

func (p *parser) interpolation() (expr, error) {
	// p.token.Type is QQSTART
	qq := &astInterpolation{Pos: p.token.Pos}
	for {
		p.next()
		switch p.token.Type {
		case QQEND:
			return qq, nil
		case STRING:
			qq.Tokens = append(qq.Tokens, astToken{Type: astSTRING, Pos: p.token.Pos, Text: p.token.Text})
		case IDENT:
			qq.Tokens = append(qq.Tokens, astToken{Type: astIDENT, Pos: p.token.Pos, Text: p.token.Text})
		case ERROR:
			return qq, p.errorf("%s", p.token.Text)
		default:
			return qq, p.errorf("reserved keyword: %s", p.token.Text)
		}
	}
}

// pDoExprs finalizes parsing of a slice of expressions.
func pDoExprs(es exprs) {
	for i := 0; i < len(es)/2; i++ {
		es[i], es[len(es)-i-1] = es[len(es)-i-1], es[i]
	}
	// Recognize some special forms. Currently: #'=
	for i, e := range es {
		if i == len(es)-1 {
			break
		}
		if i == 0 {
			continue
		}
		switch e := e.(type) {
		case *astToken:
			if e.Text != "=" {
				break
			}
			ne, ok := es[i+1].(*astDerivedVerb)
			if !ok || ne.Adverb.Text != "'" {
				break
			}
			v, ok := ne.Verb.(*astToken)
			if !ok {
				break
			}
			if v.Text == "#" {
				e.Text = "icount"
				es[i+1] = &astNop{}
			}
		}
	}
}
