# TLC Port Handoff

Updated: October 9, 2026. Active branch: `master`.

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
four requiring reconciliation. Their complete assertion bodies are now staged in
`tlc_distributed_java_test.go`, behind `tlago_disabled_distributed_tests`.
Upstream's shared setup unconditionally assumes false for OffHeapDiskFPSet.
The draft runs the original Ant off-heap/512 KiB profile with CPU-derived workers
and native process joins instead of JVM exit interception. On 48 workers,
DieHard, remote EWD840 and TSnapShot pass; local EWD840 fails with GENERAL during
final off-heap CheckFPs. Java's selector likewise retains a previously shut-down
flusher when the new partitions are too small. Do not reset it to hide this
source behavior, weaken the assertion, manufacture a skip or award completion
credit. The opt-in draft records a known failure, not an ordinary green gate.

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
Named publication, discovery, file loading, interning, settings, registration,
manager snapshots and keepalive are wired into the production CLI. Roles are
`server`, `worker`, `fpserver` and `worker-fpserver`. The combined role shares its
listener and waits for both lifetimes. Signal shutdown occurs at the source
shutdown-hook boundary. Accepted replies drain before orderly host shutdown;
forced connection closure remains independent of computation/storage lifetime.
Discovery captures a stable native coordinator endpoint ID. Rebinding a name
affects fresh lookups; existing references retain their original coordinator.
Unbinding removes discovery only; explicit coordinator removal invalidates its
references. Lookup replies now include the ID, so roles must use matching builds.
Concurrent host closes join the same callback/accepted-connection teardown and
retain its cleanup error. A closed admission gate or an empty reply count alone
cannot let a later close return before cleanup. Forced close can still interrupt
a graceful reply drain; graceful callers continue waiting for accepted replies.
Short TCP checks also hold a worker's accepted fingerprint lookup open while
keepalive/cache/exit calls use the same worker connection. Control calls remain
responsive; exit rejects new calls and lets the accepted computation finish.
The full N=7 model also holds TCP relay traffic to one of two FP hosts, with
separate bidirectional, request-only and reply-only cases, without closing
connections. Coordinator status/manager and worker alive/cache probes remain
responsive; routing stays distinct and no failover is
reported. Releasing traffic restores the original 114,942-state result. This
controlled byte-relay stall does not prove arbitrary network blackholes.

Coordinator file requests use a fresh resolver, observing current default
library and user directories and owning separate temporary resource copies.
Explicit search overrides remain captured; worker basename caching is unchanged.
Read diagnostics preserve dot components, including symlink traversal, in the
absolute pathname rather than reporting a lexically collapsed location.
Upstream RMI names describe source references only, not a Go transport requirement.

Transport does not redial or replay ambiguous fingerprint mutations. Manager
failover retains the source algorithm, including its forward reassignment and
slot-based statistics. Workers receive independent manager snapshots with shared
registration-wrapper aliases; the coordinator recovery trace is omitted.
Generated worker, fingerprint and coordinator-owned fingerprint references include
a random native host identity and a per-host sequence. A fresh host reusing an
address cannot receive an old object's delayed call. Stale lazy fingerprint
references fail through the existing manager failover path; replacement storage
remains untouched. Explicit caller-selected endpoint names retain their meaning.
Worker registration also receives a reference before opening its callback
connection. TLC wakes stuck queue consumers before its first GetURI call;
callback refusal must not skip that wake. Native callback owners close unused
references safely, share concurrent first calls and never replay failed calls.

`DistributedOperationError` traits drive retry, removal, discovery and exit
choices. Causes remain available through `errors.Is`/`errors.As`; native sender
frames and error graphs survive payload transfer. Memory exhaustion permits
smaller-batch retry; executor rejection does not. Lost fingerprint exit replies
use an exit-only trait, allowing shutdown to visit later registrations. Prior
closed clients and insertion/checkpoint failures remain reportable. The obsolete
`UnmarshalException` carrier and RMI-specific exit branch are removed. The unused
RMI RemoteException, ServerException, NoSuchObjectException, ConnectException,
ExportException and NotBoundException carriers/class metadata are also removed.
TLC's fingerprint-registration rejection retains its source application name,
message and checked-I/O category through the generic IOException carrier; it
does not require an RMI base class. Generic diagnostic carriers remain intact.

Native connection-owner cleanup ignores already-closed errors only when all
causes are benign. Joined errors that also contain a real callback cleanup
failure remain reportable through discovery and connection owners, preserving
their original context and `errors.Is`/`errors.As` causes.

Named `MultiFPSet` checkpoint and recovery operations run child stores
concurrently and join before returning. Child I/O failures propagate as operation
failures with native Go error wrapping; the manager must not ignore them as remote
server outages. Unnamed begin/commit operations retain source sequential ordering.
Nested fingerprint-distance and invariant checks also run children concurrently
and join started work. Child I/O failures become operation failures with retained
causes; the manager does not treat them as callable I/O fallback results. Signed
minima, empty reductions and invariant short-circuiting retain source behavior.
Nested size statistics also sum children concurrently, preserving long overflow.
`MultiFPSet.GetStatesSeen` returns the parent lookup counter only; child counters
are independent and must not be added to the distributed manager's count.
Nested `AddThread` inherits the source no-op. `IncWorkers` visits children,
but heap disk children also inherit a no-op; off-heap children register with the
shared eviction barrier. Only a direct disk `AddThread` adds a reader. Local/TCP
checks verify successful/failed reader addition and retained pool ownership.
Nested initialization joins its children, ignores returned replacements and
wraps checked I/O once. Unchecked failures retain their native identity; no
Java ForkJoin exception copying is performed.
Public disk-store close releases readers while retaining their closed slots and
resetting the pool cursor. Later disk I/O fails through those closed owners
instead of reopening storage. Internal allocation rollback still clears its
temporary reader arrays; native host cleanup releases all descriptors.
Heap disk invariant checks retain table locks on flush/open failure. Once the
scan opens, successful close releases locks even on invalid/truncated contents;
expected-count comparison follows release. Close failure retains source ownership.
Short Linux syscall checks verify actual close I/O failure overrides scan results
and retains locks; successful-close controls release them.
Memory and nested expected-count invariant overloads inherit the base true
result. Nested no-argument checks visit children; disk overloads enforce counts.
Memory recovery retains complete prefixes and prior membership on truncated or
duplicate input. Local/TCP checks cover all seven partial-long lengths: base
memory runtime failures stop manager recovery, while packed-memory I/O failures
warn and continue healthy registrations. Duplicates stop both without failover.
`MemFPSet1` reads checkpoint primitives directly from the file, matching
`FileUtil.newDFIS`. It must not eagerly buffer ahead of backing-set assignments.
Short syscall checks cover each header/key read failure, partial field/table
mutation, one file close and manager continuation; native I/O causes survive.
Its checkpoint writer also follows unbuffered `FileUtil.newDFOS`. Failed writes
retain the completed temporary-file prefix and old checkpoint without promotion;
native cleanup closes once without retrying writes. Healthy registrations continue.
The buffered `MemFPSet` and `MemFPSet2` writers likewise close the raw owner on
failure, without replaying the failed buffer. Their source flush boundaries and
I/O conversion remain intact; syscall checks verify full-buffer/final-flush and
close failures, retained prefixes and skipped promotion.
Those two stores also use the source's eager 8 KiB input buffer during recovery.
Refill belongs to the current fingerprint read, before insertion; failure must
retain that ordering. Syscall checks cover constructor and both refill failures,
exact membership prefixes, one file close and manager continuation.
Coordinator startup checks retain recovered trace/queue state before fingerprint
failure. Runtime failures prevent publication; packed-memory truncation warns,
continues healthy recovery and prints actual recovery counts before publication.
Memory queue checkpoints open the configured literal directory directly. They
do not create parents, choose temporary directories or clean symlink traversal.
Disk queues now use the same source path contract through one constructor for
public and coordinator use. Checkpoint and spill operations do not create parents.
Background state-pool read/write failures report the source pool diagnostic and
exit the native process with status 1. They cannot silently stop a goroutine and
leave queue waiters alive. Synchronous pool calls still return errors to callers.
Pool reads publish each empty state before decoding its fields, retaining partial
mutation in a failed slot and leaving later slots untouched. Successful reads
propagate close errors; native failure cleanup closes the stream only once.
Disk queue synchronous pool failures become the source coded reading/writing
states runtime assertions. Detail text and prior mutations remain; fatal
categories escape the ordinary catch and no extra cause is attached.
The disk queue cleaner warns on every failed deletion, including missing files,
using canonical paths, and continues through later pools. Canonicalization
failures follow the source error diagnostic and native process exit status 1.
Short TCP cases also cover completed recovery with a lost reply for Mem/LSB/MSB
storage. The manager warns once, continues to the next registration and leaves
routing intact; the broken connection does not replay the completed recovery.
Recovered borrowed storage remains readable through a fresh host.
Named memory, disk and DFID checkpoint methods use an explicitly empty name
literally (`.fp.tmp`/`.fp.chkpt`), without selecting `fpset` or the initialized
name. Empty disk backing names likewise create `.fp`; unnamed checkpoint
methods retain their distinct source behavior.
Empty fingerprint storage directories retain the literal separator prefix;
checkpoint paths are not cleaned or redirected to a temporary directory.
DFID checkpoint creation opens the supplied file directly without creating
missing parent directories. Native checks keep all paths inside temporary roots.
Disk initialization assigns paths before checking a negative worker count and
allocates worker/pool reader slots before opening storage. Reinitialization
retains source membership/index metadata; replaced native reader owners close,
and failed opens retain allocated slots while releasing partial handles.
Concrete memory and disk fingerprint exits now report `TLC_FP_COMPLETED` with
the native local host after optional cleanup. Source-ignored directory-removal
failure does not suppress that diagnostic. Go role ownership retains process
lifetime; accepted TCP exit replies still drain before host closure.

Failed coordinator construction stops and joins its queue workers and closes
owned trace/fingerprint handles, preserving files and the original failure.
Allocation order and base-constructor error precedence remain unchanged;
successful construction transfers resources to the coordinator.

The native worker/FP network owner retains local fingerprint storage through
unpublication and registration errors. Host shutdown drains accepted replies,
then closes each owned storage object once without invoking Exit or deleting
files. Direct `Host` publications remain caller-owned; worker runtime shutdown
is still separate.
Graceful transport shutdown expires socket reads before draining, so a header
whose body never arrives cannot hold shutdown open. Accepted handlers retain
their response writes, including Exit beside an incomplete request on one TCP
connection. Ordinary operation deadlines and retry policy are unchanged.
Before registration begins, FP startup failures close allocated storage handles
without deleting files or replacing the failure. Once registration starts,
startup/reporting errors retain storage for the native host or caller to release.

State/result payloads retain identity and nil/empty distinctions across states,
values, strings, predecessor graphs, caches, byte buffers and partition vectors.
Cached nil-supplier `LazySupplierValue` objects retain their supplier type and
cache graph. Nil-supplier evaluation fails even with a cached value. Executable
Go suppliers are explicitly rejected at transfer without invocation, rather
than silently becoming ordinary lazy values. Uncached wrappers retain the
existing lazy-transfer failure before supplier checks.

Attached model-value data supports native `[]Value` through the same array graph,
preserving cycles, shared backing storage and typed nil/empty arrays across RPC.
Scalar character attachments now use `uint16`, retaining every UTF-16 code unit
and distinct map-key type. Out-of-range payloads fail. Peers carrying this new
scalar tag need updated native builds; opaque evaluator objects remain rejected.

Primitive attachments also support `[]bool`, signed integer slices, `[]uint16`
and both floating-point slice widths. Native array references preserve sharing,
separate equal arrays, typed nil/empty slices and receiver isolation. Integer bit
payloads retain signed zero and NaN bits; invalid kinds, references and narrow
representations fail explicitly. These finite Go types require updated peers.
Native `[]string` attachments also retain shared slice storage, separate equal
arrays, typed nil/empty slices and exact string bytes across mixed graphs and
worker requests/results. They use a separate typed array table with validated
references and isolated receiver ownership.
Attached `*TLCStateMut` objects reuse root/predecessor state identities, including
attached-only states, cache/value back-references and native map keys. Decoder
states are allocated before attachments are resolved. Typed nil references and
receiver isolation survive; attached states retain the existing metadata checks.
Attached `[]*TLCStateMut` slices preserve backing storage and can alias the
invocation's root slice. A state-array table reserves identities before following
cycles; ordinary unaliased roots retain their existing inline representation.
Typed nil/empty arrays, null entries and independent equal arrays survive.
Native `map[string]Value` attachments also preserve shared map identity, cycles,
nil entries and typed nil/empty maps in requests, results and failure contexts.
Native `map[any]any` attachments additionally retain supported scalar and value
keys, their concrete Go types and shared references. Non-comparable decoded keys,
duplicate keys and unsupported entries fail explicitly. General keys are not
evaluated or fingerprinted to sort them. Peers need the updated payload schema.
Attached `*UniqueString` values and keys share the existing name graph with
value fields. Text, token, location, identity and typed nil references survive;
equal distinct names are not re-interned or merged.

Attached `[]*UniqueString` arrays reuse the same native name-array table as
record and record-set fields. Direct/TCP/error-context checks preserve shared
backing storage, distinct equal arrays, nil/empty arrays and nil entries, with
isolated receiver ownership. Invalid array/name references fail explicitly.

Finite `OpRcdValue` domain-row containers and result arrays retain constructor
sharing with other operators and model-data attachments. Native `[][]Value`
attachments use that row graph; receiver updates affect shared operator
application without reaching the sender. Invalid row references fail explicitly.
Finite operator evaluation rejects nil argument arrays/rows, and initialization
rejects nil rows, arguments and results. Real empty argument rows remain valid;
null inputs must not become zero-argument matches or silently initialized values.
Mixed `[]any` and `map[string]any` attachments retain supported scalars, bytes,
TLC values, typed containers and recursive mixed-container graphs with the same
ownership guarantees. Their compact entries share the scalar data tags; peers
must use the updated native payload schema.
Attached `*LongVec` objects now share the state graph's vector table with result
fingerprints. Active elements and object aliases survive; spare capacity and
backing-array aliases do not. Null/empty vectors and map-key identity are retained.
Attached state vectors likewise share their table with result partitions,
preserving vector/state cycles, null entries, active capacity and receiver
ownership. Decoded vectors retain TLCStateVec's unbounded collection policy.
Typed vector-array attachments also share the result's partition arrays. Native
array IDs preserve nonempty aliases and cycles, null/empty arrays, null entries
and receiver isolation. Result payloads now use graph array IDs directly.
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

Thread construction retains its supplied proxy/selector and schedules keepalive;
it does not register itself. `RegisterWorker` inserts the thread after
construction and before Start, preserving both URI callback boundaries.
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

Server-thread sent/received counters and final cache ratios use a native mutex
for concurrent getter access. Counter updates retain signed 32-bit overflow;
cache values retain NaN and the failure sentinel. Initialize fields before Run
and use getters during execution. The short concurrency check adds no original
method credit; no worker/network operation holds this statistics mutex.

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

The full-model shared-process-loss row gives two distinct worker endpoints on
one listener real nonempty RPC blocks, then kills their process. It requires
two loss/deregistration events and the source's single deduplicated cache warning
before checking that the coordinator still reports unfinished work. A fresh
worker finishes the unchanged model with 114,942 distinct states and an empty
queue. The gate is in the test endpoint, not production evaluation or cleanup.

The full model also kills the first of two fingerprint hosts with assigned work
and resumes actual evaluation. Worker and coordinator fail over independently.
Two partition slots alias the survivor, so source statistics report 229,884 in
this separate failure row; ordinary original rows still require 114,942.
Another row kills that host after a real successor `putBlock` has inserted new
fingerprints but before its reply returns. Stored membership is verified before
the kill. Source callable failover retries against the survivor and completes
with the same 229,884 slot-based count and empty queue. Only this row requires
one coordinator EOF diagnostic for the deliberately lost insertion reply.
The insertion-reply-loss case also covers two-child LSB/MSB hosts. Before killing
the first host, it flushes both child files and verifies stored membership; both
hosts must report the configured disk implementation. Native host memory budgets
do not change the original N=7 model bounds or source failover assertions.

A separate full-model row kills the host after a nonempty `containsBlock` lookup
completes but before its answer returns. The worker receives the single expected
EOF diagnostic and retries through source callable reassignment; the coordinator
independently reassigns its publication endpoint. Both retain the surviving host,
and the run finishes with 229,884 slot-counted distinct states and an empty queue.
This covers lookup reply loss, not arbitrary partitions, and adds no original
method completion credit.
LSB/MSB lookup-loss rows flush both children before lookup and require actual
disk fingerprint reads before holding the reply. Both hosts retain their source
two-child factory layout; final failover/count/diagnostic assertions are unchanged.

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

The remote restart matrix also loses the first host's reply after its real named
checkpoint commit. The coordinator must print the source warning and retain two
distinct available registrations. Fresh processes restore the committed files
and exact partition membership before resuming the original N=7 model. Mem,
LSB and MSB rows cover both one worker and two workers sharing an application
in each generation. The two-worker cases retain actual work and separate
statistics from both replacement workers. This covers completed-commit reply
loss, not incomplete remote commits or atomicity.

Short native checks also lose the reply after a real named begin. They require
one begin, no commit or replay, unchanged registrations and exact pending file
contents. Fresh storage leaves those files unpromoted. Missing committed Mem
snapshots print the source warning and let the manager recover the healthy host;
nested disk child I/O failures propagate and stop before the later host. This
storage-specific behavior must not be replaced by an atomic recovery assumption.

Fresh-process LSB/MSB checks also remove one committed child snapshot from the
first host after a real mid-run checkpoint and role crashes. Recovery reports
the missing child before publication, joins the sibling's native recovery and
does not recover the later host. Source main prints one GENERAL and closes with
cleanup disabled; its caught failure does not change process status. Retained
coordinator/remote checkpoint bytes survive, and the missing child stays absent.

Accepted-request connection-loss checks cover begin, commit and recovery with
memory and two-child LSB/MSB stores. The manager continues to the healthy host
before releasing the old handler, without replay or reassignment. Accepted work
then finishes with exact child snapshot bytes/membership; committed snapshots
also reopen in fresh stores. This covers transport loss, not host death.

## Remaining distributed work

1. Extend restart/recovery coverage to other failure phases. Verified cases use
   two-host Mem/LSB/MSB registered endpoints and cover complete checkpoints,
   completed-commit reply loss and missing committed LSB/MSB child snapshots.
   Keep the fresh-process remote-FP CLI limitation separate:
   Java recovers before publication/registration while its dynamic manager is
   empty. Do not reorder startup to manufacture support.
2. Cover additional full-model fingerprint failure phases and general network
   partitions. Forced TCP closure/reply loss is covered; it does not prove every
   partition or stalled-connection case. One gated accepted fingerprint lookup
   verifies worker control responsiveness and orderly exit. A controlled
   full-model TCP relay stall covers both directions together and each direction
   independently; arbitrary network blackholes and other partition topologies
   remain unproved.
3. Exact interruption between the trace and intern method calls remains unproved.
   Short local Linux syscall checks now kill before intern-file mutation, inside
   intern commit at deletion entry. Queue/trace promote their new generations;
   intern/fingerprint snapshots retain the old generations and pending files.
   This requires `strace` and does not establish full-model/remote recovery at
   that boundary. Do not add a production hook, reorder commits or use a timing
   race. Removing `vars.tmp` deletes the old checkpoint before failing and tests
   a different boundary.
4. Finish the source ownership/constructor and native cleanup audit where evidence
   identifies actual remaining shortcuts. Consult prior audits before repeating
   completed checks. Missing-component behavior alone cannot prove full parity.
   Built-in selector arithmetic, startup capture, queue bounds, statistics and
   smart-proxy timing have a current source audit and focused green receipt.
   Custom factory selection now uses linked Go constructors registered through
   `RegisterBlockSelectorFactory`, with the original startup property. The
   coordinator and threads retain the returned `BlockSelection` policy. Keep
   source fallback/panic boundaries; do not add a JVM loader.
   The application constructor attaches its original tool, application and
   flags before calling the factory. Failed callbacks retain native resource
   rollback; subclass registration setup still follows base factory creation.
   Raw byte-queue paths, partial reads and pool failure catches now match the
   source. Background reader/writer failures and cleaner exceptions print the
   pool error and exit. Ordinary failed deletions warn with canonical paths and
   continue, including missing files. Synchronous failures raise the coded queue
   assertion without advancing work.
   Failed raw writes close native files without reflushing or publishing the
   checkpoint marker. Original writer wake/finish and stream tests remain green.
5. Reconcile opaque custom data and evaluator metadata against actual source
   transferability. Preserve explicit rejection until a faithful native contract
   is established. No Java object serialization or reflection runtime is wanted.
6. Reconcile the four opt-in original model bodies with the upstream disabled
   harness and its retained-flusher failure. The source-profile failure is now
   located; staged bodies and native coverage earn no completion credit.

Bulk queues publish their logical length after the whole enqueue loop. Failed
disk spills retain the inserted prefix without counting it. Deque storage tracks
occupancy independently. The source memory queue's bulk slot-overwrite quirk is
preserved; distributed retry fixtures use the actual disk queue.
All four queues reject nil array/vector batches before mutation using an ordinary
Go error. Explicit empty batches remain valid; rejection releases queue locks
and preserves existing work. Original queue tests and native TCP retry/loss
checks pass without changing their assertions.

Raw byte-queue checkpoint/pool files retain literal symlink traversal and never
create missing parents. Recovery preserves inactive slots, publishes allocations
before reading, and retains the source's zero-padded short final entry. Empty
configured directories retain their source paths rather than selecting `/tmp`.
Byte queue reads extract raw entries under the queue lock and decode afterward.
A decode failure retains the complete batch removal and published length.
Null raw entries retain the source peek/dequeue/assertion boundaries; oversized
bulk requests retain the source wrapper's post-removal access failure.
Byte-state conversion rejects nil input rather than allocating an empty state.
Array conversion finishes before queue publication, so a failed conversion
cannot enqueue its valid prefix, including in zero-variable models.

The production checkpoint order is queue begin, trace begin, FP begin, queue
resume, intern begin, queue commit, trace commit, intern commit, FP commit.
Mixed files after interruption are source behavior, not an atomic transaction.
Trace, worker, state queues, intern table and memory fingerprint stores follow
the source existence/delete/rename order using native filesystem calls. A
dangling old checkpoint link survives failed promotion; a live link is removed
first, without changing its target. Short cases cover those eight owners and
five supporting integer/object/byte-array/DFID stores. Their commit errors retain
I/O classification and source text; byte-array pool deletion keeps partial
progress on failure without advancing checkpoint markers.
Disk, off-heap and nested-store trace recovery requires the existing trace; it
cannot fall back to a named snapshot when the trace is missing. MemFPSet retains
the source file-based recovery that deliberately ignores the trace parameter.
All three memory stores now propagate a final close failure after successful
recovery, retaining reconstructed membership. Native failure cleanup closes once
without replacing an earlier read/reconstruction error. Sixteen Linux syscall
checks verify normal/faulted close, manager warnings and healthy continuation.
Intern-table recovery holds its native interning mutex before opening and reading
the checkpoint, through token publication and replay. Linux FIFO checks verify
the blocked-header boundary and lock release on complete/truncated input.

Named disk snapshot startup now has 40 short LSB/MSB, direct/nested, local/TCP
cases. Direct missing-file I/O warns and continues; nested I/O stops startup
after both children join. Empty, duplicate and descending snapshots retain
partial disk count/index/write mutations and stop before the next registration
or publication. Trailing partial records after two complete fingerprints follow
the source EOF catch and complete recovery. These checks add native coverage,
not original-method credit or full-model restart guarantees.

Off-heap batch operations now reject nil vectors before storage access, matching
the inherited source FPSet contract. Local/TCP checks distinguish nil from empty
batches, retain membership/seen counts and verify normal calls after rejection.

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
