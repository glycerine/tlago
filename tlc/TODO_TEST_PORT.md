# Original TLC test-port inventory

Snapshot: 2026-10-06. Java source checkout: `8f4bc8b73ad1202774a6bf70143436f8ba50aab0`; Go baseline: `5946d3e`, with the current working tree inspected. This is a source-to-source inventory, not a new test run or a declaration of full Java behavioral parity.

## Scope and counting

The main inventory covers every test-bearing concrete class in `../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/`, including JUnit 3 tests, annotated JUnit 4 methods, inherited methods, and concrete subclasses with no test methods of their own. Abstract bases and support classes are dependencies, not additional runnable test classes. Separate appendices cover shared `test/util/`, `test-long/`, `test-concurrent/`, `test-verify/`, and benchmarks.

- **617 of 626 non-`@Ignore` concrete classes have all their logical test methods mapped (98.6%).**
- **1,260 of 1,269 non-`@Ignore` logical test-method contexts have confirmed translations (99.3%).** A context is `(concrete Java class, method)`; inherited methods count once for each concrete subclass.
- **9 logical method contexts across 9 classes remain to port or reconcile.** Related older Go checks may cover parts of the remaining cases.
- There are 660 test-bearing concrete classes in the source tree overall. 34 are wholly disabled by `@Ignore`; 39 ignored method contexts are tracked separately below and excluded from the percentages.

Model-test setup reconciliation (2026-10-06): the common Go runner now retains
its parsed fingerprint configuration and applies the original Ant off-heap /
512 KiB settings before parsing. The previous small-MSB replacement was a
translation shortcut. This correction changes no original assertions and adds
no method-count credit; verification receipts belong in `PORT_PROGRESS.md`.

These are conservative translation counts. Credit requires an explicit original-method translation or an inspected match of the original inputs/assertions; merely testing the same Go type, running a model manually, having an ephemeral Java comparison, or copying a fixture is insufficient. Related Go tests whose full original cases/assertions have not been reconciled remain on the checklist. The percentages measure test translation, not TLC implementation completion or source-code coverage.

Methods are counted **before parameter expansion**, and non-`@Ignore` does not imply every upstream Ant target executes them: assumptions, target exclusions, platforms, and JVM requirements still apply. For example, `OffHeapIndexerEquivalenceTest.testInfiniteInfMult` has **7,254 parameter rows**; each of the three concrete `OffHeapIndexerParameterizedTest` subclasses inherits five methods over **1,104 rows** (16,560 contexts before assumptions). `GetScopedIdentifiersTests` has all 18 rows ported. Preserve complete matrices rather than substituting a few samples.

Excluded from this TLC count: SANY, PlusCal, formatter, Toolbox UI suites, and CommunityModules (a separate project). SANY `semantic.TestLevelChecking.testAll` is **Port complete** with all 51 original rows and separate generation/level assertions; see [the SANY test-port notes](../sany_tests/README.md). SANY `semantic.NestedModuleInstanceTest.testTopLevelInstanceOfNestedModule` is also **Port complete**; its LET-instance sibling retains the original Java ignore. SANY `semantic.SemanticCorpusTests.test` is **Port complete** in [the root translation](../sany_semantic_corpus_java_test.go), retaining all 28 original parameter rows, the NegativeOpTest assumption, canonical reference/comment assertions and checked levels. The older AST facade checks remain supplementary. All three SANY `semantic.TestSubexpressionSelectors` methods are **Port complete** in [the root translation](../sany_subexpression_selectors_java_test.go), with the original generation-only helper and exact error-code/message/location assertions. All five methods of `semantic.IncrementalSemanticParseTests` are **Port complete** in [the root translation](../sany_incremental_semantic_java_test.go), including both standalone LET methods with actual dependency/module level checks, syntax identity, concrete graph classes and imported source-reference assertions. SANY `semantic.TestBuiltInOperatorInitialization.testInitAndReInit` is **Port complete** across all 72 properties and both global-context passes. SANY `xml.TestDecimalXMLExport.test` is **Port complete** with its original fixture/assertions and source numeric metadata. This adds no TLC inventory credit. The complete CommunityModules Ant test target already has its own Go translation in [community_modules_java_test.go](../community_modules_java_test.go). Email reporting and dependencies pursued for email remain excluded under the user’s scope directive.

SANY lexer production reconciliation (2026-10-07): candidate scanning is replaced
by the generated Java DFA/NFA over the actual UTF-16 stream. The existing original
TokenizerTests remain translated and pass; this production correction adds no
original-method credit. Bounded Java comparisons and their limits are recorded in
`PORT_PROGRESS.md`.

SANY parser reconciliation (2026-10-07): **Port complete** for
`parser.BelchDefTests.runTestCase` (all five original rows) in
[sany_belchdef_java_test.go](../sany_belchdef_java_test.go), and all three
`parser.IncrementalSyntaxParseTests` methods in
[sany_incremental_syntax_java_test.go](../sany_incremental_syntax_java_test.go).
The original token-boundary and standalone parser assertions are now retained;
the earlier module-wrapper checks remain supplementary. No change to the main
TLC method/class totals. All 74 generated-lookahead entry points now have source
production callers, and all 130 direct-choice sites have translated paths.
Native failure-span estimation is removed. Ordered expected-token entries match
Java on 341 external cases (14,741 sequences across 117 nonempty results).
Exhaustive malformed rescan contexts and corpus AST comparison remain separate
implementation requirements; these observations add no original-method credit.

SANY syntax corpus: both `TlaPlusSyntaxCorpusTests.testAll` and
`testAllTlaPlusNodesUsed` are **Port complete**, retaining 355 original parameter
contexts each. Full recursive SANY-to-DSL translation, AST equality and source
known-failure/error handling are installed. All 355 actual translated-output rows
match Java; the original Java class passes all 710 contexts. Source metadata and
expected trees also match across all unchanged cases. Main TLC totals are unchanged.

SANY fidelity audit (2026-10-08): all 25 findings in
[SANY_TESTS_TO_FIX.md](../SANY_TESTS_TO_FIX.md) are repaired. All 59 affected
original method contracts are restored, including the full XML schema/command/
library/error paths. Dedicated SANY and root SANY tests pass. All 227 mirrored
fixtures and the current embedded XSD match pinned upstream bytes. XML
validation requires `xmllint` (libxml2) on PATH; it cannot silently skip when
required. Original ignored/empty methods remain explicitly identified, and the
main TLC inventory totals do not change. Full normal verification passes with zero failures. Final inventory reconciles
all 96 primary methods: 95 pass (including one empty source body), one upstream
ignore; this does not claim full SANY parity beyond the source assertions.

## Topic totals for the main suite

| Topic | Non-ignored classes | Classes fully mapped | Logical methods | Mapped methods | Pending methods |
| --- | ---: | ---: | ---: | ---: | ---: |
| CLI, REPL, messages, and trace-spec output | 5 | 5 | 37 | 37 | 0 |
| Presentation models | 5 | 5 | 13 | 13 | 0 |
| Standard modules, constants, native overrides, and random values | 44 | 44 | 75 | 75 | 0 |
| Evaluation, initial states, next states, and action composition | 45 | 45 | 51 | 51 | 0 |
| Safety checking, diagnostics, and checker lifecycle | 34 | 31 | 34 | 31 | 3 |
| Issue regressions in the evaluator and checker | 98 | 98 | 101 | 101 | 0 |
| Traces, aliases, dump/load, and generated trace specs | 17 | 16 | 48 | 47 | 1 |
| Generated TTrace recheck variants | 45 | 45 | 45 | 45 | 0 |
| Liveness and fairness model regressions | 101 | 100 | 101 | 100 | 1 |
| Liveness graph, tableau, and expression helpers | 6 | 6 | 48 | 48 | 0 |
| Simulation and multithreaded simulation | 20 | 20 | 55 | 55 | 0 |
| Coverage | 20 | 20 | 23 | 23 | 0 |
| Debugger and scoped identifiers | 16 | 16 | 35 | 35 | 0 |
| Checkpoint and recovery models | 2 | 2 | 2 | 2 | 0 |
| Distributed TLC | 11 | 7 | 41 | 37 | 4 |
| Fingerprint sets, indexers, arrays, and iterators | 20 | 20 | 165 | 165 | 0 |
| Queues and pool writers | 2 | 2 | 11 | 11 | 0 |
| Values, lazy functions, enumeration, and value streams | 17 | 17 | 189 | 189 | 0 |
| Collections, buffered files, combinatorics, and statistics | 14 | 14 | 91 | 91 | 0 |
| Numbered legacy model suite | 104 | 104 | 104 | 104 | 0 |
| **Total** | **626** | **617** | **1,269** | **1,260** | **9** |

## Porting rules and proposed order

Current priority (2026-10-06): development returned to `master`. Postpone the
new rpc25519/Tube job service until the remaining faithful Java TLC port is
complete. Core checkpoint model ports are now complete; pursue remaining Java
parity work without importing the unfinished distributed-service implementation.
The older topic deferrals below describe the previous work order.

Current user priority (2026-10-03): first finish the pending production work,
get the suite green, and commit all current changes; then port original
correctness tests. Temporarily skip **Debugger and scoped identifiers**,
**Checkpoint and recovery models**, **Distributed TLC**, **JPF concurrency
verification**, and **Benchmarks and supporting fixtures**. Keep their inventory
entries visible. If a test fails, check the mechanical translation and inspect
production TLC for shortcuts; implement the missing behavior before moving on,
without weakening the original assertions. Mark finished entries **Port complete**.

Testing directive (2026-10-05): run long workloads normally with their complete
original bounds. Reserve `-race` for short, focused concurrency checks; never
combine it with long workloads or broad selections that include them.

1. Reconcile the remaining source-failing and JVM-specific assertions. `DumpLoadTraceTest` has all methods translated, including both enabled EWD840 binary methods, but the full root gate exposed intermittent JSON auto-worker round-trip failures in `testSafetyDumpLoadTraceJSONAutoWorkers` and `testSafetyDieHardAliasSub2DumpLoadTraceJSONAutoWorkers`. Resolve their production behavior before treating this class as green; retain original worker counts and prefix assertions. Reconcile any older value/module/utility checks with complete original inputs and assertions.
2. Port small direct value, collection, module, and output tests against their existing production implementations. Fix actual production gaps before completing the affected original method.
3. Port numbered suite/evaluation-order and checker/model regressions in coherent topics, then simulation and liveness models; checkpoint/recovery remains deferred. Carry constructor settings and inherited coverage/exit/trace assertions with each test.
4. Port generated TTrace variants after their original run and trace-spec pipeline work accurately. Preserve ordering and generated-artifact dependencies.
5. Deferred for now: complete distributed wire transport before translating full remote integration tests. Keep the four upstream assumption-disabled transport models visible in the backlog.
6. Port fingerprint parameter matrices and long/concurrent suites with their original semantics; schedule resource-heavy execution separately. Keep JPF verification and benchmarks deferred. Give JVM-specific inlining tests an explicit disposition instead of claiming they are ordinary Go test ports.

Translate the **whole original Java method** after its feature is in Go, including expected exceptions, setup/teardown, constructor options, random seeds, parameter rows, inherited assertions, and fixture bytes. Do not invent substitute regression/unit tests or weaken assertions to get green results. Persistent fixtures belong in `tlc/test_vectors/`. Upstream-only skips must retain their original reason. The initial inventory added documentation only; subsequent status changes record complete original-method translations.

## Missing or incomplete main-suite tests, by topic

Every unchecked entry below names the exact original class and the remaining methods. **Partial** means other methods in that class already have confirmed translations. **Reconcile** means related Go checks exist but complete original-method equivalence has not been established. **Missing** means no complete translation was identified. The latter two statuses both remain pending; they do not assert the absence of all related implementation or testing.

### CLI, REPL, messages, and trace-spec output

Command-line options, runtime flags, REPL input, MP console rendering, warning policy, and trace-spec text.

- [x] [tlc2/REPLTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/REPLTest.java) — **Port complete**: `testProcessInput`.
  Go translation: [repl_java_test.go](../repl_java_test.go). All thirteen original expressions and exact results run sequentially through one REPL instance. Production preserves generated file order, default Randomization import, ToolIO reset, narrow exception catches, diagnostic parameters and buffered writer/finally behavior.
- [x] [tlc2/TLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/TLCTest.java) — **Port complete**: `testHandleParametersAbsoluteInvalid`, `testHandleParametersAbsoluteValid`, `testHandleParametersFractionInvalid`, `testHandleParametersAllocateLowerBound`, `testHandleParametersAllocateUpperBound`, `testHandleParametersAllocateHalf`, `testHandleParametersAllocate90`, `testHandleParametersMaxSetSize`, `testHandleParametersSimulateFileNum`, `testRuntimeConversion`.
  Go translation: [tlc/cli_java_test.go](cli_java_test.go). All ten original methods, exact memory ratios and signed bounds, heap-allocation assumptions, global set-size checks, simulation arguments and six runtime-format assertions retained. Production fixes Java floating-to-long saturation and uses a stable native heap budget rather than reserved memory.
- [x] [tlc2/output/MPTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/output/MPTest.java) — **Port complete**: `testPrintErrorInt`, `testPrintErrorIntString`, `testPrintErrorIntStringArray`, `testPrintProgressStats`.
  Go translation: [tlc/output_mp_java_test.go](output_mp_java_test.go). Full per-method ToolIO setup, actual production println capture, all exact counts/overload substitutions, six formatted progress parameters and both original locale alternatives retained. Production console/buffering and Java integral locale formatting are implemented. `MP.getMessage0` now constructs all 245 source templates before sequential substitution, including exact argument-count, tool/debug, null and empty-PID branches; its 225 literal cases are generated from the pinned Java source. The four original methods remain unchanged. Debug diagnostics and getTLCBug now follow source ordering, stream routing and recorder behavior. Subsequent production audit restores raw recorder ordering, runtime-exception object events, suppression and throwable stack policies. Manual observations verify these boundaries without increasing original-method counts.
- [x] [tlc2/output/SpecTraceExpressionWriterTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/output/SpecTraceExpressionWriterTest.java) — **Port complete**: `testInitNextWithNoError`, `testInitNextWithError`, `testInitNextWithErrorAndTraceExpression`, `testMultilineTraceExpression`.
  Go translation: [spec_trace_writer_java_test.go](../spec_trace_writer_java_test.go). All four original methods, independent temp files, exact preamble/error states/expressions, two-buffer ordering and named-Formula assertion retained. Generated files run through the legacy SANY frontend; its source return-code semantics are preserved, including ordinary semantic diagnostics.
- [x] [tlc2/output/WarningControlTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/output/WarningControlTest.java) — **Port complete**: `testSuppressMessagesSanyCode`, `testSuppressMessagesSanyErrorCodeFails`, `testSuppressMessagesTlcCode`, `testSuppressMessagesMultipleCodes`, `testSuppressMessagesUnknownCodeFails`, `testSuppressMessagesMissingArgFails`, `testMessagesAsErrorsSanyWarningCode`, `testMessagesAsErrorsTlcCode`, `testMessagesAsErrorsUnknownCodeFails`, `testMessagesAsErrorsMissingArgFails`, `testNowarningConflictWithSuppressMessages`, `testNowarningConflictWithmessagesAsErrors`, `testSameTlcCode`, `testSameSanyCode`, `testRuntimeSuppressedWarningProducesNoOutput`, `testRuntimeUnsuppressedWarningProducesOutput`, `testRuntimeWarningAsErrorThrowsTLCRuntimeException`, `testRuntimeWarningWithoutElevationDoesNotThrow`.
  Go translation: [tlc/output_warning_control_java_test.go](output_warning_control_java_test.go). All eighteen original direct CLI/runtime methods, source teardown, full set-membership/empty checks, real native stream captures and strict TLCRuntimeException expectation retained. Production separates SANY/TLC registrations and native streams now override TOOL buffering.

### Presentation models

Assignment, Formula, TypedSet, MCError, and MCState helpers.

- [x] [tlc2/model/FormulaTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/model/FormulaTest.java) — **Port complete**: `testUnnamed`, `testNamed`.
  Go translation: [tlc/model_test.go](model_test.go).
- [x] [tlc2/model/MCErrorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/model/MCErrorTest.java) — **Port complete**: `testGetErrorMessage`, `testUpdateStatesForTraceExpressions`.
  Go translation: [tlc/model_error_state_java_test.go](model_error_state_java_test.go). Shared original Utils fixtures, all six round trips and ordered state/variable assertions, record token loop, five trace states and nested variable display-name assertions are retained.
- [x] [tlc2/model/MCStateTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/model/MCStateTest.java) — **Port complete**: `testParseRoundTrips`, `testSimpleRecordPrinter`.
  Go translation: [tlc/model_error_state_java_test.go](model_error_state_java_test.go). Shared original Utils fixtures, all six round trips and ordered state/variable assertions, record token loop, five trace states and nested variable display-name assertions are retained.
- [x] [tlc2/model/TypedSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/model/TypedSetTest.java) — **Port complete**: `testParseSet1`, `testParseSet2`, `testParseSet3`, `testParseSet4`, `testParseSet5`, `testParseSet6`.
  Go translation: [tlc/model_test.go](model_test.go); original null input uses the nullable production boundary. Java trim semantics retained.

### Standard modules, constants, native overrides, and random values

Original module models and direct module-method tests, constant evaluation, module loading/overrides, and random value generation.

The nine newly translated standard-module models retain all original fixture bytes, including the complete empty-set and k-subset assumption lists. Empty-set state models retain deadlock checks and zero-uncovered assertions; constant-rank models retain no-generate-spec/no-JSON-trace overrides. Production fixes tuple conversion, ordered product emptiness, predicate-set emptiness and exact ToString override metadata before crediting the formerly failing assumptions.

- [x] [tlc2/module/ConstantContextTLCCacheTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/ConstantContextTLCCacheTest.java) — **Port complete**: `test`.
  Go translation: [tlc_module_models_java_test.go](../tlc_module_models_java_test.go), `TestJavaConstantContextTLCCache`. Complete original model/configuration and recorder assertions, inherited successful exit and class-specific setup retained.
- [x] [tlc2/module/RandomizationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/RandomizationTest.java) — **Port complete**: `testRandomSubsetNonFinite`, `testV1Valid`, `testV2Larger1`, `testSetNonFinite`, `testV1Negative`, `testV1NoIntValue`, `testV1Zero`, `testV2Zero`, `testV2Negative`, `testV3Empty`, `testV3AstronomicallyLarge`, `testV3isInfinite`, `testRSSV2Zero`, `testRSSV2Negative`, `testRSSV2Cardinality`, `testRSSV2TwiceCardinality`.
  Go translation: [tlc/modules_randomization_java_test.go](modules_randomization_java_test.go). All sixteen original methods, duplicate cases, exact cardinalities and EvalException message checks; source class seed/FP64 setup and shared generator in JUnit order retained. Production probability parsing now follows Java Double.valueOf.
- [x] [tlc2/module/SequencesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/SequencesTest.java) — **Port complete**: `testTailString`, `testHeadString`, `testHeadStringEmpty`, `testAppendString`, `testAppendString2`, `testAppendStringNonString`, `testConcatStringToSeq`, `testConcatSeqToString`, `testConcatStringToString`, `testConcatIntToSeq`, `testConcatSeqToInt`, `testConcatIntToInt`, `testSubseq`.
  Go translation: [tlc/modules_sequences_tlc_java_test.go](modules_sequences_tlc_java_test.go). Every original method, fixture, specific EvalException code/UniqueString equality, normalization order, MaxInt32 interval side, full permutation loop and Value.hashCode/equals HashSet assertion are retained.
- [x] [tlc2/module/TLCExtTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/TLCExtTest.java) — **Port complete**: `test`.
  Go translation: [tlc_module_models_java_test.go](../tlc_module_models_java_test.go), `TestJavaTLCExtModel`. Complete original model/configuration and recorder assertions, inherited successful exit and class-specific setup retained.
- [x] [tlc2/module/TLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/TLCTest.java) — **Port complete**: `testA`, `testB`, `testCombineMaxIntIntervalOnLeft`, `testCombineMaxIntIntervalOnRight`, `testPermutations`.
  Go translation: [tlc/modules_sequences_tlc_java_test.go](modules_sequences_tlc_java_test.go). Every original method, fixture, specific EvalException code/UniqueString equality, normalization order, MaxInt32 interval side, full permutation loop and Value.hashCode/equals HashSet assertion are retained.
- [x] [tlc2/tool/BagsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/BagsTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_module_models_java_test.go](../tlc_module_models_java_test.go), `TestJavaBagsModel`. Complete original model/configuration and recorder assertions, inherited successful exit and class-specific setup retained.
- [x] [tlc2/tool/ConstantRank1TLCEvalTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ConstantRank1TLCEvalTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_module_models_java_test.go](../tlc_module_models_java_test.go), `TestJavaConstantRank1TLCEval`. Complete original model/configuration and recorder assertions, inherited successful exit and class-specific setup retained.
- [x] [tlc2/tool/ConstantRank2AssertErrorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ConstantRank2AssertErrorTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_module_models_java_test.go](../tlc_module_models_java_test.go), `TestJavaConstantRank2AssertError`. Complete original model/configuration and recorder assertions, inherited successful exit and class-specific setup retained.
- [x] [tlc2/tool/EmptySetEqAssumeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EmptySetEqAssumeTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_module_models_java_test.go](../tlc_module_models_java_test.go), `TestJavaEmptySetEqAssume`. Complete original model/configuration and recorder assertions, inherited successful exit and class-specific setup retained.
- [x] [tlc2/tool/EmptySetEqStatesRcdTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EmptySetEqStatesRcdTest.java) — **Port complete**: `testRcdSpec`.
  Go translation: [tlc_module_models_java_test.go](../tlc_module_models_java_test.go), `TestJavaEmptySetEqStatesRcd`. Complete original model/configuration and recorder assertions, inherited successful exit and class-specific setup retained.
- [x] [tlc2/tool/EmptySetEqStatesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EmptySetEqStatesTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_module_models_java_test.go](../tlc_module_models_java_test.go), `TestJavaEmptySetEqStates`. Complete original model/configuration and recorder assertions, inherited successful exit and class-specific setup retained.
- [x] [tlc2/tool/KSubsetAssumeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/KSubsetAssumeTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_module_models_java_test.go](../tlc_module_models_java_test.go), `TestJavaKSubsetAssume`. Complete original model/configuration and recorder assertions, inherited successful exit and class-specific setup retained.
- [x] [tlc2/tool/RandomElementSimulationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementSimulationTest.java) — **Port complete**: `test` → [tlc_random_element_models_java_test.go](../tlc_random_element_models_java_test.go); original constructor, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/RandomElementT4Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementT4Test.java) — **Port complete**: `test` → [tlc_random_element_models_java_test.go](../tlc_random_element_models_java_test.go); original constructor, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/RandomElementTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementTest.java) — **Port complete**: `test` → [tlc_random_element_models_java_test.go](../tlc_random_element_models_java_test.go); original constructor, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/RandomElementXandYTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementXandYTest.java) — **Port complete**: `test` → [tlc_random_element_models_java_test.go](../tlc_random_element_models_java_test.go); original constructor, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/RandomSubsetATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetATest.java) — **Port complete**: `testSpec` (from `RandomSubset`).
  Go translation: [tlc_random_subset_models_java_test.go](../tlc_random_subset_models_java_test.go), `TestJavaRandomSubsetA`. Full original fixture, seed/setup, recorder and trace/value assertions retained.
- [x] [tlc2/tool/RandomSubsetBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetBTest.java) — **Port complete**: `testSpec` (from `RandomSubset`).
  Go translation: [tlc_random_subset_models_java_test.go](../tlc_random_subset_models_java_test.go), `TestJavaRandomSubsetB`. Full original fixture, seed/setup, recorder and trace/value assertions retained.
- [x] [tlc2/tool/RandomSubsetEmptyTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetEmptyTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_random_subset_models_java_test.go](../tlc_random_subset_models_java_test.go), `TestJavaRandomSubsetEmpty`. Full original fixture, seed/setup, recorder and trace/value assertions retained.
- [x] [tlc2/tool/RandomSubsetNextT4Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetNextT4Test.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_random_subset_models_java_test.go](../tlc_random_subset_models_java_test.go), `TestJavaRandomSubsetNextT4`. Full original fixture, seed/setup, recorder and trace/value assertions retained.
- [x] [tlc2/tool/RandomSubsetNextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetNextTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_random_subset_models_java_test.go](../tlc_random_subset_models_java_test.go), `TestJavaRandomSubsetNext`. Full original fixture, seed/setup, recorder and trace/value assertions retained.
- [x] [tlc2/tool/RandomSubsetNextTuplesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetNextTuplesTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_random_subset_models_java_test.go](../tlc_random_subset_models_java_test.go), `TestJavaRandomSubsetNextTuples`. Full original fixture, seed/setup, recorder and trace/value assertions retained.
- [x] [tlc2/tool/RandomSubsetSetOfFcnsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetSetOfFcnsTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_random_subset_models_java_test.go](../tlc_random_subset_models_java_test.go), `TestJavaRandomSubsetSetOfFcns`. Full original fixture, seed/setup, recorder and trace/value assertions retained.
- [x] [tlc2/tool/RandomSubsetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_random_subset_models_java_test.go](../tlc_random_subset_models_java_test.go), `TestJavaRandomSubsetModel`. Full original fixture, seed/setup, recorder and trace/value assertions retained.
- [x] [tlc2/tool/SetPredValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/SetPredValueTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_module_models_java_test.go](../tlc_module_models_java_test.go), `TestJavaSetPredValueModel`. Full original fixture, seed/setup, recorder and trace/value assertions retained.
- [x] [tlc2/tool/StandardModulesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/StandardModulesTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_module_models_java_test.go](../tlc_module_models_java_test.go), `TestJavaStandardModulesModel`. Full original fixture, seed/setup, recorder and trace/value assertions retained.
- [x] [tlc2/tool/SubseteqNextStateTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/SubseteqNextStateTest.java) — **Port complete**: `testSpec` → [tlc_subseteq_next_state_java_test.go](../tlc_subseteq_next_state_java_test.go); original constructor, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/UserModuleOverrideAnnotationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/UserModuleOverrideAnnotationTest.java) — **Port complete**: `testSpec` → [tlc_user_module_override_java_test.go](../tlc_user_module_override_java_test.go); original native fixture callbacks, constructor/index/classpath setup, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/UserModuleOverrideFromJarTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/UserModuleOverrideFromJarTest.java) — **Port complete**: `testSpec` → [tlc_user_module_override_java_test.go](../tlc_user_module_override_java_test.go); original native fixture callbacks, constructor/index/classpath setup, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/UserModuleOverrideTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/UserModuleOverrideTest.java) — **Port complete**: `testSpec` → [tlc_user_module_override_java_test.go](../tlc_user_module_override_java_test.go); original native fixture callbacks, constructor/index/classpath setup, inherited exit, and all active assertions retained.

The eight random-subset model translations preserve both fixed seeds and exact traces, every original state/queue count, four-worker setup, complete tuple/component checks and zero-uncovered assertions. The original `RandomSubsetTest` y-bound condition uses `firstX` for its upper bound; that literal condition is retained. Predicate-set and standard-module models run their complete original inputs through the production parser/checker.

- [x] [tlc2/tool/TLCGetLevelTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCGetLevelTest.java) — **Port complete**: `testSpec` in [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go). Restored inherited JSON trace-dump setup; all original assertions and model settings unchanged. Fully asserted helper also generates the original first-phase artifact for the TTrace recheck.

### Evaluation, initial states, next states, and action composition

Bindings, assignment, quantified evaluation, LET, INSTANCE, actions, enabledness, initial-state enumeration, and evaluation order.

- [x] [tlc2/tool/ActionCompositionATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionCompositionATest.java) — **Port complete**: `testSpec` → [tlc_action_composition_java_test.go](../tlc_action_composition_java_test.go); original constructor/configuration, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/ActionCompositionBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionCompositionBTest.java) — **Port complete**: `testSpec` → [tlc_action_composition_java_test.go](../tlc_action_composition_java_test.go); original constructor/configuration, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/AssignmentInitExpensiveTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssignmentInitExpensiveTest.java) — **Port complete**: `test` → [tlc_assignment_models_java_test.go](../tlc_assignment_models_java_test.go); original constructor/configuration, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/AssignmentInitNegTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssignmentInitNegTest.java) — **Port complete**: `test` → [tlc_assignment_models_java_test.go](../tlc_assignment_models_java_test.go); original constructor/configuration, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/AssignmentInitTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssignmentInitTest.java) — **Port complete**: `test` → [tlc_assignment_models_java_test.go](../tlc_assignment_models_java_test.go); original constructor/configuration, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/AssignmentNext2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssignmentNext2Test.java) — **Port complete**: `test` → [tlc_assignment_models_java_test.go](../tlc_assignment_models_java_test.go); original constructor/configuration, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/AssignmentNext3Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssignmentNext3Test.java) — **Port complete**: `test` → [tlc_assignment_models_java_test.go](../tlc_assignment_models_java_test.go); original constructor/configuration, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/AssignmentNextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssignmentNextTest.java) — **Port complete**: `test` → [tlc_assignment_models_java_test.go](../tlc_assignment_models_java_test.go); original constructor/configuration, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/CdotWithContextATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CdotWithContextATest.java) — **Port complete**: `testSpec` → [tlc_cdot_context_java_test.go](../tlc_cdot_context_java_test.go); source constructor/inherited settings and all assertions retained.
- [x] [tlc2/tool/CdotWithContextBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CdotWithContextBTest.java) — **Port complete**: `testSpec` → [tlc_cdot_context_java_test.go](../tlc_cdot_context_java_test.go); source constructor/inherited settings and all assertions retained.
- [x] [tlc2/tool/CdotWithContextCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CdotWithContextCTest.java) — **Port complete**: `testSpec` → [tlc_cdot_context_java_test.go](../tlc_cdot_context_java_test.go); source constructor/inherited settings and all assertions retained.
- [x] [tlc2/tool/CdotWithContextDTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CdotWithContextDTest.java) — **Port complete**: `testSpec` → [tlc_cdot_context_java_test.go](../tlc_cdot_context_java_test.go); source constructor/inherited settings and all assertions retained.
- [x] [tlc2/tool/ChainedCdotsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ChainedCdotsTest.java) — **Port complete**: `testSpec` → [tlc_cdot_context_java_test.go](../tlc_cdot_context_java_test.go); source constructor/inherited settings and all assertions retained.
- [x] [tlc2/tool/EmptyExistentialQuantifierTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EmptyExistentialQuantifierTest.java) — **Port complete**: `testSpec` → [tlc_evaluation_remaining_java_test.go](../tlc_evaluation_remaining_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/EvalControlTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EvalControlTest.java) — **Port complete**: `test`, `testIfEnabled` → [tlc/eval_control_java_test.go](eval_control_java_test.go); all original flag transitions and fifteen assertions retained.
- [x] [tlc2/tool/EvaluatingValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EvaluatingValueTest.java) — **Port complete**: `testSpec` → [tlc_evaluating_value_java_test.go](../tlc_evaluating_value_java_test.go); original native fixture callbacks, constructor/index/classpath setup, inherited exit, and all active assertions retained.
- [x] [tlc2/tool/LetDef1BoxedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef1BoxedTest.java) — **Port complete**: `testSpec` → [tlc_let_boxed_java_test.go](../tlc_let_boxed_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/LetDef1BoxedbTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef1BoxedbTest.java) — **Port complete**: `testSpec` → [tlc_let_boxed_java_test.go](../tlc_let_boxed_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/LetDef1BoxedcTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef1BoxedcTest.java) — **Port complete**: `testSpec` → [tlc_let_boxed_java_test.go](../tlc_let_boxed_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/LetDef1BoxeddTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef1BoxeddTest.java) — **Port complete**: `testSpec` → [tlc_let_boxed_java_test.go](../tlc_let_boxed_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/LetDef2BoxedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef2BoxedTest.java) — **Port complete**: `testSpec` → [tlc_let_boxed_java_test.go](../tlc_let_boxed_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/LetDef2BoxedbTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef2BoxedbTest.java) — **Port complete**: `testSpec` → [tlc_let_boxed_java_test.go](../tlc_let_boxed_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/LetDef2BoxedcTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef2BoxedcTest.java) — **Port complete**: `testSpec` → [tlc_let_boxed_java_test.go](../tlc_let_boxed_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/LetDef2BoxeddTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef2BoxeddTest.java) — **Port complete**: `testSpec` → [tlc_let_boxed_java_test.go](../tlc_let_boxed_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/MinimalSetOfInitStatesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/MinimalSetOfInitStatesTest.java) — **Port complete**: `testSpec` → [tlc_evaluation_remaining_java_test.go](../tlc_evaluation_remaining_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/MinimalSetOfNextStatesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/MinimalSetOfNextStatesTest.java) — **Port complete**: `testSpec` → [tlc_evaluation_remaining_java_test.go](../tlc_evaluation_remaining_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/SetOfStatesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/SetOfStatesTest.java) — **Port complete**: all six original methods → [tlc/set_of_states_java_test.go](set_of_states_java_test.go); original dummy-state identity/virtual equality, full32/64 collision matrices, iterator reset/fingerprint sum and HashSet assertions retained.
- [x] [tlc2/tool/UndeclaredRecursionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/UndeclaredRecursionTest.java) — **Port complete**: `testSpec` → [tlc_evaluation_remaining_java_test.go](../tlc_evaluation_remaining_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/evalorder/InitEvalOrder1Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/evalorder/InitEvalOrder1Test.java) — **Port complete**: `test` → [tlc_init_eval_order_java_test.go](../tlc_init_eval_order_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/evalorder/InitEvalOrder2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/evalorder/InitEvalOrder2Test.java) — **Port complete**: `test` → [tlc_init_eval_order_java_test.go](../tlc_init_eval_order_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/evalorder/InitEvalOrder3Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/evalorder/InitEvalOrder3Test.java) — **Port complete**: `test` → [tlc_init_eval_order_java_test.go](../tlc_init_eval_order_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/evalorder/InitEvalOrder4Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/evalorder/InitEvalOrder4Test.java) — **Port complete**: `test` → [tlc_init_eval_order_java_test.go](../tlc_init_eval_order_java_test.go); full original settings, inherited exit, and assertions retained.
- [x] [tlc2/tool/evalorder/InitEvalOrderBasicTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/evalorder/InitEvalOrderBasicTest.java) — **Port complete**: `test` → [tlc_init_eval_order_java_test.go](../tlc_init_eval_order_java_test.go); full original settings, inherited exit, and assertions retained.

### Safety checking, diagnostics, and checker lifecycle

Invariants, state/action properties, assumptions/postconditions, deadlocks, DFID, views, model errors, and completion/cleanup.

- [x] [tlc2/tool/ASTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ASTest.java) — **Port complete**: `testSpec` → [tlc_as_java_test.go](../tlc_as_java_test.go); all original assertions and source settings retained.
- [x] [tlc2/tool/AbsoluteSpecPathTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AbsoluteSpecPathTest.java) — **Port complete**: `test` → [tlc_absolute_spec_path_java_test.go](../tlc_absolute_spec_path_java_test.go); all original assertions and source settings retained.
- [x] [tlc2/tool/ActionLevelPropATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionLevelPropATest.java) — **Port complete**: `testSpec` → [tlc_action_level_prop_java_test.go](../tlc_action_level_prop_java_test.go); all original assertions and source settings retained.
- [x] [tlc2/tool/ActionLevelPropBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionLevelPropBTest.java) — **Port complete**: `testSpec` → [tlc_action_level_prop_java_test.go](../tlc_action_level_prop_java_test.go); all original assertions and source settings retained.
- [x] [tlc2/tool/ActionLevelPropCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionLevelPropCTest.java) — **Port complete**: `testSpec` → [tlc_action_level_prop_java_test.go](../tlc_action_level_prop_java_test.go); all original assertions and source settings retained.
- [x] [tlc2/tool/ActionLevelPropDTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionLevelPropDTest.java) — **Port complete**: `testSpec` → [tlc_action_level_prop_java_test.go](../tlc_action_level_prop_java_test.go); all original assertions and source settings retained.
- [x] [tlc2/tool/ActionLevelPropETest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionLevelPropETest.java) — **Port complete**: `testSpec` → [tlc_action_level_prop_java_test.go](../tlc_action_level_prop_java_test.go); all original assertions and source settings retained.
- [ ] [tlc2/tool/AssertExpressionStack.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssertExpressionStack.java) — **Reconcile**: `testSpec`. Unchanged Java at the inventory revision fails `assertNoTESpec` and inherited SUCCESS exit: the source base enables generation and EC maps the assertion failure to exit14. Go produces the same trace, generation and exit. Keep pending; no weakened test, new skip, or incompatible production change. Exact translation remains an ignored scratch draft until upstream expectations are reconciled.
- [x] [tlc2/tool/ContinueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ContinueTest.java) — **Port complete**: `testSpec` → [tlc_continue_empty_java_test.go](../tlc_continue_empty_java_test.go); full source worker counts, options, coverage, exit and recorder assertions retained.
- [ ] [tlc2/tool/DepthFirstTerminate.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DepthFirstTerminate.java) — **Reconcile**: `testSpec`. Unchanged Java requests availableProcessors workers but current TLC rejects DFID with multiple workers (issue548); on this48-core host it records GENERAL and exits255, violating the original assertions. Preserve the worker setting and keep pending; no new skip or single-worker substitute.
- [x] [tlc2/tool/DiameterTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DiameterTest.java) — **Port complete**: `testSpec` → [tlc_diameter_java_test.go](../tlc_diameter_java_test.go); full source worker counts, options, coverage, exit and recorder assertions retained.
- [x] [tlc2/tool/DotConstrainedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DotConstrainedTest.java) — **Port complete**: `testSpec` in [tlc_dot_constrained_java_test.go](../tlc_dot_constrained_java_test.go). Original writer override delegates before atomically observing constrained flags; exact trace, exit, statistics, register, postcondition and coverage assertions preserved.
- [x] [tlc2/tool/ElevatedSanyWarning.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ElevatedSanyWarning.java) — **Port complete**: `testSpec` → [tlc_elevated_sany_warning_java_test.go](../tlc_elevated_sany_warning_java_test.go); original absolute corpus input, message elevation, actual ToolIO stdout, parsing-failure recorder assertion and inherited ERROR_SPEC_PARSE exit/settings preserved.
- [x] [tlc2/tool/EmptyTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EmptyTest.java) — **Port complete**: `testSpec` → [tlc_continue_empty_java_test.go](../tlc_continue_empty_java_test.go); full source worker counts, options, coverage, exit and recorder assertions retained.
- [x] [tlc2/tool/EvalExceptionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EvalExceptionTest.java) — **Port complete**: `testSpec` in [tlc_eval_exception_java_test.go](../tlc_eval_exception_java_test.go). Restored inherited JSON dump setting; preserves the original coverage override, complete trace/actions/ordinals, exact evaluation diagnostic and expression stack, stats and ERROR exit. The fully asserted helper also provides the original first phase for the TTrace test.
- [x] [tlc2/tool/FingerprintExceptionHangTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/FingerprintExceptionHangTest.java) — **Port complete**: `testSpec` → [tlc_fingerprint_hang_java_test.go](../tlc_fingerprint_hang_java_test.go); exact source diagnostic stacks/messages, recorder checks, coverage assertions where present, inherited options and FAILURE_SPEC_EVAL exit retained.
- [x] [tlc2/tool/FingerprintExceptionInitTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/FingerprintExceptionInitTest.java) — **Port complete**: `testSpec` → [tlc_fingerprint_exceptions_java_test.go](../tlc_fingerprint_exceptions_java_test.go); exact source diagnostic stacks/messages, recorder checks, coverage assertions where present, inherited options and FAILURE_SPEC_EVAL exit retained.
- [x] [tlc2/tool/FingerprintExceptionNextCallstackTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/FingerprintExceptionNextCallstackTest.java) — **Port complete**: `testSpec` → [tlc_fingerprint_exceptions_java_test.go](../tlc_fingerprint_exceptions_java_test.go); exact source diagnostic stacks/messages, recorder checks, coverage assertions where present, inherited options and FAILURE_SPEC_EVAL exit retained.
- [x] [tlc2/tool/FingerprintExceptionNextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/FingerprintExceptionNextTest.java) — **Port complete**: `testSpec` → [tlc_fingerprint_exceptions_java_test.go](../tlc_fingerprint_exceptions_java_test.go); exact source diagnostic stacks/messages, recorder checks, coverage assertions where present, inherited options and FAILURE_SPEC_EVAL exit retained.
- [ ] [tlc2/tool/InliningTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/InliningTest.java) — **Reconcile**: `testSpec` combines checker assertions with HotSpot CompilerInlining/JFR recording, reflective `ExpectInlined` annotations and exact callee descriptor checks. Go compile-time inlining does not provide those JVM records; keep the entire method visible and uncredited rather than porting only its checker assertions.
  JVM compilation/inlining assertion needs a documented Go-specific disposition. `test-dist` explicitly selects this class in its slow batch; only the generic batch excludes it.
- [x] [tlc2/tool/InvParameterizedATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/InvParameterizedATest.java) — **Port complete**: `testSpec` → [tlc_inv_parameterized_java_test.go](../tlc_inv_parameterized_java_test.go); source invariant text/level, embedded config, options, inherited exit and recorder assertions retained.
- [x] [tlc2/tool/InvParameterizedBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/InvParameterizedBTest.java) — **Port complete**: `testSpec` → [tlc_inv_parameterized_java_test.go](../tlc_inv_parameterized_java_test.go); source invariant text/level, embedded config, options, inherited exit and recorder assertions retained.
- [x] [tlc2/tool/InvParameterizedCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/InvParameterizedCTest.java) — **Port complete**: `testSpec` → [tlc_inv_parameterized_java_test.go](../tlc_inv_parameterized_java_test.go); source invariant text/level, embedded config, options, inherited exit and recorder assertions retained.
- [x] [tlc2/tool/MinimumDiameterTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/MinimumDiameterTest.java) — **Port complete**: `testSpec` → [tlc_post_assumption_minimum_java_test.go](../tlc_post_assumption_minimum_java_test.go); full source worker counts, options, coverage, exit and recorder assertions retained.
- [x] [tlc2/tool/MonolithSpecTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/MonolithSpecTest.java) — **Port complete**: `testSpec` → [tlc_monolith_spec_java_test.go](../tlc_monolith_spec_java_test.go); all original assertions, source options, worker counts and inherited exit preserved.
- [x] [tlc2/tool/PostAssumptionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PostAssumptionTest.java) — **Port complete**: `testSpec` → [tlc_post_assumption_minimum_java_test.go](../tlc_post_assumption_minimum_java_test.go); full source worker counts, options, coverage, exit and recorder assertions retained.
- [x] [tlc2/tool/TSnapShotTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TSnapShotTest.java) — **Port complete**: `testSpec` → [tlc_tsnapshot_java_test.go](../tlc_tsnapshot_java_test.go); all original assertions, source options, worker counts and inherited exit preserved.
- [x] [tlc2/tool/ViewMapTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ViewMapTest.java) — **Port complete**: `testSpec` → [tlc_view_map_java_test.go](../tlc_view_map_java_test.go); all original assertions, source options, worker counts and inherited exit preserved.

### Issue regressions in the evaluator and checker

The original Github, Bugzilla, and CodePlex cases under tool/. Read each original model/config and assertions before selecting its implementation slice.

- [x] [tlc2/tool/BugzillaBug279Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/BugzillaBug279Test.java) — **Port complete**: `testSpec` → [tlc_bugzilla_279_java_test.go](../tlc_bugzilla_279_java_test.go); full original model/config, constructor and overrides, trace/register/coverage assertions where present, source defaults and inherited exit retained.
- [x] [tlc2/tool/CodePlexBug21Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CodePlexBug21Test.java) — **Port complete**: `testSpec` → [tlc_codeplex_21_java_test.go](../tlc_codeplex_21_java_test.go); full original model/config, constructor and overrides, trace/register/coverage assertions where present, source defaults and inherited exit retained.
- [x] [tlc2/tool/Github1087Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1087Test.java) — **Port complete**: `testSpec` → [tlc_github_1087_java_test.go](../tlc_github_1087_java_test.go); full original model/config, constructor and overrides, trace/register/coverage assertions where present, source defaults and inherited exit retained.
- [x] [tlc2/tool/Github1134aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1134aTest.java) — **Port complete**: `testSpec` → [tlc_github_1134_java_test.go](../tlc_github_1134_java_test.go); exact original diagnostics/counts/coverage and inherited exit, MC.tla input, config and source noGenerateSpec/doDumpTrace overrides preserved.
- [x] [tlc2/tool/Github1134bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1134bTest.java) — **Port complete**: `testSpec` → [tlc_github_1134_java_test.go](../tlc_github_1134_java_test.go); exact original diagnostics/counts/coverage and inherited exit, MC.tla input, config and source noGenerateSpec/doDumpTrace overrides preserved.
- [x] [tlc2/tool/Github1134cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1134cTest.java) — **Port complete**: `testSpec` → [tlc_github_1134_java_test.go](../tlc_github_1134_java_test.go); exact original diagnostics/counts/coverage and inherited exit, MC.tla input, config and source noGenerateSpec/doDumpTrace overrides preserved.
- [x] [tlc2/tool/Github1134dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1134dTest.java) — **Port complete**: `testSpec` → [tlc_github_1134_java_test.go](../tlc_github_1134_java_test.go); exact original diagnostics/counts/coverage and inherited exit, MC.tla input, config and source noGenerateSpec/doDumpTrace overrides preserved.
- [x] [tlc2/tool/Github1134eTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1134eTest.java) — **Port complete**: `testSpec` → [tlc_github_1134_java_test.go](../tlc_github_1134_java_test.go); exact original diagnostics/counts/coverage and inherited exit, MC.tla input, config and source noGenerateSpec/doDumpTrace overrides preserved.
- [x] [tlc2/tool/Github1134fTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1134fTest.java) — **Port complete**: `testSpec` → [tlc_github_1134_java_test.go](../tlc_github_1134_java_test.go); exact original diagnostics/counts/coverage and inherited exit, MC.tla input, config and source noGenerateSpec/doDumpTrace overrides preserved.
- [x] [tlc2/tool/Github1145Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1145Test.java) — **Port complete**: `testSpec` → [tlc_github_1145_java_test.go](../tlc_github_1145_java_test.go); original FINISHED assertion and inherited success exit, default generation/coverage/DOT/debugger/JSON flags retained.
- [x] [tlc2/tool/Github1145bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1145bTest.java) — **Port complete**: `testSpec` → [tlc_github_1145_java_test.go](../tlc_github_1145_java_test.go); exact initial Inv violation with x = 1 and inherited safety exit; all five source instrumentation overrides retained.
- [x] [tlc2/tool/Github1147Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1147Test.java) — **Port complete**: `testSpec` → [tlc_github_1147_java_test.go](../tlc_github_1147_java_test.go); full original golden DOT line/EOF comparison, 51/50/47 counts, depth3, postcondition diagnostics, zeroUncovered and safety exit retained; signed fingerprint rendering fixed in production.
- [x] [tlc2/tool/Github1161Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1161Test.java) — **Port complete**: `testSpec` → [tlc_github_1161_java_test.go](../tlc_github_1161_java_test.go); all seven configured properties and their original wrapper variants, FINISHED/SUCCESS/formula diagnostics and inherited success exit; original monolith and instrumentation overrides retained.
- [x] [tlc2/tool/Github1161ViolatedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1161ViolatedTest.java) — **Port complete**: `testSpec` → [tlc_github_1161_java_test.go](../tlc_github_1161_java_test.go); original PropertyViolated diagnostic, counterexample and complete s=0..5 trace/ordinals/action metadata; inherited safety exit and source instrumentation overrides retained.
- [x] [tlc2/tool/Github1198aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198aTest.java) — **Port complete**: `testSpec` → [tlc_github_1198_java_test.go](../tlc_github_1198_java_test.go); exact original tautology warning presence, 3/2/0 counts, depth2, zeroUncovered and success exit; original coverage override and remaining default runner settings retained.
- [x] [tlc2/tool/Github1198bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198bTest.java) — **Port complete**: `testSpec` → [tlc_github_1198_java_test.go](../tlc_github_1198_java_test.go); exact original tautology warning presence, 3/2/0 counts, depth2, zeroUncovered and success exit; original coverage override and remaining default runner settings retained.
- [x] [tlc2/tool/Github1198cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198cTest.java) — **Port complete**: `testSpec` → [tlc_github_1198_java_test.go](../tlc_github_1198_java_test.go); exact original tautology warning presence, 3/2/0 counts, depth2, zeroUncovered and success exit; original coverage override and remaining default runner settings retained.
- [x] [tlc2/tool/Github1198dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198dTest.java) — **Port complete**: `testSpec` → [tlc_github_1198_java_test.go](../tlc_github_1198_java_test.go); exact original tautology warning absence, 3/2/0 counts, depth2, zeroUncovered and success exit; original coverage override and remaining default runner settings retained.
- [x] [tlc2/tool/Github1198fTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198fTest.java) — **Port complete**: `testSpec` → [tlc_github_1198_java_test.go](../tlc_github_1198_java_test.go); exact original tautology warning absence, 3/2/0 counts, depth2, zeroUncovered and success exit; original coverage override and remaining default runner settings retained.
- [x] [tlc2/tool/Github1198hTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198hTest.java) — **Port complete**: `testSpec` → [tlc_github_1198_java_test.go](../tlc_github_1198_java_test.go); exact original tautology warning absence, 3/2/0 counts, depth2, zeroUncovered and success exit; original coverage override and remaining default runner settings retained.
- [x] [tlc2/tool/Github1244Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1244Test.java) — **Port complete**: `testSpec` → [tlc_github_1244_java_test.go](../tlc_github_1244_java_test.go); original FINISHED/SUCCESS, 3/2/0 counts and inherited success exit; exact embedded TLA configuration and noGenerateSpec/doDumpTrace overrides retained.
- [x] [tlc2/tool/Github1244bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1244bTest.java) — **Port complete**: `testSpec` → [tlc_github_1244_java_test.go](../tlc_github_1244_java_test.go); original FINISHED/SUCCESS, 3/2/0 counts and inherited success exit; exact configuration and noGenerateSpec/doDumpTrace overrides retained.
- [x] [tlc2/tool/Github1244cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1244cTest.java) — **Port complete**: `testSpec` → [tlc_github_1244_java_test.go](../tlc_github_1244_java_test.go); original FINISHED/SUCCESS, 3/2/0 counts and inherited success exit; exact configuration and noGenerateSpec/doDumpTrace overrides retained.
- [x] [tlc2/tool/Github1302Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1302Test.java) — **Port complete**: `testSpec` → [tlc_github_1302_java_test.go](../tlc_github_1302_java_test.go); original simulation num=3/depth3, exact 10/1/0 stats and zeroUncovered, success exit and full original runner overrides; original embedded TLA configuration retained.
- [x] [tlc2/tool/Github1302bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1302bTest.java) — **Port complete**: `testSpec` → [tlc_github_1302_java_test.go](../tlc_github_1302_java_test.go); exact original initial invariant violation and both function states, absence of Stats/GENERAL, safety exit and forced trace generation with other instrumentation disabled; original embedded TLA configuration retained.
- [x] [tlc2/tool/Github1302cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1302cTest.java) — **Port complete**: `testSpec` → [tlc_github_1302_java_test.go](../tlc_github_1302_java_test.go); exact original 2/1/0 stats, absence of GENERAL, success exit and forced trace generation with other instrumentation disabled; original embedded TLA configuration retained.
- [x] [tlc2/tool/Github1389CountingTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389CountingTest.java) — **Port complete**: `testSpec` → [tlc_github_1389_java_test.go](../tlc_github_1389_java_test.go); original CountAtMostFour temporal violation, counterexample and both postcondition diagnostics, inherited safety exit; original ten-state witness postcondition retained.
- [x] [tlc2/tool/Github1389LoopsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389LoopsTest.java) — **Port complete**: `testSpec` → [tlc_github_1389_java_test.go](../tlc_github_1389_java_test.go); original SYSTEM_STACK_OVERFLOW and absence of GENERAL, inherited Error exit; recoverable native stack-byte exhaustion implemented without a semantic recursion cutoff.
- [x] [tlc2/tool/Github1389StateGuardTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389StateGuardTest.java) — **Port complete**: `testSpec` → [tlc_github_1389_java_test.go](../tlc_github_1389_java_test.go); original SYSTEM_STACK_OVERFLOW and absence of GENERAL, inherited Error exit; unresolvable state guard follows original recursive expansion to recoverable resource exhaustion.
- [x] [tlc2/tool/Github1389Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389Test.java) — **Port complete**: `testSpec` → [tlc_github_1389_java_test.go](../tlc_github_1389_java_test.go); original FINISHED/SUCCESS and absence of unsupported-formula diagnostic, inherited success exit; all source runner overrides retained.
- [x] [tlc2/tool/Github1389ViolatedBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389ViolatedBTest.java) — **Port complete**: `testSpec` → [tlc_github_1389_java_test.go](../tlc_github_1389_java_test.go); original PropViolated diagnostic, full two-state x=0,1 trace/ordinals/action metadata and back-to-state1, inherited liveness exit; actual expanded base-case level preserved.
- [x] [tlc2/tool/Github1389ViolatedCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389ViolatedCTest.java) — **Port complete**: `testSpec` → [tlc_github_1389_java_test.go](../tlc_github_1389_java_test.go); original PropViolated/counterexample and both postcondition diagnostics, inherited liveness exit; exact original postcondition retains lasso shape over x in1..98.
- [x] [tlc2/tool/Github1389ViolatedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389ViolatedTest.java) — **Port complete**: `testSpec` → [tlc_github_1389_java_test.go](../tlc_github_1389_java_test.go); original PropViolated diagnostic, full two-state x=0,1 trace/ordinals/action metadata and back-to-state1, inherited liveness exit.
- [x] [tlc2/tool/Github179aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github179aTest.java) — **Port complete**: `testSpec` → [tlc_github_179_java_test.go](../tlc_github_179_java_test.go); original assumption exit, FINISHED and exact native PrintT signature/failure message; all source runner defaults retained.
- [x] [tlc2/tool/Github179bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github179bTest.java) — **Port complete**: `testSpec` → [tlc_github_179_java_test.go](../tlc_github_179_java_test.go); original spec-evaluation exit, FINISHED, exact native signature/failure message and every nested-expression frame/range/newline; all source runner defaults retained.
- [x] [tlc2/tool/Github179cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github179cTest.java) — **Port complete**: `testSpec` → [tlc_github_179_java_test.go](../tlc_github_179_java_test.go); original spec-evaluation exit, FINISHED, exact native signature/failure message and every nested-expression frame/range/newline; all source runner defaults retained.
- [x] [tlc2/tool/Github362Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github362Test.java) — **Port complete**: `testSpec` → [tlc_github_362_391_407_java_test.go](../tlc_github_362_391_407_java_test.go); all six original output substrings for instance names/constants, FINISHED/SUCCESS and inherited success exit; embedded config and all runner defaults retained.
- [x] [tlc2/tool/Github391Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github391Test.java) — **Port complete**: `testSpec` → [tlc_github_362_391_407_java_test.go](../tlc_github_362_391_407_java_test.go); original zero-state stats/search depth, FINISHED and inherited success exit; all runner defaults retained.
- [x] [tlc2/tool/Github407Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github407Test.java) — **Port complete**: `testSpec` → [tlc_github_362_391_407_java_test.go](../tlc_github_362_391_407_java_test.go); original 9/4/0 stats, depth3, FINISHED, inherited success exit, complete golden dump/EOF and zero-uncovered assertions; original doDump override retained and plain state writer newline corrected.
- [x] [tlc2/tool/Github432Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github432Test.java) — **Port complete**: `testA`, `testB`, `testC`, `testD` → [tlc_github_432_java_test.go](../tlc_github_432_java_test.go); all four original config-substitution rows, exact warning cardinality/parameters or absence, original exit-assertion override and all runner defaults; warning names extracted from syntax images as Java does.
- [x] [tlc2/tool/Github461Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github461Test.java) — **Port complete**: `testSpec` → [tlc_github_461_525_597_java_test.go](../tlc_github_461_525_597_java_test.go); original assertion diagnostic, full x=0..4 trace/ordinals/extended-state metadata, exact nested stack/newline, zero uncovered and inherited assertion exit; all runner defaults.
- [x] [tlc2/tool/Github525Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github525Test.java) — **Port complete**: `testSpec` → [tlc_github_461_525_597_java_test.go](../tlc_github_461_525_597_java_test.go); original FINISHED/unsupported-formula diagnostic/no GENERAL and inherited Error exit; production exit-code mapping corrected to source.
- [x] [tlc2/tool/Github597Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github597Test.java) — **Port complete**: `testSpec` → [tlc_github_461_525_597_java_test.go](../tlc_github_461_525_597_java_test.go); original random fp/seed and embedded dekker config, coverage/DOT disabled, exact 4356/1500/0 stats, Termination violation/counterexample/state/back-to-state assertions and inherited liveness exit.
- [x] [tlc2/tool/Github648Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github648Test.java) — **Port complete**: `testSpec` → [tlc_github_648_652_java_test.go](../tlc_github_648_652_java_test.go); original one-worker settings, FINISHED, depth1, 1332/36/0 stats, all 35 exact coverage/count/cost lines and inherited success exit; nested TLCEval global reentrancy and Empty next-state argument implemented.
- [x] [tlc2/tool/Github648wNTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github648wNTest.java) — **Port complete**: `testSpec` → [tlc_github_648_652_java_test.go](../tlc_github_648_652_java_test.go); original ten-worker settings, FINISHED, depth1, 1332/36/0 stats, all 35 exact per-worker coverage/count/cost lines and inherited success exit; no worker-count/model-size reduction.
- [x] [tlc2/tool/Github652Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github652Test.java) — **Port complete**: `testSpec` → [tlc_github_648_652_java_test.go](../tlc_github_648_652_java_test.go); original embedded config, FINISHED, no GENERAL, zero-state stats/depth and inherited success exit; all runner defaults.
- [x] [tlc2/tool/Github680aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github680aTest.java) — **Port complete**: `testSpec` → [tlc_github_680_java_test.go](../tlc_github_680_java_test.go); original hidden UNCHANGED warning parameters, FINISHED/no GENERAL, depth1, 1/1/0 stats and inherited success exit; original warning suppression, embedded config and debugger override retained.
- [x] [tlc2/tool/Github680bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github680bTest.java) — **Port complete**: `testSpec` → [tlc_github_680_java_test.go](../tlc_github_680_java_test.go); original hidden UNCHANGED warning parameters, FINISHED/no GENERAL, depth1, 1/1/0 stats and inherited success exit; original warning suppression, embedded config and debugger override retained.
- [x] [tlc2/tool/Github680cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github680cTest.java) — **Port complete**: `testSpec` → [tlc_github_680_java_test.go](../tlc_github_680_java_test.go); original hidden UNCHANGED warning parameters, FINISHED/no GENERAL, depth1, 1/1/0 stats and inherited success exit; original warning suppression, embedded config and debugger override retained.
- [x] [tlc2/tool/Github687Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687Test.java) — **Port complete**: `testSpec` → [tlc_github_687_java_test.go](../tlc_github_687_java_test.go); all original recorded diagnostics, state-count and zero-uncovered assertions where present, and inherited exit; original embedded/config override, noGenerateSpec=true and doDumpTrace=false retained.
- [x] [tlc2/tool/Github687fifoTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687fifoTest.java) — **Port complete**: `testSpec` → [tlc_github_687_java_test.go](../tlc_github_687_java_test.go); all original recorded diagnostics, state-count and zero-uncovered assertions where present, and inherited exit; original embedded/config override, noGenerateSpec=true and doDumpTrace=false retained.
- [x] [tlc2/tool/Github687specATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687specATest.java) — **Port complete**: `testSpec` → [tlc_github_687_java_test.go](../tlc_github_687_java_test.go); all original recorded diagnostics, state-count and zero-uncovered assertions where present, and inherited exit; original embedded/config override, noGenerateSpec=true and doDumpTrace=false retained.
- [x] [tlc2/tool/Github687specBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687specBTest.java) — **Port complete**: `testSpec` → [tlc_github_687_java_test.go](../tlc_github_687_java_test.go); all original recorded diagnostics, state-count and zero-uncovered assertions where present, and inherited exit; original embedded/config override, noGenerateSpec=true and doDumpTrace=false retained.
- [x] [tlc2/tool/Github687specCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687specCTest.java) — **Port complete**: `testSpec` → [tlc_github_687_java_test.go](../tlc_github_687_java_test.go); all original recorded diagnostics, state-count and zero-uncovered assertions where present, and inherited exit; original embedded/config override, noGenerateSpec=true and doDumpTrace=false retained.
- [x] [tlc2/tool/Github687specDTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687specDTest.java) — **Port complete**: `testSpec` → [tlc_github_687_java_test.go](../tlc_github_687_java_test.go); all original recorded diagnostics, state-count and zero-uncovered assertions where present, and inherited exit; original embedded/config override, noGenerateSpec=true and doDumpTrace=false retained.
- [x] [tlc2/tool/Github687specInitNextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687specInitNextTest.java) — **Port complete**: `testSpec` → [tlc_github_687_java_test.go](../tlc_github_687_java_test.go); all original recorded diagnostics, state-count and zero-uncovered assertions where present, and inherited exit; original embedded/config override, noGenerateSpec=true and doDumpTrace=false retained.
- [x] [tlc2/tool/Github687specPrimeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687specPrimeTest.java) — **Port complete**: `testSpec` → [tlc_github_687_java_test.go](../tlc_github_687_java_test.go); all original recorded diagnostics, state-count and zero-uncovered assertions where present, and inherited exit; original embedded/config override, noGenerateSpec=true and doDumpTrace=false retained.
- [x] [tlc2/tool/Github696Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github696Test.java) — **Port complete**: `testSpec` → [tlc_github_696_java_test.go](../tlc_github_696_java_test.go); original typed-model-value failure message, FINISHED, inherited failure exit, and original depth/stats assertions where present; all runner defaults retained.
- [x] [tlc2/tool/Github696bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github696bTest.java) — **Port complete**: `testSpec` → [tlc_github_696_java_test.go](../tlc_github_696_java_test.go); original typed-model-value failure message, FINISHED, inherited failure exit, and original depth/stats assertions where present; all runner defaults retained.
- [x] [tlc2/tool/Github715Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github715Test.java) — **Port complete**: `testSpec` → [tlc_github_715_java_test.go](../tlc_github_715_java_test.go); original FINISHED/diagnostic presence and absence, exact invalid-spec variable/location parameters, depth2 and 3/2/0 stats where present, and inherited exit; all four original configurations and doCoverage=false retained.
- [x] [tlc2/tool/Github715bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github715bTest.java) — **Port complete**: `testSpec` → [tlc_github_715_java_test.go](../tlc_github_715_java_test.go); original FINISHED/diagnostic presence and absence, exact invalid-spec variable/location parameters, depth2 and 3/2/0 stats where present, and inherited exit; all four original configurations and doCoverage=false retained.
- [x] [tlc2/tool/Github715cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github715cTest.java) — **Port complete**: `testSpec` → [tlc_github_715_java_test.go](../tlc_github_715_java_test.go); original FINISHED/diagnostic presence and absence, exact invalid-spec variable/location parameters, depth2 and 3/2/0 stats where present, and inherited exit; all four original configurations and doCoverage=false retained.
- [x] [tlc2/tool/Github715dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github715dTest.java) — **Port complete**: `testSpec` → [tlc_github_715_java_test.go](../tlc_github_715_java_test.go); original FINISHED/diagnostic presence and absence, exact invalid-spec variable/location parameters, depth2 and 3/2/0 stats where present, and inherited exit; all four original configurations and doCoverage=false retained.
- [x] [tlc2/tool/Github725Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725Test.java) — **Port complete**: `testSpec` → [tlc_github_725_java_test.go](../tlc_github_725_java_test.go); original FINISHED/SUCCESS (or FINISHED/no GENERAL for725f) and inherited success exit. Original noGenerateSpec=true, coverage/debugger=false, DOT default and original JSON overrides retained; no production changes required.
- [x] [tlc2/tool/Github725bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725bTest.java) — **Port complete**: `testSpec` → [tlc_github_725_java_test.go](../tlc_github_725_java_test.go); original FINISHED/SUCCESS (or FINISHED/no GENERAL for725f) and inherited success exit. Original noGenerateSpec=true, coverage/debugger=false, DOT default and original JSON overrides retained; no production changes required.
- [x] [tlc2/tool/Github725cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725cTest.java) — **Port complete**: `testSpec` → [tlc_github_725_java_test.go](../tlc_github_725_java_test.go); original FINISHED/SUCCESS (or FINISHED/no GENERAL for725f) and inherited success exit. Original noGenerateSpec=true, coverage/debugger=false, DOT default and original JSON overrides retained; no production changes required.
- [x] [tlc2/tool/Github725dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725dTest.java) — **Port complete**: `testSpec` → [tlc_github_725_java_test.go](../tlc_github_725_java_test.go); original FINISHED/SUCCESS (or FINISHED/no GENERAL for725f) and inherited success exit. Original noGenerateSpec=true, coverage/debugger=false, DOT default and original JSON overrides retained; no production changes required.
- [x] [tlc2/tool/Github725eTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725eTest.java) — **Port complete**: `testSpec` → [tlc_github_725_java_test.go](../tlc_github_725_java_test.go); original FINISHED/SUCCESS (or FINISHED/no GENERAL for725f) and inherited success exit. Original noGenerateSpec=true, coverage/debugger=false, DOT default and original JSON overrides retained; no production changes required.
- [x] [tlc2/tool/Github725fTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725fTest.java) — **Port complete**: `testSpec` → [tlc_github_725_java_test.go](../tlc_github_725_java_test.go); original FINISHED/SUCCESS (or FINISHED/no GENERAL for725f) and inherited success exit. Original noGenerateSpec=true, coverage/debugger=false, DOT default and original JSON overrides retained; no production changes required.
- [x] [tlc2/tool/Github725gTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725gTest.java) — **Port complete**: `testSpec` → [tlc_github_725_java_test.go](../tlc_github_725_java_test.go); all original diagnostics, full single-state trace/actions/ordinal and stuttering2, exact Prop violation/depth1/2/1/0 stats, and inherited liveness exit. Original noGenerateSpec=true, coverage/debugger=false, DOT default and original JSON overrides retained; no production changes required.
- [x] [tlc2/tool/Github725hTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725hTest.java) — **Port complete**: `testSpec` → [tlc_github_725_java_test.go](../tlc_github_725_java_test.go); original FINISHED/no GENERAL, exact Inv initial-state diagnostic/state string/newline and inherited safety exit, with original Github725g model/config override. Original noGenerateSpec=true, coverage/debugger=false, DOT default and original JSON overrides retained; no production changes required.
- [x] [tlc2/tool/Github726Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github726Test.java) — **Port complete**: `testSpec` → [tlc_github_726_757_java_test.go](../tlc_github_726_757_java_test.go); original exact write(String) log membership/absence, FINISHED, depth0/zero stats and inherited success exit; constructor output capture, debugger=false and other defaults retained.
- [x] [tlc2/tool/Github742Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github742Test.java) — **Port complete**: `testSpec` → [tlc_github_726_757_java_test.go](../tlc_github_726_757_java_test.go); original FINISHED/unsupported-liveness diagnostic/no GENERAL and inherited Error exit; full6×6 fairness formula retained, signed32-bit DNF multiplication overflow handling ported from LNConj.
- [x] [tlc2/tool/Github743Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github743Test.java) — **Port complete**: `testSpec` → [tlc_github_726_757_java_test.go](../tlc_github_726_757_java_test.go); original FINISHED/no GENERAL and inherited safety exit; complete four-state trace, actions/ordinals and extended-state metadata; Tool.getState reconstruction skips incomparable candidates using the exact source exception catch.
- [x] [tlc2/tool/Github746Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github746Test.java) — **Port complete**: `testSpec` → [tlc_github_726_757_java_test.go](../tlc_github_726_757_java_test.go); original FINISHED/no GENERAL/COMPUTING_INIT and inherited safety exit; exact Inv initial-state string, field order and trailing newline; all runner defaults retained.
- [x] [tlc2/tool/Github757Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github757Test.java) — **Port complete**: `testSpec` → [tlc_github_726_757_java_test.go](../tlc_github_726_757_java_test.go); original FINISHED/no GENERAL/no postcondition false/evaluation-error and inherited safety exit; embedded config, noGenerateSpec=true, doDumpTrace=false and explicit JSON dump retained.
- [x] [tlc2/tool/Github766SimulateTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github766SimulateTest.java) — **Port complete**: `testSpec` → [tlc_github_766_java_test.go](../tlc_github_766_java_test.go); original simulation flag, explicit JSON dump, embedded config, noGenerateSpec=true, doDumpTrace=false, FINISHED/no GENERAL, no postcondition failure/evaluation-error and inherited safety exit.
- [x] [tlc2/tool/Github766Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github766Test.java) — **Port complete**: `testSpec` → [tlc_github_766_java_test.go](../tlc_github_766_java_test.go); original explicit JSON dump, embedded config, noGenerateSpec=true, doDumpTrace=false, FINISHED/no GENERAL, no postcondition failure/evaluation-error and inherited safety exit.
- [x] [tlc2/tool/Github798ITest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github798ITest.java) — **Port complete**: `testSpec` → [tlc_github_798_807_java_test.go](../tlc_github_798_807_java_test.go); original FINISHED/no GENERAL, 2/1/0 stats, depth1, success exit and embedded config/noGenerateSpec/no-JSON settings.
- [x] [tlc2/tool/Github798NTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github798NTest.java) — **Port complete**: `testSpec` → [tlc_github_798_807_java_test.go](../tlc_github_798_807_java_test.go); original FINISHED/no GENERAL, 3/2/0 stats, depth2, success exit and embedded config/noGenerateSpec/no-JSON settings.
- [x] [tlc2/tool/Github807Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github807Test.java) — **Port complete**: `testSpec` → [tlc_github_798_807_java_test.go](../tlc_github_798_807_java_test.go); original complete three-state trace, exact Add action/location, 3/3/0 stats, depth3, postcondition checks and inherited safety exit; debugger=false/noGenerateSpec=true/no JSON retained. Fixed production action decomposition to use the applied operator declaration for state-level arguments, as Java does.
- [x] [tlc2/tool/Github817Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817Test.java) — **Port complete**: `testSpec` → [tlc_github_817_java_test.go](../tlc_github_817_java_test.go); original FINISHED/SUCCESS, 3/3/0 stats and inherited success exit; original assertZeroUncovered retained; embedded/alternate configs, noGenerateSpec=true and doDumpTrace=false retained.
- [x] [tlc2/tool/Github817bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817bTest.java) — **Port complete**: `testSpec` → [tlc_github_817_java_test.go](../tlc_github_817_java_test.go); original FINISHED, 2/2/0 stats, action-property violation, complete init/in-progress trace with actions/ordinals and inherited liveness exit; embedded/alternate configs, noGenerateSpec=true and doDumpTrace=false retained.
- [x] [tlc2/tool/Github817cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817cTest.java) — **Port complete**: `testSpec` → [tlc_github_817_java_test.go](../tlc_github_817_java_test.go); original FINISHED/SUCCESS, 3/3/0 stats and inherited success exit; embedded/alternate configs, noGenerateSpec=true and doDumpTrace=false retained.
- [x] [tlc2/tool/Github817dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817dTest.java) — **Port complete**: `testSpec` → [tlc_github_817_java_test.go](../tlc_github_817_java_test.go); original FINISHED/SUCCESS, 3/3/0 stats and inherited success exit; embedded/alternate configs, noGenerateSpec=true and doDumpTrace=false retained.
- [x] [tlc2/tool/Github817eTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817eTest.java) — **Port complete**: `testSpec` → [tlc_github_817_java_test.go](../tlc_github_817_java_test.go); original FINISHED/SUCCESS, 3/3/0 stats and inherited success exit; embedded/alternate configs, noGenerateSpec=true and doDumpTrace=false retained.
- [x] [tlc2/tool/Github819Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github819Test.java) — **Port complete**: `testSpec` → [tlc_github_819_849_java_test.go](../tlc_github_819_849_java_test.go); original liveness-tautology diagnostic and inherited FAILURE_LIVENESS_EVAL exit; embedded config/noGenerateSpec=true/no JSON retained.
- [x] [tlc2/tool/Github849Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github849Test.java) — **Port complete**: `testSpec` → [tlc_github_819_849_java_test.go](../tlc_github_819_849_java_test.go); original FINISHED/no method-override diagnostic and inherited safety exit; embedded config and default coverage/debugger/DOT/JSON/forced trace generation retained.
- [x] [tlc2/tool/Github858Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github858Test.java) — **Port complete**: `testSpec` → [tlc_github_858_java_test.go](../tlc_github_858_java_test.go); original FINISHED/safety exit and exact complete state/action at ordinal100; original seeded simulation, JSON plus tlcaction dump, DOT/forced trace generation, coverage/debugger=false retained. Classpath supplies the original CommunityModules archive.
- [x] [tlc2/tool/Github866Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github866Test.java) — **Port complete**: `testSpec` → [tlc_github_866_java_test.go](../tlc_github_866_java_test.go); original two workers, FINISHED/no GENERAL/no postcondition false/evaluation error and inherited success exit; original no coverage/DOT/debugger/JSON/trace generation retained. Fixed native file.separator default in production system-property lookup; explicit overrides still take precedence.
- [x] [tlc2/tool/Github971aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971aTest.java) — **Port complete**: `testSpec` → [tlc_github_971_java_test.go](../tlc_github_971_java_test.go); all original recorded diagnostics, exact stats/depth where asserted, full state/action/ordinal trace, original property name, loop-back2 where asserted and inherited exit; two workers, original three/four-count queue latch and complete MemStateQueue delegation, original lncheck/config/no coverage/DOT/debugger/JSON/trace generation retained.
- [x] [tlc2/tool/Github971bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971bTest.java) — **Port complete**: `testSpec` → [tlc_github_971_java_test.go](../tlc_github_971_java_test.go); all original recorded diagnostics, exact stats/depth where asserted, full state/action/ordinal trace, original property name, loop-back2 where asserted and inherited exit; two workers, original three/four-count queue latch and complete MemStateQueue delegation, original lncheck/config/no coverage/DOT/debugger/JSON/trace generation retained.
- [x] [tlc2/tool/Github971cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971cTest.java) — **Port complete**: `testSpec` → [tlc_github_971_java_test.go](../tlc_github_971_java_test.go); all original recorded diagnostics, exact stats/depth where asserted, full state/action/ordinal trace, original property name, loop-back2 where asserted and inherited exit; two workers, original three/four-count queue latch and complete MemStateQueue delegation, original lncheck/config/no coverage/DOT/debugger/JSON/trace generation retained.
- [x] [tlc2/tool/Github971dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971dTest.java) — **Port complete**: `testSpec` → [tlc_github_971_java_test.go](../tlc_github_971_java_test.go); all original recorded diagnostics, exact stats/depth where asserted, full state/action/ordinal trace, original property name, loop-back2 where asserted and inherited exit; two workers, original three/four-count queue latch and complete MemStateQueue delegation, original lncheck/config/no coverage/DOT/debugger/JSON/trace generation retained.
- [x] [tlc2/tool/Github971eTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971eTest.java) — **Port complete**: `testSpec` → [tlc_github_971_java_test.go](../tlc_github_971_java_test.go); all original recorded diagnostics, exact stats/depth where asserted, full state/action/ordinal trace, original property name, loop-back2 where asserted and inherited exit; two workers, original three/four-count queue latch and complete MemStateQueue delegation, original lncheck/config/no coverage/DOT/debugger/JSON/trace generation retained.

### Traces, aliases, dump/load, and generated trace specs

Trace reconstruction, alias evaluation, trace races, external trace serialization, and generated trace-expression specifications.

- [ ] [tlc2/tool/DistributedTrace.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DistributedTrace.java) — **Reconcile**: `testSpec`. Unchanged Java fails assertNoTESpec and the inherited SUCCESS exit (actual 12, invariant violation). Temporary full Go translation matches both failures and passes original eleven-state trace/value/action/ordinal assertions with four workers. No weakened assertion or new persistent skip; keep uncredited pending disposition.
  Unchanged Java JUnit preflight (2026-10-04) fails two original assertions: a trace-exploration spec is generated despite assertNoTESpec, and inherited SUCCESS (0) differs from actual VIOLATION_SAFETY (12). Preserve both source assertions and keep this context pending; no Go translation or new skip credited.
- [x] [tlc2/tool/DumpAsDotTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DumpAsDotTest.java) — **Port complete**: `testSpec` → [tlc_dump_as_dot_java_test.go](../tlc_dump_as_dot_java_test.go); original exact DOT master bytes, FINISHED/no GENERAL, 18/11/0 stats, register42 first IntValue18, no postcondition failure/evaluation-error, zero-uncovered and inherited liveness exit. Original DOT/colorize/actionlabels/stuttering, default JSON/coverage/debugger/forced trace generation retained. Fixed production stuttering rank/snapshot continuation, Java HashMap/HashSet ordering (including resize and tree bins), and source legend/label formatting; no master-file normalization or weakened assertion.
- [x] [tlc2/tool/DumpLoadTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DumpLoadTraceTest.java) — **Port complete**: all 32 enabled methods and the three original `@Ignore` translations, including `testLivenessEWD840MC3DumpLoadTraceTLC` and `testLivenessEWD840MC3DumpLoadTraceTLCAutoWorkers`, in [tlc_dump_load_trace_java_test.go](../tlc_dump_load_trace_java_test.go).
  Execution qualification (October 6): full root snapshot `8d13d91` fails the JSON auto-worker safety and alias-sub2 methods. A 100-repeat targeted run on `795594a` reproduces the first method's prefix mismatch; ten earlier repeats passed. All methods remain explicit translations, so translation counts are unchanged, but execution parity is unresolved. Logs: `/mnt/oldrog/tmp/tlago-proof-statement-root-full.log` and `/mnt/oldrog/tmp/tlago-trace-roundtrip-reproduce-100.log`. Do not weaken assertions, change workers or infer completion from a passing repeat. A deterministic scratch reconstruction of the plain safety failure now produces identical seven-state replay and 36/7/1 statistics in pinned Java and Go; the earlier actual alias dump also has a matching Java replay receipt. These comparisons identify a source expectation limitation, but do not resolve the original prefix assertions or grant whole-suite pass credit.
  Other original enabled methods: `testLivenessMCDumpLoadTraceJSON`, `testLivenessMCDumpLoadTraceTLC`, `testLivenessMCDumpLoadTraceJSONAutoWorkers`, `testLivenessMCDumpLoadTraceTLCAutoWorkers`, `testSafetyDumpLoadTraceJSON`, `testSafetyDumpLoadTraceTLC`, `testSafetyDumpLoadTraceJSONAutoWorkers`, `testSafetyDumpLoadTraceTLCAutoWorkers`, `testLivenessBidirectionalDumpLoadTraceJSON`, `testLivenessBidirectionalDumpLoadTraceTLC`, `testLivenessBidirectionalDumpLoadTraceJSONAutoWorkers`, `testLivenessBidirectionalDumpLoadTraceTLCAutoWorkers`, `testSafetyTESpecEqAliasDumpLoadTraceJSON`, `testSafetyTESpecEqAliasDumpLoadTraceTLC`, `testSafetyTESpecEqAliasDumpLoadTraceJSONAutoWorkers`, `testSafetyTESpecEqAliasDumpLoadTraceTLCAutoWorkers`, `testLivenessExample1DumpLoadTraceJSON`, `testLivenessExample1DumpLoadTraceTLC`, `testLivenessExample1DumpLoadTraceJSONAutoWorkers`, `testLivenessExample1DumpLoadTraceTLCAutoWorkers`, `testSafetyDieHardAliasSubDumpLoadTraceJSON`, `testSafetyDieHardAliasSubDumpLoadTraceTLC`, `testSafetyDieHardAliasSub2DumpLoadTraceJSON`, `testSafetyDieHardAliasSub2DumpLoadTraceTLC`, `testSafetyDieHardAliasSub2DumpLoadTraceJSONAutoWorkers`, `testSafetyDieHardAliasSub2DumpLoadTraceTLCAutoWorkers`, `testSafetyDieHardAliasSupDumpLoadTraceJSON`, `testSafetyDieHardAliasSupDumpLoadTraceTLC`, `testSafetyDieHardAliasSupDumpLoadTraceJSONAutoWorkers`, `testSafetyDieHardAliasSupDumpLoadTraceTLCAutoWorkers`.
  Whole original dump/load phases, binary format, fingerprint polynomial 4, single-worker/auto-worker settings, disabled deadlock, nonempty file and trace checks, finished/exit/violation records, full state/ordinal equality or prefix assertions all retained. Both EWD840 methods pass unchanged Java and Go, including three complete repeated race runs each; the persistent complete Go class passes all 32 enabled methods normally and with the race detector. Historical paced Java execution can violate the replay-length assertion when partial liveness checks select another cycle. That source timing limitation remains documented; no partial checks, assertions, bounds or skip conditions were changed.

- [x] [tlc2/tool/PrintTraceRaceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PrintTraceRaceTest.java) — **Port complete**: `testSpec` in [tlc_print_trace_race_java_test.go](../tlc_print_trace_race_java_test.go). Restored inherited JSON dump setting; preserves four workers, original first-two-state/ordinal assertions, payload shape, diagnostics, stats, uncovered location and FAILURE_SAFETY_EVAL exit. The fully asserted helper also provides the original first phase for the TTrace test.

- [x] [tlc2/tool/TraceWithLargeSetOfInitialStatesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TraceWithLargeSetOfInitialStatesTest.java) — **Port complete**: `testSpec` in [tlc_trace_model_java_test.go](../tlc_trace_model_java_test.go). Restored inherited JSON trace-dump setup; all original assertions and model settings unchanged. Fully asserted helper also generates the original first-phase artifact for the TTrace recheck.

### Generated TTrace recheck variants

Every concrete *_TTraceTest and *_TTrace class, including inherited testSpec. These require the original first run and generated trace-spec artifacts; do not replace them with a rerun of the source model.

- [x] [tlc2/tool/BugzillaBug279Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/BugzillaBug279Test_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_ttrace_subset_depth_first_java_test.go](../tlc_ttrace_subset_depth_first_java_test.go). Rechecks the actual Go-generated artifact after the fully asserted original model run; preserves original settings, complete trace/ordinals, diagnostics, safety exit and zero-uncovered assertion. Retains deadlock checking, disabled DOT, 3/3/0 stats and the lazily represented subset. JSON conversion now invokes each lazy value’s source `toSetEnum` method rather than manually enumerating with different allocation, ordering and coverage behavior.
- [x] [tlc2/tool/DepthFirstDieHardTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DepthFirstDieHardTest_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_ttrace_subset_depth_first_java_test.go](../tlc_ttrace_subset_depth_first_java_test.go). Rechecks the actual Go-generated artifact after the fully asserted original model run; preserves original settings, complete trace/ordinals, diagnostics, safety exit and zero-uncovered assertion. The first phase retains DFID and blank action labels; the recheck asserts the exact original generated `_init`/`_next` source locations.
- [x] [tlc2/tool/DepthFirstErrorTraceTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DepthFirstErrorTraceTest_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_ttrace_subset_depth_first_java_test.go](../tlc_ttrace_subset_depth_first_java_test.go). Rechecks the actual Go-generated artifact after the fully asserted original model run; preserves original settings, complete trace/ordinals, diagnostics, safety exit and zero-uncovered assertion. The first phase retains DFID and blank action labels; the recheck asserts the exact original generated `_init`/`_next` source locations.
- [x] [tlc2/tool/EvalExceptionTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EvalExceptionTest_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_ttrace_eval_race_java_test.go](../tlc_ttrace_eval_race_java_test.go). Rechecks its actual generated Go artifact after the fully asserted original model run, with original JSON/DOT/no-generation/no-coverage settings and safety exit. Preserves FINISHED, 6/6/0 stats and all six exact states/actions/ordinals. Production `TLCState.getVals` now uses Java HashMap iteration and `MCState` consumes that order when generating invariants, preserving source record-field token order.
- [x] [tlc2/tool/Github461Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github461Test_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_ttrace_first_java_test.go](../tlc_ttrace_first_java_test.go). Executes the fully asserted original Go model first, then rechecks its actual generated monolithic TTrace artifact with the original basename/config, resolver, no-generation and no-coverage settings. Preserves the safety exit, complete five-state trace/actions/ordinals and zero-uncovered assertion.
- [x] [tlc2/tool/Github597Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github597Test_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_ttrace_first_java_test.go](../tlc_ttrace_first_java_test.go). Executes the fully asserted original Go model first, then rechecks its actual generated monolithic TTrace artifact with the original basename/config, resolver, no-generation and no-coverage settings. Preserves random fp/seed, disabled DOT, liveness exit, FINISHED/no GENERAL, temporal violation, counterexample, trace and loop-back existence assertions.
- [x] [tlc2/tool/PrintTraceRaceTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PrintTraceRaceTest_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_ttrace_eval_race_java_test.go](../tlc_ttrace_eval_race_java_test.go). Rechecks its actual generated Go artifact after the fully asserted original model run, with original JSON/DOT/no-generation/no-coverage settings and safety exit. Preserves four workers without debugger, FINISHED, 2/2/0 stats, no GENERAL, behavior diagnostic, original first-two-state/ordinal and payload-shape assertions.
- [x] [tlc2/tool/RandomElementSimulationTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementSimulationTest_TTraceTest.java) — **Port complete**: `test`.
  Go translation: [tlc_ttrace_random_element_java_test.go](../tlc_ttrace_random_element_java_test.go). Rechecks the actual generated Go artifact after the fully asserted original model run, preserving seeds, JSON/DOT, no-generation/no-coverage, worker/debugger settings and inherited safety exit. First phase remains simulation num=1 without debugger; recheck uses the source BFS defaults and checks all eleven states and exact generated _init/_next locations plus zero-uncovered.
- [x] [tlc2/tool/RandomElementT4Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementT4Test_TTraceTest.java) — **Port complete**: `test`.
  Go translation: [tlc_ttrace_random_element_java_test.go](../tlc_ttrace_random_element_java_test.go). Rechecks the actual generated Go artifact after the fully asserted original model run, preserving seeds, JSON/DOT, no-generation/no-coverage, worker/debugger settings and inherited safety exit. Both phases use four workers; all eleven original y/x component, bound and ordinal assertions remain.
- [x] [tlc2/tool/RandomElementTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementTest_TTraceTest.java) — **Port complete**: `test`.
  Go translation: [tlc_ttrace_random_element_java_test.go](../tlc_ttrace_random_element_java_test.go). Rechecks the actual generated Go artifact after the fully asserted original model run, preserving seeds, JSON/DOT, no-generation/no-coverage, worker/debugger settings and inherited safety exit. Preserves FINISHED/no TLC_BUG, behavior diagnostic, 11/11/0 stats, all eleven exact states/actions/ordinals and zero-uncovered.
- [x] [tlc2/tool/RandomElementXandYTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementXandYTest_TTraceTest.java) — **Port complete**: `test`.
  Go translation: [tlc_ttrace_random_element_java_test.go](../tlc_ttrace_random_element_java_test.go). Rechecks the actual generated Go artifact after the fully asserted original model run, preserving seeds, JSON/DOT, no-generation/no-coverage, worker/debugger settings and inherited safety exit. Preserves FINISHED/no TLC_BUG, behavior diagnostic, all three exact states/actions/ordinals and zero-uncovered.
- [x] [tlc2/tool/RandomSubsetATest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetATest_TTraceTest.java) — **Port complete**: `testSpec` (from `RandomSubset_TTrace`).
  Go translation: [tlc_ttrace_random_subset_java_test.go](../tlc_ttrace_random_subset_java_test.go). Rechecks its actual generated Go artifact after the fully asserted original model run, retaining JSON/DOT/no-generation/no-coverage and inherited safety exit. Preserves inherited RandomSubset_TTrace settings, FINISHED/no GENERAL, 1 initial state, 2/2/0 stats, depth 2 and both exact seeded states/actions/ordinals.
- [x] [tlc2/tool/RandomSubsetBTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetBTest_TTraceTest.java) — **Port complete**: `testSpec` (from `RandomSubset_TTrace`).
  Go translation: [tlc_ttrace_random_subset_java_test.go](../tlc_ttrace_random_subset_java_test.go). Rechecks its actual generated Go artifact after the fully asserted original model run, retaining JSON/DOT/no-generation/no-coverage and inherited safety exit. Preserves inherited RandomSubset_TTrace settings, FINISHED/no GENERAL, 1 initial state, 2/2/0 stats, depth 2 and both exact seeded states/actions/ordinals.
- [x] [tlc2/tool/RandomSubsetNextT4Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetNextT4Test_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_ttrace_random_subset_java_test.go](../tlc_ttrace_random_subset_java_test.go). Rechecks its actual generated Go artifact after the fully asserted original model run, retaining JSON/DOT/no-generation/no-coverage and inherited safety exit. Preserves four workers in both phases without debugger, FINISHED/no TLC_BUG, behavior diagnostic, all eleven state components/bounds/ordinals and fresh getVals snapshots.
- [x] [tlc2/tool/RandomSubsetNextTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetNextTest_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_ttrace_random_subset_java_test.go](../tlc_ttrace_random_subset_java_test.go). Rechecks its actual generated Go artifact after the fully asserted original model run, retaining JSON/DOT/no-generation/no-coverage and inherited safety exit. Preserves source seed 15041980, FINISHED/no TLC_BUG, behavior diagnostic, 11/11/0 recheck stats, all eleven exact states/actions/ordinals and zero-uncovered.
- [x] [tlc2/tool/RandomSubsetTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetTest_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_ttrace_random_subset_java_test.go](../tlc_ttrace_random_subset_java_test.go). Rechecks its actual generated Go artifact after the fully asserted original model run, retaining JSON/DOT/no-generation/no-coverage and inherited safety exit. Preserves FINISHED/no GENERAL, 1 initial state, 2/2/0 stats, depth 2, original generated _init/_next locations, getVals snapshots, every component/tuple/bound/UNCHANGED assertion and zero-uncovered. Retains firstX in the original y upper-bound expression.
- [x] [tlc2/tool/TLCGetLevelTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCGetLevelTest_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_ttrace_level_view_java_test.go](../tlc_ttrace_level_view_java_test.go). Rechecks its actual generated Go artifact after the fully asserted original model run, with original JSON/DOT/no-generation/no-coverage/debugger settings. Preserves liveness exit, FINISHED, 4/4/0 stats, no GENERAL, temporal-violation/counterexample/trace diagnostics, all four exact states/actions/ordinals and stuttering state5.
- [x] [tlc2/tool/TraceWithLargeSetOfInitialStatesTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TraceWithLargeSetOfInitialStatesTest_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_ttrace_level_view_java_test.go](../tlc_ttrace_level_view_java_test.go). Rechecks its actual generated Go artifact after the fully asserted original model run, with original JSON/DOT/no-generation/no-coverage/debugger settings. Preserves -maxSetSize10 in both phases, safety exit, FINISHED/no GENERAL/no TLC_BUG, behavior diagnostic, both exact states/ordinals, every original generated _init/_next source location and zero-uncovered.
- [x] [tlc2/tool/ViewMapTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ViewMapTest_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_ttrace_level_view_java_test.go](../tlc_ttrace_level_view_java_test.go). Rechecks its actual generated Go artifact after the fully asserted original model run, with original JSON/DOT/no-generation/no-coverage/debugger settings. Preserves -view in both phases, safety exit, FINISHED/no GENERAL/no TLC_BUG, behavior diagnostic, all eight full states including pc, ordinals and every original generated _init/_next source location.
- [x] [tlc2/tool/checkpoint/CheckpointOnViolationTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/checkpoint/CheckpointOnViolationTest_TTraceTest.java) — **Port complete**: `testSpec`. Translation: [tlc_checkpoint_models_java_test.go](../tlc_checkpoint_models_java_test.go). Generates the actual original trace before rechecking; preserves inherited exit, checkpoint interval, exact 7/7/0 counts, all seven trace states and uncovered assertion.
- [x] [tlc2/tool/liveness/BidirectionalTransitions1BxTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions1BxTest_TTraceTest.java) — **Port complete**: `testSpec` (from `BidirectionalTransitions1B_TTrace`).
  Go translation: [tlc_bidirectional_liveness_java_test.go](../tlc_bidirectional_liveness_java_test.go). Complete original method and constructor settings, JSON/DOT/debugger/coverage/generation defaults, FINISHED/no GENERAL and exact stats retained. Preserves liveness exit, temporal-violation/counterexample/trace diagnostics, every exact state/action/ordinal and back-to-state1. Runs its fully asserted original model first, then rechecks the actual generated Go artifact with source no-generation/no-coverage settings and exact 4/3/0 or 5/4/0 recheck stats.
- [x] [tlc2/tool/liveness/BidirectionalTransitions1ByTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions1ByTest_TTraceTest.java) — **Port complete**: `testSpec` (from `BidirectionalTransitions1B_TTrace`).
  Go translation: [tlc_bidirectional_liveness_java_test.go](../tlc_bidirectional_liveness_java_test.go). Complete original method and constructor settings, JSON/DOT/debugger/coverage/generation defaults, FINISHED/no GENERAL and exact stats retained. Preserves liveness exit, temporal-violation/counterexample/trace diagnostics, every exact state/action/ordinal and back-to-state1. Runs its fully asserted original model first, then rechecks the actual generated Go artifact with source no-generation/no-coverage settings and exact 4/3/0 or 5/4/0 recheck stats.
- [x] [tlc2/tool/liveness/BidirectionalTransitions2CxTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions2CxTest_TTraceTest.java) — **Port complete**: `testSpec` (from `BidirectionalTransitions2C_TTrace`).
  Go translation: [tlc_bidirectional_liveness_java_test.go](../tlc_bidirectional_liveness_java_test.go). Complete original method and constructor settings, JSON/DOT/debugger/coverage/generation defaults, FINISHED/no GENERAL and exact stats retained. Preserves liveness exit, temporal-violation/counterexample/trace diagnostics, every exact state/action/ordinal and back-to-state1. Runs its fully asserted original model first, then rechecks the actual generated Go artifact with source no-generation/no-coverage settings and exact 4/3/0 or 5/4/0 recheck stats.
- [x] [tlc2/tool/liveness/BidirectionalTransitions2CyTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions2CyTest_TTraceTest.java) — **Port complete**: `testSpec` (from `BidirectionalTransitions2C_TTrace`).
  Go translation: [tlc_bidirectional_liveness_java_test.go](../tlc_bidirectional_liveness_java_test.go). Complete original method and constructor settings, JSON/DOT/debugger/coverage/generation defaults, FINISHED/no GENERAL and exact stats retained. Preserves liveness exit, temporal-violation/counterexample/trace diagnostics, every exact state/action/ordinal and back-to-state1. Runs its fully asserted original model first, then rechecks the actual generated Go artifact with source no-generation/no-coverage settings and exact 4/3/0 or 5/4/0 recheck stats.
- [x] [tlc2/tool/liveness/ChooseTableauSymmetryTestA_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ChooseTableauSymmetryTestA_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_choose_tableau_symmetry_java_test.go](../tlc_choose_tableau_symmetry_java_test.go). Preserves liveness exit, FINISHED/no GENERAL, temporal-violation/counterexample/trace diagnostics, all five exact states/ordinals/action labels, both original loop-back3 parameters and zero-uncovered. Rechecks the actual generated Go artifact after the fully asserted original phase, with source JSON/DOT/debugger/no-generation/no-coverage, 6/5/0 stats and exact generated _init/_next source locations.
- [x] [tlc2/tool/liveness/CodePlexBug08AgentRingTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08AgentRingTest_TTraceTest.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Fully asserted original first run and actual generated-artifact recheck, including exact graph file sizes and complete counterexample.
- [x] [tlc2/tool/liveness/CodePlexBug08EWD840FL1Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL1Test_TTraceTest.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Fully asserted original first run and actual generated-artifact recheck, including exact graph file sizes and complete counterexample.
- [x] [tlc2/tool/liveness/CodePlexBug08EWD840FL2Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL2Test_TTraceTest.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Fully asserted original first run and actual generated-artifact recheck, including exact graph file sizes and complete counterexample.
- [x] [tlc2/tool/liveness/CodePlexBug08EWD840FL3Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL3Test_TTraceTest.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Fully asserted original first run and actual generated-artifact recheck, including exact graph file sizes and complete counterexample.
- [x] [tlc2/tool/liveness/CodePlexBug08EWD840FL4Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL4Test_TTraceTest.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Fully asserted original first run and actual generated-artifact recheck, including exact graph file sizes and complete counterexample.
- [x] [tlc2/tool/liveness/CodePlexBug08Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08Test_TTraceTest.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Fully asserted original first run and actual generated-artifact recheck, including exact graph file sizes and complete counterexample.
- [x] [tlc2/tool/liveness/CodePlexBug08aTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08aTest_TTraceTest.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Fully asserted original first run and actual generated-artifact recheck, including exact graph file sizes and complete counterexample.
- [x] [tlc2/tool/liveness/ErrorTraceConstructionTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ErrorTraceConstructionTest_TTraceTest.java) — **Port complete**: `testSpec` → [tlc_error_trace_construction_java_test.go](../tlc_error_trace_construction_java_test.go). Fully asserted original run followed by actual generated-artifact recheck, preserving all eight states, ordinals, exact graph sizes and generated loop-back action.
- [x] [tlc2/tool/liveness/LoopTestForcedPartial_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LoopTestForcedPartial_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_liveness_remaining_ttrace_java_test.go](../tlc_liveness_remaining_ttrace_java_test.go). Whole original method, constructor and inherited runner/exit settings retained. Rechecks the actual generated artifact after all original model assertions; preserves exact statistics, traces, loop/stuttering, graph sizes and generated action locations where asserted.
- [x] [tlc2/tool/liveness/LoopTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LoopTest_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_liveness_remaining_ttrace_java_test.go](../tlc_liveness_remaining_ttrace_java_test.go). Whole original method, constructor and inherited runner/exit settings retained. Rechecks the actual generated artifact after all original model assertions; preserves exact statistics, traces, loop/stuttering, graph sizes and generated action locations where asserted.
- [x] [tlc2/tool/liveness/OneBitMutexNoSymmetryTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/OneBitMutexNoSymmetryTest_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_liveness_remaining_ttrace_java_test.go](../tlc_liveness_remaining_ttrace_java_test.go). Whole original method, constructor and inherited runner/exit settings retained. Rechecks the actual generated artifact after all original model assertions; preserves exact statistics, traces, loop/stuttering, graph sizes and generated action locations where asserted.
- [x] [tlc2/tool/liveness/Test3_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Test3_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_liveness_remaining_ttrace_java_test.go](../tlc_liveness_remaining_ttrace_java_test.go). Whole original method, constructor and inherited runner/exit settings retained. Rechecks the actual generated artifact after all original model assertions; preserves exact statistics, traces, loop/stuttering, graph sizes and generated action locations where asserted.
- [x] [tlc2/tool/liveness/UnsymmetricModelCheckerTestA_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/UnsymmetricModelCheckerTestA_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_liveness_remaining_ttrace_java_test.go](../tlc_liveness_remaining_ttrace_java_test.go). Whole original method, constructor and inherited runner/exit settings retained. Rechecks the actual generated artifact after all original model assertions; preserves exact statistics, traces, loop/stuttering, graph sizes and generated action locations where asserted.
- [x] [tlc2/tool/liveness/simulation/Example1Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/Example1Test_TTraceTest.java) — **Port complete**: `testSpec` (from `AbstractExample_TTrace`).
  Go translation: [tlc_simulation_ttrace_java_test.go](../tlc_simulation_ttrace_java_test.go). Whole original constructor, inherited runner/exit settings and assertions retained; rechecks the actual generated artifact after the full original simulation assertions.
- [x] [tlc2/tool/liveness/simulation/Example2Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/Example2Test_TTraceTest.java) — **Port complete**: `testSpec` (from `AbstractExample_TTrace`).
  Go translation: [tlc_simulation_ttrace_java_test.go](../tlc_simulation_ttrace_java_test.go). Whole original constructor, inherited runner/exit settings and assertions retained; rechecks the actual generated artifact after the full original simulation assertions.
- [x] [tlc2/tool/liveness/simulation/LiveCheckExample1Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckExample1Test_TTraceTest.java) — **Port complete**: `testSpec` (from `AbstractExample_TTrace`).
  Go translation: [tlc_simulation_ttrace_java_test.go](../tlc_simulation_ttrace_java_test.go). Whole original constructor, inherited runner/exit settings and assertions retained; rechecks the actual generated artifact after the full original simulation assertions.
- [x] [tlc2/tool/liveness/simulation/LiveCheckExample2Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckExample2Test_TTraceTest.java) — **Port complete**: `testSpec` (from `AbstractExample_TTrace`).
  Go translation: [tlc_simulation_ttrace_java_test.go](../tlc_simulation_ttrace_java_test.go). Whole original constructor, inherited runner/exit settings and assertions retained; rechecks the actual generated artifact after the full original simulation assertions.
- [x] [tlc2/tool/liveness/simulation/LiveCheckSimulationTest2a_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckSimulationTest2a_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_simulation_ttrace_java_test.go](../tlc_simulation_ttrace_java_test.go). Whole original constructor, inherited runner/exit settings and assertions retained; rechecks the actual generated artifact after the full original simulation assertions.
- [x] [tlc2/tool/liveness/simulation/SimulationTest2a_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/SimulationTest2a_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_simulation_ttrace_java_test.go](../tlc_simulation_ttrace_java_test.go). Whole original constructor, inherited runner/exit settings and assertions retained; rechecks the actual generated artifact after the full original simulation assertions.
- [x] [tlc2/tool/liveness/simulation/StutteringTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/StutteringTest_TTraceTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_simulation_ttrace_java_test.go](../tlc_simulation_ttrace_java_test.go). Whole original constructor, inherited runner/exit settings and assertions retained; rechecks the actual generated artifact after the full original simulation assertions.

### Liveness and fairness model regressions

Temporal semantics, fairness, constraints, symmetry, double negation, counterexample construction, and original Examples models.

- [x] [tlc2/tool/NoFairnessButLivePropATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropATest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_no_fairness_models_java_test.go](../tlc_no_fairness_models_java_test.go). Full original testSpec and constructor configuration, VIOLATION_LIVENESS exit, FINISHED/3/2/0/depth2, exact warning assertions and zero-uncovered retained. Coverage=false override, JSON/DOT/debugger/forced generation and byte-identical model/config retained.
- [x] [tlc2/tool/NoFairnessButLivePropBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropBTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_no_fairness_models_java_test.go](../tlc_no_fairness_models_java_test.go). Full original testSpec and constructor configuration, VIOLATION_LIVENESS exit, FINISHED/3/2/0/depth2, exact warning assertions and zero-uncovered retained. Coverage=false override, JSON/DOT/debugger/forced generation and byte-identical model/config retained.
- [x] [tlc2/tool/NoFairnessButLivePropDTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropDTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_no_fairness_models_java_test.go](../tlc_no_fairness_models_java_test.go). Full original testSpec and constructor configuration, VIOLATION_LIVENESS exit, FINISHED/3/2/0/depth2, exact warning assertions and zero-uncovered retained. Coverage=false override, JSON/DOT/debugger/forced generation and byte-identical model/config retained.
- [x] [tlc2/tool/NoFairnessButLivePropGTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropGTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_no_fairness_models_java_test.go](../tlc_no_fairness_models_java_test.go). Full original testSpec and constructor configuration, SUCCESS exit, FINISHED/3/2/0/depth2, exact warning assertions and zero-uncovered retained. Coverage=false override, JSON/DOT/debugger/forced generation and byte-identical model/config retained.
- [x] [tlc2/tool/NoFairnessButLivePropHTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropHTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_no_fairness_models_java_test.go](../tlc_no_fairness_models_java_test.go). Full original testSpec and constructor configuration, VIOLATION_LIVENESS exit, FINISHED/3/2/0/depth2, exact warning assertions and zero-uncovered retained. Coverage=false override, JSON/DOT/debugger/forced generation and byte-identical model/config retained.
- [x] [tlc2/tool/NoFairnessButLivePropNoWarningCustomFairTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropNoWarningCustomFairTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_no_fairness_models_java_test.go](../tlc_no_fairness_models_java_test.go). Full original testSpec and constructor configuration, SUCCESS exit, FINISHED/3/2/0/depth2, exact warning assertions and zero-uncovered retained. Coverage=false override, JSON/DOT/debugger/forced generation and byte-identical model/config retained.
- [x] [tlc2/tool/NoFairnessButLivePropNoWarningNoFairTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropNoWarningNoFairTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_no_fairness_models_java_test.go](../tlc_no_fairness_models_java_test.go). Full original testSpec and constructor configuration, SUCCESS exit, FINISHED/3/2/0/depth2, exact warning assertions and zero-uncovered retained. Coverage=false override, JSON/DOT/debugger/forced generation and byte-identical model/config retained.
- [x] [tlc2/tool/NoFairnessButLivePropNoWarningWithFairTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropNoWarningWithFairTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_no_fairness_models_java_test.go](../tlc_no_fairness_models_java_test.go). Full original testSpec and constructor configuration, SUCCESS exit, FINISHED/3/2/0/depth2, exact warning assertions and zero-uncovered retained. Coverage=false override, JSON/DOT/debugger/forced generation and byte-identical model/config retained.
- [x] [tlc2/tool/NoFairnessButLivePropOTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropOTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_no_fairness_models_java_test.go](../tlc_no_fairness_models_java_test.go). Full original testSpec and constructor configuration, VIOLATION_LIVENESS exit, FINISHED/3/2/0/depth2, exact warning assertions and zero-uncovered retained. Coverage=false override, JSON/DOT/debugger/forced generation and byte-identical model/config retained.
- [x] [tlc2/tool/NoFairnessButLivePropPTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropPTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_no_fairness_models_java_test.go](../tlc_no_fairness_models_java_test.go). Full original testSpec and constructor configuration, SUCCESS exit, FINISHED/3/2/0/depth2, exact warning assertions and zero-uncovered retained. Coverage=false override, JSON/DOT/debugger/forced generation and byte-identical model/config retained. Preserves the no-temporal-violation assertion and absence of the deferred fairness warning on success.
- [x] [tlc2/tool/PossibleFailActionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PossibleFailActionTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_possible_models_java_test.go](../tlc_possible_models_java_test.go). Full original testSpec and constructor configuration retained, with JSON/DOT/coverage/debugger/forced generation and byte-identical fixtures. VIOLATION_ASSUMPTION exit, FINISHED, unwitnessed diagnostic and exact original predicate name in the first record's first parameter preserved.
- [x] [tlc2/tool/PossibleFailMixedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PossibleFailMixedTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_possible_models_java_test.go](../tlc_possible_models_java_test.go). Full original testSpec and constructor configuration retained, with JSON/DOT/coverage/debugger/forced generation and byte-identical fixtures. VIOLATION_ASSUMPTION exit, FINISHED, unwitnessed diagnostic and exact original predicate name in the first record's first parameter preserved.
- [x] [tlc2/tool/PossibleFailNoTransTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PossibleFailNoTransTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_possible_models_java_test.go](../tlc_possible_models_java_test.go). Full original testSpec and constructor configuration retained, with JSON/DOT/coverage/debugger/forced generation and byte-identical fixtures. VIOLATION_ASSUMPTION exit, FINISHED, unwitnessed diagnostic and exact original predicate name in the first record's first parameter preserved.
- [x] [tlc2/tool/PossibleFailStateTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PossibleFailStateTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_possible_models_java_test.go](../tlc_possible_models_java_test.go). Full original testSpec and constructor configuration retained, with JSON/DOT/coverage/debugger/forced generation and byte-identical fixtures. VIOLATION_ASSUMPTION exit, FINISHED, unwitnessed diagnostic and exact original predicate name in the first record's first parameter preserved.
- [x] [tlc2/tool/PossibleTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PossibleTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_possible_models_java_test.go](../tlc_possible_models_java_test.go). Full original testSpec and constructor configuration retained, with JSON/DOT/coverage/debugger/forced generation and byte-identical fixtures. SUCCESS exit, FINISHED and absence of GENERAL/postcondition-false/unwitnessed diagnostics preserved.
- [x] [tlc2/tool/liveness/BidirectionalTransitions1BxTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions1BxTest.java) — **Port complete**: `testSpec` (from `BidirectionalTransitions1BTest`).
  Go translation: [tlc_bidirectional_liveness_java_test.go](../tlc_bidirectional_liveness_java_test.go). Complete original method and constructor settings, JSON/DOT/debugger/coverage/generation defaults, FINISHED/no GENERAL and exact stats retained. Preserves liveness exit, temporal-violation/counterexample/trace diagnostics, every exact state/action/ordinal and back-to-state1. Preserves original Prop1Bx/Prop1By/Prop2Cx/Prop2Cy first-record property name and 13/3/0 or 9/4/0 stats; the 2C variants retain config arguments without a .cfg suffix.
- [x] [tlc2/tool/liveness/BidirectionalTransitions1ByTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions1ByTest.java) — **Port complete**: `testSpec` (from `BidirectionalTransitions1BTest`).
  Go translation: [tlc_bidirectional_liveness_java_test.go](../tlc_bidirectional_liveness_java_test.go). Complete original method and constructor settings, JSON/DOT/debugger/coverage/generation defaults, FINISHED/no GENERAL and exact stats retained. Preserves liveness exit, temporal-violation/counterexample/trace diagnostics, every exact state/action/ordinal and back-to-state1. Preserves original Prop1Bx/Prop1By/Prop2Cx/Prop2Cy first-record property name and 13/3/0 or 9/4/0 stats; the 2C variants retain config arguments without a .cfg suffix.
- [x] [tlc2/tool/liveness/BidirectionalTransitions1Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions1Test.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_bidirectional_liveness_java_test.go](../tlc_bidirectional_liveness_java_test.go). Complete original method and constructor settings, JSON/DOT/debugger/coverage/generation defaults, FINISHED/no GENERAL and exact stats retained. Preserves successful exit, exact depth and zero-uncovered assertion.
- [x] [tlc2/tool/liveness/BidirectionalTransitions2CxTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions2CxTest.java) — **Port complete**: `testSpec` (from `BidirectionalTransitions2CTest`).
  Go translation: [tlc_bidirectional_liveness_java_test.go](../tlc_bidirectional_liveness_java_test.go). Complete original method and constructor settings, JSON/DOT/debugger/coverage/generation defaults, FINISHED/no GENERAL and exact stats retained. Preserves liveness exit, temporal-violation/counterexample/trace diagnostics, every exact state/action/ordinal and back-to-state1. Preserves original Prop1Bx/Prop1By/Prop2Cx/Prop2Cy first-record property name and 13/3/0 or 9/4/0 stats; the 2C variants retain config arguments without a .cfg suffix.
- [x] [tlc2/tool/liveness/BidirectionalTransitions2CyTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions2CyTest.java) — **Port complete**: `testSpec` (from `BidirectionalTransitions2CTest`).
  Go translation: [tlc_bidirectional_liveness_java_test.go](../tlc_bidirectional_liveness_java_test.go). Complete original method and constructor settings, JSON/DOT/debugger/coverage/generation defaults, FINISHED/no GENERAL and exact stats retained. Preserves liveness exit, temporal-violation/counterexample/trace diagnostics, every exact state/action/ordinal and back-to-state1. Preserves original Prop1Bx/Prop1By/Prop2Cx/Prop2Cy first-record property name and 13/3/0 or 9/4/0 stats; the 2C variants retain config arguments without a .cfg suffix.
- [x] [tlc2/tool/liveness/BidirectionalTransitions2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions2Test.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_bidirectional_liveness_java_test.go](../tlc_bidirectional_liveness_java_test.go). Complete original method and constructor settings, JSON/DOT/debugger/coverage/generation defaults, FINISHED/no GENERAL and exact stats retained. Preserves successful exit, exact depth and zero-uncovered assertion.
- [x] [tlc2/tool/liveness/ChooseTableauSymmetryTestA.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ChooseTableauSymmetryTestA.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_choose_tableau_symmetry_java_test.go](../tlc_choose_tableau_symmetry_java_test.go). Preserves liveness exit, FINISHED/no GENERAL, temporal-violation/counterexample/trace diagnostics, all five exact states/ordinals/action labels, both original loop-back3 parameters and zero-uncovered. Preserves original model/config/library imports, JSON/DOT/debugger/coverage/forced generation, 13/6/0 stats and both original violated-property aliases.
- [x] [tlc2/tool/liveness/CodePlexBug08AgentRing790Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08AgentRing790Test.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Original successful exit and presence/absence assertions, explicit config and coverage=false.
- [x] [tlc2/tool/liveness/CodePlexBug08AgentRingTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08AgentRingTest.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Full original assertions/settings, exact graph file sizes, complete states/actions/ordinals and postconditions where asserted.
- [x] [tlc2/tool/liveness/CodePlexBug08EWD840FL1Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL1Test.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Full original assertions/settings, exact graph file sizes, complete states/actions/ordinals and postconditions where asserted.
- [x] [tlc2/tool/liveness/CodePlexBug08EWD840FL2FromCheckpointTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL2FromCheckpointTest.java) — **Port complete**: `testSpec`. Translation: [tlc_checkpoint_recovery_java_test.go](../tlc_checkpoint_recovery_java_test.go). Restores unchanged Java checkpoint.zip into isolated temporary storage, retaining three workers, gzip, base liveness exit, exact recovery/final counts, graph sizes, ten-state trace and loop-back assertions. Production now restores intern tokens in TLC.Process before tool construction, matching Java instead of reloading after parsing.
- [x] [tlc2/tool/liveness/CodePlexBug08EWD840FL2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL2Test.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Full original assertions/settings, exact graph file sizes, complete states/actions/ordinals and postconditions where asserted.
- [x] [tlc2/tool/liveness/CodePlexBug08EWD840FL3Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL3Test.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Full original assertions/settings, exact graph file sizes, complete states/actions/ordinals and postconditions where asserted.
- [x] [tlc2/tool/liveness/CodePlexBug08EWD840FL4Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL4Test.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Full original assertions/settings, exact graph file sizes, complete states/actions/ordinals and postconditions where asserted.
- [x] [tlc2/tool/liveness/CodePlexBug08Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08Test.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Full original assertions/settings, exact graph file sizes, complete states/actions/ordinals and postconditions where asserted.
- [x] [tlc2/tool/liveness/CodePlexBug08aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08aTest.java) — **Port complete**: `testSpec` → [tlc_codeplex08_liveness_java_test.go](../tlc_codeplex08_liveness_java_test.go). Full original assertions/settings, exact graph file sizes, complete states/actions/ordinals and postconditions where asserted.
- [x] [tlc2/tool/liveness/EmptyOrderOfSolutionsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/EmptyOrderOfSolutionsTest.java) — **Port complete**: `testSpec` → [tlc_error_trace_construction_java_test.go](../tlc_error_trace_construction_java_test.go). Original tautology diagnostic and inherited FAILURE_LIVENESS_EVAL exit, JSON/DOT/debugger/coverage/generation settings.
- [x] [tlc2/tool/liveness/ErrorTraceConstructionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ErrorTraceConstructionTest.java) — **Port complete**: `testSpec` → [tlc_error_trace_construction_java_test.go](../tlc_error_trace_construction_java_test.go). Complete original eight-state counterexample, ordinals and named actions, exact graph sizes, exact loop-back action, coverage and runner settings.
- [x] [tlc2/tool/liveness/ExamplesACPNBTLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesACPNBTLCTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesAbaAsynByzTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesAbaAsynByzTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesAsyncTerminationDetectionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesAsyncTerminationDetectionTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesBcastByzTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesBcastByzTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesBcastFolkloreTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesBcastFolkloreTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesBlockingQueuePoisonAppleTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesBlockingQueuePoisonAppleTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesBufferedRandomAccessFileTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesBufferedRandomAccessFileTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesCf1sFolkloreTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesCf1sFolkloreTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesCoffeeCan100BeansTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesCoffeeCan100BeansTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesEWD840Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesEWD840Test.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesEWD998ChanIDTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesEWD998ChanIDTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesEWD998Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesEWD998Test.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesEnvironmentControllerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesEnvironmentControllerTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesHuangTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesHuangTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesLiveHourClockTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesLiveHourClockTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesLockHSTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesLockHSTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesMCAlternatingBitTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCAlternatingBitTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesMCDistributedReplicatedLogTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCDistributedReplicatedLogTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured. This is a local single-worker liveness model, not distributed TLC transport; preserves 271/37/0 statistics, all six states and the named Extend(n2) back-edge to state 2.
- [x] [tlc2/tool/liveness/ExamplesMCEWD687aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCEWD687aTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesMCLiveInternalMemoryTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCLiveInternalMemoryTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesMCLiveWriteThroughCacheTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCLiveWriteThroughCacheTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesMCWriteThroughCacheTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCWriteThroughCacheTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesMCYoYoNoPruningTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCYoYoNoPruningTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesNbacgGuer01Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesNbacgGuer01Test.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesPrisonersTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesPrisonersTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesSDPAttackNewSolutionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSDPAttackNewSolutionTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesSDPAttackTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSDPAttackTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured. Preserves exact 1,103/526/175 statistics and all eleven original states using the complete 21-argument formatter and fourteen constant strings.
- [x] [tlc2/tool/liveness/ExamplesSchedulingAllocatorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSchedulingAllocatorTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesSimpleAllocatorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSimpleAllocatorTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesSingleLaneBridgeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSingleLaneBridgeTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesSpanTreeRandomTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSpanTreeRandomTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesSpanTreeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSpanTreeTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/ExamplesSyncTerminationDetectionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSyncTerminationDetectionTest.java) — **Port complete**: `testSpec` → [tlc_examples_liveness_java_test.go](../tlc_examples_liveness_java_test.go). All original presence/absence assertions, exact property alias where asserted, inherited exit and five settings overrides, final liveness mode, pristine fixtures and original CommunityModules manifest classpath retained; postconditions compare the complete counterexample where configured.
- [x] [tlc2/tool/liveness/Github1037Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github1037Test.java) — **Port complete**: `testSpec` → [tlc_liveness_success_initial_error_java_test.go](../tlc_liveness_success_initial_error_java_test.go). Preserves the complete original method, inherited exit checks, constructor configuration and all five settings overrides, default liveness strategy, and pristine fixtures.
- [x] [tlc2/tool/liveness/Github317Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github317Test.java) — **Port complete**: `testSpec` → [tlc_liveness_diagnostics_symmetry_java_test.go](../tlc_liveness_diagnostics_symmetry_java_test.go). Complete original method/settings/exit, full diagnostic stacks and single-report counts, source worker counts, actual graph bytes, statistics, named trace actions, exact coverage and postcondition register preserved where asserted; 14 byte-identical fixtures. Undefined fairness-variable failures retain Java coded exception parameters and expression/context metadata.
- [x] [tlc2/tool/liveness/Github317aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github317aTest.java) — **Port complete**: `testSpec` → [tlc_liveness_diagnostics_symmetry_java_test.go](../tlc_liveness_diagnostics_symmetry_java_test.go). Complete original method/settings/exit, full diagnostic stacks and single-report counts, source worker counts, actual graph bytes, statistics, named trace actions, exact coverage and postcondition register preserved where asserted; 14 byte-identical fixtures. Undefined fairness-variable failures retain Java coded exception parameters and expression/context metadata.
- [x] [tlc2/tool/liveness/Github604Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github604Test.java) — **Port complete**: `testSpec` → [tlc_liveness_four_simple_java_test.go](../tlc_liveness_four_simple_java_test.go). Full original method, inherited successful exit and constructor/settings retained; monolith configs, 17/4/0 and 10,507/103/0 statistics, depth, temporal-violation absence and zero-uncovered checks retained where asserted. All eight fixtures are byte-identical to Java.
- [x] [tlc2/tool/liveness/Github702Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github702Test.java) — **Port complete**: `testSpec` → [tlc_liveness_four_simple_java_test.go](../tlc_liveness_four_simple_java_test.go). Full original method, inherited successful exit and constructor/settings retained; monolith configs, 17/4/0 and 10,507/103/0 statistics, depth, temporal-violation absence and zero-uncovered checks retained where asserted. All eight fixtures are byte-identical to Java.
- [x] [tlc2/tool/liveness/Github710aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710aTest.java) — **Port complete**: `testSpec` → [tlc_liveness_finite_and_loop_java_test.go](../tlc_liveness_finite_and_loop_java_test.go). Full original constructor/settings, exit, diagnostic/property aliases, complete trace/actions/ordinals, postcondition register, actual graph sizes and exact coverage retained where asserted; 22 pristine fixture files. Forced partial checking retains Java result-code semantics and continues inserting the complete graph.
- [x] [tlc2/tool/liveness/Github710bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710bTest.java) — **Port complete**: `testSpec` → [tlc_liveness_finite_and_loop_java_test.go](../tlc_liveness_finite_and_loop_java_test.go). Full original constructor/settings, exit, diagnostic/property aliases, complete trace/actions/ordinals, postcondition register, actual graph sizes and exact coverage retained where asserted; 22 pristine fixture files. Forced partial checking retains Java result-code semantics and continues inserting the complete graph.
- [x] [tlc2/tool/liveness/Github710cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710cTest.java) — **Port complete**: `testSpec` → [tlc_liveness_finite_and_loop_java_test.go](../tlc_liveness_finite_and_loop_java_test.go). Full original constructor/settings, exit, diagnostic/property aliases, complete trace/actions/ordinals, postcondition register, actual graph sizes and exact coverage retained where asserted; 22 pristine fixture files. Forced partial checking retains Java result-code semantics and continues inserting the complete graph.
- [x] [tlc2/tool/liveness/Github710dFairTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710dFairTest.java) — **Port complete**: `testSpec` → [tlc_liveness_finite_and_loop_java_test.go](../tlc_liveness_finite_and_loop_java_test.go). Full original constructor/settings, exit, diagnostic/property aliases, complete trace/actions/ordinals, postcondition register, actual graph sizes and exact coverage retained where asserted; 22 pristine fixture files. Forced partial checking retains Java result-code semantics and continues inserting the complete graph.
- [x] [tlc2/tool/liveness/Github710dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710dTest.java) — **Port complete**: `testSpec` → [tlc_liveness_finite_and_loop_java_test.go](../tlc_liveness_finite_and_loop_java_test.go). Full original constructor/settings, exit, diagnostic/property aliases, complete trace/actions/ordinals, postcondition register, actual graph sizes and exact coverage retained where asserted; 22 pristine fixture files. Forced partial checking retains Java result-code semantics and continues inserting the complete graph.
- [x] [tlc2/tool/liveness/Github710eFairTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710eFairTest.java) — **Port complete**: `testSpec` → [tlc_liveness_finite_and_loop_java_test.go](../tlc_liveness_finite_and_loop_java_test.go). Full original constructor/settings, exit, diagnostic/property aliases, complete trace/actions/ordinals, postcondition register, actual graph sizes and exact coverage retained where asserted; 22 pristine fixture files. Forced partial checking retains Java result-code semantics and continues inserting the complete graph.
- [x] [tlc2/tool/liveness/Github710eTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710eTest.java) — **Port complete**: `testSpec` → [tlc_liveness_finite_and_loop_java_test.go](../tlc_liveness_finite_and_loop_java_test.go). Full original constructor/settings, exit, diagnostic/property aliases, complete trace/actions/ordinals, postcondition register, actual graph sizes and exact coverage retained where asserted; 22 pristine fixture files. Forced partial checking retains Java result-code semantics and continues inserting the complete graph.
- [x] [tlc2/tool/liveness/Github790Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github790Test.java) — **Port complete**: `testSpec` → [tlc_liveness_success_initial_error_java_test.go](../tlc_liveness_success_initial_error_java_test.go). Preserves the complete original method, inherited exit checks, constructor configuration and all five settings overrides, default liveness strategy, and pristine fixtures.
- [x] [tlc2/tool/liveness/IncompatibleTypesLiveTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/IncompatibleTypesLiveTest.java) — **Port complete**: `testSpec` → [tlc_liveness_four_simple_java_test.go](../tlc_liveness_four_simple_java_test.go). Full original method, inherited successful exit and constructor/settings retained; monolith configs, 17/4/0 and 10,507/103/0 statistics, depth, temporal-violation absence and zero-uncovered checks retained where asserted. All eight fixtures are byte-identical to Java.
- [x] [tlc2/tool/liveness/InitialLivenessEvaluationErrorInvariantOnlyTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/InitialLivenessEvaluationErrorInvariantOnlyTest.java) — **Port complete**: `testSpec` → [tlc_liveness_success_initial_error_java_test.go](../tlc_liveness_success_initial_error_java_test.go). Preserves the complete original method, inherited exit checks, constructor configuration and all five settings overrides, default liveness strategy, and pristine fixtures. Original configuration-error/safety alternatives and exact diagnostic/statistics branches retained; invariant-only control requires safety failure and the complete two-state trace. Production now retains the original exception code before diagnostic replay.
- [x] [tlc2/tool/liveness/InitialLivenessEvaluationErrorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/InitialLivenessEvaluationErrorTest.java) — **Port complete**: `testSpec` → [tlc_liveness_success_initial_error_java_test.go](../tlc_liveness_success_initial_error_java_test.go). Preserves the complete original method, inherited exit checks, constructor configuration and all five settings overrides, default liveness strategy, and pristine fixtures. Original configuration-error/safety alternatives and exact diagnostic/statistics branches retained; invariant-only control requires safety failure and the complete two-state trace. Production now retains the original exception code before diagnostic replay.
- [ ] [tlc2/tool/liveness/LivenessSymmetryWarning.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LivenessSymmetryWarning.java) — **Reconcile**: `testSpec`. Unchanged Java and the complete Go draft both pass the three original assertions (FINISHED present, unsupported-liveness-symmetry warning present, GENERAL absent), then fail the inherited successful-exit assertion: expected 0, actual 13 from April25MC. The Go draft retains base debugger, coverage, DOT, JSON dump, forced trace generation, single worker, deterministic fingerprint/seed, disabled deadlock and completed-graph liveness settings. No production mismatch found; original assertions remain intact in ignored scratch, without translation credit or a new skip.
- [x] [tlc2/tool/liveness/LoopTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LoopTest.java) — **Port complete**: `testSpec` → [tlc_liveness_finite_and_loop_java_test.go](../tlc_liveness_finite_and_loop_java_test.go). Full original constructor/settings, exit, diagnostic/property aliases, complete trace/actions/ordinals, postcondition register, actual graph sizes and exact coverage retained where asserted; 22 pristine fixture files. Forced partial checking retains Java result-code semantics and continues inserting the complete graph.
- [x] [tlc2/tool/liveness/LoopTestForcedPartial.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LoopTestForcedPartial.java) — **Port complete**: `testSpec` → [tlc_liveness_finite_and_loop_java_test.go](../tlc_liveness_finite_and_loop_java_test.go). Full original constructor/settings, exit, diagnostic/property aliases, complete trace/actions/ordinals, postcondition register, actual graph sizes and exact coverage retained where asserted; 22 pristine fixture files. Forced partial checking retains Java result-code semantics and continues inserting the complete graph.
- [x] [tlc2/tool/liveness/LoopTestWeakFair.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LoopTestWeakFair.java) — **Port complete**: `testSpec` → [tlc_liveness_finite_and_loop_java_test.go](../tlc_liveness_finite_and_loop_java_test.go). Full original constructor/settings, exit, diagnostic/property aliases, complete trace/actions/ordinals, postcondition register, actual graph sizes and exact coverage retained where asserted; 22 pristine fixture files. Forced partial checking retains Java result-code semantics and continues inserting the complete graph.
- [x] [tlc2/tool/liveness/NoSymmetryTableauModelCheckerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/NoSymmetryTableauModelCheckerTest.java) — **Port complete**: `testSpec` → [tlc_liveness_diagnostics_symmetry_java_test.go](../tlc_liveness_diagnostics_symmetry_java_test.go). Complete original method/settings/exit, full diagnostic stacks and single-report counts, source worker counts, actual graph bytes, statistics, named trace actions, exact coverage and postcondition register preserved where asserted; 14 byte-identical fixtures. Undefined fairness-variable failures retain Java coded exception parameters and expression/context metadata.
- [x] [tlc2/tool/liveness/OneBitMutexNoSymmetryTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/OneBitMutexNoSymmetryTest.java) — **Port complete**: `testSpec` → [tlc_liveness_diagnostics_symmetry_java_test.go](../tlc_liveness_diagnostics_symmetry_java_test.go). Complete original method/settings/exit, full diagnostic stacks and single-report counts, source worker counts, actual graph bytes, statistics, named trace actions, exact coverage and postcondition register preserved where asserted; 14 byte-identical fixtures. Undefined fairness-variable failures retain Java coded exception parameters and expression/context metadata.
- [x] [tlc2/tool/liveness/TemporalDoubleNegationAlwaysDoubleNegationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationAlwaysDoubleNegationTest.java) — **Port complete**: `testValidProperty` → [tlc_temporal_double_negation_java_test.go](../tlc_temporal_double_negation_java_test.go). Preserves the complete original method, inherited exit checks, constructor configuration and all five settings overrides, default liveness strategy, and pristine fixtures.
- [x] [tlc2/tool/liveness/TemporalDoubleNegationAlwaysDualTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationAlwaysDualTest.java) — **Port complete**: `testValidProperty` → [tlc_temporal_double_negation_java_test.go](../tlc_temporal_double_negation_java_test.go). Preserves the complete original method, inherited exit checks, constructor configuration and all five settings overrides, default liveness strategy, and pristine fixtures.
- [x] [tlc2/tool/liveness/TemporalDoubleNegationEventuallyDoubleNegationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationEventuallyDoubleNegationTest.java) — **Port complete**: `testValidProperty` → [tlc_temporal_double_negation_java_test.go](../tlc_temporal_double_negation_java_test.go). Preserves the complete original method, inherited exit checks, constructor configuration and all five settings overrides, default liveness strategy, and pristine fixtures.
- [x] [tlc2/tool/liveness/TemporalDoubleNegationFairEventuallyDoubleNegationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationFairEventuallyDoubleNegationTest.java) — **Port complete**: `testValidProperty` → [tlc_temporal_double_negation_java_test.go](../tlc_temporal_double_negation_java_test.go). Preserves the complete original method, inherited exit checks, constructor configuration and all five settings overrides, default liveness strategy, and pristine fixtures.
- [x] [tlc2/tool/liveness/TemporalDoubleNegationFairLeadsToDualTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationFairLeadsToDualTest.java) — **Port complete**: `testValidProperty` → [tlc_temporal_double_negation_java_test.go](../tlc_temporal_double_negation_java_test.go). Preserves the complete original method, inherited exit checks, constructor configuration and all five settings overrides, default liveness strategy, and pristine fixtures.
- [x] [tlc2/tool/liveness/TemporalDoubleNegationImplicationDoubleNegationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationImplicationDoubleNegationTest.java) — **Port complete**: `testValidProperty` → [tlc_temporal_double_negation_java_test.go](../tlc_temporal_double_negation_java_test.go). Preserves the complete original method, inherited exit checks, constructor configuration and all five settings overrides, default liveness strategy, and pristine fixtures.
- [x] [tlc2/tool/liveness/TemporalDoubleNegationImplicationTautologyTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationImplicationTautologyTest.java) — **Port complete**: `testValidProperty` → [tlc_temporal_double_negation_java_test.go](../tlc_temporal_double_negation_java_test.go). Preserves the complete original method, inherited exit checks, constructor configuration and all five settings overrides, default liveness strategy, and pristine fixtures.
- [x] [tlc2/tool/liveness/TemporalDoubleNegationLeadsToDualTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationLeadsToDualTest.java) — **Port complete**: `testValidProperty` → [tlc_temporal_double_negation_java_test.go](../tlc_temporal_double_negation_java_test.go). Preserves the complete original method, inherited exit checks, constructor configuration and all five settings overrides, default liveness strategy, and pristine fixtures.
- [x] [tlc2/tool/liveness/Test3.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Test3.java) — **Port complete**: `testSpec` → [tlc_liveness_finite_and_loop_java_test.go](../tlc_liveness_finite_and_loop_java_test.go). Full original constructor/settings, exit, diagnostic/property aliases, complete trace/actions/ordinals, postcondition register, actual graph sizes and exact coverage retained where asserted; 22 pristine fixture files. Forced partial checking retains Java result-code semantics and continues inserting the complete graph.
- [x] [tlc2/tool/liveness/TwoPhaseCommitTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TwoPhaseCommitTest.java) — **Port complete**: `testSpec` → [tlc_liveness_four_simple_java_test.go](../tlc_liveness_four_simple_java_test.go). Full original method, inherited successful exit and constructor/settings retained; monolith configs, 17/4/0 and 10,507/103/0 statistics, depth, temporal-violation absence and zero-uncovered checks retained where asserted. All eight fixtures are byte-identical to Java. Original zero-uncovered assertion is retained across all reports: extra Go race instrumentation can make periodic reports expose transient zeros before final coverage, and unchanged Java reproduces the same assertion failure under slower/earlier-report pacing. Default normal Java and Go runs pass; no final-report filtering or timing flags were substituted.
- [x] [tlc2/tool/liveness/UnsymmetricModelCheckerTestA.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/UnsymmetricModelCheckerTestA.java) — **Port complete**: `testSpec` → [tlc_liveness_diagnostics_symmetry_java_test.go](../tlc_liveness_diagnostics_symmetry_java_test.go). Complete original method/settings/exit, full diagnostic stacks and single-report counts, source worker counts, actual graph bytes, statistics, named trace actions, exact coverage and postcondition register preserved where asserted; 14 byte-identical fixtures. Undefined fairness-variable failures retain Java coded exception parameters and expression/context metadata.
- [x] [tlc2/tool/liveness/UnsymmetricModelCheckerTestB.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/UnsymmetricModelCheckerTestB.java) — **Port complete**: `testSpec` → [tlc_liveness_diagnostics_symmetry_java_test.go](../tlc_liveness_diagnostics_symmetry_java_test.go). Complete original method/settings/exit, full diagnostic stacks and single-report counts, source worker counts, actual graph bytes, statistics, named trace actions, exact coverage and postcondition register preserved where asserted; 14 byte-identical fixtures. Undefined fairness-variable failures retain Java coded exception parameters and expression/context metadata.

- [x] [tlc2/tool/PossibleCountsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PossibleCountsTest.java) — **Port complete**: `testSpec` in [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go). Restored inherited JSON dump setup; all six original assertions, successful exit and model settings retained.

### Liveness graph, tableau, and expression helpers

Direct Java graph, node/table, particle-closure, and live-expression tests.

- [x] [tlc2/tool/liveness/DiskGraphTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/DiskGraphTest.java) — **Port complete**: `testGetPathWithoutInitNoTableau` (from `test/tlc2/tool/liveness/DiskGraphTest.java`), `testGetMinimalPathWithoutTableau` (from `test/tlc2/tool/liveness/DiskGraphTest.java`), `testPathWithTwoInitNodes` (from `test/tlc2/tool/liveness/DiskGraphTest.java`), `testAddSameGraphNodeTwice` (from `test/tlc2/tool/liveness/DiskGraphTest.java`), `testLookupExistingNode` (from `test/tlc2/tool/liveness/DiskGraphTest.java`), `testAddSameGraphNodeTwiceCorrectSuccessors` (from `test/tlc2/tool/liveness/DiskGraphTest.java`), `testGetPathPartialGraph` (from `test/tlc2/tool/liveness/DiskGraphTest.java`) → [liveness_disk_graph_java_test.go](liveness_disk_graph_java_test.go). All original method bodies, inputs and assertions retained. Production graph storage now uses the original BufferedRandomAccessFile, including logical length/cursor, flush and DiskGraph reset order; checkpoints retain source unbuffered DataInputStream/DataOutputStream primitives.
  Related Go checks: [tlc/liveness_graph_test.go](liveness_graph_test.go).
- [x] [tlc2/tool/liveness/GraphNodeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/GraphNodeTest.java) — **Port complete**: `testAllocateRealign` (from `test/tlc2/tool/liveness/GraphNodeTest.java`), `testRealign` (from `test/tlc2/tool/liveness/GraphNodeTest.java`), `testAllocateNested` (from `test/tlc2/tool/liveness/GraphNodeTest.java`), `testAllocateNestedRandom` (from `test/tlc2/tool/liveness/GraphNodeTest.java`), `testAllocateNegative` (from `test/tlc2/tool/liveness/GraphNodeTest.java`), `testAllocateAndSuccessorSize` (from `test/tlc2/tool/liveness/GraphNodeTest.java`) → [liveness_graph_node_java_test.go](liveness_graph_node_java_test.go). All original method bodies, inputs and assertions retained; complete allocation loops and Java Random seed4711 retained.
  Related Go checks: [tlc/liveness_graph_test.go](liveness_graph_test.go).
- [x] [tlc2/tool/liveness/LiveExprNodeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LiveExprNodeTest.java) — **Port complete**: `testLNBool` (from `test/tlc2/tool/liveness/LiveExprNodeTest.java`), `testLNState` (from `test/tlc2/tool/liveness/LiveExprNodeTest.java`) → [liveness_expression_java_test.go](liveness_expression_java_test.go). All original method bodies, inputs and assertions retained; four commented LNNext source TODO assertions remain comments.
  Related Go checks: [tlc/liveness_process_test.go](liveness_process_test.go).
- [x] [tlc2/tool/liveness/TBParTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TBParTest.java) — **Port complete**: `testParticleClosureInconsistentConstantLevel` (from `test/tlc2/tool/liveness/TBParTest.java`), `testParticleClosureInconsistentStateLevel` (from `test/tlc2/tool/liveness/TBParTest.java`), `testParticleClosureConsistentConstantLevel` (from `test/tlc2/tool/liveness/TBParTest.java`), `testParticleClosureConsistentStateLevel` (from `test/tlc2/tool/liveness/TBParTest.java`), `testParticleClosureExampleConstLevel` (from `test/tlc2/tool/liveness/TBParTest.java`), `testParticleClosureExampleStateLevel` (from `test/tlc2/tool/liveness/TBParTest.java`) → [liveness_tbpar_java_test.go](liveness_tbpar_java_test.go). All original method bodies, inputs and assertions retained; source object-identity assertions retain pointer equality.
  Related Go checks: [tlc/liveness_process_test.go](liveness_process_test.go).
- [x] [tlc2/tool/liveness/TableauDiskGraphTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TableauDiskGraphTest.java) — **Port complete**: `testGetPathWithoutInitNoTableau` (from `test/tlc2/tool/liveness/DiskGraphTest.java`), `testGetMinimalPathWithoutTableau` (from `test/tlc2/tool/liveness/DiskGraphTest.java`), `testPathWithTwoInitNodes` (from `test/tlc2/tool/liveness/DiskGraphTest.java`), `testAddSameGraphNodeTwice` (from `test/tlc2/tool/liveness/DiskGraphTest.java`), `testLookupExistingNode` (from `test/tlc2/tool/liveness/DiskGraphTest.java`), `testAddSameGraphNodeTwiceCorrectSuccessors` (from `test/tlc2/tool/liveness/DiskGraphTest.java`), `testGetPathPartialGraph` (from `test/tlc2/tool/liveness/DiskGraphTest.java`), `testGetShortestPath` (from `test/tlc2/tool/liveness/TableauDiskGraphTest.java`), `testUnifyingNodeInPath` (from `test/tlc2/tool/liveness/TableauDiskGraphTest.java`), `testUnifyingNodeShortestPath` (from `test/tlc2/tool/liveness/TableauDiskGraphTest.java`), `testPathWithTwoInitNodesWithTableau` (from `test/tlc2/tool/liveness/TableauDiskGraphTest.java`), `testGetPathWithTwoInits` (from `test/tlc2/tool/liveness/TableauDiskGraphTest.java`), `testNodeSetDone` (from `test/tlc2/tool/liveness/TableauDiskGraphTest.java`), `testGetPathWithTwoNodesWithSameFingerprint` (from `test/tlc2/tool/liveness/TableauDiskGraphTest.java`), `testLookupExistingNodeWithTidx` (from `test/tlc2/tool/liveness/TableauDiskGraphTest.java`), `testWhatsDoneIsDoneRRS` (from `test/tlc2/tool/liveness/TableauDiskGraphTest.java`), `testWhatsDoneIsDoneRSR` (from `test/tlc2/tool/liveness/TableauDiskGraphTest.java`), `testWhatsDoneIsDoneSRR` (from `test/tlc2/tool/liveness/TableauDiskGraphTest.java`) → [liveness_disk_graph_java_test.go](liveness_disk_graph_java_test.go). All original method bodies, inputs and assertions retained; includes all seven inherited DiskGraphTest methods with the original tableau index0 override. Production tableau graph storage now uses the original BufferedRandomAccessFile; reset retains source setLength operations and order without invoking DiskGraph.reset.
  Related Go checks: [tlc/liveness_graph_test.go](liveness_graph_test.go).
- [x] [tlc2/tool/liveness/TableauNodePtrTableTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TableauNodePtrTableTest.java) — **Port complete**: `testSetDoneBFSOrder` (from `test/tlc2/tool/liveness/TableauNodePtrTableTest.java`), `testSetDoneNoOrder` (from `test/tlc2/tool/liveness/TableauNodePtrTableTest.java`), `testSetDone` (from `test/tlc2/tool/liveness/TableauNodePtrTableTest.java`), `testSetDone2` (from `test/tlc2/tool/liveness/TableauNodePtrTableTest.java`), `testSetDone3` (from `test/tlc2/tool/liveness/TableauNodePtrTableTest.java`), `testIsDoneSPP` (from `test/tlc2/tool/liveness/TableauNodePtrTableTest.java`), `testIsDonePPS` (from `test/tlc2/tool/liveness/TableauNodePtrTableTest.java`), `testIsDonePSP` (from `test/tlc2/tool/liveness/TableauNodePtrTableTest.java`), `testRedundantMethodYieldSameResult` (from `test/tlc2/tool/liveness/TableauNodePtrTableTest.java`) → [liveness_tableau_node_ptr_java_test.go](liveness_tableau_node_ptr_java_test.go). All original method bodies, inputs and assertions retained.
  Related Go checks: [tlc/liveness_graph_test.go](liveness_graph_test.go).

### Simulation and multithreaded simulation

Both tool/ and tool/simulation/ simulator/worker classes, original simulation models, and liveness/simulation models. Preserve seeds, termination predicates, trace assertions, and worker counts.

- [x] [tlc2/tool/SimulationWorkerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/SimulationWorkerTest.java) — **Port complete**: `testGetTraceTLCState0`, `testGetTraceTLCState1`, `testGetTraceTLCState2`, `testGetTraceTLCState3`, `testGetTraceTLCState4`, `testGetTraceTLCState5`, `testGetTraceTLCState6`.
  Go translation: [tlc/simulation_worker_trace_java_test.go](simulation_worker_trace_java_test.go). All seven whole original methods retain their constructors, equality dummy and 37 size/initial-state/fingerprint/predecessor-identity assertions. Production StateVec storage and SimulationWorker trace traversal retain polymorphic state behavior, source synchronization and corrected predecessor levels after dropping finite stuttering.
  Related Go checks: [tlc/simulator_test.go](simulator_test.go), [tlc/worker_test.go](worker_test.go).
- [x] [tlc2/tool/SimulatorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/SimulatorTest.java) — **Port complete**: `testPrintBehaviorShouldPrintErrorState`.
  Go translation: [tlc_simulator_error_state_java_test.go](../tlc_simulator_error_state_java_test.go). Whole original method retains Github726 FastTool construction, zero workers, unbounded trace depth, seed/aril zero, the empty initial state, uncoded runtime exception and TLC_ERROR_STATE assertion. Production exposes the source exception-printing overload without a summary.
  Related Go checks: [tlc/simulator_test.go](simulator_test.go).
- [x] [tlc2/tool/liveness/simulation/Example1Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/Example1Test.java) — **Port complete**: `testSpec` (from `AbstractExampleTestCase`).
  Go translation: [tlc_simulation_counterexamples_java_test.go](../tlc_simulation_counterexamples_java_test.go). Whole original method, constructor and inherited runner/exit settings retained.
- [x] [tlc2/tool/liveness/simulation/Example2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/Example2Test.java) — **Port complete**: `testSpec` (from `AbstractExampleTestCase`).
  Go translation: [tlc_simulation_counterexamples_java_test.go](../tlc_simulation_counterexamples_java_test.go). Whole original method, constructor and inherited runner/exit settings retained.
- [x] [tlc2/tool/liveness/simulation/LiveCheckExample1Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckExample1Test.java) — **Port complete**: `testSpec` (from `AbstractExampleTestCase`).
  Go translation: [tlc_simulation_counterexamples_java_test.go](../tlc_simulation_counterexamples_java_test.go). Whole original method, constructor and inherited runner/exit settings retained. Source assumeTrue(false) is retained after every counterexample and trace assertion; only postcondition assertions are bypassed as in Java.
- [x] [tlc2/tool/liveness/simulation/LiveCheckExample2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckExample2Test.java) — **Port complete**: `testSpec` (from `AbstractExampleTestCase`).
  Go translation: [tlc_simulation_counterexamples_java_test.go](../tlc_simulation_counterexamples_java_test.go). Whole original method, constructor and inherited runner/exit settings retained. Source assumeTrue(false) is retained after every counterexample and trace assertion; only postcondition assertions are bypassed as in Java.
- [x] [tlc2/tool/liveness/simulation/LiveCheckSimulationTest2.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckSimulationTest2.java) — **Port complete**: `testSpec` (from `SuccessfulSimulationTestCase`).
  Go translation: [tlc_simulation_success_java_test.go](../tlc_simulation_success_java_test.go); whole original method, constructor and inherited runner/exit settings retained.
- [x] [tlc2/tool/liveness/simulation/LiveCheckSimulationTest2a.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckSimulationTest2a.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_simulation_counterexamples_java_test.go](../tlc_simulation_counterexamples_java_test.go). Whole original method, constructor and inherited runner/exit settings retained.
- [x] [tlc2/tool/liveness/simulation/SimulationTest2.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/SimulationTest2.java) — **Port complete**: `testSpec` (from `SuccessfulSimulationTestCase`).
  Go translation: [tlc_simulation_success_java_test.go](../tlc_simulation_success_java_test.go); whole original method, constructor and inherited runner/exit settings retained.
- [x] [tlc2/tool/liveness/simulation/SimulationTest2PostCondition.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/SimulationTest2PostCondition.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_simulation_success_java_test.go](../tlc_simulation_success_java_test.go); whole original method, constructor and inherited runner/exit settings retained.
- [x] [tlc2/tool/liveness/simulation/SimulationTest2a.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/SimulationTest2a.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_simulation_counterexamples_java_test.go](../tlc_simulation_counterexamples_java_test.go). Whole original method, constructor and inherited runner/exit settings retained.
- [x] [tlc2/tool/liveness/simulation/SimulationTestAssumption.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/SimulationTestAssumption.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_simulation_success_java_test.go](../tlc_simulation_success_java_test.go); whole original method, constructor and inherited runner/exit settings retained.
- [x] [tlc2/tool/liveness/simulation/StutteringTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/StutteringTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_simulation_counterexamples_java_test.go](../tlc_simulation_counterexamples_java_test.go). Whole original method, constructor and inherited runner/exit settings retained.
- [x] [tlc2/tool/simulation/Github1191Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/Github1191Test.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_simulation_models_java_test.go](../tlc_simulation_models_java_test.go). Whole original `testSpec`, constructor, inherited ModelCheckerTestCase flags and exit assertion, and byte-identical inputs. Both original finished/no-GENERAL assertions and SUCCESS constructor retained.
- [x] [tlc2/tool/simulation/Github1191aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/Github1191aTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_simulation_models_java_test.go](../tlc_simulation_models_java_test.go). Whole original `testSpec`, constructor, inherited ModelCheckerTestCase flags and exit assertion, and byte-identical inputs. Both original finished/no-GENERAL assertions, zero-uncovered check and VIOLATION_LIVENESS constructor retained.
- [x] [tlc2/tool/simulation/Github602Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/Github602Test.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_simulation_models_java_test.go](../tlc_simulation_models_java_test.go). Whole original `testSpec`, constructor, inherited ModelCheckerTestCase flags and exit assertion, and byte-identical inputs. Both original finished/no-GENERAL assertions and exact first progress parameters 156/1/100/0/0 retained.
- [x] [tlc2/tool/simulation/NQSpecTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/NQSpecTest.java) — **Port complete**: `testSpec`.
  Go translation: [tlc_simulation_models_java_test.go](../tlc_simulation_models_java_test.go). Whole original `testSpec`, constructor, inherited ModelCheckerTestCase flags and exit assertion, and byte-identical inputs. Both original finished/no-GENERAL assertions, inherited debugger setting and full 100-trace run retained. Production replaces the blanket recursive-body prime rejection with Java’s formal-argument maximum-level rule.
- [x] [tlc2/tool/simulation/SimulationWorkerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/SimulationWorkerTest.java) — **Port complete**: `testSuccessfulRun`, `testInvariantViolation`, `testActionPropertyViolation`, `testInvariantBadEval`, `testActionPropertyBadEval`, `testUnderspecifiedNext`, `testDeadlock`, `testModelStateConstraint`, `testModelActionConstraint`, `testWorkerInterruption`, `testTraceDepthObeyed`, `testStateAndTraceGenerationCount`.
  Go translation: [tlc_simulation_worker_correctness_java_test.go](../tlc_simulation_worker_correctness_java_test.go). All twelve whole methods retain their seeds, constructors, 126 assertions, exact trace values/levels, error codes, queue emptiness, interruption/join/liveness and generation counters. All thirteen BasicMultiTrace vectors match Java bytes. Production replaces the bounded result channel with FIFO queue behavior and preserves InterruptedException termination/reporting without counting an interrupted trace as completed.
  Related Go checks: [tlc/simulator_test.go](simulator_test.go).
- [x] [tlc2/tool/simulation/SimulatorMultiThreadTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/SimulatorMultiThreadTest.java) — **Port complete**: `testSuccessfulSimulation` (from `SimulatorTest`), `testInvariantViolationInitialState` (from `SimulatorTest`), `testInvariantViolation` (from `SimulatorTest`), `testInvariantBadEvalInitState` (from `SimulatorTest`), `testInvariantBadEvalNonInitState` (from `SimulatorTest`), `testUnderspecifiedInit` (from `SimulatorTest`), `testInvariantViolationContinue` (from `SimulatorTest`), `testDontContinueOnRuntimeSpecError` (from `SimulatorTest`), `testLivenessViolation` (from `SimulatorTest`), `testLivenessViolationIgnoresContinue` (from `SimulatorTest`).
  Go translation: [tlc_simulator_correctness_java_test.go](../tlc_simulator_correctness_java_test.go). All ten inherited whole methods and 22 assertions run with the original numWorkers() = 4, alongside the base class’s one-worker contexts; no worker-count substitution.
  Related Go checks: [tlc/simulator_test.go](simulator_test.go).
- [x] [tlc2/tool/simulation/SimulatorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/SimulatorTest.java) — **Port complete**: `testSuccessfulSimulation`, `testInvariantViolationInitialState`, `testInvariantViolation`, `testInvariantBadEvalInitState`, `testInvariantBadEvalNonInitState`, `testUnderspecifiedInit`, `testInvariantViolationContinue`, `testDontContinueOnRuntimeSpecError`, `testLivenessViolation`, `testLivenessViolationIgnoresContinue`.
  Go translation: [tlc_simulator_correctness_java_test.go](../tlc_simulator_correctness_java_test.go). All ten whole original methods retain 22 assertions, seed zero, FP64 setup, trace limits and continuation settings. Initial-state evaluation exceptions now return their reported diagnostic codes as in Java, rather than escaping after printing.
  Related Go checks: [tlc/simulator_test.go](simulator_test.go).

### Coverage

All original coverage model tests and OpApplNodeWrapper reporting methods have explicit translations.

No remaining non-`@Ignore` original methods identified in this topic. See the mapping appendix for the translations.

### Debugger and scoped identifiers

All 16 concrete debug/ test classes have explicit translations; this does not establish completion of DAP transport or debugger features outside those methods.

No remaining non-`@Ignore` original methods identified in this topic. See the mapping appendix for the translations.

### Checkpoint and recovery models

Original checkpoint-on-violation and time-bound models; generated recheck variants are listed with TTrace tests.

- [x] [tlc2/tool/checkpoint/CheckpointOnViolationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/checkpoint/CheckpointOnViolationTest.java) — **Port complete**: `testSpec`. Translation: [tlc_checkpoint_models_java_test.go](../tlc_checkpoint_models_java_test.go). Original inherited exit, coverage, event/trace assertions and checkpoint interval retained; time-bound test preserves five seconds and unchanged model/configuration.
- [x] [tlc2/tool/checkpoint/CheckpointWhenTimeBoundTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/checkpoint/CheckpointWhenTimeBoundTest.java) — **Port complete**: `testSpec`. Translation: [tlc_checkpoint_models_java_test.go](../tlc_checkpoint_models_java_test.go). Original inherited exit, coverage, event/trace assertions and checkpoint interval retained; time-bound test preserves five seconds and unchanged model/configuration.

### Distributed TLC

Current inventory: **37 port complete and four Reconcile**. Ordinary test entries
now preserve the upstream unconditional OffHeapDiskFPSet setup assumption as
`t.Skip`, before any roles start. Assertion bodies remain unchanged in shared
helpers. Separate `TestDiagnosticJava...` entries behind
`tlago_disabled_distributed_tests` deliberately bypass that assumption, retaining
Ant's off-heap/512 KiB profile, CPU-derived workers, fixtures and assertions.
Three source-profile runs passed before the correction. Local EWD840 exposed
Java's reuse of a closed flusher when final-check partitions become too small.
The user-authorized Go fix selects sequential preparation; the unchanged local
EWD840 diagnostic now passes in 74.80 seconds. See
[JAVA_BUG_FOUND.md](../JAVA_BUG_FOUND.md). Source skips and diagnostic passes do
not resolve harness reconciliation. No original-method credit is added.

Native worker resource-failure checks cover local/TCP calls, returned/panicked
memory exhaustion and executor rejection. Memory exhaustion requeues both
assigned states in order and reduces the block limit; rejection remains
non-recoverable. Cause details, URI diagnostics, computing cleanup, counters and
worker liveness are retained without a Java remote-exception envelope. Upstream
has no direct methods for these cases; no original-method credit is added.

Native worker-loss checks also cover absent and empty assigned blocks on all
four queue backends. Cleanup substitutes an explicit empty batch before queue
entry, preserves suspension, wakes consumers and decrements the worker count
once. Timer FIFO checks use the production disk queue without changing their
work identity/order assertions. No original method covers the absent-block case;
the original-method inventory is unchanged.

Full N=7 native process checks also kill a fingerprint host after a strict batch
insertion prefix. Mem/LSB/MSB rows verify each prefix/suffix membership bit before
host death, retain the source callable reassignment and finish with 229,884
slot-counted distinct states and an empty queue. Disk rows flush actual child
files before the kill. No direct original method covers partial batch loss;
these checks add no original-method completion credit.

Native endpoint-removal checks cover all five worker operations and repeated
direct exit over local/TCP boundaries, failure-payload round trips and unchanged
completion counts. Shutdown and keepalive continue past an already removed
worker with their distinct logging behavior. TCP checks distinguish endpoint
removal from connection closure. These supplemental cases add no original
method credit; broader distributed parity remains separately pending.

External GDB verification now stops the existing concrete checkpoint fixture
at the caller instruction after trace commit succeeds and before intern commit
is invoked. Process death preserves new queue/trace and old intern/fingerprint
generations; fresh recovery confirms both generations and leaves pending files
unpromoted. This is a short local manual receipt, separate from the earlier
syscall stop inside intern commit and from full-model/remote recovery. It adds
no original-method credit or production hook.

An opt-in external GDB check now covers the same caller boundary in the full
N=7 model with two nested LSB fingerprint hosts. A completed baseline precedes
the interrupted advancing checkpoint. Disassembly confirms trace commit returned
before intern commit was called; files retain new queue/trace, old intern plus
its pending file, and new remote fingerprints committed earlier by the source
remote path. Fresh roles restore exact partition membership/frontier and finish
at 114,942 states with an empty queue and no GENERAL. This closes that specific
full-model/remote verification gap without original-method credit. See
`TLC_ARCH.md` for the optimized `go test -c` binary and Linux/GDB invocation.

Native coordinator catalog checks cover local/TCP missing lookup/unbind, failure
payloads, retained retry delays, duplicate creation, shutdown-hook binding guards
and repeated coordinator removal. A lazy-reference check distinguishes an absent
catalog from a missing binding and observes later publication. These cases add
no original-method credit; broader distributed parity remains pending.

Native fingerprint transport checks preserve network causes through
`errors.Is`/`errors.As` for failed dial and closed-client operations. A completed
batch insertion followed by socket closure retains storage mutation, reports
failure and verifies no replay/redial. Original manager failover/checkpoint tests
remain unchanged. Standalone fingerprint registration fixtures now supply the
distributed coordinator's registration state; base-coordinator rejection remains
intact. No original-method credit is added for these native checks.

Distributed failure fixtures now use native Go operation traits rather than
fabricated Java remote exceptions. Existing retry/loss, shutdown, timer, finalizer
and bootstrap assertions remain intact. The sender-stack debug assertion exposed
missing Go frame capture; endpoint creation now captures its originating frames,
and actual local/TCP worker cases require the computation stack. Original Java
manager and smart-proxy test assertions are unchanged; no method credit is added.

Supplemental trace-evaluation checks preserve returned reconstruction/alias
failure identity, partial output and coordinator ordinary/fatal catch boundaries.
Original Alias safety and distributed initializer/model checks remain green;
noninitial printing branches `3`, `4` and `5` also preserve nil-reconstruction
diagnostics and fatal exit in child processes. Fingerprint-sequence recovery
branch `2` also preserves fatal exit and signed fingerprint diagnostics through
real tool lookups. Supplemental sequence checks cover normal-only random
restoration, missing initial behavior and returned info metadata. Tool lookup
overloads preserve source missing-match distinctions. Five native printing
metadata cases cover branch-specific UID/worker writes before alias evaluation.
Missing initial-transition checks preserve partial output and ordinary coordinator
catch/completion. Native disk-traversal checks cover four partial-read failures,
included/excluded successful restoration and native descriptor failure during
restoration. These verify normal-only cursor restoration and failure propagation.
Concurrent reconstruction checks cover anchor-only/invalid-empty inputs, normal
and supplied initial metadata, failure/random lifetime, missing initial dereference
and actual missing-successor exit. Six disk-backed public-entry checks preserve
ordinary/fatal reconstruction error identity and stop before behavior/state
printing. Required-worker checks preserve twelve public-entry index/null
failures, predecessor lookup failures and initial/equal-state short paths. Native
worker records verify full and requested predecessor ranges, while reconstruction
checks verify the trace monitor is held and released on success/failure. Native
checkpoint/level checks require each worker in order, preserve earlier mutations
and suppress later worker/marker operations after a missing slot. Level checks
retain the source worker-only maximum and minimum of one. Native enumerator
checks preserve exhausted-index and closed-cursor failures, first-failure close
ordering, required worker/writer/file access, unflushed reader construction,
neighbor-only advancement, genuine zero fingerprints and selector-only reset.
Native tool-ownership checks preserve empty/provided-initial short paths,
lookup-time failures and random-generator lifetime in both trace implementations.
Printing/disk checks omit tool access only on source short paths and preserve
missing alias results rather than fabricate output. No original-method credit
is added.

Supplemental partial trace-write checks preserve the attempted-record pointer,
partial bytes and unchanged state/record publication for initial, distributed
successor and shared single-process successor writers. Six additional native
worker cases preserve attempted pointer, partial predecessor/worker/fingerprint
bytes and earlier depth updates before write failure. Successful worker/mirror
cases retain generated actions and the source predecessor metadata policy.
Eight depth-limit cases preserve signed 32-bit maximum selection, completed
record/UID/worker updates, extended predecessor assignment and failure-time
counter/mirror behavior. Worker recovery checks cover publication before
missing-owner, closed-owner and seek failures, preservation of the existing
owner, and every truncated checkpoint length from zero through fifteen bytes.
Seven worker checkpoint creation cases cover required owner access, flush-before-
temporary ordering, stale creation errors, retained owners/pointers, exact metadata
bytes and lock release. Worker commit joins the existing missing-temporary,
blocked-delete and successful-promotion matrix, including native I/O payloads and
file mutation ordering.
Five worker read-owner cases retain original owner/last-pointer identity, stale
error handling, closed-owner mark-before-seek order and high-bit fingerprints.
Every partial length from zero through twelve bytes retains its consumed cursor;
the complete thirteen-byte record restores the saved writer position. No original
method directly covers these boundaries; completion counts remain unchanged.

Eight native null-state write cases retain predecessor-before-owner ordering,
target dereference after completed initial/successor records, closed-owner error
precedence, exact bytes and earlier depth/last-pointer updates. Failures suppress
state/count/mirror publication and release the worker lock. No original method
directly covers these boundaries; completion counts remain unchanged.

Eight native writer-owner cases cover initial/successor writes with missing,
memory-only, closed and healthy owners with a saved creation error. Writes cannot
open or replace the owner or publish state/count/mirror data after owner failure.
Successor depth updates precede owner access. Native fixtures initialize owners
during setup; original assertions remain unchanged. No original method directly
covers these boundaries; completion counts remain unchanged.

Six native predecessor-assignment cases cover ordinary/extended metadata with
mutable, polymorphic and typed nil inputs. Extended states clear the predecessor
before missing-predecessor failure; level and unrelated metadata stay unchanged.
No original method directly covers these boundaries; completion counts remain
unchanged. The seven original simulation trace methods retain their assertions.

Six native trace-registration cases cover missing trace/worker references,
negative/out-of-range indices, an empty owner list and valid slot replacement.
Invalid registration leaves the fixed owner list unchanged. No original method
directly covers this boundary; completion counts remain unchanged.

Eight native source-disk-trace checkpoint cases cover begin/recovery with
missing/closed owners and saved creation errors. Checks retain owner identity,
flush-before-temporary-creation and read-before-seek mutation ordering, exact
metadata bytes and lock release. No original method directly covers this
boundary; completion counts remain unchanged.

Eighteen native disk-writer cases cover initial/record/successor writes with
missing/closed owners and saved errors, missing trace receivers and predecessor-
before-owner failure order. Invalid writes retain owner identity and suppress
record/state publication. No original method directly covers these boundaries;
completion counts remain unchanged. Original partial-write assertions are intact.

Twenty-four native read-owner cases cover predecessor/fingerprint/depth reads,
public depth, reporting and enumeration with missing/closed owners and saved
creation errors. One additional public-depth case retains the consumed cursor on
a truncated predecessor read. Missing owners cannot yield zero or memory data;
disk errors propagate. No original method directly covers these boundaries;
completion counts remain unchanged.

Eleven native enumerator-owner cases cover position/read/close with a missing
reader or enumerator, and reset with missing enumerator/trace/writer/reader.
Reset publishes length before old-cursor access; an explicit position opens a
replacement even without an old reader. No original method directly covers these
boundaries; completion counts remain unchanged.

Nine native trace-print entry cases cover ordinary/concurrent/shared printers
with missing current state, trace or both. Failures precede behavior-header and
state output. No original method directly covers these boundaries; completion
counts remain unchanged.

Six native worker-construction cases cover the passed tool/checker directory,
missing checker/tool/trace, rejected trace index and file-open failure. Owner
creation precedes trace registration; trace acceptance precedes checker-list
publication. Rejected native registration releases its file handle. No original
method directly covers these boundaries; completion counts remain unchanged.

Six native post-construction mutation cases cover enabled/disabled deadlock,
liveness and debugger-mode settings. Deadlock diagnostics, set allocation and
stuttering graph insertion use captured settings. No original method directly
covers this boundary; completion counts remain unchanged.

Five native worker queue cases cover replacement or removal of the checker's
queue, dequeue failure and successor publication with a retained or missing
checker worker slot. Worker-owned operations retain the captured queue and
executing worker. No original method directly covers these boundaries;
completion counts remain unchanged.

Eight native successor-owner cases cover replacement/removal of the fingerprint
set or state writer, excluded state/action constraint reasons and unsatisfied
writer failure propagation. Captured owners retain insertion and transition
output; liveness stuttering keeps the source's current-checker writer access.
No original method directly covers these boundaries; counts remain unchanged.

Four native successor-evaluation cases cover eligible/excluded states after
replacement/removal of the checker's tool. Ordered validity, constraints,
exclusion reasons, invariant and implied-action evaluation retain the executing
worker's tool. No original method directly covers these boundaries; original
completion counts remain unchanged.

Three native worker error cases cover deadlock, invariant and implied-action
failures after checker-tool replacement. Pairwise alias and counterexample
postcondition evaluation use the worker tool; checker reporting remains under
the same lock. Fixtures supply the initial trace and reconstructable transitions.
No original method directly covers tool replacement; counts remain unchanged.

Four native postcondition alias checks cover an escaping native or fatal failure
at the first or second alias. Failure identity is retained, later aliases stop,
and the postcondition is not evaluated. Tool-level displayable alias errors keep
their source handling. No original method directly covers this boundary;
original completion counts remain unchanged.

Four native postcondition reconstruction checks cover escaping native/fatal
errors from initial and transition state recovery. Original failure identity is
retained; alias and postcondition evaluation cannot proceed after failed recovery.
No original method directly covers these returned-error boundaries; completion
counts remain unchanged.

Three native final-state recovery checks distinguish initial fingerprint
selection from noninitial state-equality reconstruction and verify returned
metadata and transition action. Worker ID and trace pointer are not copied from
the target onto generated states. No original method directly covers these
helper boundaries; completion counts remain unchanged.

Six native missing-result cases cover initial/transition recovery, first/last
nil aliases and first/last missing counterexample entries. No raw-state or
empty-record substitute is created; failures retain source alias/constructor
ordering. Existing native alias/deadlock fixtures now supply valid reconstruction
setup without changing assertions. No original method directly covers these
boundaries; completion counts remain unchanged.

Eight native postcondition context cases cover missing current state/tool,
retained initial prefix, missing trace owners and empty/nil-last prefixes.
Failure precedes postcondition evaluation; no fresh trace or initial-state
recovery is substituted for a missing owner or last prefix state. No original
method directly covers these boundaries; completion counts remain unchanged.

Nine native liveness-tail checks cover missing current/set/checker/writer/tool/
liveness owners, current checker-tool selection, earlier writer failure and
call-stack replay ownership. Stuttering insertion and output retain source
ordering; no replacement set or silent tail skip is allowed. No original method
directly covers these owner boundaries; completion counts remain unchanged.

Twelve native filename/checkpoint checks cover empty path components, preserved
dot components, empty-directory configuration, required opening context and
cached empty checkpoint filenames. Begin flushes before temporary creation,
commit deletes before promotion, and recovery publishes metadata before owner
access. Literal suffix I/O is isolated in temporary working directories; no
filesystem-root filename is opened. No original method directly covers these
boundaries; completion counts remain unchanged.

Supplemental initial-publication checks preserve metadata, fingerprint/trace/
queue/property order, missing-owner and I/O failure context, suppression of
later elements and seen/excluded-state owner access. No original method directly
covers these boundaries; original initializer methods remain green.

Supplemental coordinator-label checks preserve the statistics prefix, counter
padding, ASCII URI rendering and unchanged endpoint metadata. No original
method directly covers these labels; completion counts remain unchanged.

Supplemental role-usage checks preserve invalid-argument early returns and
banner/error/usage order while naming native Go commands. Server usage checks
cover both ToolIO output modes. These add no original-method completion credit.

Supplemental coordinator-mode checks preserve base-server registration rejection
before manager access, locally and over native TCP. Distributed registration and
original manager tests remain green. No original method directly covers this
boundary; completion counts remain unchanged.

Supplemental worker-loss checks cover unconditional suspended-consumer wakeup
for all four queues, cleanup/requeue-before-wakeup, wake-before-decrement and
duplicate loss reports. Focused concurrency checks pass with race detection;
no upstream method directly covers this boundary, so counts remain unchanged.

Supplemental batch-statistics checks preserve required timer/coordinator owners,
received-count/timestamp/delta ordering and worker-loss requeueing after an
absent timer. No original method covers these boundaries; counts are unchanged.

Supplemental fresh-process selector checks preserve startup setting capture,
flag precedence, boolean parsing, deferred static-size initialization and failed
constructor timing. Upstream has no direct methods for these boundaries;
original-method completion counts remain unchanged.

Supplemental retry-queue checks preserve required requeueing before limit
updates, partial mutation and failure identity for I/O/runtime/fatal failures,
assigned work and the inner catch boundary. These add no original-method credit.

Supplemental selector-ownership checks cover constructor retention, missing
selector error handling/finally cleanup and requeue-before-selector-failure
ordering. No upstream test directly covers these boundaries, so original-method
completion counts and the missing distributed harness entries are unchanged.

Supplemental application traversal checks preserve evaluator-driven replacement,
shrink and growth of generation/invariant/implied-init/implied-action arrays.
Violation names/context and generation failure identity are retained. No original
method covers these mutations; original-method completion counts are unchanged.

Supplemental fingerprint lifecycle checks preserve explicit null failures at
begin/commit/recover, including live replacement between phases. Distributed
close reports and continues; local lifecycle calls propagate the failure.
Existing original manager and short TCP checkpoint checks pass. No original
method covers these boundaries, so completion counts are unchanged.

Supplemental coordinator handler checks preserve missing-trace diagnostics,
required queue shutdown, skipped notification after queue failure and thread
finally cleanup. Block selection fails for a missing queue. Existing selector,
finalizer, worker-loss and short native TCP checks pass; no original-method
completion credit is added.

Supplemental missing-manager checks preserve constructor ownership, generation/
statistics/failure ordering, predecessor context and computing cleanup locally
and through native TCP. Empty-manager behavior remains distinct. No original
method covers this boundary; completion counts remain unchanged.

Supplemental application successor-array checks preserve container isolation,
shared state objects, larger-vector order and validation-time replacement/shrink/
growth behavior. Existing worker/TCP checks pass; no original-method credit is
added because upstream has no direct successor-array methods.

Supplemental shutdown traversal checks preserve late startup publication,
earlier exits/latch mutations before a nil runnable failure, lifecycle lock
release and source unpublished-worker skips. Three short checks pass with race
instrumentation; existing short TCP lifecycle checks pass normally. No original
method covers this boundary, so completion counts are unchanged.

Supplemental gob/TCP manager checks preserve nil endpoints separately from empty
registration slots, shared/distinct wrapper identity, metadata and operation-time
failover. Contradictory null references fail before resolution. Existing original
manager methods pass; no original-method credit is added for transfer checks.

Supplemental local/TCP worker-registration checks preserve required queue wakeup
before worker contact, null-worker ordering and monitor release after wakeup
failure. Existing registration and original smart-proxy checks pass. No upstream
method covers these failure boundaries; original-method counts are unchanged.

Supplemental fingerprint-check boundaries preserve null-endpoint failed task
completions, healthy results and null-registration submission failures. Original
manager methods and native TCP I/O checks pass; no original-method credit is added.

Supplemental live-registration checks cover checkpoint phase replacement,
post-first-call tail selection, current failure hostnames, fixed initial counts,
shutdown's captured-next ordering and wrapper identity. Existing original dynamic
manager methods and short native TCP checkpoint checks pass. No upstream method
covers these registration changes; original-method completion counts are unchanged.

Full native EWD840 recovery now passes between the two nested FP commits.
Parent checks require first-child promotion and changed contents, a byte-identical
old second-child checkpoint and its retained temporary file. Fresh recovery
preserves all final model assertions. No disabled Java harness credit is added;
isolated trace-commit interruption and broader failures remain pending.

Full native EWD840 recovery now passes after intern-table commit and before
FP commit, retaining committed queue/trace/intern files and both FP temporaries.
It preserves all final model assertions and adds no disabled Java harness credit.
Nested FP-commit interruption is also verified; isolated trace interruption
remains pending.

Full native EWD840 recovery now passes after interruption immediately following
queue commit: exact new queue/old trace metadata and full trace reconstruction
precede the unchanged final model assertions. This adds no disabled Java harness
credit; interruption after intern commit is also verified. Isolated trace and
nested FP commit coverage is now verified; isolated trace interruption remains
pending.

Supplemental disk checkpoint checks retain lock/flag ownership after I/O or
runtime flush/copy failures, release on success and native manager continuation
to healthy partitions without reassignment. No original-method credit is added.

Supplemental memory queue recovery checks cover untouched storage/cursor,
partial state publication, fixed-capacity overflow and coordinator stop ordering.
Original short memory queue methods pass; no original-method credit is added.

Supplemental checkpoint commit checks cover queue/trace I/O categories, retained
file/pool mutation and coordinator stop ordering. Original short queue methods
remain green; these checks add no original-method completion credit.

Supplemental memory checkpoint checks cover missing parents for all three
implementations locally and through native TCP, plus zero/negative backing-set
constructor sizes. These checks add no original Java method completion credit.

Supplemental MemFPSet1 checks retain source backing-set recovery counts, growth,
zero membership and field-by-field mutation, including complete/truncated files
through native TCP. No upstream SetOfLong test exists; no original-method credit
is added for these checks.

Supplemental disk file recovery checks retain sequential reader replacement,
partial close/open failure state and completed file/index writes. Existing
original buffered-file methods pass; these checks add no original-method credit.

Supplemental duplicate recovery checks now require the runtime assertion category
for memory sets and parent MultiFPSet, including native failure payloads and
real TCP manager recovery. These add no original Java method completion credit.

Native nested recovery checks now require selected-child RecoverFP dispatch,
source disk/off-heap duplicate failure or warning-and-continue behavior, exact
partition routing and retained earlier insertions. Child I/O failure must escape
after real queue recovery and before coordinator publication, without replacing
its cause. No enabled original method covers this complete dispatch boundary;
supplemental checks add no original-method credit.

Native queue recovery failure checks cover enqueue/dequeue records truncated
before UID or level. They require untouched slots, publication of the new empty
state before reading, completed header-field mutations and failure before FP
recovery/publication after successful trace recovery. Original short queue and
value-stream translations remain green. No enabled original method covers these
failure boundaries; supplemental checks add no original-method credit.

Native coordinator recovery checks now cover missing trace checkpoint metadata,
all 0–15 byte truncations and a complete negative cursor. They require I/O
propagation before queue recovery/publication/initialization and the source
saved-pointer update before failed seek. Closed-owner recovery must retain the
same read-before-seek order without reopening the handle. No enabled original
method covers these boundaries; supplemental checks add no original-method credit.

Native trace recovery now propagates enumerator creation/cursor/read/reset I/O
failures. Short checks cover truncated predecessor/fingerprint fields, missing
files, closed handles, valid zero fingerprints and reset after trace growth.
MultiFPSet and direct disk/off-heap reconstruction reject failed records without
inserting zero. Existing short original MultiFPSet translations and full native
pre-commit interruption recovery pass. No enabled original method covers these
error boundaries; supplemental checks add no original-method completion credit.

Native block-selector checks cover the source null-server assertion, zero-worker
arithmetic, saturated numeric conversions, NaN/infinite overhead, proportional
and static limit no-ops, limiting/statistical limit updates, queue request bounds
and actual-dequeue averages. Existing original smart-proxy methods and native
TCP coordinator retry/loss pass. Upstream has no selector test methods; these
supplemental checks add no original-method completion credit.

A short native selector concurrency check reproduces and verifies removal of
races in average publication and transfer-limit updates/reads. Atomic average
operations retain the source lossy calculation and signed overflow; sequential
zero/rounding/overflow and fixed-average checks pass. No original-method credit.

A native coordinator keepalive check runs the unchanged ten-second initial
delay and verifies uncaught runtime/fatal failures terminate only the timer,
retain diagnostics and leave coordinator/assigned work unchanged. Timer owners
are joined. TCP worker-loss/keepalive and original smart-proxy checks pass.
Upstream has no enabled method for this boundary; no original-method credit.

Native fingerprint-check callable failure checks retain the source I/O-only
catch, GENERAL diagnostic, MaxInt64/false sentinels, unchanged partition ownership
and unchecked execution-failure path. Both returned and panicked I/O failures
pass locally and over TCP; host availability remains intact. Original manager
methods pass. No upstream methods cover this catch; no completion credit added.

Native NextStateResult getter checks retain null receiver/partition failures,
array references, null/empty distinctions and signed counter overflow. Payload,
original smart-proxy and TCP worker result/retry checks pass. No original getter
test methods exist; supplemental checks add no original-method credit.
Coordinator cache-ratio classification and canonical numeric output now match
all 271 Java reference rows, including NaN/infinities, signed zero, decimal ties
and deterministic bit patterns. Locale-specific separators/digits pass 1,860
reference rows (1,068 locale keys and 792 numbering variants) and fresh-process
configuration/initialization checks. POSIX disables worker-statistics grouping
without changing MP's explicit grouping. No original test method exists;
supplemental coverage adds no credit.

Native value-array graph checks cover state, tuple, record, function, operator argument
row, tuple-product and record-set sharing, cycles, nil/empty arrays, malformed
references and isolated receiver mutations across worker request/result TCP.
State checks retain backing arrays shared between distinct states and composite
values, recursive state/value arrays, separate equal-content arrays, nil/empty
storage and malformed-reference rejection.
No enabled original Java method directly covers this boundary; no original-method
credit is added. Native ValueVec checks also cover vector identity, active count,
full capacity, shared backing arrays, unused cyclic slots and malformed references
across gob and worker TCP. Native record/record-set name-array checks now retain
array sharing, separate equal-content arrays, shared name objects, nil/empty
arrays and receiver isolation, with malformed-reference rejection. These cases
add no original-method completion credit.

The full native MC06 process matrix now includes `partitioned_fingerprints`
with two standalone FP processes and private temporary storage. Their two
distinct nonempty initial partitions must total 16,384 fingerprints before
worker launch; the unchanged N=7 model retains FINISHED, 114,942 distinct/0 queued,
two registrations and no GENERAL across all roles. This adds no disabled
upstream harness completion credit.

FP reference registration now preserves source capacity/latch behavior without
an early network probe. Native checks accept a stopped-store reference, reject
an extra unreachable reference with the exact capacity category/detail and
reject incomplete references without consuming slots. Existing original dynamic
manager tests pass; these supplemental checks add no original-method credit.

Fingerprint snapshots now defer connection setup until an operation, matching
source endpoint-reference behavior. Native checks receive a snapshot with a
stopped store, require operation-time partition failover and isolated coordinator
registrations, verify concurrent first calls and close both used/unused owned
references. No enabled original method directly covers this native boundary;
supplemental checks add no original-method completion credit.

Native worker graph codec failures now enter the remote I/O worker-loss catch,
with their causes preserved. Short TCP/coordinator checks require exact assigned
state requeue, one deregistration, no GENERAL and continued unfinished work.
Unevaluated lazy values retain the source runtime category and exact diagnostic.
No enabled original method directly covers these boundaries; supplemental
checks add no original-method completion credit.

The native worker resolver API is `DistributedFilenameToStreamResolver`.
Coordinator TCP file/cache checks retain basename keys, flag-independent fetches,
binary/empty data, deletion/refetch and independent temporary directories.
No direct upstream resolver test exists; this adds no original-method credit.

Malformed native coordinator locations now retain a distinct Go error category.
Short discovery/keepalive checks require the source log-and-continue handling
without exit, cancellation or latch release, across existing validation cases.
These supplemental checks add no original-method completion credit.

Model-value scalar transfer now retains `int8`, `int16` and `float32`, with
checked integer/bit ranges. Gob and worker TCP cases cover both float widths,
signed zero, extrema, subnormals, infinities and NaN. Integer IEEE payload bits
fix negative-zero loss in the former float64 field. No enabled upstream method
directly tests this boundary; these checks add no original-method credit.

Primitive-array attachment transfer now covers nine native slice types: bool,
int/int8/int16/int32/int64, uint16 and float32/float64. Gob and worker TCP checks
retain exact element types/bits, shared references through mixed containers,
separate equal arrays, typed nil/empty slices and isolated receiver mutation.
Malformed kinds, references and narrow representations are rejected. The
existing complete ModelValueTest translation remains green; these supplemental
transport checks add no original-method credit.

String-array attachments now have native `[]string` graph support. Focused gob
and worker TCP checks preserve shared storage through mixed containers, separate
equal arrays, nil/empty distinctions, Unicode/NUL/arbitrary native string bytes
and receiver mutation isolation. Invalid array references fail explicitly.
The existing complete ModelValueTest translation remains green; these native
transport checks add no original-method completion credit.

Accepted native checkpoint calls now have connection-loss coverage for begin,
commit and recovery. The source I/O catch must warn once, continue to the healthy
store and leave partition registrations/availability unchanged. Queue/trace
checkpoints commit and the queue resumes while the disconnected handler is
still gated; accepted storage work finishes after release. These short native
checks add no original-method credit and do not prove process-crash recovery
or checkpoint atomicity.

Partitioned assigned-block checkpoint/recovery now uses two native TCP stores
with separate metadata directories. Reopened tables must recover the initial
and successor fingerprints into different original partitions while retaining
the exact queue/trace identities. Existing local/single-store checks remain.
This native boundary test adds no original-method completion credit.
Fresh-process remote-FP CLI recovery is reconciled as a pinned-source limitation:
its dynamic manager is empty when recovery runs before publication/registration.
Adding that capability would change the original algorithm; registered-endpoint
remote recovery is covered separately.

Native model-value byte data now preserves shared buffer identity across states
and worker-result partitions, with separate equal-content buffers, isolated
receiver ownership and nil/empty distinctions. Gob and worker TCP checks retain
mutation-visible sharing; malformed byte references are rejected. No enabled
Java test directly covers this graph boundary, so no original-method credit is
added. Opaque custom model data remains separately pending.

Coordinator system-failure diagnostics retain debug-enabled throwable stacks
and original catch/finally cleanup order. Eight short returned/panicked cases
cover stack-overflow/out-of-memory with debug off/on. They also require completed
ToolIO stack lines, exposing and fixing a shared printer shortcut. Original Java
MP/WarningControl and native keepalive checks remain green; no enabled direct
Java coordinator-main test exists, so no original-method credit is added.

Four short native in-flight FP block checks now cover sequential/concurrent
put and contains reassignment after TCP host closure. They require original
partition ordering, survivor wrapper sharing, one warning and the exact backing
store effects. Transport closure does not cancel the old owned storage handler;
all test jobs/handlers are joined. No enabled direct Java test covers this native
boundary, so these cases add no original-method completion credit. Process crash
and recovery models remain pending.

Native worker keepalive checks now cover finished, unbound, disconnected and
status-failure coordinators, with debug stacks enabled/disabled. They retain
activity suppression, idle shutdown, completion latch, cancelled timer and
unavailable callback behavior. The production failure diagnostic now uses the
source throwable overload. No enabled direct Java test covers these boundaries;
the additional native checks add no original-method completion credit.

Worker/coordinator fatal endpoint errors now use the same native remote I/O
policy for returns and panics as fingerprint RPC. Ten short TCP cases retain
cause/suppressed sharing and actual Go diagnostics; successful subsequent calls
verify host availability and failed Exit retains worker publication. Existing
RPC and source manager checks remain green. These boundary cases add no
original-method completion credit.

Coordinator required-component checks now cover checkpoint, recovery and close
access order. Missing owners retain preceding file writes, recovery reads and
trace closure, without later resume, commits or metadata deletion. Failed queue
suspension still bypasses later owners. Focused checkpoint/TCP checks pass.
Upstream has no direct methods for missing coordinator components; these native
checks add no original-method completion credit or changes to Missing entries.

Coordinator management stop/suspend/resume controls now preserve source
synchronization, queue failure order and stop’s reporting-wait notification.
Short native checks cover six queue success/failure cases, nine missing-owner
cases and a joined real reporting wait. Upstream has no direct methods for
these controls; no original test statuses or completion credit change.

Native printable-state transfer now retains shared record/value/cache identity,
default display, extra fields and fingerprints over gob/TCP and worker exception
contexts. A real transfer check exposed and fixed ToState’s pointer comparison
where the source uses token equality. Existing original record/alias methods
remain green; no direct original printable-state transfer tests exist, so these
supplemental checks add no original-method status or completion credit.

Native state-cache transfer now retains map identity, separate equal-content
maps, nil/empty maps, null entries and shared/recursive values across state
roots and predecessors. Gob/TCP requests, results and exception contexts pass.
Invalid map/value references, duplicate keys and out-of-range keys fail. No
direct upstream cache-transfer test methods exist, so these supplemental checks
add no original-method status or completion credit.

Native predecessor graph transfer now passes gob/TCP request/result and worker
exception checks for shared parents, non-root ancestors, nulls, cycles, stored
levels and shared values. Invalid IDs and unsupported ancestor metadata fail
explicitly. Upstream has no direct predecessor-transfer test methods; these
supplemental checks do not change original-method status or credit. Remaining
evaluator metadata/custom-state coverage is tracked separately.

Distributed worker partition vectors now retain TLCStateVec’s unbounded
growth, capacity-ten default and backing-array indexing, independent of the
tool StateVec’s SetBound. Native result decoding restores the same collection
policy while transferring active entries only. Supplemental local/gob/TCP and
malformed-selection/publication checks pass; no direct upstream TLCStateVec
test methods exist, so no original method status or credit changes.

Distributed management queries now require the source-owned server and running
components. Generated counts cannot silently omit the FP manager, remaining
work cannot omit the queue and selector averages have no fabricated fallback.
Supplemental checks preserve explicit inactive/empty sentinels, signed counter
overflow and queued plus assigned work. No direct upstream query methods exist,
so no original test status or method credit changes.

Periodic/final coordinator progress now uses the existing message locale
formatter instead of fixed comma grouping. Supplemental native process checks
verify nine source locale reference rows, signed limits, rates, plain final
statistics and tool/success conditions. Original MP tests remain green. No
direct original coordinator-formatting method exists; no method credit is added.

Coordinator trace depth I/O failures now stop periodic/final reporting instead
of publishing stale depth. Native failure checks preserve partial cursor reads,
monotonic successful depth, management `-1` catches and final reporting's
shutdown/count mutation order. Existing original TLCGetLevel and its TTrace
methods pass; supplemental failure tests add no original-method credit.

Server-thread finalizer checks now cover returned/panicked remote cache failures,
uncaught runtime/fatal cache failures and fatal trace printing inside the catch.
Native goroutine ownership preserves source catch/finally order without killing
the process, recatching handler failures or executing later finally steps after
a cache failure. Focused RPC/codec/checkpoint and original smart-proxy checks pass.
No direct original finalizer methods exist; supplemental checks add no credit.

Coordinator timer-task checks now cover returned/panicked remote failures, false
liveness, unchecked local failures, activity suppression and native TCP fatal
status payloads. Remote failures enter worker-loss cleanup once, preserving
assigned work and deregistration-only diagnostics. Local unchecked failures
retain ownership. No direct original timer-task methods exist; no credit is added.

Final worker-exit and shutdown-hook checks now preserve their distinct catches
for local/native-payload failures and return/panic forms. Final completion removes
the exited thread even on failure; the hook retains registrations. Native worker
unavailability has a separate wire trait from the broader final-exit ignore
category. Twelve completion processes, 56 hook cases and actual unavailable TCP
workers pass. No direct original methods exist for these boundaries; no credit
is added.

SimpleCache raw ratio arithmetic now passes 271 source counter rows and native
TCP nonfinite/signed-zero/extreme checks. Missing cache owners retain source
failure behavior. Native worker cache replies carry IEEE bits to preserve
negative zero through gob. No original SimpleCache methods exist; no method credit is
added. Its worker exit-message compact formatter now passes 271 counter rows,
6,610 additional numeric rows, 1,860 locale rows and six fresh locale processes.
Native Go formatting preserves locale/special symbols and source integer
rounding; this is distinct from coordinator worker statistics. No original
method status changes, since upstream has no direct tests for this formatter.

Returned fatal local endpoint errors now escape exactly like fatal panics;
scalar/block/statistics, checkpoint/recovery and close checks verify no retry,
warning or availability mutation. Original Java manager translations and native
RPC checks remain green. These short Go boundary checks have no enabled direct
Java counterpart and add no original-method completion credit.

Fingerprint RPC failures now use the existing native Go failure payload, retaining
nullable messages, causes, suppressed-error sharing and actual Go diagnostics.
Fatal endpoint failures retain the remote I/O category needed by TLC failover.
Six short returned/panicked failure cases pass normally and under a focused race
check. They add no original-method completion credit and provide no Java RMI or
serialization compatibility.

Finite configured constant-operator payloads (`OpRcdValue`) now retain argument
rows, results, sharing, cycles and operator application through native gob and
worker TCP request/result checks. Upstream has no direct test of this transfer;
these short Go boundary checks add no original-method completion credit.
The combined `worker-fpserver` command also passes full unchanged MC06/N=7
exploration in a single worker/FP process with a separate coordinator. Both
processes finish normally with 114942 distinct states, zero queued states and
no GENERAL event. This adds native command-lifetime coverage, not completion
credit for the disabled original harness.
Coordinator publication failure checks now preserve source failure after FP
insertion: selected missing/null partitions and null visited vectors fail,
and trace write failures terminate the server thread without another dequeue.
These native checks have no enabled direct Java counterpart and add no original
method credit. Full unchanged coordinator-FP MC06 verification remains green.
The assigned-block checkpoint barrier is now covered by a short native TCP
test with real disk queue/trace and both local and native TCP FP storage. It checks suspension,
publication before checkpoint, worker continuation and storage reopening with
the exact recovered frontier/trace identities. This is additional boundary
coverage without original-method credit. A separate full native MC06 row now
verifies mid-run local MemFPSet recovery in fresh coordinator/worker processes
after abrupt post-commit producer exit. The native pre-commit interruption row
now requires source old-queue recovery and fingerprints from the full persisted
trace, then the original full model result. It adds no original-method credit.
Interruption immediately after queue commit and after intern commit is verified;
nested FP-commit interruption is also verified; isolated trace interruption
remains pending.
Fresh-process remote-FP CLI recovery
is a verified source startup limitation, as documented above.
Loss of the last worker is now exercised by the full native MC06
`all_workers_lost` row. It waits for cleanup to finish, verifies the available
coordinator is not done before replacement registration, then requires the
original final model counts and exact one-time loss/cache-warning behavior.
This adds native failure coverage, not original disabled-harness completion
credit. A short native fingerprint-server process check now covers abrupt kill,
manager reassignment, insertion on surviving storage and fresh-process recovery
of committed MemFPSet, LSBDiskFPSet and MSBDiskFPSet membership. It distinguishes
committed fingerprints from a later insertion in an uncommitted snapshot. Disk
rows also verify flushed live membership, byte-identical committed checkpoints
and no promotion of pending snapshots after process loss. It preserves the dead
client's shutdown cause and existing manager registrations. No original method
directly covers this native process boundary, so it adds no credit.
The full native MC06 fingerprint_server_loss row also kills the first of two
fingerprint hosts with real worker work assigned, resumes unchanged evaluation
and requires completion without GENERAL. Source size() sums aliased partition
slots after failover, so this separate failure row requires 229,884 reported
states; existing ordinary model rows keep their 114,942 assertions. This adds
native failure coverage, not completion credit for the disabled Java harness.
The native N=7 model now covers a complete mid-run checkpoint across two remote
hosts with Mem, LSB or MSB storage, followed by loss of all original roles and
recovery with empty replacement stores registered before source recovery. It verifies committed
partition membership, disk queue counts, no regenerated initialization and final
114,942-distinct/empty-queue completion without GENERAL. This uses the existing
library lifecycle; ordinary Java CLI startup still recovers before registration.
No production startup change or disabled-harness method credit is added. Other
failure phases and general network partitions remain open. Disk hosts retain the
source two-child MultiFPSet layout and restore fingerprint high bits from each
child's committed file when checking complete recovered membership.
Completed remote commit reply loss now also runs with two workers sharing one
application in both process generations, across Mem/LSB/MSB storage. Existing
exact warning, two retained registrations, committed snapshot bytes, recovered
membership and final model assertions remain intact; both replacement workers
must report actual work on distinct endpoints of the same listener. These native
combination checks add no original-method completion credit.
Named nested-store checkpoint traversal now has native concurrency/join and
failure-boundary checks, plus fresh-store Mem/LSB/MSB recovery checks preserving
both high-bit partitions. The source MultiFPSetTest has no corresponding named
checkpoint methods; this adds no original-method completion credit.
Intern-table recovery now locks before opening/reading its checkpoint, matching
the whole source synchronized method. Short Linux FIFO cases prove the blocked
header boundary and lock release on success/truncation. No direct original
InternTable recovery test exists, so these native checks add no method credit.
Native checkpoint symlink checks now cover trace, worker, both state queues,
intern table, all three memory fingerprint stores and five supporting
integer/object/byte-array/DFID stores. They preserve the source
existence/delete/rename sequence and target ownership on success and failure.
These 52 cases and the byte-array partial pool-deletion check add no
original-method completion credit; the existing original StateQueue and
MemIntQueue methods remain green.
Completed fingerprint recovery with a lost TCP reply now has native Mem/LSB/MSB
checks. They retain exact membership in both high-bit partitions, the source
warning/continuation behavior, registration identities and no automatic replay.
A fresh host reads the retained borrowed storage. No enabled original method
directly covers this transport boundary, so it adds no completion credit.
Explicit empty fingerprint checkpoint names now have native local/TCP checks
for Mem/1/2 and LSB/MSB, plus DFID and empty disk backing-name checks. The
literal snapshot stays separate from `fpset` and the initialized name; recovery
retains exact membership and MemFPSet1's source count increment. No original
method directly covers this filename boundary; native checks add no credit.
Six native empty-directory checks now cover memory/disk and DFID stores using
literal absolute paths inside temporary roots. A DFID missing-parent check
retains direct file-open failure and verifies no directory creation or membership
change. These source path contracts have no direct original methods and add no
completion credit. Six native LSB/MSB initializer cases now retain filename
assignments before negative-array failure, reader allocation before failed open,
source metadata across reinitialization and native old-reader retirement. No
original method directly covers these partial mutations; they add no credit.
Fingerprint completion diagnostics now have 20 local/TCP Mem/1/2/LSB/MSB exit
cases covering retained/removed files and exact host/message records. A native
failed-directory-removal case retains the source's ignored cleanup result and
completion. No enabled original method directly covers this exit diagnostic;
these checks add no original-method credit.
Nested fingerprint/invariant checks now have native concurrency/join, signed
minimum, empty-result and child-failure checks. TCP verifies that child I/O is
wrapped as an operation failure and reaches the manager's failed-task path.
The existing original MultiFPSet getFPSet method remains green; no original
method directly checks this concurrency/failure boundary, so no credit is added.
Nested size/statistics now have native joined concurrency, overflow and local/TCP
parent-counter checks. Original MultiFPSet getFPSet and manager nested-partition
methods remain green. No original method directly tests parent/child counter
isolation or the parallel size boundary; these native checks add no credit.
Native local/TCP nested thread-registration checks verify inherited `AddThread`
is a no-op while `IncWorkers` visits children. Heap disk children inherit the
same worker-registration no-op; direct disk `AddThread` separately adds one
reader and retains existing reader/pool ownership on open failure. Corrected an
earlier native assertion and Go implementation that incorrectly allocated heap
readers from `IncWorkers`. Related original nested methods remain green; no
direct original test exists, so no credit is added.
Native nested initialization checks cover failures from either child, joined
ownership, unchanged arguments and ignored replacement returns. Checked I/O
has one operation wrapper; unchecked/fatal errors retain identity without Java
ForkJoin copying. Original nested-store methods and actual startup rollback
checks remain green. No direct original method exists, so no credit is added.
Native LSB/MSB local/TCP close checks retain closed worker/pool reader identities,
ignore individual close I/O failures and reject subsequent actual disk lookup
without reopening storage. Corrected public Close's reuse of the array-clearing
rollback helper. Host draining, startup rollback, reader recovery and original
nested-store methods remain green. No direct original post-close endpoint method
exists, so no credit is added.

A short native Linux syscall check now kills a concrete checkpoint child at
intern deletion entry, after queue/trace promotion and before intern-file
mutation. File bytes, pending generations and fresh queue/trace/intern/FP
recovery are checked against a complete control. It requires strace and uses no
production hook or timing race. This covers a local syscall boundary inside the
intern commit method; exact between-method interruption and full-model/remote
recovery remain unproved. No direct enabled original method exists, so no credit
is added.
Native invariant-overload checks cover local/TCP memory, nested and LSB/MSB
stores. Memory/nested expected-count calls inherit the base true result; disk
calls enforce counts and nested no-argument checks visit children. Related
original methods remain green; no direct original overload test earns credit.
Native memory recovery now checks 32 local/TCP cases: base/packed storage with
seven partial-long lengths after complete records, plus duplicates before an
unvisited record. Source manager warning/continuation, exact runtime diagnostics,
prefix/prior membership and unchanged registrations are verified. No enabled
original method covers this boundary, so no completion credit is added.
Eight native coordinator startup cases extend this through actual trace/queue
recovery and local/TCP fingerprint endpoints. Runtime failures prevent
publication and later recovery; packed-memory I/O reports the warning, recovers
the healthy endpoint and emits actual recovery counts before publication.
Related original manager methods remain green; no new completion credit added.
Native memory queue path checks reproduce and fix invented parent creation and
lexical path cleaning. Begin/commit/recovery now retain source literal paths;
original state-queue methods and related active-block checkpoint checks pass.
No direct original method covers these filesystem boundaries; no credit added.
Disk queue native path checks cover missing parents, symlink traversal through
begin/commit/recovery and empty-directory constructor paths. The nine short
inherited original queue methods remain green; the separate two-billion-state
growth method was not rerun. No new original-method completion credit is added.
Native subprocess checks cover state-pool background missing-file, truncated-read
and nil-write failures: one source diagnostic, reader basename, actual exit 1,
and no normal return/deferred cleanup. Synchronous failure checks retain pending
work and caller error handling. Original DiskPoolWriter methods remain green;
no direct original method covers fatal failures, so no credit is added.
Native pool-reader checks cover pending/direct/cache reads across missing worker,
UID and level fields plus successful input. Source partial destination publication,
unvisited slots and pending-work retention are verified. Relevant original pool
writer and nine short inherited queue methods pass; no new method credit added.
Native disk queue write/dequeue/peek failure cases verify source coded runtime
assertions, detail parameters, no attached cause and unchanged queue mutations.
Related pool/recovery checks and short original queue methods pass. No direct
original method covers this catch boundary; no new completion credit is added.
Native cleaner checks verify missing/blocked deletion warnings, canonical
symlink paths, continued later deletion and range advancement. A subprocess
checks canonicalization failure diagnostic/exit; only the two short cleaner
cases run with race instrumentation. No direct original method earns new credit.
Attached model-value `[]Value` data now transfers through the native array graph.
Short payload/TCP checks cover nil/empty arrays, self-references, sharing with
state and tuple backing arrays, receiver ownership and invalid array references.
Original ModelValue methods remain green; no original attached-data transport
method exists, so this native supplement adds no completion credit. Opaque custom
data and other evaluator metadata remain a separate transferability audit.
Short native worker reply-loss cases additionally retain fully computed but
unreceived results behind a gate, close TCP and require source coordinator
retry/requeue/deregistration before releasing the old reply. They verify no
FP/trace/statistics publication, exact source diagnostics and no repeated cleanup,
with the worker runtime still alive. These cases add no original-method credit
and do not complete full-model network partition coverage.
The native MC06 `computed_worker_reply_loss` row also holds a fully computed
batch in a separate worker process before closing its TCP host. It confirms the
runtime remains alive, waits for coordinator cleanup and verifies unfinished work
before registering a replacement. Original N=7 final counts remain 114,942
distinct states and an empty queue. One source smaller-block retry, one
deregistration and the exact cache-statistic warning are required. This supplements
connection-loss coverage without credit for the four disabled original methods;
general network partitions remain missing.
The native MC06 `shared_worker_process_lost` row assigns nonempty RPC blocks
to two distinct workers sharing one listener, then kills their whole process.
It requires both loss/deregistration events and exactly the single cache warning
printed by Java MP's warning deduplication. With all workers removed, the
coordinator must still report unfinished work before a fresh worker registers.
The unchanged model finishes with 114,942 distinct states and an empty queue.
This native supplement adds no original-method completion credit.
Four native missing-coordinator cases now distinguish direct loss cleanup from
timer-triggered cleanup and successful status replies. They preserve assigned
work and worker counts at the source failure boundary. The production fix rejects
missing ownership instead of silently skipping deregistration. Existing native
retry/concurrency fixtures now supply their required owner without changing
assertions. There is no direct original method for this boundary; credit is unchanged.
Three native local `MultiFPSet` adapter cases now retain one storage endpoint
instead of exposing nested children as distributed servers. They verify original
storage membership and identity, scalar/block answers and internal high-bit
routing for memory/MSB/LSB storage. Existing original nested storage/manager tests
remain separate; these supplemental cases add no original-method credit.
Three native coordinator-run startup cases now reject missing thread/coordinator
ownership, retain the source worker-increment boundary and keep startup failures
outside catch/finally. Direct and owned-goroutine checks preserve assigned states,
keepalive, cleanup flag and cache-read behavior. No direct original method covers
these invalid-owner cases; original-method completion credit is unchanged.
Two native fingerprint exit-reply cases now exercise an actual completed exit
whose TCP reply is lost, directly and through manager shutdown. They retain the
exact net/rpc UnexpectedEOF cause, continue to later registrations and require
one exit without replay. Prior closed clients stay reportable; completed insertion
reply loss remains non-ignorable. The RMI-specific exception branch is replaced
with native exit traits. No direct original method covers this native boundary;
method completion credit is unchanged.
Null FP answers no longer become successful empty worker results: the shared
iterator preserves the source null failure, and native replies retain the
distinction between null vectors, null words and initialized empty words.
Short direct/TCP checks require worker predecessor context and KeepCallStack.
Successor validation also preserves the source distinction between a missing
state and an ordinary incomplete state. Four supplemental local/TCP cases cover
the failure cause, predecessor, call-stack flag and unchanged statistics; no
direct original Java method covers this boundary, so method credit is unchanged.
Default successor evaluation also rejects missing actions/predicates instead of
fabricating empty vectors. Six supplemental local/TCP cases distinguish those
failures from an ordinary false-action deadlock and preserve worker context and
statistics. No original method directly covers this boundary; credit is unchanged.
Literal disk traces with empty directory prefixes now commit and clean up their
owned files. Three supplemental cases cover commit/recovery, delete-before-failed-
promotion order and native resource deletion. No original method directly covers
this prefix; original-method credit is unchanged and isolated trace interruption
remains pending.
Call-stack construction now rejects missing evaluator owners rather than creating
a fresh default tool. Five additional native initialization/fingerprint checks
verify the existing source constraint-before-fingerprint boundary and suppression
of later work after failure; no fingerprint implementation change was needed.
These owner/initialization cases have no direct original method and add no credit.
Existing original BitVector printing and dynamic-manager methods remain green;
the new failure checks add no original-method completion credit.
Failover warning routing now matches ToolIO for scalar, batch and statistics
calls, including configured streams, captured messages and null detail text.
A short native TCP check verifies surviving-server reassignment after an
unambiguous pre-insertion disconnect and captures the warning. Original manager
translations remain green; broader FP failure models remain pending.
Worker readiness/completion and distributed option diagnostics now route through
ToolIO too. Short system/capture checks retain worker shutdown/unpublication/
latch behavior and exact option warning/error messages. Native CLI help and
startup rejection checks remain green; these stream checks add no original
method completion credit.

Short native worker TCP checks now hold an accepted remote fingerprint lookup
while keepalive/cache/exit calls use the same worker connection. Control calls
remain responsive, exited workers reject new calls and accepted computation
finishes after lookup release without replay or FP insertion. These checks add
no original-method credit or full-model/general-network-partition coverage.
Final fingerprint reporting now rejects a missing coordinator/manager instead
of fabricating zero statistics. Native direct and model-lifecycle checks require
failure before final-count publication, rate reset, success, summary or cleanup,
while retaining the preceding executor shutdown. No direct original method
covers the missing-owner boundary; method completion credit is unchanged.
Local disk/off-heap/nested fingerprint recovery now requires its supplied trace
instead of substituting named-file recovery. MemFPSet's source file-based path
still ignores the trace. Native cases cover primitive/nested stores, valid
snapshot controls and unchanged membership at failure. This boundary has no
direct original method and earns no additional method completion credit.

Coordinator file requests now use fresh resolvers. Native checks cover changing
default library/user directories over TCP, captured explicit overrides, separate
resource copies and directory cleanup. Existing original filename tests remain
green on Linux, preserving the Windows platform guard. No direct original method
covers changing coordinator defaults; original method completion credit is unchanged.

Native failed-constructor checks cover trace-open, null configuration, partial
nested disk initialization, invalid manager count and negative registration count.
They require joined queue workers and closed metadata handles while preserving
files, error precedence and successful ownership transfer. The existing original
DiskPoolWriter tests remain unchanged. No direct original method tests coordinator
rollback; these supplements add no original method completion credit.

Native fingerprint host ownership checks require storage to survive source
rejection/unpublication and a registration failure reported after acceptance.
Shutdown drains a blocked real lookup before closing disk handles, preserves
files, deduplicates repeated publication and rejects late registration. Existing
original manager constructor/concurrent-order tests remain unchanged. No original
method tests native host ownership; these checks add no method completion credit.

Native pre-registration FP startup checks now cover LSB/MSB partial initialization,
returned/panicked hostname failure and a missing coordinator. Disk handles close
while files, failure identity and announcement order remain intact. Registration
and reporting failures retain live storage until its owner closes it. These
isolated native checks add no original-method completion credit.

Repeated registration of the same native worker now has full MC06 N=7 coverage:
two coordinator threads share one worker identity, final distinct states remain
114,942, and the queue is empty. Two worker statistics and the source second-exit
warning are required; only a possible single source cache warning is allowed.
The native host remains alive through coordinator completion and then drains.
No original method covers this scenario; the four disabled methods remain Missing.

Attached model-value `map[string]Value` data now transfers through the native
payload graph. Focused checks require shared map identity, cycles, distinct-map
ownership, nil/empty maps, nil-valued entries and receiver mutations across
requests, results and worker exceptions. Invalid IDs, duplicate keys and invalid
entry references fail explicitly. Existing original ModelValue tests remain
green; no original method covers attached-map networking and no credit is added.

Mixed attached `[]any` and `map[string]any` graphs now transfer in native Go.
Direct/TCP checks cover map/array cycles, shared identity, distinct equal containers,
typed nil/empty containers, nil entries, Unicode/NUL keys, scalar bits and references
to byte and TLC value/typed-container graphs. Receiver mutations cannot change
sender storage. Invalid references, duplicate keys, unknown tags and opaque
entries fail explicitly. Existing original ModelValue methods remain green;
no original method directly tests mixed attachment networking, so no credit is added.

A full-model fingerprint insertion-reply-loss row now verifies that a host has
actually stored new successor fingerprints before it is killed without returning
the putBlock answer. Native callable failover redirects to the surviving host;
worker and coordinator snapshots independently converge on the survivor. The
existing source slot-based final count is 229,884 with an empty queue and FINISHED.
The deliberately interrupted coordinator call requires one EOF diagnostic;
other roles retain their no-EOF check, and all roles reject GENERAL. This adds
native failure-phase coverage without original-method completion credit.
The same insertion-reply-loss model now also runs with factory-created two-child
LSB/MSB fingerprint hosts. Both child files flush before the failed host withholds
its reply; real membership is checked after flushing. Both hosts must retain the
requested backend and all existing failover, final-count and diagnostic assertions.
The original N=7 bounds remain unchanged; this adds no original-method credit.

The full-model fingerprint lookup-reply-loss row holds a real nonempty
`containsBlock` answer, then kills its host. The worker must report exactly one
lost-reply EOF and reassign through the source callable; the coordinator must
independently converge on the same surviving registration. Both partition slots
alias that host. The unchanged N=7 model finishes with 229,884 slot-counted
distinct states, an empty queue and FINISHED, with no GENERAL in any role. Other
roles retain their no-EOF assertions. This native supplement leaves all four
disabled original methods missing.

The lookup-reply-loss full model also covers LSB/MSB hosts. Both child files
flush before lookup, and positive disk read counters are required before the
answer is held and the process killed. Existing N=7 bounds, source failover,
229,884 slot-counted final states, empty queue and diagnostic checks remain.
Both hosts must report the requested two-child backend. No original-method
credit is added.

Full MC06 N=7 coverage also runs two workers in one native process, exercising
the source shared application, fingerprint manager, executor and exit latch.
Both worker identities must share one listener, each must report actual sent
and received states, and final counts remain 114,942 distinct with an empty
queue and FINISHED. Both roles reject GENERAL and unexpected EOF. This native
coverage does not complete any of the four disabled original methods.

A short native server-thread statistics check observes real batch execution
and finalization concurrently, retaining signed 32-bit counter overflow and
NaN cache ratios. Getter/update synchronization fixes three native data races.
No direct original method covers this boundary; the original counts remain
37 port complete and four Missing.

Unused RMI exception scaffolding is removed. Existing native registration
failure checks now also retain I/O/transport categories and the rejection's
absent cause across payload transfer. Original manager and smart-proxy assertions
are unchanged; this cleanup adds no original-method completion credit.

Native host closure checks join callback cleanup for all four combinations of
forced/graceful callers, retain its failure and close each callback once.
Existing TCP checks still require graceful reply delivery and forced interruption
of a reply drain. No original Java method directly covers this native ownership
boundary; original method totals are unchanged.

Memory fingerprint recovery now propagates successful-read close failures.
Sixteen short Linux strace cases cover all three memory stores with real EIO
injection and controls, preserving earlier read failures, partial membership,
manager warnings and healthy continuation. Existing original buffered-input,
manager and MultiFPSet methods remain unchanged; native checks add no credit.

MemFPSet1 recovery now follows FileUtil.newDFIS's unbuffered primitive reads.
Seven short Linux syscall cases cover each header/key read failure and success,
retaining earlier field assignments, table identity/insertion counts, one close
and manager continuation. Existing close cases now retain the native EIO cause
for all three memory stores. No dedicated upstream MemFPSet1/SetOfLong test
exists; distributed inventory remains 37 port complete and four Missing.

MemFPSet1 checkpoint writes now follow FileUtil.newDFOS's unbuffered primitives.
Eight short Linux syscall cases cover success, all six write failures and final
close failure. Completed temporary prefixes, old checkpoint preservation, skipped
failed promotion, healthy continuation and one close without write retry are
required. Existing original manager and stream methods remain unchanged; no
dedicated source test exists and no original-method credit is added.

The two buffered memory stores now close their raw checkpoint file without
reflushing a failed buffer. Eight short Linux syscall cases preserve source
8192-byte buffering and cover normal, full-buffer/final-flush and close failure.
Exact prefixes, no write replay, one close, old checkpoint preservation and
healthy continuation are required. No dedicated original method covers this
boundary; inventory remains 37 port complete and four Missing.

Native coordinator discovery checks now preserve endpoint identity across name
replacement: old references stay with their coordinator, fresh lookups resolve
the replacement, unbind retains existing references and removal invalidates them.
Concurrent publication and connection cleanup remain covered. The earlier native
assertion that a captured reference follows a rebind was a transport shortcut;
it now requires the source behavior. No original method covers this boundary and
no original-method credit is added.

Buffered memory recovery now uses the source eager 8192-byte input stream,
preserving refill failure before the current fingerprint's insertion. Eight
Linux syscall cases cover both stores, initial/refill I/O failure and success,
exact membership prefixes, one close and healthy continuation. Existing original
buffered-input/manager tests and local/TCP partial/duplicate/close/startup checks
remain unchanged. No original-method credit is added.

Custom selector factories now have a linked Go registration boundary using the
original factory-name property. Nine isolated native cases preserve startup
capture, fresh construction, custom precedence, fallback/error/panic/nil
boundaries and coordinator/thread selection/statistics/retry dispatch. Existing
built-in and nine original smart-proxy contexts remain intact. No original
factory test exists; distributed inventory remains 37 port complete and four
Missing.

Four native factory-startup cases also exercise real application constructors
with local disk or distributed fingerprint managers. They preserve app/tool/flag
visibility before selection, subclass registration order and resource rollback
after callback panic. Existing constructor ownership checks remain unchanged;
no dedicated original factory test exists and no method credit is added.

Server-thread construction now preserves source separation from registration.
Native ownership checks require no map mutation and allow a nil underlying
worker proxy. Existing short registration URI/wake and lost-computed-reply checks
retain their assertions; the manually started fixture registers explicitly.
No upstream constructor method exists and original-method totals are unchanged.

Finite LongVec attachments now use the same native graph table as returned
fingerprint vectors. Direct/TCP/error-context checks preserve object aliases,
active elements, null/empty vectors, map-key identity and receiver isolation;
spare capacity and backing-array aliases are discarded as in the source.
Invalid references fail explicitly. Original LongVec/GrowingLongVec methods
remain port complete and unchanged; no transport method credit is added.

BitVector attachments now retain their complete word array, trailing zero
words, pointer identities and shared nonempty storage through the native graph.
Direct/TCP/result/error-context checks cover map keys, nil/empty distinctions,
receiver isolation and invalid object/word/kind references. Existing primitive
array checks include uint64 word arrays. Both original BitVector printing
assertions and all 44 ModelValue methods pass unchanged. There is no direct
original attachment method, so completion counts remain unchanged.

State-vector attachments now share the native graph with result partitions.
Direct/TCP/error-context checks preserve vector/state cycles, repeated and distinct
objects, active capacity, null/empty vectors, map keys and receiver ownership.
Malformed references and evaluator metadata remain rejected; decoded collections
retain TLCStateVec policy. No original TLCStateVec transport method exists and
no original-method status or credit changes.

Typed state/fingerprint vector arrays now share the native graph with result
partition arrays. Direct/TCP/error-context checks retain nonempty array aliases,
cycles, repeated/distinct arrays, null/empty arrays, null entries and receiver
isolation. Invalid attachment/nested/root/element references remain errors.
Existing original methods are unchanged; no native transport-method credit added.

The full local mid-run checkpoint/recovery harness now also uses two workers in
one shared-runtime process before the snapshot and two fresh replacement workers
after recovery. It checks producer registrations, persisted/recovered frontier
counts, both replacement identities and actual work, FINISHED, 114,942 distinct
states and an empty queue. No GENERAL or unexpected EOF is permitted. This adds
native checkpoint-barrier coverage without original-method completion credit.

The complete two-host remote checkpoint/restart matrix now also runs with two
workers sharing one process application/runtime, across Mem, LSB and MSB stores.
It preserves the existing single-worker cases and requires exact persisted queue
counts, committed file identity after crashes and full recovered partition
membership before evaluation. Both replacement worker identities must share a
listener and report actual work. All three unchanged N=7 models must finish at
114,942 distinct states and an empty queue, with no GENERAL. This verifies the
registered-endpoint library recovery lifecycle; Java's CLI recovery-before-FP-
registration limitation is unchanged and disabled methods receive no credit.

The same Mem/LSB/MSB remote restart matrix additionally drops one reply after the
actual fingerprint checkpoint commit. It requires the source warning, unchanged
two-host availability, committed snapshot files after crashes, exact recovered
membership before workers start and the ordinary full N=7 completion counts.
This is native transport failure coverage, with no original-method credit.

The full N=7 remote restart fixture additionally loses the first host's reply
after completed recovery. Both restored partitions retain exact checkpoint
membership before worker startup. The following size call triggers source
failover without replay; after the disowned host is crashed and joined, the
survivor completes at 114,942 stored fingerprints, with 229,884 reported across
the two aliased routing slots and an empty queue. Mem/LSB/MSB coverage adds no
original-method credit and preserves the ordinary recovery-row assertions.

A full N=7 direct-memory completed-begin reply-loss row now also passes: the
failed host receives no commit, pending bytes survive unpromoted, fresh recovery
warns/continues with that host empty, and replacement evaluation finishes at
114,942 states with an empty queue. Nested LSB/MSB begin-loss rows instead stop
fresh recovery before the healthy host/publication, preserving both pending
children without promotion and all retained checkpoints. Checked-I/O warning and
failover catches must not swallow this nested operation failure. These rows add
no original-method credit.

Short completed-begin reply-loss checks cover Mem/LSB/MSB storage, real pending
snapshot membership, skipped commit, retained registrations and fresh-store
recovery without pending-file promotion. They verify the source distinction
between ignored direct Mem recovery I/O and propagated nested disk recovery
failures. No original-method completion credit is added.

Native connection-owner checks additionally verify that joined already-closed
errors cannot hide other callback cleanup failures, including discovery-owned
coordinator cleanup. They preserve error causes and later-owner cleanup without
adding original Java method completion credit.

Native failed-client checks also cover coordinator, worker and fingerprint
transports. A terminal RPC failure closes its codec promptly, preserves the
original cause and retains the failed client without redial/replay. Method-level
RPC errors keep the connection usable. This Go transport ownership coverage has
no direct original Java method and adds no original-method completion credit.

Native worker-cache constructor checks now require the source typed negative
capacity error and dimension detail for wrapped shift inputs 31, 63 and -1.
No original SimpleCache test exists; this adds no original-method credit.

Native listener ownership checks cover accept failure, combined accept/close
failure, continued use of an accepted connection, and retained mixed shutdown
errors on coordinator/worker role owners. They add no original-method credit.

Native corrupt-queue recovery checks use a real N=7 checkpoint and fresh nested
LSB hosts. EOF stops startup before remote recovery or worker publication; hosts
remain empty, all roles join and retained snapshots stay unchanged. This adds
native phase coverage without original-method completion credit. The truncated
trace variant makes the queue checkpoint independently unavailable and requires
the trace EOF, verifying trace-before-queue ordering with the same shutdown and
retained-snapshot assertions.

Native response-flush checks also require a failed server write to close the
transport, unblock the peer and server reader, drain accepted-reply accounting
and perform the accepted fingerprint operation exactly once. The short isolated
race selection passes. This native codec check adds no original-method credit.

Native model-data payload checks now also cover map[any]any with scalar/value
keys, exact key types, cycles and shared references through requests, results and
WorkerException contexts. Nil/empty maps, receiver isolation and malformed keys
are checked. Existing original ModelValue tests pass; general-map network checks
add no original-method completion credit.

Attached native UniqueString checks cover sharing with value fields, metadata,
equal distinct names, typed nil values/keys and invalid IDs in direct payloads,
TCP results and WorkerException contexts. Original ModelValue/StringHelper tests
remain green; these networking checks add no original-method completion credit.

Attached native `[]*UniqueString` arrays now share the record-name array graph.
Direct/TCP/result/WorkerException checks retain backing-array sharing with record
and record-set fields and nested attachments, separate equal-content arrays,
native nil/empty array types, nil entries, name metadata and receiver isolation.
Invalid array and name references fail explicitly. The original ModelValue
methods pass; these new networking checks add no original-method completion
credit.

Finite operator container checks retain shared domain rows and result arrays
across operators and model-data attachments. Receiver updates change the sibling
operator's application without affecting the sender. Direct/TCP/failure-context
checks cover malformed references and nil/empty rows. These native checks add no
original-method completion credit.

Finite operator null-input checks additionally verify that nil rows/arguments
fail evaluation and nil rows/arguments/results fail initialization, while actual
empty rows remain valid. Local/TCP worker checks retain the evaluation failure
context. Original value-initialization methods remain green; native checks add
no original-method completion credit.

Lazy supplier native checks now reject type erasure during transfer. Cached
nil-supplier wrappers retain their own type and shared cached-value graph across
direct payloads, TCP results and WorkerException context. Nil-supplier evaluation
fails in source order regardless of cache contents. Executable Go suppliers are
rejected explicitly without evaluation or worker dispatch; uncached wrappers
keep the source lazy-transfer assertion. Invalid cache references fail decoding.
Original TRACE/TRACE-alias tests remain green. There is no direct original
supplier-network method, so these checks add no original-method completion credit.

Native attached `*TLCStateMut` checks retain shared root/predecessor identities,
attached-only states, cyclic value/cache back-references, native state map keys
and typed nil state references. Direct/TCP/result/WorkerException checks verify
receiver isolation; invalid graph references and unsupported attached evaluator
metadata still fail explicitly. Original ModelValue methods remain green. These
new networking checks add no original-method completion credit.

Native attached `[]*TLCStateMut` checks retain shared slice backing storage,
including aliases of invocation roots, nested containers and recursive state
references. Typed nil/empty arrays, null entries, independent equal arrays and
receiver isolation survive direct/TCP/result/WorkerException transfer. Malformed
array/element IDs, conflicting root forms and unsupported state metadata fail
explicitly. Existing legacy root/result payloads and original ModelValue methods
remain green; new native checks add no original-method completion credit.

The native callable-mode check verifies that ordinary model-checking states
ignore `TLCDefer` callbacks and remain transferable, while extended states execute
them and reject executable metadata transfer. This corrected an unconditional
Go setter; no direct original callable test exists. Seven existing original
simulation-trace test ports and focused payload/application checks pass unchanged.
This adds no original-method completion credit.

Native extended-copy checks cover shallow/deep copies with no predecessor,
consistent or changed predecessor levels, and the maximum-depth failure. Copies
now reapply the source predecessor setter rather than copying its fields. No
direct upstream copyExt test exists; existing state/vector/alias and seven
original simulation trace checks pass. Original-method counts are unchanged.

Native cache-mode checks now cover ordinary, extended mutable, functional and
print states. Only extended mutable states store values. Native cache graph
fixtures explicitly select that mode; print graph fixtures retain their cache
on a separate mutable state rather than inventing a wrapper cache. No direct
upstream accessor test exists; this adds no original-method completion credit.

Native base-state metadata checks cover print/functional states in extended mode:
action/callable setters remain no-ops, while predecessor assignment changes only
the level. Existing state/TCP graph and seven original simulation trace checks
pass. No direct source method tests this boundary; no original credit is added.
Separate print-wrapper ownership and delegated return values are now implemented.
Native checks preserve independent metadata, copy/bind identities, source equality
and name-identity behavior, shared owner roots and cached-data cycles. Malformed
owner graphs fail explicitly. Original RecordValue and Alias checks plus native
local/remote distributed DieHard process traces pass. No direct original wrapper
ownership/transfer method exists; original completion counts remain unchanged.

State-info conversion now preserves a print wrapper's original record identity,
including in counterexample nodes and received TCP states. Ordinary states still
produce fresh variable records. No direct original conversion test exists; native
identity checks and existing original record, alias and trace dump/load test ports
pass without changing original-method completion counts.

Native negative-location trace checks cover included/predecessor disk lookups
and the coordinator's worker-error catch. Failed seeks propagate I/O errors
instead of empty prefixes, retain pointer mutation and finish the queue without
spurious behavior output. Existing trace checks and three original buffered-file
seek methods pass. No direct original trace test covers this invalid location;
original-method completion counts remain unchanged.

Full N=7 coverage also pauses TCP traffic to one of two fingerprint hosts while
keeping connections open. Separate cases hold both directions, only requests
or only replies; markers require the selected direction and reject an opposite
direction block in asymmetric cases. Test control probes
require responsive coordinator status/manager and worker alive/cache calls,
unfinished model status and unchanged partition routing. After release the
ordinary 114,942 distinct states, empty queue and FINISHED are required, with no
GENERAL, EOF or failover warning. The relay preserves raw request/reply bytes and
drains replies before closing. This adds controlled network-stall coverage,
without original-method credit or exhaustive network-partition claims.

Native graceful-shutdown checks now also cover a decoded request header whose
sender never supplies its body. Shutdown expires pending socket reads while
retaining accepted handlers and reply writes. A second case combines an accepted
Exit with an incomplete request on the same TCP connection and checks the actual
successful Exit reply. These transport checks have no direct original Java test
and add no original-method completion credit.

Forty native named disk snapshot startup cases cover LSB/MSB stores, direct and
nested ownership, and local/TCP endpoints with missing, empty, duplicate,
descending and trailing-partial snapshots. They verify source I/O versus runtime
failure boundaries, joined sibling recovery, retained partial file/index/write
counts and reader ownership, and independent recovery-end statistics before
publication. The original LSB/MSB `testFPSetRecovery` methods remain green with
their unchanged 99,999 bound. This supplemental matrix has no direct enabled
original method and earns no additional original-method completion credit.

Native off-heap batch checks now cover nil rejection before storage access,
actual empty-vector acceptance, unchanged membership/seen counts and ordinary
batch continuation through local/TCP endpoints. Fixed the off-heap overrides'
nil-to-empty shortcut to match inherited FPSet behavior. Existing four original
ShortDiskFPSet batch methods and 18 short OffHeapDiskFPSet storage methods pass
unchanged. No direct original nil-batch method exists, so these native checks
earn no additional original-method completion credit.

Native queue batch checks cover nil array/vector rejection across memory, deque,
disk-state and byte-array queues. Rejection preserves queued work and releases
locks; explicit empty batches and subsequent work remain valid. The nine original
StateQueue methods and nine inherited DiskStateQueue methods pass unchanged,
as do native coordinator TCP retry/loss checks. No direct original null-batch
test exists, and these eight native cases add no original-method credit.

Eight native byte-queue read rows now cover raw extraction before decoding,
complete batch removal on decode failure, null raw entry boundaries and the
source oversized-request failure after removing available entries. Existing
DiskPoolWriter, BufferedDataInputStream and ValueInputOutputStream methods remain
unchanged and green. No original direct decode-boundary method exists; these
checks add no original-method completion credit.

Native byte-queue conversion checks additionally reject missing states before
publication in zero/one-variable models, including an array's valid prefix.
Eight rows verify direct conversion, ordinary/synchronized enqueue and arrays,
with continued queue use. Existing writer/stream Java methods pass unchanged;
there is no direct original null-state method and no additional test credit.

Remote server/worker integration, init failures, fingerprint-manager failover, and smart-proxy calculations. Native Go fingerprint, worker and coordinator TCP calls, ordinary state/value/result payloads, structured worker failures, manager snapshots, discovery and worker/coordinator/FP lifecycle publication are implemented and unit-verified. Native CLI entry points are wired with focused help/property/address checks. Native coordinator signal shutdown and separate-process DieHard execution with local and standalone remote fingerprint storage are verified. The native process harness also checkpoints and recovers the full MC06 initial frontier through the real `-recover` CLI in fresh coordinator/worker processes. It requires 16384 recovered fingerprints/queued states, no repeated initialization, and the original final 114942 distinct/0 queued result. This is additional native coverage, not original disabled-harness completion credit. A native worker-loss row also verifies requeueing after killing a worker with an assigned block, survivor/replacement completion, exact one-time deregistration and the original cache-warning behavior. Fresh-process mid-run local MemFPSet recovery now also completes the full unchanged model after abrupt post-commit producer exit, preserving exact checkpoint/recovery counts, no repeated initialization and the original final result. It adds no original-method credit. A pre-commit interruption row also checks old queue recovery with full-trace MultiFPSet reconstruction and the original final model counts. Interruption after queue commit now passes with exact committed queue/full-trace recovery counts and unchanged old trace metadata. Interruption after intern commit also passes with committed trace/intern metadata before fingerprint commit. Nested fingerprint commit interruption now passes with independently inspected mixed checkpoint files. Isolated trace-commit interruption, extended/custom values and broader network/process failure coverage remain pending. Fresh-process remote-FP CLI recovery is a source startup limitation, not implemented Java behavior. The new Go boundary tests add no original-method completion credit.

Native address-reuse checks now cover generated coordinator-owned fingerprint,
standalone fingerprint and worker references. The original counter-only names
allowed a stale lazy reference to mutate a different store at a reused address.
Generated names now include a random native host identity. Old references fail
through the existing remote I/O and manager reassignment paths; fresh references
work and replacement storage remains isolated. No original Java method directly
tests this native transport case; no original-method credit is added.

Worker callback registration now preserves the source queue-wake-before-getURI
ordering on connection refusal. Native TCP checks cover the formerly skipped wake
and deferred callback connection ownership, including concurrent first calls,
closed unused references, failed-client retention and idempotent closure. The
original smart-proxy methods remain unchanged; these native boundary checks add
no original-method completion credit.

Bulk queue accounting now follows the source post-loop length update. Native
checks cover partial disk-spill failures and independent deque occupancy/growth.
The Go-only memory-vector check now preserves the source's overwritten-slot
quirk; distributed retry/worker-loss fixtures use the actual disk queue with
their state identity/order assertions unchanged. The nine original StateQueue
methods and nine inherited disk methods remain unchanged. These native cases
add no original-method credit.

Native raw byte-queue storage checks cover missing-parent failures, literal
symlink traversal, empty configured paths, partial slot publication, inactive
slot retention and source short-final-read behavior. The original DiskPoolWriter
wake/finish methods and buffered/value stream methods pass unchanged. No direct
original raw queue storage method exists; these cases add no completion credit.

Native raw-pool failure checks now require process exit and one source pool
diagnostic for background I/O/allocation/nil-entry failures. Synchronous queue
checks preserve coded errors, nil causes, detail fallback and unchanged work.
Failed checkpoint buffering/marker publication is checked directly. Existing
original DiskPoolWriter and stream methods remain unchanged and green; these
native failure cases add no original-method credit.

The existing native cleaner matrix now covers both state and raw byte queues,
including failed missing/nonempty-directory deletes, canonical symlink paths,
later successful deletion and range progress. Raw canonicalization errors now
have child-process coverage for one error-severity event and process exit.
Original writer tests remain unchanged; these native cases add no test credit.

Fresh-process remote recovery now also covers a missing first-host committed
child snapshot with LSB/MSB storage. It checks source failure reporting before
publication, no later-host recovery, joined sibling storage, caught-failure
shutdown without cleanup, and unchanged retained checkpoint bytes. Existing
successful recovery/model completion assertions remain unchanged. This native
failure coverage does not translate the disabled distributed model harness and
adds no original-method credit.

A separate direct MemFPSet missing-snapshot row verifies the source's checked-I/O
continuation through fresh processes. It requires one warning, unchanged routing,
empty damaged-host membership, exact healthy-host membership, no replacement of
the missing file and unchanged persisted queue count before workers start. The
full N=7 model must finish at 114,942 states with an empty queue and no GENERAL.
This native failure-phase coverage adds no original-method completion credit.

Source-style disk traces no longer retain every written state/action in the
native in-memory mirror. New ownership checks cover initial, record and successor
writes, actual disk positions/fingerprints and depth traversal. The active
checkpoint fixture uses file depth/last-pointer observations and its existing
recovered-file enumeration. No direct original method tests this ownership
boundary; original trace assertions and completion counts remain unchanged.

Native scalar character attachments now preserve `uint16` alongside character
arrays. Direct transfer checks all 65,536 code units; TCP result/error-context
checks preserve typed character map keys, and malformed ranges are rejected.
All 44 original ModelValue methods remain unchanged and green. There is no
dedicated upstream character-attachment test, so original credit is unchanged.

Native coordinator read checks in `distributed_server_file_path_test.go` verify
relative/absolute symlink traversal, exact directory/open-failure diagnostics
and retained nested file-open causes. Absolute diagnostics now preserve dot
components rather than naming a different, collapsed location. There is no
direct original distributed method for this case; test credit is unchanged.

The native accepted-checkpoint connection-loss matrix now covers memory and
factory-created two-child LSB/MSB stores across begin, commit and recovery.
Exact pending/committed child bytes, fresh-store committed recovery, continued
healthy-host work and retained registration identities are checked. Accepted
storage work finishes after transport closes, without replay or reassignment.
This adds no original-method credit and does not cover process death or atomicity.

Native LSB/MSB invariant ownership checks cover flush/open/scan I/O failure and
valid/invalid order through local/TCP endpoints. Corrected premature table-lock
release on failures before the source finally region and moved expected-count
comparison after release. No original method directly tests this failure
boundary; these twenty short cases earn no original-method credit.
A separate twelve-case Linux syscall matrix verifies real scan-close EIO
overrides valid/invalid/truncated scan results and retains locks, with successful
close controls. It requires strace and adds no original-method credit.

Native final-reporting coverage also runs the unchanged EWD840 N=7 model with
a remote Mem host returning or raising checked I/O failure from CheckFPs.
It requires one GENERAL, the source Long.MAX_VALUE distance fallback, subsequent
statistics and cleanup, 114,942 distinct fingerprints and an empty queue.
The normal model harness retains its zero-GENERAL gate. No original Java method
directly tests this failure branch; original-method credit remains unchanged.
The reporting boundary now also preserves signed distances and exact two-digit
decimal rounding. Focused native checks cover local checked-I/O `-1`, MinInt64,
decimal ties and adjacent large integers, plus zero-distance failure and the
empty-model bypass. No direct original method tests this calculation; these
checks do not alter the inventory count.
Local-manager constructor checks also preserve delayed missing-reference failure,
valid counts/empty batches and worker transfer without fabricated endpoint
publication. Existing local lifecycle checks now use the real constructor.
No direct original method covers this boundary; the inventory is unchanged.
The full native N=7 final-reporting matrix additionally covers death of the sole
Mem host after its real CheckFPs completes and before its reply. It preserves
GENERAL, later states-seen exhaustion warnings, captured distinct/queue counts,
source success fallback and joined surviving roles. This native fault case does
not weaken the original zero-GENERAL model assertions or add original credit.
Native fingerprint RPC now delegates trace recovery with a nil argument instead
of rejecting it unconditionally. Four short memory/disk/nested cases preserve
actual checkpoint recovery or the source unchecked missing-trace failure, plus
one accepted call and continued connection use. Non-null traces stay local.
No direct original method covers this transport boundary; credit is unchanged.

- [ ] [tlc2/tool/distributed/DieHardDistributedTLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/DieHardDistributedTLCTest.java) — **Reconcile**: `testSpec`.
  Source setup/body: [tlc_distributed_java_test.go](../tlc_distributed_java_test.go). The ordinary entry preserves the exact upstream unconditional skip. Explicit body diagnostic: [tlc_distributed_java_diagnostic_test.go](../tlc_distributed_java_diagnostic_test.go), build tag `tlago_disabled_distributed_tests`. All seven trace states, ordinal/action-label checks, FINISHED, BEHAVIOR and no GENERAL are staged; the source Ant-profile run passes with 48 workers. The restored source assumption is separate from diagnostic execution; no passing-body or completion credit is added.
- [x] [tlc2/tool/distributed/DistributedDoInitFunctorInvariantContinueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/DistributedDoInitFunctorInvariantContinueTest.java) — **Port complete**: `testSpec`. Translation: [tlc_init_model_java_test.go](../tlc_init_model_java_test.go). Original inherits ordinary ModelCheckerTestCase rather than a remote server harness. Identical original NotNine model/config bytes, inherited exit and all diagnostic assertions retained; continuation also preserves exact counts and uncovered assertion. Unchanged Java and Go pass.
- [x] [tlc2/tool/distributed/DistributedDoInitFunctorInvariantTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/DistributedDoInitFunctorInvariantTest.java) — **Port complete**: `testSpec`. Translation: [tlc_init_model_java_test.go](../tlc_init_model_java_test.go). Original inherits ordinary ModelCheckerTestCase rather than a remote server harness. Identical original NotNine model/config bytes, inherited exit and all diagnostic assertions retained; continuation also preserves exact counts and uncovered assertion. Unchanged Java and Go pass.
- [ ] [tlc2/tool/distributed/EWD840DistributedTLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/EWD840DistributedTLCTest.java) — **Reconcile**: `test`.
  Source setup/body: [tlc_distributed_java_test.go](../tlc_distributed_java_test.go). The ordinary entry preserves the exact upstream unconditional skip. Explicit body diagnostic: [tlc_distributed_java_diagnostic_test.go](../tlc_distributed_java_diagnostic_test.go), build tag `tlago_disabled_distributed_tests`. FINISHED, exact 114942 distinct/0 queued STATS and no GENERAL are staged. The source Ant-profile run originally failed during final CheckFPs on a retained closed flusher. After the user-authorized Go lifecycle fix, the unchanged diagnostic passes in 74.80 seconds; see [JAVA_BUG_FOUND.md](../JAVA_BUG_FOUND.md). The restored source assumption is separate from diagnostic execution; no original-method completion credit is added.
- [ ] [tlc2/tool/distributed/EWD840DistributedWithFPSetTLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/EWD840DistributedWithFPSetTLCTest.java) — **Reconcile**: `test`.
  Source setup/body: [tlc_distributed_java_test.go](../tlc_distributed_java_test.go). The ordinary entry preserves the exact upstream unconditional skip. Explicit body diagnostic: [tlc_distributed_java_diagnostic_test.go](../tlc_distributed_java_diagnostic_test.go), build tag `tlago_disabled_distributed_tests`. FINISHED, exact 114942 distinct/0 queued STATS and no GENERAL are staged with one standalone fingerprint role; the source Ant-profile run passes with 48 workers. The restored source assumption is separate from diagnostic execution; no passing-body or completion credit is added.
- [x] [tlc2/tool/distributed/TLCSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/TLCSetTest.java) — **Port complete**: `testSpec`. Translation: [tlc_distributed_test.go](../tlc_distributed_test.go). Full server/application initialization with unchanged original model/config, source default FPSet ratio and MSB dummy equivalent, isolated registry namespace, and all three diagnostic assertions. Native FPSet exit preserves the process as Java dummy exit does. Unchanged Java and Go pass; no network transport substitute is claimed.
- [ ] [tlc2/tool/distributed/TSnapShotDistributedTLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/TSnapShotDistributedTLCTest.java) — **Reconcile**: `test`.
  Source setup/body: [tlc_distributed_java_test.go](../tlc_distributed_java_test.go). The ordinary entry preserves the exact upstream unconditional skip. Explicit body diagnostic: [tlc_distributed_java_diagnostic_test.go](../tlc_distributed_java_diagnostic_test.go), build tag `tlago_disabled_distributed_tests`. FINISHED, STATS queue=0, no GENERAL and BEHAVIOR are staged; the source Ant-profile run passes with 48 workers. The restored source assumption is separate from diagnostic execution; no passing-body or completion credit is added.

### Fingerprint sets, indexers, arrays, and iterators

Disk/memory/off-heap factories, recovery, duplicate merging, high/low fingerprints, indexers, arrays, and iterators. Preserve parameter matrices and original resource/concurrency behavior.

- [x] [tlc2/tool/fp/Bug210DiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/Bug210DiskFPSetTest.java) — **Port complete**: `testDiskLookupWithOverflow`.
  Go translation: [tlc/fpset_disk_bugs_java_test.go](fpset_disk_bugs_java_test.go). All whole original methods, constructor/configuration inputs and catch rules are retained. Retains the complete index of MaxInt32/1024+8 entries and three MaxInt64 offsets. Production widens page indices before multiplication as Java does; original test passes on amd64 and 386.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [x] [tlc2/tool/fp/Bug242DiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/Bug242DiskFPSetTest.java) — **Port complete**: `testDiskFPSetWithHighMem`, `testDiskFPSetIntMaxValue`, `testDiskFPSetIntMinValue`, `testDiskFPSetZero`, `testDiskFPSetOne`.
  Go translation: [tlc/fpset_disk_bugs_java_test.go](fpset_disk_bugs_java_test.go). All whole original methods, constructor/configuration inputs and catch rules are retained. Retains memory2097153638/MaxInt32/MinInt32/0/1, ratio1, uninitialized dummy constructors, catch-only-OutOfMemoryError allowance for the two large inputs and catch(Exception) excluding Error.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [x] [tlc2/tool/fp/Bug246DiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/Bug246DiskFPSetTest.java) — **Port complete**: `testLinearFillup`, `testFlushDiskFPSet`.
  Go translation: [tlc/fpset_bug246_java_test.go](fpset_bug246_java_test.go). Whole original runtime-memory-derived fill loop and exception assertions retained, including zero-reader initialization, ratio1, descending MaxInt64 fingerprints, duplicate rejection and last-observed OOM statistics. Native race run retains all 67,108,863 insertions on this machine. Production preserves zero dedicated readers through initialization/reopening. Original `testFlushDiskFPSet` contains no active statements and remains empty; that translation does not establish flush coverage.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [x] [tlc2/tool/fp/FPSetFactoryTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/FPSetFactoryTest.java) — **Port complete**: `testGetDiskFPSet`, `testGetFPSetMSB`, `testGetFPSetLSB`, `testGetFPSetOffHeap`, `testGetFPSetMSBWithMem`, `testGetFPSetLSBWithMem`, `testGetFPSetMSBWithMemAndRatio`, `testGetFPSetLSBWithMemAndRatio`, `testGetFPSetMultiFPSet`, `testGetFPSetLSBMultiFPSet`, `testGetFPSetOffHeapMultiFPSet`, `testGetFPSetMultiFPSetWithMem`, `testGetFPSetLSBMultiFPSetWithMem`, `testGetFPSetOffHeapMultiFPSetWithMem`, `testGetFPSetOffHeapMultiFPSet42`. Complete original methods and helpers in [fpset_factory_java_test.go](fpset_factory_java_test.go).
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [x] [tlc2/tool/fp/LSBDiskFPsetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/LSBDiskFPsetTest.java) — **Port complete**: `testCtorLLMinus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorLL` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorLLPlus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorLLNextPow2Min1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16Minus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16Plus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16NextPow2Min1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorULMinus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorUL` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorULPlus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorULNextPow2Min1` (from `AbstractHeapBasedDiskFPSetTest`), `testFPSetRecovery` (from `AbstractHeapBasedDiskFPSetTest`), `testFPSetRecovery2` (from `AbstractHeapBasedDiskFPSetTest`), `testFPSetRecoveryDuplicate` (from `AbstractHeapBasedDiskFPSetTest`). Whole inherited methods → [heap_disk_fpset_java_test.go](heap_disk_fpset_java_test.go); raw-memory DummyFPSetConfiguration override, all 12 exact budgets/three inequalities, complete 99,998-fingerprint trace recovery, 1,024 forced-flush calls and original duplicate-warning behavior retained. Full class passes Java, Go normal and Go race checks.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [x] [tlc2/tool/fp/LongArrayTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/LongArrayTest.java) — **Port complete** → [long_array_java_test.go](long_array_java_test.go). All seven original methods, complete 100-element signed extremes/CAS loops, both typed AssertionError catches, original zero-memory loop conditions/increments, 10,321-element swap, full unseeded JavaRandom 21,383-element swap and source architecture assumption retained. Production bounds use Java AssertionError and preserve Error catch/class metadata.
  Related Go checks: [tlc/long_array_test.go](long_array_test.go).
- [x] [tlc2/tool/fp/LongArraysTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/LongArraysTest.java) — **Port complete** → [long_arrays_java_test.go](long_arrays_java_test.go). All six original methods and complete comparator/partition/sentinel/member/wrapped-range/order/count helpers, all 44 Basic2 literals and 15 range assertions retained. Original zero-position empty indexer remains literal; production infinite-precision constructor accepts zero as Java does.
  Related Go checks: [tlc/long_array_test.go](long_array_test.go).
- [x] [tlc2/tool/fp/MSBDiskFPSetTest2.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/MSBDiskFPSetTest2.java) — **Port complete**: `testCtorLLMinus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorLL` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorLLPlus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorLLNextPow2Min1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16Minus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16Plus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16NextPow2Min1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorULMinus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorUL` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorULPlus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorULNextPow2Min1` (from `AbstractHeapBasedDiskFPSetTest`), `testFPSetRecovery` (from `AbstractHeapBasedDiskFPSetTest`), `testFPSetRecovery2` (from `AbstractHeapBasedDiskFPSetTest`), `testFPSetRecoveryDuplicate` (from `AbstractHeapBasedDiskFPSetTest`), `testGetLast`, `testHighFingerprint1`, `testHighFingerprint2`, `testGetLastNoBuckets`. Whole 19-method class → [msb_disk_fpset_java_test.go](msb_disk_fpset_java_test.go), using full inherited helpers from [heap_disk_fpset_java_test.go](heap_disk_fpset_java_test.go). Preserve lower256/upper2^31, all exact budgets/recovery ranges, raw100-fingerprint helper and its assertion, discarded-new/old-iterator check, exact high fingerprints and all original duplicate/exception assertions.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [x] [tlc2/tool/fp/OffHeapBitshiftingIndexerParameterizedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapBitshiftingIndexerParameterizedTest.java) — **Port complete**; all five inherited methods (`testZero`, `testOne`, `testLongMin`, `testLongMax`, `testSome`) → [fpset_offheap_indexer_parameterized_java_test.go](fpset_offheap_indexer_parameterized_java_test.go). All 1,104 original rows, duplicate rows, exact 1,024-iteration loop, fresh constructors at each call, subclass indexer types and original assumptions retained. All three unchanged Java classes pass; complete Go matrices pass normal and race. No production changes.
  Related Go checks: [tlc/offheap_indexer_test.go](offheap_indexer_test.go).
- [x] [tlc2/tool/fp/OffHeapBitshiftingIndexerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapBitshiftingIndexerTest.java) — **Port complete**: `testBitshifting`, `testBitshifting2`, `testShift1_268435456`, `testBitshiftOvershoot`, `testNoOverflowErrorBitShifting`. Whole original methods and complete helper sweeps → [offheap_bitshifting_indexer_java_test.go](offheap_bitshifting_indexer_java_test.go); preserve both 128-position boundary sweeps, wraparound probes, overshoot fingerprint and catch-only-TLCRuntimeException constructor check.
  Related Go checks: [tlc/offheap_indexer_test.go](offheap_indexer_test.go).
- [x] [tlc2/tool/fp/OffHeapDiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapDiskFPSetTest.java) — **Port complete**: `testInsertAndEvict1`, `testInsertAndEvict2`, `testInsertAndEvict3`, `testInsertAndEvict4`, `testInsertAndEvict5`, `testInsertAndEvict6`, `testInsertAndEvict7`, `testInsertAndEvict8`, `testInsertAndEvict9`, `testInsertAndEvict10`, `testInsertAndEvict11`, `testInsertAndEvict12`, `testInsertAndEvict13`, `testInsertAndEvict14`, `testInsertAndEvict15`, `testInsertAndEvict16`, `testOffset1Page`, `testOffset3Page`, `testOffset5Page`, `testOffset9Page`, `testWriteIndex`, `testMergeDuplicate`, `testMergeDistinct`. Whole 23-method class → [offheap_disk_fpset_java_test.go](offheap_disk_fpset_java_test.go); all 16 original seeds/lengths, complete insertion and post-eviction memory-only assertions, exact shrinking-set offset loops, full 99,999,999-entry index overflow test, dummy reader/iterator overrides and all duplicate/distinct merge output and warning-set assertions retained. Production now uses the non-heap base constructor, sorts the actual array, marks entries through the original iterator, merges streams and writes indexes with Java addressing and source assertion checks. All 23 pass Java and Go; the 22 methods other than the huge index traversal additionally pass race checks.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [x] [tlc2/tool/fp/OffHeapIndexerEquivalenceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapIndexerEquivalenceTest.java) — **Port complete**; parameterized: preserve all original rows: `testInfiniteInfMult`. Full 7,254-row matrix and complete original random/uniform/boundary helper → [offheap_indexer_equivalence_java_test.go](offheap_indexer_equivalence_java_test.go); preserve assumption, unseeded bounded sampling, all 1,024-iteration limits and source monotonicity/equality assertions. Both @Ignore methods remain uncredited.
  Related Go checks: [tlc/offheap_indexer_test.go](offheap_indexer_test.go).
- [x] [tlc2/tool/fp/OffHeapInfPrecisionIndexerParameterizedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapInfPrecisionIndexerParameterizedTest.java) — **Port complete**; all five inherited methods (`testZero`, `testOne`, `testLongMin`, `testLongMax`, `testSome`) → [fpset_offheap_indexer_parameterized_java_test.go](fpset_offheap_indexer_parameterized_java_test.go). All 1,104 original rows, duplicate rows, exact 1,024-iteration loop, fresh constructors at each call, subclass indexer types and original assumptions retained. All three unchanged Java classes pass; complete Go matrices pass normal and race. No production changes.
  Related Go checks: [tlc/offheap_indexer_test.go](offheap_indexer_test.go).
- [x] [tlc2/tool/fp/OffHeapIteratorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapIteratorTest.java) — **Port complete**: `testNext`, `testMarkNext`.
  Go translation: [tlc/offheap_iterator_java_test.go](offheap_iterator_java_test.go). Both whole original methods retain the LongArray platform assumption, all 64 entries, elements32, infinite-precision indexer64/1, complete traversal/count/preservation/mark assertions and the source repeated array.get(32) upper-half loop. Production ports markNext and the original do/while, wrap assertion, exhaustion and distance-task exception handling.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [x] [tlc2/tool/fp/OffHeapMult1024IndexerParameterizedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapMult1024IndexerParameterizedTest.java) — **Port complete**; all five inherited methods (`testZero`, `testOne`, `testLongMin`, `testLongMax`, `testSome`) → [fpset_offheap_indexer_parameterized_java_test.go](fpset_offheap_indexer_parameterized_java_test.go). All 1,104 original rows, duplicate rows, exact 1,024-iteration loop, fresh constructors at each call, subclass indexer types and original assumptions retained. All three unchanged Java classes pass; complete Go matrices pass normal and race. No production changes.
  Related Go checks: [tlc/offheap_indexer_test.go](offheap_indexer_test.go).
- [x] [tlc2/tool/fp/ShortDiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/ShortDiskFPSetTest.java) — **Port complete**: `testWithoutZeroFP`, `testWithoutMinFP`, `testWithoutMaxFP`, `testZeroFP`, `testMinFP`, `testMinMin1FP`, `testNeg1FP`, `testPos1FP`, `testMaxFP`, `testValues`, `testDiskLookupWithFpOnLoPage`, `testMemLookupWithZeros`, `testMemLookupWithMin`, `testMemLookupWithMax`, `testDiskLookupWithZeros`, `testDiskLookupWithMin`, `testDiskLookupWithMax`, `testDiskLookupWithMaxOnPage`, `testDiskLookupWithZerosOnPage`, `testDiskLookupWithLongMinValueOnPage`, `testComparePutAndPutBlock`, `testCompareContainsAndContainsBlock`, `testContainsBlock`, `testPutBlock`. Whole 24-method class → [short_disk_fpset_java_test.go](short_disk_fpset_java_test.go); retain all six original runKnown conditional early returns and their full bodies, the complete 1,125-tuple interpolation matrix, 3,072-fingerprint low-page traversal, original memory configurations, manual zero/min/max flushes, duplicate page loops and block comparisons. Production now preserves Java double-to-long conversion and fixed-length LSB flush buffers including unfilled zeros.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [x] [tlc2/tool/fp/iterator/TLCIterator1Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/iterator/TLCIterator1Test.java) — **Port complete**: all three original/inherited methods (`testNext`, `testNoNext`, `testGetLast`) → [fpset_iterator_java_test.go](fpset_iterator_java_test.go). Complete eight-bucket buffers, all 21 reads, strict order/read count, typed exhaustion catch, largest unflushed entry and fresh setup retained for every concrete variant. Production ports source getLast/reads and exact exhaustion/monotonic exception boundaries.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [x] [tlc2/tool/fp/iterator/TLCIterator2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/iterator/TLCIterator2Test.java) — **Port complete**: all three original/inherited methods (`testNext`, `testNoNext`, `testGetLast`) → [fpset_iterator_java_test.go](fpset_iterator_java_test.go). Complete eight-bucket buffers, all 21 reads, strict order/read count, typed exhaustion catch, largest unflushed entry and fresh setup retained for every concrete variant. Production ports source getLast/reads and exact exhaustion/monotonic exception boundaries.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [x] [tlc2/tool/fp/iterator/TLCIteratorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/iterator/TLCIteratorTest.java) — **Port complete**: all three original/inherited methods (`testNext`, `testNoNext`, `testGetLast`) → [fpset_iterator_java_test.go](fpset_iterator_java_test.go). Complete eight-bucket buffers, all 21 reads, strict order/read count, typed exhaustion catch, largest unflushed entry and fresh setup retained for every concrete variant. Production ports source getLast/reads and exact exhaustion/monotonic exception boundaries.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).

### Queues and pool writers

Memory-state queues and disk/byte-array writer wakeup/finish behavior; stress and JPF suites are listed separately below.

- [x] [tlc2/tool/queue/DiskPoolWriterTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/queue/DiskPoolWriterTest.java) — **Port complete**: `testStatePoolWriterIgnoresEmptyWakeAndStopsOnFinish`, `testByteArrayPoolWriterIgnoresEmptyWakeAndStopsOnFinish`.
  Go translation: [tlc/disk_pool_writer_java_test.go](disk_pool_writer_java_test.go). Both source methods retain actual writer lifecycle observation, notification under the writer lock, 10ms polling, 5000ms deadlines, return-to-wait assertions and actual termination after finishAll.
- [x] [tlc2/tool/queue/StateQueueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/queue/StateQueueTest.java) — **Port complete**: `testEnqueue`, `testsDequeueEmpty`, `testDequeueEmpty`, `testsDequeueNotEmpty`, `testDequeueNotEmpty`, `testEnqueueAddNotSame`, `testEnqueueAddSame`, `testsDequeueAbuseEmpty`, `testsDequeueAbuseNonEmpty`.
  Go translation: [tlc/queue_test.go](queue_test.go).

### Values, lazy functions, enumeration, and value streams

Original primitive/composite value, normalization, comparison, EXCEPT, serialization, subset, and random-enumerator cases.

- [x] [tlc2/value/StringDeserializeTLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/StringDeserializeTLCTest.java) — **Port complete**: `test`.
  Go translation: [tlc_string_deserialize_java_test.go](../tlc_string_deserialize_java_test.go), with all three byte-identical original vectors and the inherited model harness/success exit assertion.
- [x] [tlc2/value/ValueInputOutputStreamTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/ValueInputOutputStreamTest.java) — **Port complete**: `testWriteShort`, `testWriteInt`, `testWriteShortNat`, `testWriteNat`, `testBlindReadStringValue`, `testBlindReadRecordValue`. Source startup useGZIP=false and explicit per-method gzip choices retained. Production constructors now use the original 8,192-byte BufferedData streams, including eager input and the File input overload’s two buffer layers; byte-queue adapters remain direct as in Java. UniqueString data primitives now use the original buffered string protocol; intern-table recovery uses its actual buffered EOF state rather than peeking past prefetched bytes.
  Go translation: [tlc/value_stream_java_test.go](value_stream_java_test.go). All six original methods, exact gzip sizes 23/26, raw cold string/record bytes and metadata -1, and read-before-intern assertions are retained.
- [x] [tlc2/value/impl/EnumerableValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/EnumerableValueTest.java) — **Port complete**: `test`.
  Go translation: [tlc/value_enumerable_java_test.go](value_enumerable_java_test.go); complete seeded size sweep 1 through 10,656.
- [x] [tlc2/value/impl/FcnLambdaValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/FcnLambdaValueTest.java) — **Port complete**: `testEmptyIntervalDomainToTuple`, `testToString`, `testToFcnRcd`, `testSelectAndApply`, `testTakeExceptOverridesValue`, `testDomainIsStableAcrossConversion`, `testToFcnRcdReturnsCachedInstance`, `testToRcdForIntervalDomainIsNull`, `testToRcdForStringDomain`, `testFingerprintStableAcrossConversion`, `testTakeExceptThenConvertToFcnRcd`, `testToFcnRcdAssertFail`, `testToFcnRcdClassCastException`, `testToFcnRcdSilentCorruption`, `testToFcnRcdSilentCorruptionFP`, `testToTupleWithExceptIntervalDomain`, `testToTupleWithExceptSetEnumDomain`, `testToTupleWithExceptFP`.
  Go translation: [tlc/value_lambda_java_test.go](value_lambda_java_test.go).
  All 18 original methods, evaluation mock and dummy-state helper, 49 equality / 3 null / 1 identity assertions, original EXCEPT conversion order and fingerprint calls are retained.
- [x] [tlc2/value/impl/FcnRcdValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/FcnRcdValueTest.java) — **Port complete**: `testSelecEmpty`, `testSelecNormalizedEmpty`, `testEmptyIntervalDomainToTuple`, `testSelect`, `testSelectNormalized`, `testSelectLinearSearchTypedMV`, `testSelectBinarySearchTypedMV`, `testMalformedExplicitFcnEqualsIntervalDoesNotWrap`, `testMalformedIntervalFcnSelectDoesNotWrap`, `testMalformedIntervalFcnExceptDoesNotWrap`, `testEmptyIntervalFcnCompareToAgreesWithEquals`, `testEmptyIntervalFcnVsEmptyTupleCompareTo`, `testEmptyIntervalFcnsNormalize`.
  Go translation: [tlc/value_function_test.go](value_function_test.go).
  Preserves the full -64..63 domain / -128..127 selection matrix and exact typed-model-value exception messages; production binary search now follows Arrays.binarySearch.
- [x] [tlc2/value/impl/InitializeValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/InitializeValueTest.java) — **Port complete**: `union`, `setcap`, `setcup`, `setdiff`, `subset`, `record`, `fcnrecord`, `tuple`, `setOfTuple`, `setOfRcds`.
  Go translation: [tlc/value_initialize_test.go](value_initialize_test.go).
- [x] [tlc2/value/impl/IntervalValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/IntervalValueTest.java) — **Port complete**: `testElementAt`, `testElementAtOutOfBoundsNegative`, `testElementAtOutOfBoundsSize`, `sizeOverflow`, `compareToOverflow1`, `testCompareExtremeIntervals`, `testEmptyIntervalEquality`, `testCompareEmptyIntervals`, `testSizeOfMaximumRepresentableInterval`, `testExtremeIntervalSize`, `testMaxIntSingletonEnumerator`, `testMaxIntSingletonSubsetEq`, `testMaxIntSingletonFingerprint`, `testMaxIntSingletonDiffCapCup`.
  Go translation: [tlc/value_interval_test.go](value_interval_test.go).
  Added production elementAt with original bounds evaluation order and source-aware failures; all fourteen original methods retain their inputs and assertions.
- [x] [tlc2/value/impl/KSubsetValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/KSubsetValueTest.java) — **Port complete**: `testEnumerateN32`, `testEnumerateN33`, `testEnumerateN63`, `testEnumerateN64`, `testNormalization`, `testKSubsetFingerprintingS009`, `testKSubsetFingerprintingS032`, `testKSubsetFingerprintingS033`, `testKSubsetFingerprintingS063`, `testInvalidKDenotesEmptySet`, `testToStringLargeSwallowsCountError`.
  Go translation: [tlc/value_ksubset_java_test.go](value_ksubset_java_test.go).
  Preserves all eleven methods, all thirty fingerprint rows, the original enumeration/null/count assertions, both invalid-k operand directions, hashes and checked strings. Fixed production Size overflow to actual IllegalArgumentException.
- [x] [tlc2/value/impl/ModelValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/ModelValueTest.java) — **Port complete**: `testEqualsUntyped`, `testEqualsStringMV`, `testEqualsIntMV`, `testEqualsBoolMV`, `testEqualsIVMV`, `testEqualsRcdMV`, `testEqualsTupeMV`, `testCompareToUntyped`, `testCompareToStringMV`, `testCompareToIntMV`, `testCompareToBoolMV`, `testCompareToIVMV`, `testCompareToRcdMV`, `testCompareToTupeMV`, `testEqualsTyped`, `testEqualsTypedMVsUntyped`, `testEqualsTwoTypedMVs`, `testEqualsStringTypedMV`, `testEqualsTypedMVString`, `testEqualsIntTypedMV`, `testEqualsTypedMVInt`, `testEqualsBoolTypedMV`, `testEqualsTypedMVBool`, `testEqualsIVTypedMV`, `testEqualsTypedMVIV`, `testEqualsRcdTypedMV`, `testEqualsTypedMVRcd`, `testEqualsTupeTypedMV`, `testEqualsTypedMVTupe`, `testCompareToTyped`, `testCompareToTypedMVsUntyped`, `testCompareToTwoTypedMVs`, `testCompareToStringTypedMV`, `testCompareToTypedMVString`, `testCompareToIntTypedMV`, `testCompareToTypedMVInt`, `testCompareToBoolTypedMV`, `testCompareToTypedMVBool`, `testCompareToIVTypedMV`, `testCompareToTypedMVIV`, `testCompareToRcdTypedMV`, `testCompareToTypedMVRcd`, `testCompareToTupeTypedMV`, `testCompareToTypedMVTupe`.
  Go translation: [tlc/value_model_test.go](value_model_test.go).
  Preserves all 44 original methods in JUnit runner order, exact comparison results, operand directions, source constructor flags and expected TLCRuntimeException types.
- [x] [tlc2/value/impl/SetOfFcnsValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/SetOfFcnsValueTest.java) — **Port complete**: `testRangeSubsetValue`, `testDomainEmpty`, `testRangeEmpty`, `testDomainAndRangeEmpty`, `testRandomSubsetAndValueEnumerator`, `testDomainModelValue`, `testDomainIntervalRangeSetEnumValueSize9`, `testDomainIntervalRangeSetEnumValueSize27`, `testDomainIntervalRangeSetEnumValueSize256`, `testRandomSubsetFromReallyLarge`, `testEmptyNonEnumerableDomain`, `testUnitNonEnumerableRange`, `testUnitNonEnumerableRangeInterval`, `testNonEnumerableRange`, `testNonEnumerableRangeInterval`, `testRandomSubsetEmptyNonEnumerableDomain`.
  Go translation: [tlc/value_fcn_set_java_test.go](value_fcn_set_java_test.go).
  Preserves indexed enumeration, all literal expected functions, the full four-set/seven-sample huge-set matrix, Java hashing/equality and exact non-enumerable failures.
- [x] [tlc2/value/impl/SetOfRcrdValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/SetOfRcrdValueTest.java) — **Port complete**: `testSimple`, `testRangeSubsetValue`, `testRandomSubset`, `testRandomSubsetVaryingParameters`, `testRandomSubsetAstronomically`, `testEmptyNonEnumerableField`, `testRandomSubsetEmptyNonEnumerableField`.
  Go translation: [tlc/value_record_set_java_test.go](value_record_set_java_test.go).
  Preserves indexed SubsetEnumerator.elementAt, Java hash/equality, all 6,684 n/m/k samples (9,246,174 records), the astronomical k=10,000 case and empty/non-enumerable field checks.
- [x] [tlc2/value/impl/SubsetEnumeratorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/SubsetEnumeratorTest.java) — **Port complete**; parameterized: all 12 original rows: `testElementsInt`, `testGetRandomSubset`.
  Go translation: [tlc/value_subset_enumerator_java_test.go](value_subset_enumerator_java_test.go).
  Preserves both methods over all eleven fractions per row (264 cases), original ASCII-decimal model values, shared parameter objects, Math.ceil, class fingerprint setup, memberships and both HashSet constructions.
- [x] [tlc2/value/impl/SubsetValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/SubsetValueTest.java) — **Port complete**: `testRandomSubsetE7F1`, `testRandomSubsetE7F05`, `testRandomSubsetE6F1`, `testRandomSubsetE5F01`, `testRandomSubsetE5F025`, `testRandomSubsetE5F05`, `testRandomSubsetE5F075`, `testRandomSubsetE5F1`, `testRandomSubsetE32F1ENeg6`, `testRandomSubsetE17F1ENeg3`, `testRandomSubsetSubset16`, `testRandomSubsetSubset256`, `testRandomSubsetSubset65536`, `testRandomSubsetSubsetNoOverflow`, `testEmptyEnumerationsAreIndependent`, `testKSubsetEnumerator`, `testKSubsetEnumeratorNegative`, `testKSubsetEnumeratorGTCapacity`, `testNumKSubset`, `testNumKSubset2`, `testNumKSubsetNeg`, `testNumKSubsetKGTN`, `testNumKSubsetUpTo62`, `testNumKSubsetPreventsOverflow`, `testUnrankKSubsets`, `testUnrank16viaRank`, `testRandomSetOfSubsets`, `testRandomSetOfSubsets300`, `testRandomSetOfSubsets400`, `testElementsNormalizedIsNormalized`, `testKElementsAreNormalized`, `testKElementsMatchElementsNormalized`, `testRandomSubsetGeneratorK0`, `testRandomSubsetGeneratorKNegative`, `testRandomSubsetGeneratorKNplus1`, `testRandomSubsetGeneratorN10`, `testRandomSubsetGeneratorN100`.
  Go translation: [tlc/value_subset_java_test.go](value_subset_java_test.go).
  Preserves all 37 methods in verified JUnit order with shared seed 15041980, independently shuffled fixtures, TreeSet comparisons, both complete 65,536-subset sweeps, all binomial rows, source sample sizes and catch-only assertions. Fixed actual overflow exception families and exposed the normalized-enumeration entry point.
- [x] [tlc2/value/impl/TupleValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/TupleValueTest.java) — **Port complete**: `testErrorMessages`.
  Go translation: [tlc/value_tuple_java_test.go](value_tuple_java_test.go).
  Preserves all four source catch/message assertions, including Java's absence of fail() after the try blocks; added the missing production array-argument overload and source-aware assertion failures.

### Collections, buffered files, combinatorics, and statistics

Original utility cases, file operations/fuzz sequences, integer queues/stacks, vectors, contexts, combinatorics, and parameterized statistics.

- [x] [tlc2/util/BufferedRandomAccessFileFuzzTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/BufferedRandomAccessFileFuzzTest.java) — **Port complete**: `fuzz`, `testWellDefined`.
  Go translation: [tlc/buffered_random_access_file_fuzz_java_test.go](buffered_random_access_file_fuzz_java_test.go). Original 10,000 shared-counter traces of 50 operations, every processor worker/seed, eight operations, rejection sampling, independent native file oracle, defined-hole model, full-read smoothing, minimizer and exact well-definedness case are retained. This is the original correctness harness, not a shortened Go fuzz substitute.
- [x] [tlc2/util/BufferedRandomAccessFileTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/BufferedRandomAccessFileTest.java) — **Port complete**: `testWrite`, `testWriteSeek`, `testWriteSeekNoLength`, `testRead`, `testReadSeekNoLength`, `testInvalidateBufferedData`, `testReadAfterSeekPastEndOfFile`, `testWriteAfterSeekPastEndOfFile`, `testObscureSetLengthBehavior`, `testIdempotentClose`, `testIOExceptionOnUseAfterClose`, `regressionTest01`, `regressionTest02`, `regressionTest03`, `regressionTest04`, `regressionTest05`, `regressionTest06`, `regressionTest07`.
  Go translation: [tlc/buffered_random_access_file_java_test.go](buffered_random_access_file_java_test.go). All eighteen original methods, complete buffer-size loops, fifteen typed closed-handle catches and every operation/literal in all seven generated traces; no exception requirement was added to the two permissive seek/no-length catches.
- [x] [tlc2/util/ByteUtilsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/ByteUtilsTest.java) — **Port complete**: `test1`, `test2`, `test3`, `test4`, `test5`, `test6`.
  Go translation: [tlc/byte_utils_java_test.go](byte_utils_java_test.go). All six original methods, per-method setup, full 10,000-value exercises, 1000-bit Java-random BigInts, original two-array split, three append attempts and IOException-only catch. Original diagnostics and timing output are retained; print-only checks have not been turned into assertions.
- [x] [tlc2/util/CombinatoricsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/CombinatoricsTest.java) — **Port complete**: `testChoose`, `testChooseBigChoose`, `testSlowChooseBigChoose`, `testBigChoose50c1`, `testBigChoose50c10`, `testBigChoose50c20`, `testBigChoose50c30`, `testBigChoose400c1`, `testBigChoose400c50`, `testBigChoose400c100`, `testBigChoose400c200`.
  Go translation: [tlc/combinatorics_java_test.go](combinatorics_java_test.go). All eleven original methods, complete 62×62 / 63×63 / 185×185 matrices, and all eight literal bit-length and exact-long/decimal cases are retained.
- [x] [tlc2/util/ContextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/ContextTest.java) — **Port complete**: `testLookupEmpty`, `testLookupBranch`, `testLookupSymbolNodeNull`, `testLookup`, `testLookupCutOffFalse`, `testLookupCutOffTrue`, `testLookupWithAtBranching`, `testLookupWithCutOffFalseAtBranching`, `testLookupWithCutOffTrueAtBranching`, `testLookupSymbolNode`.
  Go translation: [tlc/context_java_test.go](context_java_test.go).
- [x] [tlc2/util/GrowingLongVecTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/GrowingLongVecTest.java) — **Port complete**: `testReadBeyondCapacity` (from `LongVecTest`), `testAddAndReadBeyondCapacity` (from `LongVecTest`), `testRemoveBeyondCapacity` (from `LongVecTest`), `testAddRemoveBeyondCapacity` (from `LongVecTest`), `testRemoveAndGet` (from `LongVecTest`), `testRemoveWrongOrder` (from `LongVecTest`), `testGetNegative` (from `LongVecTest`), `testRemoveNegative` (from `LongVecTest`), `testGrowAndShrink`.
  Go translation: [tlc/long_vec_java_test.go](long_vec_java_test.go), retaining the original constructor hook, explicit capacity-10 cases, all twelve remove-beyond attempts, exact index-exception catch family and inherited assertions.
- [x] [tlc2/util/LongVecTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/LongVecTest.java) — **Port complete**: `testReadBeyondCapacity`, `testAddAndReadBeyondCapacity`, `testRemoveBeyondCapacity`, `testAddRemoveBeyondCapacity`, `testRemoveAndGet`, `testRemoveWrongOrder`, `testGetNegative`, `testRemoveNegative`.
  Go translation: [tlc/long_vec_java_test.go](long_vec_java_test.go), retaining the original constructor hook, explicit capacity-10 cases, all twelve remove-beyond attempts, exact index-exception catch family and inherited assertions.
- [x] [tlc2/util/MemIntQueueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/MemIntQueueTest.java) — **Port complete**: `testDequeuePastLastElement`, `testEnqueueZeros`, `testEnqueueLong`, `testEnqueueDequeueLong`, `testGrow`.
  Go translation: [tlc/int_queue_java_test.go](int_queue_java_test.go). All five original methods, irrelevant constructor arguments, default/capacity-four construction, zero-long fixture, exact growth sequence and strict NoSuchElementException catches.
- [x] [tlc2/util/statistics/BucketStatisticsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/statistics/BucketStatisticsTest.java) — **Port complete**; all original parameter rows: `testInvalidArgument`, `testMean`, `testMedian`, `testMin`, `testMin2`, `testMax`, `testStandardDeviation`, `testGetPercentile`, `testGetPercentileNaN`, `testToString`.
  Go translation: [tlc/bucket_statistics_java_test.go](bucket_statistics_java_test.go). Every original method runs on both implementation rows with fresh source constructor arguments; exact Double.compare checks, typed catches, clamp calls, repeated percentile assertion and invocation-only string checks are retained.
  Related Go checks: [tlc/bucket_statistics_test.go](bucket_statistics_test.go).
- [x] [tlc2/util/statistics/FixedSizedBucketStatisticsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/statistics/FixedSizedBucketStatisticsTest.java) — **Port complete**; all original parameter rows: `testMin`, `testMin2`, `testMax`, `testInvalidArgument`, `testGetPercentileNaN`, `testMaximum`.
  Go translation: [tlc/bucket_statistics_java_test.go](bucket_statistics_java_test.go). Every original method runs on both implementation rows with fresh source constructor arguments; exact Double.compare checks, typed catches, clamp calls, repeated percentile assertion and invocation-only string checks are retained.
  Related Go checks: [tlc/bucket_statistics_test.go](bucket_statistics_test.go).

### Numbered legacy model suite

Every concrete ETest*, Test*, and TestInvalidInvariant in tool/suite/. Inherited SuiteTestCase.testSpec and subclass coverage/assertion hooks remain part of each original test.

- [x] [tlc2/tool/suite/ETest1.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest1.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_diagnostics_java_test.go](../tlc_legacy_suite_error_diagnostics_java_test.go).
- [x] [tlc2/tool/suite/ETest10.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest10.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_models_java_test.go](../tlc_legacy_suite_error_models_java_test.go).
- [x] [tlc2/tool/suite/ETest11.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest11.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_models_java_test.go](../tlc_legacy_suite_error_models_java_test.go).
- [x] [tlc2/tool/suite/ETest12.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest12.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_models_java_test.go](../tlc_legacy_suite_error_models_java_test.go).
- [x] [tlc2/tool/suite/ETest13.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest13.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_models_java_test.go](../tlc_legacy_suite_error_models_java_test.go).
- [x] [tlc2/tool/suite/ETest14.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest14.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_models_java_test.go](../tlc_legacy_suite_error_models_java_test.go).
- [x] [tlc2/tool/suite/ETest15.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest15.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_models_java_test.go](../tlc_legacy_suite_error_models_java_test.go).
- [x] [tlc2/tool/suite/ETest16.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest16.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_diagnostics_java_test.go](../tlc_legacy_suite_error_diagnostics_java_test.go).
- [x] [tlc2/tool/suite/ETest2.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest2.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_diagnostics_java_test.go](../tlc_legacy_suite_error_diagnostics_java_test.go).
- [x] [tlc2/tool/suite/ETest3.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest3.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_models_java_test.go](../tlc_legacy_suite_error_models_java_test.go).
- [x] [tlc2/tool/suite/ETest4.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest4.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_models_java_test.go](../tlc_legacy_suite_error_models_java_test.go).
- [x] [tlc2/tool/suite/ETest5.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest5.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_diagnostics_java_test.go](../tlc_legacy_suite_error_diagnostics_java_test.go).
- [x] [tlc2/tool/suite/ETest6.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest6.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_models_java_test.go](../tlc_legacy_suite_error_models_java_test.go).
- [x] [tlc2/tool/suite/ETest7.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest7.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_diagnostics_java_test.go](../tlc_legacy_suite_error_diagnostics_java_test.go).
- [x] [tlc2/tool/suite/ETest8.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest8.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_models_java_test.go](../tlc_legacy_suite_error_models_java_test.go).
- [x] [tlc2/tool/suite/ETest9.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest9.java) — **Port complete**: whole original `testSpec`, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_models_java_test.go](../tlc_legacy_suite_error_models_java_test.go).
- [x] [tlc2/tool/suite/Test1.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test1.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test10.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test10.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test11.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test11.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test12.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test12.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test13.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test13.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test14.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test14.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test15.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test15.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test16.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test16.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test17.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test17.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test18.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test18.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test19.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test19.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_eight_java_test.go](../tlc_legacy_suite_next_eight_java_test.go).
- [x] [tlc2/tool/suite/Test2.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test2.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test20.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test20.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_eight_java_test.go](../tlc_legacy_suite_next_eight_java_test.go).
- [x] [tlc2/tool/suite/Test201.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test201.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_recursion_and_selectors_java_test.go](../tlc_legacy_suite_recursion_and_selectors_java_test.go).
- [x] [tlc2/tool/suite/Test202.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test202.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_recursion_and_selectors_java_test.go](../tlc_legacy_suite_recursion_and_selectors_java_test.go).
- [x] [tlc2/tool/suite/Test203.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test203.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_recursion_and_selectors_java_test.go](../tlc_legacy_suite_recursion_and_selectors_java_test.go).
- [x] [tlc2/tool/suite/Test204.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test204.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_recursion_and_selectors_java_test.go](../tlc_legacy_suite_recursion_and_selectors_java_test.go).
- [x] [tlc2/tool/suite/Test205.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test205.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_recursion_and_selectors_java_test.go](../tlc_legacy_suite_recursion_and_selectors_java_test.go).
- [x] [tlc2/tool/suite/Test206.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test206.java) — **Port complete**: whole inherited `testSpec` and original constructor/settings in [tlc_legacy_suite_subexpression_java_test.go](../tlc_legacy_suite_subexpression_java_test.go); full original selector/lambda/INSTANCE model and all assertions retained.
- [x] [tlc2/tool/suite/Test207.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test207.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_proof_and_selectors_java_test.go](../tlc_legacy_suite_proof_and_selectors_java_test.go).
- [x] [tlc2/tool/suite/Test208.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test208.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_proof_and_selectors_java_test.go](../tlc_legacy_suite_proof_and_selectors_java_test.go).
- [x] [tlc2/tool/suite/Test209.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test209.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_proof_and_selectors_java_test.go](../tlc_legacy_suite_proof_and_selectors_java_test.go).
- [x] [tlc2/tool/suite/Test21.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test21.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_eight_java_test.go](../tlc_legacy_suite_next_eight_java_test.go).
- [x] [tlc2/tool/suite/Test210.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test210.java) — **Port complete**: whole original `testSpec`, all ten exact output substrings, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_diagnostics_java_test.go](../tlc_legacy_suite_error_diagnostics_java_test.go).
- [x] [tlc2/tool/suite/Test212.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test212.java) — **Port complete**: whole original `testSpec`, all seven exact output substrings, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_diagnostics_java_test.go](../tlc_legacy_suite_error_diagnostics_java_test.go). Four byte-identical vectors retained. Production tracks formal argument dependencies through imported operators, LAMBDA, INSTANCE substitutions and recursive definitions rather than treating any prime in a body as non-Leibniz.
- [x] [tlc2/tool/suite/Test213.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test213.java) — **Port complete**: whole original `testSpec`, all six exact output substrings, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_diagnostics_java_test.go](../tlc_legacy_suite_error_diagnostics_java_test.go). Three byte-identical vectors retained. Production preserves temporal goals through CASE/QED subproofs, computes actual definition/argument levels, and carries ASSUME/PROVE constant constraints into explicit and implicit INSTANCE substitutions.
- [x] [tlc2/tool/suite/Test214.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test214.java) — **Port complete**: whole original `testSpec`, both exact output substrings, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_diagnostics_java_test.go](../tlc_legacy_suite_error_diagnostics_java_test.go). Two byte-identical vectors retained. Production retains complete USE/HIDE facts and applies Java’s semantic-node-kind restriction to HIDE, including proof steps and imported theorem/assumption names.
- [x] [tlc2/tool/suite/Test215.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test215.java) — **Port complete**: whole original `testSpec`, all nine exact output substrings, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_diagnostics_java_test.go](../tlc_legacy_suite_error_diagnostics_java_test.go). Two byte-identical vectors retained. Production resolves argument levels through definitions and lexical LET scopes, preserves Java’s action-wrapper node-kind checks, and reports whole-expression or individual-bound ranges with the exact Java messages. Recursive level checks retain Java’s per-iteration definition and expression caches.
- [x] [tlc2/tool/suite/Test216.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test216.java) — **Port complete**: whole original overriding `testSpec` and original constructor/settings in [tlc_legacy_suite_lifecycle_java_test.go](../tlc_legacy_suite_lifecycle_java_test.go).
- [x] [tlc2/tool/suite/Test217.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test217.java) — **Port complete**: whole original `testSpec`, all three exact output substrings, original constructor and SuiteETestCase settings in [tlc_legacy_suite_error_diagnostics_java_test.go](../tlc_legacy_suite_error_diagnostics_java_test.go). Three byte-identical vectors retained. Production propagates formal argument maximum levels through definitions, INSTANCE substitutions and builtin argument constraints, reporting the complete application range and Java diagnostic.
- [x] [tlc2/tool/suite/Test219.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test219.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_proof_and_selectors_java_test.go](../tlc_legacy_suite_proof_and_selectors_java_test.go).
- [x] [tlc2/tool/suite/Test22.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test22.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_eight_java_test.go](../tlc_legacy_suite_next_eight_java_test.go).
- [x] [tlc2/tool/suite/Test220.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test220.java) — **Port complete**: whole original overriding `testSpec` and original constructor/settings in [tlc_legacy_suite_lifecycle_java_test.go](../tlc_legacy_suite_lifecycle_java_test.go).
- [x] [tlc2/tool/suite/Test23.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test23.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_eight_java_test.go](../tlc_legacy_suite_next_eight_java_test.go).
- [x] [tlc2/tool/suite/Test24.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test24.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_eight_java_test.go](../tlc_legacy_suite_next_eight_java_test.go).
- [x] [tlc2/tool/suite/Test25.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test25.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_eight_java_test.go](../tlc_legacy_suite_next_eight_java_test.go).
- [x] [tlc2/tool/suite/Test26.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test26.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_eight_java_test.go](../tlc_legacy_suite_next_eight_java_test.go).
- [x] [tlc2/tool/suite/Test27.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test27.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_seven_java_test.go](../tlc_legacy_suite_next_seven_java_test.go).
- [x] [tlc2/tool/suite/Test28.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test28.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_seven_java_test.go](../tlc_legacy_suite_next_seven_java_test.go).
- [x] [tlc2/tool/suite/Test29.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test29.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_seven_java_test.go](../tlc_legacy_suite_next_seven_java_test.go).
- [x] [tlc2/tool/suite/Test3.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test3.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test30.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test30.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_seven_java_test.go](../tlc_legacy_suite_next_seven_java_test.go).
- [x] [tlc2/tool/suite/Test31.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test31.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_seven_java_test.go](../tlc_legacy_suite_next_seven_java_test.go).
- [x] [tlc2/tool/suite/Test32.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test32.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_seven_java_test.go](../tlc_legacy_suite_next_seven_java_test.go).
- [x] [tlc2/tool/suite/Test33.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test33.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_next_seven_java_test.go](../tlc_legacy_suite_next_seven_java_test.go).
- [x] [tlc2/tool/suite/Test34.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test34.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test35.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test35.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test36.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test36.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test37.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test37.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test38.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test38.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test39.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test39.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test4.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test4.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test40.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test40.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test41.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test41.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test42.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test42.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test43.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test43.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test44.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test44.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test45.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test45.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test46.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test46.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test47.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test47.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test48.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test48.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test49.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test49.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_modules_operators_java_test.go](../tlc_legacy_suite_modules_operators_java_test.go).
- [x] [tlc2/tool/suite/Test5.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test5.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test50.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test50.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_coverage_operators_java_test.go](../tlc_legacy_suite_coverage_operators_java_test.go).
- [x] [tlc2/tool/suite/Test51.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test51.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_coverage_operators_java_test.go](../tlc_legacy_suite_coverage_operators_java_test.go).
- [x] [tlc2/tool/suite/Test52.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test52.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_coverage_operators_java_test.go](../tlc_legacy_suite_coverage_operators_java_test.go).
- [x] [tlc2/tool/suite/Test53.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test53.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_coverage_operators_java_test.go](../tlc_legacy_suite_coverage_operators_java_test.go).
- [x] [tlc2/tool/suite/Test54.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test54.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_coverage_operators_java_test.go](../tlc_legacy_suite_coverage_operators_java_test.go).
- [x] [tlc2/tool/suite/Test55.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test55.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_coverage_operators_java_test.go](../tlc_legacy_suite_coverage_operators_java_test.go).
- [x] [tlc2/tool/suite/Test56.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test56.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_coverage_operators_java_test.go](../tlc_legacy_suite_coverage_operators_java_test.go).
- [x] [tlc2/tool/suite/Test57.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test57.java) — **Port complete**: whole original overridden `testSpec` and safety exit in [tlc_legacy_suite_instance_and_coverage_java_test.go](../tlc_legacy_suite_instance_and_coverage_java_test.go).
- [x] [tlc2/tool/suite/Test58.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test58.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_instance_and_coverage_java_test.go](../tlc_legacy_suite_instance_and_coverage_java_test.go).
- [x] [tlc2/tool/suite/Test59.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test59.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_instance_and_coverage_java_test.go](../tlc_legacy_suite_instance_and_coverage_java_test.go).
- [x] [tlc2/tool/suite/Test6.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test6.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test60.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test60.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_instance_and_coverage_java_test.go](../tlc_legacy_suite_instance_and_coverage_java_test.go).
- [x] [tlc2/tool/suite/Test62.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test62.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_instance_and_coverage_java_test.go](../tlc_legacy_suite_instance_and_coverage_java_test.go).
- [x] [tlc2/tool/suite/Test63.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test63.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_instance_and_coverage_java_test.go](../tlc_legacy_suite_instance_and_coverage_java_test.go).
- [x] [tlc2/tool/suite/Test63a.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test63a.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_instance_and_coverage_java_test.go](../tlc_legacy_suite_instance_and_coverage_java_test.go).
- [x] [tlc2/tool/suite/Test64.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test64.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_instance_and_coverage_java_test.go](../tlc_legacy_suite_instance_and_coverage_java_test.go).
- [x] [tlc2/tool/suite/Test64a.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test64a.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_instance_and_coverage_java_test.go](../tlc_legacy_suite_instance_and_coverage_java_test.go).
- [x] [tlc2/tool/suite/Test65.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test65.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_instance_and_coverage_java_test.go](../tlc_legacy_suite_instance_and_coverage_java_test.go).
- [x] [tlc2/tool/suite/Test65a.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test65a.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_instance_and_coverage_java_test.go](../tlc_legacy_suite_instance_and_coverage_java_test.go).
- [x] [tlc2/tool/suite/Test7.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test7.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test8.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test8.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test9.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test9.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_first_eighteen_java_test.go](../tlc_legacy_suite_first_eighteen_java_test.go).
- [x] [tlc2/tool/suite/Test99.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test99.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_misc_models_java_test.go](../tlc_legacy_suite_misc_models_java_test.go).
- [x] [tlc2/tool/suite/Test999.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test999.java) — **Port complete**: whole inherited `testSpec` (from `SuiteTestCase`) and original constructor/settings in [tlc_legacy_suite_misc_models_java_test.go](../tlc_legacy_suite_misc_models_java_test.go).
- [x] [tlc2/tool/suite/TestInvalidInvariant.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/TestInvalidInvariant.java) — **Port complete**: whole original overriding `testSpec` and original constructor/settings in [tlc_legacy_suite_lifecycle_java_test.go](../tlc_legacy_suite_lifecycle_java_test.go).

## Upstream ignored tests in the main suite

These do not contribute to the pending non-`@Ignore` total. Preserve the upstream ignore status and reason when translating; do not treat them as passed or silently enable them. They remain part of the original-source inventory.

- [tlc2/tool/ActionCompositionDistributiveATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionCompositionDistributiveATest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/ActionCompositionDistributiveBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionCompositionDistributiveBTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/AliasLivenessLassoTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AliasLivenessLassoTest_TTraceTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/AliasLivenessStutteringTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AliasLivenessStutteringTest_TTraceTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/AliasSafetySimuTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AliasSafetySimuTest_TTraceTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/AliasSafetyTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AliasSafetyTest_TTraceTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/DumpLoadTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DumpLoadTraceTest.java) — `testLivenessEWD840MC3DumpLoadTraceJSON`, `testSafetyDieHardAliasSubDumpLoadTraceJSONAutoWorkers`, `testSafetyDieHardAliasSubDumpLoadTraceTLCAutoWorkers` — all these methods already have matching Go skips in [tlc_dump_load_trace_java_test.go](../tlc_dump_load_trace_java_test.go).
- [tlc2/tool/EvalExceptionLivenessTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EvalExceptionLivenessTest_TTraceTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/Github1198eTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198eTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/Github1198gTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198gTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/Github297Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github297Test.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/IncompleteNextMultipleActionsTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/IncompleteNextMultipleActionsTest_TTraceTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/IncompleteNextTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/IncompleteNextTest_TTraceTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/TLCGetNonDeterminismTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCGetNonDeterminismTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/fp/OffHeapIndexerEquivalenceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapIndexerEquivalenceTest.java) — `testInfiniteBitshifting`, `testInfMultBitshifting` — no complete translation identified.
- [tlc2/tool/liveness/April20aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/April20aTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/April20bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/April20bTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/April21Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/April21Test.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/April22Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/April22Test.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/April25Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/April25Test.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/April29Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/April29Test.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/April29dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/April29dTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/ChooseTableauSymmetryTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ChooseTableauSymmetryTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/Github607Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github607Test.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/May09Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/May09Test.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/May09dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/May09dTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/NQTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/NQTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/NQaTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/NQaTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/OneBitMutexTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/OneBitMutexTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/SymmetryModelCheckerTest3.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/SymmetryModelCheckerTest3.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/SymmetryModelCheckerTest3a.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/SymmetryModelCheckerTest3a.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/SymmetryModelCheckerTestLong.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/SymmetryModelCheckerTestLong.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/SymmetryModelCheckerTestLonga.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/SymmetryModelCheckerTestLonga.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/SymmetryTableauModelCheckerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/SymmetryTableauModelCheckerTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/liveness/TableauSymmetryTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TableauSymmetryTest.java) — `testSpec` — no complete translation identified.
- [tlc2/tool/suite/Test61.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test61.java) — `testSpec` (from `SuiteTestCase`) — no complete translation identified.

## Supplementary original suites

These are outside the 1,269-method main-suite denominator above and must not disappear from the port scope merely because they live in another directory. Original Ant selection/exclusions are in [customBuild.xml](../../tlaplus/tlatools/org.lamport.tlatools/customBuild.xml).

### Shared utilities under test/util

Shared TLC/SANY file resolution, buffered input, runtime, and string helpers; telemetry preferences are listed separately from forbidden email reporting.

7 classes / 56 logical methods; 55 confirmed translations.

- [x] [util/BufferedDataInputStreamTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/BufferedDataInputStreamTest.java) — **Port complete**: `testReadStringThrowsOnShortStreamAtHalfBoundary`, `testReadStringThrowsOnShortStreamBelowHalfBoundary`, `testReadStringExactLength`, `testReadStringAcrossBufferRefill`, `testConstructorRejectsStreamReturningZero`, `testEmptyStream`, `testWriteReadStringRoundTrip` Whole seven-method class → [buffered_data_stream_java_test.go](buffered_data_stream_java_test.go); preserve exact short streams, typed EOF and TLCRuntimeException catches/error code, all 8,292 refill bytes/8,190-byte advance and exact result, initial empty-stream EOF, original string length/payload/sentinel round trip. Production input/output classes preserve eager 8,192-byte refills, source binary/string encodings, zero-read assertion, EOF state, line/skip/open/close methods and UTF-16 byte conversion. ValueInputStream/ValueOutputStream now use these original buffered classes rather than bypassing them.
- [x] [util/ExecutionStatisticsCollectorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/ExecutionStatisticsCollectorTest.java) — **Port complete**: `testCompanyLevelNoFile`, `testCompanyLevelUnreadable`, `testCompanyLevelEmptyFile`, `testCompanyLevelNoESCFile`, `testCompanyLevelRandomIdFile`, `testCompanyLevelUserDefinedIdFile`, `testNoFile`, `testUnreadableFile`, `testEmptyFile`, `testNoESCFile`, `testRandomIdFile`, `testUserDefinedIdFile` Whole twelve-method class → [execution_statistics_collector_java_test.go](execution_statistics_collector_java_test.go); retain all preference files, original unreadable-file assumption, localhost canonical routing, reserved fully qualified invalid hostname, exact identifier truncation and UUID equality/inequality/suffix assertions. The original submit subclass captures all submissions. Production preserves source identifier/property precedence, default-charset first-line decoding, opt-out/routing and UUIDv1-to-v4 fallback; tests require normal host interface access, just as the unchanged Java methods do.
- [x] [util/FileUtilTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/FileUtilTest.java) — `testReturnFromCheckpoint`, `testDuplicateStateDirCreation`, `testUseDifferentMetaDir` — mapped in [tlc/file_util_test.go](file_util_test.go).
- [x] [util/MonolithSpecExtractorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/MonolithSpecExtractorTest.java) — **Port complete**: `testExtractConfig`, `testExtractModule`, `testConfigWithWindowsPathAsName`, `testModuleWithWindowsPathAsName`, `testGetConfig` Whole five-method class → [monolith_spec_extractor_java_test.go](monolith_spec_extractor_java_test.go); preserve exact Windows-path-prefixed fixture, config result, module-name and complete module text assertions, literal Windows names/empty config/null module and both getConfig cases. Production shares exact quoted-name/ASCII-whitespace marker extraction with SANY and config loading, preserves Java trim/end-marker rules, and provides real file-backed NamedInputStream metadata and temporary-file registration.
- [x] [util/SimpleFilenameToStreamTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/SimpleFilenameToStreamTest.java) — `testResolveStandardModules`, `testResolveCommunityModules`, `testResolveByAbsolutePath`, `testResolveFromUserDir`, `testResolveWithCustomLibraryPath`, `testTLALibrarySystemProperty`, `testWindowsTLAFileCreation`, `testBizarreWorkingDirectorySearchBehavior` — mapped in [filename_to_stream_test.go](../filename_to_stream_test.go).
- [x] [util/StringHelperTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/StringHelperTest.java) — **Port complete**: `testGetWords`, `testGetWordsLeadingSpace`, `testGetWordsNull`, `testCopyString0`, `testCopyStringPos1`, `testCopyStringPos5`, `testCopyStringNeg1`, `testCopyStringNeg5`, `testOnlySpaces`, `testOnlySpacesNull`, `testTrimFront`, `testTrimFrontNull`, `testTrimFrontWhitespaces`, `testTrimEnd`, `testTrimEndNull`, `testTrimEndWhitespaces`, `testLeadingSpace`, `testLeadingSpacesNull`, `testIsIdentifier`, `testIsIdentifierNull` Whole 20-method JUnit 3 class → [string_helper_java_test.go](string_helper_java_test.go); all original word arrays, copy counts, space/trim/leading-space/identifier inputs and six strict NullPointerException cases retained. Production preserves Java UTF-16 char iteration, distinct trim/Character-whitespace/regex-whitespace rules, digit classifications, original copy-doubling algorithm and source null concatenation. Module closing tags now use that copy routine and preserve nonpositive widths; all source-commented TODO tests remain inactive.
- [ ] [util/TLCRuntimeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/TLCRuntimeTest.java) — **Reconcile**: `testIsThroughputOptimized` asserts JVM ParallelGC selected by the original Ant `-XX:+UseParallelGC` flag. Go concurrent GC does not satisfy that JVM-specific assertion; keep visible and uncredited rather than inventing a constant-success runtime response.

### Long-running fingerprint and queue tests

Four concrete classes / 22 inherited-or-local method contexts; 18 confirmed translations and 4 pending contexts. DiskStateQueueTest is the only complete concrete class. Upstream `test-dist-long` excludes DiskFPSetTest, MSBDiskFPSetTest and DiskStateQueueTest for runtime; OffHeapDiskFPSetLongTest remains selected. FPSetTest supplies the inherited fingerprint bodies.

The complete LSB and MSB `testMaxFPSetSizeRnd` methods are translated in [long_heap_fpset_stress_java_test.go](long_heap_fpset_stress_java_test.go). Run them explicitly with `go test -tags=tlc_fp_stress -run '^TestJavaLong(LSB|MSB)DiskFPSet_testMaxFPSetSizeRnd$' -timeout=0 -v ./tlc`, without race instrumentation. Their factories, default configuration, seed, all 2,147,483,648 iterations, predecessor and size assertions, checkpoints and invariant checks are retained. Translation credit does not imply full current-snapshot execution credit: LSB's complete draft passed in 20,048.87 seconds (`32959`, terminal 0); fingerprint production is unchanged since that run. MSB's older draft passed at `a915e08` in 13,343.32 seconds (`63113`, terminal 0), before flusher and file-helper fixes. Its complete compiled full run (`5144`, terminal 0) now passes in 13,063.40 seconds, including the final checkpoint, invariant and size assertions; fingerprint production is unchanged since its `132a77f` build. See HANDOFF.md for logs and snapshot qualifications. Earlier incomplete or ENOSPC runs remain uncredited.

Source `testMaxFPSetSize` starts at `-(l/2)` and ends at `l`, where `l = 2,147,483,649`; retain all 3,221,225,473 requested iterations. Static source audit finds a repeated value 1,073,741,824 at indices 1,073,741,824 and 1,073,741,825, conflicting with the per-insertion `assertFalse(put)` if reached. The final assertion also expects l rather than the larger loop span. These are source-derived observations, not full Java execution results; the original assertions and bounds remain pending.

- [ ] [tlc2/tool/fp/DiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-long/tlc2/tool/fp/DiskFPSetTest.java) — **Partial**: inherited `testSimpleFill` **Port complete** in [long_heap_fpset_java_test.go](long_heap_fpset_java_test.go), `TestJavaLongLSBDiskFPSet/testSimpleFill`, with the original concrete factory, default configuration, four put/contains pairs and initialization. `testMaxFPSetSizeRnd` **Port complete** in [long_heap_fpset_stress_java_test.go](long_heap_fpset_stress_java_test.go), with the full inherited body and concrete factory. `testMaxFPSetSize` remains pending with its complete original bounds and assertions.
- [ ] [tlc2/tool/fp/MSBDiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-long/tlc2/tool/fp/MSBDiskFPSetTest.java) — **Partial**: inherited `testSimpleFill` **Port complete** in [long_heap_fpset_java_test.go](long_heap_fpset_java_test.go), `TestJavaLongMSBDiskFPSet/testSimpleFill`, with the original concrete factory, default configuration, four put/contains pairs and initialization. `testMaxFPSetSizeRnd` **Port complete** in [long_heap_fpset_stress_java_test.go](long_heap_fpset_stress_java_test.go), with the full inherited body and concrete factory. `testMaxFPSetSize` remains pending with its complete original bounds and assertions.
- [ ] [tlc2/tool/fp/OffHeapDiskFPSetLongTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-long/tlc2/tool/fp/OffHeapDiskFPSetLongTest.java) — **Partial**: `testSimpleFill` (from `FPSetTest`), `testCollisionBucket`, and `testPosition` **Port complete** in [offheap_long_fpset_java_test.go](offheap_long_fpset_java_test.go), preserving the ratio-1.0 factory that ignores its config argument, default 64 MiB nonheap budget, exact four fingerprints, complete bucket-capacity range, MAX_VALUE and 1 assertions. `testMaxFPSetSizeRnd` (from `FPSetTest`) has its complete body in [offheap_long_fpset_stress_java_test.go](offheap_long_fpset_stress_java_test.go), using the unchanged shared random helper and the original config-ignoring ratio-1.0 factory. Its full 2,147,483,648-iteration normal run is active in session 77502; full execution and method credit remain pending. Poll that handle and `.codex-gotmp/offheap-random-full.log` before starting another copy. `testMaxFPSetSize` remains pending with all 3,221,225,473 original sequential iterations and assertions. `testMultipleFlushes` **Port complete** in [offheap_multiple_flushes_java_test.go](offheap_multiple_flushes_java_test.go), behind `tlc_fp_stress`. Retains all four rounds, the runtime-derived bound (default 8,388,608 insertions per round), Java seed 15041980, ratio-1.0 factory, one reader and all insertion/exact-count invariant assertions. The user-authorized closed-flusher correction covers direct invariant flushing after a successful eviction; failed flushes retain their existing failure state. The full default workload passes with 33,554,432 fingerprints. Unchanged pinned Java previously failed with RejectedExecutionException; see [JAVA_BUG_FOUND.md](../JAVA_BUG_FOUND.md) and PORT_PROGRESS.md for the documented divergence and receipts. No reduced bounds, weakened assertions or persistent skip.
- [x] [tlc2/tool/queue/DiskStateQueueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-long/tlc2/tool/queue/DiskStateQueueTest.java) — **Port complete**: all nine inherited `StateQueueTest` methods and `testGrowBeyondIntMaxValue` in [long_disk_state_queue_java_test.go](long_disk_state_queue_java_test.go). Preserve all 2,147,483,648 enqueues of one DummyTLCState-equivalent object, its exact base header (worker Short.MAX_VALUE, uid 0, level 1), source empty-variable context and final size assertion. All inherited bodies, identity/nil/size/invalid-count assertions and typed exception catches remain unchanged. The complete unchanged Java class passes all ten methods; the complete Go class passes with the full growth loop. No smaller bound, alternate queue implementation, invented test or skip. Growth requires approximately 14 GiB of pool storage; the original Ant exclusion remains documented above.

### Concurrent fingerprint stress tests

Four concrete classes / 17 inherited-or-local method contexts; 2 confirmed method translations and 15 pending contexts; no complete suite translations identified. Preserve generator implementations, worker coordination, partitioning, random seeds, and sizes rather than replacing them with a small -race smoke test.

The original `MultiThreadedFPSetTest.numThreads` and `.insertions` properties also expose a source coordination problem: unchanged Java's three random methods in all three concrete factories fail the upper bound with two workers and 20,000 insertions. The producer loop can keep adding while another producer has not reached its per-thread quota; batching adds a further 1,024-entry overshoot. The exact Go draft preserves the same upper-bound assertion and fails it too. With the source-supported one-worker, 20,480-insertion configuration, all nine unchanged Java methods and all nine native race methods pass, including invariant checks. These are generator/property audits, **not verification or translation credit for the default 2,147,483,649-entry contexts**; the default bounds remain intact in the staged draft. The audit found and corrected Go counter races: Java LongAdder table counts and disk-count publication now use native atomic reads/writes, and `Size` retains the source independent counter reads instead of acquiring every table lock. Heap table-count increments also run independently of the global metadata mutex, as Java LongAdder increments do; ordinary Go bucket-capacity metadata keeps its lock. A second port shortcut was removed: OffHeap puts/contains no longer serialize on a set-wide mutex, and ordinary operations check the atomic eviction flag before entering the shared barrier. CAS insertion, compare-and-set eviction selection/reset, and the original worker-count assertion now match Java. The unchanged Java two-worker scalar method with its existing insertion property set to 8,388,608 completes eviction and then fails the same upper-size assertion as the Go draft; the native race run reports no races. Full default stress verification remains pending.


- [ ] [tlc2/tool/fp/ConcurrentWriteTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-concurrent/tlc2/tool/fp/ConcurrentWriteTest.java) — **Partial**: `test` and `test3` **Port complete** in [concurrent_write_java_test.go](concurrent_write_java_test.go). Preserve all 4,000 sequential values, eight independent handles, all 400,000,000 concurrent writes and all 400,000,000 original read assertions; source invokeAll future exceptions remain captured and uninspected. `test1` and `test2` **Reconcile**: unchanged Java and exact Go drafts both fail at value 1,000,000, which reads as zero after overlapping buffer pages flush stale bytes. Their original four-million-value loops/assertions remain in ignored scratch, uncredited. `test4` **Reconcile**: source positional writes add both partition position and global index offset, leaving gaps; unchanged full Java method fails at read 50,000,000 (expected 50,000,000, actual zero). No weakened assertions, smaller bounds, new skips or divergent production changes.
- [ ] [tlc2/tool/fp/MultiThreadedLSBDiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-concurrent/tlc2/tool/fp/MultiThreadedLSBDiskFPSetTest.java) — `testMaxFPSetSizeRndBatched`, `testMaxFPSetSizeRndBlock`, and `testMaxFPSetSizeRnd` remain pending with full original bounds. **Reconcile**: inherited `testMaxFPSetSizePartitioned` completes unchanged Java and faithful Go draft with default insertion property 2,147,483,649 and 48 workers at explicit Java `-Xmx256m` / Go `GOMEMLIMIT=256MiB`; both produce 25,165,823 distinct puts, identical worker counts and zero collisions, then fail the original minimum-size assertion. This generator traverses the memory-derived bucket partition rather than INSERTIONS. Explicit memory-limit audit is not default-environment/full-size verification or translation credit. Reader lifetime now matches Java: one reader per disk search, no pool mutex for indexed workers, atomic publication of the native reader snapshot. Original assertions remain intact; draft stays ignored scratch and uncredited.
- [ ] [tlc2/tool/fp/MultiThreadedMSBDiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-concurrent/tlc2/tool/fp/MultiThreadedMSBDiskFPSetTest.java) — `testMaxFPSetSizeRndBatched`, `testMaxFPSetSizeRndBlock`, and `testMaxFPSetSizeRnd` remain pending with full original bounds. **Reconcile**: inherited `testMaxFPSetSizePartitioned` completes unchanged Java and faithful Go draft with default insertion property 2,147,483,649 and 48 workers at explicit Java `-Xmx256m` / Go `GOMEMLIMIT=256MiB`; both produce 25,165,823 distinct puts, identical worker counts and zero collisions, then fail the original minimum-size assertion. This generator traverses the memory-derived bucket partition rather than INSERTIONS. Explicit memory-limit audit is not default-environment/full-size verification or translation credit. Reader lifetime now matches Java: one reader per disk search, no pool mutex for indexed workers, atomic publication of the native reader snapshot. Original assertions remain intact; draft stays ignored scratch and uncredited.
- [ ] [tlc2/tool/fp/MultiThreadedOffHeapDiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-concurrent/tlc2/tool/fp/MultiThreadedOffHeapDiskFPSetTest.java) — `testMaxFPSetSizeRndBatched`, `testMaxFPSetSizeRndBlock`, and `testMaxFPSetSizeRnd` remain pending with their original full bounds. **Reconcile**: `testMaxFPSetSizePartitioned` (local override plus inherited `doTest`) fails unchanged pinned Java at its minimum-size assertion: the original default 64 MiB configuration and 48 workers produce 8,388,575 distinct puts, then the base demands size >= 2,147,483,649 even though this generator explicitly ignores INSERTIONS. The default-property Go draft produces identical counts for all 48 workers and fails the same assertion. Neither the subsequent invariant assertion nor the local zero-bucket-capacity assertion is reached. Full source bucket traversal retained; no smaller workload, weakened assertions, new skips or behavioral workaround. Draft remains ignored scratch and uncredited.

### JPF concurrency verification

Four classes / four methods; JVM/Java PathFinder verification rather than portable JUnit assertions. Port or specify an equivalent Go verification approach deliberately; retain original queue/wakeup/listener models and short-stream failure property. Do not emulate a JVM to satisfy this checklist.

- [ ] [tlc2/tool/fp/OffHeapDiskFPSetJPFTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-verify/tlc2/tool/fp/OffHeapDiskFPSetJPFTest.java) — `test` — pending original-method translation/reconciliation.
- [ ] [tlc2/tool/queue/DiskStateQueueJPFTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-verify/tlc2/tool/queue/DiskStateQueueJPFTest.java) — `testDeadlockFreedom` — pending original-method translation/reconciliation.
- [ ] [tlc2/tool/queue/StateQueueJPFTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-verify/tlc2/tool/queue/StateQueueJPFTest.java) — `test` — pending original-method translation/reconciliation.
- [ ] [util/BufferedDataInputStreamJPFTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-verify/util/BufferedDataInputStreamJPFTest.java) — `testReadStringMustThrowOnShortStream` — pending original-method translation/reconciliation.

### Benchmarks and supporting fixtures

The following original JMH benchmark classes have no identified complete Go benchmark translations. They are performance work, outside the correctness-test percentages. Itemized here so the scope is explicit:

- [ ] [test-benchmark/tlc2/tool/ModuleOverwritesBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/tool/ModuleOverwritesBenchmark.java).
- [ ] [test-benchmark/tlc2/tool/fp/LongArrayBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/tool/fp/LongArrayBenchmark.java).
- [ ] [test-benchmark/tlc2/tool/fp/LongArrayInitializeBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/tool/fp/LongArrayInitializeBenchmark.java).
- [ ] [test-benchmark/tlc2/tool/fp/OffHeapIndexerBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/tool/fp/OffHeapIndexerBenchmark.java).
- [ ] [test-benchmark/tlc2/tool/queue/DiskQueueBenachmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/tool/queue/DiskQueueBenachmark.java).
- [ ] [test-benchmark/tlc2/tool/queue/StateQueueBenachmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/tool/queue/StateQueueBenachmark.java).
- [ ] [test-benchmark/tlc2/tool/simulation/SimulatorBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/tool/simulation/SimulatorBenchmark.java).
- [ ] [test-benchmark/tlc2/util/CombinatoricsBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/util/CombinatoricsBenchmark.java).
- [ ] [test-benchmark/tlc2/util/FP64Benchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/util/FP64Benchmark.java).
- [ ] [test-benchmark/tlc2/value/impl/EnumerateSubsetBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/value/impl/EnumerateSubsetBenchmark.java).
- [ ] [test-benchmark/tlc2/value/impl/FcnRcdBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/value/impl/FcnRcdBenchmark.java).
- [ ] [test-benchmark/tlc2/value/impl/IntervalValueBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/value/impl/IntervalValueBenchmark.java).
- [ ] [test-benchmark/tlc2/value/impl/RandomizationBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/value/impl/RandomizationBenchmark.java).
- [ ] [test-benchmark/tlc2/value/impl/SetEnumValueBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/value/impl/SetEnumValueBenchmark.java).
- [ ] [test-benchmark/tlc2/value/impl/SetOfFcnsBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/value/impl/SetOfFcnsBenchmark.java).
- [ ] [test-benchmark/tlc2/value/impl/SubsetBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/value/impl/SubsetBenchmark.java).
- [ ] [test-benchmark/tlc2/value/impl/SubsetValueBenchmark.java](../../tlaplus/tlatools/org.lamport.tlatools/test-benchmark/tlc2/value/impl/SubsetValueBenchmark.java).

Supporting Java files are not additional tests: test-concurrent fingerprint generators, JPF queue tasks/listeners/synchronization policy, test-verify-model value-stream models, and test-model native override providers accompany their respective original tests. Freeze required original models/configs/resources under `tlc/test_vectors/` when each method is translated; fixture presence alone does not mark a test complete.

## Confirmed main-suite method mapping

This appendix records the numerator, including methods in partially translated classes, so future updates can be reviewed without relying on Go function-name counts. Listed classes with no pending methods are complete at the original-method level for this inventory; transport and other untested feature work may still remain.

| Original Java class | Mapped non-ignored methods | Go translation |
| --- | --- | --- |
| [tlc2/TraceExpressionSpecDeadlockTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/TraceExpressionSpecDeadlockTest.java) | `testSpec` (from `TraceExpressionSpecTest`) | [tlc_generated_trace_java_test.go](../tlc_generated_trace_java_test.go) |
| [tlc2/TraceExpressionSpecLassoTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/TraceExpressionSpecLassoTest.java) | `testSpec` (from `TraceExpressionSpecTest`) | [tlc_generated_trace_java_test.go](../tlc_generated_trace_java_test.go) |
| [tlc2/TraceExpressionSpecRuntimeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/TraceExpressionSpecRuntimeTest.java) | `testSpec` (from `TraceExpressionSpecTest`) | [tlc_generated_trace_java_test.go](../tlc_generated_trace_java_test.go) |
| [tlc2/TraceExpressionSpecSafetyBFSTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/TraceExpressionSpecSafetyBFSTest.java) | `testSpec` (from `TraceExpressionSpecTest`) | [tlc_generated_trace_java_test.go](../tlc_generated_trace_java_test.go) |
| [tlc2/TraceExpressionSpecSafetySimTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/TraceExpressionSpecSafetySimTest.java) | `testSpec` (from `TraceExpressionSpecTest`) | [tlc_generated_trace_java_test.go](../tlc_generated_trace_java_test.go) |
| [tlc2/debug/Debug02Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/Debug02Test.java) | `testSpec` | [tlc_debug02_debugger_java_test.go](../tlc_debug02_debugger_java_test.go) |
| [tlc2/debug/Debug03SimTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/Debug03SimTest.java) | `testSpec` | [tlc_debug03_debugger_sim_java_test.go](../tlc_debug03_debugger_sim_java_test.go) |
| [tlc2/debug/Debug03Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/Debug03Test.java) | `testSpec` | [tlc_debug03_debugger_java_test.go](../tlc_debug03_debugger_java_test.go) |
| [tlc2/debug/Debug04SimTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/Debug04SimTest.java) | `testSpec` | [tlc_debug04_debugger_sim_java_test.go](../tlc_debug04_debugger_sim_java_test.go) |
| [tlc2/debug/Debug05SimTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/Debug05SimTest.java) | `testSpec` | [tlc_debug05_debugger_sim_java_test.go](../tlc_debug05_debugger_sim_java_test.go) |
| [tlc2/debug/DebugTLCVariableTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/DebugTLCVariableTest.java) | `testFiniteEmptySetValue`, `testFiniteSetValue`, `testFiniteNestedValue`, `testInfiniteValue` | [tlc/debug_variable_java_test.go](debug_variable_java_test.go) |
| [tlc2/debug/EWD840DebuggerSimTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/EWD840DebuggerSimTest.java) | `testSpec` | [tlc_ewd840_debugger_sim_java_test.go](../tlc_ewd840_debugger_sim_java_test.go) |
| [tlc2/debug/EWD840DebuggerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/EWD840DebuggerTest.java) | `testSpec` | [tlc_ewd840_debugger_java_test.go](../tlc_ewd840_debugger_java_test.go) |
| [tlc2/debug/EWD840ErrorActionDebuggerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/EWD840ErrorActionDebuggerTest.java) | `testSpec` | [tlc_ewd840_error_action_debugger_java_test.go](../tlc_ewd840_error_action_debugger_java_test.go) |
| [tlc2/debug/EWD840ErrorDebuggerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/EWD840ErrorDebuggerTest.java) | `testSpec` | [tlc_ewd840_error_debugger_java_test.go](../tlc_ewd840_error_debugger_java_test.go) |
| [tlc2/debug/EWD998ChanDebuggerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/EWD998ChanDebuggerTest.java) | `testSpec` | [tlc_ewd998_debugger_java_test.go](../tlc_ewd998_debugger_java_test.go) |
| [tlc2/debug/EWD998TraceDebuggerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/EWD998TraceDebuggerTest.java) | `testSpec` | [tlc_ewd998_trace_debugger_java_test.go](../tlc_ewd998_trace_debugger_java_test.go) |
| [tlc2/debug/EchoDebuggerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/EchoDebuggerTest.java) | `testSpec` | [tlc_echo_debugger_java_test.go](../tlc_echo_debugger_java_test.go) |
| [tlc2/debug/ExpressionBreakpointTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/ExpressionBreakpointTest.java) | `testSpec` | [tlc_expression_breakpoint_java_test.go](../tlc_expression_breakpoint_java_test.go) |
| [tlc2/debug/GetScopedIdentifiersTests.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/GetScopedIdentifiersTests.java) | `test` | [tlc_debug_scoped_identifiers_java_test.go](../tlc_debug_scoped_identifiers_java_test.go) |
| [tlc2/debug/TLCDebuggerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/debug/TLCDebuggerTest.java) | `testStackFrameWhileRunning`, `testStackFramePaginationEmpty`, `testStackFramePaginationEmpty2`, `testStackFramePaginationEmpty3`, `testStackFramePaginationEmpty4`, `testStackFramePaginationEmpty5`, `testStackFramePaginationEmpty6`, `testStackFramePaginationEmpty7`, `testStackFramePaginationLevelNull`, `testStackFramePaginationLevel0`, `testStackFramePaginationStartFrame0`, `testStackFramePaginationStartFrame1`, `testStackFramePaginationStartFrameNegative`, `testStackFramePagination`, `testStackFramePaginationsStartFrame1`, `testStackFramePaginationsStartFrameSubList`, `testStackFramePaginationsStartFrameOutOfRange` | [tlc/debug_java_test.go](debug_java_test.go) |
| [tlc2/model/AssignmentTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/model/AssignmentTest.java) | `test` | [tlc/model_test.go](model_test.go) |
| [tlc2/model/FormulaTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/model/FormulaTest.java) | `testUnnamed`, `testNamed` | [tlc/model_test.go](model_test.go) |
| [tlc2/model/MCErrorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/model/MCErrorTest.java) | `testGetErrorMessage`, `testUpdateStatesForTraceExpressions` | [tlc/model_error_state_java_test.go](model_error_state_java_test.go) |
| [tlc2/model/MCStateTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/model/MCStateTest.java) | `testParseRoundTrips`, `testSimpleRecordPrinter` | [tlc/model_error_state_java_test.go](model_error_state_java_test.go) |
| [tlc2/model/TypedSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/model/TypedSetTest.java) | `testParseSet1`, `testParseSet2`, `testParseSet3`, `testParseSet4`, `testParseSet5`, `testParseSet6` | [tlc/model_test.go](model_test.go) |
| [tlc2/module/JsonTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/JsonTest.java) | `test` | [tlc_json_java_test.go](../tlc_json_java_test.go) |
| [tlc2/module/TLCExtTraceAliasTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/TLCExtTraceAliasTest.java) | `test` | [tlc_ext_trace_java_test.go](../tlc_ext_trace_java_test.go) |
| [tlc2/module/TLCExtTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/TLCExtTraceTest.java) | `test` | [tlc_ext_trace_java_test.go](../tlc_ext_trace_java_test.go) |
| [tlc2/tool/AliasLivenessLassoTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AliasLivenessLassoTest.java) | `testSpec` | [tlc_alias_liveness_java_test.go](../tlc_alias_liveness_java_test.go) |
| [tlc2/tool/AliasLivenessStutteringTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AliasLivenessStutteringTest.java) | `testSpec` | [tlc_alias_liveness_java_test.go](../tlc_alias_liveness_java_test.go) |
| [tlc2/tool/AliasSafetySimuTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AliasSafetySimuTest.java) | `testSpec` | [tlc_alias_safety_java_test.go](../tlc_alias_safety_java_test.go) |
| [tlc2/tool/AliasSafetyTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AliasSafetyTest.java) | `testSpec` | [tlc_alias_safety_java_test.go](../tlc_alias_safety_java_test.go) |
| [tlc2/tool/ConstLevelInvariantTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ConstLevelInvariantTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/ConstantOperatorConfigurationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ConstantOperatorConfigurationTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/CyclicRedefineInitTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CyclicRedefineInitTest.java) | `testSpec` | [tlc_cyclic_redefine_java_test.go](../tlc_cyclic_redefine_java_test.go) |
| [tlc2/tool/CyclicRedefineInstanceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CyclicRedefineInstanceTest.java) | `testSpec` | [tlc_cyclic_redefine_java_test.go](../tlc_cyclic_redefine_java_test.go) |
| [tlc2/tool/CyclicRedefineNextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CyclicRedefineNextTest.java) | `testSpec` | [tlc_cyclic_redefine_java_test.go](../tlc_cyclic_redefine_java_test.go) |
| [tlc2/tool/CyclicRedefineOpTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CyclicRedefineOpTest.java) | `testSpec` | [tlc_cyclic_redefine_java_test.go](../tlc_cyclic_redefine_java_test.go) |
| [tlc2/tool/CyclicRedefineSubActionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CyclicRedefineSubActionTest.java) | `testSpec` | [tlc_cyclic_redefine_java_test.go](../tlc_cyclic_redefine_java_test.go) |
| [tlc2/tool/CyclicRedefineVarsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CyclicRedefineVarsTest.java) | `testSpec` | [tlc_cyclic_redefine_java_test.go](../tlc_cyclic_redefine_java_test.go) |
| [tlc2/tool/DepthFirstDieHardTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DepthFirstDieHardTest.java) | `testSpec` | [tlc_depth_first_trace_java_test.go](../tlc_depth_first_trace_java_test.go) |
| [tlc2/tool/DepthFirstErrorTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DepthFirstErrorTraceTest.java) | `testSpec` | [tlc_depth_first_trace_java_test.go](../tlc_depth_first_trace_java_test.go) |
| [tlc2/tool/DumpAsDotTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DumpAsDotTest.java) | `testSpec` | [tlc_dump_as_dot_java_test.go](../tlc_dump_as_dot_java_test.go) |
| [tlc2/tool/DumpLoadTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DumpLoadTraceTest.java) | `testLivenessMCDumpLoadTraceJSON`, `testLivenessMCDumpLoadTraceTLC`, `testLivenessMCDumpLoadTraceJSONAutoWorkers`, `testLivenessMCDumpLoadTraceTLCAutoWorkers`, `testSafetyDumpLoadTraceJSON`, `testSafetyDumpLoadTraceTLC`, `testSafetyDumpLoadTraceJSONAutoWorkers`, `testSafetyDumpLoadTraceTLCAutoWorkers`, `testLivenessBidirectionalDumpLoadTraceJSON`, `testLivenessBidirectionalDumpLoadTraceTLC`, `testLivenessBidirectionalDumpLoadTraceJSONAutoWorkers`, `testLivenessBidirectionalDumpLoadTraceTLCAutoWorkers`, `testSafetyTESpecEqAliasDumpLoadTraceJSON`, `testSafetyTESpecEqAliasDumpLoadTraceTLC`, `testSafetyTESpecEqAliasDumpLoadTraceJSONAutoWorkers`, `testSafetyTESpecEqAliasDumpLoadTraceTLCAutoWorkers`, `testLivenessExample1DumpLoadTraceJSON`, `testLivenessExample1DumpLoadTraceTLC`, `testLivenessExample1DumpLoadTraceJSONAutoWorkers`, `testLivenessExample1DumpLoadTraceTLCAutoWorkers`, `testSafetyDieHardAliasSubDumpLoadTraceJSON`, `testSafetyDieHardAliasSubDumpLoadTraceTLC`, `testSafetyDieHardAliasSub2DumpLoadTraceJSON`, `testSafetyDieHardAliasSub2DumpLoadTraceTLC`, `testSafetyDieHardAliasSub2DumpLoadTraceJSONAutoWorkers`, `testSafetyDieHardAliasSub2DumpLoadTraceTLCAutoWorkers`, `testSafetyDieHardAliasSupDumpLoadTraceJSON`, `testSafetyDieHardAliasSupDumpLoadTraceTLC`, `testSafetyDieHardAliasSupDumpLoadTraceJSONAutoWorkers`, `testSafetyDieHardAliasSupDumpLoadTraceTLCAutoWorkers`, `testLivenessEWD840MC3DumpLoadTraceTLC`, `testLivenessEWD840MC3DumpLoadTraceTLCAutoWorkers` | [tlc_dump_load_trace_java_test.go](../tlc_dump_load_trace_java_test.go) |
| [tlc2/tool/EmptySubsetEqTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EmptySubsetEqTest.java) | `testSpec` | [tlc_model_java_test.go](../tlc_model_java_test.go) |
| [tlc2/tool/EvalExceptionLivenessTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EvalExceptionLivenessTest.java) | `testSpec` | [tlc_eval_exception_liveness_java_test.go](../tlc_eval_exception_liveness_java_test.go) |
| [tlc2/tool/Github1109Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1109Test.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/Github1109aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1109aTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/Github361Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github361Test.java) | `testSpec` | [tlc_constant_processing_java_test.go](../tlc_constant_processing_java_test.go) |
| [tlc2/tool/Github766SimulateTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github766SimulateTest.java) | `testSpec` | [tlc_github_766_java_test.go](../tlc_github_766_java_test.go) |
| [tlc2/tool/Github766Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github766Test.java) | `testSpec` | [tlc_github_766_java_test.go](../tlc_github_766_java_test.go) |
| [tlc2/tool/Github798ITest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github798ITest.java) | `testSpec` | [tlc_github_798_807_java_test.go](../tlc_github_798_807_java_test.go) |
| [tlc2/tool/Github798NTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github798NTest.java) | `testSpec` | [tlc_github_798_807_java_test.go](../tlc_github_798_807_java_test.go) |
| [tlc2/tool/Github807Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github807Test.java) | `testSpec` | [tlc_github_798_807_java_test.go](../tlc_github_798_807_java_test.go) |
| [tlc2/tool/Github817Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817Test.java) | `testSpec` | [tlc_github_817_java_test.go](../tlc_github_817_java_test.go) |
| [tlc2/tool/Github817bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817bTest.java) | `testSpec` | [tlc_github_817_java_test.go](../tlc_github_817_java_test.go) |
| [tlc2/tool/Github817cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817cTest.java) | `testSpec` | [tlc_github_817_java_test.go](../tlc_github_817_java_test.go) |
| [tlc2/tool/Github817dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817dTest.java) | `testSpec` | [tlc_github_817_java_test.go](../tlc_github_817_java_test.go) |
| [tlc2/tool/Github817eTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817eTest.java) | `testSpec` | [tlc_github_817_java_test.go](../tlc_github_817_java_test.go) |
| [tlc2/tool/Github819Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github819Test.java) | `testSpec` | [tlc_github_819_849_java_test.go](../tlc_github_819_849_java_test.go) |
| [tlc2/tool/Github849Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github849Test.java) | `testSpec` | [tlc_github_819_849_java_test.go](../tlc_github_819_849_java_test.go) |
| [tlc2/tool/Github858Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github858Test.java) | `testSpec` | [tlc_github_858_java_test.go](../tlc_github_858_java_test.go) |
| [tlc2/tool/Github866Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github866Test.java) | `testSpec` | [tlc_github_866_java_test.go](../tlc_github_866_java_test.go) |
| [tlc2/tool/Github971aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971aTest.java) | `testSpec` | [tlc_github_971_java_test.go](../tlc_github_971_java_test.go) |
| [tlc2/tool/Github971bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971bTest.java) | `testSpec` | [tlc_github_971_java_test.go](../tlc_github_971_java_test.go) |
| [tlc2/tool/Github971cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971cTest.java) | `testSpec` | [tlc_github_971_java_test.go](../tlc_github_971_java_test.go) |
| [tlc2/tool/Github971dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971dTest.java) | `testSpec` | [tlc_github_971_java_test.go](../tlc_github_971_java_test.go) |
| [tlc2/tool/Github971eTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971eTest.java) | `testSpec` | [tlc_github_971_java_test.go](../tlc_github_971_java_test.go) |
| [tlc2/tool/IncompleteNextMultipleActionsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/IncompleteNextMultipleActionsTest.java) | `testSpec` | [tlc_next_model_java_test.go](../tlc_next_model_java_test.go) |
| [tlc2/tool/IncompleteNextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/IncompleteNextTest.java) | `testSpec` | [tlc_next_model_java_test.go](../tlc_next_model_java_test.go) |
| [tlc2/tool/PostConditionsFailTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PostConditionsFailTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/PostConditionsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PostConditionsTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/SubsetEqTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/SubsetEqTest.java) | `testSpec` | [tlc_model_java_test.go](../tlc_model_java_test.go) |
| [tlc2/tool/TLCExtTraceSimTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCExtTraceSimTest.java) | `testSpec` | [tlc_ext_trace_java_test.go](../tlc_ext_trace_java_test.go) |
| [tlc2/tool/TLCGetAllTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCGetAllTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/TLCGetNamedUndefinedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCGetNamedUndefinedTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/TLCSetInitTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCSetInitTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/TLCSetMultiSimTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCSetMultiSimTest.java) | `testSpec` (from `TLCSetSimTest`) | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/TLCSetSimTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCSetSimTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/TLCSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCSetTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/ValueSemanticsAssumeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ValueSemanticsAssumeTest.java) | `testSpec` | [tlc_value_semantics_assume_java_test.go](../tlc_value_semantics_assume_java_test.go) |
| [tlc2/tool/coverage/ACoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/ACoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/BCoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/BCoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/CCoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/CCoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/CoverageStatisticsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/CoverageStatisticsTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/DCoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/DCoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/ECoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/ECoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/FCoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/FCoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/GCoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/GCoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/Github314CoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/Github314CoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/Github377CoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/Github377CoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/Github649CoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/Github649CoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/HCoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/HCoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/ICoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/ICoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/ImpliedCoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/ImpliedCoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/JCoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/JCoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/KCoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/KCoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/LCoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/LCoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/MCoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/MCoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/tool/coverage/OCoverageTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/OCoverageTest.java) | `testSpec` | [tlc_coverage_model_java_test.go](../tlc_coverage_model_java_test.go) |
| [tlc2/output/MPTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/output/MPTest.java) | `testPrintErrorInt`, `testPrintErrorIntString`, `testPrintErrorIntStringArray`, `testPrintProgressStats` | [tlc/output_mp_java_test.go](output_mp_java_test.go) |
| [tlc2/TLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/TLCTest.java) | `testHandleParametersAbsoluteInvalid`, `testHandleParametersAbsoluteValid`, `testHandleParametersFractionInvalid`, `testHandleParametersAllocateLowerBound`, `testHandleParametersAllocateUpperBound`, `testHandleParametersAllocateHalf`, `testHandleParametersAllocate90`, `testHandleParametersMaxSetSize`, `testHandleParametersSimulateFileNum`, `testRuntimeConversion` | [tlc/cli_java_test.go](cli_java_test.go) |
| [tlc2/output/SpecTraceExpressionWriterTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/output/SpecTraceExpressionWriterTest.java) | `testInitNextWithNoError`, `testInitNextWithError`, `testInitNextWithErrorAndTraceExpression`, `testMultilineTraceExpression` | [spec_trace_writer_java_test.go](../spec_trace_writer_java_test.go) |
| [tlc2/output/WarningControlTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/output/WarningControlTest.java) | `testSuppressMessagesSanyCode`, `testSuppressMessagesSanyErrorCodeFails`, `testSuppressMessagesTlcCode`, `testSuppressMessagesMultipleCodes`, `testSuppressMessagesUnknownCodeFails`, `testSuppressMessagesMissingArgFails`, `testMessagesAsErrorsSanyWarningCode`, `testMessagesAsErrorsTlcCode`, `testMessagesAsErrorsUnknownCodeFails`, `testMessagesAsErrorsMissingArgFails`, `testNowarningConflictWithSuppressMessages`, `testNowarningConflictWithmessagesAsErrors`, `testSameTlcCode`, `testSameSanyCode`, `testRuntimeSuppressedWarningProducesNoOutput`, `testRuntimeUnsuppressedWarningProducesOutput`, `testRuntimeWarningAsErrorThrowsTLCRuntimeException`, `testRuntimeWarningWithoutElevationDoesNotThrow` | [tlc/output_warning_control_java_test.go](output_warning_control_java_test.go) |
| [tlc2/tool/coverage/OpApplNodeWrapperTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/coverage/OpApplNodeWrapperTest.java) | `testReportCoverage01`, `testReportCoverage02`, `testReportCoverage03`, `testReportCoverage04` | [tlc/coverage_java_test.go](coverage_java_test.go) |
| [tlc2/tool/distributed/DistributedDoInitFunctorEvalExceptionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/DistributedDoInitFunctorEvalExceptionTest.java) | `testSpec` | [tlc_distributed_test.go](../tlc_distributed_test.go) |
| [tlc2/tool/distributed/TLCWorkerSmartProxyTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/TLCWorkerSmartProxyTest.java) | `testGetNetworkOverheadMaxStateOne`, `testGetNetworkOverheadMinStateOne`, `testGetNetworkOverheadZeroStateOne`, `testGetNetworkOverheadMaxStateZero`, `testGetNetworkOverheadMinStateZero`, `testGetNetworkOverheadZeroStateZero`, `testGetNetworkOverheadMinStateMax`, `testGetNetworkOverheadMaxStateMa`, `testGetNetworkOverheadZeroStateMax` | [tlc/distributed_smart_proxy_test.go](distributed_smart_proxy_test.go) |
| [tlc2/tool/distributed/fp/DynamicFPSetManagerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/fp/DynamicFPSetManagerTest.java) | `testCtorInvalidZero`, `testCtorInvalidMin1`, `testCtor1`, `testCtor10`, `testCtor31`, `testCtor32`, `testCtor33`, `testCtorMax`, `testGetIndexSingleFPSet`, `testGetIndex10FPSet`, `testReassingInvalidMin1`, `testReassingInvalid2`, `testReassingTerminate`, `testReassing`, `testFailoverPut`, `testFailoverPutBlock`, `testFailoverTerminationPutBlock`, `testFailoverTerminationPutBlockConcurrent`, `testPutBlockConcurrentOrder` | [tlc/distributed_fp_manager_test.go](distributed_fp_manager_test.go) |
| [tlc2/tool/distributed/fp/FPSetManagerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/fp/FPSetManagerTest.java) | `test2`, `test3`, `test4`, `test5`, `test8` | [tlc/distributed_fp_manager_test.go](distributed_fp_manager_test.go) |
| [tlc2/tool/doinitfunctor/DoInitFunctorEvalExceptionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/doinitfunctor/DoInitFunctorEvalExceptionTest.java) | `testSpec` | [tlc_init_model_java_test.go](../tlc_init_model_java_test.go) |
| [tlc2/tool/doinitfunctor/DoInitFunctorInvariantContinueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/doinitfunctor/DoInitFunctorInvariantContinueTest.java) | `testSpec` | [tlc_init_model_java_test.go](../tlc_init_model_java_test.go) |
| [tlc2/tool/doinitfunctor/DoInitFunctorInvariantMinimalErrorStackTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/doinitfunctor/DoInitFunctorInvariantMinimalErrorStackTest.java) | `testSpec` | [tlc_init_model_java_test.go](../tlc_init_model_java_test.go) |
| [tlc2/tool/doinitfunctor/DoInitFunctorInvariantNoContinueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/doinitfunctor/DoInitFunctorInvariantNoContinueTest.java) | `testSpec` | [tlc_init_model_java_test.go](../tlc_init_model_java_test.go) |
| [tlc2/tool/doinitfunctor/DoInitFunctorInvariantTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/doinitfunctor/DoInitFunctorInvariantTest.java) | `testSpec` | [tlc_init_model_java_test.go](../tlc_init_model_java_test.go) |
| [tlc2/tool/doinitfunctor/DoInitFunctorPropertyTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/doinitfunctor/DoInitFunctorPropertyTest.java) | `testSpec` | [tlc_init_model_java_test.go](../tlc_init_model_java_test.go) |
| [tlc2/tool/fp/MultiFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/MultiFPSetTest.java) | `testCTorLowerMin`, `testCTorMin`, `testCTorMax`, `testCTorHigherMax`, `testPutMax`, `testPutMin`, `testPutZero`, `testGetFPSet`, `testGetFPSet0`, `testGetFPSet1`, `testGetFPSetL`, `testGetFPSet0L`, `testGetFPSet1L`, `testGetFPSetOffHeap`, `testGetFPSetOffHeap0`, `testGetFPSetOffHeap1` | [tlc/multi_fpset_java_test.go](multi_fpset_java_test.go) |
| [tlc2/tool/queue/StateQueueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/queue/StateQueueTest.java) | `testEnqueue`, `testsDequeueEmpty`, `testDequeueEmpty`, `testsDequeueNotEmpty`, `testDequeueNotEmpty`, `testEnqueueAddNotSame`, `testEnqueueAddSame`, `testsDequeueAbuseEmpty`, `testsDequeueAbuseNonEmpty` | [tlc/queue_test.go](queue_test.go) |
| [tlc2/util/BitVectorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/BitVectorTest.java) | `testToString`, `testToStringRange` | [tlc/bit_vector_test.go](bit_vector_test.go) |
| [tlc2/util/ContextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/ContextTest.java) | `testLookupEmpty`, `testLookupBranch`, `testLookupSymbolNodeNull`, `testLookup`, `testLookupCutOffFalse`, `testLookupCutOffTrue`, `testLookupWithAtBranching`, `testLookupWithCutOffFalseAtBranching`, `testLookupWithCutOffTrueAtBranching`, `testLookupSymbolNode` | [tlc/context_java_test.go](context_java_test.go) |
| [tlc2/util/FP64Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/FP64Test.java) | `testExtendLongInt` | [tlc/fp64_test.go](fp64_test.go) |
| [tlc2/util/MemIntStackTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/MemIntStackTest.java) | `testPeak` | [tlc/int_stack_test.go](int_stack_test.go) |
| [tlc2/util/SynchronousDiskIntStackTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/SynchronousDiskIntStackTest.java) | `testPushIntNoWrite`, `testPushIntWrite` | [tlc/int_stack_test.go](int_stack_test.go) |
| [tlc2/value/impl/EnumerableValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/EnumerableValueTest.java) | `test` | [tlc/value_enumerable_java_test.go](value_enumerable_java_test.go) |
| [tlc2/value/impl/FcnRcdValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/FcnRcdValueTest.java) | `testSelecEmpty`, `testSelecNormalizedEmpty`, `testEmptyIntervalDomainToTuple`, `testSelect`, `testSelectNormalized`, `testSelectLinearSearchTypedMV`, `testSelectBinarySearchTypedMV`, `testMalformedExplicitFcnEqualsIntervalDoesNotWrap`, `testMalformedIntervalFcnSelectDoesNotWrap`, `testMalformedIntervalFcnExceptDoesNotWrap`, `testEmptyIntervalFcnCompareToAgreesWithEquals`, `testEmptyIntervalFcnVsEmptyTupleCompareTo`, `testEmptyIntervalFcnsNormalize` | [tlc/value_function_test.go](value_function_test.go) |
| [tlc2/value/impl/IntervalValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/IntervalValueTest.java) | `testElementAt`, `testElementAtOutOfBoundsNegative`, `testElementAtOutOfBoundsSize`, `sizeOverflow`, `compareToOverflow1`, `testCompareExtremeIntervals`, `testEmptyIntervalEquality`, `testCompareEmptyIntervals`, `testSizeOfMaximumRepresentableInterval`, `testExtremeIntervalSize`, `testMaxIntSingletonEnumerator`, `testMaxIntSingletonSubsetEq`, `testMaxIntSingletonFingerprint`, `testMaxIntSingletonDiffCapCup` | [tlc/value_interval_test.go](value_interval_test.go) |
| [tlc2/value/impl/ModelValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/ModelValueTest.java) | `testEqualsUntyped`, `testEqualsStringMV`, `testEqualsIntMV`, `testEqualsBoolMV`, `testEqualsIVMV`, `testEqualsRcdMV`, `testEqualsTupeMV`, `testCompareToUntyped`, `testCompareToStringMV`, `testCompareToIntMV`, `testCompareToBoolMV`, `testCompareToIVMV`, `testCompareToRcdMV`, `testCompareToTupeMV`, `testEqualsTyped`, `testEqualsTypedMVsUntyped`, `testEqualsTwoTypedMVs`, `testEqualsStringTypedMV`, `testEqualsTypedMVString`, `testEqualsIntTypedMV`, `testEqualsTypedMVInt`, `testEqualsBoolTypedMV`, `testEqualsTypedMVBool`, `testEqualsIVTypedMV`, `testEqualsTypedMVIV`, `testEqualsRcdTypedMV`, `testEqualsTypedMVRcd`, `testEqualsTupeTypedMV`, `testEqualsTypedMVTupe`, `testCompareToTyped`, `testCompareToTypedMVsUntyped`, `testCompareToTwoTypedMVs`, `testCompareToStringTypedMV`, `testCompareToTypedMVString`, `testCompareToIntTypedMV`, `testCompareToTypedMVInt`, `testCompareToBoolTypedMV`, `testCompareToTypedMVBool`, `testCompareToIVTypedMV`, `testCompareToTypedMVIV`, `testCompareToRcdTypedMV`, `testCompareToTypedMVRcd`, `testCompareToTupeTypedMV`, `testCompareToTypedMVTupe` | [tlc/value_model_test.go](value_model_test.go) |
| [tlc2/value/impl/FcnLambdaValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/FcnLambdaValueTest.java) | `testEmptyIntervalDomainToTuple`, `testToString`, `testToFcnRcd`, `testSelectAndApply`, `testTakeExceptOverridesValue`, `testDomainIsStableAcrossConversion`, `testToFcnRcdReturnsCachedInstance`, `testToRcdForIntervalDomainIsNull`, `testToRcdForStringDomain`, `testFingerprintStableAcrossConversion`, `testTakeExceptThenConvertToFcnRcd`, `testToFcnRcdAssertFail`, `testToFcnRcdClassCastException`, `testToFcnRcdSilentCorruption`, `testToFcnRcdSilentCorruptionFP`, `testToTupleWithExceptIntervalDomain`, `testToTupleWithExceptSetEnumDomain`, `testToTupleWithExceptFP` | [tlc/value_lambda_java_test.go](value_lambda_java_test.go) |
| [tlc2/value/impl/InitializeValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/InitializeValueTest.java) | `union`, `setcap`, `setcup`, `setdiff`, `subset`, `record`, `fcnrecord`, `tuple`, `setOfTuple`, `setOfRcds` | [tlc/value_initialize_test.go](value_initialize_test.go) |
| [tlc2/value/impl/RecordValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/RecordValueTest.java) | `testDeepCopy`, `testErrorMessages` | [tlc/value_record_test.go](value_record_test.go) |
| [tlc2/value/impl/SetOfTuplesValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/SetOfTuplesValueTest.java) | `testToStringLazy`, `testEmptyNonEnumerableComponent`, `testRandomSubsetEmptyNonEnumerableComponent`, `testRandomSubsetBeyondSetBound`, `testRandomSubsetBeyondIntMaxValue` | [tlc/value_tuple_product_java_test.go](value_tuple_product_java_test.go) + [tlc/value_setconstructors_test.go](value_setconstructors_test.go) |
| [tlc2/value/impl/ValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/ValueTest.java) | `testIsEmpty` | [tlc/value_empty_test.go](value_empty_test.go) |

## Maintaining this inventory

When a complete original method lands, record its Java class/method and Go file, remove only that method from the unchecked topic entry, and adjust the topic and overall counts. Keep inherited subclass contexts, original ignored methods, and supplementary suites separate. Recompute against the Java checkout if it changes. Reconcile related legacy Go tests before claiming additional credit; source comments or PORT_PROGRESS claims alone do not prove that every original assertion is retained.
| [tlc2/value/impl/SetOfFcnsValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/SetOfFcnsValueTest.java) | `testRangeSubsetValue`, `testDomainEmpty`, `testRangeEmpty`, `testDomainAndRangeEmpty`, `testRandomSubsetAndValueEnumerator`, `testDomainModelValue`, `testDomainIntervalRangeSetEnumValueSize9`, `testDomainIntervalRangeSetEnumValueSize27`, `testDomainIntervalRangeSetEnumValueSize256`, `testRandomSubsetFromReallyLarge`, `testEmptyNonEnumerableDomain`, `testUnitNonEnumerableRange`, `testUnitNonEnumerableRangeInterval`, `testNonEnumerableRange`, `testNonEnumerableRangeInterval`, `testRandomSubsetEmptyNonEnumerableDomain` | [tlc/value_fcn_set_java_test.go](value_fcn_set_java_test.go) |
| [tlc2/value/impl/TupleValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/TupleValueTest.java) | `testErrorMessages` | [tlc/value_tuple_java_test.go](value_tuple_java_test.go) |
| [tlc2/value/impl/SetOfRcrdValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/SetOfRcrdValueTest.java) | `testSimple`, `testRangeSubsetValue`, `testRandomSubset`, `testRandomSubsetVaryingParameters`, `testRandomSubsetAstronomically`, `testEmptyNonEnumerableField`, `testRandomSubsetEmptyNonEnumerableField` | [tlc/value_record_set_java_test.go](value_record_set_java_test.go) |
| [tlc2/value/impl/SubsetEnumeratorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/SubsetEnumeratorTest.java) | `testElementsInt`, `testGetRandomSubset` (all 12 parameter rows) | [tlc/value_subset_enumerator_java_test.go](value_subset_enumerator_java_test.go) |
| [tlc2/value/impl/SubsetValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/SubsetValueTest.java) | `testRandomSubsetE7F1`, `testRandomSubsetE7F05`, `testRandomSubsetE6F1`, `testRandomSubsetE5F01`, `testRandomSubsetE5F025`, `testRandomSubsetE5F05`, `testRandomSubsetE5F075`, `testRandomSubsetE5F1`, `testRandomSubsetE32F1ENeg6`, `testRandomSubsetE17F1ENeg3`, `testRandomSubsetSubset16`, `testRandomSubsetSubset256`, `testRandomSubsetSubset65536`, `testRandomSubsetSubsetNoOverflow`, `testEmptyEnumerationsAreIndependent`, `testKSubsetEnumerator`, `testKSubsetEnumeratorNegative`, `testKSubsetEnumeratorGTCapacity`, `testNumKSubset`, `testNumKSubset2`, `testNumKSubsetNeg`, `testNumKSubsetKGTN`, `testNumKSubsetUpTo62`, `testNumKSubsetPreventsOverflow`, `testUnrankKSubsets`, `testUnrank16viaRank`, `testRandomSetOfSubsets`, `testRandomSetOfSubsets300`, `testRandomSetOfSubsets400`, `testElementsNormalizedIsNormalized`, `testKElementsAreNormalized`, `testKElementsMatchElementsNormalized`, `testRandomSubsetGeneratorK0`, `testRandomSubsetGeneratorKNegative`, `testRandomSubsetGeneratorKNplus1`, `testRandomSubsetGeneratorN10`, `testRandomSubsetGeneratorN100` | [tlc/value_subset_java_test.go](value_subset_java_test.go) |
| [tlc2/value/impl/KSubsetValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/KSubsetValueTest.java) | `testEnumerateN32`, `testEnumerateN33`, `testEnumerateN63`, `testEnumerateN64`, `testNormalization`, `testKSubsetFingerprintingS009`, `testKSubsetFingerprintingS032`, `testKSubsetFingerprintingS033`, `testKSubsetFingerprintingS063`, `testInvalidKDenotesEmptySet`, `testToStringLargeSwallowsCountError` | [tlc/value_ksubset_java_test.go](value_ksubset_java_test.go) |
| [tlc2/value/ValueInputOutputStreamTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/ValueInputOutputStreamTest.java) | `testWriteShort`, `testWriteInt`, `testWriteShortNat`, `testWriteNat`, `testBlindReadStringValue`, `testBlindReadRecordValue` | [tlc/value_stream_java_test.go](value_stream_java_test.go) |
| [tlc2/value/StringDeserializeTLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/StringDeserializeTLCTest.java) | `test` | [tlc_string_deserialize_java_test.go](../tlc_string_deserialize_java_test.go) |
| [tlc2/tool/queue/DiskPoolWriterTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/queue/DiskPoolWriterTest.java) | `testStatePoolWriterIgnoresEmptyWakeAndStopsOnFinish`, `testByteArrayPoolWriterIgnoresEmptyWakeAndStopsOnFinish` | [tlc/disk_pool_writer_java_test.go](disk_pool_writer_java_test.go) |
| [tlc2/util/CombinatoricsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/CombinatoricsTest.java) | `testChoose`, `testChooseBigChoose`, `testSlowChooseBigChoose`, `testBigChoose50c1`, `testBigChoose50c10`, `testBigChoose50c20`, `testBigChoose50c30`, `testBigChoose400c1`, `testBigChoose400c50`, `testBigChoose400c100`, `testBigChoose400c200` | [tlc/combinatorics_java_test.go](combinatorics_java_test.go) |
| [tlc2/util/GrowingLongVecTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/GrowingLongVecTest.java) | `testReadBeyondCapacity` (from `LongVecTest`), `testAddAndReadBeyondCapacity` (from `LongVecTest`), `testRemoveBeyondCapacity` (from `LongVecTest`), `testAddRemoveBeyondCapacity` (from `LongVecTest`), `testRemoveAndGet` (from `LongVecTest`), `testRemoveWrongOrder` (from `LongVecTest`), `testGetNegative` (from `LongVecTest`), `testRemoveNegative` (from `LongVecTest`), `testGrowAndShrink` | [tlc/long_vec_java_test.go](long_vec_java_test.go) |
| [tlc2/util/LongVecTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/LongVecTest.java) | `testReadBeyondCapacity`, `testAddAndReadBeyondCapacity`, `testRemoveBeyondCapacity`, `testAddRemoveBeyondCapacity`, `testRemoveAndGet`, `testRemoveWrongOrder`, `testGetNegative`, `testRemoveNegative` | [tlc/long_vec_java_test.go](long_vec_java_test.go) |
| [tlc2/util/ByteUtilsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/ByteUtilsTest.java) | `test1`, `test2`, `test3`, `test4`, `test5`, `test6` | [tlc/byte_utils_java_test.go](byte_utils_java_test.go) |
| [tlc2/util/MemIntQueueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/MemIntQueueTest.java) | `testDequeuePastLastElement`, `testEnqueueZeros`, `testEnqueueLong`, `testEnqueueDequeueLong`, `testGrow` | [tlc/int_queue_java_test.go](int_queue_java_test.go) |
| [tlc2/util/statistics/BucketStatisticsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/statistics/BucketStatisticsTest.java) | `testInvalidArgument`, `testMean`, `testMedian`, `testMin`, `testMin2`, `testMax`, `testStandardDeviation`, `testGetPercentile`, `testGetPercentileNaN`, `testToString` | [tlc/bucket_statistics_java_test.go](bucket_statistics_java_test.go) |
| [tlc2/util/statistics/FixedSizedBucketStatisticsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/statistics/FixedSizedBucketStatisticsTest.java) | `testMin`, `testMin2`, `testMax`, `testInvalidArgument`, `testGetPercentileNaN`, `testMaximum` | [tlc/bucket_statistics_java_test.go](bucket_statistics_java_test.go) |
| [tlc2/util/BufferedRandomAccessFileTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/BufferedRandomAccessFileTest.java) | `testWrite`, `testWriteSeek`, `testWriteSeekNoLength`, `testRead`, `testReadSeekNoLength`, `testInvalidateBufferedData`, `testReadAfterSeekPastEndOfFile`, `testWriteAfterSeekPastEndOfFile`, `testObscureSetLengthBehavior`, `testIdempotentClose`, `testIOExceptionOnUseAfterClose`, `regressionTest01`, `regressionTest02`, `regressionTest03`, `regressionTest04`, `regressionTest05`, `regressionTest06`, `regressionTest07` | [tlc/buffered_random_access_file_java_test.go](buffered_random_access_file_java_test.go) |
| [tlc2/util/BufferedRandomAccessFileFuzzTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/BufferedRandomAccessFileFuzzTest.java) | `fuzz`, `testWellDefined` | [tlc/buffered_random_access_file_fuzz_java_test.go](buffered_random_access_file_fuzz_java_test.go) |
| [tlc2/module/RandomizationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/RandomizationTest.java) | `testRandomSubsetNonFinite`, `testV1Valid`, `testV2Larger1`, `testSetNonFinite`, `testV1Negative`, `testV1NoIntValue`, `testV1Zero`, `testV2Zero`, `testV2Negative`, `testV3Empty`, `testV3AstronomicallyLarge`, `testV3isInfinite`, `testRSSV2Zero`, `testRSSV2Negative`, `testRSSV2Cardinality`, `testRSSV2TwiceCardinality` | [tlc/modules_randomization_java_test.go](modules_randomization_java_test.go) |
| [tlc2/module/SequencesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/SequencesTest.java) | `testTailString`, `testHeadString`, `testHeadStringEmpty`, `testAppendString`, `testAppendString2`, `testAppendStringNonString`, `testConcatStringToSeq`, `testConcatSeqToString`, `testConcatStringToString`, `testConcatIntToSeq`, `testConcatSeqToInt`, `testConcatIntToInt`, `testSubseq` | [tlc/modules_sequences_tlc_java_test.go](modules_sequences_tlc_java_test.go) |
| [tlc2/module/TLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/TLCTest.java) | `testA`, `testB`, `testCombineMaxIntIntervalOnLeft`, `testCombineMaxIntIntervalOnRight`, `testPermutations` | [tlc/modules_sequences_tlc_java_test.go](modules_sequences_tlc_java_test.go) |
