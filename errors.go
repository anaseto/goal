package main

import "fmt"

func errs(s string) E {
	return E(s)
}

func errsw(s string) E {
	return E("left argument:" + s)
}

func errf(format string, a ...interface{}) E {
	return E(fmt.Sprintf(format, a...))
}
