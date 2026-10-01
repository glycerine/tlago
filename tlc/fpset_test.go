package tlc

import (
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestMultiFPSetConfigurationRejectsZeroFingerprintBudgetLikeJava(t *testing.T) {
	cfg := NewFPSetConfiguration()
	cfg.SetMemory(64)
	cfg.SetFPBits(21)
	defer func() {
		if recovered := recover(); recovered != "Given fpSetConfig results in zero or negative fp count." {
			t.Fatalf("panic = %v, want Java zero-fp-count message", recovered)
		}
	}()
	_ = NewMultiFPSetConfiguration(cfg)
}

func TestFPSetFactoryLoadFailuresReturnNilLikeJava(t *testing.T) {
	ClearMessageRecorders()
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	t.Cleanup(func() {
		RemoveMessageRecorder(recorder)
		ClearMessageRecorders()
	})

	for _, implementation := range []string{"com.example.DoesNotExist", "tlc2.tool.fp.DiskFPSet", "tlc2.tool.fp.HeapBasedDiskFPSet"} {
		cfg := NewFPSetConfiguration()
		cfg.NoNesting = true
		cfg.SetImplementation(implementation)
		if set := NewFPSet(cfg); set != nil {
			t.Fatalf("NewFPSet(%q) = %T, want nil like Java reflection failure", implementation, set)
		}
	}

	records := recorder.Records(ECGeneral)
	if len(records) != 3 {
		t.Fatalf("warning count = %d, want 3", len(records))
	}
	for i, record := range records {
		if record.Severity != SeverityWarning {
			t.Fatalf("record %d severity = %v, want warning", i, record.Severity)
		}
		if len(record.Params) != 1 || !strings.Contains(record.Params[0], "unsuccessfully trying to load custom FPSet class: ") {
			t.Fatalf("record %d params = %v, want Java load-failure warning", i, record.Params)
		}
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

func TestDiskFPSetDuplicateMergeUsesJavaErrorCode(t *testing.T) {
	set := NewMSBDiskFPSet(NewFPSetConfiguration())
	set.Init(1, t.TempDir(), "duplicate-merge")
	defer set.Close()

	fp := uint64(42)
	idx := set.getIndex(fp)
	set.tbl[idx] = make([]uint64, diskFPSetInitialBucketCap)
	set.tbl[idx][0] = fp
	set.tblCnt = 1
	if err := set.flushTable(); err != nil {
		t.Fatalf("initial flush returned error: %v", err)
	}

	set.tbl[idx][0] = fp
	set.tblCnt = 1
	err := set.flushTable()
	tlcErr, ok := err.(*TLCError)
	if !ok {
		t.Fatalf("duplicate flush error = %T %v, want TLCError", err, err)
	}
	if tlcErr.Code != ECTLCFPValueAlreadyOnDisk {
		t.Fatalf("error code = %d, want %d", tlcErr.Code, ECTLCFPValueAlreadyOnDisk)
	}
	want := "DiskFPSet.mergeNewEntries: 42 is already on disk.\n"
	if tlcErr.Error() != want {
		t.Fatalf("error = %q, want %q", tlcErr.Error(), want)
	}
}

func TestOffHeapDiskFPSetContainsDoesNotCountMemoryHitsLikeJava(t *testing.T) {
	cfg := NewFPSetConfiguration()
	cfg.SetMemory(64)
	set := NewOffHeapDiskFPSet(cfg)
	set.Init(1, t.TempDir(), "offheap-counters")
	defer set.Close()

	if seen := set.Put(42); seen {
		t.Fatalf("first Put returned seen")
	}
	if got := set.GetMemHitCnt(); got != 0 {
		t.Fatalf("mem hit count after fresh put = %d, want 0", got)
	}
	if !set.Contains(42) {
		t.Fatalf("Contains missed in-memory fingerprint")
	}
	if got := set.GetMemHitCnt(); got != 0 {
		t.Fatalf("offheap Contains mem hit count = %d, want 0 like Java OffHeapDiskFPSet.contains", got)
	}
	if !set.Put(42) {
		t.Fatalf("second Put did not report existing fingerprint")
	}
	if got := set.GetMemHitCnt(); got != 0 {
		t.Fatalf("offheap pre-index duplicate Put mem hit count = %d, want 0 like Java memInsert0 path", got)
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
	offHeap := &OffHeapDiskFPSet{NonCheckpointableDiskFPSet: &NonCheckpointableDiskFPSet{DiskFPSet: &DiskFPSet{}}}
	if err := offHeap.RecoverFile("offheap"); err != nil {
		t.Fatalf("RecoverFile returned error: %v", err)
	}
	multi := &MultiFPSet{Sets: []FPSet{offHeap}}
	if err := multi.BeginChkpt(); err != nil {
		t.Fatalf("MultiFPSet BeginChkpt returned error: %v", err)
	}
	if err := multi.CommitChkpt(); err != nil {
		t.Fatalf("MultiFPSet CommitChkpt returned error: %v", err)
	}
	if err := multi.BeginChkptFile("multi"); err != nil {
		t.Fatalf("MultiFPSet BeginChkptFile returned error: %v", err)
	}

	records := recorder.Records(ECGeneral)
	if len(records) != 4 {
		t.Fatalf("warning count = %d, want 4", len(records))
	}
	want := []string{
		"Checkpointing is not implemented for tlc2.tool.fp.NonCheckpointableDiskFPSet",
		"Checkpointing is not implemented for tlc2.tool.fp.NonCheckpointableDiskFPSet",
		"Checkpointing is not implemented for tlc2.tool.fp.OffHeapDiskFPSet",
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

func TestMultiFPSetExitDelegatesCleanupLikeJava(t *testing.T) {
	metadir := t.TempDir()
	artifact := filepath.Join(metadir, "owned-by-child")
	if err := os.WriteFile(artifact, []byte("preserve"), 0o644); err != nil {
		t.Fatalf("WriteFile artifact: %v", err)
	}
	child := &cleanupRecordingFPSet{NoopFPSet: NewNoopFPSet(nil)}
	set := &MultiFPSet{Sets: []FPSet{child}, metadir: metadir}
	if err := set.Exit(true); err != nil {
		t.Fatalf("Exit returned error: %v", err)
	}
	if !child.exitCalled || !child.cleanupArg {
		t.Fatalf("child exit called/cleanup = %v/%v, want true/true", child.exitCalled, child.cleanupArg)
	}
	if _, err := os.Stat(artifact); err != nil {
		t.Fatalf("MultiFPSet removed metadir directly; Java delegates cleanup to child FPSets: %v", err)
	}
}

type cleanupRecordingFPSet struct {
	*NoopFPSet
	exitCalled bool
	cleanupArg bool
}

func (s *cleanupRecordingFPSet) Exit(cleanup bool) error {
	s.exitCalled = true
	s.cleanupArg = cleanup
	return nil
}

func TestFPSetCommitCheckpointErrorsUseJavaClassNames(t *testing.T) {
	metadir := t.TempDir()
	cases := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "disk",
			err:  func() error { set := NewDiskFPSet(nil); set.metadir = metadir; return set.CommitChkptFile("disk") }(),
			want: "DiskFPSet.commitChkpt: cannot delete " + filepath.Join(metadir, "disk.fp.chkpt"),
		},
		{
			name: "mem1",
			err:  (&MemFPSet1{metadir: metadir}).CommitChkptFile("mem1"),
			want: "MemFPSet1.commitChkpt: cannot delete " + filepath.Join(metadir, "mem1.fp.chkpt"),
		},
		{
			name: "mem2",
			err:  (&MemFPSet2{metadir: metadir}).CommitChkptFile("mem2"),
			want: "MemFPSet2.commitChkpt: cannot delete " + filepath.Join(metadir, "mem2.fp.chkpt"),
		},
	}
	for _, tc := range cases {
		if tc.err == nil {
			t.Fatalf("%s CommitChkptFile returned nil, want Java-shaped error", tc.name)
		}
		if got := tc.err.Error(); !strings.Contains(got, tc.want) {
			t.Fatalf("%s CommitChkptFile error = %q, want to contain %q", tc.name, got, tc.want)
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

func TestMemFPSet2CheckFPsPreservesJavaCrossBucketQuirk(t *testing.T) {
	set := &MemFPSet2{table: make([][]byte, 3)}
	set.table[0] = []byte{0, 0, 0, 0, 0}
	set.table[2] = []byte{0xff, 0xff, 0xff, 0xff, 0x7f}
	if got := set.CheckFPs(); got != 2 {
		t.Fatalf("CheckFPs = %d, want Java's current-bucket cross-bucket distance 2", got)
	}
}
