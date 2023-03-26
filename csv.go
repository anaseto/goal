package goal

import (
	"encoding/csv"
	"io"
	"strings"
	"unicode/utf8"
)

func fCSV2(x, y V) V {
	s, ok := x.value.(S)
	if !ok {
		return Panicf("x csv y : x not a string (%s)", x.Type())
	}
	if utf8.RuneCountInString(string(s)) != 1 {
		return panics("x csv y : x not a code point character")
	}
	c, _ := utf8.DecodeRuneInString(string(s))
	if c == utf8.RuneError {
		return panics("x csv y : x not a valid code point character")
	}
	r := fCSV(c, y)
	if r.IsPanic() {
		s := string(r.value.(panicV))
		return NewPanic("x csv y" + strings.TrimPrefix(s, "csv x"))
	}
	return r
}

func fCSV(comma rune, x V) V {
	switch xv := x.value.(type) {
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
		t := rType(xv)
		if t != tAS {
			return Panicf("csv x : not an array of records")
		}
		sb := strings.Builder{}
		csvw := csv.NewWriter(&sb)
		csvw.Comma = comma
		for _, xi := range xv.elts {
			xi := xi.value.(*AS)
			csvw.Write(xi.elts)
		}
		csvw.Flush()
		return NewS(sb.String())
	default:
		return panicType("csv x", "x", x)
	}
}
