# TLC Port Handoff

This handoff is for the next Codex/model after the system upgrade. It records
the active goal, the current working rules, what is already done, what to avoid,
and the best next steps for continuing the Go port of Java TLC.

## Current Snapshot

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

Latest mail dependency progress: Activation MIME parsing/casing and Mailcap
registry parsing/discovery/lookup have their existing Java tests translated.
CommandMap eager initialization/get/set/reset/security and the bundled handler
loadClass/newInstance algorithm are now implemented, with reentrant class and
instance synchronization, permanent initialization failures and lazy resources.
Read the latest progress entry for verification and the explicit JVM boundaries.
ActivationDataFlavor's own constructors/getters/equality/normalization and its
inherited AWT metadata/comparison paths now have native ports. Preserve its
shadowed fields: inherited metadata/hash/overloads often see null base fields.
AWT MIME parsing uses a separate grammar and Hashtable enumeration; its error
and partial-mutation quirks must not be replaced by the Activation parser.
All four existing ActivationDataFlavorTest methods are translated and pass.
Bundled plain/HTML/XML text handlers now have native class lookup, flavor arrays,
transfer/content/write methods and their nine existing Java assertions translated.
Preserve default charset use, ignored MIME charset parameters, stream no-close,
pending high surrogates through flush and failed encoder-buffer state. The modern
Geronimo javamail handler body differs from this bundled activation handler.
Bundled image handlers now have native algorithms and default class lookup;
ImageIO MIME filtering/factory iterators have native ports and exact standard
fresh-reader/writer validation. Preserve the missing reader input and plain
OutputStream writer output, including TIFF's distinct rejection. Keep the
RenderedImage-before-BufferedImage branch order, unknown-type trailing space,
provider prefetch/checked-I/O deregistration/null result and no stream/graphics
close/dispose. ImageIO now has the native ordering graph and object-key hash-map
operations, wired into standard reader/writer enumeration in source registration
order. Preserve copied in-degrees with live out-edges, cycle/blocked-node omission,
node disposal, old-node iteration after clear, hash/comparable tie ordering,
partial mutation on virtual-key failures, captured table references across
reentrant key callbacks and movable=false iterator tree removal.
ServiceRegistry/SubRegistry now have native registration/leaf-class replacement,
category/overload checks, callback catch/partial mutation, captured-context clear,
reentrant monitors, live category/unordered iterators, filtering and finalizer bodies.
Standard reader/writer providers use that same registry implementation. Preserve
the unordered iterator removal that leaves graph/context entries, and clear's
callback exceptions that leave partial class-map removal and retained graph/contexts.
All 36 source methods match installed JDK bytecode; 11,080 actual-JDK/native -race
comparisons include callback reentry/errors, mutation, filtering and security-manager
null-context callback suppression. Existing ServiceRegistryRestriction and
ServiceRegistrySyncTest are translated after implementation and pass in Java/Go.
The seven OpenJDK SPI base classes now have native constructor, getter, clone,
factory-delegation and metadata-format lookup algorithms. Preserve mandatory versus
optional array validation, mutable STANDARD type-array identity, protected fields
versus virtual getters, default cache/lossless flags, null factory arguments and
metadata exception/cause boundaries. All 51 non-abstract source methods match JDK
bytecode; 7,914 actual-JDK/native -race comparisons include named-module exports.
The two existing SpiTest constructor methods are translated after implementation.
The six standard stream SPI constructors, descriptions, cache flags, class inventory
and factory branches are now ported. Preserve File/RAF catch(Exception) null returns,
output-only stack diagnostics (without the registry diagnostic header), escaping
Error, the output RAF diagnostic saying "input", and stream file/memory cache selection.
All 28 source methods/initializers match installed JDK bytecode; 16,128 comparisons
against the installed SPI classes with instrumented backing constructors pass -race.
The unchanged SpiVersionNumbers and its Go translation pass. Actual FileImage/
FileCacheImage/MemoryCacheImage stream construction, storage, disposer and operations
still require their native implementations or explicit constructor providers; missing
constructors fail explicitly outside the source catch rather than returning null.
The standard reader/writer concrete constructors/metadata, schemas and non-null AWT
type-specifier creation remain required. Module/loader/reflection/privileged operations
use explicit VM providers. Continue with those dependencies and
IIORegistry/ServiceLoader/AppContext lifetime. ImageIO must capture the
AppContext registry at class initialization.
Discovery is an explicit lazy provider boundary; no adapter means advancing lookup
fails explicitly, rather than returning an invented empty provider list. Secured
access-context capture/privileged execution requires its VM provider. The native
default runs without a SecurityManager. Official upstream OrderingTest,
DeregisterOrderedSpiTest, RegisterPluginTwiceTest and DeregisterAllSpiTest remain
for porting after IIORegistry and required SPI metadata. VM identity hashes/class/
comparable discovery, compatible-stream initialization, full concrete SPI metadata,
AWT rasterization and image codecs remain required. Continue multipart/MIME
and DataHandler object/factory/cache/buffered-pipe dependencies after these slices.
Read the latest progress entry for bytecode and Java/native verification.
Remaining AWT work includes full class
loading/initialization, text selection/readers, object and MIME externalization.
Full JVM class inventory/linking/loader isolation, stack/cast diagnostics and
native thread/URL/locale/charset providers remain required.

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

This phase is complete when the Go code has a coherent, faithful mirror of the
Java TLC architecture and behavior surfaces, with `PORT_PROGRESS.md` indicating
no major unaudited core areas remain. Translate existing Java tests
feature by feature during this phase, after their implementations are ported,
as the user requested; do not defer all tests until the entire port is complete.
