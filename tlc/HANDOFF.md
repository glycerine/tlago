# TLC Port Handoff

Updated: October 6, 2026. Full-workspace verification baseline: `7d712b0`.

This is the current restart guide for the Go TLC port. Detailed audit history
and run receipts belong in [PORT_PROGRESS.md](PORT_PROGRESS.md). Older versions
of this handoff remain in Git history; their pending jobs and next steps must
not be mistaken for current work.

## Goal and scope

Faithfully port Java TLC to Go, then translate its existing correctness tests.
When a translated test fails, inspect both the translation and the production
implementation. Fix implementation shortcuts before proceeding. Preserve the
original assertions, constructor settings, setup, teardown, seeds, parameter
matrices, fixtures, and workload bounds. Do not invent replacement tests or
weaken tests to obtain a passing result.

The Java reference checkout is `../tlaplus`, pinned to
`8f4bc8b73ad1202774a6bf70143436f8ba50aab0`. TLC sources and tests are under
`../tlaplus/tlatools/org.lamport.tlatools/`. The Go module is
`github.com/glycerine/tlago`; most TLC implementation lives in package `tlc`,
with parser integration and model tests in the repository's root package.

Current priority: finish faithful Java TLC parity on `master`. The new reusable
rpc25519/Tube distributed service is postponed until the remaining port is
complete. Do not resume its unfinished code or its service BDD work here.

The original checkpoint-on-violation and time-bound model tests are now complete.
Both pass normally with their full assertions; the time-bound test retains the
source five-second limit. The generated checkpoint trace recheck is also complete
and passes with its original checkpoint interval and exact seven-state trace.
The CodePlexBug08 recovery model also passes against the unchanged Java archive
with three workers and all original assertions. Intern-table recovery now occurs
before tool construction, matching Java and preserving checkpoint identities.
Debugger/scoped-identifier methods are already mapped.
JPF concurrency verification and benchmarks remain separately tracked; JVM-only
assertions and source-failing methods require honest reconciliation rather than
invented Go equivalents. Keep pending entries visible in
[TODO_TEST_PORT.md](TODO_TEST_PORT.md).

Email reporting is forbidden and its removal is already committed. Do not
restore JavaMail, SMTP, MIME/Activation, ImageIO/AWT, image codecs, or JVM
emulation pursued for email. Preserve generic exceptions, console output,
packaged properties, networking primitives, OpenJDK notices, and `x/text`.
Do not resume SANY XML or ApalacheIR corpus sweeps without a user request.

## Current verified state

The latest completed full normal workspace suite passes on `7d712b0`:

| Package | Result | Duration |
| --- | --- | ---: |
| Root package | Pass | 1,519.135 seconds |
| SANY tests | Pass | 1.147 seconds |
| TLC | Pass | 787.539 seconds |
| CLI command | No test files | — |

Session `15749` returned status 0 and is retired. Log:
`/mnt/oldrog/tmp/tlago-import-context-workspace.log`. This verifies its recorded
snapshot with original bounds and normal execution. It predates the subsequent
INSTANCE, direct Context, file-loading abort and module-loading corrections.
Historical failed and successful receipts belong in `PORT_PROGRESS.md`.

Full normal workspace session `28493` is live on the current module-loading
production snapshot. It preserves original bounds, failfast and a 60-minute
timeout, without race instrumentation. Both `GOTMPDIR` and `TMPDIR` use workspace
`.codex-gotmp`. Log: `/mnt/oldrog/tmp/tlago-module-loading-workspace.log`.
No success is claimed until this same handle returns terminal status. Do not
start a duplicate full workload.

Latest focused verification:

| Scope | Result | Receipt |
| --- | --- | --- |
| Existing parser/resolver/semantic/CLI/XML and original INSTANCE/action-level/trace/warning methods | Pass | 4.771 seconds, session `31624` retired |
| Original SANY method selection on final source graph | Pass | 0.717 seconds, session `31624` retired |
| Complete SANY package after inherited-context correction | Pass | 1.359 seconds, session `51638` retired; predates final search-flag preservation |
| All-package compilation | Pass | Final module-loading graph, session `51926` retired |

Source/native scratch observations verify exception types/causes, delayed output,
constructor diagnostics and storage artifacts. These are manual evidence, not
new persistent tests or test-port credit. Preserve each receipt's exact scope.
The record-lint matrix matches Java in 33 valid cases. Six full front-end plus
constructor observations now also match with the real lint phase included.
Seven manual semantic-output observations match Java for one error, dependency
errors, chains, siblings, nested modules and a successful dependency. These cover
reporting order and accumulation. Eleven further observations match generation
versus level gating, nested graphs, definition-before-top-level ordering, proof
levels, general assumption checks and unique named-theorem errors. Sixteen
import observations match Java for declarations, definitions, named facts,
class conflicts, diamonds, parameterized instances, reuse, accumulated warnings
and explicit versus implicit instance namespaces. Nine INSTANCE observations
match source collision kind/arity rules, repeated parameterized wrappers,
parameter-free reuse, local definition order and existing EXTENDS bindings.
These are bounded audit
observations, not a proof of complete semantic-node or context iteration parity.

The latest checker correction initializes coverage, liveness and cached config
in the parent constructor before storage and workers. Liveness I/O failures are
raised immediately; DFID's worker/liveness assertions follow parent construction.
BFS creates queue, trace and the selected fingerprint implementation in source
order, without an unused default fingerprint initialization. Disk graphs do not
create missing parent directories. RandomAccessFile opening failures retain the
source FileNotFoundException. The original five-second time-bound test passes
unchanged. Trace and worker constructors now propagate opening failures immediately and
preserve source path concatenation. Buffered random-access files support all four
source modes, including synchronous flags, and preserve opening versus later I/O
exception types and messages. Experimental liveness simulation creates unique
worker temporary directories itself, as Java does; failures are immediate.
Further constructor and front-end parity remains audit work.

## Long verification receipts

The earlier LSB session `59782`
failed after 19,420.414 seconds and is retired. It was compiled at `a915e08`;
log `/mnt/oldrog/tmp/tlago-long-lsb-random-unlimited-final.log`. Last logged
progress reached 2,144,786,516 of 2,147,483,648 iterations. Fingerprint merging
failed with `IOException: no space left on device` on `/mnt/oldrog/tmp`.
No full-workload success or inventory credit is claimed. Its unchanged draft and
overlay remain `/mnt/oldrog/tmp/tlago-long-random-family-draft_test.go` and
`/mnt/oldrog/tmp/tlago-long-random-family-overlay.json`.

LSB session `32959` is live on the unchanged full original random draft, using
both temporary-directory variables on the large workspace volume. Log:
`/mnt/oldrog/tmp/tlago-long-lsb-random-large-volume.log`. The actual fingerprint
file is under `.codex-gotmp/lsb-random-tmp`; placement has been verified. This
snapshot includes the latest buffered-file and trace corrections, before the
simulator correction. Preserve original bounds, factory, seed, assertions and
checkpoints; use no race instrumentation. No success or inventory credit until
the same handle returns terminal status. Do not reclaim unrelated files.

MSB session `63113` is retired with status 0. Its full random draft passes in
13,343.32 seconds, including all 2,147,483,648 insertions, checkpoint commit,
invariant check and final size. It was compiled at `a915e08` before later
flusher endpoint/count assertions and file-helper fixes. Preserve that limited
receipt; it does not verify newer changes or earn inventory credit. Both drafts
retain original factories/configuration, seed, checkpoints, assertions and
bounds, use `-timeout=0`, and have no race instrumentation.

## User-visible behavior already completed

Direct TLC arguments, `modelcheck` and `mc` use the same TLC runner and Java
flags. Do not restore `--tlc` or the bounded checker's CLI path. Help in
`cli_help.go` explains flags, defaults and Java/Toolbox correspondence. Packaged
models load their properties and files through the packaged resolver. Ordinary
and packaged loading share deferred tool construction after intern recovery.

Configuration I/O failures exit with source status `255` and omit FINISHED and
trace generation. SANY tool markers and STARTING retain source phase order.
Unexpected parsing Exceptions become checked failure; semantic Exceptions chain
through FrontEndException and TLC_PARSING_FAILED2; Java Error propagates. Delayed
SANY output releases only on the unexpected checked-exception path. See
`TLC_ARCH.md` and `PORT_PROGRESS.md` for details and remaining diagnostic gaps.

CommunityModules progress changes are committed as `840d494`. Under `go test -v`,
its subprocess output, phase progress and heartbeats remain visible. Remote check:

```bash
go test -v -count=1 -timeout=60m -run '^TestJavaCommunityModulesAnt$' .
```

The earlier model-config failure comparing `TRUE` with `"blue"` is fixed and
covered by the full suite. Reopen it only if new evidence warrants it.

## Test-port inventory

[TODO_TEST_PORT.md](TODO_TEST_PORT.md) is the authoritative class and method
inventory. Its current totals are:

| Suite | Translated method contexts | Pending contexts |
| --- | ---: | ---: |
| Main TLC | 1,260 of 1,269 | 9 |
| Shared utilities | 55 of 56 | 1 |
| Long tests | 15 of 22 | 7 |
| Concurrent tests | 2 of 17 | 15 |

In the main suite, 617 of 626 non-ignored concrete classes are complete. A method
context is a concrete Java class plus a method; inherited methods count once
per concrete subclass. These totals measure translations, not implementation
coverage or universal behavioral parity. Original ignored methods are counted
separately. Do not credit scratch probes, fixture copies, reduced workloads,
configured stress audits, or Java failures as completed translations.

`DumpLoadTraceTest` is complete: all 32 enabled methods and the three original
ignored methods are translated in `tlc_dump_load_trace_java_test.go`. The whole
unchanged Java class and the Go class passed. The Go class also passed its
previous focused race run. Those class receipts predate the final counter-lock
correction; the normal workspace receipt above verifies that counter correction.
Historical Java replay failures under deliberately paced scheduling remain
valid evidence. Current passing runs do not establish stability under every
possible schedule.

## Remaining work and known blockers

Continue the core production audit and unresolved original methods below.
Current concrete production gaps include general syntax-error ErrorDetails
rendering, generation traversal and multiple-binding context iteration order,
remaining constructor and semantic-node boundaries. Missing modules and filename/module-name mismatches now abort loading with
source E4220/E4221 details and null or actual importing-module locations.
The existing front-end exception boundary reports these failures; the native
library API returns diagnostics and preserves a previously parsed root.
Seven file-loading observations match Java. The source unresolved-module search
now exhausts EXTENDS before INSTANCE, restarts from the root after each binding,
and separates file parse units from inner modules. Cycle diagnostics retain
source E4222 and the complete filename path. Semantic order is derived from the
recorded parse-unit relationships; inner modules are generated within their
owning external module. A forward INSTANCE now preserves inherited symbol
conflicts and the external-module-table E4223 conflict. Sixteen further loading
observations match Java, with only independent extraction directory names
normalized in the scratch comparison; raw outputs remain available. General parser parity remains
unproven. The original SANY
`TestContext` method now exercises Context directly, with its original failure
result, one-error count, E4224 code and declaration/definition parameter
assertions. It lives in root `sany_context_java_test.go` so it can access the
private context implementation; the parser-fixture surrogate was removed.
The direct context and production EXTENDS path share diagnostic construction,
including structured parameters and source locations. EXTENDS conflicts now retain actual symbol
locations and definition provenance, compare source semantic-node classes, and
reuse parameter-free instance definitions. Substituted declarations are not
INSTANCE exports; only an explicit named instance creates a qualified namespace. Per-module semantic
reporting now follows the shared accumulated Errors instance; nested module
diagnostics are included in their enclosing external reporting iteration.
Generation completes before the raw shared Errors.isSuccess gate permits real
level work. Warning elevation does not change raw Errors success. Further
semantic-node parity is still required; these checks do not prove completion. Buffered-file modes,
invalid-mode exceptions and trace/worker opening boundaries were reconciled in
`db62dfb`; preserve those verified fixes. Do not add
synthetic phase output. The actual Java record linter is now ported: declaration
and formal-parameter dependencies, same-domain EXTENDS suppression, binding and
proof scopes, exact warning text and a distinct phase after successful semantic
analysis. Matching
checker output or a high test translation percentage does not prove full parity.

Five main-suite contexts have known source-failing or JVM-specific
reconciliation issues below. The other four pending contexts are original assumption-disabled
distributed transport models; checkpoint models are now complete. Consult the inventory for each exact disposition.

| Eligible class | Unresolved issue |
| --- | --- |
| `DistributedTrace` | Unchanged Java and the temporary faithful Go translation fail no-generated-TE-spec and inherited success-exit assertions (actual 12). Both produce the original eleven-state trace; retain four workers and original expectations. |
| `AssertExpressionStack` | Unchanged Java and the faithful Go draft fail `assertNoTESpec` and the inherited success-exit assertion. Trace generation is enabled by the base, and the actual exit is 14. |
| `DepthFirstTerminate` | The original runtime-derived worker setting requests 48 workers on this host. Java rejects multithreaded DFID, records `GENERAL`, and exits 255; see issue 548. Do not substitute one worker. |
| `InliningTest` | The whole method includes HotSpot/JFR compiler-inlining records, reflective annotations, and JVM callee descriptors. Porting only its checker assertions would be incomplete. |
| `LivenessSymmetryWarning` | Java and Go satisfy all three warning/diagnostic assertions but fail the inherited success exit: expected 0, actual 13. The original model produces a liveness violation. |

Keep these entries marked **Reconcile** until their original expectations or
JVM-specific requirements receive an explicit disposition. The faithful drafts
remain in ignored scratch; no new persistent skips or production divergences
have been introduced to hide the failures.

Additional pending work:

- `TLCRuntimeTest.testIsThroughputOptimized` asserts the JVM ParallelGC selection
  made by the original Ant flag. Do not fabricate a successful Go GC response.
- Six inherited long fingerprint contexts retain their full loops: 2,147,483,648
  random iterations or 3,221,225,473 sequential iterations, across three factories.
  Their full drafts are prepared but not verified or credited.
- `OffHeapDiskFPSetLongTest.testMultipleFlushes` fails in unchanged Java and Go
  when an invariant check reaches an already shut-down flusher executor. Keep
  the original four rounds and insertion counts.
- Three `ConcurrentWriteTest` methods fail identically in Java and Go because
  overlapping buffer flushes or positional-write gaps produce zero values.
  Preserve all original writes and read assertions.
- Nine concurrent fingerprint random contexts still require full stress runs.
  Their default insertion property is 2,147,483,649. Smaller configured audits
  are evidence about the implementation, not completion of these contexts.
- Three partitioned fingerprint methods traverse memory-derived buckets and
  fail their original minimum-size assertions in unchanged Java and Go. The
  default OffHeap case produces 8,388,575 puts. LSB and MSB audits with explicitly
  matched 256 MiB budgets produce 25,165,823 puts each, with all 48 worker counts
  matching Java. These source failures do not earn pass or full-size credit.

The large fingerprint runs need approximately 32 GiB for concurrent main and
merge/checkpoint files, plus reserve. The relocated `/mnt/b` filesystem has approximately 1.3 TiB available as of
October 6. Use ignored `.codex-gotmp` for large test storage and recheck
resources before each large workload. Preserve original bounds. Pending
source-behavior questions have not been resolved by silence.

A separate scheduling concern remains for alias trace replay: source-compatible
multiworker dumps can replay 10 or 11 states with differing final actions.
Java and Go check the invariant before the out-of-model trace constraint.
Do not reorder these checks or weaken replay assertions without an explicit
decision to diverge from Java. See the detailed receipts in `PORT_PROGRESS.md`.

## Implementation details to preserve

Recent fixes restore these source behaviors:

- A disk fingerprint search selects one reader for the whole search. Indexed
  workers bypass the pool mutex; pooled readers return after normal completion.
  I/O failures propagate before the return step, and close errors propagate.
  Reader arrays are published as immutable atomic snapshots.
- Heap table counters increment independently, matching Java's `LongAdder`
  behavior. Stripe locks protect table entries; the native metadata lock still
  protects bucket capacity. Counter reads remain independent.
- OffHeap insertion and lookup use atomic words and CAS rather than a set-wide
  mutex. Eviction selection and reset retain the source CAS assertions.
- Value streams use the original 8,192-byte buffered input/output layers.
  Byte-queue adapters remain direct. Graph node and pointer files use buffered
  random-access files; resets and checkpoint encodings follow Java.
- Intern-table recovery reads through its buffered stream and `atEOF` behavior.
  Unique-string primitives retain the original external string protocol.
- Disk-queue growth retains one reused Dummy-equivalent state, the original
  header, all enqueue calls, and approximately 14 GiB of raw state storage.

Prefer concrete structs and keep TLC in its existing package. Preserve source
collection iteration order wherever output, exploration, diagnostics, coverage,
or fingerprints can observe it. Use insertion-ordered maps, explicit slices,
sorting, or the existing Java HashMap ordering helpers as the source requires.
Do not introduce a second parser, compatibility fallback, or speculative
behavior changes. Consult existing audit notes before repeating old work.

Distributed integration remains deferred. The selected future transport is
`github.com/glycerine/rpc25519` with Greenpack serialization, using asynchronous
peer/circuit/fragment APIs. Preserve TLC deduplication, recovery, and termination
semantics when that work resumes; see `TLC_ARCH.md`.

## Testing and workflow

**Never combine long workloads with `-race`.** Run complete long workloads
normally. Reserve race instrumentation for short, focused concurrency checks.
Avoid broad race patterns that accidentally include long tests. Preserve the
original test bounds in both cases; do not make long tests small for convenience.

Use the local cache and temporary directory for offline Go work:

```bash
env GOCACHE="$PWD/.codex-gocache" GOTMPDIR="$PWD/.codex-gotmp" \
  GOPROXY=off GOSUMDB=off \
  go test -count=1 -failfast -timeout=60m ./...
```

Run checks appropriate to actual changes. Reuse valid results for unchanged
code instead of restarting expensive suites. Tests that inspect real network
interfaces require the unrestricted environment; do not weaken their assertions
to accommodate sandbox visibility. Poll the same live process handle while a
check runs. An observation timeout or silent log is not a terminal result.

Persistent fixtures belong in `tlc/test_vectors/`, never a directory named
`testdata`. Preserve original fixture bytes. Scratch overlays, probes, logs,
and compilation outputs stay in ignored `.codex-gotmp/` and earn no port credit.
Only reclaim regenerable compiler outputs after excluding live process references;
never delete source, vectors, or verification receipts to free space.

Update `TODO_TEST_PORT.md` manually as whole original methods become complete.
Do not regenerate it wholesale with the scratch rendering script, which loses
curated reconciliation notes. Update `PORT_PROGRESS.md` with concise audit and
verification receipts, then commit coherent, verified work. Do not back up,
rebase, stash, or push; the user pushes separately. Git writes require sandbox
escalation, but commits are already authorized. Do not request authorization
again for routine work within this scope.

## Keeping this handoff readable

Maintain this file as a current restart guide. Replace obsolete status instead
of prepending another overlapping history block. Put detailed per-run chronology
in `PORT_PROGRESS.md`. Use ordinary sentences, spaces between words and numbers,
clear tables for comparable results, and backticks for identifiers and commands.
Do not compress prose into strings such as "all32", "exit130", or "PASS10.068s".

Before resuming work, read [PLAN.md](../PLAN.md), this handoff,
[PORT_PROGRESS.md](PORT_PROGRESS.md), [TLC_ARCH.md](TLC_ARCH.md), and
[TODO_TEST_PORT.md](TODO_TEST_PORT.md). Choose the next action from current
pending entries and evidence. The overall porting goal remains incomplete; neither
the green batch nor the high translation percentage establishes completion.
