package goal

import (
	"fmt"
)

// parser builds an expr non-resolved AST.
type parser struct {
	s      *Scanner
	token  Token // current token
	pToken Token // peeked token
	oToken Token // old (previous) token
	depth  []Token
	peeked bool
}

func newParser(s *Scanner) *parser {
	p := &parser{s: s}
	return p
}

// ErrEOF signals the end of the input file.
type ErrEOF struct{}

func (e ErrEOF) Error() string {
	return "EOF"
}

// Next returns a whole expression, in stack-based order.
func (p *parser) Next() (exprs, error) {
	es := exprs{}
	for {
		e, err := p.expr()
		if err != nil {
			pRev([]expr(es))
			switch err.(type) {
			case parseSEP:
				return es, nil
			case parseEOF:
				return es, ErrEOF{}
			}
			return es, err
		}
		es = append(es, e)
	}
}

func (p *parser) errorf(format string, a ...interface{}) error {
	// TODO: in case of error, read the file again to get from pos the line
	// and print the line that produced the error with some column marker.
	return fmt.Errorf("error:%d:"+format, append([]interface{}{p.token.Pos}, a...))
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

func (p *parser) expr() (expr, error) {
	switch tok := p.next(); tok.Type {
	case EOF:
		return nil, parseEOF{}
	case NEWLINE, SEMICOLON:
		return nil, parseSEP{}
	case ERROR:
		return nil, p.errorf("error token: %s", tok)
	case ADVERB:
		return p.pAdverbs()
		//return nil, c.errorf("adverb %s at start of expression", tok)
	case IDENT:
		return &astToken{Type: astIDENT, Pos: tok.Pos, Text: tok.Text}, nil
	case LEFTBRACE, LEFTBRACKET, LEFTPAREN:
		return p.pExprBlock()
	case NUMBER, STRING:
		ntok := p.peek()
		switch ntok.Type {
		case NUMBER, STRING:
			return p.pExprStrand()
		default:
			ptok := &astToken{Pos: p.token.Pos, Text: p.token.Text}
			switch p.token.Type {
			case NUMBER:
				ptok.Type = astNUMBER
			case STRING:
				ptok.Type = astSTRING
			}
			return ptok, nil
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
		return nil, parseCLOSE{tok.Pos}
	case VERB:
		return &astToken{Type: astVERB, Pos: tok.Pos, Rune: tok.Rune}, nil
	default:
		// should not happen
		return nil, p.errorf("invalid token: %v", tok)
	}
}

func (p *parser) pExprBlock() (expr, error) {
	var bt astBlockType
	p.depth = append(p.depth, p.token)
	b := &astBlock{StartPos: p.token.Pos}
	switch p.token.Type {
	case LEFTBRACE:
		bt = astLAMBDA
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
	case LEFTBRACKET:
		switch p.oToken.Type {
		case NEWLINE, SEMICOLON, LEFTBRACKET, LEFTPAREN, NONE:
			bt = astSEQ
		default:
			// arguments being applied to something
			bt = astARGS
		}
	case LEFTPAREN:
		bt = astLIST
	}
	b.Type = bt
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
			pRev(b.Body[len(b.Body)-1])
			if b.Type == astLIST && len(b.Body) == 1 &&
				len(b.Body[0]) > 0 {
				// not a list, but a parenthesized
				// expression.
				return &astParenExpr{
					Exprs:    b.Body[0],
					StartPos: b.StartPos,
					EndPos:   err.Pos,
				}, nil
			}
			b.EndPos = err.Pos
			return b, nil
		case parseEOF:
			pRev(b.Body[len(b.Body)-1])
			opTok := p.depth[len(p.depth)-1]
			return b, p.errorf("unexpected EOF without closing previous %s at %d", opTok, opTok.Pos)
		case parseSEP:
			pRev(b.Body[len(b.Body)-1])
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

func (p *parser) pAdverbs() (expr, error) {
	// p.token.Type is ADVERB
	ads := &astAdverbs{}
	for {
		if p.token.Type == ADVERB {
			ads.Train = append(ads.Train, astToken{Type: astADVERB, Pos: p.token.Pos, Rune: p.token.Rune})
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
	st := &astStrand{}
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

func pRev(es []expr) {
	for i := 0; i < len(es)/2; i++ {
		es[i], es[len(es)-i-1] = es[len(es)-i-1], es[i]
	}
}

func bodyRev(body []exprs) {
	for i := 0; i < len(body)/2; i++ {
		body[i], body[len(body)-i-1] = body[len(body)-i-1], body[i]
	}
}
