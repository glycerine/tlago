# TLC Port Handoff

Updated: October 10, 2026. Active branch: `master`.

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

Core bridge continuation: runtime views adapt the checked SANY graphs, retaining
Context Pair history and Hashtable topology, source syntax, declarations, formals,
proofs, substitutions and module ownership. UID traversal includes Context and
substitution records. Generated expressions retain their actual child graphs;
debugger compilation preserves the running module table. Runtime helpers stay
outside user source imports and inheritance. All 12 observed Test219 whole-module
traces and 60 getter arrays match; see `TLC_ARCH.md` for the contracts and limits.

Runtime targets resolve actual external-module OpDefs. Invariant templates merge
in the processor before constraints, and configuration phases stop at their first
error. Postcondition getters retain runtime, `_POSSIBLE`, then config order and
fresh config Actions without OpDef metadata. View and alias getters resolve
current definitions and throw source runtime failures. View capture follows
configuration and precedes symmetry setup; alias lookup stays inside trace
rendering's error boundary.

Unnamed INSTANCE convenience aliases and early registration reuse canonical
source exports and symbols. Module setup retains checked
source locality and ownership, including public RECURSIVE declarations completed
by LOCAL bodies. Parser-backed action declarations read the current syntax child;
absent declarations stay unknown. Shared semantic bases and symbol views read
current syntax locations, and symbol syntax setters update the same owner.
Formal path lookup uses those getters. State setup retains declaration arrays;
spec metadata and coverage read their current source locations when requested.

Parser-owned diagnostic strings use the source semantic formatter, including
presence-based `showPlainFormulae` selection and numeric overrides. Generated
literals join that source-owned path. Native lookup aliases retain name strings.
All 20,390 observed diagnostic strings agree, and 20 independent Java formatting
rows match; detailed bounded observations and run receipts are in `PORT_PROGRESS.md`.
General source generation, graph mutation sharing and source-less fallback
lowering remain unproven.

Operator comments and human-readable images retain source failure boundaries,
the nil-versus-empty child-array branch and property-aware fallback formatting.
All 44 observed Java/native cases agree. Builtin and null syntax expose Java's
empty comment/child arrays. Variable locations and spec records match all 16
observed rows after setup, including replacement in the retained array. Null
location records use Java's `--unknown--` module. Spec metadata and coverage now
use the requested tool's declarations; counters belong to those declarations.
All 15 observed ownership/report/failure rows and 11 counter lifecycle rows match
Java. Constraint metadata now preserves Java's operator/action casts and null
failures; all 45 observed state/action cases agree. Constraint getters now retain
current processor arrays; all 56 ownership/metadata/filter observations match,
including empty/null arrays. Invariant, implied-property and temporal action/name
getters now share current processor arrays; all 128 observed rows match Java.
Config processing preserves non-null empty arrays, and metadata/liveness array
failures match. Initial/next predicates and assumptions now read current processor
values, including initial generation and axiom checking; all 53 observations match
Java. Reward and periodic predicates now read the processor with source capture
lifetimes; the simulator reporter captures before startup returns and keeps
reporting until shutdown after a false predicate. Bounded Java/native observations
and focused original checks pass; see the progress log for receipts. Simulator
initial checks now retain the constructor's invariant array; worker successor
checks re-read current arrays as they loop and fetch names only for diagnostics.
All 49 bounded capture/replacement/null/failure observations agree with Java.
Checker successor and initial-property loops now follow current arrays and names,
including source null boundaries, retained initial exceptions and empty diagnostic
names; all 147 bounded observations agree with Java. DFID preserves captured
successor property lengths, current elements/names and its shared loop index;
initial loops and early-failure arrays now follow Java. All 152 bounded DFID
observations agree. Distributed `TLCApp` now preserves source null-array/name
boundaries across tool replacement; all 109 observations agree. Checker and DFID
comparisons also pass with action inputs that fail in the predecessor state.
Distributed worker property failures now agree in six direct Java/native
observations using existing model predicates: messages, original causes, state
references, replay flags, generation counters and computation cleanup. Focused
local/TCP worker checks pass; no production correction was needed. Initial
predicates now retain the processor's `Vect` object instead of a slice snapshot.
Cached growth, shrinkage, replacement and restoration match Java in 13 bounded
observations; all 53 earlier predicate observations still agree. Focused original
checks pass, including all five generated-trace tests. Simulation statistics
preserve current actions, captured matrices and variable counter slots, source collection order and worker-owned trace IDs. Action-flow
writers preserve empty files, labels, weights and context grouping. Source-backed
Simulator statistics select the actual scoped simulation worker and retain Java's
report capture order. Direct Simulator register getters read the first registered
worker; `TLCGet`, `TLCSet` and `TLCGetOrDefault` use the calling simulation worker.
Trace getters select that worker or preserve the source single-worker assertion
and null boundaries. Bounded Java/native comparisons and focused original tests
pass; detailed coverage and limits are in the progress log and architecture notes.
The latest dispatch comparison covers 80 rows, and the original 4,224-depth
simulation checks pass at unchanged bounds. The existing short scope-isolation
race check passes. Aggregate snapshots retain first-worker domains and source
failures; register growth preserves old arrays and setter partial updates. The
indexed all-worker getter is ported, and all 43 new register observations agree.
Probabilistic disjunctions now use the active simulation worker's random stream;
ordinary and ID-only callers use the coordinator stream. All 52 bounded state,
stream and failure observations agree, and focused originals pass. Bounded-choice
and nested-disjunction checks on five existing models also match all 275 rows,
including functor stopping, callback failures and retained states. No correction
was needed for these cases. Action-composition accounting and functor delegation
also match all 98 observations across seven existing configurations in ordinary
and probabilistic modes; all seven original composition tests pass. This includes
Java's unsupported probabilistic composition paths and partial register effects.
`Vect` now keeps its signed 32-bit logical count separately from backing storage,
preserving invalid counts, overflow and mutation before failures. All 330 count
observations, 27 real-model consumer observations and 149 earlier observations
match Java. Initial generation, combined actions and `TLCGet("spec")` preserve
source iteration and first failures. Coverage creation/reporting now preserve
source null failures in initial, next, invariant and implied families, including
partial updates and report prefixes. Constraint creation now requires the source
OpDef cast, installs a fresh Action after building its cost model, and preserves
repeated-creation failures. Constraint reports retain null/cast failures. Reports
now retain the first action per predicate location, then deduplicate cost models;
module ordering follows interned tokens. Variable setup preserves partial counter
replacement and null failures; reports read each current declaration/counter in
the captured array instead of snapshotting the whole collection. All 175 coverage
observations and 30 focused original tests pass. Collector primed locations retain
their construction-time capture; all 144 graph/lifecycle observations on 18 models
match Java. Coverage traversal now follows semantic operator definitions even
after runtime pre-evaluation. Shared substitutions retain separate next/invariant
counters. `ENABLED` now binds by symbol, retaining persistent functional states
and Java's failure on mutable targets. Source-backed loading rejects configuration
failures before checker construction can mask them. All 163 evaluation/ownership
observations and 49 focused original tests pass. Next-state subscript collection
now follows Java's tuple, definition and context handling, including unsupported
expression and missing-variable warnings. All 141 entry-point, CLI-alias and
collector observations match; 48 focused original tests pass. Native graph
cleanup now accepts configuration failures that occur before a checker exists.
Warning suppression and escalation now match Java across 272 ordered
observations; eight focused original tests pass. Elevated subscript and symmetry
warnings retain non-null empty parameter arrays, preserving exception recording.
Trace reconstruction now respects state-mode metadata and source current-state
replacement. `TLCExt!Trace` owns successful-history restoration and preserves
equal-valued states as separate trace positions. All 20 worker-backed repeated-
state observations and 15 focused original methods pass; earlier metadata and
lifetime comparisons remain recorded in `PORT_PROGRESS.md`. The JSON auto-worker
prefix mismatch remains an explicit source limitation in existing deterministic
replay evidence; keep original assertions. `TLCExt!Trace` now requires Java's
checker/current-state owners instead of returning fallback traces. All 40 owner
and repeated-state observations match; 15 focused original methods pass.
Incomplete `TLCExt!Trace` states now raise the source coded runtime exception
with its parameters and formatting event; evaluator wrapping retains the full
method signature. All 48 diagnostic observations and 15 original methods pass.
TLCExt evaluating overrides, TLCEval, the trace-state helpers and sequence
replacement overrides retain source method signatures. JSON and TLC trace-state
helpers retain distinct declaring classes. Bootstrap and loaded-module metadata
comparisons match Java; the complete original CommunityModules Ant target passes.

AssertError catches only source evaluation/runtime failures, then evaluates its
expected message with the state-expression overload. PickSuccessor checks seen
fingerprints before guard evaluation and performs one lookup. Seen states skip
the guard; null and non-Boolean guards retain source exception categories.
AssertError, PickSuccessor, TLCModelValue, TLCFP and noninitial model-checking
Trace share Java's reentrant class monitor. Initial and simulation traces retain
their bypass paths; TLCCache keeps its separate read/write lock.

Twenty related original methods pass. Current bounded comparisons cover 22
PickSuccessor cases, 18 monitor cases (also checked with short race instrumentation)
and 11 AssertError boundaries; detailed receipts and limits are in PORT_PROGRESS.md.
PickSuccessor reconstructs the first matching action without changing the current
state slot. Extended states retain their stored action, including null. One
console reader retains startup input, charset and read-ahead across prompts,
with Java line endings, trimming and EOF failures. Seventy bounded console
comparisons match source across the six supported charset families, including
a healthy checker marking a successor explored. Three original iterator classes
also pass after the optional exception-message change. These observations do not
prove I/O-failure behavior, extended charset providers or all interactive timing.
TLCEvalDefinition resolves through the current root module rather than replaced
runtime bindings, preserving body selection, arity checks and null-name failures.
Eleven bounded source comparisons and 14 related original model methods pass.
State-level TLCCache keys use Java's state-expression evaluation arguments,
including EmptyState and disabled coverage, even on cache hits. Null-key/current
state failures preserve source ordering. Sixteen bounded comparisons and five
related original model methods pass. Constant-cache/hash/lock audits remain
separate and need no repetition without a new concrete gap.
TLCDefer setup failures retain the source coded runtime category and nullable
message; partial setup keeps the first callback. Ten bounded API comparisons
and 11 relevant original methods pass. Enhanced JVM NPE descriptions remain
outside the native diagnostic contract.
TLCEval conversion rejects null instead of creating UNDEF. Twenty-one bounded
comparisons match source. Assumption-only models now prepare an allocated empty
action list, preserving the strict coverage null-array check; unchanged Java
Github652 and 11 relevant original Go model methods pass.
Function combination (`@@`) converts both operands before checking either
conversion result, preserving lazy right-side effects and failure precedence.
Ten bounded comparisons, all five original TLCTest methods in Java and Go, and
four relevant original model methods pass.
`_Possible!_Counts` now accepts only function records, normalizes each source
record, preserves malformed-count failures and wraps integer sums. Its raw output
uses Java HashMap merge iteration order. Twelve bounded source comparisons,
nine relevant original Go model methods and unchanged Java PossibleCountsTest
pass; these observations do not prove every value-key collision or simulator path.
`Permutations` and `RandomElement` now propagate set-conversion failures instead
of replacing them with finite-set argument errors. Integer membership and invalid
union/difference enumeration retain runtime failure categories and source metadata.
Twelve full-result comparisons and three sourced-failure comparisons match Java;
relevant original module and four random-element model methods pass.
Scalar and model-value failures now retain Java runtime categories, formatted
arguments, source metadata and wrapped null failures. Model-value type prefixes
use UTF-16 units, including unpaired surrogates. All 84 scalar, 468 model-value
and ten additional prefix observations match source; relevant original value,
permutation and model methods pass. Keep function-record Java tests isolated:
the earlier combined class-order failure remains preserved in PORT_PROGRESS.md.
Tuple membership, explicit-set failures, interval membership/EXCEPT, function
shape comparisons and record failures preserve source runtime metadata and null
ordering. All 120 collection and 160 record observations match, including partial
record mutations after duplicate-field failure. Relevant original value and model
methods pass; Java record, tuple and interval originals also pass. Array dispatch
now preserves tuple/record arity checks and both source catch boundaries. All 56
array/sequence-helper observations and relevant original methods pass; function
records retain their default tuple-packing behavior.
Explicit and lazy function failures retain source runtime metadata. Receiver
formatting precedes null operand failure where Java formats it first, preserving
normalization and exception precedence. Lazy array application rejects null before
its source catch. All 288 explicit-function, 98 lazy-function and eight direct
vector observations match; relevant original value/model methods pass. Materialized
record flags and exact wrapper counts remain part of the lazy comparisons.
Intersection finiteness now short-circuits after a finite left operand. Set
intersection/difference and UNION failures, including EXCEPT, retain runtime
categories, source metadata and wrapper counts. Of 200 set-operation observations,
194 match exactly; six differ only in native versus JVM iterator identity text.
Later-inner UNION failures now retain their distinct diagnostic and source
metadata. Reset preserves the source's unwrapped null/cast failures, and explicit
set iterators read the current owner vector after mutation or replacement.
Of 69 iteration observations, 53 match exactly and 16 differ only in iterator
identity text. Relevant original value/model checks pass. Recovery after a caught
iterator failure remains unverified; no broader iteration parity is claimed.
Tuple-product failures, subset membership, ordinary isEmpty misuse and overridden
value Size now retain runtime categories and source metadata. Null tuple/subset
members retain typed failures; k-subset null checks follow the original empty-set
short-circuit outside the source catch. All 288 observations match exactly, and
relevant original value/model checks pass. Record-set membership, EXCEPT,
enumeration, overflow and duplicate-field construction now preserve their source
runtime boundaries. All 267 observations match, including error codes/parameters
and partially mutated constructor arrays. The full original record-set suites pass.
Function-set membership, overflow, EXCEPT and domain/range enumeration now retain
source runtime failures, and size queries preserve Java's repeated range-size call
order. UserValue comparison also retains runtime metadata and formats the receiver
before a null operand fails. All 562 observations match; all 16 original function-set
methods and relevant model checks pass. Undefined-value failures, overridden-value
EXCEPT, and default fingerprint/permutation failures now preserve runtime metadata.
ANY emptiness retains its source-less root failure inside the value catch; Nat/Int
null membership fails before constructing a coded evaluation error. All 132 special
value observations and relevant existing/original checks pass. Operator
records, lambdas and reflected/evaluating/priority/callable wrappers retain the
observed runtime metadata, typed null failures, WrongInvocationException and
argument/catch ordering. Numeric and represented core/annotated method metadata
retain actual declaring classes, signatures and annotation levels; four IO
command overrides retain level 1. Bounded Java/native comparisons and the
relevant original module/model checks pass; detailed bounds are in TLC_ARCH and
PORT_PROGRESS.

TLCSet now uses Java's value-array MethodValue registration. Argument evaluation
failures escape before its method catch; nil inputs and ownerless integer registers
retain typed source failures. All 24 boundary observations and the original BFS,
initialization, two full 4,224-depth simulation profiles and distributed startup
checks pass. Eager native constants use allocated-empty argument arrays.
TLCEval now reads its first argument without a registration arity guard,
preserving null/empty input failures, ignored extra arguments and the constant
expression's null-context failure. Evaluation, lookup, level calculation, local
definitions, lazy caching, variable recognition and state generation retain
explicit nil contexts and source dereference order. Literals can still evaluate
without a context. Initial membership assignment retains runtime error metadata;
functional ENABLED states and partial mutations match. The 14 TLCEval, 124 core,
384 cache/recognition and 144 generation observations agree with Java, and
relevant original model/debugger and focused checks pass. Detailed bounds and
receipts stay in TLC_ARCH and PORT_PROGRESS. Enumerators and action constructors
also retain their supplied contexts; 64 enumeration and 36 action observations
match, including failed-binding cursor state and successor generation. Theorem
contexts now retain their incoming context, and non-enumerable quantifier bounds
retain detailed runtime failures; 54 theorem and eight factory observations
match, with relevant original model checks passing. Ordinary/theorem context
helpers now bind by supplied argument count, retaining short arrays and source
lookup-before-index failure order. All 288 argument-boundary observations match.
Generated zero-argument applications carry allocated-empty arrays; explicit null
arrays retain their source failure. Relevant original model and generated-node
checks pass. Level calculation now retains null-expression, operator and child-array
failures at Java's dereferences, while action shortcuts still return before
reading children and null argument elements remain ignored. All 216 bounded
level observations and relevant original model checks pass. Quantifier context
construction now sizes and traverses formal groups in source order, preserving
empty groups, ignored excess domains and null/index failure precedence. All 102
factory observations and relevant original model checks pass. Argument conversion
now rejects null before lazy construction or formal indexing, and null-symbol
lookup retains Java's typed failure. All 132 related observations and relevant
original model checks pass. Boolean errors for conjunction/disjunction, negation,
IF, implication, equivalence and CASE retain the source runtime category and
failing expression/context. Quantified bodies and subset predicates retain their
extended binding contexts; failed CHOOSE and action subscripts retain their
incoming contexts. Temporal-formula failures use the same detailed runtime
carrier. All 294 observed evaluation rows and relevant original checks pass.
Initial/next generation and ENABLED now retain source boolean failure categories,
expressions and contexts for direct values, builtin results, IF, CASE and literal
predicates, optimized remaining action predicates and implication. All 330 observed
cases agree. Next user-defined value errors use `Context.Empty`; builtin and
optimized action-list failures retain their current predicate contexts.
Temporal and unbounded predicate rejections now retain source runtime failures;
all 88 bounded evaluation/generation/ENABLED comparisons agree. Temporal forall
is source-inspected through its shared branch, without a direct fixture.
Function applications now retain the whole expression for non-function failures
and the argument for non-boolean predicate results; all 224 compared rows agree.
Membership generation now retains detailed non-enumerable-domain failures; all
144 compared cases agree, including bound variables and state ownership.
Lazy/cached lambda domain failures now retain the function expression and caller
context; all 272 compared rows agree, including body bindings and cache changes.
Short lambda arguments now retain Java indexed failures in generation, Apply
and Select; all 160 compared cases agree. Function-parameter construction and
size calculation now preserve source array failures and overflow diagnostics;
all 144 combined rows agree. Parameter enumeration now preserves zero-argument
products, exhausted-product reset behavior and source runtime/null failures; all
30 bounded observations agree. Mixed tuple-formal binding matches all 80 observed
cases. Binding through retained parameter arrays now preserves source null/index
failures and read order; all 2,880 mixed/single tuple and ordinary observations
agree after normalizing process-specific tool identities. Materialization now
preserves direct tuple casts, captured arrays and typed binding/formatting
failures; all 590 observations agree, including cache state and zero arguments.
Lambda domains now retain the captured product width, component identities and
source array failures; all 608 structural/error/cache observations agree.
Uncached lambda sizing and materialization now reject null parameters before
cache publication; all 1,188 size/materialization observations agree.
Tuple conversion now retains typed parameter/domain failures without adding a
source wrapper at entry; all 1,242 conversion observations agree.
Deep normalization now retains typed null failures and partial EXCEPT/domain
normalization order; all 1,266 result/error/mutation observations agree.
Lambda EXCEPT lookup now preserves typed path failures and null-replacement body
fallback; all 1,088 application/selection observations agree.
State-generation lambda binding now rejects null parameters at the count read;
all 3,120 binding observations agree after tool-identity normalization.
Numeric override casts preserve source failure messages and last-argument-first
order; implicit null failures retain nullable details. All 36 expression and 350
override observations agree. Numeric module selection now chooses the source
arithmetic implementation as well as its signature; all 2,004 direct module and
3,000 parsed-definition observations agree, including null/zero/negative cases.
Positive-arity numeric override failures now retain source diagnostics and
wrapping; all 368 arithmetic/comparison observations agree. Integer GEQ
selection also preserves its source diagnostic label. Comparison null arguments
now retain source-ordered formatting failures; all 3,600 direct and 2,700
parsed-definition observations agree. Zero-argument STRING, EmptyBag and
JavaTime calls now retain source arity failures; all 48 bounded observations
agree. Other zero-argument registrations and mutation during evaluation remain
separate targets.
Continue concrete native override gaps against source and the original-test
inventory.
Original model-test reconciliation remains open.

## Verification baseline and test credit

Active full workload: `TestJavaOffHeapDiskFPSetLong_testMaxFPSetSizeRnd`, with
`-tags=tlc_fp_stress -timeout=0`, remains live in native exec session `27326`,
confirmed by polling the handle directly. Its log is
`.codex-gotmp/offheap-random-full-after-outage.log`; the latest saved progress is
1,518,779,649 of 2,147,483,648 iterations, without a terminal result. The previous
handoff incorrectly inferred interruption from process-list visibility. Always
poll the original session before restarting; namespace process lists alone do
not establish that a tool-owned process has stopped. An accidental duplicate,
session `7371`, was stopped with exit 130; its preserved log is
`.codex-gotmp/offheap-random-full-second-recovery.log` and earns no credit.
The earlier interrupted `.codex-gotmp/offheap-random-full.log` ends at 521,200,868
iterations without a terminal result. Preserve original bounds, assertions,
default 64 MiB direct-memory budget and files; do not add race instrumentation.
Full execution and original-method credit remain pending.

The existing native
`TestModelCheckerDoNextEnqueuesOnlyUnseenInModelSuccessorsButChecksAllImpliedActions`
now seeds the initial trace state and counts persistent records through the disk
enumerator after flushing the writer. Its existing enqueue and validation
assertions are preserved; the count requires the seed plus one unseen successor.
Disk-backed traces continue to avoid an in-memory state mirror. This native
harness reconciliation adds no original Java method credit. Details and receipts
are in `PORT_PROGRESS.md`.

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
