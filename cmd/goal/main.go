package main

import (
	"fmt"
	"goal"
	"log"
	"os"
)

func main() {
	ctx := goal.NewContext()
	ctx.SetSource("-", os.Stdin)
	for {
		fmt.Print("  ")
		v, err := ctx.RunExpr()
		if err != nil {
			_, eof := err.(goal.ErrEOF)
			if eof {
				echo(v)
				return
			}
			log.Fatal(err)
		}
		echo(v)
	}
}

func echo(v goal.V) {
	if v != nil {
		fmt.Printf("%v\n", v)
	}
}
