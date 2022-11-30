package goal

import (
	"strings"
)

// enum returns !x.
func enum(x V) V {
	x = toIndices(x)
	if x.IsErr() {
		return errf("!x : %v", x)
	}
	switch xv := x.Value.(type) {
	case I:
		return rangeI(xv)
	case AI:
		return rangeArray(xv)
	default:
		return errs("!x : x nested array")
	}
}

func rangeI(n I) V {
	if n < 0 {
		return errs("!x : x negative")
	}
	r := make(AI, n)
	for i := range r {
		r[i] = i
	}
	return NewV(r)
}

func rangeArray(x AI) V {
	cols := 1
	for _, n := range x {
		if n == 0 {
			return NewV(AV{})
		}
		cols *= n
	}
	r := make(AV, x.Len())
	reps := cols
	for i := range r {
		a := make(AI, cols)
		reps /= x[i]
		clen := reps * x[i]
		for c := 0; c < cols/clen; c++ {
			col := c * clen
			for j := 0; j < x[i]; j++ {
				for k := 0; k < reps; k++ {
					a[col+j*reps+k] = j
				}
			}
		}
		r[i] = NewV(a)
	}
	return NewV(r)
}

// where returns &x.
func where(x V) V {
	switch x := x.Value.(type) {
	case I:
		switch {
		case x < 0:
			return errf("&x : x negative (%d)", x)
		case x == 0:
			return NewV(AI{})
		default:
			r := make(AI, x)
			return NewV(r)
		}
	case F:
		if !isI(x) {
			return errf("&x : x non-integer (%g)", x)
		}
		n := I(x)
		switch {
		case n < 0:
			return errf("&x : x negative (%d)", n)
		case n == 0:
			return NewV(AI{})
		default:
			r := make(AI, n)
			return NewV(r)
		}
	case AB:
		n := 0
		for _, xi := range x {
			n += int(B2I(xi))
		}
		r := make(AI, 0, n)
		for i, xi := range x {
			if xi {
				r = append(r, i)
			}
		}
		return NewV(r)
	case AI:
		n := 0
		for _, xi := range x {
			if xi < 0 {
				return errf("&x : x contains negative integer (%d)", x)
			}
			n += xi
		}
		r := make(AI, 0, n)
		for i, xi := range x {
			for j := 0; j < xi; j++ {
				r = append(r, i)
			}
		}
		return NewV(r)
	case AF:
		n := 0
		for _, xi := range x {
			if !isI(F(xi)) {
				return errf("&x : x contains non-integer (%g)", xi)
			}
			if xi < 0 {
				return errf("&x : x contains negative (%d)", int(xi))
			}
			n += int(xi)
		}
		r := make(AI, 0, n)
		for i, xi := range x {
			for j := 0; j < int(xi); j++ {
				r = append(r, i)
			}
		}
		return NewV(r)
	case AV:
		switch aType(x) {
		case tB, tF, tI:
			n := 0
			for _, xi := range x {
				switch xi := xi.Value.(type) {
				case F:
					if !isI(xi) {
						return errf("&x : not an integer (%g)", xi)
					}
					if xi < 0 {
						return errf("&x : negative integer (%d)", int(xi))
					}
					n += int(xi)
				case I:
					if xi < 0 {
						return errf("&x : negative integer (%d)", xi)
					}
					n += int(xi)
				}
			}
			r := make(AI, 0, n)
			for i, xi := range x {
				var max I
				switch xi := xi.Value.(type) {
				case I:
					max = xi
				case F:
					max = I(xi)
				}
				for j := 0; j < int(max); j++ {
					r = append(r, i)
				}
			}
			return NewV(r)
		default:
			return errs("&x : x non-integer")
		}
	default:
		return errs("&x : x non-integer")
	}
}

// replicate returns {x}#y.
func replicate(x, y V) V {
	switch x := x.Value.(type) {
	case I:
		switch {
		case x < 0:
			return errf("f#y : f[y] negative integer (%d)", x)
		default:
			return repeat(y, int(x))
		}
	case F:
		if !isI(x) {
			return errf("f#y : f[y] not an integer (%g)", x)
		}
		n := int(x)
		switch {
		case n < 0:
			return errf("f#y : f[y] negative (%d)", n)
		default:
			return repeat(y, n)
		}
	case AB:
		if x.Len() != Length(y) {
			return errf("f#y : length mismatch: %d (f[y]) vs %d (y)", x.Len(), Length(y))
		}
		return repeatAB(x, y)
	case AI:
		if x.Len() != Length(y) {
			return errf("f#y : length mismatch: %d (f[y]) vs %d (y)", x.Len(), Length(y))
		}
		return repeatAI(x, y)
	case AF:
		ix := toAI(x)
		if ix.IsErr() {
			return errf("f#y : x %v", ix)
		}
		return replicate(ix, y)
	case AV:
		// should be canonical
		assertCanonical(x)
		return errs("f#y : f[y] non-integer")
	default:
		return errs("f#y : f[y] non-integer")
	}
}

func repeat(x V, n int) V {
	switch xv := x.Value.(type) {
	case F:
		r := make(AF, n)
		for i := range r {
			r[i] = float64(xv)
		}
		return NewV(r)
	case I:
		if isBI(xv) {
			r := make(AB, n)
			for i := range r {
				r[i] = xv == 1
			}
			return NewV(r)
		}
		r := make(AI, n)
		for i := range r {
			r[i] = int(xv)
		}
		return NewV(r)
	case S:
		r := make(AS, n)
		for i := range r {
			r[i] = string(xv)
		}
		return NewV(r)
	default:
		r := make(AV, n)
		for i := range r {
			r[i] = x
		}
		return NewV(r)
	}
}

func repeatAB(x AB, y V) V {
	n := 0
	for _, xi := range x {
		n += int(B2I(xi))
	}
	switch y := y.Value.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, xi := range x {
			if xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AF:
		r := make(AF, 0, n)
		for i, xi := range x {
			if xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AI:
		r := make(AI, 0, n)
		for i, xi := range x {
			if xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AS:
		r := make(AS, 0, n)
		for i, xi := range x {
			if xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AV:
		r := make(AV, 0, n)
		for i, xi := range x {
			if xi {
				r = append(r, y.at(i))
			}
		}
		return NewV(canonical(r))
	default:
		return errs("f#y : y not an array")
	}
}

func repeatAI(x AI, y V) V {
	n := 0
	for _, xi := range x {
		if xi < 0 {
			return errf("f#y : f[y] contains negative integer (%d)", xi)
		}
		n += xi
	}
	switch y := y.Value.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, xi := range x {
			for j := 0; j < xi; j++ {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AF:
		r := make(AF, 0, n)
		for i, xi := range x {
			for j := 0; j < xi; j++ {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AI:
		r := make(AI, 0, n)
		for i, xi := range x {
			for j := 0; j < xi; j++ {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AS:
		r := make(AS, 0, n)
		for i, xi := range x {
			for j := 0; j < xi; j++ {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AV:
		r := make(AV, 0, n)
		for i, xi := range x {
			for j := 0; j < xi; j++ {
				r = append(r, y[i])
			}
		}
		return NewV(canonical(r))
	default:
		return errs("f#y : y not an array")
	}
}

// weedOut implements {x}_y
func weedOut(x, y V) V {
	switch x := x.Value.(type) {
	case I:
		if x != 0 {
			return NewV(AV{})
		}
		return y
	case F:
		if x != 0 {
			return NewV(AV{})
		}
		return y
	case AB:
		return weedOutAB(x, y)
	case AI:
		return weedOutAI(x, y)
	case AF:
		ix := toAI(x)
		if ix.IsErr() {
			return errf("f#y : x %v", ix)
		}
		return weedOut(ix, y)
	case AV:
		//assertCanonical(x)
		return errs("f#y : f[y] non-integer")
	default:
		return errs("f_y : f[y] non-integer")
	}
}

func weedOutAB(x AB, y V) V {
	n := 0
	for _, xi := range x {
		n += 1 - int(B2I(xi))
	}
	switch y := y.Value.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, xi := range x {
			if !xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AF:
		r := make(AF, 0, n)
		for i, xi := range x {
			if !xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AI:
		r := make(AI, 0, n)
		for i, xi := range x {
			if !xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AS:
		r := make(AS, 0, n)
		for i, xi := range x {
			if !xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AV:
		r := make(AV, 0, n)
		for i, xi := range x {
			if !xi {
				r = append(r, y.at(i))
			}
		}
		return NewV(canonical(r))
	default:
		return errs("f_y : y not an array")
	}
}

func weedOutAI(x AI, y V) V {
	n := 0
	for _, xi := range x {
		n += int(B2I(xi == 0))
	}
	switch y := y.Value.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, xi := range x {
			if xi == 0 {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AF:
		r := make(AF, 0, n)
		for i, xi := range x {
			if xi == 0 {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AI:
		r := make(AI, 0, n)
		for i, xi := range x {
			if xi == 0 {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AS:
		r := make(AS, 0, n)
		for i, xi := range x {
			if xi == 0 {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case AV:
		r := make(AV, 0, n)
		for i, xi := range x {
			if xi == 0 {
				r = append(r, y[i])
			}
		}
		return NewV(canonical(r))
	default:
		return errs("f_y : y not an array")
	}
}

// cast implements x$y.
func cast(x, y V) V {
	s, ok := x.Value.(S)
	if !ok {
		return errf("s$y : s not a string (%s)", x.Type())
	}
	switch s {
	case "i":
		return casti(y)
	case "n":
		return castn(y)
	case "s":
		return casts(y)
	default:
		return errf("s$y : unsupported \"%s\" value for s", s)
	}
}

func casti(y V) V {
	switch yv := y.Value.(type) {
	case I:
		return y
	case F:
		return NewI(int(yv))
	case S:
		runes := []rune(yv)
		r := make(AI, len(runes))
		for i, rc := range runes {
			r[i] = int(rc)
		}
		return NewV(r)
	case AB:
		return y
	case AI:
		return y
	case AS:
		r := make(AV, yv.Len())
		for i, s := range yv {
			r[i] = casti(NewS(s))
		}
		return NewV(r)
	case AF:
		return toAI(yv)
	case AV:
		r := make(AV, yv.Len())
		for i := range r {
			r[i] = casti(yv[i])
			if r[i].IsErr() {
				return r[i]
			}
		}
		return NewV(r)
	default:
		return errs("\"i\"$y : non-numeric y")
	}
}

func castn(y V) V {
	switch yv := y.Value.(type) {
	case I:
		return y
	case F:
		return y
	case S:
		xi, err := parseNumber(string(yv))
		if err != nil {
			return errf("\"i\"$y : non-numeric y (%s) : %v", yv, err)
		}
		return NewV(xi)
	case AB:
		return y
	case AI:
		return y
	case AS:
		r := make(AV, yv.Len())
		for i, s := range yv {
			n, err := parseNumber(s)
			if err != nil {
				return errf("\"i\"$y : y contains non-numeric (%s) : %v", s, err)
			}
			r[i] = NewV(n)
		}
		return NewV(canonical(r))
	case AF:
		return y
	case AV:
		r := make(AV, yv.Len())
		for i := range r {
			r[i] = castn(yv[i])
			if r[i].IsErr() {
				return r[i]
			}
		}
		return NewV(r)
	default:
		return errs("\"i\"$y : non-numeric y")
	}
}

func casts(y V) V {
	switch yv := y.Value.(type) {
	case I:
		return NewS(string(rune(yv)))
	case F:
		return casts(NewI(int(yv)))
	case AB:
		return casts(fromABtoAI(yv))
	case AI:
		sb := &strings.Builder{}
		for _, i := range yv {
			sb.WriteRune(rune(i))
		}
		return NewS(sb.String())
	case AF:
		return casts(toAI(yv))
	case AV:
		r := make(AV, yv.Len())
		for i := range r {
			r[i] = casts(yv[i])
			if r[i].IsErr() {
				return r[i]
			}
		}
		return NewV(r)
	default:
		return errs("\"i\"$y : non-numeric y")
	}
}

// eval implements .s.
func eval(ctx *Context, x V) V {
	//assertCanonical(x)
	nctx := ctx.derive()
	switch x := x.Value.(type) {
	case S:
		r, err := nctx.Eval(string(x))
		if err != nil {
			return errf(".s : %v", err)
		}
		ctx.merge(nctx)
		return r
	default:
		return errType(".x", "x", x)
	}
}
