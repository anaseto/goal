package main

import "fmt"

func badtype(s string) E {
	return E(fmt.Sprintf("type error %s", s))
}

func badlen(s string) E {
	return E(fmt.Sprintf("length error %s", s))
}
