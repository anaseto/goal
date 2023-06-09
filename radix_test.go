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
	t.Run("Int8", func(t *testing.T) { testRadixSortSize(t, ctx, math.MinInt8/2, math.MaxInt8) })
	t.Run("Int16", func(t *testing.T) { testRadixSortSize(t, ctx, math.MinInt16/2, math.MaxInt16) })
	t.Run("Int32", func(t *testing.T) { testRadixSortSize(t, ctx, math.MinInt32/2, math.MaxInt32) })
	t.Run("Int8", func(t *testing.T) { testRadixGradeSize(t, ctx, 1000, 256, radixGradeSmallRange) })
	t.Run("Int8", func(t *testing.T) { testRadixGradeSize(t, ctx, math.MinInt8/2, math.MaxInt8, radixGradeSmallRange) })
	t.Run("Int16", func(t *testing.T) { testRadixGradeSize(t, ctx, math.MinInt16/2, math.MaxInt16, radixGradeAI) })
	t.Run("Int32", func(t *testing.T) { testRadixGradeSize(t, ctx, math.MinInt32/2, math.MaxInt32, radixGradeAI) })
}

func testRadixSortSize(t *testing.T, ctx *Context, min, span int64) {
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

func testRadixGradeSize(t *testing.T, ctx *Context, min, span int64, f func(*Context, *AI, int64, int64) V) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 300; i++ {
		x := make([]int64, i)
		for i := range x {
			x[i] = min + rnd.Int63n(span)
		}
		y := make([]int64, i)
		copy(y, x)
		xv := &AI{elts: x}
		xg := f(ctx, xv, min, min+span-1)
		yv := &AI{elts: y}
		p := &permutation[int64]{Perm: permRange[int64](len(y)), X: yv}
		sort.Stable(p)
		yg := NewV(&AI{elts: p.Perm})
		if !yg.Matches(xg) {
			t.Fatalf("mismatch:\n%s\nvs\n%s", xg.Sprint(ctx), yg.Sprint(ctx))
		}
	}
}
