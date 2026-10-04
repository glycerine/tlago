# Original TLC test-port inventory

Snapshot: 2026-10-03. Java source checkout: `8f4bc8b73ad1202774a6bf70143436f8ba50aab0`; Go baseline: `d68a83b`, with the current working tree inspected. This is a source-to-source inventory, not a new test run or a declaration of full Java behavioral parity.

## Scope and counting

The main inventory covers every test-bearing concrete class in `../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/`, including JUnit 3 tests, annotated JUnit 4 methods, inherited methods, and concrete subclasses with no test methods of their own. Abstract bases and support classes are dependencies, not additional runnable test classes. Separate appendices cover shared `test/util/`, `test-long/`, `test-concurrent/`, `test-verify/`, and benchmarks.

- **114 of 626 non-`@Ignore` concrete classes have all their logical test methods mapped (18.2%).**
- **357 of 1,269 non-`@Ignore` logical test-method contexts have confirmed translations (28.1%).** A context is `(concrete Java class, method)`; inherited methods count once for each concrete subclass.
- **912 logical method contexts across 512 classes remain to port or reconcile.** 1 of those classes have some confirmed methods already ported; other older Go checks may also cover parts of the remaining cases.
- There are 660 test-bearing concrete classes in the source tree overall. 34 are wholly disabled by `@Ignore`; 39 ignored method contexts are tracked separately below and excluded from the percentages.

These are conservative translation counts. Credit requires an explicit original-method translation or an inspected match of the original inputs/assertions; merely testing the same Go type, running a model manually, having an ephemeral Java comparison, or copying a fixture is insufficient. Related Go tests whose full original cases/assertions have not been reconciled remain on the checklist. The percentages measure test translation, not TLC implementation completion or source-code coverage.

Methods are counted **before parameter expansion**, and non-`@Ignore` does not imply every upstream Ant target executes them: assumptions, target exclusions, platforms, and JVM requirements still apply. For example, `OffHeapIndexerEquivalenceTest.testInfiniteInfMult` has **7,254 parameter rows**; each of the three concrete `OffHeapIndexerParameterizedTest` subclasses inherits five methods over **1,104 rows** (16,560 contexts before assumptions). `GetScopedIdentifiersTests` has all 18 rows ported. Preserve complete matrices rather than substituting a few samples.

Excluded from this TLC count: SANY, PlusCal, formatter, Toolbox UI suites, and CommunityModules (a separate project). The complete CommunityModules Ant test target already has its own Go translation in [community_modules_java_test.go](../community_modules_java_test.go). Email reporting and dependencies pursued for email remain excluded under the user’s scope directive.

## Topic totals for the main suite

| Topic | Non-ignored classes | Classes fully mapped | Logical methods | Mapped methods | Pending methods |
| --- | ---: | ---: | ---: | ---: | ---: |
| CLI, REPL, messages, and trace-spec output | 5 | 0 | 37 | 0 | 37 |
| Presentation models | 5 | 3 | 13 | 9 | 4 |
| Standard modules, constants, native overrides, and random values | 44 | 14 | 75 | 14 | 61 |
| Evaluation, initial states, next states, and action composition | 45 | 12 | 51 | 12 | 39 |
| Safety checking, diagnostics, and checker lifecycle | 34 | 7 | 34 | 7 | 27 |
| Issue regressions in the evaluator and checker | 98 | 3 | 101 | 3 | 98 |
| Traces, aliases, dump/load, and generated trace specs | 17 | 14 | 48 | 44 | 4 |
| Generated TTrace recheck variants | 45 | 0 | 45 | 0 | 45 |
| Liveness and fairness model regressions | 101 | 1 | 101 | 1 | 100 |
| Liveness graph, tableau, and expression helpers | 6 | 0 | 48 | 0 | 48 |
| Simulation and multithreaded simulation | 20 | 0 | 55 | 0 | 55 |
| Coverage | 20 | 20 | 23 | 23 | 0 |
| Debugger and scoped identifiers | 16 | 16 | 35 | 35 | 0 |
| Checkpoint and recovery models | 2 | 0 | 2 | 0 | 2 |
| Distributed TLC | 11 | 4 | 41 | 34 | 7 |
| Fingerprint sets, indexers, arrays, and iterators | 20 | 1 | 165 | 16 | 149 |
| Queues and pool writers | 2 | 1 | 11 | 9 | 2 |
| Values, lazy functions, enumeration, and value streams | 17 | 13 | 189 | 134 | 55 |
| Collections, buffered files, combinatorics, and statistics | 14 | 5 | 91 | 16 | 75 |
| Numbered legacy model suite | 104 | 0 | 104 | 0 | 104 |
| **Total** | **626** | **114** | **1,269** | **357** | **912** |

## Porting rules and proposed order

Current user priority (2026-10-03): first finish the pending production work,
get the suite green, and commit all current changes; then port original
correctness tests. Temporarily skip **Debugger and scoped identifiers**,
**Checkpoint and recovery models**, **Distributed TLC**, **JPF concurrency
verification**, and **Benchmarks and supporting fixtures**. Keep their inventory
entries visible. If a test fails, check the mechanical translation and inspect
production TLC for shortcuts; implement the missing behavior before moving on,
without weakening the original assertions. Mark finished entries **Port complete**.

1. Finish incomplete original methods/classes already represented in Go: `DumpLoadTraceTest`’s two enabled binary cases. Reconcile older value/module/utility checks with complete original inputs and assertions.
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

- [ ] [tlc2/REPLTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/REPLTest.java) — **Reconcile**: `testProcessInput`.
  Related Go checks: [repl_test.go](../repl_test.go).
- [ ] [tlc2/TLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/TLCTest.java) — **Reconcile**: `testHandleParametersAbsoluteInvalid`, `testHandleParametersAbsoluteValid`, `testHandleParametersFractionInvalid`, `testHandleParametersAllocateLowerBound`, `testHandleParametersAllocateUpperBound`, `testHandleParametersAllocateHalf`, `testHandleParametersAllocate90`, `testHandleParametersMaxSetSize`, `testHandleParametersSimulateFileNum`, `testRuntimeConversion`.
  Related Go checks: [tlc/runner_test.go](runner_test.go).
- [ ] [tlc2/output/MPTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/output/MPTest.java) — **Reconcile**: `testPrintErrorInt`, `testPrintErrorIntString`, `testPrintErrorIntStringArray`, `testPrintProgressStats`.
  Related Go checks: [tlc/output_test.go](output_test.go).
  Original console-output assertions require the complete MP console boundary.
- [ ] [tlc2/output/SpecTraceExpressionWriterTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/output/SpecTraceExpressionWriterTest.java) — **Reconcile**: `testInitNextWithNoError`, `testInitNextWithError`, `testInitNextWithErrorAndTraceExpression`, `testMultilineTraceExpression`.
  Related Go checks: [tlc/spec_writer_test.go](spec_writer_test.go).
- [ ] [tlc2/output/WarningControlTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/output/WarningControlTest.java) — **Reconcile**: `testSuppressMessagesSanyCode`, `testSuppressMessagesSanyErrorCodeFails`, `testSuppressMessagesTlcCode`, `testSuppressMessagesMultipleCodes`, `testSuppressMessagesUnknownCodeFails`, `testSuppressMessagesMissingArgFails`, `testMessagesAsErrorsSanyWarningCode`, `testMessagesAsErrorsTlcCode`, `testMessagesAsErrorsUnknownCodeFails`, `testMessagesAsErrorsMissingArgFails`, `testNowarningConflictWithSuppressMessages`, `testNowarningConflictWithmessagesAsErrors`, `testSameTlcCode`, `testSameSanyCode`, `testRuntimeSuppressedWarningProducesNoOutput`, `testRuntimeUnsuppressedWarningProducesOutput`, `testRuntimeWarningAsErrorThrowsTLCRuntimeException`, `testRuntimeWarningWithoutElevationDoesNotThrow`.
  Related Go checks: [tlc/output_test.go](output_test.go).

### Presentation models

Assignment, Formula, TypedSet, MCError, and MCState helpers.

- [x] [tlc2/model/FormulaTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/model/FormulaTest.java) — **Port complete**: `testUnnamed`, `testNamed`.
  Go translation: [tlc/model_test.go](model_test.go).
- [ ] [tlc2/model/MCErrorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/model/MCErrorTest.java) — **Reconcile**: `testGetErrorMessage`, `testUpdateStatesForTraceExpressions`.
  Related Go checks: [tlc/state_info_test.go](state_info_test.go), [tlc/trace_test.go](trace_test.go).
- [ ] [tlc2/model/MCStateTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/model/MCStateTest.java) — **Reconcile**: `testParseRoundTrips`, `testSimpleRecordPrinter`.
  Related Go checks: [tlc/state_info_test.go](state_info_test.go), [tlc/trace_test.go](trace_test.go).
- [x] [tlc2/model/TypedSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/model/TypedSetTest.java) — **Port complete**: `testParseSet1`, `testParseSet2`, `testParseSet3`, `testParseSet4`, `testParseSet5`, `testParseSet6`.
  Go translation: [tlc/model_test.go](model_test.go); original null input uses the nullable production boundary. Java trim semantics retained.

### Standard modules, constants, native overrides, and random values

Original module models and direct module-method tests, constant evaluation, module loading/overrides, and random value generation.

- [ ] [tlc2/module/ConstantContextTLCCacheTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/ConstantContextTLCCacheTest.java) — **Reconcile**: `test`.
  Related Go checks: [tlc/modules_tlc_ext_test.go](modules_tlc_ext_test.go).
- [ ] [tlc2/module/RandomizationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/RandomizationTest.java) — **Reconcile**: `testRandomSubsetNonFinite`, `testV1Valid`, `testV2Larger1`, `testSetNonFinite`, `testV1Negative`, `testV1NoIntValue`, `testV1Zero`, `testV2Zero`, `testV2Negative`, `testV3Empty`, `testV3AstronomicallyLarge`, `testV3isInfinite`, `testRSSV2Zero`, `testRSSV2Negative`, `testRSSV2Cardinality`, `testRSSV2TwiceCardinality`.
  Related Go checks: [tlc/random_generator_test.go](random_generator_test.go).
- [ ] [tlc2/module/SequencesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/SequencesTest.java) — **Reconcile**: `testTailString`, `testHeadString`, `testHeadStringEmpty`, `testAppendString`, `testAppendString2`, `testAppendStringNonString`, `testConcatStringToSeq`, `testConcatSeqToString`, `testConcatStringToString`, `testConcatIntToSeq`, `testConcatSeqToInt`, `testConcatIntToInt`, `testSubseq`.
  Related Go checks: [tlc/modules_sequences_test.go](modules_sequences_test.go).
- [ ] [tlc2/module/TLCExtTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/TLCExtTest.java) — **Reconcile**: `test`.
  Related Go checks: [tlc/modules_tlc_ext_test.go](modules_tlc_ext_test.go).
- [ ] [tlc2/module/TLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/module/TLCTest.java) — **Reconcile**: `testA`, `testB`, `testCombineMaxIntIntervalOnLeft`, `testCombineMaxIntIntervalOnRight`, `testPermutations`.
  Related Go checks: [tlc/modules_misc_test.go](modules_misc_test.go).
- [ ] [tlc2/tool/BagsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/BagsTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/ConstantRank1TLCEvalTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ConstantRank1TLCEvalTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/ConstantRank2AssertErrorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ConstantRank2AssertErrorTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/EmptySetEqAssumeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EmptySetEqAssumeTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/EmptySetEqStatesRcdTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EmptySetEqStatesRcdTest.java) — **Missing**: `testRcdSpec`.
- [ ] [tlc2/tool/EmptySetEqStatesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EmptySetEqStatesTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/KSubsetAssumeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/KSubsetAssumeTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/RandomElementSimulationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementSimulationTest.java) — **Missing**: `test`.
- [ ] [tlc2/tool/RandomElementT4Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementT4Test.java) — **Missing**: `test`.
- [ ] [tlc2/tool/RandomElementTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementTest.java) — **Missing**: `test`.
- [ ] [tlc2/tool/RandomElementXandYTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementXandYTest.java) — **Missing**: `test`.
- [ ] [tlc2/tool/RandomSubsetATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetATest.java) — **Missing**: `testSpec` (from `RandomSubset`).
- [ ] [tlc2/tool/RandomSubsetBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetBTest.java) — **Missing**: `testSpec` (from `RandomSubset`).
- [ ] [tlc2/tool/RandomSubsetEmptyTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetEmptyTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/RandomSubsetNextT4Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetNextT4Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/RandomSubsetNextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetNextTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/RandomSubsetNextTuplesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetNextTuplesTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/RandomSubsetSetOfFcnsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetSetOfFcnsTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/RandomSubsetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/SetPredValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/SetPredValueTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/StandardModulesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/StandardModulesTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/SubseteqNextStateTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/SubseteqNextStateTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/UserModuleOverrideAnnotationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/UserModuleOverrideAnnotationTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/UserModuleOverrideFromJarTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/UserModuleOverrideFromJarTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/UserModuleOverrideTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/UserModuleOverrideTest.java) — **Missing**: `testSpec`.

### Evaluation, initial states, next states, and action composition

Bindings, assignment, quantified evaluation, LET, INSTANCE, actions, enabledness, initial-state enumeration, and evaluation order.

- [ ] [tlc2/tool/ActionCompositionATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionCompositionATest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/ActionCompositionBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionCompositionBTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/AssignmentInitExpensiveTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssignmentInitExpensiveTest.java) — **Missing**: `test`.
- [ ] [tlc2/tool/AssignmentInitNegTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssignmentInitNegTest.java) — **Missing**: `test`.
- [ ] [tlc2/tool/AssignmentInitTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssignmentInitTest.java) — **Missing**: `test`.
- [ ] [tlc2/tool/AssignmentNext2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssignmentNext2Test.java) — **Missing**: `test`.
- [ ] [tlc2/tool/AssignmentNext3Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssignmentNext3Test.java) — **Missing**: `test`.
- [ ] [tlc2/tool/AssignmentNextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssignmentNextTest.java) — **Missing**: `test`.
- [ ] [tlc2/tool/CdotWithContextATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CdotWithContextATest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/CdotWithContextBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CdotWithContextBTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/CdotWithContextCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CdotWithContextCTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/CdotWithContextDTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CdotWithContextDTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/ChainedCdotsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ChainedCdotsTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/EmptyExistentialQuantifierTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EmptyExistentialQuantifierTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/EvalControlTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EvalControlTest.java) — **Reconcile**: `test`, `testIfEnabled`.
  Related Go checks: [tlc/eval_control_test.go](eval_control_test.go).
- [ ] [tlc2/tool/EvaluatingValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EvaluatingValueTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/LetDef1BoxedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef1BoxedTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/LetDef1BoxedbTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef1BoxedbTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/LetDef1BoxedcTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef1BoxedcTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/LetDef1BoxeddTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef1BoxeddTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/LetDef2BoxedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef2BoxedTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/LetDef2BoxedbTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef2BoxedbTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/LetDef2BoxedcTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef2BoxedcTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/LetDef2BoxeddTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/LetDef2BoxeddTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/MinimalSetOfInitStatesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/MinimalSetOfInitStatesTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/MinimalSetOfNextStatesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/MinimalSetOfNextStatesTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/SetOfStatesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/SetOfStatesTest.java) — **Reconcile**: `testSizeEmpty`, `testSize`, `testGrow`, `testIterate`, `testDuplicates`, `testDuplicatesButNotEqual`.
  Related Go checks: [tlc/set_of_states_test.go](set_of_states_test.go).
- [ ] [tlc2/tool/UndeclaredRecursionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/UndeclaredRecursionTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/evalorder/InitEvalOrder1Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/evalorder/InitEvalOrder1Test.java) — **Missing**: `test` (from `InitEvalOrderTest`).
- [ ] [tlc2/tool/evalorder/InitEvalOrder2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/evalorder/InitEvalOrder2Test.java) — **Missing**: `test` (from `InitEvalOrderTest`).
- [ ] [tlc2/tool/evalorder/InitEvalOrder3Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/evalorder/InitEvalOrder3Test.java) — **Missing**: `test` (from `InitEvalOrderTest`).
- [ ] [tlc2/tool/evalorder/InitEvalOrder4Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/evalorder/InitEvalOrder4Test.java) — **Missing**: `test` (from `InitEvalOrderTest`).
- [ ] [tlc2/tool/evalorder/InitEvalOrderBasicTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/evalorder/InitEvalOrderBasicTest.java) — **Missing**: `test`.

### Safety checking, diagnostics, and checker lifecycle

Invariants, state/action properties, assumptions/postconditions, deadlocks, DFID, views, model errors, and completion/cleanup.

- [ ] [tlc2/tool/ASTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ASTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/AbsoluteSpecPathTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AbsoluteSpecPathTest.java) — **Missing**: `test`.
- [ ] [tlc2/tool/ActionLevelPropATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionLevelPropATest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/ActionLevelPropBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionLevelPropBTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/ActionLevelPropCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionLevelPropCTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/ActionLevelPropDTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionLevelPropDTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/ActionLevelPropETest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ActionLevelPropETest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/AssertExpressionStack.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/AssertExpressionStack.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/ContinueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ContinueTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/DepthFirstTerminate.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DepthFirstTerminate.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/DiameterTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DiameterTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/DotConstrainedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DotConstrainedTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/ElevatedSanyWarning.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ElevatedSanyWarning.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/EmptyTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EmptyTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/FingerprintExceptionHangTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/FingerprintExceptionHangTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/FingerprintExceptionInitTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/FingerprintExceptionInitTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/FingerprintExceptionNextCallstackTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/FingerprintExceptionNextCallstackTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/FingerprintExceptionNextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/FingerprintExceptionNextTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/InliningTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/InliningTest.java) — **Missing**: `testSpec`.
  JVM compilation/inlining assertion needs a documented Go-specific disposition; excluded by test-dist, present in source.
- [ ] [tlc2/tool/InvParameterizedATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/InvParameterizedATest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/InvParameterizedBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/InvParameterizedBTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/InvParameterizedCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/InvParameterizedCTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/MinimumDiameterTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/MinimumDiameterTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/MonolithSpecTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/MonolithSpecTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/PostAssumptionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PostAssumptionTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/TSnapShotTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TSnapShotTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/ViewMapTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ViewMapTest.java) — **Missing**: `testSpec`.

### Issue regressions in the evaluator and checker

The original Github, Bugzilla, and CodePlex cases under tool/. Read each original model/config and assertions before selecting its implementation slice.

- [ ] [tlc2/tool/BugzillaBug279Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/BugzillaBug279Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/CodePlexBug21Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/CodePlexBug21Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1087Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1087Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1134aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1134aTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1134bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1134bTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1134cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1134cTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1134dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1134dTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1134eTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1134eTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1134fTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1134fTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1145Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1145Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1145bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1145bTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1147Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1147Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1161Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1161Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1161ViolatedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1161ViolatedTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1198aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198aTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1198bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198bTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1198cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198cTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1198dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198dTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1198fTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198fTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1198hTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1198hTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1244Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1244Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1244bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1244bTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1244cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1244cTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1302Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1302Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1302bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1302bTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1302cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1302cTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1389CountingTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389CountingTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1389LoopsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389LoopsTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1389StateGuardTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389StateGuardTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1389Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1389ViolatedBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389ViolatedBTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1389ViolatedCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389ViolatedCTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github1389ViolatedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1389ViolatedTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github179aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github179aTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github179bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github179bTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github179cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github179cTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github362Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github362Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github391Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github391Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github407Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github407Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github432Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github432Test.java) — **Missing**: `testA`, `testB`, `testC`, `testD`.
- [ ] [tlc2/tool/Github461Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github461Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github525Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github525Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github597Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github597Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github648Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github648Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github648wNTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github648wNTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github652Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github652Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github680aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github680aTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github680bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github680bTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github680cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github680cTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github687Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github687fifoTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687fifoTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github687specATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687specATest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github687specBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687specBTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github687specCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687specCTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github687specDTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687specDTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github687specInitNextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687specInitNextTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github687specPrimeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github687specPrimeTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github696Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github696Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github696bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github696bTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github715Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github715Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github715bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github715bTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github715cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github715cTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github715dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github715dTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github725Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github725bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725bTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github725cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725cTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github725dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725dTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github725eTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725eTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github725fTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725fTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github725gTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725gTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github725hTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github725hTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github726Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github726Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github742Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github742Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github743Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github743Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github746Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github746Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github757Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github757Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github766SimulateTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github766SimulateTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github766Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github766Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github798ITest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github798ITest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github798NTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github798NTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github807Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github807Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github817Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github817bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817bTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github817cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817cTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github817dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817dTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github817eTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github817eTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github819Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github819Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github849Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github849Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github858Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github858Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github866Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github866Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github971aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971aTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github971bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971bTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github971cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971cTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github971dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971dTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github971eTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github971eTest.java) — **Missing**: `testSpec`.

### Traces, aliases, dump/load, and generated trace specs

Trace reconstruction, alias evaluation, trace races, external trace serialization, and generated trace-expression specifications.

- [ ] [tlc2/tool/DistributedTrace.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DistributedTrace.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/DumpAsDotTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DumpAsDotTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/DumpLoadTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DumpLoadTraceTest.java) — **Partial**: `testLivenessEWD840MC3DumpLoadTraceTLC`, `testLivenessEWD840MC3DumpLoadTraceTLCAutoWorkers`.
  Already mapped: `testLivenessMCDumpLoadTraceJSON`, `testLivenessMCDumpLoadTraceTLC`, `testLivenessMCDumpLoadTraceJSONAutoWorkers`, `testLivenessMCDumpLoadTraceTLCAutoWorkers`, `testSafetyDumpLoadTraceJSON`, `testSafetyDumpLoadTraceTLC`, `testSafetyDumpLoadTraceJSONAutoWorkers`, `testSafetyDumpLoadTraceTLCAutoWorkers`, `testLivenessBidirectionalDumpLoadTraceJSON`, `testLivenessBidirectionalDumpLoadTraceTLC`, `testLivenessBidirectionalDumpLoadTraceJSONAutoWorkers`, `testLivenessBidirectionalDumpLoadTraceTLCAutoWorkers`, `testSafetyTESpecEqAliasDumpLoadTraceJSON`, `testSafetyTESpecEqAliasDumpLoadTraceTLC`, `testSafetyTESpecEqAliasDumpLoadTraceJSONAutoWorkers`, `testSafetyTESpecEqAliasDumpLoadTraceTLCAutoWorkers`, `testLivenessExample1DumpLoadTraceJSON`, `testLivenessExample1DumpLoadTraceTLC`, `testLivenessExample1DumpLoadTraceJSONAutoWorkers`, `testLivenessExample1DumpLoadTraceTLCAutoWorkers`, `testSafetyDieHardAliasSubDumpLoadTraceJSON`, `testSafetyDieHardAliasSubDumpLoadTraceTLC`, `testSafetyDieHardAliasSub2DumpLoadTraceJSON`, `testSafetyDieHardAliasSub2DumpLoadTraceTLC`, `testSafetyDieHardAliasSub2DumpLoadTraceJSONAutoWorkers`, `testSafetyDieHardAliasSub2DumpLoadTraceTLCAutoWorkers`, `testSafetyDieHardAliasSupDumpLoadTraceJSON`, `testSafetyDieHardAliasSupDumpLoadTraceTLC`, `testSafetyDieHardAliasSupDumpLoadTraceJSONAutoWorkers`, `testSafetyDieHardAliasSupDumpLoadTraceTLCAutoWorkers`.
  Two enabled EWD840 binary methods remain; three original @Ignore methods already have Go skips.

### Generated TTrace recheck variants

Every concrete *_TTraceTest and *_TTrace class, including inherited testSpec. These require the original first run and generated trace-spec artifacts; do not replace them with a rerun of the source model.

- [ ] [tlc2/tool/BugzillaBug279Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/BugzillaBug279Test_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/DepthFirstDieHardTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DepthFirstDieHardTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/DepthFirstErrorTraceTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DepthFirstErrorTraceTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/EvalExceptionTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EvalExceptionTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github461Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github461Test_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/Github597Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github597Test_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/PrintTraceRaceTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PrintTraceRaceTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/RandomElementSimulationTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementSimulationTest_TTraceTest.java) — **Missing**: `test`.
- [ ] [tlc2/tool/RandomElementT4Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementT4Test_TTraceTest.java) — **Missing**: `test`.
- [ ] [tlc2/tool/RandomElementTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementTest_TTraceTest.java) — **Missing**: `test`.
- [ ] [tlc2/tool/RandomElementXandYTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomElementXandYTest_TTraceTest.java) — **Missing**: `test`.
- [ ] [tlc2/tool/RandomSubsetATest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetATest_TTraceTest.java) — **Missing**: `testSpec` (from `RandomSubset_TTrace`).
- [ ] [tlc2/tool/RandomSubsetBTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetBTest_TTraceTest.java) — **Missing**: `testSpec` (from `RandomSubset_TTrace`).
- [ ] [tlc2/tool/RandomSubsetNextT4Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetNextT4Test_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/RandomSubsetNextTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetNextTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/RandomSubsetTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/RandomSubsetTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/TLCGetLevelTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCGetLevelTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/TraceWithLargeSetOfInitialStatesTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TraceWithLargeSetOfInitialStatesTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/ViewMapTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/ViewMapTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/checkpoint/CheckpointOnViolationTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/checkpoint/CheckpointOnViolationTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/BidirectionalTransitions1BxTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions1BxTest_TTraceTest.java) — **Missing**: `testSpec` (from `BidirectionalTransitions1B_TTrace`).
- [ ] [tlc2/tool/liveness/BidirectionalTransitions1ByTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions1ByTest_TTraceTest.java) — **Missing**: `testSpec` (from `BidirectionalTransitions1B_TTrace`).
- [ ] [tlc2/tool/liveness/BidirectionalTransitions2CxTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions2CxTest_TTraceTest.java) — **Missing**: `testSpec` (from `BidirectionalTransitions2C_TTrace`).
- [ ] [tlc2/tool/liveness/BidirectionalTransitions2CyTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions2CyTest_TTraceTest.java) — **Missing**: `testSpec` (from `BidirectionalTransitions2C_TTrace`).
- [ ] [tlc2/tool/liveness/ChooseTableauSymmetryTestA_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ChooseTableauSymmetryTestA_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08AgentRingTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08AgentRingTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08EWD840FL1Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL1Test_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08EWD840FL2Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL2Test_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08EWD840FL3Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL3Test_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08EWD840FL4Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL4Test_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08Test_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08aTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08aTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ErrorTraceConstructionTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ErrorTraceConstructionTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/LoopTestForcedPartial_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LoopTestForcedPartial_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/LoopTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LoopTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/OneBitMutexNoSymmetryTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/OneBitMutexNoSymmetryTest_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Test3_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Test3_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/UnsymmetricModelCheckerTestA_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/UnsymmetricModelCheckerTestA_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/simulation/Example1Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/Example1Test_TTraceTest.java) — **Missing**: `testSpec` (from `AbstractExample_TTrace`).
- [ ] [tlc2/tool/liveness/simulation/Example2Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/Example2Test_TTraceTest.java) — **Missing**: `testSpec` (from `AbstractExample_TTrace`).
- [ ] [tlc2/tool/liveness/simulation/LiveCheckExample1Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckExample1Test_TTraceTest.java) — **Missing**: `testSpec` (from `AbstractExample_TTrace`).
- [ ] [tlc2/tool/liveness/simulation/LiveCheckExample2Test_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckExample2Test_TTraceTest.java) — **Missing**: `testSpec` (from `AbstractExample_TTrace`).
- [ ] [tlc2/tool/liveness/simulation/LiveCheckSimulationTest2a_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckSimulationTest2a_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/simulation/SimulationTest2a_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/SimulationTest2a_TTraceTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/simulation/StutteringTest_TTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/StutteringTest_TTraceTest.java) — **Missing**: `testSpec`.

### Liveness and fairness model regressions

Temporal semantics, fairness, constraints, symmetry, double negation, counterexample construction, and original Examples models.

- [ ] [tlc2/tool/NoFairnessButLivePropATest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropATest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/NoFairnessButLivePropBTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropBTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/NoFairnessButLivePropDTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropDTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/NoFairnessButLivePropGTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropGTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/NoFairnessButLivePropHTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropHTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/NoFairnessButLivePropNoWarningCustomFairTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropNoWarningCustomFairTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/NoFairnessButLivePropNoWarningNoFairTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropNoWarningNoFairTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/NoFairnessButLivePropNoWarningWithFairTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropNoWarningWithFairTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/NoFairnessButLivePropOTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropOTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/NoFairnessButLivePropPTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/NoFairnessButLivePropPTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/PossibleFailActionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PossibleFailActionTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/PossibleFailMixedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PossibleFailMixedTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/PossibleFailNoTransTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PossibleFailNoTransTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/PossibleFailStateTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PossibleFailStateTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/PossibleTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PossibleTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/BidirectionalTransitions1BxTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions1BxTest.java) — **Missing**: `testSpec` (from `BidirectionalTransitions1BTest`).
- [ ] [tlc2/tool/liveness/BidirectionalTransitions1ByTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions1ByTest.java) — **Missing**: `testSpec` (from `BidirectionalTransitions1BTest`).
- [ ] [tlc2/tool/liveness/BidirectionalTransitions1Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions1Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/BidirectionalTransitions2CxTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions2CxTest.java) — **Missing**: `testSpec` (from `BidirectionalTransitions2CTest`).
- [ ] [tlc2/tool/liveness/BidirectionalTransitions2CyTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions2CyTest.java) — **Missing**: `testSpec` (from `BidirectionalTransitions2CTest`).
- [ ] [tlc2/tool/liveness/BidirectionalTransitions2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/BidirectionalTransitions2Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ChooseTableauSymmetryTestA.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ChooseTableauSymmetryTestA.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08AgentRing790Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08AgentRing790Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08AgentRingTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08AgentRingTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08EWD840FL1Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL1Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08EWD840FL2FromCheckpointTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL2FromCheckpointTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08EWD840FL2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL2Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08EWD840FL3Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL3Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08EWD840FL4Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08EWD840FL4Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/CodePlexBug08aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/CodePlexBug08aTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/EmptyOrderOfSolutionsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/EmptyOrderOfSolutionsTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ErrorTraceConstructionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ErrorTraceConstructionTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesACPNBTLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesACPNBTLCTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesAbaAsynByzTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesAbaAsynByzTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesAsyncTerminationDetectionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesAsyncTerminationDetectionTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesBcastByzTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesBcastByzTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesBcastFolkloreTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesBcastFolkloreTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesBlockingQueuePoisonAppleTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesBlockingQueuePoisonAppleTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesBufferedRandomAccessFileTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesBufferedRandomAccessFileTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesCf1sFolkloreTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesCf1sFolkloreTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesCoffeeCan100BeansTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesCoffeeCan100BeansTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesEWD840Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesEWD840Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesEWD998ChanIDTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesEWD998ChanIDTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesEWD998Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesEWD998Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesEnvironmentControllerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesEnvironmentControllerTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesHuangTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesHuangTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesLiveHourClockTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesLiveHourClockTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesLockHSTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesLockHSTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesMCAlternatingBitTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCAlternatingBitTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesMCDistributedReplicatedLogTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCDistributedReplicatedLogTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesMCEWD687aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCEWD687aTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesMCLiveInternalMemoryTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCLiveInternalMemoryTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesMCLiveWriteThroughCacheTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCLiveWriteThroughCacheTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesMCWriteThroughCacheTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCWriteThroughCacheTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesMCYoYoNoPruningTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesMCYoYoNoPruningTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesNbacgGuer01Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesNbacgGuer01Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesPrisonersTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesPrisonersTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesSDPAttackNewSolutionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSDPAttackNewSolutionTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesSDPAttackTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSDPAttackTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesSchedulingAllocatorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSchedulingAllocatorTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesSimpleAllocatorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSimpleAllocatorTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesSingleLaneBridgeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSingleLaneBridgeTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesSpanTreeRandomTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSpanTreeRandomTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesSpanTreeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSpanTreeTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/ExamplesSyncTerminationDetectionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/ExamplesSyncTerminationDetectionTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Github1037Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github1037Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Github317Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github317Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Github317aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github317aTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Github604Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github604Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Github702Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github702Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Github710aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710aTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Github710bTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710bTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Github710cTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710cTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Github710dFairTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710dFairTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Github710dTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710dTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Github710eFairTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710eFairTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Github710eTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github710eTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/Github790Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Github790Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/IncompatibleTypesLiveTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/IncompatibleTypesLiveTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/InitialLivenessEvaluationErrorInvariantOnlyTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/InitialLivenessEvaluationErrorInvariantOnlyTest.java) — **Missing**: `testSpec` (from `InitialLivenessEvaluationErrorTest`).
- [ ] [tlc2/tool/liveness/InitialLivenessEvaluationErrorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/InitialLivenessEvaluationErrorTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/LivenessSymmetryWarning.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LivenessSymmetryWarning.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/LoopTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LoopTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/LoopTestForcedPartial.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LoopTestForcedPartial.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/LoopTestWeakFair.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LoopTestWeakFair.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/NoSymmetryTableauModelCheckerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/NoSymmetryTableauModelCheckerTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/OneBitMutexNoSymmetryTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/OneBitMutexNoSymmetryTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/TemporalDoubleNegationAlwaysDoubleNegationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationAlwaysDoubleNegationTest.java) — **Missing**: `testValidProperty` (from `AbstractTemporalDoubleNegationTest`).
- [ ] [tlc2/tool/liveness/TemporalDoubleNegationAlwaysDualTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationAlwaysDualTest.java) — **Missing**: `testValidProperty` (from `AbstractTemporalDoubleNegationTest`).
- [ ] [tlc2/tool/liveness/TemporalDoubleNegationEventuallyDoubleNegationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationEventuallyDoubleNegationTest.java) — **Missing**: `testValidProperty` (from `AbstractTemporalDoubleNegationTest`).
- [ ] [tlc2/tool/liveness/TemporalDoubleNegationFairEventuallyDoubleNegationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationFairEventuallyDoubleNegationTest.java) — **Missing**: `testValidProperty` (from `AbstractTemporalDoubleNegationTest`).
- [ ] [tlc2/tool/liveness/TemporalDoubleNegationFairLeadsToDualTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationFairLeadsToDualTest.java) — **Missing**: `testValidProperty` (from `AbstractTemporalDoubleNegationTest`).
- [ ] [tlc2/tool/liveness/TemporalDoubleNegationImplicationDoubleNegationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationImplicationDoubleNegationTest.java) — **Missing**: `testValidProperty` (from `AbstractTemporalDoubleNegationTest`).
- [ ] [tlc2/tool/liveness/TemporalDoubleNegationImplicationTautologyTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationImplicationTautologyTest.java) — **Missing**: `testValidProperty` (from `AbstractTemporalDoubleNegationTest`).
- [ ] [tlc2/tool/liveness/TemporalDoubleNegationLeadsToDualTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TemporalDoubleNegationLeadsToDualTest.java) — **Missing**: `testValidProperty` (from `AbstractTemporalDoubleNegationTest`).
- [ ] [tlc2/tool/liveness/Test3.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/Test3.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/TwoPhaseCommitTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TwoPhaseCommitTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/UnsymmetricModelCheckerTestA.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/UnsymmetricModelCheckerTestA.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/UnsymmetricModelCheckerTestB.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/UnsymmetricModelCheckerTestB.java) — **Missing**: `testSpec`.

### Liveness graph, tableau, and expression helpers

Direct Java graph, node/table, particle-closure, and live-expression tests.

- [ ] [tlc2/tool/liveness/DiskGraphTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/DiskGraphTest.java) — **Reconcile**: `testGetPathWithoutInitNoTableau`, `testGetMinimalPathWithoutTableau`, `testPathWithTwoInitNodes`, `testAddSameGraphNodeTwice`, `testLookupExistingNode`, `testAddSameGraphNodeTwiceCorrectSuccessors`, `testGetPathPartialGraph`.
  Related Go checks: [tlc/liveness_graph_test.go](liveness_graph_test.go).
- [ ] [tlc2/tool/liveness/GraphNodeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/GraphNodeTest.java) — **Reconcile**: `testAllocateRealign`, `testRealign`, `testAllocateNested`, `testAllocateNestedRandom`, `testAllocateNegative`, `testAllocateAndSuccessorSize`.
  Related Go checks: [tlc/liveness_graph_test.go](liveness_graph_test.go).
- [ ] [tlc2/tool/liveness/LiveExprNodeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/LiveExprNodeTest.java) — **Reconcile**: `testLNBool`, `testLNState`.
  Related Go checks: [tlc/liveness_process_test.go](liveness_process_test.go).
- [ ] [tlc2/tool/liveness/TBParTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TBParTest.java) — **Reconcile**: `testParticleClosureInconsistentConstantLevel`, `testParticleClosureInconsistentStateLevel`, `testParticleClosureConsistentConstantLevel`, `testParticleClosureConsistentStateLevel`, `testParticleClosureExampleConstLevel`, `testParticleClosureExampleStateLevel`.
  Related Go checks: [tlc/liveness_process_test.go](liveness_process_test.go).
- [ ] [tlc2/tool/liveness/TableauDiskGraphTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TableauDiskGraphTest.java) — **Reconcile**: `testGetPathWithoutInitNoTableau` (from `DiskGraphTest`), `testGetMinimalPathWithoutTableau` (from `DiskGraphTest`), `testPathWithTwoInitNodes` (from `DiskGraphTest`), `testAddSameGraphNodeTwice` (from `DiskGraphTest`), `testLookupExistingNode` (from `DiskGraphTest`), `testAddSameGraphNodeTwiceCorrectSuccessors` (from `DiskGraphTest`), `testGetPathPartialGraph` (from `DiskGraphTest`), `testGetShortestPath`, `testUnifyingNodeInPath`, `testUnifyingNodeShortestPath`, `testPathWithTwoInitNodesWithTableau`, `testGetPathWithTwoInits`, `testNodeSetDone`, `testGetPathWithTwoNodesWithSameFingerprint`, `testLookupExistingNodeWithTidx`, `testWhatsDoneIsDoneRRS`, `testWhatsDoneIsDoneRSR`, `testWhatsDoneIsDoneSRR`.
  Related Go checks: [tlc/liveness_graph_test.go](liveness_graph_test.go).
- [ ] [tlc2/tool/liveness/TableauNodePtrTableTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/TableauNodePtrTableTest.java) — **Reconcile**: `testSetDoneBFSOrder`, `testSetDoneNoOrder`, `testSetDone`, `testSetDone2`, `testSetDone3`, `testIsDoneSPP`, `testIsDonePPS`, `testIsDonePSP`, `testRedundantMethodYieldSameResult`.
  Related Go checks: [tlc/liveness_graph_test.go](liveness_graph_test.go).

### Simulation and multithreaded simulation

Both tool/ and tool/simulation/ simulator/worker classes, original simulation models, and liveness/simulation models. Preserve seeds, termination predicates, trace assertions, and worker counts.

- [ ] [tlc2/tool/SimulationWorkerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/SimulationWorkerTest.java) — **Reconcile**: `testGetTraceTLCState0`, `testGetTraceTLCState1`, `testGetTraceTLCState2`, `testGetTraceTLCState3`, `testGetTraceTLCState4`, `testGetTraceTLCState5`, `testGetTraceTLCState6`.
  Related Go checks: [tlc/simulator_test.go](simulator_test.go), [tlc/worker_test.go](worker_test.go).
- [ ] [tlc2/tool/SimulatorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/SimulatorTest.java) — **Reconcile**: `testPrintBehaviorShouldPrintErrorState`.
  Related Go checks: [tlc/simulator_test.go](simulator_test.go).
- [ ] [tlc2/tool/liveness/simulation/Example1Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/Example1Test.java) — **Missing**: `testSpec` (from `AbstractExampleTestCase`).
- [ ] [tlc2/tool/liveness/simulation/Example2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/Example2Test.java) — **Missing**: `testSpec` (from `AbstractExampleTestCase`).
- [ ] [tlc2/tool/liveness/simulation/LiveCheckExample1Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckExample1Test.java) — **Missing**: `testSpec` (from `AbstractExampleTestCase`).
- [ ] [tlc2/tool/liveness/simulation/LiveCheckExample2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckExample2Test.java) — **Missing**: `testSpec` (from `AbstractExampleTestCase`).
- [ ] [tlc2/tool/liveness/simulation/LiveCheckSimulationTest2.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckSimulationTest2.java) — **Missing**: `testSpec` (from `SuccessfulSimulationTestCase`).
- [ ] [tlc2/tool/liveness/simulation/LiveCheckSimulationTest2a.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/LiveCheckSimulationTest2a.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/simulation/SimulationTest2.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/SimulationTest2.java) — **Missing**: `testSpec` (from `SuccessfulSimulationTestCase`).
- [ ] [tlc2/tool/liveness/simulation/SimulationTest2PostCondition.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/SimulationTest2PostCondition.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/simulation/SimulationTest2a.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/SimulationTest2a.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/simulation/SimulationTestAssumption.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/SimulationTestAssumption.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/liveness/simulation/StutteringTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/liveness/simulation/StutteringTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/simulation/Github1191Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/Github1191Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/simulation/Github1191aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/Github1191aTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/simulation/Github602Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/Github602Test.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/simulation/NQSpecTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/NQSpecTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/simulation/SimulationWorkerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/SimulationWorkerTest.java) — **Reconcile**: `testSuccessfulRun`, `testInvariantViolation`, `testActionPropertyViolation`, `testInvariantBadEval`, `testActionPropertyBadEval`, `testUnderspecifiedNext`, `testDeadlock`, `testModelStateConstraint`, `testModelActionConstraint`, `testWorkerInterruption`, `testTraceDepthObeyed`, `testStateAndTraceGenerationCount`.
  Related Go checks: [tlc/simulator_test.go](simulator_test.go).
- [ ] [tlc2/tool/simulation/SimulatorMultiThreadTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/SimulatorMultiThreadTest.java) — **Reconcile**: `testSuccessfulSimulation` (from `SimulatorTest`), `testInvariantViolationInitialState` (from `SimulatorTest`), `testInvariantViolation` (from `SimulatorTest`), `testInvariantBadEvalInitState` (from `SimulatorTest`), `testInvariantBadEvalNonInitState` (from `SimulatorTest`), `testUnderspecifiedInit` (from `SimulatorTest`), `testInvariantViolationContinue` (from `SimulatorTest`), `testDontContinueOnRuntimeSpecError` (from `SimulatorTest`), `testLivenessViolation` (from `SimulatorTest`), `testLivenessViolationIgnoresContinue` (from `SimulatorTest`).
  Related Go checks: [tlc/simulator_test.go](simulator_test.go).
- [ ] [tlc2/tool/simulation/SimulatorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/simulation/SimulatorTest.java) — **Reconcile**: `testSuccessfulSimulation`, `testInvariantViolationInitialState`, `testInvariantViolation`, `testInvariantBadEvalInitState`, `testInvariantBadEvalNonInitState`, `testUnderspecifiedInit`, `testInvariantViolationContinue`, `testDontContinueOnRuntimeSpecError`, `testLivenessViolation`, `testLivenessViolationIgnoresContinue`.
  Related Go checks: [tlc/simulator_test.go](simulator_test.go).

### Coverage

All original coverage model tests and OpApplNodeWrapper reporting methods have explicit translations.

No remaining non-`@Ignore` original methods identified in this topic. See the mapping appendix for the translations.

### Debugger and scoped identifiers

All 16 concrete debug/ test classes have explicit translations; this does not establish completion of DAP transport or debugger features outside those methods.

No remaining non-`@Ignore` original methods identified in this topic. See the mapping appendix for the translations.

### Checkpoint and recovery models

Original checkpoint-on-violation and time-bound models; generated recheck variants are listed with TTrace tests.

- [ ] [tlc2/tool/checkpoint/CheckpointOnViolationTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/checkpoint/CheckpointOnViolationTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/checkpoint/CheckpointWhenTimeBoundTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/checkpoint/CheckpointWhenTimeBoundTest.java) — **Missing**: `testSpec`.

### Distributed TLC

Remote server/worker integration, init failures, fingerprint-manager failover, and smart-proxy calculations. Full wire transport is still an implementation prerequisite for transport tests.

- [ ] [tlc2/tool/distributed/DieHardDistributedTLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/DieHardDistributedTLCTest.java) — **Missing**: `testSpec`.
  Upstream DistributedTLCTestCase.setUp unconditionally Assume.assumeTrue(false); retain as transport backlog, not a completed/skipped Go port.
- [ ] [tlc2/tool/distributed/DistributedDoInitFunctorInvariantContinueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/DistributedDoInitFunctorInvariantContinueTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/distributed/DistributedDoInitFunctorInvariantTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/DistributedDoInitFunctorInvariantTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/distributed/EWD840DistributedTLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/EWD840DistributedTLCTest.java) — **Missing**: `test`.
  Upstream DistributedTLCTestCase.setUp unconditionally Assume.assumeTrue(false); retain as transport backlog, not a completed/skipped Go port.
- [ ] [tlc2/tool/distributed/EWD840DistributedWithFPSetTLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/EWD840DistributedWithFPSetTLCTest.java) — **Missing**: `test`.
  Upstream DistributedTLCTestCase.setUp unconditionally Assume.assumeTrue(false); retain as transport backlog, not a completed/skipped Go port.
- [ ] [tlc2/tool/distributed/TLCSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/TLCSetTest.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/distributed/TSnapShotDistributedTLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/distributed/TSnapShotDistributedTLCTest.java) — **Missing**: `test`.
  Upstream DistributedTLCTestCase.setUp unconditionally Assume.assumeTrue(false); retain as transport backlog, not a completed/skipped Go port.

### Fingerprint sets, indexers, arrays, and iterators

Disk/memory/off-heap factories, recovery, duplicate merging, high/low fingerprints, indexers, arrays, and iterators. Preserve parameter matrices and original resource/concurrency behavior.

- [ ] [tlc2/tool/fp/Bug210DiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/Bug210DiskFPSetTest.java) — **Reconcile**: `testDiskLookupWithOverflow`.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [ ] [tlc2/tool/fp/Bug242DiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/Bug242DiskFPSetTest.java) — **Reconcile**: `testDiskFPSetWithHighMem`, `testDiskFPSetIntMaxValue`, `testDiskFPSetIntMinValue`, `testDiskFPSetZero`, `testDiskFPSetOne`.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [ ] [tlc2/tool/fp/Bug246DiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/Bug246DiskFPSetTest.java) — **Reconcile**: `testLinearFillup`, `testFlushDiskFPSet`.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [ ] [tlc2/tool/fp/FPSetFactoryTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/FPSetFactoryTest.java) — **Reconcile**: `testGetDiskFPSet`, `testGetFPSetMSB`, `testGetFPSetLSB`, `testGetFPSetOffHeap`, `testGetFPSetMSBWithMem`, `testGetFPSetLSBWithMem`, `testGetFPSetMSBWithMemAndRatio`, `testGetFPSetLSBWithMemAndRatio`, `testGetFPSetMultiFPSet`, `testGetFPSetLSBMultiFPSet`, `testGetFPSetOffHeapMultiFPSet`, `testGetFPSetMultiFPSetWithMem`, `testGetFPSetLSBMultiFPSetWithMem`, `testGetFPSetOffHeapMultiFPSetWithMem`, `testGetFPSetOffHeapMultiFPSet42`.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [ ] [tlc2/tool/fp/LSBDiskFPsetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/LSBDiskFPsetTest.java) — **Reconcile**: `testCtorLLMinus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorLL` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorLLPlus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorLLNextPow2Min1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16Minus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16Plus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16NextPow2Min1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorULMinus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorUL` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorULPlus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorULNextPow2Min1` (from `AbstractHeapBasedDiskFPSetTest`), `testFPSetRecovery` (from `AbstractHeapBasedDiskFPSetTest`), `testFPSetRecovery2` (from `AbstractHeapBasedDiskFPSetTest`), `testFPSetRecoveryDuplicate` (from `AbstractHeapBasedDiskFPSetTest`).
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [ ] [tlc2/tool/fp/LongArrayTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/LongArrayTest.java) — **Reconcile**: `testGetAndSet`, `testOutOfRangePositive`, `testOutOfRangeNegative`, `testGetAndTrySet`, `testZeroMemory`, `testSwap`, `testSwapRandom`.
  Related Go checks: [tlc/long_array_test.go](long_array_test.go).
- [ ] [tlc2/tool/fp/LongArraysTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/LongArraysTest.java) — **Reconcile**: `testEmpty1`, `testEmpty2`, `testBasic1`, `testBasic2`, `test0`, `testIsInRange`.
  Related Go checks: [tlc/long_array_test.go](long_array_test.go).
- [ ] [tlc2/tool/fp/MSBDiskFPSetTest2.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/MSBDiskFPSetTest2.java) — **Reconcile**: `testCtorLLMinus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorLL` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorLLPlus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorLLNextPow2Min1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16Minus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16Plus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorPow16NextPow2Min1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorULMinus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorUL` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorULPlus1` (from `AbstractHeapBasedDiskFPSetTest`), `testCtorULNextPow2Min1` (from `AbstractHeapBasedDiskFPSetTest`), `testFPSetRecovery` (from `AbstractHeapBasedDiskFPSetTest`), `testFPSetRecovery2` (from `AbstractHeapBasedDiskFPSetTest`), `testFPSetRecoveryDuplicate` (from `AbstractHeapBasedDiskFPSetTest`), `testGetLast`, `testHighFingerprint1`, `testHighFingerprint2`, `testGetLastNoBuckets`.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [ ] [tlc2/tool/fp/OffHeapBitshiftingIndexerParameterizedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapBitshiftingIndexerParameterizedTest.java) — **Reconcile**; parameterized: preserve all original rows: `testZero` (from `OffHeapIndexerParameterizedTest`), `testOne` (from `OffHeapIndexerParameterizedTest`), `testLongMin` (from `OffHeapIndexerParameterizedTest`), `testLongMax` (from `OffHeapIndexerParameterizedTest`), `testSome` (from `OffHeapIndexerParameterizedTest`).
  Related Go checks: [tlc/offheap_indexer_test.go](offheap_indexer_test.go).
- [ ] [tlc2/tool/fp/OffHeapBitshiftingIndexerTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapBitshiftingIndexerTest.java) — **Reconcile**: `testBitshifting`, `testBitshifting2`, `testShift1_268435456`, `testBitshiftOvershoot`, `testNoOverflowErrorBitShifting`.
  Related Go checks: [tlc/offheap_indexer_test.go](offheap_indexer_test.go).
- [ ] [tlc2/tool/fp/OffHeapDiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapDiskFPSetTest.java) — **Reconcile**: `testInsertAndEvict1`, `testInsertAndEvict2`, `testInsertAndEvict3`, `testInsertAndEvict4`, `testInsertAndEvict5`, `testInsertAndEvict6`, `testInsertAndEvict7`, `testInsertAndEvict8`, `testInsertAndEvict9`, `testInsertAndEvict10`, `testInsertAndEvict11`, `testInsertAndEvict12`, `testInsertAndEvict13`, `testInsertAndEvict14`, `testInsertAndEvict15`, `testInsertAndEvict16`, `testOffset1Page`, `testOffset3Page`, `testOffset5Page`, `testOffset9Page`, `testWriteIndex`, `testMergeDuplicate`, `testMergeDistinct`.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [ ] [tlc2/tool/fp/OffHeapIndexerEquivalenceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapIndexerEquivalenceTest.java) — **Reconcile**; parameterized: preserve all original rows: `testInfiniteInfMult`.
  Related Go checks: [tlc/offheap_indexer_test.go](offheap_indexer_test.go).
- [ ] [tlc2/tool/fp/OffHeapInfPrecisionIndexerParameterizedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapInfPrecisionIndexerParameterizedTest.java) — **Reconcile**; parameterized: preserve all original rows: `testZero` (from `OffHeapIndexerParameterizedTest`), `testOne` (from `OffHeapIndexerParameterizedTest`), `testLongMin` (from `OffHeapIndexerParameterizedTest`), `testLongMax` (from `OffHeapIndexerParameterizedTest`), `testSome` (from `OffHeapIndexerParameterizedTest`).
  Related Go checks: [tlc/offheap_indexer_test.go](offheap_indexer_test.go).
- [ ] [tlc2/tool/fp/OffHeapIteratorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapIteratorTest.java) — **Reconcile**: `testNext`, `testMarkNext`.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [ ] [tlc2/tool/fp/OffHeapMult1024IndexerParameterizedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/OffHeapMult1024IndexerParameterizedTest.java) — **Reconcile**; parameterized: preserve all original rows: `testZero` (from `OffHeapIndexerParameterizedTest`), `testOne` (from `OffHeapIndexerParameterizedTest`), `testLongMin` (from `OffHeapIndexerParameterizedTest`), `testLongMax` (from `OffHeapIndexerParameterizedTest`), `testSome` (from `OffHeapIndexerParameterizedTest`).
  Related Go checks: [tlc/offheap_indexer_test.go](offheap_indexer_test.go).
- [ ] [tlc2/tool/fp/ShortDiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/ShortDiskFPSetTest.java) — **Reconcile**: `testWithoutZeroFP`, `testWithoutMinFP`, `testWithoutMaxFP`, `testZeroFP`, `testMinFP`, `testMinMin1FP`, `testNeg1FP`, `testPos1FP`, `testMaxFP`, `testValues`, `testDiskLookupWithFpOnLoPage`, `testMemLookupWithZeros`, `testMemLookupWithMin`, `testMemLookupWithMax`, `testDiskLookupWithZeros`, `testDiskLookupWithMin`, `testDiskLookupWithMax`, `testDiskLookupWithMaxOnPage`, `testDiskLookupWithZerosOnPage`, `testDiskLookupWithLongMinValueOnPage`, `testComparePutAndPutBlock`, `testCompareContainsAndContainsBlock`, `testContainsBlock`, `testPutBlock`.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [ ] [tlc2/tool/fp/iterator/TLCIterator1Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/iterator/TLCIterator1Test.java) — **Reconcile**: `testNext` (from `TLCIteratorTest`), `testNoNext` (from `TLCIteratorTest`), `testGetLast` (from `TLCIteratorTest`).
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [ ] [tlc2/tool/fp/iterator/TLCIterator2Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/iterator/TLCIterator2Test.java) — **Reconcile**: `testNext` (from `TLCIteratorTest`), `testNoNext` (from `TLCIteratorTest`), `testGetLast` (from `TLCIteratorTest`).
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).
- [ ] [tlc2/tool/fp/iterator/TLCIteratorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/fp/iterator/TLCIteratorTest.java) — **Reconcile**: `testNext`, `testNoNext`, `testGetLast`.
  Related Go checks: [tlc/fpset_test.go](fpset_test.go).

### Queues and pool writers

Memory-state queues and disk/byte-array writer wakeup/finish behavior; stress and JPF suites are listed separately below.

- [ ] [tlc2/tool/queue/DiskPoolWriterTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/queue/DiskPoolWriterTest.java) — **Reconcile**: `testStatePoolWriterIgnoresEmptyWakeAndStopsOnFinish`, `testByteArrayPoolWriterIgnoresEmptyWakeAndStopsOnFinish`.
  Related Go checks: [tlc/queue_test.go](queue_test.go).
- [x] [tlc2/tool/queue/StateQueueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/queue/StateQueueTest.java) — **Port complete**: `testEnqueue`, `testsDequeueEmpty`, `testDequeueEmpty`, `testsDequeueNotEmpty`, `testDequeueNotEmpty`, `testEnqueueAddNotSame`, `testEnqueueAddSame`, `testsDequeueAbuseEmpty`, `testsDequeueAbuseNonEmpty`.
  Go translation: [tlc/queue_test.go](queue_test.go).

### Values, lazy functions, enumeration, and value streams

Original primitive/composite value, normalization, comparison, EXCEPT, serialization, subset, and random-enumerator cases.

- [ ] [tlc2/value/StringDeserializeTLCTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/StringDeserializeTLCTest.java) — **Missing**: `test`.
- [ ] [tlc2/value/ValueInputOutputStreamTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/ValueInputOutputStreamTest.java) — **Reconcile**: `testWriteShort`, `testWriteInt`, `testWriteShortNat`, `testWriteNat`, `testBlindReadStringValue`, `testBlindReadRecordValue`.
  Related Go checks: [tlc/value_stream_test.go](value_stream_test.go).
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
- [ ] [tlc2/value/impl/KSubsetValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/KSubsetValueTest.java) — **Reconcile**: `testEnumerateN32`, `testEnumerateN33`, `testEnumerateN63`, `testEnumerateN64`, `testNormalization`, `testKSubsetFingerprintingS009`, `testKSubsetFingerprintingS032`, `testKSubsetFingerprintingS033`, `testKSubsetFingerprintingS063`, `testInvalidKDenotesEmptySet`, `testToStringLargeSwallowsCountError`.
  Related Go checks: [tlc/value_ksubset_test.go](value_ksubset_test.go).
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
- [ ] [tlc2/value/impl/SubsetValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/SubsetValueTest.java) — **Reconcile**: `testRandomSubsetE7F1`, `testRandomSubsetE7F05`, `testRandomSubsetE6F1`, `testRandomSubsetE5F01`, `testRandomSubsetE5F025`, `testRandomSubsetE5F05`, `testRandomSubsetE5F075`, `testRandomSubsetE5F1`, `testRandomSubsetE32F1ENeg6`, `testRandomSubsetE17F1ENeg3`, `testRandomSubsetSubset16`, `testRandomSubsetSubset256`, `testRandomSubsetSubset65536`, `testRandomSubsetSubsetNoOverflow`, `testEmptyEnumerationsAreIndependent`, `testKSubsetEnumerator`, `testKSubsetEnumeratorNegative`, `testKSubsetEnumeratorGTCapacity`, `testNumKSubset`, `testNumKSubset2`, `testNumKSubsetNeg`, `testNumKSubsetKGTN`, `testNumKSubsetUpTo62`, `testNumKSubsetPreventsOverflow`, `testUnrankKSubsets`, `testUnrank16viaRank`, `testRandomSetOfSubsets`, `testRandomSetOfSubsets300`, `testRandomSetOfSubsets400`, `testElementsNormalizedIsNormalized`, `testKElementsAreNormalized`, `testKElementsMatchElementsNormalized`, `testRandomSubsetGeneratorK0`, `testRandomSubsetGeneratorKNegative`, `testRandomSubsetGeneratorKNplus1`, `testRandomSubsetGeneratorN10`, `testRandomSubsetGeneratorN100`.
  Related Go checks: [tlc/value_setoffcns_test.go](value_setoffcns_test.go), [tlc/value_ksubset_test.go](value_ksubset_test.go).
- [x] [tlc2/value/impl/TupleValueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/value/impl/TupleValueTest.java) — **Port complete**: `testErrorMessages`.
  Go translation: [tlc/value_tuple_java_test.go](value_tuple_java_test.go).
  Preserves all four source catch/message assertions, including Java's absence of fail() after the try blocks; added the missing production array-argument overload and source-aware assertion failures.

### Collections, buffered files, combinatorics, and statistics

Original utility cases, file operations/fuzz sequences, integer queues/stacks, vectors, contexts, combinatorics, and parameterized statistics.

- [ ] [tlc2/util/BufferedRandomAccessFileFuzzTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/BufferedRandomAccessFileFuzzTest.java) — **Missing**: `fuzz`, `testWellDefined`.
- [ ] [tlc2/util/BufferedRandomAccessFileTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/BufferedRandomAccessFileTest.java) — **Reconcile**: `testWrite`, `testWriteSeek`, `testWriteSeekNoLength`, `testRead`, `testReadSeekNoLength`, `testInvalidateBufferedData`, `testReadAfterSeekPastEndOfFile`, `testWriteAfterSeekPastEndOfFile`, `testObscureSetLengthBehavior`, `testIdempotentClose`, `testIOExceptionOnUseAfterClose`, `regressionTest01`, `regressionTest02`, `regressionTest03`, `regressionTest04`, `regressionTest05`, `regressionTest06`, `regressionTest07`.
  Related Go checks: [tlc/buffered_random_access_file_test.go](buffered_random_access_file_test.go).
- [ ] [tlc2/util/ByteUtilsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/ByteUtilsTest.java) — **Reconcile**: `test1`, `test2`, `test3`, `test4`, `test5`, `test6`.
  Related Go checks: [tlc/byte_utils_test.go](byte_utils_test.go).
- [ ] [tlc2/util/CombinatoricsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/CombinatoricsTest.java) — **Reconcile**: `testChoose`, `testChooseBigChoose`, `testSlowChooseBigChoose`, `testBigChoose50c1`, `testBigChoose50c10`, `testBigChoose50c20`, `testBigChoose50c30`, `testBigChoose400c1`, `testBigChoose400c50`, `testBigChoose400c100`, `testBigChoose400c200`.
  Related Go checks: [tlc/combinatorics_test.go](combinatorics_test.go).
- [x] [tlc2/util/ContextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/ContextTest.java) — **Port complete**: `testLookupEmpty`, `testLookupBranch`, `testLookupSymbolNodeNull`, `testLookup`, `testLookupCutOffFalse`, `testLookupCutOffTrue`, `testLookupWithAtBranching`, `testLookupWithCutOffFalseAtBranching`, `testLookupWithCutOffTrueAtBranching`, `testLookupSymbolNode`.
  Go translation: [tlc/context_java_test.go](context_java_test.go).
- [ ] [tlc2/util/GrowingLongVecTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/GrowingLongVecTest.java) — **Reconcile**: `testReadBeyondCapacity` (from `LongVecTest`), `testAddAndReadBeyondCapacity` (from `LongVecTest`), `testRemoveBeyondCapacity` (from `LongVecTest`), `testAddRemoveBeyondCapacity` (from `LongVecTest`), `testRemoveAndGet` (from `LongVecTest`), `testRemoveWrongOrder` (from `LongVecTest`), `testGetNegative` (from `LongVecTest`), `testRemoveNegative` (from `LongVecTest`), `testGrowAndShrink`.
  Related Go checks: [tlc/long_vec_test.go](long_vec_test.go).
- [ ] [tlc2/util/LongVecTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/LongVecTest.java) — **Reconcile**: `testReadBeyondCapacity`, `testAddAndReadBeyondCapacity`, `testRemoveBeyondCapacity`, `testAddRemoveBeyondCapacity`, `testRemoveAndGet`, `testRemoveWrongOrder`, `testGetNegative`, `testRemoveNegative`.
  Related Go checks: [tlc/long_vec_test.go](long_vec_test.go).
- [ ] [tlc2/util/MemIntQueueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/MemIntQueueTest.java) — **Reconcile**: `testDequeuePastLastElement`, `testEnqueueZeros`, `testEnqueueLong`, `testEnqueueDequeueLong`, `testGrow`.
  Related Go checks: [tlc/int_queue_test.go](int_queue_test.go).
- [ ] [tlc2/util/statistics/BucketStatisticsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/statistics/BucketStatisticsTest.java) — **Reconcile**; parameterized: preserve all original rows: `testInvalidArgument`, `testMean`, `testMedian`, `testMin`, `testMin2`, `testMax`, `testStandardDeviation`, `testGetPercentile`, `testGetPercentileNaN`, `testToString`.
  Related Go checks: [tlc/bucket_statistics_test.go](bucket_statistics_test.go).
- [ ] [tlc2/util/statistics/FixedSizedBucketStatisticsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/util/statistics/FixedSizedBucketStatisticsTest.java) — **Reconcile**; parameterized: preserve all original rows: `testMin`, `testMin2`, `testMax`, `testInvalidArgument`, `testGetPercentileNaN`, `testMaximum`.
  Related Go checks: [tlc/bucket_statistics_test.go](bucket_statistics_test.go).

### Numbered legacy model suite

Every concrete ETest*, Test*, and TestInvalidInvariant in tool/suite/. Inherited SuiteTestCase.testSpec and subclass coverage/assertion hooks remain part of each original test.

- [ ] [tlc2/tool/suite/ETest1.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest1.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest10.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest10.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest11.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest11.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest12.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest12.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest13.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest13.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest14.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest14.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest15.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest15.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest16.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest16.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest2.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest2.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest3.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest3.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest4.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest4.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest5.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest5.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest6.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest6.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest7.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest7.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest8.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest8.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/ETest9.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/ETest9.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/Test1.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test1.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test10.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test10.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test11.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test11.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test12.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test12.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test13.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test13.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test14.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test14.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test15.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test15.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test16.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test16.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test17.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test17.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test18.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test18.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test19.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test19.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test2.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test2.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test20.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test20.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test201.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test201.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test202.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test202.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test203.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test203.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test204.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test204.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test205.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test205.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test206.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test206.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test207.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test207.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test208.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test208.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test209.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test209.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test21.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test21.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test210.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test210.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/Test212.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test212.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/Test213.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test213.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/Test214.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test214.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/Test215.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test215.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/Test216.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test216.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/Test217.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test217.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/Test219.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test219.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test22.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test22.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test220.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test220.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/Test23.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test23.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test24.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test24.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test25.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test25.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test26.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test26.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test27.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test27.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test28.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test28.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test29.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test29.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test3.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test3.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test30.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test30.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test31.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test31.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test32.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test32.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test33.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test33.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test34.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test34.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test35.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test35.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test36.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test36.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test37.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test37.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test38.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test38.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test39.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test39.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test4.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test4.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test40.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test40.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test41.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test41.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test42.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test42.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test43.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test43.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test44.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test44.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test45.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test45.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test46.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test46.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test47.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test47.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test48.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test48.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test49.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test49.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test5.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test5.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test50.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test50.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test51.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test51.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test52.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test52.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test53.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test53.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test54.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test54.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test55.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test55.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test56.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test56.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test57.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test57.java) — **Missing**: `testSpec`.
- [ ] [tlc2/tool/suite/Test58.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test58.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test59.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test59.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test6.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test6.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test60.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test60.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test62.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test62.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test63.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test63.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test63a.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test63a.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test64.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test64.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test64a.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test64a.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test65.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test65.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test65a.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test65a.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test7.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test7.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test8.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test8.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test9.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test9.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test99.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test99.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/Test999.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/Test999.java) — **Missing**: `testSpec` (from `SuiteTestCase`).
- [ ] [tlc2/tool/suite/TestInvalidInvariant.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/suite/TestInvalidInvariant.java) — **Missing**: `testSpec`.

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

7 classes / 56 logical methods; 11 confirmed translations.

- [ ] [util/BufferedDataInputStreamTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/BufferedDataInputStreamTest.java) — `testReadStringThrowsOnShortStreamAtHalfBoundary`, `testReadStringThrowsOnShortStreamBelowHalfBoundary`, `testReadStringExactLength`, `testReadStringAcrossBufferRefill`, `testConstructorRejectsStreamReturningZero`, `testEmptyStream`, `testWriteReadStringRoundTrip` — pending original-method translation/reconciliation.
- [ ] [util/ExecutionStatisticsCollectorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/ExecutionStatisticsCollectorTest.java) — `testCompanyLevelNoFile`, `testCompanyLevelUnreadable`, `testCompanyLevelEmptyFile`, `testCompanyLevelNoESCFile`, `testCompanyLevelRandomIdFile`, `testCompanyLevelUserDefinedIdFile`, `testNoFile`, `testUnreadableFile`, `testEmptyFile`, `testNoESCFile`, `testRandomIdFile`, `testUserDefinedIdFile` — pending original-method translation/reconciliation.
- [x] [util/FileUtilTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/FileUtilTest.java) — `testReturnFromCheckpoint`, `testDuplicateStateDirCreation`, `testUseDifferentMetaDir` — mapped in [tlc/file_util_test.go](file_util_test.go).
- [ ] [util/MonolithSpecExtractorTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/MonolithSpecExtractorTest.java) — `testExtractConfig`, `testExtractModule`, `testConfigWithWindowsPathAsName`, `testModuleWithWindowsPathAsName`, `testGetConfig` — pending original-method translation/reconciliation.
- [x] [util/SimpleFilenameToStreamTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/SimpleFilenameToStreamTest.java) — `testResolveStandardModules`, `testResolveCommunityModules`, `testResolveByAbsolutePath`, `testResolveFromUserDir`, `testResolveWithCustomLibraryPath`, `testTLALibrarySystemProperty`, `testWindowsTLAFileCreation`, `testBizarreWorkingDirectorySearchBehavior` — mapped in [filename_to_stream_test.go](../filename_to_stream_test.go).
- [ ] [util/StringHelperTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/StringHelperTest.java) — `testGetWords`, `testGetWordsLeadingSpace`, `testGetWordsNull`, `testCopyString0`, `testCopyStringPos1`, `testCopyStringPos5`, `testCopyStringNeg1`, `testCopyStringNeg5`, `testOnlySpaces`, `testOnlySpacesNull`, `testTrimFront`, `testTrimFrontNull`, `testTrimFrontWhitespaces`, `testTrimEnd`, `testTrimEndNull`, `testTrimEndWhitespaces`, `testLeadingSpace`, `testLeadingSpacesNull`, `testIsIdentifier`, `testIsIdentifierNull` — pending original-method translation/reconciliation.
- [ ] [util/TLCRuntimeTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/util/TLCRuntimeTest.java) — `testIsThroughputOptimized` — pending original-method translation/reconciliation.

### Long-running fingerprint and queue tests

Four concrete classes / 22 inherited-or-local method contexts; no complete suite translations identified. Upstream test-dist-long excludes DiskFPSetTest, MSBDiskFPSetTest, and DiskStateQueueTest for runtime; OffHeapDiskFPSetLongTest remains selected. FPSetTest is an abstract test body used by the three fingerprint subclasses.

- [ ] [tlc2/tool/fp/DiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-long/tlc2/tool/fp/DiskFPSetTest.java) — `testSimpleFill` (from `FPSetTest`), `testMaxFPSetSizeRnd` (from `FPSetTest`), `testMaxFPSetSize` (from `FPSetTest`) — pending original-method translation/reconciliation.
- [ ] [tlc2/tool/fp/MSBDiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-long/tlc2/tool/fp/MSBDiskFPSetTest.java) — `testSimpleFill` (from `FPSetTest`), `testMaxFPSetSizeRnd` (from `FPSetTest`), `testMaxFPSetSize` (from `FPSetTest`) — pending original-method translation/reconciliation.
- [ ] [tlc2/tool/fp/OffHeapDiskFPSetLongTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-long/tlc2/tool/fp/OffHeapDiskFPSetLongTest.java) — `testSimpleFill` (from `FPSetTest`), `testMaxFPSetSizeRnd` (from `FPSetTest`), `testMaxFPSetSize` (from `FPSetTest`), `testCollisionBucket`, `testPosition`, `testMultipleFlushes` — pending original-method translation/reconciliation.
- [ ] [tlc2/tool/queue/DiskStateQueueTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-long/tlc2/tool/queue/DiskStateQueueTest.java) — `testEnqueue` (from `StateQueueTest`), `testsDequeueEmpty` (from `StateQueueTest`), `testDequeueEmpty` (from `StateQueueTest`), `testsDequeueNotEmpty` (from `StateQueueTest`), `testDequeueNotEmpty` (from `StateQueueTest`), `testEnqueueAddNotSame` (from `StateQueueTest`), `testEnqueueAddSame` (from `StateQueueTest`), `testsDequeueAbuseEmpty` (from `StateQueueTest`), `testsDequeueAbuseNonEmpty` (from `StateQueueTest`), `testGrowBeyondIntMaxValue` — pending original-method translation/reconciliation.

### Concurrent fingerprint stress tests

Four concrete classes / 17 inherited-or-local method contexts; no complete original translations identified. Preserve generator implementations, worker coordination, partitioning, random seeds, and sizes rather than replacing them with a small -race smoke test.

- [ ] [tlc2/tool/fp/ConcurrentWriteTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-concurrent/tlc2/tool/fp/ConcurrentWriteTest.java) — `test`, `test1`, `test2`, `test3`, `test4` — pending original-method translation/reconciliation.
- [ ] [tlc2/tool/fp/MultiThreadedLSBDiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-concurrent/tlc2/tool/fp/MultiThreadedLSBDiskFPSetTest.java) — `testMaxFPSetSizeRndBatched` (from `MultiThreadedFPSetTest`), `testMaxFPSetSizeRndBlock` (from `MultiThreadedFPSetTest`), `testMaxFPSetSizeRnd` (from `MultiThreadedFPSetTest`), `testMaxFPSetSizePartitioned` (from `MultiThreadedFPSetTest`) — pending original-method translation/reconciliation.
- [ ] [tlc2/tool/fp/MultiThreadedMSBDiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-concurrent/tlc2/tool/fp/MultiThreadedMSBDiskFPSetTest.java) — `testMaxFPSetSizeRndBatched` (from `MultiThreadedFPSetTest`), `testMaxFPSetSizeRndBlock` (from `MultiThreadedFPSetTest`), `testMaxFPSetSizeRnd` (from `MultiThreadedFPSetTest`), `testMaxFPSetSizePartitioned` (from `MultiThreadedFPSetTest`) — pending original-method translation/reconciliation.
- [ ] [tlc2/tool/fp/MultiThreadedOffHeapDiskFPSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test-concurrent/tlc2/tool/fp/MultiThreadedOffHeapDiskFPSetTest.java) — `testMaxFPSetSizeRndBatched` (from `MultiThreadedFPSetTest`), `testMaxFPSetSizeRndBlock` (from `MultiThreadedFPSetTest`), `testMaxFPSetSizeRnd` (from `MultiThreadedFPSetTest`), `testMaxFPSetSizePartitioned` (from `MultiThreadedFPSetTest`) — pending original-method translation/reconciliation.

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
| [tlc2/tool/DumpLoadTraceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/DumpLoadTraceTest.java) | `testLivenessMCDumpLoadTraceJSON`, `testLivenessMCDumpLoadTraceTLC`, `testLivenessMCDumpLoadTraceJSONAutoWorkers`, `testLivenessMCDumpLoadTraceTLCAutoWorkers`, `testSafetyDumpLoadTraceJSON`, `testSafetyDumpLoadTraceTLC`, `testSafetyDumpLoadTraceJSONAutoWorkers`, `testSafetyDumpLoadTraceTLCAutoWorkers`, `testLivenessBidirectionalDumpLoadTraceJSON`, `testLivenessBidirectionalDumpLoadTraceTLC`, `testLivenessBidirectionalDumpLoadTraceJSONAutoWorkers`, `testLivenessBidirectionalDumpLoadTraceTLCAutoWorkers`, `testSafetyTESpecEqAliasDumpLoadTraceJSON`, `testSafetyTESpecEqAliasDumpLoadTraceTLC`, `testSafetyTESpecEqAliasDumpLoadTraceJSONAutoWorkers`, `testSafetyTESpecEqAliasDumpLoadTraceTLCAutoWorkers`, `testLivenessExample1DumpLoadTraceJSON`, `testLivenessExample1DumpLoadTraceTLC`, `testLivenessExample1DumpLoadTraceJSONAutoWorkers`, `testLivenessExample1DumpLoadTraceTLCAutoWorkers`, `testSafetyDieHardAliasSubDumpLoadTraceJSON`, `testSafetyDieHardAliasSubDumpLoadTraceTLC`, `testSafetyDieHardAliasSub2DumpLoadTraceJSON`, `testSafetyDieHardAliasSub2DumpLoadTraceTLC`, `testSafetyDieHardAliasSub2DumpLoadTraceJSONAutoWorkers`, `testSafetyDieHardAliasSub2DumpLoadTraceTLCAutoWorkers`, `testSafetyDieHardAliasSupDumpLoadTraceJSON`, `testSafetyDieHardAliasSupDumpLoadTraceTLC`, `testSafetyDieHardAliasSupDumpLoadTraceJSONAutoWorkers`, `testSafetyDieHardAliasSupDumpLoadTraceTLCAutoWorkers` | [tlc_dump_load_trace_java_test.go](../tlc_dump_load_trace_java_test.go) |
| [tlc2/tool/EmptySubsetEqTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EmptySubsetEqTest.java) | `testSpec` | [tlc_model_java_test.go](../tlc_model_java_test.go) |
| [tlc2/tool/EvalExceptionLivenessTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EvalExceptionLivenessTest.java) | `testSpec` | [tlc_eval_exception_liveness_java_test.go](../tlc_eval_exception_liveness_java_test.go) |
| [tlc2/tool/EvalExceptionTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/EvalExceptionTest.java) | `testSpec` | [tlc_eval_exception_java_test.go](../tlc_eval_exception_java_test.go) |
| [tlc2/tool/Github1109Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1109Test.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/Github1109aTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github1109aTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/Github361Test.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/Github361Test.java) | `testSpec` | [tlc_constant_processing_java_test.go](../tlc_constant_processing_java_test.go) |
| [tlc2/tool/IncompleteNextMultipleActionsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/IncompleteNextMultipleActionsTest.java) | `testSpec` | [tlc_next_model_java_test.go](../tlc_next_model_java_test.go) |
| [tlc2/tool/IncompleteNextTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/IncompleteNextTest.java) | `testSpec` | [tlc_next_model_java_test.go](../tlc_next_model_java_test.go) |
| [tlc2/tool/PossibleCountsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PossibleCountsTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/PostConditionsFailTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PostConditionsFailTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/PostConditionsTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PostConditionsTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/PrintTraceRaceTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/PrintTraceRaceTest.java) | `testSpec` | [tlc_print_trace_race_java_test.go](../tlc_print_trace_race_java_test.go) |
| [tlc2/tool/SubsetEqTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/SubsetEqTest.java) | `testSpec` | [tlc_model_java_test.go](../tlc_model_java_test.go) |
| [tlc2/tool/TLCExtTraceSimTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCExtTraceSimTest.java) | `testSpec` | [tlc_ext_trace_java_test.go](../tlc_ext_trace_java_test.go) |
| [tlc2/tool/TLCGetAllTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCGetAllTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/TLCGetLevelTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCGetLevelTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/TLCGetNamedUndefinedTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCGetNamedUndefinedTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/TLCSetInitTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCSetInitTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/TLCSetMultiSimTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCSetMultiSimTest.java) | `testSpec` (from `TLCSetSimTest`) | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/TLCSetSimTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCSetSimTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/TLCSetTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TLCSetTest.java) | `testSpec` | [tlc_checker_model_java_test.go](../tlc_checker_model_java_test.go) |
| [tlc2/tool/TraceWithLargeSetOfInitialStatesTest.java](../../tlaplus/tlatools/org.lamport.tlatools/test/tlc2/tool/TraceWithLargeSetOfInitialStatesTest.java) | `testSpec` | [tlc_trace_model_java_test.go](../tlc_trace_model_java_test.go) |
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
