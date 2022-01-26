package main

// Match returns w≡x.
func Match(w, x O) bool {
	switch w := w.(type) {
	case B:
		switch x := x.(type) {
		case B:
			return w == x
		case I:
			return B2I(w) == x
		case F:
			return B2F(w) == x
		default:
			return false
		}
	case F:
		switch x := x.(type) {
		case B:
			return w == B2F(x)
		case I:
			return w == F(x)
		case F:
			return w == x
		default:
			return false
		}
	case I:
		switch x := x.(type) {
		case B:
			return w == B2I(x)
		case I:
			return w == x
		case F:
			return F(w) == x
		default:
			return false
		}
	case S:
		switch x := x.(type) {
		case S:
			return w == x
		default:
			return false
		}
	case Array:
		// TODO: optimize common cases
		switch x := x.(type) {
		case Array:
			l := Length(w)
			if l != Length(x) {
				return false
			}
			for i := 0; i < l; i++ {
				if !Match(w.At(i), x.At(i)) {
					return false
				}
			}
			return true
		default:
			return false
		}
	default:
		return false
	}
}
