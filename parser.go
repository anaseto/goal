package main

import (
	"io"
	"fmt"
)

type Parser struct {
	ctx    *Context  // unused (for now)
	Source string    // for error messages location information (e.g. filename)
	wError io.Writer // where non-fatal error messages go (unused for now)
	s      *Scanner
	token  Token
	expr Expr // building expression
	err error
}

// ParseWithReader parses a frundis source from a reader and returns a list of
// AST blocks.
func (p *Parser) ParseWithReader(reader io.Reader) error {
	s := &Scanner{reader: reader, wError: p.wError}
	s.Init()
	p.s = s
	_, err := p.parseExpr()
	if err != nil {
		return err
	}
	return nil
}

func (p *Parser) next() {
	p.token = p.s.Next()
}

type pStateFn func(*Parser) (pStateFn, error)

func (p *Parser) parseExpr() (Expr, error) {
	state := parseAny
	var err error
	for {
		state, err = state(p)
		if state == nil || err != nil {
			return p.expr, err
		}
	}
}

func parseAny(p *Parser) (pStateFn, error) {
	p.next()
	switch p.token.Type {
	case EOF:
		return nil, nil
	case ERROR:
		return nil, fmt.Errorf("token:%d:%s", p.token.Line, p.token.Text)
	case NUMBER:
		err := p.parseNum()
		if err != nil {
			return nil, err
		}
		return parseNounExpr, nil
	case SEMICOLON:
		if p.expr != nil {
			p.expr = nil
		}
		return nil, nil
	default:
		err := fmt.Errorf("unexpected token: %v", p.token)
		return nil, err
	}
}

func (p *Parser) parseNum() error {
	n := 0
	_, err := fmt.Sscanf(p.token.Text, "%v", &n) 
	if err != nil {
		return fmt.Errorf("token:%d:%s: %s", p.token.Line, p.token.Text, err)
	}
	p.expr = astInt{value: n}
	return nil
}

func parseNounExpr(p *Parser) (pStateFn, error) {
	p.next()
	switch p.token.Type {
	case VERB:
		//vt := p.token
		//p.next()
	}
	return nil, nil
}
