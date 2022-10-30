package main

import (
	"fmt"
	"io"
)

type Parser struct {
	ctx    *Context  // unused (for now)
	Source string    // for error messages location information (e.g. filename)
	wError io.Writer // where non-fatal error messages go (unused for now)
	s      *Scanner
	token  Token
	pToken Token
	depth  []Token
	peeked bool
	err    error
}

// ParseWithReader parses a frundis source from a reader and returns a list of
// AST blocks.
func (p *Parser) ParseWithReader(reader io.Reader) error {
	s := &Scanner{reader: reader, wError: p.wError}
	s.Init()
	p.s = s
	// TODO
	exprs, err := p.ppExprs()
	if err != nil {
		return err
	}
	for _, expr := range exprs {
		fmt.Printf("%v\n", expr)
	}
	return nil
}

func (p *Parser) peek() Token {
	if p.peeked {
		return p.pToken
	}
	p.pToken = p.s.Next()
	p.peeked = true
	return p.pToken
}

func (p *Parser) next() Token {
	if p.peeked {
		p.token = p.pToken
		p.peeked = false
		return p.token
	}
	p.token = p.s.Next()
	return p.token
}

func (p *Parser) errorf(format string, a ...interface{}) error {
	// TODO: in case of error, read the file again to get from pos the line
	// and print the line that produced the error with some column marker.
	return fmt.Errorf("error:%d:"+format, append([]interface{}{p.token.Pos}, a...))
}

func (p *Parser) ppExprs() ([]ppExpr, error) {
	pps := []ppExpr{}
	for {
		ppe, err := p.ppExpr()
		if err != nil {
			return pps, err
		}
		tok, ok := ppe.(ppToken)
		if ok && (tok.Type == ppSEP || tok.Type == ppEOF) {
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

func (p *Parser) ppExpr() (ppExpr, error) {
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

func (p *Parser) ppExprBlock() (ppExpr, error) {
	p.depth = append(p.depth, p.token)
	ppb := ppBlock{}
	switch p.token.Type {
	case LEFTBRACE:
		ppb.Type = ppBRACE
	case LEFTBRACKET:
		ppb.Type = ppBRACKET
	case LEFTPAREN:
		ppb.Type = ppPAREN
	}
	for {
		ppe, err := p.ppExpr()
		if err != nil {
			return ppb, err
		}
		tok, ok := ppe.(ppToken)
		if ok {
			switch tok.Type {
			case ppCLOSE:
				return ppb, nil
			case ppEOF:
				opTok := p.depth[len(p.depth)-1]
				return ppb, p.errorf("syntax:unexpected EOF without closing previous %s at %d", opTok.Type.String(), opTok.Pos)
			}
		}
		ppb.ppexprs = append(ppb.ppexprs, ppe)
	}
}

func (p *Parser) ppExprStrand() (ppExpr, error) {
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
