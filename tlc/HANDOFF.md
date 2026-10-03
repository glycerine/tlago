# TLC Port Handoff

This handoff is for the next Codex/model after the system upgrade. It records
the active goal, the current working rules, what is already done, what to avoid,
and the best next steps for continuing the Go port of Java TLC.

## Current Snapshot

Priority user directive (2026-10-02): **email reporting is forbidden.** Stop
email reporting and its dependency work immediately; it must consume no more
cycles. JavaMail, SMTP, MIME/Activation, ImageIO/AWT, image codecs and mail-driven
JVM emulation are excluded from TLC completion. Do not continue their pending
constructors, discovery, probes, tests or downloads. This overrides every older
mail-related "next", "pending" or "required" entry in these documents.

The later user instruction authorizes the main thread to remove email-only
source, tests, fixtures, resources and stale plans after detaching integration.
Preserve core distributed behavior, packaged property loading, console output,
generic exceptions, OpenJDK notices and x/text. Then resume core TLC parity and
port existing Java tests after implementing each feature. Prioritize state
counts, diagnostic/trace parity, fairness/liveness, checkpoint/recovery, storage,
coverage and simulation. Do not invent new regression/unit tests.

Latest core slice: existing Java EmptySubsetEqTest and SubsetEqTest are now
translated with their original model fixtures under tlc/test_vectors/models/.
Their end-to-end runs fixed coverage wrapper selection/source filtering,
predefined BOOLEAN values, runtime-module loading and LOCAL postcondition
binding. All four OpApplNodeWrapperTest report methods are translated too.
The existing ACoverageTest is translated too, with byte-identical model data and
expected coverage. SpecProcessor now retains declaration symbols and preserves
their source locations when applying the tool; variable coverage uses semantic
module names rather than physical filenames. Read the progress tail for
verification. CoverageStatisticsTest and CCoverageTest are now translated too:
constraint evaluation counts accepted/rejected states, action validity retains
its cost model, and the parser bridge preserves SANY junction-list and boolean
application nodes for accurate invariant/LET coverage. BCoverageTest,
DCoverageTest, ECoverageTest and FCoverageTest are now translated too.
The bridge now retains module/LET recursion flags and represents LAMBDA as an
operator argument with its own definition and parameter identities; coverage
visits those definitions. GCoverageTest, HCoverageTest, ICoverageTest and
JCoverageTest are now translated too. ENABLED retains its current cost model,
unnamed SPECIFICATION actions retain Java's unnamed marker/location labels, and
named function definitions preserve recursive/nonrecursive specification nodes,
recursive self bindings and complete source ranges. KCoverageTest, LCoverageTest,
MCoverageTest and OCoverageTest are now translated too. Instance definitions
retain SubstInNode bindings, shared source bodies and Subst identities, distinct
instancee declaration symbols and instance declaration ranges. O now reports
separate next-state and invariant substitution counts. Worker generated-state
reporting uses atomic access. Github314CoverageTest, Github377CoverageTest,
Github649CoverageTest and ImpliedCoverageTest are now translated too, completing
the current upstream coverage model testSpec translations. Imported definitions
and native aliases share their source symbol/body identities, so module-scoped
overrides reach their references. LET definitions keep scoped identities and
SANY Context graph order. Config identifiers may start with digits, and square/
angle action evaluation short-circuits as Java does. Full normal/race checks pass.
ConstLevelInvariantTest, PostConditionsTest, PostConditionsFailTest and
ConstantOperatorConfigurationTest are now translated too. Instance conversion
preserves installed native source overrides rather than replacing them with
placeholder definitions. AbstractChecker exposes the source getAllValue indexed
worker collection. The model harness records configuration-time diagnostics and
supports the source coverage-disabled setting. Read the progress tail for checks.
PossibleCountsTest, TLCSetTest, TLCGetNamedUndefinedTest, Github1109Test and
Github1109aTest are now translated too. The bridge retains declared constant
symbols/arities for startup replacement evaluation. Failed nonconstant
replacements are rejected; failed constant expressions remain deferred. Operator
record initialization follows its Java override, and ordinary constant operator
pre-evaluation catches Throwable as upstream does. Tool loading now runs within
TLC.process error handling, including the mode banner and finished message.
Undefined identifiers and assumption diagnostics preserve Java codes/locations.
TLCGetAllTest, TLCGetLevelTest and TLCSetInitTest are now translated too.
The bridge reconstructs the represented SANY module context insertion keys,
including builtin bucket occupancy, to retain Java's variable declaration and
trace order. The DOT/state writer serializes worker writes; coverage counters
use concurrent increments; ordinary next/enabled action lists retain the base
Java behavior rather than init-only action tracking. Worker disk trace writes
no longer read the shared in-memory mirror unnecessarily. The translated
harness preserves the source worker count and provides an existing trace output
directory. Read the progress tail for verification.
TLCSetSimTest and TLCSetMultiSimTest are now translated too. Unary minus
retains SANY's distinct -. operator and Java's Neg override rather than binary
Minus. Their harness preserves debugger-enabled/disabled execution and resets
checker/simulator globals to match upstream classloader isolation. The two
original model/config vectors are byte-identical; the model reaches its own
TLCSet exit at 4,225 generated states with both debugger settings. Current Java
TLC constructs Simulator for both settings despite the older test comment about
SingleThreadedSimulator; do not add that removed variant.
All six original DoInitFunctor checker testSpec methods are now translated:
invariant failure, continuation, no-continuation, minimal error stack, initial
property failure and initial evaluation exception. Their five original models
and configs are byte-identical. The bridge installs Integers.GEQ when the
external Integers module is loaded, including through trace runtime modules;
Naturals.GEQ retains Java's intentional > error label. The model harness also
resets/restores continuation to preserve upstream test-static isolation. Original
record assertions compare the first record's parameter prefix, and uncovered
assertions compare the complete set of zero-count locations. Full normal/race
checks pass. Run Java originals separately with this JDK; combined direct JUnit
leaks continuation because its classloader lacks upstream isolation.
IncompleteNextTest and IncompleteNextMultipleActionsTest are now translated too,
with their four byte-identical model/config vectors. Printed checker traces now
follow TLCTrace.printTrace and ConcurrentTLCTrace.printTrace separately from the
Worker postcondition path: recover the predecessor prefix, print an initial
state directly, and recover a partial successor through equality rather than
fingerprinting it. Source ALIAS prefix/suffix arguments and final/diff-state
printing are preserved. State variable values use Values.ppr, including null
for unassigned variables. Both models match Java's trace, action labels,
diagnostics, state counts and complete uncovered-location sets. Full normal
and race checks pass.
EvalExceptionTest (DistBakery) is now translated too with its byte-identical
original TLA+ file, retaining the embedded config, coverage-disabled setting,
24/17/5 statistics, ERROR exit, six-state trace and exact <= argument error/
nested stack. The bridge preserves Java WF/SF argument order (subscript,
action), retains runtime-added EXTENDS separately from source context imports,
and interns syntax token images in input-file order before converting operators.
Next/init bindings survive evaluation errors as Java exception unwinding does;
unbinding occurs on successful returns. This preserves the final error state.
The source harness starts fresh interner contexts to match per-test classloaders;
worker bootstrap refreshes TLCGetSet and TLCExt class-static names as well as
builtin/counterexample names. Do not remove those refreshes: stale keys make
TLCGet(level) undefined and named _POSSIBLE register lookup fail after reset.
TraceWithLargeSetOfInitialStatesTest is translated too, with its two original
vectors and source -maxSetSize 10 setting, direct initial-state trace/action
assertions and zero uncovered set. The harness restores the set bound. Full normal/race checks pass.
PrintTraceRaceTest is now translated too with four workers and three original
vectors in tlc/test_vectors/models/PrintTraceRace/. The harness distinguishes
fixture directory from root module MC, retaining the source stats, failure
status, two record states/ordinals and complete uncovered-location set.
All four Alias checker testSpec methods are translated too: safety with the
debugger disabled, simulation with num=1, lasso and stuttering. Their three
original model/config vectors are byte-identical. ALIAS record-state printing
now calls Values.ppr for each field, matching multiline TLCGet(action) records
and nested trace tuples. The safety checker case retains register 42 = 4 and
postcondition assertions; liveness cases retain exact Trace prefixes, loop-back
B action and stuttering ordinal 2. Current unchanged Java originals all pass. Full Go normal/race suites pass.
TLCExtTraceTest, TLCExtTraceAliasTest and TLCExtTraceSimTest are now translated
with their two byte-identical original vectors under
tlc/test_vectors/models/TLCExtTrace/. Basic checking retains embedded config,
success, depth 10 and 10/10/0; alias checking retains safety exit, depth 7,
7/7/0 and the exact seven states/actions/ordinals. Simulation retains num=1,
10/1/10/0/0 progress and zero uncovered locations. The unchanged Java originals
pass separately. The bridge now uses SpecProcessor's Tool.AliasSpec rather than
installAliasTarget callbacks. Generic ALIAS evaluation binds the source lazy
trace supplier, catches only EvalException/TLCRuntimeException, appends their
detail message in _ALIASEvalError, and preserves Java value-to-state conversion
and fresh print-state metadata. The native TLCExt.getTrace override has no extra
context lookup. The model harness resets ActionItemListExt.Empty to match its
fresh upstream classloader: its prev/action links otherwise leak an earlier
model's action into the next test. Do not reset it between calls within a source
runtime. All four existing Alias models remain green. Full normal/race suites
and targeted trace checks pass.
EvalExceptionLivenessTest is now translated too, retaining its byte-identical
DistBakery3aAuxMC.tla with embedded DistBakery3aAux/ProtoBakeryTest modules and
config, disabled coverage, ERROR exit, 950/555/89 statistics, exact function
comparison diagnostic, all 15 trace states and source action/ordinal assertions.
The bridge converts Cartesian-product syntax into Java's $CartesianProd node;
synthetic n-ary links flatten within their source range, while parenthesized
operands remain nested. The evaluator already implements this opcode. The three
remaining SetOfTuplesValueTest indexed-sampling methods are translated after
comparing the shared implementation and running all five unchanged Java methods:
empty Nat product, 8,000,000-element product and 64,000,000,000-element product.
Their source seed, overflow exception, enumerator kind, sample/cardinality,
HashSet-style deduplication and membership assertions are retained. Full normal
and race suites pass.
ValueSemanticsAssumeTest is now translated too, with its three byte-identical
source vectors and original no-debugger/noGenerateSpec/JSON trace-dump settings.
Go passes all
489 unchanged ASSUME clauses and matches Java success/0/0/0. Source-definition
conversion is separate from instance-export routing; each source definition
retains its own symbol, export aliases share one cached clone, and source-body
references keep source identities. The bridge emits AtNode for EXCEPT @. Native
FiniteSets IsFiniteSet/Cardinality overrides carry their source reflection
signatures, and located semantic nodes render their Java source locations in
evaluation errors. Other native signatures remain a broader metadata task.
DepthFirstErrorTraceTest is translated with its two byte-identical vectors,
-dfid 9, JSON trace dumping, safety exit, eight exact trimmed states, empty action labels, trace
ordinals and complete zero-uncovered assertion. Both unchanged Java originals
pass; targeted checks and full normal/race suites pass.
CyclicRedefineInstance/Init/Next/Op/SubAction/VarsTest are now translated after
porting their source features. LET-local INSTANCE exports have lexical symbols
and attached definitions. Source semantic identities are retained separately
from evaluation overrides; graph traversal, isDefinedWith's SubstIn rejection
and delayed application substitution implement allowCyclicRedefinitions.
Root definitions retain source context order before config overrides; config
rewrites precede constant evaluation. Formal parameters get distinct symbols
and bare operator arguments become OpArgNodes. Ordinary overrides consult the
original OpDef tool object; module overrides retain existing symbol bindings.
Definition-table indices are no longer reset after installation, which had
allowed a deferred override to overwrite TRUE's slot. The six translations
preserve disabled debugger/coverage/JSON trace dumping/generated trace spec,
retained DOT dumping, exact source exits/depth/statistics and both safety traces.
Their nine source vectors are byte-identical. All six unchanged Java originals,
targeted Go checks and full normal/race suites pass.
DepthFirstDieHardTest is translated with its two byte-identical vectors and
original -dfid 7/debugger/coverage/DOT/JSON trace-dump/generated-spec settings.
The seven exact states, empty action labels, ordinals, Finished/no GENERAL/no
STATE_PRINT1 and zero-uncovered assertions pass. Corrected DFIDWorker's early
return after semantic violations: Java finishes its inner depth loop, checking
stopCode only in the outer initial-state loop; thrown errors still leave run
immediately. Go now matches Java's 876/68 production counts as well as the trace.
The original test deliberately has no count assertions. FPIntSet statics are
initialized for each translated-test runtime, matching Java's fresh classloader;
retained levels had made the two DFID tests loop together. Both translated DFID
checks and full normal/race suites pass.
Native body lookup through SubstIn is now ported for represented operators.
Native implementations are captured before alias/source registration, then
attached to original definition bodies as Java processModuleOverrides does.
INSTANCE clones reuse those bodies. Module-specific config overrides update
the original shared body rather than just the definition table. Synthetic
ValueNode intrinsic values remain separate from explicit body tool metadata;
the existing operator-name clash assertion remains unchanged and passes.
After implementation, translated DumpLoadTraceTest's four single-worker safety
and bidirectional-liveness JSON/TLC methods. Each phase starts a fresh runtime,
preserving fp 4, exact source flags, file existence/nonempty checks, Finished,
exit and violation equality, trace ordinals and trimmed state equality. The
harness now accepts exact source arguments and omits forced generation when
noGenerateSpec applies; otherwise the binary dump was redirected to the
generated trace-spec file. DieHard uses the existing original vectors; the two
BidirectionalTransitions vectors are byte-identical. All four unchanged Java
original methods, targeted normal/race and full normal/race suites pass.
Production replay counts match Java 36/7/0 (safety) and 21/5/0 (liveness).
The four original DumpLoadTraceTest auto-worker safety/bidirectional methods
are now translated too, retaining -workers auto for dump, one worker for load
and the source prefix comparison with exact indexed states and ordinals. Those
runs exposed missing synchronized(oos) in Go's modern liveness graph path.
OrderOfSolution now owns the graph monitor; disk/tableau mutations and the Go
in-memory graph mirror share it. Tableau consistency is computed outside the
monitor and reused, while fingerprint-prefix recovery stays inside and state
regeneration/printing happens after release, matching Java's lock boundary.
All four unchanged Java auto-worker methods and all eight Go dump/load methods
pass; targeted existing liveness and the formerly remote-failing mixed config
set test pass with the race detector. Full normal/race suites pass.
Next: compare and translate remaining ALIAS/CodePlex/TESpec/Example1 dump/load
methods. The unchanged single-worker AliasSub JSON method and Go production
replay already match safety exit, seven common-variable states and 132/32/9;
translate the source intersection assertions, retaining separate dump/load
specs/configs and original vectors. Preserve the upstream Ignore annotations
for garbled EWD840 JSON and the two AliasSub auto-worker methods (AliasSub2 and
AliasSup methods remain enabled upstream).
Complete semantic module graphs and constant-processing snapshot/eligibility
metadata remain broader source work.
Trace-expression variants remain separate pending
work. Record StateString's explicit empty/non-string _format handling and full
Java String.format semantics also remain source parity work. TLCGetNonDeterminismTest
is ignored upstream by design; preserve that status when translating it. Complete SANY contexts, nested export
composition and instantiation-aware constant-processing eligibility remain
bridge parity work; the represented declaration-order slice does not establish
their completion. Original model bytes, including upstream whitespace, are
intentionally retained in test_vectors/.

- Repository: `/mnt/oldrog/home/jaten/go/src/github.com/tlaplus/tlago`.
- Java source of truth: `../tlaplus/tlatools/org.lamport.tlatools/src/tlc2`.
- Java tests to port after their features: `../tlaplus/tlatools/org.lamport.tlatools/test/tlc2`.
- Active work area: `tlc/`, package `github.com/glycerine/tlago/tlc`.
- Current branch had a clean worktree when this handoff was written.
- Last verified command before handoff:
  `env GOCACHE=/mnt/oldrog/home/jaten/go/src/github.com/tlaplus/tlago/.codex-gocache GOTMPDIR=/mnt/oldrog/home/jaten/go/src/github.com/tlaplus/tlago/.codex-gotmp go test ./...`
- Recent commits immediately before handoff:
  - `3763989 Preserve Java subseteq shortcuts`
  - `9ee99dd Scope constraint metadata by tool id`
  - `f933188 Scope TLCEval cache by tool id`
  - `489fedc Preserve parameterless TLC error codes`
  - `9181f3c Scope TLCCache by tool id`
  - `e52c5cf Mirror DiskFPSet striped locking`

## Active Goal

Continue the mechanical, breadth-first port of the Java TLC model checker to Go.
Mirror each Java feature and algorithm first, then port the Java tests for
that feature when they exist. The user updated the test-porting instruction
to this feature-by-feature sequence.

Do not switch back to the older SANY XML or ApalacheIR corpus sweeps unless the
user explicitly asks. Those are valuable, but they are paused. The current goal
is TLC.

## Operating Rules

- Keep implementation mostly in package `tlc`; avoid splitting into subpackages
  unless there is a very strong reason.
- Prefer concrete structs over interfaces, especially where Java has an
  interface or abstract class with only one meaningful production implementation.
- Use `InsMap` whenever iteration order could affect output, diagnostics,
  state exploration, fingerprints, coverage, or tests. Plain Go maps are fine
  only for lookup-only sets that are never ranged over in observable code.
- Preserve Java behavior, including load-bearing quirks. Do not "clean up"
  oddities unless the user explicitly chooses a deliberate divergence.
- Implement each Java feature accurately in Go first. Once that feature is
  ported, port its Java tests too when they exist. This is the user's latest
  instruction and supersedes the earlier instruction to defer all new tests.
- Existing fast tests may be run frequently. Use:
  `env GOCACHE=/mnt/oldrog/home/jaten/go/src/github.com/tlaplus/tlago/.codex-gocache GOTMPDIR=/mnt/oldrog/home/jaten/go/src/github.com/tlaplus/tlago/.codex-gotmp go test ./tlc`
  and, before commits, usually `go test ./...` with the same env.
- Update `tlc/PORT_PROGRESS.md` before each coherent TLC commit. It is the
  shared memory for what has been audited and what must not be revisited.
- Make regular commits after coherent chunks. Keep commit messages short.

## Documentation Map

- `PLAN.md`: broad project plan. It still contains SANY and Apalache history,
  but the current active section points to this TLC handoff.
- `ARCH.md`: detailed Java SANY architecture notes.
- `tlc/TLC_ARCH.md`: detailed Java TLC architecture, APIs, data structures,
  algorithms, performance notes, and the original mechanical port order.
- `tlc/PORT_PROGRESS.md`: living audit log and do-not-revisit ledger. Read the
  top notes and tail before starting a new audit pass.
- `tlc/HANDOFF.md`: this file, intended as the quick restart guide.

## Current Implementation State

The Go TLC port is broad and no longer skeletal. It contains concrete ports for:

- Runner/CLI option parsing and Java-shaped runtime properties.
- Model config parsing and config diagnostics.
- Spec processing over the production Go SANY bridge.
- Tool evaluation, enabledness, init/next-state generation, action metadata,
  lazy values, call-stack replay hooks, and many Java diagnostics.
- TLC state, state vectors, trace records, state writers, trace reconstruction,
  aliases, and counterexample records.
- Primitive/composite/lazy/operator values, set constructors, model values,
  fingerprints, value streams, and many Java comparison/enumeration quirks.
- Standard modules and CommunityModules overrides represented by concrete Go
  registration functions.
- ModelChecker, Worker, DFID checker/worker, simulator and simulation workers.
- Coverage cost models, TLCGet/TLCSet, TLCExt, TLCEval, JSON/trace modules, and
  `_Possible`.
- In-memory and disk queues, byte-array queues, state pools, buffered random
  access files, object/int stacks and queues.
- Memory, disk, multi, distributed, and off-heap fingerprint set machinery.
- Liveness expression processing, tableau/behavior graphs, live workers,
  disk graphs, debug DOT snapshots, and liveness counterexample reconstruction.
- Debugger/presentation helpers, model presentation structs, pretty-printing,
  and management/MX-style wrappers.

The code is not declared done. The immediate purpose is still Java-parity audit
and breadth-first correction. Port each feature's existing Java tests after its
implementation, following the user's updated sequence.

## Recently Audited Areas To Avoid Repeating

`tlc/PORT_PROGRESS.md` is the canonical list, but these are especially fresh:

- `TLCExt!TLCCache`, `TLC!TLCEval`, and `TLCGet("spec")` constraint metadata
  must be scoped by `(Tool.ID, SemanticNode UID)`, matching Java
  `getToolObject(toolId)`. Do not collapse these into shared
  `SemanticNode.ToolObject`.
- `TLCGet("spec")` constraints start as per-tool `OpDefNode` metadata from
  `SpecProcessor`; coverage later replaces the same per-tool slot with the
  constraint `Action`.
- Expression-level `S \subseteq T` now preserves Java's specialized
  `IntervalValue.isSubsetEq` and `SubsetValue.isSubsetEq` shortcut rewrites.
- Disk FP sets use Java-like striped locking, including the non-reentrant Go
  workaround for Java's reentrant write-lock flush shape.
- Dot writer filename derivation uses Java `String.replace(".dot", ...)`
  semantics, not suffix-only replacement.
- Many standard module override inventories have been checked. Do not register
  helper-only or commented-out Java methods as native exports.
- Simulator result/error classification, action-flow reduction, and RL property
  parsing have been rechecked recently.
- Liveness SCC/postfix/counterexample paths, aliasing overloads, tableau graph
  storage, and LiveCheck graph reset behavior have been rechecked recently.
- Config diagnostics, parameterless TLC error codes, checker cleanup/result
  precedence, and worker trace reconstruction have been rechecked recently.

If you touch any of these areas, re-read the relevant `PORT_PROGRESS.md` notes
first.

## Immediate Next Steps

Email integration and its dependency code have been removed under the user's
explicit authorization. Continue core TLC work; do not restore reporting.

Start by reading the top and tail of `tlc/PORT_PROGRESS.md`, then continue the
breadth-first Java source audit from areas that are not marked recently audited.
Good next slices are:

1. Refresh the remaining checker error-precedence and trace reconstruction
   audit outside the already-covered cleanup/no-action/DFID/simulator/init
   exception paths.
2. Continue comparing `tlc2/tool/impl/Tool.java`,
   `ModelChecker.java`, `Worker.java`, `Simulator.java`, and liveness support
   classes against the Go files with the same responsibility. Patch true gaps.
3. Continue the value/module audit only in parts not already marked
   do-not-loop in `PORT_PROGRESS.md`. If a Java subclass overrides a method,
   make sure the Go central helper preserves that subclass behavior.
4. Keep checking map iteration boundaries. If output, state order, diagnostic
   order, or fingerprint order can observe an iteration, use `InsMap`, slices,
   or explicit sorting.
5. Keep adding short `PORT_PROGRESS.md` notes for no-code audits. The main risk
   now is rediscovery and accidental divergence, not lack of raw code volume.

## How To Audit A Java Slice

Use this rhythm:

1. Pick a Java class or cluster from `src/tlc2`.
2. Read the Java function decomposition first.
3. Find the Go files that represent the same responsibility.
4. Compare control flow, mutation order, error precedence, and special cases.
5. Patch only real behavioral gaps.
6. Update `PORT_PROGRESS.md` with either the fix or a no-code audit note.
7. Run `go test ./tlc`; run `go test ./...` before committing.
8. Commit a coherent chunk.

Favor precise Java-facing behavior over general cleanup. The Go port can look
slightly less idiomatic when that makes source-of-truth comparison simpler.

## Things Not To Do Yet

- Do not invent new unit/regression tests. Translate existing Java tests after
  their corresponding features are implemented.
- Do not resume long SANY XML or Apalache sweeps without explicit instruction.
- Do not introduce compatibility shims or a second parser.
- Do not replace Java quirks with nicer Go behavior unless the user explicitly
  agrees to diverge.
- Do not use Go interfaces just because Java used interfaces.
- Do not use randomized Go map iteration in any user-visible or semantic path.
- Do not rely on external repo locations for frozen test data.
- Store persistent fixtures in `tlc/test_vectors/`, never `tlc/testdata/`.
  The user reserves `testdata/` for ephemeral Go fuzzer storage and cleanup;
  keep this naming rule when porting any further Java test vectors.

## Completion Definition For This Phase

Email reporting and all dependencies pursued for it are excluded by the user's
scope correction. Their unfinished work cannot prevent this phase's completion.

This phase is complete when the Go code has a coherent, faithful mirror of the
Java TLC architecture and behavior surfaces, with `PORT_PROGRESS.md` indicating
no major unaudited core areas remain. Translate existing Java tests
feature by feature during this phase, after their implementations are ported,
as the user requested; do not defer all tests until the entire port is complete.
