tlago: TLA+ tools ported to Go
==============================

(work in progress; expect lots of updates; but v0.0.1 has been tag-ed and can be experimented with).

## Command-line use

Run the full command and flag guide, or select one command's help:

```bash
go run ./cmd/tlago -help
go run ./cmd/tlago modelcheck -help
```

Help includes a one-line flag reference, detailed explanations, defaults,
examples, and mappings to Java TLC commands and Toolbox model-editor options.
`--help`, `-h`, and `help [COMMAND]` are also accepted.

Use `modelcheck --tlc` to select the Go port of Java TLC:

```bash
go run ./cmd/tlago modelcheck --tlc -workers auto -config MC.cfg MC.tla
```

This corresponds to `java -cp tla2tools.jar tlc2.TLC -workers auto -config MC.cfg MC`.
The `--tlc` switch is specific to the Go wrapper. Without it, `modelcheck` uses
the earlier bounded checker, whose options and limits are explained in help.

# tlgo -help

~~~
$ tlago -h

tlago: TLA+ parsing, semantic checking, and the Go TLC model checker

Usage: tlago COMMAND [OPTIONS] FILE...
       tlago -help | tlago help [COMMAND]
Help aliases: -help, --help, -h. Help needs no input file and exits successfully.

COMMANDS AND JAVA EQUIVALENTS
  parse                               Parse FILE... with the Go SANY parser; Java counterpart: SANY -s.
  check                               Parse and check FILE...; Java counterpart: tla2sany.SANY, Toolbox Parse Spec.
  modelcheck (mc)                     Check one model. Add --tlc to run the port of Java tlc2.TLC.
  checkimplfile (check-impl-file)     Monitor implementation trace files; Java: tlc2.tool.CheckImplFile.
  repl-expr (repl)                    Evaluate one quoted expression; Java: tlc2.REPL expression evaluation.
  apalache-json                       Export checked modules as JSON IR; no Java TLC checking-mode equivalent.
  sany-xml                            Export checked modules as XML; Java: tla2sany.xml.XMLExporter.

    Java invocation: java -cp tla2tools.jar tlc2.TLC [FLAGS] Spec. Go
    invocation: tlago modelcheck --tlc [FLAGS] Spec.tla. The --tlc switch
    belongs to tlago; subsequent TLC flags keep the Java single-dash spelling
    and case. The module's .tla and config's .cfg extensions are optional in TLC
    mode. Exactly one root module is required.
    Without --tlc, modelcheck uses the earlier bounded checker, with -config and
    -maxStates only (plus module search options). Its default state limit is
    10,000. Java TLC flags such as -workers and -simulate require --tlc. Do not
    use -maxStates with --tlc; Java TLC has no equivalent state-count cutoff
    flag.
    Toolbox model editors generate a model module and .cfg from constants,
    behavior, invariants, properties, constraints, symmetry, and model values.
    On the CLI these model choices belong in the .tla/.cfg files; they are not
    separate command-line flags. GUI names below refer to the original Toolbox.
    Other editors may label the same choices differently.
FLAG QUICK REFERENCE

Help (all commands)
  -help, --help, -h                            Print this guide without running a command.

Module search (parse, check, modelcheck, repl-expr, and exporters)
  -I DIR                                       Add a module search directory; repeat as needed.
  --prefer-library-modules                     Prefer configured library modules during resolution.

Diagnostics (check and modelcheck --tlc)
  -suppressMessages CODES                      Suppress selected diagnostic codes.
  -messagesAsErrors CODES                      Elevate selected diagnostic codes to errors.

Model-checker selection and bounded checker
  --tlc                                        Select the Go port of Java TLC.
  -config FILE                                 Select the model configuration; default SPEC.cfg.
  -maxStates N                                 Limit states in the bounded checker; default 10,000.

TLC exploration (requires modelcheck --tlc)
  -modelcheck                                  Select the ordinary model-checking command variant.
  -workers N|auto                              Set worker count; default 1.
  -dfid N                                      Use depth-first iterative deepening.
  -simulate [OPTIONS]                          Generate random traces instead of exhaustive checking.
  -generate [OPTIONS]                          Run probabilistic trace generation.
  -depth N                                     Set simulation trace length; default 100.
  -seed N                                      Set the random seed for reproducible choices.
  -aril N                                      Advance the random generator's starting position.
  -deadlock                                    Disable deadlock checking.
  -continue                                    Continue after invariant violations.
  -inv EXPR                                    Add a state-level invariant expression.
  -invlevel N                                  Report reaching a selected state-space level.
  -lncheck STRATEGY                            Select the liveness-checking schedule.
  -maxSetSize N                                Bound set enumeration; default 1,000,000.

TLC fingerprints, storage, and checkpoints (requires --tlc)
  -fp N                                        Choose fingerprint polynomial index; default random.
  -fpbits N                                    Partition fingerprint storage using high bits.
  -fpmem FRACTION                              Allocate a fraction of the fingerprint memory budget.
  -metadir DIR                                 Choose the metadata root directory.
  -checkpoint MINUTES                          Set checkpoint interval; default about 30 minutes.
  -recover DIR                                 Recover a saved checker checkpoint.
  -cleanup                                     Remove old states metadata before a fresh run.
  -gzip                                        Enable gzip for value input/output streams.

TLC output, traces, and diagnostics (requires --tlc)
  -coverage MINUTES                            Collect and periodically report coverage.
  -difftrace                                   Print only changed values in successive states.
  -terse                                       Avoid expanding values in Print output.
  -view                                        Use the configured VIEW for state presentation.
  -dump [FORMAT] FILE                          Dump explored states as text or a DOT graph.
  -dumpTrace FORMAT FILE                       Export a counterexample trace; repeat for formats.
  -loadTrace FORMAT FILE                       Constrain checking to a previously exported trace.
  -postCondition MODULE!OPERATOR               Evaluate a constant-level operator after checking.
  -generateSpecTE [nomonolith]                 Force generation of a trace-exploration specification.
  -noGenerateSpecTE, -noTE                     Disable automatic trace-exploration spec generation.
  -noGenerateSpecTEBin, -noTEBin               Disable the TE binary companion trace.
  -teSpecOutDir PATH                           Choose where trace-exploration artifacts are written.
  -userFile FILE                               Write specification user output to a file.
  -nowarning                                   Suppress all warnings.
  -tool                                        Use message-code framing for editor integration.
  -debug                                       Print internal checker debugging information.
  -debugger [OPTIONS]                          Enable the interactive specification debugger.
  -DNAME=VALUE                                 Set a ported TLC runtime property.

Expression evaluation (repl-expr)
  -spec FILE, --spec FILE                      Load a module as expression-evaluation context.
  --                                           Treat the remaining argument as the expression.

SANY XML export (sany-xml)
  -o, --offline                                Accept the Java offline XML-export option.
  -t, --terse                                  Accept the Java terse XML-export option.
  -r, --restricted                             Accept the Java restricted XML-export option.
  -u, --uncomment                              Remove comment delimiters from XML pre-comments.

Implementation trace monitor (checkimplfile; not modelcheck)
  -config FILE                                 Select the implementation-check configuration.
  -trace PREFIX                                Choose the incoming implementation trace-file prefix.
  -depth N                                     Set implementation-check depth; default 20.
  -workers N|auto                              Set implementation-check worker count.
  -deadlock                                    Disable implementation-check deadlock checking.
  -recover DIR                                 Restore implementation-check metadata.
  -coverage MINUTES                            Set implementation-check coverage interval.

FLAG DETAILS AND JAVA / TOOLBOX CORRESPONDENCE

Help (all commands)

  -help, --help, -h
    Use tlago -help for the full guide, COMMAND -help for that command's
    options, or tlago help COMMAND. Help is printed to standard output and does
    not load a spec, create metadata, or start checking.
    Java TLC: -help or -h; SANY: -help. The Go wrapper also accepts --help and
    the help subcommand.

Module search (parse, check, modelcheck, repl-expr, and exporters)

  -I DIR
    Search additional directories for imported or extended modules. Aliases:
    -include, --include, --library-path. Also accepts -IDIR, -I=DIR,
    --include=DIR, and --library-path=DIR. Paths are repeatable; this is a
    directory option, not a JVM classpath or jar loader.
    Java SANY/TLC: analogous to TLA-Library module search paths; Toolbox:
    specification library paths. -I is a tlago wrapper spelling.

  --prefer-library-modules
    Resolve competing module names using the configured library preference.
    Aliases: -preferLibraryModules and --preferLibraryModules. Use together with
    -I when selecting a library copy of a module.
    Go loader option; no same-named Java TLC flag or dedicated Toolbox switch.

Diagnostics (check and modelcheck --tlc)

  -suppressMessages CODES
    Provide a comma-separated list of warning/message codes. In check mode,
    aliases include --suppressMessages, -suppress-messages, and
    --suppress-messages; E/W-prefixed diagnostic codes are accepted there. TLC
    mode uses the exact -suppressMessages spelling and numeric Java SANY/TLC
    codes. A code cannot also be elevated. TLC rejects combining this with
    -nowarning.
    Java SANY and TLC: -suppressMessages. Toolbox: additional TLC parameters for
    a checker run; parsing warnings use the SANY diagnostic controls.

  -messagesAsErrors CODES
    Use a comma-separated code list. Check mode accepts --messagesAsErrors,
    -messages-as-errors, --messages-as-errors, and E/W-prefixed codes. TLC mode
    uses -messagesAsErrors with numeric Java codes. Suppressing and elevating
    the same code is an error.
    Java SANY and TLC: -messagesAsErrors. Toolbox: additional TLC parameters.

Model-checker selection and bounded checker

  --tlc
    Required to use the TLC flags below. Aliases: -tlc, -go-tlc, --go-tlc. The
    default modelcheck command retains the earlier bounded checker; this switch
    selects the full TLC runner.
    Go wrapper switch. Equivalent invocation target in Java is tlc2.TLC; Toolbox
    launches that checker automatically.

  -config FILE
    Accepted by both checkers. The bounded checker also accepts --config and
    requires an actual file path. TLC accepts -config, with an optional .cfg
    extension. The configuration defines constants, behavior, invariants,
    properties, constraints, symmetry, and other model choices.
    Java TLC: -config FILE. Toolbox generates this file from the model editor;
    it is not a replacement for the root specification.

  -maxStates N
    Only available without --tlc. N must be positive. Aliases: --maxStates,
    -max-states, --max-states. Reaching the limit reports a failure rather than
    claiming the model was fully checked.
    Go bounded-checker option; no equivalent Java TLC flag or Toolbox
    model-checking state limit.

TLC exploration (requires modelcheck --tlc)

  -modelcheck
    Accepted for Java CLI compatibility. A fresh TLC invocation already defaults
    to exhaustive breadth-first model checking. Choose -dfid for iterative
    deepening or -simulate for random traces.
    Java TLC: -modelcheck. Toolbox: Model-checking mode in Advanced Options.

  -workers N|auto
    N must be at least 1. auto selects the available CPU count. More workers
    change exploration scheduling and can change which error trace is found
    first. DFID requires one worker.
    Java TLC: -workers. Toolbox: worker-thread count in the execution
    configuration / Advanced Options.

  -dfid N
    Start DFID with nonnegative depth N, rather than the default breadth-first
    search. Use -workers 1. This is an exploration strategy, not a maximum
    number of reachable states.
    Java TLC: -dfid N. Toolbox: Depth-first checkbox and initial depth in
    Advanced Options.

  -simulate [OPTIONS]
    Optional comma-separated OPTIONS: num=N limits traces; file=PREFIX writes
    them and requires num=N; stats=basic or stats=full collects action
    statistics; sched=rl or sched=rlaction selects the corresponding learning
    scheduler. Default trace length is 100, with effectively unbounded trace
    count. Use -depth, -seed, and -aril to control runs. Sampling does not
    establish exhaustive correctness.
    Java TLC: -simulate with the same subarguments. Toolbox: Simulation mode,
    trace length, seed, and aril in Advanced Options; extra subarguments use
    additional TLC parameters.

  -generate [OPTIONS]
    Accepts the same subarguments as -simulate and additionally enables the
    probabilistic-generation runtime property. Use finite num=N for a bounded
    manual run.
    Java TLC: -generate. Toolbox: additional TLC parameters.

  -depth N
    Controls the length of generated simulation traces. It does not impose a
    breadth-first model-checking depth limit. Use -invlevel if you want an
    invariant that reports reaching a chosen level.
    Java TLC: -depth. Toolbox: simulation trace-length setting.

  -seed N
    Accepts a signed 64-bit integer. The default is chosen at run time.
    Reproducing simulation traces also requires the same aril, worker count,
    scheduling, and model; concurrent runs can still differ in scheduling.
    Java TLC: -seed. Toolbox: Seed field in Simulation mode.

  -aril N
    Accepts a signed 64-bit integer; default 0. Used with the seed to reproduce
    a simulation's random-generator position.
    Java TLC: -aril. Toolbox: Aril field in Simulation mode.

  -deadlock
    Despite the name, presence of this flag turns checking OFF. By default
    checking is on unless the .cfg sets CHECK_DEADLOCK FALSE. The flag disables
    checking even if the configuration enables it.
    Java TLC: -deadlock. Toolbox: uncheck Deadlock under What to check.
    Equivalent config setting: CHECK_DEADLOCK FALSE.

  -continue
    The normal behavior halts on the first violation. This option continues
    invariant checking; it is not a blanket instruction to ignore evaluation
    errors or all property failures. It also suppresses automatic trace-spec
    generation unless explicitly forced.
    Java TLC: -continue. Toolbox: additional TLC parameters.

  -inv EXPR
    Quote the TLA+ expression so the shell passes it as one argument, for
    example -inv 'x >= 0'. Repeat to add invariants. This adds a runtime
    invariant without editing the configuration.
    Java TLC: -inv EXPR. Toolbox: a named invariant in What to check is the
    model-editor counterpart; ad hoc expressions use additional TLC parameters.

  -invlevel N
    Adds the invariant TLCGet("level") < N. It reports the first encountered
    trace reaching that level and stops unless -continue is used. This is an
    invariant violation, not a proof that all deeper states satisfy the
    specification.
    Java TLC: -invlevel N. Toolbox: additional TLC parameters.

  -lncheck STRATEGY
    default performs periodic checks; final defers checking to completion;
    seqfinal uses sequential final checking; sequential selects sequential
    checks; off disables liveness checking. Disabling checks does not verify
    temporal properties.
    Java TLC: -lncheck. Toolbox: Verify temporal properties upon termination
    only corresponds to final; other strategies use additional TLC parameters.

  -maxSetSize N
    Positive integer, at most 2,147,483,647. Bounds how large a set TLC may
    enumerate. This controls expression evaluation, not the number of explored
    states or simulation traces.
    Java TLC: -maxSetSize. Toolbox: additional TLC parameters.

TLC fingerprints, storage, and checkpoints (requires --tlc)

  -fp N
    Select the indexed irreducible polynomial used for 64-bit state
    fingerprints. The index is zero-based. Changing it is useful for repeating a
    run with a different fingerprint function; it does not change the
    specification's state semantics.
    Java TLC: -fp N. Toolbox: Fingerprint seed index / Select randomly in
    Advanced Options. This is different from the simulation random seed.

  -fpbits N
    Accepts 0 through 30; the default disk-set configuration uses 1. N high bits
    select among 2^N nested fingerprint sets, affecting storage organization and
    memory distribution.
    Java TLC: -fpbits N. Toolbox: fingerprint-bit setting in Advanced Options.

  -fpmem FRACTION
    Normally a fraction between 0 and 1; default 0.25. Values greater than 1 are
    the deprecated absolute-byte form. Actual allocation depends on the selected
    fingerprint implementation and runtime memory budget. Go uses its native
    budget (including GOMEMLIMIT where applicable), not a JVM heap. This does
    not set a hard limit for the entire process.
    Java TLC: -fpmem. Toolbox: fingerprint-memory allocation. Java -Xmx and the
    Toolbox JVM heap control have no tlago flag equivalent; configure the Go
    runtime separately.

  -metadir DIR
    Store state pools, fingerprint files, and checkpoints in a generated run
    directory beneath DIR. By default the metadata root is states; recovery
    reuses the supplied checkpoint directory. Reserve disk space for the full
    run and merge/checkpoint files.
    Java TLC: -metadir. Toolbox manages metadata under its model directory; a
    CLI run chooses its own root.

  -checkpoint MINUTES
    Nonnegative integer minutes. Zero disables periodic checkpoints. Checkpoints
    capture recoverable checker state; they are not counterexample trace dumps.
    Java TLC: -checkpoint. Toolbox: additional TLC parameters; its recovery
    control selects a saved checkpoint.

  -recover DIR
    Supply the checkpoint metadata directory from an earlier run. Keep the
    specification, configuration, and relevant checker settings compatible. This
    resumes checker storage; it does not replay an exported error trace.
    Java TLC: -recover DIR. Toolbox: recover from a checkpoint of the model.

  -cleanup
    Without -recover, performs the source cleanup of the states directory before
    starting. It can delete previous run storage; keep checkpoints you need
    elsewhere. With recovery, the recovery directory is retained.
    Java TLC: -cleanup. Toolbox normally manages cleanup for its model runs.

  -gzip
    Compression is off by default. Enables the original value-stream gzip layer
    where those streams are used; it does not mean every metadata or
    random-access file is compressed.
    Java TLC: -gzip. Toolbox: additional TLC parameters.

TLC output, traces, and diagnostics (requires --tlc)

  -coverage MINUTES
    Nonnegative report interval in minutes; coverage collection is off by
    default. Zero enables collection without a positive periodic interval.
    Coverage profiles expression/action evaluation and helps find unused or
    unexpectedly expensive specification code.
    Java TLC: -coverage. Toolbox: coverage / profiling selection in Advanced
    Options and the model results.

  -difftrace
    The default prints complete state descriptions. This changes counterexample
    presentation, not the states that are explored or serialized.
    Java TLC: -difftrace. Toolbox: corresponds to viewing changes between
    error-trace states; the CLI flag controls textual output.

  -terse
    Controls value expansion in specification Print statements. Use it to keep
    user output compact; it is not the same as printing differences between
    trace states.
    Java TLC: -terse. Toolbox: additional TLC parameters.

  -view
    Enables the source VIEW-based output behavior. Define the VIEW operator in
    the model configuration; the flag does not take an expression argument.
    Java TLC: -view, together with the model's VIEW definition. Toolbox: model
    view expression; additional TLC parameters enable this CLI switch.

  -dump [FORMAT] FILE
    Omit FORMAT for a text state dump (.dump suffix). For GraphViz use dot or
    dot,colorize,actionlabels,constrained,snapshot,stuttering,strict; the .dot
    suffix is added. colorize colors transitions by action; actionlabels names
    actions; constrained includes excluded states; snapshot enables snapshots;
    stuttering includes self transitions; strict selects strict graph output.
    ${metadir} in FILE expands to the run directory. The alternate class,NAME
    form selects a registered native state writer; it does not load JVM
    bytecode.
    Java TLC: -dump with the same format switches and custom IStateWriter
    selection. Toolbox: state-graph visualization / additional TLC parameters. A
    dump contains explored states, not just an error trace.

  -dumpTrace FORMAT FILE
    Formats accepted by the port: tla, json, tlc, tlcplain, tlcTESpec,
    tlcaction, dot. Dumps are produced through the corresponding postcondition
    modules. Paths follow the source trace module's resolution relative to the
    main specification. Binary tlc output can be used with -loadTrace; it is not
    a checkpoint.
    Java TLC: -dumpTrace FORMAT FILE, including source module-backed formats
    beyond its abbreviated help. Toolbox: error-trace export is the GUI
    counterpart.

  -loadTrace FORMAT FILE
    Accepts tlc (binary) or json. Installs the corresponding trace constraint
    and view. Use for model-checking trace replay, not simulation or recovery of
    a complete exploration. Keep the original trace's variables and model
    compatible.
    Java TLC: -loadTrace FORMAT FILE. Toolbox: related to trace exploration, but
    loading this constraint is a CLI operation rather than the full GUI trace
    explorer.

  -postCondition MODULE!OPERATOR
    Repeat to add postconditions. The operator can inspect final TLC results,
    export artifacts, or assert final model conditions. Supply the module and
    operator name as one argument, such as MyChecks!CheckResult.
    Java TLC: -postCondition. Toolbox: additional TLC parameters; module
    definitions live in the specification/model.

  -generateSpecTE [nomonolith]
    Normally a suitable error produces a TE specification automatically, except
    in tool mode, continuation mode, or when checking a TE spec itself. This
    flag forces generation; nomonolith disables the monolithic form. Binary
    companion output remains controlled separately.
    Java TLC: -generateSpecTE. Toolbox: trace exploration works with generated
    trace specifications and its own model settings.

  -noGenerateSpecTE, -noTE
    Suppress the default TE artifact. Explicit -generateSpecTE overrides this
    suppression. -noTE is case-insensitive in the source parser. This does not
    suppress the checker reporting the violation itself.
    Java TLC: the same flags. Toolbox's trace-exploration workflow is the
    corresponding feature.

  -noGenerateSpecTEBin, -noTEBin
    Keep TE specification generation settings but omit its binary trace
    companion. The short alias -noTEBin is case-insensitive.
    Java TLC: the same flags. Toolbox: trace-exploration artifacts; extra
    switches use additional TLC parameters.

  -teSpecOutDir PATH
    Default is the root specification's directory. A directory selects the
    destination; a .tla path selects the TE module filename. Relative paths are
    resolved by the native filesystem from the invocation's working directory.
    Java TLC: -teSpecOutDir. Toolbox normally chooses the trace-exploration
    model's output location.

  -userFile FILE
    Creates or truncates FILE for user output such as Print. Prefer an absolute
    path, as in the Java usage contract. This is separate from the checker
    diagnostics and counterexample reports.
    Java TLC: -userFile. Toolbox: specification/user output in model results;
    this flag adds a file destination.

  -nowarning
    Warnings normally appear. Cannot be combined with -suppressMessages, which
    selects individual codes. Suppression changes reporting; it does not
    establish that the model is correct.
    Java TLC: -nowarning. Toolbox: additional TLC parameters.

  -tool
    Wraps output in the source tool protocol so an editor can parse message
    types. Plain human-readable output is the default. Tool mode disables
    automatic TE generation unless it is forced.
    Java TLC: -tool. Toolbox automatically uses this output protocol when
    launching TLC.

  -debug
    Enables developer diagnostics. It is distinct from -debugger, which attaches
    an interactive debug front end.
    Java TLC: -debug. Toolbox: additional TLC parameters.

  -debugger [OPTIONS]
    Optional comma-separated OPTIONS: nosuspend selects continuing initially;
    nohalt disables stopping on evaluation/runtime errors; port=N records the
    requested debug port (default 4712). The current Go CLI installs debugger
    evaluation hooks but does not launch a network DAP server, so this flag
    alone cannot connect a VS Code front end. Use one worker for debugging.
    Java TLC: -debugger with the same options, typically used with the TLA+ VS
    Code debug front end. This is separate from Java -debug diagnostics.

  -DNAME=VALUE
    Supply one argument, for example -Dtlc2.TLC.progressInterval=1000
    (milliseconds). Properties retain their source Java names and only take
    effect where the Go port consumes them. JVM switches such as -Xmx and -XX
    flags are not accepted by tlago.
    Java: normally a JVM -DNAME=VALUE option before tlc2.TLC; the port accepts
    it among TLC arguments. Toolbox: additional JVM arguments/system properties
    for Java; here use the TLC-mode -D form.

Expression evaluation (repl-expr)

  -spec FILE, --spec FILE
    Evaluate exactly one shell-quoted expression with this specification's
    operators and imports. -I and library-preference options are also supported.
    Without -spec, evaluation uses the built-in REPL context.
    Java: tlc2.REPL expression evaluation. This command evaluates once; it is
    not an interactive Toolbox model checker.

  --
    Ends option processing for repl-expr. The expression is still exactly one
    argument; quote it to preserve spaces and shell characters. A help-like
    argument after -- is expression text.
    Go CLI delimiter; no dedicated TLC model-checking or Toolbox equivalent.

SANY XML export (sany-xml)

  -o, --offline
    Java -o skips schema validation. The current Go exporter performs no online
    schema validation, so this accepted option does not change its output.
    Parsing and semantic checking still run.
    Java XMLExporter: -o. No TLC model-checking or Toolbox checking-mode
    equivalent.

  -t, --terse
    Java -t omits formatting tabs and newlines. The current Go exporter accepts
    this option but does not yet apply a distinct terse formatting mode. It is
    unrelated to TLC -terse.
    Java XMLExporter: -t. No TLC flag equivalent.

  -r, --restricted
    Java -r exports only the root module's declarations and definitions. The
    current Go exporter accepts the option but does not yet apply a separate
    restricted-export mode.
    Java XMLExporter: -r. No TLC checking-mode equivalent.

  -u, --uncomment
    Process operator pre-comments to extract their content, including boxed and
    single-line comment forms. Pre-comments are already exported without this
    flag; -u controls their normalization.
    Java XMLExporter: -u. No TLC checking-mode equivalent.

Implementation trace monitor (checkimplfile; not modelcheck)

  -config FILE
    Defaults to SPEC.cfg; the .cfg extension is optional. Defines the
    specification against which implementation traces are checked.
    Java: tlc2.tool.CheckImplFile -config. This monitor has no ordinary Toolbox
    model-editor launch equivalent.

  -trace PREFIX
    Defaults to SPEC_trace. The monitor polls for implementation-produced traces
    and keeps running while checking them; it is not a one-shot exported
    counterexample reader.
    Java: CheckImplFile -trace. Different from TLC -loadTrace.

  -depth N
    Sets this monitor's checking depth. It is independent of the TLC simulation
    -depth option.
    Java: CheckImplFile -depth.

  -workers N|auto
    Select a positive worker count or auto. This command uses its own
    implementation-checking workflow.
    Java: CheckImplFile -workers.

  -deadlock
    Checking is enabled by default; presence of this flag disables it, as in the
    Java monitor.
    Java: CheckImplFile -deadlock.

  -recover DIR
    Use saved metadata, including the recovered unique-string table, for the
    monitor.
    Java: CheckImplFile -recover.

  -coverage MINUTES
    Nonnegative minutes for this monitor's coverage reporting. This command does
    not accept the full set of TLC runner flags.
    Java: CheckImplFile -coverage.

EXAMPLES
  tlago modelcheck --tlc -workers auto -config MC.cfg MC.tla
  tlago modelcheck --tlc -simulate num=100 -depth 50 -seed 1 MC.tla
  tlago modelcheck --tlc -dump dot,actionlabels graph MC.tla
  tlago modelcheck --tlc -dumpTrace json error.json MC.tla
  tlago modelcheck --tlc -inv 'x >= 0' MC.tla
  tlago check -I ./modules Spec.tla
  tlago repl-expr --spec Spec.tla 'Cardinality({1, 2, 3})'
  tlago help modelcheck

    Exit status: help and successful commands return 0; argument, parsing,
    semantic, or model-checking failures return a nonzero status. The tlago
    wrapper uses its own exit statuses; do not interpret them as Java TLC's full
    error-code-to-exit-code table.
~~~


# repos with corpora/example TLA+ specs to check parsing/processing on

These are vendored under test_vectors/ now

There are several open-source repositories and standardized corpora that serve as test suites for TLA+ parsers.

Because TLA+ features tricky syntactic edge cases (such as whitespace-aligned conjunction/disjunction lists and complex operator precedence), testing against real-world specifications is essential.

Here are the best collections of TLA+ specs to test your parser against, ranked by utility:

---

### 1. `tlaplus/Examples` (The Standard Corpus)

This is the official, community-maintained repository of TLA+ specifications managed by the TLA+ Foundation. It contains hundreds of `.tla` files ranging from simple beginner puzzles to full industrial protocols (like Paxos and Raft).

* **Key specs to include in your parser suite:**
* `DieHard.tla` / `NQueens.tla`: Simple, clean syntax for initial sanity checks.
* `Paxos.tla` / `Raft.tla`: Large, complex distributed systems specifications with heavy operator nesting.
* `Specifying Systems` folder: The exact specifications featured in Leslie Lamport’s foundational textbook *Specifying Systems*.


* **Repository:** [github.com/tlaplus/Examples](https://github.com/tlaplus/Examples?utm_source=gemini)

### 2. Tree-Sitter TLA+ Corpus (Best for Parsing Edge Cases)

If you want focused, bite-sized test cases specifically tailored for grammar validation, check out the test suite written for `tree-sitter-tlaplus`.

* **Why it's useful:** Tree-sitter repositories organize tests into a structured `corpus/` folder containing hundreds of target syntax snippets paired with expected Abstract Syntax Trees (ASTs).
* **What it tests:** Highly specific parsing boundaries—such as nested alignment-based conjunction (`/\`) and disjunction (`\/`) lists, record constructs, tuple syntax, and operator fixity/precedence rules.
* **Repository:** [github.com/tlaplus-community/tree-sitter-tlaplus](https://github.com/tlaplus-community/tree-sitter-tlaplus?utm_source=gemini)


## Key Syntactic Stress-Tests for Your Parser

When testing your parser against these specifications, ensure your test suite exercises these specific TLA+ grammar features:

1. **Junction List Scoping:**
```tla
\* Ensure your parser respects column alignment to build the tree correctly
\/ A
\/ /\ B
   /\ C
\/ D

```

2. **Module Extensions & Instances:**
`EXTENDS Naturals, Sequences, TLC` and `INSTANCE ModuleName WITH x <- y`.

3. **Bounded & Unbounded Quantifiers:**
`\E x \in S : P(x)` vs. `\A x, y : P(x, y)`.

4. **Custom Infix & Prefix Operators:**
Operators like `x \div y`, `x ++ y`, or `x \cap y` with custom TLA+ precedence rules.

5. **PlusCal Comments:**
Many specs embed PlusCal algorithms inside `(* --algorithm ... *)` blocks. Your lexer/parser must either ignore these block comments or handle them cleanly without throwing syntax errors.

---------------
Author and license:

Copyright (C) 2026 Jason E. Aten, Ph.D.

MIT License. See the LICENSE file here in.

