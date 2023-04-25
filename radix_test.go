package goal

import (
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestRadix(t *testing.T) {
	ctx := NewContext()
	t.Run("Int8", func(t *testing.T) { testRadixSize(t, ctx, math.MinInt8/2, math.MaxInt8) })
	t.Run("Int16", func(t *testing.T) { testRadixSize(t, ctx, math.MinInt16/2, math.MaxInt16) })
	t.Run("Int32", func(t *testing.T) { testRadixSize(t, ctx, math.MinInt32/2, math.MaxInt32) })
}

func testRadixSize(t *testing.T, ctx *Context, min, span int64) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 300; i++ {
		x := make([]int64, i)
		for i := range x {
			x[i] = min + rnd.Int63n(span)
		}
		y := make([]int64, i)
		copy(y, x)
		xv := &AI{elts: x}
		xv = radixSortAI(ctx, xv, min, min+span-1)
		yv := &AI{elts: y}
		sort.Sort(yv)
		for i := range xv.elts {
			if xv.elts[i] != yv.elts[i] {
				t.Fatalf("mismatch at index %d (%d vs %d)", i, xv.elts[i], yv.elts[i])
			}
		}
	}
}
