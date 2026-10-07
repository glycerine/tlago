# TLC Port Handoff

Updated: October 7, 2026. Full-workspace verification baseline: `44aaf11`.

This is the current restart guide for the Go TLC port. Detailed audit history
and run receipts belong in [PORT_PROGRESS.md](PORT_PROGRESS.md). Older versions
of this handoff remain in Git history; their pending jobs and next steps must
not be mistaken for current work.

## Goal and scope

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

Current priority: finish faithful Java TLC parity on `master`. The new reusable
rpc25519/Tube distributed service is postponed until the remaining port is
complete. Do not resume its unfinished code or its service BDD work here.

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
Do not resume SANY XML or ApalacheIR corpus sweeps without a user request.

## Current verified state

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

Latest focused verification:

| Scope | Result | Receipt |
| --- | --- | --- |
| Earlier parser/semantic snapshot, original ParseErrorTests, EWD998ChanDebugger, three original fairness/liveness models, EmptyExistentialQuantifier, RandomSubsetSetOfFcns and GetScopedIdentifiers | Pass | Root 17.055 seconds, session `32902` retired |
| Current focused parser/context/bridge, original ParseErrorTests, six original proof/selector models and scoped identifier/reference checks | Pass | Root 5.509 seconds, session `55110` retired |
| Complete existing SANY package; canonical corpus AST assertions remain pending | Pass | 1.730 seconds, session `71430` retired |
| Current TLC tools, spec processing, contexts, semantic table and coverage | Pass | 0.022 seconds, session `27147` retired |
| Original TLCGetAll, ACoverage and both simulation constraint models | Pass | 2.160 seconds, session `96274` retired |
| Earlier focused TLC function context, EXCEPT/record coverage and original function-value tests | Pass | 9.705 seconds, session `98225` retired |
| All-package compilation | Pass | Final sources compile; no additional long workloads |
| Existing bounded root corpus and parser-reference checks | Pass | Earlier `44aaf11` grammar snapshot |
| Existing native exporter behavior class with Java-valid source and unordered IR | Pass | Earlier 0.035 seconds, session `62377` retired |

Earlier focused receipts belong in `PORT_PROGRESS.md`. These checks retain their
recorded scope; passing translated tests does not establish whole-method fidelity
where reconciliation gaps are documented below.

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
Java. The Java `belchDEF` token-stream operation and its production call sites
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

General JavaCC lookahead-derived expected-token sequences, remaining production
states and label error continuations remain work. The original `ParseErrorTests.testAll` lives in root
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
mechanical generator retains its 159-method dependency closure and predicates
against the scanner position, junction context and active operator stack. Saved
calls expire in source order and rescan actual error token sequences, including
earlier successful calls. All 599 complete TRACE/results and 137 selected raw
module trees match Java. The temporary bounded vocabulary matrix also matches
all 609,175 scanner verdicts. This covers the six connected entry points, not
all 74 JavaCC previews. Remaining entry points, grammar expectations, general
rescan integration and proof generation remain reconciliation work.

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
now port complete. The LET and transitive-import methods remain reconcile.
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
Java observations agree. Application checking is still a dependency: temporal
application cases use explicitly prechecked source nodes in the observer,
not an original-test completion receipt.
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
ordinary application checking still prevents closing the original LET methods.
Application and evaluator collection sharing
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
parity. The existing syntax corpus port checks parser status and node usage, without the original canonical AST
comparison. These are remaining translation gaps, not full-suite fidelity
receipts. Preserve the source harness's known-failure inversion: Java accepts
the unchanged LOCAL-in-LET error fixture, and the original expects that success.

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

Distributed integration remains deferred. The selected future transport is
`github.com/glycerine/rpc25519` with Greenpack serialization, using asynchronous
peer/circuit/fragment APIs. Preserve TLC deduplication, recovery, and termination
semantics when that work resumes; see `TLC_ARCH.md`.

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
