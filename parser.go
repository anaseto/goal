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
	return fmt.Errorf("goal:%d:"+format, append([]interface{}{p.token.Line}, a...))
}

func (p *Parser) ppExprs() ([]ppExpr, error) {
	pps := []ppExpr{}
	for {
		ppe, err := p.ppExpr()
		if err != nil {
			return pps, err
		}
		tok, ok := ppe.(Token)
		if ok && (tok.Type == EOF || tok.Type == NEWLINE || tok.Type == SEMICOLON) {
			return pps, nil
		}
		pps = append(pps, ppe)
	}
	return pps, nil
}

func (p *Parser) ppExpr() (ppExpr, error) {
	switch tok := p.next(); tok.Type {
	case EOF:
		return tok, nil
	case ERROR:
		return nil, p.errorf("invalid token:%s:%s", tok.Type, tok.Text)
	case ADVERB:
		return nil, p.errorf("adverb:%s:syntax:should follow verb", tok.Text)
	case IDENT:
		return tok, nil
	case LEFTBRACE:
		//pps := []ppExpr{}
	}
	return nil, nil
}
