package main

import (
	"bufio"
	"bytes"
	"io"
)

// Token represents a token information.
type Token struct {
	Type TokenType // token type
	Pos  int       // token's line in the source
	Text string    // content text (identifier, string, number)
}

func (t Token) String() string {
	switch t.Type {
	case ERROR:
		return "error:" + t.Text
	case ADVERB, IDENT, VERB, NUMBER:
		return t.Text
	case LEFTBRACE:
		return "{"
	case LEFTBRACKET:
		return "["
	case LEFTPAREN:
		return "("
	case RIGHTBRACE:
		return "}"
	case RIGHTBRACKET:
		return "]"
	case RIGHTPAREN:
		return ")"
	case SEMICOLON:
		return ";"
	case STRING:
		return "\"" + t.Text + "\""
	default:
		return t.Type.String()
	}
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

// Scanner represents the state of the scanner.
type Scanner struct {
	bReader   *bufio.Reader // buffered reader
	buf       bytes.Buffer  // buffer
	err       error         // scanning error (if any)
	peeked    bool          // peeked next
	pos       int           // current position in the input
	pr        rune          // peeked rune
	psize     int           // size of last peeked rune
	start     bool          // at line start
	exprStart bool          // at expression start
	token     Token         // last token
}

type stateFn func(*Scanner) stateFn

func (s *Scanner) Init(r io.Reader) {
	s.bReader = bufio.NewReader(r)
	s.exprStart = true
	s.start = true
	s.err = nil
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

const eof = -1

func (s *Scanner) peek() rune {
	if s.peeked {
		return s.pr
	}
	r, size, err := s.bReader.ReadRune()
	if err != nil {
		if err != io.EOF {
			s.err = err
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
		if err != io.EOF {
			s.err = err
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
	switch t {
	case LEFTBRACE, LEFTBRACKET, LEFTPAREN, NEWLINE, RIGHTBRACE, RIGHTBRACKET, RIGHTPAREN, SEMICOLON:
		s.exprStart = true
	default:
		s.exprStart = false
	}
	s.buf.Reset()
	return nil
}

func (s *Scanner) emitEOF() stateFn {
	if s.err != nil {
		s.buf.Reset()
		s.buf.WriteString(s.err.Error())
		return s.emit(ERROR)
	}
	return s.emit(EOF)
}

func scanAny(s *Scanner) stateFn {
	r := s.next()
	switch r {
	case eof:
		return s.emitEOF()
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
	case '-':
		s.buf.WriteRune(r)
		if s.exprStart {
			return scanMinus
		}
		return s.emit(VERB)
	case ':', '+', '*', '%', '!', '&', '|', '<', '>',
		'=', '~', ',', '^', '#', '_', '$', '?', '@', '.':
		s.buf.WriteRune(r)
		return s.emit(VERB)
	case '"':
		s.buf.WriteRune(r)
		return scanString
	case '`':
		s.buf.WriteRune('\'')
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
		switch r {
		case eof:
			return scanAny
		case '/':
			s.next()
			return scanComment
		case ' ', '\t':
			s.next()
		case '-':
			r = s.peek()
			if isDigit(r) {
				s.buf.WriteRune(r)
				return scanMinus
			}
			return scanAny
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
			return s.emitEOF()
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
			return s.emitEOF()
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
		case '\\':
			s.buf.WriteRune(r)
			nr := s.peek()
			if nr == '"' {
				s.buf.WriteRune(nr)
				s.next()
			}
		case '"':
			s.buf.WriteRune(r)
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
			s.buf.WriteRune('\'')
			return s.emit(STRING)
		default:
			s.next()
			s.buf.WriteRune(r)
		}
	}
}

func scanMinus(s *Scanner) stateFn {
	r := s.peek()
	if isDigit(r) {
		return scanNumber
	}
	return s.emit(VERB)
}

func scanNumber(s *Scanner) stateFn {
	for {
		r := s.peek()
		switch {
		case r == eof:
			return s.emit(NUMBER)
		case r == 'e':
			s.buf.WriteRune(r)
			s.next()
			r = s.peek()
			if r == '+' || r == '-' {
				s.buf.WriteRune(r)
				s.next()
				return scanExponent
			}
		case !isAlphaNum(r):
			return s.emit(NUMBER)
		default:
			s.buf.WriteRune(r)
			s.next()
		}
	}
}

func scanExponent(s *Scanner) stateFn {
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
