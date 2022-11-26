package goal

import (
	"fmt"
)

// parser builds an expr non-resolved AST.
type parser struct {
	ctx    *Context
	token  Token // current token
	pToken Token // peeked token
	depth  []Token
	peeked bool
	exprs  exprs // current sub-expression application list
}

func newParser(ctx *Context) *parser {
	p := &parser{ctx: ctx}
	return p
}

// errEOF signals the end of the input file.
type errEOF struct{}

func (e errEOF) Error() string {
	return "EOF"
}

// Next returns a whole expression, in stack-based order.
func (p *parser) Next() (exprs, error) {
	es := exprs{}
	for {
		e, err := p.expr()
		if err != nil {
			pDoExprs(es)
			switch err.(type) {
			case parseSEP:
				return es, nil
			case parseEOF:
				return es, errEOF{}
			}
			return es, err
		}
		es = append(es, e)
	}
}

func parseReturn(es exprs) exprs {
	if len(es) == 0 {
		return es
	}
	if e, ok := es[0].(*astToken); ok && e.Type == astDYAD && e.Text == ":" {
		es[0] = &astReturn{Pos: e.Pos}
	}
	return es
}

func (p *parser) errorf(format string, a ...interface{}) error {
	p.ctx.errPos = append(p.ctx.errPos, position{Filename: p.ctx.fname, Pos: p.token.Pos})
	return fmt.Errorf("parsing: "+format, a...)
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

// expr returns an applicable subexpression, or an error.
func (p *parser) expr() (e expr, err error) {
	switch tok := p.next(); tok.Type {
	case EOF:
		err = parseEOF{}
	case NEWLINE, SEMICOLON:
		err = parseSEP{}
	case ERROR:
		err = p.errorf("%s", tok)
	case ADVERB:
		e, err = p.pAdverbs()
		//return nil, c.errorf("adverb %s at start of expression", tok)
	case IDENT:
		e = &astToken{Type: astIDENT, Pos: tok.Pos, Text: tok.Text}
	case LEFTBRACE:
		e, err = p.pExprBrace()
	case LEFTBRACKET:
		e, err = p.pExprBracket()
	case LEFTPAREN:
		e, err = p.pExprParen()
	case NUMBER, STRING:
		ntok := p.peek()
		switch ntok.Type {
		case NUMBER, STRING:
			e, err = p.pExprStrand()
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
	case RIGHTBRACE, RIGHTBRACKET, RIGHTPAREN:
		if len(p.depth) == 0 {
			err = p.errorf("unexpected %s without opening matching pair", tok)
			break
		}
		opTok := p.depth[len(p.depth)-1]
		clTokt := closeToken(opTok.Type)
		if clTokt != tok.Type {
			err = p.errorf("unexpected %s without closing previous %s at %d", tok, opTok, opTok.Pos)
			break
		}
		p.depth = p.depth[:len(p.depth)-1]
		err = parseCLOSE{tok.Pos}
	case DYAD:
		e = &astToken{Type: astDYAD, Pos: tok.Pos, Text: tok.Text}
	case MONAD:
		e = &astToken{Type: astMONAD, Pos: tok.Pos, Text: tok.Text}
	default:
		// should not happen
		err = p.errorf("invalid token: %v", tok)
	}
	if err != nil {
		return e, err
	}
	if tok := p.peek(); tok.Type == LEFTBRACKET {
		p.next()
		return p.pExprApplyN(e)
	}
	return e, err
}

func (p *parser) pExprBrace() (expr, error) {
	p.depth = append(p.depth, p.token)
	b := &astLambda{StartPos: p.token.Pos}
	ntok := p.peek()
	if ntok.Type == LEFTBRACKET {
		p.next()
		args, err := p.pLambdaArgs()
		if err != nil {
			return b, err
		}
		if len(args) == 0 {
			return b, p.errorf("empty argument list")
		}
		b.Args = args
	}
	b.Body = []exprs{}
	b.Body = append(b.Body, exprs{})
	for {
		pe, err := p.expr()
		if err == nil {
			b.push(pe)
			continue
		}
		switch err := err.(type) {
		case parseCLOSE:
			pDoExprs(b.Body[len(b.Body)-1])
			b.EndPos = err.Pos + 1
			return b, nil
		case parseEOF:
			pDoExprs(b.Body[len(b.Body)-1])
			opTok := p.depth[len(p.depth)-1]
			perr := p.errorf("unexpected EOF without closing previous %s", opTok)
			p.ctx.errPos = append(p.ctx.errPos, position{Filename: p.ctx.fname, Pos: opTok.Pos})
			return b, perr
		case parseSEP:
			pDoExprs(b.Body[len(b.Body)-1])
			b.Body = append(b.Body, exprs{})
		default:
			return b, err
		}
	}
}

func (p *parser) pLambdaArgs() ([]string, error) {
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

func (p *parser) pExprBracket() (expr, error) {
	// We have a sequence, because the bracket is not applied to a previous
	// expression.
	p.depth = append(p.depth, p.token)
	return p.pExprSeq()
}

func (p *parser) pExprSeq() (expr, error) {
	b := &astSeq{StartPos: p.token.Pos}
	b.Body = []exprs{}
	b.Body = append(b.Body, exprs{})
	for {
		pe, err := p.expr()
		if err == nil {
			b.push(pe)
			continue
		}
		switch err := err.(type) {
		case parseCLOSE:
			pDoExprs(b.Body[len(b.Body)-1])
			b.EndPos = err.Pos + 1
			return b, nil
		case parseEOF:
			pDoExprs(b.Body[len(b.Body)-1])
			opTok := p.depth[len(p.depth)-1]
			perr := p.errorf("unexpected EOF without closing previous %s", opTok)
			p.ctx.errPos = append(p.ctx.errPos, position{Filename: p.ctx.fname, Pos: opTok.Pos})
			return b, perr
		case parseSEP:
			pDoExprs(b.Body[len(b.Body)-1])
			b.Body = append(b.Body, exprs{})
		default:
			return b, err
		}
	}
}

func (p *parser) pExprApplyN(e expr) (expr, error) {
	p.depth = append(p.depth, p.token)
	a := &astApplyN{
		Expr:     e,
		Args:     []exprs{{}},
		StartPos: p.token.Pos,
	}
	for {
		pe, err := p.expr()
		if err == nil {
			a.push(pe)
			continue
		}
		switch err := err.(type) {
		case parseCLOSE:
			pDoExprs(a.Args[len(a.Args)-1])
			a.EndPos = err.Pos + 1
			return a, nil
		case parseEOF:
			pDoExprs(a.Args[len(a.Args)-1])
			opTok := p.depth[len(p.depth)-1]
			perr := p.errorf("unexpected EOF without closing previous %s", opTok)
			p.ctx.errPos = append(p.ctx.errPos, position{Filename: p.ctx.fname, Pos: opTok.Pos})
			return a, perr
		case parseSEP:
			pDoExprs(a.Args[len(a.Args)-1])
			a.Args = append(a.Args, exprs{})
		default:
			return a, err
		}
	}
}

func (p *parser) pExprParen() (expr, error) {
	p.depth = append(p.depth, p.token)
	l := &astList{StartPos: p.token.Pos}
	l.Args = []exprs{}
	l.Args = append(l.Args, exprs{})
	for {
		pe, err := p.expr()
		if err == nil {
			l.push(pe)
			continue
		}
		switch err := err.(type) {
		case parseCLOSE:
			pDoExprs(l.Args[len(l.Args)-1])
			if len(l.Args) == 1 && len(l.Args[0]) > 0 {
				// not a list, but a parenthesized
				// expression.
				return &astParen{
					Exprs:    l.Args[0],
					StartPos: l.StartPos,
					EndPos:   err.Pos + 1,
				}, nil
			}
			l.EndPos = err.Pos + 1
			return l, nil
		case parseEOF:
			pDoExprs(l.Args[len(l.Args)-1])
			opTok := p.depth[len(p.depth)-1]
			perr := p.errorf("unexpected EOF without closing previous %s", opTok)
			p.ctx.errPos = append(p.ctx.errPos, position{Filename: p.ctx.fname, Pos: opTok.Pos})
			return l, perr
		case parseSEP:
			pDoExprs(l.Args[len(l.Args)-1])
			l.Args = append(l.Args, exprs{})
		default:
			return l, err
		}
	}
}

func (p *parser) pAdverbs() (expr, error) {
	// p.token.Type is ADVERB
	ads := &astAdverbs{}
	for {
		if p.token.Type == ADVERB {
			ads.Train = append(ads.Train, astToken{Type: astADVERB, Pos: p.token.Pos, Text: p.token.Text})
		}
		ntok := p.peek()
		if ntok.Type == ADVERB {
			p.next()
			continue
		}
		return ads, nil
	}
}

func (p *parser) pExprStrand() (expr, error) {
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

// pDoExprs finalizes parsing of a slice of expressions.
func pDoExprs(es exprs) {
	es = parseReturn(es)
	for i := 0; i < len(es)/2; i++ {
		es[i], es[len(es)-i-1] = es[len(es)-i-1], es[i]
	}
	//arglist := false // last expr was an argument list
	//for i, e := range es {
	//switch e := e.(type) {
	//case *astBlock:
	//arglist = e.Type == astARGS
	//case *astToken:
	//if e.Type == astDYAD && !arglist && i < len(es)-1 {
	//ne := es[i+1]
	//if isLeftArg(ne) {
	//b := &astBlock{
	//Type:     astARGS,
	//Body:     []exprs{{ne}, es[:i]},
	//StartPos: e.Pos,
	//EndPos:   e.Pos,
	//}
	//es[i+1], es[i] = es[i], b
	//es = es[i:]
	//arglist = true
	//}
	//}
	//arglist = false
	//default:
	//arglist = false
	//}
	//}
}

func bodyRev(body []exprs) {
	for i := 0; i < len(body)/2; i++ {
		body[i], body[len(body)-i-1] = body[len(body)-i-1], body[i]
	}
}
