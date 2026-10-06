# Distributed TLC batch protocol

Status: implementation plan, October 5, 2026. Development checkout:
`/mnt/b/github.com/tlaplus/tlago`, branch `distrib`.

The client requirements and reusable service boundary are defined in
[TLC_DIST_REQS.md](TLC_DIST_REQS.md). This document specifies the concrete batch
protocol that implements that contract.

## Purpose and fixed decisions

Build the rpc25519 integration with retry-safe batch operations from the start.
Use `github.com/glycerine/rpc25519` peers, circuits, and fragments, with
`github.com/glycerine/greenpack` payloads. Keep the existing concrete TLC
coordinator, workers, fingerprint manager, queue, and trace implementations.
Do not recreate Java RMI or introduce a second checker.

Preserve Java's state evaluation, fingerprint calculation, partition mask,
bit-vector meanings, successor validation order, constraints, and trace behavior.
The deliberate network-protocol improvement is that repeating a named insertion
returns its original answer, instead of performing another insertion and
answering that its fingerprints are now old.

Initial scope includes a Tube Raft reliable-membership control plane from the
start, one elected TLC coordinator, remote compute workers, and optionally remote
fingerprint servers. Surviving-process reconnects and lost replies are supported.
Process failure after a fingerprint mutation requires coordinated rollback to a
completed checkpoint. The three-machine implementation target requires an
initial recovery generation and replicated snapshot files before dispatch, so
loss of one machine does not force manual restart. Coordinator election and
replacement are provided by Tube from the first implementation; continuing an
uncheckpointed epoch after coordinator replacement and live fingerprint-shard
replacement are deferred. This differs from
Java's permissive fingerprint reassignment and must be documented as a conscious
correctness policy, not silently treated as mechanical parity.

Use Tube for reliable membership, coordinator election, authority epochs, job
control records, and committed checkpoint manifests from the start. Keep state
and fingerprint batches on rpc25519 circuits. Use jsync for stable file transfer. Distributed liveness checking is outside Java
parity; the source distributed checker does not support it. Email is excluded.

## Why an insertion needs an operation identity

Java's `FPSet.putBlock` inserts fingerprints and returns bits identifying those
newly inserted by that call. The coordinator uses those bits to write trace
records and enqueue successor states. Workers query membership but do not publish
fingerprints themselves.

If insertion succeeds but its reply is lost, a fresh insertion call returns
"already present." Using that reply can discard states never enqueued. Idempotent
set contents are insufficient: the mutation and its original reply belong to one
logical operation.

The invariant for this protocol is:

> Every accepted new fingerprint has either a coordinator-held obligation to
> publish its corresponding state, or a published trace record and queue entry.
> No successful completion or checkpoint may omit that obligation.

This is a protocol guarantee within the failure scope above, not a claim that
three existing storage components already form a crash-atomic transaction.

Source references:

- [Java coordinator merge](../tlaplus/tlatools/org.lamport.tlatools/src/tlc2/tool/distributed/TLCServerThread.java).
- [Java worker computation](../tlaplus/tlatools/org.lamport.tlatools/src/tlc2/tool/distributed/TLCWorker.java).
- [Java fingerprint operations](../tlaplus/tlatools/org.lamport.tlatools/src/tlc2/tool/distributed/fp/FPSetRMI.java).
- [Java reassignment](../tlaplus/tlatools/org.lamport.tlatools/src/tlc2/tool/distributed/fp/callable/FPSetManagerCallable.java).
- [Current Go merge](tlc/distributed.go), `computeBlockAttempt` and `publishBlock`.
- [Current Go fingerprint manager](tlc/distributed_fp_manager.go).
- [Tube membership example](/mnt/oldrog/home/jaten/rpc25519/tube/cmd/member/member.go).
- [Tube RMember, Czar, and RMVersionTuple](/mnt/oldrog/home/jaten/rpc25519/tube/czar.go).
- [Tube CAS and session APIs](/mnt/oldrog/home/jaten/rpc25519/tube/actions.go).

## Tube authority and reliable membership

Use Tube's existing `RMember` service; do not implement another election algorithm.
The member example constructs `tube.NewRMember(tableSpace, cfg)`, calls `Start`,
waits for `Ready.Chan`, and continuously services `UpcallMembershipChangeCh` and
`OperatingLeaseRenewCh`. Both channels have finite buffering and must be drained
by the control-plane event loop independently of batch computation.

Tube's OperatingLeaseRenewCh carries a time without its epoch/owner. The monitor
uses that event to request an authenticated tagged refresh, not to extend authority
from the timestamp alone. Refresh runs separately from event consumption. Candidate
admission uses the earlier of the Czar deadline and its own operating deadline
from the authoritative membership reply, with the configured drift margin.

The service elects a Czar through a lease on its tablespace's `czar` key. Its
`RMVersionTuple.CzarLeaseEpoch` comes from Tube's consensus lease epoch;
`WithinCzarVersion` orders membership changes within that epoch. Use the elected
Czar as the TLC coordinator. This is an application coordinator, distinct from
the Tube Raft cluster leader. A Raft leader change alone must not restart TLC if
the existing Czar's authority remains valid.

### Deployment and membership groups

Configure a persistent Tube cluster, initially three voting replicas, with its
addresses and credentials supplied to TLC nodes. Use a separate RMember tablespace
`tlc/<JobID>/coordinators` for each job. Only configured coordinator candidates
join that group using RMember; every candidate must be able to recover the queue,
trace, intern table, and job metadata. Keep at least two candidates available.
Do not assume a replica of the Tube cluster automatically becomes a TLC candidate.

Compute-only workers and fingerprint-only servers do not join this election group:
RMember allows its members to contend for Czar, so joining them would accidentally
make them coordinator candidates. They observe coordinator authority through the
TLC control adapter, using the service's authoritative Czar Ping response and
Tube's linearizable lease/control records. Their existing registration and
activity tracking supplies data-role membership. Nodes may combine these roles
when explicitly configured as coordinator candidates.

This avoids a second leased coordinator key or a second election. Discover the
Czar via Tube, then contact it for current reliable membership. The source warns
that the membership list and version serialized under `czar` are potentially
stale; they are discovery hints, not an authoritative current membership view.

### Epochs, leases, and admission

The control adapter tracks Czar identity, its consensus lease epoch, current
membership version, lease deadline, and coordinator process incarnation. It
accepts only monotonic membership versions and rejects expired authority. The
`Ready` signal means membership channels are usable; it is not permission to
start model checking. The candidate must also hold Czar authority and finish
job activation or checkpoint recovery.

Use `CoordinatorEpoch = CzarLeaseEpoch`, preserving Tube's positive signed 64-bit
range. Combine it with a Raft-committed `RunGeneration` to identify an execution
as `(JobID, CoordinatorEpoch, RunGeneration)`. Begin RunGeneration at 1 and
increment it through Tube CAS whenever checkpoint rollback creates another
execution under the same coordinator lease. A new Czar epoch obtains a new
activation record; never generate authority epochs randomly or derive them from
addresses, wall clocks, Raft term alone, or rpc25519 circuit IDs.

Retain random PeerIncarnation identifiers for process identity only. A restarted
candidate cannot inherit a predecessor's active execution, even if it recovers
under the same Czar lease epoch: it must activate a new RunGeneration.

All data peers verify the coordinator sender identity, execution tuple, and
unexpired authority before admitting commands. Install authority from the control
adapter's validated membership/lease information, never from an ordinary batch
message claiming a newer epoch. Returning cached replies, retirement messages,
checkpoint control, and success publication are fenced as well as insertions.

Follow Tube's operating-lease and clock-drift contract. Configure a documented
ClockDriftBound (initial deployment default 500 ms, as in the example), monitor
clock synchronization, and reject deployments that cannot satisfy it. The TLC
adapter stops admission conservatively before the observed Czar deadline by that
bound; use earlier member operating deadlines where applicable. It must recheck
its lease gate after pauses and before side effects, not just once per circuit.
Loss of quorum does not instantly invalidate a still-valid lease, but no node
may continue after its conservative deadline without a confirmed renewal.

Lease expiration closes command admission, cancels assignment, and suppresses
completion publication. Already admitted bounded mutations may finish only into
that old execution's ledger/storage. Before a new execution starts, require all
required storage owners to fence the old tuple and acknowledge that old mutations
have drained. A missing owner must be replaced with isolated storage restored
from the selected checkpoint, not reused while an old owner could still write.
This transition barrier supplies storage fencing; election does not physically
stop a paused old process.

### Raft job records and coordinator replacement

Store a versioned JobControl record in Tube containing JobID, specification hash,
Czar identity and epoch, RunGeneration, coordinator incarnation, phase, topology,
and selected committed checkpoint. Use Tube CAS with session duplicate handling
for updates, and reconcile ambiguous replies by a linearizable read. Records must
advance monotonically; peers validate that their Czar epoch matches current
lease authority. A stale coordinator's record is not valid merely because it
exists in the key/value store. Guard updates with the expected record version and
revalidate lease ownership; checkpoint references are tied to their execution
identity, so stale writers cannot overwrite newer activation records.

Phases are `Preparing`, `Active`, `Recovering`, `Completed`, and `Failed`.
The elected candidate publishes Preparing/Recovering, establishes fencing and
restores a complete checkpoint when needed, then commits Active. Data peers open
admission only for that Active tuple. Commit Completed only after the batch
completion conditions and current authority checks both pass.

Tube elects a replacement automatically, but the replacement does not adopt the
old coordinator's in-memory batch obligations. It fences the old execution,
selects the latest committed checkpoint manifest from Tube, restores every
component, and activates a new tuple. Commit a replicated initial-state generation before opening assignment so
replacement can recover even before the first periodic checkpoint. Missing all
valid snapshots is a failed deployment precondition, not a successful recovery. Storage-loss recovery under an unchanged coordinator follows the
same procedure with an incremented RunGeneration.

## Identities and wire messages

Use fixed-width unsigned integer fields and explicit byte arrays in generated
Greenpack structs. Do not serialize process pointers or reconstruct identity from
rpc25519 circuit IDs, fragment serials, addresses, or clocks.

| Field | Decision |
| --- | --- |
| `ProtocolVersion` | Start at 1; reject incompatible versions at handshake. |
| `JobID` | Cryptographically random 128-bit identifier created for a new model-checking job; persists in its checkpoint manifest. |
| `CoordinatorEpoch` | Positive signed 64-bit Tube CzarLeaseEpoch; identifies the elected coordinator authority. |
| `RunGeneration` | Positive unsigned 64-bit generation committed by Tube CAS for an execution under that authority; changes on recovery/restart. |
| `MembershipVersion` | Tube WithinCzarVersion for control-plane view ordering; ordinary membership updates do not reset batch identities. |
| `PeerIncarnation` | Random 128-bit identifier per worker/fingerprint-server process start. A changed identity is process failure, not reconnection. |
| `TopologyEpoch` | Unsigned 64-bit topology revision; initially 1, immutable during a running epoch. |
| `BatchID` | Unsigned 64-bit coordinator sequence identifying a dequeued predecessor block; never reused in an epoch. |
| `Attempt` | Unsigned 64-bit assignment generation for that batch. Increment when computation moves to another worker. |
| `StreamID` | Coordinator-issued unsigned 64-bit identity for an insertion stream bound to one fingerprint shard. |
| `Sequence` | Unsigned 64-bit insertion sequence per stream, beginning at 1. |
| `Digest` | BLAKE3-256 of the canonical logical request, excluding transport metadata. |

An insertion operation is identified by
`(JobID, CoordinatorEpoch, RunGeneration, TopologyEpoch, ShardID, StreamID, Sequence)`.
Its canonical digest covers that tuple, BatchID, Attempt, vector length, and every
64-bit fingerprint in order, using an explicitly specified big-endian encoding.
Do not hash incidental Greenpack map ordering. Preserve the original vector
indices and signed fingerprint bit patterns. Results include the request digest,
bit count, and the original newly-inserted bitmap; reject malformed padding,
lengths, and indices before coordinator mutation.

Define explicit message kinds: `Hello`, `Bootstrap`, `RegisterWorker`,
`RegisterFPServer`, `ComputeBatch`, `ComputeResult`, `ComputeError`,
`ContainsBatch`, `ContainsResult`, `InsertBatch`, `InsertResult`,
`AppliedThrough`, `RetiredThrough`, `Status`, `Heartbeat`, `Stop`, and
checkpoint prepare/commit/abort messages. Their envelopes carry job and epoch
identities and coordinator sender identity. Transport errors and TLC checker
errors remain distinct. MembershipVersion orders control updates; require the
current execution and topology for batches without invalidating requests merely
because another candidate joined the membership group.

Bootstrap supplies model/configuration identity, fingerprint polynomial,
partition mapping, checker settings, and string-interning service identity.
Install the coordinator's interning source before parsing on a worker, matching
Java. A reconnect repeats the same identities and resumes existing operations;
it cannot create a fresh job or adopt a different fingerprint database.

## Fingerprint-server execution and result retention

Create one insertion stream per shard initially. Different shards run in parallel.
A stream processes requests in increasing sequence order, so writes to the same
shard are serialized initially. Preserve concurrency with membership reads using
the FPSet's existing synchronization. Additional insertion streams are a later
performance change requiring the same per-operation guarantees.

Each stream stores `NextExpected`, `RetiredThrough`, and records for completed,
unretired requests. Admit sequences only within a bounded negotiated window.
Buffer out-of-order arrivals inside that window; reject requests beyond it with
an explicit backpressure response. The initial replay implementation reserves
a fixed byte slot per record: request bytes plus the declared maximum reply must
fit that slot. This leaves capacity for missing earlier sequences even when later
requests arrive first. Negotiate slot size before dispatch; split oversized work
without changing the search bounds. Record/callback overhead is count-bounded
separately from encoded payload/reply bytes. Empty shard vectors need no insertion request.

For a received insertion:

1. Validate job, epoch, topology, shard ownership, vector limits, and digest.
2. If its sequence is retired, return `RequestRetired`; never execute it again.
3. If a retained record exists, require its digest to match and return its exact
   bitmap. A mismatching digest is a fatal protocol violation.
4. If the request is already executing, coalesce the duplicate with that execution.
5. When it is the next sequence, reserve result storage before calling `PutBlock`.
   Hold the stream's execution ownership until both the mutation and immutable
   result record are installed. Never expose an intermediate "missing result."
6. Install the complete reply before sending it. A send failure leaves the result
   retained. Duplicate sends do not change membership, the bitmap, or counters.

The ledger is in memory in the first implementation. An exception, partial
insertion, or inability to install its result poisons the shard for this epoch:
stop accepting mutations and fail the run. Do not retry a partially executed
insertion as a fresh operation. The coordinator must roll back all components
or stop; it must not salvage an arbitrary subset of this epoch.

A read-only `ContainsBatch` may be repeated against current membership: its
answer is advisory, and publication is decided by insertion. This assumes every
in-flight insertion still has its publication obligation. Workers retain accepted
compute results rather than recomputing them after a lost result reply.

## Coordinator batch ownership and publication

Introduce a concrete coordinator batch ledger. A dequeued batch remains owned
until all its publication obligations finish. It moves through:

`Assigned -> ResultAccepted -> Inserting -> Publishing -> Applied`.

Before accepting a worker result, retain its aligned successor/fingerprint vectors,
parent UIDs, generation statistics, worker identity, and assignment generation.
Validate the complete shape first. Apply Java statistics once per accepted result;
network retransmissions must not increment them. Counters measuring actual
transport traffic may count retransmissions separately.

For each nonempty shard vector, allocate an insertion sequence and retain the
immutable request. Retry that exact identity and contents until its original reply
arrives, cancellation ends the run, or a peer/process failure requires rollback.
Timeout means unknown outcome; it never means "allocate another request ID."

Retain each received bitmap and publish its set bits through the existing trace
write and queue enqueue sequence. Give each entry a publication state:
`Pending`, `TraceWritten(uid)`, `Enqueued`. One coordinator owner mutates these
states; duplicate replies only locate the existing record. Never run publication
concurrently twice for an entry. Do not change which racing batch wins a
fingerprint merely to make trace ordering deterministic.

Once admitted under a valid lease, normal publication is not cancellable between
trace write and enqueue. Lease loss prevents new publication admission; admitted
publication finishes within old storage before the transition barrier is released. A storage
error is fatal to the epoch; partially published work cannot be called Applied.
Do not retry an enqueue with an ambiguous local outcome. Process crashes discard
this epoch and recover a coordinated checkpoint, rather than attempting live
reconstruction from these in-memory stages.

After all entries are Enqueued (or were already present), mark the insertion
Applied. Advance the contiguous applied watermark for that stream. Send
`AppliedThrough` only after every operation through that watermark is Applied.

The server advances its retirement watermark before deleting result records and
returns `RetiredThrough`. Watermarks survive reconnects within a process
incarnation. Repeated acknowledgments are harmless. The coordinator retains enough
identity/watermark state to recognize late replies after releasing batch payloads.
A `RequestRetired` for an operation the coordinator still considers unresolved is
a protocol failure, not evidence that it was applied.

## Worker attempts, cancellation, and caches

A worker retains the immutable result or checker error for each completed
`(BatchID, Attempt)` until the coordinator acknowledges acceptance. Duplicate
ComputeBatch requests replay that outcome; they do not recompute it or reuse a
new fingerprint-query answer. Bound retained worker results with credits too.

The coordinator accepts only its current assignment generation. Late results from
superseded attempts cannot insert fingerprints or mutate queue/trace/statistics.
Worker loss returns predecessor ownership once, as Java does. Once ResultAccepted,
the coordinator owns publication even if that worker disappears.

Audit worker-local fingerprint caching before enabling compute retries. A cache
entry created by an abandoned attempt must not suppress a successor that no
accepted result contains. Preserve Java's ordinary cache behavior where safe,
but keep attempt-local additions isolated until the coordinator acknowledges
acceptance of that attempt. Retaining a result at the worker alone is insufficient:
the coordinator may supersede the attempt. Discard additions for aborted or
superseded attempts; do not acknowledge acceptance before holding its complete
publication obligation. A worker restart has a new incarnation
and empty cache. Any necessary change to Java's cache lifecycle is a documented
correctness correction, not an unrecorded shortcut.

## Transport, limits, and completion

Use rpc25519 circuits for coordinator/worker and fingerprint traffic; fragments
carry Greenpack envelopes. Read replies asynchronously and correlate them with
application identities. A successful SendOneWay is not an application acceptance
or commit acknowledgment. Circuit replacement does not change operation IDs.

Negotiate byte limits, outstanding batch credits, insertion-window size, and
retained-result byte budgets during bootstrap. Start with one executing compute
batch per worker slot and one insertion window per shard. Keep multiple workers
and shards active concurrently. Reserve receiver capacity before mutations;
never evict unacknowledged insertion results to relieve memory pressure.

Retry reconnects with capped exponential backoff and jitter. Use heartbeats and
activity-aware deadlines; long successor computation is not proof of failure.
A retry budget expiring pauses dependent work or fails the epoch; it cannot turn
an unknown insertion into a successful no-op. Reject wrong epochs explicitly.
Initially abort on fingerprint-server incarnation change or unresolved loss;
do not reuse Java's manager reassignment behind a pending insertion stream.

Global completion requires: queue empty, no assigned computations, no accepted
results awaiting insertion, no unresolved insertion replies, and no unpublished
accepted successors. Perform that check under coordinator ownership so receiving
or assigning work cannot race completion. Unfinished retirement acknowledgments
need not delay exploration completion, but drain them during orderly shutdown.
Commit completion through the Tube JobControl record only under current Czar
authority. A violation/error terminates with the original TLC diagnostic and trace behavior,
not a success result.

## Checkpoint and recovery boundary

Extend existing checkpoint coordination rather than copying live files:

1. Stop assigning new work. Drain computations, accepted results, insertion
   requests, and publication. If any outcome remains ambiguous, abort checkpoint
   preparation; never checkpoint around it.
2. Confirm each insertion stream's contiguous AppliedThrough watermark. Obtain
   peer barrier acknowledgments. No mutation crosses the preparation barrier.
3. Begin existing queue, trace, FPSet, and intern-table checkpoints in an isolated
   generation directory on each owning host. Adapt destination selection so
   committing a generation cannot overwrite the previous completed generation.
   Include JobID, CoordinatorEpoch, RunGeneration, topology, polynomial,
   component identities, and file hashes in a new
   manifest. Keep original component formats where possible.
4. Commit every required component. Publish the completed manifest reference
   through Tube CAS in the current JobControl record only after all component
   commits are confirmed and current coordinator authority is revalidated. Retain the previous completed generation
   until the new generation is safely discoverable. An ambiguous/failed commit
   does not authorize continuing this epoch.
5. Resume assignment only after the manifest is complete. Partial generations
   are not recovery candidates.

On process failure, the Tube-elected coordinator fences the prior execution,
selects the latest fully completed manifest referenced by Tube, and restores
queue, trace, intern table, and all fingerprint shards from that same generation.
Commit a new CoordinatorEpoch/RunGeneration activation, clear old request
ledgers/caches, and bootstrap every peer again. Old fragments are rejected. If any required
component is absent, refuse recovery instead of mixing generations. Commit the initial generation before exploration. Every committed manifest
must reference verified component files on at least two distinct machines;
Tube metadata replication does not replicate those files. Loss of one machine
must leave a usable generation and permit automatic restoration.

Use jsync to transfer immutable model bundles and completed checkpoint files.
Verify hashes before making a transferred generation recoverable. File transfer
is not the checkpoint barrier or commit protocol.

Tube membership, coordinator authority, activation records, and checkpoint
manifest references are mandatory initial integration. Continuing unfinished
work across coordinator replacement without rollback is a later extension: it
requires durable accepted-state obligations, insertion replies, and publication
recovery. Raft election/session deduplication alone does not supply those data
recovery guarantees.

## Implementation sequence and validation

1. Integrate Tube RMember coordinator candidates and its membership/lease event
   loop. Implement Czar-derived authority, JobControl CAS activation, lease gates,
   and fencing transitions. Define Greenpack messages, canonical digests, execution
   identities, and bounds; wire peers only after the control plane can authorize
   their Active execution.
2. Add coordinator batch ownership and worker-result replay with coordinator-local
   fingerprints. Preserve checker errors, statistics, trace parents, and the
   original safety-check ordering. Audit cache behavior at this boundary.
3. Add remote fingerprint insertion streams, result ledgers, acknowledgments,
   credits, and retirement watermarks. Route remote mutations through this
   protocol; leave direct local FPSet APIs available to existing local callers.
4. Add reconnect handling and asynchronous completion accounting. Audit every
   timeout, cancellation, duplicate, and stale-result path before enabling
   automatic retry.
5. Add coordinated checkpoint manifests committed through Tube, and replacement
   coordinator recovery under a new execution tuple. Enable
   jsync transfer only for stable completed generations.
6. Port the existing Java distributed correctness tests after their features
   work. Preserve their assertions and workload bounds, and update
   `tlc/TODO_TEST_PORT.md` only as complete original methods land.

Use temporary fault-injection exercises during implementation to check lost
insertion replies, duplicate requests/replies, mismatching digests, delayed old
attempts, repeated acknowledgments, partial storage failures, peer restarts, and
checkpoint failures, Czar replacement, quorum loss, expired leases, delayed
membership updates, and stale-coordinator writes. These exercises do not count as Java test-port credit and
are supplemented by the newly authorized BDD/test-first service and
three-machine resilience tests. New service tests are not Java port credit.
Run long original workloads normally; use -race only for short focused checks.

Implementation acceptance requires evidence that each accepted new fingerprint
is queued, each accepted result affects semantic counters once, an ambiguous
mutation cannot silently become "already seen," and process recovery restores a
single complete checkpoint generation. Record intentional Java differences and
verification receipts in `tlc/PORT_PROGRESS.md`; keep `tlc/HANDOFF.md` current.
