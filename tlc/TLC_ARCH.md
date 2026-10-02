# TLC Architecture Notes for the Go Port

This document records the architecture of the Java TLC model checker in
`../tlaplus/tlatools/org.lamport.tlatools/src/tlc2` and its tests in
`../tlaplus/tlatools/org.lamport.tlatools/test/tlc2`. It is meant to guide a
faithful Go port under `tlago/tlc`.

The porting rule for this effort is strict: implement TLC by following the Java
function decomposition mechanically before translating the Java tests. The Go
code may use Go idioms for memory, errors, goroutines, and concrete structs, but
semantic boundaries, evaluation order, fingerprinting, state normalization,
queue behavior, and error precedence must match Java TLC.

Current handoff and next-step status live in `HANDOFF.md` and
`PORT_PROGRESS.md`. This architecture document explains the source-of-truth
design and original porting order; use the handoff/progress files to decide
where to resume.

## Source Inventory

As of this survey, Java TLC contains:

- `371` Java files under `src/tlc2`.
- About `96,749` lines under `src/tlc2`.
- `699` Java test files under `test/tlc2`.
- `1,135` JUnit `@Test` sites under `test/tlc2`.
- `646` `.tla` fixtures and `524` `.cfg` fixtures under `test-model`.

Top-level source package sizes:

- `tlc2/tool`: model checking, simulation, state, trace, queues, fingerprint
  sets, liveness, coverage, distributed TLC, and most core algorithms.
- `tlc2/value`: TLA+ runtime value model, serialization, random enumerable
  support, model values, and value constants.
- `tlc2/util`: FP64, custom vectors/tables/stacks/queues, disk IO helpers,
  random generator, state writers, statistics.
- `tlc2/module`: Java implementations of TLA+ standard modules and TLC
  extension operators.
- `tlc2/output`: message codes, message formatting, recorders, state printing.
- `tlc2/model`: Toolbox-oriented model/error/trace presentation classes.
- `tlc2/debug`: debugger protocol and stack-frame support.
- `tlc2/overrides`: Java module override discovery and annotations.
- `tlc2/pprint`: pretty-printer support used by output paths.
  Go mirrors this as central concrete helpers in `pretty_print.go`.
  `ValuesPPR`/`ValuesPPRString` correspond to Java `tlc2.value.Values.ppr`
  and read the Java property `tlc2.value.Values.width` with default width 80.
- `tlc2.tool.impl.SpecProcessor.vetoed` controls Java's constant-operator
  pre-evaluation veto list. The Go constant-definition pass reads the same
  property through the central TLC property map.
- Java system properties that steer runner/checker behavior are read through
  the same central property map: `tlc2.TLC.nosuspend`,
  `tlc2.TLC.nohalt`, `tlc2.TLC.stopAfter`,
  `tlc2.tool.ModelChecker.vetoCleanup`, and
  `tlc2.tool.ModelChecker.BAQueue`.
- Disk state queues and disk byte-array queues both use Java's
  `tlc2.tool.queue.DiskStateQueue.BufSize` property, defaulting to 8192.
- FPSet construction follows Java properties through the central property map:
  `tlc2.tool.fp.FPSet.impl`, `tlc2.tool.fp.DiskFPSet.logLockCnt`, and
  `tlc2.tool.fp.OffHeapDiskFPSet.probeLimit`. Disk fingerprint sets also
  honor `tlc2.tool.fp.DiskFPSet.metadirPrefix` and
  `tlc2.tool.fp.DiskFPSet.error2warning`.
- Tableau construction mirrors Java's debug export hook:
  `tlc2.tool.liveness.Liveness.tableauExportPath` writes `TBGraph` DOT output.
- Coverage cost-model creation mirrors
  `tlc2.tool.coverage.CostModelCreator.implied`, defaulting to true for
  implied init/action coverage.

Top-level test package sizes:

- `tlc2/tool`: the dominant end-to-end model-checking suite.
- `tlc2/tool/liveness`: temporal property, tableau, SCC, symmetry, simulation,
  and trace/lasso tests.
- `tlc2/tool/fp`: fingerprint set implementations and disk/off-heap indexing.
- `tlc2/tool/queue`: state queue behavior and disk pool writer tests.
- `tlc2/value/impl`: value semantics and enumerator tests.
- `tlc2/module`: standard module and TLC extension behavior.
- `tlc2/debug`: debugger protocol integration tests.
- `tlc2/util`: FP64, vectors, queues, stacks, statistics, byte encoding.
- `tlc2/model` and `tlc2/output`: Toolbox presentation and message behavior.

## Package Strategy in Go

Keep TLC mostly in package `tlc` under `tlago/tlc`, with subdirectories only
when Go requires it for commands or frozen fixtures. This follows the user's
preference to avoid circular imports. Java packages should map to file groups
inside the same Go package:

- `cli_*.go`: command parsing and top-level run orchestration.
- `tool_*.go`: `Tool`, `FastTool`, spec processing, expression evaluation.
- `state_*.go`: `TLCState`, mutable state variants, state vectors.
- `value_*.go`: value hierarchy and enumerable values.
- `fp_*.go`: fingerprinting and fingerprint sets.
- `queue_*.go`: state queues and checkpointing.
- `liveness_*.go`: temporal formula nodes, tableau, behavior graph, SCC.
- `module_*.go`: TLC/Naturals/Integers/Sequences/etc. built-ins.
- `output_*.go`: message codes and recorder-compatible output.
- `config_*.go`: `.cfg` parser and model configuration.

The Go package can still preserve Java names in type and method names where
useful during the mechanical port. After tests are green, cosmetic refactors can
be considered separately.

Prefer concrete structs over Go interfaces throughout the port. Java uses
interfaces such as `ITool`, `IStateFunctor`, `INextStateFunctor`, `IWorker`,
`FPSet`, and `IStateQueue` heavily, but the Go port should introduce an
interface only when there are already multiple meaningful implementations or
when polymorphism is itself part of the runtime value model. For single
implementations, use concrete structs such as `Tool`, `StateFunctor`,
`NextStateFunctor`, `Worker`, `MemFPSet`, `MemStateQueue`, `ModelConfig`, and
`ConfigConstant`; this keeps stack traces, debugger watches, and crash dumps
straightforward during the mechanical port.

When Java has an interface with exactly one production implementation, port the
implementation as a concrete Go struct and let call sites name that struct
directly. `IMVPerm`/`MVPerm` is the model: Java exposes an interface, but Go uses
`*MVPerm` because there is only one real permutation representation and the
fixed indexed array is important to understand in a debugger.

Any Go map whose iteration can affect output, fingerprinting, exploration
order, diagnostics, or tests must use `InsMap` from `insmap.go`. Built-in Go
maps are acceptable for lookup-only sets/tables that are never ranged over in
observable code.

## Main Execution Architecture

### `tlc2.TLC`

`tlc2.TLC` is the top-level runner. Its responsibilities are:

- Parse command-line flags.
- Select run mode: exhaustive model checking or random simulation.
- Resolve spec/config/metadir paths.
- Configure global TLC options in `TLCGlobals`.
- Initialize FP64 polynomial selection.
- Create the `Tool`/`FastTool` spec handle.
- Create an `FPSet`.
- Create a `ModelChecker`, `DFIDModelChecker`, or `Simulator`.
- Wire optional state writers (`-dump`, `-dot`).
- Start debugger mode when requested.
- Print welcome, progress, error, and summary messages.

Important fields:

- `runMode`: `MODEL_CHECK` or `SIMULATE`.
- `cleanup`: remove metadata after successful run.
- `deadlock`: whether deadlock checking is enabled.
- `seed`, `aril`, `traceDepth`, `traceNum`: simulation control.
- `mainFile`, `configFile`, `metadir`, `fromChkpt`: model input and metadata.
- `fpIndex`, `fpSetConfiguration`: fingerprint polynomial and storage config.
- `stateWriter`: optional trace/graph dumping sink.
- `tool`: volatile handle used by debugger and runtime components.

Port guidance:

- Implement a `Runner` or `TLC` struct with the same lifecycle.
- Do not hide global options initially. Java code depends heavily on
  `TLCGlobals`; mirror it with a package-level `Globals` struct to preserve
  behavior, then later reduce global mutation if tests allow.
- Keep version metadata centralized in the TLC globals layer. Java's
  `TLCGlobals.Version` feeds management beans, `TLCGet("revision")`, and
  `TLCGet("config")` install fields; Go should route all of those through the
  same helper functions.
- Preserve Java's global property defaults and live property updates:
  `tlc2.TLC.progressInterval` is seconds coerced with `max(abs(x), 1)`,
  `tlc2.TLCGlobals.chkpt` is milliseconds, and
  `tlc2.TLCGlobals.coverage` is the coverage bitmask. Go initializes these
  from the process environment and updates them when the Java-style `-D`
  parser or `TLCSet("-D...")` writes the same property names.
- Preserve exit-status categories and message codes, because the Java tests
  assert message recorder events rather than only stdout text.
- Mirror `handleParameters` as a real library parser, not only as command
  wrapper glue. Java accepts a broad command vocabulary before it constructs
  `FastTool`, and later subsystems depend on those parsed fields being present:
  `-simulate`/`-generate` with `num=`, `file=`, `stats=basic|full`, and
  `sched=rl|rlaction`; model-checking controls such as `-dfid`, `-workers`,
  `-recover`, `-metadir`, `-checkpoint`, `-coverage`, `-cleanup`, `-deadlock`;
  evaluator/output globals such as `-difftrace`, `-nowarning`, `-gzip`,
  `-terse`, `-continue`, `-view`, `-debug`, `-tool`; fingerprint controls
  `-fp`, `-fpmem`, and `-fpbits`; debugger controls `-debugger` with optional
  `port=`, `nosuspend`, and `nohalt`; trace exploration controls
  `-generateSpecTE`, `nomonolith`, `-noTE`, `-noTEBin`, and `-teSpecOutDir`;
  state/trace dump controls `-dump`, `-dumpTrace`, and `-loadTrace`; and
  runtime spec additions `-inv`, `-invlevel`, and `-postCondition`.
- The Go port keeps these in concrete structs: `Options` for direct runner
  fields, `SimulationSchedule` for RL scheduling, and `RuntimeParameters` for
  dynamically generated invariant/constraint/postcondition/view additions.
  This mirrors Java's `params` map without introducing an untyped central map
  into the Go API. `InsMap` backs parsed message-control sets to keep any later
  iteration deterministic.
- Runtime parameters follow Java `ParameterizedSpecObj` ordering. Runtime
  extendees are postcondition modules, invariant dependency modules, model
  constraint modules, action constraint modules, then view. Runtime string
  constants are bound in the same order Java's `processConstantDefns` visits
  them: model constraints, action constraints, postconditions, invariant
  dependency modules if they declare constants, then view. The current Go
  runtime invariant template form has no constant definitions of its own, but
  view constants are still carried through `RuntimeView` like Java
  `ParameterizedSpecObj.View.constDefs`. The Go bridge installs these runtime
  constants before ordinary config constants because Java
  `ParameterizedSpecObj.processConstantDefns` runs before
  `SpecObj.processConstantDefns`; if the same constant name appears in both,
  the config constant wins. Runtime invariant actions are appended after normal
  config processing, so the missing `INIT`/`NEXT` checks only see the static
  model config just as Java's `SpecProcessor` does.
- Runtime postconditions follow Java `Spec.getPostConditionSpecs`: actions from
  `ParameterizedSpecObj.getPostConditionSpecs()` come before config-file
  `POSTCONDITION(S)`, and their user-visible action name is the unqualified
  operator name even though the Go bridge resolves the qualified module
  definition internally. This matters for diagnostic/postcondition output order.
- Java's `-dump class,...` loads an `IStateWriter` by reflection. Go keeps this
  concrete with `StateWriterFactory` functions registered by class name through
  `RegisterStateWriterClass`; built-in parity names cover Java's zero-argument
  `NoopStateWriter` and `DotStateWriter`, and unknown class names fail during
  option parsing instead of silently dropping state output. The requested class
  name is still recorded in `RuntimeParameters.CustomStateWriterClass` for
  downstream visibility.
- Java's `DotStateWriter` prints full successor labels unless
  `TLCGlobals.printDiffsOnly` is set. Stuttering is an explicit
  `Visualization.STUTTERING` hint; an ordinary self-loop is still an ordinary
  edge. Keep those decisions on the concrete `StateWriter` rather than adding
  writer interfaces.
- Java's CLI DOT writer constructor always writes a `strict ` graph header
  prefix, even when the parsed `strict` sub-option is false; the sub-option only
  controls the in-memory duplicate-edge suppression set. Go models this with a
  separate `StrictPrefix` option for the CLI path. The zero-argument custom
  `DotStateWriter` and liveness DOT writer use Java's prefix-free constructors.
- DOT transition colors are assigned by action name, but transition labels use
  `Action.getInvocationSignature()`, so parameterized action instances print
  their concrete values on edges.
- `DotStateWriter` fingerprints and ranks the original states but labels
  initial and successor nodes with `state.evalStateLevelAlias()`. Successor
  tooltips remain the original state text. The local `ModelCheckerMXWrapper`
  also renders `getCurrentState()` through the same state-level alias; the
  distributed server wrapper does not.
- Java defaults simulation `traceNum` to `Long.MAX_VALUE`, not one trace. Go
  runner and simulator defaults must preserve that, with explicit `num=` or API
  options narrowing the trace count.
- Per-message suppression/elevation belongs in the output recorder path because
  Java's tests observe `MP` recorder events. In Go, `PrintWarning` honors
  `Globals.Warn` and per-code suppression, and `MessagesAsErrors` upgrades the
  recorded severity.

### `Tool`, `FastTool`, and `ITool`

`tlc2.tool.impl.Tool` is the core runtime/evaluator API. `FastTool` subclasses
`Tool` only to mark hot methods final and inline-friendly in Java. In Go, make a
single concrete `Tool` type unless a debugger/call-stack wrapper requires
separate structs.

The `ITool` boundary includes:

- `GetActions`
- `GetInitStates`
- `MakeState`
- `GetNextStates`
- `Eval`
- `IsGoodState`
- `IsInModel`
- `IsInActions`
- `EvalReward`
- `Enabled`
- `IsValid`
- `CheckAssumptions`
- `CheckPostCondition`
- `GetState` trace reconstruction
- `GetSymmetryPerms`
- `Contexts`
- getters for invariants, implied actions, temporals, constraints, aliases,
  postconditions, config, module files, and root names.

The `Tool` constructor does important work:

1. It calls `Spec`/`SpecProcessor` to parse and process the spec/config.
2. It initializes static state variable metadata on `TLCStateMut` or
   `TLCStateMutExt`.
3. It obtains the next-state specification.
4. It splits the next-state predicate into actions.
5. It assigns stable numeric IDs to init predicates and actions.

Port guidance:

- Keep `Tool` as the central object that holds processed SANY semantic nodes,
  config-derived declarations, and evaluator helpers.
- Preserve the Java call order. Many tests depend on exactly when constants,
  model values, overrides, variable locations, and action names are available.
- Port `FastTool` as either an alias/wrapper or simply as constructor options
  on `Tool`. The important thing is behavior, not Java inheritance.
- Java's `SymbolNodeValueLookupProvider` is a default-method interface mixed
  into `Tool` and `TraceApp`. In Go it should stay as concrete `Tool` methods:
  `LookupWithCutoff`, `GetVal`, `GetOpContext`, `GetVar`, `GetLevelBound`, and
  `GetLevelBoundAppl`. This keeps lookup semantics centralized without adding a
  one-implementation interface.
- Mirror Java `ITool` accessors as concrete `Tool` fields and methods. In
  particular, expose `SpecProcessor`, module-file paths, assumptions,
  assumption-axiom flags, and `CounterExample`'s backing operator definition on
  `Tool` itself instead of adding an `ITool`-style interface.

## Spec and Config Processing

### `Spec`

`tlc2.tool.impl.Spec` wraps parsed SANY semantic artifacts and exposes processed
model components:

- root module, module table, resolver, root file/config file/spec dir.
- variables and primed variable locations.
- initial predicate actions.
- next-state action.
- model constraints and action constraints.
- view, alias, symmetry, temporals, invariants, implied init/actions.
- assumptions and postconditions.
- lookup and substitution helpers.

It also collects primed variable locations for action/state-level evaluation.

### `SpecProcessor`

`SpecProcessor` converts SANY output plus the model config into TLC runtime
objects. Key jobs:

- Process constant definitions and module constants.
- Process Java module overrides.
- Apply model overrides and substitutions.
- Resolve `INIT`, `NEXT`, `SPECIFICATION`, `INVARIANT`, `PROPERTY`,
  `CONSTRAINT`, `ACTION_CONSTRAINT`, `VIEW`, `SYMMETRY`, `ALIAS`,
  `POSTCONDITION`, `_PERIODIC`, `_RL_REWARD`, and `_POSSIBLE`.
- Split temporal formulas into implied init/action/temporal pieces.
- Build `Action` objects with contexts and names.
- Initialize `ModelValue` state.
- Set variable locations on `UniqueString`.
- Set the definition table count to the number of variables before storing
  definitions. Java does this explicitly with `defns.setDefnCount(varDecls.length)`;
  the Go port keeps the same visible step instead of hiding it in `Defns`.
- Expose the same processed-spec accessor surface as Java
  `SpecProcessor`: init/next predicates, temporal and implied temporal actions,
  invariants, implied init/action checks, model/action constraints, assumptions,
  `_RL_REWARD`, `_PERIODIC`, definition snapshots, constant-definition cache,
  and postcondition specs. Go returns slice copies from these concrete methods
  to keep callers from mutating processor-owned slices by accident.

Tricky details:

- The config table is intentionally heterogeneous in Java: values can be
  strings, vectors, raw values, model values, or override descriptions. A
  cleaned-up Go representation is fine only if it preserves every parse-time
  and process-time behavior.
- Module overrides are part of semantics, not a performance-only feature.
- Model values must be initialized and ordered exactly as Java does because
  comparison, printing, and fingerprinting depend on it.
- Constants can be static, dynamic, module-scoped, or override-driven.
- `tlc2.tool.impl.ModelConfig.nosymmetry=true` makes Java's
  `ModelConfig.getSymmetry()` return the empty string even when the cfg file
  contains `SYMMETRY`; the Go port mirrors this at `GetSymmetry` so downstream
  tool/state setup sees no symmetry set.
- Config processing must rebuild every derived vector from scratch when it is
  run. The Go bridge constructs `SpecProcessor` before all definitions and
  variables are installed, then runs it again while applying the processor to a
  `Tool`; accumulated init/property/temporal vectors would diverge from Java.
  Refresh the pre-constant definition snapshot immediately before
  `ProcessConstantDefinitions`, because Java uses that snapshot for
  `_RL_REWARD` and `_PERIODIC` after ordinary constants have been evaluated.
- Config keyword validation is keyword-specific. `INIT`, `NEXT`, `VIEW`,
  `POSTCONDITION`, `_POSSIBLE`, `_PERIODIC`, and `_RL_REWARD` require
  zero-arity operator definitions; `INVARIANT` and `PROPERTY` ignore literal
  `TRUE` but reject literal `FALSE` or non-boolean values with
  `TLC_CONFIG_ID_HAS_VALUE`; `SPECIFICATION` has its own value error wording.
- Structural config decomposition has the same checks when it follows a
  zero-argument operator reference inside `SPECIFICATION` or `PROPERTY`: Java
  rejects references to parameterized operators, unknown names, literal
  `FALSE`, and non-boolean values immediately instead of falling through and
  classifying the expression by level.
- Model constraints and action constraints accept zero-arity operator
  definitions. Java stores the operator definition on the body node for later
  coverage reporting, appends the body to the constraint list, ignores literal
  `TRUE` values, and treats non-zero-arity definitions, literal `FALSE`,
  non-boolean values, and unknown names as configuration errors before model
  checking begins.
- After definitions and config substitutions are installed, Java
  `SpecProcessor.processConstantDefns` walks zero-arity operator definitions
  whose effective level is constant and pre-evaluates them into TLC `Value`s.
  Evaluation failures are deliberately swallowed for constant-level operators
  such as `Seq(S)` or failing `TLCGet` expressions so that the ordinary
  model-checking path reports the real use-site error later. The Go port mirrors
  this with `SpecProcessor.ProcessConstantDefinitions`: no new abstraction, just
  the concrete processor using the concrete `Tool`, a flat `ConstantDefns`
  cache, Java-style snapshots, and a veto list keyed by
  `tlc2.tool.impl.SpecProcessor.vetoed` or `TLAGO_SPEC_PROCESSOR_VETOED`.
- Java routes each successfully evaluated zero-arity constant operator through
  `WorkerValue.demux`. The value is deep-normalized but not eagerly initialized
  for fingerprinting at spec-processing time. Mutable values with multiple
  workers are re-evaluated once per worker under the same
  `RandomEnumerableValues` seed and stored as a `WorkerValue`; immutable
  primitives stay as plain values. The Go port keeps the same visible storage
  shape in `Defns` and on the operator tool object, while the convenience
  `ConstantDefns` cache stores worker zero's concrete value for callers that
  require a `Value`.

### Go SANY to TLC Bridge

The root package owns the production Go SANY parser and semantic tree. The TLC
runtime package must not import the root package, so the adapter lives in the
root package as `BuildTLCTool`. Import direction is one-way: root `tlago`
imports concrete `tlc` structs and builds a `Tool`; `tlc` remains independent.

Current adapter responsibilities:

- collect the root module variables and install TLC state variable locations.
- convert SANY `Definition` and `Expr` nodes into TLC `OpDefNode`,
  `OpApplNode`, `LetInNode`, quantifier, value, and action nodes.
- install config constants, including arity-bearing operator-constant rows as
  `OpRcdValue`, module-qualified constants as `Module!Name`, and install
  operator overrides into `Tool`.
- attach stable `SymbolNode` identities to bridged `OpDefNode`s so LET binding
  and later operator application use the same symbols.
- install named-instance aliases for TLC standard-module overrides and local
  LET instance aliases such as `T!PrintT`.
- evaluate `ALIAS` config operators by converting record/function-record values
  into printable TLC states. Java's `RecordValue.toState()` permits aliases
  that are subsets or supersets of the spec variables; Go mirrors that with a
  concrete record-backed state, preserving state-variable fingerprint/equality
  behavior while printing the alias record and honoring `_format`.
- resolve config-selected `INIT`, `NEXT`, `SPECIFICATION`, invariants,
  properties, constraints, view, and postconditions into `Action` or semantic
  nodes.
- preserve deterministic definition installation by sorting the collected names.

This bridge is a staging boundary, not a replacement for Java `SpecProcessor`.
The mechanical port still has to move exact Java visibility, full INSTANCE
processing for parameterized and non-standard modules, exact module-constant
scoping, action decomposition, implied init/action splitting, symmetry, aliases,
`_POSSIBLE`, `_PERIODIC`, and `_RL_REWARD` into the TLC-side `SpecProcessor`
shape. Until that is complete, the adapter should stay simple and explicit so
mismatches are easy to see.

### `Defns` and `Specs`

`Defns` is a compact definition table keyed indirectly by `UniqueString.loc`.
The same `loc` field is also used for state variable positions:

- variables occupy locations `[0, varCount)`.
- definitions start at `defnIdx`, normally initialized to `varCount`.
- `UniqueString.getDefnLoc` returns `-1` for variable locations.
- `put` assigns a new location only when the key's definition location is `-1`.
- `snapshot` copies the backing array and the current index.

`Specs.getLevel` is a static helper used by TLC to compute the effective level
of a level-checked expression under a context. It starts from the node's SANY
level, then follows each level parameter through the context. If the binding is
a `LazyValue`, it recurses on the lazy expression and context; if the binding is
an `OpDefNode`, it recurses on that operator definition. The Go semantic nodes
therefore carry concrete level metadata on `SemanticNodeBase`, and `OpDefNode`
embeds the same base rather than hiding level behind an interface.

`Specs.addSubsts` wraps an expression in the queued `SubstInNode` substitutions
in list order. This is a structural SANY helper and should remain mechanical.

`LET` context extension is shared by evaluation, enabledness, init generation,
next-state generation, and variable lookup. Zero-arity local definitions bind as
lazy values, while arity-bearing local operators bind as concrete `OpDefNode`
values in the same context. This mirrors Java's semantic-node environment
without introducing a one-implementation closure interface. Local module
instance aliases that resolve directly to built-in method values are stored as
explicit `LetBinding` entries on `LetInNode`; the same binding path is used by
evaluation, enabledness, init, next, level calculation, and variable discovery.
Parameterized theorem and assumption definitions use Java's separate
`Spec.getOpContext(ThmOrAssumpDefNode, ...)` path: bind their formal parameters
with `getVal(arg, context, cachable)` before evaluating the body. The Go
`ThmOrAssumpDefNode` therefore carries `Params`, and eval/init/next/enabled all
route through `GetThmOrAssumpContext`.

### Model Presentation

`tlc2/model` contains Toolbox-facing data holders, not checker algorithms:
`Formula`, `Assignment`, `TypedSet`, `MCVariable`, `MCState`, `MCError`, and
trace-expression metadata. Keep these as concrete structs in `model.go`.

`TypedSet` owns its own string, equality, hash, and type-validation helpers,
mirroring Java's `TypedSet` class. Spec-writing code should call those methods
rather than define model formatting locally. Its parser also preserves Java
`String.split(pattern, 0)` behavior: trailing empty pieces are discarded, so
inputs such as `", , , ,"` and `"{, , , ,}"` parse as the empty set.

`MCVariable.isTraceExplorerExpression()` is a nullness test in Java, not a
non-empty-string test. Go therefore tracks whether `SetTraceExpression` was
called separately from the stored expression text. `MCState` conjunctive
descriptions also preserve Java's optional ANSI bold/reset wrapping for trace
expressions.

`MCError` has two easy-to-clean-up Java quirks that affect trace exploration:
`toSequenceOfRecords` decides comma insertion from the original state index, not
from whether a previous non-marker record was emitted, and
`isLassoWithDuplicates` compares the full lasso trace length against the unique
normalized records from all states except the final back-to-state marker. The
method name sounds narrower than the algorithm; keep the algorithm because it
controls whether a liveness trace gets an explicit trace-view definition.

The trace-expression spec writer is also intentionally strict like Java.
Back-to-state targets and next-state variables are indexed directly, so malformed
traces fail loudly instead of producing a best-effort TE spec. The
`TLC_TRACE_EXPLORER_JSON_UNCOMMENTED` property removes comment prefixes from the
JSON `ASSUME` block just as Java does for tests, and `indentString` uses Java
split semantics by dropping trailing empty lines. The concrete
`TraceExpressionExplorerSpecWriter` stores expressions in Java `TreeMap` order,
so generated declarations and conjuncts use lexicographic variable-name order
rather than insertion order.

`StatePrinter` is the Java helper that gives state trace messages their
metadata. Preserve its small details: invariant traces use predecessor diffs
only when `TLCGlobals.printDiffsOnly` is set, incomplete states print
fingerprint `-1`, `TLC_BACK_TO_STATE` routes through `printState` only in
tool mode, and the opt-in `tlc2.output.StatePrinter.overwrite` property sleeps
for the configured milliseconds and emits terminal reset sequences unless the
state is final.

Generic spec-writer utilities also carry visible Java details. The generated
identifier counter starts like Java's `AtomicLong(1L)` and uses
`incrementAndGet()` before formatting, so a fresh process first emits suffix
`2000`; timestamps and modification-history dates use Java `Date.toString()`-
style strings rather than Go's `time.Time.String()` output.

`ErrorTraceMessagePrinterRecorder` observes printed state messages to reconstruct
the error trace. Its back-to-state path treats any parseable positive ordinal as
a terminal marker before it validates that the ordinal points at an existing
state, so Go's recorder must finish the trace before target validation too.

### `ModelConfig`

`ModelConfig` parses `.cfg` files with the TLA+ token manager rather than a
separate parser. It recognizes the keywords listed above, stores raw constants,
module constants, operator overrides, and the deadlock setting.

Port guidance:

- Keep the Go representation explicit and concrete rather than reproducing
  Java's heterogeneous `Hashtable`. Use `ConfigConstant`, `ConfigConstants`,
  typed string slices, and `InsMap` for module-scoped or override tables.
- Keep config grammar behavior separate and compatible with Java. The TLC
  package should not import the root `tlago` parser package just to tokenize
  config files, because the root package will later need to wire parsed specs
  into TLC.
- Preserve duplicate-key errors, missing identifier errors, and keyword
  pluralization behavior.
- Preserve support for configs embedded in monolithic `.tla` files.

## Action Decomposition

`Tool.getActions` splits a next-state predicate for performance. Instead of
treating `Next` as one large action, Java TLC decomposes the maximum prefix of:

- disjunctions,
- bounded existentials,
- substitutions,
- selected user-defined operator bodies.

This yields separate `Action` objects, each with:

- `pred`: semantic predicate.
- `con`: context bindings.
- `actionName`: name used in traces and coverage.
- `opDef`: optional declaration node.
- `id`: stable action/init id.
- flags for init/internal actions.
- auxiliary metadata map.
- coverage cost model.

Correctness notes:

- Decomposition must not change enabledness or generated successor states.
- Action names in error traces and coverage depend on this decomposition.
- Bounded existential action splitting is observable through trace action names.
- Some specs have unnamed actions directly inside `Spec == Init /\ [][...]_vars`.
- Java distinguishes an action declaration from its definition: declaration is
  the left-hand-side `IdentLHS` tree-node location of the operator definition,
  while definition is the predicate/body source location. Coverage roots use the
  declaration; debugger, trace, DOT, and counterexample display use the
  definition through `Action.getLocation`/`getDefinition`.

Performance notes:

- More precise splitting avoids repeated evaluation of irrelevant disjuncts.
- Context allocation during splitting and next-state generation is hot.
- Go should avoid reflection or map-heavy inner loops here.

## Expression Evaluation

`Tool.eval` and `Tool.evalAppl` are TLC's interpreter for SANY semantic nodes.
It evaluates:

- constants and variables in contexts/states.
- user-defined operators and recursive operators.
- LET/IN and substitutions.
- operator arguments and higher-order operator values.
- builtin TLA+ operators through `BuiltInOPs` opcodes.
- sets, tuples, records, functions, function application, EXCEPT.
- bounded/unbounded quantifiers where supported by TLC.
- temporal/action operators when used by enabledness/liveness machinery.
- Java module overrides.

Core control inputs:

- `Context`: linked-list binding chain with branch/cutoff behavior for
  `ENABLED`.
- `s0`: current state.
- `s1`: successor/prototype state.
- `EvalControl`: control flags for priming, enabledness, and evaluation mode.
  The helper `PartialBoolean` mirrors Java enum order `YES, NO, MAYBE`; avoid
  relying on Go's zero value as semantic "maybe", and let invalid values fail
  loudly like Java's unreachable enum default path.
- `CostModel`: coverage accounting.

Important implementation patterns:

- `evalImpl` dispatches on semantic node kind.
- `evalApplImpl` dispatches on builtin opcode or user-defined operators.
- Values returned by evaluation are `Value` implementations and may be lazy.
- When a primed variable lookup cannot be resolved, Java reports an incomplete
  state only outside `ENABLED`. Under `ENABLED`, the same primed lookup must fall
  through to the ordinary undefined-operator handling so enabledness exploration
  can continue to use its special context/control flow.
- `GetVar` must recurse through `SubstInNode`, `APSubstInNode`, `LetInNode`,
  labels, lazy values, and operator definitions before deciding that an
  operator application is a state variable.
- `GetPrimedVar` has the same decomposition shape and only succeeds when the
  expanded expression is a prime application whose argument resolves through
  `GetVar`. This is how Java recognizes assignments hidden behind labels,
  substitutions, LET aliases, lazy values, or zero-arity operator definitions.
- `GetLevelBound` is only a conservative bound. It returns temporal/action
  constants immediately for temporal/action opcodes, treats `ENABLED` as state
  level, scans bounded-quantifier ranges and arguments, and follows user
  operator definitions, lazy values, `EvaluatingValue`, and `MethodValue`.
  For `LET`, Java first takes the maximum level of every local definition body,
  then binds each local operator name to `1` before checking the `IN` body; use
  the same pattern to avoid recursive level walks and to classify hidden
  higher-level local definitions conservatively.
  Recursive functions use the same `name -> 1` sentinel for the recursive
  function body.
- `setSource` associates semantic nodes with values for fingerprint exception
  diagnostics only in `CallStackTool`; `FastTool` and `DebugTool` leave values
  untouched. Attachment overwrites earlier sources and occurs at Java's explicit
  constructor, set-operation, `DOMAIN`/`UNION`, operator-argument, and subset
  assignment sites. Lambda source attachment precedes materialization. Replay's
  later lambda/predicate copy constructors deliberately drop the source.
- `CallStackTool` freezes only for TLC runtime, evaluation, and fingerprint
  exceptions; unrelated errors unwind and pop normally. Value-side exception
  wrapping remains a separate porting concern from source attachment. Primitive,
  explicit-set, tuple, record, finite-function, lazy-value, set-constructor,
  set-operation, model/special, and operator boundaries are ported. Returned
  errors and panics both preserve the concrete receiver in the value chain.
  Detailed checked/unchecked nested string APIs still need audit.
- Set-predicate membership has a separate inner EvalException-only rewrite;
  domain, binding, and predicate evaluation are inside that catch, and ordinary
  runtime/fingerprint failures pass to its outer source wrapper. Membership uses
  the no-cost-model evaluation overload; enumeration retains its model.
  Checked predicate printing catches expansion failures before symbolic fallback.
  Lazy deep normalization is Java's inherited no-op, and direct lazy evaluation
  uses the stored cost model.
- Shared operator methods retain the concrete receiver for diagnostics, source
  wrapping, and copying. Priority evaluating wrappers own their source/model
  and retain primary method metadata while their handle list is stably sorted.
  Method/evaluating initialization skips unsupported fingerprinting.
- MethodValue invocation preserves direct evaluation exceptions and wraps other
  failures at the Java method boundary before its outer source-aware catch.
  Evaluating/priority wrappers rewrite invocation and pure-fallback failures
  through a broad catch with no outer source wrapper; callable wrappers put
  argument evaluation outside their broad invocation/state-assignment catch.
  TLCError.Runtime distinguishes those runtime override failures from legacy
  native EvalException carriers, so predicate membership does not rewrite them.
- Fingerprint exception traces read each value's current source at trace time;
  sources are not snapshots taken when an exception head is created. Null
  source entries remain present in `asTrace` like Java.
- `FingerprintException.getTrace` is intentionally recursive: Java assigns
  labels while walking the linked exception head, then prints the recursive tail
  before the current frame. Preserve that order because the formatted trace is
  user-facing diagnostic text.
- Errors are not generic exceptions; they carry TLC error codes and source
  context.

Port guidance:

- Implement the evaluator against the Go SANY semantic tree, not against raw
  parse nodes.
- Keep a mechanical switch structure close to Java's `evalApplImpl` until the
  test suite passes.
- Preserve short-circuiting for boolean operators.
- Preserve Java's eager/lazy choices. Lazy values are not optional; they avoid
  explosive enumeration and support recursive/function semantics.
- Preserve Java's special `OPCODE_fa` branch: function records and function
  lambdas evaluate `args[1]`, while tuples and records first reject
  `f[e1, ... , eN]` when `N > 1`. The tuple/record argument must not be
  evaluated before that arity check.
- The initial-state, next-state, and `ENABLED` interpreters have their own
  `OPCODE_fa` paths. When the function expression evaluates to a
  `FcnLambdaValue` without a materialized function record, Java calls
  `getFcnContext` and recurses into the lambda body under that argument-bound
  context. This preserves symbolic state generation for predicates hidden
  behind function application.
- `getFcnContext` has its own binding and diagnostics algorithm; it must not
  delegate to `FcnLambdaValue` selection/application helpers. Java checks each
  parameter domain before converting tuple arguments, reports the relevant
  argument number and function-expression source, ignores excess arguments,
  and directly indexes short argument lists. Its single tuple-mismatch error
  prints `this.toString()`: object identity for ordinary tools, and the
  overridden stack string for `CallStackTool`. Lazy-function formatting must swallow
  expansion failures so diagnostic printing can fall back to symbolic text.
- In `evalApplImpl`, if a looked-up `LazyValue` is forced with `s1 == null`,
  Java evaluates the lazy expression directly with the lazy value's saved
  context and cost model. The cached path is reserved for the `s1 != null`
  branch through `LazyValue.getValue`.
- `tlc2.value.impl.LazyValue.off=true` disables LazyValue caching in Java by
  constructing every lazy value with the `UndefValue` sentinel. Go mirrors this
  through the same Java-style property key so lazy expressions still exist as
  thunks but never cache evaluated values.
- Preserve exact undefined-value behavior. TLC distinguishes "not enumerable",
  "undefined", "not comparable", and ordinary false in user-visible ways.

## Init and Next State Generation

Initial state generation flows through:

- `Tool.getInitStates()`
- `Tool.getInitStates(IStateFunctor)`
- `Tool.makeState`
- `Tool.getInitStatesAppl`
- action item lists that delay assignments and constraints.

Next-state generation flows through:

- `Tool.getNextStates(Action, TLCState)`
- `Tool.getNextStates(INextStateFunctor, TLCState[, Action])`
- `getNextStatesImpl`
- `getNextStatesApplImpl`
- `processUnchanged`
- assignment/action item list evaluation.

Go represents Java's `IStateFunctor` and `INextStateFunctor` with concrete
callback structs, not interfaces. Their unsupported operations should remain
explicit errors: Java has no default for `IStateFunctor.addElement` or
`INextStateFunctor.addElement(predecessor, action, successor)`, while
`setElement`, plain next-state `addElement`, and `hasStates` are default
unsupported operations.

Key semantics:

- State variables are assigned by primed variable equalities.
- Partial successor states are legal during construction.
- `UNCHANGED` copies values from current to successor.
- `ENABLED` uses a special context branch/cutoff mechanism.
- Java represents delayed conjuncts with positive action-list `kind` values,
  where the integer is the original conjunct position. The special sentinels
  are `0` for the base conjunct marker, `-1` for plain predicates, `-2` for
  `UNCHANGED`, and `-3` for changed expressions. Keep these numeric values
  exact rather than replacing them with an ordinary Go enum.
- Java's `ActionItemListExt` subclasses `ActionItemList` only to carry an
  action and a previous-node pointer across `cdr()` traversal. Go folds this
  into the single concrete `ActionItemList` struct with `act` and `prev` fields
  to avoid another type while preserving `getAction()` behavior.
- `ActionItemList.cons` must mirror Java's `ActionItemList.coverage` guard: it
  descends from the parent cost model with `cm.get(pred)` only when action
  coverage is enabled. When coverage is disabled, or only variable coverage is
  enabled through the coverage bitmask, it keeps the incoming cost model.
- Preserve Java's next-state action-list decomposition inside the port:
  public `GetNextStatesFromActionList` is the coverage wrapper,
  `getNextStates0` is the ordinary recursive dispatcher, and
  `getNextStatesAllAssigned` is the `TLCGlobals.warn && s1.allAssigned()`
  branch that evaluates remaining predicates directly while routing
  `UNCHANGED` and unsatisfied predicates through the same functor hooks as Java.
- Java recognizes `OPCODE_cdot` for action composition but disables it by
  default behind the `tlc2.tool.impl.Tool.cdot` system property. The Go port
  keeps the same default via `Globals.Cdot == false` and the same property key
  through `-Dtlc2.tool.impl.Tool.cdot=true` or `TLCSet`. It returns Java's
  unsupported-action-composition message from evaluation, next-state
  generation, and `ENABLED`. When explicitly enabled, `A \cdot B` first
  collects intermediate `s -A-> t` states into a concrete `StateVec`, then
  evaluates or generates `t -B-> u`. The next-state path uses `s0.CopyWith(s1)`
  for Java's partial intermediate state, forwards composed successors through a
  concrete `NextStateFunctor`, and stamps the original predecessor/action on
  `u`. The predicate-evaluation path intentionally starts from a fresh empty
  intermediate state, matching Java's `TLCState.Empty.createEmpty()` comment.
- Java `-generate` sets `tlc2.tool.impl.Tool.probabilistic=true`. In that
  mode, next-state generation randomizes disjunction order with the simulator
  RNG's `nextDouble` start index and `nextPrime` stride, randomizes bounded
  existential and assignment enumeration, and returns as soon as the next-state
  functor reports that a successor exists. If no successor is found in the
  randomized pass for assignments, Java falls back to ordinary sequential
  enumeration. `CASE` in the next-state relation remains deliberately
  unsupported in probabilistic mode.
- A complete state must assign every declared variable.
- `isGoodState` detects incomplete or illegal states.
- Model constraints and action constraints filter states but do not replace
  invariant/action-property checks.

The model checker's init path uses a `DoInitFunctor` to avoid materializing a
large `StateVec` of all initial states. Each init state is checked and inserted
as it is generated. Fresh runs print `TLC_COMPUTING_INIT` before the functor is
invoked and then print `TLC_INIT_GENERATED1` or `TLC_INIT_GENERATED2` with
Java's generated-state pluralization and distinct-state-count rule.
Successful model-checking runs print `TLC_SUCCESS` from the checker rather than
from the top-level runner. The payload mirrors Java's collision-probability
reporting: use only the optimistic probability when it is below `1E-10`;
otherwise also compute the observed probability from `FPSet.checkFPs()`. DFID
uses the two-probability path directly, matching Java.

Correctness notes:

- Invariants on init states are checked before queue insertion.
- Implied initial conditions are checked separately.
- Successor states that fail model constraints are not queued, but implied
  action checks may still apply depending on where filtering happens.
- An unseen successor is written to the trace before invariant/implied-action
  checks so a violation can print a trace.

Performance notes:

- Avoid allocating `StateVec` for huge init sets where the Java functor path
  streams states.
- Mutable state copying and partial assignment order are hot.
- Context enumeration for quantifiers and set membership is hot.

## State Representation

### `TLCState`

`TLCState` is an abstract assignment of spec variables to explicit values.
It carries:

- `workerId`: trace fragment owner.
- `uid`: pointer into per-worker trace file.
- `level`: state graph depth, with initial states at level `1`.
- static `vars`: ordered variable declarations.
- static `Empty` and `Null` sentinels.

Core methods:

- `Bind`, `Unbind`, `Lookup`, `ContainsKey`.
- `Copy`, `DeepCopy`, `CreateEmpty`.
- `AllAssigned`, `GetUnassigned`.
- `DeepNormalize`.
- `FingerPrint`.
- `ToString` variants for traces.
- predecessor/action/cached-value hooks used by extended/debug modes.

### `TLCStateMut`

`TLCStateMut` stores values in an array indexed by `UniqueString.varLoc`.
It is optimized for model checking:

- binding is O(1).
- copying is array copy.
- fingerprinting iterates variables in declaration order.
- values are deep-normalized before sharing through queues.
- optional VIEW expression replaces raw state fingerprinting.
- optional symmetry permutations choose the lexicographically smallest
  representative before fingerprinting.
- Java installs the active `Tool` into `TLCStateMut`/`TLCStateMutExt` during
  tool construction so state fingerprints can see the VIEW expression and
  symmetry permutations. The Go port keeps the same static state context
  through `SetTLCStateTool(tool)` and must call it when constructing model
  checkers or simulators. The order is important: first select the symmetry
  representative, then fingerprint either the raw representative values or the
  value produced by evaluating VIEW on that representative.
- Configured `SYMMETRY` is not only a boolean flag. Java evaluates the named
  zero-arity operator from the unprocessed definition table, evaluates its body
  with the processed constant environment, and installs the concrete
  `MVPerms.permutationSubgroup` result before fingerprints are taken. Go mirrors
  this in `SpecProcessor.processConfigSymmetry`, leaving `HasSymmetry` config
  based while `GetSymmetryPerms` returns only the concrete installed subgroup.
  Java also warns with `TLC_SYMMETRY_SET_TOO_SMALL` when a configured symmetry
  set has fewer than two useful model values; the Go port mirrors this for
  direct `Permutations(S)` argument shapes whose constant set value is available
  through the processed tool definition table.
- `TLCStateMut.toString` also honors VIEW, but only when the global `useView`
  flag is enabled; fingerprinting uses VIEW whenever the active tool has one.
- `setPredecessor` is also the level increment path. Java fails with
  `TLC_TRACE_TOO_LONG` when the predecessor is already at `Integer.MAX_VALUE`;
  Go keeps the same 32-bit ceiling even on wider `int` platforms.

### `TLCStateMutExt`

`TLCStateMutExt` is mostly a copy of `TLCStateMut` with extra fields for:

- predecessor,
- action,
- callable,
- cached values.

It is used for simulation, executor, debugger, and modes needing richer state
metadata. The Go port should avoid duplicating the whole implementation if it
can do so without changing behavior, but should preserve the two operational
modes.

Tricky details:

- State equality in Java is value-wise but `hashCode` is intentionally absent
  for mutable states. Do not place mutable states in Go maps keyed by struct.
- Fingerprint uses the configured VIEW and symmetry. This is state identity for
  visited-state purposes.
- Symmetry reduction is expensive: all permutations are applied to find the
  lexicographically smallest representative.
- `TLCGet("level")` depends on exact level initialization.

### `StateVec`

`StateVec` is Java's mutable array-backed state vector and implements both the
plain state functor and next-state functor contracts. The Go port keeps it as a
single concrete slice-backed struct in `state.go`:

- `AddElement(state)` appends the state unchanged.
- `AddNextElement(predecessor, action, state)` stamps predecessor/action before
  appending, matching Java's `INextStateFunctor.addElement`.
- `AddElements` returns the larger vector as the receiver of the append, just as
  Java swaps `s0`/`s1` to reduce copying.
- `Remove` swaps in the last element; `RemoveAt`/`Replace` only overwrite the
  slot. These two removal styles are intentionally distinct.
- `Reset` and `Clear` both set logical size to zero while keeping capacity.
- `Copy` calls state `copy`; `DeepCopy` calls state `deepCopy`.
- `ToRecords(append)` returns all current states as record values followed by
  the appended state.
- `ToRecordsFrom(from, append)` walks backward until `from`'s fingerprint,
  then returns that suffix in forward order followed by `append`, matching
  Java's linked-list `push` behavior.

## Value System

`tlc2.value.impl.Value` is the base class for all TLA+ runtime values.

Value kind constants:

- `BOOLVALUE`
- `INTVALUE`
- `REALVALUE`
- `STRINGVALUE`
- `RECORDVALUE`
- `SETENUMVALUE`
- `SETPREDVALUE`
- `TUPLEVALUE`
- `FCNLAMBDAVALUE`
- `FCNRCDVALUE`
- `OPLAMBDAVALUE`
- `OPRCDVALUE`
- `METHODVALUE`
- `SETOFFCNSVALUE`
- `SETOFRCDSVALUE`
- `SETOFTUPLESVALUE`
- `SUBSETVALUE`
- `SETDIFFVALUE`
- `SETCAPVALUE`
- `SETCUPVALUE`
- `UNIONVALUE`
- `MODELVALUE`
- `USERVALUE`
- `INTERVALVALUE`
- `UNDEFVALUE`
- `LAZYVALUE`
- `DUMMYVALUE`

Common operations:

- `Kind`
- `CompareTo`
- `Equals`
- `Member`
- `IsFinite`
- `Size`
- `Normalize`
- `DeepNormalize`
- `IsNormalized`
- `IsDefined`
- `DeepCopy`
- `FingerPrint`
- `Permute`
- `TakeExcept`
- `Apply`
- `Select`
- `Elements`
- `RandomElement`
- `ToString`
- binary serialization/deserialization.

Java value printing has two layers. `Value.toStringImpl` creates a buffer,
invokes the concrete printer, appends a delimiter, and catches sourced runtime
failures. Nested values call the concrete buffer overload with the same offset
and checked/unchecked flag. Go represents this with `ValueToString` and concrete
`ToString(*strings.Builder, int, bool)` methods; `StringWithDelimiter` and
`StringUnchecked` expose the Java string overloads without expanding the required
`Value` interface for foreign Go implementations.

Lazy printers preserve Java's individual `catch(Throwable)` regions. Cup, cap,
difference, and predicate-set expansion catch both materialization and printing;
product printers catch materialization and print the resulting set outside that
catch. Subset printers catch the size decision only, while `UNION` propagates
expansion failures. Unchecked calls rethrow instead of taking symbolic fallback.
All these paths share the original buffer and retain partial output after a
caught failure. Function lambdas always swallow expansion/printing failures and
request checked printing of a materialized function, including from unchecked
callers. Predicate-set symbolic domains and function-parameter domains invoke
standalone checked public string conversion. `TLC!Print`, `PrintT`, and `ToString`
use unchecked conversion; file `PrintT` appends its newline before pretty printing.

Important concrete values:

- `BoolValue`: singleton true/false.
- `IntValue`: cached small-ish integer values via factory.
- `StringValue`: backed by interned `UniqueString`.
- `TupleValue`: function from `1..n` to values.
- `RecordValue`: sorted/normalized field names and values.
- `SetEnumValue`: explicit finite set with normalization and duplicate
  handling.
- `IntervalValue`: finite integer interval without materializing all elements.
- Java's `Reducible` set operations live on concrete `IntervalValue` and
  `SetEnumValue` methods in Go. Difference reduces only its left operand;
  intersection and union prefer the left reducible operand, then the right.
  These methods use membership checks instead of enumerating the other side.
  Explicit-set results preserve raw left order and duplicates; union normalizes
  its reducible right operand by enumerating it and leaves the result
  unnormalized. Empty reducible unions return the other operand directly.
- `BaseValue.CM` mirrors Java's shared `Value.cm`. Evaluating/lambda/predicate
  values use that same field. Primitive and lazy set iterators preserve Java's
  secondary-count boundaries: explicit sets count calls after exhaustion,
  intervals count returned elements, filters count examined candidates, and
  materialization additionally counts produced elements. `UNION` uses the
  concrete optimization helper and inherits its input set's cost model.
- Product enumeration counts generated fields before advancing and attaches
  the original model to each tuple, record, or function. A function set with
  an empty domain counts its sole empty function once. Normalized powerset
  and k-subset enumeration attaches models without counting generated fields;
  materialization validates the full cardinality first and counts the final
  set's elements. Empty powerset bases produce the empty subset before trying
  set conversion. A null conversion result must remain distinct from a
  materialization failure.
- Java conversion accounting has asymmetric branches: interval-domain
  `FcnRcdValue.toTuple` constructs a tuple without the model or an increment,
  while explicit-domain `FcnLambdaValue.toTuple` increments even with global
  coverage disabled. Lazy-function materialization uses raw tuple casts and
  fixed declared-size arrays, installs the cached function with its model,
  then increments the materialization count before applying EXCEPT updates.
- Pointer-returning function conversions preserve Java exceptions using typed
  TLC error panics when a Go error return is unavailable. Normalization, size,
  selection, and materialization failures must not become null conversion
  results. `Tool.Eval`/`EvalAppl` turn those typed panics into returned errors;
  other runtime panics propagate. Eager construction fails when materialization
  fails instead of falling back to a lazy function.
- `FcnRcdValue`: explicit finite function, optimized for interval domains.
  Java's `tlc2.value.impl.FcnRcdValue.threshold` controls when normalized
  finite function records switch from linear lookup to binary search; Go keeps
  the default `32` and reads the same property key.
- `FcnLambdaValue`: lazy function with params/body/tool/context.
- Function constructors stay lazy unless Java's `Tool.evalApplImpl` can prove
  every bounded domain is `Reducible` and `EvalControl.KeepLazy` is clear. In
  Go, mirror that check concretely with `IntervalValue` and `SetEnumValue`
  rather than adding a one-use `Reducible` interface. Recursive functions also
  stay lazy. `DOMAIN` on `FcnLambdaValue` reads the lambda's parameter-domain
  metadata through `GetDomain`, not by materializing the whole function into a
  `FcnRcdValue`.
- `SetOfFcnsValue`, `SetOfRcdsValue`, `SetOfTuplesValue`: lazy enumerable set
  spaces.
- `SubsetValue` and `KSubsetValue`: lazy subset enumeration and unranking.
  Go's `SubsetUnrank` keeps strict predecessor lookups in a sorted cutoff
  vector, including replacement of duplicate binomial keys, and retains Java's
  Pascal-table indices and exact long conversion. The integer random-subset
  overload is `GetRandomSetOfSubsetsUpTo`; it uses decimal scale 32/HALF_DOWN
  allocation and Java's signed-long stride/offset behavior, including zero-
  modulus failures and random draws after the output bound has been reached.
  Fixed-cardinality `KElementEnumeration` requires counts to fit an int, while
  `NumberOfKElements` requires a long. Normalized powerset traversal retains
  its separate combinations iterator. Lexicographic bitset traversal replaces
  the outer base with its enumerated set before normalization and counts the
  sizes of emitted subsets for coverage. Empty traversal returns a fresh set
  after each reset. These inherited helpers also operate on Go k-subsets.
  `KSubsetValue` keeps Java's special comparison/equality shortcuts against
  ordinary `SUBSET S`: finite cardinalities are compared without materializing
  powersets, and `kSubset(0, S)` equals `SUBSET {}` independent of `S`.
- Lazy set constructors and set operations preserve Java's
  `SetEnumValue.DummyEnum` state after `deepNormalize` with explicit Go dummy
  flags. A dummy cache means "deep-normalized but not materialized"; when later
  converted to `SetEnumValue`, the materialized set must be deep-normalized
  before the dummy flag is cleared.
- `SetPredValue`: predicate-filtered set. `Tool` eagerly evaluates
  `{x \\in S : P(x)}` only when `S` is Java `Reducible`, represented by
  `IntervalValue` and `SetEnumValue` in Go. Other values retain a lazy
  `SetPredValue` even when they support enumeration; this avoids premature
  predicate evaluation and materialization of powersets or product spaces.
- `LazyValue` and `EvaluatingValue`: deferred evaluation.
- `ModelValue`: named atoms with special comparison/permutation behavior.
- `UserValue`: module-defined values such as unbounded standard sets.
- `UndefValue`: explicit undefined marker.

Value stream details:

- Java's `ValueOutputStream` handle table is a format-level compression
  detail, not a generic "all values get IDs" rule. Only the concrete `write`
  methods that call `vos.put` can later emit `DUMMYVALUE`: `StringValue`,
  `TupleValue`, `SetEnumValue`, `FcnRcdValue`, `RecordValue`, and record field
  `UniqueString` names. `BoolValue`, `IntValue`, `IntervalValue`, and
  `ModelValue` always write themselves directly.
- `ModelValue` serializes as kind `MODELVALUE` followed by the short index into
  the already-initialized global `ModelValue.mvs` table. The stream reader does
  not build that table on demand; a null table or invalid index retains Java's
  runtime failure rather than becoming an I/O error.
- Compound lengths follow Java's encodings: tuples and function records use
  compact naturals, set enumerations and records use signed lengths to preserve
  normalizedness, and function records write an interval-domain marker byte
  before either interval bounds plus values or explicit domain/value pairs.

- Output handles retain the actual objects for the stream lifetime, as Java's
  `Object[]` does, and preserve reference identity including null. Go pointer
  addresses alone are insufficient because reclaimed addresses can be reused.
  Input handles reserve sequential indices in a table starting at capacity 16;
  assignment/lookup use the allocated bounds and null entries remain null.
- Java uses the same write format for checkpoints and external values.
  `WriteExternal` is a Go convenience alias; external reads discard saved
  UniqueString token/location metadata, re-intern the text, and leave records
  unnormalized. Compact natural writes preserve Java's comment-only
  non-negative precondition without adding validation branches.
- Cached lazy-value writes do not materialize values. Null caches throw NPE,
  while a dummy set cache writes its shared sentinel handle and kind byte before
  failing on its null ValueVec. Unsupported values/kinds use Java's
  WrongInvocationException, and unknown kind bytes print as signed Java bytes.
- File-stream truncation produces a null-message EOFException, with Go
  `errors.Is(err, io.EOF)` compatibility for recovery loops. Invalid array sizes,
  indices, and handle casts retain runtime exception types. Queue byte input
  uses no handles (`getIndex=-1`, assignment is a no-op), rejects reference
  records, and throws array-bounds failures on truncated state bytes.

Correctness notes:

- Value comparison order is semantic and must match Java for normalization,
  symmetry representatives, sorted set/record/function printing, and
  fingerprinting.
- `Normalize` is often mutating in Java. Go can choose immutable or mutable
  internals, but must preserve sharing assumptions and performance.
- Lazy enumerable values must not enumerate huge spaces prematurely.
- `isEmpty` has special mathematical cases for function sets, record sets,
  tuple sets, subsets, k-subsets, and user values.
- Fingerprint exception wrapping uses value source metadata. Preserve enough
  metadata to emit equivalent diagnostics.

Performance notes:

- The value system is the heart of TLC's runtime. Avoid interface dispatch in
  very small arithmetic/boolean paths if it becomes measurable, but first keep
  behavior obvious and close to Java.
- Enumeration should be streaming where Java streams.
- Preserve interval/function/set lazy representations to avoid catastrophic
  memory growth.

## Contexts and Enumerators

`Context` is a linked list of `(SymbolNode, value)` pairs:

- `Empty` is the base context.
- `branch(base)` inserts a null-name marker used by `ENABLED`.
- `lookup(var, cutoff)` stops at a branch marker when cutoff is true.
- `cons(name, value)` prepends a binding.

Context is used for:

- level boundedness during processing.
- runtime lexical/operator bindings during evaluation.
- quantified variables and function parameters.
- enabledness branch scoping.

`ContextEnumerator` enumerates bindings for bounded quantifiers and set-based
operator parameters. Enumeration order is user-visible in traces, random
simulation, and sometimes coverage, so preserve Java order.

## Utility Collections

`tlc2.util` contains deliberately small custom collections and IO helpers used
by values, queues, traces, liveness graphs, and fingerprint sets.

Porting guidance:

- Keep deterministic iteration order for any structure whose order reaches
  fingerprints, traces, XML/JSON output, or diagnostics.
- Java's small hash tables (`ObjLongTable`, `LongObjTable`, and the
  semantic-node variant used by primed-location coverage) expose probe-slot
  effects through returned indices and slot-order enumeration. Do not replace
  those paths with ordinary Go maps; mirror Java's `count/length/thresh`
  arrays, linear probing, `2*length+1` growth, and physical-slot scans.
- Prefer concrete Go structs. Java's `ExternalSortable` has `BigInt` as its
  practical TLC implementation, so the Go external-sort helpers operate on
  `[]*BigInt` directly instead of creating a one-implementation interface.
- Do not create public pseudo-abstract base structs whose methods rely on Go
  method dispatch that does not exist. Java abstract bases such as
  `ObjectStack` should become private shared fields plus concrete methods on
  `MemObjectStack` and `DiskObjectStack`.
- Port utility tests early. These are cheap, stable conformance checks and do
  not disturb the mechanical core model-checker port.

## Fingerprinting

`tlc2.util.FP64` implements 64-bit fingerprints over GF(2^64):

- `New()` starts from selected irreducible polynomial.
- `Extend` handles strings, chars, bytes, ints, and longs.
- `Polys` contains selectable polynomial constants.
- `TLC` selects `fpIndex` unless user specifies it.

Heap-backed disk fingerprint sets allocate Java-style striped read/write locks
at construction. The default stripe count is
`2^(floor(log2(NumWorkers)) + 8)`, with `tlc2.tool.fp.DiskFPSet.logLockCnt`
or `TLAGO_DISK_FPSET_LOG_LOCK_CNT` overriding the exponent. Go `Put` and
`Contains` use those stripes, while flushing acquires all stripes. Java's
flusher recursively acquires its already-held write lock; Go skips that stripe
during the all-stripe acquisition because its mutexes are not reentrant.
Disk fingerprint reads mirror Java `IdThread.GetId(braf.length)`: a goroutine
with a current worker id uses its corresponding fixed `BufferedRandomAccessFile`
reader; calls outside worker scope fall back to the reader pool.

Trace level is trace-authoritative when extending partial state spaces:
`CheckImpl.makeStateSpace` uses `TLCTrace.getLevel(state.uid) + depth` in Java,
so Go's `CheckImpl.MakeStateSpace` uses `TLCTrace.GetLevel(state.UID)` rather
than trusting the mutable state's cached level or mirrored object identity.

Visited-state identity is a 64-bit fingerprint of:

- normalized state values in declared variable order,
- or VIEW value when a view is configured,
- after symmetry representative reduction when symmetry is configured.

Correctness requirements:

- Go FP64 must match Java exactly for all primitive extension functions,
  including Java signed-byte and unsigned-shift behavior.
- Every `Value.FingerPrint` method must emit the same extension sequence as
  Java.
- State-level fingerprinting must normalize the same values at the same points.
- Collision handling is probabilistic just like Java TLC. Do not "improve" it
  by using full state keys unless it is explicitly a debugging mode.

## Fingerprint Sets

`FPSet` is the abstract concurrent set of 64-bit fingerprints. `put(fp)` returns
true if the fingerprint was already present and false if it was newly inserted.

Implementations:

- `MemFPSet`: synchronized in-memory hash table of buckets. Rehashes by
  doubling table capacity and splitting buckets by one bit.
- `MemFPSet1`: deprecated memory variant backed by `SetOfLong` with Java's
  open-addressing table and checkpoint format.
- `MemFPSet2`: deprecated memory variant with a `2^24` spine. The low 24 bits
  are encoded by the bucket index and only the five higher bytes are stored in
  each collision bucket.
- `DiskFPSet`: bounded memory plus sorted disk backing file with per-worker
  buffered random-access readers and reader/writer locking.
- `MSBDiskFPSet`, `LSBDiskFPSet`, `HeapBasedDiskFPSet`,
  `OffHeapDiskFPSet`: disk/off-heap variants with different indexing choices.
- `MultiFPSet`: partitions fingerprints by high-order bits over nested FPSets.
- `NoopFPSet`: testing/no-op behavior.
- distributed wrappers and managers for RMI-based distributed TLC.
- `OffHeapDiskFPSet` uses open addressing over `LongArray` and chooses one of
  three indexers: bit-shifting for power-of-two position counts, a multiply-high
  1024MiB-multiple indexer, or an exact/infinite-precision fallback. The Go port
  keeps these as one concrete `OffHeapIndexer` with a kind field.
- In the off-heap primary table, `0` means empty, a positive fingerprint means
  not yet evicted, and the same value with the high bit set means already
  evicted to the sorted disk file. Flushed slots still detect duplicates and can
  be reused for different fingerprints.

Factory/configuration behavior:

- Java property `tlc2.tool.fp.FPSet.impl` selects the implementation class;
  the Go port reads that environment key plus `TLAGO_FPSET_IMPL`.
- `getImplementations()` advertises `MSBDiskFPSet`, `LSBDiskFPSet`, and
  `OffHeapDiskFPSet`, with `MSBDiskFPSet` as the default.
- VM argument recommendations split on storage type: heap-based sets use
  `-Xmx`, while off-heap sets use `-XX:MaxDirectMemorySize`.
- `FPSetConfiguration.getMemoryInBytes()` applies the ratio to explicit memory
  too, matching Java's call to `TLCRuntime.getFPMemSize(memoryInBytes * ratio)`.
- `MultiFPSet` creates nested sets through a Go `NewMultiFPSetConfiguration`
  helper corresponding to Java `MultiFPSetConfiguration`: it copies the parent
  config, disables nesting, and divides memory across `2^fpBits` children.

Checkpointing:

- `beginChkpt`: write temporary snapshot.
- `commitChkpt`: atomically rename tmp to checkpoint.
- `recover`: load checkpoint or rebuild from trace.
- `recoverFP`: insert a recovered fingerprint and assert it was absent.
- `DiskFPSet` no-argument checkpoint methods are silent no-ops because Java
  rebuilds disk-backed fingerprints from the trace. `MultiFPSet`'s no-argument
  checkpoint methods forward to child no-argument methods, so they stay silent
  for disk-backed children. `NonCheckpointableDiskFPSet` only warns on the named
  overloads used by nesting/distributed surfaces; the warning says checkpointing
  is not implemented for the concrete Java class name.

Port guidance:

- Start with `MemFPSet` and `MultiFPSet`, then disk variants. This is an
  implementation staging order, not a semantic excuse.
- Match `put` return polarity exactly.
- Preserve thread-safety.
- Preserve checkpoint file semantics for later Java test parity.
- Disk/off-heap implementations matter for performance and large models; do not
  permanently replace them with maps.
- `DiskFPSet` keeps persistent `BufferedRandomAccessFile` readers over the
  backing `.fp` file. After a flush merges into a temporary file and replaces
  the old backing file, all readers must be closed and reopened, otherwise
  lookups can keep reading the old file handle.

## State Queues

`IStateQueue` abstracts the frontier of unexplored states.

Implementations:

- `StateQueue`: synchronization and worker coordination base class.
- `MemStateQueue`: in-memory queue.
- `StateDeque`: deque variant; Java adds every state to the front and polls
  from the front, so this is LIFO search order and it does not support
  checkpointing.
- `DiskStateQueue`: two in-memory buffers plus state-pool disk files.
- `DiskByteArrayQueue`: byte-array backed disk queue that serializes states
  before taking the queue lock and stores raw state bytes in its disk buffers.
  Java gives this queue the same reader/writer/cleaner choreography as
  `DiskStateQueue`; the Go port keeps a concrete byte-array cleaner rather
  than generalizing through an interface.
- `SynchronousDiskIntStack`: real disk-backed integer stack used by utility
  code. Java's `DiskIntStack` is documented as an unused asynchronous sketch.

`StateQueue` behavior:

- Queue selection is normally `DiskStateQueue`. Java property
  `tlc2.tool.queue.IStateQueue` selects `MemStateQueue`, `StateDeque`, or
  `DiskByteArrayQueue`; the Go port reads that environment key and the
  command-line-friendly alias `TLAGO_STATE_QUEUE`. The legacy Java boolean
  `tlc2.tool.ModelChecker.BAQueue` is mirrored by
  `TLAGO_MODEL_CHECKER_BAQUEUE`.
- `sEnqueue` adds states and wakes waiting workers.
- The array/slice enqueue overload enqueues every entry and does not filter
  `null`/`nil`; callers are expected to pass dense state arrays. The `StateVec`
  overload is different and skips nil elements. Keep this distinction even
  though it is easy to accidentally collapse the two in Go.
- `sDequeue` blocks when empty until work appears or all workers are waiting.
- `finishAll` terminates all workers and wakes main/checkpoint waiters.
- `suspendAll` stops workers at a barrier for checkpointing.
- `resumeAll` releases the checkpoint barrier.
- `resumeAllStuck` handles distributed worker death cases.

Important synchronization details:

- `numWaiting >= numWorkers` and empty queue means no work remains.
- `numWaiting >= numWorkers` while `stop` is true but the queue is not empty
  means checkpoint suspension has reached its barrier. Java notifies the
  separate `mu` monitor in this case; the Go queues broadcast on their concrete
  condition variable so `SuspendAll` can return and checkpoint the still-live
  queue.
- `finish` is volatile in Java to avoid checkpoint race deadlocks.
- `suspendAll` uses a second monitor `mu` to coordinate checkpoint waiting.
- Lock ordering is deliberate. Preserve it when translating to Go mutex/cond.

Disk queue behavior:

- `enqBuf` fills then spills to disk through `StatePoolWriter`.
- `deqBuf` drains then refills through `StatePoolReader`, which can return a
  prefetched full buffer, synchronously read a pending pool file, or fall back
  to the in-memory `enqBuf`.
- State-pool files are full-buffer writes. Java calls `TLCState.write` on every
  pool slot and therefore fails on a null entry; Go must not substitute empty
  states for nil slots because that hides queue corruption.
- A cleaner thread deletes old pool files once `loPool - lastLoPool > 100`.
  Checkpointing stops the cleaner before writing queue metadata, and after the
  first checkpoint the checkpoint commit path owns obsolete pool-file deletion.
- `DiskByteArrayQueue` uses the same cleaner threshold and checkpoint handoff
  for liveness/raw-state pool files.
- checkpoint includes queue metadata and buffered states.

Port guidance:

- Start with a faithful condition-variable based `StateQueue`.
- Recreate the Java blocking/termination semantics before optimizing.
- Add disk queue once value/state binary serialization exists.

## Model Checking Algorithm

`ModelChecker` extends `AbstractChecker`.

Main fields:

- `theFPSet`: reachable fingerprints.
- `theStateQueue`: frontier.
- `trace`: concurrent trace for error reconstruction.
- `workers`: exploration workers. Java allocates these in the constructor,
  before initial-state generation: worker 0 keeps the normal/debug-capable tool
  and later workers use `tool.noDebug()`.
- `liveCheck`: liveness subsystem. Java `AbstractChecker` creates
  `NoOpLiveCheck` only when `tool.livenessIsTrue()` is true; otherwise it
  warns for liveness plus symmetry and constructs `LiveCheck`, whose
  constructor immediately calls `Liveness.processLiveness(tool)`. The Go port
  mirrors this in `NewModelChecker` with a concrete `LiveCheck`; callers may
  still inject a prebuilt `LiveCheck` through options for tests.
- `errState`, `predErrState`, `errorCode`, `done`, `keepCallStack`.

High-level flow in `modelCheckImpl`:

1. Attempt recovery from checkpoint.
2. If starting fresh:
   - check assumptions,
   - compute initial states through `doInit`,
   - check init invariants and implied init conditions,
   - insert good in-model init states into FP set, queue, trace, liveness.
3. If no actions exist:
   - success when queue is empty,
   - otherwise error for states with no next action.
   - Java calls `cleanup(true)` in both no-action cases, even when the queue is
     non-empty and the result is `TLC_STATES_AND_NO_NEXT_ACTION`; Go preserves
     that single success-cleanup call instead of using the returned error code
     to keep artifacts.
4. Run worker exploration with `runTLC`, which starts all pre-created workers,
   periodically performs coordinator work while they run, and joins them after
   the shared queue reaches completion or an error.
5. During exploration:
   - workers dequeue states,
   - generate all action successors,
   - detect incomplete states,
   - apply model/action constraints,
   - insert unseen fingerprints,
   - write trace before invariant/implied action checks,
   - enqueue only unseen in-model non-violating states,
   - add behavior graph edges for liveness,
   - detect deadlock.
6. Run final liveness check if needed.
7. Check postconditions.
8. Print success/error summary, coverage, statistics.
9. Cleanup closes the FP set, trace, liveness checker, and state writer, then
   deletes metadata when configured and not vetoed by checkpoint-preservation
   rules.

Java wraps `modelCheckImpl` with `AbstractChecker.modelCheck`, which returns
the stored `errorCode` when the implementation result is `NO_ERROR`. This is
observable after kept-call-stack worker failures because `runTLC` deliberately
returns `NO_ERROR` so the outer code can replay the error. Go keeps the same
wrapper precedence after the cleanup decision, so cleanup still sees the
implementation result Java would have used.

`doNext` error precedence:

- incomplete successor state,
- unseen successor invariant violation,
- implied action violation,
- deadlock after all actions produce no successors,
- evaluation exceptions with call-stack replay where configured.

Continuation mode:

- Some invariant/action violations can be printed and exploration continues.
- Without continuation, first error terminates all workers.

Port guidance:

- Implement worker/error synchronization carefully. Only one worker should own
  the primary error trace unless continuation mode says otherwise.
- Java holds the checker monitor across next-state error acceptance, trace
  printing, and worker error-time postconditions, and across continuation-mode
  violation printing. Go serializes those reports with `ModelChecker.nextErrorMu`
  while using the existing state mutex for checker fields. The separate report
  mutex lets ALIAS and postconditions query/control the checker without requiring
  reentrant Go mutexes. The postcondition wrapper calls a lock-held error helper
  and reads `Done` under the state mutex before deciding the counterexample's
  `console` field.
- Preserve Java's `AbstractChecker.runTLC` coordinator role: workers do state
  generation, while the checker thread periodically suspends the queue for
  liveness/checkpoint work and then resumes or finishes the workers.
- Positive `runTLC(depth)` limits are enforced by that coordinator after
  progress reporting: if the reported level exceeds `depth`, it finishes the
  shared state queue. Workers must not skip next-state generation locally based
  on depth; that would change Java's scheduling, trace, and liveness handoff.
- Periodic coverage reporting is also driven by the coordinator, using
  `coverageInterval / progressInterval` as Java does. This is separate from the
  final coverage report emitted when model checking exits.
- `_PERIODIC` is part of that coordinator loop. Java treats a configured
  periodic expression as sufficient reason to suspend workers, evaluates it
  after any liveness work and before checkpointing, and returns
  `TLC_ASSUMPTION_FALSE` only when the evaluated value is exactly `FALSE`.
- `LiveCheck.doLiveCheck()` only decides whether liveness itself should trigger
  an expensive queue suspension. If checkpointing or `_PERIODIC` already caused
  the suspension, Java still runs `liveCheck.check` whenever liveness is enabled
  and the runtime ratio is below the configured liveness ratio.
- Java's `tlc2.TLC.stopAfter` property is a time-bound escape hatch for both
  model checking and simulation. When configured, a timer calls the active
  checker or simulator stop method. Cleanup keeps explicit checkpoint data if
  unexplored work remains and either an error was found or the run was
  time-bound.
- Simulation keeps `seed` and `aril` as separate user-visible values. Java
  seeds the simulator RNG with `seed`, advances it by `aril` `nextDouble()`
  calls, and seeds `RandomEnumerableValues` with the original `seed`, not
  `seed + aril`. `TLCGet("config")` must report the original pair.
- Java treats seed absence, not seed value zero, as the cue to draw a seed from
  a no-argument `RandomGenerator`. An explicit `-seed 0` must be replayable as
  zero. The no-argument generator follows `java.util.Random()` seeding via the
  `seedUniquifier() ^ nanoTime` pattern; DFID uses the same constructor before
  reseeding its worker RNG from `nextLong()`.
- Java's `tlc2.tool.ModelChecker.vetoCleanup` property forces metadata
  retention even when `-cleanup` was requested; Go mirrors it with the same key
  plus `TLAGO_MODEL_CHECKER_VETO_CLEANUP`.
- Cleanup has one deliberately load-bearing Java quirk: `FileUtil.deleteDir`
  receives the checker cleanup `success` argument as its `recurse` argument.
  When cleanup is called with `success=false`, Go must close checker resources
  but only attempt a non-recursive metadir delete, so a non-empty metadir with
  trace/checkpoint artifacts is preserved for debugging. When cleanup is called
  with `success=true`, Go may recursively remove the metadir. Do not copy
  Java's likely early-return cleanup leak, but do preserve this artifact
  retention rule. The unit test `TestModelCheckerCleanupPreservesFailureArtifactsLikeJava`
  is intentionally small and high-signal because this behavior is easy to
  "simplify" incorrectly.
- Preserve when traces are written relative to checks.
- Preserve generated-state counters versus distinct-state counters.
- Preserve final liveness check behavior even when no safety error occurs.
- Preserve Java's `doNextFailed` keep-call-stack polarity. Generic next-state
  failures call `setErrState(cur, succ, true, ec)` in the ordinary case; Java
  only flips that flag for stack overflow, out-of-memory, and assertion errors.
- Java `Worker.addElement(cur, action, succ)` wraps callback exceptions in a
  `WrappingRuntimeException` carrying `succ`. Go mirrors this with the concrete
  `workerNextStateError` struct so outer worker error handling can pass the
  partially generated successor to `doNextFailed` without adding an interface.
- Java exception catches and `instanceof` classify the thrown object itself;
  they do not inspect its cause chain. Go checker/call-stack/simulator/liveness
  consumers use direct concrete type checks and explicitly handle the embedded
  `LiveCounterExampleException` subclass. Fingerprint causes remain accessible
  for dedicated root-cause diagnostics. A standalone stateful exception does
  not enter the worker's EvalException/TLCRuntimeException liveness replay catch.
- Debugger control exceptions and distributed worker exception reuse also
  match only the directly thrown class. Debugger exception pushes set the
  inherited known flag once; evaluation frames catch only evaluation/runtime
  TLC exceptions. Distributed worker catch coverage includes fingerprinting
  and state routing, preserving Java's current predecessor/successor fields
  and resetting the computing flag even when a Go operation panics.
- Throwable detail messages are represented separately from Go `Error()` by
  `GetMessage() *string`: null messages and empty strings have different
  diagnostic behavior. Constructors capture Go runtime frames for Java-style
  stack rendering, including actual causes and common-frame elision. A
  fingerprint exception's `next` field is a semantic trace link, not the
  Throwable cause printed by `printStackTrace`.
- Nullable parameter arrays survive in exception carriers and
  `Message.NullableParams`. `PrintErrorNullable` preserves recorder metadata
  and Java MP's stop-at-first-null placeholder substitution. Native MethodValue
  alone substitutes a stack trace for a null detail message; evaluating and
  callable overrides keep the null array entry. GENERAL stack suffix control
  uses Java MP's historical `noDebug` startup property independently of `-debug`.
- Simulation invariant/action evaluation catches construct a dedicated worker
  error without setting its fatal exception field. Its property name and caught
  nullable detail message print with the evaluation-failed code, and the
  simulator's non-continuable-error policy stops the run.
- `TLC.process` and the simulator report EvalExceptions through GENERAL; only
  direct TLCRuntimeExceptions preserve their code and parameters there.
  Simulator exceptions overwrite an earlier worker code and stop simulation.
  `TLCError.Runtime` marks explicit Java runtime carriers; existing native
  evaluation carriers and direct system-error carriers remain distinguished.
- Call-stack replay uses a synthetic worker id `4223` like Java, but that
  worker is deliberately not registered with the shared trace. Java's comment
  says replay must not rewrite the trace file while reconstructing states for
  diagnostics. Go preserves this with a concrete `Worker.DisableTraceMirror`
  flag: replay may use its own worker trace context, but it must not append to
  `ModelChecker.Trace` through the normal worker mirror hook.
- Liveness `addNextState` call-stack replay has an additional Java bug path:
  if ordinary liveness evaluation fails but the `CallStackTool` rerun succeeds,
  Java calls `Assert.fail(EC.GENERAL, origExp)` because replay success "should
  never happen." Go preserves that as a general TLC error rather than silently
  reporting the original liveness error code.
- Preserve Java's error-time postcondition behavior: init failures with an
  `errState` call `checkPostConditionWithCounterExample(new CounterExample(errState))`,
  and worker `doNextSetErr` paths build a safety counterexample, evaluate aliases
  over that trace, and call the postcondition hook before returning the safety
  error. Java's direct `ModelChecker.doNext` helper is different: its
  `doNextSetErr` path prints and terminates but does not invoke the
  postcondition hook. Evaluation failures remain on Java's separate
  `doNextEvalFailed` path.

### CheckImpl and CheckImplFile

`CheckImpl` is Java TLC's implementation checker. It reuses the model-checking
engine to build a partial state space from a supplied starting state, tracks
covered states in `coverSet`, keeps a current-state enumerator, and exports new
traces to states that have not yet been covered by the external implementation.
It is not a normal exhaustive run: the checker repeatedly imports a concrete
trace, verifies that each adjacent state pair is reachable in the spec, marks
the visited states, and emits another trace when uncovered behavior remains.

`CheckImplFile` is the file-backed variant used to communicate with an external
simulation engine. It reads trace inputs named by a prefix plus a monotonically
increasing counter, and writes exported traces using the same prefix with
`_out_` plus a separate output counter. The output format is deliberately plain:
`STATE_n` headers followed by each TLC state's text form.

Java parses input trace files as TLA+ modules and turns operation bodies into
states through `Tool.makeState`. The Go port keeps `CheckImplFile` concrete and
uses a `LoadTraceFunc` hook at this layer so the `tlc` package does not import
the production SANY parser. The root `tlago` package wires that hook with
`NewCheckImplFileTraceLoader`: it parses the exact trace filename rather than
adding `.tla`, converts each root-module operator definition body through the
same SANY-to-TLC bridge used for normal model checking, temporarily installs the
trace definitions for helper-definition lookup, calls `Tool.MakeState`, then
restores the main tool definition table. The command option parser mirrors
Java's `CheckImplFile.main` surface: `-config`, `-deadlock`, `-recover`,
`-workers`, `-depth`, `-trace`, `-coverage`, and the root module; the root CLI
exposes it as `tlago checkimplfile`. On recovery runs, the CLI restores the
`UniqueString` checkpoint before parsing/building the tool and resets FP64 to
polynomial index 0 before construction, matching Java's `CheckImplFile.main`
ordering before `new FastTool(...)`.

## DFID Architecture

`DFIDModelChecker` is the depth-first iterative-deepening checker. Java keeps it
separate from the breadth-first `ModelChecker` and runs only with one worker and
without liveness checking.

Important structures:

- `MemFPIntSet`: fingerprint table that stores both membership and DFID level
  status bits.
- `DFIDWorker`: owns explicit stacks for states, fingerprints, successor
  vectors, and successor fingerprints up to `DFIDMax`.
- The worker randomly chooses an unfinished initial state and then randomly
  chooses unfinished successors from the current depth stack.

Important behavior:

- Each outer iteration increments `FPIntSet`'s global level and searches up to
  that level.
- `doNext` for DFID is a one-step generator. It fills successor vectors with
  states not completed at the current level and returns whether all successors
  were non-leaf.
- DFID initial-state processing records new in-model states before checking
  invariants and implied-init properties, matching Java's array/write/liveness
  ordering. Initial states excluded by model constraints are still checked
  against invariants and implied-init properties because Java leaves their
  `status` as `FPIntSet.NEW`.
- DFID `setErrState` stops the DFID workers after the abstract checker accepts
  the error state.
- Backtracking marks fingerprints leveled. States are marked done when all
  children are done or when a leaf has no new child.
- Keep DFID as worker-plus-stacks, not recursive traversal; recursion diverges
  from Java's scheduling, status updates, and randomization points.

## Worker and Trace Architecture

`Worker` is both a thread and an `INextStateFunctor`.

Worker loop:

1. `sDequeue` current state from the shared blocking queue. A `null`/`nil`
   dequeue means all workers are waiting and the frontier is empty; the worker
   sets checker completion and finishes the queue.
2. Set thread-local current state.
3. Allocate `SetOfStates` if liveness/debug mode needs successor collection.
4. Ask `Tool` to generate successors through the functor callback.
5. Detect deadlock if no successors were generated.
6. Add the stuttering self-loop to the liveness successor set, write the
   `Visualization.STUTTERING` transition to the all-state writer, and add the
   collected successor set to the liveness graph with one
   `LiveCheck.addNextState` call for the predecessor.
7. Record out-degree statistics.

Constrained dumps:

- When a generated successor is excluded by state or action constraints and
  the state writer is in constrained mode, Java rechecks each configured
  constraint and writes one transition per failed constraint with that
  constraint as the DOT reason. Generic not-in-model output is only a fallback
  when no individual reason can be identified.

Trace writing:

- Each worker owns a trace fragment file named by spec and worker id.
- Java increments the worker out-degree counter in `writeState(curState, fp,
  succState)`, immediately after recording an unseen successor in the trace.
  Keep that placement: the counter describes unseen successors discovered while
  evaluating `Next`, even if later invariant or implied-action checks stop the
  run before the state is enqueued.
- Java trace recovery for transition errors treats the stored trace as a prefix:
  `ConcurrentTLCTrace.getTrace(state)` excludes `state` itself except for the
  initial-state special case. Error printing and postcondition counterexamples
  then recover `curState` and `succState` explicitly through `Tool.getState`.
  Preserve the same prefix-plus-explicit-state construction; otherwise
  non-initial transition errors silently drop the predecessor state.
- The Go port now keeps a concrete `ConcurrentTLCTrace` on `ModelChecker`,
  writes Java-style worker fragment files as the authoritative model-checking
  path, and mirrors records into the existing in-memory `TLCTrace` for current
  Go callers.
- Error trace reconstruction follows Java's two-phase model: walk worker
  records backward to collect fingerprints and predecessor locations, then ask
  `Tool.GetState` to regenerate states forward from the initial fingerprint.
- Java backs worker trace fragments, `TLCTrace`, disk FP sets, and bit-vector
  persistence with `BufferedRandomAccessFile`. The port should preserve its
  concrete buffer state (`dirty`, `length`, `curr`, `lo`, `diskPos`, `mark`),
  8K page size, `seeek` page-read signal, and nat encodings.
- Initial state record contains previous pointer `1`, worker id, fingerprint.
- Successor record contains predecessor pointer, predecessor worker id,
  successor fingerprint.
- The successor state receives `(workerId, uid)` for later reconstruction; the
  `workerId` is the generating worker, not the predecessor's worker.
- Reads/writes are synchronized so trace printing sees consistent fragments.

Trace reconstruction and aliasing:

- Java exposes `TraceApp`/`ITool` methods to reconstruct an initial state from
  a fingerprint, reconstruct a successor from a predecessor plus fingerprint,
  and reconstruct transition metadata from a successor/predecessor pair.
- The Go port keeps those duties as concrete `Tool` methods:
  `GetInitState`, `GetStateAfter`, and `GetStateForTransition`; no Go
  `TraceApp` interface is needed.
- Java regenerates states from the init and next-state predicates. The Go port
  first checks the in-memory `KnownStates` registry when present to preserve
  identity for existing trace records, then falls back to Java-style
  regeneration.
- Unnamed actions carry Java's literal `UnnamedAction` sentinel while
  `Action.isNamed()` remains false for that sentinel. Printed
  `TLCStateInfo.info` values are action locations (`<Action ...>` or named
  variants), and unnamed initial actions use `<Initial predicate ...>` rather
  than dropping source context.
- Action locations render all non-nil bound formal parameters with ordinary
  string conversion, matching Java `Action.getLocation`. Counterexample action
  records still use `Action.getParameters`, which includes only parameter
  bindings that are TLC `Value`s. The location portion uses the action
  predicate/body source range, not the operator declaration range. The
  declaration range is retained separately on `OpDefNode` for coverage, matching
  Java `Action.getDeclaration`.
- Base `TLCStateInfo.getStateNumber()` returns the stored trace ordinal.
  Java's `AliasTLCStateInfo` overrides it to return `originalState.getLevel()`;
  the Go alias path keeps the retained original state and follows that override.
- Before replaying trace fingerprints into concrete states, Java snapshots and
  resets `RandomEnumerableValues`; restore the snapshot after replay so specs
  using randomized enumeration regenerate the same path without perturbing the
  active runtime RNG.
- Java's no-alias behavior returns the current state/info, not the successor.
  Go `EvalAlias`, `EvalAliasInfo`, and `EvalAliasInfoPair` preserve that.
- The Java default `evalAlias` overloads evaluate the resolved `ALIAS` operator
  body under `EvalControl.Clear`, convert record-like values to alias states,
  and on evaluation errors attach `_ALIASEvalError` to an alias record instead
  of throwing. Go keeps this concrete on `Tool`, not behind a separate alias
  interface. Prefix/suffix trace aliases bind `TLCExt!Trace` through the
  standard operator's `OpDefNode` symbol so aliases using `Trace()` see the
  supplied prefix records.
- Safety-error postcondition traces run Java's pairwise alias evaluation over
  each trace entry before wrapping the trace in `CounterExample`, matching
  `Worker.doPostCondition`'s `evalAlias(current, successor)` overload. Do not
  bind `TLCExt!Trace` on this path. Prefix/suffix trace aliases belong only to
  Java paths that explicitly call the prefix overload, such as checker
  `TLCTrace.printTrace` and liveness/error-trace reconstruction. The `console`
  field is omitted only when the checker was not already done at the moment the
  error was accepted.
- Java also calls `TLCTrace.printTrace(curState, succState)` for next-state
  safety errors and for continuation-mode invariant/action-property violations.
  Go mirrors this with `ModelChecker.printBehaviorTrace`, which prints
  `TLC_BEHAVIOR_UP_TO_THIS_POINT`, aliases each recovered state, and emits
  state-printer messages in ordinal order before continuing or finishing the
  queue.
- Simulator behavior printing is different: Java `Simulator.printBehavior`
  constructs `TLCStateInfo` directly from each stored simulated state and calls
  pairwise `evalAlias(current, successor)`, so Go simulator printing must not
  bind `TLCExt!Trace` either.

`ConcurrentTLCTrace` merges per-worker trace fragments to reconstruct:

- safety counterexamples,
- liveness lassos,
- trace expressions,
- alias-enhanced traces.

Port guidance:

- Go goroutines can replace worker threads, but trace fragment ownership and
  synchronization must stay equivalent.
- Normal `TLCTrace` writes the Java-compatible `MC.st`-style file: each record
  is `longNat(predecessorPointer)` followed by the state fingerprint, and
  checkpoints store only the current file pointer and `lastPtr`.
- Java's static `TLCTrace.writeBehavior(File, TLCState, StateVec)` writes a
  compressed `ValueOutputStream` containing one non-normalized `TupleValue`.
  Each tuple element is `RecordValue(TLCState)` for the corresponding trace
  state. The Go helper `TLCTraceWriteBehavior` mirrors that format so
  `IOUtils!IODeserialize` can read serialized behaviors.
- `TLCState` disk serialization follows Java field order: `shortNat(workerId)`,
  `longNat(uid)`, `shortNat(level)`, then every state variable value. It does
  not serialize variable-count headers, nil markers, predecessors, or action
  names; those are runtime/trace-reconstruction metadata.
- `TLCStateMutExt.copy` preserves level, predecessor, and action but resets
  worker id and uid; `deepCopy` preserves worker id and uid. Neither copy path
  carries deferred executor callables, matching Java's `TLCStateMutExt`.
- The Go `TLCTrace` still keeps an in-memory mirror of `TraceRecord` values as a
  transitional convenience for direct state-object trace access. This mirror is
  not the source-of-truth file format and should shrink as reconstruction moves
  fully to Java-style fingerprint replay.
- Worker-local concurrent traces keep Java's separate record format:
  `longNat(predecessorPointer)`, `shortNat(predecessorWorker)`, fingerprint.
- `ConcurrentTLCTrace` checkpointing deliberately skips the shared `TLCTrace`
  checkpoint path. It checkpoints each worker trace fragment and creates the
  shared `MC.st.chkpt` marker file for Toolbox/script compatibility, matching
  Java's override.
- `ConcurrentTLCTrace.elements()` merges per-worker trace enumerators. Java's
  `nextPos` returns sentinel `42` for "has another fingerprint" and `-1` for
  exhaustion; Go mirrors that odd API so trace-counting and management code can
  stay mechanically aligned.
- State numbering and action labels in printed traces are test-observed.

## Simulation Architecture

`Simulator` performs random behavior generation instead of exhaustive search.

Constructor responsibilities:

- Create a `FastTool` in `Simulation` mode.
- Determine deadlock/liveness settings.
- Compute trace depth and trace count.
- Create one `SimulationWorker` per configured worker.
- Use independent RNG seeds per worker.
- Use per-worker liveness checker when liveness is enabled.
- Eagerly create `TLCGet("config")` record value.
- Initialize coverage if requested.

Simulation flow:

1. Check assumptions.
2. Compute all initial states once.
3. Validate init states and invariants.
4. Filter by model constraints.
5. Normalize initial states.
6. Start progress report thread.
7. Start workers.
8. Workers randomly choose initial and next states up to depth/count.
9. Main thread collects worker results.
10. Stop on fatal error or report continuable violations depending on mode.

Important behavior:

- Simulation is not just model checking with random queue order. It has separate
  liveness checking, trace file behavior, random element/subset hooks, and
  worker result aggregation.
- Seed and aril are printed and must be reproducible. `SimulationWorker` uses
  TLC's Java-compatible `RandomGenerator` semantics, including `nextDouble`
  based element selection, worker seeds from `nextLong`, and `nextPrime` as the
  alternative-action stride.
- Simulation traces compress finite stuttering steps in `getTrace` and then
  repair predecessor links, while `getUncompressedTrace` preserves the raw
  predecessor chain.
- After worker completion, Java evaluates postconditions even on simulation
  errors. If the worker result has a trace, it passes
  `SimulationWorkerError.getCounterExample()` into
  `Tool.checkPostConditionWithCounterExample`; otherwise it calls the ordinary
  postcondition check. The returned error code replaces the worker error only
  when its mapped exit status is more severe. Initial-state invariant
  violations are different: Java evaluates postconditions with a one-state
  `CounterExample` for output side effects and still returns the invariant
  violation code.
- Java has `SimulationWorker`, `ExplorationWorker`, `RLSimulationWorker`, and
  `RLActionSimulationWorker`. The Go port keeps one concrete
  `SimulationWorker` struct. The `debug bool` selects Java
  `ExplorationWorker` behavior: generate all successors through the next-state
  functor, randomly select one successor, execute deferred callables, and
  perform liveness/post-trace checks. RL modes remain enum values because they
  select different scheduling and Q-table behavior, while still using the same
  worker struct. Java gives debugger/exploration worker construction precedence
  over RL worker construction; the Go constructor therefore keeps the configured
  scheduler on `Simulator` but forces the effective worker mode to standard when
  `debug bool` is set.
- Java can select RL simulation with `tlc2.tool.Simulator.rl` or
  `tlc2.tool.Simulator.rlaction` system properties and tune RL with
  `.rl.alpha`, `.rl.gamma`, `.rl.reward`, and `.rl.enabledOnly`. Go accepts the
  same property names through the Java-style `-Dname=value` parser and keeps
  the `TLAGO_*` environment aliases as non-Java conveniences.
- Java reports `TLCGet("config").mode` as `"generate"` when
  `tlc2.tool.impl.Tool.probabilistic` is true and `"simulate"` otherwise.
  This is separate from the scheduler name (`random`, `rl`, or `rlaction`).
- Java's simulation statistics size action-transition matrices and report
  `TLCGet("actions")` from `tool.getSpecActions()`, i.e. initial-state actions
  followed by next-state actions. Go must use `Tool.GetSpecActions()` here too;
  using only next actions shifts ids and drops init-action entries.
- Java writes `Root_actions.dot` when trace-action output is `BASIC` or `FULL`.
  `BASIC` reduces action instances by source definition; `FULL` keeps each
  instantiated action and clusters vertices by context string. Both modes
  aggregate per-worker matrices, draw weighted seen edges with the Java log-log
  formula rounded to two decimals, and draw dotted unseen edges only when the
  sink is not an init predicate.
- Java `SimulationWorker.simulateRandomTrace` calls
  `IdThread.setCurrentState` after selecting the initial state and again after
  each selected successor. Go mirrors that with concrete `SetCurrentState`
  replacement semantics; scoped `PushCurrentState` remains for nested evaluator
  calls such as next-state generation.
- Java `SimulationWorker.simulateAndReport` catches generic `Exception`,
  publishes a worker error with code `0`, and stops that worker. Go mirrors this
  with a recover boundary because many ported Java `Assert.fail` paths are
  represented as panics carrying `*TLCError`.
- In simulation mode, `TLCExt!Trace` uses Java's
  `RecordValue(TLCState, Action)` constructor: each tuple element is a record
  whose first field is `_action`, followed by the state's variables. The
  non-simulation trace constructors remain state-record only.
- Java `RecordValue(Action)` emits `name`, structured `location`, and, for
  parameterized action instances, `context` and `parameters` in parameter
  order. Go mirrors the parameter fields on the concrete `Action` record helper
  and keeps the current string location until the SANY semantic nodes carry the
  Java `Location` coordinates.
- Preserve the Java `INextStateFunctor` contract: plain `addElement(TLCState)`
  is unsupported for simulation workers; only `setElement` and the
  action-tagged successor `addElement(s, a, t)` paths are valid.
- `ActionItemListExt.cons(Action, kind)` has a subtle coverage lookup:
  Java calls `act.cm.get(this.pred)`, not `act.cm.get(act.pred)`. Keep the Go
  `ConsAction` lookup on the receiver list's current predicate so coverage
  counters line up with Java's decomposition.
- Successor-processing errors from model/action constraints, trace writes, or
  state-writer paths bubble out to the caller's `doNextFailed` path. Only
  invariant and implied-action evaluation errors use Java's dedicated
  `doNextEvalFailed` path with the property/action name parameter.
- Initial-state processing writes unseen states to the state writer and worker
  trace, enqueues them, and only then calls `liveCheck.addInitState` with
  `tool.noDebug()`, matching Java `DoInitFunctor`.
- Simulator result consumption follows Java's continuation policy. Worker
  exceptions and liveness exceptions stop the run, and
  `TLC_INVARIANT_EVALUATION_FAILED`,
  `TLC_ACTION_PROPERTY_EVALUATION_FAILED`, and
  `TLC_STATE_NOT_COMPLETELY_SPECIFIED_NEXT` are non-continuable regardless of
  `-continue`. Other behavior errors keep workers running only when
  `Globals.Continuation` is set.
- After the result loop exits, Java interrupts every simulation worker and
  joins each for up to ten seconds. The Go port mirrors this with a stop flag
  and concrete worker `Join` channel so simulation returns without leaving
  worker goroutines running.

Port guidance:

- Port model checking first, then simulation.
- Preserve `RandomGenerator` behavior before porting simulation tests.

## Liveness Architecture

The liveness subsystem implements temporal property checking by translating
properties into liveness expressions, orders of solution, optional tableau
graphs, behavior graphs, and SCC checks.

### Formula Translation

`Liveness.astToLive` converts SANY expression nodes into `LiveExprNode`:

- constant-level formulas are evaluated immediately to `LNBool`.
- variable-level formulas become `LNStateAST`.
- action-level formulas become `LNAction`.
- temporal operators become `LNAll`, `LNEven`, `LNNext`, conjunctions,
  disjunctions, negation, etc.
- bounded quantifiers may expand into conjunction/disjunction when enumerable.
- recursive temporal operators must be expanded when their level exceeds action
  level.

The translator then normalizes:

- negation pushing,
- disjunctive normal form,
- fairness decomposition,
- promises,
- state/action checks,
- possible error models.

Java decomposes startup liveness processing into two distinct phases:

- `ParseLiveness(tool)` builds `livespec /\ ~livecheck` from config-derived
  actions. It conjoins every fairness/temporal action from
  `tool.getTemporals()`. If there is one implied temporal property, it appends
  its negation directly; if there are several, it appends a disjunction of
  their negations. If both lists are empty, it returns nil.
- `astToLive` handles only the temporal constructs Java handles explicitly.
  Fallback expressions whose level is still temporal fail with
  `TLC_LIVE_CANNOT_HANDLE_FORMULA`; in particular Java deliberately rejects
  `CASE` in temporal properties instead of inheriting TLC's first-match runtime
  evaluator semantics.
- `SF_e(A)` expands to `<>[]-ENABLED <A>_e \/ []<><A>_e`, and `WF_e(A)`
  expands to `[]<>(-ENABLED <A>_e \/ <A>_e)`. Go keeps the enabled-action and
  subscripted-action pieces as ordinary concrete `LiveExprNode` state/action
  nodes with `EvalFunc`, rather than introducing Java's `LNStateEnabled` and
  subclass hierarchy.
- `<<A>>_e` (`OPCODE_aa`) becomes the same subscript-aware action node, with
  the action body and subscript preserved separately. Prime still becomes an
  action over the whole primed expression.
- Bounded temporal quantifiers are expanded only when their domains enumerate:
  `\E` becomes an `LNDisj`, `\A` becomes an `LNConj`, and empty domains collapse
  to `FALSE`/`TRUE` respectively. Enumeration or child-conversion failure falls
  back through the whole expression's level, matching Java's guarded
  try/catch around `tool.contexts`.
- Function application mirrors Java's `OPCODE_fa` handling: if the function
  expression evaluates to a non-record lambda, the application context is
  validated with `getFcnContext`, then the lambda body is translated with the
  original context just as Java does.
- `processLiveness(tool)` tags state/action predicates, pushes negation into
  positive form, simplifies, converts to DNF, classifies each DNF conjunct into
  `<>[]A`, `[]<>A`, `<>[]S`, and remaining action-free temporal formulae, bins
  equivalent temporal formulae, and creates one `OrderOfSolution` per temporal
  formula bin.

The Go port keeps that decomposition in `liveness_process.go`. `ProcessLiveness`
returns the concrete `[]*OrderOfSolution` used by `LiveCheck`; it does not hide
the work behind an interface. Fast unit tests cover the parser/normalizer shape
because this code is lightweight and easy to drift away from Java.

### `OrderOfSolution`

Each temporal formula or conjunct maps to an `OrderOfSolution`:

- optional tableau `TBGraph`.
- eventuality promises.
- `checkState`: state predicates to cache per node.
- `checkAction`: action predicates to cache per edge.
- `PossibleErrorModel` entries describing accepting cycle conditions.

### Tableau

`TBGraph` builds tableau nodes from particles:

- initial particles are closure of the temporal formula.
- successors are implied-successor particles.
- duplicate particles reuse existing nodes.
- nodes get stable integer indexes.

### Behavior Graphs

`LiveCheck` has one live checker per `OrderOfSolution`.

- non-tableau checker stores behavior graph nodes keyed by state fingerprint.
- tableau checker stores product graph nodes keyed by state fingerprint plus
  tableau node index.
- Java stores graphs on disk for scale.
- state/action check results are cached as booleans/bitvectors.
- During the Go transition, `LiveChecker` keeps the existing in-memory graph for
  the current SCC implementation and also populates the Java-style
  `DiskGraph`/`TableauDiskGraph`. Lifecycle calls (`close`, checkpoint,
  recover, reset, flush) must go to the concrete disk graph fields. Once
  `LiveWorker` is ported, disk graphs become the primary SCC input and the
  in-memory checker can be retired.
- Tableau safety-like liveness shortcuts mirror Java's `errorGraphNode`
  handoff with concrete `ErrorGraphNode` and `ErrorPrefix` fields; trace
  printing will consume those when the Java `printErrorTrace` path is ported.
- `LivenessStateWriter` embeds the concrete `StateWriter` for DOT output
  instead of introducing `ILivenessStateWriter`. `NewDotLivenessStateWriter`
  writes Java's product-graph node ids (`stateFP.tableauIndex`) and preserves
  the stuttering/dotted/default visualization hints.
- The older in-memory `BEGraph` owner keeps Java's `initNodes`, `metadir`, and
  `NodeTable` fields. Its reset and shortest-path routines are intentionally
  iterative: `ResetNumberField` uses `MemObjectStack`, and `BEGraphGetPath`
  uses `MemObjectQueue` while destructively reusing parent pointers just like
  Java.

`AbstractDiskGraph` responsibilities:

- keep two files per solution, `nodes_N` for serialized `GraphNode` successor
  records and `ptrs_N` for `(fingerprint, tableau-index, node-file-pointer)`.
- write duplicate nodes without rewriting older records; the in-memory pointer
  table determines the distinguishable graph size.
- write new records at the current file pointer, not unconditionally at EOF.
  After recovery Java seeks both graph files back to checkpoint positions and
  subsequent writes overwrite from there while stale tails are ignored.
- rebuild the pointer table from `ptrs_N` before SCC search or recovery.
- treat values below `MAX_PTR` as node-file pointers and values in
  `[MAX_PTR, MAX_LINK]` as SCC link numbers.
- support checkpoint/recover by saving and restoring the current file pointers.
- provide optional fixed-size node caching, invariant checks over all graph
  records, and DOT/string traversal helpers for debugging.
- `DebugTableauDiskGraph` is selected by Java property
  `tlc2.tool.liveness.LiveCheck.debug` only for tableau disk graphs. It is not a
  different graph algorithm: after each `addNode`, `setDone`, and `recordNode`
  it creates the normal disk-graph cache, writes `dgraph_NNN_prefix.dot`, and
  destroys the cache in a `finally` path. Snapshot write failures are diagnostic
  and print rather than changing liveness checking results.
- reconstruct counterexample prefixes with breadth-first `GetPath` searches.
  The search rebuilds the pointer table from `ptrs_N` and then destructively
  reuses element slots as predecessor links. `TableauDiskGraph` needs the same
  concrete table with a `reverse` flag that adds one packed predecessor tableau
  index per record; this mirrors Java's reverse traversable table subclass
  without adding a Go interface or duplicate table hierarchy.
- `TableauDiskGraph` must have its own recovery and traversal methods in Go.
  Java gets virtual dispatch through `AbstractDiskGraph`; Go embedding does not
  make `DiskGraph.Recover` rebuild a `TableauNodePtrTable`.

### Checking

`LiveCheck.check0`:

- decides when to run based on graph growth threshold unless final check.
- distributes live checkers to `LiveWorker` threads.
- finds accepting SCCs/cycles.
- prints liveness counterexample and lasso.
- handles worker failure precedence: liveness violation wins over checker
  failure until all live workers complete.

`LiveCheck.checkTrace` is the simulation/debug trace path. Java converts the
current trace into a temporary behavior graph by adding the first state as an
init state, then for every non-final state adding both the stuttering
self-successor and the next trace state as successors. It then adds the final
state with an empty successor set, runs a final liveness check, and resets the
graph for the next simulated behavior.

`AddAndCheckLiveCheck` is Java's testing-only subclass that synchronizes
`addInitState` and `addNextState`, then calls `check0(tool, false)` after every
addition. Go keeps this on the concrete `LiveCheck` as an `AddAndCheck` flag
and mutex instead of creating another checker type.

Tableau liveness has an important safety-like short-circuit. When
`TableauLiveChecker.addNextState` recursively finds an accepting sink tableau
node and the `OrderOfSolution` has an empty possible-error model, Java records
`errorGraphNode`, reconstructs the prefix after the surrounding graph updates
complete, prints a temporal-property counterexample, stops the main checker,
and throws the invariant-violation control exception to escape the worker
without printing a second generic error. The Go port mirrors this with
`LiveChecker.printSafetyLikeLivenessError` and the existing
`errInvariantViolated` sentinel.

`LiveWorker`:

- implements Java's iterative Tarjan SCC search over disk-backed behavior
  graphs. The DFS stack stores packed `(state, tableau-index, location,
  lowlink)` cells and uses `SCC_MARKER = -42` to revisit a node after its
  successors have been explored.
- starts from disk graph initial nodes instead of preloading every vertex;
  successors that do not satisfy the PEM's EA action are queued as fresh roots
  if they still point to disk.
- uses `TableauNodePtrTable` as the temporary SCC component set, matching Java
  even for non-tableau graphs so trace construction can share the same shape.
- dispatches to the concrete `DiskGraph` or `TableauDiskGraph` fields on
  `LiveChecker`. There is intentionally no Go `AbstractDiskGraph` interface;
  the helper methods are private concrete switches.
- ports Java's lasso construction in data form: `GetPath` reconstructs the
  prefix, `dfsPostFix` greedily finds the component segment satisfying the PEM,
  and `bfsPostFix` closes the cycle. The current Go fields store the fingerprint
  `ErrorPrefix` and `ErrorCycle`.
- reconstructs `TLCStateInfo` traces from the prefix/cycle fingerprints,
  creates a `CounterExample`, and invokes the tool post-condition hook with
  that value. The Go checker also emits the temporal-property violation,
  counterexample marker, state trace, and stuttering/back-to-state marker.
- post-hoc property attribution mirrors Java's
  `Liveness.findViolatedProperties`: evaluate each implied temporal property
  over the reconstructed lasso and print the violated property names in
  deterministic action order. `ASTToLive` converts each action predicate to
  the same `LiveExprNode` family used by the liveness checker, and the result
  is cached in the action's auxiliary map via `AttachLiveExprToAction`.
  The structural core is in place; recursive temporal-operator metadata and
  exact Java diagnostic wording still need the same breadth-first mechanical
  deepening as their Java counterparts.

`LiveCheck1` is an older in-memory implementation used by simulation and some
trace checks. It follows the Manna-Pnueli book algorithm with component
numbering ranges. The Go port keeps the supporting `BEGraph`, `BEGraphNode`,
`BTGraphNode`, and `NodeTable` concrete; Java's subclass polymorphism is mapped
onto embedded structs only where the existing code already uses it.

Port guidance:

- Treat liveness as a major milestone, not a small extension.
- First port live expression nodes and formula processing.
- Then port in-memory `LiveCheck1` for simpler tests and simulation.
- Then port disk-backed `LiveCheck` and tableau product graphs.
- Preserve warning behavior for no fairness/liveness constraints/symmetry.

## Standard Modules and Overrides

Java TLC implements standard modules in `tlc2/module`:

- `Naturals`
- `Integers`
- `Sequences`
- `FiniteSets`
- `Bags`
- `Strings`
- `TLC`
- `TLCExt`
- `TLCEval`
- `TLCGetSet`
- `Randomization`
- `Json`
- `_TLCTrace`
- `_JsonTrace`
- `_Possible`
- `TransitiveClosure`
- `AnySet`

The Go port keeps these as concrete functions in `modules_*.go`. Related Java
modules may be folded together when the behavior is still direct; for example
`Strings` and `FiniteSets` live in `modules_misc.go`, and `TLCEval`'s value
conversion helper lives beside other TLC module operators. `TLCExt` includes
the definition-by-name hook as `TLCExtTLCEvalDefinition`, which looks up a
zero-arity `OpDefNode` in the concrete `Tool` definition table and evaluates
its body in the existing context/state pair.

Standard module failures should preserve Java's specific `EC.TLC_MODULE_*`
codes and `Values.ppr` value rendering. In particular, `FiniteSets.Cardinality`
throws `TLC_MODULE_COMPUTING_CARDINALITY` for non-enumerable values, `STRING`
membership failures throw `TLC_MODULE_CHECK_MEMBER_OF`, and `STRING`/`ANY`
comparison failures throw `TLC_MODULE_COMPARE_VALUE`. These are observable
through error reporting and debugger paths, so avoid replacing them with generic
Go errors while porting module code. The numeric modules follow the same rule:
arithmetic overflow uses `TLC_MODULE_OVERFLOW`, `\div` by zero uses
`TLC_MODULE_DIVISION_BY_ZERO`, `0^0` uses `TLC_MODULE_NULL_POWER_NULL`, invalid
`%`/`^` arguments use `TLC_MODULE_ARGUMENT_ERROR`, comparison type errors use
`TLC_MODULE_ARGUMENT_ERROR_AN`, and `Nat`/`Int` membership/compare failures use
the same module membership/compare codes as Java.

`Sequences` has a few source-order details to preserve. `SubSeq` first decides
whether its first argument is a string or sequence and reports a first-argument
`TLC_MODULE_ARGUMENT_ERROR` before checking `m` and `n`; only after that does it
check natural-number arguments and domain membership. `Len`, `Head`, `Tail`,
`Cons`, `Append`, `Concat`, and `SelectSeq` should use
Java's sequence-specific `ONE_ARGUMENT_ERROR`, `APPLY_EMPTY_SEQ`,
`EVALUATING`, `ARGUMENT_ERROR`, and `ARGUMENT_NOT_IN_DOMAIN` codes. The public
Java helpers `SelectInSeq`, `Insert`, and `Remove` are not exported by the
frozen standard `Sequences.tla` interface and are not registered by the Go
standard definitions. Java's internal sequence-set object can print as
`BSeq(...)`, but `BSeq` is likewise not a standard operator name.

`Bags` distinguishes wrong-shape values from malformed bags. Operators such as
`BagCardinality`, `BagUnion`, `SqSubseteq`, `BagToSet`, and Java's odd
`SetToBag` error path use `TLC_MODULE_APPLYING_TO_WRONG_VALUE` when the value is
not the expected finite function/set shape. `BagIn`, `CopiesIn`, `BagCup`,
`BagDiff`, and `BagOfAll` use `TLC_MODULE_ARGUMENT_ERROR` or
`TLC_MODULE_ARGUMENT_ERROR_AN` for bad arguments. When `BagUnion` sees a finite
set whose element is not a bag, it reports `TLC_MODULE_BAG_UNION1` against the
whole set, not the individual element.

`TransitiveClosure.Warshall` reports a non-enumerable relation with
`TLC_MODULE_APPLYING_TO_WRONG_VALUE` and a non-pair element with
`TLC_MODULE_TRANSITIVE_CLOSURE`. The algorithm itself is Warshall over the
distinct relation endpoints in first-seen order from relation enumeration.

Core `TLC` module operators also carry observable Java validation order.
`Assert` throws `TLC_VALUE_ASSERT_FAILED` with `Values.ppr` of the second
argument. `@@` reports first/second function-shape errors with
`TLC_MODULE_ARGUMENT_ERROR`. `SortSeq` checks that its first argument converts
to a tuple, then checks that the comparator is an operator before returning for
an empty sequence; Java's first-argument message says "natural number" and the
Go port intentionally preserves that wording. `Permutations` and the finite-set
paths of `RandomElement` report `TLC_MODULE_APPLYING_TO_WRONG_VALUE`.

`Randomization` validates public arguments in Java order and reports
`TLC_MODULE_ARGUMENT_ERROR`. `RandomSetOfSubsets` checks first-argument count,
second-argument subset size, third-argument finite set, requested number of
subsets against `2^Cardinality(S)`, then subset size against `0..Cardinality(S)`.
Java also has a `RandomSubsetSet` helper that reports under the internal
operator name `RandomSubsetSetProbability`, including for probability parsing
and the requested-subsets bound. The frozen standard `Randomization.tla` module
does not export this operator, so the Go standard registry must not install it.
The helper remains useful for direct module parity. The requested-subsets bound
uses Java's
`31 - Integer.numberOfLeadingZeros(numberOfPicks) + 1 > Cardinality(S)` guard and
Java `int` left-shift overflow semantics, not a floating-point `2^n` shortcut.

`TLCGetSet` uses `TLC_MODULE_TLCGET_UNDEFINED` for missing numeric and string
registers. Invalid `TLCGet` arguments report `TLC_MODULE_ONE_ARGUMENT_ERROR`;
invalid `TLCSet` arguments report `TLC_MODULE_ARGUMENT_ERROR`. String register
names are rendered without quotes in the undefined `TLCGet(name)` message, just
as Java uses `String.valueOf(sv.val)`.

`TLCExt` one-argument operators should use Java's `TLC_MODULE_ONE_ARGUMENT_ERROR`
surface: `ToTrace` expects a `CounterExample`, `TLCModelValue` reports as
`ModelValue` and expects a string, and `TLCEvalDefinition` distinguishes
non-string names, unreachable definitions, and non-zero arity definitions through
that same error code.

`TLCExt!AssertError` is an evaluating operator with an important ordering
constraint. Java first requires the expected-error argument to be syntactically
a `StringNode`; an expression that happens to evaluate to a string is rejected.
It then evaluates the expression that is expected to fail, and only compares the
string literal with the caught exception message when that expression throws.

`Json` standard operators use `TLC_MODULE_ARGUMENT_ERROR` for the public payload
shape checks Java performs inside the module: `JsonSerialize` requires its
second argument to be a sequence or record, and `ndJsonSerialize` requires a
sequence. The separate registered `IOUtils!Serialize` path is implemented in
`modules_ioutils.go`; do not silently merge its result-record convention with
the direct `Json` module exceptions.

`TLCExt!CounterExample` is context-sensitive in Java: postcondition checking
conses the current `CounterExample` value under the actual
`CounterExample` `OpDefNode` symbol and the module operator returns that value
when present, otherwise an empty counterexample. The Go standard operator is
therefore registered as an evaluating operator rather than a plain method so it
can read the current postcondition context; do not bind by name through a
synthetic symbol.

`TLCExt!Trace` is also an evaluating operator. In simulation mode Java asks the
active `Simulator` to build the current trace. In model-checking mode Java uses
the trace file for committed states, but has a special transient-state path when
the current state still has `TLCState.INIT_UID`: it reads
`IdThread.currentState`, reconstructs the committed prefix for that predecessor,
then appends the predecessor and transient state. The Go port mirrors this with
the goroutine-local current-state scope and the concrete `ModelChecker`
trace-reconstruction helpers.

`TLCExt!PickSuccessor` is synchronized and deliberately interactive. Before
prompting, Java accepts already-seen BFS successor fingerprints because TLC
checks action constraints before it filters old states. When the guard is
`FALSE` and the successor state is complete, Java identifies the action from the
extended successor state or regenerates next states to find the first matching
action, then reads `stdin` commands: yes/blank accepts, `n` rejects, `s` prints
both states, `d` prints the state diff, and `e` marks the successor explored in
BFS mode. All prompt/status text goes through `MP.printMessage` with
`TLC_MODULE_OVERRIDE_STDOUT`, so the Go port must use `PrintMessage` for this
interactive output rather than writing directly to stdout.

Preserve Java override annotations when registering standard operators.
`TLC!TLCEval` is an evaluating override because it receives the unevaluated
expression and caches converted constant-level results on the semantic node.
Its Java implementation branches directly on the argument semantic level:
state/action/temporal expressions evaluate in the incoming context and are not
cached; constant-level expressions with a non-empty context also evaluate
without caching; only constant-level expressions under an empty context use the
static read/write-lock protected semantic-node cache. Cache reads must mux
`WorkerValue` through the active worker id before returning it. Cache writes
must mirror Java's `WorkerValue.demux`: evaluate under
`EmptyContext`/`EmptyState`, deep-normalize, and, when a mutable value is shared
across multiple workers, reevaluate it once per worker with the same
random-enumerable seed before muxing the active worker's copy and converting it
through the legacy `toSetEnum`/`toFcnRcd` path.
Ordinary symbol lookup has the same rule: a cached `WorkerValue` is muxed by the
current worker id when running inside a worker goroutine, and only falls back to
the state's worker id outside that scope. This mirrors Java's `IdThread`
selection and avoids letting copied states override the executing worker's
per-worker constant value.
`TLC!TLCGet`, `TLCExt!CounterExample`, `TLCExt!Trace`, `_TLCTrace!_TLCState`,
`_JsonTrace!_TLCState`, and `_Possible!_Counts` all carry non-constant
minimum levels in Java to prevent invalid constant folding. `TLCExt!PickSuccessor`
is action-level. Conversely, `TLCExt!TLCGetOrDefault` is a plain Java operator,
so both arguments are evaluated before it chooses between the register value and
the default. `TLCExt!TLCGetAndSet` is a TLA definition, but it calls the
Java-overridden `TLCGetOrDefault`; a direct Go implementation must therefore
evaluate `defaultVal` before reading the register rather than lazily only on a
missing register.

`Randomization!RandomSubset(k, S)` must follow Java's `EnumerableValue`
subset enumerator rather than a plain shuffled sample. Java chooses a seed
index with `Random.nextInt(|S|)`, chooses an increment with
`RandomGenerator.nextPrime`, computes the same `m/a` LCG parameters used by
`EnumerableValue.computeOptimalMandA`, and emits `k` random indices from the
LCG. The emitted `SetEnumValue` is initially unnormalized, so duplicate draws
are preserved until normal set comparison/fingerprinting forces
normalization. Randomized enumeration (`elements(Ordering.RANDOMIZED)`) uses
that same `SubsetEnumerator` path with `k = size`, so it must not be
implemented as a Fisher-Yates shuffle; the random-value consumption is
observable for seeded runs and aril replay. The LCG iterator is lazy, and reset
clears its call count while retaining its current seed. Explicit-set and interval
random iterators select directly without ordinary traversal counts. Nonempty
zero-count iterators still initialize their seed and increment. Java's module
argument checks accept negative integer counts, but `getRandomSubset` then fails
allocating `ValueVec(k)` at the Java-method override boundary.

`RandomSetOfSubsets` and `RandomSubsetSetProbability` construct a
`SubsetValue` over the input enumerable and use `CoinTossingSubsetEnumerator`.
That path normalizes the base set before tossing one coin per base element,
then accumulates generated subsets in a hash set so duplicates are dropped
during generation rather than by sorting a `ValueVec` afterward.

`TLC!RandomElement` also uses `RandomEnumerableValues.get().nextDouble()` for
intervals and finite enumerated sets. The Go port must use the shared
Java-compatible random enumerable generator here, not Go's process-global
random source, so seeded simulation and trace replay consume random values in
the same places Java does.

`RandomEnumerableValues` is mode-sensitive in Java. During BFS model checking,
worker evaluation and trace reconstruction set the current predecessor state on
the current worker thread; the random-enumerable generator is then seeded with
`enumFractionSeed XOR predecessor.fingerPrint()`, and repeated random choices
within the same predecessor continue the same stream. Initial-state generation,
DFID, and simulation use the default per-thread seed stream without predecessor
reseeding. The Go port mirrors Java `ThreadLocal<Random>` with per-goroutine
state; do not collapse it into one process-global RNG.

`Json!ToJsonObject` mirrors Java's `getObjectNode` dispatch. Records and
tuples become JSON objects, but a function record whose domain is a valid
sequence `1..n` is routed back to the array writer even under `ToJsonObject`.
This oddity is source-compatible with Java and must be preserved for
round-tripping existing specs.

Java's `Sequences` class contains helper methods such as `SelectInSeq` and
`Insert`, and its internal sequence-set object can print as `BSeq(...)`, but
the frozen `Sequences.tla` module exports only the standard sequence operators.
Do not register those helper names as Go standard definitions unless upstream
adds them to the standard module interface.

`TLCGet("diameter")` is mode-sensitive. In model checking it reports checker
progress. In simulation Java reads the current `SimulationWorker` trace count,
returning zero while initial states are generated outside a worker. The Go port
stamps simulation states with the concrete worker id and uses that id to read
the worker-local trace count, which also keeps worker-local `TLCGet`/`TLCSet`
registers aligned with Java.

`TLCGet("generated")`, `"distinct"`, and `"queue"` are direct main-checker
queries in Java. Simulation exposes generated trace/state counters through
`TLCGet("stats")`, not those direct string keys. Checker `"stats"` has fields
`queue`, `distinct`, `initial`, `generated`, `diameter`, `duration`, and
`worker`; checker `"config"` has fields `mode="bfs"`, `deadlock`, `worker`,
`seed`, `fingerprint`, and `install`. Simulator `"stats"` has fields `traces`,
`duration`, `generated`, `behavior`, `worker`, `distinct`, `distinctvalues`,
`retries`, `actions`, `levelmean`, and `levelvariance`; simulator `"config"`
has fields `mode="simulate"`, `depth`, `traces`, `deadlock`, `seed`, `aril`,
`worker`, `install`, and `sched`. Java eagerly caches checker and simulator
config records; simulator `aril` in that record is the construction-time field,
not a recomputation from the RNG after worker seeds have been drawn.
Java distinguishes direct `TLCGet` counters from stats-record counters:
`TLCGet("generated")`, `"distinct"`, `"queue"`, and `"duration"` use
`Math.toIntExact` and report overflow, while stats and coverage records use
`IntValue.narrowToIntValue`, which returns `-1` when a `long` does not fit in
TLC's 32-bit integer value.
Java enables expensive simulator extended statistics through
`tlc2.tool.Simulator.extendedStatistics` and switches exact counters on with
`.extendedStatistics.naive`. The Go port supports those property names and
`TLAGO_SIMULATOR_EXTENDED_STATISTICS(_NAIVE)` aliases, using HyperLogLog bits 8
for distinct states and 10 for distinct variable values as Java does.

Integer `TLCGet(i)`/`TLCSet(i, v)` and named registers (`"s:..."`) are
current-worker local in Java when evaluated on an `IdThread`; they do not use
the predecessor state's stored trace worker id. Outside a worker, Java
broadcasts `TLCSet` through the checker/simulator; simulator `TLCGet` fallback
reads worker 0 exactly, even when worker 0 has no value. Store these registers
on the concrete Go `Worker` and `SimulationWorker` structs, not centrally on
the checker/simulator, because Java's `AbstractChecker` and `Simulator` read
them back from worker objects. `TLCGet("all")` and `TLCGet("all:named")`
return functions whose values are per-worker tuples and, like Java, iterate
indices or named keys visible on worker 0. The Go port's ambient current-worker
helper must be goroutine-scoped, mirroring Java's thread-local worker identity.
A single process-wide worker slot is not
correct once workers run concurrently. Java `IdThread` also stores the
predecessor state while next states, alias state records, and reconstruction
states are evaluated; the Go port mirrors this with a goroutine-local
current-state stack and resets it on checker error paths. `TLCSet("exit",
TRUE)` stops the active checker and simulator. `TLCSet("pause", TRUE)` is a
blocking BFS model-checker control: Java prints
`Press enter to resume model checking.` and waits on
standard input before returning `TRUE`.

Incomplete next-state errors must carry Java's parameter vector: for a
single-action spec, the plurality fragment and comma-joined unassigned variable
names; for multi-action specs, the action name followed by those two fields.
Simulation, BFS, and DFID use this exact shape for next-state failures.

Random subsets of product-shaped values (`[S -> T]`, record sets, tuple
products) must use Java's product-index strategy from
`SetOfFcnsOrRcdsValue`. TLC converts each constituent to a `SetEnumValue`,
draws random indices over the mixed-radix product, and reconstructs the
corresponding function, record, or tuple. When the product cardinality exceeds
32-bit size, Java switches to a `BigInteger` path using `Long.MAX_VALUE - 24`
as the stride and `RandomEnumerableValues.nextLong()` as the offset.
For `SUBSET S`, Java's `SubsetValue.elements(k)` draws random bit-mask indices
when `|S| < 31` and `k <= 2^16`; otherwise it uses coin tossing over the
normalized base elements. This direct `elements(k)` API differs from both
`RandomSubset(k, SUBSET S)` and powerset randomized ordering, which first
materialize the powerset to an explicit set. `KSubsetValue` randomized ordering
uses an endless Algorithm S generator with a no-op reset and the default model.
Random-subset results inherit the receiver's cost model; product subsets also
count their result elements. Random tuple and record elements inherit the model,
while random function elements deliberately use the default model. Product
iterators initialize their random seed or offset before converting constituents
and cache their mixed-radix weights, preserving Java's ordering and laziness.

Override infrastructure:

- `TLARegistry` maps TLA+ names to Java names.
- `TLAClass` loads module classes.
- `TLCBuiltInOverrides`, `ITLCOverrides`, and annotations describe overrides.
- `MethodValue` and `EvaluatingValue` bridge operator calls to module methods.

Go mapping:

- Java standard-module static initializers are represented by
  `Tool.InstallStandardDefinitions`.
- `NewTool` installs concrete values such as `Nat`, `Int`, `STRING`, and
  `Any`, plus `MethodValue` entries for pure module operators.
- Java's `TLARegistry` aliases are installed beside the method names, e.g.
  `Plus` and `+`, `Concat` and `\o`, `MakeFcn` and `:>`.
- Context-sensitive operators such as `TLCGet` and `TLCSet` use
  `EvaluatingValue` so their implementations receive the concrete `Tool`,
  context, current state, successor state, eval control, and cost model.
- JSON, TLCExt, `_TLCTrace`, and `_Possible` overrides follow the same table.
  `TLCExt!TLCCache` uses a concrete `TLCExtCache` for constant-level
  expressions and the concrete `TLCStateMut` cache for state-level
  expressions, matching Java's split between expression tool objects and
  `TLCStateMutExt`. Java only evaluates the closure in those two cacheable
  cases; action- and temporal-level expressions bypass the cache and evaluate
  the expression directly. The state-level closure key is evaluated against the
  current state only, mirroring `tool.eval(closure, c, s0)`.
- `_TLCTrace!_TLCState` and `_JsonTrace!_TLCState` are evaluation overrides in
  Java. They ignore the syntactic level argument and return a record
  representation of the current state directly, avoiding reconstruction through
  `TLCExt!Trace`. The Go port registers `_TLCState` as an `EvaluatingValue`
  with both module-qualified aliases.
- CommunityModules `IOUtils` is also part of the practical TLC runtime surface.
  Java implements it in `tlc2.overrides.IOUtils`, not in the core
  `tlc2.module` package. The Go port keeps it in the same central `tlc`
  package with concrete functions for `IOSerialize`, `IODeserialize`,
  text/NDJSON `Serialize`, text `Deserialize`, environment lookup, process
  execution, template execution, and `atoi`. Explicit IOUtils compression uses
  gzip-wrapped Java value streams. Generic value-stream file paths that
  correspond to Java's `ValueOutputStream(File/String)` and
  `ValueInputStream(File/String)` honor the global gzip flag; raw
  `DataOutputStream`/random-access graph files remain uncompressed like Java.
- CommunityModules `Combinatorics` contributes native `factorial` and `choose`
  operator overrides in `tlc2.overrides.Combinatorics`. Those are distinct from
  the lower-level `tlc2.util.Combinatorics` table/mixed-radix helpers. The Go
  port keeps the utility helpers in `combinatorics.go`, registers the two
  module operators as native standard definitions, and has the root bridge skip
  the recursive TLA definitions from `Combinatorics.tla` when the frozen
  CommunityModules library is loaded.
- CommunityModules `Bitwise` contributes native overrides for LOCAL recursive
  helper definitions `And`, `Or`, and `Xor`, plus exported `Not` and `shiftR`.
  The exported infix operators `&`, `|`, and `^^` remain TLA definitions that
  call the native helpers. The bridge therefore skips all five annotated native
  definitions when converting `Bitwise.tla`, but named-instance export bindings
  include only `&`, `|`, `^^`, `Not`, and `shiftR`; exporting `B!And` would
  violate the module's LOCAL boundary.
- CommunityModules fold/function/set/bag helpers are practical native override
  surface too. Java overrides `DyadicRationals!Reduce` even though it is LOCAL,
  so the Go bridge skips the TLA definition but does not export `D!Reduce` for
  named instances. Java `Functions.IsInjective` uses a non-mutating O(n^2)
  duplicate check for already-user-visible tuples and interval functions, but
  sorts freshly converted tuple values in place. `Functions.AntiFunction`
  materializes a function record and normalizes the inverse record. The fold
  overrides intentionally differ in accumulator order: `Functions` and
  `FiniteSetsExt` call `op(value, acc)`, while `BagsExt.FoldBag` follows Java's
  `op(acc, bagElement)` loop for each multiplicity.
- CommunityModules `CSV` and `GraphViz` are side-effect/string rendering
  overrides. CSV appends UTF-8 text lines with Java `String.format`-style
  `%1$s` placeholders, normalizes records before header/value emission, uses
  raw `StringValue` contents for paths and delimiters, treats the delimiter in
  `CSVRead` as a Java regex, returns an empty tuple/zero count for missing
  files, and overflows beyond signed 32-bit record counts. `GraphViz.DotDiGraph`
  renders a fixed `digraph MyGraph {...}` string; node ids are Java signed
  `long` fingerprints, not unsigned Go `uint64` decimal strings.
- CommunityModules `Graphs` and `UndirectedGraphs` share Java
  `AbstractGraphs`: validate graph records, normalize node/edge sets, build
  adjacency restricted to endpoints in `G.node`, skip malformed edge values,
  enumerate simple paths with depth-first backtracking, and answer connectivity
  with breadth-first reachability. Directed edges are ordered 2-tuples;
  undirected edges are sets of one or two nodes, with singleton sets treated as
  self-loops. `UndirectedGraphs.ConnectedComponents` uses union-find. Because
  `SimplePath` and `AreConnectedIn` are exported by both modules but differ in
  semantics, the Go port registers module-qualified native values such as
  `Graphs!SimplePath` and lets the root bridge alias skipped SANY definitions
  and named instances to the module-qualified value.
- CommunityModules `VectorClocks.CausalOrder` is a ShiViz-style topological
  sort over vector-clock log entries. Java first groups entries by the node
  returned from `node(entry)`, stable-sorts each node's log by its own clock
  value, constructs parent links whenever another host's clock component
  advances beyond the per-host global clock, then repeatedly emits root entries
  with no remaining parents. The Go port keeps this phase decomposition and the
  same parent-removal behavior, but stores host logs and parent/child sets in
  explicit slices so iteration remains deterministic.
- CommunityModules `SVG` has three self-contained native overrides and one
  third-party graph-layout override. Go ports `SVGElemToString`,
  `NodeOfRingNetwork`, and `PointOnLine` directly: attribute underscores become
  dashes, attribute values are single-quoted, child elements are recursively
  serialized, `<<`/`>>` in inner text are escaped, ring coordinates use Java's
  polar conversion and integer truncation, and `PointOnLine` uses the Java
  floating division/truncation formula. `NodesOfDirectedMultiGraph` remains on
  the TLA fallback until the JGraphT/JUNG layout algorithms can be mirrored
  faithfully.
- CommunityModules `SequencesExt` mixes ordinary Java
  `@TLAPlusOperator` overrides with two string-only `@Evaluation` shortcuts.
  Go ports the native set/sequence conversion, longest-common-prefix, fold,
  search, remove, suffix, and all-subsequence helpers directly. Java returns
  `null` from `ReplaceFirstSubSeq`/`ReplaceAllSubSeqs` for non-string values so
  the pure TLA definition handles tuples; until the Go evaluator grows that
  exact `@Evaluation` fallback path, the Go helper must preserve equivalent
  tuple behavior itself. Keep the Java quirks: `SelectInSubSeq` and
  `SelectLastInSubSeq` return indices in the original sequence range, and
  `SelectInSeq` reports the non-boolean predicate position as `"third"`.
- CommunityModules `Statistics.ChiSquare` delegates to Apache Commons Math
  `ChiSquareTest`. Java normalizes/converts both inputs to function records,
  ignores the function domains after normalization, converts values to expected
  `double[]` and observed `long[]`, parses alpha from a `StringValue`, and
  returns `FALSE` when Commons Math rejects the null hypothesis. Go mirrors the
  Commons Math statistic, including rescaling expected counts by
  `sumObserved/sumExpected` when totals differ, and computes the chi-square
  survival probability through the regularized gamma Q function locally.
- `_POSSIBLE` is also a config-driven model transformation in Java
  `SpecProcessor`. Each configured predicate is wrapped in `_Possible!_Track`
  and installed as a model constraint for state-level predicates or an action
  constraint for action-level predicates. A matching `_Possible!_CheckName`
  postcondition reports `TLC_POSSIBLE_UNWITNESSED` after model checking if the
  named predicate was never witnessed. The Go bridge mirrors this with concrete
  `PossibleTrackNode` and `PossibleCheckNode` structs and stores counts in the
  named register `s:_possible`, using worker-local checker values so `_Counts`
  can merge them like Java's `TLCGet("all:named")`. Java's `_Possible.java`
  only declares `_Counts` as a native override; `_Track`, `_CheckName`, and
  `_PrintCounts` remain ordinary TLA definitions in `_Possible.tla`, so the Go
  bridge must install them as `OpDefNode`s for explicit standard-module use and
  named instances. For postcondition errors, Java evaluates the generated
  `_CheckName` predicate but reports the user's original predicate body;
  ordinary postcondition evaluation errors also include the predicate body as
  the second message parameter.

Port guidance:

- Implement built-ins as Go functions registered in a central table.
- Preserve override names and arities.
- User-defined Java override jars do not have a direct Go equivalent. Record
  this as an interoperability feature to design after core TLC parity.
- Do not silently ignore unsupported overrides; emit equivalent errors.

## Output and Error Architecture

`tlc2.output.EC` defines message/error codes. `MP` formats and prints messages
and broadcasts them to recorders. Tests commonly assert:

- that a code was recorded,
- that a code was not recorded,
- exact string values attached to a code,
- error trace state/action messages,
- coverage messages.

Important codes include:

- `TLC_FINISHED`
- `TLC_STATS`
- `TLC_INIT_GENERATED1`
- `TLC_INIT_GENERATED2`
- `TLC_SUCCESS`
- `TLC_COUNTER_EXAMPLE`
- `TLC_STATE_PRINT1/2/3`
- `TLC_BACK_TO_STATE`
- `TLC_DEADLOCK_REACHED`
- invariant/action violation codes.
- liveness violation codes.
- config/parser/general error codes.

Port guidance:

- Port `EC` constants early.
- Implement message recording before porting end-to-end tests. Use a concrete
  recorder/broadcaster unless multiple external recorder implementations are
  truly needed.
- Java records messages before suppression/`-nowarning` decides whether the
  console sees them. Go `Message.Suppressed` preserves that split for callers
  that need recorder parity while still knowing whether a diagnostic was
  user-visible. Java `printError` does not consult suppression, while
  `printMessage`, `printWarning`, `printTLCBug`, and state printing do.
- `MP.getMessage` notifies recorders before formatting without printing. Go
  marks these events `Message.FormattingOnly`; coded exception constructors
  preserve them too.
- `-messagesAsErrors` and the dynamic `tlc2.output.MP.warning2error` property
  abort `printWarning` through `Assert.fail`. Its `TLCRuntimeException`
  construction records an `MP.getMessage` event before throwing; it never
  reaches the ordinary warning event.
- Keep human text close to Java but assert primarily through codes and
  structured parameters like Java tests do.
- State string formatting is semantic output. Treat it as part of compatibility.

## Trace Exploration and Spec Writers

Java TLC has a small but important output-writing subsystem that creates TLA+
modules and configs from model data:

- `AbstractSpecWriter`: owns TLA and CFG buffers, appends module closing tags,
  writes streams/files, emits constants, formulas, views, aliases, and model
  value declarations.
- `SpecWriterUtilities`: creates generated identifiers, module primers,
  closing tags, formula/source content arrays, and override definitions.
- `SpecTraceExpressionWriter`: builds trace-exploration Init/Next relations,
  trace functions, trace-expression stubs, properties/invariants that reproduce
  error traces, lasso views, and config wrappers.
- `TraceExpressionExplorerSpecWriter`: creates the `TEExpression` helper module
  for user trace expressions, preserving a deterministic variable-expression
  map in comments.
- `TraceExplorationSpec`: coordinates the full generated `_TTrace`/`TETrace`
  spec from an `MCError` trace and registers the binary trace postcondition.

Go keeps this as concrete structs in the central `tlc` package:

- `SpecWriter` is the non-abstract buffer owner corresponding to Java's
  `AbstractSpecWriter`.
- `SpecTraceExpressionWriter` embeds `SpecWriter` and ports the trace-specific
  emitters.
- `TraceExpressionExplorerSpecWriter` stores the variable-expression map in
  `InsMap` so generated module order is deterministic.
- `TraceExplorationSpec` captures the naming/postcondition shell and can now
  generate the monolithic TE `.tla` file from an `MCError`.
  Its variable list comes from the global TLC state variables, matching Java's
  `TLCState.Empty.getVarsAsStrings()` call, rather than being reconstructed
  from the recorded `MCError`.
- `ErrorTraceMessageRecorder` mirrors Java's
  `ErrorTraceMessagePrinterRecorder`: it observes state-print/back-to-state
  message codes and builds an `MCError` for TE generation. The `TLC` runner
  subscribes it only when `Options.GenerateTraceSpec` is set.
- `-generateSpecTE` also installs Java's implicit binary-trace postcondition
  before the tool is built when binary trace generation is enabled. Go stores
  the derived TE module name in `Options.TraceSpecModuleName` so the
  postcondition path and final generated `.tla` module use the same name.

Correctness notes:

- Generated identifiers intentionally mirror Java's `scheme + currentMillis +
  counter*1000` shape.
- Writer output order is semantic. Do not use unordered Go maps in emitters.
- The generated timestamp text is not a semantic input to TLC tests; compare
  structural output or normalize that line when doing byte-level writer tests.
- Trace expressions are represented as variables and definitions, with
  temporal-level expressions initialized to `"--"` and primed according to
  Java's level rules.

## Coverage Architecture

Coverage is represented by:

- `CostModel`
- `CostModelNode`
- action wrappers and op-application wrappers.
- coverage hash tables.
- reporting through `CostModelCreator` and output messages.

`CostModelCreator` does not merely walk the syntax tree and add children. Java
reconstructs one call tree per action because the SANY semantic graph shares
operator definitions globally, while coverage must be reported per action. Its
per-action state is load-bearing:

- a stack of current `CostModelNode`s, rooted at the current `ActionWrapper`;
- a substitution map from substituted expression nodes to `Subst` identity;
- a map from higher-order operator body nodes to wrappers that should later
  receive that body as a child;
- an active `OpDefNode` set used with `CoverageHashTable` to stop recursive
  operator expansion only when a recursive definition is already on the path;
- a LET-IN map from each LET body to the IN body wrapper that should also see
  the LET part through `OpApplNodeWrapper.addLets`;
- a global context approximation used only when an operator application has
  operator arguments, so `Op(s)` can later be connected to the passed
  operator/LAMBDA body.

The Go port keeps the same side-table shape in `coverageCreator`. It uses
`InsMap` for child order and ordinary Go maps only for key lookup where no
iteration order is observable. Because Go's semantic nodes store an operator
symbol rather than Java's `SymbolNode` subclass hierarchy, the coverage creator
resolves `OpDefNode`s through `Tool.Lookup` and `Context.Lookup`; the parser
front-end must preserve recursive flags and operator-argument nodes for full
Java-equivalent coverage trees.

Coverage counts:

- expression/action hit counts,
- secondary action counts for newly discovered successor states,
- zero/non-zero coverage,
- cost coverage.

Java has two knobs. `TLCGlobals.isCoverageEnabled()` is driven by the
`-coverage` interval and enables full expression/action coverage.
`TLCGlobals.Coverage.coverage` is a bitmask read from the
`tlc2.TLCGlobals.coverage` system property: bit `1` enables action coverage and
bit `2` enables variable coverage. `Coverage.isEnabled` is true for either
interval coverage or any bit.

The guards are intentionally not interchangeable. `Spec.coverage` and
`Value.coverage` are plain `TLCGlobals.isCoverageEnabled()`; they guard
`cm.get(...)`/`cm.getAndIncrement(...)` in `eval`, `enabled`, `next`,
`processUnchanged`, substitution cost lookup, and `LazyValue` construction.
`ActionItemList.coverage` is `Coverage.isActionEnabled()`; it guards delayed
action-list cost traversal and action `incInvocations`/`incSecondary` counts.
Java's worker and simulation functor paths count action invocations one
successor at a time through `addElement`, but Java's direct bulk
`tool.getNextStates(action, state)` path increments `action.cm` by the generated
`StateVec` size when `Spec.coverage` is enabled. Keep those paths separate in
Go to avoid double counting simulation, which uses a bulk vector internally.
The Go port mirrors this with `CoverageInterval`, `CoverageFlags`,
`CoverageAnyEnabled`, `CoverageActionEnabled`, and `CoverageVariableEnabled`;
`TLAGO_COVERAGE` is accepted as the Go-friendly environment spelling of the
Java bitmask property.

Variable coverage is a separate count-distinct path. During cost-model
creation, Java installs a `CountDistinct.SyncedHyperLogLog(10)` on every state
variable declaration. `Worker.addElement` updates those counters only when an
unseen in-model successor is enqueued, and coverage reporting plus
`TLCGet("variables")` expose the resulting distinct-value estimate. The Go port
stores the same counter on each concrete `StateVariable` and updates it at the
same successor enqueue point in `processSuccessorForWorker`.

Port guidance:

- Do not wire coverage into the first evaluator pass unless the code shape
  makes it cheap, but reserve fields in actions/values to avoid later invasive
  edits.
- End-to-end test parity eventually requires coverage output.

## Debugger Architecture

`tlc2/debug` implements Debug Adapter Protocol support:

- stack frames for init, next, action, state, synthetic states.
- source breakpoints.
- debugger expression evaluation.
- scoped identifier discovery.
- goto-state events.
- attach/suspend/halt behavior.

Important Java classes:

- `IDebugTarget`: the central control surface used by `Tool` and stack frames.
  It defines the stepping enums `StepDirection`, `Granularity`, and `Step`, and
  a large family of `pushFrame`/`popFrame` overloads. In Go, avoid turning this
  into an interface while there is only one debugger path; keep the enum values
  and concrete debugger state on `TLCDebugger`.
- `TLCStackFrame`: the base debugger frame. It carries semantic-node identity,
  context, tool, optional exception, parent, and eventual value. Java uses it
  for stepping/breakpoint target checks and for `ResetEvalException`; Go mirrors
  that as concrete `TLCStackFrame`, `ResetEvalException`, and
  `AbortEvalException` structs. Java also keeps nested variable-reference
  caches on the frame; Go mirrors those as insertion-ordered maps and exposes
  concrete `TLCScope`/`DebugTLCVariable` slices for `Context`, `Constants`, and
  `Stack` scopes without pulling in DAP transport types.
- `TLCStateStackFrame` and `TLCActionStackFrame`: Java specializes frames for
  state and action evaluation. The state frame's `getS/getT` both return the
  captured state; the action frame's `getS` returns the predecessor and `getT`
  returns the successor. Go mirrors those as embedded concrete structs and keeps
  the pending debugger value as `"?"`. The state/action scope split is explicit
  because Go embedding is not Java virtual dispatch: action frames provide their
  own `Action` and `Trace` variable accessors rather than relying on inherited
  state-frame methods to override themselves.
- `TLCSyntheticStateStackFrame`: a manually inserted marker frame for trace
  display. It is still a state frame, but it also stores the successor used when
  evaluating expressions against a trace edge. Java creates these frames lazily
  in `TLCDebugger.stackTrace`, not when the worker halts, so BFS can still evict
  ordinary states if the front-end never asks for a call stack. Go mirrors this
  with `TLCStateStackFrame.GetTraceAsStackFrames` and `TLCDebugger.StackTrace`:
  stack traces are empty unless execution is halted, the first stack-trace
  request reconstructs synthetic frames from simulator traces, predecessor
  links, or `ModelChecker.GetTraceInfo`, and resume/step/goto/disconnect
  commands remove the synthetic frames again. Because Go stores the stack with
  the top at the end rather than Java's `LinkedList.push` head, synthetic trace
  frames are prepended internally in reverse order so the observable response
  remains Java top-to-bottom order.
- `TLCInitStatesStackFrame` and `TLCNextStatesStackFrame`: Java debugger frames
  that expose generated initial or successor states. Both keep a
  variable-reference-to-state map because DAP variable references are integers
  and partial states may not have fingerprints. Go mirrors this as concrete
  structs over `StateFunctor` and `NextStateFunctor`. Initial states sort
  lexicographically by state text; successor states cluster by action location
  before state text. Step-in chooses the successor with minimum string Hamming
  distance; step-over chooses maximum; step-out selects the predecessor or halts
  the functor. The Go frames expose `Initials`, `Successors`, and `Trace`
  scopes directly, maintain the variable-reference-to-state selection maps, and
  apply conditional breakpoint expressions against the generated states.
- `TLCDebugger`: owns breakpoints, exception-breakpoint filters, the active
  stack-frame list, stepping state, granularity, halt flags, and the connection
  to the debug adapter. The Go port keeps a concrete `TLCDebuggerFrame` union
  rather than an `IDebugTarget` interface; each pushed frame preserves its
  concrete kind while exposing the embedded base `TLCStackFrame` for common
  push/pop, source-frame, and breakpoint operations. The protocol transport is
  less important than preserving where model-checker/evaluator state is
  captured. Exception, unsatisfied-state, invariant-violation, and
  assumption-violation hooks duplicate the active frame just like Java; the Go
  stack layer accepts ordinary `error` values until the Java
  `StatefulRuntimeException` hierarchy has a full concrete Go mirror.
  Conditional breakpoint parse/semantic failures are reported back on the
  breakpoint object, but Java leaves the compiled `condition` field null; a null
  condition does not suppress a later location match. Go mirrors that by letting
  a breakpoint with no compiled condition op keep the incoming `fire` value.
  Continue/step-over/step-in/step-out/step-back/reverse-continue/goto-state
  commands update `Step`, `SourceFrame`, `Granularity`, and generated-state
  selection in the same place as Java's DAP handlers, while leaving the protocol
  transport itself for a later pass. Java initializes `step` to `In` so the
  first ordinary pushed frame can halt; the Go port mirrors that with
  `MaybeHaltExecution`, leaving `HaltExecution` for already-decided stops such
  as exception, spec, unsatisfied, and violation breakpoints.
- `TLCDebugger` breakpoint ownership is concrete state on the debugger:
  `breakpoints` keyed by module/source, boolean exception and invariant halt
  flags, and two conditional `TLCSourceBreakpoint` values for the Java "after
  Init/Next" and "unsatisfied" filters. The Go port mirrors Java's filter IDs
  and multi-worker warning as plain `TLCExceptionBreakpointFilter` values, not
  DAP protocol objects. Source breakpoint setting derives the module from the
  source name, replaces the module's breakpoint list, returns concrete
  `TLCBreakpoint` verification records, and keeps Java's parent-frame
  suppression rule: if an ancestor stack frame already matches the same source
  breakpoint, a nested frame does not fire it again. Verification is fuzzy like
  Java: the breakpoint is always stored, but the response is marked verified
  only when a same-line semantic node is found under a source range that
  `Location.includes` the breakpoint range. Java walks a `ModuleNode`; Go does
  not retain that exact SANY module object, so it walks the processed semantic
  definitions for the module while preserving Java's range-in-range inclusion
  logic, including the deliberately non-point breakpoint range with begin line
  `line + 1` and end line `line`. Breakpoints in modules that are not part of
  the debugged spec remain verified, matching Java's `moduleNode == null` case.
  A hit condition on the configured `Next` predicate is reported as unverified
  with Java's "A Next breakpoint does not support a hit condition." message.
- `DebugTLCVariable`: adapts TLC `Value` objects into debugger variables.
  Scalars expose `type` and `value`; enumerable/function/record/tuple values
  receive a non-zero `variablesReference` and lazily produce children.
- `TLCSourceBreakpoint`: stores source line/column, optional hit count, optional
  log message, a parsed condition operator, and the source `Location`. Location
  matching succeeds for `nullLoc`, otherwise it checks equal line and breakpoint
  column less than or equal to the semantic node begin column. Conditional
  breakpoints evaluate through `tool.noDebug().eval` and swallow evaluation
  failures so a broken debugger expression does not crash TLC. Named operator
  conditions are resolved directly from `SpecProcessor`. Arbitrary non-`TRUE`
  conditions go through `Tool.ParseDebuggerExpressionFunc`, installed by the
  root `tlcBridge`; it mirrors Java `TLCDebuggerExpression.process` by building
  a synthetic `__DebuggerModule__N` module that `EXTENDS` the root module,
  defining `__DebuggerExpr__N == <condition>`, parsing/checking it through the
  production SANY path, and converting the generated operator back into TLC's
  semantic node graph. The hook is source-location aware, matching Java's API
  shape. The bridge collects scoped definition parameters, LET definitions,
  quantifier/CHOOSE/function/comprehension bound variables, and operator-arity
  parameters from the Go AST around the breakpoint location. LET definitions are
  first emitted as parseable wrapper stubs and then replaced with the original
  converted LET operator bodies, mirroring Java's stub/substitution strategy.
- `TLCCapabilities` and `GotoStateEvent`: small protocol data types. They are
  useful in Go as plain structs even before a debug-adapter server exists.
  `TLCDebugger.InitializeCapabilities` mirrors Java `initialize`: goto-state is
  enabled only while a simulator exists, hover evaluation, terminate requests,
  exception-filter options, hit-conditional breakpoints, conditional
  breakpoints, step-back, and clipboard context are enabled; exception-info,
  log points, value-formatting options, stepping granularity, goto-targets,
  data/function/instruction breakpoints, and disassembly remain disabled. The
  returned exception-breakpoint filters are the same Java filter IDs plus the
  multi-worker warning when `NumWorkers() > 1`.
  `TLCStackTraceArguments`/`TLCStackTraceResponse` likewise model the Java
  request surface without importing DAP transport types; `startFrame` outside
  the active frame range returns an empty response, and positive `levels`
  restricts the returned slice exactly like Java's `StackTraceArguments`.
  The small Java request handlers with no deeper TLC semantics are also kept
  concrete: `threads` returns one thread `{0, "worker"}`, `setVariable` returns
  an empty response, `configurationDone` is a no-op, and `terminate` stops the
  active checker/simulator before running the same disconnect cleanup path.

Debugger variable details:

- Java's `Value.toTLCVariable` sets type to
  `<ValueClass>: <kind string>` and value to `toString()`.
- `StringValue` replaces quoted `toString()` output with the unquoted display
  string for debugger variables.
- `TupleValue` children are named with zero-padded 1-based indexes.
- `RecordValue` children are named by record field.
- `FcnRcdValue` children are named by domain element.
- `SetEnumValue` children are named by element string.
- Infinite or non-finite values must not be eagerly expanded.

Port guidance:

- Defer debugger protocol until core CLI TLC is functional.
- Preserve enough state/action metadata in core structs so debugger support does
  not require a second state model later.
- Keep debugger structures concrete. Do not port Java's `IDebugTarget` as a Go
  interface unless a second real implementation appears.
- Simulation debugging attaches only one worker to the debugger, but Java still
  uses `ExplorationWorker` for every simulator worker when `tool.isDebugger()`
  is true. In Go this means every worker is still the same concrete
  `SimulationWorker`; its `debug bool` selects exploration behavior without a
  second worker type.

## Distributed TLC Architecture

Distributed TLC uses Java RMI:

- `TLCServer`
- `TLCWorker`
- distributed FP set managers.
- block selectors.
- smart proxies.
- server/worker management beans.

Core Java flow:

1. `TLCServer` owns the state queue, trace, fingerprint-set manager, worker
   threads, error state, progress counters, and checkpoint lifecycle.
2. Workers register with the server. Registration wakes stuck server threads
   through `stateQueue.resumeAllStuck()` and creates one `TLCServerThread`.
3. `TLCServerThread` dequeues blocks of states, calls a remote `TLCWorker`, then
   merges the returned new states/fingerprints into trace, queue, and progress
   counters.
4. `TLCWorker.getNextStates` computes all successors for a block of predecessor
   states via `DistApp.getNextStates`.
5. Each successor fingerprint is checked against a worker-local `SimpleCache`.
   Cache hits are not sent to the FP-set manager, but the server later adds the
   skipped count through `addStatesGeneratedDelta`.
6. By default, remaining `(fp, successor, predecessor)` triples are sorted by
   signed fingerprint in a `TreeSet<Holder>`. Its comparison is fingerprint-only;
   two states with the same fingerprint collapse, retaining the first holder.
   The static `TLCWorker.unsorted` property instead uses holder allocation
   identity. Go retains deterministic insertion order for that mode.
7. Holders must pass Java's strict `last < fp` assertion, starting with
   `Long.MIN_VALUE`, before partitioning by `fpSetManager.getFPSetIndex(fp)` into
   parallel vectors of predecessors, successors, and fingerprints. This rejects
   a `Long.MIN_VALUE` fingerprint even in sorted mode; the source unsorted mode
   can also fail on its traversal order or duplicate fingerprints.
8. `fpSetManager.containsBlock` returns bit vectors whose set bits identify
   fingerprints not yet present.
9. Only those unseen states are checked with `work.checkState`,
   `work.isInModel`, and, only if the model constraint passes, `work.isInActions`.
   Passing states inherit the predecessor UID and return in `NextStateResult`.

The concrete `TLCApp` runtime captures implied-init, invariant, implied-action,
and next-action arrays at construction, retaining their identity even if the
tool later replaces its arrays. Worker-group members share one application.
Successor generation calls the vector overload separately for each captured
action. `StateVec.addElements` chooses the larger vector as receiver, so a
later action producing more states can move its successors before earlier
ones. Deadlock and complete-assignment checks follow that combined order,
before worker fingerprinting. Property checks use captured action arrays and
the current tool's name arrays; state/alias reconstruction and call-stack
replacement delegate to the application tool. Server initialization uses the
same snapshot runtime.

The root `LoadTLCApp` constructor performs config parsing, SANY and semantic
checking, then tool construction through the production bridge. Both local and
remote loading use `LoadOptions.FilenameResolver`. It retains the constructor's
raw root/config names and lexical spec directory, with the source post-SANY
module-table lookup (direct constructors retaining a `.tla` suffix fail).
The parser records logical module filenames separately from physical temporary
paths; `Tool.GetModuleFiles` resolves them with `isModule=false`. The application
creates a fresh InJar resolver for enumeration, using bundled standard/model
resources and the process classpath. Enumeration is deterministic parse order
in place of Java's Hashtable traversal. Constructor flags and FP configuration
retain their source identity; basic applications leave metadata/FP configuration
unset, while `LoadTLCAppWithMetadata` completes metadata setup after tool loading.
The full constructor marks recovery from any non-null checkpoint path, including
an empty one, but does not itself read the intern table: Java's `create` does
that before constructing the tool.

`FileUtil.makeMetaDir` now has the concrete `MakeMetaDir` port: verbatim
checkpoint return, global metadata-root override, spec-directory `states`
fallback, local-time timestamp formatting and the milliseconds property.
Exclusive creation creates parents first, then atomically creates the requested
directory; a collision returns an absolute temporary sibling. Lexical dot
segments are retained, including the NIO parent-creation fallback that can
create normalized parents yet fail the original final path. Checked creation
failures become source code 2163 without an I/O cause. The regular TLC command
uses the same directory-creation algorithm with its existing error-return API.
All three upstream FileUtilTest cases are translated after implementation.
The root `CreateTLCApp` now performs the source option loop and startup branches.
Options preserve nullable names, raw path separators, Java numeric syntax and
source diagnostics/continuation quirks, including checkpoint int overflow.
Normal startup recovers the interner before FP64 and config/tool creation.
Packaged startup uses `ModelInJar` resource presence, config copying and generated
properties through the same classpath adapter; it ignores explicit config,
skips interner recovery, forces tool mode and disables checkpoints. Its FP
configuration was captured before these late properties. Properties loading
uses Java Latin-1, logical continuations, Unicode/other escapes and partial
checked-I/O results. Late properties leave already-loaded globals intact.
The probabilistic Tool property freezes on first class initialization. Actions
retain their backing slice, so applications observe in-place mutation as Java
arrays do, while replacement leaves the captured array intact.

Value fingerprints extend each kind tag with a byte, matching Java overload
selection; lengths, numeric indices and model-value tokens extend with ints.
Actual-source comparisons cover all polynomials and the recovered initial state.
Registry/export/RPC and full distributed command/error/exit wiring remain
pending. Go library loading currently returns config failures to its caller;
Java ModelConfig's process exits need the source command lifecycle adapter.
Broader JVM class-loader/provider/default-resolver initialization is separate.

`NewTLCServerFromApp` now represents the source application constructor. It
retains the application, uses DiskStateQueue directly, opens its trace before
factory FPSet initialization, and wraps that local set with a canonical-host
non-distributed manager. Checkpoint names come from the metadata leaf substring,
while the trace and FPSet root names come from the application. Queue/trace
paths retain source separator concatenation and symlink/dot behavior. Missing
recovery directories fail at trace open; basic apps retain null-metadata failure.

`NewDistributedFPSetTLCServer` folds the subclass into the concrete server.
Its manager is created during base construction with the static expected count;
its readiness latch then takes the explicit constructor argument. Registration
and countdown are synchronized, rejected registrations do not release the latch,
and manager retrieval waits for it. The startup wait hook prints the expected
count and also runs for a zero latch. ModelCheck invokes it before init generation.
The port/report/expected-count/veto properties freeze at first server class
initialization, while explicit port assignment remains mutable. Close follows
trace, FP manager, then recursive metadata cleanup unless vetoed. Actual-source
constructor/registration/cleanup comparisons pass, including mismatched counts.
Recovery and initialization failure reporting/replay now follow modelCheck,
including recovery outside the init catch, deferred callback Exceptions, escaped
Errors, the already-done replay quirk and early-done cleanup. The upstream
DistributedDoInitFunctorEvalExceptionTest runs through the production Go parser,
application and server; its fixture lives in test_vectors/.

The normal completion path now waits for later worker registration instead of
rejecting an empty worker map or inventing a no-action exit. It uses the source
static REPORT_INTERVAL with monitor notifications, including the unconditional
first wait, zero/negative timeout behavior, checkpoint-before-done ordering and
progress baselines updated after waiting. Server synchronization is reentrant
for recorder callbacks; wait releases and restores the complete acquisition
depth. Worker joins/statistics/exits precede executor shutdown and result
capture. Unexpected exit failures escape without extra close calls; only the
source remote exception families are ignored. Final distinct-state results are
shared across instances, success uses that snapshot, and failure completion
switches the application to CallStackTool. Reported worker violations return
normally; close failures escape even after a violation. Registry publication and
removal now follow modelCheck through concrete naming/export/flush boundaries.
The primary binding is published before FP readiness and initialization, while
the worker binding appears only after successful init/recovery. Early-done init
retains the primary binding. Completion flushes after FINISHED, closes, removes
worker then primary bindings and unexports with force=false, ignoring its boolean
result. Exceptional paths stop at the source phase without extra cleanup; the
registry itself is never unexported by this method.

RunWorkerShutdownHook skips empty registrations, performs one registry guard at
the current mutable port and preserves source checked catches before direct
worker exits. It retains registrations and propagates unchecked failures.
TLCRegistryNamespace supplies local identity/binding lifetime, lazy registry
references and JDK duplicate fixed-ObjID behavior, including cached port-zero
endpoints. It has no wire listener, serialization, socket occupancy detection or
actual OS ephemeral-port assignment. Separate namespaces isolate names only;
FP64/interner class globals still need process/runtime isolation. The translated
init test uses its own naming namespace like the source per-test JVM fork.
Actual constructor autoexport, active remote calls, MP console flush and the
complete command/exit lifecycle remain pending. Full DistributedTLCTestCase transport tests follow those
features (the Java harness currently disables them for OffHeapDiskFPSet).

Distributed discovery now has source worker and FP-server lookup loops with
explicit Naming.lookup and Thread.sleep adapters. Only direct RMI connect
failures with a net-connect cause and missing bindings retry. The source signed
int backoff counter wraps, producing zero sleeps after its negative/zero phase.
The root worker startup adapter joins this discovery to production tool/group
loading. Typed checked-exception carriers preserve the distinct net/RMI families.
Actual source loop bodies match deterministic boundary comparisons; wire
registry/export/RPC and full command lifecycle remain pending. No dedicated upstream lookup tests exist; the complete upstream
distributed harness follows the missing transport features.

DistributedFPSet.main now has a concrete native RunDistributedFPServer boundary.
It preserves ToolIO/System stream separation, argument early returns, two clock
reads, raw temporary metadata path, ratio-one configuration/two nested FPSets,
factory initialization and class diagnostics, hostname timing, registration
rejection/unexport and skipped flush. Other Throwables record GENERAL before
the failure line and final flush. Registration, lookup, hostname, clock and
Object.wait can supply native provider boundaries; normal registration uses
the actual local server. MP console rendering and remote wire calls remain
separate from this recorder/stream implementation.

The report loop owns the FPSet monitor, reports signed long size and waits five
minutes. Shutdown sets the static running flag without waking it; another main
invocation does not reset that flag. Base FPSet.exit sets it before synchronized
notify, which wakes one waiter. Concrete sets carry their inherited wait state;
memory sets reuse the same reentrant monitor as their synchronized operations.
MultiFPSet init initializes children concurrently, ignores returned replacement
objects and wraps checked I/O. Worker RuntimeExceptions retain the ForkJoin
cause-copy behavior; general JVM pool scheduling/provider details remain
separate. Disk init retains lexical paths, explicit prefix mkdirs and coded
file-open failures. File.mkdirs tries the original spelling before canonical
parent creation, so missing components before .. can still fail at file open.
All sixteen upstream MultiFPSetTest methods are translated after implementation,
including identity partitioning, sign-bit collisions and the OffHeap guard.
Twenty actual-source command observations and race checks pass; parallel error
ordering/selected child is allowed to vary. Multi/Noop statesSeen uses atomic
loads/stores for snapshots while retaining the source separate read/modify/write.

A two-FP-server native integration loads the real parser/tool, registers both
sets, checks five distinct states and wakes both report waits on close under
-race. This harness loads worker classes before evaluation and registers after
init. Arbitrary concurrent native worker loading/server evaluation still races
shared FP64/interner globals; Java isolates those globals by JVM process. The
source raw init enqueue also assumes its startup phase. Process/runtime class
isolation, transport, concrete FP process exit/completion messages and full
worker/server/combined command lifecycle remain unfinished.

Worker registration is keyed by server-thread identity, not URI or worker
identity. Java can register the same worker more than once; each registration
has its own assigned block, statistics, keepalive, and removal. Go keeps one
protected `InsMap[*TLCServerThread, *DistributedWorker]`, snapshots thread keys
for iteration, and removes only the chosen thread. The URI remains display
metadata. Worker counts and block selectors use the number of registrations;
new-state counts sum queue size and each registered thread's entire assigned
block, including remote calls waiting at a worker's synchronized computation.

Registration and new-state counting preserve Java's monitor serialization.
Per-worker successor computation is serialized as well. Thread worklist
references use atomic publication for timer/progress readers, preserving null
versus empty arrays. Keepalive begins during construction before `Start`.
Graceful completion joins each thread, prints statistics, exits the worker,
then removes that registration quietly; lost-worker removal emits the
idempotent deregistration diagnostic. Sent/received counters wrap as Java ints,
and result/statistics/timestamp/delta processing remains in the inner remote/NPE
catch. Null results/first partitions take the lost-worker path; an empty
partition array takes the outer model-error path.

`DistributedWorkerRuntime` represents the static executor, keepalive timer,
ordered worker group, and completion latch of a Java worker JVM. A new Go
worker has its own runtime; `DistributedWorkerGroup` instead assembles one
shared runtime and publishes each runnable's worker before registration. The
group starts all registration goroutines, then schedules the timer and prints
readiness without waiting for registration. Standalone local registration still
starts its convenience timer after starting the server thread. Exit prints
completion, shuts down the executor,
cancels the shared timer, forcibly unexports that worker, then decrements the
latch. It does not acquire the computation lock or wait for accepted tasks.
Repeated direct exits print again and fail at unexport without decrementing;
proxy calls to an unexported endpoint fail with a direct NoSuchObjectException.
Direct `isAlive` remains true. Shutdown ignores only direct NoSuchObjectException
and does not recreate the executor or latch. AwaitTermination waits for the
latch, then sleeps ten seconds before returning.

Worker keepalive captures the registry URL from startup and performs a fresh
lookup followed by isDone after inactivity. Local groups supply a direct-server
adapter; discovery groups retain the original Naming.lookup boundary. Timeout
freezes at first task initialization, with Java int multiplication overflow.
Computing workers or recent/future invocation timestamps suppress lookup; zero
invocation forces it. The task retains its original runnable array and reads
workers dynamically, so later worker publication remains visible and a missing
worker causes the source NPE. RunKeepAliveOnce represents Java's public run.

RemoteException and NotBoundException print the source nullable error detail,
then exit workers in order. A finished server prints its completion diagnostic
before the same exit loop. Malformed URLs and NoSuchObjectException from worker
exit reach an explicit FINEST logging boundary. Unexpected failures escape;
task cancellation follows the exit loop, while each worker exit still cancels
the shared timer earlier. Timer scheduling preserves the ten-second initial
delay and sixty-second period measured from the preceding actual start;
uncaught failures terminate its goroutine. JVM wall-clock changes and complete
uncaught-thread/logging-provider integration remain separate. Twenty-seven
actual-source task/worker comparisons and native race probes match the activity,
lookup/query, property, error, cancellation and export observations. No dedicated
upstream timer tests exist; complete distributed tests await transport.

Server keepalive and final cache reads preserve their RemoteException catches;
final exit ignores only the three Java dead-worker exception families, warns,
and removes the registration in finally order. Actual registry/export/RPC
transport remains pending. Worker
fingerprint lookups and server block inserts now use
their respective shared executors; accepted tasks finish after shutdown,
while new submissions are rejected. The worker preserves the rejection's
RemoteException and the proxy's ServerException envelope.

`RMIFilenameToStreamResolver` now fetches through the `GetFile` contract into a
private temporary directory, caches paths by basename, and fetches again only
when the cached file disappears. Java ignores constructor library paths and
the `isModule` flag here; remote failures print a stack trace and still create
an empty file. Null returned bytes throw after creating the file and before
caching it. `GetFullPath` retains the source's comparison against each key's
UTF-16 length, with deterministic insertion order in place of HashMap order.
File deletion follows Java's process-exit lifetime, rather than worker exit.
The command wrapper calls `CleanupDistributedFiles`; library hosts call that
hook on process shutdown themselves.

`TLCServer.GetFile` strips the request to its basename and performs the
filename-only InJar/SimpleFilenameToStream search: packaged `/model/` assets,
user/spec directory, explicit library directories, then embedded standard-module
assets supplied by the parser bridge. The superclass's one-argument resolve
invokes the virtual two-argument method, retaining InJar's packaged-model
precedence. Failed InJar copies fall back to Simple; failed Simple copies
return the attempted file. Server reads keep
the directory/Integer.MAX_VALUE checks, single-read zero-filled buffer, and
nested RuntimeException wrapping with FileNotFoundException/IOException causes.
`LoadDistributedWorkerTool` parses configuration first and routes all module
loads through the worker resolver using the existing parser, semantic checker,
and TLC bridge. It now installs a fresh worker interning context before any
configuration or semantic values are created, with FP64 initialization from
the server polynomial before installing that source.
`StartDistributedWorkerGroup` then creates the shared local group and starts
asynchronous registration. The server exposes the TLCApp command-line deadlock
flag (default true), constant preprocess flag and current FP64 polynomial;
ModelConfig CHECK_DEADLOCK does not supply the distributed application flag.
Java Integer.getInteger decoding now governs distributed integer properties,
including signs, hex/octal prefixes, signed-int bounds, BMP digits and invalid
value fallback. FP64 random initialization uses Java Random, indexed
initialization preserves typed bounds failures, and the unrolled integer
extension has a separate source-loop implementation. The missing hex digit in
polynomial 65 is corrected; all 131 constants and 655 source-Java integer/long
extension pairs match. The upstream FP64Test is ported after implementation.
Network invocation, registry discovery/retry, endpoint export and full command
startup/error/shutdown wiring remain pending.
`FilenameToStream` now carries the local/remote resolver contract. Both concrete
resolvers return `TLAFile`, retaining existence-at-construction URI provenance,
library flags and the resolver used for `.class` override lookup. Remote cached
files keep their object identity and null original URI when constructed before
writing, even after the file starts to exist. The production parser and config
loader use `GetPath` at their OS file-reading boundaries.

`SimpleFilenameToStream` snapshots the user/CWD and explicit or TLA-Library
paths at construction, then searches standard-module and bare classpath roots.
Explicit filesystem libraries remain non-standard; both classpath locations
are standard. Module loading normalizes the .tla suffix and warns/truncates at
newlines; non-module loading keeps the filename. Missing results retain the
historical parent/child TLAFile behavior instead of returning null. File paths
retain dot components and provenance URIs preserve Unicode and escape spaces.

Go supplies ordered classpath directory/archive entries or mounted bundled
fs.FS resources. The parser bridge supplies its frozen standard assets; default
native classpath entries come from java.class.path/CLASSPATH or the CWD. Archive
resources keep their jar:file: URI while copying to one per-resolver temporary
directory. Filesystem resources retain their path. The descriptive search path
keeps Java's nested separators. Generic JVM class-loader providers, manifest
classpath expansion and module-loader VM hooks are not implemented by this
native resource adapter.

InJar behavior is folded into the same concrete resolver: raw packaged names
are tried before extension/newline handling, copied modules remain non-library,
and copy failure falls back to Simple. Simple copy errors print and return the
attempted file; opening a discovered resource and outer close errors preserve
the RuntimeException/IOException boundary. Copying uses Java's 1024-byte loop
and registers files for process-exit deletion. Native stream closes are
idempotent; resources are also closed on failure rather than leaking them.
`DistributedServerFiles` now delegates to this resolver instead of duplicating
its search/copy algorithm. Its temporary directory is retained across requests.

All eight upstream SimpleFilenameToStreamTest cases are ported after the
implementation, using existing frozen standard/community fixtures under
`test_vectors/`. The Windows-specific case retains its source platform guard;
verification here ran on Linux. Direct probes compiled the actual Java
resolver/locator/TLAFile sources and matched user/library/missing/standard/
community provenance, original Unicode/space URI spelling, path descriptions,
uppercase suffix handling, override lookup and archive predicates. The real
worker tool-loading/cache/failure/bootstrap probe also passes under -race.

`InternTable` is a concrete linear-probing table with Java String hash codes,
half-capacity growth thresholds, `2 * length + 1` growth, and signed int token
increments. `Put` grows even before a cache hit. With an `InternSource`, absent
strings retain the returned token/location and leave the local token counter
unchanged; exceptions become the cause-free `Failed to intern ...` runtime
assertion, while Java Error carriers escape. A null response increments count
without occupying a slot, preserving the source quirk.

Interning is serialized per logical JVM/table. Individual array accesses have
a separate Go lock, so `Get`, `Find`, `ToMap`, and checkpoint scans retain Java's
live slot iteration and do not wait across an in-flight remote create. `ToMap`
uses InsMap to make the resulting map iteration deterministic. Checkpoints
write the token counter followed by occupied slots in their current table order,
using Java's legacy UniqueString byte encoding. Recovery updates the counter
and inserts into the existing table, including duplicate strings/tokens and
growth, rather than replacing or deduplicating it. Incomplete records preserve
the nullable checkpoint-corruption assertion; a truncated header remains EOF.

The local worker interning source captures the master's table and copies raw
UniqueString fields to model Java serialization between JVMs. The worker then
recreates Go's cached built-in and counterexample/action/location names after
source installation. This avoids using package-initialization tokens or
sharing mutable locations with the master. Construct one worker Tool per fresh
interning context and share it among that JVM's worker threads; context reset
belongs to bootstrap before concurrent checking.

Server checkpoint ordering now includes the interner: queue/trace/FP begin,
queue resume, interner begin, queue/trace/interner commit, then FP commit. Errors
before queue resume leave the queue suspended like Java; completion messages
retain the source's recorder arguments. TLCApp restores interning before tool
construction, separately from TLCServer's trace/queue/FP recovery method.

Worker construction now retains Java's immutable raw URI metadata in the form
`rmi://hostname:port/threadId`. `DistributedWorkerAddress` supplies the address
of a future exported endpoint; local construction uses the machine hostname
and port zero, Java getPort's fallback. The raw URI is used for registration,
statistics and getURI; worker failures use toASCIIString. Host classification
follows Java's server/registry authority distinction rather than Go net/url,
including null hosts for Unicode or underscore registry names, bracketed IPv6
hosts, scope IDs, and Java IPv4/hostname rules. URI validation preserves the
component masks, escaped-pair checks, UTF-16 failure indices and the
IllegalArgumentException -> URISyntaxException message/cause boundary.

ASCII conversion normalizes to NFC, encodes non-ASCII UTF-8 bytes with uppercase
percent escapes, and preserves existing ASCII spelling/escapes. Unicode tables
come from the cached golang.org/x/text v0.28.0 dependency. Since its normalizer
inserts stream-safe CGJs into long combining sequences and Java does not, a
canonical decomposition/order/composition fallback removes that behavioral
difference while retaining genuine input CGJs. Java TLC has no dedicated URI
test; 232 fixtures captured from the installed OpenJDK 21.0.12.1 verify worker
URI construction, host/null classification, syntax failures and ASCII output.
An integration check covers the executor-rejection message and proxy envelope.
Worker-group construction now performs the local canonical-host lookup:
resolve the machine name, prefer IPv4 by default (or the configured IPv6/system
policy), then verify the reverse name against a forward lookup. Failed
canonicalization returns the numeric address, with Java's uncompressed IPv6
spelling. Successful local-host addresses have a five-second monotonic cache;
each cached address retains its canonical name. Initial lookup failure keeps
the nested UnknownHostException cause and the runnable's RuntimeException
wrapper. Go supplies native DNS/hosts resolution; JVM resolver providers,
jdk.net.hosts.file, general DNS TTL caches, scoped-address metadata and VM
security hooks are not represented. Endpoint allocation and reflection/VM
diagnostics remain with the pending startup/export transport feature.

FP-manager registrations retain Java FPSets wrapper identity, cached hostname,
and availability. Reassignment shares wrappers across partitions while keeping
the original partition count. It marks the selected wrapper unavailable,
chooses the next available successor, and retains Java's non-wrapping
`for (j=index; j<next; j++)` replacement loop even when the successor index
wrapped to the start. Exhaustion caches managerIsBroken. Scalar and block calls
catch Exception, reassign and retry; Java Errors escape. Exhausted block calls
use Java's BitVector(size,true), including its closed-range initialization
quirk. Statistics failures reassign without retrying that slot, and size/seen
totals count every partition, including shared wrappers. Close and checkpoint
coalesce adjacent wrappers and trailing copies of the first wrapper, rather
than globally deduplicating underlying FPSet objects.

Concurrent block calls submit one callable per partition and collect results
by their saved index, regardless of completion order. Submission retries retain
the three-retry bound, Java random one-to-five-second delay and shutdown check.
ExecutionException is logged and leaves a null result slot; checkFPs and
checkInvariant also use concurrent completion collection, with Java's signed
minimum and early false return. NonDistributedFPSetManager bypasses executors
and preserves its IOException fallback contracts. Empty distributed managers
now report zero servers and indexing throws ArithmeticException.

All nineteen DynamicFPSetManagerTest cases and five FPSetManagerTest cases have
been ported after the feature implementation, including FaultyFPSet's scalar
and block virtual dispatch. The nested factory tests retain the production
factory and two high-bit subpartitions with a bounded Go memory budget.

Local failure contracts retain Java's exception structure:

- The worker preserves direct WorkerExceptions and translates direct
  OutOfMemoryError/RejectedExecutionException failures to RemoteExceptions.
  Other failures retain predecessor/successor metadata in WorkerException.
- The local smart proxy supplies Java RMI's ServerException envelope for a
  worker-thrown RemoteException. This represents the observable error contract
  before actual Go networking is introduced.
- Server recoverability checks a direct EOF cause with a null detail message,
  or a direct remote cause with an OutOfMemoryError cause. It does not search
  arbitrary cause chains. Recoverable multi-state blocks print the reduction
  message, requeue the block, and halve the selection limit in Java order.
- Other remote failures and direct null-pointer failures deregister the worker;
  null-pointer diagnostics include the stack. Other exceptions reach the
  server's outer model-error catch and keep their error-state metadata.
- RemoteException's detail field supplies its cause and nested message text.
  Concrete EOF, memory, rejection, and null-pointer classes preserve null
  versus empty messages and saved Go stack frames. Timing uses differences of
  epoch millisecond readings; empty-block network-overhead division retains
  Java's floating-point result.

Important Java data structures:

- `NextStateResult`: carries `TLCStateVec[] nextStates`,
  `LongVec[] nextFingerprints`, computation time, and raw states-computed count.
  `getStatesComputedDelta()` returns `statesComputed - nextStates.length`.
- `IFPSetManager`: partitions fingerprint space, performs block
  contains/put/checkpoint operations, and reports states-seen/distinct counts.
- `IBlockSelector`: chooses how many states to hand to each worker. Java keeps
  this separate from server threads for performance tuning.
- `TLCWorkerSmartProxy`: hides dead/slow worker behavior from server threads.

Port guidance:

- Do not port RMI mechanically as networking first.
- Preserve semantics in local concrete abstractions first.
- Later choose Go RPC/gRPC only after single-process behavior is conformant.
- Port a feature's Java tests after implementing that feature in Go. Keep
  transport-dependent tests with the transport feature they exercise.
- The Go port should keep `TLCServer`, `DistributedWorker`,
  `DistributedFPSetManager`, and `NextStateResult` concrete. Transport can wrap
  these structs later.
- Java's `DynamicFPSetManager` is represented as dynamic fields on the same
  concrete Go manager. It rejects registrations beyond the expected server
  count and computes the low-bit mask with Java's loop
  `while expected > 0 { expected /= 2; log++ }`, yielding `(1<<log)-1`.
  Preserve that mask calculation even when it looks different from a direct
  power-of-two helper.
- Java's `NonDistributedFPSetManager` is also folded into the concrete Go
  manager. It sends every `LongVec` block to the single wrapped FP set, reports
  `getStatesSeen()` as the wrapped set's `size()`, performs checkpoint begin
  and commit in separate calls, recovers by replaying the local trace, and
  closes by calling both `Close` and `Exit(cleanup)` on the wrapped set.
- Distributed fingerprint checkpointing is filename-paired. Java
  `FPSetManager.Checkpoint.run` calls `beginChkpt(filename)` and
  `commitChkpt(filename)` on the same remote FP set; the later manager-level
  `commitChkpt()` is a no-op for the distributed manager. The Go concrete
  `DistributedFPSetManager` performs the begin/commit pair immediately within
  `Checkpoint(filename)`; its later `CommitCheckpoint()` is a no-op. Java calls
  checkpoint Thread.run directly, so that work remains sequential. Distributed
  checkpoint/recovery I/O failures print the cached hostname diagnostic and
  continue; local checkpoint/recovery returns errors to the caller.
- Any worker/server map that is iterated for progress output must use `InsMap`.
  Fingerprint holder de-duplication may use a Go map only if iteration is over a
  separately maintained sorted fingerprint slice.
- Distributed TLC does not support `TLCGet`/`TLCSet` in Java; preserve that
  eventual error behavior rather than silently sharing mutable variables across
  workers.

## Management Architecture

Java exposes TLC progress through JMX:

- `TLCStatisticsMXBean` defines status accessors and control actions.
- `ModelCheckerMXWrapper` wraps single-process model checking.
- `TLCServerMXWrapper` wraps distributed TLC.
- `TLCStandardMBean` provides version/revision metadata and registration.
- `StateMonitor` attaches to a Java VM and periodically prints
  `getCurrentState()`.

Statistics exposed by both wrappers:

- generated states,
- distinct generated states,
- state queue size,
- generated/distinct states per minute,
- progress/search depth,
- worker count,
- average distributed block count,
- runtime ratio for liveness,
- current state string,
- spec/model names,
- checkpoint, live-check, stop, suspend, and resume controls.

Port guidance:

- Do not port JMX. Keep the wrapper names and accessors as plain Go structs so
  CLI status, HTTP, or future RPC layers can call them.
- `ModelCheckerMXWrapper` should delegate directly to `ModelChecker` and
  `StateQueue`; `TLCServerMXWrapper` should delegate to the distributed server
  shell.
- `checkpoint()` maps to `ForceCheckpoint()`.
- `liveCheck()` maps to a concrete flag on `LiveCheck` that is consumed by the
  next `LiveCheck.Check` call.
- `suspend()` and `resume()` map to the state queue's existing synchronization
  methods.

## Checkpointing and Recovery

Checkpointing spans:

- state queue.
- fingerprint set.
- trace fragments.
- liveness disk graphs.
- checker metadata.

Pattern:

1. Main checker suspends workers.
2. Components write `.tmp` checkpoint files.
3. Components atomically commit to `.chkpt`.
4. Workers resume.
5. Recovery rebuilds in-memory structures from checkpoint or trace.

Correctness notes:

- Queue suspension races have historically had subtle bugs. Preserve the Java
  condition-variable protocol.
- Checkpoint commits must be atomic enough for crash recovery assumptions.
- Disk file formats matter if tests inspect sizes or recovery behavior.
- `vars.chkpt` follows Java `InternTable`: a single `tokenCnt` integer,
  followed by `UniqueString` records until EOF. Each record is token, variable
  location, Java string length, and the low byte of each UTF-16 code unit. Do
  not add record counts or value-stream handle encoding.

## Test Architecture

### Harnesses

`CommonTestCase`:

- installs a `TestMPRecorder`.
- provides trace assertions.
- provides liveness lasso assertions.
- provides coverage assertions.
- checks generated trace-expression specs.
- validates disk graph file sizes in some tests.

`ModelCheckerTestCase`:

- runs `tlc2.TLC` with a spec/config path and extra arguments.
- records actual exit status.
- provides setup/teardown isolation.
- supports liveness and normal model checker tests.

`SuiteTestCase`:

- common expectations for the historical `suite` models.
- asserts finished, stats, init count, no general error, and coverage.

`TTraceModelCheckerTestCase`:

- variants that check trace expressions or Toolbox trace behavior.

`TLCDebuggerTestCase`:

- runs TLC under debugger configuration and checks DAP-like events.

### Test Fixture Strategy in Go

The Java tests rely heavily on `test-model`. For Go:

- Freeze required `.tla` and `.cfg` fixtures under `tlago/tlc/test_vectors/`
  or another self-contained path before enabling translated tests.
- Do not use a directory named `testdata/` for persistent fixtures. The user
  reserves that name for ephemeral Go fuzzer storage and cleanup. Persist
  vectors in `test_vectors/` instead, including generated Java oracle data.
- Keep path conventions compatible with module names.
- Prefer generated golden recorder-event JSON only if direct translated tests
  would be too expensive; otherwise port assertions directly.

### Test-Port Ordering

Per user instruction, do not port tests until the function-level mechanical
implementation exists. When test porting starts, port in this order:

1. Utility tests: FP64, byte utils, vectors, queues/stacks, statistics.
2. Value tests: primitive values, tuple/record/set/function/subset semantics.
3. Config parser tests and model-config behavior.
4. Fingerprint set tests.
5. State queue tests.
6. Small `suite` model checker tests.
7. General model checker tests in `tlc2/tool`.
8. Simulation tests.
9. Liveness tests.
10. Coverage tests.
11. Checkpoint/recovery tests.
12. Standard module tests.
13. Debugger tests.
14. Distributed TLC tests.

### Behavioral Axes Covered by Java Tests

The Go test suite must eventually cover:

- exact FP64 output and value fingerprinting.
- value equality, comparison, normalization, and enumeration order.
- model value creation, ordering, and symmetry permutation.
- config parsing and constant/operator override handling.
- init/next action decomposition.
- bounded quantifier and CHOOSE behavior.
- incomplete state detection.
- invariant and implied-action violation reporting.
- assumption and postcondition checking.
- deadlock detection.
- model/action constraints.
- VIEW and symmetry reduction.
- alias rendering in traces.
- random element/subset behavior under seeds.
- simulation trace counts and depths.
- liveness tableau construction and accepting-cycle detection.
- liveness safety-property short-circuiting.
- fairness and stuttering behavior.
- checkpoint and recovery.
- coverage output.
- standard module operators.
- trace-expression generation.
- debugger source/stack/scoped identifier behavior.

## Correctness Hotspots

Pay extra care to these areas:

- FP64 exactness, including signed byte and unsigned shift behavior.
- `Value.compareTo` ordering across all value kinds.
- Set/record/function normalization and duplicate elimination.
- Lazy enumerable values and their enumeration order.
- Model value initialization and global ordering.
- Symmetry permutations and representative selection.
- VIEW fingerprinting.
- Context branch/cutoff behavior for `ENABLED`.
- Prime and unprime handling in action evaluation.
- `UNCHANGED` and tuple-of-vars handling.
- Action splitting and action names.
- Exact distinction among constraints, invariants, implied actions, and
  assumptions.
- Error precedence when multiple workers find errors.
- Trace write timing before invariant/implied-action checks.
- Deadlock definition: no successors generated, not merely no unseen successors.
- Liveness formula level handling and recursive temporal operator expansion.
- Tableau product graph fingerprint composition.
- Stuttering steps in liveness.
- Checkpoint queue suspension races.
- Output message codes and parameter order.

## Performance Hotspots

Java TLC performance depends on:

- mutable array-backed states.
- interned `UniqueString` variable locations.
- lazy values and streaming enumerators.
- action decomposition.
- FP64 speed.
- concurrent workers with shared FPSet and state queue.
- disk-backed queues/FPSets for large models.
- per-worker trace fragments.
- avoiding full `StateVec` allocation for huge initial states.
- avoiding repeated config record construction for `TLCGet("config")`.
- liveness action checks outside coarse graph locks.

Go performance guidance:

- Use slices for state values and vectors.
- Use integer variable locations, not map lookups, in hot state binding.
- Keep `Context` as a linked list initially; optimize only with tests.
- Use goroutines with explicit `sync.Mutex`/`sync.Cond` for queue parity.
- Use `uint64` internally for FP64 but match Java signed rendering where needed.
- Avoid `interface{}` churn and interface dispatch in inner loops where
  possible, but do not abstract before behavior is stable.
- Implement disk-backed structures before declaring performance parity.

## Initial Go API Shape

A minimal public API should support:

```go
type Options struct {
    SpecFile   string
    ConfigFile string
    MetaDir    string
    Mode       RunMode
    Workers    int
    Deadlock   bool
    Seed       int64
    Aril       int64
    TraceDepth int
    TraceNum   int64
}

type Result struct {
    ExitStatus      int
    ErrorCode       int
    StatesGenerated uint64
    DistinctStates  uint64
    QueueSize       uint64
    Messages        []Message
}

func ModelCheck(ctx context.Context, opts Options) (*Result, error)
func Simulate(ctx context.Context, opts Options) (*Result, error)
```

Command-line tools can wrap this API, but the library must be directly usable
from Go tests and downstream translators.

## Mechanical Port Checklist

Port implementation in this dependency order:

1. Output message constants and recorder API.
2. Utility collections needed by values: vectors, bit vectors, long vectors,
   object tables, random generator.
3. FP64 and fingerprint helper tests against Java constants.
4. Value interfaces and primitive values.
5. Composite and lazy values.
6. Value serialization streams.
7. `Context`, action item lists, and context enumerators.
8. `ModelConfig` parser.
9. `Spec` and `SpecProcessor` over the existing Go SANY semantic tree.
10. `Action`, `Tool`, evaluator, enabledness.
11. `TLCState`, state vectors, state printers.
12. Standard module built-ins and override registry.
13. `MemFPSet`, then `MultiFPSet` and disk FP sets if/when multiple concrete
    implementations are needed.
14. `MemStateQueue`, then disk queues if/when multiple concrete
    implementations are needed.
15. `ConcurrentTLCTrace`.
16. `AbstractChecker`, `ModelChecker`, `Worker`.
17. `Simulator` and simulation workers.
18. Liveness expression nodes and formula processing.
19. `LiveCheck1`.
20. Disk-backed liveness graph and `LiveCheck`.
21. Coverage.
22. Checkpoint/recovery.
23. Debugger support.
24. Distributed support.

Only after this implementation pass should the Java `tlc2` test suite be
translated into Go tests, as requested.

## Compatibility Definition

The Go TLC port is ready when:

- It parses specs using the production Go SANY parser.
- For TLC-supported TLA+, it produces the same reachable-state counts as Java
  TLC for the frozen test corpus.
- It reports the same primary error code for every Java test-model case.
- It prints equivalent counterexample traces, including state order, action
  labels, stuttering/back-to-state markers, alias output, and lasso shape.
- It matches Java value semantics, including comparison, normalization,
  membership, enumeration order, fingerprinting, and serialization.
- It matches Java liveness decisions for safety-like and true liveness
  properties.
- It can checkpoint and recover models that Java tests cover.
- It performs within the same rough complexity class as Java TLC on large
  models, with disk-backed queues and fingerprint sets available.
