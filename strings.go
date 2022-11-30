package goal

func applyS(s S, x V) V {
	switch xv := x.Value.(type) {
	case I:
		if xv < 0 {
			xv += I(len(s))
		}
		if xv < 0 || xv > I(len(s)) {
			return errf("s[i] : i out of bounds index (%d)", xv)
		}
		return NewV(s[xv:])
	case F:
		if !isI(xv) {
			return errf("s[x] : x non-integer (%g)", xv)
		}
		return applyS(s, x)
	case AB:
		return applyS(s, fromABtoAI(xv))
	case AI:
		r := make(AS, xv.Len())
		for i, n := range xv {
			if n < 0 {
				n += len(s)
			}
			if n < 0 || n > len(s) {
				return errf("s[i] : i out of bounds index (%d)", n)
			}
			r[i] = string(s[n:])
		}
		return NewV(r)
	case AF:
		z := toAI(xv)
		if isErr(z) {
			return z
		}
		return applyS(s, z)
	case AV:
		r := make(AV, xv.Len())
		for i, xi := range xv {
			r[i] = applyS(s, xi)
			if isErr(r[i]) {
				return r[i]
			}
		}
		return NewV(canonical(r))
	default:
		return errf("s[x] : x non-integer (%s)", xv.Type())
	}
}

func applyS2(s S, x V, y V) V {
	var l int
	switch y := y.Value.(type) {
	case I:
		if y < 0 {
			return errf("s[x;y] : y negative (%d)", y)
		}
		l = int(y)
	case F:
		if !isI(y) {
			return errf("s[x;y] : y non-integer (%g)", y)
		}
		l = int(y)
	case AI:
	case AB:
		if Length(x) != y.Len() {
		}
		return applyS2(s, x, fromABtoAI(y))
	case AF:
		z := toAI(y)
		if isErr(z) {
			return z
		}
		return applyS2(s, x, z)
	default:
		return errType("s[x;y]", "y", y)
	}
	switch xv := x.Value.(type) {
	case I:
		if xv < 0 {
			xv += I(len(s))
		}
		if xv < 0 || xv > I(len(s)) {
			return errf("s[i;y] : i out of bounds index (%d)", xv)
		}
		if _, ok := y.Value.(AI); ok {
			return errf("s[x;y] : x is an atom but y is an array")
		}
		if int(xv)+l > len(s) {
			l = len(s) - int(xv)
		}
		return NewV(s[xv : int(xv)+l])
	case F:
		if !isI(xv) {
			return errf("s[x;y] : x non-integer (%g)", xv)
		}
		return applyS2(s, x, y)
	case AB:
		return applyS2(s, fromABtoAI(xv), y)
	case AI:
		r := make(AS, xv.Len())
		if z, ok := y.Value.(AI); ok {
			if z.Len() != xv.Len() {
				return errf("s[x;y] : length mismatch: %d (#x) %d (#y)",
					xv.Len(), z.Len())
			}
			for i, n := range xv {
				if n < 0 {
					n += len(s)
				}
				if n < 0 || n > len(s) {
					return errf("s[i;y] : i out of bounds index (%d)", n)
				}
				l := z[i]
				if n+l > len(s) {
					l = len(s) - n
				}
				r[i] = string(s[n : n+l])
			}
			return NewV(r)
		}
		for i, n := range xv {
			if n < 0 {
				n += len(s)
			}
			if n < 0 || n > len(s) {
				return errf("s[i;y] : i out of bounds index (%d)", n)
			}
			l := l
			if n+l > len(s) {
				l = len(s) - n
			}
			r[i] = string(s[n : n+l])
		}
		return NewV(r)
	case AF:
		z := toAI(xv)
		if isErr(z) {
			return z
		}
		return applyS2(s, z, y)
	case AV:
		r := make(AV, xv.Len())
		for i, xi := range xv {
			r[i] = applyS2(s, xi, y)
			if isErr(r[i]) {
				return r[i]
			}
		}
		return NewV(canonical(r))
	default:
		return errf("s[x;y] : x non-integer (%s)", xv.Type())
	}
}

func bytes(x V) V {
	switch x := x.Value.(type) {
	case S:
		return NewI(len(x))
	case AS:
		r := make(AI, x.Len())
		for i, s := range x {
			r[i] = len(s)
		}
		return NewV(r)
	case AV:
		r := make(AV, x.Len())
		for i, xi := range x {
			r[i] = bytes(xi)
			if isErr(r[i]) {
				return r[i]
			}
		}
		return NewV(canonical(r))
	default:
		return errType("bytes x", "x", x)
	}
}
