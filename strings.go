package goal

func applyS(s S, x V) V {
	switch x := x.(type) {
	case I:
		if x < 0 {
			x += I(len(s))
		}
		if x < 0 || x > I(len(s)) {
			return errf("s[i] : i out of bounds index (%d)", x)
		}
		return s[x:]
	case F:
		if !isI(x) {
			return errf("s[x] : x non-integer (%g)", x)
		}
		return applyS(s, x)
	case AB:
		return applyS(s, fromABtoAI(x))
	case AI:
		r := make(AS, x.Len())
		for i, n := range x {
			if n < 0 {
				n += len(s)
			}
			if n < 0 || n > len(s) {
				return errf("s[i] : i out of bounds index (%d)", n)
			}
			r[i] = string(s[n:])
		}
		return r
	case AF:
		z := toAI(x)
		if err, ok := z.(errV); ok {
			return err
		}
		return applyS(s, z)
	case AV:
		r := make(AV, x.Len())
		for i, v := range x {
			r[i] = applyS(s, v)
			if err, ok := r[i].(errV); ok {
				return err
			}
		}
		return canonical(r)
	default:
		return errf("s[x] : x non-integer (%s)", x.Type())
	}
}

func applyS2(s S, x V, y V) V {
	var l int
	switch y := y.(type) {
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
		if err, ok := z.(errV); ok {
			return err
		}
		return applyS2(s, x, z)
	default:
		return errType("s[x;y]", "y", y)
	}
	switch x := x.(type) {
	case I:
		if x < 0 {
			x += I(len(s))
		}
		if x < 0 || x > I(len(s)) {
			return errf("s[i;y] : i out of bounds index (%d)", x)
		}
		if _, ok := y.(AI); ok {
			return errf("s[x;y] : x is an atom but y is an array")
		}
		if int(x)+l > len(s) {
			l = len(s) - int(x)
		}
		return s[x : int(x)+l]
	case F:
		if !isI(x) {
			return errf("s[x;y] : x non-integer (%g)", x)
		}
		return applyS2(s, x, y)
	case AB:
		return applyS2(s, fromABtoAI(x), y)
	case AI:
		r := make(AS, x.Len())
		if z, ok := y.(AI); ok {
			if z.Len() != x.Len() {
				return errf("s[x;y] : length mismatch: %d (#x) %d (#y)",
					x.Len(), z.Len())
			}
			for i, n := range x {
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
			return r
		}
		for i, n := range x {
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
		return r
	case AF:
		z := toAI(x)
		if err, ok := z.(errV); ok {
			return err
		}
		return applyS2(s, z, y)
	case AV:
		r := make(AV, x.Len())
		for i, v := range x {
			r[i] = applyS2(s, v, y)
			if err, ok := r[i].(errV); ok {
				return err
			}
		}
		return canonical(r)
	default:
		return errf("s[x;y] : x non-integer (%s)", x.Type())
	}
}

func bytes(x V) V {
	switch x := x.(type) {
	case S:
		return I(len(x))
	case AS:
		r := make(AI, x.Len())
		for i, s := range x {
			r[i] = len(s)
		}
		return r
	case AV:
		r := make(AV, x.Len())
		for i, z := range x {
			r[i] = bytes(z)
			if err, ok := r[i].(errV); ok {
				return err
			}
		}
		return canonical(r)
	default:
		return errType("bytes x", "x", x)
	}
}
