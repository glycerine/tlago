// Copyright (c) 2016 Markus Alexander Kuppe. All rights reserved.
package tlc

import "testing"

func assumeJavaLongArraysSupported(t *testing.T) {
	t.Helper()
	if !LongArrayIsSupported() {
		t.Skip("source Assume: LongArray.isSupported()")
	}
}
func javaLongArraysIsInRange(idx, reprobe, pos, size int64) bool {
	if idx+reprobe >= size && pos < idx {
		return pos <= (idx+reprobe)%size
	}
	return idx <= pos && pos <= idx+reprobe
}
func doJavaLongArraysTest(t *testing.T, expected []int64, partitions, reprobe int64, indexer *OffHeapIndexer) {
	t.Helper()
	array := NewLongArrayFrom(expected)
	comparator := func(fpA, posA, fpB, posB int64) int {
		if fpA <= 0 || fpB <= 0 {
			return 0
		}
		wrappedA := indexer.GetIdx(uint64(fpA)) > posA
		wrappedB := indexer.GetIdx(uint64(fpB)) > posB
		if wrappedA == wrappedB && posA > posB {
			if fpA < fpB {
				return -1
			}
			return 1
		} else if wrappedA != wrappedB {
			if posA < posB && fpA < fpB {
				return -1
			}
			if posA > posB && fpA > fpB {
				return -1
			}
		}
		return 0
	}
	length := int64(len(expected)) / partitions
	for i := int64(0); i < partitions; i++ {
		start := i * length
		end := start + length
		if i+1 == partitions {
			end = array.Size() - 1
		}
		LongArraysSortRange(array, start, end, comparator)
	}
	for i := int64(0); i < partitions; i++ {
		end := (i + 1) * length
		if i+1 == partitions {
			end = array.Size() - 1
		}
		LongArraysSortRange(array, end-reprobe, end+reprobe, comparator)
	}
	sentinel := 0
Outer:
	for j, l := range expected {
		if l == 0 {
			if got := array.Get(int64(j)); got != 0 {
				t.Fatalf("sentinel(%d)=%d, want0", j, got)
			}
			sentinel++
		} else if l < 0 {
			if got := array.Get(int64(j)); got != l {
				t.Fatalf("sentinel(%d)=%d, want%d", j, got, l)
			}
			sentinel++
		} else {
			for k := int64(0); k < array.Size(); k++ {
				if array.Get(k) == l {
					continue Outer
				}
			}
			t.Fatalf("long %d not found", l)
		}
	}
	for pos := int64(0); pos < array.Size(); pos++ {
		l := array.Get(pos)
		if l <= 0 {
			continue
		}
		idx := indexer.GetIdx(uint64(l))
		if !javaLongArraysIsInRange(idx, reprobe, pos, array.Size()) {
			t.Fatalf("%d, pos:%d idx:%d r:%d outside range", l, pos, idx, reprobe)
		}
	}
	pos := int64(0)
	seen := make([]int64, 0, len(expected))
	for pos < array.Size() {
		e := array.Get(pos)
		if e <= 0 || indexer.GetIdx(uint64(e)) > pos {
			pos++
			continue
		}
		seen = append(seen, e)
		pos++
		break
	}
	for ; pos < array.Size()+reprobe; pos++ {
		actual := array.Get(pos % array.Size())
		if actual <= 0 {
			continue
		}
		idx := indexer.GetIdx(uint64(actual))
		if pos < array.Size() && idx > pos {
			continue
		}
		if pos > array.Size()-1 && idx+reprobe < pos {
			continue
		}
		seen = append(seen, actual)
	}
	for i := 1; i < len(seen); i++ {
		lo, hi := seen[i-1], seen[i]
		if !(lo < hi) {
			t.Fatalf("%d > %d", lo, hi)
		}
	}
	if got, want := len(seen), len(expected)-sentinel; got != want {
		t.Fatalf("seen=%d, want%d", got, want)
	}
}

// Original LongArraysTest.testEmpty1 and source setup/helpers.
func TestJavaLongArraysEmpty1(t *testing.T) {
	assumeJavaLongArraysSupported(t)
	expected := []int64{}
	doJavaLongArraysTest(t, expected, 1, 0, NewInfinitePrecisionOffHeapIndexer(int64(len(expected)), 1))
}

// Original LongArraysTest.testEmpty2 and source setup/helpers.
func TestJavaLongArraysEmpty2(t *testing.T) {
	assumeJavaLongArraysSupported(t)
	expected := []int64{0, 0, 0, 0}
	doJavaLongArraysTest(t, expected, 1, 2, NewInfinitePrecisionOffHeapIndexer(int64(len(expected)), 1))
}

// Original LongArraysTest.testBasic1 and source setup/helpers.
func TestJavaLongArraysBasic1(t *testing.T) {
	assumeJavaLongArraysSupported(t)
	expected := []int64{5, 8, 1, 7, 0, 3}
	array := NewLongArrayFrom(expected)
	LongArraysSort(array)
	for i := int64(1); i < array.Size(); i++ {
		if !(array.Get(i-1) < array.Get(i)) {
			t.Fatalf("values%d,%d are not strictly increasing", array.Get(i-1), array.Get(i))
		}
	}
}

// Original LongArraysTest.testBasic2 and source setup/helpers.
func TestJavaLongArraysBasic2(t *testing.T) {
	assumeJavaLongArraysSupported(t)
	expected := []int64{74236458333421747, 9185197375878056627, 9017810141411942826, 481170446028802552, 587723185270146839, 764880467681476738, 1028380228728529428, 1246117495100367611, 1353681884824400499, 1963327988900916594, 2157942654452711468, 2211701751588391467, 2197266581704230150, 2391118405386569995, 2754416910109403115, 3528296600587602855, 3766154305485605955, 4172091881329434331, 4273360576593753745, 4338054185482857322, 4487790251341705673, 4760603841378765728, 4897534821030901381, 5057347369431494228, 5185984701076703188, 5255556356599253415, 4911921657882287345, 5512811886280168498, 5627022814159167180, 5630009759945037387, 5592096823142754761, 5880489878946290534, 6796173646113527960, 6887096685265647763, 6946033094922439935, 7100083311060830826, 7575172208974668528, 8240485391672917634, 8572429495433200993, 8804495173596718076, 8771524479740786626, 8986659781390119011, 9136953010061430590, 9195197379878056627}
	array := NewLongArrayFrom(expected)
	LongArraysSort(array)
	for i := int64(1); i < array.Size(); i++ {
		if !(array.Get(i-1) < array.Get(i)) {
			t.Fatalf("values%d,%d are not strictly increasing", array.Get(i-1), array.Get(i))
		}
	}
}

// Original LongArraysTest.test0 and source setup/helpers.
func TestJavaLongArrays0(t *testing.T) {
	assumeJavaLongArraysSupported(t)
	expected := []int64{22102288204167208, 225160948165161873, 0, 1638602644344629957, 1644442600000000000, 0}
	doJavaLongArraysTest(t, expected, 1, 3, NewInfinitePrecisionOffHeapIndexer(int64(len(expected)), 1))
}

// Original LongArraysTest.testIsInRange and source setup/helpers.
func TestJavaLongArraysIsInRange(t *testing.T) {
	assumeJavaLongArraysSupported(t)
	if got := javaLongArraysIsInRange(0, 0, 0, 4); got != true {
		t.Fatal("isInRange(0, 0, 0, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(0, 0, 1, 4); got != false {
		t.Fatal("isInRange(0, 0, 1, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(0, 0, 2, 4); got != false {
		t.Fatal("isInRange(0, 0, 2, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(0, 0, 3, 4); got != false {
		t.Fatal("isInRange(0, 0, 3, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(0, 0, 4, 4); got != false {
		t.Fatal("isInRange(0, 0, 4, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(0, 1, 1, 4); got != true {
		t.Fatal("isInRange(0, 1, 1, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(0, 1, 2, 4); got != false {
		t.Fatal("isInRange(0, 1, 2, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(0, 2, 2, 4); got != true {
		t.Fatal("isInRange(0, 2, 2, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(0, 2, 3, 4); got != false {
		t.Fatal("isInRange(0, 2, 3, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(0, 3, 3, 4); got != true {
		t.Fatal("isInRange(0, 3, 3, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(0, 3, 4, 4); got != false {
		t.Fatal("isInRange(0, 3, 4, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(3, 0, 3, 4); got != true {
		t.Fatal("isInRange(3, 0, 3, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(3, 1, 0, 4); got != true {
		t.Fatal("isInRange(3, 1, 0, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(3, 2, 1, 4); got != true {
		t.Fatal("isInRange(3, 2, 1, 4) differs from source expectation")
	}
	if got := javaLongArraysIsInRange(3, 2, 2, 4); got != false {
		t.Fatal("isInRange(3, 2, 2, 4) differs from source expectation")
	}
}
