# TLC requirements for a reusable distributed job coordinator

Status: requirements and service contract, October 6, 2026.

TLC is the first client of a reusable distributed job management service built
on rpc25519 and Tube Raft's RMember. This document defines what TLC needs from
that service, especially batch ownership, transaction boundaries, and observable
ordering. [BATCH_PROTOCOL.md](BATCH_PROTOCOL.md) supplies the current concrete
protocol design; its mechanics implement these requirements rather than defining
a TLC-specific scheduler that other clients cannot use.

The contract below is normative for the proposed service. It describes required
behavior, not features already implemented. Existing Java TLC is the reference
for model-checking semantics; retry-safe replies and fenced coordinator authority
are intentional improvements to its distributed protocol.

## TLC's work and the reusable abstraction

TLC explores a graph whose size and fan-out are unknown in advance. A work item
is a predecessor state. Expanding it generates successor states; accepted new
successors become more work. A batch groups predecessor states for efficiency.
An empty result is valid completed work. Neither a fixed task count nor a
predeclared dependency graph can be required by the job service.

TLC also has partitioned visited-state storage, a trace recording parent links,
a disk-backed frontier queue, an intern table, and semantic progress counters.
These are application resources. The job service must coordinate their ownership
and publication without understanding TLA+ values, fingerprint bit patterns,
invariants, parent UIDs, or the Java storage formats.

| Generic service concept | TLC client meaning |
| --- | --- |
| Job definition and artifact identity | Specification, configuration, module closure, checker settings, fingerprint polynomial. |
| Execution identity and authority | One fenced coordinator authorized by RMember for a particular execution. |
| Work item and batch | One predecessor state and a vector of predecessor states. |
| Assignment attempt | A particular worker's authority to expand that batch. |
| Result payload | Successor vectors, fingerprints, parent references, errors, and statistics. |
| Conditional claim resource | Fingerprint set returning which keys this operation newly inserted. |
| Publication obligation | Write each winning successor's trace record and enqueue its state. |
| Completion barrier | No frontier work, active assignments, accepted results, or unfinished publication. |
| Snapshot participants | Queue, trace, fingerprint shards, intern table, and client metadata. |

Use opaque payloads with explicit lengths, checksums, and client codec versions.
Application keys are scoped to a job; equal fingerprints in different jobs do
not imply equal work. Resource operations can use a client-defined key codec.
Do not mandate that every reusable job has fingerprints or conditional claims.

## Authority, identity, and resource fencing

**R1. One authority.** RMember's elected Czar is the job coordinator. Only configured
coordinator candidates participate in its election group. Compute/resource nodes
may be candidates only when configured and capable of recovery. The Tube Raft
leader and the job coordinator are distinct roles.

**R2. Execution identity.** Use JobID, consensus CzarLeaseEpoch, and a Tube-committed
RunGeneration. A retry preserves this tuple. Rollback/restart creates another
generation; coordinator replacement uses the new Czar authority. Ordinary member
updates do not invalidate all outstanding requests. Peer incarnations distinguish
process restarts from connection replacements.

**R3. Fenced admission.** Every command, result acceptance, resource mutation,
acknowledgment, snapshot commit, and terminal job decision is associated with its
execution and authorized sender. Reject stale authority. A message cannot elect
its sender or install a new epoch. Observe the RMember lease/clock-drift contract;
fail closed after the conservative lease deadline without renewal.

**R4. Transition barrier.** A new coordinator cannot open a new execution against
storage still being mutated by the old execution. Fence resource admission and
drain admitted operations, or restore replacement resources into isolated storage.
An unreachable former owner is not assumed stopped. Raft election supplies
logical authority; resource adapters must enforce that authority at side effects.

## Batch ownership and retry requirements

**R5. Continuous ownership.** Removing work from the frontier creates an outstanding
batch obligation before that work can become invisible to completion detection.
Assignment, result acceptance, resource claims, publication, and completion must
not leave a gap where nobody owns that obligation. Scheduling may be asynchronous.

**R6. Attempt exclusivity.** A batch may execute more than once physically after
worker failure or reassignment, but only its accepted attempt may publish results.
Acceptance and supersession are mutually exclusive coordinator transitions.
Reassignment releases predecessor ownership once. A delayed old result must not
publish children, increment semantic counters, or report successful completion.

**R7. Result ownership transfer.** Accept a result only after its complete payload
and publication obligations are retained. Thereafter the coordinator owns it,
even if the producing worker disappears. A worker acceptance acknowledgment is
not acknowledgment that its descendants are already queued or checkpointed.

**R8. Immutable retries.** An operation has a stable identity and content digest.
Duplicate requests replay the original outcome; concurrent copies join one
execution. The same identity with different content is an error. A timeout is
an unknown outcome, not permission to allocate a fresh identity for a mutation.

**R9. Retry scope.** Request/reply replay works across connection loss while the
owning processes and ledgers survive. The initial service does not promise that
an in-memory operation ledger survives process crashes. Such failure invalidates
the execution and requires coordinated recovery or a reported job failure.
Expose this recovery policy as a service capability, not an implicit TLC rule.

**R10. Safe result retirement.** Retain mutation outcomes until the client confirms
that their publication obligations are Applied. Retirement acknowledgments must
be repeatable. Contiguous watermarks or equivalent tombstones prevent a delayed
request from executing again after its reply is discarded. Never discard an
unacknowledged outcome simply because a timer or memory budget expired.

## Transactionality required by TLC

The service must distinguish these stages rather than give one ambiguous "done"
response:

| Observable stage | Required meaning |
| --- | --- |
| ResultAccepted | The coordinator owns a complete attempt result and its remaining obligations within this execution. |
| ClaimResolved | The conditional resource operation has an immutable original answer. |
| Applied | All winning states have trace records and frontier entries; duplicate delivery cannot publish them again. |
| CheckpointCommitted | All required resources belong to one completed snapshot generation discoverable through Tube. |
| JobCompleted | Global exploration is exhausted and the current authority has committed the terminal decision. |

Accepted or Applied acknowledgments are **execution-scoped**, not a claim of
crash durability. Recovery may roll back work after the last completed checkpoint.
The service must expose a generation change and recovery status so clients never
mistake rollback for transparent continuation of the same execution history.

**R11. Atomic conditional claim outcome.** For each fingerprint, at most one distinct
claim operation may receive "new" in a surviving execution. Repeating that
winning operation must return "new" again as part of its original bitmap; it must
not become a losing operation just because insertion already occurred. Preserve
input/result index alignment. TLC's fingerprint equivalence remains probabilistic
as in Java; the service must not silently change the collision policy.

**R12. Claim-to-publication obligation.** A successful claim cannot be detached from
the retained state it obligates TLC to enqueue. Other batches may observe that
fingerprint as present while publication is pending, provided the winner still
owns the obligation. An abandoned owner requires execution failure/rollback;
it cannot be silently forgotten while exploration continues.

**R13. Publish once.** Trace writing and queue insertion form an application-defined
publication sequence. Track completed stages to prevent duplicate replies from
repeating them. Retain the chosen parent relationship and trace UID. A partial
storage failure poisons the execution unless the adapter can prove safe repair.
The initial policy fails/rolls back; it does not guess whether enqueue happened.

**R14. Defined atomicity granularity.** Require atomic outcome identity and per-key
claim semantics, not a database transaction spanning an entire batch, all shards,
trace files, and the queue. A batch can publish incrementally. It becomes Applied
only after every relevant obligation finishes. Operations that partially mutate
then fail cannot be acknowledged as a successful batch.

**R15. Accounting.** Increment semantic generation/acceptance counters once per
accepted logical result, preserving Java's definitions. Transport-byte and
physical-computation counters may count retries but must be labeled separately.
A stale or duplicate result cannot alter the logical counters.

## Linearity and client-visible ordering

"Linearity" here means linearizability of the specified operations: each completed
logical operation has one effective point between invocation and response, and
non-overlapping operations respect real-time order. It does not require a global
order of all worker computations, deterministic traversal, or byte-identical
traces across concurrent schedules.

**R16. Control operations.** Job activation, execution changes, snapshot references,
and terminal decisions use Tube's linearizable records/CAS. A successfully
completed control change precedes later dependent control reads. An ambiguous
reply must be resolved by operation identity or authoritative read, not assumed
successful or failed. This guarantee is conditional on the configured Tube
persistence and lease assumptions.

**R17. Conditional claims.** Claims on the same key have one authoritative order.
A claim completed before a different claim begins must be visible to the latter.
Overlapping claims can choose either winner, consistent with the actual resource
operation. Different shards need no global ordering, and a batch bitmap need not
represent one atomic cross-key instant. Duplicate replay is the response of the
original logical operation, not a fresh membership query.

**R18. Membership reads.** TLC's contains operations may return a pointwise current
view rather than a snapshot of every shard. False "unseen" results can lead to
extra candidate work, subsequently deduplicated by authoritative claims. "Seen"
must mean that a committed fingerprint or a live publication obligation exists
in this execution; it cannot come from an abandoned attempt or old generation.
The TLC adapter must also preserve the source check/constraint ordering.

**R19. Publish observation.** Applied is returned only after all required publication
stages finish. Later coordinator operations observe those stages. A repeated
commit request observes the same Applied outcome. The generic API need not expose
an intermediate trace-write/queue-insert gap as a completed publication.

**R20. Result/error ordering.** An accepted TLC checker violation retains its original
error information and predecessor/successor states. Completion cannot race past
an accepted violation. Choose the terminal decision once under current authority;
record it through Tube. Do not select "success" just because a queue-empty report
arrived before a still-outstanding worker error.

An illustrative history is:

1. Claim operation A inserts fingerprint F; its reply is lost.
2. Different operation B claims F and receives "already present."
3. Retried A receives its original "new" bit.
4. A publishes F's state and becomes Applied once.

This is valid: A owns the pending publication throughout. Replacing step 3 with
"already present" and dropping A's state violates R8, R11, and R12.

## Completion, recovery, and bounded resources

**R21. Dynamic completion.** Complete only when the frontier is empty and no assigned
batch, accepted result, unresolved claim, or publication obligation can add work.
Check under coordinator ownership that excludes racing admissions. Outstanding
control authority/recovery transitions prohibit success. Worker-local idleness
is insufficient. Safe retirement acknowledgments may drain during shutdown.

**R22. Coordinated snapshot.** Stop admission and reach a consistent cut across all
obligations before snapshot preparation. Commit queue, trace, intern table,
fingerprint resources, and metadata as one named generation. Publish its manifest
reference through Tube only after every required component commit is confirmed.
Keep the previous completed generation available. A partial snapshot is never
an eligible recovery point.

**R23. Recovery.** After coordinator/storage-owner failure, restore all participants
from the same completed snapshot under new authority/generation and clear stale
attempts, caches, and replay ledgers. Refuse mixed-generation recovery. Without
a complete checkpoint, report failure and require an explicit initial-state
restart. Electing a replacement coordinator does not restore lost application
payloads or provide uninterrupted execution.

**R24. Bounded flow.** Support count and byte credits for assignments, results, claims,
and retained replies. Reserve outcome capacity before resource mutation. Apply
backpressure without dropping ownership or outcomes. Control-plane lease events
and progress reporting must continue while data traffic is blocked. Preserve
original long workload bounds; do not make capacity limits silent state cutoffs.

**R25. Artifact and codec identity.** Verify immutable model bundles and resource
snapshots before use. TLC supplies fingerprints, interning identity, value/state
codecs, module dependencies, and checker settings. The service verifies declared
identities and transfer integrity. jsync is an optional transfer mechanism; it
cannot establish snapshot consistency by copying changing files.

## Proposed reusable service boundary

These are conceptual operations, not final Go interface declarations:

| Service operation | Client-visible contract |
| --- | --- |
| Create/observe job | Immutable definition identity, current authority, execution generation, recovery policy, and phase. |
| Register role/capability | Identify candidate, worker, and resource capabilities without making all workers election candidates. |
| Assign/reassign batch | Transfer tracked frontier ownership to a fenced attempt. |
| Submit/replay result | Accept one current attempt; replay its acceptance or rejection by identity. |
| Resolve conditional resource operation | Invoke a client's claim adapter once and retain its original outcome. |
| Apply/replay publication | Advance application stages once, retaining obligations until Applied. |
| Acknowledge outcomes | Safely release retained payloads without allowing old operations to execute anew. |
| Prepare/commit snapshot | Establish a consistent cut, collect component acknowledgments, commit the manifest reference. |
| Recover/activate execution | Fence old resources, restore one generation, and publish the new active tuple. |
| Observe/cancel/finish | Report explicit phases, outstanding work, authority, and one terminal decision. |

The reusable service owns identities, authority integration, attempts, replay
ledgers, credits, obligation accounting, barriers, and job-control records.
Adapters own application execution, conditional resource semantics, publication
stages, and snapshot/restore logic. An adapter reports failures precisely and
honors fencing; wrapping an arbitrary non-idempotent side effect does not make
that effect transactional or crash-safe.

TLC retains successor evaluation, safety checks, source cache semantics where
safe, fingerprint partitioning, trace construction, queue storage, diagnostics,
and Java-derived statistics. It supplies adapters for expansion, FPSet claims,
trace/queue publication, and snapshot participants. It must prevent speculative
worker-cache entries from suppressing states whose result was never accepted.

Keep these boundaries implementable with concrete Go structs and functions.
Opaque payloads and adapter callbacks provide reuse without duplicating the TLC
checker or inventing a general distributed database.

## Scope and implementation guidance

RMember/Tube authority, batch replay, and obligation accounting are required from
the first network integration. Use Raft for control decisions and snapshot
references, with rpc25519 circuits/Greenpack for high-volume payloads. The initial
contract provides checkpoint-based process recovery, not durable continuation
of every acknowledged batch. A later durable mode must persist both outcomes
and publication obligations before advertising stronger acknowledgments.

No requirement here adds distributed temporal/liveness checking, deterministic
exploration order, arbitrary cross-shard ACID transactions, or email reporting.
Ordinary user cancellation is an explicit terminal policy, not proof of exhaustive
exploration. External irreversible side effects beyond TLC's managed storage
would require a stronger client adapter contract before claiming recovery safety.

Implement generic lifecycle and resource boundaries first, then connect the
existing TLC production path. Translate existing Java distributed tests after
their features work, preserving assertions and workload bounds. Temporary
fault-injection exercises may establish retry/fencing evidence but are not Java
port credit or authorization to invent persistent tests. Keep long workloads
separate from focused race checks.

Verification must demonstrate that every accepted new fingerprint remains
publishable, duplicate traffic does not duplicate logical effects, old authority
cannot contaminate a new execution, and success cannot omit outstanding work.
Use the requirement IDs above to explain implementation decisions and evidence
in `tlc/PORT_PROGRESS.md`.

## Reference behavior

- [Java coordinator batch merge and worker loss](../tlaplus/tlatools/org.lamport.tlatools/src/tlc2/tool/distributed/TLCServerThread.java).
- [Java worker expansion and membership queries](../tlaplus/tlatools/org.lamport.tlatools/src/tlc2/tool/distributed/TLCWorker.java).
- [Java conditional fingerprint operations](../tlaplus/tlatools/org.lamport.tlatools/src/tlc2/tool/distributed/fp/FPSetRMI.java).
- [Java queue completion accounting](../tlaplus/tlatools/org.lamport.tlatools/src/tlc2/tool/queue/StateQueue.java).
- [Java distributed checkpoint orchestration](../tlaplus/tlatools/org.lamport.tlatools/src/tlc2/tool/distributed/TLCServer.java).
- [Go distributed implementation](tlc/distributed.go) and [fingerprint manager](tlc/distributed_fp_manager.go).
- [Tube RMember example](/mnt/oldrog/home/jaten/rpc25519/tube/cmd/member/member.go) and [membership implementation](/mnt/oldrog/home/jaten/rpc25519/tube/czar.go).
