// structural functions (Length, Reverse, Take, ...)

package main

import "sort"

// Length returns ≠x.
func Length(x O) I {
	switch x := x.(type) {
	case nil:
		return 0
	case Array:
		return x.Len()
	default:
		return 1
	}
}

func reverse(x O) {
	switch x := x.(type) {
	case AB:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case AF:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case AI:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case AS:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case AO:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case sort.Interface:
		sort.Reverse(x)
	}
}

// Reverse returns ⌽x.
func Reverse(x O) O {
	switch x := x.(type) {
	case Array:
		r := cloneShallow(x)
		reverse(r)
		return r
	default:
		return badtype("⌽")
	}
}

// First returns ↑x.
func First(x O) O {
	switch x := x.(type) {
	case Array:
		if x.Len() == 0 {
			return badlen("↑")
		}
		return x.At(0)
	default:
		return x
	}
}

// Tail returns ↓x.
func Tail(x O) O {
	switch x := x.(type) {
	case Array:
		if x.Len() == 0 {
			return badlen("↓")
		}
		return x.Slice(1, x.Len())
	default:
		return badtype("↓")
	}
}

// Drop returns w↓x.
func Drop(w, x O) O {
	i := 0
	switch w := w.(type) {
	case B:
		i = B2I(w)
	case I:
		i = w
	case F:
		i = I(w)
	default:
		// TODO: improve error messages
		return badtype("↓")
	}
	switch x := x.(type) {
	case Array:
		if x.Len() == 0 {
			return badlen("↓")
		}
		switch {
		case i >= 0:
			if i > x.Len() {
				i = x.Len()
			}
			return x.Slice(i, x.Len())
		default:
			i = x.Len() + i
			if i < 0 {
				i = 0
			}
			return x.Slice(0, i)
		}
	default:
		return x
	}
}
