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
	Pos  int       // token's line in the source
	Text string    // content text (identifier, string, number)
}

type TokenType int

const (
	EOF TokenType = iota
	ERROR
	ADVERB
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
	ctx     *Context      // unused (for now)
	reader  io.Reader     // reader to scan from
	wError  io.Writer     // writer for scanning errors
	bReader *bufio.Reader // buffered reader
	buf     bytes.Buffer  // buffer
	peeked  bool          // peeked next
	pos     int           // current position in the input
	pr      rune          // peeked rune
	psize   int
	start   bool // at line start
	token   Token
}

type stateFn func(*Scanner) stateFn

func (s *Scanner) Init() {
	s.bReader = bufio.NewReader(s.reader)
	if s.wError == nil {
		s.wError = os.Stderr
	}
	s.token = Token{Type: EOF, Pos: s.pos}
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
	// TODO: in case of error, read the file again to get from pos the line
	// and print the line that produced the error with some column marker.
	fmt.Fprintf(s.wError, "scan error:%d: %s\n", s.pos, msg)
}

const eof = -1

func (s *Scanner) peek() rune {
	if s.peeked {
		return s.pr
	}
	r, size, err := s.bReader.ReadRune()
	if err != nil {
		if err != io.EOF {
			s.error(err.Error())
		}
		r = eof
	}
	s.peeked = true
	s.pr = r
	s.psize = size
	return s.pr
}

func (s *Scanner) next() rune {
	if s.peeked {
		s.updateInfo(s.pr)
		s.peeked = false
		s.pos += s.psize
		return s.pr
	}
	r, sz, err := s.bReader.ReadRune()
	s.pos += sz
	if err != nil {
		// end of file
		if err != io.EOF {
			s.error(err.Error())
		}
		return eof
	}
	s.updateInfo(r)
	//fmt.Printf("[%c]", r)
	return r
}

func (s *Scanner) updateInfo(r rune) {
	if r == '\n' {
		s.start = true
	} else {
		s.start = false
	}
}

func (s *Scanner) emit(t TokenType) stateFn {
	s.token = Token{t, s.pos, s.buf.String()}
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
	case '/':
		if s.start {
			return scanCommentLine
		}
		s.buf.WriteRune(r)
		return s.emit(ADVERB)
	case '\'', '\\':
		s.buf.WriteRune(r)
		return s.emit(ADVERB)
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
	case ':', '+', '-', '*', '%', '!', '&', '|', '<', '>',
		'=', '~', ',', '^', '#', '_', '$', '?', '@', '.':
		s.buf.WriteRune(r)
		return s.emit(VERB)
	case '"':
		return scanString
	case '`':
		return scanSymbolString
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

func scanCommentLine(s *Scanner) stateFn {
	r := s.peek()
	if r == '\n' {
		return scanMultiLineComment
	}
	return scanComment
}

func scanMultiLineComment(s *Scanner) stateFn {
	for {
		r := s.next()
		switch {
		case r == eof:
			return s.emit(EOF)
		case r == '\\' && s.start:
			r := s.next()
			if r == '\n' {
				return scanAny
			}
		}
	}
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

func scanSymbolString(s *Scanner) stateFn {
	for {
		r := s.peek()
		switch {
		case r == eof:
			return s.emit(ERROR)
		case !isAlpha(r) && (s.buf.Len() == 0) || !isAlphaNum(r):
			return s.emit(STRING)
		default:
			s.next()
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
