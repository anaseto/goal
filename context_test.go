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

func getScriptMatchTests(s string) ([]matchTest, error) {
	d := os.DirFS("testdata/scripts")
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
		body := strings.SplitN(text, "\n/RESULT:\n", 2)
		if len(body) != 2 {
			log.Printf("%s: bad script", fname)
			continue
		}
		left := body[0]
		right := body[1]
		mts = append(mts, matchTest{
			Fname: fname,
			Left:  strings.TrimSpace(left),
			Right: strings.TrimSpace(right),
		})
	}
	return mts, nil
}

func TestEval(t *testing.T) {
	mts, err := getMatchTests("*.goal")
	if err != nil {
		t.Errorf("getMatchTests: %v", err)
	}
	smts, err := getScriptMatchTests("*.goal")
	if err != nil {
		t.Errorf("getScriptMatchTests: %v", err)
	}
	for _, mt := range append(mts, smts...) {
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
				t.Fatalf("compile error: %v", err)
			}
			vLeft, errLeft := ctxLeft.Run()
			ctxRight := NewContext()
			vRight, errRight := ctxRight.Eval(mt.Right)
			if errLeft != nil || errRight != nil {
				t.Log(ps)
				t.Log(matchString)
				t.Fatalf("return error: `%v` vs `%v`", errLeft, errRight)
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
	smts, err := getScriptMatchTests("errors.goal")
	if err != nil {
		t.Errorf("getScriptMatchTests: %v", err)
	}
	for _, mt := range append(mts, smts...) {
		mt := mt
		name := fmt.Sprintf("%s:%d", mt.Fname, mt.Line)
		matchString := mt.Left
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
					t.Fatalf("no error left: result: %v\nexpected: %v", v, mt.Right)
				}
			}
			e, ok := err.(*PanicError)
			if !ok {
				// should never happen
				t.Log(ps)
				t.Log(matchString)
				t.Fatalf("bad error: `%v`\nexpected:`%v`", err, mt.Right)
			}
			msg := e.Msg
			if strings.Contains(mt.Left, "\n") {
				msg = e.Error()
			}
			if !strings.Contains(e.Msg, mt.Right) {
				t.Log(ps)
				t.Log(matchString)
				t.Logf("\n   error: %s\nexpected: %v", msg, mt.Right)
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
	ctx := NewContext()
	for n := 0; n < b.N; n++ {
		ctx.Eval("{x<1000}{x+1}/1")
	}
}

func BenchmarkFoldDo(b *testing.B) {
	ctx := NewContext()
	for n := 0; n < b.N; n++ {
		ctx.Eval("1000{x+1}/1")
	}
}

func BenchmarkFib(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("fib:{?[x~0;0;x~1;1;(fib x-1)+(fib x-2)]}")
	for n := 0; n < b.N; n++ {
		ctx.Eval("fib 35")
	}
}

func BenchmarkFibTailRec(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("fibrec:{?[x~0;y;x~1;z;fibrec[x-1;z;y+z]]}")
	for n := 0; n < b.N; n++ {
		ctx.Eval("fibrec[35;0;1]")
	}
}

func BenchmarkFibDoWhile(b *testing.B) {
	ctx := NewContext()
	for n := 0; n < b.N; n++ {
		ctx.Eval("*35{x[1],+/x}/0 1")
	}
}

func BenchmarkNewContext(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewContext()
	}
}

func BenchmarkWhileN(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:!10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("100 {x+1}/a")
	}
}

func BenchmarkDosAtDef(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:!10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("100 {x[2]+:1;x}/a")
	}
}

func BenchmarkDosAtDefAV(b *testing.B) {
	ctx := NewContext()
	ctx.Eval(`a:(!10000),"a"`)
	for n := 0; n < b.N; n++ {
		ctx.Eval("100 {x[2]+:1;x}/a")
	}
}

func BenchmarkDosAt(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:!10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("100 {@[x;2;+;1]}/a")
	}
}

func BenchmarkDosAtAV(b *testing.B) {
	ctx := NewContext()
	ctx.Eval(`a:(!10000),"a"`)
	for n := 0; n < b.N; n++ {
		ctx.Eval("100 {@[x;2;+;1]}/a")
	}
}

func BenchmarkReverse(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:!10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("100 {|x}/a")
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

func BenchmarkIndex(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:!10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("1000 {a[42]}/1")
	}
}

func BenchmarkIndexAV(b *testing.B) {
	ctx := NewContext()
	ctx.Eval(`a:(!10000),"a"`)
	for n := 0; n < b.N; n++ {
		ctx.Eval("1000 {a[42]}/1")
	}
}

func BenchmarkCountAV(b *testing.B) {
	ctx := NewContext()
	ctx.Eval(`a:(!10000),"a"`)
	for n := 0; n < b.N; n++ {
		ctx.Eval("1000 {#a}/1")
	}
}

func BenchmarkWhere(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:0=30!!10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("&a")
	}
}

func BenchmarkUniq10(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:10?10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("100 {?a}/0")
	}
}

func BenchmarkUniq50(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:50?10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("20 {?a}/0")
	}
}

func BenchmarkUniq100(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:100?10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("10 {?a}/0")
	}
}

func BenchmarkUniq256(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:256?10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("4 {?a}/0")
	}
}

func BenchmarkUniq1000(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:1000?10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("?a")
	}
}

func BenchmarkSearch100In1000(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:1000?10000;b:100?10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("a?b")
	}
}

func BenchmarkSearch200In500(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:500?10000;b:200?10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("a?b")
	}
}

func BenchmarkSearch10In10000(b *testing.B) {
	ctx := NewContext()
	ctx.Eval("a:10000?10000;b:10?10000")
	for n := 0; n < b.N; n++ {
		ctx.Eval("a?b")
	}
}
