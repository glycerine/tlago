# tlago Porting Plan

`tlago` is the Go port of the command-line parts of the TLA+ tools: SANY-style
parsing and semantic checking first, then TLC-style finite-state checking. The
default architecture is one central Go library, `package tlago`, with tiny
command wrappers. When placement is unclear, put code in the central package to
avoid circular imports while the port is still converging.

## Current Active Work

The active goal at the time of handoff is the Go TLC model checker port under
`tlc/`, not the older SANY XML or ApalacheIR corpus sweeps. Read
`tlc/HANDOFF.md` first, then `tlc/PORT_PROGRESS.md`, before resuming work.

Current TLC method:

- Continue breadth-first mechanical parity against Java TLC in
  `../tlaplus/tlatools/org.lamport.tlatools/src/tlc2`.
- Keep the Go implementation mostly in package `tlc`.
- Store persistent TLC fixtures in `tlc/test_vectors/`. Do not name a fixture
  directory `testdata/`; the user reserves that name for ephemeral Go fuzzer
  storage and cleanup.
- Prefer concrete structs over interfaces.
- Use `InsMap` or explicit slices/sorting for any iterated map with observable
  order.
- Keep `tlc/PORT_PROGRESS.md` current after every no-code audit or code fix.
- Implement each Java feature accurately in Go first. Once that feature is
  ported, port its Java tests too when they exist. This is the user's latest
  instruction and supersedes the earlier instruction to defer all new tests.
- Keep the existing Go suite green as the port advances.
- Do not restart the long SANY XML or ApalacheIR corpus sweeps unless the user
  explicitly asks.

## Java Survey

The Java source of interest is under `tlaplus/tlatools/org.lamport.tlatools/src`.

- `tla2sany.SANY` is a shell that calls `tla2sany.drivers.SANY.SANYmain`.
- `tla2sany.drivers.SANY` owns the command-line contract, stage selection, exit
  codes, warning suppression/elevation, and calls into parse, semantic analysis,
  level checking, and linting.
- `tla2sany/parser` contains the JavaCC-generated parser and tokenizer. The
  original grammar is `javacc/tla+.jj`. Operator precedence and synonyms live in
  `tla2sany/parser/Operators.java`.
- `tla2sany/modanalyzer` resolves modules, parses dependency closure, enforces
  file/module name matching, records EXTENDS/INSTANCE relationships, orders
  modules for semantic processing, and rejects dependency cycles.
- `tla2sany/semantic` builds the semantic graph, resolves symbols, checks levels,
  and records standardized diagnostics in `Errors`.
- `tla2sany/StandardModules` contains built-in library modules that must be
  available without user configuration.
- `tlc2.TLC` is the TLC command-line entry point. It parses `.cfg` files,
  invokes SANY, evaluates constants and operators, enumerates initial/next
  states, checks invariants/properties, supports simulation, and reports traces.
- Java tests already group behavior usefully: syntax corpus, tokenizer state,
  operator precedence/associativity, semantic corpus/error corpus, XML export,
  driver options, and TLC model tests.

## Architecture

Keep the implementation mostly in `tlago/` as `package tlago`.

- `cmd/tlago` and future command names should only parse OS process wiring and
  call library functions.
- Shared data types stay central: source locations, diagnostics, tokens, AST,
  operator table, module resolver, semantic environment, value system, config
  model, state graph, and CLI option structs.
- Internal boundaries should be expressed as files and type names, not Go
  subpackages, until the port is mature enough to prove there are no cycles.
- Use BDD-style Go tests at the library boundary. Prefer behavior narratives
  such as "a spec extending a sibling module resolves and checks" over tiny unit
  tests for every helper.

## Parser Port Direction: Full SANY Grammar Only

The early compact parser has been removed. All public parsing, loading,
checking, CLI, and model-check source helpers now go through the Go SANY parser
port. New parser work must extend the SANY token manager, syntax parser, and
SANY-to-central-AST translator directly; do not add a second parser, fallback
parser, compatibility shim, legacy alias, or duplicate syntax path. When an API
or AST name changes, rename all use sites to the production SANY name.

The parser port is reference-driven:

- Generate Go token constants, token images, and syntax node constants from
  `TLAplusParserConstants.java` and `SyntaxTreeConstants.java`.
- Port the JavaCC token manager states from `javacc/tla+.jj`: `DEFAULT`,
  `PRAGMA`, `SPEC`, `IN_COMMENT`, `EMBEDDED`, and `IN_EOL_COMMENT`.
- Preserve Java token chains: source begin/end positions, lexical state, normal
  token links, and special-token comment/pragma links.
- Port `SyntaxTreeNode` shape before semantic expansion: node kind, heirs,
  token-backed leaves, module/file origin, and exact locations.
- Port `JunctionListContext` and `OperatorStack` as first-class parser
  machinery. Indentation-sensitive junction lists and precedence reductions are
  not optional helpers; they define valid TLA+ parsing.
- Translate all grammar productions from `tla+.jj` into central Go parser
  methods before adding more ad hoc expression grammar behavior.
- Drive the replacement with BDD conformance tests that compare Go parser
  outcomes against Java SANY over a corpus of accepted and rejected specs.

## Direct SANY XML Conformance Oracle

Pause ApalacheIR parity until the direct Go SANY parse and semantic graph are
conformant with Java SANY. The primary oracle is Java SANY's native XML export,
because it mirrors SANY parse nodes, symbols, locations, and operator
applications before Apalache rewrites or omits information.

Target workflow:

- Run Java SANY from the frozen fixture jar:
  `java -cp test_vectors/java-sany/tla2tools.jar tla2sany.xml.XMLExporter -o -I test_vectors/java-sany/StandardModules Spec.tla`.
- Use `-o` and capture stdout directly. Do not use sockets, network, local
  schema proxies, or patched Java/Scala code in the oracle path.
- Keep the Java oracle self-contained by freezing `tla2tools.jar`, the SANY
  XML schema, JavaCC grammar, generated parser constants, Java operator table,
  Java syntax corpus, Java semantic corpus, and Java standard modules under
  `test_vectors/java-sany/`.
- Run Go on the same staged spec with `tlago sany-xml Spec.tla`, loading frozen
  Java standard modules and frozen community modules from `test_vectors/`.
- Canonicalize both XML documents through an in-memory tree before diffing:
  normalize unstable UIDs, sort context entries and final module references,
  but preserve semantic child order inside expressions and declarations.
- The hard acceptance target is canonical SANY XML parity for every `.tla` file
  under `test_vectors/tla-plus-bench/specs/` and
  `test_vectors/Examples/specifications/`.

## ApalacheIR Conformance Oracle

For parser and semantic-IR conformity, use Apalache's SANY-backed import path as
an executable oracle. Apalache parses TLA+ with Java SANY, translates the
semantic graph into `TlaModule`/`TlaDecl`/`TlaEx`, and serializes ADR-005 JSON
with `TlaToJson`.

Target workflow:

- Run Apalache on the same spec: `apalache-mc parse --output=out.json Spec.tla`.
- Run Go tlago on the spec: `tlago apalache-json Spec.tla`.
- `tlago apalache-json` emits compact canonical JSON by parsing the generated
  JSON into an in-memory tree and writing object keys in lexicographic order.
  Array order remains semantic and is not globally sorted.
- Normalize sources and unstable metadata while range tracking is still being
  audited.
- Diff normalized JSON for structural conformance, then tighten toward
  byte-for-byte output once declaration ordering, source ranges, type tags, and
  all operator encodings match Apalache.
- The hard acceptance target is canonical ApalacheIR JSON parity for every
  `.tla` file under `test_vectors/tla-plus-bench/specs/` and
  `test_vectors/Examples/specifications/`, comparing against locally generated
  `apalache-mc parse --output=...` oracle output. The corpus harness parses
  Apalache's JSON into memory, lexicographically sorts object keys without
  modifying the upstream Scala/Java implementation, and sorts top-level module
  declarations and LET-local declarations by declaration identity so
  conformance is not coupled to JVM hash-table enumeration or JSON writer field
  order. Expression argument arrays remain ordered.

The Go implementation should mirror Apalache's layering without introducing a
second parser: production SANY syntax tree -> central semantic AST -> central
ApalacheIR model/JSON encoder. Operator-name mapping must be explicit and
auditable, using Apalache names such as `OPER_APP`, `PLUS`, `SET_ENUM`,
`FUN_APP`, and `IF_THEN_ELSE`.

## Command-Line Targets

Initial commands:

- `tlago parse FILE...`: parse modules and dependencies, print diagnostics, and
  exit nonzero on syntax/module-resolution failure.
- `tlago check FILE...`: parse, run semantic checks and level checks, print
  diagnostics, and exit nonzero on failure.
- `tlago sany-xml FILE...`: parse and check modules through the production SANY
  path, then emit Java-shaped SANY XML for direct conformance diffing.
- `tlago apalache-json FILE...`: parse and check modules through the production
  SANY path, then emit ADR-005 ApalacheIR JSON for conformance diffing and
  downstream translators.
Implemented early:

- `tlago modelcheck [options] SPEC`: TLC-compatible finite-state checker.

Later commands:

- `tlago simulate [options] SPEC`: TLC-compatible random simulation.
- `tlago ast` and `tlago fmt` can follow after parser parity.

SANY-compatible exit codes:

- `0`: OK.
- `2`: syntax parsing or module-resolution failure.
- `4`: semantic analysis or level-checking failure.
- `1`: CLI usage/internal tool failure. Java SANY uses `-1`, but process exit
  status is normalized by shells; Go tools should use `1`.

## Behavior Inventory

### Lexer And Parser

- Ignore arbitrary pre-module text until a `---- MODULE Name ----` header.
- Tokenize module headers, module footers, identifiers, numbers, strings,
  keywords, built-in symbolic operators, punctuation, and pragmas.
- Treat `\*` comments as end-of-line comments.
- Treat `(* ... *)` comments as block comments, including nested block comments.
- Preserve source locations as file, line, and column for all diagnostics.
- Parse root modules and nested modules delimited by TLA+ module headers and
  footers.
- Parse EXTENDS clauses and discover module dependencies.
- Parse CONSTANT/CONSTANTS and VARIABLE/VARIABLES declarations, including
  operator constants with arity.
- Parse operator definitions, including prefix/infix/postfix declarations where
  TLA+ allows them.
- Parse LET/IN, IF/THEN/ELSE, CASE/OTHER, records, tuples, sets, functions,
  quantification, CHOOSE, EXCEPT, fairness, temporal operators, proof syntax,
  and labels.
- Match SANY operator precedence, associativity, and synonyms from
  `Operators.java`.
- Reject ambiguous or non-associative operator chains, for example `A = B = C`.
- Keep an AST form stable enough for semantic checking, model checking, and
  optional parse-tree conformance tests.

### Module Resolution

- If a root argument omits `.tla`, append it.
- Resolve root-relative sibling modules before library modules.
- Resolve `TLA-Library`-style search paths.
- Resolve bundled standard modules without configuration.
- Enforce that a file named `Foo.tla` contains top-level module `Foo`.
- Detect missing modules with the importing module named in the diagnostic.
- Detect circular EXTENDS/INSTANCE dependency cycles.
- Analyze dependencies before dependents.
- Mark standard modules so linter/model checker can treat them specially.

### Semantic Checking

- Build one semantic environment per module, with contexts following EXTENDS,
  INSTANCE, and nested-module visibility rules.
- Resolve identifiers to declarations/definitions/built-ins.
- Reject duplicate declarations and illegal redefinitions.
- Check arity of operator calls, declared operator constants, and substitutions.
- Check INSTANCE and module substitution semantics.
- Validate recursive declarations and recursive operator definitions.
- Run level checking: constants, variables, actions, temporal formulas, priming,
  ENABLED, UNCHANGED, fairness, and proof-related restrictions.
- Report standardized diagnostics with stable codes, severity, source location,
  and message text.
- Support message suppression and message-as-error behavior.
- Support parse-only (`-s`) and semantic-without-level-check (`-l`) modes.
- Warn when PlusCal translations are stale; translation itself can be deferred.
- Add linter passes only after semantic parity for core specs.

### TLC-Style Checking

- Parse `.cfg` files: SPECIFICATION, INIT, NEXT, CONSTANT(S), INVARIANT(S),
  PROPERTY, CONSTRAINT, ACTION_CONSTRAINT, VIEW, SYMMETRY, ALIAS, POSTCONDITION,
  and overrides.
- Evaluate TLA+ expressions over finite values: booleans, integers, strings,
  model values, sets, tuples, records, functions, sequences, and standard
  modules.
- Support finite enumeration for sets, quantifiers, functions, records, tuples,
  subsets, and bounded integer intervals.
- Build initial states from INIT or SPECIFICATION.
- Generate next states from NEXT/action formulas, including primed variables and
  unchanged variables.
- Detect deadlocks unless disabled.
- Check invariants and implied init/action properties.
- Check temporal properties and fairness after safety behavior is stable.
- Produce readable counterexample traces and later TLC-compatible trace formats.
- Implement deterministic fingerprints and duplicate-state detection.
- Support worker parallelism after single-threaded correctness is proven.
- Add simulation mode with seed/aril/depth/trace count.
- Integrate Z3 for optional bounded/symbolic checks after explicit-state TLC
  behavior exists; Z3 must not replace normal finite-state TLC semantics.

## BDD Test Plan

Each behavior test should read like a scenario and may cover many helper
functions.

- Parser accepts a minimal module and preserves module name and locations.
- Parser ignores pre-module text and nested comments.
- Parser handles declarations, EXTENDS, ASSUME, THEOREM, and operator
  definitions in one ordinary module.
- Parser follows SANY operator precedence and rejects non-associative chains.
- Resolver loads a sibling dependency and a bundled standard module.
- Resolver reports missing modules and file/module name mismatches.
- Semantic checker accepts a small counter spec with `Init` and `Next`.
- Semantic checker rejects duplicate declarations and undefined identifiers.
- Level checker rejects illegal priming of constants.
- CLI `parse` and `check` return SANY-style exit codes.
- Config parser selects INIT/NEXT/INVARIANT entries from a `.cfg` file.
- Model checker explores a tiny finite counter, reports state count, and finds
  an invariant violation with a trace.
- Simulator uses seed/depth settings reproducibly.

## Implementation Phases

1. Central package scaffold, diagnostics, resolver, AST shell, CLI wrappers, and
   BDD behavior tests.
2. Reference-synchronized SANY constants, token images, syntax node IDs, lexer
   states, token links, syntax tree nodes, junction contexts, and operator-stack
   scaffolding.
3. Full JavaCC token manager port, including pragma/spec/comment modes,
   Unicode operators, proof-step lexemes, source positions, and special tokens.
4. Full `tla+.jj` grammar production port into central Go parser methods, with
   the SANY syntax tree as the primary output.
5. Keep the single SANY parser path wired through CLI, resolver, semantic
   checking, and model checking.
6. Module dependency closure and standard module embedding.
7. Semantic environment and common SANY diagnostics.
8. Level checking for core TLA+.
9. `.cfg` parser and expression evaluator.
10. Single-threaded explicit-state model checker, TLC output/traces,
    simulation, compatibility options, parallel checking, liveness, XML/AST
    export, formatter, and optional Z3-assisted workflows.

## Completed BDD Slices

The first executable slice is passing. These behaviors keep the CLI and early
checker usable through the production SANY parser path:

- A minimal module parses.
- Pre-module text and nested comments are ignored.
- EXTENDS and declarations are recorded.
- Operator precedence for ordinary arithmetic/comparison/connective expressions
  matches SANY for representative cases.
- Non-associative operator chains fail.
- A root spec can resolve a sibling module and a standard module.
- File/module name mismatches are reported.
- A basic semantic check catches duplicate declarations and undefined names.
- `tlago parse`, `tlago check`, and `tlago modelcheck` expose behavior through
  one CLI.
- `tlago apalache-json` emits a compact ADR-005 root/module/declaration/
  expression JSON tree for the supported central AST subset, matching
  Apalache's untyped serializer field names and operator names for constants,
  variables, ordinary operator declarations, assumptions, literals, user
  operator applications, arithmetic, set membership, set enumeration, and
  conditionals.
- A gated ApalacheIR oracle test runs `apalache-mc parse --output=...` when
  `APALACHE_MC` or a PATH binary is available, removes known-unstable `source`
  fields, and structurally diffs the resulting JSON against Go output.
- ApalacheIR module emission applies a stable define-before-use declaration
  sort over the central AST, mirroring Apalache's layer-by-layer parser pass
  before JSON output.
- ApalacheIR expression emission now resolves nullary user-defined operators as
  `OPER_APP(NameEx(...))` while preserving constants, variables, formal
  parameters, and bound variables as `NameEx`, matching Apalache's SANY
  importer behavior.
- ApalacheIR operator declarations now preserve `RECURSIVE` status in the
  `isRecursive` flag, including the Apalache oracle coverage for recursive
  operator definitions.
- The central AST now preserves names on assumption and theorem facts via
  `NamedExpr`; ApalacheIR emits named assumptions as `TlaAssumeDecl` with the
  Apalache `name` field.
- ApalacheIR behavior coverage now includes structured expression forms:
  tuples, records, record sets, function constructors/applications, set filters
  and maps, LET/IN, bounded and unbounded quantifiers, temporal quantifiers,
  bounded and unbounded CHOOSE, CASE/OTHER, and EXCEPT updates.
  Temporal/action coverage includes prime, ENABLED, UNCHANGED, square and
  angle actions, temporal modalities, leads-to/guarantees, and weak/strong
  fairness. These are now covered by a gated Apalache oracle test in addition
  to local shape tests.
- ApalacheIR conformance now covers representative core built-ins for boolean
  logic, equality/inequality, set membership and set algebra, standard sets,
  function sets, composition, and Cartesian product. Java `N_Times` operands
  are preserved through the SANY bridge, and ApalacheIR emits n-ary
  `SET_TIMES` to match the oracle for products such as `x \X x \X x`.
- ApalacheIR conformance now covers ordinary function definitions, multi-index
  function application, and recursive function definitions. Recursive function
  definitions emit `FUN_REC_CTOR`, while ordinary function constructors remain
  `FUN_CTOR`, matching Apalache's SANY import path.
- ApalacheIR source-range conformance now has oracle coverage with source
  fields enabled for declarations, assumptions, operator bodies, nested
  expressions, function constructors, LET bodies, CASE/OTHER, and EXCEPT.
  The central AST carries end positions and per-name declaration positions, and
  ApalacheIR emits module-style source filenames to match Apalache.
- `tlago modelcheck` accepts `-maxStates N`/`--max-states N` to bound explicit
  state exploration from the command line.
- `.cfg` parsing supports INIT, NEXT, INVARIANT(S), PROPERTY/PROPERTIES,
  SPECIFICATION, integer CONSTANT(S), CONSTRAINT(S), ACTION_CONSTRAINT(S),
  VIEW, SYMMETRY, ALIAS, POSTCONDITION, and CHECK_DEADLOCK FALSE.
- The first explicit-state model checker explores deterministic finite integer
  specs with guarded primed assignments, detects deadlocks, checks invariants,
  and reports counterexample traces.
- The model checker also supports finite nondeterminism through integer set
  enumeration and membership assignments such as `x \in {0, 1}` and
  `x' \in {x, x + 1}`.
- Expression evaluation supports integer/boolean `IF/THEN/ELSE`.
- The model checker applies `.cfg` state constraints to initial and successor
  states.
- The model checker supports `UNCHANGED x` and `UNCHANGED <<x, y>>` action
  conjuncts.
- The model checker evaluates `.cfg` `POSTCONDITION(S)` once after successful
  exploration as constant-level, zero-arity boolean operators.
- The model checker evaluates module `ASSUME` clauses in the constant
  environment before computing initial states.
- Reference tests now require generated Go SANY token and syntax-node tables to
  match the local Java implementation before grammar-port work continues.
- Legacy bootstrap-parser entry points and files have been removed; tests and
  tools call the production SANY parser path directly.
- Common community helper modules and TLC runtime helper operators are bundled
  enough for parser/semantic passes to progress through more corpus specs.
- Junction-list operands following infix operators parse without being split at
  the outer connective.
- A SANY token manager now emits Java token kinds for default, pragma, and spec
  lexer states; it preserves token locations, longest-match literals, nested
  block comments, line comments, proof-step lexemes, and special-token chains.
- The SANY operator table is generated from `Operators.java`, including
  canonical operators, low/high precedence, fixity, associativity, and Unicode
  or symbolic synonyms used by `OperatorStack`.
- The JavaCC grammar production inventory is generated from `tla+.jj` so the Go
  parser port has an explicit coverage checklist for every SANY production.
- A parser production coverage guard now checks the generated JavaCC inventory
  against `SanyParser` methods and requires every production to have either a
  parser method or an explicit folded/token-level coverage note; deferred
  parser productions fail the guard.
- The first SANY parser shell builds syntax-tree nodes for the compilation
  unit/module spine, EXTENDS, CONSTANT, VARIABLE, and flat operator-definition
  body items while preserving Java's `zero`/`one` heir split and pre-comments.
- The SANY parser now accepts nested module body items, producing nested
  `N_Module` syntax-tree nodes inside the enclosing `N_Body`.
- `OperatorStack` now performs SANY-style precedence reductions for prefix,
  infix, postfix, and n-fix operators, including non-associative conflict
  detection.
- The SANY parser shell now reduces top-level operator-definition expressions
  through `OperatorStack` instead of storing expression tokens as a flat list.
- The `tlago parse` command exercises the shared SANY token manager and syntax
  parser path.
- The SANY-to-central-AST translator exposes `ParseSanyModuleSource` and
  `CheckSanySource`, translating modules, declarations, simple operator
  definitions, infix/prefix/postfix expressions, tuples, sets, conditionals,
  quantifiers, LET/IN, and operator applications enough for semantic duplicate,
  undefined-name, and primed-constant checks.
- `ModelCheckSanySource` now runs the explicit-state model checker over the
  SANY-translated AST for the finite counter behavior, proving the bridge can
  feed executable specs for the supported expression subset.
- The SANY model-check bridge now covers finite nondeterministic set membership
  and `IF/THEN/ELSE` invariants, including translation of `TRUE` and `FALSE`
  identifiers to central boolean literals.
- The SANY model-check bridge now covers `UNCHANGED` action conjuncts for both
  single variables and tuple variables through translated prefix and tuple
  expressions.
- The SANY model-check bridge now covers parameterless `LET/IN` definitions in
  actions and invariants through translated local definitions.
- The model checker now inlines parameterless operator references and calls to
  parameterized operator definitions in actions and invariants.
- The model checker now inlines user-defined unary and binary operator
  definitions, including SANY fixity-style infix operators.
- The model checker now indexes loaded dependency definitions with qualified
  names, so SANY-loaded specs can call helpers such as `Helper!Inc(x)`.
- The SANY model-check bridge now covers finite universal and existential
  quantifiers over integer sets through translated `N_BoundQuant` nodes.
- The SANY model-check bridge now translates multi-variable bounded quantifiers
  such as `\A a, b \in S` and `\A a \in S, b \in T` into nested finite
  quantifiers.
- The SANY model-check evaluator now enumerates bounded existential action
  quantifiers such as `\E n \in S: x' = n` to generate successor states.
- The SANY model-check bridge now covers integer interval expressions such as
  `0..N` in initialization and quantified domains.
- `LoadSanySpec` now loads root modules, sibling dependencies, configured
  library paths, and embedded standard modules through the SANY parser bridge;
  sibling-module model checking works with qualified names such as
  `Helper!Zero`.
- The `tlago check` and `tlago modelcheck` commands load modules through
  `LoadSanySpec`.
- The CLI usage text now advertises the active `parse`, `check`, and
  `modelcheck` command surface.
- `tlago check` now supports SANY-style warning controls:
  `-suppressMessages`/`--suppress-messages` hides selected diagnostic codes,
  and `-messagesAsErrors`/`--messages-as-errors` elevates selected warnings to
  semantic failures.
- The old bootstrap parser files and old generic source helper names were
  removed; tests and callers now use the SANY-named API surface directly.
- The SANY parser and model-check bridge now cover prefix conjunction and
  disjunction lists such as `/\ A /\ B` and `\/ A \/ B`, including indented
  action-style junction lists.
- The SANY expression parser now recognizes simple weak and strong fairness
  expressions such as `WF_<<x>>(A)` and `SF_<<x>>(A)`, producing
  `N_FairnessExpr` nodes.
- The SANY parser now handles Java semantic-corpus fairness forms such as
  `WF_RefersTo(vars, "vars")(A)` by splitting leading `WF_`/`SF_` identifier
  tokens in expression context and parsing the subscript through JavaCC-style
  restricted expressions.
- The SANY semantic bridge now also translates call-style fairness parse trees
  such as `WF_vars(Next)` and `SF_vars(Next)`, allowing real examples such as
  `CigaretteSmokers.tla` to check cleanly.
- The SANY theorem parser now recognizes terminal proof forms such as
  `PROOF OMITTED`, `PROOF OBVIOUS`, and simple `BY` references, producing
  `N_TerminalProof` nodes after theorem expressions.
- The SANY proof parser now recognizes simple zero-step `PROOF QED` blocks,
  producing `N_Proof`, `N_ProofStep`, and `N_QEDStep` nodes.
- The SANY proof parser now recognizes numbered assertion proof steps,
  including `SUFFICES` assertions and bare-level `QED` steps.
- The SANY proof parser now recognizes named proof command steps for `HAVE`,
  `TAKE`, `WITNESS`, `PICK`, `CASE`, `USE`, `HIDE`, and `DEFINE`.
- The SANY proof parser now recognizes bounded `TAKE` proof steps and
  unbounded identifier-list `PICK` proof steps.
- The SANY proof parser now recognizes `ASSUME ... PROVE ...` assertion
  bodies, including their use under `SUFFICES`.
- The SANY theorem parser now recognizes `ASSUME ... PROVE ...` theorem
  bodies instead of treating them as ordinary expression token streams.
- The SANY ASSUME/PROVE parser now recognizes `NEW` declarations for fresh
  symbols, variables, constants, and state/action/temporal declarations.
- The SANY expression parser now recognizes simple labeled expressions of the
  form `Label:: expr`, producing `N_Label` nodes with `N_GeneralId` labels.
- The SANY expression parser now recognizes temporal quantifiers `\EE` and
  `\AA`, producing `N_UnboundQuant` nodes.
- The SANY expression parser has behavior coverage for temporal prefix
  operators such as `ENABLED A` and `<>A`.
- The SANY semantic bridge now translates record constructors and record field
  access so `tlago check` accepts record-heavy specs without model evaluation.
- The SANY semantic bridge now translates function constructors and function
  applications, including local binder handling for constructor bodies.
- The SANY semantic bridge now translates SANY `N_FunctionDefinition` module
  items into function-valued definitions with bounded parameters.
- The SANY semantic bridge now translates fixity-style definition LHS nodes
  (`N_InfixLHS`, `N_PrefixLHS`, and `N_PostfixLHS`) into named definitions
  with operand parameters.
- The SANY semantic bridge now translates subset and mapped set comprehensions,
  including local binder handling for comprehension bodies and predicates.
- The SANY semantic bridge now translates record-set and function-set
  expressions such as `[a: S]` and `[S -> T]`, traversing their component sets
  during semantic checks and model-checking expression rewrites.
- The SANY semantic bridge now translates `EXCEPT` updates with dot and indexed
  components, including traversal of update paths and replacement expressions.
- The SANY semantic bridge now translates labeled expressions as transparent
  wrappers so `Label:: expr` participates in semantic checks and rewrites.
- The SANY semantic bridge now translates square/angle action wrappers and
  weak/strong fairness expressions, traversing their action and subscript
  expressions during checks and rewrites.
- The SANY semantic bridge now translates open-expression binder forms:
  `CASE`, bounded/unbounded `CHOOSE`, unbounded quantifiers, and `LAMBDA`,
  preserving local-name scope during checks and expression rewrites.
- The SANY semantic bridge now translates JavaCC-compatible `N_Real` literals
  as numeric literals.
- The semantic checker now reports operator-call arity mismatches for top-level
  and local operator definitions while respecting binder shadowing.
- The semantic checker now reports duplicate formal parameters on top-level and
  LET-local operator definitions.
- The semantic checker now records arity for declared operator constants such
  as `CONSTANT F(_)` and reports mismatched calls.
- The SANY model-check evaluator now executes transparent labels and ordered
  `CASE` expressions in boolean, integer, finite-set, and action contexts.
- The SANY model-check evaluator now executes bounded finite integer `CHOOSE`
  expressions by enumerating the domain and applying the predicate.
- The SANY model-check evaluator now enumerates finite integer set
  comprehensions, including mapped elements, finite domains, and predicates.
- The SANY model-check evaluator now executes finite integer function
  applications backed by function constructors or function definitions with
  bounded integer domains.
- The SANY model-check evaluator now executes integer-valued record field
  access over explicit record constructors.
- The SANY model-check evaluator now executes simple integer-valued record
  `EXCEPT` updates with `!.field` paths and `@` replacement values.
- The SANY model-check evaluator now executes simple finite integer function
  `EXCEPT` updates with `![index]` paths and `@` replacement values.
- The SANY model-check evaluator now enumerates `DOMAIN` for finite integer
  function constructors and function `EXCEPT` values.
- The bundled `FiniteSets` module now exposes `Cardinality`, and the
  model-check evaluator computes it over finite integer and string literal
  sets.
- The bundled `TLC` module now exposes `Assert`, and the model-check evaluator
  treats true assertions as boolean successes and reports string messages for
  failed assertions.
- The bundled `TLC` module now exposes `Print`, and the model-check evaluator
  returns its value argument in boolean and integer contexts.
- The SANY model-check evaluator now enumerates finite integer set union,
  intersection, and set difference expressions.
- The SANY model-check evaluator now checks finite integer set equality,
  subset, strict subset, membership, and non-membership predicates.
- The SANY model-check evaluator now supports TLA+ integer division and
  remainder operators (`\div` and `%`).
- The SANY model-check evaluator now supports boolean equivalence operators
  (`<=>` and `\equiv`).
- The SANY model-check evaluator now executes square action wrappers such as
  `[Step]_x` as action-or-stutter next-state relations.
- The SANY model checker now accepts `.cfg` `SPECIFICATION` entries for simple
  safety specs of the form `Init /\ [][Next]_vars`, decomposing them into init
  and next relations.
- The SANY model-check evaluator now computes `ENABLED A` for simple finite
  actions by reusing the action assignment evaluator.
- The model checker now uses integer-valued `.cfg` `VIEW` operators for state
  identity, collapsing states with equal view values.
- Plain non-local `INSTANCE M` declarations now load sibling modules and expose
  their declarations and definitions for semantic checking and SANY model
  checking.
- `LOCAL` operator definitions now remain visible inside their defining module
  while staying private from modules that import the definition's module through
  `EXTENDS`.
- Non-local `INSTANCE` declarations in an extended module now export their
  instantiated symbols to importing modules for semantic checking and model
  checking, while `LOCAL INSTANCE` remains private to its defining module.
- Simple `INSTANCE M WITH name <- expr` substitutions are now represented in
  the central module model and applied to instantiated definitions during SANY
  model checking.
- `INSTANCE M WITH ...` substitutions now also apply to qualified uses such as
  `M!Op` during SANY model checking.
- Named module-instance definitions such as `Alias == INSTANCE Inner WITH ...`
  now produce `N_ModuleDefinition`, register nested modules with the loader,
  and expose instantiated definitions through `Alias!Op`; the Java semantic
  corpus `ConstantModuleSubstitution.tla` checks cleanly.
- The SANY parser now recognizes the JavaCC definition LHS family for identifier
  parameters, function definitions with quantifier bounds, and infix operator
  definitions, producing `N_IdentLHS`, `N_FunctionDefinition`, `N_QuantBound`,
  and `N_InfixLHS` nodes.
- The SANY parser now preserves JavaCC-style operator constant declaration
  syntax such as `CONSTANT F(_), G(_, _)` in `N_ParamDeclaration` trees.
- The SANY parser now recognizes additional JavaCC module-body productions:
  `RECURSIVE`, `INSTANCE ... WITH`, substitutions, named assumptions, and named
  theorems, producing the corresponding SANY syntax-tree nodes.
- The SANY expression parser now wraps primitive number, real, string, and
  parenthesized expressions in JavaCC-compatible `N_Number`, `N_Real`,
  `N_String`, and `N_ParenExpr` nodes.
- The SANY expression parser now recognizes common JavaCC open expressions:
  `IF/THEN/ELSE`, bounded quantifiers, and unbounded quantifiers, producing
  `N_IfThenElse`, `N_BoundQuant`, and `N_UnboundQuant` nodes.
- The SANY expression parser now lets open-expression bodies honor enclosing
  expression stop conditions, so parenthesized quantified/IF/LET/CHOOSE/LAMBDA
  expressions do not consume following operators.
- The SANY expression parser now recognizes simple set enumeration and tuple
  expressions, producing `N_SetEnumerate` and `N_Tuple` nodes.
- The SANY expression parser now recognizes simple subset and set-comprehension
  brace forms, producing `N_SubsetOf`, `N_SetOfAll`, and `N_QuantBound` nodes.
- The SANY expression parser now recognizes simple square-bracket record and
  function constructors, producing `N_RcdConstructor`, `N_FieldVal`,
  `N_FcnConst`, and `N_QuantBound` nodes.
- The SANY expression parser now recognizes simple square-bracket record-set
  and function-set forms, producing `N_SetOfRcds`, `N_FieldSet`, and
  `N_SetOfFcns` nodes.
- The SANY expression parser now recognizes simple function-application
  suffixes such as `F[x, y]`, reducing them through the operator stack to
  `N_FcnAppl` nodes.
- The SANY expression parser now recognizes simple square-bracket `EXCEPT`
  updates with dot and index components, producing `N_Except`,
  `N_ExceptSpec`, and `N_ExceptComponent` nodes.
- The SANY expression parser now recognizes simple square and angle bracket
  action expressions such as `[A]_x` and `<<A>>_x`, producing `N_ActionExpr`
  nodes.
- The SANY parser now treats action subscripts after `]_` and `>>_` as JavaCC
  restricted expressions, so primed/action/temporal step expressions remain
  one call argument when nested inside operator applications.
- The SANY operator stack now reduces dot expressions such as `r.a` to
  `N_RecordComponent` nodes instead of generic infix expressions.
- The SANY expression parser now recognizes simple bang-qualified identifiers
  such as `M!Op`, producing `N_GeneralId`, `N_IdPrefix`, and
  `N_IdPrefixElement` nodes.
- The SANY expression parser now recognizes JavaCC bang structural selectors
  such as `M!<<`, `M!>>`, `M!2`, `M!@`, and `M!:!Field`, producing
  `N_StructOp` nodes inside the final `N_GeneralId` selector tree.
- The SANY expression parser now recognizes JavaCC bang operator selectors
  such as `M!\oplus`, `M!\oplus(1, 2)`, `M!DOMAIN!Field`, and `M!^+`;
  the SANY bridge preserves operator selector images so qualified calls like
  `Helper!\oplus(1, 2)` inline during model checking.
- The SANY expression parser now carries JavaCC `OpArgs` on intermediate bang
  selector-chain elements, including documented shapes such as
  `Foo(1)!Bar!Baz(2, 3)` and `M!F(1)!G`.
- The SANY expression parser now recognizes JavaCC bang argument-only selector
  elements such as `M!(1, 2)` and `M!(1)!Field`.
- The SANY expression parser now recognizes simple operator applications such
  as `F(a, b)`, producing `N_OpApplication` and `N_OpArgs` nodes.
- The SANY expression parser now recognizes simple qualified operator
  applications such as `M!F(a)`, attaching final `N_OpArgs` to the
  `N_GeneralId` as an `N_OpApplication`.
- The SANY expression parser now recognizes additional JavaCC open expressions:
  `LET/IN`, `CASE`, `CHOOSE`, and `LAMBDA`, including LET-local definition
  boundaries and their corresponding syntax-tree nodes.
- `CASE` arm values now honor their enclosing expression boundary, so
  Java semantic-corpus action/temporal definitions such as `action == v'` and
  `temporal == []c` can be used in single-arm and multi-arm CASE expressions
  without consuming the following module body item.
- The model checker applies `.cfg` action constraints using environments with
  both current-state identifiers and primed successor identifiers.
- Parser, semantic checking, and evaluator support parameterless `LET/IN`
  expressions in actions and invariants.
- LET-local definitions now use the same JavaCC LHS family as top-level
  definitions, including infix/postfix symbol definitions and recursive symbol
  references; the Java semantic-corpus `LetInTest.tla` checks cleanly.
- Parser, semantic checking, and evaluator support finite integer-set universal
  and existential quantifiers with `\A` and `\E`.
- The parser accepts prefix conjunction and disjunction lists such as
  `/\ A /\ B` and `\/ A \/ B`.
- The SANY parser now preserves indentation-sensitive nested action junctions,
  including alternating `\/ /\ ...` bodies and prefixed forms such as
  `~ /\ A /\ B`, without treating standalone junction symbols as lists when
  they are passed as higher-order operator values.
- The evaluator enumerates integer intervals such as `0..N` as finite integer
  sets for initialization, membership, and quantifier domains.
- The model checker can evaluate simple integer definitions from loaded
  dependency modules through qualified names such as `Helper!Zero`, and through
  unqualified names imported by the root module's EXTENDS list.
- Semantic checking rejects primed declared constants such as `C'`.
- The model checker checks `.cfg` state-predicate PROPERTY entries on every
  reached state and reports violating traces.
- The `tlago modelcheck` CLI now prints counterexample traces as deterministic
  sorted assignments such as `x = 1, y = 2` instead of raw Go maps.
- The semantic checker now registers declarations and definitions under
  module-qualified names, so `Helper!Missing` is reported as undefined and
  `Helper!Inc(1, 2)` reports an arity mismatch during `check`.
- The SANY semantic bridge now translates expression-only
  `ASSUME ... PROVE ...` theorem bodies into boolean implication expressions,
  so `tlago check` traverses identifiers on both sides instead of reporting
  unsupported `N_AssumeProve` nodes.
- `ASSUME ... PROVE` translation now treats identifier `NEW` declarations as
  theorem-local binders, including optional `\in` domains, so those fresh names
  are not reported as undefined during `check`.
- The semantic checker now validates `INSTANCE ... WITH` substitution targets
  and replacement arities, catching cases such as substituting a scalar for a
  unary operator definition.
- The SANY semantic bridge now carries `RECURSIVE` declarations into the
  central module model, and the semantic checker reports missing recursive
  definitions and recursive arity mismatches.
- The embedded `Sequences` module now exposes `Len`, and the model-check
  evaluator computes `Len(<<...>>)` over tuple literals.
- Finite integer tuple sequences now support `Head`, `Tail`, and `Append`
  through the embedded `Sequences` module during model checking.
- Tuple-backed finite sequences now behave as 1-indexed functions for simple
  integer applications such as `S[2]` during model checking.
- The model-check evaluator now computes `DOMAIN` of tuple-backed sequences as
  `1..Len` and supports `Sequences!SubSeq` over finite integer tuples.
- `Sequences!SelectSeq` now filters finite integer tuple sequences with unary
  `LAMBDA` predicates during model checking.
- Boolean equality and inequality now compare finite integer tuple sequences
  element-by-element, preserving sequence order.
- Boolean equality and inequality now compare string literals in model-checking
  invariants and properties.
- The SANY model-check evaluator now supports finite string literal set
  membership, non-membership, equality, subset, strict subset, union,
  intersection, and difference predicates.
- The SANY model-check evaluator now enumerates finite integer and string
  powersets for `SUBSET`, including subset membership and `Cardinality`.
- The SANY model-check evaluator now compares finite integer and string
  set-of-set literals and computes their `Cardinality`.
- `.cfg` constants now accept opaque TLC-style model values such as
  `CONSTANTS Red = Red Blue = Blue`; the model checker compares them, checks
  finite model-value set membership/subset/equality, and uses them in
  assumptions and invariants.
- The model-check evaluator now executes Java syntax-corpus bitfield number
  literal formats (`\b`, `\B`, `\o`, `\O`, `\h`, and `\H`) as integer values.
- The SANY parser and semantic bridge now accept Java semantic-corpus tuple
  quantifier bounds such as `<<x, y>> \in S` in set comprehensions and
  function literals, preserving the tuple-bound names as local binders.
- The SANY semantic bridge now translates Cartesian product expressions
  (`\X`, Java `N_Times`) as ordinary binary expressions, unblocking examples
  that use product domains in function and set expressions.
- Named `ASSUME`, `ASSUMPTION`, and `AXIOM` module items now become referable
  zero-arity definitions while still contributing assumption expressions, so
  the Java semantic-corpus `AssumeTest.tla` checks cleanly.
- The embedded standard-module set now includes `Reals`, and semantic checking
  recognizes the built-in standard sets `BOOLEAN`, `Int`, and `Real`; the Java
  semantic-corpus `NumbersTest.tla`, `ExceptTest.tla`, and
  `RecordsWarning3Test.tla` check cleanly.
- The SANY parser now treats `PROPOSITION`, `LEMMA`, and `COROLLARY` as theorem
  body items at expression boundaries, and named theorem/proposition items are
  exported as referable zero-arity definitions; Java semantic-corpus
  `TheoremTest.tla` checks cleanly.
- Module body-boundary detection now recognizes operator-starting definitions
  such as `-. x == x`, so Java semantic-corpus negative operator definitions
  parse after assumptions without being swallowed into the previous expression.
- The SANY proof parser now handles corpus proof references with operator-call
  facts and `DEF` references, without mistaking proof-call syntax for operator
  definitions.
- The SANY proof parser now handles labeled nested `ASSUME ... PROVE` items,
  tuple-bound `TAKE`/`PICK` proof steps, proof-step module instance
  definitions, and nested implicit `QED BY` proofs without swallowing enclosing
  QED steps.
- The SANY semantic bridge now ignores proof syntax nodes when translating
  theorem and assumption bodies, while still checking the theorem expression.
- `tlago check` now accepts every module in the local Java SANY semantic corpus
  under `tlatools/org.lamport.tlatools/test/tla2sany/semantic/corpus`, excluding
  the helper module `Semantics.tla` that is loaded as a dependency.
- The SANY parser now accepts tuple binders in bounded and unbounded `CHOOSE`
  expressions and mixed bound groups such as `x, y \in S, <<z>> \in T`.
- The SANY parser now accepts higher-order operator references as expression
  values, including infix, prefix, postfix, junction, Unicode, parenthesized,
  and argument-bearing nonfix operator references.
- LET-local named module instances such as `LET M == INSTANCE N IN ...` now
  parse inside expression contexts, including conjunction and disjunction lists.
- LET definition RHS checking now scopes the definition's own name like SANY:
  the definition name is not visible in its own body, so bound expressions such
  as `LET x == CHOOSE x \in S : TRUE IN x` check cleanly; this also unblocks
  real examples such as `Chameneos.tla`.
- Labels now parse with optional argument lists and as embedded operands, so
  forms such as `Lbl(a,b)::P(a,b)` and `a + Lab:: b * c` match SANY syntax.
- A Go regression test now sweeps the accepted Java SANY syntax corpus under
  `tlatools/org.lamport.tlatools/test/tla2sany/corpus`, skipping Java's
  documented expected-failure scenarios.
- The semantic checker now rejects declared-name reuse by operator parameters
  and bound symbols, and reports zero-argument uses of operators that require
  parameters; the corresponding early Java semantic error-corpus cases are
  detected.
- The semantic checker now rejects user redefinitions of built-in symbols such
  as `TRUE`, while still allowing embedded standard modules to define their
  exported built-in names.
- The semantic checker now rejects fundamental level errors for primed primes
  and direct variable-level assumptions, covering Java semantic error-corpus
  cases `E4205_Test.tla` and `E4206_Test.tla`.
- INSTANCE checking now preserves duplicate `WITH` substitutions, requires
  constants and variables of imported modules to be explicitly or implicitly
  assigned, and rejects substitutions for non-parameter operators; this covers
  Java semantic error-corpus cases `E4240_Test.tla`, `E4241_Test.tla`, and
  `E4242_Illegal_Test.tla`.
- INSTANCE substitutions into modules that declare variables now enforce basic
  TLA+ level matching for constant and variable parameters, covering
  `E4245_Constant_Sub_Test.tla` and `E4245_Variable_Sub_Test.tla`.
- Module-level `USE`/`HIDE` blocks are now parsed as structured SANY nodes,
  including expression and `DEF` item lists with bare operator references, and
  `USE DEF` names participate in undefined-reference checking; this covers
  `E4200_Use_Test.tla`.
- EXTENDS/module import checking now rejects distinct imported modules with the
  same name and incompatible imported symbol kinds, including the Java
  `Naturals`/`_+_` conflict cases `E4223_Test.tla` and `E4224_Test.tla`.
- SANY warning diagnostics now preserve non-fatal severity and cover same-kind
  symbol ambiguity from `EXTENDS` imports (`W4800`), same-name ambiguity across
  plain `INSTANCE` imports (`W4801` for instance-vs-instance exports), and
  record constructor field names that clash with existing symbols (`W4802`).
- `Nat`, `Int`, and `Real` now live in embedded standard modules instead of
  the global built-in set; plain `INSTANCE` imports warn and skip unqualified
  exports that collide with local symbols or earlier instance exports, covering
  both Java `W4801` fixtures.
- PlusCal translation checksum validation now preserves module source text,
  computes the TLA+ translation CRC, recognizes the semantic error-corpus
  PlusCal algorithm checksum, and reports stale translation warnings `W4803`
  through `W4805`.
- Function definitions now contribute bracket-application arity metadata, and
  semantic checking rejects statically visible wrong-arity applications while
  allowing the single tuple-argument form for multi-parameter functions; this
  covers `E4260_1_Expected_2_Provided_Test.tla` and
  `E4260_2_Expected_3_Provided_Test.tla`.
- Record constructor checking now rejects duplicate field names, covering
  `E4262_Test.tla`.
- Higher-order operator parameters now preserve expected operator arities, and
  call checking rejects expression arguments where operators are required,
  wrong-arity operator and LAMBDA arguments, and LAMBDA arguments where an
  expression parameter is expected; this covers `E4270_Test.tla`, the
  arity-focused `E4271_*_Test.tla` files, `E4274_*_Test.tla`, and
  `E4275_Test.tla`.
- Recursive operator definitions now reject prime operators in their bodies,
  covering `E4290_Test.tla`.
- Recursive definition sections now preserve LET-level `RECURSIVE`
  declarations, require same-level definitions, and reject declaration,
  theorem/proof, module, assume, and use/hide items between top-level
  recursive declarations and their definitions, covering `E4293` and `E4294`.
- Temporal/action level-composition checking now rejects bare temporal
  operators over action-level formulas, action operands to leads-to, mixed
  action/temporal boolean operands, temporal quantifier bounds, and action
  bounds for temporal quantifier bodies, covering `E4310` through `E4315`.
- Label semantic checking now preserves label parameters and rejects repeated,
  missing, and unnecessary label parameters, labels outside definitions/proof
  steps, labels in EXCEPT updates, and ambiguous duplicate sibling labels,
  covering `E4330` through `E4337`.
- Proof-step semantic checking now preserves compact proof metadata and rejects
  named implicit steps, non-expression proof steps used as expressions,
  non-constant HAVE/TAKE/WITNESS/CASE material under temporal proof goals,
  non-constant bounds on temporal PICK formulas, ASSUME/PROVE definitions used
  as ordinary expressions, and invalid HIDE facts, covering `E4350` through
  `E4357`.
- The proof parser now tracks Java SANY-style proof levels, so higher-level
  child proofs, terminal `BY`/`OBVIOUS` proofs, and QED-attached child proofs
  parse without swallowing sibling steps.
- Proof-step summaries now preserve nesting depth, and temporal proof-step
  restrictions apply to the theorem proof's own top-level steps rather than
  nested subproof obligations.
- The embedded standard modules now include a compact `TLAPS` proof-helper
  shell with common backend pragma names such as `SMT`, `SMTT`, and `PTL`;
  the Bakery example checks cleanly through `tlago check`.
- The embedded standard modules now include compact SANY-parsed shells for
  common community helper modules: `Functions`, `FunctionTheorems`,
  `FiniteSetTheorems`, `SequenceTheorems`, `SequencesExt`,
  `SequencesExtTheorems`, `FiniteSetsExt`, `FiniteSetsExtTheorems`,
  `NaturalsInduction`, `WellFoundedInduction`, `GraphTheorems`,
  `DyadicRationals`, and `Apalache`. The `Functions` and sequence shells use
  real operator-parameter syntax such as `op(_, _)`, so higher-order arity
  checks continue to flow through the production SANY parser path.
- The embedded `TLC` module now exports common runtime helper operators such
  as `TLCSet`, `Permutations`, `RandomElement`, `ToString`, `TLCEval`, and
  `JavaTime`, resolving model-wrapper symmetry definitions and TLC register
  idioms through the normal checker.
- Semantic import environments now include unqualified exports inherited
  through transitive `EXTENDS` chains, including standard-module symbols such
  as `Nat` from a spec module extended by a model wrapper; the Bakery and
  Boulanger model wrappers check cleanly.
- The embedded `FiniteSets` module now exports `IsFiniteSet`, and the
  explicit-state evaluator treats supported finite set values as finite;
  examples such as `Disruptor_SPMC.tla`, `CRDT.tla`, and `Reachable.tla`
  check past their finite-set assumptions.
- Square-bracket lookahead now treats action-subscript `]_` as a hard
  boundary, so temporal action forms such as `[][Next]_vars` do not scan into
  later function constructors while deciding whether `[` starts `[i \in S |-> e]`.
- Expression parsing now allows a junction-list bullet that would otherwise be
  an expression stop token to serve as the missing right operand when the
  operator stack is waiting after an infix operator, covering SANY forms like
  `A =>` followed by an indented `/\ ...` list.
- Proof `USE`/`HIDE` parsing now accepts inline proof-step references such as
  `USE <2>1 DEF Foo` without mistaking the reference for a new nested proof
  step; examples such as `CoffeeCan_proof.tla` check past this proof form.
- The SANY semantic bridge now translates generic prefix/infix/postfix
  operator-reference nodes as higher-order operator values or direct calls,
  with built-in operator symbols participating in undefined-name and arity
  checks; examples using `FoldFunction(+, ...)` and `ReduceSet(\intersect, ...)`
  now advance to library-resolution work.
- Function constructors whose domains contain set comprehensions and indexing,
  including record field values such as `[p \in {q \in S : f[q] # Zero} |-> e]`,
  parse and check far enough to expose semantic gaps instead of syntax errors.
- Root modules tolerate trailing tool commands or prose after the closing
  `====` footer while still preserving lexer diagnostics inside the module.
- `STRING` is recognized as a Java SANY built-in standard set, alongside
  `BOOLEAN`, while `Int` and `Real` continue to come from standard modules;
  the example `LevelSpec.tla` checks cleanly.
- Embedded community helper shells now export real `Functions!Range` and
  `FiniteSetsExt!Min`/`Max` definitions, unblocking examples such as
  `DieHardest.tla` that extend those modules.
- `FiniteSetsExt` now re-exports the `Functions` helper surface transitively,
  so examples such as `MCReplicatedLog.tla` can use `Range` by extending only
  `FiniteSetsExt`, as in the community module.
- Embedded `SequencesExt` now exports real sequence helper definitions for
  `RemoveAt`, `Front`, `IsPrefix`, `IsStrictPrefix`, `Prefixes`,
  `CommonPrefixes`, and `LongestCommonPrefix`.
- Bounded quantifier bodies inside prefix junction-list operands now retain
  same-line infix conjunction/disjunction in the quantified body, so fairness
  forms such as `\A s \in Servers : WF_vars(Extend(s)) /\ WF_vars(Copy(s))`
  check correctly.
- Named higher-order operator arguments now use the callee's operator-parameter
  signature during arity checks, so calls such as `Apply(Add, 1, 2)` and
  Huang's `FoldFunction(Add, Zero, ...)` check without treating `Add` as an
  erroneous zero-argument call.
- The `DyadicRationals` embedded module now exports `IsDyadicRational` using
  the community module predicate shape, so `Huang.tla` checks cleanly apart
  from existing record-field warnings.
- `tlago sany-xml` now emits Java-shaped SANY XML from the production SANY path.
  The gated Java oracle test uses the frozen `tla2tools.jar` with
  `XMLExporter -o`, captures XML from stdout, and compares canonicalized XML
  without sockets or schema-proxy machinery. The initial `Simple.tla` and
  `Exprs.tla` fixtures pass against Java SANY.

## Current Corpus Gaps

- The latest confirmed compact sweep before ApalacheIR work was `99/120`.
  `Huang.tla` has since been checked directly and now succeeds with only
  `W4802` record-field warnings; a new sweep has intentionally not been run
  yet.
- `SimKnuthYao.tla` now parses past its post-footer shell command; remaining
  failures are missing bundled extension modules such as `CSV`, `TLCExt`,
  `IOUtils`, and `Statistics`.
- Some proof and library-heavy specs still need additional bundled helper
  definitions before meaningful semantic parity tests can pass.
- A few expression forms still need SANY parser parity work, including observed
  adjacent-expression failures in level-checking and reachability corpus specs.
- ApalacheIR parity now has an executable Apalache oracle harness and passing
  structured-expression, temporal/action, unbounded-quantifier, recursive,
  named-assumption, function-definition, source-range, core-builtin, and
  `INSTANCE ... WITH` substitution slices. Instance materialization preserves
  Apalache's source behavior: instantiated declarations point at the instance
  statement, while replacement expression spans stay anchored to the formal use
  sites in the instantiated module. It still needs theorem handling beyond
  Apalache's parse-output omissions, type-tag handling, and broader
  library/advanced-operator coverage; this work is paused while direct SANY XML
  parity is established.

## Next Slice

- For the active TLC goal, continue from `tlc/HANDOFF.md`:
  1. Read the top and tail of `tlc/PORT_PROGRESS.md`.
  2. Pick a Java TLC class or cluster not marked recently audited.
  3. Compare Java control flow, mutation order, error precedence, and special
     cases to the Go files with the same responsibility.
  4. Patch real behavioral gaps only.
  5. Update `tlc/PORT_PROGRESS.md`.
  6. Run `go test ./tlc`, then usually `go test ./...`, and commit.
- The older SANY next slice remains paused until requested: direct SANY XML
  parity over `test_vectors/tla-plus-bench/specs/` and
  `test_vectors/Examples/specifications/`, Java semantic/error corpus
  diagnostics, PlusCal checksum validation, and ApalacheIR parity.
