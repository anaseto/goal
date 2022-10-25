package main

import "io"

type Parser struct {
	ctx    *Context  // unused (for now)
	Source string    // for error messages location information (e.g. filename)
	wError io.Writer // where non-fatal error messages go (unused for now)
	line   int
	s      *Scanner
	token  Token
}
