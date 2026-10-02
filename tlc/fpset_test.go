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

func TestFPSetFactoryKnownLoadFailuresReturnNilLikeJava(t *testing.T) {
	ClearMessageRecorders()
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	t.Cleanup(func() {
		RemoveMessageRecorder(recorder)
		ClearMessageRecorders()
	})

	for _, implementation := range []string{
		"tlc2.tool.fp.FPSet",
		"tlc2.tool.fp.DiskFPSet",
		"tlc2.tool.fp.HeapBasedDiskFPSet",
		"tlc2.tool.fp.NonCheckpointableDiskFPSet",
	} {
		cfg := NewFPSetConfiguration()
		cfg.NoNesting = true
		cfg.SetImplementation(implementation)
		if set := NewFPSet(cfg); set != nil {
			t.Fatalf("NewFPSet(%q) = %T, want nil like Java reflection failure", implementation, set)
		}
	}

	records := recorder.Records(ECGeneral)
	if len(records) != 4 {
		t.Fatalf("warning count = %d, want 4", len(records))
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

func TestFPSetFactoryUnsupportedImplementationFallsBackLikeJava(t *testing.T) {
	ClearMessageRecorders()
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	t.Cleanup(func() {
		RemoveMessageRecorder(recorder)
		ClearMessageRecorders()
	})

	for _, implementation := range []string{"com.example.DoesNotExist", ""} {
		cfg := NewFPSetConfiguration()
		cfg.NoNesting = true
		cfg.SetImplementation(implementation)
		if set := NewFPSet(cfg); reflect.TypeOf(set) != reflect.TypeOf(&MSBDiskFPSet{}) {
			t.Fatalf("NewFPSet(%q) = %T, want MSBDiskFPSet fallback", implementation, set)
		}
	}

	records := recorder.Records(ECTLCFeatureUnsupported)
	if len(records) != 2 {
		t.Fatalf("unsupported warning count = %d, want 2", len(records))
	}
	for i, record := range records {
		if record.Severity != SeverityWarning {
			t.Fatalf("record %d severity = %v, want warning", i, record.Severity)
		}
		if !strings.Contains(record.Text, "Selected fingerprint set (set of visited states)") ||
			!strings.Contains(record.Text, "Reverting to default fingerprint set.") {
			t.Fatalf("record %d text = %q, want Java unsupported-feature message", i, record.Text)
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

func TestOffHeapDiskFPSetDuplicateMergeWarnsLikeJava(t *testing.T) {
	ClearMessageRecorders()
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	t.Cleanup(func() {
		RemoveMessageRecorder(recorder)
		ClearMessageRecorders()
	})

	cfg := NewFPSetConfiguration()
	cfg.SetMemory(64)
	set := NewOffHeapDiskFPSet(cfg)
	set.Init(1, t.TempDir(), "offheap-duplicate-merge")
	defer set.Close()

	if err := set.mergeOffHeapValues([]uint64{42}); err != nil {
		t.Fatalf("initial merge returned error: %v", err)
	}
	if len(set.index) != 1 {
		t.Fatalf("initial index length = %d, want Java offheap special-case length 1", len(set.index))
	}
	if err := set.mergeOffHeapValues([]uint64{42}); err != nil {
		t.Fatalf("duplicate merge returned error: %v", err)
	}
	if len(set.index) != 2 {
		t.Fatalf("duplicate index length = %d, want 2", len(set.index))
	}
	if set.fileCnt != 2 {
		t.Fatalf("fileCnt = %d, want Java table-count increment quirk 2", set.fileCnt)
	}
	records := recorder.Records(ECTLCFPValueAlreadyOnDisk)
	if len(records) != 1 {
		t.Fatalf("warning count = %d, want 1", len(records))
	}
	if records[0].Severity != SeverityWarning {
		t.Fatalf("severity = %v, want warning", records[0].Severity)
	}
	if got := records[0].Text; got != "DiskFPSet.mergeNewEntries: 42 is already on disk.\n" {
		t.Fatalf("message = %q", got)
	}
}

func TestDiskFPSetRecoverDuplicateUsesJavaCheckpointCorruptError(t *testing.T) {
	set := NewMSBDiskFPSet(NewFPSetConfiguration())
	set.Init(1, t.TempDir(), "recover-duplicate")
	defer set.Close()

	if err := set.RecoverFP(42); err != nil {
		t.Fatalf("first RecoverFP returned error: %v", err)
	}
	err := set.RecoverFP(42)
	tlcErr, ok := err.(*TLCError)
	if !ok {
		t.Fatalf("duplicate RecoverFP error = %T %v, want TLCError", err, err)
	}
	if tlcErr.Code != ECSystemCheckpointRecoveryCorrupt {
		t.Fatalf("error code = %d, want %d", tlcErr.Code, ECSystemCheckpointRecoveryCorrupt)
	}
	want := "TLC encountered the following error while restarting from a checkpoint;\n the checkpoint file is probably corrupted.\n"
	if tlcErr.Error() != want {
		t.Fatalf("error = %q, want %q", tlcErr.Error(), want)
	}
}

func TestMemoryFPSetRecoverDuplicateUsesJavaFPNotInSetError(t *testing.T) {
	cases := []struct {
		name string
		set  FPSet
	}{
		{name: "MemFPSet", set: NewMemFPSet()},
		{name: "MemFPSet1", set: NewMemFPSet1(NewFPSetConfiguration())},
		{name: "MemFPSet2", set: NewMemFPSet2(NewFPSetConfiguration())},
		{name: "MultiFPSet", set: &MultiFPSet{Sets: []FPSet{NewMemFPSet()}, Shift: 64}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.set.RecoverFP(42); err != nil {
				t.Fatalf("first RecoverFP returned error: %v", err)
			}
			err := tc.set.RecoverFP(42)
			tlcErr, ok := err.(*TLCError)
			if !ok {
				t.Fatalf("duplicate RecoverFP error = %T %v, want TLCError", err, err)
			}
			if tlcErr.Code != ECTLCFPNotInSet {
				t.Fatalf("error code = %d, want %d", tlcErr.Code, ECTLCFPNotInSet)
			}
			if tlcErr.Error() != "The fingerprint is not in set." {
				t.Fatalf("error = %q", tlcErr.Error())
			}
		})
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

func TestMemFPSetRehashMirrorsJavaBucketSplitOrder(t *testing.T) {
	set := &MemFPSet{
		table:     make([][]uint64, 4),
		threshold: 80,
		mask:      3,
	}
	set.table[0] = []uint64{0, 4, 8, 12}
	set.count = uint64(len(set.table[0]))
	set.rehash()

	if got, want := set.table[0], []uint64{8, 0}; !reflect.DeepEqual(got, want) {
		t.Fatalf("left split bucket = %v, want Java reverse-copy order %v", got, want)
	}
	if got, want := set.table[4], []uint64{12, 4}; !reflect.DeepEqual(got, want) {
		t.Fatalf("right split bucket = %v, want Java reverse-copy order %v", got, want)
	}
}

func TestMemFPSetRehashReusesOneSidedBucketsLikeJava(t *testing.T) {
	set := &MemFPSet{
		table:     make([][]uint64, 4),
		threshold: 80,
		mask:      3,
	}
	original := []uint64{0, 8, 16}
	set.table[0] = original
	set.count = uint64(len(original))
	set.rehash()

	if len(set.table[0]) == 0 || &set.table[0][0] != &original[0] {
		t.Fatalf("one-sided bucket was copied; Java reuses the original bucket slice")
	}
	if set.table[4] != nil {
		t.Fatalf("empty split bucket = %v, want nil like Java", set.table[4])
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

	overflow := NewMemFPSet()
	for _, value := range []uint64{0, 1 << 63} {
		overflow.Put(value)
	}
	if got := overflow.CheckFPs(); got != 1<<63 {
		t.Fatalf("overflow CheckFPs = %d, want Java Math.abs(Long.MIN_VALUE) bit pattern", got)
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
