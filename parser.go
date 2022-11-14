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

type ErrEOF struct{}

func (e ErrEOF) Error() string {
	return "EOF"
}

// Next returns a whole expression, in stack-based order.
func (p *parser) Next() (exprs, error) {
	ps := exprs{}
	for {
		pe, err := p.expr()
		if err != nil {
			pRev([]expr(ps))
			return ps, err
		}
		tok, ok := pe.(*astToken)
		if ok && (tok.Type == astSEP || tok.Type == astEOF) {
			pRev([]expr(ps))
			if tok.Type == astEOF {
				return ps, ErrEOF{}
			}
			return ps, nil
		}
		ps = append(ps, pe)
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
		return &astToken{Type: astEOF, Pos: tok.Pos}, nil
	case NEWLINE, SEMICOLON:
		return &astToken{Type: astSEP, Pos: tok.Pos}, nil
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
		return &astToken{Type: astCLOSE, Pos: tok.Pos}, nil
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
	pb := &astBlock{}
	switch p.token.Type {
	case LEFTBRACE:
		bt = astLAMBDA
		ntok := p.peek()
		if ntok.Type == LEFTBRACKET {
			p.next()
			args, err := p.pLambdaArgs()
			if err != nil {
				return pb, err
			}
			if len(args) == 0 {
				return pb, p.errorf("empty argument list")
			}
			pb.Args = args
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
	pb.Type = bt
	pb.Body = []exprs{}
	pb.Body = append(pb.Body, exprs{})
	for {
		pe, err := p.expr()
		if err != nil {
			return pb, err
		}
		tok, ok := pe.(*astToken)
		if !ok {
			pb.push(pe)
			continue
		}
		switch tok.Type {
		case astCLOSE:
			pRev(pb.Body[len(pb.Body)-1])
			if pb.Type == astLIST && len(pb.Body) == 1 &&
				len(pb.Body[0]) > 0 {
				// not a list, but a parenthesized
				// expression.
				return astParenExpr(pb.Body[0]), nil
			}
			return pb, nil
		case astEOF:
			pRev(pb.Body[len(pb.Body)-1])
			opTok := p.depth[len(p.depth)-1]
			return pb, p.errorf("unexpected EOF without closing previous %s at %d", opTok, opTok.Pos)
		case astSEP:
			pRev(pb.Body[len(pb.Body)-1])
			pb.Body = append(pb.Body, exprs{})
		default:
			pb.push(pe)
		}
	}
}

func (p *parser) pLambdaArgs() ([]string, error) {
	// c.token.Type is LEFTBRACKET
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
	// c.token.Type is NUMBER or STRING for current and peek
	pb := astAdverbs{}
	for {
		switch p.token.Type {
		case ADVERB:
			pb = append(pb, &astToken{Type: astADVERB, Pos: p.token.Pos, Rune: p.token.Rune})
		}
		ntok := p.peek()
		switch ntok.Type {
		case ADVERB:
			p.next()
		default:
			return pb, nil
		}
	}
}

func (p *parser) pExprStrand() (expr, error) {
	// p.token.Type is NUMBER or STRING for current and peek
	pb := astStrand{}
	for {
		switch p.token.Type {
		case NUMBER:
			pb = append(pb, &astToken{Type: astNUMBER, Pos: p.token.Pos, Text: p.token.Text})
		case STRING:
			pb = append(pb, &astToken{Type: astSTRING, Pos: p.token.Pos, Text: p.token.Text})
		}
		ntok := p.peek()
		switch ntok.Type {
		case NUMBER, STRING:
			p.next()
		default:
			return pb, nil
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
