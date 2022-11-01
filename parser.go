package main

import (
	"fmt"
)

// Parser builds an Expr AST from a ppExpr.
type Parser struct {
	pp   *parser
	prog *AstProgram
}

func (p *Parser) Init(s *Scanner) {
	pp := &parser{}
	pp.Init(s)
	p.pp = pp
	p.prog = &AstProgram{}
}

func (p *Parser) Next() ([]Expr, error) {
	pps, err := p.pp.Next()
	if err != nil {
		return nil, err
	}
	for _, ppe := range pps {
		//switch ppe := ppe.(type) {
		switch ppe.(type) {
		case ppBlock:
		case ppStrand:
		case ppExprs:
		case ppToken:
		}
	}
	return nil, nil
}

// parser builds a ppExpr pre-AST
type parser struct {
	ctx    *Context // unused (for now)
	s      *Scanner
	token  Token
	pToken Token
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
	if p.peeked {
		p.token = p.pToken
		p.peeked = false
		return p.token
	}
	p.token = p.s.Next()
	return p.token
}

// Next returns a whole expression, in stack-based order.
func (p *parser) Next() (ppExprs, error) {
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
			return pps, nil
		}
		pps = append(pps, ppe)
	}
}

func closeToken(op TokenType) TokenType {
	switch op {
	case LEFTBRACE:
		return RIGHTBRACE
	case LEFTBRACKET:
		return RIGHTBRACKET
	case LEFTPAREN:
		return RIGHTPAREN
	default:
		panic(fmt.Sprintf("not an opening token:%s", op.String()))
	}
}

func (p *parser) ppExpr() (ppExpr, error) {
	switch tok := p.next(); tok.Type {
	case EOF:
		return ppToken{Type: ppEOF, Pos: tok.Pos}, nil
	case NEWLINE, SEMICOLON:
		return ppToken{Type: ppSEP, Pos: tok.Pos}, nil
	case ERROR:
		return nil, p.errorf("invalid token:%s:%s", tok.Type, tok.Text)
	case ADVERB:
		return nil, p.errorf("syntax:adverb %s at start of expression", tok.Text)
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
			return nil, p.errorf("syntax:unexpected %s without opening matching pair", tok.Text)
		}
		opTok := p.depth[len(p.depth)-1]
		clTokt := closeToken(opTok.Type)
		if clTokt != tok.Type {
			return nil, p.errorf("syntax:unexpected %s without closing previous %s at %d", tok.Text, opTok.Type.String(), opTok.Pos)
		}
		p.depth = p.depth[:len(p.depth)-1]
		return ppToken{Type: ppCLOSE, Pos: tok.Pos}, nil
	case VERB:
		return ppToken{Type: ppVERB, Pos: tok.Pos, Text: tok.Text}, nil
	default:
		// should not happen
		return nil, p.errorf("invalid token type:%s:%s", tok.Type, tok.Text)
	}
}

func (p *parser) ppExprBlock() (ppExpr, error) {
	p.depth = append(p.depth, p.token)
	ppb := ppBlock{}
	ppb.ppexprs = []ppExprs{}
	switch p.token.Type {
	case LEFTBRACE:
		ppb.Type = ppLAMBDA
	case LEFTBRACKET:
		ppb.Type = ppARGS
	case LEFTPAREN:
		ppb.Type = ppLIST
	}
	ppb.ppexprs = append(ppb.ppexprs, ppExprs{})
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
			ppRev(ppb.ppexprs[len(ppb.ppexprs)-1])
			if ppb.Type == ppLIST && len(ppb.ppexprs) == 1 &&
				len(ppb.ppexprs[0]) > 0 {
				// not a list, but a parenthesized
				// expression.
				return ppb.ppexprs[0], nil
			}
			return ppb, nil
		case ppEOF:
			ppRev(ppb.ppexprs[len(ppb.ppexprs)-1])
			opTok := p.depth[len(p.depth)-1]
			return ppb, p.errorf("syntax:unexpected EOF without closing previous %s at %d", opTok.Type.String(), opTok.Pos)
		case ppSEP:
			ppRev(ppb.ppexprs[len(ppb.ppexprs)-1])
			ppb.ppexprs = append(ppb.ppexprs, ppExprs{})
		default:
			ppb.push(ppe)
		}
	}
}

func (p *parser) ppExprStrand() (ppExpr, error) {
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
