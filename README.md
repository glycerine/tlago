# tlago

TLA+ parsing, semantic checking, and the TLC model checker, ported to Go.

Core model checking, simulation,
liveness checking, traces, checkpoint/recovery, and distributed roles are
implemented. Full Java TLC parity is still being reconciled. See [ALPHA.md](ALPHA.md)
for the current scope, known limitations, and feedback checklist.

## Try the included model

Unpack the supplied archive for your operating system and architecture, then run
these commands from the unpacked directory:

```sh
./tlago -help
./tlago -workers 2 -config examples/counter/Counter.cfg examples/counter/Counter.tla
```

The counter cycles through four states and satisfies `TypeOK`. The checker
should report successful completion. On Windows, use `tlago.exe` in place of
`./tlago`.

The second configuration deliberately violates an invariant and produces a
counterexample:

```sh
./tlago -workers 1 -config examples/counter/CounterViolation.cfg examples/counter/Counter.tla
```

A nonzero exit is expected for that invocation. TLC writes metadata and may
write trace-exploration files near the model, so use a writable working copy.
The binary includes the standard and bundled CommunityModules definitions;
ordinary checking does not require Java or the repository checkout.

## Build from source

Use Go 1.26 or newer. From the repository root:

```sh
make build
./bin/tlago -workers 2 -config examples/counter/Counter.cfg examples/counter/Counter.tla
```

Without `make`, use `go build -p 8 -o bin/tlago ./cmd/tlago` (on Windows, choose
`bin/tlago.exe` as the output). Builds need the Go
module dependencies on the first run. Python 3 is required only for the test
runner and archive packaging, not for the executable.

## Commands

TLC is the default command and accepts Java TLC command-line flags:

```sh
tlago -workers auto -config MC.cfg MC.tla
tlago modelcheck -workers auto -config MC.cfg MC.tla
tlago mc -workers auto -config MC.cfg MC.tla
```

All three use the same TLC runner. Do not use a `--tlc` selection flag.

| Command | Purpose |
| --- | --- |
| `parse FILE...` | Parse TLA+ modules with the Go SANY parser. |
| `check FILE...` | Parse and perform semantic checking. |
| `modelcheck` / `mc` | Run TLC; also the default without a subcommand. |
| `repl-expr 'EXPR'` | Evaluate a quoted expression; use `-spec FILE` for context. |
| `checkimplfile` | Monitor implementation trace files. |
| `server`, `worker`, `fpserver`, `worker-fpserver` | Run distributed TLC roles. |
| `sany-xml` | Export the checked SANY graph as XML. |
| `apalache-json` | Export JSON IR; corpus parity remains incomplete. |

Run `tlago -help` for the maintained flag reference, defaults, examples, and
Java/Toolbox mappings. Use `tlago help modelcheck` or `tlago help server` for
command-specific help. `-h`, `--help`, and `help [COMMAND]` are accepted.
Use `-I DIR` to add a module search directory.

Model-checking exit statuses follow the TLC runner: success is 0; violations,
invalid arguments, and execution failures return nonzero statuses. Parsing and
export commands have their own status mappings. Scripts should inspect both
the exit status and diagnostics.

## Development and alpha bundles

```sh
make test      # Guarded default TLC suite; under two minutes.
make vet       # Static checks, including test files.
make dist      # Native archive plus SHA-256 checksum in dist/.
```

The test runner enforces a 9.5-second safety cutoff for an active leaf test and
118 seconds for the suite. Tests known to exceed ten seconds retain their bodies
with `t.Skip("test is too slow")`. `make long` is disabled. Do not remove these
skips or launch long/race workloads to prepare an alpha.

Linux Make targets share CPUs `0-7` through `taskset`; Go work uses
`GOMAXPROCS=8`. On systems without `taskset`, the targets omit CPU affinity.
`RUN` can be set explicitly for another execution wrapper. `make test` needs
local loopback sockets for the distributed tests; sandbox-denied socket failures
are not passing checks. The JSON log is saved to `.cache/test-results.jsonl`.

`make vet` includes test files. It disables `stdmethods` because the
ported Java `ReadByte`, `WriteByte`, and `Seek` APIs deliberately have different
signatures from Go's similarly named interfaces. Plain `go vet ./...` still
reports those Java method signatures.

To cross-build an archive, for example:

```sh
python3 scripts/package.py --goos darwin --goarch arm64
```

Cross-built archives require testing on their target platform before sharing.
Packaging includes the executable, this guide, the alpha notes, the example,
and third-party notices. It does not publish, tag, or upload anything.

See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for bundled attributions,
[TLC_ARCH.md](https://github.com/glycerine/tlago/blob/master/tlc/TLC_ARCH.md)
for implementation contracts, and
[TODO_TEST_PORT.md](https://github.com/glycerine/tlago/blob/master/tlc/TODO_TEST_PORT.md)
for the source-test inventory. Detailed development history is in
[PORT_PROGRESS.md](https://github.com/glycerine/tlago/blob/master/tlc/PORT_PROGRESS.md).

## License

Copyright © 2026 Jason E. Aten, Ph.D., and the upstream contributors identified
in the sources. Distributed under the [MIT license](LICENSE), with additional
notices for incorporated code and modules.
