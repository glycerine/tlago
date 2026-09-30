# TLC Port Progress

## Notes To Future Us

- Do not re-run the old TLA+ XML or Apalache corpus sweeps unless explicitly asked. The active goal is the Go TLC model checker under `tlc/`.
- The source of truth is Java TLC under `../tlaplus/tlatools/org.lamport.tlatools/src/tlc2` and, later, its tests under `../tlaplus/tlatools/org.lamport.tlatools/test/tlc2`.
- The current experiment is breadth-first mechanical porting first. Do not start porting the Java TLC test suite yet. Existing fast Go tests may be run; small utility tests are acceptable.
- Keep the Go code mostly in package `tlc`, prefer concrete structs over interfaces, and use `InsMap` whenever deterministic iteration matters.
- Already audited recently; do not loop on these unless touched: `fpset.go` MemFPSet1/MemFPSet2, `fpset_disk.go` high-level DiskFPSet checkpoint/recovery/merge API shape, `liveness_tables.go`, `liveness_disk_graph.go`, `liveness_tableau_disk_graph.go`, `liveness_process.go`, `liveness_graph.go`, liveness DOT/debug writer call sites in `liveness_check.go`, `liveness_check1.go`, `random_generator.go`, `object_collections.go`, `int_stack.go`, `int_queue.go`, `state_pool.go`, `disk_state_queue.go`, and `simulation_worker.go`/`simulation_worker_modes.go`.
- Already audited recently; do not loop on the standard TLC module override registry unless touched: Java's built-in overrides are `TLCGetSet`, `TLCEval`, `TLCExt`, `Json`, `_TLCTrace`, `_JsonTrace`, and `_Possible`; the older static module surfaces for Naturals/Integers/Sequences/FiniteSets/Bags/TLC/Randomization/TransitiveClosure/Strings are represented in `standard_definitions.go` and `modules_*.go`. Public-looking `Remove`, `FApply`, `FSum`, and FiniteSets list helpers in Java source are commented-out code, not active override surface.
- `CheckImplFile` deliberately keeps trace parsing behind `LoadTraceFunc` for now: the production SANY-to-TLC bridge is in the root `tlago` package and already imports `tlc`, so `tlc` cannot import it without a circular dependency. Do not add a second parser here; wire the loader from the command/front-end layer or move the bridge mechanically if we later choose that architecture.
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
  - Exploration/debug simulation now preserves Java's `ExplorationWorker.halt()` behavior: debugger step-out can command the next-state functor to halt, while state generation separately polls the halted flag.
  - `CheckImpl` now mirrors more of Java's visible control flow: partial-state-space start/completion/failure messages, illegal-transition standalone state printing, implied-action standalone state printing, false-result short-circuiting in `checkTrace`, and Java-shaped `CheckImplFile` trace polling output.
  - Worker trace checkpoint commit now fails when the worker `.tmp` checkpoint cannot be renamed, matching Java's `Worker.commitChkpt()` behavior instead of silently ignoring a missing checkpoint.
  - Core TLC message formatting now covers Java-shaped run/progress/statistics/checkpoint/coverage strings for the message codes already emitted by the Go checker, simulator, and coverage paths.
  - The TLC runner now supplies Java-shaped startup mode banner parameters and finished-runtime strings for BFS, DFID, and simulation runs.
  - BFS model checking now follows Java's recover-before-fresh-start ordering and preserves the `TLC_LIVE_FORMULA_TAUTOLOGY` guard.
  - BFS model checking now emits Java's final safety progress snapshot immediately before final liveness checking.
  - Worker liveness failures now follow Java's call-stack replay path for `EvalException`/stateful runtime failures and preserve the original error after replay.
  - `ModelChecker.doNextFailed` now preserves `EvalException` error codes/parameters and Java's keep-call-stack behavior for known fatal/system-like TLC errors.
  - Checkpoint recovery now rebinds the checker metadir and all trace fragments to `FromCheckpoint`, matching Java's `FileUtil.makeMetaDir(..., fromChkpt)` behavior for resumed runs.
  - BFS `RunTLC` now follows Java's `keepCallStack` return convention and `ModelCheck` replays next-state failures with `CallStackTool` before final summary output.
  - Initial-state exceptions now print the Java-shaped init failure message and replay init generation with `CallStackTool`, including fingerprint-exception handling.
  - Periodic-work liveness/periodic-condition failures now leave the state queue suspended for outer termination, matching Java's error-return path.
  - DFID clean termination now reports success instead of `GENERAL`, and DFID next-state replay can run through `CallStackTool` like Java's `DFIDModelChecker`.
  - DFID init now prints Java-shaped invariant/implied-init diagnostics and replays init exceptions with `CallStackTool`.
  - DFID workers now push the current worker id during `Run`, matching Java `IdThread` behavior for `TLCGet("worker")` and worker-local values.
  - DFID checkpoint recovery now rebinds `Metadir` to `FromCheckpoint`, matching Java's resumed-run metadir setup.
  - Liveness `Check`/`FinalCheck` now honor Java's `LNCheck` gates: periodic checks obey `DoLiveness`, and final checks skip when liveness checking is `off`.
  - Liveness SCC checks now emit Java-style temporal-property start/end messages with graph size and current/complete mode.
  - Disk-backed liveness SCC checking now creates the graph cache for the whole PEM pass, records graph size on the disk graph like Java, and lets liveness violations take precedence over checker failures after all checkers have run.
  - Liveness counterexample reconstruction now keeps Java's raw-vs-printable trace split: raw states feed violated-property attribution and `CounterExample`, printed states go through `ALIAS`, and stuttering lassos emit Java's fairness/specification warning when applicable.
  - Disk-backed liveness graphs now collect Java-style out-degree samples at node insertion and expose in/out-degree recomputation through `LiveCheck`'s auxiliary statistics methods.
  - Safety-like liveness counterexample prefix reconstruction now uses Java's non-prefix `ALIAS` overload instead of passing a `TLCExt!Trace` context.
  - Liveness DOT/debug writer calls now mirror Java's disk-backed liveness paths: initial states are written for non-tableau/tableau checkers, non-tableau transitions carry bit-vector labels and seen/unseen status, tableau cross-product edges are emitted on insertion, and recursive done-expansion writes dotted edges for generated state successors.
  - Legacy in-memory `LiveCheck1` is now ported as a concrete Go struct: trace graph construction, incremental BE/BT graph updates, Tarjan SCC passes, PEM subcomponent checking, and liveness error trace reconstruction are present. `BEGraphNode` now preserves tableau identity so Java's BT-node equality semantics are not collapsed to state fingerprints.
  - Simulator workers now mirror Java's liveness selection: default per-worker checks use concrete `LiveCheck1` with a shared error flag, while disk-backed simulation liveness is only selected via the Java-style `tlc2.tool.Simulator.experimentalLiveness` property or `TLAGO_SIMULATOR_EXPERIMENTAL_LIVENESS`.
  - Liveness check/worker/error-trace skeletons with concrete disk graph fields.
  - Distributed server publishing now writes accepted successors through a Java-style master `TLCTrace.writeState` primitive, assigning the returned trace UID before enqueueing instead of using the worker-local concurrent trace writer.
  - Distributed block selection now follows Java's `BlockSelectorFactory` property surface while keeping one concrete Go selector: static, unlimiting/proportional, limiting, and default statistical modes are selected through the Java property names, and static mode honors `tlc2.tool.distributed.TLCServerThread.BlockSize`.
  - Distributed worker registration now mirrors Java's server lifecycle: registering a worker wakes stuck queue threads, creates the corresponding master `TLCServerThread`, registers it, and starts it; explicit thread construction still uses the lower-level thread registration hook.
  - Distributed server lifecycle now has Java-shaped checkpoint/recover/close, init-state generation through a distributed `DoInitFunctor`, joinable server threads, periodic checkpoint/progress reporting, final summary/success reporting, and a Go-library `ModelCheck` path for explicitly registered in-process workers.
  - Distributed TLC output codes and Java-shaped message text are now present for server-ready, worker register/deregister, worker stats, worker lost, recoverable block-size reduction, FPSet wait/register, server-not-running, server-finished, and VM-version diagnostics; the currently ported distributed paths emit the matching lifecycle messages.
  - `TLCSet("pause", TRUE)` now mirrors Java's synchronized state-queue pause path more closely by holding the concrete queue monitor while waiting for stdin, without using the blocking `SuspendAll` helper.
- Added and kept green many fast Go tests for utility behavior and already-ported pieces.

## Left To Do

- Finish breadth-first audit of Java TLC source against Go implementation, patching true behavioral gaps as they are found.
- Complete faithful model-checker behavior:
  - `ModelChecker`, `Worker`, `DFIDModelChecker`, `DFIDWorker`, checkpoint/recovery, termination, progress reporting, coverage reporting, and error precedence.
  - Trace reconstruction and counterexample printing through safety and liveness paths.
  - State queue interaction with fingerprint set and trace file at full Java fidelity.
- Complete liveness checker parity:
  - final disk-backed SCC/cycle checking flow, accepting-component error reporting, lasso reconstruction, violated-property attribution, and any remaining DOT formatting differences.
  - optional `LIVENESS_STATS` raw statistics printing still needs an output-channel decision; graph collection/recomputation helpers are now present.
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

- Latest committed chunks before the `CheckImpl` parity checkpoint:
  - `6773c66 Wire liveness DOT writer paths`
  - `7af6643 Port in-memory LiveCheck1 core`
  - `e7a0e89 Use LiveCheck1 for simulation liveness`
  - `1570bae Mirror exploration halt command`
- Current checkpoint:
  - `CheckImpl`, `CheckImplFile`, worker trace checkpoint commit behavior, core TLC reporting/coverage message formatting, runner startup/finish banners, standard module override surface audit, `TLCSet("pause")`, distributed master trace writes, distributed block selector properties, distributed worker registration, distributed server lifecycle, and distributed reporting messages have just been tightened against Java.
- Last verified command:
  - `go test ./tlc`
- Immediate next steps:
  1. Continue breadth-first audit into remaining checkpoint/distributed/debugger surfaces and any disk-backed liveness final-SCC details.
  2. Then continue into remaining model-checker/worker parity gaps discovered from the Java source.
  3. Keep `PORT_PROGRESS.md` current before each coherent TLC commit.
