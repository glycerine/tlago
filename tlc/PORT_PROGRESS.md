# TLC Port Progress

## Notes To Future Us

- Do not re-run the old TLA+ XML or Apalache corpus sweeps unless explicitly asked. The active goal is the Go TLC model checker under `tlc/`.
- The source of truth is Java TLC under `../tlaplus/tlatools/org.lamport.tlatools/src/tlc2` and, later, its tests under `../tlaplus/tlatools/org.lamport.tlatools/test/tlc2`.
- The current experiment is breadth-first mechanical porting first. Do not start porting the Java TLC test suite yet. Existing fast Go tests may be run; small utility tests are acceptable.
- Keep the Go code mostly in package `tlc`, prefer concrete structs over interfaces, and use `InsMap` whenever deterministic iteration matters.
- Already audited recently; do not loop on these unless touched: `fpset.go` MemFPSet1/MemFPSet2, `liveness_tables.go`, `liveness_disk_graph.go`, `liveness_tableau_disk_graph.go`, `liveness_process.go`, `liveness_graph.go`, `random_generator.go`, `object_collections.go`, `int_stack.go`, `int_queue.go`, `state_pool.go`, `disk_state_queue.go`, and `simulation_worker.go`/`simulation_worker_modes.go`.
- Parser-backed liveness coverage already exists in top-level `tlc_liveness_parser_test.go`; it parses tiny TLA+ specs, builds a TLC tool, and drives `tlc.ParseLiveness`.
- Commit after each coherent chunk. Keep commit messages short and do not add authorship boilerplate.

## Accomplished

- Documented TLC architecture in `tlc/TLC_ARCH.md`, including major Java subsystems, data structures, algorithms, correctness risks, and performance-sensitive areas.
- Established Go TLC package structure with central concrete implementations rather than a Java-style package/interface hierarchy.
- Ported core runtime scaffolding:
  - TLC globals, error codes, output/reporting helpers, registry/system-property helpers.
  - Semantic/action nodes, contexts, call stack, evaluation control, standard definitions, model config parsing.
  - Tool/spec processor skeletons with init/next/action/invariant/property/liveness collections.
- Ported substantial value layer:
  - primitive, model, tuple, record, function, set, lazy, operator, interval, stream, and set-constructor values.
  - Java fingerprinting/order behavior where it is known to affect traces, fingerprints, and diagnostics.
  - random enumerable value behavior folded into `random_generator.go`.
- Ported utility collections and IO helpers:
  - Java-style `ObjLongTable`, `LongObjTable`, `SetOfLong`, `Vect`, `LongVec`, `List`, `BitVector`.
  - byte utilities, buffered/random access files, BigInt/external sort helpers, deprecated BigSet shape.
  - concrete object stacks/queues, integer stacks/queues, synchronous disk int stack.
  - disk/byte/state queue pool readers/writers and state pool cleaner.
- Ported state machinery:
  - mutable states, state vectors, set-of-states, state info, state printer/writer, trace records.
  - state fingerprints, normalization hooks, model value permutations, symmetry setup.
- Ported fingerprint sets:
  - base FPSet APIs, NoopFPSet, MemFPSet, MemFPSet1, MemFPSet2, MultiFPSet, DiskFPSet, OffHeapDiskFPSet scaffolding.
  - Recent commit `e985339` replaced placeholder MemFPSet1/2 aliases with Java-shaped implementations.
- Ported liveness front end and graph structures:
  - AST-to-live conversion, `ParseLiveness`, `ProcessLiveness`, PEM/order/tableau structures.
  - `TBGraph`, `TBGraphNode`, particles, promises, possible error models.
  - BEGraph, graph nodes/transitions, disk graph and tableau disk graph checkpoint/recovery/traversal basics.
  - Recent commit `ca4060d` fixed disk graph `AddNode` to honor recovered file pointers instead of always seeking EOF.
- Ported major checker/simulator scaffolding:
  - Model checker/DFID/check-impl/worker outlines and selected worker behavior.
  - Simulator and SimulationWorker with standard, debug-exploration, RL, and RL-action modes folded into one concrete worker plus mode flags.
  - Simulator now creates default per-worker live checks when Java would check liveness, using worker-specific graph directories under the simulator metadir.
  - Simulator now has a Java-style progress reporter for simulation progress, coverage/action-flow updates, and `_PERIODIC` false termination.
  - Simulator now prints Java-shaped worker-error behaviors and final simulation summaries (`TLC_STATS_SIMU`) with coverage/action-flow finalization.
  - BFS model checking now follows Java's recover-before-fresh-start ordering and preserves the `TLC_LIVE_FORMULA_TAUTOLOGY` guard.
  - BFS model checking now emits Java's final safety progress snapshot immediately before final liveness checking.
  - Liveness check/worker/error-trace skeletons with concrete disk graph fields.
- Added and kept green many fast Go tests for utility behavior and already-ported pieces.

## Left To Do

- Finish breadth-first audit of Java TLC source against Go implementation, patching true behavioral gaps as they are found.
- Complete faithful model-checker behavior:
  - `ModelChecker`, `Worker`, `DFIDModelChecker`, `DFIDWorker`, checkpoint/recovery, termination, progress reporting, coverage reporting, and error precedence.
  - Trace reconstruction and counterexample printing through safety and liveness paths.
  - State queue interaction with fingerprint set and trace file at full Java fidelity.
- Complete liveness checker parity:
  - final SCC/cycle checking flow, accepting-component error reporting, lasso reconstruction, violated-property attribution, DOT output edge cases.
  - retire or clearly reconcile any temporary in-memory liveness paths once disk graph worker parity is complete.
- Complete FPSet/DiskFPSet parity:
  - fine-grained striped locking and block IO concurrency, disk flushing, checkpoint/recovery edge cases, management/MX statistics.
  - distributed FP set manager/proxy behavior.
- Complete value and tool audit:
  - remaining operator-evaluation edge cases, module functions, TLC standard modules, debugging/lazy-value caching semantics, probabilistic mode gaps that Java supports.
  - ensure no Go map iteration leaks into user-visible order or fingerprints.
- Complete simulator parity:
  - finish comparison to `Simulator.java`, `SimulationWorker.java`, `RLSimulationWorker.java`, and `RLActionSimulationWorker.java`.
  - verify trace-file names/content, continuation behavior, post-condition/error-code precedence, periodic condition handling, and action-flow graph output.
- Complete management/distributed/debugger surfaces only after core local checker behavior is faithful.
- After breadth-first mechanical port is coherent, then port the Java TLC test suite under `test/tlc2` into Go `_test.go` files in `tlc/`.

## Current Position

- Latest commits:
  - `4d8e5e8 Track TLC port progress`
  - `ca4060d Honor disk graph recovery pointers`
  - `e985339 Port memory FPSet variants`
  - `835dd2b Align LongObjTable growth`
  - `bbba936 Mirror ObjLongTable probing`
  - `d282d2b Mirror byte-array queue pool cleanup`
  - `0b3230c Port DiskStateQueue pool cleaner`
  - `0de29bb Add parser-backed liveness operator test`
- Last verified commands before this file:
  - `go test ./tlc`
  - `go test -run TestParseLivenessFromParsedSANYSpec ./`
- Immediate next steps:
  1. Continue breadth-first into `ModelChecker.java` and `Worker.java`, focusing on checkpoint/recovery cleanup, trace reconstruction, and error precedence.
  2. Then continue into remaining liveness/checkpoint/distributed pieces.
  3. Keep `PORT_PROGRESS.md` current before each coherent TLC commit.
