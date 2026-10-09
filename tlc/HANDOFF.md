# TLC Port Handoff

Updated: October 8, 2026. Active branch: `master`.

This is the current restart guide. Detailed audit history and verification
receipts belong in [PORT_PROGRESS.md](PORT_PROGRESS.md); implementation contracts
belong in [TLC_ARCH.md](TLC_ARCH.md). Older handoffs remain in Git history and
must not be read as current assignments.

## Goal and boundaries

Continue the original Java distributed TLC algorithm in Go. Leave the
rpc25519/Tube alternative aside. Use native Go networking, goroutines, ownership
and errors; do not implement RMI, Java serialization or a Java runtime. Preserve
the coordinator, worker, fingerprint-manager, trace and checkpoint behavior.

The Go module is `github.com/glycerine/tlago`. The canonical checkout is
`/mnt/b/github.com/tlaplus/tlago`; the environment's workspace path also resolves
there. Java is read-only in `../tlaplus`, pinned to
`8f4bc8b73ad1202774a6bf70143436f8ba50aab0`. Its TLC sources and tests are under
`../tlaplus/tlatools/org.lamport.tlatools/`. Most implementation is package `tlc`;
model loading, CLI integration and process model tests also live in the root.

Email reporting and its JavaMail/SMTP/MIME/Activation/ImageIO/AWT/codec work are
forbidden and excluded from completion. Removal is committed. Preserve generic
exceptions, packaged properties, console output, OpenJDK notices and `x/text`.

The listed SANY audit contracts are complete; this does not prove general SANY
parity beyond those contracts. Do not restart XML or ApalacheIR corpus sweeps.
Other TLC test reconciliation remains visible in [TODO_TEST_PORT.md](TODO_TEST_PORT.md);
JVM-only GC/JPF assertions are not native Go implementation requirements.

## Verification baseline and test credit

The user supplied a green full-suite baseline. Do not rerun that approximately
45-minute suite. The earlier recorded full-workspace run verified `23f046e`:
root 1,552.199 seconds, TLC 771.039 seconds, SANY 2.419 seconds, status 0.
Later distributed changes have focused receipts; they do not establish a new
full-workspace pass. Reuse verification for unchanged code.

The distributed inventory records 41 original methods: 37 port complete and
four missing. The four model methods inherit an unconditional false assumption
in `DistributedTLCTestCase`; do not invent replacement skips or award original
method credit for the native process harness:

| Original class | Missing method |
| --- | --- |
| `DieHardDistributedTLCTest` | `testSpec` |
| `EWD840DistributedTLCTest` | `test` |
| `EWD840DistributedWithFPSetTLCTest` | `test` |
| `TSnapShotDistributedTLCTest` | `test` |

Existing original manager, smart-proxy, initializer, storage, vector and trace
methods retain their source assertions. Native checks supplement transport and
ownership boundaries without changing original-method counts. The overall TLC
and distributed port goals remain incomplete.

An earlier distributed model run failed with its assertion lost to tool output
truncation, then passed unchanged with captured output. Its original failure is
unresolved. Later green runs do not diagnose it. Preserve the note and receipts
in `PORT_PROGRESS.md`; capture output before running future long checks.

Alias trace replay also has a documented scheduling concern. Genuine replay
liveness failures can stop before final actions; both Java and Go check
invariants on successors excluded by the trace constraint. Preserve that ordering
and the original replay assertions. The inventory separately tracks source-failing
fingerprint/concurrent-I/O contexts and missing full-size long workloads; no green
focused selection resolves them or earns their completion credit.

## Current distributed implementation

Native `net/rpc` over TCP supplies coordinator, worker and fingerprint endpoints.
Named publication, discovery, file loading, interning, settings, registration,
manager snapshots and keepalive are wired into the production CLI. Roles are
`server`, `worker`, `fpserver` and `worker-fpserver`. The combined role shares its
listener and waits for both lifetimes. Signal shutdown occurs at the source
shutdown-hook boundary. Accepted replies drain before orderly host shutdown;
forced connection closure remains independent of computation/storage lifetime.
Short TCP checks also hold a worker's accepted fingerprint lookup open while
keepalive/cache/exit calls use the same worker connection. Control calls remain
responsive; exit rejects new calls and lets the accepted computation finish.
The full N=7 model also holds bidirectional TCP relay traffic to one of two FP
hosts without closing connections. Coordinator status/manager and worker
alive/cache probes remain responsive; routing stays distinct and no failover is
reported. Releasing traffic restores the original 114,942-state result. This
controlled byte-relay stall does not prove arbitrary network blackholes.

Coordinator file requests use a fresh resolver, observing current default
library and user directories and owning separate temporary resource copies.
Explicit search overrides remain captured; worker basename caching is unchanged.
Upstream RMI names describe source references only, not a Go transport requirement.

Transport does not redial or replay ambiguous fingerprint mutations. Manager
failover retains the source algorithm, including its forward reassignment and
slot-based statistics. Workers receive independent manager snapshots with shared
registration-wrapper aliases; the coordinator recovery trace is omitted.

`DistributedOperationError` traits drive retry, removal, discovery and exit
choices. Causes remain available through `errors.Is`/`errors.As`; native sender
frames and error graphs survive payload transfer. Memory exhaustion permits
smaller-batch retry; executor rejection does not. Lost fingerprint exit replies
use an exit-only trait, allowing shutdown to visit later registrations. Prior
closed clients and insertion/checkpoint failures remain reportable. The obsolete
`UnmarshalException` carrier and RMI-specific exit branch are removed.

Named `MultiFPSet` checkpoint and recovery operations run child stores
concurrently and join before returning. Child I/O failures propagate as operation
failures with native Go error wrapping; the manager must not ignore them as remote
server outages. Unnamed begin/commit operations retain source sequential ordering.

Failed coordinator construction stops and joins its queue workers and closes
owned trace/fingerprint handles, preserving files and the original failure.
Allocation order and base-constructor error precedence remain unchanged;
successful construction transfers resources to the coordinator.

The native worker/FP network owner retains local fingerprint storage through
unpublication and registration errors. Host shutdown drains accepted replies,
then closes each owned storage object once without invoking Exit or deleting
files. Direct `Host` publications remain caller-owned; worker runtime shutdown
is still separate.
Before registration begins, FP startup failures close allocated storage handles
without deleting files or replacing the failure. Once registration starts,
startup/reporting errors retain storage for the native host or caller to release.

State/result payloads retain identity and nil/empty distinctions across states,
values, strings, predecessor graphs, caches, byte buffers and partition vectors.
Attached model-value data supports native `[]Value` through the same array graph,
preserving cycles, shared backing storage and typed nil/empty arrays across RPC.
Native `map[string]Value` attachments also preserve shared map identity, cycles,
nil entries and typed nil/empty maps in requests, results and failure contexts.
Mixed `[]any` and `map[string]any` attachments retain supported scalars, bytes,
TLC values, typed containers and recursive mixed-container graphs with the same
ownership guarantees. Their compact entries share the scalar data tags; peers
must use the updated native payload schema.
Symbolic values follow the source materialization rules. Unsupported opaque
custom data and evaluator metadata fail explicitly; they are not silently
removed. Audit actual source transferability before extending the codec:
`OpLambdaValue` retains a non-serializable Context, MethodValue has reflection
handles, and ordinary Action predicate/context objects are not made transferable
by Action's Serializable declaration alone.

Coordinator publication inserts fingerprints before writing traces and queueing
selected states. It uses the incoming successor UID as the predecessor location.
Final fingerprint statistics require the existing manager. Missing ownership
fails before publishing a fabricated zero count or success diagnostic.
Missing selected partitions/states/visited vectors fail at source access points;
unused partitions retain lazy access. Worker checks precede constraints. TLCApp
retains captured action/property arrays, source vector concatenation, separate
returned-array ownership and predecessor metadata policy. Replay requires its
existing ordinary evaluator rather than creating a fresh default Tool.

Local `MultiFPSet` adaptation retains one endpoint to the original storage.
Its high-bit child routing must not become the manager's low-bit server routing.
Explicitly registered distributed servers retain their own manager routing.

Thread startup increments workers before coordinator queue capture, outside
catch/finally. Worker-loss cleanup cancels keepalive, claims its one-time flag,
removes the coordinator registration, requeues, clears assigned states, wakes
queue consumers and decrements workers in that order. Missing ownership cannot
silently skip removal. Timer-triggered cleanup accesses the coordinator queue
before entering that sequence. Repeated loss reports do not duplicate work or
registration removal; wakeup preserves queue suspension.

Keepalive retains its original ten-second first run and sixty-second period.
Run-handler and finalizer failures preserve earlier mutations, cache-warning
behavior and uncaught goroutine/timer boundaries. Native trace cleanup closes
owned handles; do not emulate JVM resource leaks. Literal empty-directory trace
prefixes remain disk-backed and support checkpoint commit and cleanup.

## Verified process coverage

`tlc_distributed_process_test.go` runs the unchanged EWD840/MC06 N=7 model in
separate processes. Ordinary coordinator-owned, standalone, partitioned and
combined fingerprint roles require 114,942 distinct states, an empty queue,
FINISHED and no GENERAL. Native DieHard/TSnapShot process models preserve their
original error/trace assertions in `tlc_distributed_trace_process_test.go`.

A two-worker process row exercises the source shared application, fingerprint
manager, executor and exit latch. It requires distinct worker endpoints on one
native listener, work and statistics from each worker, 114,942 distinct states,
an empty queue and clean shutdown with no GENERAL or EOF on either role.

Repeated registration of one native worker preserves two coordinator threads
and shared worker identity. The full model retains 114,942 distinct states and
an empty queue. First exit removes the endpoint; the second receives the native
removed-endpoint failure and the source warning. This adds no original-method
credit and requires no RMI compatibility.

Worker process death with a survivor/replacement, loss of the sole worker, and
loss of a fully computed worker reply are covered. Sole-worker cleanup must
finish while the coordinator remains available with unfinished work before
replacement registration. The reply-loss row keeps its disconnected runtime
alive, requires one EOF smaller-block retry, one deregistration and exactly the
source cache-statistic warning, then requires ordinary final model counts.

The full model also kills the first of two fingerprint hosts with assigned work
and resumes actual evaluation. Worker and coordinator fail over independently.
Two partition slots alias the survivor, so source statistics report 229,884 in
this separate failure row; ordinary original rows still require 114,942.
Another row kills that host after a real successor `putBlock` has inserted new
fingerprints but before its reply returns. Stored membership is verified before
the kill. Source callable failover retries against the survivor and completes
with the same 229,884 slot-based count and empty queue. Only this row requires
one coordinator EOF diagnostic for the deliberately lost insertion reply.

Fresh-process local recovery covers the complete 16,384-state initial frontier,
a mid-run checkpoint and interruptions before queue commit, after queue commit,
after intern commit and after the first nested fingerprint commit. Checks inspect
actual files and recovered counts, not just producer markers. A short assigned-
block barrier check reopens local, one-remote-store and two-remote-store snapshots
with matching queue/trace/fingerprint identities.

A local full-model mid-run checkpoint also runs two workers sharing one process
application and runtime. Fresh coordinator and two replacement workers restore
the persisted frontier and finish with the original 114,942 distinct states.
Both producer registrations, both replacement statistics and actual work by
each replacement worker are required. This does not establish remote multi-worker
restart by itself; see the remote matrix below. Checkpoint atomicity remains
unproved.

Short process checks kill MemFPSet, LSBDiskFPSet and MSBDiskFPSet hosts after a
committed checkpoint plus later insertion/pending snapshot. Fresh hosts recover
only committed membership. Disk rows verify flushed live membership, unchanged
committed bytes and no pending-snapshot promotion. These checks do not establish
full-model remote restart/recovery or checkpoint atomicity.

`tlc_distributed_remote_recovery_test.go` now covers a complete mid-run checkpoint
with two remote hosts using Mem, LSB or MSB storage, followed by loss of the
original coordinator, worker and both hosts. It runs with either one worker or
two workers sharing one application/runtime in each worker process. Disk hosts
use the source factory's two-child MultiFPSet layout. Empty replacement stores
retain committed files and
are registered before source recovery. Their exact partition membership and the
disk queue count must match the snapshot before replacement worker startup.
The recovered N=7 model requires 114,942 distinct states, an empty queue, one
recovery and no GENERAL. Two-worker rows require distinct endpoints on one
listener and actual work/statistics from both replacement workers. This verifies
the registered-endpoint library lifecycle. It does not add CLI startup support
or establish checkpoint atomicity.

## Remaining distributed work

1. Extend restart/recovery coverage beyond the verified complete-checkpoint,
   two-host Mem/LSB/MSB registered-endpoint cases, including other failure phases.
   Keep the fresh-process remote-FP CLI limitation separate:
   Java recovers before publication/registration while its dynamic manager is
   empty. Do not reorder startup to manufacture support.
2. Cover additional full-model fingerprint failure phases and general network
   partitions. Forced TCP closure/reply loss is covered; it does not prove every
   partition or stalled-connection case. One gated accepted fingerprint lookup
   verifies worker control responsiveness and orderly exit. A controlled
   full-model TCP relay stall is also verified; arbitrary network blackholes and
   other partition topologies remain unproved.
3. Isolate interruption after trace commit and before intern commit. Trace and
   intern owners are concrete. Do not add a production test-only hook, reorder
   commits or use a timing race. Removing `vars.tmp` causes intern commit to
   delete the old checkpoint before failing; that is a different failure boundary.
4. Finish the source ownership/constructor and native cleanup audit where evidence
   identifies actual remaining shortcuts. Consult prior audits before repeating
   completed checks. Missing-component behavior alone cannot prove full parity.
5. Reconcile opaque custom data and evaluator metadata against actual source
   transferability. Preserve explicit rejection until a faithful native contract
   is established. No Java object serialization or reflection runtime is wanted.
6. Keep the four disabled original methods missing until their actual harness
   contracts have a justified translation. Native coverage earns no such credit.

The production checkpoint order is queue begin, trace begin, FP begin, queue
resume, intern begin, queue commit, trace commit, intern commit, FP commit.
Mixed files after interruption are source behavior, not an atomic transaction.
Disk, off-heap and nested-store trace recovery requires the existing trace; it
cannot fall back to a named snapshot when the trace is missing. MemFPSet retains
the source file-based recovery that deliberately ignores the trace parameter.
Intern-table recovery holds its native interning mutex before opening and reading
the checkpoint, through token publication and replay. Linux FIFO checks verify
the blocked-header boundary and lock release on complete/truncated input.

## Testing, commits and documentation

Implement the feature faithfully, then port its original Java tests. Preserve
setup, teardown, assertions, matrices, seeds, fixtures and workload bounds. When
there is no direct original test, the user's later instruction permits focused
native unit checks. Fix production shortcuts before proceeding; never weaken a
ported assertion for a green result.

Run only relevant checks. Never combine long workloads with `-race`; use race
instrumentation only for short concurrency selections. Do not shorten original
bounds. For offline work, use:

```bash
env GOCACHE="$PWD/.codex-gocache" GOTMPDIR="$PWD/.codex-gotmp" \
  GOPROXY=off GOSUMDB=off \
  go test ./tlc -run '^RelevantTest$' -count=1 -timeout=2m
```

Capture long output in ignored `.codex-gotmp/` logs. Only an explicit terminal
exit establishes completion; poll the same live process handle. Real network
checks need the appropriate local-listener/interface permission. Do not weaken
them to accommodate sandbox restrictions. Never delete live compiler outputs,
verification receipts, source or vectors during disk cleanup.

Persistent fixtures belong in `tlc/test_vectors/`, never a directory named
`testdata`. CLI TLC arguments are Java's flags by default; direct arguments and
`modelcheck`/`mc` use the same runner. Do not restore `--tlc` or the bounded CLI.

Commit each green chunk, aiming for every 10–15 minutes. Do not back up, rebase,
stash, amend or push; the user pushes separately. Commits are already authorized.
Keep the workspace clean between chunks.

Update [TODO_TEST_PORT.md](TODO_TEST_PORT.md) manually when a whole original
method becomes complete; retain curated reconciliation and subclass notes.
Put concise audit/run receipts in [PORT_PROGRESS.md](PORT_PROGRESS.md), preserving
its existing line endings. Append with unique EOF context and inspect the diff.
Do not rewrite that file wholesale. Keep this handoff short and current rather
than adding overlapping history blocks. Before resuming, read [PLAN.md](../PLAN.md),
this guide, the progress log, architecture notes and test inventory.
