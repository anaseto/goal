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
	depth []Token
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
	case EOF, NEWLINE, SEMICOLON:
		return ppToken{Type: ppSEP, Line: tok.Line}, nil
	case ERROR:
		return nil, p.errorf("invalid token:%s:%s", tok.Type, tok.Text)
	case ADVERB:
		return nil, p.errorf("syntax:adverb %s at start of expression", tok.Text)
	case IDENT:
		return ppToken{Type: ppIDENT, Line: tok.Line, Text: tok.Text}, nil
	case LEFTBRACE, LEFTBRACKET, LEFTPAREN:
		return p.ppExprBlock()
	case NUMBER, STRING:
		return p.ppExprStrand()
	case RIGHTBRACE, RIGHTBRACKET, RIGHTPAREN:
		if len(p.depth) == 0 {
			return nil, p.errorf("syntax:unexpected %s without opening matching pair", tok.Text)
		}
		opTok := p.depth[len(p.depth)-1]
		clTokt := closeToken(opTok.Type)
		if clTokt != tok.Type {
			return nil, p.errorf("syntax:unexpected %s without closing previous %s at %d", tok.Text, opTok.Type.String(), opTok.Line)
		}
		p.depth = p.depth[:len(p.depth)-1]
		return ppToken{Type: ppCLOSE, Line: tok.Line}, nil
	case VERB:
		return ppToken{Type: ppVERB, Line: tok.Line, Text: tok.Text}, nil
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
		if ok && (tok.Type == ppCLOSE) {
			return ppb, nil
		}
		ppb.ppexprs = append(ppb.ppexprs, ppe)
	}
}

func (p *Parser) ppExprStrand() (ppExpr, error) {
	return nil, nil
}
