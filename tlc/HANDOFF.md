# TLC Port Handoff

Updated: October 6, 2026. Full-workspace verification baseline: `44aaf11`.

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
| Existing parser/semantic classes, original ParseErrorTests, EWD998ChanDebugger, three original fairness/liveness models, EmptyExistentialQuantifier, RandomSubsetSetOfFcns and GetScopedIdentifiers | Pass | Root 17.055 seconds, session `32902` retired |
| Complete existing SANY package with original corpus assertions | Pass | 1.477 seconds, session `73231` retired |
| Existing focused TLC function context, EXCEPT/record coverage and original function-value tests | Pass | 9.705 seconds, session `98225` retired |
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

Current MSB session `5144` is live, using the isolated binary
`/mnt/oldrog/tmp/tlago-heap-fp-stress.test`, compiled from `132a77f` production
and the installed test translation. Log:
`/mnt/oldrog/tmp/tlago-heap-random-msb-current-full.log`. It started at
18:17:43 CDT on October 6 and last reported 1,386,019,479 of 2,147,483,648
insertions. Preserve this run and poll the same handle; do not restart it or
claim a full pass before terminal completion. Its temporary files use the large
workspace volume. Both full runs use `-timeout=0` and no race instrumentation.

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
General JavaCC rescan, the theorem's three-token Assume-Prove selection,
`UseOrHideOrBy`, TAKE/PICK/WITNESS grammar and remaining proof productions are
reconciliation work; this is not complete Proof grammar parity.

Function and set-comprehension bridge nodes retain one group per syntactic
bound, including multi-name lists and distinct adjacent tuple bounds. Each domain
is converted once before the formals enter the context, matching Java generation.

Node constructors now use the actual 446-entry source `SyntaxNodeImage` table,
which differs from node-kind constant names. Unknown selector syntax retains its
source image and zero-valued slot, then reports constructor errors before name
resolution. Reconstructed tuple membership retains both structural diagnostics.
General selector and constructor fidelity remains reconciliation work.

Current bounded observations match Java: 365 complete parser TRACE/results, 41
output routing/format cases, 12 LAMBDA semantic cases, 13 CHOOSE semantic cases,
31 selected declaration/LHS trees, 54 substitution target/arrow trees, 22 quantified
semantic observations, three CHOOSE runtime probes and ten quantified metadata/
runtime probes, plus 93 selected expression trees including ranges, four function-application
runtime probes, ten bracket constructor/group metadata and runtime probes,
16 brace semantic observations, nine brace metadata/runtime probes, all 446
node-image entries, and 16 selected complete definition trees with kinds, images
and ranges, plus 25 selected complete module trees with kinds, images, ranges, original images
and proof levels.
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
