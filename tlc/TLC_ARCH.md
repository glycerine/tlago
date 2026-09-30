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
  them: model constraints, action constraints, then postconditions. Runtime
  invariant actions are appended after normal config processing, so the missing
  `INIT`/`NEXT` checks only see the static model config just as Java's
  `SpecProcessor` does.
- Java's `-dump class,...` loads an `IStateWriter` by reflection. Go records the
  requested class name in `RuntimeParameters.CustomStateWriterClass` and uses a
  no-op writer until a Go extension hook exists; do not add a one-method writer
  interface only for this placeholder.
- Java's `DotStateWriter` prints full successor labels unless
  `TLCGlobals.printDiffsOnly` is set. Stuttering is an explicit
  `Visualization.STUTTERING` hint; an ordinary self-loop is still an ordinary
  edge. Keep those decisions on the concrete `StateWriter` rather than adding
  writer interfaces.
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
- evaluate `ALIAS` config operators by converting complete record values back
  into TLC states, falling back to the original state if evaluation or
  conversion fails.
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
- `setSource` associates semantic nodes with values for better fingerprint
  exception diagnostics.
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
- In `evalApplImpl`, if a looked-up `LazyValue` is forced with `s1 == null`,
  Java evaluates the lazy expression directly with the lazy value's saved
  context and cost model. The cached path is reserved for the `s1 != null`
  branch through `LazyValue.getValue`.
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

Important concrete values:

- `BoolValue`: singleton true/false.
- `IntValue`: cached small-ish integer values via factory.
- `StringValue`: backed by interned `UniqueString`.
- `TupleValue`: function from `1..n` to values.
- `RecordValue`: sorted/normalized field names and values.
- `SetEnumValue`: explicit finite set with normalization and duplicate
  handling.
- `IntervalValue`: finite integer interval without materializing all elements.
- `FcnRcdValue`: explicit finite function, optimized for interval domains.
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
- `SetPredValue`: predicate-filtered set.
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
  the global `ModelValue.mvs` table. The Go port mirrors this with
  `ModelValueAtIndex`.
- Compound lengths follow Java's encodings: tuples and function records use
  compact naturals, set enumerations and records use signed lengths to preserve
  normalizedness, and function records write an interval-domain marker byte
  before either interval bounds plus values or explicit domain/value pairs.

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
or `TLAGO_DISK_FPSET_LOG_LOCK_CNT` overriding the exponent. The current Go
disk FP implementation still protects the table with the coarse `DiskFPSet.mu`;
fine-grained per-stripe `put`/`contains`/flush locking should be ported as a
separate disk-FP concurrency chunk.
Disk fingerprint reads mirror Java `IdThread.GetId(braf.length)`: a goroutine
with a current worker id uses its corresponding fixed `BufferedRandomAccessFile`
reader; calls outside worker scope fall back to the reader pool.

Trace level is trace-authoritative when extending partial state spaces:
`CheckImpl.makeStateSpace` uses `TLCTrace.getLevel(state.uid) + depth` in Java,
so Go's `CheckImpl.MakeStateSpace` uses `TLCTrace.GetLevelForState(state)`
rather than trusting the mutable state's cached level when a trace is present.
That helper uses the state's UID when it is the in-memory trace index and falls
back to the mirrored state object identity for worker-local trace UIDs.

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
- a cleaner thread deletes old pool files.
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
- `liveCheck`: liveness subsystem.
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
- Java's `tlc2.tool.ModelChecker.vetoCleanup` property forces metadata
  retention even when `-cleanup` was requested; Go mirrors it with the same key
  plus `TLAGO_MODEL_CHECKER_VETO_CLEANUP`.
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
- Preserve Java's error-time postcondition behavior: init failures with an
  `errState` call `checkPostConditionWithCounterExample(new CounterExample(errState))`,
  and worker `doNextSetErr` paths build a safety counterexample, evaluate aliases
  over that trace, and call the postcondition hook before returning the safety
  error. Evaluation failures remain on Java's separate `doNextEvalFailed` path.

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
uses a `LoadTraceFunc` hook at this layer until SANY semantic-node-to-state
wiring is available without introducing package cycles. The command option
parser mirrors Java's `CheckImplFile.main` surface: `-config`, `-deadlock`,
`-recover`, `-workers`, `-depth`, `-trace`, `-coverage`, and the root module.

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
  ordering.
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
  bindings that are TLC `Value`s.
- Base `TLCStateInfo.getStateNumber()` returns the stored trace ordinal.
  Java's `AliasTLCStateInfo` overrides it to return `originalState.getLevel()`;
  the Go alias path keeps the retained original state and follows that override.
- Before replaying trace fingerprints into concrete states, Java snapshots and
  resets `RandomEnumerableValues`; restore the snapshot after replay so specs
  using randomized enumeration regenerate the same path without perturbing the
  active runtime RNG.
- Java's no-alias behavior returns the current state/info, not the successor.
  Go `EvalAlias`, `EvalAliasInfo`, and `EvalAliasInfoPair` must preserve that.
- The Java default `evalAlias` overloads build a prefix supplier from prefix
  and suffix arrays. Go uses concrete `Tool` helper methods that build the same
  supplier closure without introducing a separate interface.
- Safety-error postcondition traces run alias evaluation over each trace entry
  before wrapping the trace in `CounterExample`, matching Java's
  `Worker.doPostCondition`. The `console` field is omitted only when the checker
  was not already done at the moment the error was accepted.
- Java also calls `TLCTrace.printTrace(curState, succState)` for next-state
  safety errors and for continuation-mode invariant/action-property violations.
  Go mirrors this with `ModelChecker.printBehaviorTrace`, which prints
  `TLC_BEHAVIOR_UP_TO_THIS_POINT`, aliases each recovered state, and emits
  state-printer messages in ordinal order before continuing or finishing the
  queue.

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
  worker struct.
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
- Simulator result consumption follows Java's continuation policy. Worker
  exceptions and liveness exceptions stop the run, and
  `TLC_INVARIANT_EVALUATION_FAILED`,
  `TLC_ACTION_PROPERTY_EVALUATION_FAILED`, and
  `TLC_STATE_NOT_COMPLETELY_SPECIFIED_NEXT` are non-continuable regardless of
  `-continue`. Other behavior errors keep workers running only when
  `Globals.Continuation` is set.

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
- rebuild the pointer table from `ptrs_N` before SCC search or recovery.
- treat values below `MAX_PTR` as node-file pointers and values in
  `[MAX_PTR, MAX_LINK]` as SCC link numbers.
- support checkpoint/recover by saving and restoring the current file pointers.
- provide optional fixed-size node caching, invariant checks over all graph
  records, and DOT/string traversal helpers for debugging.
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
  The structural core is in place; bounded quantifier expansion, function
  lambda values, and exact WF/SF subscript handling still need the same
  breadth-first mechanical deepening as their Java counterparts.

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

`TLCExt!CounterExample` is context-sensitive in Java: postcondition checking
conses the current `CounterExample` value into the evaluation context and the
module operator returns that value when present, otherwise an empty
counterexample. The Go standard operator is therefore registered as an
evaluating operator rather than a plain method so it can read the current
postcondition context.

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
BFS mode.

Preserve Java override annotations when registering standard operators.
`TLC!TLCEval` is an evaluating override because it receives the unevaluated
expression and caches converted constant-level results on the semantic node.
Its Java implementation branches directly on the argument semantic level:
state/action/temporal expressions evaluate in the incoming context and are not
cached; constant-level expressions with a non-empty context also evaluate
without caching; only constant-level expressions under an empty context use the
static read/write-lock protected semantic-node cache. Cache reads must mux
`WorkerValue` through the active worker id before returning it, and cache writes
must evaluate under `EmptyContext`/`EmptyState` before converting through the
legacy `toSetEnum`/`toFcnRcd` path.
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
normalization. This is observable for seeded runs and aril replay. Java does
not reject a negative integer `k` for `RandomSubset`; it simply yields the
empty set because the enumerator has no next element.

`TLC!RandomElement` also uses `RandomEnumerableValues.get().nextDouble()` for
intervals and finite enumerated sets. The Go port must use the shared
Java-compatible random enumerable generator here, not Go's process-global
random source, so seeded simulation and trace replay consume random values in
the same places Java does.

`RandomEnumerableValues` is mode-sensitive in Java. During BFS model checking,
worker evaluation and trace reconstruction set the current predecessor state;
the random-enumerable generator is then seeded with
`enumFractionSeed XOR predecessor.fingerPrint()`, and repeated random choices
within the same predecessor continue the same stream. Initial-state generation,
DFID, and simulation use the default seed stream without predecessor reseeding.

`Json!ToJsonObject` mirrors Java's `getObjectNode` dispatch. Records and
tuples become JSON objects, but a function record whose domain is a valid
sequence `1..n` is routed back to the array writer even under `ToJsonObject`.
This oddity is source-compatible with Java and must be preserved for
round-tripping existing specs.

`Sequences!Insert` is part of Java's live `Sequences` surface and performs an
insertion-sort style scan from the right, calling the supplied binary predicate
and also requiring the inserted value to compare less than the shifted element.
The historical `Sequences!Remove` helper is commented out in Java and should
not be registered unless upstream re-enables it.

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
normalized base elements. The Go randomization module follows those branches
instead of enumerating the powerset first.

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
- `_POSSIBLE` is also a config-driven model transformation in Java
  `SpecProcessor`. Each configured predicate is wrapped in `_Possible!_Track`
  and installed as a model constraint for state-level predicates or an action
  constraint for action-level predicates. A matching `_Possible!_CheckName`
  postcondition reports `TLC_POSSIBLE_UNWITNESSED` after model checking if the
  named predicate was never witnessed. The Go bridge mirrors this with concrete
  `PossibleTrackNode` and `PossibleCheckNode` structs and stores counts in the
  named register `s:_possible`, using worker-local checker values so `_Counts`
  can merge them like Java's `TLCGet("all:named")`.

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
- `ErrorTraceMessageRecorder` mirrors Java's
  `ErrorTraceMessagePrinterRecorder`: it observes state-print/back-to-state
  message codes and builds an `MCError` for TE generation. The `TLC` runner
  subscribes it only when `Options.GenerateTraceSpec` is set.

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
- `TLCDebugger`: owns breakpoints, exception-breakpoint filters, the active
  stack-frame list, stepping state, granularity, halt flags, and the connection
  to the debug adapter. The protocol transport is less important than preserving
  where model-checker/evaluator state is captured.
- `DebugTLCVariable`: adapts TLC `Value` objects into debugger variables.
  Scalars expose `type` and `value`; enumerable/function/record/tuple values
  receive a non-zero `variablesReference` and lazily produce children.
- `TLCSourceBreakpoint`: stores source line/column, optional hit count, optional
  log message, a parsed condition operator, and the source `Location`. Location
  matching succeeds for `nullLoc`, otherwise it checks equal line and breakpoint
  column less than or equal to the semantic node begin column. Conditional
  breakpoints evaluate through `tool.noDebug().eval` and swallow evaluation
  failures so a broken debugger expression does not crash TLC.
- `TLCCapabilities` and `GotoStateEvent`: small protocol data types. They are
  useful in Go as plain structs even before a debug-adapter server exists.

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
6. Remaining `(fp, successor, predecessor)` triples are sorted by fingerprint in
   a `TreeSet<Holder>`. Equality is fingerprint-only; two states with the same
   fingerprint collapse at this stage just as Java's `Holder.compareTo` does.
7. Sorted holders are partitioned by `fpSetManager.getFPSetIndex(fp)` into
   parallel vectors of predecessors, successors, and fingerprints.
8. `fpSetManager.containsBlock` returns bit vectors whose set bits identify
   fingerprints not yet present.
9. Only those unseen states are checked with `work.checkState`,
   `work.isInModel`, and `work.isInActions`. Passing states inherit the
   predecessor UID and are returned in `NextStateResult`.

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
- Tests under `tlc2/tool/distributed` should remain late-stage tests.
- The Go port should keep `TLCServer`, `DistributedWorker`,
  `DistributedFPSetManager`, and `NextStateResult` concrete. Transport can wrap
  these structs later.
- Distributed fingerprint checkpointing is filename-paired. Java
  `FPSetManager.Checkpoint.run` calls `beginChkpt(filename)` and
  `commitChkpt(filename)` on the same remote FP set; the later manager-level
  `commitChkpt()` is a no-op for the distributed manager. The Go concrete
  `DistributedFPSetManager` records the active checkpoint filename so
  `Checkpoint(filename)` and `CommitCheckpoint()` commit the same file-named
  checkpoint.
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
- Do not use a directory named `testdata/`.
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
