package goal

import "strings"

func applyS(s S, x V) V {
	if x.IsI() {
		xv := x.I()
		if xv < 0 {
			xv += int64(len(s))
		}
		if xv < 0 || xv > int64(len(s)) {
			return Panicf("s[i] : i out of bounds index (%d)", xv)
		}
		return NewV(s[xv:])
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("s[x] : x non-integer (%g)", x.F())
		}
		return applyS(s, NewI(int64(x.F())))
	}
	switch xv := x.value.(type) {
	case *AB:
		return applyS(s, fromABtoAI(xv))
	case *AI:
		r := make([]string, xv.Len())
		for i, n := range xv.Slice {
			if n < 0 {
				n += int64(len(s))
			}
			if n < 0 || n > int64(len(s)) {
				return Panicf("s[i] : i out of bounds index (%d)", n)
			}
			r[i] = string(s[n:])
		}
		return NewAS(r)
	case *AF:
		z := toAI(xv)
		if z.IsPanic() {
			return z
		}
		return applyS(s, z)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = applyS(s, xi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewAV(r))
	default:
		return Panicf("s[x] : x non-integer (%s)", x.Type())
	}
}

func applyS2(s S, x V, y V) V {
	var l int64
	if y.IsI() {
		if y.I() < 0 {
			return Panicf("s[x;y] : y negative (%d)", y.I())
		}
		l = y.I()
	} else if y.IsF() {
		if !isI(y.F()) {
			return Panicf("s[x;y] : y non-integer (%g)", y.F())
		}
		l = int64(y.F())
	} else {
		switch yv := y.value.(type) {
		case *AI:
		case *AB:
			return applyS2(s, x, fromABtoAI(yv))
		case *AF:
			z := toAI(yv)
			if z.IsPanic() {
				return z
			}
			return applyS2(s, x, z)
		default:
			return panicType("s[x;y]", "y", y)
		}
	}
	if x.IsI() {
		xv := x.I()
		if xv < 0 {
			xv += int64(len(s))
		}
		if xv < 0 || xv > int64(len(s)) {
			return Panicf("s[i;y] : i out of bounds index (%d)", xv)
		}
		if _, ok := y.value.(*AI); ok {
			return Panicf("s[x;y] : x is an atom but y is an array")
		}
		if xv+l > int64(len(s)) {
			l = int64(len(s)) - xv
		}
		return NewV(s[xv : xv+l])

	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("s[x;y] : x non-integer (%g)", x.F())
		}
		return applyS2(s, NewI(int64(x.F())), y)
	}
	switch xv := x.value.(type) {
	case *AB:
		return applyS2(s, fromABtoAI(xv), y)
	case *AI:
		r := make([]string, xv.Len())
		if z, ok := y.value.(*AI); ok {
			if z.Len() != xv.Len() {
				return Panicf("s[x;y] : length mismatch: %d (#x) %d (#y)",
					xv.Len(), z.Len())

			}
			for i, n := range xv.Slice {
				if n < 0 {
					n += int64(len(s))
				}
				if n < 0 || n > int64(len(s)) {
					return Panicf("s[i;y] : i out of bounds index (%d)", n)
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
				return Panicf("s[i;y] : i out of bounds index (%d)", n)
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
		if z.IsPanic() {
			return z
		}
		return applyS2(s, z, y)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = applyS2(s, xi, y)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewAV(r))
	default:
		return Panicf("s[x;y] : x non-integer (%s)", x.Type())
	}
}

func bytes(x V) V {
	switch xv := x.value.(type) {
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
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewAV(r))
	default:
		return panicType("bytes x", "x", x)
	}
}

// cast implements s$y.
func cast(s S, y V) V {
	switch s {
	case "i":
		return casti(y)
	case "n":
		return castn(y)
	case "s":
		return casts(y)
	default:
		return Panicf("s$y : unsupported \"%s\" value for s", s)
	}
}

func casti(y V) V {
	if y.IsI() {
		return y
	}
	if y.IsF() {
		return NewI(int64(y.F()))
	}
	switch yv := y.value.(type) {
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
		return toAI(floor(y).value.(*AF))
	case *AV:
		r := make([]V, yv.Len())
		for i := range r {
			r[i] = casti(yv.At(i))
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return NewAV(r)
	default:
		return panics("\"i\"$y : non-numeric y")
	}
}

func castn(y V) V {
	if y.IsI() || y.IsF() {
		return y
	}
	switch yv := y.value.(type) {
	case S:
		xi, err := parseNumber(string(yv))
		if err != nil {
			return Panicf("\"i\"$y : non-numeric y (%s) : %v", yv, err)
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
				return Panicf("\"i\"$y : y contains non-numeric (%s) : %v", s, err)
			}
			r[i] = n
		}
		return Canonical(NewAV(r))
	case *AF:
		return y
	case *AV:
		r := make([]V, yv.Len())
		for i := range r {
			r[i] = castn(yv.At(i))
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewAV(r))
	default:
		return panics("\"i\"$y : non-numeric y")
	}
}

func casts(y V) V {
	if y.IsI() {
		return NewS(string(rune(y.I())))
	}
	if y.IsF() {
		return casts(NewI(int64(y.F())))
	}
	switch yv := y.value.(type) {
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
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewAV(r))
	default:
		return panics("\"i\"$y : non-numeric y")
	}
}

func drops(s S, y V) V {
	switch yv := y.value.(type) {
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
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return NewAV(r)
	default:
		return panicType("s_y", "y", y)
	}
}

// trim returns s^y.
func trim(s S, y V) V {
	switch yv := y.value.(type) {
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
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return NewAV(r)
	default:
		return panicType("s^y", "y", y)
	}
}

func replace(x, y, z V) V {
	switch xv := x.value.(type) {
	case S:
		return replaceS(xv, y, z)
	case *AS:
		return replaceAS(xv, y, z)
	case *AV:
		r := xv.reuse()
		for i, v := range xv.Slice {
			r.Slice[i] = replace(v, y, z)
			if r.Slice[i].IsPanic() {
				return r.Slice[i]
			}
		}
		return NewV(r)
	default:
		return panicType("sub[x;y;z]", "x", x)
	}
}

func replaceS(s S, y, z V) V {
	switch yv := y.value.(type) {
	case S:
		zv, ok := z.value.(S)
		if !ok {
			return Panicf("sub[s;s;z] : non-string z (%s)", z.Type())
		}
		return NewS(strings.ReplaceAll(string(s), string(yv), string(zv)))
	case *AS:
		zv, ok := z.value.(*AS)
		if !ok {
			return Panicf("sub[s;S;z] : z not a string array (%s)", z.Type())
		}
		if yv.Len() != zv.Len() {
			return Panicf("sub[s;y;z] : length mismatch for y and z (%d vs %d)", yv.Len(), zv.Len())
		}
		oldnews := make([]string, 0, 2*yv.Len())
		for i, s := range yv.Slice {
			oldnews = append(oldnews, s, zv.At(i))
		}
		rep := strings.NewReplacer(oldnews...)
		return NewS(rep.Replace(string(s)))
	default:
		return panicType("sub[s;y;z]", "y", y)
	}
}

func replaceAS(xv *AS, y, z V) V {
	switch yv := y.value.(type) {
	case S:
		zv, ok := z.value.(S)
		if !ok {
			return Panicf("sub[s;s;z] : non-string z (%s)", z.Type())
		}
		r := xv.reuse()
		for i, s := range xv.Slice {
			r.Slice[i] = strings.ReplaceAll(string(s), string(yv), string(zv))
		}
		return NewV(r)
	case *AS:
		zv, ok := z.value.(*AS)
		if !ok {
			return Panicf("sub[s;S;z] : z not a string array (%s)", z.Type())
		}
		if yv.Len() != zv.Len() {
			return Panicf("sub[s;y;z] : length mismatch for y and z (%d vs %d)", yv.Len(), zv.Len())
		}
		oldnews := make([]string, 0, 2*yv.Len())
		for i, s := range yv.Slice {
			oldnews = append(oldnews, s, zv.At(i))
		}
		rep := strings.NewReplacer(oldnews...)
		r := xv.reuse()
		for i, s := range xv.Slice {
			r.Slice[i] = rep.Replace(string(s))
		}
		return NewV(r)
	default:
		return panicType("sub[s;y;z]", "y", y)
	}
}
