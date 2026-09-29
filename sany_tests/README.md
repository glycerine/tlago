# Java SANY Test Port

This directory contains the Go porting surface for the Java SANY tests from:

`tlaplus/tlatools/org.lamport.tlatools/test/tla2sany`

Fixtures and corpus inputs are mirrored under `sany_tests/test_vectors/tla2sany/`.
Do not use a `testdata/` directory here; local cleanup/fuzzer workflows may
delete it.

The first pass is intentionally compile-only. Each Java `@Test` method is
represented by a Go test with `t.Skip("tla2sany wip")` at the top. Porting work
should remove one skip at a time, translate that Java test's assertions
faithfully, and make it green against the Go SANY implementation.
