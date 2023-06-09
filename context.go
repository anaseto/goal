// Copyright (c) 2022 Yon <anaseto@bardinflor.perso.aquilenet.fr>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

// Package goal provides an API to goal's interpreter.
//
// In order to evaluate code in the goal programming language, first a new
// context has to be created.
//
//	ctx := goal.NewContext()
//
// This context can then be used to Compile some code, and then Run it. It is
// possible to customize the context by registering new unary and binary
// operators using the RegisterMonad and RegisterDyad methods.
//
// See tests in context_test.go, as well as cmd/goal/main.go, for usage
// examples.
package goal

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strings"
)

// Context holds the state of the interpreter. Context values have to be
// created with NewContext.
type Context struct {
	// program representations (AST and compiled)
	gCode   *globalCode
	lambdas []*lambdaCode

	// execution and stack handling
	stack     []V
	frameIdx  int32
	callDepth int32
	lambda    int // currently executed lambda (if any)

	// values
	globals        []V            // global variables
	constants      []V            // constants
	sconstants     map[string]int // string constants index (avoiding dups)
	variadics      []VariadicFun  // variadic functions
	variadicsNames []string       // variadic function names

	// symbol handling
	gNames       []string       // ID: name
	gIDs         map[string]int // name: ID
	gPrefix      string         // current name prefix
	gAssignLists [][]int        // index: assign list ids

	// parsing, scanning
	scanner  *Scanner
	compiler *compiler
	fname    string               // filename
	sources  map[string]string    // filename: source
	keywords map[string]IdentType // special keyword names
	vNames   map[string]variadic  // variadic keywords

	// error positions stack
	errPos []position

	// rand
	rand *rand.Rand

	// miscellaneous
	sortBuf32  []int32 // radix sort buffer
	sortBuf16  []int16 // radix sort buffer
	sortBuf8   []int8  // radix sort buffer
	assigned   bool    // last instruction was opAssignGlobal
	compactFmt bool    // compact value sprint formatting

	Log  io.Writer // output writer for logging with \expr and rt.log
	Prec int       // floating point formatting precision (default: -1)
	OFS  string    // output field separator (default: " ")
}

// NewContext returns a new context for compiling and interpreting code, with
// default parameters.
func NewContext() *Context {
	ctx := &Context{}
	ctx.gCode = &globalCode{}
	ctx.stack = make([]V, 0, 32)
	ctx.gIDs = make(map[string]int, 8)
	ctx.sources = make(map[string]string, 4)
	ctx.constants = []V{constAV: NewV(&AV{flags: flagImmutable})}
	ctx.sconstants = map[string]int{}
	ctx.Prec = -1
	ctx.OFS = " "
	ctx.initVariadics()
	return ctx
}

const (
	constAV = iota
)

// RegisterMonad adds a variadic function to the context, and generates a new
// monadic keyword for that variadic (parsing will not search for a left
// argument). The variadic is also returned as a value.
// Note that while that a keyword defined in such a way will not take a left
// argument, it is still possible to pass several arguments to it with bracket
// indexing, like for any value.
func (ctx *Context) RegisterMonad(name string, vf VariadicFun) V {
	id := len(ctx.variadics)
	_, ok := ctx.keywords[name]
	if ok {
		panic(fmt.Sprintf("RegisterMonad: keyword %s already in use", name))
	}
	ctx.variadics = append(ctx.variadics, vf)
	ctx.variadicsNames = append(ctx.variadicsNames, name)
	ctx.keywords[name] = IdentMonad
	ctx.vNames[name] = variadic(id)
	return newVariadic(variadic(id))
}

func (ctx *Context) registerVariadic(name string, vf VariadicFun) V {
	id := len(ctx.variadics)
	ctx.variadics = append(ctx.variadics, vf)
	ctx.variadicsNames = append(ctx.variadicsNames, name)
	ctx.vNames[name] = variadic(id)
	return newVariadic(variadic(id))
}

// GetVariadic returns the variadic value registered with a given keyword or
// symbol, along its associated variadic function. It returns a zero value and
// nil function if there is no registered variadic with such name.
func (ctx *Context) GetVariadic(name string) (V, VariadicFun) {
	v, ok := ctx.vNames[name]
	if !ok {
		return V{}, nil
	}
	return newVariadic(v), ctx.variadics[v]
}

// RegisterDyad adds a variadic function to the context, and generates a new
// dyadic keyword for that variadic (parsing will search for a left argument).
// The variadic is also returned as a value.
func (ctx *Context) RegisterDyad(name string, vf VariadicFun) V {
	id := len(ctx.variadics)
	_, ok := ctx.keywords[name]
	if ok {
		panic(fmt.Sprintf("RegisterDyad: keyword %s already in use", name))
	}
	ctx.variadics = append(ctx.variadics, vf)
	ctx.variadicsNames = append(ctx.variadicsNames, name)
	ctx.keywords[name] = IdentDyad
	ctx.vNames[name] = variadic(id)
	return newVariadic(variadic(id))
}

// AssignGlobal assigns a value to a global variable name.
func (ctx *Context) AssignGlobal(name string, x V) {
	id := ctx.global(name)
	x.IncrRC()
	ctx.globals[id] = x
}

// GetGlobal returns the value attached to a global variable with the given
// name.
func (ctx *Context) GetGlobal(name string) (V, bool) {
	id, ok := ctx.gIDs[name]
	if !ok {
		return V{}, false
	}
	return ctx.globals[id], true
}

// Compile parses and compiles code from the given source string. The loc
// argument is the location used for error reporting and represents, usually a
// filename.
func (ctx *Context) Compile(loc string, s string) error {
	s = strings.Trim(s, " \n")
	if len(ctx.gCode.Body) > 0 {
		ctx.gCode.Body = ctx.gCode.Body[:0]
		ctx.gCode.Pos = ctx.gCode.Pos[:0]
		ctx.gCode.last = 0
	}
	ctx.fname = loc
	ctx.sources[loc] = s
	ctx.scanner = NewScanner(ctx.keywords, s)
	ctx.compiler = newCompiler(ctx)
	llen := len(ctx.lambdas)
	err := ctx.compiler.ParseCompile()
	if err != nil {
		ctx.gCode.Body = ctx.gCode.Body[:0]
		ctx.gCode.Pos = ctx.gCode.Pos[:0]
		ctx.gCode.last = 0
		ctx.lambdas = ctx.lambdas[:llen]
		ctx.assigned = false
		return ctx.getError(err, true)
	}
	ctx.checkAssign()
	return nil
}

// Run runs compiled code, if not already done, and returns the result value.
func (ctx *Context) Run() (V, error) {
	if len(ctx.gCode.Body) == 0 {
		return V{}, nil
	}
	err := ctx.exec()
	if err != nil {
		return V{}, err
	}
	return ctx.pop(), nil
}

// Eval calls Compile with the given string and an empty location, and then
// Run.  You cannot call it within a variadic function, as the evaluation is
// done on the current context, so it would interrupt compilation of current
// file.  Use EvalPackage for that.
func (ctx *Context) Eval(s string) (V, error) {
	ofname := ctx.fname
	defer func() {
		ctx.fname = ofname
	}()
	err := ctx.Compile("", s)
	if err != nil {
		return V{}, err
	}
	return ctx.Run()
}

// ErrPackageImported is returned by EvalPackage for packages that have already
// been processed (same location).
type ErrPackageImported struct{}

func (e ErrPackageImported) Error() string {
	return "ErrPackageImported"
}

// EvalPackage calls Compile with the string s as source, loc as error location
// (used for caching too, usually a filename), pfx as prefix for global
// variables (usually a filename without the extension), and then Run.  If a
// package with same location has already been evaluated, it returns
// ErrPackageImported. This means that even though Goal allows to evaluate
// (also via import) with the same location several times (which can be useful
// if separate files using the same package can be used together or alone),
// only the first one counts.  The package is evaluated in a derived context
// that is then merged on successful completion, so this function can be called
// within a variadic function.
func (ctx *Context) EvalPackage(s, loc, pfx string) (V, error) {
	ofname := ctx.fname
	defer func() {
		ctx.fname = ofname
	}()
	oprefix := ctx.gPrefix
	if pfx != "" {
		ctx.gPrefix = pfx
		defer func() {
			ctx.gPrefix = oprefix
		}()
	}
	_, ok := ctx.sources[loc]
	if ok {
		return NewI(0), ErrPackageImported{}
	}
	nctx := ctx.derive()
	err := nctx.Compile(loc, s)
	if err != nil {
		ctx.merge(nctx)
		return V{}, err
	}
	r, err := nctx.Run()
	ctx.merge(nctx)
	if err != nil {
		return V{}, err
	}
	return r, nil
}

// AssignedLast returns true if the last compiled expression was an assignment.
func (ctx *Context) AssignedLast() bool {
	return ctx.assigned
}

func (ctx *Context) checkAssign() {
	if len(ctx.gCode.Body) == 0 {
		ctx.assigned = false
	}
	switch ctx.gCode.Body[ctx.gCode.last] {
	case opAssignGlobal:
		ctx.assigned = true
	default:
		ctx.assigned = false
	}
}

func (ctx *Context) exec() error {
	ip, err := ctx.execute(ctx.gCode.Body)
	if err != nil {
		ctx.stack = ctx.stack[0:]
		ctx.push(V{})
		ctx.updateErrPos(ip, nil)
		ctx.gCode.Body = ctx.gCode.Body[:0]
		ctx.gCode.Pos = ctx.gCode.Pos[:0]
		ctx.gCode.last = 0
		return ctx.getError(err, false)
	}
	ctx.gCode.Body = ctx.gCode.Body[:0]
	ctx.gCode.Pos = ctx.gCode.Pos[:0]
	ctx.gCode.last = 0
	if len(ctx.stack) == 0 {
		// should not happen
		return ctx.getError(errors.New("no result: empty stack"), false)
	}
	return nil
}

func (ctx *Context) getError(err error, compile bool) error {
	e := &PanicError{
		Msg:       err.Error(),
		positions: ctx.errPos,
		sources:   ctx.sources,
		compile:   compile,
	}
	ctx.errPos = nil
	return e
}

func (ctx *Context) updateErrPos(ip int, lc *lambdaCode) {
	fname := ctx.fname
	if lc != nil {
		fname = lc.Filename
	}
	if len(ctx.gCode.Body) == 0 {
		// should not happen during execution
		ctx.errPos = append(ctx.errPos, position{Filename: fname})
		return
	}
	if lc != nil {
		if ip >= len(lc.Body) || ip < 0 {
			ip = len(lc.Body) - 1
		}
		pos := lc.Pos[ip]
		ctx.errPos = append(ctx.errPos, position{Filename: fname, Pos: pos, lambda: lc})
	} else {
		if ip >= len(ctx.gCode.Body) || ip < 0 {
			ip = len(ctx.gCode.Body) - 1
		}
		pos := ctx.gCode.Pos[ip]
		ctx.errPos = append(ctx.errPos, position{Filename: fname, Pos: pos})
	}
}

// Show returns a string representation with debug information about the
// context.
func (ctx *Context) Show() string {
	return ctx.programString()
}

func (ctx *Context) storeConst(x V) int {
	if ctx.compiler.scope() != nil {
		x.MarkImmutable()
	}
	ctx.constants = append(ctx.constants, x)
	return len(ctx.constants) - 1
}

func (ctx *Context) global(s string) int {
	if ctx.gPrefix != "" && !strings.ContainsRune(s, '.') {
		s = ctx.gPrefix + "." + s
	} else {
		s = strings.TrimPrefix(s, "main.") // main namespace
	}
	id, ok := ctx.gIDs[s]
	if ok {
		return id
	}
	ctx.globals = append(ctx.globals, V{})
	ctx.gIDs[s] = len(ctx.gNames)
	ctx.gNames = append(ctx.gNames, s)
	return len(ctx.gNames) - 1
}

// derive returns a context derived from ctx, suitable for eval.
func (ctx *Context) derive() *Context {
	nctx := &Context{}
	nctx.gCode = &globalCode{}
	nctx.stack = make([]V, 0, 32)

	nctx.variadics = ctx.variadics
	nctx.variadicsNames = ctx.variadicsNames
	nctx.keywords = ctx.keywords
	nctx.vNames = ctx.vNames
	nctx.rand = ctx.rand
	nctx.Log = ctx.Log

	nctx.constants = ctx.constants
	nctx.sconstants = ctx.sconstants
	nctx.lambdas = ctx.lambdas
	nctx.globals = ctx.globals
	nctx.gNames = ctx.gNames
	nctx.gIDs = ctx.gIDs
	nctx.sources = ctx.sources
	nctx.gPrefix = ctx.gPrefix
	nctx.Prec = ctx.Prec
	nctx.OFS = ctx.OFS
	return nctx
}

// merge integrates changes from a context created with derive.
func (ctx *Context) merge(nctx *Context) {
	ctx.constants = nctx.constants
	ctx.sconstants = nctx.sconstants
	ctx.lambdas = nctx.lambdas
	ctx.globals = nctx.globals
	ctx.gNames = nctx.gNames
	ctx.gIDs = nctx.gIDs
	ctx.sources = nctx.sources
	ctx.gPrefix = nctx.gPrefix
	ctx.Prec = nctx.Prec
	ctx.OFS = nctx.OFS
}
