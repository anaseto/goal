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
	col     int           // current column number
	colprev int           // previous column number
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

func (s *Scanner) error(msg string) {
	line := s.line
	col := s.col
	if s.ch == '\n' {
		line--
		col = s.prevcol
	}
	fmt.Fprintf(s.Werror, "scan error:%d:%d: %s\n", line, col, msg)
}

const eof = -1

func (s *Scanner) next() rune {
	r, _, err := s.bReader.ReadRune()
	if err != nil {
		// end of file
		//s.state = scanEnd
		if err != io.EOF {
			s.error(err.Error())
		}
		return eof
	}
	//fmt.Printf("[%c]", r)
	if r == '\n' {
		s.line++
		s.colprev = s.col
		s.col = 0
	} else {
		s.col++
	}
	return r
}

func (s *Scanner) emit(t TokenType) stateFn {
	s.token = Token{t, t.line, s.buf.String()}
	s.buf.Reset()
	return nil
}

func scanAny(s *Scanner) stateFn {
	switch r := s.next(); r {
	case eof:
		return nil
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
	//switch {
	//case
	//}
	// TODO: identifiers, numbers
	return nil
}

func scanSpace(s *Scanner) stateFn {
	for {
		r, _, err := s.bReader.ReadRune()
		switch {
		case err != nil:
			s.bReader.UnreadRune()
			return scanAny
		case r == '/':
			s.col++
			return scanComment
		case r == ' ' || r == '\t':
			s.col++
		default:
			s.bReader.UnreadRune()
			return scanAny
		}
	}
}

func scanComment(s *Scanner) stateFn {
	for {
		r, _, err := s.bReader.ReadRune()
		if err != nil || r == '\n' {
			s.bReader.UnreadRune()
			return scanAny
		}
		s.col++
	}
}

func scanColon(s *Scanner) stateFn {
	r, _, err := s.bReader.ReadRune()
	if err != nil || r != ':' {
		s.bReader.UnreadRune()
		return s.emit(COLON)
	}
	s.buf.WriteRune(':')
	return s.emit(DOUBLECOLON)
}

func scanAdverb(s *Scanner) stateFn {
	r, _, err := s.bReader.ReadRune()
	if err != nil || r != ':' {
		s.bReader.UnreadRune()
	} else {
		s.buf.WriteRune(':')
	}
	return s.emit(ADVERB)
}

func scanVerb(s *Scanner) stateFn {
	r, _, err := s.bReader.ReadRune()
	if err != nil || r != ':' {
		s.bReader.UnreadRune()
	} else {
		s.buf.WriteRune(':')
	}
	return s.emit(VERB)
}

func scanString(s *Scanner) stateFn {
	// TODO: improve
	for {
		r, _, err := s.bReader.ReadRune()
		switch {
		case err != nil:
			s.bReader.UnreadRune()
			return scanAny
		case r == '"':
			return s.emit(STRING)
		default:
			s.col++
			s.buf.WriteRune(r)
		}
	}
}
