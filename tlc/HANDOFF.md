# TLC Port Handoff

Updated: October 9, 2026. Active branch: `master`.

The user authorized fixing the upstream off-heap flusher lifecycle bug in Go.
The fix and upstream report are in [JAVA_BUG_FOUND.md](../JAVA_BUG_FOUND.md).
The correction covers both final-check selection and direct invariant flushing
after an executor shuts down. Focused regressions and the formerly failing
original-profile EWD840 diagnostic pass. The original off-heap multiple-flush
test passes behind `tlc_fp_stress`, retaining all four default rounds of
8,388,608 insertions and exact invariant counts. This authorization does not
cover the separate checkpoint-error formatting hang. Four original model methods remain
Reconcile; no skipped body is counted as passing. Continue concrete port gaps
without manufacturing fault scenarios.

This is the current restart guide. Detailed audit history and verification
receipts belong in [PORT_PROGRESS.md](PORT_PROGRESS.md); implementation contracts
belong in [TLC_ARCH.md](TLC_ARCH.md). Older handoffs remain in Git history and
must not be read as current assignments.
Use [DISTRIBUTED_PORT_MAP.md](DISTRIBUTED_PORT_MAP.md) to locate the native
implementation of each of the 35 upstream distributed source files. Mapping
coverage is separate from behavioral verification and original-test credit.

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

Core bridge continuation: runtime LET traversal imports SANY Context state,
including Pair history and Hashtable topology. Canonical adapters now retain
ASSUME/PROVE, theorem and proof graphs, source module ownership, label formals
and OpDef step links. Module views share the original SANY base and syntax range.
Separate assumption/theorem vectors preserve constant processing; generated
SANY top-level graphs still omit some AST-only INSTANCE expressions, including
original model 219's line 42 assumption. Keep those runtime assumptions and
assertions while completing source generation. Complete visitor callbacks,
mutation/null boundaries, builtin adapters and native INSTANCE lowering remain
pending; see `TLC_ARCH.md` for the exact boundaries.

## Verification baseline and test credit

Active full workload: `TestJavaOffHeapDiskFPSetLong_testMaxFPSetSizeRnd`, with
`-tags=tlc_fp_stress -timeout=0`, is running in native exec session `77502`.
Its log is `.codex-gotmp/offheap-random-full.log`; compiler-only receipt is
`.codex-gotmp/offheap-random-full-compile.log`. Poll that same handle before
starting another copy. It preserves all 2,147,483,648 iterations and original
assertions, with the default 64 MiB direct-memory budget and no race
instrumentation. Source translation is committed, but full execution and
original-method credit remain pending. Temporary storage is under
`.codex-gotmp/`. Do not remove live files or infer completion from progress lines.

The user supplied a green full-suite baseline. Do not rerun that approximately
45-minute suite. The earlier recorded full-workspace run verified `23f046e`:
root 1,552.199 seconds, TLC 771.039 seconds, SANY 2.419 seconds, status 0.
Later distributed changes have focused receipts; they do not establish a new
full-workspace pass. Reuse verification for unchanged code.

The distributed inventory records 41 original methods: 37 port complete and
four requiring reconciliation. `tlc_distributed_java_test.go` preserves the
source shared setup's unconditional false assumption with `t.Skip`, using its
exact OffHeapDiskFPSet reason before any roles start. Every assertion remains
in its shared body helper. `tlc_distributed_java_diagnostic_test.go`, behind
`tlago_disabled_distributed_tests`, exposes `TestDiagnosticJava...` entries that
deliberately bypass that assumption. A source skip is not an executed body pass;
these separate diagnostics earn no original-method completion credit.

The diagnostic runs the original Ant off-heap/512 KiB profile with CPU-derived
workers and native process joins instead of JVM exit interception. On 48 workers,
DieHard, remote EWD840 and TSnapShot passed before this fix. Local EWD840 exposed
Java's reuse of a shut-down flusher when final-check partitions become too small.
The explicitly authorized Go correction selects the existing sequential path.
The unchanged local EWD840 diagnostic now passes in 74.80 seconds, including
no GENERAL and normal process exits. See [JAVA_BUG_FOUND.md](../JAVA_BUG_FOUND.md)
for the source analysis and regression. Source-skipped entries remain separate.

| Original class | Pending method |
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
Production roles are `server`, `worker`, `fpserver` and `worker-fpserver`; the
combined role shares a listener and waits for both lifetimes. Discovery,
publication, file loading, interning, settings, registration, manager snapshots
and keepalive are wired into the CLI. Upstream RMI names identify source files
only. They are not Go implementation requirements. Native network shutdown errors
reach CLI stderr and a failure exit status, including after a successful TLC
command body; earlier command diagnostics remain visible. Native discovery accepts
bare or bracketed IPv6 hosts and escapes zone suffixes for URLs, restoring them
for TCP dialing. Coordinator and callback advertisements normalize IP brackets,
retain scoped hosts and actual bound ports, and reject all wildcard spellings.
Listener bind hosts also accept bare or bracketed IPv6.
Native worker construction/publication formats validated TCP addresses with
Go URLs, carrying that address through the registration runnable. Scoped
interface names with hyphens are supported and zones are URL-escaped. The
existing source-construction URI fixtures remain intact. IPv6 loopback
discovery/status and worker callbacks have focused coverage;
scoped link-local routing across real interfaces remains outside that coverage.

Use [DISTRIBUTED_PORT_MAP.md](DISTRIBUTED_PORT_MAP.md) to find implementation
files, and [the distributed architecture](TLC_ARCH.md#distributed-tlc-architecture)
for detailed contracts. Preserve these boundaries during further work:

| Area | Required behavior |
| --- | --- |
| Transport | Failed calls retain the failed client and original cause; no automatic redial or replay. Remote method errors leave the connection usable. Failed response flushes close the connection; failed accepts release the listener while existing connections remain independent. |
| Publication | Discovery captures stable endpoint identity. Rebinding affects new lookups; unbinding removes discovery only. Generated references include a host incarnation and sequence so address reuse cannot receive stale calls. Matching peer builds are required. |
| Ownership | Concurrent closes join teardown and retain mixed cleanup failures. Worker/FP endpoint closes join their explicit client release and retain its result for callback owners. Coordinator views share one owner for primary-client release followed by callbacks, retaining both results for discovery shutdown; callbacks rejected after shutdown return any real release failure. Discovery closes join cached coordinator release and retain its result for repeated callers. Accepted connections close once, retaining errors after tracking removal. Orderly shutdown drains accepted replies; forced close can interrupt transport independently of accepted computation/storage. Incomplete request bodies cannot block shutdown. Host-owned storage closes once after draining, without Exit or file deletion; direct publications remain caller-owned. |
| Registration | Lazy callback references precede connection opening. Wake stuck consumers before the first worker URI callback. Thread construction precedes insertion, then Start; startup increments workers before queue capture. |
| Worker loss | Cancel keepalive, claim cleanup once, remove registration, requeue, clear assigned states, wake consumers and decrement workers in source order. Convert an absent assigned block to an explicit empty batch. Preserve queue suspension and the ten-second/sixty-second keepalive schedule. |
| Failure categories | Native operation traits drive retry/removal/exit. Retain causes and sender frames. Memory exhaustion permits smaller batches; executor rejection does not. Ignore closed errors only when every cause is benign. Exit-only lost replies must not suppress earlier insertion/checkpoint failures. |
| Fingerprint routing | Preserve forward reassignment, aliased wrappers and slot-based statistics. Worker snapshots omit the coordinator trace. A local MultiFPSet remains one endpoint with high-bit child routing; manager server routing uses low bits. |
| Nested storage | Named checkpoint/recovery, size and checks join concurrent children. Child I/O becomes an operation failure, outside manager checked-I/O warning/fallback. Unnamed checkpoint operations remain sequential. Parent states-seen counters exclude child counts. |
| Storage lifetime | Retain source partial mutations, reader slots, lock ownership and failure precedence. Roll back newly owned native descriptors without reopening closed stores, retrying writes or promoting pending snapshots. Constructors retain source allocation/error order. |
| Paths and files | Preserve literal empty names/directories, dot components and symlink traversal. Queue/trace/FP/DFID operations must not invent parents or temporary fallback directories. File requests use fresh default resolvers; explicit overrides and worker basename caching remain captured. |
| Queues | Publish bulk logical length only after the loop; failed spills retain uncounted prefixes. Reject nil batches before mutation; empty batches remain valid. Byte decoding follows removal, while array conversion finishes before publication. Preserve partial pool reads and source diagnostics; background pool/cleaner exceptions exit the process rather than silently killing a goroutine. |
| State ownership | Ordinary states retain base no-op action/cache/callable accessors. Extended copies reapply predecessor setters and depth limits. Print wrappers own a separate mutable state and independent metadata; their record conversion returns the existing record. Disk traces store RAF links/fingerprints without retaining an in-memory state/action mirror. Negative trace seeks fail after pointer mutation. |
| Evaluation | Worker checks precede constraints. TLCApp retains captured action/property arrays, vector concatenation, separate returned-array ownership and predecessor policy. Replay uses its existing evaluator. Coordinator publication inserts fingerprints before trace and queue writes; missing owners cannot fabricate success or zero statistics. |

Native payloads preserve supported graph identity, cycles, nil/empty distinctions,
concrete scalar/container types and isolated receiver ownership. State, value,
name, buffer, vector, map and finite-operator tables have validated references.
LongVec transfers active entries rather than spare capacity; BitVector retains
all words. Invalid finite-operator nil arguments/rows/results fail, while actual
empty argument rows remain valid. Cached nil-supplier wrappers retain their type
and graph but still fail evaluation; executable suppliers are rejected without
invocation. Symbolic materialization follows the source.

Core metadata producers are audited: neither core TLC nor its tests calls
`Value.setData` or `ModelValue.setData`. Ordinary distributed model-checking
states store no action; populated extended actions retain evaluator objects that
cannot be transferred. Unsupported opaque/custom data remains an explicit error.
Extend transfer only for an actual producer and a native contract, not arbitrary
Java objects. See the metadata producer audit in `TLC_ARCH.md`.

Collision reporting interprets fingerprint-distance bits as a signed long and
rounds the exact reciprocal to two decimal significant digits, half up, before
floating-point conversion. Preserve local `-1` and MinInt64 distances, the
nonempty zero-distance failure and the source empty-model bypass.
The local fingerprint-manager constructor retains a missing reference until
use; counts and empty batches remain valid. The ordinary distributed constructor
still rejects it immediately, as its source evaluates the reference's string.
Distributed manager diagnostics use the shared native local-host lookup and
report lookup failures before falling back to `Unknown`. Local-manager hostnames
remain captured from their storage owner.

## Checkpoint and recovery contracts

Preserve this production order:

1. Queue begin, trace begin, fingerprint begin.
2. Resume queue, intern begin.
3. Queue commit, trace commit, intern commit, fingerprint commit.

Distributed named fingerprint checkpoints commit during the fingerprint-begin
phase; their final manager commit is a no-op. Local fingerprints commit in the
final phase. Mixed generations after interruption are source behavior. Do not
turn these checkpoints into an atomic transaction or promote pending files.
Application creation restores intern tokens before parsing and server construction;
a missing committed file or truncated header fails before trace/queue/FP recovery
or publication. A complete pending intern file is retained without promotion. Incomplete
intern records retain the recovered counter and complete prefix, then throw
checkpoint corruption. Their null error parameter exposes a reproduced upstream
GENERAL-formatting loop; do not count this as green process recovery. See the
architecture notes on incomplete intern records. Server
recovery runs trace, queue, then fingerprints. Intern recovery holds its mutex
from before open through token publication/replay.

Direct checked-I/O failures warn and continue healthy registrations. Nested
child I/O propagates after child joins and stops later registrations/publication.
`MemFPSet` recovery ignores its trace parameter; disk/off-heap/nested trace
recovery requires the actual trace. Preserve partial recovered membership, file mutation,
close-error precedence and source existence/delete/rename order, including old
checkpoint symlinks. Runtime recovery failures prevent worker publication.
Native TCP permits a nil trace argument and delegates it to storage: direct Mem
recovers its own named checkpoint, while replay stores retain their null failure.
Non-null traces remain coordinator-owned and cannot cross the transport.

Fresh CLI remote recovery still has the source limitation: recovery precedes
publication/registration, so the dynamic manager is empty. Registered-endpoint
library restart coverage does not establish CLI support. Do not reorder startup
to manufacture it.

## Verified process coverage

These are focused receipts, not proof of complete distributed parity. Details,
terminal statuses and log paths are in [PORT_PROGRESS.md](PORT_PROGRESS.md).
Reuse unchanged receipts rather than repeating long models.

| Scenario | Verified scope |
| --- | --- |
| Ordinary model roles | EWD840/MC06 N=7 coordinator-owned, standalone, partitioned and combined fingerprint roles: 114,942 distinct states, empty queue, FINISHED, no GENERAL. Native DieHard/TSnapShot retain their error/trace assertions. |
| Multiple workers | Shared application/runtime, distinct endpoints on one listener, actual work and statistics per worker. Repeated registration retains separate coordinator threads and shared worker identity, including the second exit warning. |
| Worker successor checks | Local/TCP cases verify fingerprint lookup before invariants, invariants before constraints, model/action short-circuiting and exact failure context. Seen states skip checks; only accepted states receive predecessor UIDs. The worker does not insert fingerprints. |
| Worker failure | Sole-worker, survivor/replacement, fully computed reply loss and shared-process loss with two assigned endpoints. Source retry, deregistration, warning and unfinished-work checks precede replacement completion. |
| Fingerprint failure | Host death, complete/partial insertion reply loss and lookup reply loss for Mem/LSB/MSB. Disk cases use two actual children and flushed/read storage. Aliased survivor slots report 229,884 while the survivor stores 114,942 fingerprints. Ordinary rows still require 114,942. |
| Final fingerprint check | Full native N=7 with a live remote Mem host returning or raising checked I/O failure: one GENERAL, source Long.MAX_VALUE distance fallback, success summary, 114,942 distinct/zero queued and normal cleanup. The check is not replayed and the live host is not reassigned. |
| Final fingerprint reply loss | Sole remote Mem host completes its real distance check over 114,942 entries, then dies before replying. One GENERAL precedes final states-seen exhaustion warnings and source fallback success reporting. Captured final counts remain 114,942 distinct/zero queued; surviving roles join normally. A two-Mem-host case also preserves 114,942 distinct/zero queued, uses the survivor's real distance and states-seen count, skips retrying the failed statistics slot, and exits the aliased survivor once. The same full-model survivor case also passes with two nested LSB or MSB children on each host; actual final flushes are checked. An off-heap survivor case verifies actual evicted file membership plus retained memory entries, matching the source subset check without forcing a final flush. General network partitions remain unproved. |
| Controlled stalls | Bidirectional, request-only and reply-only TCP relay stalls retain distinct routing and responsive status/control calls; release restores the ordinary model result. A gated accepted lookup also preserves worker control responsiveness. Arbitrary blackholes/topologies remain unproved. |
| Coordinator timer | With a real computed worker reply held, callback removal is detected by the unchanged ten-second timer over TCP. Assigned FIFO work is requeued once and the timer joins. Later transport loss does not repeat cleanup or shrink the block; the late reply is drained without publication. Statistics and the final cache warning retain source behavior. This is a native scheduler fixture, not a full-process partition. |
| Worker activity policy | During an actual TCP computation blocked in fingerprint lookup, synchronous keepalive invocation makes no discovery call after loss of an already used coordinator. The actual invocation-start timestamp also suppresses discovery after this short computation; completion does not refresh it. Computation returns intact and the executor remains open. This verifies task policy, not an additional scheduler period or full-process partition. |
| Idle worker timer | The actual ten-second scheduler detects coordinator loss through an already used discovery connection, exits two published idle workers, shuts down their shared executor and releases their latch. The timer goroutine joins and both callbacks reject further calls. This is not an active-computation or full-process partition test. |
| Local restart | Complete initial frontier, mid-run checkpoints, two-worker restart and interruptions before queue commit, after queue commit, after intern commit and after the first nested FP commit. Files/counts are inspected. |
| Remote restart | Fresh coordinator, workers and two Mem/LSB/MSB hosts recover exact committed partition membership/frontier, then complete N=7. One- and two-worker generations cover complete checkpoints and completed-commit reply loss. |
| Remote off-heap checkpoints | Named begin/commit/recover retain Java's warning-only no-ops. An actual two-child off-heap host creates no snapshots and restores no membership; a later remote Mem host still commits/restores normally, without failover. This limitation differs from coordinator-local trace replay. |
| Recovery reply loss | Real recovery completes before acknowledgement loss. Warning/continuation precede source size-query reassignment; no transport replay. Fresh inspection verifies membership, then survivor evaluation completes with source slot-counted statistics. |
| Begin reply loss | Real begin leaves pending files without commit. Direct Mem warns, restores the healthy host and completes evaluation. Nested LSB/MSB stop before the healthy host/publication and preserve both pending children. |
| Missing/corrupt snapshots | Missing Mem committed files warn/continue; truncated/duplicate direct Mem records stop before the next host or publication, retaining partial inserts and checkpoint bytes. A failure at the second Mem host also retains the first host's complete reconstruction. Missing, empty or out-of-order nested disk snapshots stop after sibling joins. Ordering failures retain partial writes and the source full-file counter. Trailing partial records preserve complete membership and permit full-model completion. Truncated coordinator queue stops before FP recovery; truncated trace plus missing queue reports trace EOF first. Retained bytes and process joins are checked. |
| Late commit failures | Missing intern temporary stops final local FP commit, retaining its old snapshot and new pending file; remote FP has already committed. Missing local FP temporary fails after queue, trace and intern promotion. Failed rename retains source deletion of the older destination. These are short real-file checks, including a TCP FP endpoint, not full-model restart proof. |
| Accepted request loss | Short begin/commit/recovery cases let accepted work finish after connection loss, with exact files/membership and no replay/reassignment. This is not host-death coverage. |
| Trace-to-intern boundary | External GDB/disassembly verifies local and full remote N=7 interruption after trace commit and before intern commit. Queue/trace and intern generations differ; remote FP is new, local FP old. Remote recovery completes. Opt-in Linux/amd64 fixture requires ptrace, GDB and an optimized symbol-bearing `go test -c` binary. |

Short storage/startup checks also cover direct/nested local/TCP malformed named
snapshots, all partial-long lengths, native syscall read/write/close failures,
source buffer boundaries, literal paths, off-heap nil batches and queue/intern
ownership. Those checks add no original-method credit and do not prove full-model
restart or checkpoint atomicity. See architecture notes for their exact limits.

## Remaining distributed work

The source inventory is fully mapped, but mapping is not proof of parity. The
four model reconciliations have translated assertion bodies and preserve Java's
shared unconditional skip; they are not four absent implementations. The
upstream off-heap flusher defect has an explicitly authorized Go fix. The separate
intern-error formatting hang still requires a decision before diverging.
Further failure-phase and partition checks below are verification gaps, not
identified missing source features. Choose concrete source behavior and consult
existing receipts before adding scenarios; do not turn this into an unbounded
fault-testing project. Original-method credit remains unchanged.

1. Extend recovery coverage to additional failure phases beyond the matrix above.
   Intern-header, trace and queue startup failure ordering is verified; other
   phases remain open. The reproduced intern-record GENERAL-formatting hang is
   also an upstream limitation requiring an explicit decision before diverging.
   Keep the CLI empty-manager limitation separate from
   registered-endpoint library recovery. Do not claim checkpoint atomicity from
   successful restarts.
2. Cover additional full-model fingerprint failure phases and general network
   partitions beyond forced closure, reply loss and the controlled relay stalls.
3. Continue source ownership/constructor and native cleanup comparisons only
   where concrete evidence identifies remaining shortcuts. Consult prior audits
   before repeating completed checks; missing-component tests alone cannot prove
   parity. Linked Go selector factories already preserve startup capture,
   fallback/panic boundaries and application attachment order; no JVM loader is
   required.
4. Keep unsupported metadata explicit. Extend the completed core producer audit
   only when a concrete additional producer and native transfer contract warrant
   it. Do not invent action/cache data on ordinary states to expand codec scope.
5. Continue reconciling the four model bodies and native harness adaptation.
   The flusher defect is fixed and the formerly failing diagnostic passes.
   The exact source assumption remains preserved; skipped entries and opt-in
   diagnostics do not earn original-method completion credit.

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
Distributed role help includes local/remote storage launch sequences, the
original backend selector property, storage-owner option placement and native
heap/direct-memory settings. Default coordinator storage rejects remote FP
registrations; the remote example sets `expectedFPSetCount=1`. Remote off-heap
named-checkpoint limitations remain explicit.
Both checker and server `-fpmem` deprecation warnings give Go memory-budget
guidance (`GOMEMLIMIT`), preserving the source option's acceptance and allocation.

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
