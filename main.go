package main

import (
	"fmt"
)

func main() {
	fmt.Printf("Hello, world!\n")
	fmt.Printf("Add2:%#v\n", Add(I(3), I(5)))
	fmt.Printf("Add2:%#v\n", Add(F(3), I(5)))
	fmt.Printf("Add1:%#v\n", Conjugate(I(3)))
	fmt.Printf("Add2:%#v\n", Add(I(3), AI{5, 3, 8}))
	fmt.Printf("Add2:%#v\n", Add(AF{1, 2, 3}, AI{5, 3, 8}))
	fmt.Printf("Add2:%#v\n", Add(AF{1, 2, 3, 4}, AI{5, 3, 8}))
	fmt.Printf("Add2:%#v\n", Add(AO{AF{1, 2}, AF{3, 4}}, AI{3, 8}))
}
