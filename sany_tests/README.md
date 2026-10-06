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
