package goal

import (
	"encoding/csv"
	"io"
	"strings"
	"unicode/utf8"
)

func fCSV2(ctx *Context, x, y V) V {
	s, ok := x.bv.(S)
	if !ok {
		return panicType("x csv y", "x", x)
	}
	if utf8.RuneCountInString(string(s)) != 1 {
		return panics("x csv y : x is not a code point character")
	}
	c, _ := utf8.DecodeRuneInString(string(s))
	if c == utf8.RuneError {
		return panics("x csv y : x is not a valid code point character")
	}
	r := fCSV(ctx, c, y)
	if r.IsPanic() {
		s := string(r.bv.(panicV))
		return NewPanic("x csv y" + strings.TrimPrefix(s, "csv x"))
	}
	return r
}

func csvStringVs(x []V, ctx *Context) []string {
	r := make([]string, len(x))
	for i, xi := range x {
		switch xiv := xi.bv.(type) {
		case S:
			r[i] = string(xiv)
		default:
			r[i] = xi.Sprint(ctx)
		}
	}
	return r
}

func fCSV(ctx *Context, comma rune, x V) V {
	switch xv := x.bv.(type) {
	case S:
		sr := strings.NewReader(string(xv))
		csvr := csv.NewReader(sr)
		csvr.Comma = comma
		csvr.FieldsPerRecord = -1
		nlines := strings.Count(string(xv), "\n")
		r := make([]V, 0, nlines)
		for {
			record, err := csvr.Read()
			if err != nil {
				if err == io.EOF {
					return NewAV(r)
				}
				return Errorf("%v", err)
			}
			r = append(r, NewAS(record))
		}
	case *AS:
		sb := strings.Builder{}
		csvw := csv.NewWriter(&sb)
		csvw.Comma = comma
		csvw.Write(xv.elts)
		csvw.Flush()
		return NewS(sb.String())
	case *AV:
		sb := strings.Builder{}
		csvw := csv.NewWriter(&sb)
		csvw.Comma = comma
		for _, xi := range xv.elts {
			switch xiv := xi.bv.(type) {
			case *AS:
				csvw.Write(xiv.elts)
			case *AB:
				r := stringIntegers(xiv.elts)
				csvw.Write(r)
			case *AI:
				r := stringIntegers(xiv.elts)
				csvw.Write(r)
			case *AF:
				r := stringFloat64s(xiv.elts, ctx.Prec)
				csvw.Write(r)
			case *AV:
				r := csvStringVs(xiv.elts, ctx)
				csvw.Write(r)
			default:
				return panicType("csv x", "xi", xi)
			}
		}
		csvw.Flush()
		return NewS(sb.String())
	default:
		return panicType("csv x", "x", x)
	}
}
