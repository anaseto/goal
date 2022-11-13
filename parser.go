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
func (c *parser) Next() (exprs, error) {
	ps := exprs{}
	for {
		pe, err := c.expr()
		if err != nil {
			pRev([]expr(ps))
			return ps, err
		}
		tok, ok := pe.(pToken)
		if ok && (tok.Type == pSEP || tok.Type == pEOF) {
			pRev([]expr(ps))
			if tok.Type == pEOF {
				return ps, ErrEOF{}
			}
			return ps, nil
		}
		ps = append(ps, pe)
	}
}

func (c *parser) errorf(format string, a ...interface{}) error {
	// TODO: in case of error, read the file again to get from pos the line
	// and print the line that produced the error with some column marker.
	return fmt.Errorf("error:%d:"+format, append([]interface{}{c.token.Pos}, a...))
}

func (c *parser) peek() Token {
	if c.peeked {
		return c.pToken
	}
	c.pToken = c.s.Next()
	c.peeked = true
	return c.pToken
}

func (c *parser) next() Token {
	c.oToken = c.token
	if c.peeked {
		c.token = c.pToken
		c.peeked = false
		return c.token
	}
	c.token = c.s.Next()
	return c.token
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

func (c *parser) expr() (expr, error) {
	switch tok := c.next(); tok.Type {
	case EOF:
		return pToken{Type: pEOF, Pos: tok.Pos}, nil
	case NEWLINE, SEMICOLON:
		return pToken{Type: pSEP, Pos: tok.Pos}, nil
	case ERROR:
		return nil, c.errorf("error token: %s", tok)
	case ADVERB:
		return c.pAdverbs()
		//return nil, c.errorf("adverb %s at start of expression", tok)
	case IDENT:
		return pToken{Type: pIDENT, Pos: tok.Pos, Text: tok.Text}, nil
	case LEFTBRACE, LEFTBRACKET, LEFTPAREN:
		return c.pExprBlock()
	case NUMBER, STRING:
		ntok := c.peek()
		switch ntok.Type {
		case NUMBER, STRING:
			return c.pExprStrand()
		default:
			ptok := pToken{Pos: c.token.Pos, Text: c.token.Text}
			switch c.token.Type {
			case NUMBER:
				ptok.Type = pNUMBER
			case STRING:
				ptok.Type = pSTRING
			}
			return ptok, nil
		}
	case RIGHTBRACE, RIGHTBRACKET, RIGHTPAREN:
		if len(c.depth) == 0 {
			return nil, c.errorf("unexpected %s without opening matching pair", tok)
		}
		opTok := c.depth[len(c.depth)-1]
		clTokt := closeToken(opTok.Type)
		if clTokt != tok.Type {
			return nil, c.errorf("unexpected %s without closing previous %s at %d", tok, opTok, opTok.Pos)
		}
		c.depth = c.depth[:len(c.depth)-1]
		return pToken{Type: pCLOSE, Pos: tok.Pos}, nil
	case VERB:
		return pToken{Type: pVERB, Pos: tok.Pos, Rune: tok.Rune}, nil
	default:
		// should not happen
		return nil, c.errorf("invalid token: %v", tok)
	}
}

func (c *parser) pExprBlock() (expr, error) {
	var bt pBlockType
	c.depth = append(c.depth, c.token)
	pb := pBlock{}
	switch c.token.Type {
	case LEFTBRACE:
		bt = pLAMBDA
		ntok := c.peek()
		if ntok.Type == LEFTBRACKET {
			c.next()
			args, err := c.pLambdaArgs()
			if err != nil {
				return pb, err
			}
			if len(args) == 0 {
				return pb, c.errorf("empty argument list")
			}
			pb.Args = args
		}
	case LEFTBRACKET:
		switch c.oToken.Type {
		case NEWLINE, SEMICOLON, LEFTBRACKET, LEFTPAREN, NONE:
			bt = pSEQ
		default:
			// arguments being applied to something
			bt = pARGS
		}
	case LEFTPAREN:
		bt = pLIST
	}
	pb.Type = bt
	pb.Body = []exprs{}
	pb.Body = append(pb.Body, exprs{})
	for {
		pe, err := c.expr()
		if err != nil {
			return pb, err
		}
		tok, ok := pe.(pToken)
		if !ok {
			pb.push(pe)
			continue
		}
		switch tok.Type {
		case pCLOSE:
			pRev(pb.Body[len(pb.Body)-1])
			if pb.Type == pLIST && len(pb.Body) == 1 &&
				len(pb.Body[0]) > 0 {
				// not a list, but a parenthesized
				// expression.
				return pParenExpr(pb.Body[0]), nil
			}
			return pb, nil
		case pEOF:
			pRev(pb.Body[len(pb.Body)-1])
			opTok := c.depth[len(c.depth)-1]
			return pb, c.errorf("unexpected EOF without closing previous %s at %d", opTok, opTok.Pos)
		case pSEP:
			pRev(pb.Body[len(pb.Body)-1])
			pb.Body = append(pb.Body, exprs{})
		default:
			pb.push(pe)
		}
	}
}

func (c *parser) pLambdaArgs() ([]string, error) {
	// c.token.Type is LEFTBRACKET
	args := []string{}
	for {
		c.next()
		switch c.token.Type {
		case IDENT:
			args = append(args, c.token.Text)
		case RIGHTBRACKET:
			return args, nil
		default:
			return args, c.errorf("expected identifier or ] in argument list, but got %s", c.token)
		}
		c.next()
		switch c.token.Type {
		case RIGHTBRACKET:
			return args, nil
		case SEMICOLON:
		default:
			return args, c.errorf("expected ; or ] in argument list but got %s", c.token)
		}
	}
}

func (c *parser) pAdverbs() (expr, error) {
	// c.token.Type is NUMBER or STRING for current and peek
	pb := pAdverbs{}
	for {
		switch c.token.Type {
		case ADVERB:
			pb = append(pb, pToken{Type: pADVERB, Pos: c.token.Pos, Rune: c.token.Rune})
		}
		ntok := c.peek()
		switch ntok.Type {
		case ADVERB:
			c.next()
		default:
			return pb, nil
		}
	}
}

func (c *parser) pExprStrand() (expr, error) {
	// c.token.Type is NUMBER or STRING for current and peek
	pb := pStrand{}
	for {
		switch c.token.Type {
		case NUMBER:
			pb = append(pb, pToken{Type: pNUMBER, Pos: c.token.Pos, Text: c.token.Text})
		case STRING:
			pb = append(pb, pToken{Type: pSTRING, Pos: c.token.Pos, Text: c.token.Text})
		}
		ntok := c.peek()
		switch ntok.Type {
		case NUMBER, STRING:
			c.next()
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
