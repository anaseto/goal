package main

import (
	"fmt"
	"strings"
)

func main() {
	testPrimitives()
	testScanner()
	testParser()
}

func testPrimitives() {
	fmt.Printf("Add:%#v\n", Add(I(3), I(5)))
	fmt.Printf("Add:%#v\n", Add(F(3), I(5)))
	fmt.Printf("Add:%#v\n", Add(I(3), AI{5, 3, 8}))
	fmt.Printf("Add:%#v\n", Add(AF{1, 2, 3}, AI{5, 3, 8}))
	fmt.Printf("Add:%#v\n", Add(AF{1, 2, 3, 4}, AI{5, 3, 8}))
	fmt.Printf("Add:%#v\n", Add(AV{AF{1, 2}, AF{3, 4}}, AI{3, 8}))
	fmt.Printf("Equal:%#v\n", Equal(AI{1, 3, 8, 2}, AI{5, 3, 8, 1}))
	//fmt.Printf("Add2:%#v\n", float64(1.0)/float64(0.0))
	fmt.Printf("Divide:%#v\n", Divide(F(2), F(0)))
	fmt.Printf("Sort:%#v\n", SortUp(AI{3, 2, 1}))
	fmt.Printf("Sort:%#v\n", SortUp(AV{3, 2, 1}))
	fmt.Printf("Sort:%#v\n", SortUp(AV{3, 2, AI{}, 1, AI{2, 2}}))
	fmt.Printf("Take:%#v\n", Take(5, AI{2, 3, 4}))
	fmt.Printf("Take:%#v\n", Take(2, AI{2, 3, 4}))
	fmt.Printf("Take:%#v\n", Take(-5, AI{2, 3, 4}))
	fmt.Printf("Take:%#v\n", Take(-2, AI{2, 3, 4}))
	fmt.Printf("ShiftBefore:%#v\n", ShiftBefore(AI{2, 3}, AI{1, 4, 5}))
	fmt.Printf("ShiftBefore:%#v\n", ShiftBefore(7, AF{1, 4, 5}))
	fmt.Printf("ShiftAfter:%#v\n", ShiftAfter(7, AF{1, 4, 5}))
	fmt.Printf("Flip:%#v\n", Flip(AI{1, 2, 3}))
	fmt.Printf("Flip:%#v\n", Flip(AV{AF{1}, AF{4}, AF{5}}))
	fmt.Printf("Flip:%#v\n", Flip(AV{AF{1, 2}, I(4), I(5)}))
	fmt.Printf("Flip:%#v\n", Flip(AV{AF{1, 2}, I(4), I(5), "patata"}))
	fmt.Printf("Classify:%#v\n", Classify(AV{AF{1, 2}, I(4), AF{1, 2}, "patata"}))
	fmt.Printf("Classify:%#v\n", Classify(AI{1, 2, 3, 2, 2, 4, 5, 3}))
	fmt.Printf("Range:%#v\n", Range(10))
	fmt.Printf("Range:%#v\n", Range(AI{4, 2, 3}))
	fmt.Printf("Indices:%#v\n", Indices(AI{0, 1, 0, 0, 1}))
	fmt.Printf("Indices:%#v\n", Indices(AI{3, 0, 1}))
	fmt.Printf("MarkFirts:%#v\n", MarkFirts(AI{3, 3, 1, 2, 4, 2, 4}))
	fmt.Printf("OccurrenceCount:%#v\n", OccurrenceCount(AB{false, false, true, false, true, true}))
	fmt.Printf("Windows:%#v\n", Windows(3, Range(7)))
	fmt.Printf("MemberOf:%#v\n", MemberOf(AS{"two", "twelve", "five", "one", "one", "nine"}, AS{"one", "two", "four"}))
	fmt.Printf("MemberOf:%#v\n", MemberOf(5, AI{2, 3, 6}))
	fmt.Printf("MemberOf:%#v\n", MemberOf(3, AV{2, 3, 6}))
	fmt.Printf("MemberOf:%#v\n", MemberOf(AI{4, 3, 3, 3, 5, 2, 6}, AI{2, 3, 6}))
	fmt.Printf("Group:%#v\n", Group(AI{0, 3, 2, 2, 0, 3}))
	fmt.Printf("Group:%#v\n", Group(AB{false, true, false}))
}

func testScanner() {
	sr := strings.NewReader("%!:+/\n&/: /comment\nident 23 \"string\"")
	sc := &Scanner{reader: sr}
	sc.Init()
	for tk := sc.Next(); tk.Type != EOF; tk = sc.Next() {
		fmt.Printf("%v", tk)
		if tk.Type == NEWLINE {
			fmt.Print("\n")
		}
	}
	fmt.Print("\n")
}

func testParser() {
	s := "23 45 + {(x-23)+|x} 23 43 + fun[2;3];"
	sr := strings.NewReader(s)
	p := &Parser{}
	fmt.Println(s)
	p.ParseWithReader(sr)
}
