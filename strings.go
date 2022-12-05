package goal

import "strings"

func applyS(s S, x V) V {
	if x.IsInt() {
		xv := x.Int()
		if xv < 0 {
			xv += int64(len(s))
		}
		if xv < 0 || xv > int64(len(s)) {
			return errf("s[i] : i out of bounds index (%d)", xv)
		}
		return NewV(s[xv:])
	}
	switch xv := x.Value.(type) {
	case F:
		if !isI(xv) {
			return errf("s[x] : x non-integer (%g)", xv)
		}
		return applyS(s, x)
	case *AB:
		return applyS(s, fromABtoAI(xv))
	case *AI:
		r := make([]string, xv.Len())
		for i, n := range xv.Slice {
			if n < 0 {
				n += int64(len(s))
			}
			if n < 0 || n > int64(len(s)) {
				return errf("s[i] : i out of bounds index (%d)", n)
			}
			r[i] = string(s[n:])
		}
		return NewAS(r)
	case *AF:
		z := toAI(xv)
		if z.IsErr() {
			return z
		}
		return applyS(s, z)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = applyS(s, xi)
			if r[i].IsErr() {
				return r[i]
			}
		}
		return canonicalV(NewAV(r))
	default:
		return errf("s[x] : x non-integer (%s)", x.Type())
	}
}

func applyS2(s S, x V, y V) V {
	var l int64
	if y.IsInt() {
		if y.Int() < 0 {
			return errf("s[x;y] : y negative (%d)", y.Int())
		}
		l = y.Int()
	} else {
		switch yv := y.Value.(type) {
		case F:
			if !isI(yv) {
				return errf("s[x;y] : y non-integer (%g)", yv)
			}
			l = int64(yv)
		case *AI:
		case *AB:
			if Length(x) != yv.Len() {
			}
			return applyS2(s, x, fromABtoAI(yv))
		case *AF:
			z := toAI(yv)
			if z.IsErr() {
				return z
			}
			return applyS2(s, x, z)
		default:
			return errType("s[x;y]", "y", y)
		}
	}
	if x.IsInt() {
		xv := x.Int()
		if xv < 0 {
			xv += int64(len(s))
		}
		if xv < 0 || xv > int64(len(s)) {
			return errf("s[i;y] : i out of bounds index (%d)", xv)
		}
		if _, ok := y.Value.(*AI); ok {
			return errf("s[x;y] : x is an atom but y is an array")
		}
		if xv+l > int64(len(s)) {
			l = int64(len(s)) - xv
		}
		return NewV(s[xv : xv+l])

	}
	switch xv := x.Value.(type) {
	case F:
		if !isI(xv) {
			return errf("s[x;y] : x non-integer (%g)", xv)
		}
		return applyS2(s, x, y)
	case *AB:
		return applyS2(s, fromABtoAI(xv), y)
	case *AI:
		r := make([]string, xv.Len())
		if z, ok := y.Value.(*AI); ok {
			if z.Len() != xv.Len() {
				return errf("s[x;y] : length mismatch: %d (#x) %d (#y)",
					xv.Len(), z.Len())
			}
			for i, n := range xv.Slice {
				if n < 0 {
					n += int64(len(s))
				}
				if n < 0 || n > int64(len(s)) {
					return errf("s[i;y] : i out of bounds index (%d)", n)
				}
				l := z.At(i)
				if n+l > int64(len(s)) {
					l = int64(len(s)) - n
				}
				r[i] = string(s[n : n+l])
			}
			return NewAS(r)
		}
		for i, n := range xv.Slice {
			if n < 0 {
				n += int64(len(s))
			}
			if n < 0 || n > int64(len(s)) {
				return errf("s[i;y] : i out of bounds index (%d)", n)
			}
			l := l
			if n+l > int64(len(s)) {
				l = int64(len(s)) - n
			}
			r[i] = string(s[n : n+l])
		}
		return NewAS(r)
	case *AF:
		z := toAI(xv)
		if z.IsErr() {
			return z
		}
		return applyS2(s, z, y)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = applyS2(s, xi, y)
			if r[i].IsErr() {
				return r[i]
			}
		}
		return canonicalV(NewAV(r))
	default:
		return errf("s[x;y] : x non-integer (%s)", x.Type())
	}
}

func bytes(x V) V {
	switch xv := x.Value.(type) {
	case S:
		return NewI(int64(len(xv)))
	case *AS:
		r := make([]int64, xv.Len())
		for i, s := range xv.Slice {
			r[i] = int64(len(s))
		}
		return NewAI(r)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = bytes(xi)
			if r[i].IsErr() {
				return r[i]
			}
		}
		return canonicalV(NewAV(r))
	default:
		return errType("bytes x", "x", x)
	}
}

// cast implements s$y.
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
	if y.IsInt() {
		return y
	}
	switch yv := y.Value.(type) {
	case F:
		return NewI(int64(yv))
	case S:
		runes := []rune(yv)
		r := make([]int64, len(runes))
		for i, rc := range runes {
			r[i] = int64(rc)
		}
		return NewAI(r)
	case *AB:
		return y
	case *AI:
		return y
	case *AS:
		r := make([]V, yv.Len())
		for i, s := range yv.Slice {
			r[i] = casti(NewS(s))
		}
		return NewAV(r)
	case *AF:
		return toAI(yv)
	case *AV:
		r := make([]V, yv.Len())
		for i := range r {
			r[i] = casti(yv.At(i))
			if r[i].IsErr() {
				return r[i]
			}
		}
		return NewAV(r)
	default:
		return errs("\"i\"$y : non-numeric y")
	}
}

func castn(y V) V {
	if y.IsInt() {
		return y
	}
	switch yv := y.Value.(type) {
	case F:
		return y
	case S:
		xi, err := parseNumber(string(yv))
		if err != nil {
			return errf("\"i\"$y : non-numeric y (%s) : %v", yv, err)
		}
		return xi
	case *AB:
		return y
	case *AI:
		return y
	case *AS:
		r := make([]V, yv.Len())
		for i, s := range yv.Slice {
			n, err := parseNumber(s)
			if err != nil {
				return errf("\"i\"$y : y contains non-numeric (%s) : %v", s, err)
			}
			r[i] = n
		}
		return canonicalV(NewAV(r))
	case *AF:
		return y
	case *AV:
		r := make([]V, yv.Len())
		for i := range r {
			r[i] = castn(yv.At(i))
			if r[i].IsErr() {
				return r[i]
			}
		}
		return canonicalV(NewAV(r))
	default:
		return errs("\"i\"$y : non-numeric y")
	}
}

func casts(y V) V {
	if y.IsInt() {
		return NewS(string(rune(y.Int())))
	}
	switch yv := y.Value.(type) {
	case F:
		return casts(NewI(int64(yv)))
	case *AB:
		return casts(fromABtoAI(yv))
	case *AI:
		sb := &strings.Builder{}
		for _, i := range yv.Slice {
			sb.WriteRune(rune(i))
		}
		return NewS(sb.String())
	case *AF:
		return casts(toAI(yv))
	case *AV:
		r := make([]V, yv.Len())
		for i := range r {
			r[i] = casts(yv.At(i))
			if r[i].IsErr() {
				return r[i]
			}
		}
		return canonicalV(NewAV(r))
	default:
		return errs("\"i\"$y : non-numeric y")
	}
}

func drops(s S, y V) V {
	switch yv := y.Value.(type) {
	case S:
		return NewS(strings.TrimPrefix(string(yv), string(s)))
	case *AS:
		r := make([]string, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = strings.TrimPrefix(string(yi), string(s))
		}
		return NewAS(r)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = drops(s, yi)
			if r[i].IsErr() {
				return r[i]
			}
		}
		return NewAV(r)
	default:
		return errType("s_y", "y", y)
	}
}

// trim returns s^y.
func trim(s S, y V) V {
	switch yv := y.Value.(type) {
	case S:
		return NewS(strings.Trim(string(yv), string(s)))
	case *AS:
		r := make([]string, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = strings.Trim(string(yi), string(s))
		}
		return NewAS(r)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = trim(s, yi)
			if r[i].IsErr() {
				return r[i]
			}
		}
		return NewAV(r)
	default:
		return errType("s^y", "y", y)
	}
}
