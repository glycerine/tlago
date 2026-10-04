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
External binary trace deserialization now calls ReadExternal, matching Java's
read(internTbl.toMap()) delegation. Records loaded into a different intern table
must normalize against its field-name order; ordinary Read had made binary
ALIAS replay report a missing counterexample field that was visibly present.
Modern disk liveness checkers now use their disk graph alone as Java does;
they no longer evaluate source/target predicates again for an extra memory
mirror. The memory path remains for checkers without disk storage.
DumpLoadTraceTest now has 33 of its 35 original methods translated: 30 enabled
methods and the three original @Ignore skips. Ten enabled DieHard ALIAS cases
retain separate dump/load specs/configs, nonempty common-variable comparisons,
ordinals and source equality/prefix assertions. Four each TESpec EqAlias,
Example1 and MC cases retain exact source flags and worker counts. Their 16
copied model/config vectors are byte-identical. All enabled unchanged Java
methods pass normally; targeted Go race and full normal/race suites pass.
The two enabled EWD840 binary translations remain pending. Their ordinary Go
production runs match Java's eight states and 15986/1566/0 then 7198/677/0,
but race instrumentation triggers partial liveness checks and a longer replay.
Pacing the unchanged Java test reproduces the same assertion failure: an
eight-state dump replays as 15 or 19 states. Beyond the dumped prefix the
constraint admits other states and another cycle can be selected. Do not
weaken its equality/prefix assertions, invent a skip or disable partial checks
to hide this source timing limitation. The source garbled-JSON @Ignore remains
ported; these two enabled methods have not been accepted as test ports.
TraceExpressionSpecSafetyBFSTest, TraceExpressionSpecSafetySimTest and
TraceExpressionSpecRuntimeTest are now translated after comparing the unchanged
Java originals and production Go generated tools. Their inherited assertions
retain one action/invariant/init, the next relation, good/in-model/valid states,
the four x/y safety states or five runtime-error records, ALIAS and disabled
deadlock checking. Original safety/runtime configs and runtime model are copied
byte-for-byte under test_vectors/models/TESpecTest/. Generation and replay use
the same interner, the source debugger/noGenerateSpecTEBin flags and metadir;
the resolver includes both the original user directory and generated directory.
The ModelCheckerTestCase harness now scopes/restores Double.MAX_VALUE for its
liveness threshold. DumpLoadTraceTest retains its default periodic checks.
Full normal and race suites pass, along with targeted generated/replay checks.
TraceExpressionSpecDeadlockTest and TraceExpressionSpecLassoTest are now
translated after retaining the represented semantic module graph and contexts.
SpecProcessor exposes its actual ExternalModuleTable and root ModuleNode. The
bridge preserves dependency order, builtin/operator/declaration identities,
context Pair and Hashtable orders, EXTENDS relationships, instantiated flags,
LOCAL instance definitions and shared bodies. Java creates distinct LOCAL
parameter-free instance definitions sharing their source bodies; applications
in the owning module now bind to those definitions. On the same Java-generated
monoliths, all module/declaration/operator orders match for 11 deadlock and 12
lasso modules, as do all 334 and 395 operator/body identity relationships.
The original deadlock test retains all four good/in-model/invariant-valid states,
ALIAS and three module lookups; its commented deadlock TODO remains unasserted.
The lasso retains default periodic checks, one property/no invariants, all three
0/FALSE -> 1/TRUE -> 0/FALSE states, module lookups and LiveCheck1 failure.
Three additional source vectors are byte-identical under test_vectors/.
Periodic liveness reads graph size before suspending workers; publish the two
node-pointer-table primitive int counts atomically while retaining the existing
solution monitor. Constant pre-evaluation now requires lookup to return an
OpDefNode, matching Java and preserving native/value overrides rather than
evaluating their TLA+ placeholder bodies. Full normal and race suites pass.
SpecProcessor now processes constants by actual module and declaration/operator
identity, retaining WorkerValue storage and Java's instantiation eligibility,
source origins, inner-module recursion and immediate node updates. Its global
entry is replaced only when it still denotes the exact evaluated operator.
Snapshot is captured after config overrides and before pre-evaluation; symmetry,
_RL_REWARD and _PERIODIC use that unprocessed snapshot. Root registration retains
transitive EXTENDS exports, source contexts retain original nodes rather than
colliding named INSTANCE exports, and native overrides follow module order.
TLCGetAndSet remains a TLA+ definition as in the source TLCExt module. Debugger
constants use the module map, worker mux and retained compound instance names.
The REPL reads the actual root operator instead of rebuilding a second graph.
Across eight unchanged models, Java/Go comparisons match 24 constant records,
941 operator-origin records and 941 unprocessed/current snapshot records.
After implementation, translated Github361Test with its original two workers,
Finished, 2/1/0 statistics and one initial state; its two vectors are byte-identical
under test_vectors/. The unchanged Java JUnit method passes too.
Validation of this slice is recorded at the end of PORT_PROGRESS.md.

Debugger workers now halt on the debugger monitor and resume on actual command
notifications, releasing the monitor for stack/variable requests. Retain source
factory override and suspend/nohalt defaults, exit-frame stops/presentation,
state-selection granularity, virtual-frame cleanup and reset control flow.
Variable rendering runs in DebugEvalDebugger mode to prevent recursive debugger
entry while displaying lazy function constants. DAP frame IDs combine actual
semantic UIDs with random IDs, preserving repeated-node frames. Root assumptions
include direct extendees' inherited vectors in source order (including repeated
paths), share source expression identities, exclude INSTANCE imports, and retain
AXIOM flags. Runtime string constants are installed on their actual declarations
before constant-map processing.
Both complete unchanged EchoDebuggerTest and EWD998ChanDebuggerTest pass in
separate original Java runs. The production Go Echo run matches its first nine
source stops: all 29 frame locations/contexts, three groups and nine constants,
then finishes with success under race instrumentation. This is a source probe,
not a partial persistent translation of either complete test. After this core
implementation, translated all 17 original TLCDebuggerTest pagination methods;
unchanged Java originals and targeted Go race verification pass.

The production SANY syntax nodes now retain source depth and parent links;
central expressions/definitions and TLC semantic nodes preserve actual syntax
identity. Step-over/out compare syntax depth and owning N_OperatorDefinition
identity. Retain one-item junction list/item nodes and one semantic quantifier
for each source variable/bound list. Generated Init state frames reuse the active
tool/node/context. The concrete debugger frame union preserves state/action
breakpoint conditions and hit counts, ancestor suppression, and action/next
trace getS/addT behavior. Debugger supplier evaluation also swaps the state's
fingerprint tool to FastTool and restores it, matching Java. VIEW evaluation uses
Java's State-mode overload, and workers reuse a successor's fingerprint when
collecting states instead of evaluating VIEW twice.
The full unchanged Echo command sequence now matches all 30 stops, 221 frame
locations/contexts and 221 syntax-depth/operator-owner records under the race
detector, plus the original constant groups/entries. This remains an ephemeral
source comparison, not the complete persistent Echo test translation.

Variable trace display now reconstructs the disk prefix and source in-memory
suffix, retains action addT=false versus next addT=true, and filters incomplete
simulation states. Next simulation traces keep padded names and selection IDs;
model-checking trace display does not add simulation-only selection IDs. State
and action variable rendering runs through the debugger supplier boundary,
remembers nested references, and shows Java's interleaved unprimed/primed action
record fields (including the trailing space), pending-value text/type and
fingerprint fallback. Context lazy values evaluate with concrete frame states
without updating their own caches. FastTool retains its original tool mode and
remains the parameterless state-fingerprint tool, matching Java construction.
Workers still fingerprint explicitly with the active debug tool.
After these features, the complete EchoDebuggerTest.testSpec is translated with
all source assertions and used inherited frame/context/state/trace/successor
helpers. Its four upstream vectors are byte-identical under test_vectors/models/
Echo. Unchanged Java and Go normal/race executions pass. The expanded source
probe matches all 30 stops/221 frames plus 696 scopes, 137 context variables,
158 state-variable records and 838 trace-variable records. Fingerprint type
presence matches; raw fingerprint numbers depend on each runtime's intern-token
namespace (observed NoNode token Java=1, Go=127) and are not claimed equal.
Debugger stack variables now use SyntaxTreeNode.getHumanReadableImage and
source Variable equality, preserving source order and nested references.
Concrete frames expose nullable expression requests/results and getWatch with the
source state/action/synthetic overloads, supplier mode, exception boundaries,
parameter-name lookup, nested expansion and TLCExt CounterExample context.
The expression compiler now follows the retained semantic child path, emits
LOCAL stubs and reconnects them to actual LET operator/symbol identities rather
than reconverting source definitions. Syntax, semantic and represented level
errors retain Java messages and locations. The path traversal enters its root
without preemption, matching Java walkChildren; preemption applies to children.
Module top-level assumptions retain actual shared evaluator expressions and
inherited source order. Broader top-level nodes/proofs remain pending.
After the production work, the complete original ExpressionBreakpointTest is
translated, including its embedded config, hit/column/condition, state value,
context map and lazy-cache assertions. Its one original vector is byte-identical
under test_vectors/models/ExpressionBreakpointTest/. Unchanged Java and targeted
Go tests pass. Final full offline go test ./... and go test -race ./... pass
(root 32.759s/197.687s; TLC 0.862s/3.425s). The live EWD998 source/Go probe
matches eight stops, 52 frame
locations, 211 expression responses, 156 getWatch responses and 21 stack-variable
records, and finishes successfully under race instrumentation. This is an
ephemeral comparison, not a partial persistent EWD998 test.
Hover now follows actual semantic paths and formal identities, with concrete
base/state/action handling of primed variables, pending assignments, lazy values,
record fields, source types and nested references. Ordinary operator/function,
quantifier, CHOOSE, LAMBDA and comprehension parameters retain distinct formal
symbols and actual parser locations/syntax; all bounded domains convert before
installing any bound symbol, matching Generator.processQuantBoundArgs. Parameterized
INSTANCE/proof metadata still requires broader source work. Lookup retains an
operator's actual definition before a same-name alias, preserving source versus
INSTANCE identities and removing an extra refinement substitution frame.
Protocol Evaluate dispatch now retains hover/variables/repl/watch/clipboard,
nullable results and source monitor/frame lookup behavior. Hover URI parsing
checks Java path/query/fragment character masks, UTF-16 escapes, source coordinate
splitting and exception boundaries. Lazy display catches depend on the concrete
frame type, including base-only NullPointerException handling.
Breakpoint verification walks the selected ModuleNode's actual syntax children,
retaining root visitation without preemption and fuzzy child range inclusion.
INSTANCE modules retain their own assumptions for location lookup, while checker
assumption multiplicity remains restricted to the root EXTENDS closure.
The corrected ephemeral Java/Go hover probe matches five stops, all 79 frame
locations, 120 semantic paths, 16 hover responses, all 171 original line-verification
results and 21 URI response/exception cases; Go finishes successfully under race
instrumentation. Empty and malformed-syntax condition results match too. The
missing-dependency condition LET T == INSTANCE DoesNotExist IN T!YOLO now
reports Java's exact located semantic error. Conditions use Java isBlank and the
selected module's graph, preserving exact-name operator lookup without trimming
or a global-name shortcut; frame expression evaluation retains the processor root.
Debugger dependency loading follows the actual parser dependency list, including
completed internal-module names, recursively reusing the live external table.
Successful dependencies persist across later expression failures; temporary
wrapper modules stay out of that table and do not replace its root. Existing
source/config/native identities survive incremental context construction.
The separate preConstantSnapshot now contains TRUE/FALSE/BOOLEAN and the native
Strings.STRING MethodValue before ordinary definitions or config bindings.
Source processConstants traverses represented modules, operator/LET/label bodies,
substitutions, assumptions, bounded expressions and operator arguments, retaining
an identity set for processed definitions. Initial and dynamic modules use this
snapshot separately from constant pre-evaluation. Numerals retain their original
radix/image and big-integer metadata; large integers and decimals fail during
constant processing with Java messages. Dynamic wrappers run constant processing
and represented native module overrides before reconnecting LOCAL stubs through
ModuleNode graph substitution. Integers' own GEQ override is available when first
loaded dynamically, including its inherited Naturals operators.
Final ephemeral source/Go comparisons match all 419 hover/breakpoint records,
448 expression/watch/stack records and 60 dependency records under race
instrumentation. Dependency checks retain Bags after a failed expression, reuse
its overrides, load Randomization and preserve the root/transient-wrapper rules.
After production work, the entire original EWD998ChanDebuggerTest.testSpec is
translated, including the source Set<Variable> inherited lazy-cache overload.
All 154 equality assertions, truth/false/reference assertions and 5 base-frame,
21 state-frame, 7 action-frame and 1 next-state-frame calls are retained. The
source response non-null assertion is guaranteed by Go's concrete response type.
All four EWD998 model/config vectors are byte-identical to Java's originals under
test_vectors/models/EWD998. No invented tests or vectors were added.
The whole method exposed and fixed two production gaps: context maps/display
names use source declaration names rather than qualified evaluator lookup keys;
semantic context bindings use Java signatures and full human-readable operator
definitions, including attached comments and one-child spacing, without a type.
State constraints invoke the source one-state eval overload, selecting State
mode and the current state's level for hit-count breakpoints. Action constraints
retain the two-state overload. The inherited helper also preserves a typed nil
expected context as Java's null, without comparing it against the actual context.
Unchanged Java JUnit reports one run/zero failures/zero ignored. The complete Go
method and full offline normal/race suites pass, alongside Echo and
ExpressionBreakpoint. The original EWD998TraceDebuggerTest production comparison
also matches: the _TETrace hover returns the exact two-state tuple, TupleValue
type and nested reference, and the run ends with liveness violation exit 13.
After verifying those features, translated its entire method, retaining both
equality assertions and the nonzero-reference assertion, constructor config
arguments and expected exit status. The shared source harness now accepts each
original test's expected exit, instead of assuming success; existing tests still
expect success. One more byte-identical vector, EWD998_TTrace.tla, contains its
embedded config and three original modules. Unchanged Java JUnit reports one
run/zero failures/zero ignored; full normal Go and all four complete debugger
model tests under race instrumentation pass. No production correction was
necessary for this trace test; no invented tests or vectors were added.
The original EWD840DebuggerTest production comparison now matches all 624
frame/context/state records over 37 stops and ends with Java's safety exit 12,
127 generated states, 38 distinct states, queue 3 and depth 6. Initialization
suppresses debugger stops until a checker/simulator exists. Source definition
symbols retain actual declaration names, syntax and locations; lazy semantic
images use Java's SemanticNode.toString rather than evaluator lookup names.
Unsupported expression evaluation retains the detailed runtime exception's
expression/context and located message. Initial and next-state invariant control
exceptions are distinct fresh objects; next-state exceptions retain their own
known flag. Worker wrapping exceptions keep explicit expression/state accessors
without a cause/message. Debugger catches inspect the directly thrown class and
use the source base/action frame overloads, removing a false stop while unwinding.
After production parity, translated the entire EWD840DebuggerTest.testSpec:
35 equality, one true and one non-null assertion, 3 base-frame, 7 state-frame,
34 action-frame, 1 init-frame and 1 next-frame call, including both source loops.
The inherited init-frame helper preserves the source's unimplemented successor
count assertion; continue_(steps) retains the original alias stepping sequence.
All three original vectors are byte-identical under test_vectors/models/EWD840.
Unchanged Java JUnit and the complete Go method pass; final full offline normal
and race suites pass. No invented tests or vectors were added.
Both original EWD840 error debugger methods are now completely translated.
The initial-state comparison matches 31 frame/exception/context/state records;
the action-error comparison matches 61 records, both under race instrumentation.
Legacy exception variables retain source human-readable locations, nullable
throwable detail messages, Java simple class names and no nested reference.
Action.UNKNOWN retains SemanticNode.nullSN's builtin location, syntax image and
minimum-integer kind. Parameterized LET definitions do not enter the evaluation
context; the source zero-arity lazy bindings remain. The harness checks Java's
pending actualExitStatus sentinel -1 before resuming for cleanup, since both
original methods finish while TLC is paused. It refreshes IntValue's cached
statics as the source per-test classloader does; prior CallStackTool source
metadata otherwise moves the action exception stop up to SelectSeq. Source
metadata reads/writes are synchronized to preserve Java reference atomicity when
worker-local lazy values attach a source to shared cached integer values.
EWD840ErrorDebuggerTest retains all 7 equality, 6 null and 1 non-null assertions,
1 base-frame and 7 state-frame calls. EWD840ErrorActionDebuggerTest retains all
7 equality assertions, the null-exception loop, 1 non-null assertion and its
complete action-frame assertion. Error02.tla and Error03.tla are byte-identical
original vectors in test_vectors/models/EWD840. Unchanged Java JUnit methods,
both complete Go translations and full offline normal/race suites pass.
No invented tests or vectors were added.
The entire original EWD840DebuggerSimTest.testSpec is translated too after a
production comparison matching all 911 frame/exception/context/state records
across 39 stops under race instrumentation, 35 generated states, one trace and
safety exit 12. SimulationWorkerError now extends the source invariant exception,
initializes its own known flag through its constructor and displays its formatted
error-code/parameter message. Catch dispatch recognizes this actual subclass,
restoring the invariant halt before alias evaluation and preventing duplicate
handling while unwinding. Every simulation error construction path uses the
same initialized source class; its stored exception remains separate from cause.
The whole test retains all 38 equality, 3 true and 1 non-null assertion, 3 base-frame,
7 state-frame and 27 action-frame calls, all three loops, conditional level > 3
spec breakpoint, synthetic trace levels, exact Stop invariant message, constructor
config/seed/fingerprint/simulation arguments and safety exit. One more original
vector, MC02Sim.tla, is byte-identical under test_vectors/models/EWD840. Unchanged
Java JUnit, the whole Go method and full offline normal/race suites pass. No
invented tests or vectors were added. The intermittent DieHard auto-worker
binary trace prefix mismatch is now reproduced by the unchanged original Java
method under temporary worker scheduling delays: exactly the same empty-big vs
pour-big-to-small state at index 6. Both sources check invariants on rejected
successors, so a nonminimal dump can replay to an earlier off-trace violation.
Preserve source behavior and all original assertions; do not add retries or
weaken the comparison. Read the progress tail for the scheduling evidence.
The whole original Debug03Test and Debug03SimTest methods are now translated
after production comparisons matching 17 checker and 572 simulation records.
Next-state and synthetic frames retain source locations, contexts, nine sorted
successors, trace variables and selection references. Simulation keeps all four
loops, forward x=0..8 selections, backward navigation, stepping out to the two
initial states and selecting the second initial state. Source assertions remain
complete: checker 1 equality/1 next-frame/1 synthetic-frame call; simulation
11 equality/2 true/2 init-frame/3 next-frame calls. Debug03.tla is byte-identical
under test_vectors/models/debug. Both Java originals and complete Go methods
pass; checker counts are 92/10/0/depth 2, and simulation generates 1,055 states
and two traces.
The complete original Debug02Test.testSpec is translated after matching all 24
hover/frame/state records under -race. Existing production behavior retains exact
semantic source ranges, pending assignments, TRUE/FALSE state values, nullable
type responses and source type strings while stepping through state/action/next
frames. The original 51 equality, 8 true and 4 false assertions, constant module
view and single next-frame helper call remain intact. Its embedded-config vector
Debug02.tla is byte-identical under test_vectors/models/debug. Unchanged Java and
the complete Go method pass; source production counts are 3/2/0/depth 2. The
source hover harness constructs the absolute module URI, symbol query, source
coordinate fragment and top-frame selection.
The complete original Debug04SimTest.testSpec is translated after the full
production sequence was checked under -race. All 1,410 observation records have
matching structure: 1,288 match exactly; 122 reflect different successors tied
for maximum Hamming distance. Java's stepOver selects from HashSet iteration,
and its test requires consecutive x values to differ. Each runtime's trace,
expression results and fingerprint labels agree with its chosen states; both
complete with 543 generated states and four traces. Preserve that source
assertNotEquals rather than imposing a particular tied successor. Existing
production code required no correction. The whole test retains 34 equality,
2 array equality, 16 true and 1 not-equal assertion sites, all five loops,
3 init-frame/9 next-frame calls, all 215 expression evaluations, both idempotence
checks and all conditional/unconditional breakpoint commands. Shared stepOut
supports the original count overload. Debug04.tla is byte-identical under
test_vectors/models/debug. The unchanged Java original and complete Go method
pass.
Debug05SimTest's production features and complete original method are now
ported. Named LET instances with WITH substitutions stop at the enclosing IN;
embedded standard modules retain declared constant substitutions; source TLA+
operator bindings keep their instantiated SubstIn body; and a dynamically
imported declaration reuses its original module-context identity. CounterExample
lookup reads the live shared SpecProcessor definitions, matching Java Spec and
avoiding the fast debugger tool's stale copy after TLCExt import. The whole
production sequence matches all 950 frame/context/state/variable/expression
records, including 21 stops and 63 evaluations, under -race. All nineteen JSON
payloads match. Java and Go readers each decode all 38 Java/Go binary exports to
the same original trace payloads; raw streams contain runtime-specific string
tokens/record ordering. Both runtimes finish successfully with 51 generated
states/one trace/depth 25/seed 1/aril 0. The complete Go test retains every source
assertion (8 equality/1 true sites), all six module expressions and the full
nineteen-iteration export/readback loop. Source doDumpTrace=false and BASE_DIR
resolver overrides are retained; runtime files use an isolated temporary
directory. Debug05.tla is byte-identical under test_vectors/models/debug.
GetScopedIdentifiersTests and DebugTLCVariableTest are now translated in full.
The source scoped-symbol helper retains symbol identity across LET definitions,
operator parameters, quantified variables and LAMBDA arguments; generated
signatures preserve operator arity. Debugger expression construction uses that
helper and removes LET names before formatting its parameter set, matching the
source's LOCAL stub handling. All eighteen original scope records match Java
exactly, including the upstream known infix-parameter quirk. All four original
nested-variable methods match nine expansion/five child records, including
names, types, values and expandability. Existing production nested-value code
needed no change. The whole Go translations preserve all four assertions per
scope case, all eight nested-variable equality sites and the two-child loop;
source inline inputs are unchanged and no invented vectors/tests were added.
Both unchanged Java classes pass (18 and 4 cases/methods respectively), as do
the whole Go translations and existing debugger expression tests under -race.
RecordValue.toState/StateString now preserves Java's format selection: absent
_format selects the default, an explicit empty format suppresses output, and
non-string/null format values throw their source exceptions. StringValue's
DebuggerValue subtype is accepted. Duplicate record fields bind in source order,
so the last matching field supplies the state variable. State rendering and CSV
now share Java's formatter behavior for their actual string arguments, including
whole-template validation, indexing/reuse, flags, width, UTF-16 precision,
boolean/hash conversion, uppercase and typed formatting failures. CSV writes the
source UTF-8 bytes and line separator after successful formatting. The shared
string printer preserves supplementary and unpaired surrogate units rather than
replacing an emoji with two replacement characters. All 60,551 template records,
all Unicode scalar uppercase mappings, all 65,536 code-unit print observations,
18 record cases and ten CSV cases match Java. The existing RecordValueTest is
now translated completely, with both methods and all fifteen source assertions,
including select. No invented persistent tests or vectors were added.
The complete original CommunityModules Ant test target is now translated in
community_modules_java_test.go after fixing the source features exposed by its
unchanged AllTestsUnix and ShiViz runs. Both translated phases pass. Source
operator arguments retain builtin/imported identities and arities, mismatched
annotated native methods are rejected, quoted string images retain their data,
simple graph paths own their elements, IOUtils templates use JavaFormatStrings
with Java validation order, and unnamed non-LOCAL parameter-free instances reuse
source definitions so config overrides remain visible. Read the progress tail
for original inputs, all 195 Java binding comparisons and sixty IOUtils records.
Final full normal workspace checks including this target pass (314.766s root).
The expanded full race check passes (2601.730s root), and the reviewed-harness
race target also passes (2364.191s). For the expanded race suite,
use go test -race -timeout 60m ./...; the original vector-clock assumptions are
expensive. Continue broader semantic/native registration and core I/O fidelity
work after verification; no reduced Community runner/config is needed.
Core TXT charset fidelity now follows OpenJDK's actual Files.writeString/
readString paths for the six guaranteed charsets and their JDK aliases. TXT
rejects malformed/unmappable input before opening a write destination. Ordinary
OutputStreamWriter/FileWriter constructors retain replacement; Json.textSerialize
uses Files.newBufferedWriter's REPORT encoder, as recorded below. UTF-16 empty
output has no BOM, and strict UTF-8 error lengths follow String's optimized file path.
Typed coding/charset exceptions retain Java class names, length messages and
IOException catch classification. Read charset validation precedes file access;
NUL paths precede charset validation. Both 1,234,164-record production comparisons
(strict Files and replacement stream/String codecs) match Java exactly; ten
path/charset-priority records match too. These are owned ephemeral observations,
not new persistent tests/vectors. The existing complete Ant suite includes the
original IOUtils tests; keep running that whole target. Broader provider charsets,
process default-charset discovery/decoding, file error/option precedence and
invalid-value casts still need their source features. IOUtils.atoi now uses
Java decimal parsing and StringValue-subclass acceptance, preserving signed
limits, BMP digit scripts, supplementary-digit rejection and source error
parameters. All 65,562 non-null production observations match exactly normally
and under -race; null retains its distinct NullPointerException family. VM
helpful-null diagnostic text remains source work. Its existing assumptions stay
in the complete Ant target with no invented tests or reduced runner.
Core TXT argument/exception and option-validation boundaries now retain source
casts, class names, evaluation order and Error propagation. Charset lookup,
enum mapping, strict encoding and incompatible file-option checks preserve their
separate source stages. Empty/nonempty options keep their distinct defaults;
SYNC/native DSYNC and umask-filtered creation are represented. The registered
priority chain retains its source primary reflected method signature while
executing JSON 25 before TXT 50. All 165 direct source calls and 87 full-chain
calls match normally and under -race; the real parser bridge already retains
unknown-format TLA fallback. Final workspace normal checks and three platform
builds pass. The TXT full race run completed successfully (2577.919s root,
snapshot 183fd47); retire handle 25120. It predates the Linux filesystem and
current JSON slices. Linux TXT filesystem
boundaries now preserve normalized separators without resolving dot/parent
components, empty syscall paths, strict UTF-8 native path encoding and NIO error
families. DELETE_ON_CLOSE unlinks immediately and refuses final symlinks; ignored
unlink errors and CREATE_NEW empty/final-dot quirks are retained. Open errors
retain paths, channel errors do not; close failures cannot turn into success.
The Files read-size limit retains its Error boundary. Owned observations match
763 bounded TXT calls, 128 open-channel calls, 120 nullable exception constructors
and three sparse-file/charset calls; bounded TXT and sparse calls match under
-race too. Final full normal and TLC race checks pass. Continue NDJSON stream/public
convenience boundaries, native/default charset discovery and other filesystem
providers. No new persistent tests/vectors were added; the original whole Ant
suite and TXT round-trip assumptions remain unchanged. See the progress tail.
Core Json.textSerialize now retains the actual Files.newBufferedWriter path:
REPORT encoding, separate 8192 UTF-16-character/8192-byte buffers, opening before
node conversion, partial output, close-error suppression and RuntimeException
wrapping only inside the source try. Options and payload conversions occur once
before destination evaluation. Required casts stay outside the catch. Gson string
escaping preserves UTF-16 units and accepts the represented StringValue subclass;
RecordValue.apply returns a present null component instead of treating it as absent.
All 932 actual-source observations match normally and under -race, including
56 suppressed-failure cases and normalized stack headings; all 65,536 single-unit
JSON strings match Java. VM helpful-null text and real Java stack frames are not
claimed. The complete original core JsonTest is now translated after the feature,
with its byte-identical JsonTests.tla under test_vectors/ and private output cwd.
Actual unchanged source JUnit and translated normal/race runs pass. This does not
establish ordinary JSON FileWriter/reader/default-charset, parser leniency,
provider charset or static-monitor parity; those remain source work. The final
full workspace normal run passed (313.379s root), core-json-full-normal-final.log.
The full race run for that preceding JSON snapshot passed (2551.695s root),
core-json-full-race-final.log; retire handle 6743.
Ordinary Json.serialize/ndSerialize now retain FileWriter replacement encoding,
opening before node conversion, ignored mkdirs results and unwrapped primary
failures with suppressed close errors. Linux legacy File paths retain logical
UTF-16 diagnostics while native filename access replaces malformed units; NUL
and empty-path errors differ from NIO. Directory canonical fallback preserves
hard parent failures. Default charset keeps frozen startup file.encoding and
cached resolution; runtime property changes do not alter it. Six standard
charsets/aliases are represented; extended providers and COMPAT/native discovery
remain. The three synchronized static writers share the existing reentrant class
monitor, including argument evaluation in NDJSON. All 7,008 actual-source calls
(584 inputs across twelve startup charset names) match normally/under -race;
584 production CLI startup-charset calls, fifteen canonical/mkdirs side-effect
calls and the reentrant/contended monitor call match under -race too. Consecutive
high-surrogate input at encoder boundaries retains the next pending surrogate.
Canonical fallback respects the kernel's forty-symlink limit rather than Go's
more permissive resolver. Original JsonTest remains whole and passes. Full
normal checks passed after the encoder correction (312.380s root); final normal
checks after the native symlink-limit correction passed (309.577s root),
core-json-files-baseline-normal.log; retire handle 42423. Final TLC race checks passed (3.555s),
core-json-files-baseline-tlc-race.log; retire handle 96625. The full workspace race run covering the ordinary-writer/encoder snapshot
passed (2525.182s root), core-json-files-full-race-final.log; retire handle 83512.
That binary predates only the native symlink-limit correction, covered by the
final full normal and TLC race checks plus fifteen actual-source race observations.
JSON readers/Gson leniency and other native providers remain source work;
read the progress tail for final verification status.

The requested clean baseline is committed as d68a83b, with full normal and TLC
race checks green. The first correctness-test slice completes ContextTest (ten
methods), InitializeValueTest (all ten original methods with shared vectors),
and FormulaTest (both methods, restoring the omitted nested-LET case).
Unchanged Java JUnit passes all 22 methods; complete Go translations and TLC
race checks pass (3.581s). TODO_TEST_PORT.md marks all three "Port complete";
235/1269 logical methods and 103/626 classes are mapped. Full workspace normal
checks for this test slice passed (312.442s root, 1.084s SANY, 0.930s TLC),
correctness-java/initialize-context-formula-full-normal.log; retire handle 17599.
Next correctness slices: original queue, value/enumeration and module
methods, reconciling all original inputs/assertions, then model regressions.

Current user direction: prioritize mechanically porting original correctness tests.
Use TODO_TEST_PORT.md for the inventory and update completed entries to
"Port complete". Temporarily skip debugger/scoped identifiers, checkpoint/recovery
models, distributed TLC, JPF verification, and benchmarks/supporting fixtures.
When a ported original test fails, inspect for shortcuts in production TLC and
finish the accurate implementation before moving on; preserve the source test.

Attaching DAP
transport/capability events, full SANY level metadata,
theorem/proof/top-level children, broader nested and parameterized instance
metadata/export composition and complete native override registration remain
source work. Full debugger and full TLC completion are not established.
Operating-system default-locale discovery and other UTF-16 consumers remain
source work; formatter startup language/category and line-separator properties
are represented and frozen on first use. TLCGetNonDeterminismTest
is ignored upstream by design; preserve that status when translating it.
Original model bytes, including upstream whitespace, remain in test_vectors/.

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

## Distributed Transport Decision

User decision (2026-10-03): use the user's `~/rpc25519` system
(`github.com/glycerine/rpc25519`) for distributed TLC communications, with
Greenpack serialization (`github.com/glycerine/greenpack`). The local source is
`/mnt/oldrog/home/jaten/rpc25519`.

Use its peer/circuit/fragment actor API to pipeline state and fingerprint
batches without waiting synchronously for each remote response. Correlate
asynchronous results with outstanding batches; dependent work still requires
its results. Preserve Java TLC's deduplication, checkpoint, failure-recovery and
termination semantics. This records the selected transport; network integration
remains pending. See `TLC_ARCH.md` for the intended mapping.

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
- `tlc/TODO_TEST_PORT.md`: original Java test-method inventory and current port status.
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


Correctness continuation: StateQueueTest is now port complete in queue_test.go,
with all nine original methods separated, exact Integer.MIN_VALUE/MAX_VALUE,
ten repeated identical/distinct enqueue cases, and original assertion order.
Opaque zero-valued Go state identities replace unused DummyTLCState abstract
stubs; no production changes needed. Unchanged Java JUnit with -ea passes nine
methods; Go targeted methods and full TLC -race pass (3.322s); retire handles
92874 and 34280. Current inventory: 240/1269 methods (18.9%), 104/626 complete
classes (16.6%), 1029 pending contexts across 522 classes, two partial classes.
Next: DiskPoolWriterTest requires faithful WAITING/BLOCKED/alive/join observations;
current Go writers expose neither native lifecycle completion nor wait state.
Do not replace its assertions with a sleep-and-no-error approximation. Alternatively
advance direct value/module tests while preserving this implementation prerequisite.


EnumerableValueTest is also port complete: exact seeded sweep across n=1..10656,
all original bounds and uniqueness assertions. The test calls Go's production
SubsetEnumerator equivalent directly because the Java dummy only forwards size;
its unused abstract stubs contribute no behavior. Current original implementation
and test pass Java -ea JUnit (1.392s); Go sweep passes (1.184s); full TLC -race
passes (7.642s). Inventory now 241/1269 mapped methods, 105/626 complete classes,
1028 pending methods across 521 classes. Retire handles 39012, 66448, 95532.
Current follow-up candidates: original value/stream/module methods and model
regressions, fixing production shortcomings when exposed. Preserve the original
DiskPoolWriterTest wait-state/join requirements when taking up that feature.

Full TLC normal checks for the queue/enumeration snapshot pass (1.749s); retire
handle 21325. Baseline and first batch full workspace normal results remain green;
this continuation changes original-test translations and documentation only.


FcnRcdValueTest is now port complete: all thirteen original methods, complete
selection matrices and exact typed-model-value exception messages. Its binary
search case exposed a production shortcut (Go lower_bound versus Java
Arrays.binarySearch); Select now follows Java's inclusive high bound, immediate
compare==0 return and subsequent equality check. Preserve the unchanged A_Z
exception expectation. Original current Java production class plus original
JUnit tests pass (13 methods, 0.303s); Go translations pass (2.344s); full TLC
-race passes (26.815s). Retire handles 43177, 49416, 96528. Full normal workspace
checks passed (314.927s root, 1.068s SANY, 3.594s TLC),
correctness-java/fcn-record-full-normal.log; retire handle 37843. This coherent
slice is ready to commit. Inventory now 254/1269 mapped methods (20.0%), 106/626
complete classes (16.9%), 1015 pending methods across 520 classes.
Next value-stream reconciliation must preserve real files, close/reopen and
23/26-byte gzip file-size assertions; current byte-buffer approximations omit
those source assertions. Do not weaken them. Remaining direct value/module tests
and model regressions remain in scope; five user-deferred topics stay skipped.

IntervalValueTest follow-up: original elementAt (three methods) has no Go
production method yet. Port its short-circuit 0 <= idx && idx < size() check,
IntValue.gen(low+idx) and exact Assert.fail message/source before the tests.
The remaining interval methods also need original overflow exception family/text,
comparison extremes, both post-reset enumerator assertions and Diff/Cap/Cup sizes.


IntervalValueTest continuation: all fourteen original methods translated; production
ElementAt is now implemented with Java evaluation order and source-aware failures.
Size overflow also retains detailed source/context. Java original current production
class and original tests pass (14 methods); targeted Go tests pass. TODO marks port
complete, 268/1269 mapped methods, 107/626 complete classes, 1001 pending contexts
across 519 classes. Initial normal workspace checks at 89185 failed on TLC disk-full writes;
root and SANY passed (interval-full-normal.log). Retire 89185.
Initial race build 56510 failed because disk filled. After pruning only inspected
old disposable cache objects, retry 97304 passed (27.431s; retire handle),
interval-tlc-race-retry.log. Poll live normal handle 89185; do not restart it. Update terminal results
and commit when green. Retire handles 59627 and 56510. Scope exclusions unchanged.

Full normal handle 89185 terminated: root passed (313.822s), SANY passed
(1.006s), but TLC had failed on disk-full TempDir/testlog writes before cache
cleanup. Retire 89185; the old live statements above are superseded. Full
workspace normal retry passed with free disk space (315.060s root, 1.043s SANY,
3.576s TLC); retire 55823. Preserve original failure log interval-full-normal.log
and passing interval-full-normal-retry.log.

Next ModelValueTest: current original ModelValue.java and unchanged class
compile and pass JUnit -ea (44 methods, 0.032s). Not yet credited as Go ports.
Preserve JUnit method execution order because UniqueString.compareTo uses intern
IDs, and exact +/-1 comparisons depend on creation order across the class.
Actual JUnit runner order is in owned ignored correctness-java/model-value-junit-order.txt.
Keep source testCompareToTupeMV using StringValue("foo") as written, despite
its name. Keep BoolFalse versus BoolTrue direction-specific inputs and all
expected TLCRuntimeException families; old Go checks accepted any error/sign.

ModelValue class isolation can save/restore the package internTable pointer
and modelValues count/table/mvs under its mutex, using a fresh InternTable while
executing the original JUnit order. Do not reset the intern table per method,
hoist all input creation, or replace exact compare results with sign tests.
