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
The primary `semantic.TestErrors` and `semantic.TestErrorMessageFormatting`
translations now live in root [sany_errors_java_test.go](../sany_errors_java_test.go)
and [sany_error_message_formatting_java_test.go](../sany_error_message_formatting_java_test.go).
They call production logging/rendering and generation-only semantics rather
than constructing expected diagnostic slices or formatting inside the tests.
The three `drivers.WarningControlTest` settings methods live in root
[sany_warning_control_java_test.go](../sany_warning_control_java_test.go), calling
the production settings-controlled driver and observing WARNING/ERROR output.
Its nine CLI methods and IllegalOperator retain combined streams, original
arguments and exact rendered diagnostics in this directory.
`drivers.Github429Test` lives in root
[sany_github429_java_test.go](../sany_github429_java_test.go), using the initialized
frontend, actual parsing and generation-only semantic driver with silent output.
Its source contract is nonthrowing phases, without added diagnostic-success checks.
The completed fidelity repairs and original method ledger are tracked in
[SANY_TESTS_TO_FIX.md](../SANY_TESTS_TO_FIX.md); its repair status identifies the verified scope.
Keep translated assertions faithful to the Java source, and add any copied
fixtures under `sany_tests/test_vectors/` instead of relying on external checkout
locations.

Both original `parser.TlaPlusSyntaxCorpusTests` methods are **port complete**,
retaining all 355 parameter contexts each. `testAll` uses the full source
SANY-to-DSL translator and recursive AST equality assertions. The runner retains
SKIP attributes, case-insensitive known-failure names, inverted acceptance for
known failures, and the distinction between checked translation errors and
structural assertion failures. No expected tree, source input or assertion is
weakened. `testAllTlaPlusNodesUsed` checks unused AST DSL kinds after loading every
expected tree, excludes PlusCal kinds and FAIR, and retains the zero-unused
assertion. The earlier native SANY-kind surrogate is removed.

The corpus loader retains source bytes, multiline names, ERROR/SKIP attributes
and expected DSL trees. All 356 external metadata/expected-tree rows and all 355
actual translated-output rows match Java. The original Java class passes all 710
test contexts. Helper comparisons also retain the 552 operator-literal and 470
identifier/kind/list/tuple observations. This establishes the original corpus
methods, not exhaustive parser/semantic graph parity. Main TLC totals are unchanged.

`parser.BelchDefTests.runTestCase` is **port complete** in root
`sany_belchdef_java_test.go`, retaining all five original parameter rows. It
checks the actual lazy parser stream: initialization leaves the cursor before
its first token, the first module token is not EOF, every expected definition
period ends at DEFBREAK, and consuming DEF invokes belchDEF. Preserve the original
space-count periods and header/footer consumption counts. Java's initial EOF
dummy maps to the Go parser's unconsumed cursor. The older full-module parse
checks remain supplementary, without duplicate original-method credit.

All three `parser.IncrementalSyntaxParseTests` methods are **port complete** in
root `sany_incremental_syntax_java_test.go`: `testParseBasicOpDef`,
`testParseBasicExpression`, and `testParseConjunctionList`. Parse the original
standalone input in SPEC state after belchDEF, using actual production grammar
entry points. Preserve nonnull results, concrete node kinds, exact token images
and the three-heir conjunction count. The earlier enclosing-module checks remain
supplementary with their assertions intact. Unchanged Java passes all eight rows
across these two classes; Go and complete SANY pass. This completes four original
methods and does not establish full parser lookahead or AST equality parity.

`semantic.SemanticCorpusTests.test` is **port complete** in root
`sany_semantic_corpus_java_test.go`. All 28 original parameter rows are retained,
including Semantics itself and the original NegativeOpTest assumption. Parse and
semantic success, no semantic warnings, actual operator-reference resolution,
source comments and checked levels match the original assertions. Reference
search uses the actual canonical walkGraph with postVisit callbacks, not AST
name matching or estimated levels. Original Java JUnit and Go pass; all 28
fixtures match the pinned Java source byte for byte. The older AST facade checks
in this directory remain supplementary and add no duplicate method credit.
This completes the original method, not general canonical allocation/AST equality
or evaluator sharing, and does not change the main TLC inventory totals.

All three `semantic.TestSubexpressionSelectors` methods are **port complete** in
root `sany_subexpression_selectors_java_test.go`. Preserve the original module
bodies, syntax/dependency phase, semantic-generation-only helper, abort catch,
failure assertion, empty internal-error list, and first error's exact code,
message and complete location. The earlier native API checks remain supplementary.
Original Java JUnit and Go pass. The all-navigation case exposed lost selector
syntax at a flattened call callee; generation now retains its actual source
selector for unresolved-name/location accumulation. This completes these three
methods, not the full selectorToNode engine or canonical malformed graph paths.

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
Literal getter guards and canonical mutable metadata are now integrated with
the same TLC body nodes. Canonical composite level algorithms are now translated; remaining generation
and evaluator integration do not establish whole-frontend completion.

`semantic.IncrementalSemanticParseTests.basicOpDefTest` is **port complete** in
the root translation. It parses `op == 0` incrementally and generates the actual
operator without an enclosing module. It preserves both log-success assertions,
non-null result, name, arity, recursion flag, actual level checking, syntax-node
identity, constant level and concrete NumeralNode body type.

`semantic.IncrementalSemanticParseTests.letInExpressionTest` and
`letInExpressionWithTransitiveDepsTest` are **port complete** in root
`sany_incremental_semantic_java_test.go`. They parse the original standalone
expressions, retain exactly one original parser dependency, load/check actual
canonical dependency modules, and generate against the external module table
with a dummy module. Preserve both log-success assertions, non-null result,
actual recursive level checking, syntax identity, constant level, concrete
LetInNode/OpApplNode/OpDefNode classes and the exact imported source reference.
The four embedded standard-module sources match the pinned Java files byte for
byte. The dependency helper uses production loading/generation and checks each
actual module before entering it in the external table. It does not wrap the
expression in a synthetic module or skip missing children.

All five original methods in this class are now translated. The unchanged Java
JUnit class passes all five; all five Go translations pass. Older AST facade
checks in this directory remain supplementary and add no duplicate method credit.
This completes this class, not the broader SANY frontend or TLC port.

`semantic.TestBuiltInOperatorInitialization.testInitAndReInit` is **port complete**
in root `sany_builtin_initialization_java_test.go`. Its complete helper checks
all 72 source property rows before and after actual global Context reinitialization:
symbol presence, OpDefNode class, name/kind, builtin identity, standard/parameter/
local flags, null body, arity, null variadic parameters or exact formal count,
level, both metadata arrays, and exhaustive context membership/count. Nil and
empty arrays remain distinct. The property table matches Java mechanically.
Builtins now use the common OpDef node rather than a separate symbol class;
ordinary operator graph construction remains pending. This whole SANY class
adds no TLC inventory credit.

`xml.TestDecimalXMLExport.test` is **port complete** with its byte-identical
original fixture and both original XML substring assertions. Unchanged Java and
Go pass. The production decimal representation now preserves source mantissa,
exponent, original image parts and overflow unscaled-value/scale fields. The
exporter reads those fields, including Java's positive scale in its overflow
branch; it does not reinterpret that branch as a negative decimal exponent.
Integer XML also uses the generated numeral representation, preserving decimal
leading-zero and TLA radix semantics instead of Go radix auto-detection. Bounded
scratch constructor/XML comparisons earn no additional original-test credit.

XML parity tests and non-offline XML export require `xmllint` (libxml2) on PATH.
The exact pinned Java `sany.xsd` is packaged with Go via embedding; validation
uses the complete libxml2 XSD engine with network access disabled. It checks
captured module-test output independently of production validation. A missing
validator fails explicitly rather than skipping or passing the check. The
original `-o` flag skips production schema validation; its source module test
still independently validates when the packaged schema is present.

The XML help methods now live in root
[sany_xml_help_java_test.go](../sany_xml_help_java_test.go), independently
capturing the complete production usage printer. The decimal method uses the
library exporter with its original four explicit false options, including
non-offline validation and no pretty printing. Error and string-escape methods
exercise both typed library failures and exact command exit categories.
