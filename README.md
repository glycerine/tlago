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


## License

Copyright © 2026 Jason E. Aten, Ph.D.

[MIT license](LICENSE)

