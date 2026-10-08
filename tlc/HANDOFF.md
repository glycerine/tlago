# TLC Port Handoff

Updated: October 8, 2026. Full-workspace verification baseline: `23f046e`.
The user subsequently supplied a green full-suite baseline. Current distributed
changes have focused verification; do not rerun the full suite.

This is the current restart guide for the Go TLC port. Detailed audit history
and run receipts belong in [PORT_PROGRESS.md](PORT_PROGRESS.md). Older versions
of this handoff remain in Git history; their pending jobs and next steps must
not be mistaken for current work.

## Goal and scope

User clarification (October 8): distributed TLC uses native Go networking and
concurrency. Do not implement RMI or pretend to provide a Java runtime.
Distributed retry, remote-failure, worker-availability and exit decisions now use
`DistributedOperationError` traits exclusively. Java remote-exception hierarchy
and nested-cause inference are removed from those decisions. Native endpoint
errors capture Go frames where they originate; payload decoding retains sender
stacks. Generic source diagnostic carriers remain separate from transport logic.

Memory exhaustion permits smaller-batch retry; executor rejection does not.
Shutdown and keepalive tolerate prior worker removal, continue to later workers
and avoid decrementing completion again. Coordinator missing bindings retain
discovery retry and suppress shutdown-hook traversal. Connection closure remains
distinct from endpoint removal. Source algorithm, checkpoint, trace ownership
and the four assumption-disabled original remote harnesses still require work.

Fingerprint transport failures use `DistributedOperationError` and preserve Go
causes for `errors.Is`/`errors.As`, replacing the flattened endpoint error. They
retain manager I/O/failover handling without acquiring worker smaller-batch retry
or exit traits. Native checks cover closed clients and a completed insertion whose
reply is lost: storage remains mutated, and the client does not replay or redial.
Standalone fingerprint fixtures now supply distributed coordinator registration
state, preserving the base coordinator's rejection behavior.

Faithfully port Java TLC to Go, then translate its existing correctness tests.
When a translated test fails, inspect both the translation and the production
implementation. Fix implementation shortcuts before proceeding. Preserve the
original assertions, constructor settings, setup, teardown, seeds, parameter
matrices, fixtures, and workload bounds. Do not invent replacement tests or
weaken tests to obtain a passing result.

The Java reference checkout is `../tlaplus`, pinned to
`8f4bc8b73ad1202774a6bf70143436f8ba50aab0`. TLC sources and tests are under
`../tlaplus/tlatools/org.lamport.tlatools/`. The Go module is
`github.com/glycerine/tlago`; most TLC implementation lives in package `tlc`,
with parser integration and model tests in the repository's root package.

The SANY audit goal is complete: every defect in
[SANY_TESTS_TO_FIX.md](../SANY_TESTS_TO_FIX.md) is repaired, the original methods
are verified against the Go port, and full normal workspace verification passes.
Current priority is the original distributed TLC algorithm on `master`. Leave
the rpc25519/Tube alternative aside. Port its behavior using Go interfaces and
transport, without Java RMI, Java serialization or JVM machinery. Java remote
interfaces are behavioral references, not compatibility targets. Canonical
adapter/context gaps remain separately tracked below. The user has supplied a green full-suite
baseline; run focused checks for changed code rather than repeating that suite.
Commit each tested chunk, aiming for every 10–15 minutes.

Distributed role usage errors name the native `tlago server`, `tlago worker`
and `tlago fpserver` commands. Explicit Java counterpart descriptions in help
remain reference information; error usage does not require a Java installation.
Coordinator worker labels retain the source `TLCWorkerThread-` statistics
prefix, minimum three-digit counter formatting and ASCII URI rendering.
Formatting reuses the existing URI formatter and leaves endpoint metadata intact.

Coordinator threads retain their supplied block selector. They do not borrow
the server's selector or construct a replacement when it is absent. Missing
selectors enter the model-error handler and thread cleanup; during a recoverable
batch failure, requeueing precedes the selector failure, preserving pending work.
Focused selector, error-handler, original smart-proxy and short TCP checks pass.
Recoverable-batch retries also require a queue: they cannot lower the transfer
limit or report continuation after skipping requeueing. Requeue failures retain
preceding queue mutations and escape the inner worker-failure catch.
Block-selector mode settings are captured once for the process, with static,
unlimiting and limiting precedence. The static batch size is captured separately
when first constructing a static selector, including a failed construction.
Later setting changes cannot silently replace the process's selection policy.
Batch-result handling requires its timer task and coordinator: received-state
counts update first, then the keepalive timestamp, then generated-state delta.
An absent timer enters worker-loss cleanup and requeues assigned work rather
than returning a publishable result. Direct-construction fixtures supply the
timer task normally created by the production constructor.
Worker-loss cleanup wakes queue consumers unconditionally after requeueing and
clearing assigned work, before decrementing the worker count. It uses native
`WakeAllWaiters`, preserving suspension; registration still uses the distinct
`ResumeAllStuck` operation. All four queue implementations and short concurrency
checks cover this boundary.
Fingerprint-server registration belongs to distributed coordinators. Base
coordinators reject it before accessing the manager, preserving the source
server-level error even with a dynamic or absent manager. Local and native TCP
checks cover rejection without registration mutations or acceptance messages.

The original checkpoint-on-violation and time-bound model tests are now complete.
Both pass normally with their full assertions; the time-bound test retains the
source five-second limit. The generated checkpoint trace recheck is also complete
and passes with its original checkpoint interval and exact seven-state trace.
The CodePlexBug08 recovery model also passes against the unchanged Java archive
with three workers and all original assertions. Intern-table recovery now occurs
before tool construction, matching Java and preserving checkpoint identities.
Debugger/scoped-identifier methods are already mapped.
JPF concurrency verification and benchmarks remain separately tracked; JVM-only
assertions and source-failing methods require honest reconciliation rather than
invented Go equivalents. Keep pending entries visible in
[TODO_TEST_PORT.md](TODO_TEST_PORT.md).

Email reporting is forbidden and its removal is already committed. Do not
restore JavaMail, SMTP, MIME/Activation, ImageIO/AWT, image codecs, or JVM
emulation pursued for email. Preserve generic exceptions, console output,
packaged properties, networking primitives, OpenJDK notices, and `x/text`.
The current user request authorizes the SANY XML repairs listed in the audit.
ApalacheIR corpus sweeps remain deferred.

SANY audit progress: all F01–F25 repairs are complete. Every original primary
method has a verified current mapping and full-run receipt: 95 pass, including
one empty source body, and one preserves the upstream ignore. All 59 repaired
and 35 unchanged reviewed functional contracts pass with original parameter
matrices. All 227 mirrored fixtures and the embedded schema remain byte-identical
to pinned Java. XML export requires `xmllint` (libxml2) on PATH unless `-o` is
selected; validation uses no network. This completes the listed source test
contracts, not general SANY parity beyond their original assertions.

Full normal `go test ./... -count=1 -timeout=60m` has passed with status 0 and
zero failures: root 1,552.199 seconds, TLC 771.039 seconds, SANY 2.419 seconds.
Handle 53216 is retired. All original workload bounds are retained, with no
race instrumentation. Local-interface tests require the unrestricted environment.
The run verifies implementation and test sources at `23f046e`. Distributed
changes after `d7c029b` have focused receipts in PORT_PROGRESS.md.

## Current verified state

Error-trace printing propagates returned state-reconstruction and alias errors
instead of fabricating fallback states. Ordinary printing errors reach the
coordinator catch; fatal errors escape before queue completion/notification.
Noninitial error-trace reconstruction now follows source fatal branches `3`,
`4` and `5`: recovery and bug diagnostics precede exit status 1; branches `4`
and `5` additionally print the unrecovered standalone state. Child-process
checks verify output ordering and that exit bypasses deferred cleanup. Fingerprint-
sequence reconstruction also preserves branch `2` fatal diagnostics with signed
fingerprint formatting. It passes the predecessor state to `Tool.GetState`,
restores the random generator only on normal completion and retains returned
info metadata. Missing initial reconstruction remains a nil array element when
it is the sole fingerprint, or fails on dereference before another lookup.
`Tool.GetState` returns no match for fingerprint-only/predecessor-state forms;
the predecessor-info form returns the source evaluation error. Printing copies
UID/worker metadata only for a current state reconstructed after a prefix and
a successor reconstructed in the noninitial branch. Empty-prefix current and
initial-transition reconstruction retain their own metadata before alias evaluation.
A missing initial transition fails at the state printer after printing the initial
state; the ordinary coordinator catch reports it and completes/notifies the queue.
Disk fingerprint traversal restores its saved cursor only after normal completion
and propagates a failed restoration. Partial-read failures retain the consumed
cursor; the predecessor chain follows the source initial-state sentinel without
a self-link shortcut. Concurrent reconstruction excludes the anchor record from
the result length, restores randomness only after normal completion and retains
source UID/worker updates before later lookup failures. Missing initial results
fail at metadata dereference; missing successors print branch `2` diagnostics
and exit. Public concurrent trace methods propagate reconstruction errors before
printing rather than discarding them for fallback traces. Record lookup requires
the selected worker slot and preserves index/null failures instead of substituting
an in-memory trace. Collection rereads the end record under the trace monitor,
follows the predecessor chain to the source initial/requested-fingerprint boundary,
and retains the monitor through reconstruction. Initial/equal-state short paths
do not require workers. Checkpoint begin, commit and recovery require each
worker in source order, preserving earlier mutations and stopping before later
workers or marker publication on failure. Begin and level reporting hold the
trace monitor. Level reporting uses worker maxima with a minimum of one; it
does not substitute the base trace level. Concurrent enumeration creates required
worker readers under the trace monitor. Reader creation snapshots the existing
writer cursor without flushing or reopening the writer. Cursor failures propagate
before selector advancement; exhausted fingerprint access fails instead of
returning zero. Closing stops at the first failure and retains the closed reader
owner. Neighbor-only advancement and selector-only reset preserve source behavior.
Reconstruction requires the tool only at actual lookup calls; empty and supplied-
initial paths retain normal randomness restoration and returned info metadata.
Printing requires its lookup/alias tool and preserves a returned nil alias instead
of displaying an unaliased substitute. Disk reconstruction cannot select an
in-memory fallback merely because the tool is absent. Initial-only printing
still requires no tool. Worker record writers assign their last pointer before
writing predecessor, worker and fingerprint bytes. Partial failures retain the
attempted pointer, consumed bytes and earlier depth update, without state/counter/
mirror publication. Successful successor writes preserve the generated action
and predecessor metadata policy; the native mirror records the published state
without modifying it. Worker depth comparison uses signed 32-bit next-level
arithmetic, preserving the previous maximum when addition wraps. At the depth
limit, completed record and UID/worker updates survive the later predecessor
failure; extended states retain the predecessor assigned before that failure.
Worker recovery publishes the checkpoint pointer after its complete read, then
closes the checkpoint reader and seeks the existing trace owner. It cannot
reopen missing or closed owners; closing retains the closed handle. Truncated
checkpoint reads leave the pointer unchanged. Checkpoint creation also requires
the existing owner: it flushes before opening/truncating temporary metadata and
does not consult a saved creation error. Commit failures preserve source I/O
classification, text and delete-before-promotion mutations. Worker construction,
filename/context guards, other owner access and trace cleanup boundaries still
need audit.

Distributed initial-state publication changes only the state UID, preserving
worker, predecessor, action and level metadata. Fingerprint insertion precedes
the required root trace write and queue enqueue; property checks follow. Missing
trace/queue failures retain earlier mutations and suppress later initialization
elements. Seen or excluded states do not require unused publication owners.

Checkpoint and recovery checks preserve source mutation, failure categories and
phase ordering through native Go interfaces and transport. Detailed receipts
belong in PORT_PROGRESS.md; the supplemental checks below add no completion
credit to the four assumption-disabled Java remote model harnesses.

The full native EWD840/MC06 process harness retains N=7, the original final
114,942 distinct states and zero queued states, exactly one recovery, no repeated
initialization or GENERAL/lost replies, normal fresh coordinator/worker exits
and joined child processes. Initial and complete mid-run checkpoint recovery
pass. Latest interruption receipts are:

| Interruption boundary | Recovered fingerprints | Recovered queue | Checkpoint files |
| --- | ---: | ---: | --- |
| Before queue commit | 20,480 | 16,384 | Old queue and trace metadata |
| After queue commit | 20,480 | 12,288 | New queue, old trace metadata |
| After intern commit | 20,480 | 12,288 | New queue/trace/intern, old FP checkpoints |
| After first nested FP commit | 20,480 | 12,288 | First FP promoted, second FP byte-identical to its old checkpoint |

The harness inspects committed queue headers, trace metadata, temporary-file
promotion and prior FP file contents independently of producer count markers.
FP recovery reconstructs the complete persisted trace. Commit failpoints wrap
the same configured production memory FP factory in test code; there is no
production checkpoint hook. Isolated trace-commit interruption and broader
process/network failures remain pending; this does not establish atomic recovery.

Focused recovery contracts are also verified:

- Trace writes publish the attempted record pointer before predecessor and
  fingerprint bytes. Partial write failures retain that pointer and buffered
  bytes without publishing a record or changing the state UID/metadata.
- Trace metadata reads update the saved pointer before seek, including closed
  owners. Enumeration creation/cursor/read/reset errors propagate without zero
  insertion or false end-of-trace; reset retains source cursor/length behavior.
- Memory and disk queues retain untouched slots and publish empty states before
  reading them. Completed state-header fields survive later failures. Memory
  recovery retains its cursor and fixed capacity; successful disk recovery closes
  input before reader restart. Queue failures precede FP recovery/publication.
- MultiFPSet trace recovery selects each child's RecoverFP. Disk/off-heap
  duplicate errors or warnings, memory/parent runtime assertions, routing and
  prior insertions survive. MemFPSet1 recovery retains source field assignments,
  count/growth/zero quirks and exact zero/negative allocation boundaries.
- Disk file recovery replaces readers one slot at a time, retains earlier
  replacements and untouched later readers on failure, and resets the pool
  cursor only on success. Named checkpoint creation retains locks/flusher state
  on flush/copy failure; success advances the marker and releases ownership.
- Memory checkpoint creation opens files without creating missing parents.
  Queue/trace commit errors retain their I/O category through native payloads;
  partial pool deletion and earlier commits remain when later phases fail.
  Native manager checks retain source diagnostics and healthy-partition
  continuation without reassignment or resetting failed storage ownership.

Unused RMI-specific error carriers are removed. Native Go failure traits drive
retry/shutdown decisions; Java remote machinery is outside the port.

Coordinator checkpoint, recovery and close require their source-owned queue,
trace and fingerprint manager. Missing components fail at their actual access;
earlier checkpoint files, recovery reads and trace closure remain observable,
and later resume, commit or metadata deletion does not run. Failed suspension
still returns without accessing later owners. Focused native ordering and TCP
checkpoint checks pass; there are no direct original methods for these missing
component cases, so they add no original-method completion credit.

Coordinator management controls now retain the source monitor and operation
order. `Stop` marks done, finishes the queue, then wakes the reporting wait;
`Suspend` and `Resume` hold the same monitor around their queue calls. Missing
owners and queue failures escape without a fabricated success or premature
notification, and the monitor is released on failure. Focused ordering, actual
report-wait wakeup and checkpoint checks pass. Upstream has no direct management
control methods, so supplemental checks add no original-method credit.

Management queries now retain missing-owner failures instead of fabricating
zero, `"N/A"` or a partial generated-state count. Running depth requires its
trace, generated count requires the fingerprint manager, and remaining-work
count requires the queue. Average block count comes only from the selector;
the unused fallback field is removed. The source’s explicit inactive sentinels
and missing-manager distinct-count sentinel remain. Focused checks verify
queued plus assigned work, signed count overflow, unchanged rates, non-consuming
current-state reads and empty-queue `"N/A"`. These checks add no original-method
credit; broader distributed completion remains unproven.

Periodic coordinator progress and final tool progress now use the existing
locale-aware message integer formatter, matching Java's `MP.format` calls.
Nine fresh native locale processes verify grouping, digits, negative affixes,
signed limits and the plain final `TLC_STATS` parameters. Existing worker locale
checks and the four original MP methods pass. These supplemental coordinator
formatting checks add no original-method completion credit.

Trace depth reporting now exposes I/O failures through an explicit Go error
return. Coordinator periodic/final reporting stops at that failure; management
getters retain their source I/O catches and `-1` result. Disk depth traversal
restores its cursor only on success and retains partial reads on failure.
Focused failure, monotonic-depth, fresh-process final-reporting, native TCP and
original TLCGetLevel/TTrace checks pass. Direct source tests for these failure
boundaries do not exist, so supplemental checks add no method credit.

Server-thread catch and finally run in source order even if trace printing in
the catch fails. Uncaught handler/finalizer failures print a native diagnostic
and stop only the owned goroutine. A final cache-read remote failure, returned
or panicked, warns and continues; an unchecked cache failure skips the remaining
finally operations. Five joined native process checks and focused worker RPC,
codec, checkpoint and original smart-proxy tests pass. Supplemental finalizer
checks have no direct original methods and add no completion credit.

Coordinator keepalive status calls now apply the same remote-failure catch to
returned and panicked endpoint failures. Worker loss deregisters once, requeues
the assigned block in order and decrements the worker count once. Unchecked
local failures still escape without ownership changes; recent/future activity
suppresses status calls. Focused local and native TCP checks pass. These timer
boundaries have no direct original methods and add no completion credit.

Final coordinator worker exit normalizes returned/panicked failures before its
three-family dead-worker catch and always removes that thread's registration.
The shutdown hook has a narrower silent catch, represented by the native
`WorkerUnavailable` payload trait; server and other I/O failures are reported
and iteration continues. Payload round trips, 12 fresh completion processes,
56 hook cases and actual unavailable TCP endpoints pass. All communicating
roles need the current payload build. Supplemental checks add no method credit.

SimpleCache ratio arithmetic now retains source floating-point division,
including zero-denominator infinity/NaN and signed zero. Missing cache ownership
fails instead of reporting zero. All 271 source counter reference rows and
native TCP edge values pass. The RPC reply carries raw IEEE bits because gob's
float-field omission lost negative zero; all roles need the current build.
The worker exit-message compact formatter now uses native Go arithmetic and
locale symbols: at most three fractional digits, grouping, negative affixes,
signed zero, infinity/NaN labels and legacy/numbering variants. It passes the
271 counter rows, 6,610 additional numeric rows, 1,860 locale rows and six fresh
locale processes. This differs from the coordinator’s two-decimal worker
statistics formatter. These supplemental checks add no original-method credit.

Fingerprint check tasks now catch I/O failures before executor completion
wrapping, print `GENERAL` and return the source sentinels (`MaxInt64` for
fingerprint distance, `false` for invariants). Unchecked task failures retain
the source execution-failure path. Focused local/TCP checks and original manager
methods pass; supplemental failure checks add no original-method credit.

`NextStateResult` getters retain source null failures rather than fabricated
default values; the delta also rejects a null state-partition array. Direct
arrays, empty-array distinctions and signed counter overflow remain intact.
Focused payload/smart-proxy and worker TCP result/retry checks pass. Coordinator
cache-ratio output now retains NaN/infinities, signed zero and decimal half-up
rounding. All 271 canonical Java reference rows pass, including large values
and deterministic bit patterns. Locale-specific separators/digits now pass
1,860 reference rows covering the 1,068 existing locale keys and 792 numbering
system variants. Fresh-process checks verify configured locale selection and
either initialization order with message formatting. POSIX worker statistics
suppress grouping while MP's explicit grouping remains intact. These checks
add no original-method credit or claim of complete distributed parity.

Native state transfer now preserves shared backing value arrays for states, tuples,
records, functions, configured operator argument rows, tuple products and record
sets. Receiver mutations remain visible through shared arrays without touching
sender storage; recursive arrays and nil/empty distinctions survive gob and TCP.
State arrays use the same graph table as composite values, including arrays
shared between distinct states and tuples. Invalid or conflicting state-array
references fail decoding. Peers use the same build for the updated payload.
This is native Go graph transfer. ValueVec identity, active count and full backing
capacity are also preserved, including shared storage and unused recursive slots.
Record and record-set name arrays also retain shared storage and isolated
receiver ownership across gob and worker TCP. Further metadata/custom-data and
process failure/recovery work remains pending.
These short native checks add no original-method completion credit.

Fingerprint checkpoint/recovery and shutdown now traverse live registrations
with a fixed initial count, preserving wrapper identity and source call order.
Checkpoint phases and caught-I/O hostnames resolve the current slot separately;
shutdown captures the next wrapper before exiting the current one. Focused
manager, failure-boundary and native TCP checkpoint checks pass. Supplemental
registration-change checks add no original-method credit.
Fingerprint and invariant tasks also preserve null-endpoint failure categories
and healthy completion results. Null registrations still fail during submission;
these supplemental boundary checks add no original-method credit.
Native manager payloads now distinguish empty slots from registrations with nil
endpoints, retaining shared/distinct wrappers and availability. Gob/TCP receipt
defers failover until an operation, with worker-local mutations. All communicating
roles need the current payload build; original-method counts remain unchanged.
Lifecycle endpoint access also preserves explicit null failures: checkpoint
stops at the failing phase, while distributed close reports and continues.
Focused local and short TCP lifecycle checks pass without original-method credit.

Record-backed printable states now retain their record through the native value
graph, including sharing with caches, ordinary state values and other printable
states. Gob/TCP request/result/failure checks preserve default display, extra
record fields, stored metadata and underlying fingerprints. `RecordValue.ToState`
now compares variable-name tokens as the source does, so received records bind
local spec variables correctly. The source’s separate `_format` object-identity
behavior remains covered. Invalid print-record IDs/types are rejected. All roles
need the current payload build. Existing original record and alias checks pass;
new transfer checks add no original-method credit.

Native cached-state transfer now retains populated, empty and nil maps. Shared
cache maps stay shared across roots and predecessor states; equal-content
separate maps remain separate. Cached values share the invocation’s existing
value graph, including null entries and recursive values. Gob/TCP request,
result and worker-exception checks pass with isolated receiver ownership.
Invalid references, duplicate keys and keys outside the source signed-int range
fail explicitly. All roles need the current payload build. These supplemental
checks add no original-method credit; unsupported evaluator/custom values remain
subject to the existing explicit codec errors.

Native state transfer now supports predecessor references in the shared state
graph. Shared parents, non-root ancestors, typed nulls and cycles retain their
identity and stored levels without invoking metadata-mutating setters. Gob/TCP
request, result and worker-exception checks preserve receiver isolation and
shared values. Invalid references and unsupported ancestor evaluator objects
are rejected. All communicating roles need the updated payload build. Other
evaluator metadata and custom state/data types remain separately tracked; this
adds no original-method credit or Java serialization/RMI support.

Worker predecessor/successor/result partitions now retain the source’s separate
unbounded vector behavior: default capacity 10, doubling on growth and indexing
the backing array. They no longer inherit the tool `StateVec`’s `SetBound`
limit. Native result decoding restores that policy with capacity equal to the
active count; unused source slots are still omitted from the wire. Local/gob/TCP
checks verify 11 successors with `SetBound = 1`, later receiver growth, malformed
FP-selection failure context and FP-before-trace publication order. Ordinary
bounded vector checks still pass. No direct original vector tests exist, so
these supplemental checks add no original-method credit.

Block selectors now retain the source null-server assertion and numeric
conversion behavior, including zero workers, NaN and infinities. Proportional
and static selectors ignore transfer-limit updates; limiting and statistical
selectors apply them. Focused queue-bound and actual-dequeue average checks,
the original smart-proxy methods and native TCP retry/loss check pass. Upstream
has no selector test methods; supplemental checks add no original-method credit.
The average uses independent atomic reads/writes, retaining source lossy updates
and signed overflow. Transfer-limit access is protected for concurrent worker
retries. An exact short concurrency race check passes; long workloads remain
separate from race instrumentation.

Coordinator keepalive now reports uncaught failures and stops only its timer
rather than terminating the Go process. It retains the ten-second first call
and sixty-second scheduling from actual invocation start, matching the worker
timer. A normal real-timer check verifies runtime/fatal failures, diagnostics,
unchanged assigned work and joined timer owners. Native TCP keepalive/loss and
original smart-proxy checks pass; no original-method credit is added.

The native EWD840 process harness now covers two standalone FP servers. Before
starting its worker, it reads the coordinator's published manager reference
graph and checks two distinct nonempty stores totaling the complete 16,384-state
initial frontier. Each FP process owns private temporary storage. The unchanged
MC06/N=7 workload must then finish with 114,942 distinct states and an empty
queue, exactly two FP registrations, no GENERAL across any role and no lost RPC
reply. This is supplemental native coverage, not disabled Java harness credit.

Native fingerprint registration now stores the endpoint reference without an
extra connection probe. This retains the dynamic manager's source capacity
check and registration-latch order even when the referenced store is stopped.
Short TCP checks require one acceptance/countdown, the exact capacity rejection
for an extra unreachable reference, and no consumed slot for incomplete native
references. Original dynamic-manager and native FP lifecycle checks pass.
Worker registration requires the queue wakeup before either URI call or thread
creation. Missing queues and wakeup failures cannot create partial registrations;
focused local/TCP checks preserve failure categories and monitor release.
Worker registration retains its required URI calls.

Coordinator error handling requires the trace and queue at their source access
points. Missing traces report the trace-printing failure; missing queues retain
prior error state and skip completion notification. Block selection no longer
treats a missing queue as completed work. Thread finally cleanup still runs.
Focused handler/selector/finalizer and short TCP checks pass; supplemental checks
add no original-method completion credit.

Workers now retain the supplied fingerprint manager, including nil, rather than
inventing an empty manager. Missing-manager failures occur after generation and
its statistics update, preserving predecessor/error context and computing cleanup.
Generation failures still take precedence; empty managers remain distinct.
Focused local/TCP checks pass and add no original-method completion credit.

Distributed application generation/property loops now read their current arrays
at each index, observing evaluator-driven replacement, shrink and growth. Initial
constructor capture of tool arrays is retained. Focused traversal/native TCP
checks and the three original distributed initialization/TLCSet models pass;
these supplemental cases add no original-method completion credit.

Distributed application successor results now have separate fixed-size array
storage, preserving shared state objects and larger-vector merge order. Tool
accumulator mutations cannot rewrite returned slots; validation-time shrink/grow
retains source result length and failure ordering. Focused worker/TCP checks and
both unchanged DieHard native process variants pass with the original trace.
These supplemental checks add no original-method completion credit.

Worker-process shutdown reads each runnable’s worker at its turn, so later
startup publication during an earlier exit is observed. A later nil runnable
retains earlier exit/latch mutations and releases the lifecycle lock; unpublished
workers still receive the source skip. Focused normal/race and short TCP lifecycle
checks pass. Supplemental checks add no original-method completion credit.

Worker fingerprint-manager snapshots now create native endpoint references
without dialing every store. A stopped fingerprint host no longer prevents
snapshot receipt; the first operation reaches TLC's existing failover logic.
Reference ownership is committed after complete graph decoding. Closing the
coordinator client closes established child connections and disables unused
references too. Concurrent first calls safely share a published client; no
operation is replayed and failed established connections are not redialed.

Worker request/result graph validation and representation failures now retain
the remote I/O category used by the coordinator's worker-loss catch. Focused
TCP checks require codec causes, no automatic retry or unintended dispatch,
continued host availability and exact assigned-block requeue/deregistration
without marking the unfinished coordinator done. Source exceptions raised
while materializing values retain their application category. Unevaluated lazy
values now report the source runtime diagnostic rather than a generic codec
error, and never dispatch a worker request.

The worker file resolver is now named `DistributedFilenameToStreamResolver`,
with constructor `NewDistributedFilenameToStreamResolver`; the misleading RMI
Go API name is removed. Worker bootstrap, root integration and process helpers
use the native name. TCP checks retain basename file caching, deletion/refetch,
binary/empty files and separate temporary-directory ownership. The Java class
remains a behavior reference, not a transport compatibility target.

Native discovery now classifies malformed coordinator locations with
`DistributedLocationError`, retaining the location and parser/validation cause.
Keepalive logs this category and continues, matching the source catch rather
than terminating its timer as an uncaught failure. Parser escapes, unsupported
schemes, missing/nested bindings, user information, queries and fragments retain
their existing validation rules. Focused checks require repeated logging without
worker exit, executor shutdown, timer cancellation or completion-latch release.

Native model-value data now supports `int8`, `int16` and `float32` as well as
the existing scalar types. Floating-point data uses integer IEEE bit patterns:
gob's omitted zero struct fields otherwise lose negative zero, including on
the old `float64` path. Integer ranges and float32 bit widths are checked.
Gob and worker TCP checks retain types, extrema, signed zero, subnormals,
infinities and NaN. The native payload layout changed; peers use the same build.

Native TCP connection loss during accepted fingerprint checkpoint begin,
commit and recovery calls is now verified. The source I/O catch emits one
warning and continues to the healthy store without partition reassignment or
availability changes. The coordinator resumes its queue and commits queue/trace
checkpoints despite the caught FP failure. Accepted disconnected storage work
remains alive until explicitly released and joined. This is connection-loss
coverage, not process-crash recovery or an atomic distributed checkpoint.

Assigned-block checkpoint recovery now covers two independent native TCP FP
stores. Initial and successor fingerprints occupy different partitions, stored
in separate metadata directories and reopened behind fresh hosts/tables. The
queue frontier and disk trace keep their exact committed identities; recovery
must neither merge nor swap the FP partitions. Local and single-store cases
retain their existing assertions.

Fresh-process remote-FP CLI recovery is a verified limitation of the pinned
Java flow: the dynamic manager starts empty, recover runs before publication
and FP registration, and recovery iterates that empty registration list. The
Go startup retains that order. Restoring remote FP stores through the CLI would
require an enhancement to the source algorithm. Registered-endpoint recovery
is verified separately. Fresh-process mid-run local MemFPSet recovery now also
completes the full unchanged MC06 model after an abrupt post-commit coordinator
exit. Interruption before the first replacement-file commit now has a full native
model check: the old queue survives while default MultiFPSet rebuilds from the
full persisted trace. Interruption between commits and broader process-crash
coverage remain pending.

Model-value byte data uses native graph references instead of per-value copies.
Shared buffers stay shared across states and result partitions; equal-content
separate buffers remain distinct, and receiver data stays isolated from sender
data. Nil/empty byte buffers remain distinct. Graph and worker TCP checks pass;
opaque custom data and evaluator metadata still require their own audit.

Coordinator system-failure diagnostics use the source throwable overload for
stack-overflow and out-of-memory categories, retaining debug stacks and cleanup
order. The shared stack printer now completes each ToolIO line with println,
as Java does, instead of leaving a raw stack in the unfinished-message buffer.
Focused process, original output and native keepalive checks pass.

Native fingerprint failover now has in-flight block coverage for put/contains
in sequential and concurrent manager execution. A closed TCP host must trigger
survivor reassignment while its accepted storage handler is still paused; result
partition order, shared survivor wrapper and exact warning count are retained.
Transport closure leaves owned storage work alive, and the test joins it. This
covers connection loss, not fingerprint-process crash/recovery.

Worker keepalive now reports coordinator failures through the throwable printer,
retaining debug-enabled sender stacks. Short native TCP checks cover completion,
missing binding, disconnected coordinator and status failure. Computing/recently
active workers stay running; idle workers exit, release their latch, cancel
keepalive and reject callbacks. No timer interval or activity timeout is reduced.

Worker, coordinator and fingerprint RPC handlers share fatal-error encoding.
Returned and panicked fatal endpoint failures both receive the native remote
I/O category, retaining the original diagnostic/cause graph. Ordinary failure
traits and real worker evaluation wrappers remain intact. Exact short TCP
checks cover next-state, aliveness, cache and exit calls plus a coordinator
settings call; existing native RPC and original manager checks remain green.

Local fingerprint endpoint fatal errors now escape manager catches whether
returned or panicked. Scalar/block/statistics, checkpoint/recovery and close
paths retain the source fatal category without failover or availability changes.
Remote fatal failures still arrive as native I/O operation errors. Focused
original manager and native RPC checks pass; no transport compatibility changes
are involved.

Permanent native DieHard and TSnapShot process checks now retain the original
active model assertions with coordinator-owned and standalone TCP fingerprint
storage. DieHard requires its exact seven-state trace; TSnapShot requires an
empty final queue. Both require FINISHED, BEHAVIOR and no GENERAL across all
roles. Fixtures are byte-identical to Java; process progress is visible under
`go test -v`. These supplemental checks do not complete the disabled upstream
harness or add original-method completion credit.

The TLC bridge now reads the checked canonical OpDef/ThmOrAssumpDef level,
replacing the XML-exporter's estimate. Canonical construction for validated
subexpression selectors now retains actual selected bodies, formal parameters,
caller arguments, substitution prefixes, LET contexts and selector ownership.
Original Test206 and Test209 pass unchanged; 233 external Java/Go graph rows
agree, including observed sharing and `subExpressionOf` references. This is
not a complete port of `Generator.selectorToNode`: invalid-selector diagnostics,
allocation order and general graph coverage still require faithful translation.
The bridge's AST-based export and runtime graph construction also need further
canonical integration. No estimated-level fallback is restored.

Named/unnamed INSTANCE generation no longer compares canonical export counts
with the native AST's export list. That comparison is absent in Java and wrongly
rejects modules with recursive LOCAL definitions: Java retains the recursive
declaration's locality flag. Actual child/body completeness checks remain.

Historical full normal workspace verification of `4cd17ea` has finished. Session `12025`
returned status 1 and is retired. TLC passed in 775.694 seconds and SANY passed
in 1.525 seconds. Root failed in 1,561.759 seconds on three native XML fixtures:
the two missing-Naturals cases documented below, and an explicit PROOF location
fixture with an empty BY. Java reports the same two empty-BY errors. That older run applied no fixture
change and enabled no optional XML/Apalache sweep. The current audit fixes all
three invalid native fixtures, and current full verification passes. Log:
`/mnt/oldrog/tmp/tlago-canonical-driver-workspace.json`. This earlier snapshot
does not verify the subsequent selector and traversal changes.

Canonical graph traversal now follows Java's actual node edges, UID registration,
context enumeration and pre/post callbacks. Source reachability and delayed
operator substitution are translated over these nodes. The original semantic
corpus class is port complete in root `sany_semantic_corpus_java_test.go`: all 28
parameter rows are retained, including Semantics itself and the original
NegativeOpTest assumption. Assertions use actual operator references, source-node
comments and checked levels. The older AST facade checks remain supplementary.
This original translation exposed missing ExternalModuleTable root publication;
the driver now publishes the final external module after its level check, as Java
does. The original Java class and Go translation pass. No full graph-allocation,
invalid-selector or evaluator-sharing completion is claimed.

All three original TestSubexpressionSelectors methods are now port complete in
root `sany_subexpression_selectors_java_test.go`. Their helper parses and generates
without level checking and retains errors on abort. Require Java's exact first
error code, message and location, plus rejection without internal errors. The
all-navigation method exposed lost selector syntax at a flattened call callee;
generation now retains the actual selector for unresolved-name and range
accumulation, including the source UniqueString diagnostic parameter. Original
Java and Go pass; 16 external full-detail observations agree. The earlier native
API checks remain supplementary. Full selector generation/diagnostics are still
pending; these methods do not prove every invalid path.

Canonical child traversal now separately implements Java's containment edges,
filtered child lists, visitor callbacks and `pathTo`. Module children retain the
source cached array; application children are fresh and ordered ranges before
operands. Null entries remain visible in raw children but do not make a node a
nonleaf. The runtime path helper now uses this same null-filtering rule. Token
syntax constructors retain their source filename, fixing incorrect path matches
without a cached-location fallback.

External Java/Go comparisons agree on 2,885 child rows, 6,321 traversal callbacks,
5,770 exact/nonexact paths and 106 cache/callback-mutation observations across
original Test206, Test207 and Test209. These observations earn no test-port
credit and do not establish complete malformed-node or evaluator graph parity.
Existing original scoped-identifier and frontend checks pass after the filename
fix, as does complete SANY and all-package compilation. All handles are retired.
The full-workspace receipt above remains an earlier failing snapshot.

Runtime SemanticContext now retains Java Hashtable entry links and exposes
source-style symbol enumeration. Rehashing relinks shared entries; enumerators
retain their original bucket array and advance before callbacks. Runtime module
graph traversal resolves each enumerated key against the current binding.
Presence checks distinguish a stored null symbol from a missing key, and null
keys/exhausted enumeration preserve Java exception families. All 195 external
enumeration and traversal observations agree with Java; existing original model,
coverage and scoped checks pass (root 6.973 seconds), as do focused TLC tool checks
and compilation. Runtime LET context reconstruction and full canonical/evaluator
sharing remain pending; these observations add no original test-port credit.

Original parser BelchDefTests and IncrementalSyntaxParseTests are now port
complete in the root package. The former checks all five source token-boundary
rows; the latter uses the three original standalone inputs and concrete syntax
assertions. Earlier module-wrapper checks remain supplementary. Java passes
all eight rows; the current related frontend gate passes in 2.912 seconds,
complete SANY in 1.955 seconds, and all-package compilation passes. Production
is unchanged from the semantic-context commit. Main TLC inventory is unchanged;
full parser lookahead/AST parity and runtime LET integration remain pending.

Both `CheckSanySpecLevels` and the `CheckSpec`/TLC driver now invoke the actual
generated `ModuleNode.levelCheck`. The driver preserves Java's external-module
order, shared diagnostic log and raw-success gate. Integration exposed the
missing EXCEPT `AtNode.levelCheck` and ASSUME integer diagnostic parameter; both
are now translated. Existing temporal diagnostic expectations require the seven
exact Java messages/ranges. The native qualified-module fixture now declares
the named INSTANCE required by both Java and Go, retaining its two-state check.
The full SANY suite, focused frontend/config tests, all five original incremental
methods, 26 original TLC model methods and all-package compile checks pass.
The unchanged Java level-check originals pass all 52 cases. The INSTANCE
assertion uses `int32`, preserving expected values 1 and 3. No new full-workspace
verification baseline is established by these focused checks.

The latest completed full normal workspace suite passes on `44aaf11`:

| Package | Result | Duration |
| --- | --- | ---: |
| Root package | Pass | 1,547.242 seconds |
| SANY tests | Pass | 1.298 seconds |
| TLC | Pass | 780.089 seconds |
| CLI command | No test files | — |

Session `16991` returned status 0 and is retired. Log:
`/mnt/oldrog/tmp/tlago-expression-grammar-workspace.log`. This verifies the
parser-footer and expression-grammar snapshot with original bounds and normal
execution. It predates `aed6180` stack corrections and newer semantic lookup and
real-module packaging. Historical receipts belong in `PORT_PROGRESS.md`.

Normal workspace session `22503` returned status 1 and is retired. Its root
package failed the invalid native unary-minus fixture, which is now corrected
against Java and passes focused checks. Its complete TLC package passed in
769.038 seconds; CLI has no tests and SANY passed in 1.253 seconds. The root
failure prevents full-workspace pass credit. Log:
`/mnt/oldrog/tmp/tlago-operator-resolution-workspace.log`.

Normal full root-package session `45205` returned status 0 in 1,656.783
seconds and is retired. It verifies the `008083a` recursive-function and INSTANCE
snapshot, before stateful LET and symbol-constructor changes. Log:
`/mnt/oldrog/tmp/tlago-recursive-function-instance-root-full.log`.

Normal full root-package session `77352` returned status 1 in 1,564.264 seconds
and is retired. It tested `4085e45`, before proof-reference changes, and exposed
`TestJavaEWD998ChanDebugger`: a LET in an INSTANCE substitution gave imported
`Len` a fabricated zero arity. Preserve unknown signature metadata instead of
replacing the real signature. The unchanged original test now passes, alone in
13.186 seconds and in the final focused gate. Log:
`/mnt/oldrog/tmp/tlago-symbol-instance-root-full.log`.

Normal full root-package session `99670` returned status 1 in 1,605.816 seconds
and is retired. It tested `8d13d91`, before mixed DEFINE and INSTANCE changes.
Three failing test classes contain five failures: native XML snippets
TerminalByPrefixFactXML and ProofPickBoundLevelXML use `..` without Naturals;
proof `@` shorthand is not generated; unchanged original
TestJavaSafetyDumpLoadTraceJSONAutoWorkers and
TestJavaSafetyDieHardAliasSub2DumpLoadTraceJSONAutoWorkers report differing
replayed trace states. Log: `/mnt/oldrog/tmp/tlago-proof-statement-root-full.log`.
Prioritize these failures before further feature work. Do not weaken the original
trace assertions or change worker settings. The original fixture and proposed
Naturals imports have separate Java source comparisons; automatic approval review
rejected both proposed native fixture edits. Explicit authorization is pending.
No fixture edit was applied. Proof `@` generation is now corrected: the unchanged
native case passes in 0.023 seconds and 11 bounded Java comparisons match. The
trace failures remain unresolved. Ten normal repetitions passed, but 100 further
repetitions reproduced one JSON auto-worker mismatch in 155.708 seconds. Preserve
that failure receipt; passing repeats alone do not prove a fix. A scratch
reconstruction of its ten console states now replays identically in pinned Java
and Go: seven states and 36/7/1 statistics. The earlier captured alias artifact
also replays identically in Java and Go. Both comparisons support retaining the
source invariant checks on excluded successors. Keep the original prefix
assertions unchanged and the source expectation issue unresolved; detailed
receipts are in `PORT_PROGRESS.md` and the LSB replay section of `TLC_ARCH.md`.
No current complete workspace pass is claimed.

Earlier focused verification receipts belong in PORT_PROGRESS.md. Current
checks retain their recorded scope; passing translated tests does not establish
whole-method fidelity where reconciliation gaps remain.

Source/native scratch observations verify exception types/causes, delayed output,
constructor diagnostics and storage artifacts. These are manual evidence, not
new persistent tests or test-port credit. Preserve each receipt's exact scope.
The record-lint matrix matches Java in 33 valid cases. Six full front-end plus
constructor observations now also match with the real lint phase included.
Seven manual semantic-output observations match Java for one error, dependency
errors, chains, siblings, nested modules and a successful dependency. These cover
reporting order and accumulation. Eleven further observations match generation
versus level gating, nested graphs, definition-before-top-level ordering, proof
levels, general assumption checks and unique named-theorem errors. Sixteen
import observations match Java for declarations, definitions, named facts,
class conflicts, diamonds, parameterized instances, reuse, accumulated warnings
and explicit versus implicit instance namespaces. Nine INSTANCE observations
match source collision kind/arity rules, repeated parameterized wrappers,
parameter-free reuse, local definition order and existing EXTENDS bindings.
These are bounded audit
observations, not a proof of complete semantic-node or context iteration parity.

The latest checker correction initializes coverage, liveness and cached config
in the parent constructor before storage and workers. Liveness I/O failures are
raised immediately; DFID's worker/liveness assertions follow parent construction.
BFS creates queue, trace and the selected fingerprint implementation in source
order, without an unused default fingerprint initialization. Disk graphs do not
create missing parent directories. RandomAccessFile opening failures retain the
source FileNotFoundException. The original five-second time-bound test passes
unchanged. Trace and worker constructors now propagate opening failures immediately and
preserve source path concatenation. Buffered random-access files support all four
source modes, including synchronous flags, and preserve opening versus later I/O
exception types and messages. Experimental liveness simulation creates unique
worker temporary directories itself, as Java does; failures are immediate.
Further constructor and front-end parity remains audit work.

## Long verification receipts

The earlier LSB session `59782`
failed after 19,420.414 seconds and is retired. It was compiled at `a915e08`;
log `/mnt/oldrog/tmp/tlago-long-lsb-random-unlimited-final.log`. Last logged
progress reached 2,144,786,516 of 2,147,483,648 iterations. Fingerprint merging
failed with `IOException: no space left on device` on `/mnt/oldrog/tmp`.
No full-workload success or inventory credit is claimed. Its unchanged draft and
overlay remain `/mnt/oldrog/tmp/tlago-long-random-family-draft_test.go` and
`/mnt/oldrog/tmp/tlago-long-random-family-overlay.json`.

LSB session `32959` returned status 0 and is retired. The unchanged random
draft passed in 20,048.87 seconds with all 2,147,483,648 insertions, original
seed and factory, per-insertion assertions, checkpoint commit, invariant and
final-size check. Log: `/mnt/oldrog/tmp/tlago-long-lsb-random-large-volume.log`.
Both temporary-directory variables used the large workspace volume. This
snapshot includes buffered-file and trace corrections, before the simulator
correction; the TLC fingerprint implementation is unchanged since that run.
The abstract method, setup/progress/teardown and concrete factories have now
been reviewed against Java. The complete LSB and MSB random bodies are installed
in `long_heap_fpset_stress_java_test.go`, under `tlc_fp_stress`. Java excludes
these heap classes from `test-dist-long`; the explicit Go target preserves the
full workloads without adding hours to ordinary test runs. No duplicate LSB run
is needed for unchanged fingerprint code. Do not reclaim unrelated files.

MSB session `63113` is retired with status 0. Its full random draft passes in
13,343.32 seconds, including all 2,147,483,648 insertions, checkpoint commit,
invariant check and final size. It was compiled at `a915e08` before later
flusher endpoint/count assertions and file-helper fixes. Preserve that limited
receipt; it does not verify newer production changes. The installed translations
retain original factories/configuration, seed, checkpoints, assertions and bounds.
Their two method contexts now receive translation credit; current full execution
is qualified separately.

MSB session `5144` is retired with status 0. The complete current fingerprint
run passes in 13,063.40 seconds, including all 2,147,483,648 insertions,
checkpoint commit, invariant and final-size assertions. The isolated binary
`/mnt/oldrog/tmp/tlago-heap-fp-stress.test` was compiled from `132a77f`
fingerprint production and the installed test translation; subsequent semantic
and parser changes do not alter that fingerprint code. Log:
`/mnt/oldrog/tmp/tlago-heap-random-msb-current-full.log`. The run finished at
21:54:04 CDT on October 6. Both full heap runs use `-timeout=0` and no race
instrumentation. Do not repeat them for unrelated semantic changes.

The explicit stress target compiles and lists both original methods. The ordinary
target omits them. Existing LSB/MSB factory and simple-fill tests pass normally in
3.459 seconds. Run the full methods with:
`go test -tags=tlc_fp_stress -run '^TestJavaLong(LSB|MSB)DiskFPSet_testMaxFPSetSizeRnd$' -timeout=0 -v ./tlc`.
OffHeap random, all three sequential contexts and OffHeap multiple-flush
reconciliation remain pending.

## User-visible behavior already completed

Direct TLC arguments, `modelcheck` and `mc` use the same TLC runner and Java
flags. Do not restore `--tlc` or the bounded checker's CLI path. Help in
`cli_help.go` explains flags, defaults and Java/Toolbox correspondence. Packaged
models load their properties and files through the packaged resolver. Ordinary
and packaged loading share deferred tool construction after intern recovery.

Configuration I/O failures exit with source status `255` and omit FINISHED and
trace generation. SANY tool markers and STARTING retain source phase order.
Unexpected parsing Exceptions become checked failure; semantic Exceptions chain
through FrontEndException and TLC_PARSING_FAILED2; Java Error propagates. Delayed
SANY output releases only on the unexpected checked-exception path. See
`TLC_ARCH.md` and `PORT_PROGRESS.md` for details and remaining diagnostic gaps.

CommunityModules progress changes are committed as `840d494`. Under `go test -v`,
its subprocess output, phase progress and heartbeats remain visible. Remote check:

```bash
go test -v -count=1 -timeout=60m -run '^TestJavaCommunityModulesAnt$' .
```

The earlier model-config failure comparing `TRUE` with `"blue"` is fixed and
covered by the full suite. Reopen it only if new evidence warrants it.

## Test-port inventory

[TODO_TEST_PORT.md](TODO_TEST_PORT.md) is the authoritative class and method
inventory. Its current totals are:

| Suite | Translated method contexts | Pending contexts |
| --- | ---: | ---: |
| Main TLC | 1,260 of 1,269 | 9 |
| Shared utilities | 55 of 56 | 1 |
| Long tests | 17 of 22 | 5 |
| Concurrent tests | 2 of 17 | 15 |

In the main suite, 617 of 626 non-ignored concrete classes are complete. A method
context is a concrete Java class plus a method; inherited methods count once
per concrete subclass. These totals measure translations, not implementation
coverage or universal behavioral parity. Original ignored methods are counted
separately. Do not credit scratch probes, fixture copies, reduced workloads,
configured stress audits, or Java failures as completed translations.

`DumpLoadTraceTest` is complete: all 32 enabled methods and the three original
ignored methods are translated in `tlc_dump_load_trace_java_test.go`. The whole
unchanged Java class and the Go class passed. The Go class also passed its
previous focused race run. Those class receipts predate the final counter-lock
correction; the normal workspace receipt above verifies that counter correction.
Historical Java replay failures under deliberately paced scheduling remain
valid evidence. Current passing runs do not establish stability under every
possible schedule.

## Remaining work and known blockers

Continue the core production audit and unresolved original methods below.
Current concrete production gaps include syntax-error lookahead and residual
stack reporting, generation traversal and multiple-binding context iteration
order, remaining constructor and semantic-node boundaries. The parser now reads
tokens lazily and catches actual lexical failures. It reports source
TokenMgrError text before ParseUnit's single E4003 abort, including EOF and
UTF-16 character details. Fifteen lexical/trailing-text observations match
Java. Comment scanning now retains Java's MORE/SPECIAL_TOKEN segments,
actual special-token kinds and backward links, shared signed nesting counter,
lexical state at EOF and checked failure behavior. EOF after a completed special
segment remains accepted by the token manager even inside an outer comment, as
in Java; an unfinished MORE segment fails. Syntax nodes retain chronological
pre-comment images. The parser wrapper emits the source multiline open-comment
message from actual scanner begin/end lines. All 200 bounded token-stream rows
and 28 state/parser-output rows match Java; these do not establish every lexical
rule or concurrent interleaving. PRAGMA now selects NUMBER, identifiers,
junctions and module headers by source longest match and token-kind priority;
its one-character SKIP wins ties with single-character identifiers. BAND/BOR
prefixes follow CASE1b/c, CASE2b/c and CASE6b/c, excluding underscores, lone W/S
and WF/SF prefixes. All 1,080 bounded stream rows match Java. SwitchTo rejects
invalid states with source TokenMgrError reason 2 and unchanged current state;
all 66 valid/invalid transition rows match. The generated Java DFA/NFA is now
translated in `sany_scanner_generated.go` and runs over the actual UTF-16
SimpleCharStream in `sany_char_stream.go`. Candidate-based token selection is
removed. All 10,392 character/operator rows match Java, resolving the 13 earlier
SPEC malformed-string and incomplete-prefix differences. All 16,247 scanner
state rows also agree on counters, raw image buffers and state-array hashes.
The stream mechanics separately match 11,452 bounded observations, including
buffer growth/reuse, backup, reinitialization, signed position wraparound and
line adjustment. These bounded comparisons do not establish every constructor,
arbitrary reader/exception boundary or concurrent interleaving. Production
module and debugger-dependency loading now decode UTF-8 with Java replacement
lengths, preserving columns for malformed prefixes. All 49,430 byte-decoding/token
rows and 6,166 parser-output rows agree with Java; seven direct file-loader cases
pass. Monolith fallback now calls the actual ported `MonolithModule` extractor,
reads and closes its NamedInputStream, and uses an explicit extracted-source flag.
It retains default-charset extraction before UTF-8 parsing, temporary-file metadata
and root provenance. All 108 bounded observations across UTF-8, US-ASCII and
UTF-16 defaults agree with Java; 27 prior US-ASCII differences are resolved.
General encoding constructors and I/O-failure paths remain audit work.
Another 300 ReInit/round-reset rows agree, including complete state arrays and
invalid lexical-state handling. Continue the remaining parser and canonical
semantic-graph audit. Preserve source error
text; lexical diagnostics use E1200 rather than heuristic error categories.

LET, operator/function-body and bound-expression generation now preserve source
context, label-parameter and function stacks when a body or domain throws.
Successful generation pops them explicitly; domains are generated inside the
new empty context, before quantified formals are introduced. LET also retains
the source level guard, array-bounds boundary and lower-level clamp. All 4,256
bounded state/diagnostic observations across 23 cases agree with Java, and the
247 earlier LET graph rows remain exact. These receipts cover the observed body/
domain failures; constructor-abort paths and runtime LET context sharing still
need source reconciliation. ASSUME/PROVE now retains the source 100-entry
`inScopeOfAPDecl` array, signed 32-bit depth and successful-path context/depth
cleanup. Label restrictions scan the active declaration scopes from depth 2,
as Java does. All 1,077 bounded state observations across 21 cases agree,
including nested/domain failures, array bounds and signed endpoints; the earlier
225 AP graph rows remain exact. Other generation boundaries remain audit work.

Ordinary operator generation now retains its label scope after popping the
formal context, through OpDef construction, registration and recursion-field
assignment. Only normal completion pops the scope and attaches its label table.
The early label-table copy on the source expression is removed. Twenty external
observations across 10 direct source `processOperator` cases match, including
retained registration and labels when recursion-field assignment throws. The
247 normal LET graph rows and 4,256 body/domain failure rows remain exact.
This does not complete canonical evaluator integration or all constructor paths.

The TLC bridge now caches an actual canonical LET's runtime adapter before
converting children and uses original local Definition pointers. Declaration
formals are adapted once by canonical identity for definitions, quantifiers,
CHOOSE, functions, LAMBDA, comprehensions, INSTANCE parameters and selectors.
Original Test206/Test209 initially exposed captured-formal failures; these are
fixed in production. Temporary native selector LET wrappers have different
bodies and no longer claim the original LET's canonical identity. All 40 focused
original methods and 19 coverage methods pass; seven external Java/Go identity/
evaluation observations agree. Context transfer, full semantic-base sharing and
remaining native selector reconstruction still need implementation.
Canonical formal adapters now retain the actual `SemanticNodeBase` pointer,
including UID and indexed tool objects. Standalone runtime formals allocate
their own base; adapter paths do not allocate throwaway formal nodes. Runtime
OpDef symbols retain their definition's base, and lookup checks indexed symbol/
body objects for the active tool. Config constants, constant pre-evaluation and
native/module overrides now store values in the active tool's indexed slots.
Actual declared constants retain their canonical semantic base. Symbol lookup no
longer falls back to generic definition/body caches; 192 bounded Java/Go lookup
observations agree across tuple, numeral and string bodies. The config/native
and coverage gates pass. The broader run exposed original Test219's nested
INSTANCE prefix failure: the check double-charged earlier parameters. Java's
remaining-arity counter is now translated and Test219 passes unchanged. Across
the two original-model runs, all 601 selected top-level Go methods are accounted
for: 596 pass and five retain source ignores/assumptions. Full SANY and all 33
related selector/scoped/legacy executions pass. This composite receipt does not
establish a green full workspace; the known native XML failures remain.
The producer/arity changes are committed as `57274fb`. The current
literal patch reads indexed slots through worker muxing and preserves fresh
constant-value identity. Record/EXCEPT fields and selected scalar bodies retain
actual source nodes. All 69 bounded Java observations and 72 selected original
model/coverage executions pass; full SANY and compilation pass. Additional
original Test216 exposed missing module theorem initialization. The bridge now
retains the actual theorem statement vector and shared statement bases, and
constant processing visits it after assumptions. Test216 and the expanded
original model/coverage selection pass unchanged, as does the original SANY
package. The native XML fixture failures remain. Base-less aliases, complete
runtime ASSUME/PROVE definition bodies and remaining canonical Context/wrapper
adapters are still incomplete; this is not full lookup/graph parity.

The literal/theorem changes are committed as `5428acc`. The record-field
slice removes direct literal/generic-cache reads from record constructors and
record sets, and removes field re-evaluation from record selection. These now
follow Java's indexed-slot casts and worker muxing. Record selection preserves
the detailed failure reason and expression/context; CounterExample retains
Java's RecordValue subclass dispatch. All 102 bounded source comparisons agree.
The original RecordValue, SetOfRcrdValue and value-stream checks pass. Current
run details and remaining gates belong in PORT_PROGRESS.md. Canonical runtime
Context transfer and ASSUME/PROVE definition bodies remain next bridge gaps.
The record-field changes are committed as `6c91722`.

The current operator adapter slice retains the actual SANY semantic base rather
than allocating a separate UID/tool-slot identity. Runtime symbols share the
same base, and definition shells are cached before body adaptation. Actual
canonical-node caches preserve aliases and debugger dependency reuse. Original
model/debugger/coverage and SANY gates pass; 190 inspected explicit definitions
in existing fixtures retain their base, UID, syntax, location and indexed slots.
This prepares canonical LET Context transfer, which is still pending. Preserve
source Context Pair history and Hashtable buckets rather than reconstructing
them from filtered `getLets` definitions; module-instance and theorem entries
must not disappear. Full run receipts belong in PORT_PROGRESS.md.

The operator identity changes are committed as `6336b72`. Context transfer
infrastructure now preserves separate source Pair history and Hashtable bucket
chains, with exact runtime null-duplication failures. All 1,381 bounded Java
Content/lookup/null observations match. Full LET bridge wiring remains pending;
the user's SANY test fidelity audit takes priority before that work resumes.

Recursive declarations now update the actual node and unresolved counters inside
`endRecursiveDefinition`, before label-scope completion. Canonical named functions
also complete there rather than during native preregistration. Thirty direct Java
`processOperator` observations agree on normal completion, counter underflow and
invalid-level failures. The earlier LET, named-function and body/domain graph/
state observations still agree. Native paths without actual declaration nodes
remain explicitly incomplete; broader generation and runtime sharing need work.
Replacing completion syntax with null now clears the actual tree interface,
native position and source location. A null declaration throws the source
NullPointerException family. All 32 direct source completion observations agree.

The Java `belchDEF` token-stream operation and its production call sites
are now ported. Definition recognition requires the inserted marker, and
`DefStep` leaves it for the definition parser as Java does. Twenty-one scratch
comparisons match actual non-EOF token kinds, images, positions and marker
placement. This is not an EOF-position or full parser-parity receipt. The actual parser now maintains module-production message frames and expecting
state, and throws a typed ParseException at the failed footer consumption.
Its source message comes from that real failure state, separately from native
text; the loader reports it before the existing E4003 abort. EOF positions
retain Java's last-character coordinates, including empty input, CRLF, tabs
and UTF-16 text. Twelve EOF observations match. The missing-expression and
missing-footer observations now match complete Java front-end output; fifteen
lexical observations still match. The expression parser now follows the source prefix sequence, operand,
postfix-extension loop and optional recursive infix continuation. It leaves
unrelated following tokens for the enclosing production. Operator arguments
and substitution values use `OpOrExpr`, with source operator-token alternatives
and lookahead; lambdas belong to that argument production. Ordinary identifier
operands retain source `N_GeneralId` wrappers. The semantic bridge preserves
unqualified Boolean literal translation through those wrappers. A failed
bracket consumption raises the actual typed failure with its real production
frames, and LOCAL retains a single Definition frame. All eighteen established
lexical and syntax observations now match full Java front-end output.

Operator-stack failures now carry source messages and locations, and final
reduction preserves accumulated errors before the ordinary-constructor exception.
Record selection uses the dedicated reduction. Prefix reduction preserves its
raw operator node, including unary minus; original precedence assertions compare
the input symbols unchanged. Twenty-one direct stack observations match Java.
All thirteen additional full-front-end observations now match after actual
semantic lookup: undeclared `-.` and `^+` fail before generating operands.
Twenty-two further observations match declarations, aliases, local operator
scopes and diagnostic ordering. Raw operator names remain in messages; lookup
alone resolves aliases. Both missing-operator messages retain Java error code
4004 (`SUSPECTED_UNREACHABLE_CHECK`), preserving message-control behavior. Built-in arity comes from the actual initial context,
including the binary `\times`, rather than parser metadata or XML projection.

Packaged module loading now uses the existing byte-exact Java standard modules,
CommunityModules, TLAPS modules and the single Apalache module source. Abbreviated
module bodies and synthesized arithmetic exports are removed. Genuine declarations
and source locations drive imports. This is source loading, not a claim that every
module runtime override or semantic construct is complete. Native fixture corrections
retain their assertions and are checked against Java; original Java fixtures are
unchanged. The exporter sorting check starts with Java-valid source, then deliberately
reverses its IR definitions before checking the unchanged output-order assertion.

Expression generation now uses contexts built by visiting actual module-body
and LET syntax nodes in order. Ordinary definitions and named facts become visible
after their bodies; `RECURSIVE` names become visible at their declarations. Function
domains precede their temporary self-binding, which is available in the body.
Selector preparation uses the same visibility boundary and preserves unfinished
recursive definitions. Selecting their bodies reports source code 4005 at the
named token. Failed selectors retain the null-operator placeholder without a
second undefined-name error. Unresolved symbolic expressions retain Java's distinct
null result, including function-application argument suppression. The existing
arity and operator-argument checks do not inspect expressions that failed generation.
Seventy-four bounded observations match Java's actual semantic error counts,
codes, ranges, messages and ordering; this does not establish full Generator parity.
Receipt: `/mnt/oldrog/tmp/tlago-source-visibility-final-audit-corrected.log`.

Module generation now dispatches actual body syntax units rather than grouped
AST categories. Selector preparation runs within each unit. Declarations,
instances, definitions, assumptions, theorems and proofs are visited in source
order; recursive-section checks run before the unit body, and unfinished
recursive operators are reported in declaration order. Invalid operand selectors
retain Java's error code 4005, individual selector range and original message.
Twenty additional source-order and recursive-section observations match Java;
the established 74 visibility observations still match on this snapshot.
Receipts: `/mnt/oldrog/tmp/tlago-module-unit-order-audit-corrected.log` and
`/mnt/oldrog/tmp/tlago-module-unit-final-visibility-audit.log`.

Nested modules are now generated at their actual body units. They share Java's
module recursive counters, including the source exception on invalid input where
an unfinished recursive section spans a nested module. `Spec.SemanticDiags`
retains semantic errors independently of successful generation; parent errors
precede child errors even when an exception interrupts generation. The original
semantic-error corpus helper's `WrongInvocationException` catch is restored,
without swallowing other exceptions. Twenty-eight bounded source comparisons
match Java's diagnostics, exception type and message; established visibility
observations also remain matching. Receipt:
`/mnt/oldrog/tmp/tlago-nested-unit-final-audit-28.log`.

The semantic-error corpus now retains all four original assertion families:
failure severity, diagnostic parameter counts, absence of suspected-unreachable
checks and presence of the expected error code. The original helper returns
parse diagnostics when any are present, including warnings, and retains the
source exception catch. All 66 ErrorCode metadata entries are ported. Production
diagnostics retain actual Java arguments rather than padding parameter lists.
Invalid label selection reports source code 4337 at the label token; failed
selection suppresses duplicate label checks. INSTANCE operator arguments remain
operator arguments even when their arity mismatches. Duplicate module conflicts
are generated at their module unit instead of by an earlier approximation.

The fixed-parameter diagnostic comparison matches Java in all 121 unchanged
primary error-corpus fixtures. Four supporting modules are loaded as dependencies,
not additional primary fixtures. Earlier receipt prose incorrectly said 126;
the actual original fixture selection and comparison contain 121. Receipt:
`/mnt/oldrog/tmp/tlago-proof-statement-final-parameters.log`.
This checks fixed-code order/count and displayed parameter values; it is not a
complete diagnostic message/range/type comparison or proof of full semantic parity.

Module recursive functions now preserve rejected declarations, their original
operator arity and undefined bodies. Function domains are generated before
recursive arity validation, followed by the body. INSTANCE declaration-level
matching uses Java's `ModuleNode.isConstant` rule: variables, operator bodies and
EXTENDS theorems, including local operators but excluding instantiated theorem
context definitions. Declared-operator operands preserve their levels and level
parameters. INSTANCE errors retain source range, message, arguments and order.
Twenty-eight bounded Java/Go observations match full diagnostic output, and all
74 established visibility observations remain matching. These are scratch
observations, not invented tests or test-port credit. The exposed empty operator application discrepancy is now corrected: optional
arguments use the source two-token lookahead, and `OpArgs` requires its first
argument. Invalid argument starts fail before entering an expression; actual
argument and bang-extension frames retain source residual-stack output. Twenty-two
bounded parse-message comparisons match Java, including empty calls, nested and
qualified calls, commas, missing delimiters and valid argument forms. The 196
accepted lookahead token kinds match the pinned generated Java parser. These
checks do not establish complete lookahead or parser parity.

Restricted expressions and fairness now use Java's identifier-only subscript
production, detached final arguments and shared fairness hook. The original
five-heir fairness node replaces call-style reconstruction. Lexing excludes
`WF_`/`SF_` from ordinary identifiers, and no parser-side identifier fabrication
remains. Fairness/action builtin level maxima check children first and suppress
redundant parent errors. All 42 bounded parse/tree/semantic observations match
Java. One handwritten native lexer expectation is corrected to Java's exact
tokens with its input and full assertion intact. Original Java tests and fixtures
are unchanged. Empty-node location sentinels and complete semantic-level graphs
remain separate audit work; receipts are in `PORT_PROGRESS.md`.

LET generation now carries declaration identities, levels and shared module
recursion counters through actual expression traversal. Its declaration vector
survives lexical scope exit. The level decreases before IN generation, while
the LET symbol context remains visible. Rejected nested definitions preserve
unfinished outer bindings; functions follow domains, validation and body phases.
LET selector diagnostics are reported when generation reaches their expressions,
and retained bindings supply application arities. All 29 bounded lower-phase
comparisons match Java errors and recursive exception boundaries. Existing
native LET acceptance source now uses distinct inner/outer operator names,
because Java rejects the prior shadowing fixture; assertions remain unchanged.

Module operators now retain the first declaration/definition and its signature.
Duplicate validation precedes the body; the new operator constructor's conflict
follows it, as in Java. Function constructor failures preserve whether the bound
context is pushed. Function domains are generated once per syntactic group,
then bound names, before validating the definition's symbol. Numeric/string leaf
selectors and bound-name collisions retain source diagnostic details.

Incomplete module-instance names validate arity before generating operands.
Expression operands retain arity errors; operator operands follow receiving
formal arities and source expression/operator/lambda failure branches before
incomplete-name validation. Preserve Java's GeneralId operator-argument behavior:
attached expression arguments are not generated on that path. All 50 bounded
full-message comparisons match Java, including warnings and ordering. The 74
visibility and 29 LET comparisons remain matching. No original tests or fixture
bytes changed, and scratch probes earn no inventory credit.

Proof projection now retains BY facts, DEF entries, MODULE entries, qualified
step references and DEFINE bindings. Reference generation preserves source
fact/expression mode, argument errors, rejected DEF entries, Empty BY and
statement-versus-SUFFICES NEW visibility through nested subproofs. All 49 bounded
semantic-phase observations match Java's errors, warnings, ranges, messages and
ordering; every source parses successfully. The existing 121 fixed-parameter,
74 visibility, 29 LET and 50 symbol/operand comparisons still match. Scratch
observations do not add test-port credit.

Proof statements and DEFINE bodies now generate in lexical order. DEFINE shares
operator/function phases with LET, retaining first bindings and recursive identity.
Domains precede function construction; an absent definition-vector entry preserves
Java's no-message ArrayIndexOutOfBoundsException and earlier diagnostics. PICK
names are visible in its formula, hidden during its whole subproof, then installed
for following steps. Pseudo-expression errors now arise at actual expression
selection, before later BY errors. Formal operator arities and declaration locations
are retained. Higher-order operands use receiving-formal generation followed by
operator-constructor matching; existing level constraints remain checked.

All 58 further comparisons match: 57 complete semantic diagnostic observations
and one lower-phase exception/type/null-message/retained-diagnostics observation.
Logs use `/mnt/oldrog/tmp/tlago-proof-statement-final-*.log`. The existing native
proof-scope fixture is corrected against Java: PICK keeps I in scope, so a later
quantifier must use a fresh k. Its named <1>I collision and assertions are unchanged.

Mixed DEFINE steps now retain their syntax and visit operator, function and module
definition heirs in Java order, in both expression generation and selector preparation.
Four bounded source comparisons match, including a rejected forward instance reference.
Existing 58 statement and 49 reference comparisons still match. Complete SANY passes
in 1.881 seconds; the existing focused gate passes in 13.476 seconds. These checks
add no permanent test inventory credit.

Non-local INSTANCE proof steps are now projected and generated. Proof-local WITH
clauses generate before imported bindings: implicit defaults, explicit operands,
duplicate detection, implicit arity checks, then missing substitutions. Illegal
targets omit operand generation. Failed operator operands retain Java's nullOpArg
and subsequent zero-location arity error; lambda body errors preserve lambda arity.
Module-definition formals are scoped during substitutions. Conflicting imports
retain the first binding and its arity. DEF on an INSTANCE step emits Java's
non-definition error. Named-instance prefix arguments are checked at their actual
selector component. Context.getByClass follows Java Hashtable enumeration, not
reversed insertion links: keep builtin and unrelated definition entries during
rehashing, then filter declarations. Import operators first and theorem/assumption
symbols afterward, in Hashtable order within each class. Theorem conflicts retain
the original source location. All 31 INSTANCE and four additional context-order
observations match full diagnostics, including inherited declarations and bucket
changes. Scratch comparisons add no permanent test credit. Existing SANY passes
in 1.433 seconds and the focused gate in 14.323 seconds. Receipts:
`/mnt/oldrog/tmp/tlago-proof-instance-context-final-*.log`.

Operator operands now reject arguments on a named selector prefix before generating
its arguments or resolving later name components. This shared selectorToNode
boundary supplies both higher-order calls and INSTANCE substitutions. Nine symbolic
observations match Java, including an unknown prefix, ignored invalid argument
expressions and subsequent constructor/substitution errors. Existing 31 INSTANCE,
58 statement and 50 symbol observations still match. Complete SANY passes in
1.838 seconds; focused existing classes pass in 13.960 seconds. Receipts use
`/mnt/oldrog/tmp/tlago-proof-instance-symbolic-*.log`. No inventory credit.

Proof assertion `@` references now retain the preceding infix RHS per proof depth,
matching generateProof's `$Nop` reuse without repeating prior diagnostics. Nested
proofs keep independent history; non-assertion and ASSUME/PROVE steps reset it.
SUFFICES expression behavior follows the source branch. Undefined `@` retains
Java's exact diagnostic and nullOAN result. Existing SANY passes in 1.894 seconds;
focused classes pass in 14.423 seconds. Receipts use
`/mnt/oldrog/tmp/tlago-proof-at-*.log`. Complete proof-level/graph parity remains
unproven; these scratch observations add no test inventory credit.

After the full-run failures are resolved, continue proof INSTANCE inherited-context/constructor identity and level checks,
original hierarchical proof level checks,
complete typed NEW/bound contexts and constructor/application failure boundaries.
The current generation path does not establish complete proof-graph parity. Broader
module namespace resolution, formal/bound contexts, general operator-argument
generation and application failure boundaries still need reconciliation,
including compound selectors and LET instances. Preserve original assertions
and diagnose implementation shortcuts before installing more original tests.

JavaCC expected-token entries now agree on the bounded observations documented
below; exhaustive production states and label error continuations remain work. The original `ParseErrorTests.testAll` lives in root
`sany_parse_error_java_test.go` and uses the shared actual parsing phase. It
asserts recorded parser output with the original input and assertion text,
replacing the semantic-diagnostic surrogate. The original percent-error output
assertion now checks `token "%"` in actual recorded output. The three existing
`TestSanyOutputFormatting` methods use production `SimpleSanyOutput` and the
actual in-memory syntax parser, retaining their original input and assertions.
Simple/Silent/OutErr routing, level ordinals, verbatim messages without arguments,
platform line separators and applicable Java string formatting are ported. The
parser emits TRACE at actual production entry/exit and reports caught errors
inside its parse wrapper. Production exits do not run while propagating exceptions.
Restored frames retain Java's original spelling and LET production name.

Constant/recursive declarations use `ConstantDeclarationItems`, its two-token
argument lookahead and typed operator leaves. Formal operator declarations retain
their own frame and error wording. INSTANCE's two-token body lookahead and later
keyword reclassification match source timing; subsequent substitutions require
three-token comma/target/arrow lookahead. The exact 177-token target set includes
Unicode operators and excludes '.'. LAMBDA uses direct identifier parameters;
semantic translation consumes that tree. Its E4274 arity error uses the operator
name range, independently of the subsequent full-call E4271 diagnostic.

CHOOSE, optional domain binding and identifier tuples retain their source frames,
expectations and mandatory-token failures. Preserve empty tuple syntax and Java's
`processChoose` formal-count arithmetic: `<<>>` creates a tuple formal named `>>`.
Do not replace it with a scalar binding or a zero-formal tuple. Direct lower-level
Java/Go TLC probes match tuple flags, names, selected values and an invalid-domain
error prefix; they do not invoke Java TLC reporting.

Quantified forms use Java's identifier-list/colon lookahead, mandatory bounded
lists and actual production frames. Temporal quantifiers retain the source frame
name `Bound Quantified Expression` and map to `$TemporalForall`/`$TemporalExists`.
Generate every domain of one bounded source quantifier before introducing any
formal; explicitly nested quantifiers remain separate scopes. Arity and operator
argument checks retain that ordering, and conflict diagnostics use the previous
symbol's actual source range. Empty tuple bounds preserve Java's one `>>` formal.
ContextEnumerator reports the source coded `TLC_ARGUMENT_MISMATCH` runtime failure
for a tuple type or length mismatch.

Primitive expression parsing preserves Java's mandatory delimiter failures and
one-token expression lookahead for optional tuple elements. String syntax is a
leaf with decoded escapes and retained quote marks; semantic translation strips
only the surrounding marks. The existing native parser expectation now follows
that source image. Original Java assertions remain unchanged. Generic expression
operators are typed token leaves, and syntax location aggregation retains Java's
integer extrema for empty nodes. Junction list/item frames follow actual grammar
entry and exit. The source junction context retains list type and column, matches
new bullets without an added line predicate, and is terminated only on normal
completion. Indentation checks traverse descendants, stop at nested junction lists,
and throw Java's exact exception after leaving the item frame. General JavaCC
lookahead and surrounding expression grammar remain reconciliation work.

Record fields and EXCEPT paths retain `Field Value`, `Field Set`, `Except Spec`
and `Except Component` frames. Each EXCEPT path requires at least one component;
index components require an expression. Preserve the source `= or ,` expectation,
typed equality token and keyword field reclassification. `!.@` records an error
without throwing, then parsing continues so later errors retain their source order.
Shared `Identifier` consumption now throws immediately instead of returning nil.
Brace forms now use Java's function preview and immediate identifier/tuple
comma/colon lookaheads. Preserve source membership reconstruction, typed IN,
subset expectations and the explicit complex-membership comprehension error.
Do not search ahead for a later colon to choose the form. Square-bracket parsing now
uses Java's function preview and field lookaheads, followed by one mandatory
expression and its actual continuation token. Function bounds, record fields,
applications, function sets, EXCEPT and action forms retain source delimiters and
failure boundaries; no search ahead for the final form's separator remains.

IF/CASE/arm parsing retains Java's original frames and mandatory token failures.
CASE consumes ordinary arms before its optional final OTHER arm, with separators
strictly right of the active junction alignment. LET requires one or more actual
LOCAL/DEFBREAK/RECURSIVE definition tokens, then mandatory IN; malformed tokens
are not skipped to reach IN. Function application reduction places the function
in zero and the bracket node's heirs in one, matching source `reducePostfix`.
The translator no longer compensates for a fabricated application wrapper, so a
real nested function application remains an argument.

Definition parsing now preserves Java's expectation state and mandatory tokens
for function bounds, operator bodies, formal lists and higher-order declarations.
An operator parameter list requires its first formal, including nested underscore
lists. Infix head selection uses the source two-token lookahead; malformed right
identifiers fail inside `Infix LHS`. Module headers throw at a missing separator,
and marked malformed definitions enter their production instead of being skipped.
Module bodies use the source one-token entry lookahead and two-token unit
selection, including the `USE ONLY` exclusion. Definition lookahead validates only
the tokens within its budget, including in proof definition lists. Assumptions
retain their own frame, optional named-head marker and source expectations;
variable/EXTENDS lists and theorem statement failures also retain their source
expectation state. Hierarchical proofs require numbered QED steps and a nonempty
sequence of definitions in `DefStep`. `Step`, `QEDStep`, `DefStep`, `HaveStep`,
`CaseStep` and `AssertStep` retain source frames and failure expectations. Invalid
levels and forbidden nested proofs throw immediately. Numbered step nodes retain
both original and corrected images; generation checks the original image for
illegal named implicit steps. Proof-level stacks pop only on normal completion.
`UseOrHideOrBy` now owns the shared source command grammar and frame. It requires
an item after each comma and `DEF`, permits `ONLY` only for USE/BY, and retains
flat MODULE/identifier children. Step references in expressions preserve source
outside-proof and `<+>` failures; implicit references use the enclosing level.
`StructOp` retains its frame and rejects real selectors after normal frame exit.
Operator and symbolic structural selectors are typed token leaves. Projection
retains empty commands and generation counts appended entries, including failed
facts, before reporting source empty-BY/USE/HIDE diagnostics.
TAKE and PICK use source full identifier-list previews, mandatory choices and
shared `QuantBound` nodes. Unbounded PICK uses identifier leaves, without
fabricated declaration wrappers. TAKE/PICK/WITNESS retain source frames and
expectations, including missing items, commas, domains and PICK's colon/body.
Remove the separate proof-bound scanner and its unused search helper.
Assume-Prove retains its source frame, mandatory assumptions/PROVE/body and
expectations. Nested labels use the exact identifier/`::`/recursive production,
without the former broad label-name scan. Proof nesting admits 100 levels and
rejects 101 before entering a new Proof frame, matching Java.
Numeric proof-step levels now use Java's signed 32-bit conversion. Overflow
raises `NumberFormatException` with the source message instead of accepting a
64-bit level or silently treating it as an unrecognized token. Thirteen source
boundary observations match, including leading zeros and implicit levels.

Expression labels are formed after parsing the primitive and postfix extensions.
Their source shape is validated at `::`; parameterized labels retain an
`N_OpApplication`. The operator stack removes the label operand before parsing
its body, then uses source last-operator state for the precedence check. Projection
reads the source callee's final name and arguments instead of fabricated wrappers.
NEW declarations now retain source frames, expectations and actual two-token
JavaCC alternatives. Reject an operator declaration with arguments before `\in`
at the source boundary; its ordinary domain token remains an `IN` leaf. Theorem's
three-token Assume-Prove selection now runs the actual generated scanner. The
mechanical generator now retains all 268 source scanner methods and predicates
against the scanner position, junction context and active operator stack. Saved
calls expire in source order and rescan actual error token sequences, including
earlier successful calls. All 599 complete TRACE/results and 137 selected raw
module trees match Java. The temporary bounded vocabulary matrix also matches
all 609,175 scanner verdicts for the six entry points connected at that snapshot.
All 74 entry points are now callable, and optional argument parsing uses actual
source calls 50, 68 and 72 with budget 2. Another 4,440 external observations
match results, remaining budgets and current/farthest positions across all 74
entry points in default parser contexts. Related original frontend/model checks
pass in 9.878 seconds; complete SANY passes in 1.979 seconds and compilation passes.
Module Body now runs source calls 1–5 in Java's decision order. Definition heads
run calls 8–11 with Integer.MAX_VALUE, and Identifier-LHS definitions choose the
body with call 7 only after consuming DEF and running belchDEF. Native early
module-instance preview and manual definition-head error-span retention are
removed. The actual generated scanners now supply those saved rescan calls.
Another 266 external production rows agree on syntax kinds/images/coordinates
and complete malformed-input messages across 23 cases. Existing original frontend
and model checks pass in 9.602 seconds, complete SANY in 1.780 seconds, and
compilation passes. Declaration/fact parsing now uses source calls 6, 12, 13, 14, 15, 21 and 23
with their actual budgets. Constant operator parameters, repeated WITH
substitutions, optional fact names and first/subsequent ASSUME/PROVE expression
clauses retain actual saved calls rather than manually estimated failed spans.
Proof alternatives and commands, definition repetition, TAKE/PICK selection,
assertion expressions and decimal preview now use source calls 24–37 with Java's
budgets and saved-call order. Proof consumes PROOF only on its source preview;
failed terminal/structured alternatives retain the actual source token position.
Obsolete proof-bound/identifier scanners and decimal failure-span estimates are
removed. Operator arguments, quantifiers, sets, brackets and tuples now also use
source calls 38–49 with Java's budgets, predicate order and saved-call indices.
Set parsing previews its expression before the function-head predicate; bracket
parsing preserves the keyword-field reclassification choice after record preview.
Unused native previews and their manual error-span estimates are removed.
Fairness uses source call 51 (budget 2). Junctions start their indentation context
before source disjunction/conjunction calls 52/53 (Integer.MAX_VALUE). Initial
expression prefixes and open expressions use calls 54/55; infix right operands
retain distinct calls 62/63, all with Integer.MAX_VALUE. Initial extendable
operands use junction preview 56 before the operator-stack predicate, then
primitive preview 57 (budget 1) after direct parenthesized-form choices.
Expression extensions now use the source loop preview 58 (budget 1), then
postfix, record-field and function-argument calls 59–61 (Integer.MAX_VALUE)
before their column predicates. Continuation preview 66 (budget 1) precedes
infix/label calls 64/65 (Integer.MAX_VALUE). Predicate-failure branches retain
Java's empty expected-token list rather than estimating a following identifier.
PrimitiveExp now previews String and Number with source calls 69/70 and their
column predicates, then preserves direct identifier/infix/postfix choices before
nonexpressive-prefix preview 67 (all Integer.MAX_VALUE). BangExtension uses 73
(budget 1), then identifier preview 71 (Integer.MAX_VALUE) with the `@` exclusion,
operator alternatives and argument preview 72 (budget 2). Its other branch keeps
direct OpArgs before structural preview 74 (budget 1).
All 74 source lookahead entry points now have production callers. JavaCC's 130
direct-choice expectation masks are now generated from upstream, and all 130
source sites have translated paths, including the guarded switch failures.
Expectations expire by token generation, are ordered by token kind, and precede
saved-call rescans and their duplicate checks. Successful
consumption now performs Java's 101-token cleanup of expired lookahead references.
CompilationUnit now parses the source Prelude's identifier/number sequence and
resets the tokenizer to DEFAULT only after successful module parsing. Module body
choices retain Java's nested direct-choice recording order before saved previews.
Definitions, identifier tuples, formal parameters, INSTANCE substitutions and
assumptions now record source sites 19–37. First and subsequent formal parameters
retain distinct expectation sites 25/27 in the shared helper. These converted
failure branches use source masks rather than native estimates. ASSUME/PROVE,
new-symbol declarations, optional bounds, theorem keywords and terminal proofs
now also record source sites 38–54. First/subsequent clauses retain sites 40/42;
state/action/temporal formal choices use their separate site 50. USE/HIDE/BY,
fact and DEF lists, proof-step choices and TAKE/WITNESS/PICK/SUFFICES now record
source sites 55–76. Shared fact parsing receives each actual source site rather
than merging ordinary and DEF-item expectations. Expression arguments, quantifiers
and set forms now record source sites 77–90, 92–96 and 98. Shared set loops retain
their actual caller sites, and open/parenthesized expressions reject token kinds
outside their Java productions. Brackets, EXCEPT components, tuples, restricted
expressions, fairness, LET, junction items, CHOOSE and lambda now record source
sites 99–118. Identifier and keyword record fields retain distinct sites 100/101.
The earlier 641 state observations remain qualified evidence; all 5,823 current
parser rows across 341 cases agree. Operator-token productions and OpenStart
also match Java on 1,475 observations over all 295 token kinds, including
acceptance, consumption and expected-token lists. Prefix operands preserve the
Infix Op frame for unary minus. Prelude, set continuations and proof lexemes
retain their guarded source failures.
The obsolete identifier-definition expression preview and its manual span map
are removed. Generated exceptions retain the actual consumed token and expected
sequences; short messages use only the longest generated sequence. Junction
failure contributes no invented bullet tokens, and formal declarations require
their actual caller site. All 341 external cases also match Java's ordered
expected-token entries, including 14,741 sequences in 117 nonempty results.
Generated exceptions now own Java's full getMessage and short-message
formatters; ordinary exceptions return their supplied message. Full messages
escape following tokens and list alternatives; short messages escape only the
prior token. Both match Java across 2,400 external formatting observations.
The 341 expected-entry and 5,823 syntax/message rows remain equal. Current
original frontend/model checks pass in 9.867 seconds, complete SANY in 1.967
seconds and compilation passes. Successful consumption now advances generation
and cleanup for EOF too; subsequent token requests retain distinct EOF identities
and linked tokens rather than clamping to the last token. Token-manager and
pretokenized paths each match 330 Java consumption-state observations, including
cleanup at token 101. All 641 existing direct-choice/saved-call state rows remain
equal. EndModule uses the generated consumption path; unused native consuming
helpers are removed. Saved lookaheads now retain Java's actual current consumed
token instead of the next token. A persistent initial dummy token is shared by
lookaheads and generated exceptions; rescans start immediately after that saved
token. All 1,998 saved-reference observations across 74 entry points and 18
repeated-exception observations agree with Java. Exception generation now
uses the failed token kind, active masks and saved rescans exclusively. The kind
is cleared after generation; arbitrary expected-sequence injection is removed.
All 4,144 generation observations match Java, including repeated calls and mask
unions. Generation alone retains the current token without fetching a successor. Exhaustive malformed rescan contexts,
syntax AST parity and proof generation remain reconciliation work.
These bounded observations do not establish every semantic-predicate context or
full parser parity.

Semantic symbols now embed TLC's `SemanticNodeBase` and use its shared
`NewSemanticNodeBase` constructor. This corrects the separate SANY counter added
in `1470b93`: Java uses one UID counter across all semantic subclasses. UID,
kind and hash access delegate to the common base. Assigned UIDs remain stable
when the signed counter wraps, including UID -1; assignment state is separate
from the stored UID bits. Existing Go zero-value base support remains separate
from normal eager source construction. Twenty-six bounded source observations
match shared allocation, hash/kind and tool-slot operations, repeated UID reads
and signed boundaries. Indexed tool objects now belong to each base node, as in Java, rather than a
global UID/hash map. SANY and TLC use the same per-node slots. Sparse growth,
null writes, retained array length and negative-index exceptions follow source;
distinct nodes retain separate slots even when UIDs collide. Thirty-five source
slot observations match. Tool IDs now use Java's signed 32-bit width throughout the cache APIs.
FrontEnd's allocator starts at zero and wraps as a Java int. TLC retains one
static ID across Tool construction and spec processors, matching Spec; creating
a new tool does not allocate another cache namespace. A maximum positive slot
write raises the source NegativeArraySizeException before changing storage.
Thirteen exact ID observations and the sixteen prior slot observations match.
TLCEval now stores its converted constant value on the expression's indexed
tool slot, as Java does, replacing its global UID-keyed map. Read pre-existing
values through WorkerValue muxing, then preserve the source read/write-lock
recheck and cache write. Worker selection follows the current worker, defaulting
to zero outside a worker; invalid indices raise source bounds exceptions instead
of silently selecting zero. Thirteen exact cache observations and six unchanged
original model methods pass. TLCCache's constant path now also owns a HashMap
on the expression's tool slot. Its class-wide reentrant lock permits nested
calls; hash filtering and lookup-key equality follow Java, including collision
tree bins. Seven unchanged original cache/extension/DOT/coverage models pass.
Source comparison includes mixed keys, equal fresh objects, UID collisions,
recursive cache calls, cast cleanup and 32 colliding keys through actual tree
bins. All 27 source observations match; sixteen concurrent callers share one
cached value in the isolated short race probe, which passes in 1.039 seconds.
Random-enumerable seed changes now reset only the caller's generator, retaining
peer RNG streams and the independent predecessor-state scope. RNG and checker
paths now share IdThread's one current-state slot; source error-state reset clears
it before trace recovery. The unchanged RandomElement model retains all eleven
trace states. Thirteen exact source observations, six original models and a short
race check pass. RNG behavior and initialized predecessor now reside on the
saved JavaRandom instance. The factory captures default/BFS behavior once;
restoring a generator preserves its stream and predecessor marker. Setter get
hooks, explicit null and plain-Random exceptions follow source. Seventeen new
instance observations and the thirteen thread observations match; nine unchanged
original models pass. Thread-lifetime cleanup, unindexed APIs, formal-parameter
graph construction, concrete class equality, complete allocation order and wider
cache/worker semantics remain pending. Detailed receipts are in PORT_PROGRESS.md; no whole-superclass or
full-workspace parity claim.

WorkerValue demux now follows source global worker-count decisions, mandatory
deep normalization, seed replay and array ownership. Constant preprocessing uses
EvalControl.Clear through the source Tool overload, rather than the former
EvalConst shortcut. Nil evaluator/results throw source exception types. Fifteen
exact demux observations and eleven unchanged original models pass. Ordinary
lookup now selects worker zero outside an IdThread scope, independent of state
worker metadata. Eleven direct-source lookup observations match; context/body
results remain unmuxed. Indexed SymbolNode storage and the full source lookup
provider graph are still pending.

Label generation now follows source guard order and returns immediately for
forbidden labels. Allowed labels check the body before their own parameters.
LET definitions retain ambient EXCEPT and nested NEW restrictions while resetting
the required bound-parameter stack. Repetition and missing-parameter ErrorDetails
use source messages. Fifteen detailed source comparisons match; existing focused
parser/semantic checks pass. This does not establish complete label-node parity.

Definition and proof-INSTANCE parameter generation now allocates concrete
FormalParamNodes before binding-conflict checks. Definition bodies retain the
ordered parameter nodes; identifier occurrences retain the resolved formal.
Nodes share the SemanticNode UID allocator and own indexed tool slots, syntax,
source location and native module ownership. Sixteen source observations match.
LevelNode data, visitors, evaluator graph sharing and full source allocation order
remain pending. TAKE/PICK constructor progress is described below.

Bounded, unbounded and temporal quantifiers now retain newly constructed formal
nodes in source order. Domains generate in the enclosing scope before parameter
allocation; conflicting declarations retain the earlier binding. Body and nested
domain references retain their resolved nodes. Nineteen source node observations
and eleven complete diagnostic cases match; existing focused checks pass.
Quantifier arrays here are flattened native metadata, not complete OpApplNode
bound-group construction or level-check data. Full source allocation order,
including generation below rejected labels, remains pending.

CHOOSE now retains its constructed formals and uses the same binding-preserving
scope as quantifiers. Generate the domain before allocating formals; bounded and
tuple forms use identifier syntax, while the source unbounded scalar constructor
uses the CHOOSE token's syntax/location. Twenty-five node observations and
thirteen complete diagnostic cases match Java, including empty tuple behavior.
Existing focused checks and the earlier quantifier observations pass.

Function-constructor, LAMBDA, set-of-all and filtered-set expressions now retain
formal nodes after generating all domains in the enclosing scope. Body references
resolve actual bindings; filtered sets generate their predicate rather than the
derived native element. Thirty-three node observations and eighteen complete
diagnostic cases match Java, including domain-before-conflict error order.
Existing focused checks pass. Full bound grouping, level data, source allocation
order and evaluator sharing are still pending.

Named-function preparation now retains bound formals and the temporary self
formal, resolving the function name after introducing its bounds. Accepted bodies
reuse that context; rejected names follow the source body-scope rules. Recursive
declarations retain their existing binding rather than binding the temporary self
formal. Twenty-five node observations and fifteen complete diagnostics match,
including function-name/bound conflicts and builtin-name resolution. The temporary
creation record does not model the final nonrecursive OpApplNode's pruned array.
Full graph construction, allocation order and evaluator sharing remain pending.

TAKE and PICK now retain their constructed formal nodes, including rejected
declarations. Domains generate before parameters. TAKE installs accepted bindings
immediately; PICK uses them in its predicate, hides them from its own proof, and
installs them after that proof. Conflicts preserve the original binding and its
location. Twenty-three source node observations and seventeen complete diagnostic
cases match Java; existing focused checks pass. Actual TAKE/PICK proof-node
graphs and scoped allocation order are now compared below. Inherited level data
and allocation order across all graph features remain pending.

Label expression generation now retains resolved formal arrays after generating
its body. Existing formals are shared; each non-formal argument occurrence gets
its own dummy node using the argument's complete syntax and location. Twenty-two
actual source-node observations match; existing focused checks pass. Label
argument diagnostics now report each non-formal occurrence at its own argument
syntax before repetition checks. Repetition uses retained UID identity, so distinct
dummy nodes with the same spelling do not count as repeated formals. Twenty-two
argument/repetition observations across twelve scratch cases match Java.
Required parameters now use ordered formal groups from the current label scope,
removing matched UIDs in source sequence order. Rejected same-named nodes remain
distinct requirements. Quantifier domains are checked together before the group
enters scope; nested labels and LET definitions reset that scope. Fourteen required
parameter observations across sixteen scratch cases match; all sixteen now match
complete diagnostics. Extra parameters use Java HashSet membership, removal and
iteration to produce one aggregate diagnostic. The earlier nine-case diagnostic
comparison now matches completely. HashMap removal includes tree rebalancing and
conversion back to lists; 3,470 operation/iteration observations match Java.

Builtin operator construction now retains source phony FormalParamNodes for fixed
arity, including empty arrays for arity zero and nil arrays for variadic operators.
Each is a fresh zero-arity local with null syntax, unknown location and no module.
Builtin operators retain their own zero-kind builtin syntax and location. Correct
null-syntax formal construction to retain null rather than substitute `nullSN`.
All 668 constructor observations match Java across 72 builtins and 77 formals.
Full frontend entry now initializes/rebuilds the global context before parsing,
including failed parses, and retains that context on the spec. Earlier specs keep
their builtin identities after later resets. Named-function builtin resolution
retains the actual context node. Eleven lifecycle observations match Java; four
unchanged original TLC models pass. General selector/evaluator builtin graph
sharing remains incomplete.

Generator construction now retains its four source sentinel nodes in allocation
order, including their links and empty/nil arrays. External modules receive fresh
sets; nested modules and expression generation share their owner's set. Twenty-three
actual constructor observations match Java. This is constructor metadata, not
complete ordinary OpDefNode/OpApplNode/OpArgNode/LabelNode or failure-path graphs.

Module generation now retains a concrete ModuleNode before its body, with source
arity -2, whole-module syntax/location, zero-based nesting and separate contexts.
Nested nodes enter the parent's definition list; formals retain the constructed
module owner. Fourteen module observations match Java. Generator class initialization
also retains its process-wide `$$InAssume` declaration before instance sentinels.
Twenty-eight sentinel/marker observations match. Ordinary constants and variables
now retain concrete OpDeclNode identities, whole declaration-item syntax,
module ownership and constructor level data. Constants include themselves in
both parameter sets; variables do not. Generation allocates rejected declaration
nodes and keeps the earlier accepted binding. Local accepted declarations enter
the module context, and identifier expressions retain their declaration node.
Thirty-eight actual Java observations match, including a rejected duplicate's
allocation gap. SymbolTable's registration primitive now matches Java's check
order, return value and structured diagnostics; warnings keep the earlier binding
and return true. Canonical operator origin checks use source identity and actual
module parameters. All 149 observations across 77 registration cases match Java.
The existing Java context test now uses its actual kind-zero OpDefNode constructor
and OpDeclNode with nullSTN. Its unchanged assertions exposed and corrected a
production context-classification shortcut; Java JUnit and Go both pass.
Module generation now retains an external table or a copied enclosing stack with
its own nested context. Declarations retain their original table. Direct EXTENDS
merges the available retained contexts in source order; imported and enclosing
bindings reuse original declaration nodes. All 61 Java observations match for
transitive/diamond identities, original owners/tables and enclosing references.
Declaration arity comes from each syntax occurrence, rather than a name-keyed map.
Actual repeated declarations use the registration primitive: same kind/arity
warns and keeps the first binding; differing kind/arity errors. Nine complete Java
diagnostic comparisons match, including the first location in later bound-name
conflicts. The native duplicate-variable test is corrected against its exact
unchanged Java fixture: warning 4801 followed by the undefined-name error. Its
replacement assertion checks the exact warning code, range and message.
The TLC bridge now reads accepted generated declaration objects for source
locations and arities, and omits rejected local declarations from runtime module
contexts, constant registration and INSTANCE target discovery. It retains the
existing metadata path for native AST APIs without generated source graphs.
All 34 lower Java FastTool observations match across repeated declarations,
enclosing rejections and EXTENDS rejections: complete arrays/counts, locations,
signatures and initial-state counts/values. Six unchanged original TLC models pass.
Module names alone no longer create expression namespaces. EXTENDS imports
unqualified names; a named INSTANCE creates its qualified exports, including an
instance with the same name as its module. INSTANCE defaults resolve only names
available at that source point, including preceding LOCAL definitions. Scalar
default construction precedes WITH; remaining operator-default arity checks
precede missing-substitution errors, in source context enumeration order.
All 30 complete Java diagnostic comparisons match. Four native model/reference
fixtures now declare the actual named INSTANCE their qualified calls require;
the missing-substitution assertion checks Java's exact code, message and range.
Existing focused root/SANY checks, six unchanged original models and all-package
compilation pass. This establishes no new full-workspace pass or inventory credit.
Module and proof INSTANCE processing now share the same RHS generator. Scalar
substitutions generate expressions; operator substitutions generate operator
arguments and retain the source nullOpArg failure diagnostics. The former
module-only generator has been removed; deferred level checking remains separate.
All 50 namespace/RHS cases match complete Java diagnostics. Original Test210 and
Test212 also match complete output and pass unchanged assertions: nested labels
retain whole syntax ranges, ASSUME/PROVE selectors use code 4005, and unapplied
parameterized INSTANCE prefixes remain legal in operator-argument contexts.
Final focused root checks include all original Test206–220 methods; complete
SANY and compilation pass. LET INSTANCE units now use that same substitution
generator and register actual exports in their temporary context. The blanket
I! namespace shortcut is removed: unknown exports error, signatures are retained,
and exports leave scope with the LET. Proof and LET export registration share
one path, preserving kind/arity conflicts and the first accepted binding. Named
exports retain their instantiation location separately from original source
syntax identity; shared parameter-free origins do not erase arity conflicts.
All 28 LET comparisons match complete Java diagnostics. Existing focused root,
complete SANY, seven original model methods and compilation pass. No tests or
fixtures were changed, and no new inventory credit or full-workspace pass is
claimed. INSTANCE generation now retains the resolved substitution array, including
implicit defaults and WITH replacements. LET level checks visit these instances
after definitions and the body, using the lexical operator bindings. All 24 module/
LET level comparisons match Java; existing focused root checks, Test206–220,
seven original models, complete SANY and compilation pass normally. The existing
native level-message assertion now checks Java's exact messages and ranges,
with its fixture unchanged. No full-workspace pass or inventory credit is added.
Module argument constraints now propagate through the symbolic definition
analysis, including user-defined substitutions, indirect arguments, higher-order
forwarding and theorem expressions. The builtin-only argument scan is removed.
All 24 additional paired module/LET comparisons and the previous 24 level cases
match complete Java diagnostics; focused root checks, original related models,
complete SANY and compilation pass normally. The existing TestInstanceNode check
now retains every original method assertion: separate parsing/dependency loading,
successful semantic generation, failed level checking, the exact error count,
code and [1, 3] parameters. The unchanged Java method and faithful Go translation
both pass. SANY methods remain excluded from TLC inventory totals.
The direct co-parameter scan is now removed. INSTANCE checks propagated
ArgLevelParam relationships, including compound and indirect arguments,
higher-order forwarding and LET definitions. Retain set membership and the
source hash sum of declaration identities and argument position; use the shared
Java HashMap algorithm for iteration and collision trees. Native object identity
supplies the source same-class tie-break, so collision-tree order is runtime
identity dependent. All 74 ordinary comparisons match complete Java diagnostics;
a separate 12-relationship collision probe matches diagnostic membership, codes,
ranges and messages, without claiming identical native identity order.
Nonconstant modules initialize their constant constraints to zero before unioning
expression constraints by minimum, matching ModuleNode and suppressing the
previous fabricated second bound. Focused root checks, original related models,
complete SANY and compilation pass normally. No full-workspace pass or inventory
credit is added.
Next finish retained InstanceNode/LetInNode and ordinary canonical operator graphs.
Separate production entry points now expose parsing/dependency loading,
semantic generation and root level checking. Programmatic generation links shared
imported definitions and copied top-level nodes; legacy CLI reporting retains its
per-external-module sequence. Expression checks now propagate each child's
levelCorrect result separately from its diagnostics, including LET instances and
substitution expressions. A reported non-Leibniz substitution does not turn its
parent's result false when Java's InstanceNode returns true. Action/fairness
argument gates use child results, preserving an enclosing error that a raw
Errors check previously suppressed. Definition, assumption and theorem plans
retain these expression results. Unary operands also check before their own
application maxima, using their actual validity to suppress redundant enclosing
errors. Remove the extra Go-only constant-prime rejection: Java allows constant,
literal and boolean priming, and builtin maxima diagnose double priming.
All 12 prime, 27 phase and 74 preceding legacy diagnostic comparisons match
Java. Native expectations use the unchanged source fixtures and exact Java diagnostics;
scratch comparisons add no original-method inventory credit.
ASSUME-PROVE now checks each assumption and PROVE expression in lexical
scope. NEW domains check before their declarations enter scope; uses retain
the declaration's own level. Preserve Java's unusual return rule: PROVE errors
are reported but do not alone make the AP node return false. Temporal bounds
report exact 4356 messages and whole NEW ranges, including indirect definitions.
All 29 parseable AP phase comparisons and 31 legacy cases (including two source
parser rejections) match. Module substitution constraints now distinguish own
LOCAL definitions from EXTENDS imports. Unexported LOCAL bodies do not add
independent constraints to an importer; exported references still retain their
body dependencies. Nonconstant target module bounds survive nested INSTANCE
substitutions, including LOCAL INSTANCE. All 30 expanded comparisons match
(29 parseable modules plus one grammar rejection). Instance constraint collection
also merges every retained substitution expression's own constraints, even when
the target declaration is unused. Use the actual resolved default/WITH array
instead of reconstructing defaults by name. All 27 additional parseable RHS
cases and prior bounded comparisons match Java. Current focused root/model checks
pass in 39.140 seconds, complete SANY passes in 1.818 seconds, and compilation
passes. Original corpus fixtures and assertions are unchanged; no inventory
credit or new full-workspace pass is added.

LET now imports only its retained instances' co-parameter relationships, matching
LetInNode's field-specific propagation. Scalar/argument constraints come from
the body and generated OpDefs. Generation now retains accepted LET instance
export references; their constraints follow symbolic Subst LC/ALC/ALP translation.
The two previously failing imported bounds are resolved. All 22 expanded cases
match, including operator aliases, nested wrappers and LOCAL negative controls.
Retained instances share that symbolic translation instead of inlining target
bodies through WITH. Do not merge the entire target module into LET.
Ordinary expression/signature LET summaries now preserve body-only dependencies
and merge all retained definition constraint fields. Captured formal generation
and shadowing diagnostics match source. Symbolic operator-level conditions also
retain Java's higher-order Leibniz propagation; local caches distinguish captured
formal identities and actual operators between specializations. All 22 additional
signature comparisons match, and earlier bounded comparisons remain exact.
The original SANY `TestLevelChecking.testAll` now retains all phase assertions
and all 51 parameter rows, mechanically verified against source. Unchanged Java
passes all 51; full Go SANY passes in 2.009 seconds. This is SANY method credit
only and changes no TLC inventory totals. No production changes in that test-port
commit; reuse the verified production receipts.
The original nested-module SANY class now preserves its active top-level test
and Java's ignored LET-instance method, including the exact ignore reason. Both
fixtures match source bytes; unchanged Java and full Go SANY pass. The original incremental
radix-overflow method now checks actual generated NumeralNodes and all three
source big-integer values. The basic expression method also retains actual
numeral level checking, syntax identity and ConstantLevel. Numerals are
constructed during semantic generation and reused by the TLC bridge; their
iteration tracking matches source. The other three incremental methods remain
uncredited: native-AST checks omit canonical node and syntax identity, actual
levels, dependency tables and imported operator source identity. Implement those
production graphs before translating the omitted assertions. See `sany_tests/README.md` for the method requirements.
Builtins now use the common OpDef node. Their constructor preserves source
null/empty metadata arrays, variadic Leibniz flags, defined state and checked
level state. All 72 property rows match source. The complete original builtin
initialization/reinitialization class passes in Java and Go; no TLC count changes.
The ordinary OpDef constructor now preserves source parameter/body/syntax
identity, initializes argument metadata and recursion defaults, and registers
after field initialization. Graph links can retain actual shared TLC literals.
All 32 bounded constructor observations match Java. Ordinary generation still
needs complete actual bodies, constructor wiring, registration timing and
recursive completion before crediting its original test.
Generated decimals now retain source image parts, signed-long mantissa/exponent
and overflow unscaled value/scale. TLC and XML reuse their representation.
The exporter preserves source's overflow scale sign and uses generated numeral
values for leading-zero/radix integers. Ten constructor cases, eight decimal XML
metadata cases and ten numeral XML values match Java. The original decimal XML
method remains unchanged and passes in Java/Go. This is bounded evidence and
SANY method credit only; ordinary OpDef and other canonical graphs remain pending.
String generation also retains its actual node, interned value, syntax/location
and level-check iteration. TLC/XML reuse the node. XML no longer decodes an
already-decoded value a second time, preserving quotes and literal backslashes
that are data. Eleven generation/TLC-node cases and ten XML values match Java;
focused models, full SANY and the original string-deserialization model pass.
No original-method count or full-workspace completion credit is added.
Actual formal nodes register in temporary SymbolTable contexts alongside the
native formal map. Scope exit restores both representations; rejected duplicates
preserve earlier bindings. Literal nodes and builtin, bound, record, CASE and
operator applications retain actual children and source syntax. Token nodes
intern raw images at parse time. Call generation resolves symbols and arity
before operands; concrete matching preserves false-result versus thrown-error
behavior. OpArgs and LAMBDA retain their actual symbols, formals and module
ownership, including the owning Generator's failure sentinels.

Ordinary local/top-level OpDefs construct after body generation and formal-scope
restoration. LET retains its actual context through IN and preserves definition
order. Named functions retain their preparation context and construct/register
before generating the body. Their actual function stack detects bracketed and
bare recursion; nonrecursive specifications clear the temporary self formal.

Explicit RECURSIVE entries now allocate actual declaration/dummy-formal/OpDef
nodes and retain all three source module vectors, including rejected duplicates.
Definitions complete the same declared OpDef, preserving original arity,
constructor-sized metadata arrays and localness. Formal arrays are replaced
before operator bodies; zero-arity functions complete before their bodies.
Section numbers/flags and wrong-level construction match source. Preserve Java's
final-formal overwrite behavior in its recursive arity check. Calls use actual
formal arrays, including completed recursive operators in LET scopes.

If a body graph remains unported, preserve the declaration's correct completion
and syntax state without fabricating its body. Its enclosing LET remains
incomplete. Recursive graph and diagnostic receipts belong in PORT_PROGRESS.md.

Ordinary expression labels now retain actual LabelNodes, body/formal identities
and nested label tables. Definition and label bodies push separate LS frames;
quantifiers, CHOOSE, comprehensions, functions and LAMBDA push actual bound formal
groups. Ordinary operator parameters do not enter that sequence. Generate a
label body before resolving parameters and constructing its node. Duplicate
registration retains the earlier node. Reuse the owning Generator's nullLabelNode
at implemented guards, without generating the rejected body.

EXCEPT now constructs its actual node before specs and each mutable pair before
its RHS. Active stacks own AtNode EXCEPT/pair references; base, modifier and pair
syntax retain source identity. Nested bases/indices resolve the outer context.
Label rejection uses those same stacks. All 22 EXCEPT graph cases match Java's
350 complete output rows, and all 22 label graphs now match their 287 rows.
Both complete 22-case diagnostic sets agree on codes, ranges and messages.
Existing focused root/TLC tests, full sany_tests and compile-all pass. Detailed
receipts belong in PORT_PROGRESS.md. No permanent tests or original-method credits
were added in those construction slices. The later level-checking integration
closes basicOpDefTest; see the current level-checking status below.

Label tables now preserve Java's default Hashtable bucket/chain enumeration and
rehash behavior. LabelNode and OpDef accessors retain nullable table identity,
earlier duplicate entries and shared mutable tables. LabelNode exposes its name,
arity, body, goal and single body child. The 22 bounded table/accessor cases match
all 869 rows; 25 generated-label cases match all 602 graph/enumeration rows and
complete diagnostics. Existing focused root/TLC, SANY and compile gates pass.

NEW now constructs actual declaration and NewSymbNode graphs, generating the
domain before registration. Retain declaration/wrapper syntax, source kind/level,
arity, synonyms, selected earlier bindings and actual domain identities. Native
proof scopes carry these selected declaration pointers. Bare higher-arity symbols
now reject during generation and preserve actual nullOAN in enclosing graphs.
The 25 bounded generateNewSymb cases match all 234 output rows and their complete
frontend diagnostics. Existing root/TLC, SANY and compile checks pass; the 56
formal and 59 AP phase/substitution comparisons still match. Surrounding canonical
named AP ownership and NEW level/visitor/evaluator integration remain pending.

Unnamed AP bodies now retain their actual node, assumption/prove pointers,
declaration-scope array, boxed flag and source proof-state transition. Ordinary
AP registers the actual shared marker; delimiter mismatches and boxed AP inside
ordinary assumptions now report source E4005 during generation. Clause tracking
matches ordinary labels following these AP bodies. All 20 cases match their 227
complete graph/metadata rows and frontend diagnostics. Existing focused root/TLC,
full sany_tests, compile-all and AP/formal comparisons pass. Receipts belong in
PORT_PROGRESS.md.

Concrete theorem/assumption definition and owner constructors retain source,
module, body, proof, parameter-array and backlink identities. Registration precedes
parameter installation. Leaf-proof construction aliases its arrays; module
assertion/top-level getters preserve Java's cached mutable arrays.

Assumption generation creates a named definition after its body, then allocates
the separate owner. Theorem generation creates a named provisional goal before
its body, completes/registers it before the proof, then allocates its owner after
the proof. Outer AP and labels retain the actual goal; nested AP goals remain null.
Top-level AP declaration and marker contexts remain open through proof processing,
with named theorem registration in the enclosing module context. Ordinary and
labeled AP bodies preserve actual children and clause metadata. AP theorem names
used as expressions fail before application allocation, returning the actual
source failure node. Real owners are retained for complete bodies with no proof,
OBVIOUS/OMITTED, or complete canonical BY proofs. BY allocates its temporary USE
node before a leaf sharing the actual arrays; theorem completion reuses that leaf.
Top-level USE/HIDE retain actual vectors and source fact-validation order. Their
constructors preserve nullable step names, array aliases and fresh child copies.
Missing canonical BY vectors or proof children keep their owners incomplete.

NonLeafProofNode and DefStepNode constructors now preserve their supplied arrays,
context and nullable interned step number. Non-leaf children are null for empty
steps; definition-step children preserve a nonnull empty array and fail on null
definitions. Both copy populated child arrays and retain null entries. Their
32-case constructor comparison matches Java in all 80 rows. Numbered-step
OpDefNodes retain their actual step backlinks, register with
zero arity, and preserve Java's null body/parameter/level arrays and false flags.
Their children remain a fresh array containing the ordinary null body slot.
All 72 rows across 32 constructor cases and three registration collisions match
Java. Structured generation now enters actual proof contexts, retains complete
DEFINE/USE/HIDE and ASSERT/HAVE/CASE/WITNESS/QED steps, handles ordinary-expression
SUFFICES and finishes nested proofs before their theorem owners. Every named step
preserves the source provisional-goal allocation, including discarded non-theorem
goals. Actual symbols, label selections and @ shorthand retain their shared nodes.
Step syntax replacement also updates cached locations. All 993 rows across 21
whole modules match Java. AP steps now retain real goals and declaration contexts:
ASSERT declarations live only through their proof; SUFFICES declarations become
visible after their proof and are merged into the enclosing context at proof end.
Source AP/definition/owner flags and nested scopes retain their actual identities.
Context merging uses Java Hashtable enumeration rather than insertion history.
All 2,281 rows across 29 AP proof modules and 37 direct table observations match
Java. TAKE/PICK now construct actual bounded/unbounded applications, retain
source formal groups and label parameters, and install PICK bindings after its
proof using the captured context. All 2,424 rows across 35 whole TAKE/PICK
modules match Java, including nested scope restoration and rejected bindings.
InstanceNode constructor/accessors/children and Subst storage/mutation/identity
lookup now match all 582 rows across 192 constructor combinations and mutation
observations. Supplied arrays retain identity, null arrays become empty, and
null entries preserve source failures. SubstIn/APSubstIn copy/default constructors,
array ownership, explicit mutation and completeness checks now match 518 rows
across 180 source scenarios. Preserve their distinct source diagnostic codes.
Production INSTANCE substitution templates now retain actual defaults, explicit
RHS nodes and shared Subst mutations, including canonical label rejection outside
a definition. All 51 rows across 17 direct processSubst scenarios match Java.
Unnamed INSTANCE now retains actual shared/copied imported definitions, source
pointers, wrappers, module vectors and proof-context bindings. Targets retain
the instantiated flag; proof instances preserve body syntax and Java raw-array
localness. All 347 rows across 12 valid whole modules and one retained parser
rejection match Java. Named INSTANCE now constructs actual qualified definitions
and module-name symbols with source formal syntax, compound identifiers and
caller-owned LET/proof instance arrays. Qualified calls retain actual imported
operators; bare module names are rejected as expressions before application
construction and retained directly as facts/DEF references. All 605 observed rows
across 18 whole modules match Java, including duplicates, higher-order formals,
empty targets and nested named instances. Qualified GeneralId arguments now
check actual symbol arity before allocating OpArg; instance-prefix application
errors precede final lookup, and terminal module names remain incomplete
operators. Qualified instance fact/DEF references retain actual symbols and proof
array slots. All 1,750 observed rows across 41 whole modules match Java, including
actual operands, proof reference arrays and rejection order. EXTENDS inheritance,
complete instance vectors and general subexpression/fixity selectors remain
incomplete. Context/module collection getters now preserve source enumeration,
definition history and lazy array snapshots. Inner modules register actual nodes
in the enclosing context, with a shared loader-ordered external-module table.
Forward inner references fail before template allocation. All 5,697 observed rows
across 43 whole modules and 17 loader observations match Java. EXTENDS inheritance
and complete module/level/visitor/evaluator graph parity remain pending.
EXTENDS now copies actual assumption, theorem and top-level vectors in source
order, preserving duplicate references through diamonds. Its inherited instances
enter top-level vectors, while the separate instance vector follows Java's code
and remains local. Actual extendee arrays and separate direct/recursive mutable
set caches preserve copy ownership and cache lifetime. Imported expression
metadata shares actual definitions. All 7,519 observed rows across 47 whole modules
and associated cache/null observations match Java. Inherited level, visitor
and evaluator work remain
incomplete. Original-definition comparison now requires the actual operator or
assertion class, matching immediate source pointers and cached declaration arrays
from the source module. A live context scan no longer changes previously frozen
parameter-freedom decisions. All 8,318 observed rows across 47 whole modules and
17 comparison pairs per module match Java, including source chains, mixed classes
and nulls. This evidence adds no original-method or full-workspace completion
credit. Definition-path accessors now preserve actual compound-array ownership,
local-name lookup and counted path joins. `UniqueStringJoinN` follows Java
assertions, null behavior, literal `!` and intermediate interning order. All
16,872 manual comparison rows across the retained 47 modules match Java. These
observations add no permanent tests or original-method credit. Context duplication
now copies history independently but shares symbol nodes, rebuilding lookup
newest-first with plain node names exactly as Java does. The oldest repeated
name wins and module keys become plain names in the copy. All 1,578 direct
comparison rows and the retained 16,872 module rows match. EXTENDS Context
merge now snapshots history, derives keys from actual classes/names, compares
concrete classes and reads current syntax-tree locations. All 1,115 direct
merge rows and the retained module rows match Java. Module generation now retains
those canonical merge diagnostics in direct extendee order; native scans only
supply expression metadata. All 24,529 rows across 68 modules match Java,
including 25 explicit diagnostic location/parameter rows and the complete earlier
observations. Missing contexts on resolved extendees now log the source internal
error and continue vector copies and body generation. Each repeated EXTENDS
occurrence retains its own token position and UniqueString parameter. All 118
comparison rows across 20 root/nested scenarios match Java. Missing-module
resolution now records the source internal error and throws, stopping nested and
enclosing generation while retaining earlier diagnostics and copied vectors.
The semantic driver chains a checked SemanticException and the legacy entry point
returns ERROR; unexpected runtime failures retain their propagating boundary.
All 150 direct abort rows, 12 driver rows and 3 legacy-boundary observations match
Java. Stack frames remain native to each implementation. The retained 24,529
whole-module rows also match. General shared Errors ownership remains part of
the broader semantic audit. Bounded generation
evidence does not complete inherited level checking, visitors or evaluator graph
sharing.

All 1,083 comparison rows across 47 bounded theorem modules match Java, including
syntax kinds, UID order, exact goal/reference pointers, label tables, declaration
scope, failure nodes and diagnostic codes/ranges/messages. Assumption and direct
constructor comparisons also pass. USE/HIDE/BY comparisons add 427 exact rows
across 40 modules and 116 constructor rows. Existing focused tests, full
sany_tests and compile-all are the affected gates; receipts belong in
PORT_PROGRESS.md. These observations add no original-method or full-workspace
completion credit.

Remaining module vectors, general qualified selectors, INSTANCE/fact/imported identities,
recursive inherited level checks, visitors, shared Errors/exception integration
and evaluator graph sharing remain pending. Other Context iteration callers and
live/concurrent Hashtable enumeration still require source audits.
Missing canonical children or earlier native-only import identities keep owners
incomplete. Next semantic work is remaining selector/instance identities,
level checking and evaluator sharing.
Complete allocation order across all graphs remains unproven.

Canonical level-checking prerequisites now include the original
`ParamAndPosition` and `ArgLevelParam` classes. They retain actual symbol
references, Java signed hashes, reference equality, nullable formatting and
source `occur` behavior. All 5,125 temporary Java/Go comparison rows agree.
The original `SetOfLevelConstraints` and `SetOfArgLevelConstraints` are also
ported, including tightening puts, raw copy constructors, nullable values,
source key equality and HashMap iteration. All 680 temporary map observations
agree with Java. Canonical SANY constructors now own the inherited `LevelNode`
data, actual mutable symbol/argument sets and source getter guards. Formal and
declaration level checks and canonical subnode aggregation agree with Java across
575 direct observations. The retained 24,529 whole-module graph/diagnostic rows
still agree. Operator arguments now perform the source level check and retain
shared operator collections, while keeping their own non-Leibniz set. The full
operator-definition level algorithm and metadata accessors are translated;
792 direct source observations agree, including recursive bounds, weights,
higher-order conditions, partial failures and numbered steps. Canonical nodes and
actual TLC literal bodies now share their level and iteration cells. Literal
views retain actual mutable canonical sets and constraints on the body node and
allocate no semantic identity; direct TLC checks and setters observe the same
cells. All 60 direct literal observations and the retained 575, 792 and 24,529
comparison rows agree with Java. The full original incremental basicOpDefTest is
now port complete. The complete LET and transitive-import translations follow below.
LetInNode now performs the complete source check: ordered component traversal,
body parameter-set copies, retained constraint/dependency merges and filtering
of dependencies bound by local formals. It preserves the source omission of
non-Leibniz propagation. All 68 direct Java observations agree, including full
metadata formatting, partial failures and HashSet copy capacity/order. These
observations do not close the two original incremental LET tests.
LabelNode now checks formals and its actual body, preserving the source cached
`true` result and direct delegation of six metadata getters. Its inherited
non-Leibniz set remains separate. All 81 direct Java observations agree, including
label-bodied operator checks and exact getter guards. These observations add no
original-method completion credit.
Theorem/assumption definitions now use their complete source metadata algorithm
and guarded accessors. Preserve source fresh-table resets, non-monotonic levels,
unchanged weights and the Leibniz allocation inside the formal loop. All 458
direct observations agree, including theorem definitions in LET checks.
USE/HIDE, leaf proofs, definition steps and non-leaf proofs now invoke the
source ordered subnode aggregation. Resolve each actual graph child only when
visited, preserving prior writes on later failure. Non-leaf proofs copy steps
followed by instances before checking; definition steps retain the array passed
at entry while observing in-place element changes. All 282 direct Java rows
and the retained 575 common rows agree. No original-method credit is added.
NewSymbNode now checks the declaration and optional set, preserves the source
exact-temporal diagnostic E4356, and shares five actual set collections while
retaining its own non-Leibniz set. Removing a set retains previously shared
metadata and correctness. All 195 Java observations agree, including exact
error text, deduplication, failure order and typed-null set handling.
AssumeNode now checks its expression and optional definition, delegates six
metadata getters and ports its source-specific formatter. Its Java shadow
iteration counter remains distinct from the inherited cell: the next-iteration
overload reads the inherited counter. Diagnostics do not force the returned
result false, and temporal-constant constraints update the assumption's own
collections. All 160 detailed Java observations and 575 retained common rows
agree. TheoremNode now preserves its own shadow counter and inherited subnode
aggregation, including PICK and recursive temporal-proof checks. All 223 direct
Java observations agree. Those temporal proof cases use explicitly prechecked
source applications and add no original-test credit; fresh application checking
is translated below.
AssumeProveNode now follows the complete source check: two assumption passes,
ignored PROVE boolean result, retained metadata merges and temporal-constant
constraints only when assumption checks succeed. Preserve null failure order
and repeated virtual level reads. All 143 direct Java rows agree. No original
method completion credit is added.
Subst now translates the five static parameter/constraint/dependency helpers,
including first-match reference lookup, shared replacement sets, fresh unmatched
singletons and source tightening order. Typed-null lookup matches Java null. All
130 direct Java rows agree; runtime failures compare exception class rather than
JVM-specific enhanced messages. SubstIn/APSubstIn and Instance callers are
translated below. No original-method credit is added.
ModuleNode now performs the source recursive-section initialization and two
checking passes, ordered module/definition/top-level checks and retained
constraint merges. Its isConstant method checks actual operator bodies and
theorem levels; getLevel preserves the source prohibition. The module formatter
uses Java collection forms. All 332 direct Java rows agree, including diagnostics,
cache behavior and partial failures. No original-method credit is added.
SubstInNode and APSubstInNode now perform their complete source level checks:
ordered child checking, retained level parameters and fresh constraint/dependency
translations. SubstIn copies and sequentially rewrites both all/non-Leibniz sets;
APSubstIn retains all parameters and leaves its non-Leibniz set untouched. All
768 direct Java rows agree across signed iterations, failed children, duplicate
and chained substitutions, identity checks and malformed metadata. These manual
observations add no original-method completion credit.
InstanceNode now performs the complete source checks for replacement levels,
non-Leibniz operators, argument bounds and co-parameter dependencies. It filters
exported constraints by actual formal references, resets only level parameters
and retains its other metadata. Proof-instance getter overrides and raw formatter
follow Java. Parameterized level diagnostics retain their original arguments for
duplicate equality, including recursive-module diagnostics. All 545 direct Java
rows and the retained 332 module rows agree. These are manual observations;
standalone generation and source-reference assertions are now retained in the
complete original LET translations below.
OpApplNode now translates its complete source level algorithm for declared and
AnyDef operators, including theorem/assumption definitions, higher-order bounds,
weights, non-Leibniz propagation, bound filtering, LC/ALC/ALP translation and all
temporal checks. Preserve source retained/reset collection choices, false-result
caching, short-circuit/repeated checks and failure writes. Private getArg matches
formal references rather than UIDs. All 2,897 Java observations and retained
545 instance/332 module rows agree. Temporal cases include fresh child
applications. Diagnostic symbol parameters compare references, including null
normalization. These observations add no original-method completion credit.
All five original IncrementalSemanticParseTests methods are now port complete.
The two standalone LET tests preserve original dependency tables, actual module
and expression checks, syntax identity, constant levels, concrete graph classes
and exact imported source references. They use the original embedded module
sources, not enclosing-module surrogates. Unchanged Java JUnit and all five Go
methods pass. OpDefNode getSource/hasSource now preserve the source immediate
reference and self fallback, including typed-null handling. No TLC inventory
count changes; broader standalone generation and evaluator integration remain.
Standalone generation, remaining graph fidelity and evaluator collection sharing
remain pending: the legacy TLC symbol-parameter API still returns its separate
TLC symbol projection. No TLC inventory count changes.

Function and set-comprehension bridge nodes retain one group per syntactic
bound, including multi-name lists and distinct adjacent tuple bounds. Each domain
is converted once before the formals enter the context, matching Java generation.

Node constructors now use the actual 446-entry source `SyntaxNodeImage` table,
which differs from node-kind constant names. Unknown selector syntax retains its
source image and zero-valued slot, then reports constructor errors before name
resolution. Reconstructed tuple membership retains both structural diagnostics.
General selector and constructor fidelity remains reconciliation work.

Current bounded observations match Java: 599 complete parser TRACE/results, 41
output routing/format cases, 12 LAMBDA semantic cases, 13 CHOOSE semantic cases,
31 selected declaration/LHS trees, 54 substitution target/arrow trees, 22 quantified
semantic observations, three CHOOSE runtime probes and ten quantified metadata/
runtime probes, plus 93 selected expression trees including ranges, four function-application
runtime probes, ten bracket constructor/group metadata and runtime probes,
16 brace semantic observations, 25 command-generation observations, nine binder-generation observations,
nine brace metadata/runtime probes, all 446
node-image entries, and 16 selected complete definition trees with kinds, images
and ranges, plus 94 selected complete module trees with kinds, images,
ranges, original images and proof levels.
Keep each scope distinct. Whole-module canonical AST assertions, general source ranges and complete parser/semantic graph fidelity are not established.
Detailed source comparisons and verification receipts are in `PORT_PROGRESS.md`.
Numeric/general-Object driver formatting, PrintStream error-state queries and remaining production-frame
coverage still require reconciliation. Do not claim full SanyOutput or parser
parity. Both original syntax corpus methods are now port complete across all
355 parameter contexts each. The full recursive SANY-to-DSL translator, source
known-failure runner and AST equality assertions are installed. Source bytes,
expected trees and ERROR/SKIP attributes are unchanged. All 355 actual translated
outputs match Java, including rejection and checked translation errors; the
original Java class passes all 710 test contexts. The earlier 356 metadata/expected
AST rows and 552 operator/470 reparser helper rows remain separate receipts.
Preserve the source known-failure inversion and distinguish checked translation
errors from structural assertion failures. These corpus results do not establish
all malformed production states, canonical integration or evaluator sharing.

Missing modules and filename/module-name mismatches now abort loading with source E4220/E4221 details and null or
actual importing-module locations. The existing front-end exception boundary
reports these failures; the native library API returns diagnostics and
preserves a previously parsed root. Seven file-loading observations match Java.
The source unresolved-module search now exhausts EXTENDS before INSTANCE,
restarts from the root after each binding, and separates file parse units from
inner modules. Cycle diagnostics retain source E4222 and the complete filename
path. Semantic order is derived from the recorded parse-unit relationships;
inner modules are generated within their owning external module. A forward
INSTANCE now preserves inherited symbol conflicts and the external-module-table
E4223 conflict. Sixteen further loading observations match Java, with only
independent extraction directory names normalized in the scratch comparison;
raw outputs remain available. General parser parity remains unproven. The
original SANY `TestContext` method now exercises Context directly, with its
original failure result, one-error count, E4224 code and declaration/definition
parameter assertions. It lives in root `sany_context_java_test.go` so it can
access the private context implementation; the parser-fixture surrogate was
removed. The direct context and production EXTENDS path share diagnostic
construction, including structured parameters and source locations. EXTENDS
conflicts now retain actual symbol locations and definition provenance, compare
source semantic-node classes, and reuse parameter-free instance definitions.
Substituted declarations are not INSTANCE exports; only an explicit named
instance creates a qualified namespace. Per-module semantic reporting now
follows the shared accumulated Errors instance; nested module diagnostics are
included in their enclosing external reporting iteration. Generation completes
before the raw shared Errors.isSuccess gate permits real level work. Warning
elevation does not change raw Errors success. Further semantic-node parity is
still required; these checks do not prove completion. Buffered-file modes,
invalid-mode exceptions and trace/worker opening boundaries were reconciled in
`db62dfb`; preserve those verified fixes. Do not add synthetic phase output.
The actual Java record linter is now ported: declaration and formal-parameter
dependencies, same-domain EXTENDS suppression, binding and proof scopes, exact
warning text and a distinct phase after successful semantic analysis. Matching
checker output or a high test translation percentage does not prove full
parity.
Five main-suite contexts have known source-failing or JVM-specific
reconciliation issues below. The other four pending contexts are original assumption-disabled
distributed transport models; checkpoint models are now complete. Consult the inventory for each exact disposition.

| Eligible class | Unresolved issue |
| --- | --- |
| `DistributedTrace` | Unchanged Java and the temporary faithful Go translation fail no-generated-TE-spec and inherited success-exit assertions (actual 12). Both produce the original eleven-state trace; retain four workers and original expectations. |
| `AssertExpressionStack` | Unchanged Java and the faithful Go draft fail `assertNoTESpec` and the inherited success-exit assertion. Trace generation is enabled by the base, and the actual exit is 14. |
| `DepthFirstTerminate` | The original runtime-derived worker setting requests 48 workers on this host. Java rejects multithreaded DFID, records `GENERAL`, and exits 255; see issue 548. Do not substitute one worker. |
| `InliningTest` | The whole method includes HotSpot/JFR compiler-inlining records, reflective annotations, and JVM callee descriptors. Porting only its checker assertions would be incomplete. |
| `LivenessSymmetryWarning` | Java and Go satisfy all three warning/diagnostic assertions but fail the inherited success exit: expected 0, actual 13. The original model produces a liveness violation. |

Keep these entries marked **Reconcile** until their original expectations or
JVM-specific requirements receive an explicit disposition. The faithful drafts
remain in ignored scratch; no new persistent skips or production divergences
have been introduced to hide the failures.

Additional pending work:

- `TLCRuntimeTest.testIsThroughputOptimized` asserts the JVM ParallelGC selection
  made by the original Ant flag. Do not fabricate a successful Go GC response.
- Four inherited long fingerprint contexts remain pending: the OffHeap random
  method retains 2,147,483,648 iterations, and all three sequential methods retain
  3,221,225,473 iterations. Their full drafts remain outside the persistent suite
  and uncredited. LSB/MSB random translations are installed; see receipts above.
- `OffHeapDiskFPSetLongTest.testMultipleFlushes` fails in unchanged Java and Go
  when an invariant check reaches an already shut-down flusher executor. Keep
  the original four rounds and insertion counts.
- Three `ConcurrentWriteTest` methods fail identically in Java and Go because
  overlapping buffer flushes or positional-write gaps produce zero values.
  Preserve all original writes and read assertions.
- Nine concurrent fingerprint random contexts still require full stress runs.
  Their default insertion property is 2,147,483,649. Smaller configured audits
  are evidence about the implementation, not completion of these contexts.
- Three partitioned fingerprint methods traverse memory-derived buckets and
  fail their original minimum-size assertions in unchanged Java and Go. The
  default OffHeap case produces 8,388,575 puts. LSB and MSB audits with explicitly
  matched 256 MiB budgets produce 25,165,823 puts each, with all 48 worker counts
  matching Java. These source failures do not earn pass or full-size credit.

The large fingerprint runs need approximately 32 GiB for concurrent main and
merge/checkpoint files, plus reserve. The relocated `/mnt/b` filesystem has approximately 1.3 TiB available as of
October 6. Use ignored `.codex-gotmp` for large test storage and recheck
resources before each large workload. Preserve original bounds. Pending
source-behavior questions have not been resolved by silence.

A separate scheduling concern remains for alias trace replay: source-compatible
multiworker dumps can replay 10 or 11 states with differing final actions.
Java and Go check invariants even on successors excluded by the trace constraint.
Do not reorder these checks or weaken replay assertions without an explicit
decision to diverge from Java. See the detailed receipts in `PORT_PROGRESS.md`.

## Implementation details to preserve

Recent fixes restore these source behaviors:

- A disk fingerprint search selects one reader for the whole search. Indexed
  workers bypass the pool mutex; pooled readers return after normal completion.
  I/O failures propagate before the return step, and close errors propagate.
  Reader arrays are published as immutable atomic snapshots.
- Heap table counters increment independently, matching Java's `LongAdder`
  behavior. Stripe locks protect table entries; the native metadata lock still
  protects bucket capacity. Counter reads remain independent.
- OffHeap insertion and lookup use atomic words and CAS rather than a set-wide
  mutex. Eviction selection and reset retain the source CAS assertions.
- Value streams use the original 8,192-byte buffered input/output layers.
  Byte-queue adapters remain direct. Graph node and pointer files use buffered
  random-access files; resets and checkpoint encodings follow Java.
- Intern-table recovery reads through its buffered stream and `atEOF` behavior.
  Unique-string primitives retain the original external string protocol.
- Disk-queue growth retains one reused Dummy-equivalent state, the original
  header, all enqueue calls, and approximately 14 GiB of raw state storage.

Prefer concrete structs and keep TLC in its existing package. Preserve source
collection iteration order wherever output, exploration, diagnostics, coverage,
or fingerprints can observe it. Use insertion-ordered maps, explicit slices,
sorting, or the existing Java HashMap ordering helpers as the source requires.
Do not introduce a second parser, compatibility fallback, or speculative
behavior changes. Consult existing audit notes before repeating old work.

Distributed integration is active again. Follow the original TLC coordinator,
worker and fingerprint-server algorithm through Go endpoint boundaries; the
rpc25519 alternative and Java RMI implementation are outside this work.
`DistributedWorkerEndpoint` now supplies worker calls to server threads, the
smart proxy and shutdown. Registration preserves both source URI calls and
propagates failures, including failures after the thread starts. The nine
original smart-proxy cases now use the public method with the source dummy
worker rather than bypassing the endpoint call. `DistributedServerEndpoint` now also supplies discovery, worker bootstrap,
file loading, coordinator interning, fingerprint-server registration and
keepalive calls. All remote settings/status calls return errors. The local
adapter pins coordinator interning identity before worker initialization and
copies mutable metadata for workers. The root worker loader reads specification,
configuration and deadlock arguments in source order and stops on failure.
`DistributedFingerprintEndpoint` now supplies fingerprint operations to the
manager, coordinator registration and FP server startup. Its local adapter
retains storage ownership. Endpoint-returned failures enter the same manager
failover paths as local failures. Workers receive independent manager snapshots,
preserving aliased partition wrappers and shared storage endpoints while omitting
the coordinator-only trace. The fingerprint network endpoint now runs over Go net/rpc and TCP. The host
supports named fingerprint objects, concurrent calls and listener/connection
shutdown independent of storage lifetime. Requests are not retried by the
transport. Source-style manager failover consumes connection/storage failures.
Scalar/batch answers, all 64 fingerprint bits, null/empty vectors and checkpoint
filenames cross the wire. Null batches now fail across all six native storage
backends, matching the source dereference rather than being treated as empty.
Trace recovery stays coordinator-local. A native state/value graph payload now preserves full int32 levels, signed
UIDs, state/value/string object sharing, symbolic set representations and caches,
with the source function/predicate/lazy materialization rules. Finite configured
constant operators (`OpRcdValue`) retain their argument rows, results and operator
application behavior through native payloads. Gob round-trip and worker TCP
request/result checks pass. The codec still rejects extended evaluator state
metadata, opaque custom user values/data and evaluator-backed operators; those
are not silently discarded. Continue auditing actual transferable data rather
than implementing Java object serialization: source `OpLambdaValue` retains a
non-serializable `Context`, while `MethodValue` retains reflection/method handles.
`Action` declares Serializable, but its ordinary predicate/context objects do
not; that declaration alone does not establish transferable extended states.
`DistributedResultPayload` now carries worker result partitions and signed
counters. Repeated state/fingerprint vectors retain identity, and all partitions
share one state/value graph. Null arrays and partitions remain distinct from
empty arrays and vectors. Only active vector entries cross the wire, matching
the source vector serialization contract. Focused gob round-trip checks and
the related original vector/smart-proxy tests pass.
`NetworkWorkerEndpoint` now supplies all five worker calls over TCP. The shared
RPC host publishes named workers and removes them on successful exit without
closing other endpoints or cancelling in-flight computations. Worker request
and result payloads are integrated. Native operation errors preserve the
coordinator's connection, null-failure and block-size retry decisions. Worker
failures retain both error states, their sharing, `KeepCallStack`, nullable
messages, cause/suppression graphs and sender diagnostic stacks. Shutdown
recognizes unavailable native workers. Focused TCP, retry/loss and codec checks
pass, as does one short concurrent computation/keepalive race check.
`NetworkServerEndpoint` now supplies every coordinator call over TCP, including
settings, files, interning, worker/fingerprint registration and manager snapshots.
Registration sends named TCP endpoint references; local unpublished endpoints
are rejected. Coordinator-owned storage is published on its advertised listener
for worker snapshots. Manager wrapper sharing and availability are preserved,
while failover state and trace ownership remain isolated. Acquired callback and
snapshot connections have explicit owners and close on host/client shutdown.
Native failures now enter the existing resolver and worker keepalive catches.
Native discovery now probes binding presence without invoking coordinator
settings. It preserves connection-refused versus reachable/not-ready retry
behavior and retains owned connections for keepalive relookup. Worker network
environments publish workers on a real advertised TCP listener before exposing
them to shutdown/keepalive, then register their named references with the
coordinator. Bootstrap retains its polynomial/interner/resolver/app/manager
order. Worker addresses now use `tcp://`; source URI authority/index/NFC checks
are retained with the scheme change required by the user. An actual empty-block
worker callback passes over TCP. `DistributedCoordinatorNetwork.Publication`
now supplies the existing model-checking lifecycle with a listener opened at
`CreateRegistry`, native binding replacement/removal and local shutdown lookup.
Unpublishing one coordinator leaves other hosted endpoints intact. Initialization
failure retains the master binding and skips worker publication, as in the
source; the process owner closes the retained listener. Actual model-checking
init-error and focused publication checks pass. FP network environments now
publish owned storage, register its TCP reference and remove its publication at
the command's rejection or reporting-loop shutdown points. Native registration
rejection retains the source early return without flush, and leaves a worker on
the same host available. Reporting preserves the five-minute wait boundary;
remote exit wakes it through the existing storage lifecycle. Remaining payload
classes and separate-process full model coverage remain pending. Native CLI entry points now expose `server`, `worker`,
`fpserver` and `worker-fpserver`, with dedicated help and startup property
consumption before role initialization. Worker command lifetime follows its
exit latch; the combined command waits for both roles. Callback publication
accepts a reachable host with the OS-selected port. These entry points still
need broader separate-process model coverage. Native coordinator signal shutdown
is now implemented at the original shutdown-hook registration boundary. A separate-process native coordinator/worker run of the unchanged DieHard
model now completes normally and matches all seven upstream expected trace
states. It exposed and repaired two source deviations: TLCApp injected
predecessor/action objects that Java does not attach, and coordinator publication
used those objects instead of the incoming successor UID as the predecessor
trace location. Publication now writes the incoming UID before replacing it,
without adding evaluator metadata. Focused source-contract and native-boundary
checks pass. The probe ran from the model directory with one worker and no
checkpoints; it is supplemental evidence, not original remote-harness port
completion (that harness remains unconditionally assumption-disabled upstream).
Native coordinator SIGINT/SIGTERM handling now runs the existing worker
shutdown hook before process exit, and unregisters signal handling on normal
return. A separate-process MC06 startup probe confirms SIGTERM exits the
coordinator with 143 and its registered worker with 0. This is an interruption
check, not full MC06 model verification. A three-process DieHard run with one
standalone FP server also matches all seven trace states and exits all roles
normally. Full remote test translation remains pending.

Native EWD840/MC06 separate-process coverage now runs both coordinator-owned
and remote fingerprint storage in `tlc_distributed_process_test.go`, including
standalone FP servers and the combined `worker-fpserver` command. The combined
row uses one process and one native listener for both roles, with a separate
coordinator; it verifies full exploration and normal command lifetime through
fingerprint shutdown. No production change was needed for that case.
It preserves the unchanged N=7 model and configuration, checks actual TLC event
codes, requires exactly 114,942 distinct states and zero queued states, and
rejects GENERAL. The helper isolates each role's process globals and logs role
output under `go test -v`; all child processes are joined on failure or normal
completion. Fixtures are byte-identical to pinned upstream. The remote row
explicitly uses the supported MemFPSet implementation and one worker thread.
The ordinary full workloads pass normally, without race instrumentation. This is native
transport coverage using the original assertion bodies, not completion credit
for the source's unconditionally disabled in-JVM harness. Orderly worker/FP process cleanup now drains accepted native RPC replies before
closing connections. The gob server codec tracks requests until response flush,
preventing the reporting-loop return from cutting off its triggering Exit reply.
Explicit forced Close remains available and does not acknowledge unfinished
calls. No insertion retry or exception suppression was introduced. Focused
shutdown checks pass normally and under an exact short race selection; broader
failure/recovery coverage remains pending.

Distributed fingerprint checkpoint/recovery now preserves the original
IOException-only catch: sequential begin/commit or recovery stops on an unchecked
failure, while I/O failures report through ToolIO and continue to the next grouped
server. Fatal local failures escape. All three memory fingerprint implementations
retain their Java commit IOException category. MemFPSet and MemFPSet2 now distinguish
clean EOF from a truncated eight-byte record, preserving their respective coded
runtime failure and IOException. Valid checkpoint recovery and corrupt-file
rejection are verified through the native TCP boundary, alongside existing
original manager methods. Native MC06 initial-frontier checkpoint recovery is now exercised in fresh
processes by the `checkpoint_recovery` row of the existing process test. The
producer commits queue, trace, local MemFPSet and intern-table files through
TLCServer.Checkpoint with all 16,384 original initial states queued. The fresh
coordinator uses the real `-recover` CLI, must report those exact recovered
fingerprint/queue counts without regenerating initialization, and must finish
with 114,942 distinct states and zero queued states. The producer and recovery
roles retain process-global isolation. This covers a quiescent initial frontier,
not fresh-process recovery after an assigned-block checkpoint or remote-FP
CLI recovery; those broader boundaries remain pending.

The native MC06 `worker_loss` row now kills an owned worker after a real block
arrives and successor evaluation is paused. A second worker remains registered,
and a replacement joins after the kill. The coordinator must report loss,
deregister exactly once and complete the full 114,942-distinct/zero-queued model.
Java's exact GENERAL cache-statistic warning for a dead worker is required in
this new failure case; any other GENERAL event fails. Existing no-failure rows
retain their original no-GENERAL requirement. A short race check verifies that
simultaneous keepalive/RPC loss reports requeue the block and decrement worker
count only once. This covers one interrupted worker with a survivor/replacement,
not loss of every worker or arbitrary network partitions. No production shortcut
or warning suppression was needed for this case.

The `all_workers_lost` MC06 row now covers loss of the sole registered worker
with an assigned block. It waits for deregistration and the source finally-block
warning before registering a replacement, then uses a native IsDone call to
require that the coordinator is available with unfinished work. Replacement
exploration must finish with 114,942 distinct states and zero queued states,
exactly one deregistration and only the original dead-worker cache warning.
The full row passes normally; no production change was needed. Arbitrary network
partitions and fingerprint-server failure/recovery remain separate coverage gaps.

Coordinator publication now preserves the source failure boundary after FP
insertion. Missing/nil selected state partitions, null states and null visited
vectors fail instead of silently dropping successors. Trace write errors return
to the server thread's outer failure handler and terminate that thread; it does
not perform another dequeue after marking the run failed. Partitions with no
selected bits retain Java's lazy dereference behavior. Focused checks and one
full unchanged coordinator-FP MC06 process run pass.

The assigned-block checkpoint barrier now has a focused native TCP check.
A real server thread completes and publishes its held block before the disk
frontier, trace and local or native TCP fingerprints are checkpointed. The worker resumes,
and reopening storage recovers the selected frontier's exact UID and trace FP
positions. Normal and exact short race checks pass; no production fix was needed.
That short check reopens storage in the same process. Separate full-model
coverage now restores a mid-run local MemFPSet checkpoint in fresh coordinator
and worker processes after abrupt producer exit. Interruption before the first commit is now also checked with the source old-queue
and full-trace recovery behavior. Interruption after queue commit is now verified
with the new queue and old trace metadata. Interruption after intern-table commit
is also verified with committed trace/intern metadata before FP commit. Isolated
trace-commit interruption remains pending; nested FP-commit interruption now
passes with mixed committed files;
fresh-process remote-FP CLI recovery is the source startup
limitation described above.

The shared bit-vector iterator now rejects null input and uninitialized words,
matching Java rather than treating missing FP answers as empty results. This
protects both coordinator publication and worker contains-block processing.
Native FP replies distinguish null vectors, uninitialized word arrays and
initialized empty arrays explicitly, preserving valid empty iteration through
gob. Worker failures retain predecessor context and KeepCallStack locally and
through native error traits. Original bit-vector formatting and manager tests,
short native checks and focused race checks pass. Full remote-FP verification
is recorded in PORT_PROGRESS.md; no Java wire protocol is introduced.

Fingerprint failover warnings now use ToolIO, matching the original manager and
callables. Configured system streams and tool-mode message collection retain
the exact two println calls, embedded failure-detail newline and null-message
text. Scalar, batch and statistics routing checks and the original manager
translations pass. A short TCP test closes an endpoint before any insertion,
then verifies surviving-server reassignment and warning capture; this does not
retry an ambiguously completed insertion. Broader FP failure models remain open.

Worker exit/readiness and distributed option warnings/errors now honor ToolIO
as well. Native CLI-assigned streams receive these messages; tool-mode callers
capture the original println messages. Exit still shuts down the executor,
cancels keepalive, unpublishes and releases the latch in source order. Focused
system/capture checks, one exact short worker-lifecycle race selection and the
existing CLI help/invalid-startup checks pass. Direct stderr exception traces
retain their original System.err destination.



These focused checks do not prove distributed completion.

## Testing and workflow

Normal full TLC session `23915` is retired with status 0, passing in 769.474
seconds at `6392374`. Log `/mnt/oldrog/tmp/tlago-static-tool-id-full-tlc.log`.
This binary predates the subsequent TLCEval and TLCCache node-cache corrections;
retain that snapshot qualification. Focused current checks are green; no new
full-workspace pass is established.

**Never combine long workloads with `-race`.** Run complete long workloads
normally. Reserve race instrumentation for short, focused concurrency checks.
Avoid broad race patterns that accidentally include long tests. Preserve the
original test bounds in both cases; do not make long tests small for convenience.

Use the local cache and temporary directory for offline Go work:

```bash
env GOCACHE="$PWD/.codex-gocache" GOTMPDIR="$PWD/.codex-gotmp" \
  GOPROXY=off GOSUMDB=off \
  go test -count=1 -failfast -timeout=60m ./...
```

Run checks appropriate to actual changes. Reuse valid results for unchanged
code instead of restarting expensive suites. Tests that inspect real network
interfaces require the unrestricted environment; do not weaken their assertions
to accommodate sandbox visibility. Poll the same live process handle while a
check runs. An observation timeout or silent log is not a terminal result.

Persistent fixtures belong in `tlc/test_vectors/`, never a directory named
`testdata`. Preserve original fixture bytes. Scratch overlays, probes, logs,
and compilation outputs stay in ignored `.codex-gotmp/` and earn no port credit.
Only reclaim regenerable compiler outputs after excluding live process references;
never delete source, vectors, or verification receipts to free space.

Update `TODO_TEST_PORT.md` manually as whole original methods become complete.
Do not regenerate it wholesale with the scratch rendering script, which loses
curated reconciliation notes. Update `PORT_PROGRESS.md` with concise audit and
verification receipts, then commit coherent, verified work. Do not back up,
rebase, stash, or push; the user pushes separately. Git writes require sandbox
escalation, but commits are already authorized. Do not request authorization
again for routine work within this scope.

## Keeping this handoff readable

Maintain this file as a current restart guide. Replace obsolete status instead
of prepending another overlapping history block. Put detailed per-run chronology
in `PORT_PROGRESS.md`. Use ordinary sentences, spaces between words and numbers,
clear tables for comparable results, and backticks for identifiers and commands.
Do not compress prose into strings such as "all32", "exit130", or "PASS10.068s".

Before resuming work, read [PLAN.md](../PLAN.md), this handoff,
[PORT_PROGRESS.md](PORT_PROGRESS.md), [TLC_ARCH.md](TLC_ARCH.md), and
[TODO_TEST_PORT.md](TODO_TEST_PORT.md). Choose the next action from current
pending entries and evidence. The overall porting goal remains incomplete; neither
the green batch nor the high translation percentage establishes completion.

Fingerprint RPC failures use the shared native Go failure payload rather than
flattening errors to strings. Nullable messages, cause/suppressed-error graphs,
sharing and sender Go stacks survive the boundary. Fatal returned or panicked
storage failures receive the remote I/O category with the original failure as
cause; ordinary failures retain their existing catch traits. This is TLC
behavioral parity over Go transport, with no RMI or Java serialization support.

Assigned-block checkpoint verification now also reopens a native TCP FP store
from its committed file, starting with an empty new table. Queue, trace and both
fingerprints recover consistently. The Java CLI recovers before coordinator
publication/FP registration; fresh-process remote-FP CLI recovery remains a
verified source startup limitation, not a feature proved by this method-level
check.
