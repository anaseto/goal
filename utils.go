package main

import "math"

func B2I(b B) (i I) {
	if b {
		i = 1
	}
	return
}

func B2F(b B) (f F) {
	if b {
		f = 1
	}
	return
}

func isNum(x Object) bool {
	switch x.(type) {
	case I, F:
		return true
	default:
		return false
	}
}

func isArray(x Object) bool {
	switch x.(type) {
	case AO, AI, AF, AS:
		return true
	default:
		return false
	}
}

func sign(x F) (sign int) {
	if x > 0 {
		sign = 1
	} else if x < 0 {
		sign = -1
	}
	return sign
}

func divide(w, x F) F {
	if x == 0 {
		return F(math.Inf(sign(w)))
	}
	return w / x
}
