# SANY Java-to-Go test fidelity audit

Audit completed: 2026-10-08 (source review began 2026-10-07). Status: findings documented; fixes not implemented by this audit.

## Scope and result

This audit covers every `@Test` method under upstream `tlatools/org.lamport.tlatools/test/tla2sany`: **34 test classes and 96 methods**. Parameterized methods are counted once in that inventory; their complete parameter sets and shared assertion helpers were also reviewed. This is SANY's dedicated test tree, not every TLC test that happens to invoke SANY, nor every possible SANY behavior.

Java baseline: `../tlaplus`, commit `8f4bc8b73ad1202774a6bf70143436f8ba50aab0`. Go baseline: commit `6c9172247e3763c658e82ff14df10dcda160c8af` plus the working-tree test sources inspected during this audit. The main development thread was active; unrelated TLC edits were left untouched.

All 96 Java methods have Go counterparts. **59 methods have a recorded assertion, setup, helper or exercised-entry-point divergence**, grouped into 25 findings. That includes smaller omissions and overly strict translations as well as serious false-positive paths; it does not mean 57 production bugs. **35 methods have no identified divergence in their reviewed test contract**, one preserves an upstream ignore, and one preserves an upstream empty test body. Finding counts overlap: one method can need several repairs.

The concrete defects below are established by comparing test source and helper behavior. No production fixes or test changes were made, and no test workload was run for this audit. Where a restored assertion would expose missing production behavior, the implementation requirement is stated explicitly. “No identified divergence” is a source-review result, not certification that the SANY implementation is bug-free or that every helper has been formally proved equivalent.

## Repair status

Repairs are now in progress under the user’s instruction to fix this audit and
then fix the production failures found by the restored tests. The finding text
below records the original audit baseline; this status and the method ledger
track the current repairs.

- **F01–F03: port complete for all nine affected original methods.** The four
  Errors methods and five formatting methods now live in root-package tests,
  exercise production logging/rendering, and preserve generation-only semantics.
  Targeted tests pass, as do the remaining dedicated SANY package and selected
  original root SANY methods. No original parameter or assertion was weakened.
- **F04–F06: port complete for the eleven affected original methods.** All
  twelve WarningControl methods and IllegalOperator pass. The three settings
  methods use the production settings-controlled driver and record WARNING/ERROR
  output. CLI methods retain source arguments, combined streams, exit mapping and
  exact substrings; IllegalOperator uses its extensionless path. The driver now
  renders SANY summaries and ErrorDetails and propagates warning elevation to its
  named exit status. This covers these contracts, not every SanySettings option.
- **F07–F09: port complete for all three affected original methods.** Github429
  initializes the real frontend and runs parsing plus generation without levels
  or linting, preserving its nonthrowing contract. Location checks the source
  nested coordinate comparisons independently of Compare. Vector checks the
  actual ArrayIndexOutOfBoundsException, and production now throws that type.
  All original inputs and bounds remain unchanged; source selections pass.
- **F20–F21 and F25: port complete for all four affected original methods.**
  Precedence uses fixed wrapper heir navigation and the original GeneralId
  predicate over the full source matrix. Builtin initialization calls the actual
  context membership operation before retrieval on both passes. The two CLI
  diagnostic methods retain shared capture and source substrings without added
  exit assertions. The dedicated SANY suite and related root methods pass.
- **F10–F19 and F22–F24: pending.** A green run of their current counterparts does not resolve
  their recorded test-contract defects. Final verification of the fully restored
  96-method inventory has not yet occurred.

## Evidence and counting rules

- The Go copies of all **227 files** in `sany_tests/test_vectors` were compared with their corresponding upstream `test/tla2sany` or `test-model` originals: all match byte for byte; none of those Go files lacks an upstream counterpart. No upstream syntax, semantic or semantic-error corpus fixture is absent from its mirrored corpus directory.
- The syntax corpus contains **42 files / 355 cases**; the semantic corpus has **28 .tla files**, including the upstream-excluded `NegativeOpTest.tla`; the semantic error corpus has **121 matching cases**. The tokenizer has nine rows, belchDEF five, proofs four, and level checking 51.
- The independently written operator fixture table has **98 rows**, matching Java's fixity, ordered synonyms, precedence bounds and associativity exactly. No operator row was missing in the current tree.
- The auxiliary `test_vectors/java-sany/xml/sany.xsd` is **not** one of those 227 mirrored test files and does **not** match the pinned upstream schema. Its substantive differences are F23.
- Root-package replacements are the primary translations where present. Older surrogate tests in `sany_tests` marked supplementary are not mistaken for the primary tests. In particular, the former approximations for incremental parsing, semantic corpus traversal, selectors, built-in initialization, context merge and parser-error output have current replacements.
- Java test infrastructure without `@Test` methods was reviewed where used: `SANYTest`, `RecordedSanyOutput`, test frontend/resolvers, `SyntaxCorpusRunner`, `SyntaxCorpusFileParser`, `AstNode`, and `TlaPlusParserOutputTranslator`. These are helpers, not additional test methods.
- Go subtests, `t.TempDir`, explicit output writers, pointer boxes for Java reference identity and a slice for the vector enumeration snapshot can preserve the source behavior without emulating a JVM. Do not mistake such representation changes for weakened assertions.

## Fix rules

Restore the original operations, inputs, stage boundaries, assertions and exception categories. Use a root-package test when private production APIs must be inspected. Expected results must be derived independently of the actual result, following Java's helper, not assigned directly to the value under test.

If the faithfully restored test fails, repair the actual Go port before considering the test complete. Do not add post-hoc deduplication, allow arbitrary failures, omit flags, suppress a failing parameter, replace exact checks with broad substrings, or silently skip unavailable schema validation. Preserve actual upstream ignores and commented-out bodies. Do not invent new regression cases in this repair pass.

Run the restored original tests normally with their original bounds. Long corpus or operator workloads must not be combined with `-race`. Store any required fixture under `test_vectors`, never a directory named `testdata`. The original audit itself authorized no source changes or git actions; the user’s later repair request now authorizes this implementation work.

## Detailed findings

### F01 — `TestErrors`: constructed diagnostic slices substitute for the production logger

**Affected:** all four methods in [semantic_testerrors_test.go](sany_errors_java_test.go). Java: `semantic/TestErrors.java`.

Java constructs `Errors`, calls `addMessage`, and checks the logger's returned strings, details, severity, counts and summary. Go constructs already-classified `Diagnostic` values and puts them directly into a slice. Deep equality against those same values tests slice filtering, not the original logging operation.

Specific lost or changed behavior:

1. Java error messages use `INTERNAL_ERROR` (4003); Go supplies `E1300`, a different code. All error-bearing methods therefore miss original code classification and retained error details.
2. Java compares complete warning/error string arrays: `location.toString() + "\n\n" + message`. Go compares struct slices and merely searches summaries for message text. Wrong locations, module names, separators and rendering can pass.
3. Java checks complete `getMessages`, `getWarningDetails` and `getErrorDetails` results, including empty opposite-severity lists. Go does not exercise these original detail-access paths. The mixed test also omits explicit total-message and error counts; the duplicate test omits error count, complete detail lists and summary checks.
4. Java sends a **null location** through `addMessage` and verifies canonical `Location.nullLoc` in both details and strings. Go assigns `Position{}` directly, avoiding normalization.
5. Java's location helper uses a changing seed starting at zero, a distinct `Test%d.tla` name, and coordinates `seed*3`, `seed*5`, `seed*7`, `seed*11`. Go fixes the filename to `Test.tla` and starts at positive explicit seeds, losing the initial zero-coordinate case and distinct-name input.
6. The duplicate test calls `addMessage` three times per message and expects automatic insertion-time deduplication. Go explicitly calls `.Deduplicated()` on a manually assembled slice, concealing a logger that stores or exposes duplicates.

**Fix:** port/use the production equivalent of `Errors.addMessage` and its unformatted overload. `appendSanyDiagnostics`, `sanyJavaErrorDetails` and `sanyErrorsString` are relevant existing production pieces, but direct use of slice construction plus a test-local cleanup is not an adequate replacement. Restore the logger-facing API if required. Call it with Java's codes, text and generated locations, including an actual absent-location input. Reproduce all string arrays, detail lists, success/failure and count assertions. Build expected details independently. Use `E4003` for `INTERNAL_ERROR`. Preserve the generated-location formula and zero seed without introducing reliance on Go test execution order. In the duplicate case, make six logger calls and inspect the logger immediately; no test-side deduplication.

### F02 — percent-format logger tests do not call the formatter they claim to test

**Affected:** the first two methods of [semantic_testerrormessageformatting_test.go](sany_error_message_formatting_java_test.go).

The literal-text test assigns `Message: message` then checks `log[0].Message == message`. The parameter test calls `fmt.Sprintf` **inside the test**, assigns its result, and checks it. Neither invokes the SANY error logger or its detail renderer. A broken production formatter has no effect on either test. Both also invent a string-valued code name instead of using the real `SUSPECTED_UNREACHABLE_CHECK` metadata (4004).

**Fix:** invoke the actual production logger with code 4004 and null location. First supply the literal `Couldn't resolve infix operator symbol `%'.` with no arguments; then supply the format `Couldn't resolve infix operator symbol `%s'.` with the argument `%%`. Read/render the first production error detail and compare to the exact original expected string. Preserve the distinction between no arguments and one argument; never preformat in the test. Implement the production overload/rendering distinction first if absent.

### F03 — semantic percent-error reproducers can pass on the wrong phase or severity

**Affected:** `testUnresolvedPercentOperator`, `testUnresolvedNonfixPercentOperator`, `testUnresolvedDoublePercentOperatorIsNotRenamed` in the same file.

Java requires successful syntax/dependency processing, runs semantic generation **without level checking**, and renders only `log.getErrors()`. Parse/abort exceptions are not accepted as successful observations. Go's helper returns parse diagnostics as soon as parsing fails, otherwise runs all of `CheckSpec`, then searches `Diagnostics.Error()` across the entire result. A parse error or warning mentioning the source `%`/`%%` can satisfy the check instead of the intended rendered semantic error.

**Fix:** port Java's helper phase for phase: parse the exact module, require syntax success, resolve dependencies, run `GenerateSanySpec`, and let unexpected semantic aborts fail. Do not run levels/linting. Filter/render actual error details only, preferably through the original SANY error-string operation, and preserve the original substring assertions. Do not strengthen those substrings into new assertions not present upstream.

### F04 — warning settings tests bypass the configured SANY driver and output recorder

**Affected:** `WarningControlTest`'s three API-level methods in [drivers_warningcontroltest_test.go](sany_tests/drivers_warningcontroltest_test.go).

Java passes `SanySettings` to `SANY.parse`, captures `RecordedSanyOutput` at WARNING threshold, checks the driver's enumerated exit, and inspects emitted log levels. Go obtains diagnostics with `LoadSanySpec`/`CheckSpec`, then applies `DiagnosticOptions` afterward. It does not check configuration flowing through the driver, emission suppression, emitted WARNING/ERROR levels or the driver's result. The default case permits either `Foo` or `W4802`, whereas Java requires `Foo` in a WARNING message. The elevation case omits Java's explicit strict-error-code setting.

**Fix:** expose/use the production settings-enabled parse driver and output interface. Port a recording output helper that retains message level/text and obeys the original threshold. Use the original settings values and resolver directory. Assert OK plus a WARNING containing `Foo`; OK plus no emitted WARNING when suppressed; semantic/level failure plus an ERROR containing `Warning treated as error` when elevated. Suppression/elevation must happen in the driver at the correct stages, not in a test-side filter.

### F05 — warning CLI tests change argument and stream contracts

**Affected:** seven CLI methods other than the two missing-argument methods.

Java points `ToolIO.out` and `ToolIO.err` at the same capture. Go passes nil stdout, which `RunCLI` turns into `io.Discard`, and inspects stderr alone. The two negative suppression assertions can therefore pass even if the warning appears on stdout. Positive cases impose a stderr-only requirement that Java does not impose.

The suppression, elevation and comma-separated suppression methods additionally omit Java's **`-error-codes`** argument. Thus they do not test warning control in the original strict-code configuration. The overlap test weakens the required substring from `codes were set to both -suppressMessages and -messagesAsErrors` to `both -suppressMessages and -messagesAsErrors`.

**Fix:** use a shared capture for both writers, matching the Java setup, for all seven output-checking methods. Restore `-error-codes` in the three original invocations. If the Go SANY option parser/driver does not accept and implement it, port that functionality first. Restore the full overlap substring. Retain original exact named success/error categories through a documented SANY-to-Go exit mapping; never use arbitrary nonzero as a substitute. The two missing-argument tests only assert the ERROR category upstream and need no invented output assertion.

### F06 — `IllegalOperatorTest` loses the CLI and exact rendered diagnostic contract

**Affected:** [drivers_illegaloperatortest_test.go](sany_tests/drivers_illegaloperatortest_test.go).

Java calls `SANYmain` with an **extensionless** `IllegalOperatorTest` path and inspects combined output for four exact substrings. Go loads an explicit `.tla` path, runs `CheckSpec`, manually deduplicates errors, then checks coordinates and three shortened message fragments. It loses extensionless resolution, the `*** Errors: 1\n` summary, the rendered module name/newline, the complete `Argument number 1 to operator 'D' \n` text and final punctuation. Manual deduplication can hide duplicate reporting.

**Fix:** invoke the actual Go SANY CLI with the original extensionless filename and a shared output capture. Assert all four source strings verbatim. Do not deduplicate in the test. Correct resolver behavior, production deduplication and SANY rendering if those assertions fail. Coordinate checks can remain supplementary, but cannot replace the original output assertions.

### F07 — `Github429Test` changes semantic-analysis mode

**Affected:** [drivers_github429test_test.go](sany_tests/drivers_github429test_test.go).

Java initializes the frontend, parses, then calls semantic analysis with **level checking disabled** (`false`). Its contract is that these operations do not throw an exception; it does not require an empty/error-free log. Go runs `CheckSpec`, including level checking, and requires no error diagnostics. This is a different and stricter workload, rather than the original nonthrowing phase test.

**Fix:** retain the original fixture and frontend initialization, then invoke parse followed by generation-only semantic analysis. Unexpected exceptions/panics must fail naturally; do not add levels/linting or substitute diagnostic success for the original exception contract. Keep any useful stronger completed-spec test separately labelled supplementary.

### F08 — comparator correctness is checked using the comparator itself

**Affected:** `LocationTest.testComparator` in [st_locationtest_test.go](sany_tests/st_locationtest_test.go).

Both versions parse 16 locations, sort them and require 14 unique values. Java then independently checks ascending begin line, begin column, end line and strictly increasing end column when preceding fields tie. Go sorts with `Compare` and then checks adjacent `Compare < 0`. A consistently reversed or incorrectly prioritized comparator can satisfy its own check.

**Fix:** keep the original 16 inputs and 14-unique assertion. Replace the final comparator call with Java's exact nested comparisons on `Line`, `Column`, `EndLine`, `EndColumn`. The expected ordering must not call `Compare`. Fix production comparison if the independent oracle fails.

### F09 — vector insertion accepts any panic, not the required bounds exception

**Affected:** `VectorTest.insertElementAtRejectsAppendPosition` in [utilities_vectortest_test.go](sany_tests/utilities_vectortest_test.go).

Java catches only `ArrayIndexOutOfBoundsException` when inserting index 2 into a two-element vector. Go treats any nonnil panic as success, including unrelated failures. Production `SanyVector.InsertElementAt` currently panics with a string, so the mismatch is also visible in source.

**Fix:** preserve the original invalid insertion and assert the specific bounds-failure type. A matching exception type already exists as `tlc.ArrayIndexOutOfBoundsException`; use the production typed equivalent consistently. Nil recovery must fail, and any other recovered value must fail/repanic. Fix the production bounds exception rather than allowing arbitrary panic text. No need to add new boundary inputs.

### F10 — every XML module test drops upstream XSD validation

**Affected:** all 20 methods in [xml_testxmlexportermodule_test.go](sany_tests/xml_testxmlexportermodule_test.go).

Java's direct methods or shared `export` helper parse the complete document and validate it against bundled `sany.xsd`. All but the offline test explicitly require schema presence; offline still validates if the resource is available. Go only parses XML and makes selected tree/text assertions. None of the 20 tests validates the schema.

This particularly defeats `testUseHideDefsExportsModuleReference`: the upstream regression was a **schema** choice missing `ModuleNodeRef`, while the exporter already emitted the correct element. Go's reference/count checks cannot detect that regression. Likewise, well-formed XML can still violate element order, required children, occurrence bounds, primitive types or reference-kind choices.

**Fix:** provide real validation against the **current pinned upstream schema**, then restore the helper-level validation call and required resource check. Use a Go-capable validator that enforces the schema's actual constraints; a few hand-picked tag checks, XML tokenization or golden-text comparison are not equivalent. Package the schema in a reproducible way under `test_vectors`/embedded resources. Preserve the offline test's conditional resource semantics; do not extend that conditional skip to the other 19 tests. The production non-offline exporter also needs its Java validation path; the tests must exercise it and still retain Java's independent validation of captured output.

### F11 — 15 XML module tests bypass the command entry point and include option

**Affected:** all module methods except offline, terse, restricted and the two uncomment-flag methods.

Java runs `XMLExporter.run` and examines captured XML/output. The shared `export` helper also passes `-I BASE_PATH`. Go's `checkedXMLExporterModule` calls `checkedSANYXMLForPath`, which directly loads/checks a spec and calls `SanyXML`. Driver argument parsing, exit status, stdout serialization, required quiet stderr and explicit include-directory behavior are not tested. Finding F10 remains even in the five methods that do use `runSANYXMLCommand`.

**Fix:** mechanically port each original invocation. For the shared helper, call the Go XML driver with `-I` and the original model directory and model path. Capture both writers, assert the source success/quiet-stderr requirements, parse stdout and perform F10 validation. Direct early methods must preserve their own source assertions rather than being assigned a universal stronger contract. Keep library-only export checks as supplementary if useful.

### F12 — DieHard optional attributes and unnamed modules can silently escape assertions

**Affected:** `testExportDieHardModule`.

Java tests `hasAttribute("filename")` and, when present, requires its value to end in `DieHard.tla`. Go checks only when the value is nonempty, accepting a present empty attribute. Java requires the first descendant `uniquename` of every module to belong to the allowed set; Go skips the membership check when `xmlUniqueName` is empty, accepting empty/missing names.

**Fix:** distinguish attribute absence from a present empty value with the map lookup boolean. For every exported module, retrieve the source's expected name node, fail if missing, and require membership even for an empty value. Keep the original allowed module set unchanged.

### F13 — recursive-section tests conflate a missing element with empty text

**Affected:** `testRecursiveSectionGroupsJointDeclaration` and `xmlRecursiveSection`.

Java returns null when no `recursiveSection` exists and empty string when one exists but has empty text. It asserts presence for f/h, equal f/g, distinct h, and **absence** for nonRecursive. Go returns `""` for both cases; its nonRecursive check therefore accepts `<recursiveSection/>`. It also requires all recursive values to be nonempty, stronger than Java's presence checks.

**Fix:** return `(value, present)` or a pointer from the helper. Port Java's null/presence checks separately from text equality. Require no element for nonRecursive, not merely empty text. Do not add a nonempty-value assertion unless kept as a separately identified supplementary check.

### F14 — uncommented comment checking delays an assertion Java makes immediately

**Affected:** `testUncommentFlagWithTLACommentStyles` and shared `requireXMLPreComments`.

Java immediately fails when a recognized operator's encountered comment lacks the required content, then removes that operator from its expected map. Go merely leaves an entry in the map on a mismatch and continues. A later duplicate comment with the correct content can rescue a wrong earlier comment. Java's `testRelationPreComments` and `testUncommentFlagWithRelations` do use the deferred-match behavior; their loops must not be changed to the stricter style by accident.

**Fix:** faithfully port this method's own assertion/removal loop, or give the helper an explicit immediate-assert policy. On the first still-expected named comment, require the substring before removing it. Preserve each other method's original loop semantics. This method also does not require every pre-comment to be nonempty: its expected map covers styles 1, 2 and 4, while style 3 is outside that map. The Go helper currently adds a blanket nonempty assertion. Remove that added requirement from the mechanical translation or keep it separately as supplementary. Keep all existing expected multiline text and indentation.

### F15 — XML error tests omit the library contract and collapse result types

**Affected:** all seven methods in [xml_testxmlexportererrors_test.go](sany_tests/xml_testxmlexportererrors_test.go).

Each Java test exercises both `moduleToXML` and `run`. Help must not throw; invalid cases must throw an `XMLExportingException` carrying the exact `XMLExporterExitCode`, and the CLI must return that same named category. Go's first five methods test only `RunCLI`; its last two use a load/check/export diagnostic helper and omit the CLI. No translated method tests the original structured library result.

Java's relevant categories are OK=0, ARGS_PARSING_FAILURE=1, SPEC_PARSING_FAILURE=2 and XML_UNREPRESENTABLE_CHARACTER=7. The generic Go CLI currently distinguishes semantic failures as 4 and has no corresponding tested exporter-specific typed error. Semantic spec failures are not Java XML_CONFIGURATION_FAILURE=4 and must not be interpreted as such merely because integers coincide.

**Fix:** port an exporter-specific structured error/result type and library entry point following Java's control flow; preserve code and message (and bug classification where asserted). Restore both invocations in each test. Test the named exporter categories directly, and define an explicit adapter if the top-level CLI intentionally uses a wider exit enum. For faithful Java command parity, preserve the XML exporter's original numeric exit codes; do not remap a spec semantic failure into a different XML category. The four currently exact CLI checks should remain exact, not become merely nonzero.

### F16 — XML spec-failure test changes its fixture and permits every failure

**Affected:** `TestXMLExporterErrors.testSpecParseFailure`.

Java chooses `SemanticErrorCorpusTests.getTestFiles().get(0)` from the sorted corpus, currently **`E4200_Instance_Test.tla`**, and requires SPEC_PARSING_FAILURE for both entry points. Go hardcodes `E4200_Test.tla` and accepts any CLI code other than OK. Argument failure, exporter failure or internal failure would all satisfy it.

**Fix:** reuse the faithful sorted corpus enumeration and select its first entry, matching upstream. Restore the library assertion and exact exporter-specific SPEC_PARSING_FAILURE result. A changed fixture or an arbitrary error is not the source test.

### F17 — XML control-character reproducers lose exact rejection classification and CLI checks

**Affected:** the two null-character methods and `TestXMLExporterStringEscapes.testFormFeedStringEscapeIsRejected`.

Java asserts XML_UNREPRESENTABLE_CHARACTER through both library and CLI paths; the two null tests also assert `isBug()==false`. Go accepts `HasErrors()` plus `U+0000` or `U+000C` anywhere in accumulated diagnostics. Its helper can return syntax/semantic failure before XML export, so a wrong-stage error mentioning the character can pass. The CLI and null-case nonbug assertions are absent.

**Fix:** parse/generate successfully as the original exporter does, call the production library exporter, inspect its typed error code 7 and the original exact character substring, and run the CLI against the same fixture expecting code 7. Restore `isBug()==false` in the two Java null tests. Retain the original generated module bodies and form-feed fixture unchanged. Do not accept any prior-phase error as proof of XML character rejection.

### F18 — XML help tests replace complete shared usage output with keywords

**Affected:** both methods in [xml_testxmlexporterhelptext_test.go](sany_tests/xml_testxmlexporterhelptext_test.go).

Java independently obtains the full output of `XMLExporter.printUsage` and requires the selected CLI stream to contain it. Go checks only `sany-xml` and `FILE` for help, and accepts either `sany-xml` **or** `file` for no arguments. A one-line error about a file passes without printing any usage. The current `runSanyXML` no-argument branch only prints `at least one file is required`; this is a concrete missing production behavior visible in source, not a test execution reported by this audit.

**Fix:** place the tests in the root package if needed to call `printSanyXMLUsage`. Capture that printer separately, then require the entire result in stdout for help and stderr for no arguments, preserving the source empty-opposite-stream and named exit assertions. Go command branding can differ in the printer; the complete printer output must still be present. Inspect top-level `RunCLI` help interception too, so it uses the same usage contract rather than a differently formatted help implementation.

### F19 — supported string-escape export can pass despite malformed trailing XML

**Affected:** `TestXMLExporterStringEscapes.testSupportedStringEscapes` and `soleStringValueFromXML`.

Java invokes the XML CLI, requires OK, parses the **complete** document, requires exactly one StringValue, and compares its decoded contents exactly. Go bypasses the CLI via `checkedSANYXMLForPath`. Its token loop stops on **every** error instead of distinguishing EOF from malformed XML. Once it has read the expected StringValue, a malformed tail can be ignored and the test can pass.

**Fix:** restore the actual command invocation and original success category. Parse stdout completely; treat only `io.EOF` as successful completion, fail on every other decoding error, and require a single properly closed document root. Retain exactly-one-StringValue and the unchanged decoded value `"\\ \n \r \t \""`. Preserve qualified-name semantics as in F22. Do not add XSD validation as a new independent assertion to this method; restore its original production non-offline export path instead.

### F20 — precedence helpers permit malformed AST structure Java would reject

**Affected:** `OperatorPrecedenceTests.testOperatorCombination` in [parser_operatorprecedencetests_test.go](sany_tests/parser_operatorprecedencetests_test.go).

Java `getOpImage` reads the precise operator-wrapper child and its heir `[1]`. Go's `sanyOperatorTokenImage` recursively returns the first token image anywhere underneath the wrapper. An incorrectly shaped wrapper retaining the expected token elsewhere can pass.

Java chooses an infix node's higher-precedence child by testing whether its left child is **N_GeneralId**. Go tests whether that child is one of three operator-expression kinds, otherwise choosing the right child. Those predicates agree on the normal current fixture trees but are not equivalent for a malformed/unexpected left-child kind.

**Fix:** port the fixed Java heir navigation directly and its N_GeneralId predicate. Do not search recursively to accommodate an incorrectly ported AST. Preserve the complete table, matrix and #893 skip; fix production structure if the restored original traversal fails.

### F21 — built-in initialization omits the separate context membership assertion

**Affected:** `TestBuiltInOperatorInitialization.testInitAndReInit` in [sany_builtin_initialization_java_test.go](sany_builtin_initialization_java_test.go).

Java calls `context.occurSymbol(expected.Name)` before retrieval. That operation uses `table.containsKey`, independently of `getSymbol`. Go checks only that retrieval returns a nonnil node. All the substantial property and reinitialization assertions are now present; this is a narrower remaining API-coverage omission, not evidence that built-in initialization is generally absent or wrong.

**Fix:** retain/add the production context membership operation if missing and assert it for every expected built-in before retrieval, both before and after reinitialization. Do not implement the assertion as a test-local repeated `getSymbol != nil` check. Preserve existing actual-node property checks and complete enumeration.

### F22 — XML inspection discards namespace-qualified names

**Affected:** all 20 module methods through `parseSANYXMLTestDocument`, and the supported-string-escape method through `soleStringValueFromXML`.

The Java DOM is namespace-aware and its `getNodeName`/`getElementsByTagName` operations retain qualified names; schema validation further constrains permissible namespaces. Go stores and searches only `Name.Local`, including attribute keys. Thus a prefixed or differently namespaced element/attribute can masquerade as the expected unqualified name. F10 already prevents many namespace mistakes, but fixing validation alone does not faithfully restore the helper.

**Fix:** preserve QName/namespace information in the inspection tree and port Java's lookup semantics for these unprefixed expected names. Require the expected no-namespace schema names when validating; do not erase `Name.Space`. Handle attribute presence using the proper qualified key. Compare complete qualified end/start names and preserve complete-document validation. An XML parser accepting a document as well-formed is not proof that its names match the source assertion.

### F23 — available auxiliary XML schema is from an older contract

**Affected:** the resource needed to repair the 20 module tests and the decimal library path. File: [sany.xsd](test_vectors/java-sany/xml/sany.xsd). Upstream: `src/tla2sany/xml/sany.xsd`.

The existing auxiliary schema differs substantively from the pinned Java version, not just in commentary or formatting. Examples confirmed by parsing both schemas:

- Its DecimalNode lacks the current node group and `integralPart`/`fractionalPart` children.
- Its UserDefinedOpKind lacks current original-operator/module information, pre-comments, recursiveSection and local fields.
- Its ModuleNode lacks `extends` and the nested `ModuleNodeRef` choice.
- Its UseOrHideNode `defs` choice lacks `ModuleNodeRef`, precisely the regression covered by the upstream HIDE test.

This file is not currently used to validate the audited Go tests, so this finding is a required fixture repair when restoring F10/F24, not a claim that present tests actively validate against the wrong schema.

**Fix:** use a byte-for-byte copy of the pinned upstream schema for the current parity suite and package it for both production/test resource lookup. Do not patch only individual schema choices to make selected tests pass. If an older schema must remain for separately identified historical compatibility work, retain it under a clearly versioned name and do not use it for these source tests.

### F24 — decimal export takes a different option/validation path

**Affected:** [xml_testdecimalxmlexport_test.go](sany_tests/xml_testdecimalxmlexport_test.go).

The two final substring assertions match Java. However, Java calls `specToXMLString(emt, false, false, false, false)`: unrestricted, comments unchanged, **no pretty printing**, **not offline**. The original library path validates the document against the embedded schema before returning. Go calls default `SanyXML`, avoiding the original explicit formatting options and schema-validation contract.

**Fix:** invoke the faithful library equivalent with those exact four options and propagate its structured failure. Implement non-offline production validation, using F23's current schema. Keep both exact integral/fractional substring assertions; do not replace them with integer-normalized text that loses leading or trailing zeros. Java does not add a separate test-side validator here, so restore library behavior rather than inventing extra original assertions.

### F25 — two original diagnostic CLI tests gain an extra failure-exit assertion

**Affected:** `Github723Test.test` and `RecursiveDefDeclMismatchTest.test`.

Both Java tests call the void `SANYmain` entry point and assert exact diagnostic substrings in shared output. Their Go translations retain those substrings but additionally fail when `RunCLI` returns OK. This is an added success criterion, not a loss of the original diagnostic coverage. It can nevertheless turn a faithful source test into a different regression test when CLI exit policy changes.

**Fix:** keep the original fixture, invocation and all diagnostic substring assertions in the mechanical translation. Do not require a nonzero result there unless the original entry-point adapter itself has to signal an unexpected exception. Move any deliberate CLI exit-policy assertion into a clearly labelled supplementary test rather than treating it as an upstream requirement. This finding is low priority compared with the tests that can pass without exercising production behavior.

## Implementation order and completion criteria

1. Restore logger APIs and the original diagnostic tests (F01-F03), then warning-driver settings/recording and CLI arguments (F04-F06). These give reliable observations for subsequent SANY work.
2. Correct the small phase/helper/type omissions (F07-F09, F20-F21, F25) without replacing original parameter sets.
3. Bring in the current schema and faithful non-offline XML validation/error APIs (F10, F15-F17, F23-F24). Restore command capture, include paths and complete help output (F11, F18-F19).
4. Restore individual XML predicates, immediate assertions and qualified-name handling (F12-F14, F22).

Each affected method is complete only when its original setup, operation, expected result and all assertions are present and the faithfully ported method passes. Shared-helper fixes must be checked against every listed dependent method; success in one dependent test is insufficient. Where a supplementary test deliberately adds stricter checks, keep those separate from the mechanical source translation. Mark actual repairs elsewhere only when done; this document records outstanding work and does not declare it implemented.

## Method-by-method review ledger

`Fix` means at least one confirmed translation divergence below. `Reviewed` means no divergence identified in the inspected source test contract. `Ignored upstream` and `Empty upstream` do not count as executed functional coverage. Each method links to its primary Go implementation; the Java class link provides the original methods and helpers.

### Github723Test

Java: [Github723Test.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/Github723Test.java).

Reviewed: Original fixture and complete arity diagnostic substring retained. Both output streams captured; Go additionally requires a failing CLI exit.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`test`](sany_tests/github723test_test.go:13) | port complete | F25 |

### RecursiveDefDeclMismatchTest

Java: [RecursiveDefDeclMismatchTest.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/RecursiveDefDeclMismatchTest.java).

Reviewed: Both original complete diagnostic substrings and fixture retained. Both output streams captured; Go adds a nonzero CLI exit assertion.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`test`](sany_tests/recursivedefdeclmismatchtest_test.go:13) | port complete | F25 |

### Bug156TEStackOverflowTest

Java: [drivers/Bug156TEStackOverflowTest.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/drivers/Bug156TEStackOverflowTest.java).

Reviewed: Java test body is commented out. Go preserves the empty body, but does not reproduce the unused SpecObj construction/global initialization in @Before. Neither body exercises the historical bug.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testFrontEndParse`](sany_tests/drivers_bug156testackoverflowtest_test.go:7) | Empty upstream | — |

### Github429Test

Java: [drivers/Github429Test.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/drivers/Github429Test.java).

Repaired: Original fixture and initialized frontend retained; parsing and generation-only semantic analysis preserve the nonthrowing contract (F07).

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testForFailedParse`](sany_github429_java_test.go) | port complete | F07 |

### IllegalOperatorTest

Java: [drivers/IllegalOperatorTest.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/drivers/IllegalOperatorTest.java).

Repaired: Original extensionless CLI argument, shared output stream and all four exact rendered-output assertions restored (F06).

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`test`](sany_tests/drivers_illegaloperatortest_test.go:12) | port complete | F06 |

### WarningControlTest

Java: [drivers/WarningControlTest.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/drivers/WarningControlTest.java).

Reviewed: All 12 methods and original warning fixture retained. Settings/output, CLI argument and stream differences repaired (F04–F05). The two missing-argument cases preserve their tested failure category.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testWarningAppearsWithDefaultSettings`](sany_warning_control_java_test.go) | port complete | F04 |
| [`testSuppressMessagesViaSettings`](sany_warning_control_java_test.go) | port complete | F04 |
| [`testMessagesAsErrorsViaSettings`](sany_warning_control_java_test.go) | port complete | F04 |
| [`testCLISuppressMessagesSilencesWarning`](sany_tests/drivers_warningcontroltest_test.go:49) | port complete | F05 |
| [`testCLIMessagesAsErrorsCausesFailure`](sany_tests/drivers_warningcontroltest_test.go:60) | port complete | F05 |
| [`testCLIMultipleCodesSuppressed`](sany_tests/drivers_warningcontroltest_test.go:71) | port complete | F05 |
| [`testCLIUnknownCodeInSuppressMessages`](sany_tests/drivers_warningcontroltest_test.go:82) | port complete | F05 |
| [`testCLIUnknownCodeInMessagesAsErrors`](sany_tests/drivers_warningcontroltest_test.go:93) | port complete | F05 |
| [`testCLISuppressMessagesMissingArgument`](sany_tests/drivers_warningcontroltest_test.go:104) | Reviewed | — |
| [`testCLIMessagesAsErrorsMissingArgument`](sany_tests/drivers_warningcontroltest_test.go:112) | Reviewed | — |
| [`testCLIOverlapBetweenSuppressMessagesAndMessagesAsErrors`](sany_tests/drivers_warningcontroltest_test.go:120) | port complete | F05 |
| [`testCLIErrorLevelCodeInSuppressMessages`](sany_tests/drivers_warningcontroltest_test.go:131) | port complete | F05 |

### TestSanyOutputFormatting

Java: [output/TestSanyOutputFormatting.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/output/TestSanyOutputFormatting.java).

Reviewed: Actual SimpleSanyOutput invoked: verbatim % and %% with no parameters; interpolation with parameters; actual parser rejection and captured token "%" output. Platform line separator retained.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testPercentSignInMessageWithoutArguments`](sany_tests/output_testsanyoutputformatting_test.go:12) | Reviewed | — |
| [`testArgumentsAreStillInterpolated`](sany_tests/output_testsanyoutputformatting_test.go:23) | Reviewed | — |
| [`testParseErrorMentioningPercentOperator`](sany_tests/output_testsanyoutputformatting_test.go:33) | Reviewed | — |

### BelchDefTests

Java: [parser/BelchDefTests.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/BelchDefTests.java).

Reviewed: All five source rows retained. Actual lazy token stream and belchDEF used; initialization, current DEFBREAK, header/body/footer consumption checked. Go unconsumed cursor represents the Java initial EOF dummy.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`runTestCase`](sany_belchdef_java_test.go:32) | Reviewed | — |

### CanonicalOperatorTests

Java: [parser/CanonicalOperatorTests.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/CanonicalOperatorTests.java).

Reviewed: All 98 independent operator fixture rows retained, including synonyms in source order; actual operator existence and canonical synonym resolution asserted.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testCanonicalOperatorCorrectness`](sany_tests/parser_canonicaloperatortests_test.go:11) | Reviewed | — |

### IncrementalSyntaxParseTests

Java: [parser/IncrementalSyntaxParseTests.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/IncrementalSyntaxParseTests.java).

Reviewed: Standalone SPEC-state parsing and belchDEF retained. Original op == 0, numeral 0, and three-item conjunction inputs; actual node kinds, identifier/numeral images and list length asserted.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testParseBasicOpDef`](sany_incremental_syntax_java_test.go:36) | Reviewed | — |
| [`testParseBasicExpression`](sany_incremental_syntax_java_test.go:61) | Reviewed | — |
| [`testParseConjunctionList`](sany_incremental_syntax_java_test.go:73) | Reviewed | — |

### OperatorAssociativityTests

Java: [parser/OperatorAssociativityTests.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/OperatorAssociativityTests.java).

Reviewed: All infix rows and Cartesian products of their synonyms retained; original acceptance iff associative predicate retained. Uses syntax parsing, not semantic checking.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testOperatorAssociativity`](sany_tests/parser_operatorassociativitytests_test.go:7) | Reviewed | — |

### OperatorPrecedenceTests

Java: [parser/OperatorPrecedenceTests.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/OperatorPrecedenceTests.java).

Reviewed: All 98 rows, synonym pairs, eight constructible fixity combinations, conflict predicates and original #893 exclusion retained. Two AST navigation helpers loosen the structural checks (F20).

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testOperatorCombination`](sany_tests/parser_operatorprecedencetests_test.go:11) | port complete | F20 |

### ParseErrorTests

Java: [parser/ParseErrorTests.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/ParseErrorTests.java).

Reviewed: Original one-row matrix retained; generated filename matches module name; actual parsing-phase failure and recorded full multiline error substring checked.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testAll`](sany_parse_error_java_test.go:35) | Reviewed | — |

### ProofTests

Java: [parser/ProofTests.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/ProofTests.java).

Reviewed: All four input/expected proof trees retained; node kinds, levels, step names, child counts, optional PROOF token and recursive nested proof comparison retained. The asymmetric nested-proof check also exists in Java and is not a new Go omission.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`test`](sany_tests/parser_prooftests_test.go:12) | Reviewed | — |

### TlaPlusSyntaxCorpusTests

Java: [parser/TlaPlusSyntaxCorpusTests.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/TlaPlusSyntaxCorpusTests.java).

Reviewed: All 42 corpus files and 355 cases retained; SKIP/ERROR attributes, all 12 named known failures, checked translator error versus assertion failure distinction, recursive AST kinds/children/field names, and unused TLA+ kind check retained. Reviewed runner, DSL parser and recursive translator; 115 numeric dispatch labels and 111 translated assertions match source counts. Counts support inspection, not a proof of helper equivalence.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testAll`](sany_tests/parser_tlaplussyntaxcorpustests_test.go:14) | Reviewed | — |
| [`testAllTlaPlusNodesUsed`](sany_tests/parser_tlaplussyntaxcorpustests_test.go:82) | Reviewed | — |

### TokenizerTests

Java: [parser/TokenizerTests.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/TokenizerTests.java).

Reviewed: All nine source cases retained: four module cases, three pragma cases, EOL comment and block comment. Initial, explicitly selected and final lexical states plus complete token-kind sequences including EOF retained.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`runTokenizerCase`](sany_tests/parser_tokenizertests_test.go:12) | Reviewed | — |

### IncrementalSemanticParseTests

Java: [semantic/IncrementalSemanticParseTests.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/IncrementalSemanticParseTests.java).

Reviewed: Five primary root tests exercise actual canonical nodes. Original standalone parsing, semantic success, explicit level checking where present, syntax identity, name/arity/recursive flag, numeral type and big-radix fallback retained. LET tests check dependency list, transitive modules and original imported operator source identity.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`basicOpDefTest`](sany_incremental_semantic_java_test.go:119) | Reviewed | — |
| [`basicExpressionTest`](sany_incremental_semantic_java_test.go:81) | Reviewed | — |
| [`bigRadixNumeralTest`](sany_incremental_semantic_java_test.go:35) | Reviewed | — |
| [`letInExpressionTest`](sany_incremental_semantic_java_test.go:162) | Reviewed | — |
| [`letInExpressionWithTransitiveDepsTest`](sany_incremental_semantic_java_test.go:165) | Reviewed | — |

### NestedModuleInstanceTest

Java: [semantic/NestedModuleInstanceTest.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/NestedModuleInstanceTest.java).

Reviewed: Original two fixtures retained. Active test checks parse and semantic success. Second test preserves the exact Java @Ignore reason; it is not counted as executed coverage. No original output-content assertion was lost.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testTopLevelInstanceOfNestedModule`](sany_tests/semantic_nestedmoduleinstancetest_test.go:32) | Reviewed | — |
| [`testLetInstanceOfNestedModule`](sany_tests/semantic_nestedmoduleinstancetest_test.go:37) | Ignored upstream | — |

### SemanticCorpusTests

Java: [semantic/SemanticCorpusTests.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/SemanticCorpusTests.java).

Reviewed: All 28 original .tla files retained; NegativeOpTest keeps upstream issue #1130 exclusion. Actual graph walk, definition/source identity, first attached comment ID and four level-symbol identities checked; semantic warnings rejected. Java has a TODO for syntax-level assertion discovery, which is not an implemented assertion to port.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`test`](sany_semantic_corpus_java_test.go:36) | Reviewed | — |

### SemanticErrorCorpusTests

Java: [semantic/SemanticErrorCorpusTests.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/SemanticErrorCorpusTests.java).

Reviewed: All 121 matching source cases retained and sorted. Expected filename code/severity, every fixed parameter count, absence of 4004, and presence of expected code checked. Parse messages short-circuit semantics; WrongInvocationException allowance preserved.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`test`](sany_tests/semantic_semanticerrorcorpustests_test.go:15) | Reviewed | — |

### TestBuiltInOperatorInitialization

Java: [semantic/TestBuiltInOperatorInitialization.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestBuiltInOperatorInitialization.java).

Reviewed: Before/after actual global-context reinitialization, all original symbol properties, null body, variadic-null versus zero-arity nonnull parameter arrays, max-level/weight arrays and complete symbol enumeration checked. One separate membership operation is omitted (F21).

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testInitAndReInit`](sany_builtin_initialization_java_test.go:10) | port complete | F21 |

### TestContext

Java: [semantic/TestContext.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestContext.java).

Reviewed: Actual context merge used, not a fixture proxy; false merge result, exactly one error, E4224 and incoming/existing declaration/definition parameter positions retained.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testDifferentSymbolClassesDiagnostic`](sany_context_java_test.go:33) | Reviewed | — |

### TestErrorMessageFormatting

Java: [semantic/TestErrorMessageFormatting.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestErrorMessageFormatting.java).

Reviewed: All five original methods present; production logger bypass in first two and semantic error phase/rendering boundary differences in last three (F02/F03).

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testPercentSignInMessageTextIsNotAFormatSpecifier`](sany_error_message_formatting_java_test.go) | Port complete | F02 |
| [`testPercentSignInMessageParameterIsNotAFormatSpecifier`](sany_error_message_formatting_java_test.go) | Port complete | F02 |
| [`testUnresolvedPercentOperator`](sany_error_message_formatting_java_test.go) | Port complete | F03 |
| [`testUnresolvedNonfixPercentOperator`](sany_error_message_formatting_java_test.go) | Port complete | F03 |
| [`testUnresolvedDoublePercentOperatorIsNotRenamed`](sany_error_message_formatting_java_test.go) | Port complete | F03 |

### TestErrors

Java: [semantic/TestErrors.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestErrors.java).

Reviewed: Original message texts retained, but logger insertion, original error code, generated locations, rendered strings, details and insertion-time deduplication are not faithfully exercised (F01).

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testWarningMessages`](sany_errors_java_test.go) | Port complete | F01 |
| [`testErrorMessages`](sany_errors_java_test.go) | Port complete | F01 |
| [`testMixedMessageLevels`](sany_errors_java_test.go) | Port complete | F01 |
| [`testDuplicateErrorsIgnored`](sany_errors_java_test.go) | Port complete | F01 |

### TestInstanceNode

Java: [semantic/TestInstanceNode.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestInstanceNode.java).

Reviewed: Original nested module body retained; separate semantic success and level failure, exactly one error, E4246 and one-based position/required-level parameters [1,3] retained.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testOperatorArgumentMinimumLevelDiagnostic`](sany_tests/semantic_testinstancenode_test.go:33) | Reviewed | — |

### TestLevelChecking

Java: [semantic/TestLevelChecking.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestLevelChecking.java).

Reviewed: All 51 expression/result rows retained. Separate syntax, semantic and level stages; log success versus returned level result and expected boolean retained. The redundant semantic-log recheck after levels is also in Java.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testAll`](sany_tests/semantic_testlevelchecking_test.go:34) | Reviewed | — |

### TestSubexpressionSelectors

Java: [semantic/TestSubexpressionSelectors.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestSubexpressionSelectors.java).

Reviewed: All three inputs retained; generation without level checking, specific semantic abort recovery, user-facing rejection excluding E4003, first error E4200, complete rendered message and exact source location retained.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testUnresolvedCompoundOperatorName`](sany_subexpression_selectors_java_test.go:87) | Reviewed | — |
| [`testConsecutiveTreeNavigationSelectors`](sany_subexpression_selectors_java_test.go:92) | Reviewed | — |
| [`testAllTreeNavigationSelectors`](sany_subexpression_selectors_java_test.go:97) | Reviewed | — |

### LocationTest

Java: [st/LocationTest.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/st/LocationTest.java).

Reviewed: All includes examples and original 16 parsed comparator inputs retained; 14 unique locations retained. Independent nested coordinate oracle restored (F08).

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testContains`](sany_tests/st_locationtest_test.go:12) | Reviewed | — |
| [`testContains2`](sany_tests/st_locationtest_test.go:59) | Reviewed | — |
| [`testComparator`](sany_tests/st_locationtest_test.go:91) | port complete | F08 |

### VectorTest

Java: [utilities/VectorTest.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/utilities/VectorTest.java).

Reviewed: Pointer boxes preserve Java object identity in contains/append assertions; order, size and snapshot contents retained. Go slice snapshot is an idiomatic representation change. Insert now requires the original bounds exception class (F09).

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`containsUsesIdentityComparison`](sany_tests/utilities_vectortest_test.go:12) | Reviewed | — |
| [`insertElementAtRejectsAppendPosition`](sany_tests/utilities_vectortest_test.go:25) | port complete | F09 |
| [`elementsReturnsSnapshot`](sany_tests/utilities_vectortest_test.go:38) | Reviewed | — |
| [`appendNoRepeatsUsesIdentity`](sany_tests/utilities_vectortest_test.go:51) | Reviewed | — |

### TestDecimalXMLExport

Java: [xml/TestDecimalXMLExport.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestDecimalXMLExport.java).

Reviewed: Both exact integral/fractional XML substrings and Decimal.tla retained. Original non-offline library export and its validation/format option behavior not retained (F24).

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`test`](sany_tests/xml_testdecimalxmlexport_test.go:33) | Fix | F24, F23 |

### TestXMLExporterErrors

Java: [xml/TestXMLExporterErrors.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestXMLExporterErrors.java).

Reviewed: All seven methods present. Source requires library and CLI paths and exact named error categories, not just any error. Missing paths and distinctions detailed in F15-F17.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testHelpReturnsOk`](sany_tests/xml_testxmlexportererrors_test.go:15) | Fix | F15 |
| [`testNoArgs`](sany_tests/xml_testxmlexportererrors_test.go:23) | Fix | F15 |
| [`testIncludeDirWithoutSpec`](sany_tests/xml_testxmlexportererrors_test.go:31) | Fix | F15 |
| [`testCannotFindSpec`](sany_tests/xml_testxmlexportererrors_test.go:39) | Fix | F15 |
| [`testSpecParseFailure`](sany_tests/xml_testxmlexportererrors_test.go:47) | Fix | F15, F16 |
| [`testNullCharacterInStringLiteral`](sany_tests/xml_testxmlexportererrors_test.go:56) | Fix | F15, F17 |
| [`testNullCharacterInComment`](sany_tests/xml_testxmlexportererrors_test.go:60) | Fix | F15, F17 |

### TestXMLExporterHelpText

Java: [xml/TestXMLExporterHelpText.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestXMLExporterHelpText.java).

Reviewed: Exit and empty opposite-stream checks retained; complete shared usage text check replaced by loose keywords (F18).

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testPrintHelpText`](sany_tests/xml_testxmlexporterhelptext_test.go:13) | Fix | F18 |
| [`testPrintHelpTextOnNoArgs`](sany_tests/xml_testxmlexporterhelptext_test.go:27) | Fix | F18 |

### TestXMLExporterModule

Java: [xml/TestXMLExporterModule.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestXMLExporterModule.java).

Reviewed: All 20 methods and original model/comment fixtures retained. Reviewed UID resolution, nesting order/reachability, LET reference kinds/names/order, theorem/assumption references, recursive sections and HIDE defs. Shared schema, driver and namespace omissions affect every method; individual differences detailed below.

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testExportDieHardModule`](sany_tests/xml_testxmlexportermodule_test.go:17) | Fix | F10, F22, F23, F11, F12 |
| [`testExportCaseOtherModule`](sany_tests/xml_testxmlexportermodule_test.go:57) | Fix | F10, F22, F23, F11 |
| [`testExportWithOfflineMode`](sany_tests/xml_testxmlexportermodule_test.go:67) | Fix | F10, F22, F23 |
| [`testExportWithTerseMode`](sany_tests/xml_testxmlexportermodule_test.go:74) | Fix | F10, F22, F23 |
| [`testExportWithRestrictedMode`](sany_tests/xml_testxmlexportermodule_test.go:81) | Fix | F10, F22, F23 |
| [`testRelationPreComments`](sany_tests/xml_testxmlexportermodule_test.go:91) | Fix | F10, F22, F23, F11 |
| [`testTLACommentStylesPreComments`](sany_tests/xml_testxmlexportermodule_test.go:110) | Fix | F10, F22, F23, F11 |
| [`testNestedModuleIsChildOfEnclosingModule`](sany_tests/xml_testxmlexportermodule_test.go:150) | Fix | F10, F22, F23, F11 |
| [`testNestedModuleIsNotInheritedThroughExtends`](sany_tests/xml_testxmlexportermodule_test.go:162) | Fix | F10, F22, F23, F11 |
| [`testLetInstanceOfEmptyModuleExportsModuleInstanceRef`](sany_tests/xml_testxmlexportermodule_test.go:171) | Fix | F10, F22, F23, F11 |
| [`testLetInstanceWithNothingToInlineExportsModuleInstanceRef`](sany_tests/xml_testxmlexportermodule_test.go:183) | Fix | F10, F22, F23, F11 |
| [`testLetInstanceInlinesDefinitionsOfInstancee`](sany_tests/xml_testxmlexportermodule_test.go:195) | Fix | F10, F22, F23, F11 |
| [`testLetExportsEachModuleDefinitionIndependently`](sany_tests/xml_testxmlexportermodule_test.go:208) | Fix | F10, F22, F23, F11 |
| [`testLetAlwaysExportsModuleInstanceRef`](sany_tests/xml_testxmlexportermodule_test.go:213) | Fix | F10, F22, F23, F11 |
| [`testTopLevelInstanceOfEmptyModuleExportsInstanceNode`](sany_tests/xml_testxmlexportermodule_test.go:226) | Fix | F10, F22, F23, F11 |
| [`testUncommentFlagWithTLACommentStyles`](sany_tests/xml_testxmlexportermodule_test.go:240) | Fix | F10, F22, F23, F14 |
| [`testUncommentFlagWithRelations`](sany_tests/xml_testxmlexportermodule_test.go:271) | Fix | F10, F22, F23 |
| [`testLetInstanceExportsInstantiatedTheoremAndAssumption`](sany_tests/xml_testxmlexportermodule_test.go:290) | Fix | F10, F22, F23, F11 |
| [`testRecursiveSectionGroupsJointDeclaration`](sany_tests/xml_testxmlexportermodule_test.go:304) | Fix | F10, F22, F23, F11, F13 |
| [`testUseHideDefsExportsModuleReference`](sany_tests/xml_testxmlexportermodule_test.go:327) | Fix | F10, F22, F23, F11 |

### TestXMLExporterStringEscapes

Java: [xml/TestXMLExporterStringEscapes.java](../tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestXMLExporterStringEscapes.java).

Reviewed: Original supported decoded string and form-feed fixture retained; CLI/library categories, complete document parsing and qualified names are weakened (F17/F19/F22).

| Java method / primary Go translation | Result | Findings |
| --- | --- | --- |
| [`testSupportedStringEscapes`](sany_tests/xml_testxmlexporterstringescapes_test.go:13) | Fix | F19, F22 |
| [`testFormFeedStringEscapeIsRejected`](sany_tests/xml_testxmlexporterstringescapes_test.go:20) | Fix | F17 |

## Recheck procedure

For a later audit, enumerate the Java `@Test` methods again rather than relying on filename counts or port-complete labels. Resolve each counterpart to the current primary Go function and inspect its shared helpers, fixtures and phase/exit paths. Recompare mirrored fixtures and the schema against the pinned upstream revision. Reconcile this ledger if the main thread changes a test after this snapshot. Do not treat a green Go suite as fidelity evidence until the source assertions above have been restored.
