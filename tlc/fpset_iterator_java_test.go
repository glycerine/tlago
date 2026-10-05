// Copyright (c) 2012 Microsoft Corporation. All rights reserved.
package tlc

import "testing"

// Original TLCIteratorTest.getArray and all three concrete getBuffer variants.
func javaTLCIteratorBuffer(variant int) [][]uint64 {
	getArray := func(length int, offset uint64, elements int) []uint64 {
		a := make([]uint64, length)
		for i := 0; i < length && i < elements; i++ {
			a[i] = uint64(i) + offset
		}
		return a
	}
	buff := make([][]uint64, 8)
	buff[0] = getArray(8, 1, 8)
	buff[1] = getArray(8, 9, 6)
	buff[4] = getArray(8, 15, 3)
	buff[6] = getArray(8, 18, 4)
	if variant == 1 {
		buff[3] = []uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}
	}
	if variant == 2 {
		buff[4] = getArray(8, 15, 7)
		buff[6] = []uint64{22 | (uint64(1) << 63), 0, 0, 0}
	}
	return buff
}

// Original TLCIteratorTest's three methods; each uses fresh @Before state,
// including inherited methods under TLCIterator1Test and TLCIterator2Test.
func runJavaTLCIteratorSourceMethods(t *testing.T, variant int) {
	t.Helper()
	t.Run("testNext", func(t *testing.T) {
		itr := newMSBDiskIterator(javaTLCIteratorBuffer(variant))
		var predecessor int64 = -1
		i := 0
		for i < 21 {
			i++
			if !itr.hasNext() {
				t.Fatal("hasNext false")
			}
			next, err := itr.next()
			if err != nil {
				t.Fatal(err)
			}
			if predecessor != -1 && predecessor >= int64(next) {
				t.Fatalf("predecessor%d is not below%d", predecessor, next)
			}
			predecessor = int64(next)
		}
		if got := itr.reads(); got != int64(i) {
			t.Fatalf("reads=%d, want%d", got, i)
		}
	})
	t.Run("testNoNext", func(t *testing.T) {
		itr := newMSBDiskIterator(javaTLCIteratorBuffer(variant))
		for itr.hasNext() {
			if _, err := itr.next(); err != nil {
				t.Fatal(err)
			}
		}
		if itr.hasNext() {
			t.Fatal("hasNext true at end")
		}
		_, err := itr.next()
		if _, ok := err.(*NoSuchElementException); !ok {
			t.Fatalf("Must throw NoSuchElementException; got%T:%v", err, err)
		}
	})
	t.Run("testGetLast", func(t *testing.T) {
		itr := newMSBDiskIterator(javaTLCIteratorBuffer(variant))
		got, err := itr.getLast()
		if err != nil {
			t.Fatal(err)
		}
		if got != 21 {
			t.Fatalf("last=%d, want21", got)
		}
	})
}

func TestJavaTLCIterator(t *testing.T)  { runJavaTLCIteratorSourceMethods(t, 0) }
func TestJavaTLCIterator1(t *testing.T) { runJavaTLCIteratorSourceMethods(t, 1) }
func TestJavaTLCIterator2(t *testing.T) { runJavaTLCIteratorSourceMethods(t, 2) }
