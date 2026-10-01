package tlc

import (
	"math"
	"reflect"
	"testing"
)

func TestFPSetConfigurationFactoryBehaviors(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.LSBDiskFPSet")
	cfg := NewFPSetConfiguration()
	if got := cfg.GetImplementation(); got != "tlc2.tool.fp.LSBDiskFPSet" {
		t.Fatalf("implementation from property = %q, want LSBDiskFPSet", got)
	}
	if !isDiskFPSetImplementation("tlc2.tool.fp.NonCheckpointableDiskFPSet") {
		t.Fatalf("NonCheckpointableDiskFPSet should be recognized as disk-backed")
	}
	if got := GetFPSetVMArguments("tlc2.tool.fp.MSBDiskFPSet", 2048); got != "-Xmx2048m" {
		t.Fatalf("heap VM args = %q", got)
	}
	if got := GetFPSetVMArguments("tlc2.tool.fp.OffHeapDiskFPSet", 2048); got != "-XX:MaxDirectMemorySize=2048m" {
		t.Fatalf("offheap VM args = %q", got)
	}
	want := []string{
		"tlc2.tool.fp.MSBDiskFPSet",
		"tlc2.tool.fp.LSBDiskFPSet",
		"tlc2.tool.fp.OffHeapDiskFPSet",
	}
	if got := GetFPSetImplementations(); !reflect.DeepEqual(got, want) {
		t.Fatalf("implementations = %v, want %v", got, want)
	}
}

func TestDiskFPSetLockCountMirrorsJavaDefaultAndOverride(t *testing.T) {
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })

	t.Setenv(DiskFPSetLogLockCntProperty, "")
	t.Setenv("TLAGO_DISK_FPSET_LOG_LOCK_CNT", "")
	if got := NewMSBDiskFPSet(NewFPSetConfiguration()).GetLockCnt(); got != 256 {
		t.Fatalf("default lock count with one worker = %d, want 256", got)
	}

	SetNumWorkers(3)
	if got := NewMSBDiskFPSet(NewFPSetConfiguration()).GetLockCnt(); got != 512 {
		t.Fatalf("default lock count with three workers = %d, want 512", got)
	}

	t.Setenv("TLAGO_DISK_FPSET_LOG_LOCK_CNT", "4")
	if got := NewMSBDiskFPSet(NewFPSetConfiguration()).GetLockCnt(); got != 16 {
		t.Fatalf("override lock count = %d, want 16", got)
	}
}

func TestDiskFPSetReaderSelectionUsesCurrentWorkerIDLikeIdThread(t *testing.T) {
	set := NewMSBDiskFPSet(NewFPSetConfiguration())
	set.Init(2, t.TempDir(), "reader-selection")
	defer set.Close()

	raf, pooled, err := set.openDiskReader()
	if err != nil {
		t.Fatalf("openDiskReader outside worker returned error: %v", err)
	}
	if !pooled {
		t.Fatalf("openDiskReader outside worker used fixed reader, want pool fallback")
	}
	set.poolClose(raf)

	restore := PushCurrentWorkerID(1)
	defer restore()
	raf, pooled, err = set.openDiskReader()
	if err != nil {
		t.Fatalf("openDiskReader for worker returned error: %v", err)
	}
	if pooled {
		t.Fatalf("openDiskReader for worker used pool, want worker-indexed reader")
	}
	if raf != set.braf[1] {
		t.Fatalf("openDiskReader returned %p, want braf[1] %p", raf, set.braf[1])
	}
}

func TestNonCheckpointableDiskFPSetNamedCheckpointWarnings(t *testing.T) {
	ClearMessageRecorders()
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	t.Cleanup(func() {
		RemoveMessageRecorder(recorder)
		ClearMessageRecorders()
	})

	base := &NonCheckpointableDiskFPSet{}
	if err := base.BeginChkptFile("base"); err != nil {
		t.Fatalf("BeginChkptFile returned error: %v", err)
	}
	if err := base.CommitChkptFile("base"); err != nil {
		t.Fatalf("CommitChkptFile returned error: %v", err)
	}
	offHeap := &OffHeapDiskFPSet{}
	if err := offHeap.RecoverFile("offheap"); err != nil {
		t.Fatalf("RecoverFile returned error: %v", err)
	}

	records := recorder.Records(ECGeneral)
	if len(records) != 3 {
		t.Fatalf("warning count = %d, want 3", len(records))
	}
	want := []string{
		"Checkpointing is not implemented for tlc2.tool.fp.NonCheckpointableDiskFPSet",
		"Checkpointing is not implemented for tlc2.tool.fp.NonCheckpointableDiskFPSet",
		"Checkpointing is not implemented for tlc2.tool.fp.OffHeapDiskFPSet",
	}
	for i, record := range records {
		if record.Severity != SeverityWarning {
			t.Fatalf("record %d severity = %v, want warning", i, record.Severity)
		}
		if len(record.Params) != 1 || record.Params[0] != want[i] {
			t.Fatalf("record %d params = %v, want %q", i, record.Params, want[i])
		}
	}
}

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
