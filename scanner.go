package goal

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Token represents a token information.
type Token struct {
	Type TokenType // token type
	Pos  int       // token's offset in the source
	Text string    // content text (identifier, string, number)
}

func (t Token) String() string {
	switch t.Type {
	case ADVERB, ERROR, IDENT, DYAD, NUMBER, REGEXP, QQSTART, QQEND:
		return fmt.Sprintf("{%s %s}", t.Type.String(), t.Text)
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
type TokenType int

// These constants describe the possible kinds of tokens.
const (
	NONE TokenType = iota
	EOF
	ERROR
	ADVERB
	DYAD
	DYADASSIGN
	IDENT
	LEFTBRACE
	LEFTBRACKET
	LEFTPAREN
	NEWLINE
	NUMBER
	MONAD
	REGEXP
	RIGHTBRACE
	RIGHTBRACKET
	RIGHTPAREN
	SEMICOLON
	STRING
	QQSTART
	QQEND
)

// IdentType represents the different kinds of special roles for alphanumeric
// identifiers that act as keywords.
type IdentType int

// These constants represent the different kinds of special names.
const (
	IdentVar   IdentType = iota // a normal identifier (default zero value)
	IdentMonad                  // a builtin monad (cannot have left argument)
	IdentDyad                   // a builtin dyad (can have left argument)
)

// Scanner represents the state of the scanner.
type Scanner struct {
	names   map[string]IdentType // special keywords
	reader  *strings.Reader      // rune reader
	buf     *bytes.Buffer        // buffer for building strings
	err     error                // scanning error (if any)
	npos    int                  // position of next rune in the input
	epos    int                  // next token end position
	tpos    int                  // next token start position
	psize   int                  // size of last peeked rune
	pr      rune                 // peeked rune
	r       rune                 // current rune
	exprEnd bool                 // at expression start
	delimOp bool                 // at list start
	peeked  bool                 // peeked next
	start   bool                 // at line start
	token   Token                // last token
	source  string               // source string
	state   stateFn              // starting state for Next
	qr      rune                 // quote closing rune
}

type stateFn func(*Scanner) stateFn

// NewScanner returns a scanner for the given source string.
func NewScanner(names map[string]IdentType, source string) *Scanner {
	s := &Scanner{names: names}
	s.source = source
	s.reader = strings.NewReader(source)
	s.buf = &bytes.Buffer{}
	s.start = true
	s.state = scanAny
	s.next()
	return s
}

// Next produces the next token from the input reader.
func (s *Scanner) Next() Token {
	state := s.state
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
	r, size, err := s.reader.ReadRune()
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

func (s *Scanner) next() {
	if s.peeked {
		s.peeked = false
		s.r = s.pr
		s.epos = s.npos
		s.npos += s.psize
		return
	}
	r, sz, err := s.reader.ReadRune()
	s.r = r
	s.epos = s.npos
	s.npos += sz
	if err != nil {
		if err != io.EOF {
			s.err = err
		}
		s.r = eof
	}
	//fmt.Printf("[%c]", r)
}

func (s *Scanner) emit(t TokenType) stateFn {
	s.token = Token{Type: t, Pos: s.tpos}
	s.start = t == NEWLINE
	s.delimOp = t == LEFTPAREN || t == LEFTBRACKET || t == LEFTBRACE
	switch t {
	case LEFTBRACE, LEFTBRACKET, LEFTPAREN, NEWLINE, SEMICOLON, EOF:
		// all of these don't have additional content, so we don't do
		// this test in the other emits.
		s.exprEnd = false
	default:
		s.exprEnd = true
	}
	return nil
}

func (s *Scanner) emitError(err string) stateFn {
	s.token = Token{Type: ERROR, Pos: s.tpos, Text: err}
	return nil
}

func (s *Scanner) emitString(t TokenType) stateFn {
	s.token = Token{Type: t, Pos: s.tpos, Text: s.source[s.tpos:s.epos]}
	s.start = false
	s.delimOp = false
	s.exprEnd = true
	return nil
}

func (s *Scanner) emitRegexp(text string) stateFn {
	s.token = Token{Type: REGEXP, Pos: s.tpos, Text: text}
	s.start = false
	s.delimOp = false
	s.exprEnd = true
	return nil
}

func (s *Scanner) emitNewString(t TokenType, text string) stateFn {
	s.token = Token{Type: t, Pos: s.tpos, Text: text}
	s.start = false
	s.delimOp = false
	s.exprEnd = true
	return nil
}

func (s *Scanner) emitIDENT() stateFn {
	if s.tpos == s.epos {
		return s.emitError("empty identifier")
	}
	switch s.names[s.source[s.tpos:s.epos]] {
	case IdentDyad:
		return s.emitOp(DYAD)
	case IdentMonad:
		return s.emitOp(MONAD)
	default:
		return s.emitString(IDENT)
	}
}

func (s *Scanner) emitOp(t TokenType) stateFn {
	s.token = Token{Type: t, Pos: s.tpos, Text: s.source[s.tpos:s.epos]}
	s.start = false
	s.delimOp = false
	s.exprEnd = false
	return nil
}

func (s *Scanner) emitEOF() stateFn {
	if s.err != nil {
		return s.emitError(s.err.Error())
	}
	return s.emit(EOF)
}

func scanAny(s *Scanner) stateFn {
	s.tpos = s.epos
	switch s.r {
	case eof:
		return s.emitEOF()
	case '\n':
		s.next()
		if !s.start {
			if !s.delimOp {
				return s.emit(NEWLINE)
			}
			s.start = true
		}
		return scanAny
	case ' ', '\t':
		return scanSpace
	case '/':
		if s.start {
			return scanCommentLine
		}
		s.next()
		return s.emitOp(ADVERB)
	case '\'', '\\':
		s.next()
		return s.emitOp(ADVERB)
	case '{':
		s.next()
		return s.emit(LEFTBRACE)
	case '[':
		s.next()
		return s.emit(LEFTBRACKET)
	case '(':
		s.next()
		return s.emit(LEFTPAREN)
	case '}':
		s.next()
		return s.emit(RIGHTBRACE)
	case ']':
		s.next()
		return s.emit(RIGHTBRACKET)
	case ')':
		s.next()
		return s.emit(RIGHTPAREN)
	case ';':
		s.next()
		return s.emit(SEMICOLON)
	case '-':
		if !s.exprEnd {
			return scanMinus
		}
		return scanDyadOp
	case ':':
		s.next()
		if s.r == ':' {
			s.next()
		}
		return s.emitOp(DYAD)
	case '+', '*', '%', '!', '&', '|', '<', '>',
		'=', '~', ',', '^', '#', '_', '$', '?', '@', '.', '»', '«':
		return scanDyadOp
	case '"':
		s.qr = s.r
		s.state = scanQQ
		s.next()
		return s.emit(QQSTART)
	case '`':
		return scanBackQuotedString
	}
	switch {
	case isDigit(s.r):
		return scanNumber
	case isAlpha(s.r):
		return scanIdent
	default:
		return s.emitError(fmt.Sprintf("unexpected character: %c", s.r))
	}
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

func scanDyadOp(s *Scanner) stateFn {
	s.next()
	if s.r == ':' {
		s.next()
		if s.r == ':' {
			s.next()
		}
		return s.emitOp(DYADASSIGN)
	}
	return s.emitOp(DYAD)
}

func scanSpace(s *Scanner) stateFn {
	for {
		s.next()
		switch s.r {
		case '/':
			return scanComment
		case ' ', '\t':
		case '-':
			s.tpos = s.epos
			return scanMinus
		default:
			return scanAny
		}
	}
}

func scanComment(s *Scanner) stateFn {
	for {
		s.next()
		switch s.r {
		case eof:
			return s.emitEOF()
		case '\n':
			s.next()
			if !s.start {
				if !s.delimOp {
					return s.emit(NEWLINE)
				}
				s.start = true
			}
			return scanAny
		}
	}
}

func scanCommentLine(s *Scanner) stateFn {
	s.next()
	if s.r == '\n' {
		return scanMultiLineComment
	}
	return scanComment
}

func scanMultiLineComment(s *Scanner) stateFn {
	for {
		s.next()
		switch {
		case s.r == eof:
			return s.emitEOF()
		case s.r == '\\' && s.start:
			s.next()
			if s.r == '\n' {
				s.next()
				return scanAny
			}
		}
	}
}

func scanBackQuotedString(s *Scanner) stateFn {
	for {
		s.next()
		switch s.r {
		case eof:
			return s.emitError("non terminated string: unexpected EOF")
		case '`':
			s.next()
			return s.emitString(STRING)
		}
	}
}

func scanRawString(s *Scanner) stateFn {
	s.buf.Reset()
	s.buf.WriteByte('"')
	for {
		s.next()
		switch s.r {
		case eof:
			return s.emitError("non terminated string: unexpected EOF")
		case '\n':
			s.buf.WriteString(`\n`)
		case '"':
			s.buf.WriteString(`\"`)
		case '\\':
			s.buf.WriteString(`\\`)
		case s.qr:
			if s.peek() == s.qr {
				s.next()
				s.buf.WriteRune(s.r)
				break
			}
			s.buf.WriteString(`"`)
			s.next()
			return s.emitNewString(STRING, s.buf.String())
		default:
			s.buf.WriteRune(s.r)
		}
	}
}

func scanRegexp(s *Scanner) stateFn {
	s.buf.Reset()
	for {
		s.next()
		switch s.r {
		case eof:
			return s.emitError("non terminated rx/PATTERN/: unexpected EOF")
		case '\\':
			nr := s.peek()
			if nr == s.qr {
				s.next()
			} else if nr == '\\' {
				s.next()
				s.buf.WriteRune(nr)
			}
			s.buf.WriteRune(s.r)
		case s.qr:
			s.next()
			return s.emitRegexp(s.buf.String())
		default:
			s.buf.WriteRune(s.r)
		}
	}
}

func scanNumber(s *Scanner) stateFn {
	for {
		s.next()
		switch {
		case s.r == eof:
			return s.emitString(NUMBER)
		case s.r == '.':
		case s.r == 'e' || s.r == 'E' ||
			s.r == 'p' || s.r == 'P':
			r := s.peek()
			if r == '+' || r == '-' {
				s.next()
				return scanExponent
			}
		case !isAlphaNum(s.r):
			return s.emitString(NUMBER)
		}
	}
}

func scanExponent(s *Scanner) stateFn {
	for {
		s.next()
		switch {
		case s.r == eof:
			return s.emitString(NUMBER)
		case !isDigit(s.r):
			return s.emitString(NUMBER)
		}
	}
}

func scanMinus(s *Scanner) stateFn {
	s.next()
	if isDigit(s.r) {
		return scanNumber
	}
	return s.emitOp(DYAD)
}

func scanIdent(s *Scanner) stateFn {
	dots := 0
	const delimchars = ":+-*%!&|=~,^#_?@/'"
	for {
		s.next()
		switch {
		case s.r == eof:
			return s.emitIDENT()
		case isAlphaNum(s.r):
		case s.r == '.':
			r := s.peek()
			if !isAlpha(r) {
				return s.emitIDENT()
			}
			if dots > 0 {
				return s.emitError("too many dots in identifier")
			}
			dots++
		case s.source[s.tpos:s.epos] == "qq":
			if strings.ContainsRune(delimchars, s.r) {
				s.qr = s.r
				s.state = scanQQ
				s.next()
				return s.emit(QQSTART)
			}
			return s.emitIDENT()
		case s.source[s.tpos:s.epos] == "rq":
			if strings.ContainsRune(delimchars, s.r) {
				s.qr = s.r
				return scanRawString
			}
			return s.emitIDENT()
		case s.source[s.tpos:s.epos] == "rx":
			if strings.ContainsRune(delimchars, s.r) {
				s.qr = s.r
				return scanRegexp
			}
			return s.emitIDENT()
		default:
			return s.emitIDENT()
		}
	}
}

func scanQQ(s *Scanner) stateFn {
	s.tpos = s.epos
	switch s.r {
	case eof:
		return s.emitError("non terminated string: unexpected EOF")
	case s.qr:
		s.state = scanAny
		s.next()
		return s.emit(QQEND)
	case '$':
		s.tpos++
		return scanQQIdent(s)
	default:
		s.buf.Reset()
		s.buf.WriteByte('"')
		for {
			switch s.r {
			case eof:
				return s.emitError("non terminated string: unexpected EOF")
			case '\n':
				s.buf.WriteString(`\n`)
			case '\\':
				nr := s.peek()
				switch {
				case nr == '\\' || nr == '"' && s.qr == '"':
					s.buf.WriteByte('\\')
					s.buf.WriteRune(nr)
					s.next()
				case nr == s.qr || nr == '$':
					s.buf.WriteRune(nr)
					s.next()
				case nr == '\n':
				default:
					s.buf.WriteRune(s.r)
				}
			case s.qr:
				s.buf.WriteByte('"')
				return s.emitNewString(STRING, s.buf.String())
			case '"':
				s.buf.WriteString(`\"`)
			case '$':
				s.buf.WriteByte('"')
				return s.emitNewString(STRING, s.buf.String())
			default:
				s.buf.WriteRune(s.r)
			}
			s.next()
		}
	}
}

func scanQQIdent(s *Scanner) stateFn {
	dots := 0
	braced := false
	if s.peek() == '{' {
		s.next()
		s.tpos++
		braced = true
	}
	for {
		s.next()
		switch {
		case s.r == eof:
			if braced {
				return s.emitError("interpolation ${ without closing }")
			}
			return s.emitIDENT()
		case s.r == '.':
			r := s.peek()
			if !isAlpha(r) && !braced {
				return s.emitIDENT()
			}
			if dots > 0 {
				return s.emitError("too many dots in identifier")
			}
			dots++
		case isAlphaNum(s.r):
		case braced && s.r == '}':
			f := s.emitIDENT()
			s.next()
			return f
		default:
			if braced {
				return s.emitError("invalid char in ${IDENT}")
			}
			return s.emitIDENT()
		}
	}
}
