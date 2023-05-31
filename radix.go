// Radix sort adapted from https://github.com/shawnsmithdev/zermelo which comes
// with the following license:
//
// The MIT License (MIT)
//
// Copyright (c) 2014 Shawn Smith
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package goal

import (
	"math"
	"sort"
)

const radix uint = 8
const cachedLen = 256

func radixSortAI(ctx *Context, x *AI, min, max int64) *AI {
	xlen := x.Len()
	if min >= math.MinInt16 && max <= math.MaxInt16 {
		var buf []int16
		if xlen < cachedLen {
			if ctx.sortBuf16 == nil {
				ctx.sortBuf16 = make([]int16, cachedLen*2)
			}
			buf = ctx.sortBuf16
		} else {
			buf = make([]int16, xlen*2)
		}
		r := radixSortAIWithSize(x, buf, 16, math.MinInt16)
		return r
	}
	if min >= math.MinInt32 && max <= math.MaxInt32 {
		var buf []int32
		if xlen < cachedLen {
			if ctx.sortBuf32 == nil {
				ctx.sortBuf32 = make([]int32, cachedLen*2)
			}
			buf = ctx.sortBuf32
		} else {
			buf = make([]int32, xlen*2)
		}
		r := radixSortAIWithSize(x, buf, 32, math.MinInt32)
		return r
	}
	// NOTE: given that Go's stdlib interface-based sort isn't the fastest
	// on integers, it would sometimes be better to use radix sort for
	// 64bits too, but not always.
	x = scloneAI(x)
	sort.Sort(x)
	return x
}

func radixSortAIWithSize[I signed](x *AI, buf []I, size uint, min I) *AI {
	from := radixSortInt64sWithSize[I](x.elts, buf, size, min)
	var dst []int64
	reuse := x.reusable()
	if reuse {
		dst = x.elts
	} else {
		dst = make([]int64, x.Len())
	}
	for i, n := range from {
		dst[i] = int64(n)
	}
	if reuse {
		return x
	}
	return &AI{elts: dst}
}

func radixSortInt64sWithSize[I signed](x []int64, buf []I, size uint, min I) []I {
	xlen := len(x)
	from := buf[:xlen]
	to := buf[xlen : xlen*2]
	for i, xi := range x {
		from[i] = I(xi)
	}
	radixSortWithBuffer[I](from, to, size, min)
	return from
}

type signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// radixSortWithBuffer sorts from using a radix sort. The to buffer slice
// should have same length as from, size should be the bitsize of T, and min
// should be the minimum possible value of type T.
func radixSortWithBuffer[I signed](from, to []I, size uint, min I) {
	var keyOffset uint
	for keyOffset = 0; keyOffset < size; keyOffset += radix {
		var (
			offset [256]int // Keep track of where room is made for byte groups in the buffer
			prev   I        = min
			key    uint8
			sorted = true
		)

		// Compute counts by byte type at current radix
		for _, elem := range from {
			key = uint8(elem >> keyOffset)
			offset[key]++
			if sorted {
				sorted = elem >= prev
				prev = elem
			}
		}

		if sorted {
			break
		}

		// Compute target bucket offsets from counts
		var sum int
		if keyOffset == size-radix {
			// Negatives
			for i := 128; i < len(offset); i++ {
				count := offset[i]
				offset[i] = sum
				sum += count
			}
			// Positives
			for i := 0; i < 128; i++ {
				count := offset[i]
				offset[i] = sum
				sum += count
			}
		} else {
			for i, count := range offset {
				offset[i] = sum
				sum += count
			}
		}

		// Swap values between the buffers by radix
		for _, elem := range from {
			key = uint8(elem >> keyOffset)
			to[offset[key]] = elem
			offset[key]++
		}

		from, to = to, from
	}

	// copy from buffer if done during odd turn
	if radix&keyOffset == radix {
		copy(to, from)
	}
}

func radixGradeSmallRange(ctx *Context, x *AI, min, max int64) V {
	xlen := x.Len()
	var buf []int8
	if xlen < cachedLen {
		if ctx.sortBuf8 == nil {
			ctx.sortBuf8 = make([]int8, cachedLen)
		}
		buf = ctx.sortBuf8
	} else {
		buf = make([]int8, xlen)
	}
	from := buf[:xlen]
	if min >= math.MinInt8 && max <= math.MaxInt8 {
		for i, xi := range x.elts {
			from[i] = int8(xi)
		}
	} else {
		for i, xi := range x.elts {
			from[i] = int8(xi - min - math.MinInt8)
		}
	}
	if xlen < 256 {
		p := make([]byte, xlen)
		radixGradeInt8(from, p)
		return NewAB(p)
	}
	var p []int64
	if x.reusable() {
		p = x.elts
	} else {
		p = make([]int64, xlen)
	}
	radixGradeInt8(from, p)
	return NewAI(p)
}

func radixGradeAI(ctx *Context, x *AI, min, max int64) V {
	xlen := x.Len()
	if min >= math.MinInt16 && max <= math.MaxInt16 {
		var buf []int16
		if xlen < cachedLen {
			if ctx.sortBuf16 == nil {
				ctx.sortBuf16 = make([]int16, cachedLen*2)
			}
			buf = ctx.sortBuf16
		} else {
			buf = make([]int16, xlen*2)
		}
		if xlen < 256 {
			r := radixGradeAIBytes[int16](x, buf, 16, math.MinInt16)
			return NewAB(r)
		}
		r := radixGradeAIIs[int16](x, buf, 16, math.MinInt16)
		return NewAI(r)
	}
	if min >= math.MinInt32 && max <= math.MaxInt32 {
		var buf []int32
		if xlen < cachedLen {
			if ctx.sortBuf32 == nil {
				ctx.sortBuf32 = make([]int32, cachedLen*2)
			}
			buf = ctx.sortBuf32
		} else {
			buf = make([]int32, xlen*2)
		}
		if xlen < 256 {
			r := radixGradeAIBytes[int32](x, buf, 32, math.MinInt32)
			return NewAB(r)
		}
		r := radixGradeAIIs[int32](x, buf, 32, math.MinInt32)
		return NewAI(r)
	}
	if xlen < 256 {
		p := &permutation[byte]{Perm: permRange[byte](xlen), X: x}
		sort.Stable(p)
		return NewAB(p.Perm)
	}
	p := &permutation[int64]{Perm: permRange[int64](xlen), X: x}
	sort.Stable(p)
	return NewAI(p.Perm)
}

func radixGradeAIIs[I signed](x *AI, buf []I, size uint, min I) []int64 {
	xlen := x.Len()
	from := buf[:xlen]
	to := buf[xlen : xlen*2]
	for i, xi := range x.elts {
		from[i] = I(xi)
	}
	var fromp, top []int64
	if x.reusable() {
		fromp = x.elts
		top = make([]int64, xlen)
	} else {
		bufp := make([]int64, xlen*2)
		fromp = bufp[:xlen]
		top = bufp[xlen : xlen*2]
	}
	for i := range fromp {
		fromp[i] = int64(i)
	}
	radixGradeWithBuffer[I, int64](from, to, fromp, top, size, min)
	return fromp
}

func radixGradeAIBytes[I signed](x *AI, buf []I, size uint, min I) []byte {
	xlen := x.Len()
	from := buf[:xlen]
	to := buf[xlen : xlen*2]
	for i, xi := range x.elts {
		from[i] = I(xi)
	}
	bufp := make([]byte, xlen*2)
	fromp := bufp[:xlen]
	top := bufp[xlen : xlen*2]
	for i := range fromp {
		fromp[i] = byte(i)
	}
	radixGradeWithBuffer[I, byte](from, to, fromp, top, size, min)
	return fromp
}

// radixGradeWithBuffer sorts from using a radix sort. The to buffer, fromp,
// top slices should have same length as from, size should be the bitsize of T,
// and min should be the minimum possible value of type T.
func radixGradeWithBuffer[J signed, I integer](from, to []J, fromp, top []I, size uint, min J) {
	var keyOffset uint
	for keyOffset = 0; keyOffset < size; keyOffset += radix {
		var (
			offset [256]I // Keep track of where room is made for byte groups in the buffer
			prev   J      = min
			key    uint8
			sorted = true
		)

		// Compute counts by byte type at current radix
		for _, elem := range from {
			key = uint8(elem >> keyOffset)
			offset[key]++
			if sorted {
				sorted = elem >= prev
				prev = elem
			}
		}

		if sorted {
			break
		}

		// Compute target bucket offsets from counts
		var sum I
		if keyOffset == size-radix {
			// Negatives
			for i := 128; i < len(offset); i++ {
				count := offset[i]
				offset[i] = sum
				sum += count
			}
			// Positives
			for i := 0; i < 128; i++ {
				count := offset[i]
				offset[i] = sum
				sum += count
			}
		} else {
			for i, count := range offset {
				offset[i] = sum
				sum += count
			}
		}

		// Swap values between the buffers by radix
		for i, elem := range from {
			key = uint8(elem >> keyOffset)
			j := offset[key]
			offset[key]++
			to[j] = elem
			top[j] = fromp[i]
		}

		from, to = to, from
		fromp, top = top, fromp
	}

	// copy from buffer if done during odd turn
	if radix&keyOffset == radix {
		copy(to, from)
		copy(top, fromp)
	}
}

// radixGradeInt8 sorts p by from, and puts sorted from into to.
func radixGradeInt8[I integer](from []int8, p []I) {
	var (
		offset [256]I // Keep track of where room is made for byte groups in the buffer
		key    uint8
	)

	// Compute counts by byte type at current radix
	for _, elem := range from {
		key = uint8(elem)
		offset[key]++
	}

	// Compute target bucket offsets from counts
	var sum I
	// Negatives
	for i := 128; i < len(offset); i++ {
		count := offset[i]
		offset[i] = sum
		sum += count
	}
	// Positives
	for i := 0; i < 128; i++ {
		count := offset[i]
		offset[i] = sum
		sum += count
	}

	// Swap values between the buffers by radix
	for i, elem := range from {
		key = uint8(elem)
		j := offset[key]
		offset[key]++
		p[j] = I(i)
	}
}

// radixGradeUint8 sorts p by from.
func radixGradeUint8[I integer](from []uint8, p []I) {
	var (
		offset [256]I // Keep track of where room is made for byte groups in the buffer
	)

	// Compute counts by byte type at current radix
	for _, elem := range from {
		offset[elem]++
	}

	// Compute target bucket offsets from counts
	var sum I
	for i, count := range offset {
		offset[i] = sum
		sum += count
	}

	// Swap values between the buffers by radix
	for i, elem := range from {
		j := offset[elem]
		offset[elem]++
		p[j] = I(i)
	}
}
