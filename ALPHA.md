# Alpha testing

Use a disposable, writable copy of your model. Start with the included counter,
then try a small model whose result you already know. Compare results with your
usual TLC setup when available: invariants, state counts, deadlocks, liveness,
and the counterexample itself are more useful than matching elapsed times.

## Suggested checks

1. Run the included `Counter.cfg`: successful checking, four distinct states.
2. Run `CounterViolation.cfg`: invariant violation and a counterexample.
3. Parse and check your own model with `tlago check Spec.tla`.
4. Check it with the Java TLC flags you normally use, including its `.cfg` file.
5. If relevant, try simulation, trace export/import, or checkpoint/recovery on
   a small model. Use `tlago help modelcheck` for supported arguments.

The ten-second limit applies to repository test executions. Keep alpha smoke
checks small as well; do not use the disabled stress tests as an alpha checklist.

## Current limitations

- This is an alpha port. The source inventory and broad historical test passes
  do not establish complete Java TLC parity. Some original Java tests have
  failing expectations or disabled setup, and remain explicitly reconciled
  rather than credited as passing.
- Large fingerprint stress workloads and other tests over ten seconds are
  skipped. They are not part of routine validation or an alpha acceptance claim.
- Distributed execution uses native Go boundaries. Java RMI interoperability is
  not provided. Use the same tlago build for coordinator, workers, and storage
  servers. General network-partition behavior remains incompletely verified.
- Recovery follows the translated TLC phases; it is not an atomic transaction
  across all checkpoint files and distributed stores. Keep an independent copy
  of important model inputs and checkpoints.
- JVM-specific instrumentation and compiler/GC assertions do not have native
  Go equivalents. Some compiler-dependent enhanced null messages differ.
- SANY XML and Apalache JSON corpus parity remain incomplete. Those long oracle
  sweeps are disabled under the current testing policy.
- Email reporting is excluded. Console and file output remain available.

## Feedback

Report issues at <https://github.com/glycerine/tlago/issues>. Include:

- The command, operating system/architecture, and relevant environment settings.
- A small `.tla` / `.cfg` reproducer and any imported modules needed to run it.
- Expected and observed results, the exit status, and complete console output.
- Whether Java TLC gives the expected result with the same model and flags.

Remove private model content and credentials before attaching files. For
nondeterministic results, include the worker count and reported random seed.
The command's runtime help is the authoritative user-facing flag reference.

## Preparing a bundle

Run `make test`, `make build`, and the two small counter invocations from the
README. `make dist` produces a native archive and its `.sha256` file in `dist/`.
Unpack it into a fresh directory and repeat the counter checks there, away from
the source checkout, to confirm the included module resources resolve.

On Linux, verify an archive with `sha256sum -c FILE.sha256` from `dist/`.
On macOS, use `shasum -a 256 -c FILE.sha256`. Keep the accompanying notices in
any redistributed bundle. Cross-building proves compilation only; test the
archive on its target system before sharing it.
