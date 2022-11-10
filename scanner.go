package main

import (
	"bufio"
	"bytes"
	"io"
)

// Token represents a token information.
type Token struct {
	Type TokenType // token type
	Rune rune      // context text when only one rune is enough
	Pos  int       // token's offset in the source
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

// TokenType represents the different kinds of tokens.
type TokenType int32

// These constants describe the possible kinds of tokens.
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

// Init initializes the scanner with a given reader. It can be reused again
// with a new reader, but position information will be reset.
func (s *Scanner) Init(r io.Reader) {
	if s.bReader != nil {
		buf := s.buf
		*s = Scanner{}
		s.buf = buf
		s.buf.Reset()
	}
	s.bReader = bufio.NewReader(r)
	s.exprStart = true
	s.start = true
	s.token = Token{Type: EOF, Pos: s.pos}
}

// Next produces the next token from the input reader.
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
	s.token = Token{Type: t, Pos: s.pos}
	switch t {
	case LEFTBRACE, LEFTBRACKET, LEFTPAREN, NEWLINE, RIGHTBRACE, RIGHTBRACKET, RIGHTPAREN, SEMICOLON:
		// all of these don't have additional content, so we don't do
		// this test in the other emits.
		s.exprStart = true
	default:
		s.exprStart = false
	}
	s.exprStart = false
	return nil
}

func (s *Scanner) emitString(t TokenType) stateFn {
	s.token = Token{Type: t, Pos: s.pos, Text: s.buf.String()}
	s.exprStart = false
	s.buf.Reset()
	return nil
}

func (s *Scanner) emitRune(t TokenType, r rune) stateFn {
	s.token = Token{Type: t, Pos: s.pos, Rune: r}
	s.exprStart = false
	return nil
}

func (s *Scanner) emitEOF() stateFn {
	if s.err != nil {
		s.buf.Reset()
		s.buf.WriteString(s.err.Error())
		return s.emitString(ERROR)
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
		return s.emitRune(ADVERB, r)
	case '\'', '\\':
		return s.emitRune(ADVERB, r)
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
		if s.exprStart {
			return scanMinus
		}
		return s.emitRune(VERB, r)
	case ':', '+', '*', '%', '!', '&', '|', '<', '>',
		'=', '~', ',', '^', '#', '_', '$', '?', '@', '.':
		return s.emitRune(VERB, r)
	case '"':
		s.buf.WriteRune(r)
		return scanString
	case '`':
		s.buf.WriteRune('`')
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
		return s.emitString(ERROR)
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
			s.next()
			return scanMinus
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
			s.buf.WriteString("non terminated string: unexpected EOF")
			return s.emitString(ERROR)
		case '\\':
			s.buf.WriteRune(r)
			nr := s.peek()
			if nr == '"' {
				s.buf.WriteRune(nr)
				s.next()
			}
		case '"':
			s.buf.WriteRune(r)
			return s.emitString(STRING)
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
			s.buf.WriteString("non terminated string: unexpected EOF")
			return s.emitString(ERROR)
		case !isAlpha(r) && (s.buf.Len() == 0) || !isAlphaNum(r):
			s.buf.WriteRune('`')
			return s.emitString(STRING)
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
			return s.emitString(NUMBER)
		case r == '.':
			s.buf.WriteRune(r)
			s.next()
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
			return s.emitString(NUMBER)
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
			return s.emitString(NUMBER)
		case !isDigit(r):
			return s.emitString(NUMBER)
		default:
			s.buf.WriteRune(r)
			s.next()
		}
	}
}

func scanMinus(s *Scanner) stateFn {
	r := s.peek()
	if isDigit(r) {
		s.buf.WriteRune('-')
		return scanNumber
	}
	return s.emitRune(VERB, '-')
}

func scanIdent(s *Scanner) stateFn {
	for {
		r := s.peek()
		switch {
		case r == eof:
			return s.emitString(IDENT)
		case r == '.':
			r = s.peek()
			if !isAlpha(r) {
				return s.emitString(IDENT)
			}
			s.buf.WriteRune('.')
			s.buf.WriteRune(r)
			s.next()
		case isAlphaNum(r):
			s.buf.WriteRune(r)
			s.next()
		default:
			return s.emitString(IDENT)
		}
	}
}
