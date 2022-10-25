package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

// Token represents a token information.
type Token struct {
	Type TokenType // token type
	Line int       // token's line in the source
	Text string    // content text (identifier, string, number)
}

type TokenType int

const (
	EOF TokenType = iota
	ERROR
	ADVERB
	COLON
	DOUBLECOLON
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
	reader  io.Reader     // reader to scan from
	wError  io.Writer     // writer for scanning errors
	bReader *bufio.Reader // buffered reader
	buf     bytes.Buffer  // buffer
	pos     int           // current position in the input
	line    int           // current line number
	token   Token
}

type stateFn func(*Scanner) stateFn

func (s *Scanner) Init() {
	s.bReader = bufio.NewReader(s.reader)
	s.line = 1
	if s.wError == nil {
		s.wError = os.Stderr
	}
	s.token = Token{Type: EOF, Line: s.line}
}

func (s *Scanner) Next() Token {
	state := scanAny
	for {
		state = state(s)
		if state == nil {
			return s.token
		}
	}
}

func (s *Scanner) error(msg string) {
	if s.wError == nil {
		return
	}
	line := s.line
	if s.token.Type == NEWLINE {
		line--
	}
	fmt.Fprintf(s.wError, "scan error:%d: %s\n", line, msg)
}

const eof = -1

func (s *Scanner) peek() rune {
	r, _, err := s.bReader.ReadRune()
	if err != nil {
		return eof
	}
	s.bReader.UnreadRune()
	return r
}

func (s *Scanner) next() rune {
	r, sz, err := s.bReader.ReadRune()
	s.pos += sz
	if err != nil {
		// end of file
		if err != io.EOF {
			s.error(err.Error())
		}
		return eof
	}
	//fmt.Printf("[%c]", r)
	if r == '\n' {
		s.line++
	}
	return r
}

func (s *Scanner) emit(t TokenType) stateFn {
	s.token = Token{t, s.line, s.buf.String()}
	s.buf.Reset()
	return nil
}

func scanAny(s *Scanner) stateFn {
	r := s.next()
	switch r {
	case eof:
		return s.emit(EOF)
	case '\n':
		return s.emit(NEWLINE)
	case ' ', '\t':
		return scanSpace
	case '\'', '/', '\\':
		s.buf.WriteRune(r)
		return scanAdverb
	case ':':
		// TODO: actually, definitions should follow identifier
		s.buf.WriteRune(r)
		return scanColon
	case '{':
		return s.emit(LEFTBRACE)
	case '[':
		return s.emit(LEFTBRACKET)
	case '(':
		return s.emit(LEFTPAREN)
	case '}':
		return s.emit(RIGHTBRACE)
	case ']':
		return s.emit(RIGHTBRACKET)
	case ')':
		return s.emit(RIGHTPAREN)
	case ';':
		return s.emit(SEMICOLON)
	case '+', '-', '*', '%', '!', '&', '|', '<', '>',
		'=', '~', ',', '^', '#', '_', '$', '?', '@', '.':
		s.buf.WriteRune(r)
		return scanVerb
	case '"':
		return scanString
	}
	switch {
	case isDigit(r):
		s.buf.WriteRune(r)
		return scanNumber
	case isAlpha(r):
		s.buf.WriteRune(r)
		return scanIdent
	default:
		s.buf.WriteRune(r)
		return s.emit(ERROR)
	}
	return nil
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAlpha(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func isAlphaNum(r rune) bool {
	return isAlpha(r) || isDigit(r)
}

func scanSpace(s *Scanner) stateFn {
	for {
		r := s.peek()
		switch {
		case r == eof:
			return scanAny
		case r == '/':
			s.next()
			return scanComment
		case r == ' ' || r == '\t':
			s.next()
		default:
			return scanAny
		}
	}
}

func scanComment(s *Scanner) stateFn {
	for {
		r := s.next()
		switch r {
		case eof:
			return s.emit(EOF)
		case '\n':
			return s.emit(NEWLINE)
		}
	}
}

func scanColon(s *Scanner) stateFn {
	r := s.peek()
	switch r {
	case ':':
		s.buf.WriteRune(':')
		s.next()
		return s.emit(DOUBLECOLON)
	default:
		return s.emit(COLON)
	}
}

func scanAdverb(s *Scanner) stateFn {
	r := s.peek()
	if r == ':' {
		s.buf.WriteRune(':')
		s.next()
	}
	return s.emit(ADVERB)
}

func scanVerb(s *Scanner) stateFn {
	r := s.peek()
	if r == ':' {
		s.buf.WriteRune(':')
		s.next()
	}
	return s.emit(VERB)
}

func scanString(s *Scanner) stateFn {
	for {
		r := s.next()
		switch r {
		case eof:
			return s.emit(ERROR)
		case '"':
			return s.emit(STRING)
		default:
			s.buf.WriteRune(r)
		}
	}
}

func scanNumber(s *Scanner) stateFn {
	for {
		r := s.peek()
		switch {
		case r == eof:
			return s.emit(NUMBER)
		case !isDigit(r):
			return s.emit(NUMBER)
		default:
			s.buf.WriteRune(r)
			s.next()
		}
	}
}

func scanIdent(s *Scanner) stateFn {
	for {
		r := s.peek()
		switch {
		case r == eof:
			return s.emit(IDENT)
		case !isAlphaNum(r):
			return s.emit(IDENT)
		default:
			s.buf.WriteRune(r)
			s.next()
		}
	}
}
