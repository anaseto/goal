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

const (
	radix     uint = 8
	cachedLen      = 256
)

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
	// on integers, it would often be better to use it for 64bits too, but
	// the advantage isn't as clear in all cases.
	x = shallowCloneAI(x)
	sort.Sort(x)
	return x
}

func radixSortAIWithSize[T signed](x *AI, buf []T, size uint, min T) *AI {
	xlen := x.Len()
	from := buf[:xlen]
	to := buf[xlen : xlen*2]
	for i, xi := range x.elts {
		from[i] = T(xi)
	}
	radixSortWithBuffer[T](from, to, size, min)
	var dst []int64
	reuse := reusableRCp(x.rc)
	if reuse {
		dst = x.elts
	} else {
		dst = make([]int64, xlen)
	}
	for i, n := range from {
		dst[i] = int64(n)
	}
	if reuse {
		return x
	}
	return &AI{elts: dst}
}

type signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// radixSortWithBuffer sorts from using a radix sort. The to buffer slice
// should have same length as from, size should be the bitsize of T, and min
// should be the minimum possible value of type T.
func radixSortWithBuffer[T signed](from, to []T, size uint, min T) {
	var keyOffset uint
	for keyOffset = 0; keyOffset < size; keyOffset += radix {
		var (
			offset [256]int // Keep track of where room is made for byte groups in the buffer
			prev   T        = min
			key    uint8
			sorted = true
		)

		for _, elem := range from {
			// For each elem to sort, fetch the byte at current radix
			key = uint8(elem >> keyOffset)
			// inc count of bytes of this type
			offset[key]++
			if sorted { // Detect sorted
				sorted = elem >= prev
				prev = elem
			}
		}

		if sorted { // Short-circuit sorted
			break
		}

		// Find target bucket offsets
		var watermark int
		if keyOffset == size-radix {
			// Handle signed values
			// Negatives
			for i := 128; i < len(offset); i++ {
				count := offset[i]
				offset[i] = watermark
				watermark += count
			}
			// Positives
			for i := 0; i < 128; i++ {
				count := offset[i]
				offset[i] = watermark
				watermark += count
			}
		} else {
			for i, count := range offset {
				offset[i] = watermark
				watermark += count
			}
		}

		// Swap values between the buffers by radix
		for _, elem := range from {
			key = uint8(elem >> keyOffset) // Get the byte of each element at the radix
			to[offset[key]] = elem         // Copy the element depending on byte offsets
			offset[key]++
		}

		// Reverse buffers on each pass
		from, to = to, from
	}

	// copy from buffer if done during odd turn
	if radix&keyOffset == radix {
		copy(to, from)
	}
}
