# Original distributed TLC: source-to-Go map

Updated October 9, 2026. Java reference:
`../tlaplus` at `8f4bc8b73ad1202774a6bf70143436f8ba50aab0`.
Paths in the Java column are relative to
`tlatools/org.lamport.tlatools/src/tlc2/tool/distributed/` in that checkout.

This maps all 35 Java files in that directory tree to their Go implementation.
It is a navigation and scope inventory, not a claim of complete behavioral
parity. A mapped class can still contain a port defect. Verification receipts
are in [PORT_PROGRESS.md](PORT_PROGRESS.md); original-method completion remains
in [TODO_TEST_PORT.md](TODO_TEST_PORT.md).

The target is the original coordinator/worker/fingerprint algorithm. Go uses
TCP, `net/rpc`, goroutines and explicit ownership. Source names containing RMI
identify Java reference files only. Java transport compatibility, JVM services,
reflection and Java object serialization are not implementation targets. The
rpc25519/Tube alternative and email reporting are outside this work.

## Application and coordinator

| Java source | Go implementation | Behavior to preserve |
| --- | --- | --- |
| `DistApp.java` | [distributed_app.go](distributed_app.go), [trace.go](trace.go) | Application settings, initial/successor enumeration, state/action constraints, checks, state reconstruction and call-stack replay. |
| `TLCApp.java` | [distributed_app.go](distributed_app.go) | CLI/application construction, action/property snapshots, complete-state/deadlock checks and their evaluation order. |
| `TLCServer.java` | [distributed.go](distributed.go), [distributed_server_startup.go](distributed_server_startup.go), [distributed_server_init.go](distributed_server_init.go), [distributed_server_completion.go](distributed_server_completion.go), [distributed_server_command.go](distributed_server_command.go) | Construction, recovery before publication, initial states, registration, progress/checkpoint loop, joins, final results, close and shutdown hook. |
| `DistributedFPSetTLCServer.java` | [distributed_server_startup.go](distributed_server_startup.go) | Dynamic manager construction, expected-registration latch and registration rejection. |
| `TLCServerThread.java` | [distributed.go](distributed.go) | Select work, invoke worker, reduce recoverable oversized blocks, insert fingerprints before trace/queue publication, one-time worker-loss cleanup and final statistics. |
| `RMIFilenameToStreamResolver.java` | [distributed_files.go](distributed_files.go) | Private worker files, basename cache, refetch after deletion and failure/output boundaries. |
| `TLCServerRMI.java` | [distributed_server_endpoint.go](distributed_server_endpoint.go), [distributed_server_rpc.go](distributed_server_rpc.go) | Coordinator operation contract over local and TCP endpoints. |
| `InternRMI.java` | [intern_table.go](intern_table.go), [distributed_server_endpoint.go](distributed_server_endpoint.go), [distributed_server_rpc.go](distributed_server_rpc.go) | Coordinator-assigned intern tokens and location metadata, installed before worker parsing. |

## Worker

| Java source | Go implementation | Behavior to preserve |
| --- | --- | --- |
| `TLCWorker.java` | [distributed.go](distributed.go), [distributed_worker_command.go](distributed_worker_command.go), [distributed_worker_group.go](distributed_worker_group.go), [distributed_worker_runtime.go](distributed_worker_runtime.go) | Bootstrap, serialized per-worker computation, cache/holder ordering, partitioned lookup, checks before constraints, predecessor UID, counters and exit. |
| `TLCWorkerRMI.java` | [distributed_worker_endpoint.go](distributed_worker_endpoint.go), [distributed_worker_rpc.go](distributed_worker_rpc.go) | Next-state, liveness, exit, address and cache-statistic calls. |
| `TLCWorkerSmartProxy.java` | [distributed.go](distributed.go) | Invocation timing, computation-time handling and network-overhead observation. |
| `TLCTimerTask.java` | [distributed_worker_keepalive.go](distributed_worker_keepalive.go), [distributed_worker_runtime.go](distributed_worker_runtime.go) | Computing/recent-activity suppression, coordinator relookup, completion/loss reporting and worker exit. |
| `NextStateResult.java` | [distributed.go](distributed.go), [distributed_result_payload.go](distributed_result_payload.go) | Partition arrays, computed-state delta and timing; native transfer preserves graph identities and nil/empty distinctions. |

## Fingerprint servers and managers

| Java source | Go implementation | Behavior to preserve |
| --- | --- | --- |
| `fp/IFPSetManager.java` | [distributed_fp_manager.go](distributed_fp_manager.go) | Routing, scalar/batch calls, statistics, checks, registration and checkpoint/recovery/close operations. |
| `fp/FPSetManager.java` | [distributed_fp_manager.go](distributed_fp_manager.go), [distributed_manager_payload.go](distributed_manager_payload.go) | Registration-wrapper identity, forward reassignment, partition-indexed results, lifecycle traversal and slot-based statistics. |
| `fp/DynamicFPSetManager.java` | [distributed_fp_manager.go](distributed_fp_manager.go) | Expected-count validation, low-bit mask and bounded registration. |
| `fp/NonDistributedFPSetManager.java` | [distributed_fp_manager.go](distributed_fp_manager.go) | Single local store, executor bypass, trace recovery and local I/O fallbacks. |
| `fp/FPSetRMI.java` | [distributed_fingerprint_endpoint.go](distributed_fingerprint_endpoint.go), [distributed_rpc.go](distributed_rpc.go) | Fingerprint operation contract, including named/unnamed checkpoints and coordinator-local trace recovery. |
| `fp/FPSetManagerException.java` | [java_exceptions.go](java_exceptions.go), [distributed_failure_payload.go](distributed_failure_payload.go) | TLC registration-rejection message and checked-I/O category; no RMI exception hierarchy. |
| `fp/DistributedFPSet.java` | [distributed_fp_server.go](distributed_fp_server.go), [distributed_fp_network.go](distributed_fp_network.go) | Discovery retry, source storage factory, registration, reporting, shutdown and output/failure boundaries. |
| `fp/TLCWorkerAndFPSet.java` | [distributed_combined_command.go](distributed_combined_command.go), [distributed_worker_network.go](distributed_worker_network.go), [cli_distributed.go](../cli_distributed.go) | Start FP role before worker role, independent command lifetimes and shared native host ownership. |

## Batch and check callables

| Java source | Go implementation | Behavior to preserve |
| --- | --- | --- |
| `fp/callable/FPSetManagerCallable.java` | [distributed_fp_manager.go](distributed_fp_manager.go) | Warn, reassign and retry the selected partition; no-server result marks all input fingerprints new. |
| `fp/callable/PutBlockCallable.java` | [distributed_fp_manager.go](distributed_fp_manager.go) | Insert the selected partition and retain its index across reassignment. |
| `fp/callable/ContainsBlockCallable.java` | [distributed_fp_manager.go](distributed_fp_manager.go) | Lookup the selected partition and retain its index across reassignment. |
| `fp/callable/BitVectorWrapper.java` | [distributed_fp_manager.go](distributed_fp_manager.go) | Native `distributedBitVectorResult` associates each answer with its input partition. |
| `fp/callable/CheckFPsCallable.java` | [distributed_fp_manager.go](distributed_fp_manager.go) | Checked-I/O reporting/fallback before task-failure handling; signed minimum collection. |
| `fp/callable/CheckInvariantCallable.java` | [distributed_fp_manager.go](distributed_fp_manager.go) | Checked-I/O reporting/false fallback and the manager's early false result. |

Concurrent batch submission captures the partition count and submits one job
per partition. Completion order cannot reorder the returned vector array.
Task failures retain the source logged failure and absent result slot. Executor
shutdown/rejection remains distinct from memory exhaustion at the worker.
Native transport never redials or replays an ambiguously completed mutation;
the original manager owns reassignment. This is not the postponed idempotent
batch protocol.

## Selection and management

| Java source | Go implementation | Behavior to preserve |
| --- | --- | --- |
| `selector/IBlockSelector.java` | [distributed_selector_factory.go](distributed_selector_factory.go) | Block selection, transfer limit and average block count. |
| `selector/BlockSelector.java` | [distributed.go](distributed.go) | Base/proportional policy, queue bounds and source arithmetic. |
| `selector/LimitingBlockSelector.java` | [distributed.go](distributed.go) | Base policy capped by maximum transfer size. |
| `selector/StaticBlockSelector.java` | [distributed.go](distributed.go) | Configured fixed block size. |
| `selector/StatisticalBlockSelector.java` | [distributed.go](distributed.go) | Statistical size calculation and retained measurements. |
| `selector/BlockSelectorFactory.java` | [distributed.go](distributed.go), [distributed_selector_factory.go](distributed_selector_factory.go) | Property selection and fallback boundaries; linked Go constructors supply custom policies. |
| `management/TLCStatisticsMXBean.java` | [management.go](management.go) | Statistics and stop/suspend/resume/checkpoint operation contract. |
| `management/TLCServerMXWrapper.java` | [management.go](management.go) | Running/stopped sentinels, queue/trace observations and synchronized controls. |

## Supporting ownership and transfer

These source classes depend on TLC facilities outside the distributed directory.
They remain part of the port's behavior; the directory inventory does not prove
their completeness.

- Coordinator work uses `DiskStateQueue`; supporting memory/deque/byte queues
  retain their own source contracts. See [queue.go](queue.go),
  [disk_state_queue.go](disk_state_queue.go) and
  [disk_byte_array_queue.go](disk_byte_array_queue.go).
- Trace storage/reconstruction and intern checkpoints retain their source
  ordering. See [trace.go](trace.go) and [intern_table.go](intern_table.go).
- Local and remote hosts use the existing TLC storage factory and fingerprint
  implementations. See [fpset.go](fpset.go), [fpset_disk.go](fpset_disk.go) and
  [offheap_fpset.go](offheap_fpset.go).
- Native state/result/manager/failure payloads preserve the supported TLC graph
  contracts. See [distributed_state_payload.go](distributed_state_payload.go),
  [distributed_result_payload.go](distributed_result_payload.go),
  [distributed_manager_payload.go](distributed_manager_payload.go) and
  [distributed_failure_payload.go](distributed_failure_payload.go).
  Opaque custom data remains explicitly rejected until source transferability
  and a native contract are established; arbitrary Java object transfer is not
  a requirement.
- Native publication/discovery replaces Java naming infrastructure. See
  [distributed_coordinator_network.go](distributed_coordinator_network.go),
  [distributed_registry.go](distributed_registry.go) and
  [distributed_network_discovery.go](distributed_network_discovery.go).
  Publication, discovery and endpoint lifetime are separate operations.

## Evidence and open work

The current original distributed inventory is 41 method contexts: 37 complete
and four Reconcile. The four staged model bodies preserve their original
assertions and Ant profile behind `tlago_disabled_distributed_tests`. On the
recorded 48-worker run, local EWD840 fails at final off-heap fingerprint checking;
the source selector also retains the closed flusher in that configuration.
The shared Java harness disables these tests. Do not weaken the assertions,
silently count disabled methods as passing or alter the pinned algorithm to
hide that source limitation.

Focused evidence includes original manager and smart-proxy tests, native TCP
operations, full EWD840 model runs, worker/FP host loss, complete/partial insertion
reply loss, controlled transport stalls and registered-endpoint checkpoint
restarts, including completed recovery followed by lost acknowledgement and
source size-query failover. These receipts have different scopes; none establishes
complete distributed parity alone. Use the detailed entries in the progress log
rather than rerunning unchanged long workloads.

The exact trace-to-intern caller boundary is verified by external GDB/disassembly
in both a local storage fixture and a full N=7 model with two nested LSB hosts.
Fresh-process recovery preserves their distinct local/remote fingerprint commit
generations. The remote model finishes at 114,942 states with an empty queue.
Neither receipt establishes atomic checkpointing.

The remaining assignments are maintained in
[HANDOFF.md](HANDOFF.md#remaining-distributed-work):

1. A corrupt coordinator queue now stops real-model checkpoint startup before
   remote fingerprint recovery/publication, with joined roles and unchanged
   retained checkpoint bytes. A truncated trace plus missing queue additionally
   verifies trace-before-queue failure ordering. Other phases remain unproved. Fresh CLI
   remote recovery before registration has the source's empty-manager limitation.
2. Additional full-model fingerprint failure phases and general partitions
   remain unproved beyond the recorded loss and controlled-stall cases.
3. Continue ownership/cleanup comparison where concrete source evidence reveals
   a shortcut. Consult prior receipts before repeating completed audits.
4. Core metadata producers and transferability are audited; see `TLC_ARCH.md`.
   New custom-data support needs a concrete producer and native contract.
5. Resolve the four original model-test dispositions without manufacturing a
   green gate or completion credit.

No missing distributed Java file was identified by this mapping. Method bodies,
failure boundaries and the open assignments still require evidence before the
overall goal can be marked complete.
