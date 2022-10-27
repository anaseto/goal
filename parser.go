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
	expr, err := p.ppExpr()
	if err != nil {
		return err
	}
	fmt.Printf("%v", expr)
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
	return fmt.Errorf("error:%d:"+format, append([]interface{}{p.token.Line}, a...))
}

func (p *Parser) ppExprs() ([]ppExpr, error) {
	pps := []ppExpr{}
	for {
		ppe, err := p.ppExpr()
		if err != nil {
			return pps, err
		}
		tok, ok := ppe.(ppToken)
		if ok && (tok.Type == ppSEP) {
			return pps, nil
		}
		pps = append(pps, ppe)
	}
	return pps, nil
}

func (p *Parser) ppExpr() (ppExpr, error) {
	switch tok := p.next(); tok.Type {
	case EOF, NEWLINE, SEMICOLON:
		return ppToken{Type: ppSEP, Line: tok.Line}, nil
	case ERROR:
		return nil, p.errorf("invalid token:%s:%s", tok.Type, tok.Text)
	case ADVERB:
		return nil, p.errorf("syntax:adverb %s at start of expression", tok.Text)
	case IDENT:
		return ppToken{Type: ppIDENT, Line: tok.Line, Text: tok.Text}, nil
	case LEFTBRACE:
		return p.ppExprBrace()
	case LEFTBRACKET:
		return p.ppExprBracket()
	case LEFTPAREN:
		return p.ppExprParen()
	case NUMBER, STRING:
		return p.ppExprStrand()
	case RIGHTBRACE, RIGHTBRACKET, RIGHTPAREN:
		return nil, p.errorf("syntax:unexpected %s at start of expression", tok.Text)
	case VERB:
		return ppToken{Type: ppVERB, Line: tok.Line, Text: tok.Text}, nil
	default:
		// should not happen
		return nil, p.errorf("invalid token type:%s:%s", tok.Type, tok.Text)
	}
}

func (p *Parser) ppExprBrace() (ppExpr, error) {
	return nil, nil
}

func (p *Parser) ppExprBracket() (ppExpr, error) {
	return nil, nil
}

func (p *Parser) ppExprParen() (ppExpr, error) {
	return nil, nil
}

func (p *Parser) ppExprStrand() (ppExpr, error) {
	return nil, nil
}
