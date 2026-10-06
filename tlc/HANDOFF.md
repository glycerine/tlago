# TLC Port Handoff

Updated: October 6, 2026. Full-workspace verification baseline: `cbdf35a`.

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

The full normal workspace suite passes on `cbdf35a`:

| Package | Result | Duration |
| --- | --- | ---: |
| Root package | Pass | 1,706.569 seconds |
| SANY tests | Pass | 1.068 seconds |
| TLC | Pass | 781.592 seconds |
| CLI command | No test files | — |

Session `44100` returned status 0 and is retired. Log:
`/mnt/oldrog/tmp/tlago-master-streaming-parity-workspace.log`. This verifies
buffered streaming fingerprint storage, the current CLI, all enabled original
model ports, and the full original disk-queue growth workload. It excludes the
outside-suite fingerprint random overlay.

The subsequent full normal TLC package suite also passes on `a915e08`, in
791.920 seconds. Session `63702` returned status 0 and is retired. Log:
`/mnt/oldrog/tmp/tlago-named-recovery-final-tlc.log`. That snapshot adds faithful
named-file checkpoint recovery, source invariant semantics and runtime rename
causes. These results supersede the older `305a13f` workspace baseline.

Subsequent source-parity corrections preserve scan I/O exceptions, off-heap
flusher lifecycle, validation before replacement and the reopened-file scan.
The model-test runner now applies Java Ant's off-heap / `512k` settings before
parsing instead of substituting a small MSB set. Latest checkpoint file helpers
rename directly, copy without creating parents, replace destination links and
propagate typed I/O failures. Detailed implementation/run chronology is in
`PORT_PROGRESS.md`; the full-suite receipts above predate these later changes.

Focused verification:

| Scope | Result | Duration / receipt |
| --- | --- | --- |
| Eight original safety/checkpoint/recovery/liveness/legacy models with Ant settings | Pass | 158.473 seconds, `d611cba` valid-setting path |
| Original factory/CLI checks, default and source memory settings | Pass | 1.072 / 0.963 seconds |
| Complete original heap/MSB/ShortDisk/off-heap classes and manager checks | Pass | 583.638 seconds, `f764bb1`, session `80888` retired |
| Latest file-helper correction: existing focused fingerprint/commit checks | Pass | 0.181 seconds, session `27462` retired |
| State-functor defaults: existing tool/debugger checks and original models | Pass | Core 0.024 seconds; seven init models 0.733 seconds; three debugger models 0.519 seconds |
| Numeric CLI correction: original TLC/WarningControl classes and checkpoint models | Pass | 0.017 / 5.638 seconds |
| Simulation argument correction: original CLI classes and existing scheduler check | Pass | 0.017 / 0.016 seconds |
| DFID/fingerprint CLI branches: original CLI classes and four DFID model methods | Pass | 0.016 / 1.529 seconds |
| CheckImplFile/parser and MP.getError correction: original MP/CLI classes | Pass | 0.016 seconds |
| Ant class isolation: original 48-worker TLCGetAll then five-second checkpoint | Pass | 5.329 seconds, session `10096` retired |
| Persistent CLI parameter state: original CLI/debugger/dump-load methods | Pass | 0.023 / 1.249 seconds |
| Packaged TLC model integration: existing CLI behaviors and original subset methods | Pass | 4.638 seconds; original CLI classes 0.017 seconds |
| Off-heap barrier failure semantics: original 22 short cases and 48-worker/checkpoint sequence | Pass | 0.077 / 5.319 seconds; sessions `19096` / `18062` retired |
| Runner completion and memory banners: original message/CLI checks and trace/checkpoint models | Pass | 6.721 / 1.756 seconds; sessions `9440` / `98861` retired |
| Original factory followed by complete off-heap class with Ant isolation | Pass | 46.331 seconds, session `33613` retired; includes 99,999,999-entry index |
| Recovery/seed ordering: original simulation/recovery/CodePlex models | Pass | 298.728 seconds, session `62916` retired; focused component checks 6.354 seconds |
| Remaining 186 TLC tests after original off-heap class isolation | Pass | 125.929 seconds, session `14660` retired; retains full index/long bounds |
| Missing MP diagnostic templates: original MP/warning checks and trace/liveness models | Pass | 0.014 / 1.900 seconds, sessions `38266` / `93875` retired |
| Latest all-package compilation | Pass | Session `12720` retired |

Manual source/native observations also confirm named checkpoint contents,
missing-parent I/O failure, destination-link replacement and unchanged referent.
These are manual verification, not additional unit tests or inventory credit.

The latest off-heap recovery correction uses the current flusher directly when
recovery fills the table, retaining source statistics and exception boundaries.
Only probe exhaustion falls back to ordinary put/eviction. Final 22 short original
off-heap cases pass in 0.067 seconds; three original checkpoint/recovery models
pass in 4.304 seconds; final all-package compilation passes (session `97004`).
Manual unchanged-Java/native counts match for full-table recovery and configured
probe exhaustion. No test assertions, bounds or inventory counts changed. The
full-class receipt above predates this correction; see progress for exact scopes.

The shared off-heap barrier now stops on the first eviction failure and retains
Java Phaser's incomplete phase and pending-flush state. It does not continue
flushing other sets or release waiting workers after a failed callback. Further
arrival reports the source illegal-state diagnostic; worker registration retains
the 65,535-party limit and waits behind an incomplete advance. Successful phase
advancement preserves source flag-clearing order and signed 31-bit phase wrap.
Manual unchanged-Java/native failure observations agree; original short methods
and the 48-worker/checkpoint sequence pass. No new tests or inventory credit.

State-generation functor defaults now preserve Java's message-less
`UnsupportedOperationException` for unimplemented `setElement`, `hasStates`,
and unary next-state insertion. Existing callbacks retain their dispatch.
The focused receipts above verify the correction; no tests or assertions changed.

Numeric CLI flags now preserve Java signed-32/64-bit parsing, BMP decimal digits,
worker-auto trimming and interval overflow/assignment order. Original CLI classes
and checkpoint models pass. Manual unchanged-Java/native edge cases match; these
are observations, not new test ports. The follow-up simulation audit also restores
the source `NumberFormatException` boundary, replace-all filename handling and
empty-file presence check. Original CLI classes and the existing scheduler check
pass; manual unchanged-Java/native outcomes match. DFID/fingerprint numeric
branches also retain their individual source diagnostics and assignment order;
four original DFID model methods pass. The separate `CheckImplFile` parser now
preserves its own worker, coverage and message-code behavior, string presence
and exception boundary. Source `MP.getError` formatting is implemented; original
MP/CLI tests and manual unchanged-Java comparisons pass. See progress for receipts.

Repeated debugger/load-trace options now preserve source option state and the
first implicit view. Post-condition parsing uses source trailing-empty split
semantics. `nomonolith` is accepted and consumed as Java does; neither the pinned
source nor Go changes output for that token. Help and README explain this.
Original CLI/debugger/dump-load methods and manual source/native observations
pass; their receipts predate completion of the retired workspace snapshot below.

Runner completion now follows Java: ignore user-output cleanup I/O failures,
print the finished message, then generate a trace spec. Trace-generation I/O
reports its original diagnostic without changing the checker result. The
original generator's ignored mkdirs result and resource-close catch are retained.
Both trace-generation messages now use the source MP text. Startup banners report
the configured heap and direct-memory budgets instead of reserved Go memory and
zero. Original message/CLI, trace/checkpoint and writer tests pass; manual Java
and Go observations confirm completion order, budgets and preserved results.
These changes postdate the retired full TLC snapshot `73adde2`.

Recovery now completes before fingerprint/random initialization. A failed
intern-table recovery leaves the polynomial and enumerable seed unchanged, as
Java does. Simulation with an automatically chosen seed ignores aril for RNG
advancement while preserving its parsed option value; the enumerable seed is
set once before tool construction. Original simulation/recovery/CodePlex checks
and source/native manual observations pass. No assertions or counts changed.

Missing MP storage, metadata, CHOOSE, liveness and trace diagnostics now retain
Java's text instead of numeric fallback. Pool cleanup preserves its error/warning
variants; temporal violations preserve zero/single/plural-property wording.
Back-to-state and debug-state formatting retain source tool/debug branches.
Unknown codes use Java's wrong-invocation diagnostic. Manual source/native
observations match 1,344 rows across those branches; original tests pass.
The 186-test remainder is retired with status 0. A fresh full workspace receipt
is still needed after these production corrections.

Current normal verification jobs:

| Job | Compiled source snapshot | Session | Log under `/mnt/oldrog/tmp` |
| --- | --- | --- | --- |
| Full original MSB random draft, no timeout | `a915e08` | `63113` | `tlago-long-msb-random-unlimited-final.log` |
| Full original LSB random draft, no timeout | `a915e08` | `59782` | `tlago-long-lsb-random-unlimited-final.log` |

Poll these handles before launching duplicate suites. Workspace `71687` is
retired with status 1: root passes in 1,519.820 seconds and SANY in 1.071 seconds,
but TLC fails in 46.483 seconds in the older handwritten
`TestOffHeapDiskFPSetDuplicateMergeWarnsLikeJava`. It expected a duplicate full
flush to succeed, whereas unchanged Java asserts `Broken disk index.` before
publishing the file count. The inaccurate non-Java test and its unused slice helper were removed. The
complete original `testMergeDuplicate` retains all ten output values and six
warnings. The final 22 original short off-heap
cases pass with source validation unchanged. This receipt is not a green
workspace result and predates the correction.

 The workspace run at `d611cba` is now retired with status 1: root timed out
in `TestJavaCheckpointWhenTimeBound`; TLC passed in 820.531 seconds and SANY
in 1.057 seconds. The log is `tlago-source-ant-fpset-workspace.log` under
`/mnt/oldrog/tmp`. The source Ant runner forks each test class; the Go model
helper had retained the off-heap static barrier from the earlier 48-worker
`TLCGetAll` model. Focused original sequence `17783` reproduced the wait and is
retired with status 1. The model setup now initializes a fresh off-heap singleton,
matching the source fork. Ordinary runtime registration remains unchanged. The
same original sequence passes in 5.329 seconds (`10096` retired), retaining all
workers, interval settings, time bound and assertions. The failed workspace
snapshot predates this setup correction and later production corrections. The completed broader fingerprint receipt includes
the original 99,999,999-entry index method and retains all source bounds.
No job in this table has full-run credit yet. None uses `-race`.

Full TLC `33649` is retired with status 1 at `73adde2`, in 672.555 seconds.
The original off-heap class hit an uninitialized fingerprint set retained from
an earlier factory class. Ant uses a fresh JVM per class; the native original
class now initializes its own singleton before its methods. The exact factory
then off-heap sequence reproduced the failure in 0.047 seconds and now passes
in 46.331 seconds, including the complete original index method. Production
singleton lifetime and original assertions are unchanged. This resolves that
failure; it is not yet a full-suite success receipt.

The MSB and LSB random jobs retain all 2,147,483,648 iterations, default
factories/configuration, seed, checkpoint calls and assertions, with `-timeout=0` and no `-race`. Expect hours.
LSB completed its first 536,870,912-entry flush and resumed insertion without a
reported failure. Quiet flushes are not terminal jobs. Earlier sessions `98446`,
`83220` and `59015` are retired without full-run credit; see progress receipts
for their deliberate termination reasons.

The full random family draft for LSB/MSB/OffHeap remains outside the enabled
suite, without completion credit. Its LSB and MSB cases use the live jobs above;
do not edit their shared draft while those runs are in progress. Its full source
loops and factory settings
are in `/mnt/oldrog/tmp/tlago-long-random-family-draft_test.go`, with overlay
`/mnt/oldrog/tmp/tlago-long-random-family-overlay.json`. The separate full MSB
draft and overlay use prefix `/mnt/oldrog/tmp/tlago-long-msb-random-full-`.

The Java archive recovery port exposed and fixed late intern-table restoration.
The full new recovery method, five short component checks, and all-package
compilation pass. Eight related original checkpoint, recovery, EWD840 and DFID
model tests pass normally in 230.719 seconds. This is focused verification of
those model ports; the full `cbdf35a` workspace receipt above also covers them.

TLC is now the default CLI runner. Direct invocations such as
`tlago -workers auto MC.tla`, `tlago modelcheck -workers auto MC.tla`, and
`tlago mc -workers auto MC.tla` use the same Java-shaped flags and checker.
The selection flags `--tlc`, `-tlc`, `--go-tlc`, and `-go-tlc` have been removed,
as has the bounded checker's CLI path and its `-maxStates` option. The older
bounded-checker library API is still covered by its existing tests.

The ordinary CLI also handles Java's packaged-model mode: omit the module when
`CLASSPATH` contains `/model/MC.tla` and its configuration. It loads generated
properties, uses the packaged resolver, selects tool output and disables
checkpoints. Direct invocation, `modelcheck` and `mc` use the same runner.
Manual unchanged-Java/native runs of adapted original fixtures match failure
code 2147 and successful 3-generated/2-distinct state counts. These observations
add no test-port credit. The retired `71687` workspace snapshot predates this correction.

The guide in `cli_help.go` prints flag summaries, explanations, Java/Toolbox
correspondence, defaults, examples, and current implementation limits. Help
exits successfully without loading files or starting a checker. README includes
an updated help transcript. Focused normal checks pass: root 6.365 seconds,
SANY 0.014 seconds, and TLC 0.014 seconds. Manual checks covered 34 direct,
modelcheck, and mc invocations, including exhaustive checking, simulation,
violation traces, removed-option errors, and help. Original Java-derived tests
and the test-port inventory are unchanged. The full `cbdf35a` workspace receipt
above includes these CLI changes. Historical interrupted race runs have no pass
credit; never mix the long workloads with `-race`.

The CommunityModules progress changes are already committed as `840d494`.
Under `go test -v`, the test streams subprocess output and reports phase progress
with periodic heartbeats. To check it remotely from the repository root:

```bash
go test -v -count=1 -timeout=60m -run '^TestJavaCommunityModulesAnt$' .
```

The earlier model-config failure comparing `TRUE` with `"blue"` was fixed and is
covered by the green workspace suite. Reopen it only if new evidence warrants it.

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
