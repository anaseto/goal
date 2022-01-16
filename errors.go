package main

import "fmt"

func badtype(s string) E {
	return fmt.Errorf("type error %s", s)
}

func badlen(s string) E {
	return fmt.Errorf("length error %s", s)
}
