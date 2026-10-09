# Java TLC bug: off-heap checks reuse a shut-down flusher

Recorded October 9, 2026. Upstream checkout: `tlaplus/tlaplus`, commit
`8f4bc8b73ad1202774a6bf70143436f8ba50aab0`.

## Summary and impact

`OffHeapDiskFPSet.getFlusher` can return an earlier concurrent flusher after its
executor has been shut down by a completed merge. This occurs when an earlier
flush qualifies for parallel preparation but the final fingerprint check uses
more threads and no longer satisfies the minimum partition-size condition.
`checkFPs()` then calls `prepareTable()` on the closed executor and receives
`java.util.concurrent.RejectedExecutionException`.

Direct invariant checks expose the same lifecycle defect without invoking the
selector. `DiskFPSet.checkInvariant(long)` calls the retained flusher's
`flushTable()`; after an eviction closes the executor, newly inserted entries
cause the next invariant check to submit to that closed executor. The original
`OffHeapDiskFPSetLongTest.testMultipleFlushes` reaches this path.

In the faithful Go translation, the distributed EWD840 model completes state
exploration but final fingerprint checking reports `EC.GENERAL` (1000), violating
the original test's no-GENERAL assertion. This is a local storage/executor
lifecycle error, independent of Java RMI or the Go transport. These observations
do not establish fingerprint loss or disk corruption.

## Affected Java source

All paths below are relative to `tlatools/org.lamport.tlatools/` in the pinned
upstream checkout. Line numbers refer to that commit.

| Location | Relevant behavior |
| --- | --- |
| `src/tlc2/tool/fp/OffHeapDiskFPSet.java:116` | Default `PROBE_LIMIT` is 1024. |
| Same file, lines 190–210 | Eviction selects a flusher using the store's reader/thread count. |
| Same file, lines 228–234 | Parallel eligibility creates a new flusher; fallback returns `this.flusher`. |
| Same file, lines 957–965 | Concurrent flusher owns a fixed thread pool. |
| Same file, lines 972–1035 | `prepareTable()` submits tasks to that pool. |
| Same file, lines 1124–1132 | Merge submits work, shuts down the pool and awaits termination; finally also shuts it down. |
| Same file, lines 498–500 | Final check selects using `TLCGlobals.getNumWorkers()`, then calls `prepareTable()`. |
| `src/tlc2/tool/fp/DiskFPSet.java:810–838` | Invariant checks flush the retained instance directly and compare its size with the expected count. |

The problematic selector is:

```java
private Flusher getFlusher(final int numThreads, final long insertions) {
    if (array.size() >= 8192
            && Math.floor(array.size() / (double) numThreads) > 2 * PROBE_LIMIT) {
        return new ConcurrentOffHeapMSBFlusher(array, PROBE_LIMIT, numThreads, insertions);
    } else {
        return this.flusher;
    }
}
```

A concurrent flusher is effectively single-use after merge. Returning it from
the fallback neither restores its executor nor adapts its partition geometry.

## Deterministic reproduction recipe

In a Java test in package `tlc2.tool.fp`, initialize an off-heap set with exactly
8192 fingerprint slots, default probe limit 1024, and one reader thread. Save
and restore the global TLC worker count around the test, and close the set.

1. Insert fingerprints 41 and 97.
2. Call `evict()` to perform an actual concurrent preparation and merge.
   With one thread, 8192 / 1 > 2048, so selection creates a concurrent flusher.
   The merge commits the entries and shuts down that flusher's executor.
3. Insert fingerprints 131 and 197 into the now available memory table.
4. Set `TLCGlobals`' worker count to 48 and call `checkFPs()`.
   floor(8192 / 48) = 170, so parallel eligibility fails. The fallback returns
   the old concurrent flusher. Its next task submission is rejected.

Four final-check workers also reproduce the boundary: 8192 / 4 = 2048, and the
eligibility comparison is strictly greater than 2048. Three workers still
qualify and select a fresh concurrent flusher.

The production-profile reproduction uses the original distributed EWD840 test,
Ant's off-heap/512 KiB fingerprint profile, and 48 CPU-derived workers. Two
local children receive 32768 slots apiece. Eviction with one thread qualifies;
final checking with 48 threads gives floor(32768 / 48) = 682 and falls back to
the closed flusher.

## Evidence and limits

The selector and shutdown lifecycle above were inspected in the pinned Java
source. The real-merge regression and full distributed reproduction described
here were executed in the Go translation. **An unmodified Java execution of
this new small regression was not performed for this report.** The Java recipe
above is supplied for upstream confirmation; it is not a claimed Java test run.

Before the correction, the small Go regression panicked with
`RejectedExecutionException` through `invokeAll -> prepareTable -> CheckFPs`.
The earlier full distributed EWD840 diagnostic failed with GENERAL through
`invokeAll -> prepareTable -> CheckFPs -> MultiFPSet.CheckFPs`.

The upstream shared harness already contains an unconditional false assumption
at `test/tlc2/tool/distributed/DistributedTLCTestCase.java:68`, with the reason
`DistributedTLCTestCase broken with OffHeapDiskFPSet.` This prevents ordinary
execution from detecting this regression. We have not established that this
particular defect was the historical reason for that assumption.

## Suggested Java correction and Go divergence

Return a sequential flusher when the current partition sizes do not qualify:

```java
} else {
    return new OffHeapMSBFlusher(array);
}
```

The sequential constructor already exists at line 1191. Keep creation of a fresh
concurrent flusher in the eligible branch. Do not reopen a closed executor or
weaken the final-check assertions to accept GENERAL.

The user explicitly authorized this correction in Go. In
[tlc/offheap_concurrent_flusher.go](tlc/offheap_concurrent_flusher.go), the fallback
clears `concurrentFlusher`; nil selects the existing sequential preparation and
merge implementation. Its comment records the deliberate divergence from buggy
Java. Direct flushing also clears a retained concurrent flusher when its executor
is already shut down after a successful flush, covering invariant checks and
recovery without counting an extra eviction or reopening the executor. Failed
flushes retain their existing failure state. The off-heap closest-pair calculation
still approximates using retained in-memory fingerprints; this fix does not
change it into a full disk scan.

## Go verification

Focused regressions are in
[tlc/distributed_offheap_flusher_selection_test.go](tlc/distributed_offheap_flusher_selection_test.go).
They cover the strict partition boundary, fresh eligible selection, an actual
merge followed by final checking, and preservation of all four fingerprints.
`TestOffHeapInvariantAfterConcurrentFlush` additionally reproduces the direct
invariant path after an actual executor shutdown and checks four successive
membership counts. It failed with `RejectedExecutionException` before the
direct-flush correction.

```sh
go test ./tlc -run '^(TestDistributedOffHeapFlusherFallsBackWhenPartitionsTooSmall|TestDistributedOffHeapFinalCheckAfterConcurrentFlush)$' -count=1 -v
go test . -tags=tlago_disabled_distributed_tests -run '^TestDiagnosticJavaEWD840Distributed$' -count=1 -timeout=10m -v
go test ./tlc -run '^TestOffHeapInvariantAfterConcurrentFlush$' -count=1 -v
go test ./tlc -tags=tlc_fp_stress -run '^TestJavaOffHeapDiskFPSetLong_testMultipleFlushes$' -count=1 -timeout=30m -v
```

After the fix, the focused selection plus original Java insertion/merge controls
passed in 0.078 seconds. The full original-profile EWD840 diagnostic passed in
74.80 seconds with its original assertions, 114942 distinct states, zero queued
states, no GENERAL, and normal worker/server exits. Source-skipped ordinary test
entries remain separate from this explicit diagnostic; no original-method
completion credit was added. No full suite or race-instrumented workload ran.

The original `OffHeapDiskFPSetLongTest.testMultipleFlushes` is now translated in
[tlc/offheap_multiple_flushes_java_test.go](tlc/offheap_multiple_flushes_java_test.go),
behind `tlc_fp_stress`. With the direct-flush correction it passes all four
default rounds of 8,388,608 insertions, all put assertions and every exact-count
invariant check in 317.29 seconds. It retains seed 15041980, the source ratio-1.0
factory and one reader. This earns one supplementary long-test method credit;
distributed original-method credit remains unchanged. The final focused controls
pass in 0.272 seconds. Receipts:
`.codex-gotmp/offheap-invariant-flusher-before.log`,
`.codex-gotmp/offheap-invariant-flusher-final.log`, and
`.codex-gotmp/offheap-multiple-flushes-final.log`.

Local execution receipts (ignored development logs):
`.codex-gotmp/offheap-flusher-bug-before.log`,
`.codex-gotmp/offheap-flusher-bug-fixed.log`, and
`.codex-gotmp/offheap-flusher-ewd840-fixed.log`.
