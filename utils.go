package main

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
