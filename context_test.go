package goal

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"testing"
)

type matchTest struct {
	Fname string
	Line  int
	Left  string
	Right string
}

func getMatchTests(s string) ([]matchTest, error) {
	d := os.DirFS("testdata/")
	fnames, err := fs.Glob(d, s)
	if err != nil {
		return nil, err
	}
	mts := []matchTest{}
	for _, fname := range fnames {
		bs, err := fs.ReadFile(d, fname)
		if err != nil {
			return nil, err
		}
		text := string(bs)
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) == 0 || line[0] == '/' {
				continue
			}
			left, right, found := strings.Cut(line, " /")
			if !found {
				log.Printf("%s:%d: bad line", fname, i+1)
				continue
			}
			mts = append(mts, matchTest{
				Fname: fname,
				Line:  i + 1,
				Left:  strings.TrimSpace(left),
				Right: strings.TrimSpace(right),
			})
		}
	}
	return mts, nil
}

func TestEval(t *testing.T) {
	mts, err := getMatchTests("*.goal")
	if err != nil {
		t.Errorf("getMatchTests: %v", err)
	}
	for _, mt := range mts {
		if mt.Fname == "errors.goal" {
			continue
		}
		mt := mt
		name := fmt.Sprintf("%s:%d", mt.Fname, mt.Line)
		matchString := fmt.Sprintf("(%s) ~ (%s)", mt.Left, mt.Right)
		t.Run(name, func(t *testing.T) {
			ctxLeft := NewContext()
			err := ctxLeft.Compile("", mt.Left)
			ps := ctxLeft.programString()
			if err != nil {
				t.Log(ps)
				t.Log(matchString)
				t.Logf("compile error: %v", err)
				t.Fail()
				return
			}
			vLeft, errLeft := ctxLeft.Run()
			ctxRight := NewContext()
			vRight, errRight := ctxRight.Eval(mt.Right)
			if errLeft != nil || errRight != nil {
				t.Log(ps)
				t.Log(matchString)
				t.Logf("return error: `%v` vs `%v`", errLeft, errRight)
				t.Fail()
				return
			}
			if !Match(vLeft, vRight) {
				t.Log(ps)
				t.Log(matchString)
				if vLeft != (V{}) {
					t.Logf("results: %s vs %s\n", vLeft.Sprint(ctxLeft), vRight.Sprint(ctxRight))
				} else {
					t.Logf("results: %v vs %s\n", vLeft, vRight.Sprint(ctxRight))
				}
				//t.Logf("results (go): %#v vs %#v", vLeft, vRight)
				t.Fail()
			}
		})
	}
}

func TestErrors(t *testing.T) {
	mts, err := getMatchTests("errors.goal")
	if err != nil {
		t.Errorf("getMatchTests: %v", err)
	}
	for _, mt := range mts {
		mt := mt
		name := fmt.Sprintf("%s:%d", mt.Fname, mt.Line)
		matchString := fmt.Sprintf("%s", mt.Left)
		t.Run(name, func(t *testing.T) {
			ctx := NewContext()
			err := ctx.Compile("", mt.Left)
			ps := ctx.programString()
			if err == nil {
				var v V
				v, err = ctx.Run()
				if err == nil {
					t.Log(ps)
					t.Log(matchString)
					t.Errorf("no error left: result: %v\nexpected: %v", v, mt.Right)
				}
			}
			e, ok := err.(*PanicError)
			if !ok {
				// should never happen
				t.Log(ps)
				t.Log(matchString)
				t.Errorf("bad error: `%v`\nexpected:`%v`", err, mt.Right)
			}
			if !strings.Contains(e.Msg, mt.Right) {
				t.Log(ps)
				t.Log(matchString)
				t.Logf("\n   error: %v\nexpected: %v", e.Msg, mt.Right)
				t.Fail()
				return
			}
		})
	}
}

func BenchmarkFoldMinus(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:!1000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("-/a")
	}
}

func BenchmarkFoldPlus(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:!1000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("+/a")
	}
}

func BenchmarkFoldPlusFloat(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:0.1+!1000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("+/a")
	}
}

func BenchmarkFoldLambdaPlus(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:!1000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("{x+y}/a")
	}
}

func BenchmarkFoldLambdaPlusFloat(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:0.1+!1000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("{x+y}/a")
	}
}

func BenchmarkFoldWhile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("{x<1000}{x+1}/1")
	}
}

func BenchmarkFoldDo(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("1000{x+1}/1")
	}
}

func BenchmarkFib(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("fib:{?[x~0;0;x~1;1;(fib x-1)+(fib x-2)]}; fib 35")
	}
}

func BenchmarkFibTailRec(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("fibrec:{?[x~0;y;x~1;z;fibrec[x-1;z;y+z]]};fibrec[35;0;1]")
	}
}

func BenchmarkFibDoWhile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("*35{x[1],+/x}/0 1")
	}
}

func BenchmarkNewContext(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewContext()
	}
}

func BenchmarkWhileN(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("100 {x+1}/!10000")
	}
}

func BenchmarkReverse(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("100 {|x}/!10000")
	}
}

func BenchmarkAppend(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:!10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("500 {x,1}/a")
	}
}

func BenchmarkAppendGlobal(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:!10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("b:a;500 {b,::1}/a")
	}
}

func BenchmarkAppend2(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:!10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("500 {x:x,1;x,1}/a")
	}
}

func BenchmarkDrop2(b *testing.B) {
	ctx := NewContext()
	for n := 0; n < b.N; n++ {
		ctx.push(V{})
		ctx.push(V{})
		ctx.drop2()
	}
}

func BenchmarkDropN2(b *testing.B) {
	ctx := NewContext()
	for n := 0; n < b.N; n++ {
		ctx.push(V{})
		ctx.push(V{})
		ctx.dropN(2)
	}
}
