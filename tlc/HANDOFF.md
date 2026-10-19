# TLC Port Handoff

Updated: October 19, 2026. Active branch: `master`.

Alpha preparation: `make build`, `make test`, `make vet`, and `make dist` now
provide the local build, guarded test, static-check, and native archive paths.
The alpha guide is `ALPHA.md`. A fresh-directory Linux amd64 bundle smoke check
covers success, invariant violation, default config discovery, sibling imports,
and relative paths with spaces. The CLI now normalizes qualified relative model
paths through its existing absolute-path resolver. No dedicated version command
or bug-report versioning was added. See the progress log for validation receipts.

Off-heap flusher lifecycle now follows upstream issue #1470 commits
`4c83d5e63`, `8c9533d44`, and `210df3af8`: concurrent flushers belong to
individual operations, inherited invariant/recovery flushing stays sequential,
and final checks use initialized reader counts and close preparation executors.
See `PORT_PROGRESS.md` for the review and validation.

Current testing policy (user, October 18): tests longer than 10 seconds are
forbidden. Preserve their bodies and use `t.Skip("test is too slow")`.
The routine `tlc` suite must finish under two minutes. This overrides all older
instructions below to execute full long bounds. The two sequential heap jobs
were terminated with exit 143 at the user's request; do not resume or restart
them. Historical results and translation credits describe past verification,
not permission to run slow tests. Slow source methods and stress-tagged entry
points are now skipped, with their original bodies and bounds retained.
The reduced default suite passes in 47.296 seconds on CPUs 0-7; its slowest
completed test took 8.39 seconds. A temporary watchdog enforced a 9.5-second
leaf-test cutoff and a 118-second package cutoff during verification.

Native test path observations now share `canonicalNativeTestDirectory`:
resolve the existing directory before appending an uncreated checkpoint name,
and use the absolute resolved path for `strace` filters, descriptor comparisons
and canonical cleanup warnings. Application I/O retains its original paths.
Startup and cleanup checks share `nativeTestDescriptorsUnder`, so symlinked
temporary directories cannot hide live handles. This correction was reviewed
by source analysis only at the user's request; remote verification is pending.

CPU limit: all port builds, tests and comparisons must share CPUs `0-7` via
`taskset -c 0-7`, for at most eight CPUs total. Surviving workloads were pinned
across all threads after the user's resource correction. The user killed a Java
process; revalidate recorded handles before relying on their prior live status.
Do not restart the killed workload without explicit authorization. Preserve
original worker counts and workload bounds when applying affinity.

The current source-surface pass restored ModelChecker and DFID worker startup,
continuation, assumptions, reporting, summary and Tool-taking successor
boundaries, including both DFID `DoNext` forms and its source cleanup form;
both checker implementations now expose their protected `ModelCheckImpl`
boundary, while the public `ModelCheck` wrapper applies the stored error code;
AbstractChecker's worker-array coverage-reporting form is restored;
all translated AbstractChecker instance and register boundaries now enforce
Java's receiver-first `NullPointerException`, including constant-returning and
unsupported base methods;
ModelChecker checkpointing, recovery, cleanup, statistics reporting, worker
coordination and periodic-work paths likewise reject nil receivers instead of
returning successful defaults;
checkpoint and recovery also require their queue, trace, fingerprint and
enabled-liveness owners at each source dereference, including typed-nil stores,
while rereading the public owners between phases;
concurrent trace commit now requires its embedded base only after all worker
checkpoints commit, preserving source mutation order before marker publication;
its synchronized level, begin-checkpoint, enumerator and reconstruction phases
also require that base before acquiring their translated trace monitor;
concurrent close retains native worker-descriptor cleanup first, then requires
the inherited base trace at Java's final close phase; base `TLCTrace.close`
remains outside the trace monitor like its unsynchronized source method;
cleanup requires the constructor-owned fingerprint set, trace, optional
liveness checker and state writer, and its explicit-checkpoint predicate
requires the queue;
when liveness is enabled, startup tautology detection, initial-state
publication, periodic and final checking, and liveness statistics all require
the constructor-owned `LiveCheck` rather than treating a missing owner as
disabled liveness;
progress reporting and worker startup require the constructor-owned fingerprint
set instead of fabricating a zero count or skipping `IncWorkers`, and startup
and coordinator joining require every published worker slot instead of skipping
null entries, including a public-array replacement between those phases;
successful outdegree reporting likewise requires every worker and its captured
statistics owner instead of omitting null entries;
each worker's out-degree statistics title is the constructor's initial
`TLC Worker <id>` thread name, captured before Java applies its later
`TLCWorkerThread-<id>` runtime name;
ordinary/model-checking, DFID, liveness, distributed and simulation worker
constructors narrow IDs, worker counts, DFID target levels and simulation trace
depths to signed Java `int` before storing or using them;
the synchronized global worker-count setter and incrementer also retain Java
`int` conversion and signed addition overflow;
global maximum-set-size validation narrows direct native inputs before applying
the source positive-bound check;
no-action validation, periodic suspension/resume, depth termination, successor
and initial-state publication, and error shutdown require the live state queue
at their source phases instead of skipping missing or typed-nil queues;
RunTLC no longer preflights that live queue before worker startup; completed
workers retain their constructor-captured queue, and the periodic fast path only
requires the live queue when Java reaches a queue operation;
ordinary and DFID RunTLC depths narrow to signed Java `int` before the
below-two startup guard;
their protected continuation helpers also narrow coverage countdowns and depths
before zero-count coverage or BFS depth termination;
RunTLC likewise starts workers before the periodic phase requires the live Tool,
preserving captured worker execution and coordinator error timing;
coordinator completion inspection and periodic coverage require the embedded
AbstractChecker owner instead of treating its absence as completion or no work;
periodic work samples the checkpoint timer and then requires its Tool before
testing the no-work fast path, matching Java's periodic-expression lookup;
DFID post-level termination also requires its indexed worker, and DFID completion
and coverage require the same embedded base owner at their source phases;
model-checker error reporting requires its trace before queue shutdown instead
of silently omitting the behavior trace when both trace owners are missing;
both concrete checkers require their Tool when constructing initial-error
diagnostic replay instead of returning a fallback result for a missing owner;
both concrete ModelCheckImpl paths now perform recovery before their first Tool-
dependent assumption/initialization phase, without a native Tool preflight;
DFID cleanup now follows Java's three source regions: recovery, tautology,
assumption and non-exceptional initialization returns escape without cleanup;
initialization exceptions and no-action success clean explicitly; and the later
search region always cleans when it exits;
ModelChecker initialization exceptions likewise become reported TLC result
codes after replay, with only a subsequent cleanup failure escaping as a Go
error;
ModelChecker search and final-liveness exceptions are reported and consumed by
Java's outer `catch(Exception)`, while Java `Error` values bypass that catch and
escape after final summary and cleanup;
the search `finally` runs summary, optional liveness statistics and cleanup as
one ordered sequence, so an earlier finalization failure skips later phases and
a cleanup failure replaces an already pending throwable;
DFID's search `finally` likewise prints its summary before cleanup in one
ordered defer, with the same failure short-circuit and replacement behavior;
DFID worker-run and terminal-liveness failures share the source outer
`catch(Exception)`: ordinary exceptions are consumed after reporting,
`LiveException` retains its code, and Java `Error` values escape;
both search regions also route direct panics through those catch boundaries
before their ordered finalizers, preserving the Exception/Error distinction;
DFID's separate pre-search `catch(Throwable)` now covers liveness-tautology
evaluation, assumptions and initialization, including direct panics, while its
ordinary early-result paths still return without cleanup;
ModelChecker's initialization `catch(Throwable)` likewise includes its progress
setup and the postcondition after a non-exceptional initialization result, with
diagnostics and cleanup performed after that isolated region reports a failure;
both concrete cleanup methods now stop at the first checkpoint or close failure,
so later owners and metadata deletion are skipped as in Java's straight-line
cleanup methods;
cleanup is no longer suppressed by a native completion latch; a later call
repeats Java's source sequence, including retrying the phase that previously
failed;
heap-backed fingerprint insertion now releases its owned write stripe only on
Java's normal return paths; a flush failure retains every acquired stripe plus
the flusher-selection and forced-flush state because the source exits before
those releases and resets; heap disk lookup failures likewise retain the read
stripe acquired by `Put` or `Contains` because Java unlocks only after a
successful lookup; `CheckFPs` also retains all table stripes when its initial
flush fails and releases them only after a successful flush; disk binary-search
I/O failures emit Java's optional negative-midpoint GENERAL event and mandatory
`SYSTEM_DISKGRAPH_ACCESS` event before propagating the same failure; disk merge
index sizing preserves signed-long overflow, coded GENERAL rejection and
32-bit Java narrowing before array allocation, and its off-heap exact-page
adjustment preserves Java `int` overflow and typed negative-array allocation;
disk lookup distinguishes a nil index from a retained empty index, incrementing
the lookup counter before the latter raises Java's array-bounds failure; the
off-heap disk-offset path preserves the same nil/empty distinction and typed
empty-index failure, and selects a disk reader only after index-only return
paths have been exhausted;
public off-heap initialization, checkpoint-warning, membership, recovery, close,
unexport, verification, capacity/statistics, force-flush and worker-registration
boundaries reject a nil receiver before performing their first source-body
operation;
the `NonCheckpointableDiskFPSet` initialization, close, unexport and
named-checkpoint warning boundaries enforce the same receiver-first failure
before delegation or output;
the `DiskFPSet` initialization, membership, recovery, verification,
checkpoint, close, unexport, worker, exit and management/statistics boundaries
likewise reject a nil receiver before no-op returns, argument validation or
source-body work;
the concrete heap-based, LSB and MSB initialization, close and unexport wrappers
also reject their own nil receivers before dereferencing Go's embedded base
pointer;
the MSB disk `TLCIterator` is exposed as `MSBTLCIterator` with its public
constructor and four methods; it preserves a null backing array separately from
a non-null empty array, traversal methods raise the translated null failure and
`Reads` remains valid on the constructed iterator;
both off-heap indexer lookup forms raise the translated null-receiver failure
instead of an untyped native panic;
the bit-shifting indexer constructor preserves Java assertion types and the
fingerprint-space assertion; zero positions retain a defensive native rejection
because the source constructor does not terminate for that input;
the off-heap iterator's bounded traversal form rejects a nil receiver with the
translated `NullPointerException` before reading its current position;
the primary `MemFPSet` applies that receiver-first failure across its close,
unexport, initialization, membership, block, worker, exit, invariant,
checkpoint, recovery and statistics boundaries;
`MemFPSet1` independently enforces the same receiver-first contract across its
close, unexport and corresponding set, checkpoint and recovery surface;
`MemFPSet2` independently enforces that contract across its close,
unexport, compressed-table, checkpoint and recovery surface;
`NoopFPSet` enforces the same contract even for close, unexport,
constant-returning and no-op methods, so nil receivers cannot silently succeed;
`MultiFPSet` enforces receiver-first failure before aggregate child traversal,
block validation, checkpoint/recovery work and its inherited no-op thread hook;
retained null or typed-nil nested fingerprint sets remain observable through
the source accessors but raise translated null failures at each actual child
method invocation, including parallel traversals;
off-heap `Sizeof`, `CheckFPs` and `CheckInvariant` retain the subclass's
unsynchronized/no-op table-lock behavior and do not acquire the native per-set
eviction mutex; the Phaser advance action likewise invokes off-heap eviction
without an additional per-set lock;
all concrete FP-set batch lookups preserve Java's plain `statesSeen += size`
as separate atomic snapshots, allowing the source's lost-update behavior
without introducing a Go data race or taking the set's main mutex;
heap disk bucket-capacity allocation uses the same separate-snapshot treatment
for Java's plain long and does not acquire the native object mutex;
disk grow-count and flush-time statistics likewise use signed Java-width
separate snapshots without that mutex and update flush time only after success;
the named-checkpoint success marker uses the same signed 32-bit snapshot
representation and advances only after the copy succeeds;
disk `AddThread` allocates and copies the replacement reader array before
opening its new final slot, publishing only after that open succeeds;
named disk fingerprint recovery relies on externally exclusive access without
acquiring table locks; its checkpoint/current file owners and trace replay's
enumerator follow the native descriptor-rollback rule on failure while retaining
Java's field mutation, error precedence and reader-replacement order; trace
replay itself is unsynchronized like the source;
`DiskFPSet.Close` closes retained reader slots without table locks like Java;
search calls assign their returned result only after successful completion, so
an uncaught Java `Error` preserves the prior result used by the ordered
`finally` success decision in both checkers; DFID also skips its post-level
`Done` reset when `RunTLC` throws;
ModelChecker periodic work resumes its suspended queue only after a successful
non-checkpoint pass; exceptional liveness, expression and ratio exits leave it
suspended, while checkpointing retains its own source-ordered resume phase;
periodic liveness timing and runtime-ratio updates use Java's signed epoch
millisecond samples and whole-millisecond arithmetic rather than Go's monotonic
nanosecond elapsed time;
liveness graph check completion messages use the same signed
epoch-millisecond start/end samples for both violation and successful paths;
liveness disk-graph degree scans plus plain/tableau string and DOT rendering
print `SYSTEM_DISKGRAPH_ACCESS` and raise Java's status-1 process exit on checked
I/O instead of returning or discarding the failure;
checker and simulator statistics plus direct `TLCGet("duration")` likewise
subtract signed epoch milliseconds before integer division by 1000, while
retaining their separate narrow-to-`-1` and 32-bit-cast policies;
the ordinary TLC runner and distributed server final banners also format the
signed difference of epoch-millisecond samples, matching their Java process
and server lifecycle boundaries;
coverage overhead selection, disk fingerprint-set flush counters and
`RandomAccessFile` operation counters use Java's signed wall-clock millisecond
subtraction; the random-access time fields are millisecond integers like their
source `long` fields;
trace-exploration generation owns the recovered-checker suppression after
confirming a multi-state error trace; the runner still invokes that method from
its Java-ordered finalization phase;
BFS progress-rate reporting uses the same epoch-millisecond subtraction and
whole-millisecond progress interval as Java; its continuation depth comparison
also applies to every integer without a native nonpositive-depth sentinel;
the explicit-start `PrintSummary` overload treats every negative millisecond
value as Java's periodic-rate sentinel while zero and positive values remain
real epoch timestamps;
DFID summary reporting rereads the fingerprint-set size for each emitted
message instead of caching across Java's separate calls;
the public success-reporting helper requires a non-null, non-typed-nil
fingerprint set at Java's first dereference instead of fabricating a zero-state
success report;
the shared double formatter used by success probabilities and action-graph
weights now uses Java's plain/scientific notation boundaries, exponent spelling
and closest two-digit selection when the shortest significand has one digit;
4,148 bounded diagnostic values match the installed JDK 21, including notation
boundaries and subnormals, without claiming exhaustive floating-point parity;
bucket-statistics reports use decimal half-up rounding and localized digits and
separators at each source Formatter call, while retaining the source's literal
`%n` text outside those calls; complete reports match pinned Java for English,
German and Arabic with normal and custom line separators;
both recovery paths reread the fingerprint-set size after printing the recovery
message before publishing the recovered-state counter, matching Java's two
separate calls;
ModelChecker initial-state reporting likewise reads the fingerprint-set size
once for branch selection and again for the distinct-count parameter on the
two-count message;
non-exceptional initialization-failure postcondition checking also requires the
ModelChecker and Tool before choosing the initial-state counterexample form;
ModelChecker successor-error replay likewise requires its receiver and Tool and
enters the diagnostic path even when the recorded predecessor is missing;
initial and successor storage require their fingerprint set and state writer,
including typed-nil interfaces, and the direct no-worker compatibility path
requires its base trace before writing an unseen state;
its public/protected model-check, initialization and successor wrappers plus
their directly exposed implementation forms reject the receiver before their
panic-to-error conversion boundaries;
DFIDModelChecker now applies the same receiver boundary across model checking,
worker coordination, initialization, successor generation, error replay,
periodic work, checkpoint/recovery, cleanup and reporting; its fingerprint set
and state writer are required at Java's direct accesses, including typed-nil
writers, and enabled liveness requires its checker instead of becoming a skip;
DFID worker exhaustion and failure handling also require the checker callback
owner instead of silently suppressing stop, error and completion publication;
the complete translated DFID worker method surface rejects a nil receiver before
locking, random selection, stack access, trace output or lifecycle handling;
Worker.Run resolves its Java-final checker, Tool and captured queue owners before
the loop and always publishes queue shutdown on completion, stop and failure,
including typed-nil queue interfaces;
Worker.DoNext requires its final Tool and checker, its liveness stop requires the
captured queue, and liveness output rejects typed-nil writers after inserting
the stuttering state in Java's source order;
worker successor callbacks wrap a missing checker as Java's successor-associated
runtime failure after counting the state, and unsatisfied output rejects a
typed-nil captured writer;
Worker recovery and unsatisfied-state names plus its two state-writing overloads;
ConcurrentTLCTrace overrides, records and protected reconstruction; and TLCTrace
level, prefix-printing and behavior-writing forms.
SimulationWorker's protected explicit-RNG selection forms and source statistics
post-trace form are also exposed. Its protected interrupt checkpoint is exposed
and owns all simulation-mode trace interruption checks. Its protected post-trace
hook retains the source boolean contract in ordinary and RL modes. Core trace
execution, RL-mode behavior and indexed/named register methods now uniformly
enforce Java's receiver-first `NullPointerException`.
Indexed worker/simulator register arguments, extended-state cache keys and RL
disabled-action indices narrow to Java `int` before lookup, growth or bounds
checks.
`AbstractChecker` worker-array indices likewise narrow before selecting BFS or
DFID register owners for indexed and named operations.
`AbstractChecker` register reads, writes, broadcasts and aggregate snapshots no
longer acquire the checker monitor, matching the source API. Native per-worker
register storage still prevents Go data races; multi-worker aggregation may
observe interleaved worker updates as Java does.
`AbstractChecker.SetDone` likewise remains outside the checker monitor. Its
completion field uses atomic snapshots, but the method keeps Java's separate
read then write rather than an atomic exchange, so concurrent callers may both
observe the old false value. Error-state mutations retain caller-equivalent
checker monitor ownership.
Worker error reporting, trace printing and postcondition replay now hold that
same reentrant checker monitor through their source synchronized sections,
including safety-like liveness diagnostics. Coordinator stop, suspend and
resume calls cannot interleave with those sections.
The breadth-first outer next-state failure handler classifies the throwable
before claiming the error state, then inspects its message and parameters only
if that claim wins, matching Java's reporting order.
DFID successor diagnostics also hold that checker monitor across the source
error-state and trace-printing sections, including continuation, deadlock,
property evaluation and the outer next-state failure handler.
That outer handler constructs its GENERAL diagnostic only after winning
`setErrState`, so a losing worker does not format or inspect its throwable.
The `DFIDWorker.run` catch holds the same monitor from error-state publication
through GENERAL output and completion notification, after its source-order
`setStop(2)` call.
Simulator's protected state-run, behavior-printing and error-policy boundaries
are exposed as well. All translated Simulator instance methods now enforce
Java's receiver-first `NullPointerException`; scheduler selection remains
instance-independent because the source helper is static.
Tool now retains and exposes the resolver used by the production application
loader, its static probabilistic-mode query uses the source name, and its
no-argument initial-state collector, function-context overloads and ordinary
context-enumeration overloads are exposed. Its value and operator-context methods
also support the source forms without explicit cost models, and its
context-taking next-state overload uses the caller's evaluation context. Initial
and successor rejection also route through the protected unsatisfied-state
hooks. Predicate validity supports the source form without an explicit context.
Module-file traversal treats a typed-nil resolver as Java null at the first file.
The native application-node function-context helper, its implementation body
and the ordinary context-enumeration implementation require a Tool owner before
argument, function or quantifier inspection.
Tool's action-item and action-list `Enabled` overloads are exposed, and recursive
evaluation plus liveness use the cost-aware source boundary. Its protected
application and `UNCHANGED` implementation boundaries are distinct from their
call-stack wrappers. Tool's protected value-source hook is exposed and owns
evaluator source assignment.
`Tool.GetActions` retains Java's backing-array identity and raises the typed
`NullPointerException` when invoked through a nil tool.
Action constructors eagerly allocate their source final auxiliary map, including
`UNKNOWN`, and `SetID` stores Java's signed 32-bit value.
The folded `Spec.getPrimedLocs` and `collectPrimedLocs` methods likewise require
their Tool receiver, and collection rejects a null predicate at Java's immediate
`getKind` boundary while retaining the established traversal behavior.
Folded `getVar` and `getPrimedVar` dispatch also requires a Tool receiver while
retaining Java's null-expression result for a valid Tool.
The three symbol lookup entry points now enforce the same receiver-first failure
before their established symbol, context, cutoff and state resolution.
The theorem/assumption context helper, both level-bound methods and randomized
context enumeration also enforce receiver-first Java dispatch while retaining
their established valid-Tool behavior.
Public `Eval`, `EvalPure`, and the directly exposed implementation/application
evaluation boundaries now require a Tool owner; native signature validation at
`Eval` still precedes receiver access as documented.
DebugTool's evaluator wrapper, protected implementation and constant-, state-
and action-level evaluator methods likewise require their Tool owner before
fast-tool selection, mode switching or debugger-frame work.
`Tool.SetDebugger` exposes the source mode switch and returns the prior mode;
debugger-triggered evaluation uses it before restoring mode and state-tool data.
`ToolMode.String` and `ToolMode.Ordinal` expose Java `Tool.Mode` names and
ordinals explicitly: `Simulation=0`, `MC=1`, `MC_DEBUG=2`, and `Executor=3`.
The native `MC_DFS` execution distinction has no Java enum ordinal and returns
`-1`; this preserves `ModeMC` as Go's zero value.
The nested debugger placeholder value exposes its source kind and type-string
methods; both return the user-facing pending-evaluation text, and debugger
variable rendering obtains the type through that method.
State completeness plus aggregate/individual model and action constraint methods,
including their directly exposed implementation forms, likewise require a Tool
owner before inspecting states or constraints.
Expression/assumption validation, all three action-validity overloads, and
assumption checking now apply that owner boundary across wrappers and directly
exposed implementation forms.
Reward evaluation applies the same boundary in its public and implementation
forms while retaining the configured-expression and fallback behavior.
Initial-state collection, state construction, vector/functor successor
generation and their directly exposed init/make implementations require a Tool;
the per-action functor form also retains Java's required action/predicate access.
Protected-equivalent init/successor predicate, action-list and application
wrapper/implementation boundaries enforce the same owner before call-stack,
debugger, coverage or enumeration work.
Protected-equivalent `ENABLED` predicate, action-list, application and
`UNCHANGED` wrapper/implementation boundaries plus next-state
`ProcessUnchanged` wrapper/implementation boundaries also require a Tool owner
before call-stack, debugger, coverage, evaluation or terminal-list work.
The source identity and mode getters also reject nil receivers rather than
inventing tool ID zero or model-checking mode.
Tool's source constraint, initial/spec-action, invariant and temporal property
getters likewise reject nil owners while retaining their established slice and
vector ownership for real tools.
Its root/configuration/spec-processor metadata, implied init/action, next-state
and view getters now enforce the same source receiver boundary.
Alias, postcondition, constraint, assumption, RL and counterexample-definition
getters also reject nil tools without changing valid configuration resolution.
Symmetry permutation/presence and liveness-triviality queries enforce the same
receiver contract while preserving valid empty-spec results.
The source `NoDebug`, `IsDebugger` and module lookup methods reject nil owners;
ordinary tools retain their existing non-debug identity and optional lookup.
The state-returning `EvalAlias` overload also rejects a nil Tool before applying
the existing no-alias identity behavior.
The next-state predicate, application and `UNCHANGED` implementation boundaries
are likewise distinct from their call-stack/debug wrappers.
TLCTrace's protected fingerprint-vector reconstruction overloads are also
exposed and own the disk recovery path. Concurrent trace enumeration implements
the source error-returning reset contract, and worker trace enumeration exposes
`HasMoreFP` with its checked-I/O result.
Direct `ConcurrentTraceEnumerator` and `WorkerTraceEnumerator` method calls
reject nil receivers before returning through their checked-I/O channels.
Base disk-enumerator reset reads the outer RAF length without acquiring the
outer trace monitor, matching the source inner-class method.
The base location-reconstruction overload and both concurrent reconstruction
overloads reject nil trace receivers before state or arity handling.
The base trace `GetLevel` overload adapter likewise checks its receiver before
rejecting excess native arguments.
Complete disk-trace reconstruction materializes Java `HashMap<Long, ...>`
values order through the shared OpenJDK-derived map instead of Go map order.
Base and concurrent trace `Close` calls reject nil receivers; explicit-position
enumerator reset can still replace an absent old reader before seeking.
Base trace begin, commit and recovery checkpoint calls likewise reject nil
receivers; populated disk recovery and concurrent lifecycle behavior is unchanged.
Only base trace begin holds the source monitor; commit and recovery retain
Java's unsynchronized entry behavior.
Worker trace begin and commit remain synchronized, while worker recovery now
retains Java's unsynchronized entry behavior.
All three `MemStateQueue` checkpoint methods retain Java's unsynchronized entry
behavior after the checker has suspended queue workers.
Its plain enqueue/dequeue, size and emptiness paths also remain outside the
queue monitor; atomic snapshots preserve Java's separate plain-long updates
without introducing a Go data race or atomic-add semantics.
All three `DiskStateQueue` checkpoint methods do the same; begin still finishes
the cleaner first, and recovery still restarts the retained pool reader.
Its plain length reads and mutations use the same unsynchronized atomic-snapshot
contract as `MemStateQueue`; synchronized worker operations retain their monitor.
`DiskByteArrayQueue` follows that same unsynchronized outer-queue contract while
retaining its raw checkpoint stream and byte-pool helpers.
Its plain length paths now use the shared atomic-snapshot contract as the two
state queues; synchronized raw worker operations retain their monitor.
TLCTrace's fingerprint-only and predecessor-taking state-write overloads are
exposed; the original heap FPSet recovery translation uses them directly.
The `checkimplfile` CLI now completes its production external-trace loop. Its
temporary trace bridge uses the checked tool's `SpecProcessor`, reuses that
tool's variable declarations and processes temporary definition constants
before state construction; valid and illegal transition traces both advance to
the next polled filename. The extensionless trace file is the readiness marker,
while SANY loads the corresponding `.tla` module. Setup and loop failures use
Java's outer `CHECK_FAILED_TO_CHECK` report and successful process exit. Parsed
option errors print the source usage message and likewise return success. Trace
polling uses `ToolIO`, and export preserves `PrintWriter`'s suppressed output
errors after the destination opens. A nonzero partial-space result preserves
Java's mapped `System.exit` status, and scheduled exports retain the timestamp
captured before trace generation. The export schedule stores wall-clock
milliseconds and uses signed Java `long` subtraction, including overflow. State
checking and uncovered-trace generation
dereference their fingerprint sets, trace and tool only on the source branches
that consume them; missing required owners raise typed null failures. Violated
action properties and invariants read names through the tool getters and retain
Java's null-array and short-array failures. Trace-front-end exceptions retain
`CHECK_COULD_NOT_READ_TRACE`, while accumulated syntax or semantic diagnostics
retain `TLC_PARSING_FAILED` as the cause of the outer checker report. Metadata
directory creation uses the shared `FileUtil.makeMetaDir` port, including the
global override, collision naming and typed creation failure.
StateVec exposes the source removal/replacement names, and its `ToList` method
retains abstract `TLCState` entries instead of narrowing them to mutable states.
Explicit capacities for `StateVec` and the distributed `TLCStateVec` adapter
are narrowed to Java `int` before negative-array checking and allocation.
StateVec element access and replacement indices also narrow before array bounds
checks.
StateVec and generic Vect expose source `Stream` methods as lazy Go sequences
with fixed active ranges and shared captured backing storage.
SetOfStates action filtering returns the source set abstraction rather than a
flattened slice.
SetOfLong exposes its ignored-load-factor constructor and fails typed nil
receivers at every public collection boundary. Its constructors narrow to Java
`int`, growth preserves signed overflow and failure mutations, and zero-length
nonzero lookup raises Java's typed division-by-zero failure.
ObjLongTable, its SemanticNode specialization and LongObjTable now reject nil
receivers across all public lookup, mutation, sizing, merge, snapshot and
enumeration boundaries. Null merge sources also fail, and all three table
constructors preserve Java's negative-array-size rejection.
ObjLongTable, LongObjTable and liveness NodeTable constructors narrow capacities
to Java `int`; the utility tables also preserve signed growth overflow and
raise Java `ArithmeticException` for zero-length lookups.
Liveness `NodePtrTable`, `TableauNodePtrTable` and its reverse variant likewise
narrow constructor capacities to Java `int`, while retaining their distinct
source growth and failure-mutation order.
Direct liveness-table locations, `NodeTable` tableau-index lookups and
`TableauNodePtrTable` tableau-index inputs also narrow to Java `int` before
source validation, probing or array access.
Direct tableau-record indices and stored parent locations narrow before source
array access and parent-range validation, including the `-1` sentinels.
LiveWorker's public DOT renderer also narrows its initial tableau index before
matching graph records, preserving Java `int` identity for direct native calls.
Disk-graph initial-node storage and retained path indices do the same; tableau
graph node, pointer, link, record and path entries narrow before their source
assertions, comparisons, lookups or diagnostics.
The public `ObjLongTable.Enumerator` constructor is exposed directly, and the
`Keys` factory uses the same live-table iterator path.
FPSetFactory's synchronous source names and both `GetFPSet` overloads are
exposed; the older Go factory names remain delegates.
Its asynchronous initialized-set Future is also exposed, and the corresponding
ModelChecker constructor option trusts the Future-owned initialized store.
BitVector distinguishes its no-argument read-target constructor from an
initialized zero-capacity vector and exposes `HashCode` and `TrueCnt`.
Its iterator exposes both Java constructors: the default form remains
uninitialized until `Init`, and the vector-taking form initializes immediately.
Focused checker, DFID, worker and trace selections pass, including one short
DFID race run, and all packages compile offline. The base TLCTrace zero-argument
all-states `getTrace()` now reconstructs the trace file directly, retains the
source record-location semantics, and restores the file cursor. No translated
original test targets this overload; the focused trace and cursor selections
cover its surrounding paths. Detailed receipts and commit chronology remain in
[PORT_PROGRESS.md](PORT_PROGRESS.md).

The public `TLC` runner surface now includes resolver and state-writer setters
and getters, run-mode/main-file/start-time accessors, module-file enumeration,
and the source simulation/model-checking runtime maps. `HandleParameters`
preserves dependencies installed by Java-style programmatic harnesses, and the
CLI loader consumes an installed resolver at its explicit loading boundary.
`RunMode.String` and `StateVisualization.String` expose the exact Java enum
names in declaration order for the runner and state-writer public surfaces.
`HandleParameters`, `GetSpecName` and `GetModelName` reject nil receivers.
Both name getters use the main file as Java's fallback and honor the `specName`
and `modelName` system properties.
Typed-nil optional resolvers and state writers now follow Java null semantics:
parameter parsing does not preserve them, model-checker construction substitutes
the no-op writer, model checking skips a logically null injected writer, and
liveness construction applies the same no-op default.
Typed-nil injected fingerprint sets and state queues likewise select the
model-checker constructor's standard stores.
The pinned Java artifact's public signatures, the original `TLCTest`, temporary
focused runner probes and all-package offline compilation pass.

The public `SanyExitCode` values expose the driver enum's exact numeric codes,
and `SanyExitCodeFromCode` preserves its typed invalid-code failure.
`SanyExitException` carries that enum through the source numeric and enumerated
getters, including nullable construction and typed receiver/null-code failures.
`SanyErrorCode` exposes its standard-value predicate and metadata getters,
including the source `SanyErrorLevel` enum and variadic-parameter sentinel.
Public `SanySettings` factories and `ParseSanyWithSettings` now drive the same
frontend settings path as the CLI. Error-code sets are copied and immutable;
AST-only settings suppress linting and PlusCal translation validation.
Public `SanySemanticAbort` and `SanySemanticException` preserve the source
detail, error-log and cause identities, including nullable constructor values
and typed null-pointer boundaries.
`SanyErrors` and `SanyErrorDetails` expose the ordered diagnostic log, severity
filters, duplicate suppression, detail metadata and fresh result collections.
The legacy semantic `FrontEnd` holder and its three deliberate null stubs are
exposed through `SanyFrontEnd*`; its tool-ID function shares the production
zero-based atomic counter.
Canonical selector generation treats a missing selector-argument node as the
source zero-argument case while public syntax-node methods retain strict nil
receiver failures.

`TraceExpressionExplorerSpecWriter.getVariableExpressionMapFromTLAFile` is
exposed as `GetVariableExpressionMapFromTLAFile`. It reads with Java's default
charset and `BufferedReader` line boundaries, preserves missing-file and
file-open behavior, and returns an empty nonnil map when generated declaration
lines exist but contain no assignment.

`MultiFPSet` exposes its source child-construction, fingerprint-selection and
fresh child-array boundaries as `GetNestedFPSets`, `GetFPSet` and `GetFPSets`.
The translated original MultiFPSet and factory tests use these methods directly.

The folded `ParameterizedSpecObj.RuntimeInvariantTemplate` exposes source
construction plus `getModules` and `getAction` as
`NewRuntimeInvariantTemplate`, `GetModules` and `GetAction`; construction retains
the supplied module collection, and both CLI invariant forms use it. Production
runtime invariant compilation goes through the holder rather than duplicating it.
The folded `PostCondition`, `View` and `Constraint` carriers expose all source
constructor families, retain supplied constant maps and process every mapped
string constant; a supplied null map fails at the later source use boundary.

`JavaRandom.GetAril` exposes source-unsynchronized `RandomGenerator.getAril`;
simulator accounting uses it directly, while the prior `Aril` spelling remains
a delegate. `NextDouble` publishes Java's plain `aril++` before its two separate
synchronized `next` calls, so another RNG caller can interleave between those
seed steps. Atomic snapshots avoid Go races without turning the aril update into
an indivisible addition. Seed reset and Gaussian generation retain their source
monitor ownership.
`Next(int)` narrows its bit count and applies Java's masked long-shift distance,
including negative and overflowing inputs at the native boundary.

`TLCExtAssertError` now shares `TLCExt.class` with the source synchronized
successor picker, model-value creator and fingerprint operator. Trace
reconstruction retains its narrower synchronized block; `TLCNoOp`, defer and
the remaining unsynchronized operators stay outside the class monitor.

`MP.getMessage` rendering now uses a shared reentrant class monitor across
ordinary messages, errors, warnings, states and TLC bug messages. Recorder
callbacks still precede that monitor, while warning-history insertion and
`ToolIO` printing follow it, matching the Java release boundary. The complete
severity prefix or tool envelope and the first-warning hint are constructed
inside the monitor.
`ToolIO.setMode(int)` narrows before validating the two legal modes, so native
oversized inputs reproduce Java-width selection and rejection.
`MemoryRecorder.Recorded` and `Records` narrow query codes to Java `int`, matching
`TestMPRecorder` lookups; native `Message` injection retains the supplied code.

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

The two ignored `ActionCompositionDistributive` methods now have complete Go
translations and byte-identical model/config vectors. They remain immediate
`upstream @Ignore` skips and receive no passing-method credit. The active
`ActionCompositionA/B` translations retain their inherited debugger mode; only
the ignored distributive classes disable it, matching the Java constructors and
overrides.

`Action` now exposes both source `getLocation` forms through variadic
`GetLocation`, and its source-named `GetDeclaration` and `GetDefinition` return
structured `SourceLocation` values. Coverage, debugger, simulator, state-info
and counterexample consumers use these methods. The older `*Location` and
`GetLocationNamed` spellings remain compatibility delegates.
Source-location construction narrows all four line/column coordinates to Java
`int` before storage and diagnostic formatting.
Native trace-metadata fixtures construct actions through the source-compatible
constructor so definition formatting never relies on an impossible nil
predicate.

`CallStack.Freeze` now represents only Java's zero- and one-argument overloads;
extra fingerprint exceptions fail before the stack is frozen. The translated
action-composition, action-level, cyclic-redefinition, incomplete-next,
possible-action and fingerprint-exception model selections pass.
CallStackTool source assignment also dereferences null values like Java; the
ordinary FastTool/DebugTool-shaped path continues to return them unchanged.
Its evaluation wrappers now push null expressions too, preserving Java's
push/freeze/pop order and the frozen stack's leading-null suppression.
The CallStackTool-facing stack-presence and rendering accessors now reject nil
tool receivers while preserving ordinary tools' false and empty results.

ConfigFileException, WorkerException, StatefulRuntimeException and LiveException
constructors now accept only their Java overload shapes. Their cause-taking
forms normalize typed-nil Go errors to Java null causes.

Defns now enforces its two Java constructor forms and String/UniqueString key
surface, plus the internal SymbolNode-to-name adapter required by the native
parser. Nil receivers and keys fail at invocation, unsupported Go key types are
rejected, and invalid backing-array sizes and negative stored locations use the
corresponding typed Java exceptions. Explicit definition counts, constructor
addition, index increments and resize calculations retain signed Java `int`
semantics and their source failure mutations.
Specs level and substitution helpers now retain Java's null-dereference timing
and typed substitution-cast failures; the SubstIn copy constructor rejects a
null source directly.
The utility List now rejects nil receivers and null copy/append operands across
its Java method surface, and rendering dereferences null elements like Java.
TLCStateFun's constant-returning and unsupported-operation methods now reject
nil receivers before returning or raising their method-specific exception, and
its unassigned-variable result is a concrete empty collection.
TLAPlusExecutor and its Mapping now retain Java's null invocation failures on
mapping mutation/context access and executor map, step, lock and constant access;
the native optional definition-table constructor accepts at most one table.
`GetLock` exposes the source accessor, while `Lock` remains its compatibility
delegate and returns the same reentrant lock.
`TLCDebuggerFactoryGetInstance` exposes the nested debugger factory method;
tool attachment delegates through it, preserving override identity and fresh
instance suspend/continue and halt configuration.
`ErrorConstantToExitStatus` and `ExitStatusToCrash` expose the nested
`EC.ExitStatus` methods, including Java `int` narrowing and the source's exact
known-status set.
The complete public `MP.Colors` ANSI table is exposed through 57
`MPColors...` constants with byte-identical escape sequences.
ContextEnumerator now rejects null enumeration arrays and nil receivers, uses
typed array bounds for short variable arrays, and reports invalid binding-array
types as Java ClassCastException through the existing checked error channel.
Context lookup, conversion, rendering, accessor, predicate, depth and copy
methods now reject nil receivers. `Context.Empty.depth()` also preserves the
source null-child dereference, while its other empty-context behavior remains
unchanged.
StateFunctor and NextStateFunctor source methods now reject nil receivers before
dispatching callbacks or defaults. The default GetStates behavior is tested on
a concrete empty functor, matching Java interface invocation.
Trace alias evaluation now preserves Java's distinct supplier and pairwise
overloads: only the supplier form installs the lazy imported-trace binding.
Nil Tool receivers and enabled-alias null inputs fail at the corresponding Java
dereference, while disabled alias evaluation still returns its input unchanged.
Native postcondition fixtures use a concrete null semantic predicate so their
actions remain renderable through the stricter source-compatible Action path.
TLCStateInfo's action-taking constructor now accepts only Java's zero- and
one-action overload shapes instead of silently ignoring extra Go arguments.
Buffered data-stream constructors likewise accept at most one stream, and
`ReadFully` accepts only Java's whole-array or offset-and-length forms.
Buffered input/output byte ranges, input string/skip lengths and output character
ranges narrow offsets and counts to Java `int` before their existing source-order
copy, cursor and failure behavior.
SimpleCache accepts only its zero- and one-size constructors, and its public
cache lookup/counter methods reject nil receivers like Java invocation.
`GetHitRatioAsString` performs that receiver check before initializing its
package formatter state, preserving Java's receiver-first method boundary.
SetOfStates constructor and insertion adapters accept only the zero- or one-tool
forms corresponding to Java's distinct overloads.
Its explicit capacity is narrowed to Java `int` before negative-array checking
and allocation, preventing 64-bit Go sizes from bypassing Java overflow behavior.
ValueVec's optional-capacity adapter accepts only Java's zero- and one-capacity
constructor forms. Explicit `ValueVec`, `Vect` and `LongVec` capacities are
narrowed to Java `int` before negative-array checking and allocation.
`LongVec` indexed access and removal also narrow before range and backing-array
checks, preserving Java's distinct positive and negative index failures.
MemStateQueue's optional metadata-directory adapter likewise accepts only its
two Java constructor forms.
TLCTrace metadata construction is bounded to the port's supported zero-, one-
and two-string forms, and ConcurrentTLCTrace accepts at most one worker override.
SynchronousDiskIntStack's optional-capacity adapter now accepts only Java's
default- and explicit-capacity constructor forms.
Packaged-model and filename-resolver adapters accept at most one classpath or
option set, preserving their captured Java class-loader configuration.
Named input streams, TLA files and simple filename resolvers reject nil receivers
with typed `NullPointerException` before getters, resolution, rendering or
reference-counted close work; typed-nil module-override resolvers count as null.
The native TLA class loader retains its supplied resolver interface, while a
typed-nil resolver skips resolver lookup and uses the source system-classpath
fallback.
Value input/output file constructors accept at most one explicit compression
flag, matching Java's default and boolean constructor families.
The shared BaseValue constructor path accepts at most one CostModel, enforcing
the no-model/one-model overload boundary across composite value constructors.
The operator-lambda, lazy, set-predicate and function-lambda constructors plus
product, K-subset and materializing enumerators enforce the same boundary;
function-set domain enumeration accepts at most one retained domain vector.
CostModel invocation and secondary counters accept only Java's no-value and
one-value overloads, including when recording is disabled.
CostModel also exposes the source wrapper's semantic-node and root queries;
child insertion keys through the restored node accessor. `AddLets` restores
the LET-wrapper insertion form, deriving its key from the supplied wrapper and
returning the receiver; coverage creation uses that source path while `AddLet`
remains available to existing Go callers.
`ActionWrapper` is exposed as the folded `CostModel` handle, and
`NewActionWrapper` accepts the exact four-value `ActionWrapperRelation` enum and
retains the action, mapped reporting relation and fresh substitution map.
Unlike the production absent-action sentinel, direct construction preserves a
null action until a later source dereference.
ValueExcept copy construction accepts only Java's current-index and explicit-
index forms, narrowing the copied cursor to Java `int` in both cases.
Singleton-record normalization, state-record defaults and map-function
normalization accept only their Java zero- or one-value overload forms.
SimulationWorkerError construction accepts only Java's no-exception and
one-exception forms.
Simulation worker results, concurrent trace records and liveness graph
transitions narrow their worker or tableau indices to Java's signed 32-bit
`int` range at their source constructor boundaries.
Assignment formatting, MCState conjunctive descriptions and MCError message
construction accept only their Java zero- or one-optional-argument forms.
LiveExprNode body access rejects nil receivers and excess indices, and junction
index failures use Java array-bounds exceptions instead of returning nil.
Junction counts and additions reject nil receivers. Adding a null child retains
Java's mutation-before-`NullPointerException` order.
Liveness level, action-containment and positive-form queries reject nil
receivers, and unary queries read their current child instead of stale metadata.
State-body, context and tag accessors plus tag mutation also reject nil
receivers with Java's typed null exception.
Direct and lasso liveness evaluation reject nil receivers instead of treating
a missing expression as true.
Liveness transformation, tagging and AE/EA/general classification methods also
reject nil receivers at Java's instance-invocation boundary.
Liveness equality now follows unary and junction node shape, and equality,
string and DOT rendering preserve Java's null receiver and child failures.
CheckImpl construction accepts at most one fingerprint-set configuration, and
an explicitly supplied null configuration raises `NullPointerException`; the
file-backed subclass inherits both boundaries. Constructor depth and partial-
space level addition use Java's signed 32-bit arithmetic. Its command
parses the model configuration before SANY, preserving `FastTool` failure order.
The abstract state-input and trace-export adapters reject nil `CheckImpl`
receivers before consulting their installed Go callbacks.
Property checking captures each source array length once and reloads the array
for every indexed property, including typed null and shrink failures.
Successful file-trace loading publishes a fresh non-null state array, including
for a trace module with zero operator definitions.
The production loader publishes that array before state conversion, so a later
conversion failure retains Java's partially filled replacement and old indices.
The retained slice-returning loader publishes a non-null partial result on error
with the same counter behavior.
State, input-trace and output-trace counters increment with Java signed 32-bit
overflow.
Negative state cursors preserve Java post-increment before their typed array
bounds failure.
CheckImpl captures its start time before model-checker storage construction and
passes the selected FP-set configuration through the source constructor order;
known non-instantiable implementations retain their typed null failure.
Tool's action-list Enabled and expression-validity adapters accept at most one
cost model or context, respectively.
LongArray zeroing, TLCTrace level lookup and SpecWriter source creation accept
only their Java zero- or one-optional-argument forms.
All 22 SpecWriter instance methods reject nil receivers before validation,
identifier generation, buffer mutation, stream writes or file creation.
All 21 SpecTraceExpressionWriter methods likewise require their receiver and
embedded base writer before overload validation or trace-spec mutation.
Its static `indentString` count narrows to Java `int` before indent allocation.
Its three static init/next-to-buffer entry points require the TLA destination
after building fragments and applying optional CFG-buffer side effects. The
shared builder and all delegating entry points reject a null trace while
retaining the source empty-list result.
The static trace-function writer likewise rejects a null input list or null
state element while retaining empty-tuple output for an allocated empty list.
Arrow-assignment generation requires both its assignment and TLA destination,
evaluating the assignment first; the instance and static forms share this path.
SpecWriter formula-list serialization rejects a null list before allocation and
null elements after their per-entry identifier allocation.
Spec and init/next definition writers preserve Java's partial CFG/TLA output
before typed null-array or array-bounds failures at each indexed access.
Raw constants, both assignment-constant passes and model-value-set emission now
require their Java collection and element owners while retaining empty inputs.
Formula-list emission requires its outer list and both required fields per row,
preserving CFG header/label writes before typed row failures.
SpecTraceExpressionWriter requires explicit extra-module, stub-variable and
expression-data collections plus a CFG owner for view configuration. Generated
TE specs allocate Java's nonnull empty raw-constants list.
Trace-property and info-comment generation now require their trace/data owners,
state entries, valid lasso indexes, state variables and property CFG output;
each property helper writes its CFG declaration before any property TLA.
Trace-function INIT/NEXT generation preserves Java's staged variable, trace and
multi-state final-state failures after their preceding CFG/TLA output.
The static init/next builder uses typed state, back-target and variable-array
access while retaining a null next iterator element as the source loop sentinel.
TraceExpressionExplorerSpecWriter construction requires its expression list;
an allocated empty list still generates the complete empty explorer module.
Trace configuration wrapping requires its CFG buffer, and negative indentation
counts raise Java's typed negative-array exception.
ErrorTraceMessageRecorder exposes source `getMCErrorTrace` as
`GetMCErrorTrace`, enforces receiver dispatch, replaces state ordinals
unconditionally and parses both back-state payload forms with Java int bounds.
BroadcastRecorder exposes source `subscribe` and `unsubscribe` as `Subscribe`
and `Unsubscribe`, requires its receiver, and preserves HashSet-style null
subscription, idempotence, removal and dispatch-time failure semantics.
StatePrinter narrows runtime, explicit invariant and back-state ordinals to
Java `int` before formatting and recorder publication.
The shared MP state-info printer also narrows its public ordinal argument before
publishing direct state events.
MP message-code entry points and exit-status lookup narrow their public codes to
Java `int` before recorder notification, formatting, return and classification.
Core TLC runtime, configuration, evaluation, liveness, simulation-worker and
process-exit carriers retain error codes only after the same Java `int` narrowing.
AbstractChecker error state, the distributed server's native coded extension and
DFID stop propagation likewise store only narrowed Java `int` codes.
DFID trace-format dispatch and simulator fatal-error classification also narrow
before comparing codes, so rendering and control flow use the same Java value.
MP suppression and warning-as-error controls narrow codes on insertion, lookup
and read-only view membership, including the native SANY code representation.
MP throwable error, cause, parameter, coded-runtime and warning entry points
narrow before GENERAL branching, recorder publication, formatting and return.
TraceExpressionExplorerSpecWriter also requires its receiver, base writer and
expression map across its five configuration/generation helpers.
TraceExplorationSpec generation and direct spec writing require their receiver
before applying the established empty-trace and missing-output shortcuts.
LongArray zeroing also requires its receiver and preserves Java's zero-thread
division failure and negative-thread executor argument failure before any
empty-array shortcut.
LongArray construction from a collection, size/string/range/array conversion,
and LongArrays sorting/conversion reject null owners. Explicit-comparator sort
dereferences a null comparator only after the source empty/no-iteration checks;
the public default comparator is exposed through `NewLongArraysLongComparator`
and `Compare`, and range sorting accepts Java's with/without-comparator overload
forms. Range rendering keeps only Java's `end == -1` empty shortcut.
DotActionWriter methods reject nil or uninitialized receivers, and edge writing
accepts only Java's zero- or one-weight overload forms.
Action-node and edge identifiers narrow to signed Java `int` before formatting.
Its node and edge writes share Java's writer monitor; subgraph delimiters and
close remain outside that monitor like their source methods.
Limiting and static distributed block-selector adapters accept at most one
maximum or block-size override.
All BlockSelector instance methods now reject nil receivers with the typed null
exception. Protected sizing, average-statistics mutation and limiting-maximum
helpers are exposed as `GetBlockSize`, `SetAverageBlockCnt` and `GetMaximum`.
DistributedWorkerSmartProxy's network-overhead accessor now rejects a nil
receiver instead of returning its valid unmeasured sentinel.
TLCServer's direct lifecycle, registry, statistics, configuration and FP-set
methods now reject nil receivers with the typed null exception. Its Go
remote/error adapters retain their checked-error channels, and `PrintSummary`
retains Java's static-call behavior.
Its summary level narrows to signed Java `int` before progress and search-depth
message formatting.
The public nested `FPSetManager.FPSets` registration wrapper is exposed as
`DistributedFPSets`. Its availability, endpoint, hostname, scalar/batch,
statistics and lifecycle methods own the same production manager calls as Java;
the translated original manager test now registers through `GetFPSet` and the
source-named `DistributedFPSetManager.Register` boundary.
All public `DistributedFPSetManager` methods reject a nil receiver before field
access, locking, overload adaptation, endpoint calls or executor creation.
`Reassign` narrows its slot index to signed Java `int` before bounds validation
and synchronized availability mutation.
Typed-nil endpoint interfaces normalize to Java null across wrapper construction,
registration, local-manager construction, lifecycle access and native transfer.
The local fingerprint endpoint adapter also rejects nil owners and nil or
typed-nil stores before every storage operation.
The local coordinator endpoint adapter likewise requires its receiver and
retained `TLCServer` before each coordinator call.
Fingerprint-server startup recognizes a typed-nil coordinator as Java null at
the registration dereference, after storage initialization, type reporting and
local-host resolution.
The network coordinator endpoint requires its receiver and transport client
before argument adaptation or RPC, including typed-nil registration arguments.
The network fingerprint endpoint requires its receiver before lazy connection,
request/vector construction or remote trace-argument rejection.
The network worker endpoint likewise requires its receiver before state-graph
encoding, lazy connection or request construction.
The native RPC host rejects typed-nil coordinator, worker and fingerprint
endpoints at publication/reference creation, and treats any retained typed-nil
entry as unavailable during dispatch.
Manager payload decoding preserves explicit null endpoint nodes, while a
non-null endpoint reference must resolve to a literal and typed non-null
fingerprint endpoint.
`DistributedWorkerSmartProxy` normalizes typed-nil worker interfaces to Java
null and applies that check across every delegated worker method.
TLCServerThread and its timer task now enforce the same receiver boundary across
thread controls, block processing, worker-loss cleanup, state/statistics access,
URI access and keepalive helpers. Valid repeated cleanup remains idempotent.
Server-thread construction retains the exact supplied block-selector reference;
selector invocation and retry recognize a typed-nil interface as Java null.
Server-thread queue use similarly treats typed-nil `StateQueue` interfaces as
Java null across run, retry, publication, error and worker-loss paths.
Coordinator worker registration and average-block reporting apply the same
typed-null access checks to their queue, worker and selector interfaces.
Coordinator checkpoint, recovery, model-check reporting, initialization and
state-count paths require typed-null queues at each source dereference phase.
DistributedWorker's direct URI, evaluation, constraint, liveness, cache and
computing-state methods reject nil receivers; the RPC computation adapter still
returns its typed null failure through the error channel.
The public worker `Holder` carrier is exposed as `DistributedStateHolder`, with
source successor/predecessor accessors and signed Java-long fingerprint
comparison. Worker successor partitioning constructs and consumes that carrier.
DistributedWorkerGroup and TLCWorkerRunnable launch/inspection methods now
reject nil receivers before goroutine startup or error conversion. The worker
bootstrap, direct group construction and runnable registration treat typed-nil
coordinator interfaces as Java null at their first source dereference. Runnable
registration still publishes its worker before that failure. The worker runtime
continues to model Java's static shutdown and await operations.
`NewDistributedWorkerRunnable` exposes the public nested-class constructor with
Java `int` narrowing and a direct one-worker runtime; group and process startup
use the same initializer with their existing shared runtime.
The worker-side keepalive task rejects nil receivers before timer panic handling
or checked-error conversion across cancellation, execution, activity scanning,
worker exit, logging and worker lookup. Its direct coordinator status adapter
and its discovery-backed status wrapper also recognize a typed-nil endpoint as
Java null.
DistributedExecutor rejects nil receivers across submission and shutdown state,
and rejects a nil task synchronously instead of starting a panicking goroutine.
UniqueString now rejects nil receivers across its full token, location, string,
comparison, equality, fingerprint and substring surface. UniqueString peer
comparison/equality also dereference null peers like Java, while String-null
equality remains false.
Stored variable/definition locations narrow to Java `int` at `SetLoc`.
InternTable retains its supplied remote-source interface, but a typed-nil source
takes Java's local token-allocation branch when a missing string is created.
Its public put, source mutation, token lookup, string lookup and map conversion
methods reject nil receivers before monitor or live-table access; checkpoint
methods already enforce the same boundary.
Token lookup narrows its identifier to signed Java `int` after that receiver
check.
BigInt now rejects nil or invalid wrapper receivers across value, fingerprint,
comparison, serialization and rendering. `equals(null)` remains false, while
`compareTo(null)` raises the typed null exception.
FP64 exposes the source `getIrredPoly` and `InitRnd` names as
`FP64GetIrredPoly` and `FP64InitRnd`. Checker and distributed-server production
paths use the source-named getter; the prior spellings remain delegates.
Distributed server Tool/error-code adapters, worker/group address adapters and
FP-set registration accept at most one optional value each.
StateWriter transition adapters accept at most one predicate, matching the
Java action-and-predicate write form.
The distributed filename resolver accepts only Java's no-path or one-path-array
constructors, and its public methods reject nil receivers with typed exceptions.
Typed-nil server endpoints behave as Java null only when a cache miss dereferences
the server; cached files remain available without that access.
Coordinator file reads repeat the source `File.length()` observation for array
allocation and narrow that later value to signed Java `int`, including overflow.
Their absolute diagnostic paths are computed only inside the source failure
branches; successful reads do not perform that work.
DiskGraph and TableauDiskGraph construction accept at most one out-degree
statistics object and reject extras before opening graph files.
LivenessStateWriter enforces the same zero- or one-predicate transition forms
as StateWriter before nil and no-op shortcuts.
DiskGraph and TableauDiskGraph DOT writing accept at most one label map before
receiver, cache and file work.
Debugger halt adapters accept at most one level, and the level-taking form now
checks step and breakpoint matches before halting like Java.
Tool and TLCApp fingerprint-state reconstruction accept at most one predecessor,
and debugger expression parsing accepts at most one semantic root.
Tool's four fingerprint-state reconstruction entry points also require their
source receiver before custom dispatch or missing-state handling, while overload
count validation still precedes receiver access.
Unchecked value string conversion accepts only Java's no-delimiter or
one-delimiter overload forms across all concrete value adapters.
Semantic path lookup and extended-module collection accept at most one exact or
recursive flag, matching their Java overload pairs.
Fingerprint invariant checks accept only Java's no-count or one-expected-count
forms across concrete sets, local/network endpoints, RPC dispatch and the
distributed manager; excess counts fail before receiver, storage or transport
work.
Distributed manager block lookup and insertion accept at most one executor,
matching their sequential and executor-taking Java overloads.
SpecTraceExpressionWriter property, Init/Next and trace-function adapters now
accept only their exact Java overload shapes, including the distinct
expression-data and identifier forms.
Its primer accepts at most one extra-module collection, matching the remaining
two source overloads.
Invariant-violation state printing accepts only Java's state-info-only,
previous-state-plus-index and previous-state-plus-index-plus-final forms, with
typed native adapter arguments.
The shared ordinary/nullable message formatter accepts at most one message
class, matching the default and explicit-class Java paths.
Tool constraint evaluation accepts at most one successor, preserving the
distinct state-constraint and action-constraint evaluator forms.
Tool evaluation now dispatches only the six Java evaluator signatures, with
positional context, state, control and cost-model validation instead of a
permissive unordered native argument scan. Signature validation occurs before
call-stack or receiver-dependent work.
Generic Java exception constructors using the shared optional-message path now
accept only their no-message and one-message forms.

The four ignored Alias TTrace methods now have complete Go translations in
`tlc_alias_ignored_ttrace_java_test.go`. Each remains an immediate
`upstream @Ignore: TESpec Bug` skip and receives no passing-method credit. Their
dormant bodies retain the generated trace-spec recheck and complete source
assertions for safety, simulation, lasso and stuttering behavior.

The ignored EvalExceptionLiveness TTrace method and Github1198 e/g methods are
also complete skipped translations. EvalExceptionLiveness reuses the enabled
test's exact 15-state result checker for its generated artifact. Github1198 e/g
retain their full unsupported warning-composition reasons and byte-identical
configs. None receive passing-method credit.

`Github297Test.testSpec` is now a complete skipped translation with its original
debugger override, success/count/depth assertions and byte-identical monolithic
model. It retains the upstream issue-297 ignore and receives no passing credit.

Both ignored IncompleteNext TTrace methods now have complete generated-artifact
rechecks. They share the enabled originals' exact exit, counts, partial-state
traces, actions, incomplete-variable diagnostics and uncovered assertions while
retaining the upstream trace-exploration issue URL as immediate skips.

The ignored four-worker `TLCGetNonDeterminismTest` is also translated in full,
including its inherited SUCCESS expectation, liveness diagnostics, four-state
trace and stuttering ordinal. It retains `No known fix/By design` and uses
byte-identical model/config vectors without receiving passing credit.

Ignored symmetry-liveness inventory now includes complete April20a, April20b,
April21, April22, April25, April29, April29d, May09, May09d, NQ and NQa
translations. They preserve inherited SUCCESS expectations, violation or
no-violation outcomes, generated property identifiers, exact traces and loop
actions. April29 and April29d also retain their disk-graph assertions; NQa
retains its final source failure demanding a trace-completeness check. The
complete OneBitMutex translation adds exact statistics, disk-file sizes,
13-state trace and loop action. ChooseTableauSymmetry adds its exact `13/6/0`,
six-state trace and loop action. The four short/long SymmetryModelChecker
variants preserve their distinct statistics, disk sizes, four- or ten-state
traces and loop-action overloads. SymmetryTableauModelChecker preserves the exact
liveness-reconstruction exception, while TableauSymmetry preserves its graph
sizes, six-state trace and loop action. All fifty-two symmetry source vectors
remain byte-for-byte identical, and the upstream incorrect-symmetry reason
remains.

The complete skipped Github607 translation preserves its distinct eager-tableau
issue reason, disabled coverage, embedded `.tla` config, inherited SUCCESS,
FINISHED/no-GENERAL, depth 1 and `2/1/0`. Its source vector is byte-identical.

Ignored legacy-suite Test61 now has its complete inherited `SuiteTestCase` body:
default `2/1/0`, one initial state, FINISHED/no-GENERAL and zero uncovered. It
retains the class-level JSpec/ISpec bug reason and two byte-identical vectors.

Both ignored OffHeap indexer-equivalence methods now retain the full 21,762-row
source matrix, exact assumptions, indexer ordering, random/uniform/boundary
comparisons and monotonicity checks behind `Checked in older commits` skips.
They receive no passing credit. Every method in the current upstream-ignored
inventory now has a complete Go body or an already matching explicit Go skip.
The nested public off-heap iterator is exposed as `OffHeapIterator`, with its
three constructor shapes, wrap policy, position, next, marking and availability
methods. The original iterator tests use this public surface directly.
`OffHeapDiskFPSetIsSupported`, `Evict` and `IsHigher` expose the remaining main
class boundaries. Factory architecture checks use the source support predicate,
and disk-offset search enters through the public comparison method.
The distinct public `InfinitePrecisionMult1024Indexer` constructor and support
predicate are also available. Index calls preserve each Java implementation's
signed-long behavior: the two multiplication variants differ for negative
fingerprints and probes, while ordinary off-heap storage continues to pass its
masked nonnegative fingerprints.

The latest liveness audit matches two `GraphNode` formatting quirks. Java
unconditionally removes the presumed trailing comma and space, so an empty
`<0,0>` node ends in `--`. Its filter-taking DOT overload also discards the
filter, and `LiveWorker` therefore emits every outgoing edge. Native now
preserves both behaviors. Graph serialization also now raises `AssertionError`
before writing an overallocated node, instead of persisting unused transition
slots. Frozen-jar probes, the focused graph/worker selection and all-package
compilation pass.

Disk-graph SCC mutation now also preserves Java's assertion boundaries: links
must be in `[MAX_PTR, MAX_LINK)`, and tableau lookup, recording and link methods
require nonnegative tableau indices. An assertion-enabled frozen-jar probe
confirms the typed failures; focused disk/tableau/worker checks pass.

Disk-graph recovery now rebuilds the existing pointer tables as Java does,
rather than replacing them first. This retains in-memory max-link entries and
tableau done-only fingerprints that have no pointer-file record. Java and native
probes agree on both retained cases; original disk/tableau recovery tests pass.

Liveness resource closing now preserves Java's fail-fast order: a node-file
close failure skips the pointer file, a disk-graph close failure skips the
liveness writer, and a checker close failure skips later checkers. The explicit
`VETO_CLEANUP` property still retains disk-graph handles while closing the
writer, matching the installed Java behavior.

Disk-graph on-disk size accounting now queries the pointer file before the node
file, preserving Java's exception order when either graph handle is invalid.

Graph transitions now store both fingerprint halves as signed Java `int`
values. High-bit fingerprints therefore survive disk round trips without
defeating duplicate-transition lookup. Ordinary disk-backed liveness checking
also restores Java's `TLCRuntimeException(EC.GENERAL)` guard when a nonempty
node gains no transition while a main checker is active.

Modern `LiveCheck.GetChecker` now raises typed
`ArrayIndexOutOfBoundsException` for invalid indices like Java's array access;
the folded no-op implementation continues to return null for every index.

The explicit `LiveCheck` writer constructor now preserves Java's null ordering:
a null solution array fails immediately, a null writer fails on the first
solution, and an empty solution array never dereferences its null writer.
Internal no-writer construction supplies a concrete no-op writer.

Liveness degree-statistics methods now retain a supplied null accumulator.
Empty graphs return it unchanged; the first actual sample raises typed
`NullPointerException`, matching Java instead of manufacturing a histogram.

Fixed-size bucket-statistics construction and empty-bucket sampling now expose
Java's typed array failures. Both fixed ordinary and concurrent constructor
capacities narrow to Java `int` before allocation. `BucketStatistics.Add`
accepts every statistics implementation through the source interface shape and
rejects Java-null inputs.
Ordinary, fixed and concurrent sample amounts narrow to Java `int` before
negative validation and bucket lookup or overflow selection.
Ordinary, fixed, concurrent, fixed-concurrent and dummy statistics now reject
nil receivers across mutation, observation, sample, derived-metric and rendering
paths. A valid dummy returns a concrete empty sample collection as Java does.
`CounterStatistic` also dereferences its Boolean supplier and public receiver
like Java instead of converting null into a disabled or zero-valued counter.
Its enabled implementation renders the numeric adder, while its disabled
implementation preserves Java's inherited nested-class identity rendering.
`CountDistinct` now does the same for public and internal instance methods.
Its no-op variant still accepts null values and states, while active variants
raise typed `NullPointerException`. HyperLogLog construction and indexing use
Java's masked `int` and `long` shift distances, including the source behavior
for zero, negative and wrapped bit counts and typed array-size/index failures.
The naive implementation now shares a per-instance monitor across hash insertion
and counting, matching the synchronized `MemFPSet.put` and `size` calls to which
Java delegates. Ordinary HyperLogLog remains unsynchronized; its explicitly
synced subclass computes the register index and rank outside its monitor, then
locks only the register update; exceptional index access releases that monitor
like Java's synchronized block. Counting remains synchronized across the full
estimate.
`ConcurrentBucketStatistics` likewise follows its source concurrent carriers
instead of a whole-object mutex: observations and existing buckets update
atomically, samples take a weakly consistent sorted snapshot, and simultaneous
first insertion retains Java's ignored `putIfAbsent` loser.
Bucket-derived metrics preserve Java's independent live reads and their order
instead of freezing one observation count and sample snapshot. This includes
the mean's two observation reads, standard deviation's nested mean reads,
percentile's NaN rejection before any statistics access, and median's source
runtime failure when a live observation count exhausts its sample snapshot.
Successful model-checker outdegree reporting aggregates each worker through
that same source `BucketStatistics.add` boundary, retaining its separate
observation and sample reads instead of reconstructing observations from the
sample counts.

Tableau graph node lookup now retains Java `Vect` backing-array behavior. A
new graph starts with capacity 10 and doubles on growth; lookup can return null
from an unused slot within that capacity and raises typed
`ArrayIndexOutOfBoundsException` outside it. `GraphNode.GetTNode` now also
raises typed `NullPointerException` for a null graph.

Legacy `BTGraphNode` index, flag, tableau lookup and diagnostic methods now
raise typed `NullPointerException` when invoked through a null node. Tableau
lookup also fails on a null graph instead of returning null.

Base `BEGraphNode` methods now preserve direct null-receiver failures and typed
successor-array bounds. Empty parent lookup fails at index zero, and a null
transition query returns false only while the successor array is empty before
failing at the first equality call. Successor indices narrow to Java `int`.

`BEGraph` initial nodes now retain their Java `Vect` capacity and growth.
Lookup can return null from unused backing slots and raises typed array bounds
outside capacity; lookup indices narrow to Java `int`, and graph methods
dereference stored null nodes in source order.

`BEGraphGetPath` now preserves Java's null ordering: a null start raises typed
`NullPointerException`, a null destination exhausts the search and returns the
coded path-construction failure, and a stored null successor fails when its
visited flag is read.

`TBGraphNode` successor lookup now retains the inherited ten-slot Java `Vect`
and exact doubling. Unused slots return null, indices outside capacity raise
typed array bounds, and construction with a null particle raises typed
`NullPointerException`.

The two `TBGraph` constructors are now distinct: `NewEmptyTBGraph` represents
Java's no-argument empty graph, while `NewTBGraph` processes its formula. A
null or otherwise unknown closure term raises the coded
`TLC_LIVE_ENCOUNTERED_ACTIONS` runtime failure.

`TBGraphNode` public access, consistency, accepting and DOT methods now raise
typed `NullPointerException` on a null receiver. Construction also rejects a
null expression inside its particle at Java's level lookup. The documented nil
`String()` convenience remains an empty native string.

`TBGraph` size, initial-count and DOT methods now likewise type null receiver
failures. Its documented nil `String()` convenience and the internal optional
export helper remain unchanged.

`TBPar` and `TBParVec` indexed access now retains the backing capacity of
Java's inherited `Vect`. Unused slots within a requested capacity return null,
growth doubles exactly and indices at capacity raise typed array bounds.
Removing the last particle expression also clears its former backing slot.

`Assignment.SetFormula` now raises the typed Java
`UnsupportedOperationException("Not implemented yet")` instead of an untyped
Go string panic. Its existing translated check now verifies type and message.
`Assignment.SetRight` and `GetSetOfModelValues` also share the source object
monitor, preserving atomic right-side replacement/cache invalidation and lazy
typed-set publication. Other assignment getters and mutators remain outside it.

Invalid assignment symmetry and model-value mutations now raise typed
`IllegalArgumentException` with the source messages. Creating a negative-size
empty-string array likewise raises typed `NegativeArraySizeException`.

Formula name parsing now includes Java's vertical-tab regex whitespace and
uses `String.trim` boundaries for matched sides. It removes code units through
U+0020 while preserving nonbreaking spaces.

`TypedSet` now distinguishes Java's unset type from an explicitly empty type.
That preserves `hasType`, underscore rendering, unsetting and equality while
retaining the existing exported string representation.

`TypedSet.Value` now narrows its index to Java `int` and raises typed
`ArrayIndexOutOfBoundsException` for negative indices. Upper indices retain the
existing empty-string representation of Java null.

Formula-list deserialization now raises typed `NullPointerException` for a null
list and typed `StringIndexOutOfBoundsException` for an empty serialized entry,
while retaining Java's enabled-entry filtering.

`MCVariable` now exposes the source `GetName`, `GetTLCValue` and
`GetValueAsString` methods, plus `GetSingleLineDisplayName` and
`GetValueAsStringReIndentedAs`. All public variable access, trace and formatting
methods raise typed `NullPointerException` on a null receiver. Trace-expression
storage, its setter and getter preserve Java's nullable string, including
clearing an existing annotation. Older shortened spellings remain compatibility
delegates.
Nullable constructor adapters distinguish Java's string and `IValue` overloads:
both reject a null name with coded `EC.GENERAL`, the string overload also rejects
a null value, and the value overload accepts null as empty rendered text.
Nullable indentation renders a null prefix as Java's `null` text on every line.

`TypedSet` exposes the source-named `GetValuesAsList`, `GetValueCount`,
`GetValue`, and `ToStringWithoutBraces` boundaries. `TypedSetGetTypeOfID`
preserves the static helper's nullable input and result; the prior native names
remain compatibility delegates. Spec generation uses the source-named methods.
Nullable adapters preserve `getType`, `setType`, `contains`, and `getValue`
nulls while the established string helpers remain available to non-null callers.
`TypedSetGetTypeOfID` indexes Java UTF-16 code units, including non-ASCII BMP
prefixes, rather than treating UTF-8 bytes as Java characters.
The value array also preserves nullable entries and caller-owned array mutation
through nullable accessors. Rendering, typed-list construction and indexed lookup
append a null entry as Java's `null` text; equality still distinguishes null from
an empty string, and the source numeric scan dereferences null entries.

`ModelConfig` initializes its Java collection fields as concrete empty slices and
returns the live raw-constant and section storage rather than detached copies.
`GetConstantsAsList` remains a newly derived collection and uses Java trim and
trailing-split rules for both raw newlines and assignment separators.
Specification detection uses Java `String.trim` characters,
and nullable reverse-override lookup preserves an absent mapping and rejects a
null key before the processor consumes it. All 24 public configuration query
paths reject a null receiver with typed `NullPointerException`.
Those query paths now also share Java's `ModelConfig` object monitor; nested
source getter calls use lock-held field lookups to retain reentrant behavior.
The public error-returning `Parse` method now fills an existing ModelConfig from
its file or monolith source, and `ParseModelConfigFile` delegates to that same
path. Like Java's instance method, `Parse` leaves the global model-value array
unset; the Spec-equivalent application loader finalizes it after successful
configuration parsing and before SANY. The source/file convenience helpers
still return ready-to-use finalized configurations. Its optional
`FilenameToStream` constructor argument is retained and used
to resolve configuration files, including classpath, embedded and distributed
resources. The instance `Parse` method passes its stored logical name to that
resolver unchanged; caller-level helpers apply the Java `.cfg` default before
constructing the instance. A missing or typed-nil argument creates the standard
resolver in the constructor, capturing its user and library search paths before
any later parse; the zero-value parse fallback follows the same rule. Repeated
singleton sections replace the stored value before raising the
duplicate-keyword error, matching Java's `Hashtable.put` order. A repeated valid
`CHECK_DEADLOCK` boolean follows the same replacement order, while an invalid
value fails before mutation.
Raw constant text is accumulated per parse and replaces the receiver's raw list
only after full success. A failed reparse retains the prior raw list while other
semantic collections keep Java's already-applied partial mutations; a successful
constant-free parse publishes a concrete empty list.
Configuration string values retain Java's raw token contents after removing the
quotes, so valid escape sequences remain backslash text rather than being decoded
by Go. The lexer accepts exactly Java's six string escapes and converts invalid,
multiline and unterminated strings to the same synthetic EOF boundary; value
parsing preserves the resulting Java null-token dereference.
The lexer also recognizes Java's decimal, octal, binary and hexadecimal number
token shapes. `ModelConfig` still passes every such token to decimal
`Integer.parseInt`, so decimal overflow and all base-prefixed forms raise the
source-compatible typed `NumberFormatException` with its exact input message.
`ModelConfigAllKeywords` publishes Java's ordered, mutable 22-entry
`ALL_KEYWORDS` array. Parser recognition remains independent of mutations to
that public array, matching Java's separately populated configuration table.
Identifier scanning follows SANY's special forms: `ℕ`, `ℤ` and `ℝ` remain
single identifiers, an underscore prefix requires a later ASCII letter, and
`$` remains its standalone operator token rather than starting an identifier.
The generated literal inventory also retains the `@` alternative embedded in
SANY's mixed `IDENTIFIER` production.
Characters outside the complete SANY token vocabulary now become Java's
synthetic EOF token instead of native-only one-byte symbols. Top-level,
singleton and value contexts consequently retain their distinct source failure
or successful-termination behavior.
List and constant sections retain that recovered EOF for the outer parse loop,
rather than continuing the token manager past the lexical failure. Qualified
constant left sides preserve Java's physical-EOF and null recovery-token
behavior after `!`. Constant punctuation checks dereference the synthetic
token's null image at the same source boundaries in arguments, module targets
and sets. An empty argument list first creates the partial `)` model value and
then raises Java's general configuration error.
Unterminated block comments and line comments that reach physical EOF use that
same marked token. Both LF and CR terminate line comments normally, including
CRLF through the following whitespace pass.
Configuration token positions count bare CR and LF as new lines and count CRLF
once, matching `SimpleCharStream` diagnostics inside and outside comments.
The `SF_` and `WF_` fairness literals win before general identifier scanning, so
following letters form a second token while valid `SF1_x` and `S_F` identifiers
remain whole, matching SANY's CASE2/CASE3 productions.
The regex-only CASE1b/c, CASE2b/c and CASE6b/c junction bullets also use their
complete prefixed token images; their source exclusions for `SF`, lone `S`/`W`
and underscores remain effective.
SPEC-state whitespace is exactly space, tab, CR and LF. Form feed now follows
Java's lexical-error EOF path instead of being skipped as native whitespace.
Resolver results that are null or name a missing file return the adapted
`CFG_ERROR_READING_FILE` with Java's logical filename and fixed `File not found.`
reason instead of a Go nil panic or host-specific filesystem text.
Resolved-file reading follows Java's stat/open/read phases. Directories and
open failures print FileUtil's internal FileInputStream line and return the same
missing-file error; failures after a successful open retain normalized I/O text.
Configuration bytes are decoded with Java's startup default charset and decoder
replacement policy. Monolith extraction additionally preserves the source's
UTF-8 re-encoding followed by a second default-charset decode.
The production TLC application loader now constructs this resolver-backed
`ModelConfig` directly instead of duplicating file reads and source parsing.
Configuration I/O failures still print their Java diagnostic and cross the
`System.exit` adaptation, while syntax failures remain `ConfigFileException`s.

`TraceExpressionInformation` exposes the complete Java holder surface: four
getters, four setters, its constructor and `CreateTraceExpressionInformationHolders`.
The factory now rejects null lists and null elements at the source dereference
boundaries. Its three string fields and their public constructor, setters and
getters preserve Java null references. Trace-expression writers use the public
accessors and reproduce `StringBuffer.append(String)` by rendering null as
`null`. Holder creation skips a `Formula` whose text reference is null or empty,
leaving the corresponding array slot null.

`ParseMCState` now rejects missing separators, malformed or overflowing state
numbers and ordinary headers without a variable-line newline at Java's typed
exception boundaries. Marker-only stuttering states remain valid.

The `MCState` marker-copy constructor rejects a null source, shares the source
variables and descriptive fields, and leaves its record cache null as Java does.
All public state getters and formatting methods likewise raise typed
`NullPointerException` on a null receiver.

`MCState` now retains its nullable location as `*SourceLocation`, matching the
Java carrier instead of flattening it to text. Parsing captures canonical TLC
line/column/module coordinates. `GetLocation` returns that structured value,
and record rendering distinguishes a null location (`"null"`) from the unknown
location value. `GetConjunctiveDescription` exposes both source overloads; the
older shortened spelling remains a compatibility delegate.
State names and labels also retain null separately from explicit empty strings;
parsed marker states keep their null name, while TLCStateInfo construction uses
the source's explicit empty fields. Nullable getters expose the direct values,
record headers render null names as `"null"`, and a null conjunctive indent is
omitted. All three variable renderers dereference a null variable array as Java
does rather than treating it as an allocated empty array.

`MCError` now owns a concrete empty state slice, types all null-receiver and
stored-null-state failures, and preserves the source order for null
trace-expression maps: empty states succeed and the first variable lookup
fails. Its message carrier and getter preserve Java null separately from an
explicit empty string, including the no-argument and cause constructors.
The nullable trace-expression map adapter ignores keys mapped to null and retains
nonnull expression references. Trace updates and record-sequence rendering
dereference a state's null variable array at the Java boundary.

`MCError.IsLassoWithDuplicates` now requires each pre-marker state's record as
Java does. Missing records raise typed `NullPointerException`; the removed
native presentation-string fallback could silently classify distinct states.

`MCState` variable parsing now follows Java `String.split` trailing-field rules,
unconditionally skips the source separator character after `/\`, and raises
typed string or array bounds for malformed assignments instead of dropping
them.

An `MCState` label beginning directly with `line ` now preserves Java's
`substring(0, -1)` failure instead of accepting a location without an action
name. Canonical named locations and ordinary names remain unchanged.

`MCState.AsRecord(true)` now emits the source `_TEAction` field and inserts
state names and locations raw between quotes rather than Go-escaping their
contents. Direct state construction also narrows ordinals to Java `int`.

All six public `Formula` access, mutation, rendering and name-splitting methods
now raise typed `NullPointerException` on a null receiver instead of returning
empty defaults or ignoring mutation. Its backing text and nullable constructor,
getter and setter adapters preserve Java null separately from an empty string;
the established string surface renders a null reference as Java's `null` text.

All 23 public `Assignment` overloads and methods now likewise type null receiver
failures. In particular, null `SetFormula` fails before reaching the real
receiver's deliberate `UnsupportedOperationException`.
The nullable formula-setter adapter is overridden at the same boundary and
cannot mutate an assignment through its embedded `Formula`.
Nullable construction and accessors preserve null labels and right sides, while
formula and pretty-print concatenation renders them as Java's `null` text.
Nullable tab and delimiter adapters retain the corresponding overload behavior.
Assignment overrides restore Java's virtual Formula name-splitting fallback to
the formatted assignment rather than the embedded raw right side.

The TLC value pretty-printer exposes Java's public parse-tree pipeline as
`PPrintParse`, `Node` and `PPrintFormat`. Node type/range, child, sibling,
source, formatted-value, mutation, append and rendering methods are available.
Parsing, node ranges, substring values and formatting widths use Java UTF-16
code-unit indices and `Character` classification rather than Go byte offsets.
The public parser entry index narrows to Java `int` before source range checks
and subvalue selection.
Public formatter widths and every nested layout addition/subtraction preserve
signed Java `int` conversion and overflow.
Node constructor/setter metadata, source offsets and range length/end arithmetic
also preserve signed Java `int` narrowing and overflow.
Nullable node strings and formatted values remain distinct from empty strings.
`UniqueString` counted joins, substring starts and stored variable counts also
narrow at their Java `int` boundaries; token lookup narrows in `InternTable.Get`.

The nullable parameter-array constructor, setter and getter retain the caller's
array, trim elements in place, preserve partial mutation before a null-element
failure, and render a null inserted after construction as Java's `null` text.
`SetParams` mutates the supplied slice in place and applies Java `String.trim`
semantics, preserving nonbreaking and other Unicode spaces above U+0020.

All 16 public `TypedSet` instance methods now type null receiver failures,
including equality and string rendering. A real set still compares false to a
null argument as Java does.

In-memory `IntStack` indexed peek now addresses the 1,024-cell backing array,
including unused zero cells. Empty pop decrements size to -1 before raising
typed array bounds, and public stack methods type null receiver failures.
Its source-spelled `PeakInt` and `PeakLong` overload families are primary in
LiveWorker and the translated original test; corrected `Peek` names delegate.
Push, pop and peak methods share Java's reentrant `MemIntStack` monitor through
lock-held helpers; inherited size and reset remain unsynchronized.

`SynchronousDiskIntStack` now preserves Java's constructor, null-receiver and
disk-failure boundaries. Negative capacity raises typed negative-array size;
explicit capacities narrow to Java `int` before the source maximum-capacity
clamp and allocation;
empty pop and failed page I/O raise coded state-writing failures without
mutating size, and page flushes do not create missing parent directories.

Memory and disk object stacks now preserve their shared Java `ObjectStack`
boundaries. Empty synchronized pop raises typed backing-array bounds without
changing size, negative bulk counts and null bulk inputs use their Java typed
failures, and every public stack/checkpoint method rejects a null receiver.
Plain push, pop and size remain outside the stack monitor through a shared
signed-32-bit atomic snapshot; synchronized stack methods retain their monitor.

`MemObjectQueue` now rejects null receivers across all five source methods.
Its empty dequeue continues to return null; the Go-only size convenience keeps
its documented behavior and is not treated as a Java API method.

`ObjectPoolStack` now preserves the frozen Java implementation's broken monitor
notification boundary: first write and positive recovery mutate pool metadata,
then raise typed `IllegalMonitorStateException` before disk I/O. The alternative
`DiskObjectStack` wraps pool failures in the source coded stack I/O errors.

Object-stack bulk push now defers logical-length mutation until every enqueue
succeeds. This preserves `MemObjectStack`'s repeated-slot overwrite quirk and
`DiskObjectStack`'s unchanged size after partial bulk failure. Disk buffer stores
also post-increment the index before typed array bounds failure, as Java does.

`MemObjectStack` checkpoint writing now preserves the source loop's double
index increment and writes only even slots. Recovery keeps the declared length
and decoded prefix before typed EOF or fixed-array bounds failure; checkpoint
opens preserve Java file-not-found boundaries instead of creating parents.

`DiskObjectStack` checkpoint paths now use the same source file-open failures.
Checkpoint writing types overflowed buffer indices, and recovery decodes each
object before fixed-buffer bounds checks while retaining already-read metadata
and values after failure.

`CallStack` now retains Java's 64-slot backing array and explicit signed index.
Empty pop underflows to -1, failed push post-increments back to zero, freeze
overloads preserve their null ordering, and all six public methods type null
receiver failures.

`FingerprintException` now types null receivers across root-cause, trace,
trace-list and message access. Trace lists retain null sources, consecutive
trace suppression compares semantic-node UIDs, and empty trace lists are
concrete rather than nil. `FingerprintExceptionGetNewHead` restores the source
head-construction path: an existing fingerprint exception gains a new value
head, while another cause starts a new chain. Constructors treat typed-nil Go
interfaces as Java null, and production value-failure wrapping uses this public
entry point.

`WorkerException` now preserves nullable detail messages through construction,
throwable wrapping and distributed decoding. Its inherited message and cause
accessors raise typed null-receiver failures, while cause and state ownership
remain unchanged.

`StatefulRuntimeException` preserves default, empty and null detail messages,
cause identity and the source known-flag transition. Message, cause and both
known-flag methods raise typed null-receiver failures.

The combined `TLCError` carrier applies the same receiver contract to its
Java-facing message, cause and known-flag methods for both `EvalException` and
`TLCRuntimeException` roles. Its Go-only `Error` and `Unwrap` methods remain
nil-safe.

`EvalException` retains and returns its source parameter array instead of
copying it. Caller mutation is visible through the getter while the constructed
message remains stable; its three declared accessors type null receivers.

`ConfigFileException` retains its construction-time formatted message and
cause identity. Inherited message and cause access now raise typed
null-receiver failures.

The stateful next-state and plain initialization invariant exceptions remain
distinct. Their inherited message, cause and known-state surfaces type null
receivers, and typed-null initialization errors no longer pass the source
exception predicate.

`Worker.WrappingRuntimeException` keeps its evaluation exception and successor
state separate from Java's null throwable cause. Message, cause and both unwrap
accessors now raise typed null-receiver failures.

`SimulationWorkerError` retains mutable parameter-array, state-trace and payload
exception ownership while keeping the throwable cause null. Nullable parameter
arrays are no longer copied, and inherited stateful accessors type null outer
receivers.

`LiveException` and `LiveCounterExampleException` preserve all nullable-message
constructor forms, null throwable causes and the source's shadowed base/subclass
counterexample fields. The native-only cause channel is removed, and throwable
accessors type null receivers.

`FPSetManagerException` accepts Java's nullable constructor message and types
message/cause access on null receivers. It retains the documented native
`IOException` substitute, absent cause, source class name and registration-
rejection classification after removal of dormant RMI scaffolding.

`FrontEndException` preserves both SANY constructors, including nullable direct
messages and cause-only message derivation from the cause's Java `toString`.
Message and cause access type null outer receivers.

The shared `RuntimeException` and `IOException` carriers expose Java's default,
nullable-message, message-plus-cause and cause-only construction semantics.
Cause identity, derived messages and typed null receivers are preserved.

`RejectedExecutionException` now exposes all four Java constructors with
nullable messages, cause-derived text, retained cause identity and typed null
receiver access. Distributed worker rejection classification is unchanged.

`ExecutionException` preserves both public Java constructors, and
`EOFException` preserves default versus nullable message construction. Both
carriers retain their cause rules and type null receiver access.

`ExceptionInInitializerError` preserves default, throwable and nullable-string
construction. Throwable construction keeps a null message while both legacy
`GetException` and `GetCause` return the saved exception; all accessors type
null receivers.

`NoClassDefFoundError` preserves default/nullable messages, and
`ClassNotFoundException` preserves default, nullable-message and
message-plus-cause construction. Its legacy `GetException` aliases the retained
cause, with typed null receivers across both carriers.

`IllegalAccessException`, `InstantiationException` and `NoSuchMethodException`
preserve default and nullable-string construction. `InvocationTargetException`
preserves default, target-only and target-plus-nullable-message construction;
both `GetCause` and legacy `GetTargetException` return the target. All message,
cause and legacy accessors type null receivers.

`SecurityException` and `ConcurrentModificationException` preserve all four
public constructor forms: default, nullable message, cause-only and nullable
message plus cause. Cause-only construction derives Java throwable text, both
carriers retain cause identity, and their accessors type null receivers.

`NoSuchElementException` and `UnsupportedOperationException` likewise preserve
default, nullable-message, cause-only and nullable-message-plus-cause forms.
Cause-only messages use Java throwable text, causes retain identity, and
message/cause access types null receivers.

`IllegalArgumentException` and `IllegalStateException` preserve the same four
public forms, including Java cause-derived messages and retained cause identity.
Their message and cause accessors type null receivers.

`IllegalMonitorStateException` and `IllegalThreadStateException` preserve
default versus nullable-string construction, report their exact Java class
names, retain null causes, and type null receiver access.

`StackOverflowError`, `OutOfMemoryError` and `NullPointerException` preserve
default versus nullable-string construction and null causes, with typed message
and cause access on null receivers. Existing TLC system-failure classification
remains intact.

`NegativeArraySizeException`, `ClassCastException` and `ArithmeticException`
preserve default versus nullable-string construction and null causes, with
typed message and cause access on null receivers.

The three index exception carriers preserve their public default, nullable
message and numeric constructors separately from existing JVM-generated
index-and-length messages. `IndexOutOfBoundsException` includes both `int` and
`long` numeric forms. Public `int` indices and generated index/length operands
narrow to Java width, while the `long` overload remains wide; all accessors type
null receivers.

Generic `JavaError` preserves the four public `java.lang.Error` constructor
forms, including cause-derived text and retained cause identity. Its accessors
type null receivers, and Java Error-family classification remains intact.

`AssertionError` preserves default, object, boolean, UTF-16 character, integer,
long, float, double and message-plus-cause construction. Object construction
uses Java-style value text and retains throwable objects as causes; floating
messages preserve Java decimal/scientific thresholds. Accessors type null
receivers.

`ZipException`, `UnknownHostException`, `FileNotFoundException` and
`UnsupportedEncodingException` preserve default versus nullable-string
construction, null causes, checked-I/O classification and typed null receiver
access.

`NumberFormatException` preserves public default and nullable-message
construction separately from Java's decimal/radix input-string factories. It
narrows factory radix metadata to Java `int`, retains null causes and types null
outer/embedded receivers.

The formatter's package-private `IllegalFormatArgumentIndexException` narrows
its stored index to Java `int`, including the special `Integer.MIN_VALUE`
message used when a parsed argument index is not representable.

`InvalidPathException` preserves indexed and non-indexed construction, null and
negative-index validation, input/reason/index metadata, composed messages, null
causes and typed null receiver access. Explicit indices narrow to Java `int`
before validation.

`URISyntaxException` preserves indexed and non-indexed construction, null and
negative-index validation, input/reason/index metadata, composed messages, null
causes and typed null receiver access. Explicit indices narrow to Java `int`
before validation.

`WrongInvocationException` preserves its sole nullable-string constructor, null
cause and typed null receiver access directly from the checked-out upstream
source.

The five distributed lookup carriers preserve their JDK default versus
nullable-string construction, null causes, checked-I/O classification where
applicable, and typed null receiver access.

`FileSystemException` and its four translated subclasses now expose their
complete file, other-file, reason, message and cause accessors with Java's
null-receiver failures. Message composition retains nullable metadata and the
JDK's `file -> other: reason` ordering. Installed bytecode and the native probe
agree, focused file/checkpoint tests pass, and all packages compile.

`BufferedRandomAccessFile` now distinguishes a missing receiver from a valid
closed handle across shared open-file checks, close, public `seeek` and all mark
methods. Nil invocation raises the typed Java `NullPointerException`; closed
files retain their source `IOException`, and repeated close remains successful.
Compact short-natural and natural writes narrow their inputs to signed Java
`int` before selecting the one-, two- or four-byte encoding.
The unbuffered `RandomAccessFile` applies the same typed receiver boundary to
close, seek, scalar and array read/write operations. Null byte arrays fail after
the source operation counter increments.

The five core charset exceptions now retain their concrete Java class through
native error formatting and guard message, cause and metadata access. Coding
lengths narrow to and accept every Java `int`; unsupported and illegal charset
names retain nullable names while using Java's `"null"` constructor message.
JDK and native probes agree, and the original JSON model test passes.

Both core JSON deserializers now preserve `FileReader`'s construction boundary.
Direct null paths return the JVM's exact helpful `NullPointerException`, while
missing files, invalid paths and directories return typed
`FileNotFoundException` with the source messages. The three direct JSON
conversion entry points also retain the exact helpful null-value message. The
ordinary JSON serializers preserve their value-before-path null ordering and
both exact helpful messages. JSON numbers now follow Gson's `getAsInt()` path:
decimal/exponent values truncate toward zero, wider values narrow to Java
`int`, and Gson's number-length and scale guards retain their exact failures.
Deserializing JSON `null` returns Java's typed `IOException` and exact unsupported
value message in both ordinary and newline-delimited paths. Empty and
whitespace-only ordinary documents follow Gson's null result, while empty
newline-delimited input remains an empty tuple.

Malformed-document parity now uses a UTF-16-unit Gson-shaped reader. It parses
the complete lenient first value before TLC conversion, then performs the strict
trailing-content check. Comments, unquoted and single-quoted tokens, alternate
separators, missing array values, case-insensitive literals, BOM and Gson's
non-execute prefix follow the checked Java behavior. Failures retain the
represented `JsonSyntaxException`, `MalformedJsonException` or `EOFException`
cause hierarchy, path, UTF-16 line/column position and troubleshooting URL.
Lenient quoted strings preserve an escaped physical line feed, as Gson does.
An initial BOM is zero-width for first-line diagnostics; a BOM after leading
space remains an ordinary unquoted token. A first JSON null bypasses Gson's
strict trailing-content check before TLC rejects the null as unsupported.
NDJSON also follows Java `BufferedReader.readLine` splitting and `String.trim`
semantics. Arbitrary Gson streaming APIs and resource-exhaustion behavior remain
outside this module boundary.

`_TLCTrace` binary I/O now preserves Java `File`/stream construction. Null paths
and values fail in source order with exact helpful messages; write-open failures
are typed `FileNotFoundException`. Invalid or missing read paths return the
empty record, while an existing directory fails during input-stream open. A
null value leaves Java's exact 20-byte empty gzip member because stream close
runs from `finally`; close failures replace earlier returned I/O failures.

Value streams now expose Java's two-method data carrier interfaces. Sequential
streams return their concrete buffered input/output object; direct byte-queue
streams return themselves. Calls through either view share the same cursor and
buffer, matching `getInputStream()` and `getOutputStream()`. The deprecated
`read(Map)` compatibility overload is also represented: it ignores even a null
or conflicting supplied table and delegates to external interning. `_TLCTrace`
uses that overload as Java does. File/String constructor equivalents default to
the global gzip flag or accept the explicit boolean overload. File input retains
Java's two eager buffered layers; output retains one. `ValueWrite` restores the
public value-receiver serialization boundary over the stream-owned Go dispatch;
state and `_TLCTrace` serialization enter through it. String, tuple, enumerated
set, function-record and record values expose their public `createFrom` and
`createFromExternal` boundaries, and stream kind dispatch enters through them.
Sequential input handle assignment and lookup narrow indices to Java `int`
before accessing the source handle array.

`UniqueString` now exposes both source `concat` overloads, string equality and
Java `hashCode`. Its length, prefix and substring operations use UTF-16 code
units, including prefixes and slices that end within a surrogate pair. Bounds
raise the represented `StringIndexOutOfBoundsException`. The shared Java string
hash and `InternTable` equality also preserve unpaired surrogates and recognize
different native encodings of the same UTF-16 units. Nullable string helpers
retain Java concat's `"null"` text and false equality result.

`BoolValue.GetVal` and `StringValue.GetVal` expose their source accessors. They
return the stored primitive or exact `UniqueString` reference, including a null
stored string reference, and reject null receivers with typed
`NullPointerException`.
Primitive, interval, model, lazy, undefined, user and all four operator value
forms also reject null receivers at their inherited no-op `DeepNormalize`
boundary. Shared string and operator receiver dispatch performs the same typed
check before returning a concrete owner. Their identity and constant
normalization, normalizedness, definedness, deep-copy and permutation paths
likewise reject null receivers before returning or examining arguments. Their
kind and kind-string introspection also dispatch through the receiver first,
including the four concrete operator carriers rather than Go's promoted-field
panic. The remaining primary composite, set-constructor, set-operation,
predicate, function and k-subset values enforce the same kind boundary; their
kind-string path already enters through the nil-aware shared helper. Priority
and callable evaluator subclasses, lazy suppliers, debugger
placeholders and counterexamples enforce the same boundaries explicitly;
identity-returning operations retain the dynamic subclass receiver.
`GetCostModel` now has an explicit receiver boundary across the complete value
hierarchy, including the shared operator implementation and the five subclass
wrappers. Valid calls return the exact stored cost-model reference.
Those five wrappers also explicitly dispatch comparison, equality, membership,
finiteness, size, fingerprinting, EXCEPT and string rendering through their
outer receiver. Empty counterexample EXCEPT batches retain the counterexample
identity. Operator-lambda construction no longer renders a nullable definition;
name rendering occurs at the source `toString` boundary.
The four concrete operator carriers now apply the same outer-receiver dispatch
to comparison, equality, membership, finiteness, size, fingerprinting, EXCEPT
and both string-rendering forms instead of relying on promoted base methods.
All primary value carriers now reject nil receivers at `IsFinite`, including
the constant-true interval, tuple, enumerated-set, record and function-record
implementations.
The same complete primary hierarchy now rejects nil receivers at `Size`; the
three previously correct string and product-set implementations remain intact.
Membership also dispatches through the receiver first across the primary
hierarchy, before either null or non-null element handling.
Equality now has the same receiver-first ordering for both null and ordinary
operands across all primary value carriers.
Comparison likewise rejects the receiver before inspecting either null or
ordinary operands throughout the primary hierarchy.
Fingerprinting now rejects nil receivers before normalization, unsupported-value
diagnostics or hash work across every primary value carrier.
Single-value EXCEPT application also rejects nil receivers before validating a
null descriptor or returning the replacement for an empty path.
Batched EXCEPT application likewise rejects the receiver before validating a
null slice or returning from an empty non-null batch.
Ordinary string rendering now rejects nil receivers across every primary value
carrier, including the formerly constant `UndefValue` image.
Builder-based rendering likewise rejects the receiver before either null or
valid builder handling throughout the primary hierarchy.
The 15 composite, set and function carriers now also enforce receiver-first
dispatch across normalization, deep normalization, normalizedness, definedness,
deep copy and permutation, completing those inherited `Value` boundaries beyond
the earlier primitive, operator and subclass passes.
The shared `InitializeValue` entry point now treats both a nil interface and a
typed-nil value as a Java null receiver, raising the translated
`NullPointerException` before optional concrete initialization or the default
deep-normalize-and-fingerprint path.

`SumIntValues` now preserves Java `IntValue.sum`: cached small results and signed
32-bit overflow remain intact, while null left/right operands raise the exact
helpful `n`/`m` field-access messages.
`IntValue.NBits` rejects a nil outer receiver before entering its source catch
region and otherwise retains the signed-width bit-count algorithm.

Special-value comparison now follows Java `instanceof` for native typed nils.
`UndefValue` compares greater and unequal to null; `UserValue` formats its
receiver first and then raises the exact helpful `obj.toString()` null failure.
`NewUndefValue` represents the public constructor, and `ValUndef` is initialized
through it.
`ValueGetNonEnumerableErrorMsg` exposes the base source diagnostic and
dispatches the `UserValue.GetNonEnumerableErrorMsg` override. Nat, Int and Seq
provide their detailed messages through the optional
`UserObjWithNonEnumerableErrorMsg` interface; other values retain the ordinary
value diagnostic. Bounded quantifier context construction uses this public
boundary.

`NewValueVec()` now represents Java's no-argument constructor with size zero
and capacity 10. Passing an explicit capacity retains the corresponding Java
constructor, including the distinct zero-capacity growth path.

`TupleValue` now exposes singleton and pair constructors, `GetElem`/`GetElems`,
identity `ToTuple`, and empty-only `ToRcd`. The element slice accessor preserves
the source backing-array identity. `GetElem` narrows its index to Java `int`
before backing-array access. Generic tuple and record conversion reaches the
direct methods.

`NewSingletonRecordValue` represents both singleton `RecordValue` constructors:
the omitted normalization flag defaults to false and an explicit flag is
retained. `RecordValue.ToRcd` returns the exact receiver, and generic record
conversion uses that identity method.

`FcnRcdValue.ToFcnRcd` likewise returns the exact receiver. Its source-named
`ToRcd` conversion handles string-domain functions, while the older native
`ToRecord` spelling delegates to it. Generic conversion uses the direct methods.
`NewFcnRcdValueVec` also represents the public ValueVec-domain constructor and
copies the vector's active elements. Set-of-functions enumeration uses that
constructor, so each generated function owns an independent domain slice.
`GetDomainAsValues` exposes the source accessor and returns the interval's fresh
value array or the explicit domain backing slice; the older `DomainAsValues`
name remains a compatibility delegate.
Direct linear selection, thresholded binary selection and non-normalizing size
on `FcnRcdValue` reject nil receivers before domain or value-array access.

`IntervalValue.IsSubsetEq` exposes the source subset method. Interval operands
use Java's bound-containment fast path; other right-hand values use ordinary
enumeration. The evaluator dispatches through the direct method.
`IntervalValue.ElementAt` narrows its index to Java `int` before the source
negative short circuit, size check and integer-value construction.
Tuple comparison and equality, explicit-set random selection and tuple
conversion, and interval comparison, equality, subset, indexed-element,
random-element, fingerprint and value-array access all enforce Java's
receiver-first null behavior.

`ModelValueCompareTo`, `ModelValueEquals`, and `ModelValueMember` expose the
source's public reverse-dispatch hooks. Primitive, composite, set and module
callers now invoke those public methods directly.
All three reject a nil model-value receiver before type inspection or fallback
results.

`ValueExcept` now exposes its path/value constructor and both shallow cursor-copy
forms through `NewValueExcept` and `NewValueExceptCopy`. `String` preserves the
source cursor/path rendering, and advancement shares path and replacement data.
`CheckArg` validates single arguments or flattened tuple arguments against the
owning function lambda's formal domains.

`FcnParams.GetFormals` and `GetDomains` expose the exact backing slices. The
public `IsTuples` field remains the Go representation of both Java's public
tuple-flag field and same-named accessor.

`FcnLambdaValue` now exposes `GetBody`, `GetRcd`, `GetParams`, `GetCon` and
`HasRcd`, plus the one-argument copy constructor and source-named `ToRcd`.
Liveness and function-context binding use these public carrier boundaries.
`MakeRecursive` rejects a nil outer receiver before entering its source catch
region; retained null contexts still fail inside that region before changing
the lazy-evaluation control flag.
`EvalControl` flag queries, setters, conditional priming and semantic-equivalence
comparison narrow their inputs to Java `int`; setters return only Java-width
control values.
Function-lambda and predicate-set constructors store only Java-width controls;
lazy-value cache lookup, evaluation and cache publication use the same boundary.
Function-lambda application and operator method/evaluating/priority/callable
dispatch narrow controls before callbacks or recursive argument evaluation.
`Tool.Eval` parsing and callback dispatch, plus the direct pure, implementation
and application evaluator entries, narrow controls before semantic dispatch.

`OpLambdaValue` constructors retain a supplied null context and copy-overload
tool. The copy overload throws for a null source lambda, matching Java field
access during constructor delegation. Evaluation no longer substitutes an
empty context or undefined result for retained null context/tool carriers.

`MethodValue.GetMinLevel` and `EvaluatingValue.GetMinLevel`/`GetOpDef` expose
the native operator metadata used by level analysis and trace/counterexample
definition lookup. Formal-parameter arity, action-item kind, operator-lambda
control, method minimum level, and evaluating minimum-level/priority metadata
narrow to signed Java `int` at their constructor boundaries.
`PriorityEvaluatingValue.Add` rejects a nil wrapper receiver before validating
the added handle, preserving Java instance-dispatch precedence.
`OpRcdValue.AddLine` rejects a nil receiver before its source catch region and
before vector-length allocation or element access.

`LazyValue.GetNumTimesThatANewValueWasCached` exposes the cache insertion count
used by the upstream debugger tests.

`LazyValue.Eval` accepts Java's zero-, one-, and two-state forms. The zero-state
form uses the current `EmptyState`, the one-state form supplies a null successor,
and public cached lookup throws on a nil receiver. Direct evaluation also
rejects a nil outer receiver before overload selection or tool validation.
`LazySupplierValue.GetValue` likewise rejects a nil subclass receiver before
discarding its evaluation arguments or reading the retained supplier.
Java's static `TLCState.isSubset` entry point is exposed as
`TLCStateIsSubset`; lazy-value cache eligibility uses it directly, while
`IsStateSubset` remains a compatibility delegate to the same partial-state
algorithm.

`NewLazyValueWithCostModel` represents the source constructor without an
explicit cache flag and preserves its coverage lookup delegation. Supplier lazy
values use that overload and retain the source default cache marker.
`LazyEvalOff` captures `tlc2.value.impl.LazyValue.off` once per represented
classloader and emits the source disablement notice once. The reset helper is
for idle isolated runtimes.
The direct lazy cacheability predicate rejects a nil receiver before inspecting
the retained value, matching Java instance dispatch.

Every `Enumerable` now exposes `IsSubsetEq`. The evaluator dispatches through
that interface, intervals retain their bounds shortcut, powersets retain their
base-set reduction, and K-subsets retain naive enumeration.

Every `Enumerable` also exposes `GetRandomSubset`. Lazy sets, set operations,
powersets, and K-subsets use the Java superclass path: convert to an explicit
set, then delegate to its random-subset implementation. Requested sample counts
narrow to signed Java `int` across indexed, coin-tossing and product
enumerators, materialized random subsets and probability-based powerset
sampling. Bounded-length powerset sampling also narrows its requested count and
maximum subset length, performs `maxLength + 1` at Java width, and preserves the
resulting typed negative-array, zero-length index and zero-modulus failures.

Java's `Enumerable.elements(int)` overload is exposed uniformly as `ElementsK`.
It retains specialized explicit-set, interval, product, powerset and K-subset
enumerators and the superclass collect-then-sample path for other sets.

`ElementsOrdered` and `EnumerableOrdering` expose the ordering overload.
Powersets traverse normalized without eager materialization, K-subsets retain
their endless randomized generator, and bounded `CHOOSE` calls this API.
`RandomSubsetGenerator` now exposes that public Java nested class directly;
constructors on both folded subset owner types normalize the retained base,
narrow `k` to Java `int` and preserve the typed invalid-range failure.
All 12 concrete enumerable carriers now validate their receiver before ordinary
or sampled traversal, ordering selection, random-subset allocation and subset
comparison. Shared default adapters preserve that receiver-first ordering for
typed-nil interface values.
The 16 powerset and K-subset helper entry points likewise validate the receiver
before reading the base set for unranking, counting, bounded enumeration,
materialization, randomized selection or lexicographic traversal.
The nested unrank and K-element enumerator APIs also validate receivers before
subset lookup, randomized selection, sorting or set collection.

`ValueEnumerationAll`, `ValueEnumerationForEach`, and
`ValueEnumerationAsSet` expose the Java interface's default operations as Go
adapters. They consume from the current position and raise deferred enumerator
failures at the matching `nextElement` boundary.
All 20 concrete value enumerators reject nil receivers at both `Reset` and
`NextElement`, including no-op and exhausted-enumerator paths.

`FunctionValue` now unifies tuple, record, counterexample, function-record and
function-lambda dispatch. The first three expose source-named control/domain
methods, and `FcnRcdValue.ApplyArgs` retains Java's default tuple wrapping.
All five carriers validate their outer receiver before argument-array handling,
threshold lookup, source wrapping or embedded-record delegation; counterexamples
dispatch explicitly instead of losing their outer receiver through promotion.
`SelectPath` exposes final Java `Value.select(Value[])`, rejects null receivers
and paths, and walks each component through `FunctionValue.Select`.

Boolean, integer, string, model, interval and undefined values expose their
source `Mutates` override. `ValueMutates` dispatches through that method and
rejects a null value at the Java receiver boundary.

`OperatorValue` now unifies all six `OpValue` implementations. Operator lambdas,
operator records and native methods expose the source default tool-aware
evaluation path; direct and tool-aware dispatch use the shared interface.
All six carriers reject nil outer receivers across both evaluation forms before
argument handling, method metadata access or promoted subclass delegation.

Boolean, integer, string and model values expose `IsAtom`; `ValueIsAtom`
represents the `IValue` default and recursive `SetEnumValue.IsSetOfAtoms`
dispatches through it with the source null diagnostic.

`ValueHasData`, `ValueGetData`, and `ValueSetData` expose the base `Value`
defaults and dispatch the mutable attachment override on `ModelValue`.
The three direct `ModelValue` data-attachment methods reject nil receivers
before reading or mutating the attachment.

`ValueToSetEnum`, `ValueToFcnRcd`, `ValueToRcd`, and `ValueToTuple` expose the
base nullable conversion methods and their existing concrete overrides.
`ValueToState` exposes the adjacent base method; function records now retain
Java's string-domain conversion through `RecordValue.ToState`.
`NewPrintTLCState` exposes the public printable-state constructor through the
folded mutable-state carrier, retaining the record and underlying state by
identity without eagerly dereferencing either nullable argument.
All 14 concrete tuple, record, function-record and function-lambda conversion
methods reject nil receivers before identity returns, metadata inspection or
materialization.
`ValueToUnquotedString` and `ValueGetTypeString` expose the common debugger and
IOUtils rendering boundaries; `StringValue.ToUnquotedString` exposes the source
override and the prior `UnquotedString` name remains a delegate.
`ValueKindString` uses Java's signed-byte image index and preserves typed bounds
and source-aware fingerprint failures instead of returning an unknown-kind
fallback.
`ValueToTLCVariable` and `ValueGetTLCVariables` expose debugger-variable
construction and nested expansion, including the existing concrete overrides.
`ValueSetCostModel`, `ValueSetSource`, `ValueGetSource`, and `ValueHasSource`
expose inherited value metadata without widening the minimal Go value interface.
The corresponding direct `BaseValue` methods enforce Java receiver-first null
behavior while preserving normal cost-model and source set/get/clear semantics.
`FcnParams.IsTuples` completes the source parameter-carrier interface and returns
the shared tuple-flag slice; Tool function application uses that accessor.

`NewSingletonSetOfTuplesValue` and `NewPairSetOfTuplesValue` expose Java's
one-factor and two-factor Cartesian-product constructors. Tuple-product and
record-set construction enforce the source nonempty and matching-length
assertions after null-array checks.
`DummyEnum` is the shared public lazy-set cache sentinel. Product sets,
powersets, set operations, streams and distributed payloads preserve its exact
pointer identity instead of parallel native flags.

`NewRecordValueFromLocation`, `NewRecordValueFromOpDef`,
`NewRecordValueFromAction`, and `NewRecordValueFromActionField` expose the
source metadata/action record constructors used by TLCGet and counterexamples.
`NewRecordValueFromOpDecl` and `NewRecordValueFromOpDeclField` provide the
matching declaration metadata constructors used by `TLCGet("variables")`.
`NewExtendedRecordValue` exposes Java's copy-and-append record constructor and
owns `_ALIASEvalError` state construction.
`NewRecordValueFromStateInfo` exposes the ordinary `TLCStateInfo` constructor;
`TLCStateInfo.ToRecordValue` retains printable-state record identity before
entering that constructor. `GetOriginalState` and `GetAction` expose the source
getter names; the older shortened spellings remain compatibility delegates.
Checker, liveness, debugger, model and counterexample paths use the source-named
methods.
`NewFcnRcdValueFromValues` and `NewFcnRcdValueFromMap` expose the source list
and map constructors. Checker and simulator register aggregation use the map
path with OpenJDK-compatible hash iteration.

`NewBoolValue` now represents allocating `new BoolValue`; `BoolValueOf` returns
the canonical `BoolTrue`/`BoolFalse` values used by evaluators and streams.

`OpRcdValue`, `MethodValue`, and `EvaluatingValue` expose their source
`Initialize` overrides. `InitializeValue` dispatches to them before applying the
generic deep-normalize/fingerprint default.

`Reducible` now unifies explicit sets and intervals for eager difference,
intersection, union, membership, size and traversal. Evaluator and union paths
dispatch through the interface rather than concrete type lists.
Both carriers reject nil receivers across all six operations before operand
membership, empty-set identity or result allocation.

`UserObjIsEmpty` and `UserObjGetNonEnumerableErrorMsg` expose Java's overridable
defaults without requiring optional methods on every Go user object. Generic
emptiness and `UserValue` diagnostics use these adapters.
The built-in Nat, Int, STRING, ANY and sequence objects implement the optional
buffer printer used by `UserValue`. Sequence-set rendering therefore preserves
the caller's offset and swallow flag while recursively printing its range.
The pointer-backed sequence object rejects nil receivers across comparison,
membership, finiteness, emptiness, diagnostics and both rendering forms; the
other built-in user objects are value-backed.
`UserValue` also rejects a nil backing `UserObj` at every Java dereference point:
user-value comparison, membership, finiteness, diagnostics and rendering. Its
comparison with `ModelValue` still returns first without touching the backing
object, matching the Java branch order.
All pointer-backed buffer printers now reject nil receivers before formatting;
this includes the `KSubsetValue` symbolic form when expansion is disabled.
All 12 concrete set-materialization methods reject nil receivers before cache
inspection, traversal, allocation or identity return.

`Enumerable` now embeds the full `Value` contract, including source `size` and
`member`. `PermutationSubgroup` accepts `Enumerable` directly, and symmetry
validation passes its already-validated explicit set.

`ValueVec` exposes source-named public aliases for element insertion, indexed
access, capacity/size, first/last access, emptiness, and indexed mutation.
Its indexed access, mutation and insertion arguments plus minimum-capacity
requests narrow to Java `int` before array or growth behavior.
`NewValueVecFromCollection` copies collection elements into fresh storage.
Generic `Vect` applies the same narrowing to indexed access, mutation,
insertion and search starts, plus explicit minimum-capacity requests.

`MVPerm.GetAllModelValues` exposes the retained non-null permutation range in
model-value index order. Symmetry validation uses this public accessor.

`DebugTLCVariable` exposes the `TLCVariable` setter surface and `GetNested`.
Constructors, accessors, cloning, nesting, comparison, and menu-context methods
reject nil receivers or required values at the Java boundary.
Variable-reference setter inputs narrow to signed Java `int` before storage.
Debugger scope frame IDs and variable-request references narrow to the same
width before lookup or random child-reference seeding.
Base, state, action, initial-state and next-state frame variable dispatch also
narrows before comparing context, state and nested-value references.
Synthetic trace-frame display widths narrow before constructing the padded
state-level label.
Goto-state references narrow before initial/successor state-map lookup at both
the debugger request and frame selection boundaries.
Stack-trace start/level pagination narrows inputs and adds the requested range
at Java `int` width; an overflowed reversed sublist raises the typed Java
argument exception.
Source-breakpoint line and optional column inputs narrow before response storage,
condition parsing and Java-width synthetic-location bound arithmetic.
Breakpoint hit conditions parse with Java `Integer.parseInt` range semantics;
invalid or overflowing decimal text yields the source fallback of zero hits.
Derived debugger context, stack and state-trace references add their fixed
offsets at Java `int` width before lookup or scope publication.

`CounterExample` exposes the trace/console, trace/loop, and trace/action/loop
constructor overloads. Safety and liveness builders use them. Loop ordinals
and their decremented indices use signed Java `int` arithmetic before lasso
selection. `ToTrace` now preserves Java cast, null, tuple-index, and
destination-index failures.

`SetEnumValue` now exposes its no-argument/cost-model and singleton constructors
and recursive `IsSetOfAtoms` query. Atomic sets contain only model, integer,
string or Boolean values at any nested set depth; malformed null vectors and
elements retain their exact helpful failures. Its direct `ToSetEnum` conversion
returns the unchanged receiver and does not normalize it; generic set conversion
uses that identity path.

State-backed `RecordValue` construction now retains the JVM's exact helpful
null messages: null states fail at `TLCState.getVars()`, and a null action after
a valid state fails at `Action.getParameters()`. The paired debugger constructor
checks both state and successor up front with their distinct source names. A
null installed variable declaration retains the exact `vars[i]`/`getName()`
diagnostic across the shared constructors. Direct action-record construction,
including the extra-field form used by coverage metadata, also rejects a null
action at `Action.getParameters()`; callers that need an unknown action supply
`UnknownAction` explicitly. The map-backed constructor rejects a null map at
`Map.entrySet()` instead of treating it as an empty record. Constraint metadata
also retains the exact `OpDefNode.getName()` failure for a null operator.
Direct nil `OpDefNode` access now raises the typed Java `NullPointerException`
across locality/source, instance-path, arity/recursive/declaration metadata,
string, signature, comment and human-readable-image methods.
`SymbolNode` now enforces that receiver contract across coverage, source/tree,
tool-object, name/signature/image, declaration-kind and string operations.
`SemanticNodeBase` enforces it across tree/image, identity/hash, tool slots,
source/location, standard-module and level-metadata operations. Its public
identity accessor uses Java's exact `GetUid` spelling throughout semantic graph
traversal, hashing, debugger IDs and fingerprint trace comparison. `GetUID`
remains a compatibility delegate, and old adapters retain fallback recognition.
`ModuleNode` enforces it across identity/status, declarations/definitions,
statements, EXTENDS traversal and semantic path lookup.
`NumeralNode`, `DecimalNode` and `StringNode` enforce it across their direct
value, rendering and level operations while preserving valid literal metadata.
Their public SANY depth renderers retain semantic UID/kind/comment prefixes,
large-number details, source images, escaping and UTF-16 string lengths.
Formal parameters, operator declarations and operator arguments expose the same
source depth rendering with arity, defining-module and operator-UID details.
The recursive application renderer covers real nested and higher-order calls,
bounded quantifiers and `subExpressionOf` definition links. Runtime definitions
retain the extra recursion, LET-level, parameter-nullness and symbol-table state
required by that rendering. `AtNode` also exposes its complete depth renderer.
`LetInNode` now exposes its source depth renderer, including filtered context
entries in Pair-history order and recursive rendering of the `IN` body.
`LabelNode` and `ThmOrAssumpDefNode` expose their source depth renderers;
canonical labels retain Assume/Prove, goal-clause and subexpression metadata.
`Subst`, `SubstInNode`, `APSubstInNode` and `AssumeProveNode` expose their
source depth renderers. Imported substitution wrappers retain both module
owners after late bridge binding. `OpApplNode` depth rendering is now public.
`NewSymbNode`, `TheoremNode` and `LeafProofNode` also expose their source depth
renderers, covering bounded `NEW` assumptions and leaf proof facts/definitions.
`NonLeafProofNode`, `DefStepNode`, `UseOrHideNode` and `InstanceNode` expose
their depth renderers, including proof contexts and module substitutions.
`AssumeNode` and `ModuleNode` expose their source depth renderers, including
constant-module classification, Java HashSet EXTENDS order and top-level lists.
`ExternalModuleTable` and its entries expose their dependency-ordered depth
renderers, including the source null-module diagnostic branch.
`ModuleNode.print`, `ExternalModuleTableEntry.printMTE` and
`ExternalModuleTable.printExternalModuleTable` are exposed with stdout and
owned-writer forms, preserving depth cutoffs, initial-context filtering,
dependency insertion order and direct-print indentation.
`SanySemanticNodeToString` provides concrete runtime dispatch, while the public
semantic-base, operator-base and `OpDefNode` functions preserve source `super`
call boundaries.
`SanyContextGetContextEntryStringVector` exposes filtered or complete context
vectors with source Pair-history order and recursive semantic rendering.
`SanyLevelNodeDefaultLevelDataToString` and
`SanyLevelNodeLevelDataToString` expose inherited and polymorphic level-debug
rendering over the canonical runtime graph. Shared parser metadata preserves
the module, operator-definition, assumption, theorem and instance overrides.
`SanyLevelNodeLevelCheck` and `SanyLevelNodeLevelCheckNext` expose the explicit
iteration and incrementing source overloads through the parser-owned checker,
including live error-log mutation and metadata updates.
`SanyLevelNodeLevelCheckSubnodes` exposes the shared child-aggregation helper,
retaining module exclusions, lazy child resolution and Java's partial
`levelChecked` mutation before null-array or null-child failures.
`SanyLevelNodeGetLevel`, `SanyLevelNodeGetLevelParams`,
`SanyLevelNodeGetAllParams` and `SanyLevelNodeGetNonLeibnizParams` expose the
scalar level and live Java-order symbol sets. Repeated getters retain carrier
identity, and set mutation updates the parser-owned level metadata.
`SanyLevelNodeGetArgLevelParams` similarly returns a cached live
`SanyArgLevelParamSet`, preserving operator/position/parameter equality,
Java-order iteration and backing-set mutation.
`SanyLevelNodeHashSetToString` and `SanyLevelNodeALPHashSetToString` expose the
two remaining static `LevelNode` set renderers over those live carriers.
`WalkSemanticGraph` accepts external module tables and entries, reconstructing
Java Hashtable enumeration order from retained insertion history.
`AssumeNode`, `TheoremNode` and the exposed `ThmOrAssumpDefNode` methods now
enforce the same boundary; `AssumeNode.GetIsAxiom` exposes its source query.
`OpApplNode` exposes its source operator, operand, quantified-symbol, tuple,
bound, opcode and child queries plus operand/operator mutation. Backing arrays
retain Java ownership, while flattened symbols and children are fresh results.
Built-in token lookup, opcode classification and both semantic `HasOpcode`
queries narrow integer arguments to Java width. Null operator lookup and
negative token lookup retain Java's null and array-index failures.
`OpArgNode` retains and exposes its source operator, constructor-time name and
arity, and containing module; the canonical SANY bridge transfers all four.
`LabelNode` likewise retains and exposes its source name, arity, body and goal;
both canonical and direct AST bridges now preserve available label identity.
`AtNode` retains the enclosing EXCEPT and replacement-component applications,
and exposes their shared references plus source base/modifier derivation.
`Subst`, `SubstInNode` and `APSubstInNode` expose their source metadata and
mutators. Canonical conversion retains substitution syntax/implicit flags,
module provenance, backing arrays and shared substitution-record mutation.
Substitution access indices narrow to Java `int` before backing-array access.
`LetInNode` exposes its cached filtered-definition array, body and children;
empty valid LET definitions retain a concrete zero-length backing array.
`AssumeProveNode` and `NewSymbNode` expose their source flags, statements,
goals, declarations, bounds and child-array semantics with typed receivers.
Leaf/non-leaf proofs, definition steps, USE/HIDE steps and INSTANCE proof nodes
also expose their source metadata, step-name mutation and exact child ownership.
`ThmOrAssumpDefNode` now retains and exposes theorem/assumption flags, proof,
parameters/arity, module provenance, expression kind and Leibniz metadata.
Operator definitions, theorem definitions and labels share a source-shaped label
table with exact bucket enumeration, lookup, duplicate rejection and rehashing.
Source `GetChildren` methods now cover operator definitions, labels, assumptions,
theorems and modules. `ModuleNode` retains Java's cached backing child array, and
`SemanticChildren` delegates to these source methods.
`OpDefNode` also exposes its source parameter, body, arity and numbered-step
methods. Parameter arrays retain their backing storage, while arity remains an
independent field and the canonical bridge preserves variadic `-1` arity.
The linked operator-definition and runtime-symbol arity setters narrow to Java
`int` before updating both views.
Its level-bound, weight, operator-condition and Leibniz queries now preserve
checked-state failures, variadic indexing, result ownership and parser metadata.
Their one-, two- and three-dimensional indices narrow through the shared Java
array-access boundary after checked-state and variadic handling.
Runtime symbols now expose source arity/locality, declaration/formal parameter
identity, module provenance, occurrence and shared-source comparison; canonical
binding retains these distinctions for module and proof-local declarations.
`ModuleNode` retains source definition, instance, record and top-level
collections with Java's direct, cached or fresh result ownership, including the
two deliberately unimplemented null theorem/assumption accessors.
Theorem/assumption definitions expose their checked maximum-level, weight,
minimum-maximum-level and operator-condition queries, with parser-computed
matrices retained through canonical and instantiated runtime views.
`AssumeNode` delegates level and level-parameter queries to its expression as
Java does. Canonical conversion retains the checked expression parameter set.
Its all-parameter query likewise delegates through the complete expression
graph, and generic all-parameter traversal treats the wrapper as transparent.
Checked canonical semantic nodes now retain their parser-computed level
parameter symbols generally, including operator definitions consumed by
configuration invariant validation.
Semantic nodes expose source kind mutation and syntax-attached pre-comments.
The complete public `ASTConstants` integer kind surface is available through
`Semantic*Kind` values, and `SanySemanticKindNames` is the live 38-entry source
name array used by level-check failures. Its proof operator names are exposed
through the shared interned `tlc.Op*` constants alongside the earlier expression
operator names.
Comment access retains the syntax array and Java's null/cast failures, and its
print form preserves `SyntaxTreeNode.PreCommentToString` spacing exactly. The
translated semantic-corpus test now uses this public surface.
`SanySyntaxNode` exposes the non-conflicting Java `TreeNode` filename, interned
filename, location, image, interned image, pre-comment, first-child, generic-op
and local-layout accessors. `GetHeirs` returns a fresh non-nil empty slice when
there are no children, matching Java's shared empty-array contract. Public Go
fields continue to represent source methods whose names would collide with
those fields.
Its remaining static text helpers are exposed as `SanyPreCommentToString`,
`SanyUnboxComment` and `SanyUnboxBackslashStarComment`; node string rendering
and static operator-definition traversal also use the source behavior. Nullable
overloads retain Java's null-in/null-out contract.
`PrintST` completes the executable public tree surface with source indentation,
operator annotations, pre-comments and null-child output; `PrintSTTo` provides
the same bytes on a caller-owned Go writer.
`SanyParser.Parse` restores the public `ParseTree.parse` wrapper: it captures
source parse/token-manager failures, reports parser messages, publishes the root
only after a complete parse and returns the source success boolean. `RootNode`,
`ModuleName` and the existing fresh `Dependencies` expose the remaining contract.
The generated parser surface also exposes token-manager construction and
reinitialization, current/lookahead token access, generated parse-exception
creation and tracing no-ops. A null installed manager is retained until token
demand instead of selecting the pretokenized EOF adapter.
Its live token-manager carrier is public as `TokenSource`; direct replacement
changes subsequent generated-parser token reads, and reinitialization updates
the same field.
The current-token carrier is likewise public as `Token`; lookahead and
consumption retain its live `Next` links, direct replacement adopts the
caller-owned chain, and reinitialization installs a fresh dummy token.
The generated parser's `PrefixOpToken`, `NEPrefixOpToken`, `InfixOpToken` and
`PostfixOpToken` entry points expose the existing exact JavaCC choice sets and
parse-exception behavior.
Their `PrefixOp`, `NonExpPrefixOp`, `InfixOp` and `PostfixOp` syntax-node
productions retain source node kinds, token leaves and canonical last-operator
updates.
`ConstantDeclarationItems` exposes the generated declaration-item production
with identifier arity and prefix, infix and postfix declaration forms; the
earlier singular Go spelling remains an alias.
`ParenthesesExpression` exposes the generated choice among ordinary
parentheses, braces, square brackets, tuples/actions and fairness expressions;
the earlier `PrimitiveExpression` Go spelling remains an alias.
`Expression` exposes the generated no-argument operator-stack production while
the `ExpressionUntil` family continues to carry Go's context-specific stop
predicates.
`BangExt` exposes the generated intermediate `N_IdPrefixElement`; the existing
`BangExtension` tuple form remains the internal adapter used to assemble final
general identifiers.
`Junctions`, `DisjList` and `ConjList` expose the generated junction-list
boundaries over a shared item accumulator; `Junctions` returns the list while
the latter two retain their source void consumption contract.
`JuncItem(int)` exposes the generated item boundary over the same parser path,
using the caller-supplied syntax-node kind without validating it against the
consumed conjunction or disjunction bullet.
`OpSuite` restores the final missing generated-production name as the source
void wrapper over `OpOrExpr`; the latter accepts an optional Go stop predicate
while remaining directly callable with the Java signature.
`OpenExpression`, `LetIn`, `Case`, `CaseArm`, `OtherArm`,
`UnboundOrBoundChoose`, `Lambda`, `IfThenElse`, `SomeQuant` and `SomeTQuant`
likewise accept their zero-argument generated signatures; optional Go stop
predicates preserve context-specific callers.
`ExtendableExpr` has its generated public void signature over an inactive
standalone operator stack, retaining Java's consume-then-null-failure direct
call boundary; active expression parsing uses the private stack-aware helper.
The generated public `LookingAhead` flag selects `GetToken`'s retained scan
cursor. Its initial null cursor, indexed null failure, post-scan cursor and
reinitialization retention match JavaCC state.
The generated public `jj_nt` carrier is exposed as `NextToken`. Generated
next-kind decisions distinguish uncached fetches from prelinked tokens and
reuse the cached decision until consumption, while direct token APIs and
reinitialization leave the public carrier unchanged.
The parser root carrier is public as `ParseTree`; `RootNode` returns the live
field, reinitialization and failed parsing retain it, and successful parsing
replaces it only after the complete compilation unit is available.
The final public dependency vector is exposed as `DependencyList` with interned
entries. `Dependencies` returns fresh string snapshots, while caller additions,
parser additions, identity and contents survive reinitialization.
`SanyVector` covers the source legacy wrapper's constructors, indexed access and
mutation, first/last, clearing, append, identity-based append without repeats,
snapshot enumeration and exact list rendering. Invalid capacity, bounds, nil
receiver/input and null-element rendering use typed Java failures.
`SanyIntWrapper` preserves both constructors, 32-bit wrapping increment/set/value
operations and typed nil receivers. `SanyIndent` and `SanyIndentSB` share the
source character result while reading the live nine-entry `SanyStringBlanks`
array, including clamping, terminal-newline and negative-index behavior.
The deprecated `SanyAssert` constructor and both assertion helpers preserve the
source class's intentional no-op behavior for every argument.
Generated tokens expose the source `SpecialToken` field name. Regular tokens
point to the last preceding special token, backward `SpecialToken` and forward
`Next` links retain the complete special-token chain, and syntax comment
attachment traverses that live carrier.
Their source position carriers are likewise public as `BeginLine`,
`BeginColumn`, `EndLine` and `EndColumn`. Scanner, parser diagnostics,
lookahead, rebasing and token-backed syntax nodes read those live scalars;
`SourceFile` separately retains Go's filename context for reconstructed
positions and ranges.
Its Reader-style string constructor and reinitializer preserve the same
character-stream and token-manager identities across input replacement;
production syntax parsing now enters through that constructor.
Syntax-tree proof-level, level and parent accessors/mutators enforce Java's typed
nil-receiver boundary; recursive level/parent initialization also rejects null
children at the source dereference point.
Filename and image accessors distinguish absent Java fields from valid interned
empty strings, preserving null unique-string results and dereference failures.
The source public `OriginalImage` carrier retains raw numbered proof-step text
independently of the normalized `Image`, and proof translation reads that live
field when identifying implicit steps.
`SanySyntaxNodeImages` exposes the source public interned image array. Node
constructors snapshot its selected pointer, including null, while `PrintST`
consults the live table only for an absent node image as Java does.
`First` raises the typed array bounds failure for an empty selected child array.
`SANYNullSyntaxNode`, `NewEmptySanySyntaxNode` and
`NewSanyBuiltinSyntaxNode` restore the parser-layer sentinel and base constructor
states, including built-in provenance, nonnil empty child arrays and `-1` levels.
All syntax constructors retain Java's concrete shared empty pre-comment state,
and nonnil empty child inputs remain nonnil so `Local` observes their layout.
Ordinary constructors reject null child entries while
`NewSanyIgnoredSyntaxNode` preserves the special Generator `@` constructor's
caller-owned, possibly-null child array and deliberately uninitialized metadata.
`SanyOperatorStack` exposes the source empty/size, push, top/bottom and pop
boundaries. An inactive stack retains Java's distinct null/current-top behavior
and typed failures instead of implicitly creating a stack or returning size zero.
Its source `OSelement` adaptation exposes node/operator access and operator
classification on values returned by `TopOfStack`, with typed nil failures.
`SanyOperatorInfo` exposes the source singleton void operator, string and
identifier access, all fixity/associativity predicates, prefix-declaration
classification and table-existence query. Its live public `Fix` and
`Associativity` carriers drive those methods after direct caller mutation.
String and interned-identifier constructors preserve the supplied precedence,
carriers and identifier identity; the interned form retains Java's nullable
identifier boundaries. Instance methods reject nil receivers with Java's typed
failure.
The source-shaped operator lookup, existence, mixfix and synonym APIs accept
interned identifiers and return shared operator pointers. Canonical spellings
and every synonym therefore expose the same mutable `Fix` and `Associativity`
carriers. String/value helpers remain for internal parser callers.
`SanyPrintOperatorTable` reproduces the source debug output; its generated key
order simulates Java `Hashtable` growth, collision chains and enumeration with
UTF-16 string hashes rather than relying on Go map iteration.
`SanyToken` exposes JavaCC's source `newToken` factory and image rendering. A
default token retains kind zero and a null image, while lexer-produced tokens
distinguish an assigned empty image; instance rendering rejects a nil receiver.
`SanyParseError` and `SanyParseErrors` preserve the parser's custom/default
message pairs, ordered shared collector and fresh returned error arrays. The
operator stack now reports incomplete reductions through that source carrier.
`SanyParseException` exposes JavaCC's default, message and generated
constructors, public current-token/expectation/token-image carriers, full message
rendering and `ParseExceptionExtended` short rendering. Generated constructors
retain caller-owned arrays and all instance/static renderers reject nil inputs.
`SanyTokenMgrError` exposes JavaCC's four reason codes, default, message and
lexical constructors, message access and protected escape/lexical helpers. The
token manager now throws this public carrier while retaining native diagnostics.
`SanyTokenManager` exposes JavaCC's explicit-state constructor, default and
explicit-state reinitialization, debug-stream setter and typed next-token
receiver boundary. Its source-shaped constructors and reinitializers retain a
caller-supplied `SanyCharStream`; string forms remain convenience delegates.
Its live public `DebugStream`, literal-image table, lexical-state names and
next-state table retain JavaCC's caller mutation semantics; scanning reads the
exported tables directly.
`SanyCharStream` exposes JavaCC `SimpleCharStream` construction, tab-size,
token/character reads, position queries, backup, string reinitialization,
image/suffix, cleanup and begin-position adjustment over its UTF-16 buffer. Its
source-public `BufPos` field tracks the live circular-buffer cursor directly.
`SanyJunctionListContext` is now the parser's single junction-context
implementation. Its public start, terminate, bullet-alignment and indentation
queries preserve Java's nested stack behavior, short-circuit invalid non-bullets
and raise typed null, empty-stack and invalid-token failures.
`ParseSANYLocation` and `ParseSANYLocations` recognize all four Java `Location`
formats with source anchoring, trimming, module-name, 32-bit integer and
pattern-order semantics, including duplicate extraction across overlapping
patterns.
The four named matchers and `SANYAllLocationPatterns` are public. Single-location
parsing uses the named matchers, while multi-location extraction reads the live
aggregate slice so caller element replacement changes its traversal as Java's
public array does.
The location adaptation also preserves intern-token module ordering, exact
same-module containment, raw end-coordinate comparison, coordinate-array
clamping and Java location text/line queries. `Position.String` remains the
generic diagnostic renderer; `SANYLocationString` exposes Java's location form.
Coordinate-string parsing retains Java split and typed-failure behavior;
interned-source access, hash calculation and zero-based PlusCal region conversion
are exposed through the same adaptation.
`SemanticLevelToString` preserves Java's numeric and named rendering for all
four legal levels and illegal values, narrowing direct inputs to Java `int`.
Temporal debugger-expression rejection uses it and reports `3 (Temporal)`
instead of a shortened hardcoded name.
`SemanticConstantLevel` through `SemanticTemporalLevel`, the min/max bounds and
the live nullable `SemanticLevels` array expose `LevelConstants` scalar and
boxed-level metadata.
The public `StandardModules` utility is exposed through node/name membership
queries and in-place set filtering. `SemanticNodeBase.IsStandardModule`
delegates to the same exact nine-name table.
Semantic application construction now accepts module symbols through Java's
stub `ModuleNode.match` branch, which returns false without a diagnostic.
`ModuleNode.GenerateUnusedName` formats the current millisecond through Java's
string formatter and retries against the visible semantic context. Debugger
wrapper module and expression names now use this source method.
Runtime modules retain the canonical SANY `isInstantiated` flag, including
instances introduced only inside structured proofs, so constant-definition
processing skips parameterized instance targets as Java does.
Runtime symbols and operator definitions expose the inherited built-in query,
and modules expose their inherited name signature. Debugger hover rendering
keeps built-ins location-shaped while ordinary semantic nodes honor
`SemanticNode.showPlainFormulae`, matching Java's `TLCStackFrame` branch.
Debugger hover lookup also recovers Java's shared `SymbolNode` view for
operator definitions, theorem or assumption definitions and modules before
consulting context and semantic tool values. Standalone theorem-definition
adapters share their semantic base with that view, like parser-backed nodes.
Runtime declaration symbols now expose inherited `OpDefOrDeclNode.getComment`;
declarations and definitions use the same source pre-comment renderer.
`SanyOpDefOrDeclNodeGetSymbolTable` exposes their shared parser symbol table.
The bridge publishes table shells before adapting recursive context symbols and
binds module owners after module-shell creation, preserving source identity.
`SanyModuleNodeGetSymbolTable` retains the separate source method's hardcoded
null result and typed null-receiver failure.
`SanyModuleNodeIsConstant` invokes the parser-owned module level checker before
examining inherited definition and theorem levels, retaining live diagnostics.
`SanySymbolNodeMatch` dispatches runtime declarations, definitions, formal
parameters, theorem definitions and modules to their parser-owned match
implementations, retaining diagnostic and abort behavior.
`SanyFormalParamNodeMatchSemantic` exposes the separate formal-parameter
overload that ignores its semantic-node argument and tests only zero arity.
`SanyBuiltinOperator` and the live `SanyBuiltinOperatorProperties` slice expose
the complete 72-entry `BuiltInOperators.Properties` metadata. Context rebuilds
and metadata lookup consume that public table, retaining its argument arrays.
`SanyContext` exposes the shared semantic context implementation, constructor,
key and enumeration types. Its process-wide global context is rebuilt alongside
the canonical SANY context; built-in classification follows current node identity.
`SanySymbolMatcher` and `SanyNameAndTypeMatcher` expose source prefix/type
filtering, and `ModuleNode.GetSymbols` returns matching context symbols in source
position order while excluding internal dollar-prefixed built-ins.
`SanyExternalModuleTable` exposes the shared ordered module table and nested
entry carrier. Its name overloads, duplicate suppression, fresh module arrays,
root lookup and null-key boundaries follow the source class.
Generic `SanyContextGetByClass` filters current Hashtable bindings by assignable
runtime type. Context and external-table entries expose the source dummy level
text, and the context depth/error rendering overload preserves its fixed notice.
`SanyContextMergeExtendContext` merges retained Pair history oldest-to-newest,
skips local symbols, preserves existing bindings, suppresses shared-origin
duplicates and records the source 4800/4224 diagnostics through `SanyErrors`.
`SanySymbolTable` exposes the external and inner constructor shapes, copied
context stack, external/current contexts, module owner, stack mutation and
symbol/module resolution with the source namespace and failure behavior. Its
registration methods preserve built-in rejection, hard symbol/module conflicts,
shared-source suppression and warning-success ambiguity diagnostics. Rendering
retains Pair-history order, current-binding lookup, UIDs and context separators.
`SanyParamAndPosition` and `SanyArgLevelParam` expose the semantic level-key
carriers with reference equality, Java hashes, occurrence checks and rendering.
`SanySetOfLevelConstraints` and `SanySetOfArgLevelConstraints` preserve the
source HashMap order, copy behavior and minimum/maximum tightening rules.
`SanyLevelNodeGetLevelConstraints` and
`SanyLevelNodeGetArgLevelConstraints` return identity-stable live adapters over
the parser-owned maps; mutations and copy construction retain Java iteration
and capacity behavior.
`SanySubst` shares the runtime substitution carrier and exposes its four-field
construction, live mutation, identity lookup, basic parameter-set transforms
and fixed source level-data string. Its scalar-constraint, argument-constraint
and argument-level-parameter transforms use exact runtime symbol identities,
canonical level rechecking and Java collection order.
`SanyLevelException` exposes the semantic level checker's checked-exception
carrier with its nullable message, absent cause and exact Java class name.

The complete formatter exception hierarchy now retains its JDK constructor,
metadata, derived-message and null-cause behavior. Required strings reject null,
conversion metadata uses a Java-width UTF-16 code unit, and precision and width
metadata narrow to Java `int`. Concrete accessors raise typed
`NullPointerException` on null receivers. The private illegal-index sentinel
message is also preserved. Paired probes and focused SANY/value checks pass.

`OffHeapDiskFPSet.OffHeapRuntimeException` now exposes its sole source cause
constructor, concrete message/cause accessors and exact nested Java class name.
Null causes and null outer/embedded receivers retain Java behavior. A probe
against the checked-out compiled class and focused off-heap checks pass.

Debugger reset/abort control exceptions now retain their source default runtime
metadata and exact `IDebugTarget` nested class names. Reset construction enforces
the source assertion that its frame is non-null; concrete accessors type null
receivers. Assertion-enabled/default Java probes and focused debugger checks
pass.

The retained generic `NamingException` carrier now distinguishes default,
explicit null, empty and nonempty messages, keeps the constructors' null cause,
and types null message/cause receiver access. This adds no JNDI name model or
mail functionality. JDK and native probes agree, and all packages compile.

`BigInt(String)` now follows Java `BigInteger` decimal parsing instead of
delegating malformed input to a raw Go panic. It preserves typed constructor
failures, sign validation, nine-code-unit failure groups, Java BMP decimal
digits and supplementary-digit rejection. All 65,536 BMP digit classifications
match the installed JDK. Paired probes and the existing
BigInt/ByteUtils/external-sort checks pass.
The random-bit constructor narrows its bit count and computes
`(numBits+7)/8` at Java `int` width, including wrapped negative arguments and
byte-array lengths.

`Combinatorics.choose`, `sumChoose` and `choosePairToInt` narrow arguments and
pairing arithmetic to Java `int`. `sumChoose` masks `long` shift distances like
the JVM for its full-set and all-but-one shortcuts.
`fact`, `bigChoose`, `slowBigChoose`, `bigSumChoose` and `binomial` likewise
narrow their scalar inputs and use Java-width loop and subtraction arithmetic;
negative `BigInteger.shiftLeft` distances retain Java's right-shift meaning.
Mixed-radix explicit lengths and Pascal-table dimensions narrow to Java `int`;
negative `toSeq` and overflowed Pascal allocations raise typed Java array-size
failures after their source checks.

Deprecated `BigSet` now rejects nil receivers across its complete public surface,
rejects a null public backing set, and dereferences null keys in `Put` and
`Contains` like Java's underlying `Set`. Its retained empty payload-writing
behavior is unchanged. Constructor sizes narrow to Java `int`; maximum sizes
remain unnormalized, while the underlying set rejects nonpositive initial
sizes with Java's message-less `IllegalArgumentException`.

`JavaRandom` now rejects nil receivers across its complete native method
surface, before bound or allocation checks. `NextBytes` rejects a null array,
and `NextIntN` raises the typed Java
`IllegalArgumentException("bound must be positive")` for nonpositive bounds
before consuming RNG state. Java/native probes and existing randomization checks
agree on failure and representative bounded-result rows.

`Sx.car`/`cdr` now raise exact `WrongInvocationException` failures, and a null
reference remains distinct from the `Sx.nil` singleton during append,
membership traversal and pair rendering. Null pair elements now fail with typed
`NullPointerException`. `Sx.Print` restores the source streaming renderer for
atoms, integers, proper lists and dotted pairs; `String` collects through that
method. Java/native probes and existing Sx/SimpUtil checks pass.
The public `Sx.SxInt` and `Sx.SxPair` constructors are exposed as `NewSxInt`
and `NewSxPair`; the static factories delegate through them, integer values
narrow to Java's signed 32-bit range and pair constructors retain nullable
members.

The two public `MemIntQueue.DetailedFormatter` helpers are available under
class-prefixed Go names. They preserve the source physical backing-array walk,
signed fingerprint/pointer reconstruction and exact diagnostic text.
All eleven public queue and inherited set operations enforce Java receiver-first
null behavior across mutation, removal, size queries and checkpoint lifecycle.
The explicit `MemIntQueue` capacity narrows to Java `int` before backing-array
allocation, including the wrapped zero and negative-capacity boundaries.
All memory, deque, disk-state and raw-byte batch dequeue counts narrow to Java
`int` before assertion, waiting or allocation. Both disk pool-cleaner targets
narrow before their monotonic threshold comparison.

`StateDeque` now maps `ArrayDeque.addFirst(null)` to typed
`NullPointerException` before changing storage or logical size. Existing batch
prefix retention is unchanged. Its three checkpoint operations now return the
source `UnsupportedOperationException` and exact message. Focused queue checks
and all-package compilation pass.
Inherited plain length operations remain outside the deque monitor through the
same separate atomic-snapshot contract as the concrete queue implementations.

The internal nonfair reentrant read/write lock now matches Java ownership
failures: invalid write release raises null-message
`IllegalMonitorStateException`, while invalid read release retains Java's exact
message. Cross-goroutine failures leave the owner's hold intact. Focused lock and
cache checks pass. `Striped` lock-array capacities and public indexed access
narrow to Java `int` through the constructor, static factory and `getAt` surface.

Random subset hash-set construction now preserves Java's typed negative-capacity
failure and exact message. The translated Randomization and SubsetValue tests
pass.

Subset unranking and random-set allocation now preserve typed
`ArithmeticException` failures from Java's `BigInteger.longValueExact`, including
the exact out-of-range message. `KSubsetValue` constructor cardinalities narrow
to signed Java `int` before the value stores or evaluates them. Deterministic
subset unranking, cardinality, bounded enumeration, `kElements` and `kSubset`
arguments use the same Java width, including their wrapped zero and negative
paths. Existing subset and randomization checks pass.

`ExternalSortable` and the generic `ExSortUtils` read, write and append methods
are restored, including prototype-driven element construction. The existing
BigInt-specific names delegate through that source surface. Empty, truncated
and negative-length input preserve source exception types, and element-write
failures retain Java's fixed `IOException` replacement message; focused stream
checks pass.
ByteUtils fixed lengths and read/range indices plus ByteUtils/ExSort subarray
bounds narrow to Java `int` before allocation, stream access or iteration.

`DiskGraph.CommitChkpt` now returns Java's typed `IOException` for both old-file
deletion and temporary-file rename failures, retaining the established exact
message and ordering. Existing disk-graph checks pass.

Tableau reverse-path cycle detection now returns Java's typed `RuntimeException`
with the exact capitalized source message. Existing tableau disk-graph checks
pass.

`TLAPlusExecutor.Map` now returns the source `RuntimeException` and exact message
when no action/context/value mapping matches, instead of a coded `TLCError`.
`Step` now returns Java's null-message `AssertionError` when successor generation
is empty. `GetConstant` distinguishes missing definitions from present non-value
definitions and raises Java's exact `ClassCastException` for the latter. The
exposed executor lock is now reentrant like Java, including nested `Step` calls
under caller ownership. `Step` snapshots mapping context once before locking and
reuses it across iterations, matching source mutation visibility. Empty initial
state generation now fails through `StateVec.First` with Java's exact typed
array-bounds exception. Mapping assignments to unknown parameter names retain
Java's null-symbol context binding instead of being silently discarded. `Step`
now returns the JVM's typed and exact null-mapping/null-action failures before
locking. Mapping construction with a null action also preserves Java's lazy
failure: an empty parameter list succeeds, while the first named parameter
raises the exact typed null dereference. Calling `Context` on the successfully
constructed null-action mapping raises that same source failure.
An action with a null context likewise constructs an empty mapping, then raises
Java's distinct exact `Context.lookupName` null dereference when parameters are
requested. Its empty mapping returns the action's null context verbatim.
If a value is subsequently assigned, context construction raises Java's exact
typed null-`Context.cons` failure.
`TLAPlusExecutor.Map` also preserves the separate exact `Context.lookup` null
failure when a name-matching action has a null context.
Its lookup predicate now also stops on a branch/null-symbol context node with
Java's exact typed `SymbolNode.getName` failure.
Mapping a real bound value against a null requested value now returns Java's
exact typed helpful-null failure before native value comparison.
The action-context predicate also preserves Java's separate exact failure when
a non-null symbol has a null name.
The outer map loop now likewise stops on a null action entry with Java's exact
typed `Action.getName` failure.
`Mapping` construction now also stops on a branch/null-symbol context node with
the source predicate's exact typed failure, and preserves the distinct exact
failure for a non-null symbol whose name is null.

Probabilistic next-state `CASE` now returns Java's explicit
`UnsupportedOperationException` with the existing exact message, instead of a
coded general TLC error. It still fails before evaluating any CASE arm.

Disabled action composition now returns Java's detailed runtime failure from
evaluation, `ENABLED` and next-state generation. Each path retains the original
`\cdot` semantic node and context along with the exact existing message.

`Tool.MakeState` now returns Java's detailed runtime failure for both a
non-unique predicate and an incomplete resulting state, retaining the predicate
and `Context.Empty` with the existing exact messages.

Applying a tuple, record or counterexample record with multiple arguments now
returns Java's detailed runtime failure, retaining the application node and
context with the existing exact message.

Applying `DOMAIN` to a non-function and `\subseteq` with a non-enumerable left
operand now return Java's detailed runtime failures, retaining their expression
and context with the exact source messages.

The final unknown-opcode evaluation fallback now returns Java's detailed runtime
failure, retaining the malformed application and context with its exact message.

Nested `EXCEPT` selection through a non-function now preserves Java's value
source boundary: unsourced roots return a plain runtime failure, while sourced
roots retain the source in a detailed cause wrapped by `FingerprintException`.

Evaluating an `OpLambdaValue` with a null operator definition now returns the
JVM's exact helpful `NullPointerException`; sourced values wrap that cause in
`FingerprintException` like Java.

`IOUtils!IOExec` with an empty command now returns Java's typed
`ArrayIndexOutOfBoundsException` and exact message instead of a generic error
whose text merely named that exception.

`RandomAccessFile` now accepts Java's `rws` and `rwd` modes in addition to `r`
and `rw`, using the platform synchronous-write flags. Invalid modes return the
exact typed Java `IllegalArgumentException`.
Negative seeks return Java's typed `IOException` and exact `Negative seek
offset` message while retaining the wrapper's seek counter without adding
failed-call time. Construction also
normalizes Java `File` paths, checks invalid paths after mode validation,
converts open failures to typed `FileNotFoundException`, and rejects directories
in read-only mode as OpenJDK does. Operational failures return typed Java I/O
messages; failed seek, read and write calls increment counters but not elapsed
times, and close is idempotent.

The last passing normal offline full gate is `f6ab0cc`: session `9728`, exit
0, root 4,974.101s, TLC 798.269s, SANY 2.355s. Subsequent installed changes
have no current repository-wide passing gate. The current `tlc` package alone
passes normally on `a1e642e`: 25,581 passing test entries, zero failures,
526.763s, with CPUs 0-7 and no Java process. Its terminal JSON receipt is
`.codex-gotmp/full-tlc-after-metadir-cleanup-20261016.jsonl`.
The first current root Go-only gate, on `f8dfbae`, finished in 5,588.088s
with one failing test: `TestDistributedDoInitFunctorEvalException` inherited
the class-style coverage flag from earlier model tests. The fixture now resets
that flag as Java's fresh-JVM harness does (`bace515`); a focused run with a
coverage test first passes. The next root Go-only gate, session `90037`,
finished in 5,577.372s with that distributed test passing and two
`Github971` leaves failing intermittently. Java's worker empty-queue path
holds the checker monitor while setting `done` and notifying; Go had omitted
that monitor. The source-order correction is `a2f8bb5`. Before it, repeated
Github971a/c runs passed 13/50 and 11/50; after it, both passed 50/50, as
did all five Github971 tests across 50 repetitions. A short focused race
check passes. The root receipt is
`.codex-gotmp/full-root-go-only-after-distributed-coverage-20261016.jsonl`.
The full `tlc` package gate on `a2f8bb5` passed in session `18256`: 25,582
passing entries, zero failures and 531.307s, including the full disk-queue
workload. Its receipt is
`.codex-gotmp/full-tlc-after-worker-monitor-20261016.jsonl`. The corrected
root Go-only gate was interrupted in session `23759`, with partial receipt
`.codex-gotmp/full-root-go-only-after-worker-monitor-20261016.jsonl`.
This gate compiled the earlier worker-monitor snapshot. A further source parity
correction, `c9cf0d2`, holds the BFS checker monitor through the coordinator's
completion test and progress, depth and coverage continuation, then reacquires
it for the post-wait count and completion test. Focused checks, 20 repetitions
of all five Github971 tests and a short focused race check pass. A full `tlc`
package gate on this correction passed in session `24400`: 25,582 passing
entries, zero failures and 542.317s, including the unchanged full disk-queue
workload. Its receipt is
`.codex-gotmp/full-tlc-after-continuation-monitor-network-20261016.jsonl`.
Root gates exclude the three Java-launching tests. The separate `sany_tests`
package passes on `c9cf0d2` in 2.178s, and the all-package compile-only check
passes. A root Go-only gate compiled with the BFS continuation correction was
interrupted in session `50250`, partial receipt
`.codex-gotmp/full-root-go-only-after-continuation-monitor-20261016.jsonl`.
The later DFID post-wait monitor correction is `2993135`; 20 repetitions of
existing DFID lifecycle checks, original depth-first model tests and a short
focused race selection pass. Its full `tlc` package gate passed in session
`70329`: 25,582 passing entries, zero failures and 832.554s, including the
unchanged full disk-queue workload. The receipt is
`.codex-gotmp/full-tlc-after-dfid-monitor-20261016.jsonl`.
The separate `sany_tests` package passes on `2993135` in 2.164s, and its
all-package compile-only check passes.
Both root gates compiled before the DFID edit, so repository-wide verification
of `2993135` remains pending. A root Go-only gate for that correction was
interrupted in session `98546`, partial receipt
`.codex-gotmp/full-root-go-only-after-dfid-monitor-20261016.jsonl`. All three
root gates excluded the same three Java-launching tests and shared CPUs 0-7.
Commit `4762056` removes a premature DFID `workers[0]` read at `RunTLC`
startup. Java first accesses worker slots in its later join path; the Go
helper never used that early worker value. Twenty repetitions of the existing
DFID lifecycle checks and the four original depth-first model tests pass.
All three root gates compiled before this edit. Its full `tlc` package gate
passed in session `85282`: 25,582 passing entries, zero failures and 555.971s,
including the unchanged full disk-queue workload. The receipt is
`.codex-gotmp/full-tlc-after-dfid-slot-timing-20261016.jsonl`. Its root gate
was interrupted in session `51582`, partial receipt
`.codex-gotmp/full-root-go-only-after-dfid-slot-timing-20261016.jsonl`. It
excludes the same three Java-launching tests; a three-hour harness timeout
allowed for the concurrent CPU-limited gates without changing test bounds. An
all-package compile-only check passes on `4762056`. The four interrupted root
gates had no failing test leaves in their partial JSON receipts; none proves a
root-package pass. The later root gate on `13409b9` passed in session `39091`:
1,635 passing entries, 50 skips, zero failures and 5,569.142s. It includes
the previously failing distributed checkpoint duplicate case and both
Github971 methods.
The user's separate plain-text root run failed only
`TestJavaGithub971a` and `TestJavaGithub971b` on an earlier compiled snapshot;
both pass 100 repetitions each on the current source.
Gate `85707` on `34f4161`
failed with three native checkpoint reader errors and a debugger panic. The hover
lookup panic was corrected in `34a87ab`; the unchanged original debugger test
passes. Logs are under `.codex-gotmp/full-port-after-writer-fixes/`.

The next normal offline gate on `34a87ab`, session `74633`, is terminal exit 1:
root failed in 4,509.998s on one leaf,
`TestNativeDistributedRemoteCheckpointCorruptLaterMemSnapshot/duplicate`;
TLC passed in 1,085.483s and SANY in 2.439s. All other reported tests passed.
Its exact command, tracked input hashes and JSON log are under
`.codex-gotmp/full-port-after-filesystem-hover/`. The earlier three checkpoint
leaves passed in this run, but their prior failures are not dismissed. The
original observer stopped after the gate; its captures are retained under
`.codex-gotmp/checkpoint-filesystem-observe/`.

A single focused execution of the new failing leaf with an ignored reader-error
overlay reproduced the failure and identified `EOFException` on pool file `1`.
A separate external observer captured that file at checkpoint promotion and
close: both copies contain only 7,041 of 8,192 expected states (360,448 bytes).
Pool file `0` contains all 8,192 states (419,309 bytes); queue metadata decodes
with 12,288 queued states. Artifact hashes, complete decoder rows and zero
observer errors are in `.codex-gotmp/checkpoint-filesystem-observe-duplicate-later/`.
The observer is stopped; the artifacts are retained. These observations prove
an incomplete pool file in the focused failing run. A bounded external probe
against the frozen Java TLC jar now confirms the same upstream crash boundary:
after 16,385 valid states, immediate `Runtime.halt` after checkpoint publication
leaves pool `1` at 2,547,712 of 2,670,592 bytes and recovery fails with
`EOFException`; a one-second delayed control writes the full file and recovers
all states. The probe and hashes are under `.codex-gotmp/java-checkpoint-crash/`.
Do not weaken assertions or change the writer lifecycle without the applicable
authorization. The separate checkpoint error-formatting hang remains outside
the authorized flusher fix.

The user explicitly approved the source-divergent `BeginChkpt` writer-drain
correction on October 16, 2026. It is now installed: `BeginChkpt` holds the
queue mutex and waits for the pending pool writer after draining the cleaner,
before publishing queue metadata. The unchanged duplicate-later memory case,
both full-frontier interruption cases and the related LSB commit-reply-loss
case pass with the installed correction. The retained preparation and original
failure evidence remain under `.codex-gotmp/checkpoint-writer-drain-prepared/`
and `.codex-gotmp/checkpoint-filesystem-observe-duplicate-later/`.

The later final-component ownership correction exposed a native harness defect:
the mid-run checkpoint fixture replaced the public queue and fingerprint-manager
mirrors after their owners had already been captured. Application-backed server
construction now has a native configuration callback that installs adapters
before ownership capture, and the fixture uses it for the queue and fingerprint
commit failpoints. The unchanged duplicate-later case now checkpoints a real
20,480-distinct/12,288-queued frontier and fails in `StatePoolReader` without
the writer patch. With the isolated overlay it reaches the intended later-host
fingerprint corruption and passes. The after-intern and after-first-fingerprint
interruption cases also reach their exact markers without the overlay, then fail
on the same incomplete pool; both pass with the overlay at original bounds.
These observations supplied the evidence for the now-installed correction.

After gate `74633` ended, three source-backed corrections were installed without
replacing stale files. `Tool.Lookup` retains the original declaration for an
unbound state; Action constructors preserve explicit empty names, labels and
parameterized missing-context failures; predicate validation preserves supplied
contexts and detailed non-boolean diagnostics. All source/prepared SHA256s matched
before promotion, and the installed file hashes match the prepared copies. The
combined focused original/model/native selection passes: root 11.006s, TLC
0.016s, session `99261`. The source observations remain under
`.codex-gotmp/tool-lookup-unbound/`, `.codex-gotmp/action-name-overloads/` and
`.codex-gotmp/validation-context-details/`; no original-test inventory credit
was added. There is no full passing gate on this newest production snapshot.

The validation comparison used 36 records from actual parsed EWD840 MC06
expressions, initial state and contexts. Baseline Go differed on 20; all match
the installed correction. The debugger's assumption dispatch is now separate:
ordinary false expression validation leaves its frame stack intact, while a
false assumption adds one violation frame. Seven actual parsed `NoNext == FALSE`
source/native values and stack counts match; related original checks pass in
10.363s. Arbitrary reset scheduling remains unproven.

Fairness liveness leaves now retain Java's action/subscript labels. Parsed
`ENABLED` syntax and junction child traversal also match Java's DOT path;
focused original models and existing liveness checks pass. Details and the
ignored parsed-model observer are in the progress log. The full gate has not
been rerun since the latest source-backed corrections.

Liveness state-enabled, action and ordinary state nodes now preserve Java's
null-tool failure boundary while retaining the boxed/subscripted enabled early
return and custom evaluator ownership. Focused original liveness models pass;
the ignored parsed-model observer is documented in the progress log. Folded
state/action leaves also expose their retained context and mutable source tag.
Parsed state-AST leaves expose their retained semantic expression separately
from the temporal-child `GetBody` accessor.

`LiveExprNext` now recursively flattens singleton junctions and binaryizes
multiway junctions before tableau construction, matching `LNNext`. Focused
original tableau/liveness models pass; see the progress log for the receipt.

`LiveExprNext` direct evaluation now shifts to the successor state, and its
lasso entry preserves Java's enabled-assertion failure because next must have
been rewritten first. An actual parsed-tableau observer and original liveness
models pass; details are in the progress log.

Lasso evaluation no longer normalizes empty lists or invalid positions. Boolean
and temporal nodes retain Java's empty-loop results, while state/action access
uses typed `IndexOutOfBoundsException`. Parsed observers and original liveness
models pass; see the progress log.

Counterexample property attribution now reconstructs each implied temporal as
Java does, preserves trace positions, enforces source lasso assertions, reports
conversion/evaluation failures without a synthetic fallback, and includes the
stuttering-cycle overload.

`LiveExprNext` now has Java's action level and derives `containAction` from its
body. An actual generated tableau plus original liveness/generated-trace models
pass; details are in the progress log.

`LiveExprNext.IsGeneralTF` now inherits Java's unconditional base result instead
of inspecting its body. Ordinary parsed state predicates also use their syntax
images for DOT output, as Java `LNStateAST` does. Next-node DOT output now
recurses through that specialized body rendering, matching `LNNext`. The parsed
observer and focused translated liveness selections pass; details are in the
progress log.

Action-node binaryization now raises Java's `WrongInvocationException`, and
next-node tagging preserves Java's inherited no-op. The actual parsed observer
and focused translated liveness selections pass.

Direct conjunction/disjunction constructors now preserve nested junctions;
only `AddConj`/`AddDisj` flatten them, matching Java's constructor/add split.
Binaryization still uses the flattening add path. Focused translated and
original liveness models pass.

The bounded-exists liveness regression now uses a valid constant semantic body
instead of an impossible unknown-kind operator node. The complete AST and
broader liveness selectors pass; this was a synthetic-fixture correction, not a
production lookup change.

Bounded-forall liveness conversion now preserves Java's coded failure and the
message from a caught `TLCRuntimeException`. A parsed model with a
non-enumerable temporal quantifier bound prints the same formula and bound
diagnostic as Java. Successful bounded-exists and bounded-forall expansion also
uses the expanded live expression's level, matching the source fallback.
Conjunction, disjunction, conditional, negation and implication now preserve
that actual translated level too; function application uses Java's static
expression level after its speculative lambda expansion.
Constant liveness subformulas also propagate evaluation failures and reject
non-boolean values at tableau construction with `TLC_EXPECTED_VALUE`, matching
Java instead of deferring the failure to state exploration.
Multi-bin liveness debug records now report Java's total DNF possible-error-model
count for every solution rather than the size of each grouped solution bin.
Unsupported temporal nesting around actions now raises Java's coded
`TLC_LIVE_WRONG_FORMULA_FORMAT` startup error instead of an unexpected legacy
evaluation exception.
Stuttering counterexample warnings now identify the complete `Spec` operator
declaration span, matching Java's `OpDefNode.getLocation()` output.
Liveness state/action type failures, direct temporal evaluation failures and
tableau action rejection now retain Java's `TLCRuntimeException` family rather
than being misclassified as value `EvalException`s.

The optional tableau DOT export now runs at Java's `processLiveness` boundary,
after binaryization, rather than from every `NewTBGraph` call. Write failures
print a stack trace to stderr as source does. The actual parsed observer and
focused original liveness models pass.

Java's missing `TBGraph.toString` and recursive `TBGraphNode.toString` renderers
are now ported with source node/successor ordering and cycle suppression. The
actual parsed tableau observer and focused translated checks pass. Both
renderers reject nil receivers with the translated Java exception.

`TBGraph` and `TBGraphNode` also expose the source initial-count, node-index and
particle accessors. Negative node indices preserve Java's enabled assertion.
Initial counts, assigned node indices and graph/successor lookup indices narrow
to Java `int`; nil node receivers fail before the assignment assertion.
Tableau expansion triples expose their three expressions, alpha/beta
classification and exact multiline source rendering, including null components.

`GraphNode.GetTransitions` now returns set semantics like Java
`getTransition`, collapsing duplicate fingerprint/tableau/action triples. Its
no-argument form and transition value accessors are also available. DOT
rendering includes the source overload that filters outgoing tableau edges by
a `TableauNodePtrTable`. Graph-node equality and hashing use only the state
fingerprint and tableau index with Java's 32-bit hash arithmetic. Transition
values also expose Java's check-vector/fingerprint/tableau equality and hash.
The shared abstract graph-node constructor retains a supplied null check vector;
typed null failures occur only when a source method actually dereferences it.
`BTGraphNode.SetIndex` also preserves Java's raw 32-bit OR semantics, allowing
the supplied index's high bits to set the done and dummy flags.
Live-checker solution numbers, disk-graph solution filenames/metadata and
checker lookup indices narrow to Java `int` at their public boundaries.

`PossibleErrorModel` and `OrderOfSolution` now retain the arrays supplied to
their constructors and setters, matching Java ownership. Actual parsed fairness
solutions, focused translated checks and original liveness models pass. The
solution's tableau, promise, state-check, action-check and PEM getters expose
those same retained values. The PEM accessor uses Java's exact `GetPems`
spelling; `GetPEMs` remains a compatibility delegate.
Its public state-check and legacy action-check evaluators use the source
`CheckState` and `CheckAction` names; the suffixed Go spellings remain delegates,
and the bit-vector overload remains distinct because Go has no overloading.

`OrderOfSolution` also requires Java's non-null promise array. Production
no-tableau solutions supply a distinct zero-length slice, while tableau
solutions retain the non-null slice returned by promise extraction.
`LiveExprNode` also exposes the caller-supplied `TBPar` promise-extraction
overload. Its convenience form returns a distinct non-null zero-length slice
when no promises exist.

`TBPar` and `TBParVec` now retain their inherited Java `Vect` construction,
indexing, empty-removal and null-dereference boundaries. Empty membership tests
still preserve Java's short-circuit result without dereferencing a null query.

The missing `OrderOfSolution.printPromises` and `Liveness.printTBGraph`
diagnostics are ported. They preserve one-line-per-promise output, ToolIO
ownership, the null-tableau text and PrintStream-style ignored write failures.

Stuttering checks now raise typed array bounds failures for invalid AE-state
indices, and the action bit-vector overload preserves a nil result until a true
action requires `set`, matching Java. Parsed and focused checks pass.
NodeTable exposes source `GetBTNodeWithHint`; the prior pluralized Go spelling
remains a delegate to the same direct-slot and fallback lookup.

`PossibleErrorModel` now preserves Java's null-array access order across empty,
stuttering and formatting methods. Formatting also raises typed array bounds
and null-expression failures instead of native slice panics or `<nil>` text.

Legacy tableau behavior graphs now retain Java's polymorphic `BTGraphNode`
rendering, including tableau indices throughout recursive output. Calling the
base behavior node's tableau accessor raises the source
`WrongInvocationException`. BE/BT equality and hashing retain the source class
split: BE uses the fingerprint, while BT also uses the masked tableau index.
The actual parsed graph observer and focused translated checks pass.

Legacy `LiveCheck1` graph-statistics methods now return fresh dummy statistics
as Java does, ignoring a supplied real accumulator, and its missing indexed
checker accessor returns null.
Its public single-transition overload is also available as `AddNextStatePair`,
retaining the same checker monitor and graph mutation path as set insertion.

Modern `LiveCheck` exposes Java's metadata-directory and indexed-checker
accessors. The folded `NoOpLiveCheck` also retains its source tool accessor;
its checker accessor returns null for every index. Returned `LiveChecker`
instances expose their owned solution and concrete plain/tableau disk graph.
`LiveCheck.BeginChkpt` shares the object monitor used by synchronized
add-and-check insertion, while the folded no-op override remains lock-free.
Modern and legacy liveness containers expose source `GetNumChecker`; checker
startup guards use it, while the prior `NumChecker` name remains a delegate.

Installed corrections since that baseline:

- `74a785c`: DFID retains worker throwables, typed stack bounds and the checker's
  accepted error code; simulation retains supplied queues and typed null write
  failures; state-aware symbol lookup preserves missing-state failures while
  operator arguments use symbol-only lookup.
- `2b68be8`: the full worker/statistics constructors retain supplied counters,
  including nil; `NewSimulationWorkerWithDefaultStatistics` implements the
  source short constructor with fresh counters.
- `81b923d` and `24d9648`: trace-action modes and trace-file prefixes are nullable
  through CLI options, simulator and worker APIs. Only nil disables collection
  or output; an empty filename prefix writes `_0_0`.
- `c81692d`: lifecycle/getter calls reject nil receivers; negative join timeouts
  raise the source argument exception before checking startup.
- Simulator and SimulationWorker entry points now reject nil receivers before
  defaults or error conversion. Their result queue applies the same typed
  receiver boundary to insertion, wakeup, removal and inspection paths.
- `TLCStandardMBean` rejects nil receivers across version, revision,
  registration and unregistration. Its source null-object subclass remains a
  true no-op for both lifecycle operations. `ModelCheckerMXWrapper` also
  enforces receiver and owned checker/TLC boundaries across all declared
  metrics and controls. `DiskFPSetMXWrapper` likewise requires both its receiver
  and owned fingerprint set for every metric and control.
  `BucketStatisticsMXWrapper` enforces the same receiver and statistics-owner
  boundary for object identity and all aggregate metrics. Its metrics owner is
  structural, so every valid `IBucketStatistics`-shaped implementation receives
  Java's virtual dispatch instead of falling through a closed concrete-type
  switch. All declared
  `TLCServerMXWrapper` methods now enforce receiver dispatch, including its
  server-independent checkpoint, runtime-ratio and liveness no-op bodies.
- `TeeOutputStream` enforces receiver, byte-array and owned-stream null
  boundaries while retaining main-before-branch write/flush order and Java's
  branch-closing `finally` precedence. Its three write forms share the source
  object monitor; flush and close remain outside that monitor.
- `DelayedPrintStream` rejects a null byte array while retaining buffered
  writes, swallowed PrintStream I/O errors and reset-after-release behavior.

Focused original DFID, worker, simulator, FCoverage and Github362 checks pass.
Bounded Java observations also match, including actual file bytes and worker
state/counter/RNG outcomes. Detailed receipts and scope limits belong in
[PORT_PROGRESS.md](PORT_PROGRESS.md) and [TLC_ARCH.md](TLC_ARCH.md). No inventory
credit was added. Arbitrary simultaneous lifecycle mutation and concurrent
reporting remain unproven. The old combined overlay under
`.codex-gotmp/prepared-core-parity/` is historical evidence; its pre-nullability
API copies must not replace the current production files.

All 15 earlier model failures are resolved by declaration-name lookup fixes.
The completed recovery matrix `13856` passed 13 top-level methods/20 subtests
in 562.455 seconds. Checkpoint/gzip harness corrections and the original
assertions remain intact. TSnapShot's earlier zero-queue failure is retained as
historical timing-sensitive evidence; five focused repetitions and the full
baseline passed without weakening it.

Recovery flags now retain nil versus empty through runner options and checker
constructors. An explicit empty DFID flag attempts recovery from its initialized
store instead of skipping it. Six actual missing/saved-store observations and
four CLI observations match Java. `-recover` preserves literal path spelling
and appends a separator, including for empty paths. Focused original DFID and
checkpoint regressions pass. BFS recovery now retains the initialized trace,
queue and fingerprint store paths and dispatches to the selected fingerprint
backend. `FPSetTrace` accepts standalone and concurrent enumeration; memory sets
recover their snapshots, while disk and multi sets enumerate the actual trace.
All 24 real-checkpoint comparison rows across six backends match Java for result,
store ownership, counts and console output. Missing memory snapshots retain
`FileNotFoundException`. Concurrent cursor failures now return their checked I/O
error through the shared enumerator API rather than panicking. Memory snapshot
and buffered-stream file opens now reject directories before reading, retain
`FileNotFoundException` details and leave existing stream state intact on open
failure. Twelve real-checkpoint and 15 direct stream rows match Java. Focused
original model/storage checks pass; checkpoint atomicity and exhaustive
failure-phase parity remain unproven.

`AbstractChecker` now exposes Java's base lifecycle/query defaults: progress is
`-1`, suspend/resume raise their exact unsupported-operation failures, and
recovery status reflects whether a checkpoint argument was supplied. Both trace
overloads raise their source failures, and queue/generated/distinct/initial
counters return `-1`; base statistics and config use those inherited values.
DFID inherits these defaults while concrete BFS behavior overrides them.
The concrete BFS statistics getter delegates to the shared final-method model so
virtual metric calls retain `ModelChecker` values. Its config getter returns the
constructor-captured record directly. Both getters require a receiver, matching
the inherited Java methods. BFS trace-info getters likewise require their
receiver, destination state and trace owner before delegating; an initial
destination still permits a null `from` state in the two-state source overload.
Concrete BFS stop, suspend and resume require the checker base and state queue;
successful queue control still wakes the coordinator under the checker lock.
Progress requires a trace owner and converts only checked trace I/O failure to
`-1`, as Java does.
Concrete generated, initial, queue-size and distinct-state queries require their
checker base and constructor-owned storage; generated totals dereference every
worker slot instead of skipping a missing worker.

BFS checker cleanup now follows the source's phase boundaries. Failed recovery,
failed assumptions, liveness tautology and a non-exception initialization stop
retain initialized resources. Initialization exceptions and exploration still
run cleanup. Six actual checker observations verify base trace ownership, return
codes and state counts; focused originals pass. Java worker descriptors have a
separate lifetime from its base trace, so worker enumeration alone cannot prove
whether checker cleanup ran. Native owned worker descriptors still close during
actual cleanup. Worker enumeration snapshots the writer cursor and opens its
separate reader without acquiring the worker monitor, matching the unsynchronized
Java `elements()` method. `BufferedRandomAccessFile` remains unsynchronized at
the object level, as its source documentation specifies; only its static buffer
pool is synchronized. Worker record reads enforce the source assertions after
each decoded field: nonnegative pointer, worker-array bounds and fingerprint-set
membership, before restoring the writer cursor. Concurrent reconstruction then
asserts that the collected chain fits the workers' reported level and, when no
starting state is supplied, that its final record is initial. Both public
state-to-trace overloads also preserve Java `ArrayList`'s level-derived negative
capacity failure before accessing a worker record, including signed-int overflow.
Reconstructed non-null state-info wrappers must carry a state before predecessor
metadata assignment, yielding typed `NullPointerException` like Java dispatch.
`LiveCheck.check0` prints its start event and then rejects an empty checker array
with `IllegalArgumentException` before resetting the shared liveness-error owner,
matching construction of Java's zero-capacity `ArrayBlockingQueue`.
After all liveness workers complete, multiple failures retain the last failure
in completion order like Java's overwritten `ExecutionException`; any reported
liveness violation still takes precedence over all worker failures.
The worker pool joins every goroutine before consuming result records, and each
worker boundary captures panics as the Future cause that `check0` diagnoses.
LiveChecker graph-size and last-check-size reads require both the checker and its
ordinary or tableau disk graph, preserving `check0` and `doLiveCheck`
dereference failures instead of returning native zero/in-memory fallbacks.
Its begin-checkpoint, commit, flush and recovery adapters enforce the same owner
requirements before delegating; missing owners cannot turn lifecycle work into
successful no-ops.
Close requires the graph unless cleanup is vetoed and always requires the state
writer after graph close. Reset and both disk-graph degree calculations require
every checker and graph rather than skipping absent owners.
The successful non-final `check0` pointer-table restore also requires every
checker and graph and stops before its completion message on the first failure.
`LiveChecker.CheckSccs` requires its receiver and disk graph, creates the graph
cache, and only then requires its solution; a missing solution retains the cache
like Java's failed `LiveWorker.call` setup.
`LiveWorker.CheckSccs` likewise requires its checker before rebuilding graph
pointers, then requires its PEM and solution in source order instead of treating
missing SCC owners as a successful empty check.
Modern `LiveChecker.AddInitState` requires its solution and ordinary or tableau
disk graph; it mutates the graph before invoking the required liveness writer,
so writer failure retains the recorded root. The obsolete in-memory fallback is
removed.
Ordinary `AddNextState` reads the successor set before requiring its solution,
then acquires the solution monitor and requires the disk graph; its obsolete
in-memory successor path is likewise removed.
Tableau `AddNextState` computes the successor/tableau consistency matrix before
acquiring the solution monitor, then requires the tableau disk graph inside the
monitor. Its obsolete in-memory successor path is removed as well.
Recursive expansion for a new tableau node on an already-completed state
requires its solution, tableau node and disk graph. Solution state checks still
run before the node and graph are dereferenced, matching Java's helper order.
The now-unreachable `LiveChecker.Graph`, `Initial` and in-memory node-construction
helpers are removed; modern liveness graph ownership is exclusively disk-backed
as in Java.
Tableau successor insertion now applies Java's `GENERAL` assertion when an
existing node gains no transition while a main checker is present; only
simulation may legitimately take that unchanged-node branch.
`LiveCheck1.AddNextState` leaves the set-level method unsynchronized and enters
its synchronized pair overload once per successor, allowing Java's monitor
interleaving between successor pairs.
Worker liveness call-stack replay holds the shared checker monitor from the
winning `printedLivenessErrorStack` claim through iterator reset, replay and
nested-expression output. Competing worker error reports therefore wait until
the original liveness failure is ready to propagate, matching Java's enclosing
`synchronized (this.tlc)` block.
Safety-like liveness trace reconstruction uses that same checker monitor through
state regeneration, output, checker stop/error publication and postcondition
evaluation. A reporter that finds the shared flag already set returns normally
before touching its Tool or prefix, as Java's `LiveCheck.printErrorTrace` does.
Its reconstruction preserves direct nil/empty owner failures, uses the source
`EvalException` for a missing initial state, dereferences a missing successor,
and retains null ALIAS results in the trace. The detected error graph node is
cleared only after successful postcondition handling; earlier failures and a
losing reporter leave it installed.
Fingerprint-prefix recovery likewise destroys its temporary disk-graph cache
only after `GetPath` succeeds. A checked path failure retains both the cache and
error node because Java's cache bracket has no `finally` cleanup.
The enclosing SCC worker lifecycle follows the same rule: successful PEM checks
destroy the graph cache before recording graph size, while any checked SCC setup
or traversal failure retains the cache and skips size recording.
SCC component checking now asserts the source five-slot Tarjan record shape
before popping a location, tableau index and fingerprint from its component stack.
Each nontrivial extracted component record must also carry a disk-graph file
pointer before its graph link is retired.
SCC root initialization requires each nonnegative initial link to remain a file
pointer and rejects tableau `UNDONE` or `DONE` sentinels during a final check.
DFS traversal preserves the remaining source assertions for stack locations,
active and previously assigned link ranges, final-check `UNDONE` successors,
component-table bounds, complete component-stack draining and seen self-loops.
After successful cleanup and size recording, SCC checking evaluates the selected
disk graph's invariants with the solution's state/action check widths. Checked
storage failures propagate; a false invariant raises `AssertionError`.
Lasso prefix/cycle reconstruction reuses this enclosing SCC cache. It neither
replaces nor destroys the cache itself, including when path or cycle recovery
fails; the enclosing normal-return lifecycle remains its sole owner.
Prefix path lookup, state regeneration and prefix ALIAS evaluation run in a
dedicated goroutine concurrently with cycle construction, matching Java's
single-thread executor. Joining unwraps an `EvalException` cause and returns all
other task failures as `ExecutionException`; cycle failures remain direct.
For an SCC counterexample, rendering completes before the Tool postcondition is
evaluated. The postcondition remains inside the enclosing cache/report lifecycle,
and a missing Tool fails only after the trace has been emitted.
The lasso loop ordinal narrows `TLCStateInfo.StateNumber` through signed 32-bit
Java `int` conversion before marker output and `CounterExample` construction.
Liveness body indices, tags, tag traversal and lasso cycle/current positions
also use Java-width arguments; tag and next-position addition preserve signed
overflow before returning or indexing the state list.
For a non-stuttering lasso without VIEW, closing-state regeneration must equal
the cycle state before the back-to-state marker is accepted. VIEW preserves the
source short circuit around this assertion for Github issue 1045.

Built-in state dumps, DOT and action-flow output preserve `FileUtil.newBFOS`
constructor diagnostics and `PrintWriter` suppression of output/close I/O errors.
Real base and DOT state writes retain their Java monitors; no-op writes,
explicit snapshots and close remain outside the monitor like their source methods.
Actual `/dev/full` runs retain the same checker result and state counts as Java,
and failed flushes release the owned output descriptor. Regular dump/DOT files
remain byte-identical. Custom writer callbacks and explicit snapshot file
operations retain their separate error behavior; this does not settle arbitrary
custom-writer cleanup failure ordering. Five real-action renderer observations
also match constructor failures, full graph bytes and closed-writer behavior.
Specialized liveness DOT output now uses signed fingerprints and literal label
newlines. Graphs without a tableau delegate to the ordinary DOT renderer, matching
Java's inherited overload without action-check labels. Three complete graph files
from actual `BidirectionalTransitions1Bx` states, generated tableau nodes and
computed action checks match Java byte for byte. This verifies rendering, not a
new original test completion. The auto-worker trace-prefix failure is an
intermittent unchanged-Java scheduling limitation. The October 19 harness
correction retains exact prefix states and ordinals but validates a final
constrained-out safety successor against the replay tool's transition relation
and the same violated invariant. Worker settings and production invariant checks
are preserved. This correction has source review only; no local tests were run.
Plain and tableau liveness disk graphs now expose Java's debug `writeDotViz`
entry points, including cache creation/destruction, concrete renderer dispatch
and stderr stack traces for write failures. `LiveWorker` also renders filtered
SCC tables, skips debug files for trivial components, and exposes Java's
component/DFS detailed stack formatters.
`TableauNodePtrTable` exposes its source detailed formatter for bucket, pointer,
link, predecessor and seen-state inspection.
DiskGraph's public location-based node lookup and no-tableau pointer lookup are
also exposed. The location overload preserves Java's required cache and typed
negative-pointer failures; the complete original disk-graph tests use it.
`DiskGraph.BeginChkpt` now shares the same graph mutex as Java's synchronized
cached and direct node reads; commit and recovery remain unsynchronized.
DOT snapshots now use the existing replacement-copy path before appending the
closing trailer. Six real-model output profiles match Java, including destination
symlinks, empty directories and same-file/hard-link destinations. Seven further real-file observations match copy/append failures and retained
ownership: missing sources/parents, non-directory parents, nonempty destinations
and directory sources. Linux failures retain the source filesystem exception
classes and path details; `DirectoryNotEmptyException` remains a checked I/O error.
Explicit snapshots retain checked I/O failures. Automatic snapshots after state
or transition writes print the cause's native stack and wrap it in a runtime
exception, as Java does. Four actual writer/checker observations match return
values, causes, console output, graph bytes and state counts; stderr exception
headers match while stack frames describe Go. Unobserved transfer failures,
races and other filesystem providers remain unproven.
The shared `FileUtil.replaceFile` correction is now installed. Linux replacement
moves check source/destination identities, replace destination entries, copy
across filesystems with metadata, and roll back a new destination if source
removal fails. Other platforms retain the previous rename fallback. Copy and
snapshot operations normalize separator spelling, validate paths and retain
public filenames and exception argument spelling. Empty paths use the provider
system path. Copy creation preserves special mode bits for the native filesystem
to apply its creation/write rules. All 73 actual Java/Go file, link, directory,
permission, metadata, path and graph-byte comparison records match against the
installed implementation. Special files, non-Linux source equivalence and
filesystem races remain unproven. The retained preparation patch, hash manifest
and observer receipts are under `.codex-gotmp/fileutil-move-boundary/`.

The explicit-maximum infinite-precision indexer constructor is installed in
[offheap_indexer.go](offheap_indexer.go); all 120 index/failure observations
agree with unchanged Java. All four original supplementary JPF workloads now
have native bounded/exhaustive equivalents behind `tlc_verify`, with completed
original Java source comparisons. Current scopes, receipts and the full tagged
TLC gate are recorded below. The earlier missing-Unsafe startup failure was
resolved by using the installed Java 11 runtime with its existing configuration;
no source/runtime emulation was introduced.

The queue production corrections retain Java's monitor ordering: `FinishAll`
signals suspension while holding the queue monitor, then disk subclasses release
that monitor before stopping pool workers. `ResumeAllStuck` separately rechecks
the non-stopped, nonempty and waiting conditions after its initial observation.
Both native pool readers check `finished` after each condition wait. State and
byte-array pool buffer sizes narrow to Java `int` before allocation checks.
The disk verification fixture now uses the original raw-long stream models
through actual queue/pool workers. Its appended `INIT_UID=-1` records therefore
follow the source model's signed representation; ordinary production streams
retain their natural-number encoding. Earlier real-byte-file fixture failures
are retained only as the explicit diagnostic described below. They are not an
unresolved production scheduler defect. Detailed run history belongs in
`PORT_PROGRESS.md`.

The existing `MCDiskStateQueueWorkload` specification and configuration also
pass unchanged in Java and the native CLI: 44,886 generated states, 14,805
distinct states, depth 66 and identical outdegree/success messages. Its five
invariants, locality property and original load constraint remain enabled.
This verifies checker behavior on the bounded source model, with no JPF or
production-thread verification credit. Receipts are in the progress log.

The full native MSB batched random method completed with exit 1 in 817.52
seconds: all 48 producers finished with zero collisions, and 2,148,818,944 puts
matched final size. The original upper-size bound of 2,147,483,697 then failed;
the final invariant assertion was not reached. Original bounds, seeds, memory
settings and assertions were intact, on CPUs 0-7 without race instrumentation.
Receipt: `.codex-gotmp/concurrent-msb-batched-full-20261017/`. This retains
source-failure reconciliation and earns no passing credit. Do not repeat the
unchanged workload for credit; no Java or interrupted sequential job was restarted.

The full native off-heap block-random method deadlocked after exceeding its
insertion target. Session `59740` ended with exit 2 after diagnostic SIGQUIT;
this is not a natural assertion result or passing credit. Live debugger evidence
showed 13 producer goroutines waiting at an eviction barrier registered for 48
parties (phase 287), with eviction pending and no producer errors in the log.
The stable set counters were 2,141,323,601 disk entries and 7,540,852 table
entries: 2,148,864,453 total, above the original 2,147,483,649 target. The other
35 producers had exited, leaving too few parties to advance the barrier. Java
uses the same fixed-party phaser and its generators do not deregister on exit;
this identifies a source-test lifecycle conflict, not a completed Java comparison.
Receipts: `.codex-gotmp/concurrent-offheap-block-full-20261017/`, including
`gdb-barrier-counts.txt`, `gdb-size.txt` and terminal stacks in `run.log`.
Original bounds, 48 workers, seeds, memory and assertions were preserved on
CPUs 0-7 without race instrumentation. Do not repeat this unchanged workload
for passing credit. No Java or interrupted sequential workload was restarted.
The prepared LSB native-width count correction remains isolated in
`.codex-gotmp/lsb-wide-count-correction-20261017/`, pending the user's divergence
decision; production retains Java's count behavior.

Original fingerprint stress translations are now installed behind `tlc_fp_stress`:
[long_sequential_fpset_stress_java_test.go](long_sequential_fpset_stress_java_test.go)
contains the three complete sequential bodies, and
[concurrent_fpset_stress_java_test.go](concurrent_fpset_stress_java_test.go) contains
all four generators and twelve concrete concurrent contexts. The tagged suite
compiles; nine random bodies agree with Java using existing one-worker/20,480-
insertion properties. Default off-heap partitioning agrees for all 48 producer
counts and retains the same source minimum-size failure. Sequential source
contradictions and default random-workload verification remain pending; no new
passing/inventory credit. Sequential sessions `69494`/`85088` stopped after
the power outage and remain uncredited; no job was restarted. Default build/test
file and import lists are unchanged by these additions; the completed full gate
retained its original inputs. Detailed source comparisons are in the progress log.
The full-bound native LSB scalar-random method now has an observed source-shaped
failure: its 2,147,483,648-entry table maximum overflows the LSB flusher's signed
`int` count. At 2,147,483,778 entries a producer fails and the original latch
cannot finish. Short Java/Go preparation probes match the count conversion and
failure. The owned run was terminated after stack/field capture; no method
credit or production correction was added. See [the LSB count report](../JAVA_LSB_COUNT_BUG.md)
and `PORT_PROGRESS.md`; do not repeat this unchanged failing workload for credit.
The full-bound native MSB scalar-random method completes all 48 producers and
validates its first large flush, then fails the original final size upper bound:
2,149,182,743 actual versus 2,147,483,697 allowed. This is consistent with the
source generator's previously reproduced quota/size overshoot; the final
invariant assertion is not reached and no passing credit is added. Failed
producers now report their exception class and captured stack even for empty
messages, retaining the original failed-producer latch behavior.
The full-bound native MSB block-random method also finishes with all producers
and batch-membership checks intact, then fails the same upper bound at
2,148,993,024 entries. Its zero-collision batch quotas alone require at least
2,147,500,032 puts, above the allowed 2,147,483,697. Both MSB runs are terminal;
retain their original assertions and do not repeat them for passing credit.
The three remaining ConcurrentWrite bodies are installed in
[concurrent_write_source_failures_java_test.go](concurrent_write_source_failures_java_test.go)
behind the same tag. Full four-million-value writes in `test1`/`test2` retain
the same Java/Go failure at read 1,000,000.
The complete 400-million-value `test4` body has now finished normally without
race instrumentation: session `5717`, exit 1, 684.419 seconds. It fails at read
50,000,000 with actual zero, matching the earlier full Java failure. Log:
`.codex-gotmp/concurrent-write-test4-full-native.log`. Keep the original bounds
and assertions, retain reconciliation status, and do not repeat this completed
workload for a passing claim.
The full default off-heap batched-random native method terminated in session
`57353` with exit 137, verified after the benchmarking pause. Java session `16867`
also terminated with exit 137 after the user killed its process; do not restart
that Java workload without explicit authorization. Neither run earns completion
credit. Neither was restarted. Both
startup logs confirm 48 workers and 2,147,483,649 insertions, with original
bounds and no race. Logs: `.codex-gotmp/fpset-concurrent-full/{java,native}-full.log`.
The native binary predates the counter-constructor correction, which does not
change fingerprint code. Exact commands are in that folder's `commands.json`.
These reconciliation cases earn no passing credit. Source-failing
`DistributedTrace` and `AssertExpressionStack` bodies and exact fixtures are also
installed in [tlc_source_reconciliation_java_test.go](../tlc_source_reconciliation_java_test.go),
behind `tlc_reconcile`. Java and Go retain the original trace-generation and
SUCCESS-exit contradictions. The latter also retains the source helper’s
conflict with its generic action label; neither earns passing credit. The same
tagged file also installs `LivenessSymmetryWarning`: its three method assertions
pass in Java and Go, while inherited SUCCESS fails against actual exit 13.
The original assertion remains intact and uncredited. `DepthFirstTerminate`
is installed there too: its original processor-derived worker count and depth 50
produce matching Java/Go GENERAL and exit 255 failures on this 48-core host.
The test remains uncredited; do not substitute one worker.

The earlier sandboxed gate at `acfd10b` failed. All 15 model failures now pass after
the declaration-name correction. Focused existing RPC/process-recovery/tracing
checks and the capture-only ExecutionStatisticsCollector tests also pass with the
required permissions. A direct native observation confirms sandboxed interface
enumeration fails with netlink operation-not-permitted; permitted enumeration
succeeds. Source-defined random-ID fallback explains the two sandbox telemetry
suffix failures. No assertion, skip or implementation was changed to accommodate
the sandbox. The collector's identifier, file-backed identifier, synchronous
collection and protected submission boundaries now raise typed
`NullPointerException` on nil receivers. Its source `Selection` values and
default constructor path/domain inputs are exposed and captured once at lazy
class initialization like Java's static finals. Both default and injected
construction trigger that capture; the translated original
preference-file setup uses the same enum rendering. Preference existence,
enabled-state lookup, default-path writes, selection recovery and the static
prompt decision now preserve the source file and DNS ordering. Public collection
always runs asynchronously and captures worker panics like `CompletableFuture`;
the wait property registers a ten-second process-exit wait used by normal and
distributed-signal shutdown. BFS, DFS and simulation startup invoke the collector
immediately after printing their mode banner and reuse the same runtime snapshot;
telemetry removes the four source-only display/runtime keys and preserves
version, mode, environment, storage, toolbox and IDE ordering. REPL startup
reports its smaller source environment payload before evaluation, and the
standalone collector entry invokes synchronous collection. Detailed receipts
belong in the progress log. The full port remains active.

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
JVM-only GC assertions are not native Go implementation requirements. The JPF
short-stream property now has a complete native exhaustive equivalent in
[buffered_data_stream_verify_java_test.go](buffered_data_stream_verify_java_test.go),
behind `tlc_verify`: all 300 original length pairs pass and agree with Java.
All four supplementary verification methods now have native bounded/exhaustive
equivalents. The disk-queue method combines its 7,045-state source protocol
explorer with real Go queue/pool workers using the original raw-long stream
models and shared state read/write logic. A private captured factory selects
those streams only for the fixture; ordinary constructors retain byte files.
Original bounds, UIDs and writer/task joins are preserved. Fifty native runs
pass, including observed negative-UID prefetch, and all four methods pass
normally and with focused race instrumentation. The unchanged original Java
disk-queue test and pool models pass actual JPF in 1,570,688 states, without
errors or constraints. Receipt directories:
`.codex-gotmp/disk-queue-jpf-java11-20261017/` and
`.codex-gotmp/disk-queue-native-model-20261017/`.
The earlier real-byte-file fixture mismatch remains an explicit diagnostic via
`TLAGO_DISK_QUEUE_VERIFY_BYTES=1`; it earns no original-method credit and does
not change production encoding. Main-suite counts and remaining long/concurrent
source-failure reconciliations are unchanged. These proofs cover the original
bounded workloads, not arbitrary production scheduling or hardware memory models.
Memory-queue/off-heap source comparisons and deterministic short-stream evidence
are detailed in `PORT_PROGRESS.md` and `TODO_TEST_PORT.md`.
The full native `tlc` package regression with `tlc_verify` passed on source
`8d247d7`: 25,885 passing entries, 5,727 skips and zero failures in 562.91
seconds, with package terminal pass and session `61259` exit 0. This covers
all default TLC tests and all supplementary methods after the stream-boundary
refactoring. Original bounds/workers were retained, with no exclusions or
`-race`; all builds/tests/children shared CPUs 0-7 and `GOMAXPROCS=8`.
Receipts: `.codex-gotmp/full-tlc-stream-models-20261017/run.jsonl` and its
`commands.json`. Focused file/error-order and original binary trace-model
checks also pass, and all default packages compile. The newer native root baseline and snapshot qualifications are recorded below.

Detailed bridge, evaluator, simulator, DFID, value and override audits are
preserved in [PORT_PROGRESS.md](PORT_PROGRESS.md#october-11-2026-handoff-detail-consolidation).
Public `SpecProcessor` getters and its module-override and dynamic-extendee
processing entry points reject nil receivers with the typed Java
`NullPointerException`; valid getters return their backing objects directly.
The enclosing `Spec` root-name, model-config and processor getters enforce the
same receiver boundary while retaining their existing values for real specs.
`TLARegistry` now uses synchronized Java `Hashtable` hashing, resize and bucket
enumeration semantics. `AllNames` therefore matches the pinned standard-operator
ordering exactly, and replacement retains bucket position while returning the
prior mapping through the established Go nullable-result adaptation.
The source-equivalent public `TLCStackFrame` access, scope, matching and
evaluation surface enforces the same typed receiver boundary. Frame construction
uses detailed runtime exceptions as the authoritative expression/context and
enforces Java's required tool, node and context checks.
`TLCStateStackFrame` state, scope, variable, trace, breakpoint and evaluation
entry points likewise reject nil receivers, as do the corresponding
`TLCActionStackFrame` action-state overrides and `TLCInitStatesStackFrame`
initial-state collection and selection methods. Missing state references return
before functor access, and `TLCNextStatesStackFrame` exposes Java's `Successors`
scope while enforcing its receiver boundary. Synthetic trace frames enforce the
same boundary for current/successor access and state-pair evaluation.
Stopped events are built by the frame from Java detail messages, including
source marker-line removal and nullable-message failure behavior.
Normal and exception frame pops enforce source node identity; detailed failures
may match their carried expression, and empty or mismatched stacks fail.
The `TLCDebugger` administrative tool, granularity, capability, exception-filter
and configuration surface rejects nil receivers like Java instance dispatch.
`DebugEvalMode.String` exposes the exact `Const`, `State`, `Action` and
`Debugger` enum names in declaration order.
Breakpoint, lifecycle, stack, scope, variable and evaluation queries enforce
the same controller boundary, along with halt, continue, stepping, pause,
reverse and state-navigation commands.
The goto-state argument exposes its default constructor, boxed-integer getter
and fluent setter; it distinguishes unset, zero and nonzero references, retains
supplied boxes, accepts Java-style primitive autoboxing with `int` narrowing,
and rejects nil receivers.
`TLCCapabilities` likewise exposes its TLC-specific goto-state getter/setter and
retains Java's nullable capability state rather than collapsing unset to false.
All debugger frame lifecycle entry points require an existing controller, and
next/initial frame pops validate the source subclass before removal.
`TLCSourceBreakpoint` getters require a receiver; conditional evaluation now
requires its tool/context when Java does and swallows only source failure types.
Use [TLC_ARCH.md](TLC_ARCH.md) for implementation contracts and exact scope limits.
Those historical observations do not establish unobserved scheduling or general
parity. Do not repeat completed audits without a concrete source mismatch.

## Verification baseline and test credit

The native root regression gate on `91bd7b5` passed in session `98989`, exit 0:
1,635 passing contexts, 47 skips, zero failures, 5,561.742 seconds. Original
native bounds/workers and the prior six external-launcher exclusions were
preserved. No deferred XML/Apalache corpus or killed Java workload restarted.
All builds/tests/children shared CPUs 0-7 with GOMAXPROCS=8, without race.
Receipts: `.codex-gotmp/root-after-output-parity-20261018/`. This verifies the
output, statistics-null and recorder corrections in that snapshot; the later
IOUtils binary-null change has its separate complete suite evidence below.
Both original sequential heap workloads remain live and uncredited; this pass
does not resolve remaining source-failure reconciliation or the full port goal.

Binary `IOUtils.IOSerialize`/`IODeserialize` preserve source null path, string
carrier and compression failures before opening files. Serialization checks a
null payload after opening the stream and inside its finally-close phase.
Nine frozen Java cases and a native typed-nil variant match exact messages,
precedence and file effects, including finalized empty gzip output. All 29
existing stream/IO contexts pass. The full original CommunityModules suite on
`d9571b9` passed both phases in session `93754`, exit 0, 512.933 seconds, with
original settings/bounds on CPUs 0-7 and no race. Receipts:
`.codex-gotmp/ioutils-binary-null-20261018/`.
The completed root gate used `91bd7b5`, before the later I/O corrections; their
guard coverage comes from the separate suite and diagnostic checks.
`IOUtilsAtoi` now treats typed-nil arguments as Java null. Text-I/O format
selection rejects a null `UniqueString` instead of selecting an empty format.
Trace binary I/O likewise rejects null path carriers before filesystem access
and checks typed-nil payloads after opening its stream. Three direct Java/Go
trace observations match exact messages and file effects. The five original
trace-model checks pass, as does the complete dump/load matrix: 32 passes,
three retained upstream skips and zero failures. These focused checks do not
constitute a newer full root gate. The complete original CommunityModules
suite on `c87a0fd` passes both phases in 509.318 seconds, including the later
TXT path/payload/charset/option carrier corrections and six Java-matching error
records. Its successful invocation unsets `JAVA_TOOL_OPTIONS`: injecting it
adds JVM startup stderr and fails an original process-output assertion.
Receipts and source qualifications are in `PORT_PROGRESS.md`; no original-method
credit changed.

Disk-search IOException diagnostics preserve Java's six separate GENERAL
parameter strings and trailing spaces rather than merging the five numeric
fields. A temporary reader-open failure diagnostic captured the exact source
layout and rethrow; all 26 original short disk-set/overflow contexts pass on
the correction. No permanent test or workload bound changed. Receipts:
`.codex-gotmp/disk-lookup-error-params-20261018/`.

Statistics.ChiSquare preserves null-argument, null count-element, null-array
and null string-carrier failure phases. All 22 source/native diagnostic outcomes
match, including precedence. The complete original CommunityModules suite on
the final guard snapshot `598c970` passed both phases in session `25846`, exit 0,
507.516 seconds. Both original phases and bounds remained intact on CPUs 0-7
with GOMAXPROCS=8, without race. No permanent new test was added. Receipts:
`.codex-gotmp/statistics-null-20261018/`, including `ant-carriers.jsonl`.
Earlier snapshot results and detailed comparisons remain in `PORT_PROGRESS.md`.

The trailing original `AssertExpressionStack` nested-expression assertion now
has isolated execution evidence: moving only that existing assertion earlier
in a temporary overlay lets it pass before the known fatal trace-helper check.
All three expression locations match the preserved Java log exactly. The
repository test retains its original ordering and all source failures; this
earns no complete-method credit. Receipts:
`.codex-gotmp/assert-stack-tail-20261018/`.

The statistics-module exception-number correction (`b33c33f`) passes the full
original `TestJavaCommunityModulesAnt` translation: session `70009`, exit 0,
501.803 seconds, with both `all` and `shiviz` phases and original bounds.
Its seven temporary Commons Math/native message comparisons also match,
including grouped counts, bounds and dimensions. All children shared CPUs 0-7,
GOMAXPROCS=8, without race instrumentation. Receipts:
`.codex-gotmp/statistics-number-format-20261018/`.

The post-formatting full `tlc_verify` package gate on production `fc8a9ef` is
terminal. Restricted session `41803` exited 1 in 541.096 seconds, with 25,465
passing, 390 failing and 5,728 skipped contexts. All 138 failed parents passed
with host socket/ptrace access (549 contexts, session `93973`, exit 0 in
24.569 seconds). The one additional environment skip, the IPv6 listener parent,
also passed with its two original children on host. Replacing those complete
parent results gives 25,885 passing, 5,727 source-skipped and zero unresolved
failing contexts, matching the prior full-gate totals. This is reconciled
multi-run evidence, not a single green invocation. Original long bounds were
preserved on CPUs 0-7 with GOMAXPROCS=8, without race instrumentation.
Receipts: `.codex-gotmp/output-parity-gate-20261018/summary.json` and the
associated JSON logs. The user's original red pair, `TestJavaGithub971a` and
`TestJavaGithub971b`, also passes on current source (exit 0, 0.044 seconds).

The original `TestJavaOffHeapDiskFPSetLong_testMaxFPSetSizeRnd` workload passes
in session `27326`, terminal exit 0, after 73,649.77 seconds. The full
2,147,483,648 iterations, checkpoint operations, invariant and final-size
assertions pass with the original default 64 MiB configuration, ratio-1.0 factory
and seed, normally without race instrumentation. The fingerprint-set paths
exercised here are unchanged since its build; later string-fingerprint changes
are outside this random-long workload. The preserved log is
`.codex-gotmp/offheap-random-full-after-outage.log`. Long-test inventory credit
is now 19 of 22 contexts; the three original sequential workloads remain pending.
Do not restart this completed workload. Older interrupted runs and the duplicate
session `7371` stopped with exit 130 earn no credit. Namespace process lists
alone cannot establish that a tool-owned process has stopped; poll its original
handle before deciding whether to restart.
Former sequential Java session `69494` and native session `85088` did not
survive the power outage and process cleanup. Neither produced a terminal result
or completion credit. Their logs remain at
`.codex-gotmp/fpset-sequential/java-offheap-full.log` and
`.codex-gotmp/fpset-sequential/native-offheap-full.log`. Do not restart either
multi-hour workload without explicit authorization. The native job used the
earlier ignored driver binary; it is separate from the installed tagged
translations.
The first installed full-bound native sequential methods are running in
sessions `48346` (MSB) and `39435` (LSB), using the unchanged `dcdfea3` binary
(build `68804`, exit 0). These are first runs of the two heap methods, not
restarts of either interrupted off-heap job. Each retains all 3,221,225,473
requested iterations, its default configuration, original assertions and
checkpoint behavior. Both share CPUs 0-7 with GOMAXPROCS=8, without race or
memory overrides. The LSB factory reports its original 4,096 MiB table budget;
the pending count correction is unapplied. Poll the same handles. Receipts:
`.codex-gotmp/sequential-msb-full-20261018/` and
`.codex-gotmp/sequential-lsb-full-20261018/`. Neither has terminal evidence or
method credit. The separate LSB count correction and interrupted off-heap
restart decisions remain pending.
The original generator produces a duplicate at `i = 1,073,741,825`, equal to the
previous iteration's fingerprint. Keep this source issue visible; do not alter
assertions, bounds or production semantics to obtain a pass. Faithful drafts for
all three factories compile in ignored `.codex-gotmp/fpset-sequential/driver`;
compilation and partial execution earn no original-method credit.

The existing native
`TestModelCheckerDoNextEnqueuesOnlyUnseenInModelSuccessorsButChecksAllImpliedActions`
now seeds the initial trace state and counts persistent records through the disk
enumerator after flushing the writer. Its existing enqueue and validation
assertions are preserved; the count requires the seed plus one unseen successor.
Disk-backed traces continue to avoid an in-memory state mirror. This native
harness reconciliation adds no original Java method credit. Details and receipts
are in `PORT_PROGRESS.md`.

Earlier full-gate chronology, counts and receipts are retained in
`PORT_PROGRESS.md`; use the current baseline and active handles above for
snapshot qualifications. Concrete `FPSetStatistic` management dispatch retains
the off-heap bean's identity and invokes its capacity, size, eviction and
invariant methods. Focused race and original model/storage evidence for those
contracts is recorded there too. Passing gates do not resolve source-failing
original methods or establish whole-port completion.

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
truncation, then passed unchanged with captured output. A later captured run of
the same selection failed in the native one-worker remote TSnapShot case: its
summary had 8 distinct states and 3 queued, against the original zero-queue
assertion. Three isolated reruns had 6 distinct and 0 queued. Controlled runs
with Java's static selector show that one, two and three states per block leave
3, 1 and 0 queued respectively when the model error stops this one-worker run.
The source server thread dequeues a block before calling the worker; its outer
model-error path does not requeue that assignment. Zero queued is therefore
schedule-dependent here.
This does not prove the older lost assertion was the same failure. The
source-harness diagnostic passed with its recorded 48 workers explicitly set
under CPU affinity. Retain the exact assertion and capture full output for
future runs; see `PORT_PROGRESS.md` for receipts.

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
Local catalog lookup/creation and native coordinator listener creation narrow
registry ports to Java `int` before default-port and range handling.
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
| Coordinator thread IDs | The constructor stores Java `COUNT++` as a signed 32-bit ID, including wraparound in worker labels. |
| Worker loss | Cancel keepalive, claim cleanup once, remove registration, requeue, clear assigned states, wake consumers and decrement workers in source order. Convert an absent assigned block to an explicit empty batch. Preserve queue suspension and the ten-second/sixty-second keepalive schedule. |
| Failure categories | Native operation traits drive retry/removal/exit. Retain causes and sender frames. Memory exhaustion permits smaller batches; executor rejection does not. Ignore closed errors only when every cause is benign. Exit-only lost replies must not suppress earlier insertion/checkpoint failures. |
| Fingerprint routing | Preserve forward reassignment, aliased wrappers and slot-based statistics. Worker snapshots omit the coordinator trace. A local MultiFPSet remains one endpoint with high-bit child routing; manager server routing uses low bits. Fingerprint-bit configuration/indexer inputs, dynamic manager server counts, block-selector limits/static sizes and completion-wrapper indices narrow to Java `int` at their public boundaries. |
| Nested storage | Named checkpoint/recovery, size and checks join concurrent children. Child I/O becomes an operation failure, outside manager checked-I/O warning/fallback. Unnamed checkpoint operations remain sequential. Parent states-seen counters exclude child counts. |
| Storage lifetime | Retain source partial mutations, reader slots, lock ownership and failure precedence. `DiskFPSet.Init` is unsynchronized like Java. FP-set initialization and worker-registration counts, plus `LongArray` zeroing thread counts, narrow to Java `int` before allocation, propagation, comparison or division. Roll back newly owned native descriptors without reopening closed stores, retrying writes or promoting pending snapshots. Native disk cleanup skips reader slots that initialization or reader reopening did not fill, so it cannot replace the source open failure. Later disk lookups select retained worker/pool slots by index even when a slot is absent; they must not open a fallback reader and mask Java's null-reader failure. If native publication never completed, selection falls back to the retained source worker array rather than treating it as empty. Constructors retain source allocation/error order. |
| Paths and files | Preserve literal empty names/directories, dot components and symlink traversal. Queue/trace/FP/DFID operations must not invent parents or temporary fallback directories. File requests use fresh default resolvers; explicit overrides and worker basename caching remain captured. |
| Queues | Publish bulk logical length only after the loop; failed spills retain uncounted prefixes. Reject nil batches before mutation; empty batches remain valid. Byte decoding follows removal, while array conversion finishes before publication. Preserve partial pool reads and source diagnostics; background pool/cleaner exceptions exit the process rather than silently killing a goroutine. |
| State ownership | Ordinary states retain base no-op action/cache/callable accessors. Extended copies reapply predecessor setters and depth limits. Print wrappers own a separate mutable state and independent metadata; their record conversion returns the existing record. Disk traces store RAF links/fingerprints without retaining an in-memory state/action mirror. Negative trace seeks fail after pointer mutation. |
| Evaluation | Worker checks precede constraints. TLCApp retains captured action/property arrays, vector concatenation, separate returned-array ownership and predecessor policy. Replay uses its existing evaluator. Coordinator publication inserts fingerprints before trace and queue writes; missing owners cannot fabricate success or zero statistics. |

ActionItemList emptiness is singleton identity, not structural shape. Base
constructors retain null tails, list methods reject null receivers, and a null
Action fails. Fresh source runtimes recreate both ordinary and extended empty
singletons; extended `cdr` keeps its source predecessor/action mutation.
Both `cons` forms narrow their kind argument to Java `int`, and coverage cost
model levels narrow at their source setter before storage.
Action methods reject nil receivers, and rendering/location methods reject a nil
predicate. `IsPossible` retains Java `instanceof` behavior for a nil reference,
and PossibleAction continues to override the predicate through `GetPred`.

Native payloads preserve supported graph identity, cycles, nil/empty distinctions,
concrete scalar/container types and isolated receiver ownership. State, value,
name, buffer, vector, map and finite-operator tables have validated references.
LongVec transfers active entries rather than spare capacity; BitVector retains
all words. BitVector also preserves Java's signed constructor/index arithmetic,
masked negative shifts and null-versus-empty word storage across its public
operations. The distributed completion path uses the source `BitVectorWrapper`
carrier to associate each returned vector with its input index before restoring
request order. FPSetConfiguration rejects nil method receivers before argument
validation. Its fingerprint-width accessors use Java's exact `GetFpBits` and
`SetFpBits` spellings; the older acronym-normalized spellings remain delegates,
and production configuration and storage paths use the source names. Its
multi-set wrapper rejects a nil source configuration.
Concrete fingerprint sets retain and return the exact constructor configuration,
including nil; only the source no-argument memory-set constructor creates a default.
DFID status logic uses the source `IncLevel`, `IsCompleted`, `IsDone`, `GetLevel`
and `IsLeaf` entry points; prefixed spellings remain delegates. Memory and multi
FPIntSet methods reject nil receivers, and multi-set traversal does not skip nil
children where the source visits its child array. Inherited `AddThread` remains
a receiver-checked no-op.
Invalid finite-operator nil arguments/rows/results fail, while actual
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
checkpoint corruption. Their null error parameter exposes an upstream
GENERAL-formatting loop. The approved Go formatter correction reports this
failure without looping; a fresh coordinator observation verifies startup stops
before construction and remote recovery, unchanged checkpoint bytes and joined
roles. This is failed-startup cleanup, not successful model recovery. See the
architecture notes on incomplete intern records. Server
recovery runs trace, queue, then fingerprints. Intern recovery holds its mutex
from before open through token publication/replay.

FPSet trace recovery uses the source `Recover(trace)` boundary. Named remote
recovery remains `RecoverFile(name)` for Java's string overload; the older
`RecoverTrace` spelling is only a compatibility delegate.
MemFPIntSet closes checkpoint streams only after normal write/recovery loops;
mid-stream failures retain the descriptor and do not run an invented deferred close.

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
| Worker successor checks | Local/TCP cases verify fingerprint lookup before invariants, invariants before constraints, model/action short-circuiting and exact failure context. Seen states skip checks; only accepted states receive predecessor UIDs. Cache eviction followed by an equal fingerprint retains the first holder and predecessor. The worker does not insert fingerprints. |
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
| Recovery statistics | The recovery-end fingerprint count renders as Java's signed `long`, including `-1` after a checked local size failure. Existing local/TCP recovery checks pass; the exceptional rendering follows the pinned source but has no direct original test. |
| Begin reply loss | Real begin leaves pending files without commit. Direct Mem warns, restores the healthy host and completes evaluation. Nested LSB/MSB stop before the healthy host/publication and preserve both pending children. |
| Missing/corrupt snapshots | Missing Mem committed files warn/continue; truncated/duplicate direct Mem records stop before the next host or publication, retaining partial inserts and checkpoint bytes. A failure at the second Mem host also retains the first host's complete reconstruction. Missing, empty or out-of-order nested disk snapshots stop after sibling joins. Ordering failures retain partial writes and the source full-file counter. Trailing partial records preserve complete membership and permit full-model completion. Missing/truncated coordinator queue stops before FP recovery; missing/truncated trace plus a missing queue reports the trace failure first. Retained bytes and process joins are checked. |
| Late commit failures | Missing intern temporary stops final local FP commit, retaining its old snapshot and new pending file; remote FP has already committed. Missing local FP temporary fails after queue, trace and intern promotion. Failed rename retains source deletion of the older destination. These are short real-file checks, including a TCP FP endpoint, not full-model restart proof. |
| Accepted request loss | Short begin/commit/recovery cases let accepted work finish after connection loss, with exact files/membership and no replay/reassignment. This is not host-death coverage. |
| Trace-to-intern boundary | External GDB/disassembly verifies local and full remote N=7 interruption after trace commit and before intern commit. Queue/trace and intern generations differ; remote FP is new, local FP old. Remote recovery completes. Opt-in Linux/amd64 fixture requires ptrace, GDB and an optimized symbol-bearing `go test -c` binary. |

Short storage/startup checks also cover direct/nested local/TCP malformed named
snapshots, all partial-long lengths, native syscall read/write/close failures,
source buffer boundaries, literal paths, off-heap nil batches and queue/intern
ownership. Those checks add no original-method credit and do not prove full-model
restart or checkpoint atomicity. See architecture notes for their exact limits.

Memory fingerprint monitor ownership now matches the three source
implementations. `MemFPSet` locks size, memory-size, put and contains;
`MemFPSet1` locks only put and contains; `MemFPSet2` locks size, put and
contains. Their source-unsynchronized distance and checkpoint traversals remain
outside those monitors, and `MemFPSet1` file recovery and `recoverFP` mutate its
backing `SetOfLong` directly. `MemFPSet` and `MemFPSet2` recovery still acquire
the monitor once per source `put` call.

DFID `MemFPIntSet` locks size, memory-size and its three status operations.
Source-unsynchronized `allLeveled`, distance checking, checkpoint traversal and
direct recovery remain outside the set monitor. The model checker's periodic
checkpoint still owns that monitor around the complete checkpoint sequence,
matching Java's outer `synchronized (theFPSet)` block.
The DFID coordinator now also holds the checker monitor across its `done`
check and progress report, matching Java's level-loop ordering. The shared
checker monitor is reentrant so synchronous progress recorders may call back
into checker completion, as with Java's `synchronized` monitor. DFID also
resets `done` under that monitor after each level. The ordinary worker's fatal
error path now notifies the coordinator after finishing its captured queue,
matching Java's monitor-protected order. The four ordinary checker error-report
paths likewise notify after finishing the queue only when they win the error
claim, preserving Java's notification timing.
DFID likewise notifies after a winning property, evaluation, incomplete-state,
deadlock or general next-state failure report, matching Java's monitor order.
The distributed server thread's completed-block path now requires its captured
queue at Java's `finishAll()` call after marking done and notifying; a missing
queue reaches the outer error handler and thread finalizer.
Distributed server initialization now keeps the computing message and the
post-init queue-size announcement inside Java's init `catch(Throwable)` region;
call-stack replay still invokes `doInit` without those announcements.
Application-backed server file-name getters now read the current application,
as Java's worker-facing getters do; constructor-captured names still own trace
and checkpoint files.
Ordinary, DFID and distributed state-count messages render fingerprint sizes
as signed Java `long` values, including exceptional negative bit patterns.
Distributed rate getters now retain Java's unsynchronized access boundary:
native atomic snapshots let management queries return while another thread
owns the coordinator monitor. Periodic publication, message reads and final
zeroing use separate atomic operations; the rates remain independent statistics.
Ordinary checker progress publication and management rate queries likewise use
independent atomic snapshots, removing their observed Go read/write races while
retaining Java's unsynchronized management access.
Periodic liveness work now reads the forced-check flag at Java's decision and
call sites instead of retaining a value captured before suspension. The flag
and Java's private runtime ratio use native atomic storage; the public ratio
getter/updater methods remain available. Success-only clearing of the forced
flag retains its source timing.
Runtime-ratio updates also reread the mutable progress interval separately for
the elapsed-time subtraction and final denominator, matching Java's field-read
order instead of keeping one cached interval across the calculation.

## Remaining distributed work

The source inventory is fully mapped, but mapping is not proof of parity. The
four model reconciliations have translated assertion bodies and preserve Java's
shared unconditional skip; they are not four absent implementations. The
upstream off-heap flusher defect has an explicitly authorized Go fix. The user
also approved a separate Go correction for the intern-error formatting hang.
`output.go` now replaces each parameter's existing markers once; a
self-referential marker returns, and markers introduced for later parameters
still resolve. The original MP checks and full `tlc` package gate pass on this
change. See `PORT_PROGRESS.md` for the receipt.
Further failure-phase and partition checks below are verification gaps, not
identified missing source features. Choose concrete source behavior and consult
existing receipts before adding scenarios; do not turn this into an unbounded
fault-testing project. Original-method credit remains unchanged.
The final top-level production-name sweep leaves only Java interface shapes:
`DistApp` is the abstract contract implemented by Go's concrete `TLCApp`, and
the six unmatched value-interface names are marker or carrier capabilities
already exposed by their concrete Go values.
The corresponding nested public-type sweep is also complete. Active nested
carriers, iterators, formatters, exceptions, enums and static helpers all have
Go counterparts. `DiskStateQueue2TLA.Action` is the sole source-only nested
enum; it belongs to disabled JFR instrumentation, and every production
`trace(Action...)` call site remains commented out in the pinned source.

1. Extend recovery coverage to additional failure phases beyond the matrix above.
   Intern-header, intern-record and missing/truncated trace and queue startup
   failure ordering are verified; other phases remain open. The intern-record
   observation includes the approved Go formatter correction.
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

Original performance work now includes complete translations of
`FP64Benchmark`, `CombinatoricsBenchmark`, `LongArrayBenchmark` and
`OffHeapIndexerBenchmark`, plus the full parameter matrix of
`LongArrayInitializeBenchmark` and `IntervalValueBenchmark`. The interval work
also exposes the source `IntervalValue.getRandomSubset` API. The complete
`SetEnumValueBenchmark` and direct `SetEnumValue.getRandomSubset` API follow the
same source path. `SetOfFcnsBenchmark` is also complete, with direct random-subset
methods for the function, record and tuple product-set classes. Benchmark
work also includes both `SubsetBenchmark` methods across every source-executable
parameter pair and both active `FcnRcdBenchmark` lookup strategies across all
sizes. `FcnRcdValue` now exposes the source linear and thresholded-binary lookup
paths. Benchmark inventory entries are separate from the correctness-test
percentages. All five `EnumerateSubsetBenchmark` methods are also installed for
the complete size matrix through 19 elements. `StateQueueBenachmark` is complete
with its three balanced two-producer/two-consumer groups and seven sizes.
`DiskQueueBenachmark` is complete across both disk queue implementations, all
seven state widths and all four grouped thread counts. `SimulatorBenchmark` is
installed with its exact fixtures, six worker counts and caller-owned RNG path.
All 40 `SubsetValueBenchmark` configurations are installed with their exact
probabilistic and maximum-length workloads.
All 75 `RandomizationBenchmark` methods are installed by source name across
their four random-subset families, including the two source methods whose
function-set target does not match the method name.
`ModuleOverwritesBenchmark` is installed with all three invariants, the exact
model/config fixtures and the source native-override loader path. The original
JMH class inventory now has no unchecked entries.

Run only relevant checks. Never combine long workloads with `-race`; use race
instrumentation only for short concurrency selections. Do not shorten original
bounds. The default root `TestJavaCommunityModulesAnt` can launch a Java TLC
child, so exclude it from broad Go runs while restarting the killed Java
workload remains unauthorized. For offline work, use:

```bash
taskset -c 0-7 env GOMAXPROCS=8 \
  GOCACHE="$PWD/.codex-gocache" GOTMPDIR="$PWD/.codex-gotmp" \
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
