package tlc

import (
	"math"
	"testing"
)

func TestMemFPSetPutContainsAndSize(t *testing.T) {
	set := NewMemFPSet()
	if set.Size() != 0 {
		t.Fatalf("new set size = %d, want 0", set.Size())
	}
	if seen := set.Put(42); seen {
		t.Fatalf("first Put returned seen=true")
	}
	if set.Size() != 1 || !set.Contains(42) {
		t.Fatalf("size/contains after first put = %d/%v, want 1/true", set.Size(), set.Contains(42))
	}
	if seen := set.Put(42); !seen {
		t.Fatalf("duplicate Put returned seen=false")
	}
	if set.Size() != 1 {
		t.Fatalf("size after duplicate put = %d, want 1", set.Size())
	}
}

func TestMemFPSetRehashPreservesFingerprints(t *testing.T) {
	set := &MemFPSet{
		table:     make([][]uint64, 2),
		threshold: 2,
		mask:      1,
	}
	values := []uint64{0, 2, 1, 3, 4}
	for _, value := range values {
		if seen := set.Put(value); seen {
			t.Fatalf("Put(%d) returned seen=true before duplicate insert", value)
		}
	}
	if set.Size() != uint64(len(values)) {
		t.Fatalf("size after rehash inserts = %d, want %d", set.Size(), len(values))
	}
	if set.mask < 3 {
		t.Fatalf("mask after inserts = %d, want at least 3 after rehash", set.mask)
	}
	for _, value := range values {
		if !set.Contains(value) {
			t.Fatalf("set does not contain %d after rehash", value)
		}
	}
}

func TestMemFPSetCheckFPsMatchesJavaDistanceBehavior(t *testing.T) {
	empty := NewMemFPSet()
	if got := empty.CheckFPs(); got != uint64(math.MaxInt64) {
		t.Fatalf("empty CheckFPs = %d, want Java Long.MAX_VALUE", got)
	}

	set := NewMemFPSet()
	for _, value := range []uint64{10, 20, 13} {
		set.Put(value)
	}
	if got := set.CheckFPs(); got != 3 {
		t.Fatalf("CheckFPs = %d, want nearest distance 3", got)
	}
}
