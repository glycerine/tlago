/*
 * Copyright (c) 2009, 2013, Oracle and/or its affiliates. All rights reserved.
 * Copyright 2009 Google Inc.  All Rights Reserved.
 * DO NOT ALTER OR REMOVE COPYRIGHT NOTICES OR THIS FILE HEADER.
 *
 * This code is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License version 2 only, as
 * published by the Free Software Foundation.  Oracle designates this
 * particular file as subject to the "Classpath" exception as provided
 * by Oracle in the LICENSE file that accompanied this code.
 *
 * This code is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License
 * version 2 for more details (a copy is included in the LICENSE file that
 * accompanied this code).
 *
 * You should have received a copy of the GNU General Public License version
 * 2 along with this work; if not, write to the Free Software Foundation,
 * Inc., 51 Franklin St, Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * Please contact Oracle, 500 Oracle Parkway, Redwood Shores, CA 94065 USA
 * or visit www.oracle.com if you need additional information or have any
 * questions.
 */

// Derived from OpenJDK 21 ComparableTimSort for TLC Value arrays.
package tlc

import "math/bits"

// Compare errors unwind the sort immediately, before any subsequent comparison
// or array write. Other panics keep their original exception identity.
type javaValueSortFailure struct{ err error }

func javaSortValues(a []Value) (err error) {
	defer func() {
		if p := recover(); p != nil {
			if f, ok := p.(javaValueSortFailure); ok {
				err = f.err
			} else {
				panic(p)
			}
		}
	}()
	if a == nil {
		panic(NewNullPointerException())
	}
	n := len(a)
	if n < 2 {
		return nil
	}
	if n < 32 {
		run := javaValueSortRun(a, 0, n)
		javaValueBinarySort(a, 0, n, run)
		return nil
	}
	tlen := 256
	if n < 512 {
		tlen = n >> 1
	}
	stackLen := 49
	if n < 120 {
		stackLen = 5
	} else if n < 1542 {
		stackLen = 10
	} else if n < 119151 {
		stackLen = 24
	}
	s := javaValueTimSort{a: a, tmp: make([]Value, tlen), minGallop: 7, runBase: make([]int, stackLen), runLen: make([]int, stackLen)}
	minRun, r := n, 0
	for minRun >= 32 {
		r |= minRun & 1
		minRun >>= 1
	}
	minRun += r
	lo, remaining := 0, n
	for remaining != 0 {
		run := javaValueSortRun(a, lo, n)
		if run < minRun {
			force := min(minRun, remaining)
			javaValueBinarySort(a, lo, lo+force, lo+run)
			run = force
		}
		s.runBase[s.stackSize] = lo
		s.runLen[s.stackSize] = run
		s.stackSize++
		s.mergeCollapse()
		lo += run
		remaining -= run
	}
	for s.stackSize > 1 {
		i := s.stackSize - 2
		if i > 0 && s.runLen[i-1] < s.runLen[i+1] {
			i--
		}
		s.mergeAt(i)
	}
	return nil
}

func javaValueSortCompare(a, b Value) int {
	if a == nil {
		panic(NewNullPointerException())
	}
	c, err := a.Compare(b)
	if err != nil {
		panic(javaValueSortFailure{err})
	}
	return c
}

func javaValueBinarySort(a []Value, lo, hi, start int) {
	if start == lo {
		start++
	}
	for ; start < hi; start++ {
		pivot := a[start]
		left, right := lo, start
		for left < right {
			mid := int(uint32(left+right) >> 1)
			if javaValueSortCompare(pivot, a[mid]) < 0 {
				right = mid
			} else {
				left = mid + 1
			}
		}
		// copy has System.arraycopy's overlapping-array behavior, including n == 0.
		copy(a[left+1:start+1], a[left:start])
		a[left] = pivot
	}
}

func javaValueSortRun(a []Value, lo, hi int) int {
	runHi := lo + 1
	if runHi == hi {
		return 1
	}
	descending := javaValueSortCompare(a[runHi], a[lo]) < 0
	runHi++
	if descending {
		for runHi < hi && javaValueSortCompare(a[runHi], a[runHi-1]) < 0 {
			runHi++
		}
		for l, h := lo, runHi-1; l < h; l, h = l+1, h-1 {
			a[l], a[h] = a[h], a[l]
		}
	} else {
		for runHi < hi && javaValueSortCompare(a[runHi], a[runHi-1]) >= 0 {
			runHi++
		}
	}
	return runHi - lo
}

type javaValueTimSort struct {
	a, tmp               []Value
	minGallop, stackSize int
	runBase, runLen      []int
}

func (s *javaValueTimSort) mergeCollapse() {
	for s.stackSize > 1 {
		n := s.stackSize - 2
		if (n > 0 && s.runLen[n-1] <= s.runLen[n]+s.runLen[n+1]) || (n > 1 && s.runLen[n-2] <= s.runLen[n]+s.runLen[n-1]) {
			if s.runLen[n-1] < s.runLen[n+1] {
				n--
			}
		} else if n < 0 || s.runLen[n] > s.runLen[n+1] {
			break
		}
		s.mergeAt(n)
	}
}

func (s *javaValueTimSort) mergeAt(i int) {
	base1, len1, base2, len2 := s.runBase[i], s.runLen[i], s.runBase[i+1], s.runLen[i+1]
	s.runLen[i] = len1 + len2
	if i == s.stackSize-3 {
		s.runBase[i+1] = s.runBase[i+2]
		s.runLen[i+1] = s.runLen[i+2]
	}
	s.stackSize--
	k := javaValueGallopRight(s.a[base2], s.a, base1, len1, 0)
	base1 += k
	len1 -= k
	if len1 == 0 {
		return
	}
	len2 = javaValueGallopLeft(s.a[base1+len1-1], s.a, base2, len2, len2-1)
	if len2 == 0 {
		return
	}
	if len1 <= len2 {
		s.mergeLo(base1, len1, base2, len2)
	} else {
		s.mergeHi(base1, len1, base2, len2)
	}
}

func javaValueGallopLeft(key Value, a []Value, base, length, hint int) int {
	lastOfs, ofs := 0, 1
	if javaValueSortCompare(key, a[base+hint]) > 0 {
		maxOfs := length - hint
		for ofs < maxOfs && javaValueSortCompare(key, a[base+hint+ofs]) > 0 {
			lastOfs = ofs
			ofs = int(int32(ofs)<<1) + 1
			if ofs <= 0 {
				ofs = maxOfs
			}
		}
		if ofs > maxOfs {
			ofs = maxOfs
		}
		lastOfs += hint
		ofs += hint
	} else {
		maxOfs := hint + 1
		for ofs < maxOfs && javaValueSortCompare(key, a[base+hint-ofs]) <= 0 {
			lastOfs = ofs
			ofs = int(int32(ofs)<<1) + 1
			if ofs <= 0 {
				ofs = maxOfs
			}
		}
		if ofs > maxOfs {
			ofs = maxOfs
		}
		lastOfs, ofs = hint-ofs, hint-lastOfs
	}
	lastOfs++
	for lastOfs < ofs {
		m := lastOfs + (ofs-lastOfs)/2
		if javaValueSortCompare(key, a[base+m]) > 0 {
			lastOfs = m + 1
		} else {
			ofs = m
		}
	}
	return ofs
}

func javaValueGallopRight(key Value, a []Value, base, length, hint int) int {
	lastOfs, ofs := 0, 1
	if javaValueSortCompare(key, a[base+hint]) < 0 {
		maxOfs := hint + 1
		for ofs < maxOfs && javaValueSortCompare(key, a[base+hint-ofs]) < 0 {
			lastOfs = ofs
			ofs = int(int32(ofs)<<1) + 1
			if ofs <= 0 {
				ofs = maxOfs
			}
		}
		if ofs > maxOfs {
			ofs = maxOfs
		}
		lastOfs, ofs = hint-ofs, hint-lastOfs
	} else {
		maxOfs := length - hint
		for ofs < maxOfs && javaValueSortCompare(key, a[base+hint+ofs]) >= 0 {
			lastOfs = ofs
			ofs = int(int32(ofs)<<1) + 1
			if ofs <= 0 {
				ofs = maxOfs
			}
		}
		if ofs > maxOfs {
			ofs = maxOfs
		}
		lastOfs += hint
		ofs += hint
	}
	lastOfs++
	for lastOfs < ofs {
		m := lastOfs + (ofs-lastOfs)/2
		if javaValueSortCompare(key, a[base+m]) < 0 {
			ofs = m
		} else {
			lastOfs = m + 1
		}
	}
	return ofs
}

func (s *javaValueTimSort) ensureCapacity(capacity int) []Value {
	if len(s.tmp) < capacity {
		size := int(int32(uint32(0xffffffff)>>uint(bits.LeadingZeros32(uint32(capacity)))) + 1)
		if size < 0 {
			size = capacity
		} else {
			size = min(size, len(s.a)>>1)
		}
		s.tmp = make([]Value, size)
	}
	return s.tmp
}

func (s *javaValueTimSort) mergeLo(base1, len1, base2, len2 int) {
	a, tmp := s.a, s.ensureCapacity(len1)
	cursor1, cursor2, dest := 0, base2, base1
	copy(tmp[:len1], a[base1:base1+len1])
	a[dest] = a[cursor2]
	dest++
	cursor2++
	len2--
	if len2 == 0 {
		copy(a[dest:dest+len1], tmp[cursor1:cursor1+len1])
		return
	}
	if len1 == 1 {
		copy(a[dest:dest+len2], a[cursor2:cursor2+len2])
		a[dest+len2] = tmp[cursor1]
		return
	}
	minGallop := s.minGallop
outer:
	for {
		count1, count2 := 0, 0
		for {
			if javaValueSortCompare(a[cursor2], tmp[cursor1]) < 0 {
				a[dest] = a[cursor2]
				dest++
				cursor2++
				count2++
				count1 = 0
				len2--
				if len2 == 0 {
					break outer
				}
			} else {
				a[dest] = tmp[cursor1]
				dest++
				cursor1++
				count1++
				count2 = 0
				len1--
				if len1 == 1 {
					break outer
				}
			}
			if (count1 | count2) >= minGallop {
				break
			}
		}
		for {
			count1 = javaValueGallopRight(a[cursor2], tmp, cursor1, len1, 0)
			if count1 != 0 {
				copy(a[dest:dest+count1], tmp[cursor1:cursor1+count1])
				dest += count1
				cursor1 += count1
				len1 -= count1
				if len1 <= 1 {
					break outer
				}
			}
			a[dest] = a[cursor2]
			dest++
			cursor2++
			len2--
			if len2 == 0 {
				break outer
			}
			count2 = javaValueGallopLeft(tmp[cursor1], a, cursor2, len2, 0)
			if count2 != 0 {
				copy(a[dest:dest+count2], a[cursor2:cursor2+count2])
				dest += count2
				cursor2 += count2
				len2 -= count2
				if len2 == 0 {
					break outer
				}
			}
			a[dest] = tmp[cursor1]
			dest++
			cursor1++
			len1--
			if len1 == 1 {
				break outer
			}
			minGallop--
			if count1 < 7 && count2 < 7 {
				break
			}
		}
		if minGallop < 0 {
			minGallop = 0
		}
		minGallop += 2
	}
	s.minGallop = max(1, minGallop)
	if len1 == 1 {
		copy(a[dest:dest+len2], a[cursor2:cursor2+len2])
		a[dest+len2] = tmp[cursor1]
	} else if len1 == 0 {
		panic(NewIllegalArgumentException("Comparison method violates its general contract!"))
	} else {
		copy(a[dest:dest+len1], tmp[cursor1:cursor1+len1])
	}
}

func (s *javaValueTimSort) mergeHi(base1, len1, base2, len2 int) {
	a, tmp := s.a, s.ensureCapacity(len2)
	copy(tmp[:len2], a[base2:base2+len2])
	cursor1, cursor2, dest := base1+len1-1, len2-1, base2+len2-1
	a[dest] = a[cursor1]
	dest--
	cursor1--
	len1--
	if len1 == 0 {
		copy(a[dest-(len2-1):dest+1], tmp[:len2])
		return
	}
	if len2 == 1 {
		dest -= len1
		cursor1 -= len1
		copy(a[dest+1:dest+1+len1], a[cursor1+1:cursor1+1+len1])
		a[dest] = tmp[cursor2]
		return
	}
	minGallop := s.minGallop
outer:
	for {
		count1, count2 := 0, 0
		for {
			if javaValueSortCompare(tmp[cursor2], a[cursor1]) < 0 {
				a[dest] = a[cursor1]
				dest--
				cursor1--
				count1++
				count2 = 0
				len1--
				if len1 == 0 {
					break outer
				}
			} else {
				a[dest] = tmp[cursor2]
				dest--
				cursor2--
				count2++
				count1 = 0
				len2--
				if len2 == 1 {
					break outer
				}
			}
			if (count1 | count2) >= minGallop {
				break
			}
		}
		for {
			count1 = len1 - javaValueGallopRight(tmp[cursor2], a, base1, len1, len1-1)
			if count1 != 0 {
				dest -= count1
				cursor1 -= count1
				len1 -= count1
				copy(a[dest+1:dest+1+count1], a[cursor1+1:cursor1+1+count1])
				if len1 == 0 {
					break outer
				}
			}
			a[dest] = tmp[cursor2]
			dest--
			cursor2--
			len2--
			if len2 == 1 {
				break outer
			}
			count2 = len2 - javaValueGallopLeft(a[cursor1], tmp, 0, len2, len2-1)
			if count2 != 0 {
				dest -= count2
				cursor2 -= count2
				len2 -= count2
				copy(a[dest+1:dest+1+count2], tmp[cursor2+1:cursor2+1+count2])
				if len2 <= 1 {
					break outer
				}
			}
			a[dest] = a[cursor1]
			dest--
			cursor1--
			len1--
			if len1 == 0 {
				break outer
			}
			minGallop--
			if count1 < 7 && count2 < 7 {
				break
			}
		}
		if minGallop < 0 {
			minGallop = 0
		}
		minGallop += 2
	}
	s.minGallop = max(1, minGallop)
	if len2 == 1 {
		dest -= len1
		cursor1 -= len1
		copy(a[dest+1:dest+1+len1], a[cursor1+1:cursor1+1+len1])
		a[dest] = tmp[cursor2]
	} else if len2 == 0 {
		panic(NewIllegalArgumentException("Comparison method violates its general contract!"))
	} else {
		copy(a[dest-(len2-1):dest+1], tmp[:len2])
	}
}
