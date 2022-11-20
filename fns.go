package goal

import (
	"strings"
)

// enum returns !x.
func enum(x V) V {
	switch x := x.(type) {
	case F:
		if !isI(x) {
			return errs("!x : x non-integer")
		}
		// TODO: check whether range is too big?
		return rangeI(I(x))
	case I:
		return rangeI(x)
	case array:
		return rangeArray(x)
	default:
		return errType("!x", "x", x)
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
	return r
}

func rangeArray(x array) V {
	z, ok := x.(AI)
	if !ok {
		z = make(AI, x.Len())
		for i := range z {
			v := x.at(i)
			switch v := v.(type) {
			case I:
				z[i] = int(v)
			case F:
				if !isI(v) {
					return errs("!x : x contains non-integer")
				}
				z[i] = int(v)
			default:
				return errs("!x : x is not numeric")
			}
		}
	}
	cols := 1
	for _, n := range z {
		if n == 0 {
			return AV{}
		}
		cols *= n
	}
	r := make(AV, x.Len())
	reps := cols
	for i := range r {
		a := make(AI, cols)
		reps /= z[i]
		clen := reps * z[i]
		for c := 0; c < cols/clen; c++ {
			col := c * clen
			for j := 0; j < z[i]; j++ {
				for k := 0; k < reps; k++ {
					a[col+j*reps+k] = j
				}
			}
		}
		r[i] = a
	}
	return r
}

// where returns &x.
func where(x V) V {
	switch x := x.(type) {
	case I:
		switch {
		case x < 0:
			return errf("&x : x negative (%d)", x)
		case x == 0:
			return AI{}
		default:
			r := make(AI, x)
			return r
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
			return AI{}
		default:
			r := make(AI, n)
			return r
		}
	case AB:
		n := 0
		for _, v := range x {
			n += int(B2I(v))
		}
		r := make(AI, 0, n)
		for i, v := range x {
			if v {
				r = append(r, i)
			}
		}
		return r
	case AI:
		n := 0
		for _, v := range x {
			if v < 0 {
				return errf("&x : x contains negative integer (%d)", x)
			}
			n += v
		}
		r := make(AI, 0, n)
		for i, v := range x {
			for j := 0; j < v; j++ {
				r = append(r, i)
			}
		}
		return r
	case AF:
		n := 0
		for _, v := range x {
			if !isI(F(v)) {
				return errf("&x : x contains non-integer (%g)", v)
			}
			if v < 0 {
				return errf("&x : x contains negative (%d)", int(v))
			}
			n += int(v)
		}
		r := make(AI, 0, n)
		for i, v := range x {
			for j := 0; j < int(v); j++ {
				r = append(r, i)
			}
		}
		return r
	case AV:
		switch aType(x) {
		case tB, tF, tI:
			n := 0
			for _, v := range x {
				switch v := v.(type) {
				case F:
					if !isI(v) {
						return errf("&x : not an integer (%g)", v)
					}
					if v < 0 {
						return errf("&x : negative integer (%d)", int(v))
					}
					n += int(v)
				case I:
					if v < 0 {
						return errf("&x : negative integer (%d)", v)
					}
					n += int(v)
				}
			}
			r := make(AI, 0, n)
			for i, v := range x {
				var max I
				switch v := v.(type) {
				case I:
					max = v
				case F:
					max = I(v)
				}
				for j := 0; j < int(max); j++ {
					r = append(r, i)
				}
			}
			return r
		default:
			return errs("&x : x non-integer")
		}
	default:
		return errs("&x : x non-integer")
	}
}

// replicate returns {x}#y.
func replicate(x, y V) V {
	switch x := x.(type) {
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
		z := toAI(x)
		if err, ok := z.(errV); ok {
			return errf("f#y : x %v", err)
		}
		return replicate(z, y)
	case AV:
		z := canonical(x)
		if _, ok := z.(AV); ok {
			return errs("f#y : f[y] non-integer")
		}
		return replicate(z, y)
	default:
		return errs("f#y : f[y] non-integer")
	}
}

func repeat(x V, n int) V {
	switch x := x.(type) {
	case F:
		r := make(AF, n)
		for i := range r {
			r[i] = float64(x)
		}
		return r
	case I:
		if isBI(x) {
			r := make(AB, n)
			for i := range r {
				r[i] = x == 1
			}
			return r
		}
		r := make(AI, n)
		for i := range r {
			r[i] = int(x)
		}
		return r
	case S:
		r := make(AS, n)
		for i := range r {
			r[i] = string(x)
		}
		return r
	default:
		r := make(AV, n)
		for i := range r {
			r[i] = x
		}
		return r
	}
}

func repeatAB(x AB, y V) V {
	n := 0
	for _, v := range x {
		n += int(B2I(v))
	}
	switch y := y.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, v := range x {
			if v {
				r = append(r, y[i])
			}
		}
		return r
	case AF:
		r := make(AF, 0, n)
		for i, v := range x {
			if v {
				r = append(r, y[i])
			}
		}
		return r
	case AI:
		r := make(AI, 0, n)
		for i, v := range x {
			if v {
				r = append(r, y[i])
			}
		}
		return r
	case AS:
		r := make(AS, 0, n)
		for i, v := range x {
			if v {
				r = append(r, y[i])
			}
		}
		return r
	case AV:
		r := make(AV, 0, n)
		for i, v := range x {
			if v {
				r = append(r, y.at(i))
			}
		}
		return r
	default:
		return errs("f#y : y not an array")
	}
}

func repeatAI(x AI, y V) V {
	n := 0
	for _, v := range x {
		if v < 0 {
			return errf("f#y : f[y] contains negative integer (%d)", v)
		}
		n += v
	}
	switch y := y.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, v := range x {
			for j := 0; j < v; j++ {
				r = append(r, y[i])
			}
		}
		return r
	case AF:
		r := make(AF, 0, n)
		for i, v := range x {
			for j := 0; j < v; j++ {
				r = append(r, y[i])
			}
		}
		return r
	case AI:
		r := make(AI, 0, n)
		for i, v := range x {
			for j := 0; j < v; j++ {
				r = append(r, y[i])
			}
		}
		return r
	case AS:
		r := make(AS, 0, n)
		for i, v := range x {
			for j := 0; j < v; j++ {
				r = append(r, y[i])
			}
		}
		return r
	case AV:
		r := make(AV, 0, n)
		for i, v := range x {
			for j := 0; j < v; j++ {
				r = append(r, y[i])
			}
		}
		return r
	default:
		return errs("f#y : y not an array")
	}
}

// weedOut implements {x}_y
func weedOut(x, y V) V {
	switch x := x.(type) {
	case I:
		if x != 0 {
			return AV{}
		}
		return y
	case F:
		if x != 0 {
			return AV{}
		}
		return y
	case AB:
		return weedOutAB(x, y)
	case AI:
		return weedOutAI(x, y)
	case AF:
		z := toAI(x)
		if err, ok := z.(errV); ok {
			return errf("f#y : x %v", err)
		}
		return weedOut(z, y)
	case AV:
		z := canonical(x)
		if _, ok := z.(AV); ok {
			return errs("f#y : f[y] non-integer")
		}
		return weedOut(z, y)
	default:
		return errs("f_y : f[y] non-integer")
	}
}

func weedOutAB(x AB, y V) V {
	n := 0
	for _, v := range x {
		n += 1 - int(B2I(v))
	}
	switch y := y.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, v := range x {
			if !v {
				r = append(r, y[i])
			}
		}
		return r
	case AF:
		r := make(AF, 0, n)
		for i, v := range x {
			if !v {
				r = append(r, y[i])
			}
		}
		return r
	case AI:
		r := make(AI, 0, n)
		for i, v := range x {
			if !v {
				r = append(r, y[i])
			}
		}
		return r
	case AS:
		r := make(AS, 0, n)
		for i, v := range x {
			if !v {
				r = append(r, y[i])
			}
		}
		return r
	case AV:
		r := make(AV, 0, n)
		for i, v := range x {
			if !v {
				r = append(r, y.at(i))
			}
		}
		return r
	default:
		return errs("f_y : y not an array")
	}
}

func weedOutAI(x AI, y V) V {
	n := 0
	for _, v := range x {
		n += int(B2I(v == 0))
	}
	switch y := y.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, v := range x {
			if v == 0 {
				r = append(r, y[i])
			}
		}
		return r
	case AF:
		r := make(AF, 0, n)
		for i, v := range x {
			if v == 0 {
				r = append(r, y[i])
			}
		}
		return r
	case AI:
		r := make(AI, 0, n)
		for i, v := range x {
			if v == 0 {
				r = append(r, y[i])
			}
		}
		return r
	case AS:
		r := make(AS, 0, n)
		for i, v := range x {
			if v == 0 {
				r = append(r, y[i])
			}
		}
		return r
	case AV:
		r := make(AV, 0, n)
		for i, v := range x {
			if v == 0 {
				r = append(r, y[i])
			}
		}
		return r
	default:
		return errs("f#y : y not an array")
	}
}

// cast implements x$y.
func cast(x, y V) V {
	s, ok := x.(S)
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
	switch z := y.(type) {
	case I:
		return y
	case F:
		return I(z)
	case S:
		runes := []rune(z)
		res := make(AI, len(runes))
		for i, r := range runes {
			res[i] = int(r)
		}
		return res
	case AB:
		return y
	case AI:
		return y
	case AS:
		res := make(AV, z.Len())
		for i, s := range z {
			res[i] = casti(S(s))
		}
		return res
	case AF:
		return toAI(z)
	case AV:
		res := make(AV, z.Len())
		for i := range res {
			res[i] = casti(z[i])
			if err, ok := res[i].(errV); ok {
				return err
			}
		}
		return res
	default:
		return errs("\"i\"$y : non-numeric y")
	}
}

func castn(y V) V {
	switch z := y.(type) {
	case I:
		return y
	case F:
		return y
	case S:
		v, err := parseNumber(string(z))
		if err != nil {
			return errf("\"i\"$y : non-numeric y (%s) : %v", z, err)
		}
		return v
	case AB:
		return y
	case AI:
		return y
	case AS:
		res := make(AV, z.Len())
		for i, s := range z {
			n, err := parseNumber(s)
			if err != nil {
				return errf("\"i\"$y : y contains non-numeric (%s) : %v", s, err)
			}
			res[i] = n
		}
		return canonical(res)
	case AF:
		return y
	case AV:
		res := make(AV, z.Len())
		for i := range res {
			res[i] = castn(z[i])
			if err, ok := res[i].(errV); ok {
				return err
			}
		}
		return res
	default:
		return errs("\"i\"$y : non-numeric y")
	}
}

func casts(y V) V {
	switch z := y.(type) {
	case I:
		return S(rune(z))
	case F:
		return casts(I(z))
	case AB:
		return casts(fromABtoAI(z))
	case AI:
		sb := &strings.Builder{}
		for _, i := range z {
			sb.WriteRune(rune(i))
		}
		return S(sb.String())
	case AF:
		return casts(toAI(z))
	case AV:
		res := make(AV, z.Len())
		for i := range res {
			res[i] = casts(z[i])
			if err, ok := res[i].(errV); ok {
				return err
			}
		}
		return res
	default:
		return errs("\"i\"$y : non-numeric y")
	}
}

// eval implements .s.
func eval(ctx *Context, x V) V {
	x = canonical(x)
	nctx := ctx.derive()
	switch x := x.(type) {
	case S:
		v, err := nctx.Eval(string(x))
		if err != nil {
			return errf(".s : %v", err)
		}
		ctx.merge(nctx)
		return v
	default:
		return errType(".x", "x", x)
	}
}

// try implements .[f1;x;f2].
func try(ctx *Context, f1, x, f2 V) V {
	av := toArray(x).(array)
	for i := av.Len() - 1; i >= 0; i-- {
		ctx.push(av.at(i))
	}
	res := ctx.applyN(f1, av.Len())
	if err, ok := res.(errV); ok {
		ctx.push(S(err))
		return ctx.applyN(f2, 1)
	}
	return res
}
