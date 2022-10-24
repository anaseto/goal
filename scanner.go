package main

import (
	"bufio"
	"bytes"
	"io"
)

// Token represents a token information.
type Token struct {
	Type TokenType
	Line int
	Text string
}

type TokenType int

const (
	EOF TokenType = iota
	ERROR
	ADVERB
	COLON
	IDENT
	LEFTBRACE
	LEFTBRACKET
	LEFTPAREN
	NEWLINE
	NUMBER
	RIGHTBRACE
	RIGHTBRACKET
	RIGHTPAREN
	SEMICOLON
	STRING
	VERB
)

var tokenStrings = [...]string{
	EOF:          "EOF",
	ERROR:        "ERROR",
	ADVERB:       "ADVERB",
	COLON:        ":",
	IDENT:        "IDENT",
	LEFTBRACE:    "{",
	LEFTBRACKET:  "[",
	LEFTPAREN:    "(",
	NEWLINE:      "NEWLINE",
	NUMBER:       "NUMBER",
	RIGHTBRACE:   "}",
	RIGHTBRACKET: "]",
	RIGHTPAREN:   ")",
	SEMICOLON:    ";",
	STRING:       "STRING",
	VERB:         "VERB",
}

func (t TokenType) String() string {
	return tokenStrings[t]
}

// Scanner represents the state of the scanner.
type Scanner struct {
	ctx     *Context
	reader  io.Reader // reader to scan from
	bReader *bufio.Reader
	buf     bytes.Buffer // buffer
	cr      rune         // current rune
	col     int          // current column number
	colprev int          // previous column number
	line    int          // current line number
	token   Token
}

type stateFn func(*Scanner) stateFn

func (s *Scanner) Init() {
	s.bReader = bufio.NewReader(s.reader)
	s.line = 1
}

func (s *Scanner) Next() Token {
	s.token = Token{Type: EOF, Line: s.line}
	state := scanAny
	for {
		state = state(s)
		if state == nil {
			return s.token
		}
	}
}

func (s *Scanner) next() rune {
	r, _, err := s.bReader.ReadRune()
	if err != nil {
		s.cr = -1 // end of file
		//s.state = scanEnd
		//if err != io.EOF {
		//s.error(err.Error())
		//}
		return 0
	}
	//fmt.Printf("[%c]", r)
	s.cr = r
	if r == '\n' {
		s.line++
		s.colprev = s.col
		s.col = 0
	} else {
		s.col++
	}
	return s.cr
}

func scanAny(s *Scanner) stateFn {
	return nil
}
