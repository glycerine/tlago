# Java SANY Test Port

This directory contains the Go porting surface for the Java SANY tests from:

`tlaplus/tlatools/org.lamport.tlatools/test/tla2sany`

Fixtures and corpus inputs are mirrored under `sany_tests/test_vectors/tla2sany/`.
Do not use a `testdata/` directory here; local cleanup/fuzzer workflows may
delete it.

Each Java `@Test` method should have a corresponding Go test. Most translations
live in this directory. The direct `semantic.TestContext` translation lives in
root `sany_context_java_test.go` to access the private Context implementation and
check the original merge result, error code and structured parameters.
Keep translated assertions faithful to the Java source, and add any copied
fixtures under `sany_tests/test_vectors/` instead of relying on external checkout
locations.

`semantic.SemanticErrorCorpusTests.test` now retains all four original assertion
families: failure severity, structured argument counts, no suspected-unreachable
checks, and the expected error code. Its helper preserves the original interrupted
semantic-generation catch and returns parsing messages before semantic checking.
Passing those assertions does not establish every diagnostic's full fidelity;
remaining source differences are recorded in `tlc/HANDOFF.md` and
`tlc/PORT_PROGRESS.md`. SANY methods do not change the TLC test-port totals.

`semantic.TestInstanceNode.testOperatorArgumentMinimumLevelDiagnostic` now
retains the complete original method: syntax/dependency loading succeeds,
semantic generation succeeds, level checking fails, exactly one 4246 error is
present, and its one-based position/required-level parameters are `[1, 3]`.
The production phase entry points are `ParseSanySpecSource`, `GenerateSanySpec`
and `CheckSanySpecLevels`. This does not credit the whole Java frontend helper
API or complete canonical semantic graphs; those remain separately pending.

`semantic.TestLevelChecking.testAll` is **port complete** across all 51 original
parameter rows. Exact expressions, Unicode synonyms and expected results match
the source matrix. The method retains syntax/dependency loading, both semantic
log success assertions, level-log/result agreement, and the expected level
result. Unchanged Java and the full Go SANY package pass. This method does not
establish full frontend-helper API or canonical semantic-graph parity.

`semantic.NestedModuleInstanceTest.testTopLevelInstanceOfNestedModule` is
**port complete** with the original parse-success and semantic-success
assertions and byte-identical fixture. `testLetInstanceOfNestedModule` retains
the original Java `@Ignore` status and exact reason; it is not active-method
completion credit. Unchanged Java JUnit and the full Go SANY package agree.

`semantic.IncrementalSemanticParseTests.bigRadixNumeralTest` is **port complete**
in root `sany_incremental_semantic_java_test.go`, where it can inspect the actual
semantic-generation result. It preserves all three original radix inputs,
semantic-log success, concrete NumeralNode type, `useVal() == false`, and the
expected big-integer value. Generation retains the numeral's syntax and shares
the node with the TLC bridge instead of postponing construction until evaluation.
The unchanged original Java class passes all five methods; this credits only
the complete Go methods listed here, not the whole class.

`semantic.IncrementalSemanticParseTests.basicExpressionTest` is also **port
complete** in the root translation. It retains incremental input `0`, both
semantic-log success assertions, non-null result, actual numeral level checking,
original syntax-node identity, ConstantLevel and concrete NumeralNode type.
Numeral level checking records the supplied signed 32-bit iteration and the
no-iteration overload advances it, matching source even for decreasing iterations
and signed wraparound. The normal composite checker invokes it at iteration 1.
This does not complete the common LevelNode getter guards, other metadata, or
recursive graph iteration propagation.

The other three methods in `semantic.IncrementalSemanticParseTests` remain
**reconcile**. Their native-AST checks omit original canonical semantic-node
assertions: ordinary OpDef recursion/arity/body metadata, syntax-node identity,
actual level-check results and constant levels, dependency module tables,
concrete LetInNode/OpApplNode types, and imported operator source identity.
Implement the production graph and incremental generator behavior before
crediting these methods; parsing the same snippets is insufficient.
