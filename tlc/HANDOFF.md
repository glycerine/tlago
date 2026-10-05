# TLC Port Handoff

2026-10-05 verified original Test210 diagnostic checkpoint afteread7f9a:
Original Test210.testSpec is mechanically translated with ERROR_SPEC_PARSE,
no GENERAL, all ten exact TestPrintStream substrings (header/count9 plus nine
full location/message blocks), original SuiteETestCase flags and two byte-identical
vectors. Initial Go output had two errors instead of nine. Selector resolution
used ASSUME/PROVE's quantified expression projection and did not walk proof steps;
it lost NEW clause indices/domains, lexical declaration scope and SUFFICES handling.
Proof steps now retain their qualified names, ASSUME/PROVE bodies and SUFFICES
status. Resolve actual clause structure and nested NEW domains, track the enclosing
proof versus subsequent statements, follow Java's illegalLabelRef/illegalAPPosRef
and SUFFICES positional wrapper rules, and retain individual selector ranges.
Named assumption labels are legal. Nested proof ASSUME/PROVE labels are validated
with Java's full range/message. BY facts retain fact/expression mode, including
leaf-only proofs without numbered steps; NEW fact selections retain their source
NewSymbol rather than evaluating a projected domain. This is a selector/diagnostic
port, not a declaration of complete native proof-graph or all Generator parity.

Whole original Test210 first passes normal0.040s; focused race1.159s (38423
terminal/retired) before the final explicit domain/leaf traversal additions.
Related original legacy contexts (201–210/216/219/220/99/999/InvalidInvariant and
ETest1–16) pass race98.963s (73617 terminal/retired) before the final leaf-only
proof retention. Root parser/semantic/XML/elevated-warning race14.474s (43387
terminal/retired), SANY Java race8.226s (47389 terminal/retired).
Final leaf-retention code: complete SANY Java race7.901s (85424 terminal/retired),
log legacy-error-test210-final-sany-race.log. Final combined related root/legacy
race gate passes114.107s (42109 terminal/retired), log
legacy-error-test210-final-related-race.log. Keyword directives now use token
kinds (BY/DF/SUFFICES) rather than images, preserving keyword record fields.
That refinement passes root/Test208/Test210 race21.221s and full SANY Java
race8.109s (83989/68691 terminal/retired). Final
ASSUME/PROVE selections retain structured targets in fact mode; Java rejects a
whole AP as an expression rather than evaluating its quantified projection.
Final guard code passes related root/Test208/Test210 race20.627s (30971
terminal/retired) and full SANY Java race8.121s (90316 terminal/retired), logs
legacy-error-test210-ap-final-race.log and legacy-error-test210-ap-sany-race.log.
All logs under .codex-gotmp/. Every original assertion remains unchanged.
Inventory1082/1269 contexts (85.3%),562/626 complete classes (89.8%),187pending
across64classes; legacy99/104 complete. Next original Test212, then213/214/215/217.
Test212 requires six exact non-Leibniz substitution errors at whole INSTANCE
ranges and its two dependency modules; fix real Leibniz analysis before advancing.

Corrected INSTANCE full workspace57071 passes exit0/terminal, root1772.759s,
SANY0.932s/TLC72.721s, log legacy-error-context-final-workspace-go.log. This covers
caf0ad3 production code and ETest1–3, before ETest5/16 semantic report/validation
changes and before Test210. Full diagnostic snapshot23927 finished exit1/retired,
root1791.100s/SANY1.083s/TLC70.782s; it compiled BEFORE caf0ad3 and failed only the
three now-fixed INSTANCE-order cases. Do not credit that older snapshot as green.
Current ETest16 whole workspace75902 remains live, log
legacy-error-etest16-final-workspace-go.log. It includes7f51f9c and ETest1–16,
but not the current Test210 semantic changes. Resume this exact handle, never
restart on an observation timeout, and preserve its eventual receipt's scope.
Final current-production whole workspace launched after commit2303cdb: live15753,
log legacy-error-test210-final-workspace-go.log under .codex-gotmp/. It includes
all current Test210 selector/proof semantic changes and complete original ports.
Resume this exact handle. Do not claim its full-suite result before terminal success.
Preserve the five deferred topics and test_vectors fixture naming.
This entry is authoritative; older counts/gate/next-context notes are historical.

2026-10-05 ETest8–16 complete after6e5d315:
All16 original SuiteETestCase contexts are now complete translations. This batch
preserves the nine original constructors, all inherited runner flags, every source
assertion and18 byte-identical vectors. ETest8 uses -simulate, ERROR (the upstream
simulator TODO), checkDeadLock=false, FINISHED/STATS_SIMU, no DEADLOCK and its
exact zero-count uncovered location. ETest9–11 retain FAILURE_SPEC_EVAL, first
GENERAL record membership-error substrings and exact zero coverage. ETest12–15
retain VIOLATION_ASSUMPTION and exact first assumption-evaluation error substrings
for Cardinality overflow, unsuccessful CHOOSE and invalid integer comparison.
ETest16 initially reported one semantic error instead of Java's two: duplicate
fields in sets of records were unchecked. Generator.processRcdForms:4094 handles
both record forms. Go now validates both, reports the repeated field token range
and Java's Non-unique fields message via the existing SANY diagnostic metadata.
The source assertions and vectors are unchanged. No invented tests were added.

Final all ETest1–16 pass race40.497s (82330 terminal/retired), log
legacy-error-etest1-etest16-race.log. Individual8 race1.808s,9–11 race14.175s,
12–15 race18.217s,16 race1.132s, handles97738/43233/80532/92787 terminal/retired.
Related root parser/semantic/XML/elevated-warning race14.895s (53867 terminal/
retired), log legacy-error-record-constructors-related-race.log. Complete SANY
Java race8.013s (15921 terminal/retired), log
legacy-error-record-constructors-sany-race.log.
The three previously failing workspace contexts plus ETest1–3 also pass
race720.306s (15192 terminal/retired), log workspace-instance-context-fix-race.log.
This validates caf0ad3's actual INSTANCE filtering correction with unchanged
EvalExceptionLiveness, EWD998ChanDebugger and BufferedRandomAccessFile assertions.
Inventory1081/1269 contexts (85.2%),561/626 complete classes (89.6%),188pending
across65classes; legacy98/104 complete. Next original Test210, then212/213/214/
215/217. Test210 requires exactly nine semantic errors, exact ranges and full
ASSUME/PROVE scope, nested-label and SUFFICES diagnostics. Fix the semantic
implementation before proceeding after any failure; no generic parse-failure
substitute. The five deferred topics remain deferred.

Only two earlier whole-workspace handles remain confirmed live:23927 (diagnostic
snapshot BEFORE caf0ad3's INSTANCE context filter fix) and57071 (caf0ad3 production
code BEFORE f8bd7c9's unknown-operator metadata and the current duplicate-record-set
validation). Resume these handles; do not restart on observation timeouts. Logs
legacy-error-reporting-final-workspace-go.log and
legacy-error-context-final-workspace-go.log under .codex-gotmp/.
Final current-source full workspace launched after commit7f51f9c: live75902,
log legacy-error-etest16-final-workspace-go.log under .codex-gotmp/. It includes
the duplicate-record-set validation and all16 ETests. Resume this exact handle
and preserve earlier snapshot receipts with their accurate scope.
Do not claim full current-source green from focused evidence alone.
This entry is authoritative; earlier inventory and live-handle notes are historical.

2026-10-05 original ETest6 and ETest7 complete after f8bd7c9:
Original ETest6 passes FAILURE_SPEC_EVAL, the first GENERAL record's exact
undefined-identifier substring with line16:23 location, and both exact zero-count
uncovered locations. Its first-record substring search follows TestMPRecorder.
Original ETest7 passes ERROR_CONFIG_PARSE and exact
TLC_CONFIG_SUBSTITUTION_NON_CONSTANT parameters C/Foo. Preserve full inherited
SuiteETestCase runner flags and original constructors. All four vectors are
source-identical; neither test required an implementation change.
ETest6 race5.864s (80093 terminal/retired), ETest7 race1.764s (31611 terminal/
retired); logs legacy-error-etest6-race.log and legacy-error-etest7-race.log.
Inventory1072/1269 contexts (84.5%),552/626 complete classes (88.2%),197pending
across74classes; legacy89/104 complete. Next original context ETest8. Authoritative
source overrides earlier informal notes: its constructor uses -simulate and ERROR
(the upstream TODO notes the simulator's exit status), checkDeadLock=false;
assert FINISHED/STATS_SIMU, no DEADLOCK and exact uncovered18:15–18:22.
No substitute model, status correction or extra deadlock checking is allowed.

Last confirmed live handles:15192 (three corrected workspace failures + ETest1–3
race),23927 (older diagnostic full workspace BEFORE context-filter fix),57071
(final caf0ad3 production full workspace). Resume these, do not restart on an
observation timeout. Logs workspace-instance-context-fix-race.log,
legacy-error-reporting-final-workspace-go.log and
legacy-error-context-final-workspace-go.log under .codex-gotmp/.
The corrected three-failure normal run passes69.425s; related original legacy
race64.544s; ETest1–5 and root/SANY diagnostic race gates pass as recorded below.
Full workspace verification remains pending; no full-current green claim.
The earlier bc01728 workspace83892 failed three cases, fixed in caf0ad3, and is
terminal/retired. No changes to the five deferred topics' scope or fixture naming.
This is the authoritative next-context checkpoint; older entries are historical.

2026-10-05 original ETest5 complete after ce93212:
Original ETest5.testSpec first rejected M!Init correctly but printed a point range
and Go's generic undefined-identifier text. Generator.selectorToNode:810 reports
the complete unresolved compound-name range and Unknown operator message.
The semantic diagnostic now carries those SANY details separately, excluding
supplied argument ranges. Generic diagnostic positions and text remain available
to other frontends. No fixture-dependent reporting or assertion weakening.
ETest5 preserves ERROR_SPEC_PARSE, no GENERAL, one error, range13:15–13:20 and
exact Unknown operator text. Both fixtures match original source byte for byte.
ETest1–5 pass race11.578s (56875 terminal/retired), related root parser/semantic/
XML/elevated-warning race14.724s (89702 terminal/retired), complete SANY Java
race8.418s (75558 terminal/retired). Logs: legacy-error-etest1-etest5-race.log,
legacy-error-unknown-operator-related-race.log,
legacy-error-unknown-operator-sany-race.log under .codex-gotmp/.
Inventory1070/1269 contexts (84.3%),550/626 complete classes (87.9%),199pending
across76classes; legacy87/104 complete. Next original context ETest6, preserve its
undefined primed-variable GENERAL substring and both exact uncovered locations.
Workspace57071 is still the caf0ad3 production gate; the later ETest5 change adds
unknown-identifier SANY report metadata, covered by the focused root/SANY gates.
Older23927 compiled before the context filter correction. The three original
workspace failures have a normal PASS69.425s; their additional race gate15192
remains live. Resume existing handles; do not restart on observation timeouts.
Do not claim a full current-source green workspace while verification is pending.

2026-10-05 original ETest4 complete after caf0ad3:
Whole original ETest4.testSpec passes with FAILURE_SPEC_EVAL, FINISHED, stats0/0/0,
the exact four-frame TLC_NESTED_EXPRESSION string including its final blank line,
and the exact zero-count uncovered location. Both vectors are source-identical.
The initial port passed a literal ': 0' suffix to requireJavaTLCUncovered, whose
API already filters zero-count records and compares locations. Corrected that
mechanical translation without changing coverage semantics or implementation.
ETest1–4 pass race12.043s (82804 terminal/retired), log
legacy-error-etest1-etest4-race.log. Inventory1069/1269 contexts (84.2%),549/626
complete classes (87.7%),200pending/77classes; legacy86/104 complete.
Next context ETest5: preserve one semantic error, range13:15–13:20 and exact
Unknown operator M!Init diagnostic; fix actual semantic behavior if it fails.
Workspace57071 remains the final caf0ad3 production gate; it compiled before
ETest4 was added, so combine its eventual receipt with this focused race evidence.
Older diagnostic gate23927 compiled before caf0ad3's context filter correction.
Do not claim a full green workspace while its authoritative final gate is live.
The five deferred topics remain deferred. All process/log details below retain
snapshot scope; logs are under .codex-gotmp/.

2026-10-05 ETest1–3 and semantic diagnostic checkpoint (workspace verification pending):
Original ETest1, ETest2 and ETest3 are complete translations. All six fixture files
match the original Java vectors byte for byte. Preserve the original constructors,
SuiteETestCase flags and every assertion: exact semantic error count, argument-list
ranges and arity text for ETest1/2; deadlock exit, FINISHED, stats2/2/0, no GENERAL
and zero uncovered for ETest3. Inventory1068/1269 contexts (84.2%),548/626 complete
classes (87.5%),201pending/78classes; legacy85/104 complete.
Generator.java:900 reports incorrect arity at the final supplied argument-list
range, with the remaining signature. Diagnostics now carry optional SANY range
and text separately from their existing generic formatting. The shared TLC loader
prints actual semantic errors with Java's header/count and ErrorDetails shape;
warning elevation affects exit status without inflating the actual error count.
Existing warning-control, elevated-warning and full SANY Java race checks pass.
ETest1/2 race pass1.277s; ETest3 race pass5.849s. Related root parser/semantic/XML
and installed legacy contexts pass race60.790s before the context correction below.
No invented tests or weakened original assertions.

Full workspace83892 finished exit1, root1798.112s/SANY1.160s/TLC72.815s.
Its bc01728 snapshot failed original EvalExceptionLiveness trace variable order,
EWD998ChanDebugger variable selection, and BufferedRandomAccessFile postcondition.
This is authoritative red evidence; do not describe bc01728 as full-suite green.
The theorem-import filter admitted global builtins and ModuleNodes as theorem
entries because their declaration kind is empty. Filter those out and import
operator definitions before theorem/assumption definitions as Java Generator does.
The three unchanged failing contexts now pass69.425s (98866 terminal/retired),
log workspace-instance-context-fix-focused.log. Legacy contexts pass race64.544s
(71457 terminal/retired), log legacy-error-context-fix-race.log, including
ETest1–3 and Test201–209,216/219/220/99/999/InvalidInvariant. Current guard/final
code race recheck of the three failures plus ETest1–3 is live15192, log
workspace-instance-context-fix-race.log. Final corrected whole workspace is
live57071, log legacy-error-context-final-workspace-go.log.
Older workspace77547 completed exit0, root1806.145s/SANY1.001s/TLC70.894s,
log legacy-instance-exports-workspace-go.log; compiled before the theorem fix.
Current diagnostic workspace23927 remains live, log
legacy-error-reporting-final-workspace-go.log, but compiled BEFORE the context
filter correction above. Resume this handle; do not restart it on an observation
timeout and do not credit it as a gate for the later correction. Logs are under
.codex-gotmp/. After resolving these failures and recording current verification,
commit this batch and continue with original ETest4. Its exact nested-expression
stack and uncovered location must be preserved. The five deferred topics remain
deferred; existing tests in them must remain green. Never use testdata fixtures.
This entry supersedes older gate and next-context notes below.

2026-10-04 latest verified implementation checkpoint after63c8bc4:
Seven further original contexts are complete: Test209, Test219, Test99, Test999,
Test216, Test220 and TestInvalidInvariant. Preserve every original constructor,
flag, inherited/custom assertion, and all19 byte-identical fixture files. Test216
includes all three original dependency modules (216a/b/c); the earlier ignored
fixture draft omitted them and has been corrected before installation.
Test219 first failed semantic operator-argument checks: instance formal parameters
were counted in arity but missing from argument-position metadata. Prefix the
instance formal specifications as Java Generator.java:4880 does. Its next failure
was missing imported Inst!Foo: instantiate the whole source semantic context,
including nested imports, retaining SubstIn wrappers and source-node reuse.
Test216 then exposed omitted imported named theorems. Java imports those separately
as ThmOrAssumpDefNodes (Generator.java:4940 onward), using APSubstIn wrappers.
The bridge now preserves those context entries, actual theorem symbols, parameters,
locality, defining module and original source identity. Ordinary theorem references
resolve to their theorem symbols; no constant replacements or weakened tests.
Whole Test216 now passes28/7/0, INIT_GENERATED2 21/s/7, and the exact original
uncovered location. Test220 retains checkDeadLock=true so its config decides.
TestInvalidInvariant retains FAILURE_SAFETY_EVAL, the invariant-level diagnostic
and both GENERAL assertions. Whole99/999 normal pass1.097s/1.130s; race9.185s
(84526 terminal/retired). Whole216/220/InvalidInvariant normal pass1.374s/1.622s/
1.121s (71276/72356/71961 terminal/retired). Final focused root parser/semantic/XML,
all installed201–209/216/219/220/99/999/InvalidInvariant, cyclic INSTANCE,
native override and four bridge cases pass race82.897s (40154 terminal/retired).
Existing full SANY Java tests pass race7.978s (14888 terminal/retired), covering
the new instance-parameter semantic metadata; no later SANY semantic changes.
Inventory1065/1269 contexts (83.9%),545/626 complete classes (87.1%),204pending
across81 classes; legacy82/104 complete. Next original context: ETest1. Remaining
legacy contexts: ETest1–16, Test210/212/213/214/215/217. Preserve exact source
error counts, location ranges, diagnostic substrings, statuses and coverage.
Several ETests are runtime/config/simulation errors, not all SANY parse failures.
Do not replace their assertions with a generic failure or add invented tests.

Completed gates from the preceding snapshot: workspace44152 exits0/retired,
root1793.053s/SANY0.923s/TLC70.700s, log legacy-proof-fields-workspace-go.log.
It compiled the207/208 production code committed in63c8bc4, before the verbosity
harness and later INSTANCE/theorem changes. Whole CommunityModules verbose run
73656 exits0/retired, both unchanged phases pass306.247s, log
community-progress-verbose-go.log. That covers the progress-output change840d494
and preceding207/208 code, before the later INSTANCE/theorem implementation.
Live output, setup/loading messages and10s phase/elapsed/PID heartbeats confirmed.

Whole current TLC package race42143 passes646.992s (terminal/retired), log
legacy-theorem-imports-tlc-race.log.
Required workspace gates still active; resume existing handles, never restart:
-83892: final current-source whole workspace, legacy-theorem-imports-final-workspace-go.log.
-77547: older INSTANCE-export workspace, legacy-instance-exports-workspace-go.log;
 compiled before the theorem fix and later test files, so historical scope only.
All logs are under .codex-gotmp/. Do not claim full current-source green until
83892 reports terminal success. All focused and TLC package race checks are
green. Record
terminal gate receipts and fix any actual failure before proceeding to ETest1.
Preserve the five deferred topics and never create a testdata fixture directory.
The snapshots below are historical; this entry is authoritative.


2026-10-04 Test207/Test208 checkpoint (focused checks pass; full gates still running):
Original Test207 and Test208 are now complete translations, with all inherited
assertions, original constructor/settings, and four byte-identical source vectors.
Test207 exposed incorrect NEW declaration treatment in label validation: validate
preserved ASSUME/PROVE clauses, distinguish NEW OpDeclNodes from quantified formal
parameters, and track nested declaration scope in source order through PROVE.
Test208 exposed missing Java isFieldNameToken handling; reserved keyword record
components and EXCEPT paths now use Java's exact token ranges. Module identifiers
and the original initial keyword record-constructor branch use the same helper.
Neither original test nor fixture was weakened. Original207 passes2.032s;
original208 passes1.164s. Related parser/semantic/XML and original201–208 race
checks pass33.572s (65197 terminal/retired); earlier complete SANY Java race checks
pass8.386s. Inventory1058/1269 contexts,538/626 classes,211pending/88classes;
legacy75/104. Whole original Test209 passes1.205s in ignored overlay35080
(terminal/retired), not installed/credited. Whole original Test219 fails in
ignored overlay43516 (exit1/retired): semantic INSTANCE export metadata adds
parameterized instance arity but omits those parameter positions from operator
argument specifications. Fix the actual metadata before any later context.
Logs: .codex-gotmp/legacy-preflight209-go.log and legacy-preflight219-go.log.

CommunityModules progress-output change committed separately as840d494.
User's CommunityModules steering: keep the whole original test, but make go test
-v visibly report progress. community_modules_java_test.go now streams the child
stdout/stderr while preserving complete failure output, reports setup/dependency
and module-loading phases, and emits phase/elapsed-time/PID heartbeats every10s.
No original assertions, execution phases or production IOExec behavior changed.
Live verbose run73656 in .codex-gotmp/community-progress-verbose-go.log shows
module parsing and repeated10s heartbeats; this is evidence of progress output,
not a completed correctness gate. Run is still active: resume with write_stdin.
Whole workspace44152 in .codex-gotmp/legacy-proof-fields-workspace-go.log is
also still active; it compiled before the verbosity harness change, so its gate
covers the207/208 parser/semantic fixes. Do not restart or call either run green
until its authoritative handle/log reports terminal success. Finalize gate
receipts after these runs finish. This checkpoint records only focused green
checks; full-workspace success remains pending. Resume the Test219 metadata fix,
then install/verify/credit original209/219 once complete.
Preserve the five deferred topics and never create a testdata fixture directory.


This handoff is for the next Codex/model after the system upgrade. It records
the active goal, the current working rules, what is already done, what to avoid,
and the best next steps for continuing the Go port of Java TLC.

## Current Snapshot

Latest green code commit: `b0ee5ae` (selector resolution, original Test206,
and warning-flag isolation for original model tests). Working tree was clean
after this commit. Earlier selector foundation: `fa672eb`.
All required checks pass. Final current-source workspace2130 passes root1777.556s,
SANY1.121s and TLC70.088s (terminal/retired). Code/tests committed as `b0ee5ae`.

Final gate correction: the first workspace run84712 failed after1460.845s
at original Test63's exact coverage assertion (terminal/retired); SANY1.100s
and TLC68.170s pass for that snapshot. Do not label the root snapshot green.
Reproduced unchanged Github680a/b/c then Test63:39115 exit1/retired, while12
isolated Test63 repetitions pass9.214s (28911 retired). Cause: the Go model-test
runner omitted reset/restore of TLCGlobals.warn, so earlier original -nowarning
flags changed Tool.getNextStates0's allAssigned path and coverage. Java Tool
line1006 uses the same flag; Ant's original runner forks perTest JVMs. Restore
Warn=true before each original class's flags, then restore prior Warn on cleanup.
No evaluator/reporting/assertion changes; all original -nowarning flags retained.
Three full Github680->Test63 race repetitions pass48.668s (85432 retired).
Final module/LET binding review keeps shared cached bindings for canonical
module INSTANCEs only; independently lowered LET contexts retain their own
bindings so captured outer formals are not merged by source position. Existing
parser/semantic/XML, original Test201–206, cyclic instance and four native-module
bridge checks pass race32.615s (51123 retired). Final current-source workspace
passes root1777.556s/SANY1.121s/TLC70.088s (2130 terminal/retired); log
`.codex-gotmp/selector-resolution-isolated-final-workspace-go.log`. All required
checks green. No production changes after that compilation; committed `b0ee5ae`.

2026-10-04 selector resolution after `fa672eb`: original Test206 now passes
with its entire inherited SuiteTestCase.testSpec, original constructor/settings,
all model assumptions and three byte-identical source vectors. Ordered source
selectors are resolved before ordinary semantic checks. Preserve each argument
group's arity, source definition/module, labels, LET contexts, lifted quantifier/
function/set/CHOOSE formals, and INSTANCE wrappers. Native TLC receives $Nop or
LAMBDA/OpArg/OpAppl nodes directly; no XML roundtrip. Final selected LabelNodes
are retained, `!>>` requires exactly two operands, and INSTANCE definitions and
selected SubstIn copies share cached module-instance bindings/substitution identities.
Selected operator substitutions are compiled rather than installed as raw
flattened names. XML's argument-bearing selector lambdas now consume the same
resolved selection and retain unused-formal argument weights/level omission.
Two old Go-only XML fixtures had selected a quantifier domain before binding
it, which Java rejects. Corrected only their source selector order; retained all
assertions. Both corrected fixtures pass Java SANY. No original test weakened,
new skips or invented persistent tests.
Focused existing parser/semantic/XML + original Test201–206 pass normal6.510s
(21537 retired), race30.383s (58546 retired). Earlier complete Test206 alone
passes1.973s (24502 retired); unchanged Java JUnit passes0.521s. Existing Java
SANY suite passes race8.373s (32940 retired). Existing INSTANCE/level checks pass race95.601s (12144 retired);
all four standard-module/native bridge checks pass race2.043s (5214 retired).
Workspace84712 later failed at Test63; see final gate correction above.
The superseded workspace84999 was interrupted after source-identity changes
(exit130/retired), not a successful gate. Inventory1056/1269 contexts (83.2%),
536/626 classes (85.6%),213 pending across90 classes; legacy73/104 complete.
Remaining selector-negative diagnostics and other legacy contexts are uncredited;
this test translation does not establish complete Generator selector parity.
Next: install original Test207 alone from the retained draft. Preserve the five deferred topics.

Prepared (ignored, uninstalled/unrun/uncredited) three more whole custom legacy
methods in `.codex-gotmp/legacy-custom-lifecycle-three.go.preview`, with six source
fixture mappings in `legacy-custom-lifecycle-three-fixtures.json`: Test216 exact
28/7/0 stats, INIT_GENERATED2 21/s/7 and its full uncovered location; Test220 keeps
checkDeadLock=true so the .cfg decides, no deadlock and2/2/0; TestInvalidInvariant
retains FAILURE_SAFETY_EVAL, exact invariant-level diagnostic and both original
GENERAL assertions. Preserve full default coverage/DOT/JSON/debugger/generation
settings, no added inherited initialization/coverage assertions. Original Java
reference logs already pass all three. After the six remaining inherited cases,
these three are the next custom methods before the 22 remaining negative cases.

2026-10-04: Eighteen original legacy coverage/INSTANCE methods committed
as `dd022c6`; all normal/race checks green. Next original Test201–205 all pass:
complete lambda/recursive operator/recursive function/INSTANCE recursion and
full prefix/qualified operator assumptions retained, 13 byte-exact vectors.
Focused five normal 6.761s (89042 retired), race 22.646s (6503 retired).
No production changes. All five unchanged Java references pass.
Inventory 1055/1269 contexts (83.1%), 535/626 classes (85.5%),
214 pending across 91 classes; legacy 72/104 complete.
Discovery twelve-method run 99132 failed at Test206 after Test201–205 passed
(4.879s, retired). Java passes; Go rejects structural/operator subexpression
selectors as undefined names/incorrect arities, then graph-retention helper
panics because model construction failed. Complete Test206 and six unrun tests
remain ignored drafts and uncredited, with no weakened assertions/new skips.
Fix the real selector translation before installing them or advancing to a new
feature. No broad gates run against a known-failing draft. Committed production
and the five retained whole methods have green checks; deferred topics unchanged.

The five verified methods/13 vectors/4 docs are committed as `c1831ed`.
No repeat full workspace/TLC needed: production unchanged from verified snapshot.

2026-10-04 selector work after `c1831ed`: added Generator.Selector's ordered
step representation to SanyExprSource (`sany_selector.go`, `ast.go`). Preserve
name/null/numeric/first/last/colon/@ kinds, selector syntax and each step's raw
argument node. N_OpApplication attaches its arguments to the final selector as
Java does; argument interpretation stays delayed until operator arity is known.
Added FindingSubExpr operand primitives, distinct from generic AST traversal:
call arguments exclude callee; record access retains string field operand;
function application keeps one tuple argument; CASE retains pair structure;
fairness operand order is subscript/action; quantified/function/set domains keep
source groups; labels unwrap; EXCEPT replacements retain Java's prohibition.
These primitives are not yet wired into semantic selector resolution/TLC lowering.
Test206 remains missing/uncredited; no test assertions changed or new tests added.
An ignored inspection probe of five original Test206 forms verifies exact ordered
steps/kinds/argument groups (64712 retired). Existing SANY parser checks pass
0.022s (67271 retired); existing parser + semantic bridge + all original Test201–205
pass race22.678s (37114 retired). No full-suite claim and no new inventory credit.
Next: connect the selector state machine to source definition/label/LET scopes,
lift bound formals with exact arity, retain INSTANCE/SubstIn identities and carry
selected expressions to native TLC. Fix original Test206 before another feature.

Historical pre-resolution task (superseded by the current snapshot above):
original Test206 selector implementation, not another feature. Complete twelve-method draft preserved as
`.codex-gotmp/legacy-recursion-selectors-twelve-installed.go.preview`;
full 30-fixture mapping `.codex-gotmp/legacy-recursion-selectors-twelve-fixtures.json`.
Only Test201–205 installed; other seven source copies removed from untracked
fixture tree to keep a green baseline, originals remain untouched in Java.
Restore original Test206 alone from draft/mapping, fix implementation, then
advance to Test207/208/209/219/Test99/Test999. No source assertions may be weakened.
Unchanged Java all 104 legacy contexts passed original preflight 11601; per-class
logs `correctness-java/numbered-legacy-suite-<Class>-junit.log`.

Actual failure: Test206 lines97/94 COp!@!++ undefined and INSTANCE A5 arity0
instead5; many structural selectors <<,>>, @, colon-to-LET and nested operator
selectors likewise rejected by CheckSpec. Preserve complete source model.
Root shortcuts located:
- `sanyGeneralIDCall` in sany_translate.go ~2470 flattens every prefix's arguments
  into one list and drops argument-only selector steps. Java Generator.Selector
  keeps parallel ops/args/opsSTN/opNames vectors for every selector step.
  Original SanySyntaxNode retained in SanyExprSource; recover step structure or
  explicitly preserve it in AST, rather than guessing which binder consumes args.
- The new SanySelector step metadata now preserves all original syntax groups,
  but the old sanyParseSubexpressionSelectors still accepts only numeric/name labels, omitting
  @,<<,>>,colon and LET-defined operator selection. Semantic arity/name checks
  do not generate the actual selected lambda/operator before validation.
- TLC bridge IdentExpr simply resolves flattened name; no selected-expression
  resolution. XML exporter has only partial selector/quantifier support; do not
  assume it implements full Java algorithm or route TLC through XML as a shortcut.

Reference `../tlaplus/tlatools/org.lamport.tlatools/src/tla2sany/semantic/Generator.java`:
Selector constants/vectors ~320–450, selectorToNode ~622 onward; FindingOpName,
FollowingLabels, FindingSubExpr; retain formal params/allArgs and SubstIn wrappers.
Rules reviewed: name resolution including canonical operator aliases and LET
contexts, labeled binders, structural operands (CallExpr operands exclude callee),
CASE is pair-of-selectors, record operand selects field value, record component
second operand is string; EXCEPT only first operand legal (original restriction).
Bound identifiers: @ abstracts ALL bounded/unbounded formals; argument-only
selector instantiates ALL, including tuples; numeric/<< />> selects domains.
Labels are transparent to following structural selection; LET one body operand
but colon/name enters its definitions. Keep source module/substitution wrappers
and source positions/capture scopes. Full source Test206 exercises all of these;
Test209/Test219 exercise further instantiated/theorem/numeric selectors.
Do not silence semantic diagnostics or bypass CheckSpec: generate correct selected
expressions/arity/scope and carry them to native TLC semantic nodes.

Credit script current `.codex-gotmp/credit_legacy_recursion_operators_five.py`
idempotent; all20topic rows/13vector bytes verified. Whole 78 inherited preview now
has 71 ported and 7 pending. Legacy remaining32 contexts include those7 and25
negative/custom cases (ETest1–16,Test210/212–217/220/InvalidInvariant).
Parser-negative helper later needs faithful ToolIO capture. Deferred topics
unchanged; fixtures test_vectors only. Overall goal active.

Historical receipts/preparation follow; prior41batch is now committedfcb3f4d.



Latest active slice after committed3822bf9:17 original liveness contexts in
 tlc_liveness_finite_and_loop_java_test.go (11) and
 tlc_liveness_diagnostics_symmetry_java_test.go (6),36 pristine fixtures.
Two faithful production fixes: LiveCheck.check0 returns the violation result
code without manufacturing a thrown exception (Java AddAndCheck ignores its
intermediate result and continues graph insertion); Tool.eval emits coded
TLC_STATE_NOT_COMPLETELY_SPECIFIED_LIVE with exact parameters/expr/context
instead of generic text when a primed fairness variable is undefined.
Original assertion/settings changes: none. First11 discovery omitted Loop from
its run regex and found ForcedPartial failure; corrected complete11 PASS6.943s
normal/41.715s race. FullTLC result-code snapshot PASS65.983s. First6 discovery
failed only317 missing diagnostic; corrected complete6normal PASS14.494s.
All17 unchanged Java references were already PASS in preflights4818/95224.
Corrected6rootrace2281 PASS62.663s, terminal0/retired. Result-code broader
liveness/CodePlex/temporal snapshot7003 PASS412.569s, retired; compiled before
fairness fix/new6, not current full-source gate. Full workspace54860 PASS(root1645.594s,SANY1.010s,TLC65.352s), terminal0/
retired; JSONL liveness-seventeen-current-full-workspace.jsonl. All handles
retired. Keep leaf-specific receipts for the later iterator/matrix additions.
Also installed15 original OffHeapIndexerParameterized inherited contexts across
three subclasses in tlc/fpset_offheap_indexer_parameterized_java_test.go.
All1104rows (duplicates retained), all five methods/full1024-iteration loop,
fresh explicit indexer constructors on every source getIndexer call and exact
source assumptions. Unchanged Java all3PASS31994; Go16285normal PASS3.923s,
81316race PASS16.023s, both retired. No production changes for matrix slice.
Workspace54860 was compiled before matrices: its current production/model
coverage remains valid; complete leaf matrices verified independently.
Inventory955/1269(75.3%),460/626classes(73.5%),314pending/166classes,
onepartial; liveness99/101. All20topicrows and36fixture bytes verified.
Remaining liveness cases: deferred checkpointFL2 and source-failing warning.
All required checks green.47 intended paths staged (3production,4testfiles,
36vectors,4docs), whitespace checked excluding pristine vectors. Commit now
before next production changes.
Also installed original iterator9contexts in tlc/fpset_iterator_java_test.go,
after porting getLast reverse-positive scan/reads plus source typed exhaustion
and monotonic Assert boundaries in fpset_disk.go. All3original Java refsPASS;
Go73707 PASS0.010s, terminal0/retired. Current fullTLC72092 PASS69.871s and relevant
FPSet/DiskFP/MSB/Iterator race64996 PASS52.546s, both terminal0/retired; logs tlc-iterator-current-full-tlc-go.log
and tlc-iterator-and-fingerprint-storage-race.log. Workspace54860 predates only
this isolated iterator leaf change and added matrices/iterator source tests;
its current liveness/evaluator integration remains relevant. Keep gate receipts
honest, no redundant whole-workspace restart for these leaf methods.
Ignored previews: first18legacy whole SuiteTestCase contexts/36fixture mappings
in legacy-suite-first-eighteen.go.preview; also78whole inherited suite contexts
(with four complete additional coverage hooks) and172fixture mappings in
legacy-suite-inherited-seventy-eight.go.preview. Includes complete large
Test19/27/33 state spaces, no sampling. Both uninstalled/uncredited; inspect
before install, start first18 after this green commit. All104unchanged Java
legacy references alreadyPASS11601. Renderer never rerun wholesale.

Next arrays references verified99350: unchanged LongArrayTest all7PASS0.412s,
LongArraysTest all6PASS0.012s. Ignored long-array-seven-methods.go.preview and
long-arrays-six-methods.go.preview contain all13source methods/complete helpers.
LongArray bounds need actual Java AssertionError type before source catches can
pass; existing rangeCheck panics string. Add only core generic exception carrier,
not forbidden email/dependencies/JVM emulation. Keep testZeroMemory's original
both i<j/i++ loops unchanged even though zero iterations occur. Random-swap
retains unseeded JavaRandomDefault/full21383 values; swap uses10321.
LongArrays Basic2 copies44full signed literals; range method retains15source
assertions and all comparator/sentinel/member/wrapped-range/order/count loops.
These array ports remain UNINSTALLED/UNCREDITED; review source before use.

2026-10-04 latest checkpoint: 24 original Examples/temporal/initial contexts,
53pristine Java fixtures and two production fixes committed9acb8a7. Full current
workspace PASS(root1621.178s,SANY0.918s,TLC66.273s), original initialization/
temporal race23methods PASS72.231s; earlier full TLC race PASS624.947s.
All those handles terminal/retired; working tree clean at9acb8a7.

Now installed four original liveness methods in
 tlc_liveness_four_simple_java_test.go: Github604/Github702/IncompatibleTypesLive/
TwoPhaseCommit, eight pristine fixtures. All unchanged Java references PASS in
preflights4818/95224. Go66312 PASS19.070s, terminal0/retired.
Race9811 terminal1/retired after120.449s: first3 PASS, TwoPhaseCommit fails only
original zero-uncovered assertion because early periodic report contains zeros,
final report all covered. No data-race report. Profile97203 terminal1/retired
103.792s, same timing failure. Profiling retained original test/settings;
binary moved to ignored .codex-gotmp/two-phase-original-race.test, cpu profile
same prefix.cpu. Hot path is source-faithful symmetry/value comparison; no
speculative optimization or race-disable attributes added.

Java TestMPRecorder retains every report; Go zero-coverage helper matches it.
AbstractChecker.runTLC coverage countdown matches Go (coverage/progress interval,
3sec first wait then countdown ticks). Independent unchanged-Java diagnostic8148
reproduces identical assertZeroUncovered failure: -Xint and
-Dtlc2.TLC.progressInterval=120, original test/model/CLI unchanged,52.317s.
This diagnostic changes pacing, is NOT baseline success evidence, and does NOT
justify weakening assertions, dropping earlier coverage or changing source flags.
Keep test complete/original and document original timing limit under extra race
instrumentation; normal constructor/default Java and Go cases pass.

Source audit found separate actual production omission: periodic Go coverage
called reportCoverage directly without Java's coverage-end/overhead marker.
checker.go reportPeriodicCoverage now calls ReportCoverage(tool,StartTime),
matching AbstractChecker.reportCoverage -> CostModelCreator.report. No state
algorithm/assertion changes. Gates retired: whole4normal92360 PASS20.519s;
fullTLCnormal21664 PASS65.347s; original19coverage+first3liveness root
race83939 PASS90.236s. This race gate excludes TwoPhaseCommit.
Final source review corrected IncompatibleTypesLive coverage to its original
inherited true default; final whole4normal48581 PASS5.915s. Its focused race
41443 PASS5.915s, terminal0/retired. All current gates are terminal.

Inventory914/1269(72.0%),437/626(69.8%),355pending contexts/189classes,
onepartial, liveness82/101. All20topic rows verified. Commit14intended paths
(source test1,8pristine fixtures,4docs,checker.go) after current normal/source
coverage gates; source assertion unchanged. Whole9acb8a7 workspace receipt
plus relevant current gates suffice for this reporting-boundary leaf change.
Next: eleven finite/loop methods then six diagnostics/symmetry methods already
prepared, review and install each coherent slice, fix any production failure
before moving on. After these, liveness topic leaves only deferred checkpoint
FL2 and the unchanged-Java-failing LivenessSymmetryWarning. Keep source failure
visible; don't weaken expectations or invent a skip.

Historical receipts and independent preparation for continuation follow;
the previously described 24-context batch is now committed9acb8a7.

2026-10-04 latest: Previous six lock/protocol/memory/cache Examples committed
00893c4 after focused normal61.674s/race631.084s and fullworkspace PASS
(root1181.648s,SANY0.993s,TLC65.132s). Those handles are terminal/retired.

Current uncommitted batch: twelve remaining original Examples methods and29
byte-identical Java fixtures in test_vectors. MCYoYoNoPruning/NbacgGuer01/
Prisoners/SDPAttackNewSolution/SchedulingAllocator/SimpleAllocator/
SingleLaneBridge/SpanTreeRandom/SpanTree/SyncTerminationDetection/SDPAttack/
MCDistributedReplicatedLog. All unchanged Java references PASS (38279,62315,
90667,25396 terminal/retired). All original constructor flags, class isolation,
classpath, recorder assertions, full postconditions, traces and exit checks
retained. SDPAttack: stats1103/526/175, all11 states, full21-argument formatter
and14constant strings. ReplicatedLog is an ordinary local single-worker model,
not distributed TLC transport: stats271/37/0, all6states and named back-edge2.

Initial Go77918 terminal1/retired after594.281s:11pass, onlyYoYo full
postcondition fails. Counts and states match Java110/60/depth19, but Go wrapper
lost the source action location. Java Specs.addSubsts uses the source
SubstInNode copy constructor, retaining syntax and shared substitution array.
Production NewSubstInNodeFromSource now preserves TreeNode/Location/Substs;
SpecsAddSubsts calls it. No weakened assertion. Focused YoYo9506 PASS2.065s.
Current complete12 normal25227 PASS570.714s, terminal0/retired.
YoYo+19 original coverage race90782 PASS100.888s, terminal0/retired.
Full TLC package race7106 PASS624.947s, terminal0/retired, log
.codex-gotmp/examples-source-subst-full-tlc-race.log.
Full workspace89315 PASS(root1600.051s,SANY1.089s,TLC67.616s), terminal0/
retired. JSONL
.codex-gotmp/examples-twelve-full-workspace.jsonl (60m timeout).
An initial workspace invocation used a malformed timeout flag and terminated
before tests; corrected89315 is the actual gate. Do not restart live gates.

Next two source-only slices now installed, with no production change from
TemporalDoubleNegation tests:8whole subclass contexts/16pristine fixtures,
normal38137 PASS5.658s, race28833 PASS26.526s, terminal0/retired. Four original
Github1037/Github790/InitialLivenessEvaluationError/InvariantOnly contexts and
8pristine fixtures installed. Initial run48878 terminal1/retired: only initial
liveness error fails, Go prints same Java diagnostic/stats but exit255 versus
source allowed151(config error)/12(safety). Java always takes original thrown
exception code before diagnostic replay; Go reportInitException kept a generic
failure unless result==NoError. Production checker.go now unconditionally
uses initExceptionCode(err) for non-nil exceptions, exactly source catch logic.
All4 corrected99897 PASS4.967s, terminal0/retired; original assertions unchanged.

Current gates for this second production fix: fullTLC normal37983 PASS64.054s, terminal0/retired,
.codex-gotmp/initial-live-error-full-tlc-go.log; root race48326 PASS72.231s, terminal0/retired, covering
all12new temporal/initial cases plus11existing original initialization methods,
.codex-gotmp/initial-live-error-init-and-temporal-race.log. Workspace89315 PASS but was compiled BEFORE the checker.go initial-error fix and the
12new temporal/initial tests; do not claim that gate covers these additions.
Its root integration evidence remains useful for the source-constructor fix;
Fresh full workspace65076 PASS(root1621.178s,SANY0.918s,TLC66.273s), terminal0/
retired, compiled with both production fixes and all24new source contexts. Log
.codex-gotmp/examples-temporal-initial-current-full-workspace.jsonl. All current
gates green; no live handles. Ready for authorized commit, then install remaining
original liveness model slices. Do not restart terminal runs.

Inventory now910/1269(71.7%),433/626(69.2%),359contexts/193classes pending,
onepartial; liveness78/101,23pending. All20topic rows verified. Exactly63 intended paths staged/reviewed:3production files,3source test files,
4docs,53byte-identical Java fixtures. Latest handoff/progress receipts to be restaged
before commit. No production/test edits beyond those compiled by65076.

Independent preparation, uninstalled/uncredited: liveness-four-simple.go.preview
(4source methods/8fixture maps), liveness-finite-and-loop.go.preview (11methods/
22fixture maps). Source-only review JSON remaining-liveness-review.json covers
21methods. Java preflight4818 first12classes PASS then terminal1 on unchanged
LivenessSymmetryWarning inherited expected-success exit (actual13 from April25MC).
That source contradiction is noted in TODO; don't weaken or manufacture success.
Remaining9 graph models95224 PASS, terminal0/retired. Finite/loop preview now
regenerated with source assertion order by prepare_liveness_finite_and_loop_ordered.py.
Also prepared liveness-six-diagnostics-symmetry.go.preview (6complete original
methods/14fixture maps), preserving source assertion order, workers NumCPU, full
nested-expression strings, exact named action labels, graphs, register42 and
coverage checks. All21 eligible remaining liveness models have green unchanged
Java references. Their Go drafts remain uninstalled/uncredited; review all
fixtures/imports/settings before installation, fix any production failure first.

Offheap matrix preview offheap-indexer-parameter-matrices.go.preview preserves
all1104rows per class, five inherited methods and original assumptions (15
logical contexts/16560parameter contexts). All3 unchanged Java reference classes
31994 PASS, terminal0/retired; not installed or credited. TLC iterator9 original
methods Java PASS in preflight-tlc-iterator-source-methods.py; ignored preview
tlc-iterator-source-methods.go.preview. Production needs Java getLast/reads
and strict NoSuchElementException exhaustion instead of io.EOF before test port.
Existing msbDiskIterator has only hasNext/next; source MSBDiskFPSet.java lines186+
contains full methods. Keep native error-return adaptation with exact exception
boundary, don't weaken original catch. These methods are not installed/credited.
Independent numbered-suite preparation:104 non-ignored concrete classes
extracted from source inventory (exclude abstract SuiteTestCase and ignored
Test61). Unchanged Java preflight11601 PASS all104, terminal0/retired, script
.codex-gotmp/correctness-java/preflight-numbered-legacy-suite.py, per-class logs
numbered-legacy-suite-<Class>-junit.log. Compiles SuiteTestCase/SuiteETestCase and
original TestPrintStream. Source review JSON numbered-legacy-source-review.json.
No Go numbered test/fixture installation or inventory credit yet. Future
negative SANY cases must use actual source console capture and production error
propagation; current shared model helper requires zero parser diagnostics, so
that assumption cannot be carried into intentional parser-error cases.
SuiteTestCase: default stats2/1/0/1; preserve all subclass constructor counts,
uncovered sets and Test52/55/56/63 additional full coverage. Test19/27/33 retain
full large state spaces, no sampling. SuiteETestCase: actual ToolIO out+err
TestPrintStream capture, exact substring checks, constructor args (ETest8 uses
simulation and original ERROR TODO), deadlock overrides (ETest3/Test220) and
exit checks. TestInvalidInvariant uses root test-model directory, not suite.
Five user-deferred topics and checkpoint FL2 variant remain deferred.


2026-10-04: Six Examples tests and namespace fixes committed52959cb after
full workspace89196 PASS(root1143.104s,SANY0.908s,TLC66.125s), focused root
race12129 PASS207.031s (six examples+19 original coverage methods), and four
existing bridge tests race92665 PASS1.937s. All handles terminal/retired.
Previous nine-example batch efdd910 also green and committed.

Current batch: all48 original liveness-helper contexts translated in five new
Go test files under tlc/: liveness_expression_java_test.go,
liveness_tbpar_java_test.go,liveness_graph_node_java_test.go,
liveness_tableau_node_ptr_java_test.go,liveness_disk_graph_java_test.go.
Six source classes; TableauDiskGraph includes all seven inherited DiskGraph
methods with original tableau index0. All233 source assertion sites retained
(including inherited copies), Java Random seed4711 and full loops preserved,
TBPar Object equality assertions use pointer identity; LiveExprNode keeps four
commented LNNext TODO assertions inactive exactly as source. Original dummy
OpAppl string override and dedicated graph directories retained. No new vectors.
All48 unchanged Java references previously PASS JUnit0.064s. Initial Go
translation had two boolean-parentheses compile errors, corrected. Go70672
terminal1:46 pass, only two partial-graph error strings fail (case mismatch).
Java AbstractDiskGraph.getPath specifies Couldn't and signed long state text;
Go production DiskGraph/TableauDiskGraph now match capitalization and int64
formatting at all four failure returns. No test weakening. All48 corrected
Go23898 PASS0.041s, terminal0/retired. First gates race93882 PASS1.073s/fullTLC7989 PASS66.543s,
terminal0/retired. Further source catch audit identified untyped errors as
another shortcut: Java catch(RuntimeException) must reject IOException.
Four failure returns now use the existing generic RuntimeException with
exact message/signed state; source-test catches retain that type boundary.
Current48 Go17661 PASS0.039s, terminal0/retired. Fresh focused race30550 PASS1.078s, terminal0/retired,
liveness-helpers48-runtime-type-race.log. FullTLC47362 PASS65.935s, terminal0/retired,
liveness-helpers-runtime-type-full-tlc.log. All current gates green,
11 intended staged paths reviewed; ready for authorized commit.
Root integration success behavior/source tests unchanged since52959cb green
workspace1143.104s; no repeat broad run justified by this leaf failure boundary.

Inventory880/1269(69.3%),403/626classes(64.4%);389 contexts/223classes pending,
one partial. All20topic rows verified. Liveness helpers category now48/48.
Next six Examples Java preflight already PASS (LockHS/MCAlternatingBit/
MCEWD687a/MCLiveInternalMemory/MCLiveWriteThroughCache/MCWriteThroughCache);
translations and26 fixture maps in ignored examples-fourth-six.go.preview and
examples-fourth-six-fixtures.json, no active source/credit yet.
Five user-deferred topics and checkpoint FL2 variant remain deferred.

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


ModelValueTest is port complete in value_model_test.go: all 44 original methods
mechanically translated in the actual JUnit 4 runner order. Retained constructor
locations, false single-field RecordValue normalization flags, direction-specific
BoolFalse/BoolTrue inputs, the source's StringValue("foo") in testCompareToTupeMV,
exact +/-1/0 assertions and TLCRuntimeException expectations. Original method,
assertion and expected-exception counts reconcile. A fresh class-level intern table
and model registry isolate the original execution while restoring prior Go state;
no per-method reset, input hoisting, weaker sign checks or any-error catches.
Current original Java production/test class passes JUnit -ea (44, 0.032s);
final targeted Go passes (0.012s), full TLC normal passes (3.593s), full TLC
-race passes (26.070s). No production changes needed. Retire handles 35654,
12212, 57494 and 21067. Prior full workspace normal remains green; this slice
changes tests/docs only. Inventory now 312/1269 mapped methods (24.6%), 108/626
complete classes (17.3%), 957 pending contexts across 518 classes, two partial.
Next finish TypedSetTest.testParseSet6 with the actual Java null input; Go's
string-only ParseTypedSet API currently cannot represent it. Port the nullable
boundary first, retaining existing string callers and all four original null-set
assertions. The five user-deferred topics stay deferred.


TypedSetTest is port complete: testParseSet6 now preserves all four original
empty/null/comma/braced-comma inputs and equality assertions. Added production
ParseTypedSetNullable(*string), with existing string callers delegating to that
same implementation. Actual nil input returns a newly constructed empty set.
Source comparison also corrected Go Unicode TrimSpace to Java String.trim's
code-unit <= U+0020 rule at both trim points. Current original TypedSet.java and
unchanged original JUnit class pass all six methods (0.007s); targeted Go passes
(0.017s), full TLC -race passes (27.010s). Retire handles 3734 and 3814.
Full workspace normal checks passed (312.382s root, 1.047s SANY, 3.735s TLC),
correctness-java/typed-set-full-normal.log; retire 10817. That snapshot predates
only the final split-pattern whitespace correction; final whole-TLC normal and
race checks cover the resulting implementation (3.642s/27.112s); retire 4655/99372. Inventory now 313/1269 mapped contexts (24.7%), 109/626
complete classes (17.4%), 956 pending contexts across 517 classes; only one class
is partial. No source assertions weakened or invented cases. Deferred topics stay
skipped. ModelValueTest was committed separately as 1aba5f3.


Next lazy-function slice: current FcnLambdaValue.java and unchanged original
FcnLambdaValueTest compile with original test/tlc2/tool/TLCStates.java (required
support class absent from the frozen jar), and Java -ea JUnit passes all eighteen
(0.040s). Not yet credited as Go translations. Go Tool.EvalFunc already supplies
the same constant-result eval mock; preserve params/formal x, null semantic body,
EmptyContext, EvalClear, actual TLCStates one-variable dummy state (v0=0, uid=0),
and nil predecessor state. Retain function/EXCEPT identity, conversion ordering,
all nested EXCEPT assertions and fingerprint seeds; don't substitute parser models
or approximate bad-sharing checks. Only existing Go lazy test is a fallback-format
case, not a translation of these original methods.


FcnLambdaValueTest is port complete in value_lambda_java_test.go: all eighteen
original methods mechanically translated, with all 49 equality, three null and
one identity assertions reconciled against source. Preserved the constant-result
Tool.EvalFunc mock, formal x/arity zero, nullSN semantic body, EmptyContext,
EvalClear, TLCStates dummy v0=0/uid=0, nil predecessor, constructor locations,
conversion ordering, nested/shared EXCEPT regressions, cache identity and original
fingerprint calls/FP64.Init sites. Native zero CostModel (Node nil) represents
source null arguments. Class cleanup restores intern/state/tool/metadata/symmetry
and fingerprint globals without resetting between original methods. Source and
Go lazy-function implementation call only Tool.eval; no other mock methods need
runtime reflection boilerplate. No original assertions weakened, fixtures changed,
or invented tests added; no production changes needed. Current source Java class
and unchanged original test plus TLCStates helper pass JUnit -ea (18, 0.038s).
Targeted Go passes (0.011s); full TLC normal passes (3.516s), full TLC -race passes
(26.717s). Retire handles 36079, 13799, 50272. Prior full workspace results remain
green; this slice adds tests/docs only. Inventory now 331/1269 methods (26.1%),
110/626 complete classes (17.6%), 938 pending methods across 516 classes.
Next: complete SetOfFcnsValueTest, retaining indexed subset enumeration rather
than substituting ordinary enumeration, all empty domain/range sampling checks,
all expected function/domain/value/membership assertions, exact non-enumerable
errors and all four huge sets at sample sizes 0,1,2,799,1024,8932,16933. The commented
109031 sample is not an active source case. Do not downgrade this correctness test
to a benchmark exclusion or reduce its sample matrix. Five user-deferred topics
remain deferred.


SetOfFcnsValueTest is port complete in value_fcn_set_java_test.go: all sixteen
original methods translated with indexed SubsetEnumerator.elementAt, literal
expected functions, domain/value lengths, membership, Java-compatible HashSet
hashing/equality, empty/non-enumerable domain/range cases and exact exception
messages. Preserved all four huge function sets and all seven active sample sizes
0,1,2,799,1024,8932,16933 (28 cases), including size-overflow exception checks and
duplicate counts. The source-commented 109031 case remains inactive. No seed or
replacement cases invented and no production changes needed. Current Java source
classes and unchanged JUnit test pass all sixteen (2.478s), Go targeted passes
(8.182s), whole TLC normal passes (10.529s); whole TLC -race passes:
ok  	github.com/glycerine/tlago/tlc	91.217s
Retire handles 60975,18330,99392,54091. Prior full-workspace results remain green;
this slice adds tests/docs only. Inventory now 347/1269 contexts (27.3%), 111/626
complete classes (17.7%), 922 pending contexts across 515 classes; one partial.
All five user-deferred topics remain deferred. Next reconcile TupleValueTest's
complete original testErrorMessages, including the multi-argument apply overload.


TupleValueTest is port complete: testErrorMessages preserves all four original
catch-only contains assertions, original values and call order. Source has no
fail() after its try blocks; Go likewise checks caught TLCRuntimeException messages
without inventing stronger cases. Ported missing production TupleValue.ApplyArgs:
source argument-count assertion, single-argument delegation (control discarded),
and original catch/wrap boundary. Existing Apply/Select assertions now retain
Java Values.ppr and source-aware Assert.fail behavior (detailed failure with
EmptyContext, fingerprint wrapper, Runtime tag), following TupleValue.java and
util/Assert.java. No test weakening or new cases. Current source Java class and
unchanged original JUnit pass (1, 0.032s), Go targeted passes (0.013s).
Full workspace normal check passes on final production/test snapshot:
ok  	github.com/glycerine/tlago	316.574s
?   	github.com/glycerine/tlago/cmd/tlago	[no test files]
ok  	github.com/glycerine/tlago/sany_tests	0.971s
ok  	github.com/glycerine/tlago/tlc	10.689s
Full TLC -race passes: ok  	github.com/glycerine/tlago/tlc	95.634s
Retire handles 5019,37400,66512. Inventory now 348/1269 contexts (27.4%), 112/626
complete classes (17.9%), 921 pending contexts across 514 classes; one partial.
Five user-deferred topics remain deferred.

Next SetOfRcrdValueTest: compiled current source SetOfRcdsValue.java and unchanged
original JUnit; all seven pass Java -ea (16.119s), retire handle39273. No Go credit
yet. Draft Go translation is only ignored scratch at
.codex-gotmp/correctness-java/value_record_set_java_test.go; not compiled or
accepted. Review against original before promoting. Preserve indexed elementAt
names/field membership, HashSet equality/cardinality and all loops n=1..6,
m=1..4, kOutOfN=0..size-1, clearing the same set each time, plus ten fields of
fifty values sampled at k=10000. Keep actual non-enumerable/empty-field cases and
class FP64.Init setup. Reuse Java-compatible hash/equality helpers from the
completed function-set translation; never replace this matrix with samples or
exclude it as a benchmark. Finish needed production work before moving on.


SetOfRcrdValueTest is port complete in value_record_set_java_test.go: all seven
original methods, original helper constructors/normalization flags, @BeforeClass
FP64.Init, indexed SubsetEnumerator.elementAt for names/field membership,
ordinary/indexed HashSet equality, original length/membership/cardinality checks,
empty/non-enumerable fields and actual needBigInteger/elements(k) checks.
Preserved the entire n=1..6, m=1..4, kOutOfN=0..size-1 sweep: 6,684 samples checking
9,246,174 records, clearing the same HashSet each iteration; astronomical sample
uses ten fields of fifty values, k=10000. No fixed seed invented. Source Java
classes and unchanged original seven-method test pass (16.119s); Go targeted
passes (71.528s, predates only the shared HashSet.add helper correction). Corrected
that helper to hash once per insertion as Java HashMap.put does, retaining actual
ValueJavaHashCode and equals for collision/duplicate detection. Final record-set
snapshot whole-TLC normal passes (40.915s), whole-TLC -race passes:
ok  	github.com/glycerine/tlago/tlc	392.825s
Retire 76243,66031,35695. No production changes or weakened/invented cases.

SubsetEnumeratorTest is port complete in value_subset_enumerator_java_test.go:
both original methods over all twelve shared parameter rows and all eleven
fractions each (264 cases). Preserved original inputs/constructor flags, decimal
ASCII model values (65..74, not letters), Math.ceil, source fingerprint setup after
all parameters are constructed, all size/unique/membership assertions, both
HashSet constructions and native elements(k)/getRandomSubset production paths.
All twelve source values are finite; native randomSubsetOfEnumerable dispatch
retains its existing finite-set boundary before the source-equivalent sampling
path. Original source EnumerableValue and unchanged original JUnit pass all
24 expanded methods (0.059s). Go target passes (0.055s), targeted -race passes
(1.239s); retire 6678,79270. The whole-TLC race result above predates only this
new test file; its complete new matrix has the separate passing race run.
No production changes required. Final full-workspace normal check includes both
new classes and the shared helper correction:
ok  	github.com/glycerine/tlago	317.352s
?   	github.com/glycerine/tlago/cmd/tlago	[no test files]
ok  	github.com/glycerine/tlago/sany_tests	0.999s
ok  	github.com/glycerine/tlago/tlc	42.212s
Retire handle64633. Inventory now 357/1269 methods (28.1%), 114/626 complete classes
(18.2%), 912 pending methods across 512 classes; one partial. Five user-deferred
topics stay deferred. Scratch drafts are superseded by tracked files.

Next SubsetValueTest: read all 37 methods/helpers and verified current Java source
Combinatorics, SubsetValue and KSubsetValue against unchanged JUnit: all 37 pass
Java -ea (0.91s), subset-value-junit.log. Not credited as Go translations yet.
Correct identified production error-family gaps before translating their catches:
numberOfKElements returns fmt.Errorf for >Long.MAX_VALUE but Java throws
IllegalArgumentException("k=... and n=..."); newKElementEnumeration likewise
returns fmt.Errorf("Subset too large.") instead of IllegalArgumentException.
Keep original class seed 15041980 and FP64.Init; fixture Collections.shuffle uses
its own default java.util.Random, independently of RandomEnumerableValues.
Preserve shuffled values even when source flags them normalized; do not substitute
sorted fixtures. Retain TreeSet comparison-based duplicate detection where used,
complete 65,536-subset/unranking sweeps, all 61 binomial sum rows, original
23,131/23,071/23,077 random-set sample sizes, normalized ordering, actual exception
types, empty-enumerator independence and randomized KSubset cases. Do not replace
original class methods with existing approximate Go checks or smaller samples.


SubsetValueTest is port complete in value_subset_java_test.go: all 37 original
methods and all original helper assertions/fixtures translated. Preserve original
seed 15041980 and FP64.Init, independently shuffled fixtures using a separate
default JavaRandom and the actual Collections.shuffle descending Fisher-Yates
algorithm, including source's normalized=true flags on shuffled values. Verified
all 37 method positions against actual JUnit runner descriptions; run in that
order and share the seeded TLC generator across Go subtest goroutines, preserving
source's single-thread random stream. Restoring global state is class isolation;
no reset between original methods. TreeSet helper retains Java comparator calls,
first-key validation, red/black insertion and rotations, not hash-based substitutes.
HashSet helper now preserves key identity and null handling before equals.
Kept both complete 65,536-subset sweeps, all 61 binomial sum rows, expected boundary
counts and real IllegalArgumentException catch, all explicit sorted K-subset rows,
2,342/4,223/2,148 samples, 23,131/23,071/23,077 random-set samples, normalization
and list-order checks, empty-enumerator independence and all randomized K cases.
Catch-only testRandomSubsetSubsetNoOverflow remains catch-only as source.

Production corrected NumberOfKElements and bounded KElementEnumeration overflow
families to IllegalArgumentException; exposed direct ElementsNormalized separately
from cached Elements, retaining source-aware assertion failures. Initial faithfully
ported overflow-catch test failed because SubsetValue.Size returned an EvalException
carrier. Fixed actual production to NewTLCRuntimeException with original code and
parameters, leaving test catch/assertions intact. All 37 Go methods now pass
(1.612s targeted; predates only final non-enumerable/source-detail correction).
Current source Combinatorics, SubsetValue, KSubsetValue and unchanged original JUnit
pass all 37 Java methods (0.91s). Final snapshot full workspace normal passes:
ok  	github.com/glycerine/tlago	316.108s
?   	github.com/glycerine/tlago/cmd/tlago	[no test files]
ok  	github.com/glycerine/tlago/sany_tests	0.921s
ok  	github.com/glycerine/tlago/tlc	43.204s
Full TLC -race passes: ok  	github.com/glycerine/tlago/tlc	420.487s
Retire handles 61943 (initial red),88821,26161,21887. No tests weakened, matrices
reduced, alternate fixtures substituted or invented regression/unit cases added.
Inventory now 394/1269 contexts (31.0%), 115/626 complete classes (18.4%), 875 pending
contexts across 511 classes; one partial. All five deferred topics stay deferred.

Next: complete KSubsetValueTest's eleven original methods. Source Java class and
unchanged original JUnit already pass all eleven (0.276s), ksubset-junit.log.
Read every method and its shared doTest helper; no Go credit yet. Correct production
KSubsetValue.Size's remaining fmt.Errorf overflow to actual IllegalArgumentException
before translating testEnumerateN64's catch. Its inherited Java toSetEnum does call
size() before constructing the enumerator, matching native ordering; don't replace
that guard with another message/path. Retain all four n=32/33/63/64 enumeration
methods including null/size assertions, normalization's six exact elements, all
thirty original fingerprint matrix rows (2+10+10+8), before/after fingerprint and
explicit enumeration sizes, all eleven invalid-k assertions in both operand
directions for both k=-1 and k=4, exact hash and strings, and checked printing's
large-count swallow case. Source uses FP64.Zero for the fingerprint calls.


KSubsetValueTest is port complete in value_ksubset_java_test.go: all eleven
original methods, four n=32/33/63/64 enumerations with every non-null/size
assertion and original IllegalArgumentException catch, normalization's six literal
expected subsets, all thirty original fingerprint matrix rows (2+10+10+8), exact
Combinatorics.choose expected sizes before/after FP64.Zero fingerprinting and after
explicit enumeration. Preserve every invalid-k assertion for both -1 and 4:
size/isEmpty, both equality and comparison directions, hash equality, ordinary and
randomized null enumeration, toSetEnum equality and exact checked string. Checked
printing's large-count swallow string is unchanged. No source test assertions
weakened, matrices reduced, cases invented or commented specs made runnable.
Production KSubsetValue.Size now returns actual IllegalArgumentException with
original k/n message at the original count guard, replacing fmt.Errorf. Source
inherited toSetEnum's size-before-enumerator order remains correct.
Source Java and unchanged JUnit pass all eleven (0.276s), Go target passes (0.680s).
Final full-workspace normal check passes:
ok  	github.com/glycerine/tlago	316.529s
?   	github.com/glycerine/tlago/cmd/tlago	[no test files]
ok  	github.com/glycerine/tlago/sany_tests	0.988s
ok  	github.com/glycerine/tlago/tlc	44.256s
Full TLC -race passes: ok  	github.com/glycerine/tlago/tlc	426.589s
Retire handles 79623,24328,41661. Inventory now 405/1269 contexts (31.9%), 116/626
complete classes (18.5%), 864 pending contexts across 510 classes; one partial.
Five user-deferred topics stay deferred.

Next value streams: read all six original ValueInputOutputStreamTest methods and
StringDeserializeTLCTest plus original model references. Current source
ValueInputStream/ValueOutputStream and unchanged original six-method test pass
Java -ea (0.035s), value-stream-junit.log. Not credited as Go translations yet.
Preserve source file constructors/global compression, individual short/int and
natural operations, actual Close ordering, exact gzip file-length assertions
23 (short naturals) and 26 (naturals), actual cold raw string/record fixtures with
-1 metadata and exact source strings, and readExternal followed by first interning
for assertions. Never replace the two compressed lengths with raw lengths or
roundtrip-only checks. Ignored gzip-size-probe.go directly exercised the original
two natural cases against current production; Go returns 27/30 bytes, not 23/26.
Actual production compression must be fixed before these tests are completed.
Optional user architecture question asked whether pure Go is required or CGO/zlib
is acceptable; no response as of this commit, and default assumption after the
response window is pure Go. No compression changes made yet. Honor a later reply.
Original compressed StringDeserialize.vos, .tla and .cfg are in upstream test-model;
use tlc/test_vectors for any persistent copies and preserve complete inherited
model harness/exit assertions in addition to the two explicit recorder checks.


ValueInputOutputStreamTest is port complete in value_stream_java_test.go: all six
original methods, global compression constructors, short/int extrema and zero,
compact-natural read/write operations and the unchanged compressed lengths 23/26.
The blind string and record reads write the literal UTF-8 bytes and -1 metadata
without using Value/UniqueString before reading; retain kind, interning/text,
record cardinality and selection assertions. Test globals match Java's isolated
classloader and are restored after the class.
Production gzip output now uses a pure-Go port of JZlib's default zlib deflater
(level 6, window 15, memory level 8, NO_FLUSH/FINISH). Preserve source lazy matching,
hash chains/window sliding, token limits, Huffman heap ties/overflow repair,
run-length tree encoding, stored/fixed/dynamic block selection, bit packing,
terminal block and Java's gzip header/CRC/size trailer. No input special cases,
trailer stripping, buffering the entire input, CGO or external process dependency.
Algorithm provenance is JZlib a21be20213d66eff15904d925e9b721956a01ef7; original BSD
copyright/license notices are retained in code and JZLIB_LICENSE.txt. The optional
architecture question had no answer; the previously stated pure-Go assumption
remains in effect. Do not claim compressor byte equality for all native-zlib
versions/inputs: original cases match, and ephemeral native-zlib comparisons of
the source .vos payload repeated 0/1/2/3/4/20/100/1000/10000 times (up to 9.48MB)
were byte-identical and successfully decompressed. Those probes are not new tests.
StringDeserializeTLCTest.test is port complete in tlc_string_deserialize_java_test.go,
with all three original vectors byte-identical in test_vectors/models/StringDeserialize.
Its .vos is uncompressed (the original model calls IODeserialize(..., FALSE));
the earlier handoff description calling it compressed was incorrect. Retain the
original property path setup, CommunityModules classpath (IOUtils), full inherited
checker settings/success exit assertion and both FINISHED/GENERAL recorder checks.
The initial Go run failed to resolve IOUtils until the test supplied the same
CommunityModules jar classpath used by Java. No production shortcut or weakened
assertion was needed. Unchanged Java original passes (0.582s), Go passes (0.160s).
All six Go stream methods pass (0.017s); source Java six methods passed (0.035s).
Inventory now 412/1269 contexts (32.5%), 118/626 complete classes (18.8%), 857 pending
contexts across 508 classes; one partial. Values/value streams topic is complete:
17/17 classes, 189/189 methods. All five deferred topics stay deferred.
Final full workspace normal passes (handle 96850 retired):
ok  	github.com/glycerine/tlago	317.159s
?   	github.com/glycerine/tlago/cmd/tlago	[no test files]
ok  	github.com/glycerine/tlago/sany_tests	1.068s
ok  	github.com/glycerine/tlago/tlc	43.469s
Targeted deserialization model race passes (1.639s), handle 75318 retired.
Final targeted six stream methods pass (0.016s), handle 65645 retired.
Unchanged original Java six methods rerun: 6 pass (0.034s).
Full TLC race passes: ok  	github.com/glycerine/tlago/tlc	415.789s
Handle 6990 retired; all verification handles in this slice are terminal.
Next: reconcile both existing DiskPoolWriterTest original methods, retaining the
empty-wakeup/finish behavior and source timeouts before crediting the queue topic.


DiskPoolWriterTest is port complete in disk_pool_writer_java_test.go: both original
state and byte-array writer methods, temporary disk directories and finishAll
teardown, every shared wakeup/return-to-WAITING/aliveness/termination assertion,
5000ms deadlines and 10ms polling. Notification remains under the actual writer
mutex; the test observes its real native goroutine runtime state while the writer
cannot reacquire the lock. Actual termination substitutes for Java Thread.join;
neither a finished flag nor a notification acknowledgement replaces liveness.
Go runtime.Stack is used only in the translated test helper. Snapshot native IDs
before construction, then identify the unique newly started writer by its Start
frame so its startup wrapper remains alive even before run begins. The first
receiver-frame-only translation falsely reported termination during startup;
corrected the observation, not the assertions or production behavior. The test
fails explicitly if identity is missing/ambiguous. Thread state parsing is a Go
runtime-format dependency localized to this source-test adaptation.
Production StatePoolWriter and ByteArrayPoolWriter wait/finish loops were compared
with current Java source; the guarded empty wake and finished checks already
match and needed no production change. Current Java source and unchanged original
JUnit pass both methods (0.049s). Go originals pass 100 repetitions (1.280s), and
-race 100 repetitions (1.987s). Full TLC normal passes (43.391s). Earlier committed
068acee has full workspace normal and full TLC race green; this slice changes only
tests/docs. Retire handles 70263/17457 (initial observation reds), 30927/59827/39824.
Inventory now 414/1269 contexts (32.6%), 119/626 complete classes (19.0%), 855 pending
contexts across 507 classes; one partial. Queue/pool-writer topic is complete:
2/2 classes and 11/11 methods. All five user-deferred topics stay deferred.
Next: CombinatoricsTest's eleven methods. Read every original method: retain full
62x62 choose/binomial sweep, 63x63 choose/bigChoose exact-long sweep, 185x185
slowBigChoose/bigChoose matrix (n/k=63..247), all eight literal bit-length and
exact-long/decimal cases. Existing Go beyond-table sweep is only 11x11 and must
not substitute for the original 185x185 matrix. No CombinatoricsTest credit yet.


CombinatoricsTest is port complete in combinatorics_java_test.go: all eleven
original methods, full 62x62 choose/binomial matrix, full 63x63 choose/bigChoose
exact-long matrix, full 185x185 slowBigChoose/bigChoose matrix for n/k=63..247,
and all eight literal bit-length and exact-long/decimal methods. Exact-long
conversion checks represent Java longValueExact rather than truncating BigInt;
the three large values retain exact original decimal strings. Related old Go
checks stay intact, but their narrower beyond-table loop is not used for credit.
Production Choose/Binomial/BigChoose/SlowBigChoose/Fact and table initialization
were compared against Java; the operations exercised by this class match and
needed no production change. Java originals pass all eleven (1.319s), Go target
passes (0.647s), target -race passes (4.417s), full TLC normal passes (43.535s).
Retire handles 96593/95625/35067/15646. This slice changes only tests/docs;
production full workspace normal/full TLC race verification remains recorded in
068acee. No tests weakened, matrices reduced or regression/unit cases invented.
Inventory now 425/1269 contexts (33.5%), 120/626 complete classes (19.2%), 844 pending
contexts across 506 classes; one partial. All five deferred topics stay deferred.
Next: complete LongVecTest (eight original methods) and GrowingLongVecTest
(the same eight inherited plus testGrowAndShrink). Source was read, no credit yet.
Keep explicit new LongVec(10) in two inherited methods (do not substitute the
subclass getLongVec hook). Original remove-beyond loop is -1..10 but repeatedly
removes index zero; retain all twelve attempts. Source rangeCheck checks only
index >= elementCount and throws IndexOutOfBoundsException with Index:/Size:
message; negative indices reach the array and throw ArrayIndexOutOfBoundsException,
which the original catch accepts as an IndexOutOfBoundsException subclass.
Current Go rangeCheck's string panic is not that exception family. Fix production
before completing these tests. Growth is source double-capacity/minCapacity,
not Go append's larger-vector growth policy. Negative capacity must not be silently
clamped to zero. Preserve Growing's actual Exception catch in testGrowAndShrink.


LongVecTest and GrowingLongVecTest are port complete in long_vec_java_test.go:
all eight base methods, all eight inherited subclass contexts using its zero-capacity
factory, plus testGrowAndShrink. Preserve explicit new LongVec(10) in the two
original methods for both classes, twelve index-zero remove attempts (-1..10
loop), every value/size assertion and the original catch families. Base catches
accept actual IndexOutOfBoundsException and its array/string subclasses, never
arbitrary Go panics. Growing's extra method retains its broader Exception catch,
rejecting Java Error and foreign Go failures. Original cases/assertions intact;
no invented regression/unit methods or additional persistent vectors.
Production LongVec fixes: negative capacity throws actual NegativeArraySizeException;
Add grows with the source's double-capacity/minimum rule, retaining Java 32-bit
capacity arithmetic; rangeCheck rejects only index >= elementCount and throws
actual IndexOutOfBoundsException with Index:/Size: detail. Negative indices reach
the backing-array access and throw actual ArrayIndexOutOfBoundsException with
backing capacity (not logical size). Last retains direct array semantics. Remove
retains right-hand-side evaluation before destination-array checks; javap of the
current original class confirms laload before lastore. Pack now allocates its
source filtered array and replaces storage, removing the earlier in-place shortcut
(and its short-circuit for size <= 1). Java source methods were compared before
translation; tests were not weakened to accept the former string panic.
Unchanged Java originals pass all 17 (0.010s). Final Go originals plus existing
LongVec checks pass (0.017s), handles 3968/86950 retired. Final full workspace normal passes (handle 1678 retired):
ok  	github.com/glycerine/tlago	313.213s
?   	github.com/glycerine/tlago/cmd/tlago	[no test files]
ok  	github.com/glycerine/tlago/sany_tests	0.992s
ok  	github.com/glycerine/tlago/tlc	44.150s
Full TLC race passes: ok  	github.com/glycerine/tlago/tlc	414.234s
Handle 48906 retired; all verification handles in this slice are terminal.
Low disk space addressed by inspecting and removing only 311 owned regular
hex64-d Go cache data files older than 48h, oldest first (~2.16GB). No source,
fixtures, user data, cache directories or live test processes removed.
Inventory now 442/1269 contexts (34.8%), 122/626 complete classes (19.5%), 827 pending
contexts across 504 classes; one partial. All five deferred topics stay deferred.
Next: MemIntQueueTest's five original methods. Source methods and related Go
checks were read, no credit yet. Retain irrelevant directory/prefix constructor
arguments, source NoSuchElementException catches, zero-long fixture (old Go
long-roundtrip test uses a different number), capacity-four wrap/grow sequence,
all size assertions and only the assertions actually present in the source.
Inspect MemIntQueue/Go IntQueue production before translating; do not credit
old generic-panic expectations as the original NoSuchElementException catches.


MemIntQueueTest and ByteUtilsTest are port complete in int_queue_java_test.go
and byte_utils_java_test.go (five plus six original methods). Queue methods retain
irrelevant directory/name arguments, 4096 default and capacity-four constructors,
source zero-long fixture, all size/value assertions, the precise ring-wrap/growth
sequence and four strict NoSuchElementException catches. Production now throws
actual NoSuchElementException for empty dequeue/pop, rejects negative capacity
with actual NegativeArraySizeException, preserves supplied constructor names,
and follows MemBasedSet's long calculation/int cast/addition growth arithmetic.
Byte methods retain @Before's five 10000-entry arrays and two temp files, all
10000-int/long exercises, 10000 1000-bit BigInts, default unseeded Java Random
instances, half=4999 split, prefix discard before read-until-IOException, all
three append attempts in one IOException-only catch, source console mismatch
branches and millisecond timing lines. Original print-only checks remain print-only;
no extra exception requirement/assertion, smaller samples or substituted fixture.
Input streams are closed in Go test cleanup rather than leaked until Java exit.
Array-object diagnostic identities use native Go array identity in Java's [B@hex
form; those nondeterministic identities are not original assertions.
Production ByteUtils now preserves IOException carriers/messages and IOException-
only catch boundaries, including output failures inside append's loop catch;
known unchecked exceptions are not swallowed as EOF. Negative sized arrays and
short/invalid array access use concrete Java exception carriers. Empty BigInt
bytes retain NumberFormatException with Zero length BigInteger, matching current
JDK bytecode. Removed eager start/finish rejection before source prefix/iteration.
ReadInto performs bounds checks at the actual read, preserving its source no-op
for zero/negative length; source FileInputStream probe confirms nil-message
IndexOutOfBoundsException for invalid offsets. Native stream errors are mapped
to IOException at the io.Reader/io.Writer boundary while represented Java
exceptions retain identity. Existing Go error-message checks remain green.
BigInt's random constructor no longer substitutes math/rand or silently accepts
negative bit counts/nil RNG. It now takes *JavaRandom and follows JDK randomBits:
nonnegative check, ceil bytes, nextBytes only for positive byte count, leading-bit
mask and unsigned magnitude. Zero bits consume no RNG and need no nonnil RNG.
No previous callers of NewBigIntRandom existed; the translated original exercises
now use this actual constructor. JDK javap confirms the control flow/mask. An
ignored source comparison over ten 1000-bit values with seed 15041980 and each
subsequent nextInt is byte-identical Java/Go; it is evidence, not a new test.
Java originals: queue five pass (0.005s), bytes six pass (0.230s). Final Go originals
plus existing related checks pass (0.276s); no diagnostic mismatch lines. Retire
97824/49151/54582/14834. Final full workspace normal passes (root 314.712s, SANY 1.046s, TLC 44.578s);
full TLC race passes (417.363s). Retire handles 16479/38026. Inventory now 453/1269 contexts
(35.7%), 124/626 complete classes (19.8%), 816 pending contexts across 502 classes;
one partial. All five deferred topics stay deferred. No new regression/unit cases.
Next: full BucketStatisticsTest and FixedSizedBucketStatisticsTest parameter rows.
Read constructor matrices and every method before translating; current Java
BucketStatisticsTest has ConcurrentBucketStatistics and BucketStatistics rows.
No statistics-test translation credit yet. Preserve floating-point comparisons,
source exception families, empty/sample cases, percentile/NaN behavior and strings.


BucketStatisticsTest and FixedSizedBucketStatisticsTest are port complete in
bucket_statistics_java_test.go: all ten/six original methods, both original rows
per class (ConcurrentBucketStatistics/BucketStatistics and
FixedSizedConcurrentBucketStatistics/FixedSizedBucketStatistics), fresh per-method
instances, original titles and fixed capacity eight. Retained all samples,
Double.compare equality including signed zero/canonical NaN, source Math.round
long-to-double standard-deviation comparison, duplicate 0.5-percentile assertion,
unasserted out-of-range clamp calls, empty metrics, strict IllegalArgumentException
catches and invocation-only toString checks. No invented assertions or tests.
Original JUnit runs all 32 expanded cases successfully (0.026s); Go original and
existing related checks pass (0.011s). The initial Go run reproduced a real
failure: negative-sample catch saw a string panic rather than IllegalArgumentException.
Production negative sample/NaN failures now carry concrete IllegalArgumentException;
median's unreachable source RuntimeException retains its family/message too.
Source review also fixed percentile's saturating Java double-to-int cast (using
the existing javaDoubleToInt conversion) and logarithmic display loop: compare
integer j directly with the logarithm, rather than truncate the logarithm first.
This retains source fractional-log rounding and literal %n separators. Source
MIT/copyright headers are retained. None of these fixes weakens original checks.
Final full workspace normal passes (root 313.177s, SANY 1.120s, TLC 44.595s).
Full TLC race passes (415.625s). Retire handles 48867/9562. Retire targeted handle 10815 and failing pre-fix handle 72237.
Inventory now 469/1269 contexts (37.0%), 126/626 complete classes (20.1%), 800
pending contexts across 500 classes, one partial. The five deferred topics remain
deferred. Next utility correctness candidates: BufferedRandomAccessFileTest (all
18 methods) and BufferedRandomAccessFileFuzzTest (two methods); neither credited.
Read all setup, fixtures, original randomness/file operations and source exception
contracts before translating. Do not shorten the fuzz sequence or substitute
weaker generic panic expectations for checked Java IOException catches.
Source comparison evidence for statistics (ignored scratch probe, not a new test):
full empty/populated output for all four classes, typed negative/NaN messages and
5-billion-observation percentile with 4-billion/1-billion buckets match Java/Go
byte for byte. Next source files now read in full: BufferedRandomAccessFileTest,
BufferedRandomAccessFileFuzzTest and its AbstractFileState helper; Go buffered-file
production read; main Java buffering/close/length/read/write methods read. Fuzzer
uses available processor count workers with Java Random(threadID), one shared
run counter bounded at 10000 total traces (despite its constant's name), fifty
operations per trace and BOUND=2*8192. Retain rejection sampling, eight operations,
undefined-hole model, signed byte literals, full-read oracle smoothing, minimizer
and failure report. Do not silently change this to Go fuzzing or fewer traces.
No buffered-file test credit yet; source closed-handle catches require actual
IOException with message containing File handle closed, not errors.New alone.


BufferedRandomAccessFileTest is port complete: all eighteen original methods in
buffered_random_access_file_java_test.go. Preserve full 8192/8 long loops, original
seek/length/read/write order, all file modes and temp-name prefixes, source close
scopes, fifteen typed closed-handle catches/message assertions, and all seven
unchanged generated traces. Both seek/no-length catches remain permissive (they
accept IOException but do not require one). Complete signed byte arrays retain
all 5/5/364/701 source literals, mechanically checked against Java; no sliced-down
fixtures or invented assertions. Go wrappers represent Java throw and -1-return
boundaries while executing the actual production buffered file. Ordinary Read's
Go io.EOF sentinel maps to Java read(byte[]...) returning -1; primitive reads must
throw a concrete EOFException. Original IOException catches require represented
Java carriers, not native Go errors with a similar message. Temporary file cleanup
uses Go test cleanup instead of Java process-exit deletion.
Initial original run found errors.errorString instead of IOException after close.
Production now returns concrete IOException("File handle closed"), converts native
file IO failures at the RandomAccessFile boundary while preserving represented
exceptions, and returns actual EOFException on incomplete primitive reads. Existing
Go EOF checks still work through EOFException.Is(io.EOF).
Removed the fresh-buffer allocation shortcut: source synchronized LIFO availBuffs
pool starts at 100 slots and grows by ten; constructor borrows, close pools only
after successful flush, repeat close never pools twice. Finally sets closed before
underlying close; a close failure overrides a flush failure like Java. The original
double-close test now executes actual pooled-buffer behavior. Removed eager
read-only write rejection before buffer mutations; failures occur at actual native
write/length operations. Init uses native stat for source length without moving
the underlying pointer. Native setLength now also moves the disk pointer down to
newLength (os.File.Truncate alone does not), then refreshes diskPos and constrains
curr through the original seek path. Seek checks open at its source boundary;
Seeek retains the source internal no-check behavior. No other test was weakened.
Java originals pass all eighteen with assertions enabled (0.038s); final Go
originals plus existing buffered-file/BitVector checks pass (0.017s). Retire handles
85556/95002/24724/56352. Full workspace normal and full TLC race running with final
production/tests, superseded by a further source failure-path correction; handles 47773/46248
were deliberately stopped (exit 130), not credited. Final handles 12558/52425
now run full workspace normal/full TLC race against the corrected revision.
Inventory now 487/1269 contexts (38.4%), 127/626 complete classes (20.3%), 782
pending contexts across 499 classes, one partial. Five deferred topics stay deferred.
Next: complete BufferedRandomAccessFileFuzzTest and its AbstractFileState helper;
zero credit yet. Retain available-processor workers, per-worker Java Random seed,
shared 10000-trace bound, fifty operations, full BOUND=16384, read smoothing,
rejection sampling and minimizer. Use an independent native RandomAccessFile
adapter for the oracle, not another BufferedRandomAccessFile instance. Source
AbstractFileState uses a 1024-bit initial-capacity hint and tracks defined writes,
not concrete byte values. Disk headroom now about 1GB; full original fuzzer retains
20k temp files until test cleanup. If more headroom is required, inspect disposable
owned Go cache entries and prune only bounded old regular hex64-d data; no source,
vectors or user files. Do not shorten the original fuzzer to avoid storage needs.
An ignored direct-source probe (not a new test) confirms default JVM/Go agreement
for read-only buffered writes followed by flush/close failure, closed reads,
underlying-pointer truncation with logical cursor zero and primitive EOF. The
probe exposed early diskPos mutation: flushBuffer now changes diskPos only after
a complete successful write, and fillBuffer adds the total count after the read
loop completes. Native failure therefore retains source cached state. Original
Java tests also pass with -ea; the probe uses the normal assertions-disabled JVM
because a repeated read-only write failure intentionally violates Java's optional
native-pointer assertion. Go has no optional Java assert instrumentation here;
that distinction is not hidden by a test or comparison normalization. Retire
probe handle 14402 and final targeted handle 30498 (pass 0.017s). No further
production edits after starting final broad handles 12558/52425.
Final broad handles 12558/52425 stopped with build failures caused by disk full
(root link mapping output: no space left on device; race TLC build failure).
These are terminal and not credited as verification. Inspected and then removed
449 owned regular hex64-d Go cache files older than 48h, rechecking inode/device,
size/mtime and ownership against braf-cache-candidates.json before each deletion;
freed 3201707945 bytes (about 3.2GB). Source/vectors and cache directories untouched.
Current final retry handles 37285/46774: full workspace normal/full TLC race, logs
braf-full-normal-retry.log/braf-tlc-race-retry.log; still running, no result yet.
Ignored draft full fuzz translation prepared at
.codex-gotmp/correctness-java/buffered_random_access_file_fuzz_java_test.go.
It is NOT part of the Go test suite, has zero test-port credit, and must be reviewed,
compiled and run before installation/credit. Uses original eight operations,
independent BitSet adapter, independent native file oracle with Java pointer
truncation behavior, per-processor workers, seeds/counter, full reads with null vs
empty distinction, minimizer and testWellDefined's exact literal operations.
Review original Exception|AssertionError catch and Go panic mapping, native
oracle close/error boundaries and thread-count mapping. Retain complete traces,
10000 total runs and all fifty operations; no new invented tests. The authoritative
next action after the eighteen-method green checkpoint is finishing this original
fuzz class, not switching to a smaller statistical or timing surrogate.
Full workspace normal retry is terminal and passes (see braf-full-normal-retry.log);
retire handle 37285. Full TLC race retry handle 46774 still active, no result yet.
Fuzz draft Random calls now use actual NextIntN(int32). Replaced TLC BitVector
storage adapter with independent JDK BitSet union semantics (TLC SetRange overwrites
boundary bits, unlike BitSet.set); draft now uses exclusive ranges preserving prior
writes. Its inner comparison catch rethrows represented Java Error to the outer
Throwable handler. Original Java fuzz class and AbstractFileState compile unchanged
in ignored correctness-java/classes. Draft still uncompiled/uninstalled, zero credit;
review source catch/close/diagnostic semantics and compare generated operations
against actual Java before running full workload. Buffer correctness checkpoint
production/tests remain unchanged during final broad retry.
Final buffered-file checkpoint verification: full workspace normal passes
(root 307.300s, SANY 1.079s, TLC 46.135s); full TLC race passes (416.567s).
Retire retry handles 37285/46774. No tests or production edits after starting
these retry runs. Fuzz draft overlay compiles and its original testWellDefined
passes (0.013s), handle 97370 retired; the full fuzz method remains unrun, zero
inventory credit. Overlay/Java class outputs are ignored scratch evidence.
Next action: finish review and full workload of the draft original fuzz class,
then copy it into tlc/ and update inventory only after verified whole-class parity.


BufferedRandomAccessFileFuzzTest is translated in
buffered_random_access_file_fuzz_java_test.go: both original methods and the full
original harness. One worker per available processor (Java and Go both report 48
here), Java Random(threadID), shared atomic run IDs 1..10000, fifty operations per
trace, BOUND=16384, all eight source operations, original random array lengths/
offsets/lengths and rejection sampling. Independent AbstractFileState BitSet
adapter retains union of prior writes, exclusive range ends, undefined holes,
next-clear-bit criterion and original seek/truncate/read cursor rules. Source
1024-bit hint and dynamic word growth retained. Native unbuffered os.File oracle
implements Java RandomAccessFile EOF and setLength pointer rules; actual production
BufferedRandomAccessFile is the other side. No buffered implementation was used
as the oracle. ReadArray preserves first zero-length read, null vs empty list,
partial-read smoothing at unchanged offset and signed byte result comparison.
All source file operations remain. Minimize retains stride, deletion/order,
unchanged-source append identity cases and wellDefined short circuit. Catch
Exception|AssertionError is separate from the outer Throwable worker catch;
represented Java Error escapes to the outer boundary. Workers retain source atomic
publication and stop behavior; joins complete before failure/minimization reporting.
Go's native goroutines replace source named daemon threads (names are unused by
this original's assertions); all workers are joined, with no smaller processor cap.
Go test cleanup replaces source deleteOnExit, retaining the original two temp
files per trace until the full method finishes. Cleanup preserves first failure
while closing in source reverse order, including second-resource init failure.
No source checks inspect suppressed-exception arrays or unused RunResult hashCode.
Source wellDefined's six literal operations/offsets/length/-39 byte are exact.
Diagnostics retain signed byte lists, source labels and RuntimeException wrappers.
No invented tests, shortened workload, skipped seeds, assertions or minimizer.
Java original all two methods pass with -ea (0.960s), Go initial original run
passes (1.694s), initial complete Go race run passes (11.315s). Parsed all three
logs: exactly 10000 distinct run IDs 1..10000 each; JUnit's first console line is
prefixed by its progress dot, handled in the external evidence parser. Java/Go
actual generation comparisons for all 48 seeds, two fifty-operation traces each,
full operation byte arrays and subsequent nextLong are byte-identical (4848 lines).
These ignored probes are evidence, not added test credit. Final resource-close
review adjusted first-exception precedence after these initial runs; final focused
complete fuzz race passes (11.566s) and full TLC normal passes (46.785s).
Retire handles 61356/81395. Earlier
full TLC normal passes (45.046s); retire 73428. Retire 89777/86028/26326/27325.
No production functionality changed in this checkpoint; preceding full workspace
and full TLC race green baseline remain applicable to unchanged production/root
packages. Final checks specifically cover the added full original test harness.
Inventory now 489/1269 contexts (38.5%), 128/626 complete classes (20.4%), 780
pending contexts across 498 classes, one partial. Utility correctness topic now
complete 14/14 classes, 91/91 methods. All five deferred topics remain deferred.
Next: standard module correctness originals, beginning SequencesTest (full source
read this turn), then module/TLCTest. Inspect production source and preserve full
string fixtures, exception codes/messages, cardinality cases and complete source
function loops; do not credit existing approximate checks before reconciliation.


Standard module SequencesTest and tlc2.module.TLCTest are port complete in
modules_sequences_tlc_java_test.go: all thirteen plus five original methods.
Sequences keeps every exact string/int/tuple fixture, StringValue type checks,
UniqueString.of equality, ten strict EvalException family catches and all original
error codes. It uses the established native-module TLCError EvalException carrier
and actual EvalException where applicable, excludes TLCRuntimeException via the
existing family classifier, then verifies exact code; it does not merely accept
any error like the older approximate Go checks. Source class has no explicit
interner reset, and none was invented. TLC class retains @BeforeClass FP64.Init,
both tuple/record combine cases with normalization before ordered length/array
assertions, both MaxInt32 interval-side cases without extra normalization or
instanceof checks, and complete 1..5 permutation enumeration. Array equality uses
expected.Value.equals(actual) in source order. Source HashSet<Value> uses the
existing test adapter backed by ValueJavaHashCode and actual Value.Equal/identity,
retaining actual keys and hashing each incoming key once; it does not deduplicate
by display strings. Retain input size five, Enumerable assertion, total size120,
per-element size calls and final distinct size120. Both source copyright/MIT
notices retained. No extra assertions or invented cases. No production changes
were necessary: source/Go decomposition inspected for tested operators.
Original Java source Sequences/TLC classes and both original test classes compile
unchanged and pass all eighteen methods with -ea (0.046s). Final Go originals pass
(0.012s). Targeted complete original race passes (1.030s); full TLC normal passes
(44.939s). Retire handles 14050/87604. Retire targeted handles 20406/80410. Production/root
packages remain unchanged since full workspace/full TLC race green baseline.
Inventory now 507/1269 contexts (40.0%), 130/626 complete classes (20.8%), 762
pending contexts across 496 classes, one partial. Utility topic remains complete,
all five deferred topics deferred. Next: complete RandomizationTest (sixteen
methods); source count/location identified, no new translation credit yet.
Preserve all source cardinality/randomness fixtures, original typed exceptions
and codes, full loops and first-versus-second call behavior. Inspect production
Randomization before translating; do required functionality work before moving on.


RandomizationTest: all sixteen original methods are translated in
modules_randomization_java_test.go. Preserve source @BeforeClass seed15041980 and
FP64 initialization; JUnit DEFAULT method order shares one generator across Go
subtest goroutines. All source duplicate zero/negative cases, exact inputs,
cardinality/not-null checks, fresh normalized empty sets, membership and strict EvalException
family/message-contains catches remain. Runtime and other error families escape.
No additional tests/assertions or smaller random workload. Java original unchanged
passes all sixteen with -ea (0.053s); Go initial originals pass (0.154s).
Source audit found raw Go ParseFloat in RandomSubsetSet was a shortcut for Java
Double.valueOf. Production now uses the existing Java literal parser: suffixes,
ASCII whitespace, hex floating literals, signed NaN, syntax and overflow match.
Ignored actual-source Java/Go comparisons agree byte-for-byte for21 probability
strings, resulting sets/cardinalities, error codes/messages and subsequent RNG
nextLong; these probes earn no additional translation credit.
Final full workspace normal passes(root314.209s,SANY0.948s,TLC48.613s); retire20941.
Full TLC race passes(432.814s); retire38242. No checks remain running.
Production final checks cover the probability parser change. Final original-class
race/full TLC normal also cover the normalized empty-set fixture correction. Retire initial Go99562 and literal probe32806. Final source constructor review
corrected all three fresh empty fixtures to normalized=true, exactly new
SetEnumValue(). Complete final original class race passes(1.185s); retire12719.
Full TLC normal rerun covers that final test-only fixture correction and passes
(46.685s); retire19795.
Inventory now523/1269 contexts(41.2%),131/626 complete classes(20.9%),746 pending
contexts across495 classes,one partial. Standard module topic17/44 classes fully
mapped and48/75 methods mapped. All five deferred topics remain deferred.
Next after green checks/commit: presentation MCErrorTest and MCStateTest (two
original methods each), source tests and shared Utils read. Preserve all six
round trips, backward-compatible label space, ordered variable values, exact token
loop and error trace's five original state ordinals1,2,3,5,6. Audit production
against Java before translation; no credit until entire original assertions pass.


Presentation MCErrorTest and MCStateTest: all four original methods are port
complete in model_error_state_java_test.go. Shared original Utils buildState,
toLabelFormat and toTlcOutputFormat loops retained, with ASCII Java trim semantics.
Go's MCState API stores location text: canonical nonempty fixture strings retain
Location.toString values and the empty fixture maps to Location.nullLoc's
"Unknown location". None of the original assertions compares location objects.
MCState retains all six exact round-trip tuples, leading-space compatibility
label, names/flags/ordinals and ordered variable name/value comparisons. Record
printer retains all nine tokens, trim/prefix/substr loop and fail diagnostics;
no stronger final-empty-string assertion was invented. MCError keeps original
message, all five states with ordinals1,2,3,5,6 and x=1..5, map x->y, both loops,
state/variable counts and variable name/single-line display-name checks.
No tests were weakened, no extra cases or production changes. Actual Java
MCVariable/MCState/MCError source and original tests/Utils compile unchanged;
all four pass -ea(0.021s). Go originals pass(0.011s), complete original classes
race passes(1.028s); retire46207/69059. Full TLC normal passes(44.945s);
retire55131. No checks remain running. Full workspace normal and full TLC race
from preceding
bf642ba remain green for unchanged production/root packages.
Inventory now527/1269 contexts(41.5%),133/626 complete classes(21.2%),742 pending
contexts across493 classes,one partial. Presentation topic complete5/5 classes,
13/13 methods. Utility/value topics remain complete; all five deferred topics
remain deferred. Next: CLI/output originals. Full REPLTest(testProcessInput)
and MPTest(four methods) read; preserve all thirteen REPL calls on one instance,
exact empty-string results for invalid expressions, temp-directory construction,
per-method ToolIO TOOL/reset, exact message counts and overload substitutions,
all six progress parameters and both original locale alternatives. Inspect actual
production console boundary before translating; implement missing behavior before
crediting output methods. Root REPL translation may require full root checks.


MPTest: all four complete original output methods translated in
output_mp_java_test.go, with per-method ToolIO TOOL/reset, actual production
println capture, exact message counts/overload substitutions, six MP.format long
parameters and both original comma/dot locale alternatives. No recorder-to-string
substitute, extra assertions/cases, or weakened expected output. Source MP and
ToolIO classes plus original MPTest compile unchanged; Java -ea passes(0.045s),
Go initial complete class passes(0.011s), constructor/final follow-up passes(0.012s).
Production audit found recorder-only PrintMessage/PrintError and absent UNIT_TEST
template. Added real ToolIO system/tool output, shared out/err buffering, exact
print/println message boundaries, reset, mode validation, defensive reads and
explicit array doubling. Preserve source getAllMessages arraycopy bounds failure
when pending text follows an exactly full buffer, including concrete exception
and native arraycopy detail. Source Java bug is not repaired or hidden. Console
printing retains recorder events separately, source severity prefixes/tool
STARTMSG/ENDMSG envelopes, warnings' history/hint/suppression timing, PrintState's
rendered return, GetMessage's tool envelope and source timestamp/noTimestamps
format. UNIT_TEST uses sequential repeated substitution; absent and unused
parameters remain original. Six-parameter progress branch requires exactly six.
CLI runTLCModelCheck connects supplied system streams and restores them afterward.
MP.format's explicit ###,###.### pattern now groups by three with localized digits,
separators and negative affixes, preserving every signed-long bit (including MIN).
Raw Go CLDR defaults were not equivalent. Retained actual OpenJDK21.0.12.1 locale
symbols/affixes for1068 available locales and66 CLDR numbering keywords in generated
output_decimal_symbols.go, deduplicated into41 complete symbol sets. Existing
OpenJDK notices remain referenced. Default/category properties, script casing,
legacy language aliases/Thai variant, Unicode nu vs private-use distinction,
unknown/multipart numbering fallback and locale parent lookup represented. Cached
process-default symbols match source default Locale setup; no Go Locale.setDefault
API is asserted. Source MPDecimalSymbols/MPNumberingSymbols extraction stays in
ignored scratch as audit provenance, not runnable original test credit.
Actual Java/Go integral comparisons across1068 locales,11 values each, including
both signed64 limits match byte-for-byte(11748 lines). Direct actual-source console
probe matches all original operations in plain/tool modes plus partial/multiline
buffer boundaries and concrete exactly-full-buffer failure. Those ignored source
probes add no translation credit. Go table/core were reviewed after this evidence;
final focused original race passes(1.029s); retire78257. Final broad checks
remain to confirm final production/root edits.
Old broad handles38530/49744 deliberately stopped(exit130) for the source buffer
fix;5538/52540 deliberately stopped(exit130) after source script-case review.
Retire10129/62638/69488/50323/72001/24553/86476/68165/70274; all terminal.
Final full workspace normal passes(root312.323s,SANY1.046s,TLC48.526s); retire79366.
Full TLC race passes(428.026s); retire82520. No checks remain running.
Final original Java/Go German-locale runs pass too(Java0.048s,Go race1.028s);
retire28574. Additional actual-source comparisons of66 numbering systems in four
locales plus eight Unicode/private-use/script forms match(2992 lines). Those
ignored probes earn no extra translation credit. Final production/root is green.
Disk: inspected owned regular exact hex64-d cache entries older48h; pruned272
files2006160721 bytes, manifest mp-cache-candidates.json. Second inspected prune
187 files1506045376 bytes, manifest mp-cache-candidates-2.json. No source/vector/user
files or directories were removed.
Inventory531/1269 contexts(41.8%),134/626 complete classes(21.4%),738 pending across
492 classes,one partial. CLI/output topic1/5 classes complete,4/37 methods mapped.
All five deferred topics stay deferred. Next after green commit: WarningControlTest(all eighteen methods), then
remaining CLI/output originals including REPLTest(testProcessInput). Full original
WarningControlTest and current HandleParameters/message-control APIs read. Direct
HandleParameters currently validates/registers TLC codes; SANY codes are handled
only by root CLI extraction and require actual integration before crediting the
original direct-call methods. Preserve source @After reset/warn=true, every
registration/empty-set assertion, all conflicts, original dummy Model filename,
typed runtime exception expectation and real captured stdout/err. Native stream
assignment must reflect source ToolIO.out/err overriding buffering even if mode
remains TOOL; current SystemStreams API sets only native-system backing streams.
Add that real source capability when translating the direct runtime captures,
rather than bypassing the print boundary in the tests.
REPL audit shows original expects an instance processInput boundary with thirteen
calls; Go currently has EvaluateREPLExpression only and older generated EXTENDS
list, missing Randomization. Preserve original constructor/temp dir, generated
source, caught exception/message behavior and full default modules; fix actual
production before crediting the method. Do not simply accept Go errors in a test
adapter instead of porting the original REPL processInput behavior.


WarningControlTest correctness checkpoint (2026-10-03):
All eighteen original methods translated in output_warning_control_java_test.go,
including every direct HandleParameters result, TLC/SANY set-membership/empty
assertion, source @After reset/warn=true, native stdout/err capture and strict
runtime exception expectation. Java's boolean result maps to Go's error return;
its TLCRuntimeException maps to the established TLCError with Runtime=true. Other
panic families are rethrown. Source dummy MC and numeric arguments remain exact;
only its relative metadata directories are isolated under the test temp cwd.
Production now registers SANY and TLC controls separately, exposes read-only
backed set views, resets without replacing the views' backing sets, validates
both kinds of conflicts and bridges SANY registrations to model-check parsing.
Native ToolIO stream assignment overrides TOOL capture without changing mode.
HandleParameters emits its once-only welcome and actual MP command-line errors.
Numeric code parsing retains Java comma-split/trailing-empty, ASCII trim and
Integer.parseInt semantics. Audited actual EC declared fields: remove nested exit
status values10/75/150/255, include declared7001..7010; GENERAL1000 is TLC first.
SANY overlap errors retain actual ErrorCode enum names. Source enum MIT notice
retained. No deferred distributed implementation work is added.
Unchanged actual Java TLC/MP/ToolIO/EC/ErrorCode and original JUnit eighteen pass
with -ea(0.085s); final Go class plus original MP four pass(0.013s), focused race
passes(1.033s). Eleven direct Java/Go code-registration/error-output cases match
byte-for-byte after omitting only the runtime build greeting line; these ignored
probes add no test credit. Full workspace passes(root310.325s,SANY0.890s,
TLC47.877s); retire61045/29540. Full TLC race passes(426.282s); retire68918.
No checks remain running.
Inventory549/1269 contexts(43.3%),135/626 complete classes(21.6%),720 pending
across491 classes,one partial. CLI/output2/5 classes,22/37 methods complete.
Next after green commit: remaining CLI/output originals (TLCTest ten, REPLTest
one, SpecTraceExpressionWriterTest four). All five deferred topics stay deferred.


TLCTest correctness checkpoint (2026-10-03), final verification green:
All ten original tlc2.TLCTest methods translated in cli_java_test.go. Preserve
exact -fpmem inputs, minimum/75%/50%/99% assertions and heap-allocation assumptions,
all seven maxSetSize validations/global checks, three simulation file/num calls
and six exact runtime conversions. Temporary cwd isolates MC metadata; parent
cleanup restores class-owned globals. No invented permanent tests.
Production fixes Java float-to-long saturation for Long.MAX_VALUE, full
Double.parseDouble literal parsing, source deprecation output and invalid-ratio
failure. maxSetSize uses Java Integer.parseInt and actual error templates.
Runtime conversion retains strict millisecond thresholds, UTC calendar fields
(day-of-year/year wrap) and FORMAT-locale digits, including negative/signed-long
input. Source duplicate24h branch selects the same pattern.
The half-memory assertion exposed a shortcut: Runtime.MemStats.Sys measured
changing reserved memory, not a maximum. Configured Go GOMEMLIMIT/runtime limit
now supplies the budget. When unconfigured, native physical/container capacity
supplies a stable quarter-memory budget, applied through debug.SetMemoryLimit.
Linux honors cgroup-v1/v2 membership/mount roots/visible ancestor hard limits;
Darwin/Windows use native physical-memory APIs; other platforms require
GOMEMLIMIT. The Go runtime limit is soft; it is not a JVM hard allocation cap.
Full native checks then exposed original TestJavaMultiFPSetCTorMax requesting
2^30 children: normal timed out in that constructor; race ended with a
ThreadSanitizer address-restoration failure after heavy allocation, with no
preceding race report. Go now raises the existing concrete OutOfMemoryError for
a provably unaffordable mandatory object graph, using native interface/config/
DiskFPSet sizes against the budget. No fixed child-count cap or changed test
inputs/assertions. Temporary configs for the four nil-returning factories are
excluded. Java getNestedFPSets creates a new config each iteration; Go now copies
an independent configuration per child. The table representation was a
preliminary hypothesis only and was not changed.
Unchanged original JUnit ten passes(-ea0.118s; -Xmx256m0.121s); complete Go ten
passes0.014s, focused race1.033s, configured256MiB race1.045s. Complete original
MultiFPSet/TLC translations plus existing factory-load check race passes37.261s.
Unchanged original MultiFPSetTest.testCTorMax compiled and JUnit Request executes
it(-ea -Xmx256m): run1/failures0. Darwin/Windows amd64 builds pass.
Actual Java -Xmx256m/Go GOMEMLIMIT=256MiB literal/code/error probes(seventeen
cases) and runtime boundary/signed-long/year-wrap probes(nineteen cases) match
byte-for-byte, also in Arabic locale. Omit only build greeting lines in ignored
probes; no original assertion normalized, no extra translation credit.
Final full workspace passes(root303.042s,SANY0.977s,TLC62.059s), retire42720.
Final full TLC race passes545.514s; retire74112.
Log .codex-gotmp/correctness-java/tlc-cli-full-race-oom-final.log.
All final broad gates are green; no checks remain running.
Prior49328/93295 were terminal failures;37219/32211/91020/23607 deliberately
interrupted for production corrections; focused99827 is terminal green.
Inventory559/1269 contexts(44.1%),136/626 complete classes(21.7%),710 pending
across490 classes,one partial. CLI/output3/5 classes,32/37 methods complete.
All five deferred topics remain deferred.
Next after green commit: SpecTraceExpressionWriterTest four and REPLTest one.
Full trace-writer class source reviewed: four actual SANY front-end acceptance
checks plus named-Formula assertion. Exact temp-file setup, preamble/init/next,
two MC error states, two-buffer ordering and multiline/comment expression strings
prepared in ignored spec-trace-writer-draft.go. It is NOT activated or credited.
Port source AbstractSpecWriter.appendContentToBuffers(nullable independent
TLA/CFG strings), needed by root-package SANY integration; TLA-only Append is
insufficient. Source says "Provided for test code". Existing Go
SpecTraceExpressionBuildInitNextBuffers already returns the required two builders;
use it directly. Do not use concatenating FullToBuffers or bypass actual file
writes and full parse/semantic conclude checks. Reconcile draft with source,
port helper, activate/run/repair original methods after this green checkpoint.
Disk: two inspected owned regular hex64-d cache prunes older48h,218 files
1818337811 bytes and253 files1606262853 bytes. Manifests
 tlc-cli-cache-candidates.json and tlc-cli-cache-candidates-2.json; native
identity/type/uid/time revalidated before each unlink. No source/vector/user
files or directories removed.


SpecTraceExpressionWriterTest checkpoint (2026-10-04), final verification green:
Green previous checkpoint committed f6eb8f3; working tree was clean. All four
original writer methods now translated in spec_trace_writer_java_test.go with
exact independent TLA/CFG temp setup, preamble/init/next/two MC error states,
first-buffer/trace-function/second-buffer ordering, both trace-expression lists,
multiline swallowed-comment alternatives and original named-Formula assertion.
Actual generated files feed the legacy front-end; no string-only replacement.
Ported original AbstractSpecWriter.appendContentToBuffers(nullable strings)
in production, exported for root-package parser integration. Each non-null
buffer append remains independent and in source order; nil cfg storage throws
the established NullPointerException rather than silently skipping it.
Important source fidelity: SANY.frontEndMain defaults doStrictErrorCodes=false.
Syntax ParseException produces ERROR(-1); ordinary semantic diagnostics are
retained but still return OK(0). The unactivated draft had added a stricter
semantic-error-count assertion absent upstream; final test instead checks the
exact original return-code condition. Production SanyFrontEndMain runs existing
real parse and semantic phases and retains diagnostics on Spec; unexpected
runtime failures propagate. Syntax-invalid input returns the original ERROR(-1).
Actual Java/Go ignored legacy probes(valid, syntax-invalid, ordinary undefined
symbol) match byte-for-byte:0/errors=false,-1/errors=true,0/errors=true. They add
no new permanent tests or translation credit. Source test fixtures untouched.
Unchanged original Java JUnit four passes(-ea0.149s); complete Go four passes
0.021s and race1.075s. Retire81558/5769. Full workspace normal passes(root296.058s,SANY1.006s,TLC62.303s); retire4909.
Log .codex-gotmp/correctness-java/spec-trace-writer-full-normal.log.
No checks remain running. Prior full TLC race baseline
545.514s(f6eb8f3) remains verified: new helpers are only called by this new test
and do not change existing production paths. No full TLC race repetition needed
for these additions; root class race and full workspace cover current changes.
Inventory563/1269 contexts(44.4%),137/626 complete classes(21.9%),706 pending
across489 classes,one partial. CLI/output4/5 classes,36/37 methods complete.
Current writer checkpoint has passed its full workspace gate.
Next is REPLTest.testProcessInput(all thirteen calls/one instance); prior source
notes remain applicable. All five deferred topics remain deferred.

REPLTest checkpoint (2026-10-04), final verification green:
Previous green checkpoint a9f9277. Original REPLTest.testProcessInput now in
repl_java_test.go: all thirteen source calls, exact expressions/results and
one shared instance, no replacement sample cases or added permanent tests.
Production REPL.ProcessInput runs the actual existing SANY/checker/tool path.
Corrected default EXTENDS to Reals,Sequences,Bags,FiniteSets,TLC,Randomization,
CFG-before-TLA writes, no automatic creation of caller-owned REPL tempdir,
ToolIO TOOL/reset after both writes, optional native CommunityModules index
provider and exact spec basename/.tla handling. Source EvalException and
TLCRuntimeException catches remain narrow; unrelated exceptions/Java Errors
propagate. Catches/finally apply only after generated files are written, like
Java's inner try. SANY diagnostic parameters are retained, with source location
format and Unknown operator presentation for native undefined-name diagnostics.
Source buffered PrintWriter captures stdout at construction, does not auto-flush
and retains IO failures internally; finally flush runs after reporting errors
and TLCOutput becomes nil rather than restoring the prior writer.
TLC.Print/PrintT now use ToolIO when output is nil; non-null PrintT writers get
no added newline. Existing convenience EvaluateREPLExpression retains tempdir
management/output restoration, while using corrected source module imports.
Original Java JUnit one passes(-ea0.969s); thirteen-call Go test passes0.549s.
Nine ignored Java/Go boundary probes match byte-for-byte(print, print followed
by eval failure, repeated calls, type errors, singleton RandomElement, invalid
names). Probes are diagnostic evidence only, no permanent-test/translation credit.
Full normal draft checks1946/57825/92375 and race draft91701 were intentionally
interrupted for actual source-boundary corrections, not timeout failures.
Final full workspace passes(root304.652s,SANY1.061s,TLC62.020s);
retire94663. Full TLC race passes562.312s; retire16435. Final focused root
REPL race passes3.314s; retire98031. No checks remain running.
Logs .codex-gotmp/correctness-java/repl-{full-normal,full-race,race}.log.
Earlier focused handles21064/22321/83073/44470, Java89214 and probe4928 terminal.
Inventory564/1269(44.4%),138/626 classes(22.0%),705 pending across488 classes,
one partial. CLI/output all5/5 classes37/37 methods port complete.
Final checkpoint has passed its full workspace and race gates. Next correctness
models are original
ConstantContextTLCCacheTest and TLCExtTest (one method each, four recorder
assertions plus inherited successful exit setup/teardown). Existing root
runJavaTLCModelTest helpers support original settings; TLCExtTest overrides
runWithDebugger=false and passes -config TLCExtTest.tla. Their original fixtures
are in ../tlaplus/tlatools/org.lamport.tlatools/test-model/, not yet copied or
ported. Native cache microchecks alone do not earn those original model credits.
All five deferred topics remain deferred; no distributed or email work.

Next-class Java compilation already succeeded in ignored scratch: original
ConstantContextTLCCacheTest/TLCExtTest plus unchanged test harness dependencies,
-sourcepath upstream test directory, CP classes+tla2tools+JUnit+Hamcrest.
Log tlc-ext-model-javac.log. Do not run them with upstream BASE_PATH directly:
TTrace getTESpecOutDir writes under BASE_PATH/generated, so first copy exact
source models into a writable scratch test-model directory and set basepath.
No next-class Go test or fixture has been created yet.

Standard-module correctness models checkpoint (2026-10-04), final verification green:
Previous green checkpoint 0e4a0f6/bc2cc84. tlc_module_models_java_test.go now
mechanically translates nine original concrete classes/methods:
ConstantContextTLCCacheTest.test, TLCExtTest.test, BagsTest.testSpec,
ConstantRank1TLCEvalTest.testSpec, ConstantRank2AssertErrorTest.testSpec,
EmptySetEqAssumeTest.testSpec, EmptySetEqStatesTest.testSpec,
EmptySetEqStatesRcdTest.testRcdSpec, KSubsetAssumeTest.testSpec.
All source recorder assertions and inherited successful exit retained. Source
model/config/dependency files copied to tlc/test_vectors/models; all16 files
verified byte-identical to upstream, with complete assumption lists unmodified.
Class setup: TLCExt debugger=false and embedded .tla config; constant-rank
classes noGenerateSpec=true/doDumpTrace=false; emptiness/kSubset classes
noGenerateSpec=true but keep default JSON trace dump. State variants keep
checkDeadLock=true and complete assertZeroUncovered. Helper now has explicit
checkDeadlock option and omits -deadlock for those two classes; existing callers
retain their prior behavior. KSubset gets original CommunityModules archive on
native CLASSPATH, matching original Java tool manifest; do not replace module
source with resolver.go's convenience placeholders.
Actual new EmptySetEqAssume failed success exit at model line476. Source
SetOfTuplesValue.member starts with elem.toTuple(), then elem.toFcnRcd(). Go had
used only concrete tuple/FcnRcd assertions, wrongly rejecting valid functions.
Production now calls existing complete asTupleValue/asFcnRcdValue converters.
Next failure534 exposed finite-only guards in product IsEmptyValue. Java Value
isEmpty scans each component directly in order, so nonenumerable differences
raise errors before later emptiness can decide. Removed the guards only from
isEmpty; Java isFinite still has its own appropriate finite guards.
Next failure544 exposed entirely missing SetPredValue isEmpty branch. Added
source elements().nextElement() plus native enumeration-error propagation.
Next failure603 exposed incorrect default-isEmpty message; source uses exact
Shouldn't call isEmpty() on value ANY. Default UserObj uses its own raw String,
while Value default uses ValuesPPR, preserving source message construction.
Next failure690 was native MethodValue metadata abbreviated as ToString. Added
actual full public static Value TLC.ToString(Value) signature, preserving normal
reflective error wrapper rather than weakening AssertError or changing fixtures.
Complete EmptySetEqAssume now passes, including every remaining exact error
assertion through end of source model. State variants and KSubset all pass too.
No invented permanent tests: only original class/method translations. Ignored
empty-set-probe confirmed original failed ToString wrappers, no test credit.
Original unchanged Java JUnit all9 pass(-ea -Xmx256m), separately isolated
JVMs and writable scratch basepath. Logs tlc-cache-model-junit.log(.434s),
tlc-ext-model-junit.log(.424s), empty-set-assume-junit.log(.716s), and
original-{BagsTest(.717s),ConstantRank1TLCEvalTest(.384s),
ConstantRank2AssertErrorTest(.382s),EmptySetEqStatesTest(.470s),
EmptySetEqStatesRcdTest(.463s),KSubsetAssumeTest(.571s)}-junit.log.
The first source runs failed due to an uncompiled unchanged TTrace harness
class, not model behavior; explicitly compiled that original support class.
Final combined root race passes35.002s; retire4302. Older draft normal83364 and
TLC race94275 intentionally interrupted for final source UserObj formatting
correction. Final full normal passes(root306.199s,SANY0.943s,TLC64.204s); retire12067.
Full TLC race passes581.216s; retire46067. No checks remain running. Logs standard-models-full-normal
and standard-models-full-race under .codex-gotmp/correctness-java/.
Inventory573/1269(45.2%),147/626 classes(23.5%),696 pending across479 classes,
one partial; standard-module topic26/44 classes57/75 methods,18 pending.
Mapped JSON independently recomputed with source=test,path starts tlc2/,
nonabstract/nonignored/active filter; totals match document. Appendices excluded.
Final checkpoint has passed its full workspace and race gates. All five deferred topics stay
outside new work; no email/distributed work.
Next source models inspected: RandomSubsetATest/BTest inherit RandomSubset
one full method, with seeds15041980/918347981374 and exact two-state traces,
8008 initial/8009 generated+distinct/8007 queued. RandomSubsetTest includes
all trace value/type/tuple bounds and zero-uncovered, and an upstream typo:
its y bound assertion is firstY>=100000000 && firstX<=100000010; preserve that
literal original condition instead of inventing a stronger y upper bound.
RandomSubsetNextTest retains full67291/7729/999 counts and eleven exact trace
states. RandomSubsetNextT4Test requires4 workers and eleven trace states with
range/ordinal checks, not exact scheduler-dependent x values. Tuples variant
5461/5461/4095 and seven complete state/tuple checks. SetOfFcns variant keeps
1000 initial/2000 generated/1000 distinct/0queued/depth1/zero-uncovered;
commented-out upstream assertions are not active requirements.
No next-model fixtures or Go tests have been added yet. Candidates SetPredValue,
StandardModules and RandomSubsetEmpty retain complete original zero-state
model assumptions and source recorder checks. Do not skip random/multithreaded
correctness simply because distributed tests are deferred.

Random-subset correctness models checkpoint (2026-10-04), final verification green:
Previous green checkpoint4d028ab. New tlc_random_subset_models_java_test.go
ports all8 original RandomSubset concrete contexts: Empty, A, B, base Test,
SetOfFcns, Next, NextT4, NextTuples. Original source fixtures/configs copied
unchanged;14 new files across7 model dirs including SetPredValue/StandardModules.
The two latter original model tests were appended to tlc_module_models_java_test.go.
No production code changed in this checkpoint; existing random enumeration,
Java-seed/BigInteger product paths and previous value fixes satisfy all cases.
Every source assertion retained: fixed seeds15041980/918347981374, exact A/B
trace values,8008 initial/8009 generated+distinct/8007queued/depth2; base trace
checks5 variables,both integer bounds,booleans,all3 tuple components,UNCHANGED
values,initial/action metadata and zero-uncovered. Source y-bound typo retains
firstY>=100000000 && firstX<=100000010 literally; do not strengthen it.
Inherited CommonTestCase.assertTraceWith fully translated(length,initial/action
conditions,exact trimmed states,all ordinals). Java ASCII String.trim semantics
reuse existing root replJavaTrim; depth parsing uses signed32-bit ParseInt.
One-worker debugger setup selects extended metadata in these original classes;
helper branches remain explicit. No new debugger functionality or skipped-topic
work introduced. Four-worker case keeps4 workers,eleven trace y0..10,x1..1000,
ordinals and source absence-of-BUG condition; no deterministic x assertion added.
Next exact67291/7729/999 stats and all11 source trace states retained. Tuples
exact5461/5461/4095,initial4,trace7 and every p/q component range retained.
SetOfFcns initial1000/generated2000/distinct1000/queue0/depth1/zero-uncovered;
commented-out source assertions remain inactive. Empty model's complete
assumptions retained; predicate model evaluates and prints original LP; standard
module model uses full original EXTENDS list/config. All successful exit and
safety-violation exit expectations include inherited source teardown assertion.
All new defaults keep coverage/dot/JSON trace/generate-spec setup; Empty only
has source noGenerateSpec override. No sample matrices substituted.
Original unchanged Java JUnit all10 pass(-ea,-Xmx256m,writable scratch basepath),
separate JVMs. Logs original-{RandomSubsetEmptyTest(.537s),RandomSubsetATest
(1.349s),RandomSubsetBTest(1.279s),RandomSubsetTest(1.400s),
RandomSubsetSetOfFcnsTest(.527s),RandomSubsetNextTest(1.134s),
RandomSubsetNextT4Test(.776s),RandomSubsetNextTuplesTest(.847s),
SetPredValueTest(.651s),StandardModulesTest(.457s)}-junit.log.
Go normal individual checks all pass: Empty1.101s,A/B21.950s,base/fcns12.624s,
Next14.670s,T4/Tuples27.036s,predicate/standard3.866s.
Full8-context root random-model race passes160.534s; retire41576.
Other2-context root race passes11.161s; retire84010. No race remains running.
Full workspace normal passes(root363.854s,SANY0.946s,TLC62.365s); retire70843.
Log random-models-full-normal.log. No checks remain running. Retain verified full TLC race
581.216s baseline4d028ab: this checkpoint changes root tests/fixtures/docs only,
no TLC production or TLC package tests. Full TLC race repetition unnecessary.
Inventory583/1269(45.9%),157/626 classes(25.1%),686 pending across469 classes,
one partial. Standard-module topic36/44 classes,67/75 contexts,8 pending.
Mapped JSON counts independently recomputed and match document. All14 copied
files verified byte-identical, including original whitespace. Do not edit away
upstream trailing whitespace in persistent vectors for git diff --check.
Next8 standard-module contexts: RandomElementTest,RandomElementT4Test,
RandomElementSimulationTest,RandomElementXandYTest,SubseteqNextStateTest,
UserModuleOverrideAnnotationTest,UserModuleOverrideFromJarTest,UserModuleOverrideTest.
First4 random-element source seeds/traces ready to inspect; XandY seed
8006642976694192746 has exact3-state trace0/0,1/1,0/1+zero-uncovered. Simulation
seed8006803340504660123,-simulate num=1,debugger=false,exact11 states/actions,
zero-uncovered; preserve source action-aware assertTraceWith overload.
SubseteqNextState original68/8/0/depth2 and successful exit. User override jar
class requires original customBuild extra classpath. Both unannotated override
classes assert sorted2 mismatch records(exact Get2/Get3 method strings/path suffix),
2/1/0 stats,success,zero-uncovered. Annotation class calls three recorder.recorded
methods without asserting their results; retain actual active assertions only.
No next-context Go code/fixtures added yet. All five deferred topics remain
outside new work. Current checkpoint has passed its complete normal gate.


## 2026-10-04: original RandomElement and subseteq next-state tests

Current checkpoint adds original RandomElementTest, RandomElementT4Test,
RandomElementSimulationTest, RandomElementXandYTest (test), and
SubseteqNextStateTest (testSpec). Root files tlc_random_element_models_java_test.go
and tlc_subseteq_next_state_java_test.go. Six original TLA/CFG files copied
byte-for-byte to tlc/test_vectors/models; no production code changed.
Every original constructor/inherited successful or safety-violation exit,
coverage/dot/JSON/generate-spec setup and active assertion retained. Seeded
single-worker model checks exact932/855/388 and all11 states/ordinals/actions.
Four-worker model retains4 workers, source seed15041980, all11 y values,
x1..1000 and ordinals without adding scheduler-specific expectations.
Simulation retains debugger=false, -simulate num=1, seed8006803340504660123,
all11 exact states and exact Init/Next source location labels plus ordinals.
XandY seed8006642976694192746 retains exact3 states0/0,1/1,0/1.
Zero-uncovered checks retained where source asserts them; four-worker source
has none. Subseteq next-state retains68/8/0 and depth2; original postcondition
and _POSSIBLE model configuration unchanged.
Unchanged original Java JUnit classes each pass in separate JVM(-ea,-Xmx256m):
RandomElementTest0.522s,T4Test0.465s,SimulationTest0.352s,XandYTest0.453s,
SubseteqNextStateTest0.517s. Initial scratch run lacked flat model files and
exited file-not-found; corrected scratch layout only, then all original runs
passed. No Java source or persistent fixture was modified.
Focused all-five Go root race passes16.691s, random-element-subseteq-race.log;
retire41918. Normal new root checks pass(first1.742s,other3 3.397s,subseteq1.106s).
Full workspace passes(root366.485s,SANY0.914s,TLC59.849s); retire47901.
Log random-element-full-normal.log. No checks remain running.
Retain full TLC race581.216s baseline4d028ab; this checkpoint
changes root tests/fixtures/docs only, no TLC production/package tests.
Inventory588/1269 contexts(46.3%),162/626 complete classes(25.9%),681 pending
across464 classes, one partial. Standard-module topic72/75 methods,41/44 classes,
three pending. Mapped JSON counts independently recomputed and match document.
Remaining native-override classes: UserModuleOverrideTest,
UserModuleOverrideFromJarTest,UserModuleOverrideAnnotationTest. Inspected source
but no Go test/fixture/production changes for these yet. Bridge current native
override installer uses fixed module/member map and silently skips absent/wrong
arity native methods. Implement actual native module discovery/metadata and
source mismatch diagnostics before port credit; never substitute TLA inline
bodies/fake recorder entries. Source Get returnsTrue while TLA bodyFALSE, so
passing requires actual native callback. Get2/Get3 full original method strings
and class path suffixes required by both unannotated classes. Jar test custom
classpath setup must retain source semantics. Annotation source calls three
recorder.recorded methods without asserting them; preserve active assertions.
All five user-deferred topics remain outside new work. Goal remains active.


## 2026-10-04: native user-module discovery and original override tests

This checkpoint implements native Go linkage/discovery for user module classes
and the source SpecProcessor.processModuleOverrides registration algorithm.
Production files tlc/native_class.go and tlc/spec_processor_native.go; bridge
invokes processing after existing standard native bindings. Spec retains SANY's
FilenameResolver so class resources use the actual resolver/classpath. Source
reflection metadata is carried with linked Go callbacks through NativeClass and
NativeMethod descriptors. RegisterNativeClass returns binding-restoration closure.
RequireResource preserves source class-file lookup for legacy/native adapters;
Go implementations are linked callbacks, never execution of Java bytecode or
inline replacement TLA bodies. Test-class adapters live only in root *_test.go.
Class resource lookup follows resolver, classpath, package fallback. Resolver
catch(Exception) excludes Error; outer catch(Throwable) emits replacing-modules
failure. Registry accesses are locked; declared method slices copied on link/load.
Conventional methods enumerate all public static declarations, TLARegistry map,
root-origin arity map, annotated-method skip, missing-name/arity warnings,
loaded metadata and actual definition/body installation. Final zero-argument
methods evaluate eagerly at registration before mismatch checks. Ordered native
indexes retain built-in index name then source property/default/path separator.
Annotation order Evaluation,Callable,Operator, module/identifier checks,
minLevel/warn/silent/priority composition and source arity diagnostics retained.
Index catches direct InstantiationException/IllegalAccessException only, not
wrapped causes. MP now has exact source loaded/three mismatch message templates.
Architecture details in TLC_ARCH.md. Arbitrary Java binaries still require their
Go callback port to be linked; do not interpret resources as executable code.

Original UserModuleOverrideTest, FromJarTest, AnnotationTest.testSpec ports in
root tlc_user_module_override_java_test.go. Source Get returnsTRUE despiteFALSE
TLA body, proving actual native use; source Get2(v1)/Get3 returnFALSE. Both
unannotated tests retain sorted two mismatch records and exact source signatures,
real class resource suffixes,2/1/0,FINISHED,noGENERAL,zero-uncovered,success.
Annotation class retains source unasserted recorded calls, actual active success/
stats/coverage assertions; all five original annotated fixture methods ported.
Native index fixture TLCTestOverrides.get retains EvaluatingValueTest then
UserModuleOverrideAnnotationImpl; source property/customBuild classpath retained.
Original jar copied unmodified and still supplies its TLA module/class resource.
Jar bytecode inspected with javap: GetTRUE,Get2(Value)FALSE,Get3FALSE.
Root tlc_native_override_fixtures_java_test.go links fixture methods and metadata
through production loader. Native callbacks/metadata do not emit test records.
Evaluation fixture keeps synchronized static action via mutex and binds successor
x42 then returnsTRUE. Original EvaluatingValueTest.testSpec port in root
 tlc_evaluating_value_java_test.go: full3/2/0/depth2,FINISHED,noGENERAL,
noTEMPORAL_PROPERTY_VIOLATED/noCOUNTER_EXAMPLE,successful inherited exit.
Original TLA action x'=FALSE remains unchanged; native callback genuinely replaces
it. No invented checks/coverage requirements or relaxed model inputs.
Eight TLA/CFG files, one resolver-local class resource, and seven native fixture
files copied byte-identically from upstream (16 files total). Includes original
Java fixture sources, class resources and UserModuleOverrideFromJar.jar.
Evaluation callback's class is compiled Go linkage (source JUnit classloader's
compiled test-class counterpart), resource URI go:tlc2.tool.EvaluatingValueTest;
its method signature/default priority100/warntrue retained. No original assertion
depends on its resource path. All source jar members/hash inspected.
Original unchanged Java JUnit passes separately(-ea,-Xmx256m,source test-index
property and actual native-fixture/jar classpath): UserModuleOverrideTest0.418s,
FromJarTest0.444s,AnnotationTest0.418s,EvaluatingValueTest0.409s.
First normal Go checks pass2.007s conventional,3.592s jar+annotation,1.865sEV.
Focused four-context root race passes17.314s after final exception corrections;
retire42060. Earlier16.847s focused root pass predates final catch correction.
Initial full normal81341 and TLCrace94843 explicitly terminated(exit130) for
actual production catch corrections, not an observation timeout. Do not revive.

During long gates, ported original EvalControlTest.test and testIfEnabled in
 tlc/eval_control_java_test.go; all12+3 assertions retained, no production changes.
Original unchanged Java JUnit2methods pass0.005s; Go focused race1.029s pass,
retire74847. Existing eval_control_test.go only had unrelated PartialBoolean
checks, so prior Reconcile row now genuinely completes original methods.
Full workspace after final production changes passes(root368.645s,SANY1.051s,
TLC64.443s); retire55619, native-override-full-normal.log. It compiled before
EvalControl tests were added. Final full workspace includes those two and passes(root371.729s,SANY1.107s,
TLC62.378s); retire44838, native-override-final-normal.log. All packages reran.
Full TLC race passes557.551s; retire9423, native-override-tlc-race.log.
Its test matrix was compiled
before EvalControl additions. Focused EvalControl race covers the only two added
methods against identical final production; collectively no method omitted.
Retain that distinction in validation reporting; no need to repeat heavy old
TLC race matrix solely for these two independent pure-flag test additions.
Final root/TLC production source unchanged after all three replacement gates
began. No checks remain running. No pending production correction identified
in original four methods. Full race557.551s + focused original EvalControl
race1.029s covers every current TLC test; root new-model race17.314s also passes.
Inventory594/1269 contexts(46.8%),167/626 complete classes(26.7%),675 pending
across459 classes,one partial. Standard topic44/44 classes,75/75 methods complete.
Evaluation topic14/45 classes,15/51 contexts complete,36 pending. MappedJSON
counts independently recomputed and match doc. All five deferred topics unchanged.
Next evaluation cohort: ActionCompositionATest/BTest source module cdot/
ActionComposition, configs A/B,Tool.CDOT_KEY true,doCoverage false. A success
10/4/0/depth3/FINISHED/noGENERAL. B safety10/4/0/depth3,invariant-behavior record,
exact3state trace x0,x4,x6 (source ArrayList capacity2 is not trace length2).
AssignmentInitTest source6/5/0,FINISHED,noGENERAL,zero-uncovered,success.
Inspect all constructor/inherited settings before new credit; no new cohort
Go tests/fixtures added yet. Goal remains active, hundreds of contexts remain.

Staged diff whitespace warnings are confined to unchanged upstream Java fixture
files under native_overrides (verified byte-identical); preserve source bytes.


## 2026-10-04: original action composition and assignment models

Ported ActionCompositionATest/BTest.testSpec and all six original assignment
classes' test methods: AssignmentInit,InitNeg,InitExpensive,Next,Next2,Next3.
New root tlc_action_composition_java_test.go and tlc_assignment_models_java_test.go;
15 original TLA/CFG files copied byte-for-byte into test_vectors/models/cdot and
six assignment model directories. No production changes in this checkpoint.
Cdot classes retain source module path cdot/rootActionComposition, A/B configs,
Tool.CDOT_KEY true using isolated environment property, doCoverage=false,
default debugger/dot/JSON/generate-spec settings and inherited exits. A retains
FINISHED/noGENERAL,10/4/0/depth3/success. B retains10/4/0/depth3,safety violation,
invariant-behavior and state-print records and CommonTestCase's full trace checks:
exact x0,x4,x6, all ordinals and action-metadata conditions. Source list capacity2
is not its length; actual trace has3 states. No extra B assertions added.
Assignment models keep complete original fixtures/configurations and every source
FINISHED/noGENERAL/zero-uncovered/success assertion. Stats Init6/5/0,Neg2/1/0,
Expensive10002/1/0,Next3/2/0,Next2 3/2/0,Next3 26/5/0; only Next assertsdepth2,
as source does. Expensive retains0..10000 and CHOOSE e in SUBSET1..18:TRUE
unchanged. No bounds/expression/caching bypass or substitute model introduced.
All8 unchanged original Java JUnit classes pass separate JVMs(-ea,-Xmx256m,
writable scratch basepath): A0.444s,B0.489s,Init0.446s,Neg0.456s,
Expensive0.497s,Next0.419s,Next2 0.415s,Next3 0.449s.
Go normal A1.044s,B1.425s,Init1.075s,other5 combined5.250s pass.
Combined all8 root race31.089s passes; retire12324,
composition-assignment-race.log. Full workspace passes(root374.271s,SANY1.126s,
TLC61.723s); retire22226, composition-assignment-full-normal.log. No checks remain
running. Retain final production baseline9708a18 full TLC race557.551s plus
original EvalControl race1.029s; current checkpoint changes root tests/fixtures/
docs only, no TLC production or TLC package tests. Repetition unnecessary.
Inventory602/1269 contexts(47.4%),175/626 complete classes(28.0%),667 pending
across451 classes,one partial. Evaluation topic22/45 classes,23/51 methods
complete,28 pending. Standard topic remainsfullycomplete75/75. MappedJSONmain
counts independently recomputed and match document. All15 source files verified
byte-identical including no-final-newline configs/comment whitespace. Deferred
five user topics untouched; goal remains active, hundreds of contexts remain.

Next inspected class SetOfStatesTest still Reconcile (all6 original methods
unported). Its original DummyTLCState extends abstract TLCState, with constant
fingerprint and inherited object-identity equality. EqualityDummyTLCState extends
it and overrides equality/hash by fingerprint and id. Current Go SetOfStates
accepts/returns only *TLCStateMut; older Go tests substitute model-variable states,
sum values instead of fingerprints, use reduced collision cases and omit source
iterator/HashSet assertions. Do not credit them or smuggle fp/id into mutable
state values/test-only callbacks. Faithful port requires actual state polymorphism
at the production set/container boundary and mechanical original dummy classes.
Original matrices32 inserts,32 same-fingerprint/different-id additions,32 equal
reinserts; preserve all assertions and iterator reset/sum528, HashSet size32.
Source SetOfStates.java also catches specific TLCRuntimeException equality
failures to permit mixed value types in liveness successors; current Go put0 has
no source catch. Audit full source function, constructors, tool/no-tool fingerprint
paths, growth, clear/iterator semantics before port credit. This is required
implementation work, not a reason to weaken original tests. No SetOfStates
production/test/fixture changes made yet. Interface return changes will affect
liveness/debugger/container callers; preserve evaluator state semantics and
respect deferred topics (routine adaptations of existing callers are sufficient).
After set class, remaining cdot contexts (CdotWithContext A/B/C/D and chained
cdots) and evaluator model cases continue existing topic; read original setup.


## 2026-10-04: original SetOfStates tests and production state semantics

Completed all six original SetOfStatesTest methods in
set_of_states_java_test.go, mechanically translating DummyTLCState's constant
fingerprint/object identity and EqualityDummyTLCState's class/fingerprint/id
comparison and int32 hash. Preserved every original assertion, all32-element
loops,64 distinct same-fingerprint states,32 equal reinserts, iterator identity,
fingerprint sum528 and Java HashSet size32. Older mutable-state checks remain;
they no longer substitute for the original custom-state tests.
Production SetOfStates now stores the TLCState collection interface rather than
forcing every state to be TLCStateMut. Evaluator/liveness callers explicitly
retain mutable-state boundaries; ToSlice remains the concrete evaluator adapter.
StateVec still uses mutable states: this checkpoint does not claim a complete
abstract-state conversion of every evaluator API. TLCStateMut.Equal now propagates
value equality failures and follows the source receiver-length loop instead of
silently treating errors/unequal lengths as false. Its hash remains native object
identity, preserving Java's deliberate equals-without-hashCode behavior.
SetOfStates implements source tool/no-tool fingerprint overloads through growth,
passes actual worker tools from successor collection, preserves iteratorIndex
across clear/growth, rejects negative capacity/nil vectors, increments iteration
before bounds failure, and prints signed fingerprints with the source UTF16
last-code-unit chop. Collision handling catches only direct TLCRuntimeException
with the source constrained equality-message assertion; other failures propagate.
TLCStateSet provides source HashSet hash/equals semantics for ToSet and action-
identity subsets. No fingerprint/id values were smuggled into mutable states,
and no test-only production callback or reduced matrix was introduced.
Unchanged original Java JUnit: all6 pass0.007s. Go focused normal0.012s and final
focused race1.054s pass. Ignored Java/Go boundary mains produce byte-identical
logs for signed/chopped string, iterator after growth, tool fingerprint rehash,
mixed-type equality catch and nil-vector failure. These probes earn no test
credit and add no persistent regression/unit tests.
Full workspace passes(root379.623s,SANY1.010s,TLC63.792s); session21479 retired.
Earlier normal13214/race82477 were terminated130 for the actual nil-vector
constructor correction, superseded by final verification. Final full TLC race
passes547.729s; session20676 retired, set-of-states-full-race.log. Its compiled
matrix includes all six new methods and final production. Final focused
session51511 has finished0. No checks remain running.
Inventory608/1269 contexts(47.9%),176/626 complete classes(28.1%),661 pending
across450 classes,one partial. Evaluation topic23/45 classes,29/51 methods
complete,22 pending. MappedJSON independently reconciled. Deferred topics
unchanged. Next inspected original CdotWithContextA/B/C/D and ChainedCdots methods;
all set CDOT_KEY=true, debugger=false,coverage=false,source cdot path. B/D
noGenerateSpec=true; A/C/Chained keep generation. Inherited success exits retained.
A/B/C assert FINISHED,noGENERAL,5/3/0,depth2; D1/1/0,depth1. Chained9/4/0,depth2,
no postcondition false/evaluation error, register42 nonempty and first IntValue9.
No new tests/fixtures for these classes activated yet. Preserve constructor,
generation/dump/worker settings and all inherited assertions before credit. Goal remains active.


## 2026-10-04: original action-composition contexts and chained actions

Completed original CdotWithContextA/B/C/D.testSpec and ChainedCdots.testSpec in
root tlc_cdot_context_java_test.go. Ten TLA/CFG files copied byte-for-byte from
source cdot directory; no production changes. Every source assertion and
constructor/inherited option retained: CDOT propertytrue with isolatedt.Setenv,
coverage=false,debugger=false,workers1,deadlockdisabled,fp0/seed1/checkpoint0,
dot/JSONdumptrue,success exits. B/D noGenerateSpecTE; A/C/Chained generate as
source. A/B/C5/3/0 depth2,D1/1/0 depth1,Chained9/4/0 depth2,FINISHED/noGENERAL.
Chained retains both absent postcondition diagnostics, register42 nonempty and
first IntValue9. Actual model's postcondition also checks TLCGet(0)=63, running
through production composition/evaluation/worker registers; no substitute model
or recorder assertions. Full source fixtures retained including B/D temporal
properties and context constraints.
All5 unchanged Java JUnit pass in separate JVMs: A0.414s,B0.386s,C0.428s,
D0.420s,Chained0.457s. Go focused normal5.305s and all5 root race18.876s pass;
retire6460/32713/79004. Full workspace passes(root379.229s,SANY1.102s,TLC61.182s); retire89465,
cdot-context-full-normal.log. No checks remain running. TLC production/pkgtests unchanged; retain current
full TLC race e3b3f46 547.729s, no redundant rerun.
Inventory613/1269 contexts(48.3%),181/626 complete classes(28.9%),656 pending
across445 classes,one partial. Evaluation28/45classes,34/51methods complete,
17pending. MappedJSON independently recomputed; all10fixtures byte-identical.
Deferred five user topics untouched; goal remains active.

Next17 evaluation contexts inspected, no Go tests/fixtures activated yet:
EmptyExistentialQuantifier: -configsameTLA,checkDeadlocktrue,expecteddeadlock,
FINISHED/noGENERAL,1/1/0; default coverage/debugger/dot/JSON/generation retained.
MinimalSetOfInitStates: defaults,FINISHED/noGENERAL,INIT_GENERATED2 8/s/6,
stats14/6/0 depth1,zeroUncovered. Preserve counts despite historical comment.
MinimalSetOfNextStates: defaults,INIT_GENERATED1 1,57/7/0 depth2,FINISHED/noGENERAL,
exact uncovered lines34/42 col10..15, not alternate counts or relaxedcoverage.
UndeclaredRecursion: configTLA,success,debuggerfalse,noGeneration,noJSON,
coveragefalse; dottrue; FINISHED/noGENERAL,depth0,0/0/0.
Remaining eight LET boxed variants and five InitEvalOrder contexts need full
setup/trace/coverage audit before credit. Four InitEvalOrder1..4 methods are
inherited; Basic declares its own FINISHED/noGENERAL,2/1/0,zeroUncovered method.
All use default settings; configs1..4 select four conjunct orders in shared
InitEvalOrder.tla, which imports EvalOrder/Base.tla (must copy this dependency).
No claim of Go port completion. LETa/b liveness exit and fullTRUE/FALSEtrace;
c/d success stats3/2/0, FINISHED/SUCCESS; defaults differ by variant.

Original Java preflight of all17 remaining evaluation contexts now passes
unchanged. First InitEvalOrder1..4 scratch attempts failed because Base.tla was
missing from scratch; copied the actual source dependency and allfour pass on
rerun. No Java source/assertions changed. Logs original-<Class>-junit.log;
evaluation-remaining-javac.log. Retire32629 and42700. This earns no port credit;
all17 remain Missing until complete Go translations and verification.

Staged diff --check warnings are only original fixture bytes: source trailing
space and TLA ======= delimiters (Git labels these conflict markers). All ten
files verified byte-identical; preserve them. Go/docs diff --check is clean.


## 2026-10-04: remaining original evaluation methods

Ported all17 remaining contexts: EmptyExistentialQuantifier, MinimalSetOfInit/
NextStates,UndeclaredRecursion (tlc_evaluation_remaining_java_test.go), eight
LetDef1/2Boxed variants (tlc_let_boxed_java_test.go), InitEvalOrder1..4 inherited
methods plus Basic's declared method (tlc_init_eval_order_java_test.go).
All source settings and assertions retained; no production changes. Empty
quantifier enablesdeadlock,configTLA,expectsdeadlock,FINISHED/noGENERAL,1/1/0.
MinimalInit retains8/s/6,14/6/0,depth1,zeroUncovered. MinimalNext retains1initial,
57/7/0,depth2 and exactly zero coverage at source lines34/42,col10..15; helper
uses source CommonTestCase.assertUncovered's set semantics, not relaxedcoverage.
Recursion retains embedded Base/configTLA,debuggerfalse,coveragefalse,noJSON,
noGenerateSpec,dottrue,success,FINISHED/noGENERAL,0/0/0,depth0.
LET variantsa/b retain debuggerfalse,expectedliveness,2/2/0,FINISHED,action-property
violation,STATE_PRINT2 and fullCommonTestCase traceTRUE/FALSE (initial predicate,
action metadata, allordinals,ASCIItrim,exactlength). c/d retain defaultdebugger,
success,FINISHED/SUCCESS,3/2/0. AllLET retaincoverage/dottrue,noGenerateSpec,noJSON;
no unsolicited coverage/general/stuttering assertions. InitOrder1..4 preserve
configs/conjunctorders,defaultsettings,success,FINISHED/noGENERAL,2/1/0; Basic
alsozeroUncovered. EvalOrder/Base.tla included, not substituted. All22 original
TLA/CFG fixture files verified byte-identical. Copyright notices retained.
All17 unchanged original Java tests pass (prior checkpoint preflight, original-
<Class>-junit.log). Go focused four4.765s,LET8 4.878s,InitOrder5 2.159s pass;
retire67012/85439/29205. Combined all17 root race58.170s passes; retire3590,
evaluation-final-race.log. Full workspace passes(root380.443s,SANY1.021s,TLC61.213s); retire66868,
evaluation-final-full-normal.log. No checks remain running. No TLC production/pkgtests changed; retain
full TLC race e3b3f46 547.729s; no redundant repeat.
Inventory630/1269contexts(49.6%),198/626 completeclasses(31.6%),639pending
across428classes,onepartial. Evaluation nowfullycomplete45/45classes,51/51
contexts. MappedJSON independently recomputed. Deferred topics unchanged;
goal remains active, hundreds of original contexts remain.

Next inspected ASTest and ActionLevelPropA..E; no Go tests/fixtures activated.
AS sourcefolderAS,expected FAILURE_SPEC_EVAL,only STATES_AND_NO_NEXT_ACTION plus
inheritedexit; all defaults retained. ActionLevelProp sharedmodule/configA..E,
coveragefalse,allotherdefaults,ERROR_CONFIG_PARSE,FINISHED,source diagnostic:
A/B/C PROPERTY_ACTION_LEVEL,D PROPERTY_ACTION_LEVEL_SQUARE_A_SUB_V,
E PROPERTY_ACTION_LEVEL_ANGLE_A_SUB_V. Preserve original config/model files and
source distinction among bare, square,angle action-level properties.

Allsix upcoming AS/ActionLevelProp Java JUnit contexts pass unchanged in separate
JVM scratch runs; original-<Class>-junit.log,safety-action-level-javac.log,
retire71137. No Go port credit/tests/fixtures for those six yet.

Staged diff --check warnings are confined to unchanged source fixture trailing
whitespace in EvalOrder and MinimalSet files; all22fixtures independently
byte-compared. Go/docs staged whitespace check is clean. Preserve source bytes.


## 2026-10-04: original configuration and absolute-path diagnostics

Completed7 original contexts: ASTest.testSpec,ActionLevelPropA..E.testSpec,
AbsoluteSpecPathTest.test. New root tlc_as_java_test.go,
tlc_action_level_prop_java_test.go,tlc_absolute_spec_path_java_test.go; ten
TLA/CFG fixtures from AS/sharedActionLevelProp/Test2 copied byte-for-byte.
No production changes. AS retains defaults and only original NO_NEXT_ACTION
plus inherited FAILURE_SPEC_EVAL exit; no extra assertions. ActionLevelProp
retains coveragefalse and all other defaults,configA..E,FINISHED,
ERROR_CONFIG_PARSE and exact source codes: A/B/Cbare,D square,Eangle.
AbsolutePath retains CommonTestCase's actual singleCLIarg absoluteTest2path,
FINISHED,depth5,6/5/0,noGENERAL; no new exit,seed,deadlock ortrace-generation
assertions. The actual fixture files are copied unchanged into t.TempDir's
absolute BASE_PATH before execution so default metadata is cleaned there.
The earlier direct persistent fixture path created runtime states artifacts;
terminated fullnormal65472(130) for this actual test-isolation correction.
Removed only those newly generated Test2/states files after terminal status;
fixtures untouched. Source Test2's postcondition currently fails in both Java
and Go, but this original method has no exit/postcondition assertion. Do not
credit numbered Test2 model context from this distinct absolutepath test.
All7 unchanged Java JUnit pass: AS0.421s,A0.309s,B0.336s,C0.308s,D0.301s,
E0.300s,Absolute0.389s; original-<Class>-junit.log. Go first6normal2.427s,
combined final7normal2.634s(beforepathisolation) pass. Final isolated all7
rootrace28.203s passes; retire34215,safety-diagnostics-isolated-race.log.
Earlier race13524 passed26.291s and is superseded; notrunning. Retire20513,
93536,95015,90586; the lasttwo include upstreamAssert failure detailed below.
Final fullworkspace passes(root375.799s,SANY1.087s,TLC62.317s); retire34721,
safety-diagnostics-isolated-full-normal.log. No checks remain running. TLC production/pkgtests unchanged;
retain fullTLCrace e3b3f46 547.729s. Inventory637/1269contexts(50.2%),205/626
completeclasses(32.7%),632pendingacross421classes,onepartial. Safetytopic14/34
contexts/classes complete,20pending. Evaluation51/51 remainscomplete.
All10fixtures exactbytes, no runtime files should entercommit. Deferred topics
unchanged; overallgoalactive.

Upstream source-test mismatch: AssertExpressionStack (not a Go implementation
shortfall). Exact mechanical translation failed inheritedSUCCESS(exit0) and
assertNoTESpec; unchanged Java at rev8f4bc8b fails exactly those same assertions.
Java current EC maps VALUE_ASSERT_FAILED toVIOLATION_ASSERT14 and current
ModelCheckerTestCase defaults force -generateSpecTE. Go correctly produces
exit14, trace x0/x1, nestedcallstack and generatedTEspec, matching sourceengine.
Did not change source expectations, addskip, or make production incompatible.
Mark Reconcile and keepuncredited. The exact failing translation is only an
ignored scratch draft assert-expression-stack-draft.go.txt; its two untracked
fixture copies removed, not part of the green suite. Logs original-
AssertExpressionStack-junit.log and safety-path-stack-go.log prove mismatch.
Java workingtree clean. Still needs upstream expectations reconciled before a
faithful green port can be credited; this does not prevent other valid ports.

Next preflight ContinueTest/DepthFirstTerminate/EmptyTest source and models read.
Continue unchangedJava passes; workers3,-continue,noGenerateSpec,defaultcoverage/
dot/JSON,success,FINISHED,32/29/0,depth29,STATE_PRINT2 exactly32 records,
noTEspec,zeroUncovered. Source commented-out trace lists remainnonassertions.
Empty unchangedJava passes; defaults plus explicit second -coverage1,success,
FINISHED/noGENERAL,0/0/0. No Go translations/fixtures activated yet.
DepthFirstTerminate unchangedJava fails on48corehost: source overridesworkers
availableProcessors and -dfid50, but current TLC intentionally rejectsmultiple
DFIDworkers(issue548),recordingGENERAL/exit255 againstnoGENERAL/successasserts.
Mark Reconcile,keepuncredited; no single-worker substitute/newskip introduced.
Logs original-ContinueTest-junit.log,original-DepthFirstTerminate-junit.log,
original-EmptyTest-junit.log,safety-lifecycle-javac.log; retire23553 (1 dueJava
DFID failure) and separateEmpty run(0). No checks frompreflight remainrunning.

Source customBuild.xml default test/test-dist broad batch pattern is
**/*Test*.java (lines810/1253); neither AssertExpressionStack.java nor
DepthFirstTerminate.java matches it. Main inventory intentionally includes
nonignored concrete classes beyond the default Ant selection, so both remain
visible/pending without credit. Go DFID already has the multiple-worker guard
(dfid_checker.go); its diagnostic currently omits Java's issue548 URL, a future
source-message reconciliation detail, not justification to substituteworkers.

Staged diff --check flags only original AS.tla line18 trailing whitespace,
verified source-identical. Preserve fixture bytes. No generated runtime
metadata remains under persistent Test2 vectors; staged Go/docs check is clean.


## 2026-10-04: original continuation, diameter and postcondition methods

Completed5 original methods: ContinueTest/EmptyTest in
root tlc_continue_empty_java_test.go,DiameterTest in tlc_diameter_java_test.go,
MinimumDiameterTest/PostAssumptionTest in tlc_post_assumption_minimum_java_test.go.
Nine TLA/CFG fixtures (Continue2,Empty2,DieHardTLA2+DieHardTLAPA1,Minimum2)
copied byte-for-byte. All original source setup/assertions preserved.
Continue workers3,-continue,noGeneration,coverage/dot/JSONtrue,defaultdebugger
option (notinstalledwithworkers3),success,FINISHED,32/29/0,depth29,STATE_PRINT2,
exact32concatenatedrecords,noTEspec,zeroUncovered. Source commented-out trace
lists remain nonassertions; no traceorder substitution/singleworker shortcut.
Empty defaults plus secondexplicit -coverage1,success,FINISHED/noGENERAL,0/0/0.
Diameter workers4,defaultotherflags,success,FINISHED/noGENERAL,97/16/0,
getRecordAsInt's actual firstdepthrecord parsed32bit >=8 (not exact8/oneworker),
zeroUncovered. Minimum defaults exceptdotfalse,success,FINISHED/noGENERAL,
2/1/0,depth1,zeroUncovered. PostAssumption sharedDieHardTLA/configDieHardTLAPA
(without.cfgsuffixasoriginal),defaults,VIOLATION_ASSUMPTION,FINISHED,97/16/0,
POSTCONDITION_FALSE PostCondition,zeroUncovered; originalfails postcondition
intentionally (no noGENERAL assertion added).
Unchanged Java5 JUnit pass separateJVMs: Continue0.433s,Empty0.449s,
Diameter0.624s,Minimum0.538s,Post0.524s; safety-counts-javac.log,original-
<Class>-junit.log. Go firstnormalContinue/Empty3.713s,Diameter1.789s,
Post/Minimum3.331s pass; retire31721/86190/94580/96649.

Combined firstGo race31854 failed18.056s: Diameter's actual PossibleCounts
postcondition false,exit10 vsSUCCESS. No DataRace warning: this was a logical
lost-update bug despite individually locked register accesses. Original fixture
postcondition checks _Counts HasFour8 and PourAction14; do not change/remove it
or weaken inheritedexit, workers4,coverage ordepth assertion. Retired original
fullworkspace91424(130) for the actual production correction below.
Production possible_action.go used possibleLocalCounts/setPossibleLocalCounts
indexed by predecessorstate.WorkerID. That is the state's creator, not the
executing worker; workers evaluating successors from another worker's states
could concurrently read/write the same register and lose counts. Java generated
_Possible._Update/_Track delegates TLCGetOrDefault("s:_possible",<<>>)/TLCSet,
which use Thread.currentThread()/broadcastatbootstrap. Replaced both shortcuts
with those real production Go adapters and the source EmptyTupledefault;
removed incorrect helperfunctions. This also preserves simulatorworker-local
updates and main-thread broadcast/bootstrap instead of prior mismatches. No
new test-only hooks, locks hiding wrongownership, cannedcount, test weakening,
regression/unit methods orfixturechanges.
Unchanged fourworkerDiameter nowpassesfive consecutive -race repetitions19.557s;
retire66750,diameter-fixed-repeat-race.log. Original Java passesfive more separate
JVMrepetitions,diameter-java-repeated.log; retire88799. All5finalGo rootrace
passes21.521s; retire42421,safety-counts-fixed-root-race.log.
Fullworkspace finalproduction passes(root377.387s,SANY0.972s,TLC63.260s);
retire15462,safety-counts-fixed-full-normal.log. FullTLCrace passes560.987s; retire47789,
safety-counts-fixed-tlc-race.log. No checks remain running.
Do not restart for observationtimeouts. These supersede oldbaseline for changed
_POSSIBLE production. Inventory642/1269contexts(50.6%),210/626completeclasses
(33.5%),627pendingacross416classes,onepartial. Safetytopic19/34complete,
15pending. Deferredtopicsunchanged; goalactive.

Next inspected source DotConstrained,ElevatedSanyWarning,InvParameterizedA/B/C;
no Go tests/fixtures for these activated. DotConstrained needs full actual
getStateWriter subclass/wrapper after CLIparse: explicitdotconstrained,
default doDumpfalse,strictprefixtrue (strictdedupfalse),defaultotheropts,
callbackcalls basewrite before atomically marking IsNotInModel. Must assert
actualcallback,FINISHED/noGENERAL,4/2/0,safety,trace0/1/-1(allmetadata/ordinals),
register42nonempty/firstInt4,bothpostconditiondiagnosticsabsent,zeroUncovered.
Current root helper has no getStateWriter override; production Options.StateWriter
and StateWriter callback surface exist. Do not fake callback by reading DOT text,
setflagunconditionally or replace expectedtrace. Implement faithful setup and
real delegation beforecredit. Ninefixtures for currentbatch sourceverified.
Elevated warning needs source SANYW4802 PreTest and actualToolIOcapturedstdout,
messagesAsErrors4802,ERROR_SPEC_PARSE,ECTLCParsingFailed and substringWarning
 treatedaserror; expectedparseerrors must flow through productionratherthan
current modelhelper's requireNoErrors extra check. NotDebuggerfeaturework.
InvParameterizedA invlevel4=>26/7/1depth4,safety;B -inv~(small=3/\big=0)=>
2/2/0depth2,safety;Cgarbledexpression=>ERROR_SPEC_PARSE+FINISHED.
All3 configembeddedInvParameterizedTest.tla,noGeneration,coveragefalse,dotfalse,
JSONfalse,defaultdebuggertrue; A/BzeroUncovered evencoveragefalse.
Bridge has runtimeexpr parser; sourceCLIdiagflow and modelhelper negativeparse
handling needaudit beforetranslation. Nocredit yet; do not relaxparseexit.

Original upcoming DotConstrained and InvParameterizedA/B/C JavaJUnit now all
passunchanged; safety-next-javac.log,original-<Class>-junit.log; retire42543.
No Goportcredit yet. CLI root cli.go loader currently returns raw Diagnostics
errors from runtimeinvariant parsing; TLCrunner/modelhelper negativeparse flow
stillneeds full sourceaudit beforeC, not fakeexpectedexit. These are production
boundary discrepancies to check, not reasons to weaken C's ERROR_SPEC_PARSE.

Staged diff --check warnings are confined to original Continue/DieHardTLA
fixture trailing whitespace (all9 sourcefiles byte-identical). Preserve bytes.
Staged Go/production/docs check is clean.

2026-10-04: DotConstrainedTest.testSpec ported in
`tlc_dot_constrained_java_test.go`. Production runner/checker/liveness input
boundaries now accept IStateWriter, preserving Java custom-writer virtual
transition dispatch. Concrete default writers and liveness implementation remain.
Test embeds actual DOT writer, calls base first and atomically observes source
IsNotInModel bit. Uses original strict prefix (not dedup), constrained true,
other options false; source CLI/default setup retained. Parser-created writer is
closed before reopening the same filename to avoid buffered output clobbering.
All source assertions retained: safety exit, FINISHED/noGENERAL,4/2/0,
invariant behavior,exact0/1/-1 trace with metadata/ordinals,callbacktrue,
register42nonempty/firstInt4,no postcondition diagnostics,zeroUncovered.
Github602.tla/DotConstrained.cfg byte-identical to upstream. Unchanged Java
passes .512s; Go focused1.863s and five race repetitions20.059s pass.
Full workspace passed: root374.736s,SANY1.090s,TLC65.439s; retire44042.
Full TLCrace passed544.633s; retire56130. All checks terminal, no live runs.
Logs dot-constrained-{workspace,tlc-race,root-race}.log under scratch.
Inventory643/1269contexts(50.7%),211/626classes(33.7%);626pending415classes,
safety20/34complete14pending. Deferred topics unchanged; goal active.

Next InvParameterizedA/B/C: retain all source arguments/assertions/embeddedconfig.
Original Java all3 pass. Source SpecProcessor.java lines921..932 catches
ParseException/SemanticException/AbortException from runtime invariant generation
and Assert.fail(TLC_PARSING_FAILED2,e), not GENERAL or success. Go raw diagnostic
errors from cli.go loader currently fall through GENERAL, and positive-only root
model helper aborts before Process on diags.HasErrors. Fix actual production
loader classification and adapt original negative-case harness before C credit;
do not weaken ERROR_SPEC_PARSE exit or hardcode test results. No Go A/B/C added.

2026-10-04 InvParameterizedA/B/C original testSpec ports added in
`tlc_inv_parameterized_java_test.go`, original embedded-config fixture exactbytes.
Source defaultdebuggertrue/workers1,noGeneration,coveragefalse,dotfalse,JSONfalse
retained. Ainvlevel4=26/7/1depth4,safetyexit,FINISHED,zeroUncovered;
Bexact~(small=3 /\ big=0)=2/2/0depth2,safetyexit,FINISHED,zeroUncovered;
Cexactgarbledinput=ERROR_SPEC_PARSE+FINISHED only(asoriginalplusinheritedexit).
Initial Go A/B pass,Cfails positive-harness diagnostic becauseproductionruntime
invariant used standalone syntax-only parse and silentlyskipped invalidinput.
Fixed actual bridge to call existing shared expression compiler(root/nullLoc),
retain compiledOpDef/internalAction/sourceActionname,throwTLC_PARSING_FAILED2
oncompilererror. Removed oldparseRuntimeTLAExpression; noexpected-casehook,
positive-model requireNoErrors unchanged. Syntax-only exceptionfix was superseded
beforecredit by full sharedcompilerreuse to avoid retaining shortcut.
UnchangedJavaA/B/Cpass .411/.362/.252s. FinalfocusedGo3.548s and all3five-race
repetitions34.639s pass; retire24248/33147. Fullworkspace passes374.918/1.073/61.063s; retire78462.
Log inv-parameterized-workspace.log; all sessions terminal. TLCpackage
unchanged since f4a73e0 verified fullTLCrace544.633s; rootbridgechangecovered
by new rootrace and fullworkspace. Inventory646/1269contexts50.9%,214/626classes
34.2%,623pending412classes,onepartial;safety23/34complete11pending.
Deferredtopics unchanged; goalactive.

Next inspected originalFingerprintExceptionInitTest/NextTest/NextCallstackTest;
all3unchangedJavaJUnitpass,logsoriginal-<Class>-junit.log,compilelog
fingerprint-next-javac.log; retire85550. Exactfixtures copied onlyignoredJava
scratchmodels, no persistentGo fixtures/testports yet. Preserve complete
fingerprint diagnostic stacks/overflowmessages, InitnoGENERAL,NextGENERAL,
source stats/initcounts,coverageuncoveredsets,FAILURE_SPEC_EVALexit/defaults.
Hangvariant inspected and unchangedJavaJUnit passes .445s; compilelog
fingerprint-hang-javac.log/original-FingerprintExceptionHangTest-junit.log,
retire69454. SourceFAILURE_SPEC_EVAL,FINISHED/noGENERAL,exact3-framecallstack
and integer1/nonintegeremptysetcomparisonmessage. No Goport activated for any
of4fingerprinttests yet; no manualcomparisoncredit.

2026-10-04 four FingerprintException original testSpec ports activated:
Init/Next/NextCallstack in tlc_fingerprint_exceptions_java_test.go,Hang in
 tlc_fingerprint_hang_java_test.go. Eight exactbytes fixtures (Nextcfgsharedcopy).
Allsource defaults/debugger/options/JSON,FAILURE_SPEC_EVAL exits preserved.
InitFINISHED/noGENERAL,exacttwo-value-source-stack+overflow36,twoUncovered;
NextFINISHED/stats2/1/0/init1/GENERAL,exactvalue-stack+overflow32,oneUncovered;
CallstackFINISHED/stats1/1/0/init1/GENERAL,exactfiveframes+SUBSET42message;
HangFINISHED/noGENERAL,exactthreeframes+integer1/noninteger{}message.
InitialGo3fail(Init/Nextexit255,CallstackmissingGENERAL),Hangpass. Production
fixes: DoInitFunctor catches panics,retainsstate/cause,typedabort/delayedthrow and
OutOfMemoryTooManyInit; init/replay boundary catchespanics;Worker.addElement
catchExceptionwrapscause+successor,JavaErrorsescape;innernext-generationcatch
separatefromouterWorkerloopcatch,bothpreserveactualthrowable. LazyValue.getValue
andLazySupplier.getValue no longerattachsource(JavaonlyLazyValue.evaldoes).
IncorrectsourceattachmentturnednormalCallstackRuntimeExceptioninto nullable-
messageFingerprintException andsuppressedGENERAL. Do notrestore it.
FinalfocusedGo5.047s, unchangedJava4pass .433/.409/.424/.445s. Testsunchanged
through fixes; no inventedchecks. Retire86002/38271/38006/6229/35910.
Checks currentlylive: fullworkspace16332,fullTLCrace38449,all4five-rootrace90758;
logs fingerprint-{workspace,tlc-race,root-race}.log; retireafterterminalresults.
Inventory650/1269contexts51.2%,218/626classes34.8%,619pending408classes,
safety27/34complete7pending;deferredtopics unchanged;goalactive.
Nextinspected ViewMap(full8statetrace+all8exactactionlabels,register42firstInt43,
uncoveredline91col60..73,postconditiondiagsabsent,safety/default -view),
MonolithSpec(embeddedcfg,214/54/0,zeroUncovered,actualToolIOcapturednoMissingFile
and3parsing-pathregexes),TSnapShot4workers(MCroot,TSnapShotsubdir,FAILURE_SAFETY_EVAL,
FINISHED/noGENERAL/noBUG/behavior). NoGoportcredit yet; sourcefilesinspectedonly.
Monolith outputassertions needactualproductionSANY/TLC logging, not fabricated
text or weakerregexes. TSnapShotis multithreadedmodeldiagnostic, notcheckpoint.

UpcomingViewMap/MonolithSpec/TSnapShot originalJavaJUnitpreflight allpassunchanged
.623/.525/.501s; compilelog safety-final-next-javac.log,logs original-<Class>-junit.log;
retire8183. Scratchrun directory safety-next-model-run/test-model intentionally
preserves sourceMonolithpathregex ending test-model. ViewMapfirstpreflightfailed
because scratch target/GeneratedTESpecs directory absent(_TLCTraceSilent binary
serialization),notanupstreamtestdiscrepancy; aftercreatingexpectedharnessoutputdir,
unchangedtestpasses. Do notmarkViewMapReconcile/newskip or removePostCondition
assertionbasedonthatscratchfailure. FutureJava modelpreflights must create
BASE_PATH/../target/GeneratedTESpecs beforegeneration,matchingbuildharness.
No Goports of3nexttests yet; currentcredit only4fingerprintmethods.

Final init-outer-catch review corrected fallbackcode: typed runtime initialization
exceptions thrown outside callbacks must retain initExceptionCode(err),matching
ModelChecker.catch(Throwable),ratherthan genericGENERAL. Existingfinalfocused
cases stillrequired unchanged. Oldfullworkspace16332/fullTLCrace38449 were
terminated130 for thisactualproductioncorrection; initialrootrace90758 passed
74.117s. Finalchecks nowlive67023(fullworkspace),36836(fullTLCrace),28485(all4
five-rootrace),logs fingerprint-{workspace,tlc-race,root-race}-final.log. Retire
onlyafterterminal; no observationtimeouts treatedasstop. No finalcommit yet.

Finalall4five-rootrace passes71.626s; retire28485. Bothfinalbroadchecks remain
live67023/36836. Upcomingmonolith loader already parses/registers allmodules
fromphysicalsource in sany_translate.go loadPath; don'tassume unsupported or
extract artificialsidecarfixtures. Its originaltest stillrequiresactualparsing
path/outputmetadata on ToolIOstdout. loadPath currentlyrecordslogicalModuleFiles
but emitsnoparsingprogress; TLAFile has libraryPath accessors. Inspectsource
MonolithSpecExtractor/NamedInputStream/SANYbeforeimplementing requiredoutput.

Finalfullworkspace passes(root377.561s,SANY0.973s,TLC63.747s); retire67023.
Finalrootrace28485passed71.626s(retired). FinalTLCrace36836stilllive; waitfor
terminalresult beforecommit. Lastproductionedit onlyinitfallbackcode, no further
codechangeswhilethesechecksrun. Preserve sourcefixturewhitespace/endmarkers.

FinalfullTLCrace passes552.842s; retire36836. Allcheckshandlesterminal.
Finalworkspace377.561/0.973/63.747s and all4five-rootrace71.626s green against
finalproduction. StagedGo/docs whitespacecheckclean; onefixturetrailing-space
warning is originalNextCallstackmoduleheader, preservedbyte-identically. Ready
for greencommit of4 originaltests+realproductionfixes+inventory/documents.

2026-10-04 ViewMap,MonolithSpec,TSnapShot original testSpec translations added:
 tlc_view_map_java_test.go,tlc_monolith_spec_java_test.go,tlc_tsnapshot_java_test.go.
Six source-exact fixtures. TSnapShot4workers/defaultflags/MCroot/subdir,
FAILURE_SAFETY_EVAL,FINISHED/noGENERAL/noBUG/behavior preserved. ViewMap source
-view/defaultflags/safety,all8states+all8exactactions/ordinals,uncoveredline91
col60..73,register42firstInt43/postconditiondiagnostics preserved. InitialView
failed tupleprintedraw; productionnowSemanticValueString OpAppl override with
SemanticAllParams(sourceunion/removal/subst,fixedpointrecursion) prints labeled
variablesbyVarLoc when tuplelengthmatchesallParams. BaseprintsValuesPPR.
Original sourceallParams semantics retained ratherthan hardcoding ViewMapvars.
Rootmodelhelper restores/resetsUseView to match source per-classloader isolation.
Monolith sourceactualToolIOstreams,embeddedcfg/success,214/54/0,noMissingFile,
3provenancepatterns/zeroUncovered preserved. Regexonlyfixturepathsuffix adapted
from test-model to test_vectors/models/MonolithSpec; sourcewildcarddots retained.
Actual SANY loader fallsbackafterordinaryresolution;rootonlyparsedfirst, missing
module extractedwithdelimiters into realtempfile thenparsedondemand. Progress
canonicalpath/provenance andresolutionerror callbacksrouteactualToolIO in core
loaderandCLI. No fabricatedlogs/fixture-sidecars or weakenedtests. Tmpfiles retained
for semantic/tool/trace lifetime,asexistingclasspath extraction adapters do.
UnchangedJava3pass .623/.525/.501s; focusedGo4.824s and all3five-rootrace61.198s
pass; retire93054/72192/33722/94406/93263/19287/82439. Finalfullworkspace82686,
fullTLCrace91929 stilllive; logs safety-next-{workspace,tlc-race,root-race}.log;
retireonlyafterterminalresults. Inventory653/1269contexts51.5%,221/626classes
35.3%,616pending405classes;safety30/34complete4pending;deferredtopicsunchanged,
goalactive. No finalcommituntilbroadchecksgreen.
Next ElevatedSanyWarning unchangedJavaJUnit passes .262s with BOTH source
W4802_Pre_Test.tla/cfg copied to scratchcorpus preservingabsolutepath. Initial
scratchrun255missingcfgwascopyomission, notupstreamdiscrepancy. Logs
 original-ElevatedSanyWarning-junit.log/elevated-warning-javac.log.
No Goport activatedyet. Must apply SANYMessageControls in coreloader, reportactual
promoteddiagnosticonToolIOstdout,throwTLC_PARSING_FAILED,ERROR_SPEC_PARSE;
originaltest'sstdoutsubstringWarningtreatedaserror+parsingfailurecode+exit.
CoreCLIalreadyappliesdiagnosticCLIOptions.withTLCMessageControls;coreapp loader
currentlydoesn't. Do not fake output, dropinheritedexit or weakensourcecase.

Current batch broad-check correction: workspace82686 failed only the existing
SANY sibling-module source-position assertion (temporary path vs logical Common).
Extracted sibling syntax now retains module-relative positions and physical
SourcePath separately; original assertion unchanged. TLCrace91929 passed562.003s.
Source review also corrected APSubstInNode to simultaneous Subst.allParamSet
replacement over original body parameters; SubstInNode remains sequential.
Retire32070(cancel130 for this code correction),69453(terminalPASS).
Final checks now58698(workspace),39479(TLCrace),81848(five focused rootrace),
logs safety-next-final-{workspace,tlc-race,root-race}.log. No commit until green.

Final workspace58698 PASS: root382.944s, SANY0.926s, TLC64.393s; retire58698.
Final focused rootrace81848 PASS60.755s; retire81848. FullTLCrace39479
stilllive. StagedGo/docs whitespaceclean; fixturewarnings are byte-exact
original whitespace, deliberately preserved.

Final TLCrace39479 PASS560.863s; retire39479. All final checks terminal
and green. This batch completes ViewMap, MonolithSpec and TSnapShot original
testSpec translations plus the corresponding production fixes. Inventory
653/1269contexts51.5%,221/626classes35.3%;616contexts remain across405classes.
Commit this green checkpoint; next ElevatedSanyWarning (Java preflight passed).
The overall correctness-suite port goal remains active; deferred topics unchanged.

2026-10-04 ElevatedSanyWarning.testSpec and beforeSetUp translated in
 tlc_elevated_sany_warning_java_test.go; both original W4802 corpus fixtures
copied byte-exact to test_vectors/models/ElevatedSanyWarning. Absolute TLA
argument, messagesAsErrors4802, source JSON/DOT/coverage/debugger/worker defaults,
ERROR_SPEC_PARSE, TLC_PARSING_FAILED and actual ToolIO stdout substring retained.
Core shared loader now applies SANY message controls, reports actual diagnostics
on ToolIO.out and raises source coded parsing failure for checked errors or
promoted warnings, retaining ordinary-error parameters. No weakened assertions.
First Go run failed before loading because shared test helper used the absolute
specification as a dump filename; helper now puts its basename under metadir.
Original Java uses the class name under metadir. Absolute input retained.
Helper isolates all four MP message-control maps per source classloader and
restores prior maps, so a promoted warning cannot leak to subsequent models.
Original Java PASS0.262s; Go focused PASS0.033s; all seven selected neighboring
models five-race repetitions PASS82.197s, retire34003. Workspace93076 live;
log elevated-warning-workspace.log. tlc package unchanged since96097d0, whose
fullTLCrace PASS560.863s remains baseline. Inventory654/1269contexts51.5%,
222/626classes35.5%;615pending404classes; safety31/34complete3pending.
Next issue-regression sources BugzillaBug279,CodePlexBug21,Github1087 inspected;
unchanged original Java preflights all PASS (logs original-<Class>-junit.log).
Java generated states moved to ignored correctness-java/issues-next-states.
No Go translations/fixture credit for those three yet. Bug279 requires original
checkDeadLock=true,doDump=false (avoid enumerating SUBSET1..20), exact three-state
trace/normalized full SUBSET1..8 expected string,3/3/0,zeroUncovered,deadlock exit.
CodePlexBug21 source assumption/function-override model0/0/0/SUCCESS/noGENERAL;
Github1087 uses Github602 root/explicitcfg,2/1/0,trace1/-1,postcondition registers.
Overall goal active and deferred topics unchanged. No commit until workspace green.

Final Elevated workspace93076 PASS377.101/1.105/62.023s; retire93076.
All final handles terminal; rootrace34003 PASS82.197s; TLCpackage unchanged,
96097d0 fullrace560.863s retained. Original fixture blank EOF line preserved;
stagedGo/docs whitespaceclean. Commit warning port and continue issue regressions.
Next originalJava Bugzilla2792.209s/CodePlex21.479s/Github1087.510s all PASS.

2026-10-04 issue regressions: original BugzillaBug279Test,CodePlexBug21Test and
Github1087Test testSpec translations in tlc_bugzilla_279_java_test.go,
 tlc_codeplex_21_java_test.go,tlc_github_1087_java_test.go. Six byte-exact fixtures.
Bug279 exact source 3/3/0,deadlock/FINISHED/noGENERAL,three states/full source
Values.ppr(normalizedSUBSET1..8) expectation,zeroUncovered; checkDeadLock=true,
doDump=false and all other base defaults incl JSON/coverage/debugger retained.
The SUBSET1..20 bound is unchanged; no costly-enumeration substitute or sample.
CodePlex21 function/tuple override assumptions,0/0/0,SUCCESS/FINISHED/noGENERAL;
Github1087 Github602 root,explicitcfg,safety/FINISHED/noGENERAL,2/1/0,exacttrace
1/-1,register42firstInt2/bothpostconditiondiagnostics retained. Trace helpers
also validate source state ordinals/action metadata. No production changes
needed, no weakened/invented assertions. Original Java passes2.209/.479/.510s;
Go focused all3pass6.160s. Workspace31154 and five-rootrace38900 live; logs
issues-next-{workspace,root-race}.log; retire only after terminal results.
TLCpackage unchanged since96097d0 fullrace560.863s. Inventory657/1269contexts
51.8%,225/626classes35.9%;612pending401classes;issue topic6/98classescomplete.
Next Github1134a/b/c/d/e/f source inspected; all unchanged Java preflights PASS
(consult original-Github1134{a,b,c,d,e,f}Test-junit.log for timings). No Go tests/fixtures or inventory credit for those yet. Models
live in source test-model/gh1134 (five TLA modules plus six cfgs), rootMC.tla,
explicitMC{letter}.cfg, noGenerateSpec=true/doDumpTrace=false, otherdefaults.
A livenessinitialproperty failure Init/stateINITtuple+zeroUncovered; B/C/E
SUCCESS,9/4/0,zeroUncovered; D/F initial-evaluation failure exact tuplefunction
out-of-domain3 message and initialstate. AllFINISHED and source exits retained.
Temporary Java runs now workdir ignored scratch root, keeping generatedstates
out of repo root. Deferred topics unchanged; overallgoalactive.

Issue batch final focused rootrace38900 PASS115.445s; retire38900.
All six Github1134 Java preflights pass unchanged; no Go translations activated.
Workspace31154 remains live; no new production changes during this check.

Final issue workspace31154 PASS374.743/1.001/62.059s; retire31154.
Rootrace38900 PASS115.445s; all final handles terminal, baselineTLC fullrace
560.863s retained (no TLCpackage code changes). StagedGo/docs whitespaceclean;
Github602 original trailing blanks retained, all six fixtures byte-exact.
Commit issue tests as green checkpoint; next Github1134a..f, all source Java
preflights green and no Go translations activated. Inventory657/1269contexts
51.8%,225/626classes35.9%,612pending401classes; goalactive/deferred unchanged.

2026-10-04 Github1134a/b/c/d/e/f original testSpec translations in
 tlc_github_1134_java_test.go; eleven byte-exact source gh1134 fixtures.
Full MC.tla root/explicitMC{letter}.cfg, noGenerateSpec=true/doDumpTrace=false,
other default coverage/DOT/debugger/oneworker/checkpoint/seed/fp preserved.
A exact Init/stateINITtuple initial-property diagnostic/liveness exit and
zeroUncovered; B/C/E SUCCESS/FINISHED,9/4/0,zeroUncovered; D/F exact tuplefunction
out-of-domain3 initial-error diagnostic/state and FAILURE_SPEC_EVAL preserved.
Initial all6 failed E4200/E4242 because semantic checker treated instance formal
n as a target and lacked WITH-expression scope. Source Generator module definition
pushes formal context first: checker now uses local formals/implicit bindings,
keeps explicit-target validation separate, preserves duplicate-formal rejection.
Instance AST retains declared operator formal arities (ident/fixity); translator,
semantic checks, bridge FormalParam symbols and shared XML export carry them.
No arity-zero assumption for arbitrary instance formals. XML field propagation
only; no long XML corpus sweeps. Bridge instance params use actual FormalParam
symbols/positions. Next A failed name SingleUser!Init vs source Init: original
source operator Symbol retains declared name; qualified alias stays lookup key.
Instantiated export names unchanged. No diagnostic string stripping or test edits.
Original Java all6pass (original-Github1134{letter}Test-junit.log). Go final focused
all6 PASS5.381s; retire60855/88726/47058/17620. Workspace40684 and five-rootrace
30778 live; logs github1134-{workspace,root-race}.log. TLCpackage code unchanged,
96097d0 fullrace560.863s baseline retained. No commit until broad checks green.
Inventory663/1269contexts52.2%,231/626classes36.9%;606pending395classes,
issue topic12/98classescomplete; goalactive/deferredtopics unchanged.

Github1134 finalrootrace30778 PASS102.233s (six originals plus existing
SANY semantic bridge, five repetitions); retire30778. Workspace40684 live.
Next originalJava Github1145/1145b/1147/1161/1161Violated all pass unchanged
with complete fixtures/resourceDOT; logs original-<Class>-junit.log. No Go
tests/fixtures activated for these yet. Java tmpdirs/workdir stay in ignored
scratch. 1161Abstract is a sibling module embedded in Github1161.tla; preserve
that single original source rather than inventing a separate module fixture.
1147 requires full line-by-line golden DOT/EOF equality (actionLabels,colorize),
51/50/47,depth3,postcondition diagnostics,zeroUncovered and safety exit.
1145b deliberately disables debugger/gen/coverage/DOT/JSON to retain lazy
FcnLambda->tuple EXCEPT path; exact initial invariant Inv,x=1 and safety exit.
1161 pair has debugger/gen/coverage/DOT/JSON disabled; preserve all eight
operator/CONSTANT/recursive wrappers in original model, no sampling. Violated
case exact PropertyViolated temporal/counterexample diagnostics and trace0..5
with source VIOLATION_SAFETY exit. Overall goal active; deferred topics unchanged.

Github1134 broad check40684 FAIL600.286s: existing qualified-LOCAL helper leaked
unqualified, and originalGithub314Coverage ran unbounded then suite timed out.
The symbol-name correction exposed remaining lookup-key shortcuts. Production
bridge defineAlias now registers exact symbol identity and explicit lookup alias
without automatically exporting a source's declared name. Core module config
overrides now retrieve the named module's actual original OpDefs from ModuleTbl,
matching Java's module traversal; no qualified-name expectation on source symbols.
In-memory Defns-only callers retain alias-key fallback. Test assertions unchanged.
All six plus both affected prior tests PASS5.840s, retire22134. Package tlc now
changed, so baseline race is insufficient: rerun full workspace/fullTLCrace plus
five-focused-rootrace. Logs github1134-final-{workspace,tlc-race,root-race}.log.
Retire40684 terminalFAIL; no commit until final checks green. Inventory unchanged.

Final Github1134 handles: workspace8679,fullTLCrace95902,five-rootrace74444
(all six originals,oldGithub314Coverage,qualified-LOCAL prior test,SANYbridge).
All launched after alias/module-definition corrections. Do not reuse initial
102.233s race credit for the later production fixes; wait for final results.

Final rootrace74444 PASS121.839s; retire74444. Includes all six Github1134,
originalGithub314Coverage,qualified-LOCAL prior test,SANYsemanticbridge,5runs.
Finalworkspace8679 and fullTLCrace95902 still live. No further code edits.

Github1134 workspace8679 terminal FAIL383.406s: only existing TestJavaEchoDebugger
failed (9 frames instead of 13). This was an earlier model evaluation failure,
not missing debugger symbols: NoNode's configured model value was overwritten
by a module alias registration, so Echo evaluated its unbounded CHOOSE body.
Java SpecProcessor assigns global operator CONSTANT values to root OpDefNode
ToolObjects (source lines583..597). Go now does the same before registering the
value in Defns; shared source identity preserves replacement across aliases.
No existing test assertions changed. FullTLCrace95902 PASS556.286s before this
last correction; retire95902, rerun after actual package tlc change.
Focused six Github1134 plus Echo/Github314/qualifiedLOCAL PASS3.661s.
Final replacement handles: workspace44395,fullTLCrace15236,5-rootrace25317.
Logs github1134-config-{workspace,tlc-race,root-race}.log. Await green checkpoint.

Replacement full workspace44395 PASS: root387.278s, sany_tests1.017s,
tlc64.414s; retire44395. Replacement rootrace25317 PASS149.901s (5runs),
including all six new tests and unchanged Echo/Github314/qualifiedLOCAL.
FullTLCrace15236 remains live; no further production edits since launch.

Github1134 final fullTLCrace15236 PASS560.735s; retire15236. All final checks
now green after the global operator constant fix: full workspace387.278s,
TLC64.414s, SANY1.017s; full TLC race560.735s; focused root race5runs149.901s.
All 11 original Github1134 fixtures byte-identical; Go/docs diff checks clean.
Six source cases now Port complete; 663/1269 contexts,231/626 classes translated.
Commit this checkpoint before activating next original issue tests. Next scratch
four-test draft github1145-1161-draft.go.txt remains inactive and uncredited;
source originals1145/1145b/1147/1161/1161Violated have already passed unchanged
JavaJUnit. All known test process handles are terminal. Overall goal remains
active; continue exact source assertions and production fixes, deferred topics
unchanged. No new debugger tests/features added; Echo was a core config fix.

Github1145/1145b/1147/1161/1161Violated originals now mechanically ported,
with all ten model/resource fixtures byte-identical (9 model files + 1 goldenDOT).
No weakened test assertions. OriginalJavaJUnit all5 pass unchanged (existing
original-<Class>-junit.log). Go1145 pair PASS3.390s with source default JSON
for1145 and no instrumentation for1145b; lazy EXCEPT conversion already correct.
Initial1161 FAIL applying abs!Test3 with wrong number of arguments; violated
case passed. Bridge shortcut: exported instance symbols had arity0 until body
conversion, misclassifying an operator argument when its definition sorted later.
prepareInstanceDefinitions now records len(instance.Params)+len(source.Params)
before converting any bodies, following Java Generator.generateModuleDefinition
full parameter signature. Both1161 PASS3.043s, unchanged full s=0..5 trace.
Initial1147 failed exact goldenDOT at first high-bit fingerprint: Go rendered
unsigned decimal vs Java Long.toString signed. Production DotStateWriter now
renders signed int64 for init nodes, edges, successors and rank entries.
Exact goldenDOT/EOF, counts/depth/postcondition/zeroUncovered all PASS1.863s.
New tests stay original assertions; no additional invented regression tests.
Inventory668/1269(52.6%),236/626(37.7%);601pending contexts/390classes.
Broad checks live: workspace31317,fullTLCrace5920,focusedrootrace26186(5runs,
all5new plus sixGithub1134). Logs github1145-1161-{workspace,tlc-race,root-race}.
No commit until checks green; previous green checkpoint4772dfc. Current goal
active and deferred topics unchanged. Last focused handles17720,65699,74680,
50973,98661 all terminal; do not re-poll retired handles.

Next originalGithub1198a/b/c/d/f/h JavaJUnit pass unchanged (6classes), using
complete model+Github1198abs+sixconfigs. First scratch pass lacked abs dependency
and failed SANY; after copying exact original dependency all6 PASS. This was a
harness fixture issue, not an upstream assertion conflict. Logs original-
Github1198<letter>Test-junit.log; javac loggithub1198-javac.log. Scratch draft
.codex-gotmp/github1198-draft.go.txt only; no active Go1198 translation/credit yet.
Originals all expect3/2/0,depth2,FINISHED,success,zeroUncovered. Coveragefalse
but debugger/gen/DOT/JSON defaults retained. a/b/c require tautology2258 warning;
d/f/h require its absence. All source configs/properties remain exact; include
Github1198abs.tla even for a..f since the root always imports that module.
RetireJava handles10917/44425 terminal. BroadGo31317/5920/26186 remain live.

Focused rootrace26186 PASS182.489s (5runs, all5new + sixGithub1134); retire26186.
Full workspace31317 and fullTLCrace5920 remain live. Only later Go edit was a
comment correcting the1161 property-count wording; production/assertions unchanged.
Next Java1244/1244b/1244c/1302/1302b/1302c all PASS unchanged; original-
<Class>-junit.log, github1244-1302-javac.log, retire42507 terminal. No activeGo
translation or inventory credit yet. Preserve1244 embedded CONFIG within same
TLA file, exact -config Github1244.tla, noGenerateSpecTE/JSONfalse, otherdefaults;
variants cfgb/c route initial/action/temporal wrappers, exact3/2/0 success.
1302 original simulation num=3/depth3, -config sameTLA, stats10/1/0,zeroUncovered;
coverage/debugger/gen disabled, DOT/JSON retained. 1302b source expectedinitial
Inv exacttwo function states, noStats/noGENERAL, safety exit; coverage/debug/DOT/
JSONdisabled but forcedtracegeneration enabled. 1302c success2/1/0,noGENERAL,
samefalse instrumentation/truegeneration. PreserveactualEXCEPTfunction models
and embedded config, not handwritten substitutes. Next1198 precedes these six.

Full workspace31317 PASS393.269s root,1.051s SANY,64.088s TLC; retire31317.
Rootrace5runs182.489s already passed. FullTLCrace5920 remains live; commit after
that final result. Go/docs whitespace checks clean, all10 originalfixtures exact.

FullTLCrace5920 PASS560.033s; retire5920. Final checkpoint all green: workspace
root393.269s/SANY1.051s/TLC64.088s; fullTLCrace560.033s; focusedrootrace5runs
182.489s (all5new + sixGithub1134). All10 originalmodel/resourcefixtures exact,
Go/docs diff checks clean. Commit new1145/1145b/1147/1161/1161Violated ports and
production signature/DOT corrections now. All process handles terminal. Inventory
668/1269(52.6%),236/626classes(37.7%);601pendingcontexts/390classes,1partial.
Next activate scratch Github1198 draft and complete fixture set; all6 originals
pass JavaJUnit unchanged. Then1244/1302 originals already preflighted. Preserve
full objective/deferred scope and fix implementation on every new red test.

Current batch twelve source cases: Github1198a/b/c/d/f/h,1244/1244b/1244c,
1302/1302b/1302c. All originalJavaJUnit pass unchanged (prior preflightlogs).
InitialGo1198 ABC missing tautology warning, D/F/H pass. Production shortcut
attachSpecPropertyOrigin compared SemanticString (source location) instead of
Java getTreeNode().toString (syntax image). Now compares actual syntax image;
synthetic in-memory nodes retain image fallback. Decomposition appends each
OpAppl to stack in outer caller so level-handler sees complete original stack;
non-next boxed formulas reach common temporal handler/property tagging, asJava.
Unchanged1198 assertions now all PASS5.325s. ExactwarningpresenceABC/absenceDFH,
3/2/0,depth2,zeroUncovered and all original defaults except coveragefalse.
1244 three PASS3.646s, exactembeddedTLAconfig for first and cfgb/c, original
noGenerateSpecTE/JSONfalse + debugger/coverage/DOTtrue. Initial/action/temporal
parameter wrappers exercise production full lazy decomposition; no new fix needed.
1302 three PASS3.645s; exactsimulator num=3/depth3 stats10/1/0,zeroUncovered;
companioninitialInv exacttwofunctionstates and noStats/GENERAL,safety exit;
fingerprintcompanion success2/1/0,noGENERAL. All original runner overrides and
embeddedCONFIGs retained, including forcedtracegeneration for b/c. Only temp
translation stringescaping/build typo corrected; no originalassertion weakening.
14 originalfixture files byte-identical, inventory680/1269(53.6%),248/626(39.6%),
589pendingcontexts/378classes(1partial); all12 markedPortcomplete. Baseline26e8396.
Broadverification live: workspace54316,fullTLCrace18175,focusedrootrace68317,
5runs of all12 + prior1161pair. Logs github1198-1302-{workspace,tlc-race,root-race}.
Do not commit untilchecks green. Focusedhandles42719/75424/71085/17665 terminal.
Goal active, deferredtopics unchanged; next original1389familyaftercheckpoint.

Next originalGithub1389 family all7 JavaJUnit PASS unchanged: Counting,Loops,
StateGuard,base,Violated,ViolatedB,ViolatedC. Exact14 model/config files copied
only to ignoredJava scratch, classes compiled unchanged; logs original-<Class>-
junit.log and github1389-javac.log. Javahandle81078 terminal; retire. No active
Go1389 fixtures/tests/credit yet. All7 disable debugger/gen/DOT/JSON but retain
coverage default true. Counting exactCountAtMostFour violation/safetyexit and
postcondition assertions preserve ten-state alternating witness. Base success.
Violated/B livenessexit,exactPropViolated/two-state x=0,1 trace + back-to1.
ViolatedC livenessexit/postcondition assertions preserve shape with x-domain1..98.
Loops/StateGuard exactSystemStackOverflow + noGENERAL and inheritedERROR exit;
noFINISHED assertion. SourceLoops uses bad(FALSE) re-entering identicalarg;
StateGuard uses op(x), unresolvable state guard during tableau construction.
CurrentGo ASTToLive recurses with no recoverable execution-stack mechanism;
Go runtime overflow is fatal, whereas JavaStackOverflowError is caught by TLC.
Core already represents StackOverflowError and classifies it in runner. Faithful
port must implement recoverable runtime-resource behavior, not weaken/skip these
originals, invent arbitrary semantic recursion cutoffs, or run them in main
agent process without accounting for Go's fatal native-stack behavior. Resolve
production under test before moving to later cases. Overall goal remains active.

Focused rootrace68317 PASS220.799s (5runs all12new + prior1161pair); retire68317.
Workspace54316 and fullTLCrace18175 remain live. No production/assertion changes
since verification launch. Only documentation/inventory bookkeeping afterward.

Full workspace54316 PASS399.664s root,1.073s SANY,63.265s TLC; retire54316.
Focusedrootrace5runs220.799s passed. FullTLCrace18175 remains live; commit only
once final race check passes. Original assertions/fixtures unchanged throughout.

FullTLCrace18175 PASS561.500s; retire18175. All final checks green after core
property-origin correction: workspace399.664s root/1.073s SANY/63.265s TLC;
fullTLCrace561.500s; focusedrootrace5runs220.799s(all12new +1161pair). Original
14 fixtures byte-identical; whitespace checks clean. Commit twelvefaithful
source ports/corefix/docinventory now. Inventory680/1269(53.6%),248/626(39.6%),
589pendingcontexts/378classes. All processhandles terminal; next1389seven
originals preflightedJavahealthy but not activated in Go. No invented regression
or unit tests. Overallgoal active and full requestedscope/deferredtopics retained.

Current seven originalGithub1389 cases all mechanically ported and focusedPASS
7.256s after required production fixes, priorbaseline367465d. All originalJava
JUnit pass unchanged;14 TLA/cfg fixtures byte-identical. Originalsource coverage
true,debugger/gen/DOT/JSONfalse preserved; exact original exits/diagnostics,
fulltrace/back-to-state assertions and modelpostconditions retained. Initialfive
finitecases: base/Counting/Violated pass, B/C fail255 with translator errors.
Shortcut in astToLiveAppl recomputed static recursive level after expansion
returned state-level basecase; Java uses res.getLevel. Go now callsastToLiveLevel
with expandedlevel immediately. Allfive PASS4.797s with originalassertions.
Missing recoverable stack resource: Go nativegoroutine overflow is fatal, Java
Error propagates toTLC.process. Production now anchors actualnative stack
pointer at topASTToLive and threads it through recursiveprivatecalls; nativebyte
span guard uses referenceJVM's observeddefault1MiB stack budget (ThreadStackSize
1024KiB confirmedjava21PrintFlagsFinal loggithub1389-reference-stack-budget.log).
This is actualstackbytes, not arbitraryoperator-expansion/depth cutoff. Pointer
stays onstack and is relocated by Go stackgrowth. Both normal/race compiler
escapeanalysis explicitlyconfirm anchor and stackAnchor donotescape; no marker
heap promotion. Loggithub1389[-race]-stack-escape.log. Guardpanics actualtyped
StackOverflowError, bypassing catch(Exception) analog and using existingcore
runnercode1005/Error exit, noGENERAL. Bothloops/stateGuard PASS. No testcase
weakening, inventedregression/unit tests, sourcefixture changes or debuggerwork.
Inventory687/1269(54.1%),255/626(40.7%);582pendingcontexts/371classes(1partial).
All7 markedPortcomplete. Broadchecks live workspace40527,fullTLCrace30335,
focusedrootrace14065(5runs all7 +1161pair +1198six +1244three). Logs github1389-
{workspace,tlc-race,root-race}. Do not commit until green. Earlierfocusedhandles
32994/70915/67910/46151/91030 and compiler41080/32637 terminal; retire all.
Overallgoal active, deferred topics unchanged; next179a/b/c aftercheckpoint.

Next originalGithub179a/b/c JavaJUnit allPASS unchanged; logoriginal-<Class>-
junit.log and github179-javac.log. Sixmodel/cfg fixtures copied only toignored
Java scratch; no activeGo tests/fixtures/inventorycredit yet. Allrunnerdefaults
remainenabled: debugger,coverage,DOT,JSON,forcedtracegeneration. a assumptions
exitViolationAssumption,FINISHED,exactpublic TLC.PrintT Java method signature and
integer-vs-set normalization failure. b/c FailureSpecEval,FINISHED,exactpublic
TLC.Print/PrintT method signatures and samefailure, plus everyoriginal nested
expression stack line/column/range and trailingnewline. Do not weaken or omit
failurestack assertions. Originalclasses reuse existingcore callstack adapters,
not new debuggerfeature work. RetireJava75759 terminal; broadGo checks remain
40527/30335/14065. Baseline367465d before current uncommittedseven1389ports.

Rootrace14065 PASS329.712s (5runs all7 +1161pair +1198six +1244three); retire.
Full workspace40527 PASS397.851s root/0.992s SANY/63.793s TLC; retire40527.
FullTLCrace30335 remains live; no production/assertion changes since launch.
Repeated rootrace includes both overflowcases and allfinite cases eachrun, proving
continuedcorrectness after stack growth/unwinding with anchors on reusedgoroutine.

Final fullTLCrace30335 PASS556.285s; retire30335. Allfinalchecks green after
recursivelevel/native-stack corrections: workspace397.851s root/0.992s SANY/
63.793s TLC;fullTLCrace556.285s;rootrace5runs329.712s all7plus relatedcases.
Bothcompilerescape proofs normal/race show anchorsdonotescape; referenceJVM
budget1024KiB independentlyrecorded;14 originalfixtures exact;whitespace clean.
Commit sevenfaithful1389ports/corefixes/docs now. Inventory687/1269(54.1%),
255/626(40.7%),582pendingcontexts/371classes(1partial). Allprocesses terminal.
Next179a/b/c originals passed unchangedJavaJUnit; notactivated/credited in Go.
Goal remains active with full originalscope/deferredtopics. No inventedtests,
sourcefixture changes, debuggerfeature work or weakening originalassertions.


2026-10-04: Github179a/b/c original tests translated in
`tlc_github_179_java_test.go`, retaining all runner defaults, exit statuses,
FINISHED, exact reflected method signatures/failure text and complete nested
expression stacks (including final newline). Six original fixtures are byte
identical. Original Java JUnit passes all three. Production standard-method
metadata now includes the full Print/PrintT signatures. CallStackTool replay
starts from the ordinary evaluator rather than copying DebugTool hooks, matching
Java Tool(other)'s shared-Spec constructor and preserving inner predicate frames.
No assertions weakened or new tests invented. Five focused repetitions pass.
Inventory: 690/1269 method contexts (54.4%), 258/626 classes (41.2%);
579 contexts across 368 classes remain (one partial class).
Broad checks pending: workspace11681, full TLC race11832, focused root race53911
(five runs of the three ports and four original FingerprintException cases).
Logs: `.codex-gotmp/github179-{workspace,tlc-race,root-race}.log`.
Next: Github362 instance-scoping test; unchanged original Java JUnit passes.
Overall goal remains active; user-deferred topics remain deferred.


Github179 workspace check PASS: root401.615s, SANY1.128s, TLC63.859s;
retire11681. Focused root race PASS120.451s (five repetitions of all three
Github179 ports and four original FingerprintException ports); retire53911.
Full TLC race11832 remains running. No production edits since checks launched.
Next unchanged Java Github362, Github391 and Github407 JUnit all pass; source
assertions reviewed, no Go credit yet. Goal remains active.


Final Github179 checkpoint: full TLC race PASS571.732s; retire11832.
All checks green: full workspace, full TLC race, five focused normal repetitions
and five focused root race repetitions including original fingerprint failures.
Six fixtures match source bytes; Go/docs whitespace checks clean. Commit the
three faithful ports, native signature metadata and ordinary call-stack replay.
Inventory690/1269(54.4%),258/626(41.2%),579 pending contexts/368 classes.
No live test processes. Next Github362, then391/407; their unchanged Java tests
pass and remain uncredited until translated. Overall goal remains active.


2026-10-04: Github362/391/407 and all four Github432 methods translated.
Original Java JUnit passes all seven methods unchanged. Nine original fixtures
(including Github407.dump) are byte-identical under test_vectors. Ports retain
all original assertions and runner defaults/overrides: six instance-scope output
substrings, zero-state stats, complete plain dump line/EOF comparison and
zero-uncovered checks, four config substitutions and exact symmetry-warning
parameters/cardinality or absence. Github432's exit assertion stays disabled
as in the source. Config substitutions use isolated temporary files.
Production fixes: StateWriter mirrors println(state.toString()), adding its
missing final newline; symmetry argument names come from SyntaxTreeNode's human
readable image and the same Permutations substring/config-alias logic as Java.
Five focused normal repetitions PASS11.192s. No assertions weakened, new tests
invented, or deferred-topic work added. Inventory697/1269(54.9%),262/626(41.9%);
572 pending contexts/364 classes (one partial class). Issue topic43/98 classes,
46/101 methods. Broad checks running: workspace23520, full TLC race85376,
focused root race88198 (five runs of all seven ports).
Logs `.codex-gotmp/github362432-{workspace,tlc-race,root-race}.log`.
Baselinef4ed912; do not commit until green. Next Github461, Github525, Github597.
Overall goal remains active with user-deferred topics unchanged.


Focused root race88198 PASS120.583s (five runs of all seven new methods); retire.
Workspace23520 and full TLC race85376 remain live. All new fixtures were compared
byte-for-byte; no implementation changes since broad checks launched.
Next Github461/525/597 unchanged Java JUnit all pass; compile log
correctness-java/github461525597-javac.log, original-<Class>-junit.log. Java597
process73424 terminal. No Go tests/fixtures/inventory credit for these next cases.
461 preserves full x=0..4 trace, exact assertion message and two-frame nested
stack, zero uncovered and inherited assertion exit. 525 preserves Error exit,
FINISHED/unsupported-liveness diagnostic/no GENERAL. 597 uses original dekker.tla
embedded config, no fixed fp/seed (noRandomFPandSeed=false), coverage/DOT false,
debugger/forced TE/JSON defaults true; stats4356/1500/0, Termination violation,
counterexample/state-print/back-to-state existence and inherited liveness exit.
Use the exact-arguments helper for597 rather than retaining fp0/seed1 defaults.
Overall goal remains active; next original tests are ready after this checkpoint.


Full workspace23520 PASS399.345s root/1.020s SANY/63.549s TLC; retire23520.
Focused root race88198 PASS120.583s. Full TLC race85376 remains live; no edits
since checks launched. All seven ports and both production corrections are
ready for a green commit after that last check. Goal remains active.


Final Github362/391/407/432 checkpoint: full TLC race85376 PASS568.462s;
retire85376. All final checks green: full workspace399.345s root/1.020s SANY/
63.549s TLC; full TLC race568.462s; five focused root race runs120.583s; five
focused normal runs11.192s. All nine fixtures match original bytes. Go/docs
whitespace clean; preserve any original fixture whitespace. Commit all seven
original-method ports and the plain-dump/symmetry-warning production fixes.
Inventory697/1269(54.9%),262/626(41.9%),572 pending contexts/364 classes.
All processes terminal. Next Github461/525/597 unchanged Java JUnit passed,
with source assertion/runner requirements documented above; no Go credit yet.
The full original correctness-port goal remains active, deferred topics unchanged.


2026-10-04: Github461/525/597/648/648wN/652 original methods translated.
All six unchanged Java JUnit tests pass. Eight model/config fixtures match source
bytes. Original assertions retained: assertion message/full five-state trace,
extended-state metadata/ordinals, exact two-frame stack and zero uncovered;
unsupported-formula diagnostic and Error exit; Dekker4356/1500/0 Termination
counterexample/state/back-to-state diagnostics with random fp/seed and original
coverage/DOT overrides; zero-state652; all35 coverage/count/cost rows for both
one and ten workers in648. Both70-line coverage literals were compared exactly
to the concatenated Java literals. No model-size/worker-count reduction.
Production fixes: ExitStatusForErrorCode maps only LIVE_FORMULA_TAUTOLOGY to77,
as Java EC does; unsupported/wrong-format formulas retain generic255. TLCEval's
class-wide lock is Java's reentrant nonfair RW lock; a plain Go RWMutex deadlocked
on nested constants. New internal reentrantReadWriteLock retains concurrent
readers, writer/read reentrancy, queued-reader writer preference, writer barging,
read-to-write blocking and source65535 acquisition limits. Installed OpenJDK
NonfairSync javap confirms writerShouldBlock=false and readerShouldBlock uses
apparentlyFirstQueuedIsExclusive; logreentrant-rw-reference.log. Demux evaluates
with Empty successor state, as Java's state-only overload does, preserving
LazyValue caching and exact coverage instead of re-evaluating argument wrappers.
All six ports five repetitions PASS34.515s after these fixes.
Original deadlock run48135 terminalFAIL600.073s; trace github648652-focused.log
proves RLock inside writer-held standardTLCEvalConst. Failed build runs9791/
23396/27915 terminal: filesystem exhausted (17MB free), not source/compiler
failure. Removed only generated Go cache entries older than one hour:129.03GiB;
130GiB available, source/test_vectors untouched. Cleanup3293 terminal. Focused
handles67352/93956/37439/91628 terminal; retire all.
Inventory703/1269(55.4%),268/626(42.8%),566 pending contexts/358 classes
(one partial class); issue49/98 classes,52/101 methods. Broad checks running:
workspace22755, full TLC race76601, focused root race37270 (five runs all six
plus original1198 A/B/C/D/F/H). Logs github461652-{workspace,tlc-race,root-race}.
Baselinea29bb49; do not commit until green. Next Github680a/b/c. Full goal active;
user-deferred topics unchanged. No invented tests or weakened source assertions.


Next Github680a/b/c unchanged Java JUnit all PASS; compiler loggithub680-javac.log
and original-<Class>-junit.log. Models remain only in ignored Java scratch; no
Go credit/active fixtures. Preserve original -config <model>.tla/-nowarning,
debugger=false, coverage/DOT/forcedTE/JSON defaults, inherited success exit,
FINISHED/no GENERAL, depth1, stats1/1/0 and exact UNCHANGED_VARIABLE_CHANGED:
a x at line7col20, b x at line7col22, c x at line8col22. Suppressed console
warnings still reach the recorder in Java. Current broad handles22755/76601/
37270 remain live; no production changes since their launch. Goal active.


Workspace22755 PASS399.676s root/1.036s SANY/63.897s TLC; retire22755.
Focused root race37270 PASS389.520s: five runs of all six new ports plus original
Github1198 A/B/C/D/F/H. Retire37270. Full TLC race76601 still live, no production
edits since checks launched. Disk129GiB available; cleanup resolved compilation.
Goal remains active; next680a/b/c original Java tests already pass.


Final Github461/525/597/648/648wN/652 checkpoint: full TLC race76601
PASS573.771s; retire76601. All final checks green: workspace399.676s root/
1.036s SANY/63.897s TLC; full TLC race573.771s; five focused root race runs
389.520s including all six new cases and1198 A/B/C/D/F/H; five normal runs
34.515s. Eight fixtures byte-identical; both35-line coverage literals exact.
Go/docs whitespace clean; preserve original fixture whitespace. Commit all six
faithful ports, generic reentrant RW guard, correct Empty-state demux overload
and source exit-code mapping. Inventory703/1269(55.4%),268/626(42.8%),566 pending
contexts/358 classes (one partial). All processes terminal. Next680a/b/c original
Java JUnit passed unchanged; no Go activation/credit yet. Full goal remains
active, user-deferred topics unchanged. Source/test_vectors untouched by cache
cleanup; only old generated Go cache entries were removed.

2026-10-04: Github680a/b/c and all eight Github687 variants translated.
All eleven unchanged original Java JUnit tests pass; all twelve model/config
fixtures match original bytes. Original assertions and runner settings retained,
including suppressed warnings and the 687 trace-generation overrides. Five
focused Go repetitions pass (15.534s); five focused race repetitions pass
(190.095s). Production now applies variable context
cutoff during primed lookup, shares the branched ENABLED context with the
changed-subscript action item, and keeps persistent TLCStateFun bindings for
instantiated variables without root-state vector slots. Constant-false config
specification reports TLC_CONFIG_SPEC_IS_TRIVIAL as Java does. No assertions
weakened and no invented tests. Inventory: 714/1269 logical contexts, 279/626
fully mapped concrete classes; 555 contexts/347 classes pending, one partial.
Full offline workspace passes (root 409.440s, SANY 1.067s, TLC 65.403s);
full TLC race verification passes (567.957s). All check handles terminal.
Next pending issue family after this verified batch: Github696/696b. User's
five deferred topics remain deferred. Goal remains active.

2026-10-04: Github696/696b and Github715/b/c/d original methods translated.
All six unchanged Java JUnit tests pass; nine original fixtures byte-identical.
Five focused Go repetitions pass (12.085s); five focused race repetitions
including680/687 pass (295.983s). Exact diagnostics, presence/absence,
depth/stats and inherited exits retained, with original coverage=false for715.
696 passes without production changes. Source inspection and715 failures fixed
config decomposition: semantic lookup retains unbound symbol identity; a bare
variable in SPECIFICATION reports TLC_CONFIG_OP_IS_EQUAL with its declaration
location and spec tag, while a property variable reaches level-based handling.
Constant-false property and nonboolean config diagnostics match source branches.
Source operator definitions now retain static SANY levels using the existing
SANY level analysis, independently of coverage and runtime overrides; previously
all nonbuiltin definition levels defaulted to zero, suppressing715 warnings.
Inventory now720/1269 contexts,285/626 complete classes;549 contexts/341 classes
pending, one partial. Full offline workspace passes (root423.974s, SANY0.999s,
TLC64.192s). Full TLC race passes (570.326s); all check handles terminal.
Next pending family: Github725 through725h;
all eight unchanged Java JUnit tests already pass in the scratch reference runner.
No invented tests or weakened assertions. Goal active; deferred topics unchanged.

2026-10-04: All eight Github725/b/c/d/e/f/g/h original testSpec methods
translated. All eight unchanged Java JUnit references pass, and15 model/config
files match original bytes. First focused Go run passes (5.947s) without further
production changes. Preserve all original runner overrides: noGenerateSpec=true,
coverage/debugger=false, default DOT, JSON enabled for725..725f and disabled
for725g/h. Original assertions retained, including soundness cases:725g has
Prop liveness violation,2/1/0 stats,depth1,full single-state trace/ordinal/action
and stuttering2;725h uses725g with725h.cfg,exact Inv initial-state text/newline
and safety exit. Five normal repetitions pass12.897s and five focused race
repetitions pass136.002s. Full offline workspace passes (root415.375s,
SANY0.967s, TLC61.806s); all check handles terminal. TLC production unchanged since
ad50e68, whose full TLC race passed570.326s. Inventory728/1269 contexts,293/626
complete classes;541 contexts/333 classes pending,one partial. Next missing issue
is Github726Test; unchanged Java references for726/742/743/746/757 already
pass in the scratch runner. Goal active and user-deferred topics unchanged.

2026-10-04: Github726/742/743/746/757 original testSpec methods translated.
All five unchanged Java JUnit references pass; nine fixtures match source bytes.
Five focused Go repetitions pass9.308s. Exact original assertions retained:
726 captures complete writes and uses list membership,742 rejects the full6×6
fairness formula,743 retains all four trace states/actions/ordinals,746 keeps
initial-state field order/newline,757 retains CounterExample postcondition and
explicit JSON/noGenerateSpec overrides. Production fixes: LNConj DNF product
uses Java Math.multiplyExact's signed32-bit overflow boundary and exact diagnostic
before allocating; Tool.getState(successor,predecessor) catches only direct
TLCRuntimeException from candidate equality and continues to other successors,
preserving other exception types and state-generation failures. Initial742 run
failed with Go out-of-memory from the missing overflow check; after correction,
743 exposed the missing catch/truncated trace, now fully passing. No weakened
assertions or invented tests. Five focused root race repetitions (new cases
plus725 and alias liveness) pass261.674s. Full offline workspace passes
(root420.629s, SANY1.039s, TLC63.961s). Full TLC race passes555.004s; all
check handles terminal.
Inventory733/1269 contexts,298/626 complete classes;536 contexts/328 classes
pending,one partial. Next pending issues: Github766 and Github766Simulate;
both unchanged Java references already pass in the scratch runner.
Goal active; user-deferred topics remain deferred.

2026-10-04: Github766/766Simulate/798I/798N/807 original testSpec methods
translated with original settings and assertions. All five unchanged Java JUnit
references pass; four fixtures match original bytes. Five focused Go repetitions
pass11.056s and five focused root race repetitions (new cases plus aliases)
pass124.266s. Github807 exposed a production shortcut: action decomposition used
the enclosing operator when state-level arguments prevented further splitting.
Java instead uses the applied operator declaration. Tool.collectActionsAppl now
preserves that declaration, restoring the exact Add action/location in the full
three-state trace and CounterExample/TLCGet("spec") postcondition. No assertion
weakened; no invented tests. Github766 retains its explicit JSON/simulation
arguments; both798 cases retain their exact counts/depth;807 retains debugger=false,
noGenerateSpec=true and no JSON.
Full offline workspace passes (root425.504s, SANY1.070s, TLC62.970s).
Full TLC race passes551.188s; all check handles terminal.
Inventory738/1269 contexts,303/626 complete classes;531 contexts/323 classes
pending,one partial. Next pending family Github817/b/c/d/e; all five unchanged
Java references already pass. Later819/849/858/866/971a/b/c/d/e references also
pass unchanged. Goal active; user-deferred topics remain deferred.

2026-10-04: Github817/b/c/d/e,819,849,858,866 original testSpec methods
translated with original fixture bytes, constructor overrides and assertions.
All nine unchanged Java JUnit references pass. Five combined focused Go
repetitions pass20.610s; five focused root race repetitions pass161.558s.
817 retains zero-uncovered coverage and complete action-property counterexample;
819 retains liveness-tautology/FAILURE_LIVENESS_EVAL;849 retains forced trace
specification/default JSON and absence of native-method override diagnostic.
858 preserves seed1 simulation, both JSON and tlcaction dumps, exact complete
state100/ordinal/PassToken(2) action and forced trace generation. Initial858
failure was missing original CommunityModules classpath; supplying the same
archive as Java's tool manifest fixed setup, with model/assertions unchanged.
866 keeps two workers and its full postcondition. It exposed a missing platform
property default: tlcLookupSystemProperty now supplies native file.separator
when absent, retaining explicit property/environment overrides and undefined
property fallback. Five focused866 repetitions pass5.764s. No invented tests or
weakened assertions. Nine new fixture files are byte-identical to Java.
Full offline workspace passes (root430.386s, SANY0.979s, TLC64.206s).
Full TLC race passes567.289s; all check handles terminal.
Inventory747/1269 contexts,312/626 complete classes;522 contexts/314 classes
pending,one partial. Next pending issue family Github971a/b/c/d/e; all unchanged
Java references pass. Scratch translation preserves original three/four-count
queue latches and all delegation through the existing runner queue injection;
not yet activated or credited. Goal active; user-deferred topics unchanged.

Next trace-topic preflight: unchanged DumpAsDotTest passes, including exact DOT
bytes, postcondition register42 and zero-uncovered coverage. Unchanged
DistributedTrace fails assertNoTESpec (trace spec generated) and inherited
SUCCESS0 versus actual safety12. Keep pending; do not weaken assertions or
invent a skip. Logs original-DumpAsDotTest-junit.log and
original-DistributedTrace-junit.log in ignored correctness-java scratch.

2026-10-04: Github971a/b/c/d/e original testSpec and beforeSetUp queue wrappers
translated. All five unchanged Java JUnit references pass. Six original model/
config fixtures match bytes. First focused Go run passes5.210s; five combined
normal repetitions pass8.589s and five root race repetitions pass80.632s.
Original two workers and three/four-dequeue CountDownLatch semantics retained:
enqueue into MemStateQueue before selected-state await; count down before each
single dequeue; delegate all other queue methods unchanged. Existing runner
StateQueue injection represents Java's testing-only Factory.sq; no production
changes needed. Original lncheck/config settings, no coverage/DOT/debugger/JSON/
trace generation, diagnostics, exact counts/depth where asserted, property names,
complete state/action/ordinal traces and loop-back2 assertions preserved. Go's
Value.Equal error is propagated as Java's throwing equals, not discarded.
No weakened assertions or invented tests. Full offline workspace passes
(root424.096s, SANY1.056s, TLC62.045s); all check handles terminal. TLC production
unchanged since9083c0d, whose full race passed567.289s. Inventory752/1269 contexts,317/626 complete classes;517 contexts/309
classes pending,one partial. Issue-regression topic now98/98 classes and101/101
contexts complete. Next eligible trace case DumpAsDotTest: unchanged Java passes,
including exact DOT bytes; DistributedTrace remains pending with two original
Java assertion failures, and two EWD840 DumpLoadTrace methods retain their
previously documented timing discrepancy. Goal active; user-deferred topics
unchanged.

2026-10-04: original DumpAsDotTest.testSpec translated, with unchanged Java
JUnit reference passing and three new fixtures plus reused CodeplexBug8
byte-identical. Original exact DOT master comparison,18/11/0 counts,
FINISHED/no GENERAL, postcondition absence checks, register42 firstInt18,
zero-uncovered coverage and liveness exit retained. All default coverage,
DOT/JSON/debugger/forced trace generation and explicit colorize/actionlabels/
stuttering options preserved. First Go failure exposed actual production
shortcuts: numeric rank-node sorting, insertion-order action legend, and an
early stuttering return that skipped rank maintenance/snapshot. StateWriter
now mirrors Java's HashMap/HashSet order, source label/legend formatting,
stuttering continuation and constructor header flush. dot_hash_map.go ports
only the operations DOT uses for Comparable Integer/Long/String keys, including
put vs computeIfAbsent insertion/resize timing, bucket splits, red-black
collision trees, tree root/list order and untreeification on resize. Source
OpenJDK21 HashMap and local JDK21 bytecode consulted; copyright/license retained.
Ignored CLI probes (not permanent tests or inventory credit) match Java byte
for byte for3374 iteration/lookup snapshots (10,433,564 bytes), including signed
Long/hash collisions, UTF16 String collisions, resizing and tree splits.
No master normalization, weakened assertions or invented permanent tests.
Final five normal repetitions pass5.679s; final five focused root race
repetitions (DumpAsDot, existing DOT cases and Github391) pass61.651s.
Constructor-flush review superseded earlier broad verification:41682/29419
terminated130;87639 passed61.667s before final flush and is retired. Final full workspace51679 passes (root416.794s, SANY0.983s, TLC62.560s);
retired. Final full TLC race63793 passes551.758s; retired. All final verification
handles terminal; ready for green commit. Inventory753/1269 contexts,318/626 complete classes;
516 contexts/308 classes pending,one partial. Next eligible topic: generated
TTrace correctness variants, preserving original generation/recheck pipeline;
first source Github461Test_TTraceTest and Github597Test_TTraceTest inspected.
DistributedTrace remains pending with original Java failures, two EWD840 binary
methods retain documented partial-check timing limitation. Goal active;
user-deferred topics unchanged.

First TTrace Java preflight: Github461Test then Github461Test_TTraceTest and
Github597Test then Github597Test_TTraceTest all pass unchanged. Generated
prerequisite files were explicitly checked present before recheck, so these
are executed references, not missing-artifact assumption skips. Logs
 ttrace-first-<Class>-junit.log; javac ttrace-first-javac.log;16957 terminal.
Original TTrace base uses generated class-name TTrace.tla as both model/config,
noGenerateSpec=true, coverage=false, default JSON/DOT/debugger, resolver with
original spec path.597 additionally omits fp/seed (random), disables DOT and
keeps debugger/JSON. Its original first model is dekker with embedded config;
source first-Go port already preserves random fp/seed. No Go TTrace port credit
yet; upcoming tests must generate Go artifacts then recheck, retaining exact
461 five-state trace/actions/ordinals and597 diagnostic/loop existence checks.

2026-10-04: Github461Test_TTraceTest.testSpec and
Github597Test_TTraceTest.testSpec translated in tlc_ttrace_first_java_test.go.
Each executes the fully asserted original Go model in a subtest, then rechecks
its actual generated monolithic TTrace artifact. No source-model rerun or Java
golden substitutes for the generated Go artifact. Original first-model asserts
are unchanged; their helpers now allow the class-name trace output path.
Original TTrace no-generation/no-coverage settings, one worker, debugger, JSON,
fp0/seed1 defaults retained;597 preserves random fp/seed and disabled DOT.
461 retains safety exit, all five states/actions/ordinals and zero-uncovered;
597 retains liveness exit, FINISHED/no GENERAL, temporal violation,
counterexample, trace and back-to-state existence. Original missing-artifact
assumption retained; all credited runs actually generated and rechecked files.
Initial Go setup failures came from the explicit resolver missing the standard
classpath and an absolute config argument instead of the original basename.
Resolver now has application classpath, generated user directory and original
spec library path; model/config arguments use the generated basename. Java's
ModelConfig also passes the basename unchanged to monolith extraction: no
production change was needed or made. No weakened assertions/invented tests.
All four unchanged Java first-model/recheck JUnit references pass. Final Go
first run passes12.393s; five combined normal repetitions pass78.915s; five
combined root race repetitions pass495.934s. Full offline workspace passes
(root431.224s, SANY0.939s, TLC63.233s).53599,56526,67264 terminal and retired.
TLC production remains b1b4a11, whose full TLC race passed551.758s.
Inventory755/1269 contexts,320/626 complete classes;514 contexts/306 classes
pending,one partial. Generated TTrace topic2/45 complete. Ready for green
commit; goal active, user-deferred topics unchanged.
Next-batch Java preflight: BugzillaBug279Test/TTrace and both
DepthFirstDieHardTest/TTrace and DepthFirstErrorTraceTest/TTrace pairs pass
unchanged, with generated prerequisites explicitly checked present. Initial
DieHard scratch run lacked DieHard.cfg and exited255; copying original model/
config bytes resolved setup. Logs ttrace-next-<Class>-junit.log and
 ttrace-next-javac.log in ignored correctness-java scratch;49187 failed setup,
2670 all remaining references pass; both terminal. These three upcoming TTrace
methods are not yet translated or credited. Preserve Bug279 deadlock-check/
no-DOT settings and original subset values; DFID first phases retain blank
source action labels, while TTrace rechecks assert exact generated _init/_next
locations and complete traces. DistributedTrace and two EWD840 binary cases
remain pending for the previously documented reasons.

2026-10-04: original BugzillaBug279Test_TTraceTest.testSpec,
DepthFirstDieHardTest_TTraceTest.testSpec and
DepthFirstErrorTraceTest_TTraceTest.testSpec translated in
 tlc_ttrace_subset_depth_first_java_test.go. Each fully asserted original Go
model runs first, then its actual generated class-name TTrace artifact is
rechecked. Original Bug279 deadlock/no-DOT, lazy subset,3/3/0 stats, complete
trace/actions/ordinals and zero-uncovered retained. Original DFID first phases
retain7/9 depths and blank action labels; BFS TTrace rechecks preserve all7/8
states and exact _init/_next locations. No assertions weakened or invented
permanent tests. TTrace helper now mirrors Java isFile assumption exactly and
supports original deadlock override; existing defaults remain identical.
Initial Go focused6 tests passed before production review. Extra JSON postcondition diagnostics revealed a
production shortcut despite these original methods passing: modules_json.go
manually enumerated lazy values instead of Java's value.toSetEnum. This changed
allocation growth/bounds, cache, normalization/order and coverage semantics.
JSON value/array branches now call existing toSetEnumValue; manual fallback
removed. Original Java Bug279 first phase reaches heap-space error with256MiB
heap; recheck succeeds. Go falsely stopped atValueVec1M growth bound although
Subset.toSetEnum allocates exact1048576 size asJava. Go's fixed conversion now
serializes the whole subset without false limit error; native Go heap can
complete the original first phase too. No JSON/DOT/coverage setting disabled.
Final focused first run PASS20.836s; five normal repetitions including existing
TestJavaJson PASS113.829s. These finished38312/65220 handles are retired.
Full TLC race31965 passes560.577s; terminal and retired. Focused root race8012
(five new TTrace+Json+461TTrace repetitions,45m outer timeout for the full subset
under instrumentation) passes745.825s; terminal and retired. Every selected
method passes all five repetitions, with no races or assumption skips.
Full offline workspace38419 passes (root445.265s, SANY1.082s, TLC60.674s);
terminal and retired. All validation handles are terminal.
All six reused model/config fixtures verified byte-identical to Java originals. Inventory756/1269 contexts,321/626 complete
classes;513 contexts/305 classes pending,one partial;TTrace5/45.
Next first-phase review discovered EvalExceptionTest and PrintTraceRaceTest
omit inherited doDumpTrace=true. Reclassified both existing methods as
Reconcile and removed two contexts/classes from confirmed counts pending
correction. Restore original JSON dumps before translating their TTrace
variants; retain all existing assertions and source settings (including four
workers for PrintTraceRace). No code changes to those methods yet.
Current three new ports and JSON production fix ready for green commit.
Next reference preflight: unchanged EvalExceptionTest/TTrace and
PrintTraceRaceTest/TTrace allpass;85741 terminal. Unchanged RandomElementTest,
RandomElementT4Test,RandomElementSimulationTest,RandomElementXandYTest and all
four TTrace counterparts allpass;46715 terminal. Generated prerequisites
explicitly checkedpresent (no missing-artifact skips credited). Logs
 ttrace-eval-race-<Class>-junit.log and ttrace-random-element-<Class>-junit.log,
corresponding javaclogs in ignored correctness-java scratch. These six upcoming
methods are not yet translated/credited. Goalactive,userdeferredtopicsunchanged.

2026-10-04: EvalExceptionTest and PrintTraceRaceTest original first phases
reconciled by restoring inherited doDumpTrace=true, with all assertions intact.
Their helpers now permit class-name TTrace output paths. Original evaluation
coverage=false, JSON/DOT/debugger/fp0/seed1 settings and ERROR exit retained;
PrintTraceRace keeps four workers, coverage/DOT/JSON, no debugger and
FAILURE_SAFETY_EVAL exit. First restored-phase gate passes1.632s (62837 terminal).
An earlier scripting syntax error made no edits;21699 ran the old tests and is
not credited as verification of restored settings. Both handles are retired.
New EvalExceptionTest_TTraceTest.testSpec and
PrintTraceRaceTest_TTraceTest.testSpec translated in tlc_ttrace_eval_race_java_test.go.
Both execute the fully asserted original model first and recheck the actual
Go-generated artifact. Eval preserves FINISHED,6/6/0 stats, complete six-state
trace/actions/ordinals and safety exit. PrintTraceRace preserves four workers,
FINISHED,2/2/0 stats,no GENERAL,behavior diagnostic, original first-two-state/
ordinal and two-field payload shape assertions, and safety exit. Shared TTrace
helper now preserves workers and Java's debugger-only-with-one-worker condition;
JSON/DOT/no-generation/no-coverage and original missing-artifact assumption stay.
Initial recheck run98770: PrintTraceRace passes; Eval fails exact record field/
function order despite matching state counts and first-phase trace. Actual
production shortcut: MCState consumed declaration-ordered state values while
Java TLCState.getVals builds a HashMap. That altered the generated invariant's
variable order and the recheck's record-name intern tokens. Ported GetVals with
all declared keys, including null/unassigned values; MCState now consumes its
key-set order, while RecordValue(TLCState) retains declaration-order construction.
Generalized existing DOT HashMap helper into java_hash_map.go, preserving source
bucket/list/resize/tree behavior and adding non-Comparable-key subtree search,
identity tie-break and duplicate search. UniqueString hashes use Java UTF16
String.hashCode; its class is non-Comparable, with native Go object identities
for tie-breaking. Tree equality lookup searches both subtrees even when identity
hashes collide. Source Java TLCState/MCState, cached OpenJDK21 source and local
JDK21 bytecode consulted; notices/licenses retained. No field pre-intern hack,
rewritten trace master, weakened assertions or invented permanent tests.
Final focused4 PASS3.903s (68947 terminal); five normal repetitions of all four
plus DumpAsDot and Github461TTrace PASS24.533s (57000 terminal); five corresponding
root race repetitions PASS179.842s (33558 terminal). Full offline workspace
PASS root456.441s,SANY0.949s,TLC65.141s (87703 terminal). Full TLC race PASS557.944s
(30057 terminal). All verification handles retired; ready for green commit.
All four reused fixtures byte-identical. Unchanged Java first/recheck references
for both pairs previously passed and generated prerequisites were checked present.
Ignored CLI probes (not permanent tests/inventory credit): all3374 prior DOT
iteration/lookup snapshots (10,433,564bytes) still match Java exactly after helper
changes (57581 terminal). Additional non-Comparable probe captures actual Java
identity hashes, including an actual collision (keys62354/105841,hash2134400190),
then replays insertion/update/lookup with those identities in Go. All256 snapshots
(280,259bytes) match byte-for-byte, including tree collisions and resize splits.
Inventory760/1269 contexts,325/626 complete classes;509 contexts/301 classes
pending,one partial;TTrace7/45 complete. Both first phases restored to Port
complete, and both new rechecks marked Port complete. Goal active; deferred
Debugger/scoped identifiers,Checkpoint/recovery,Distributed TLC,JPF and
Benchmarks/supporting fixtures unchanged. Next eligible batch: four random-element
TTrace variants. Unchanged first/recheck Java pairs for RandomElementTest,
RandomElementT4Test,RandomElementSimulationTest and RandomElementXandYTest already
pass (46715 retired; ttrace-random-element-<Class>-junit.log). Their existing Go
first phases already explicitly preserve JSON; retain all seeds, workers,
simulation/debugger settings and assertions when factoring generation helpers.
Those four upcoming rechecks remain missing/uncredited.

2026-10-04 current batch: RandomElementTest_TTraceTest.test,
RandomElementXandYTest_TTraceTest.test,
RandomElementSimulationTest_TTraceTest.test and
RandomElementT4Test_TTraceTest.test translated in
 tlc_ttrace_random_element_java_test.go. Each fully asserted original model
runs first, followed by a recheck of its actual generated class-name Go TTrace
artifact. Existing first-phase methods only factor helpers accepting output-path
arguments; all assertions and source settings are unchanged. Preserve exact
seeds8006803340504660123,8006642976694192746 and15041980; simulation num=1 and
first-phase debugger=false; T4 uses four workers in both phases. Recheck defaults
remain fp0/seed1, JSON/DOT, no generation/no coverage, debugger for one worker.
Single-worker variants preserve exact11/3-state traces, named actions/ordinals,
FINISHED/no TLC_BUG, behavior diagnostic and zero-uncovered; RandomElement also
preserves11/11/0 stats. Simulation recheck asserts every original generated
_init line25..26 and_next line30..36 location. T4 preserves all11 y/x component,
bound and ordinal assertions, including fresh getVals snapshots as in Java.
No weakened assertions, substitute source-model reruns or invented tests.
All four unchanged Java first/recheck pairs previously passed, with generated
prerequisites explicitly verified present (46715 retired). Java checkout still
8f4bc8b73ad1202774a6bf70143436f8ba50aab0. All four reused fixtures verified
byte-identical; no new persistent fixtures. Final first Go8 tests PASS7.274s
(29003 terminal); five combined normal repetitions PASS21.803s (10317 terminal);
five combined root race repetitions PASS195.069s (80905 terminal). All three
handles retired. Full workspace PASS448.840s root/1.036s SANY/64.395s TLC
(67309 terminal and retired). All required checks green; ready to commit.
TLC production unchanged since8812d14, whose full TLC race passes557.944s.
Inventory764/1269 contexts,329/626 complete classes;505 contexts/297 classes
pending,one partial;TTrace11/45 complete. Goal active; deferred topics unchanged.
Next reference preflight: unchanged RandomSubsetATest/TTrace,
RandomSubsetBTest/TTrace,RandomSubsetNextTest/TTrace,
RandomSubsetNextT4Test/TTrace and RandomSubsetTest/TTrace all pass.74457 terminal;
logs ttrace-random-subset-<Class>-junit.log and
 ttrace-random-subset-javac.log in ignored correctness-java scratch. Actual
prerequisite files checked present before each recheck; no assumption skips
credited. These five upcoming TTrace contexts are not translated/credited.
Preserve both inherited RandomSubset_TTrace parameter tuples, original seeds,
first-phase worker/debugger/JSON settings, complete traces and source literal
bounds (including firstX in the original y upper-bound expression).

2026-10-04 current batch after24c2e4e: five original RandomSubset TTrace contexts
translated in tlc_ttrace_random_subset_java_test.go: RandomSubsetATest_TTraceTest,
RandomSubsetBTest_TTraceTest (inherited RandomSubset_TTrace.testSpec),
RandomSubsetNextTest_TTraceTest, RandomSubsetNextT4Test_TTraceTest and
RandomSubsetTest_TTraceTest. Each first phase runs every original assertion and
then the actual Go-generated class-name artifact is rechecked. Parent helpers
only accept output-path arguments; assertions/settings unchanged. Preserve
seeds15041980/918347981374, full two/eleven-state traces/actions/ordinals,
FINISHED/no GENERAL or TLC_BUG as originally asserted, exact init/stats/depth,
four-worker settings without debugger, JSON/DOT/no generation/no coverage,
inherited safety exit and original coverage assertions. RandomSubsetTest
preserves exact _init line31..35/_next line39..51 generated source locations,
getVals snapshots, all component/tuple bounds and UNCHANGED comparisons,
including firstX in the original y upper-bound expression. No weakened tests,
substitute source-model rechecks or invented persistent tests. All five Java
first/recheck reference pairs pass (74457 retired); actual artifacts verified
present. Four reused fixtures byte-identical; Java revision unchanged.
First focused all10 PASS150.857s (86856 terminal/retired). Five normal
four-worker recheck repetitions PASS110.829s (9631 terminal/retired).
Focused race all5 rechecks PASS165.241s (27263 terminal/retired), each including
its fully asserted original phase. No production changes; full TLC race
baseline8812d14 PASS557.944s remains applicable. Full workspace PASS513.155s
root/0.984s SANY/64.227s TLC (54194 terminal/retired). All required checks green;
ready to commit. Next-source audit found missing inherited JSON trace dump in
older TLCGetLevelTest and TraceWithLargeSetOfInitialStatesTest ports; both
conservatively changed to Reconcile and removed from numerator until restored.
Adjusted inventory767/1269 contexts,332/626 complete classes;502 contexts/294
classes pending,one partial;TTrace16/45 complete. Goal active; deferred topics
unchanged. Next eligible batch: restore those two parent setups and port their
TTrace rechecks plus ViewMapTest_TTraceTest. All three unchanged Java first/recheck
pairs pass (24269 terminal/retired; ttrace-level-view-<Class>-junit.log), with
actual prerequisite artifacts verified present. Those three TTrace contexts
remain missing/uncredited. Shared Go recheck helper needs extra CLI arguments
for original -maxSetSize10 and -view; preserve complete source assertions,
JSON/DOT/debugger, exact generated action locations and liveness stuttering5.

2026-10-04 current batch after4db4541: restored inherited JSON trace-dump setup
in TLCGetLevelTest and TraceWithLargeSetOfInitialStatesTest without changing any
original assertions; factored fully asserted generation helpers for those two
and ViewMapTest. Ported TLCGetLevelTest_TTraceTest.testSpec,
TraceWithLargeSetOfInitialStatesTest_TTraceTest.testSpec and
ViewMapTest_TTraceTest.testSpec in tlc_ttrace_level_view_java_test.go. Each runs
the fully asserted original phase then rechecks its actual Go-generated class-name
artifact, retaining original JSON/DOT/debugger/no-generation/no-coverage settings.
Shared recheck helper now accepts source extra CLI arguments after base settings,
preserving -maxSetSize10 and -view. Level recheck retains liveness exit,
FINISHED/4/4/0/no GENERAL, temporal/counterexample/trace diagnostics, complete
four-state/actions/ordinal trace and stuttering5. Large-initial-state recheck
retains safety exit, FINISHED/no GENERAL/no TLC_BUG, behavior diagnostic,
full two-state trace/ordinals, exact _init line25..26/_next line30..36 source
locations and zero-uncovered. View recheck retains safety exit, FINISHED/no
GENERAL/no TLC_BUG, behavior diagnostic, full eight-state trace including pc,
ordinals and exact _init line27..29/_next line33..41 source locations. No
weakened assertions, substitute source-model reruns, new skips or invented tests.
All three unchanged Java original/recheck pairs pass (24269 retired), with actual
prerequisite files verified present. Java revision8f4bc8b73ad1202774a6bf70143436f8ba50aab0
revalidated. Six reused fixtures byte-identical. Focused all6 PASS5.742s
(9450 terminal/retired); five normal repetitions of all6 plus prior Github461
recheck PASS22.992s (97208 terminal/retired); five root race repetitions of same
PASS205.157s (11032 terminal/retired). All20 topic totals independently agree with
source/method inventory:772/1269 contexts,337/626 complete classes;497 contexts/
289 classes pending,one partial;TTrace19/45 complete. No production changes;
full TLC race baseline8812d14 PASS557.944s remains applicable. Full workspace
PASS509.150s root/1.148s SANY/61.547s TLC (35459 terminal/retired). All required
checks green; ready to commit. Goal active, deferred topics unchanged.
Next eligible batch: all10 NoFairnessButLiveProp*Test.testSpec models (no TTrace
first-phase prerequisites left among the remaining ordinary tool/ variants).
All10 unchanged Java JUnit reference tests pass (87998 terminal/retired,
no-fairness-<Class>-junit.log and no-fairness-javac.log in correctness-java scratch).
They all override coverage=false but retain assertZeroUncovered, original
JSON/DOT/debugger/forced generation, FINISHED/3/2/0/depth2 and source-specific
SUCCESS or VIOLATION_LIVENESS exits. Preserve original exact warning-presence/
absence assertions, including P's no temporal violation and deferred fairness
warning. Shared NoFairnessButLiveProp.tla embeds module H; copy exact full original
module and all10 configs under test_vectors. Those contexts remain missing and
uncredited. Deferred checkpoint TTrace remains skipped; other pending TTrace
classes need their liveness/simulation original phases ported first.

2026-10-04 current batch after449bbff: all10 original NoFairnessButLiveProp*Test
contexts translated in tlc_no_fairness_models_java_test.go, with complete58
source assertions and original configs/exits. All override coverage=false but
retain zero-uncovered; preserve default JSON/DOT/debugger/forced generation,
FINISHED/3/2/0/depth2 and exact warning presence/absence, including P's no temporal
violation and deferred fairness warning on successful checking. Full original
NoFairnessButLiveProp.tla (including embedded H) and all10 configs copied byte
exact to test_vectors. All10 unchanged Java JUnit references pass (87998 retired).
First focused Go PASS6.629s (56683 retired); final literal-config setup retains
exact source arguments, with five normal repetitions PASS18.078s (70305 retired)
and five root race repetitions PASS169.524s (97688 retired). No production changes.
While broad verification ran, ported all5 pending PossibleTest/PossibleFail*Test
contexts in tlc_possible_models_java_test.go and restored omitted inherited JSON
trace dump in older PossibleCountsTest. Preserve complete original tests,
constructor configs, inherited SUCCESS or VIOLATION_ASSUMPTION exits,
JSON/DOT/coverage/debugger/forced generation and every diagnostic assertion.
Possible failure tests retain exact first-record/first-parameter BigJump or
Unreachable predicate names; NoTrans retains no transitions and Mixed retains
the source's Unreachable priority. PossibleCounts retains all6 assertions;
PossibleTest retains FINISHED/no GENERAL/no postcondition-false/no unwitnessed.
All6 unchanged Java references pass (20807 retired; possible-<Class>-junit.log).
Seven new Possible fixtures plus two existing counts fixtures byte-identical.
Possible focused6 PASS5.904s (65695 retired); five normal repetitions PASS10.369s
(69733 retired); five root race repetitions PASS115.192s (42540 retired).
No weakened assertions, substitute fixtures, new skips or invented tests.
Initial broad run for NoFairness-only snapshot PASS507.273s root/1.055s SANY/
63.580s TLC (73630 terminal/retired). Because five tests and restored setup were
added after that compilation, final broader run for all current source PASS
509.437s root/0.997s SANY/62.756s TLC (34744 terminal/retired). All required
checks green; ready to commit. Full TLC race baseline8812d14 PASS557.944s remains
applicable because production unchanged. All20 topic totals verified against
source/method mapping:787/1269 contexts,352/626 complete classes;482 contexts/
274 classes pending,one partial;TTrace19/45 complete. Goal active; deferred topics
unchanged. Next eligible batch: original BidirectionalTransitions1BxTest and
BidirectionalTransitions1ByTest inherited BidirectionalTransitions1BTest.testSpec,
then their actual generated TTrace rechecks inherited BidirectionalTransitions1B_TTrace.
All4 unchanged Java first/recheck references pass (58617 terminal/retired;
bidirectional1b-<Class>-junit.log and bidirectional1b-javac.log). Actual generated
prerequisites checked present before rechecks. Those4 contexts remain missing/
uncredited. Retain original JSON/DOT/coverage/debugger/forced generation settings,
liveness exits, FINISHED/no GENERAL,13/3/0 original and4/3/0 recheck stats,
property names Prop1Bx/Prop1By, all3 exact states/actions/ordinals x0,x2,x1 and
original back-to-state1 assertions (first recorded parameter, without an
additional action-label assertion).

2026-10-04 current batch after9297572: all6 original BidirectionalTransitions
model contexts and all4 generated TTrace rechecks translated in
 tlc_bidirectional_liveness_java_test.go. Successful variants preserve exact
13/3/0/depth2 and9/4/0/depth3, FINISHED/no GENERAL, SUCCESS and zero-uncovered.
1Bx/1By and2Cx/2Cy original phases retain every inherited assertion, source
configs (2C arguments without .cfg suffix), liveness exits, exact13/3/0 or9/4/0,
property names Prop1Bx/Prop1By/Prop2Cx/Prop2Cy, temporal/counterexample/trace,
complete3/4-state traces/actions/ordinals and first-record loop-back1. Rechecks
run fully asserted original phases, then actual generated class-name Go artifacts
with source JSON/DOT/debugger/no-generation/no-coverage settings and4/3/0 or5/4/0
stats, full traces and loop-back1; no extra property-name/action-label assertion
added where originals omit it. All10 unchanged Java references pass (58617 and
23512 retired; bidirectional1b/rest-<Class>-junit.log). Seven total fixtures byte
exact (five newly copied, two reused). Focused10 PASS5.959s (95121 retired), five
normal repetitions PASS24.531s (94213 retired), five root race repetitions
PASS242.389s (67894 retired). No production changes.
During broad verification, ported ChooseTableauSymmetryTestA.testSpec and its
actual generated TTrace recheck in tlc_choose_tableau_symmetry_java_test.go.
Preserve source model/config/imports, liveness exits, FINISHED/no GENERAL,
13/6/0 first phase and6/5/0 recheck, both original violated-property aliases,
temporal/counterexample/trace diagnostics, all five complete states/ordinals,
every exact source/generated action label, loop-back3 and its exact action label,
zero-uncovered and original JSON/DOT/debugger/coverage/generation overrides.
Three new fixtures byte-identical; retain original MCa.tla trailing whitespace
on lines10/15 rather than normalize source. Upstream ignored ChooseTableauSymmetryTest
disposition unchanged. Both unchanged Java references pass (41082 retired),
actual prerequisite file verified present. Focused2 PASS4.682s (11900 retired),
five normal repetitions PASS6.792s (5965 retired), five root race repetitions
PASS66.415s (4236 retired). No weakened tests, new skips or invented tests.
Initial Bidirectional-only broad snapshot PASS519.013s root/1.083s SANY/65.022s
TLC (20057 terminal/retired). Because two symmetry methods were added after its
compilation, final current-source broad PASS518.007s root/0.892s SANY/62.904s
TLC (60114 terminal/retired). All required checks green; ready to commit. Production unchanged since8812d14, verified git diff tlc/*.go empty; full TLC
race baseline557.944s remains applicable. All20 topic totals agree with source/
method mapping:799/1269 contexts,364/626 complete classes;470 contexts/262 classes
pending,one partial;TTrace24/45 complete. Goal active; deferred topics unchanged.
Next eligible batch: CodePlexBug08 and CodePlexBug08a, AgentRing/AgentRing790,
EWD840FL1..4 and their generated rechecks. Do not port/run the FromCheckpoint
class while checkpoint/recovery is deferred. Java reference preflight59633
terminal/retired: all unchanged15 classes pass (no checkpoint), with actual
generated prerequisites checked before rechecks; logs codeplex08-<Class>-junit.log, codeplex08-javac.log.
Those15 contexts remain missing/uncredited. Preserve original constructor path
CodePlexBug08, complete trace/action/ordinal/property-alias/stuttering/loop-back
assertions and source defaults; AgentRing790 overrides coverage=false, SUCCESS,
FINISHED/TLC_SUCCESS/no temporal violation. All remaining originals have liveness
exits. Source fixture names/paths discovered via rg; model inputs copied only to
ignored Java scratch for preflight, no new persistent CodePlex fixtures this turn.
