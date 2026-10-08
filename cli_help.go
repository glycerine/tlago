package tlago

import (
	"fmt"
	"io"
	"strings"
)

// Keep this guide aligned with cli.go, tlc/cli.go, and
// tlc/check_impl_file.go. Java counterparts were checked against the pinned
// TLC.java and the Toolbox model-editor pages, not inferred from flag names.
type cliHelpOption struct {
	syntax      string
	summary     string
	explanation string
	counterpart string
}

type cliHelpGroup struct {
	title   string
	options []cliHelpOption
	tlc     bool
}

func isCLIHelpFlag(arg string) bool {
	return arg == "-help" || arg == "--help" || arg == "-h"
}

func canonicalCLICommand(command string) string {
	switch command {
	case "parse", "check", "apalache-json", "sany-xml":
		return command
	case "modelcheck", "mc":
		return "modelcheck"
	case "checkimplfile", "check-impl-file":
		return "checkimplfile"
	case "repl-expr", "repl":
		return "repl-expr"
	default:
		return ""
	}
}

func printCLIHelp(w io.Writer, command string) {
	fmt.Fprintln(w, "tlago: TLA+ parsing, semantic checking, and the Go TLC model checker")
	fmt.Fprintln(w)
	if command == "" {
		fmt.Fprintln(w, "Usage: tlago [TLC FLAGS] [SPEC]")
		fmt.Fprintln(w, "       tlago COMMAND [OPTIONS] FILE...")
		fmt.Fprintln(w, "       tlago -help | tlago help [COMMAND]")
	} else {
		fmt.Fprintf(w, "Usage: tlago %s [OPTIONS] %s\n", command, cliHelpOperand(command))
	}
	fmt.Fprintln(w, "Help aliases: -help, --help, -h. Help needs no input file and exits successfully.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "COMMANDS AND JAVA EQUIVALENTS")
	for _, c := range []struct{ name, description string }{
		{"parse", "Parse FILE... with the Go SANY parser; Java counterpart: SANY -s."},
		{"check", "Parse and check FILE...; Java counterpart: tla2sany.SANY, Toolbox Parse Spec."},
		{"modelcheck (mc)", "Run the Go port of Java tlc2.TLC; also the default without a subcommand."},
		{"checkimplfile (check-impl-file)", "Monitor implementation trace files; Java: tlc2.tool.CheckImplFile."},
		{"repl-expr (repl)", "Evaluate one quoted expression; Java: tlc2.REPL expression evaluation."},
		{"apalache-json", "Export checked modules as JSON IR; no Java TLC checking-mode equivalent."},
		{"sany-xml", "Export checked modules as XML; Java: tla2sany.xml.XMLExporter."},
	} {
		fmt.Fprintf(w, "  %-35s %s\n", c.name, c.description)
	}
	fmt.Fprintln(w)
	if command == "" || command == "modelcheck" {
		writeCLIHelpParagraph(w, "A root module can be omitted when the classpath contains model/MC.tla and its configuration. Set CLASSPATH to the model archive or resource directory, as with Java TLC's packaged-model mode. TLC loads model/generated.properties, uses the packaged module resolver, enables tool-mode output, and disables checkpoints. Direct invocation, modelcheck, and mc use the same runner.")
		//writeCLIHelpParagraph(w, "Java invocation: java -cp tla2tools.jar tlc2.TLC [FLAGS] Spec. Go invocation: tlago [FLAGS] Spec.tla. The optional modelcheck (mc) subcommand uses exactly the same TLC runner and flags. TLC options keep the Java single-dash spelling and case. The module's .tla and config's .cfg extensions are optional. Exactly one root module is required.")
		//writeCLIHelpParagraph(w, "Toolbox model editors generate a model module and .cfg from constants, behavior, invariants, properties, constraints, symmetry, and model values. On the CLI these model choices belong in the .tla/.cfg files; they are not separate command-line flags. GUI names below refer to the original Toolbox. Other editors may label the same choices differently.")
	}
	groups := cliHelpGroups(command)
	fmt.Fprintln(w, "FLAG QUICK REFERENCE")
	for _, group := range groups {
		fmt.Fprintln(w)
		fmt.Fprintln(w, group.title)
		for _, option := range group.options {
			fmt.Fprintf(w, "  %-44s %s\n", option.syntax, option.summary)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "FLAG DETAILS AND JAVA / TOOLBOX CORRESPONDENCE")
	for _, group := range groups {
		fmt.Fprintln(w)
		fmt.Fprintln(w, group.title)
		for _, option := range group.options {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "  "+option.syntax)
			writeCLIHelpParagraph(w, option.explanation)
			counterpart := option.counterpart
			if counterpart == "" && group.tlc {
				counterpart = "Java TLC: the same flag. Toolbox: pass it through the additional TLC parameters field in Advanced Options."
			}
			if counterpart != "" {
				writeCLIHelpParagraph(w, counterpart)
			}
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "EXAMPLES")
	if command == "" || command == "modelcheck" {
		fmt.Fprintln(w, "  tlago -workers auto -config MC.cfg MC.tla")
		fmt.Fprintln(w, "  tlago -simulate num=100 -depth 50 -seed 1 MC.tla")
		fmt.Fprintln(w, "  tlago -dump dot,actionlabels graph MC.tla")
		fmt.Fprintln(w, "  tlago -dumpTrace json error.json MC.tla")
		fmt.Fprintln(w, "  tlago -inv 'x >= 0' MC.tla")
	}
	if command == "" || command == "check" || command == "parse" {
		fmt.Fprintln(w, "  tlago check -I ./modules Spec.tla")
	}
	if command == "" || command == "repl-expr" {
		fmt.Fprintln(w, "  tlago repl-expr --spec Spec.tla 'Cardinality({1, 2, 3})'")
	}
	if command == "sany-xml" {
		fmt.Fprintln(w, "  tlago sany-xml -o -t Spec.tla")
	}
	if command == "apalache-json" {
		fmt.Fprintln(w, "  tlago apalache-json -I ./modules Spec.tla")
	}
	if command == "checkimplfile" {
		fmt.Fprintln(w, "  tlago checkimplfile -trace implementation_trace Spec.tla")
	}
	fmt.Fprintln(w, "  tlago help modelcheck")
	fmt.Fprintln(w)
	writeCLIHelpParagraph(w, "Exit status: help and successful commands return 0; argument, parsing, semantic, or model-checking failures return a nonzero status. The tlago wrapper uses its own exit statuses; do not interpret them as Java TLC's full error-code-to-exit-code table.")
}

func cliHelpOperand(command string) string {
	if command == "repl-expr" {
		return "'EXPRESSION'"
	}
	if command == "modelcheck" {
		return "[SPEC]"
	}
	if command == "checkimplfile" {
		return "SPEC"
	}
	return "FILE..."
}

func writeCLIHelpParagraph(w io.Writer, text string) {
	line := "    "
	for _, word := range strings.Fields(text) {
		if len(line) > 4 && len(line)+1+len(word) > 80 {
			fmt.Fprintln(w, line)
			line = "    "
		}
		if len(line) > 4 {
			line += " "
		}
		line += word
	}
	fmt.Fprintln(w, line)
}

func cliHelpGroups(command string) []cliHelpGroup {
	groups := []cliHelpGroup{{title: "Help (all commands)", options: []cliHelpOption{
		{"-help, --help, -h", "Print this guide without running a command.", "Use tlago -help for the full guide, COMMAND -help for that command's options, or tlago help COMMAND. Help is printed to standard output and does not load a spec, create metadata, or start checking.", "Java TLC: -help or -h; SANY: -help. The Go wrapper also accepts --help and the help subcommand."},
	}}}
	if command != "checkimplfile" {
		groups = append(groups, cliHelpGroup{title: "Module search (parse, check, modelcheck, repl-expr, and exporters)", options: []cliHelpOption{
			{"-I DIR", "Add a module search directory; repeat as needed.", "Search additional directories for imported or extended modules. Aliases: -include, --include, --library-path. Also accepts -IDIR, -I=DIR, --include=DIR, and --library-path=DIR. Paths are repeatable; this is a directory option, not a JVM classpath or jar loader.", "Java SANY/TLC: analogous to TLA-Library module search paths; Toolbox: specification library paths. -I is a tlago wrapper spelling."},
			{"--prefer-library-modules", "Prefer configured library modules during resolution.", "Resolve competing module names using the configured library preference. Aliases: -preferLibraryModules and --preferLibraryModules. Use together with -I when selecting a library copy of a module.", "Go loader option; no same-named Java TLC flag or dedicated Toolbox switch."},
		}})
	}
	if command == "" || command == "check" || command == "modelcheck" {
		groups = append(groups, cliHelpGroup{title: "Diagnostics (check and TLC)", options: []cliHelpOption{
			{"-suppressMessages CODES", "Suppress selected diagnostic codes.", "Provide a comma-separated list of warning/message codes. In check mode, aliases include --suppressMessages, -suppress-messages, and --suppress-messages; E/W-prefixed diagnostic codes are accepted there. TLC mode uses the exact -suppressMessages spelling and numeric Java SANY/TLC codes. A code cannot also be elevated. TLC rejects combining this with -nowarning.", "Java SANY and TLC: -suppressMessages. Toolbox: additional TLC parameters for a checker run; parsing warnings use the SANY diagnostic controls."},
			{"-messagesAsErrors CODES", "Elevate selected diagnostic codes to errors.", "Use a comma-separated code list. Check mode accepts --messagesAsErrors, -messages-as-errors, --messages-as-errors, and E/W-prefixed codes. TLC mode uses -messagesAsErrors with numeric Java codes. Suppressing and elevating the same code is an error.", "Java SANY and TLC: -messagesAsErrors. Toolbox: additional TLC parameters."},
		}})
	}
	if command == "" || command == "modelcheck" {
		groups = append(groups, cliHelpGroup{title: "TLC model configuration", options: []cliHelpOption{
			{"-config FILE", "Select the model configuration; default SPEC.cfg.", "Use the Java spelling -config, with an optional .cfg extension. The configuration defines constants, behavior, invariants, properties, constraints, symmetry, and other model choices.", "Java TLC: -config FILE. Toolbox generates this file from the model editor; it is not a replacement for the root specification."},
		}})
		groups = append(groups, cliHelpGroup{title: "TLC exploration", tlc: true, options: []cliHelpOption{
			{"-modelcheck", "Select the ordinary model-checking command variant.", "Accepted for Java CLI compatibility. A fresh TLC invocation already defaults to exhaustive breadth-first model checking. Choose -dfid for iterative deepening or -simulate for random traces.", "Java TLC: -modelcheck. Toolbox: Model-checking mode in Advanced Options."},
			{"-workers N|auto", "Set worker count; default 1.", "N must be at least 1. auto selects the available CPU count. More workers change exploration scheduling and can change which error trace is found first. DFID requires one worker.", "Java TLC: -workers. Toolbox: worker-thread count in the execution configuration / Advanced Options."},
			{"-dfid N", "Use depth-first iterative deepening.", "Start DFID with nonnegative depth N, rather than the default breadth-first search. Use -workers 1. This is an exploration strategy, not a maximum number of reachable states.", "Java TLC: -dfid N. Toolbox: Depth-first checkbox and initial depth in Advanced Options."},
			{"-simulate [OPTIONS]", "Generate random traces instead of exhaustive checking.", "Optional comma-separated OPTIONS: num=N limits traces; file=PREFIX writes them and requires num=N; stats=basic or stats=full collects action statistics; sched=rl or sched=rlaction selects the corresponding learning scheduler. Default trace length is 100, with effectively unbounded trace count. Use -depth, -seed, and -aril to control runs. Sampling does not establish exhaustive correctness.", "Java TLC: -simulate with the same subarguments. Toolbox: Simulation mode, trace length, seed, and aril in Advanced Options; extra subarguments use additional TLC parameters."},
			{"-generate [OPTIONS]", "Run probabilistic trace generation.", "Accepts the same subarguments as -simulate and additionally enables the probabilistic-generation runtime property. Use finite num=N for a bounded manual run.", "Java TLC: -generate. Toolbox: additional TLC parameters."},
			{"-depth N", "Set simulation trace length; default 100.", "Controls the length of generated simulation traces. It does not impose a breadth-first model-checking depth limit. Use -invlevel if you want an invariant that reports reaching a chosen level.", "Java TLC: -depth. Toolbox: simulation trace-length setting."},
			{"-seed N", "Set the random seed for reproducible choices.", "Accepts a signed 64-bit integer. The default is chosen at run time. Reproducing simulation traces also requires the same aril, worker count, scheduling, and model; concurrent runs can still differ in scheduling.", "Java TLC: -seed. Toolbox: Seed field in Simulation mode."},
			{"-aril N", "Advance the random generator's starting position.", "Accepts a signed 64-bit integer; default 0. Used with the seed to reproduce a simulation's random-generator position.", "Java TLC: -aril. Toolbox: Aril field in Simulation mode."},
			{"-deadlock", "Disable deadlock checking.", "Despite the name, presence of this flag turns checking OFF. By default checking is on unless the .cfg sets CHECK_DEADLOCK FALSE. The flag disables checking even if the configuration enables it.", "Java TLC: -deadlock. Toolbox: uncheck Deadlock under What to check. Equivalent config setting: CHECK_DEADLOCK FALSE."},
			{"-continue", "Continue after invariant violations.", "The normal behavior halts on the first violation. This option continues invariant checking; it is not a blanket instruction to ignore evaluation errors or all property failures. It also suppresses automatic trace-spec generation unless explicitly forced.", "Java TLC: -continue. Toolbox: additional TLC parameters."},
			{"-inv EXPR", "Add a state-level invariant expression.", "Quote the TLA+ expression so the shell passes it as one argument, for example -inv 'x >= 0'. Repeat to add invariants. This adds a runtime invariant without editing the configuration.", "Java TLC: -inv EXPR. Toolbox: a named invariant in What to check is the model-editor counterpart; ad hoc expressions use additional TLC parameters."},
			{"-invlevel N", "Report reaching a selected state-space level.", "Adds the invariant TLCGet(\"level\") < N. It reports the first encountered trace reaching that level and stops unless -continue is used. This is an invariant violation, not a proof that all deeper states satisfy the specification.", "Java TLC: -invlevel N. Toolbox: additional TLC parameters."},
			{"-lncheck STRATEGY", "Select the liveness-checking schedule.", "default performs periodic checks; final defers checking to completion; seqfinal uses sequential final checking; sequential selects sequential checks; off disables liveness checking. Disabling checks does not verify temporal properties.", "Java TLC: -lncheck. Toolbox: Verify temporal properties upon termination only corresponds to final; other strategies use additional TLC parameters."},
			{"-maxSetSize N", "Bound set enumeration; default 1,000,000.", "Positive integer, at most 2,147,483,647. Bounds how large a set TLC may enumerate. This controls expression evaluation, not the number of explored states or simulation traces.", "Java TLC: -maxSetSize. Toolbox: additional TLC parameters."},
		}})
		groups = append(groups, cliHelpGroup{title: "TLC fingerprints, storage, and checkpoints", tlc: true, options: []cliHelpOption{
			{"-fp N", "Choose fingerprint polynomial index; default random.", "Select the indexed irreducible polynomial used for 64-bit state fingerprints. The index is zero-based. Changing it is useful for repeating a run with a different fingerprint function; it does not change the specification's state semantics.", "Java TLC: -fp N. Toolbox: Fingerprint seed index / Select randomly in Advanced Options. This is different from the simulation random seed."},
			{"-fpbits N", "Partition fingerprint storage using high bits.", "Accepts 0 through 30; the default disk-set configuration uses 1. N high bits select among 2^N nested fingerprint sets, affecting storage organization and memory distribution.", "Java TLC: -fpbits N. Toolbox: fingerprint-bit setting in Advanced Options."},
			{"-fpmem FRACTION", "Allocate a fraction of the fingerprint memory budget.", "Normally a fraction between 0 and 1; default 0.25. Values greater than 1 are the deprecated absolute-byte form. Actual allocation depends on the selected fingerprint implementation and runtime memory budget. Go uses its native budget (including GOMEMLIMIT where applicable), not a JVM heap. This does not set a hard limit for the entire process.", "Java TLC: -fpmem. Toolbox: fingerprint-memory allocation. Java -Xmx and the Toolbox JVM heap control have no tlago flag equivalent; configure the Go runtime separately."},
			{"-metadir DIR", "Choose the metadata root directory.", "Store state pools, fingerprint files, and checkpoints in a generated run directory beneath DIR. By default the metadata root is states; recovery reuses the supplied checkpoint directory. Reserve disk space for the full run and merge/checkpoint files.", "Java TLC: -metadir. Toolbox manages metadata under its model directory; a CLI run chooses its own root."},
			{"-checkpoint MINUTES", "Set checkpoint interval; default about 30 minutes.", "Nonnegative integer minutes. Zero disables periodic checkpoints. Checkpoints capture recoverable checker state; they are not counterexample trace dumps.", "Java TLC: -checkpoint. Toolbox: additional TLC parameters; its recovery control selects a saved checkpoint."},
			{"-recover DIR", "Recover a saved checker checkpoint.", "Supply the checkpoint metadata directory from an earlier run. Keep the specification, configuration, and relevant checker settings compatible. This resumes checker storage; it does not replay an exported error trace.", "Java TLC: -recover DIR. Toolbox: recover from a checkpoint of the model."},
			{"-cleanup", "Remove old states metadata before a fresh run.", "Without -recover, performs the source cleanup of the states directory before starting. It can delete previous run storage; keep checkpoints you need elsewhere. With recovery, the recovery directory is retained.", "Java TLC: -cleanup. Toolbox normally manages cleanup for its model runs."},
			{"-gzip", "Enable gzip for value input/output streams.", "Compression is off by default. Enables the original value-stream gzip layer where those streams are used; it does not mean every metadata or random-access file is compressed.", "Java TLC: -gzip. Toolbox: additional TLC parameters."},
		}})
		groups = append(groups, cliHelpGroup{title: "TLC output, traces, and diagnostics", tlc: true, options: []cliHelpOption{
			{"-coverage MINUTES", "Collect and periodically report coverage.", "Nonnegative report interval in minutes; coverage collection is off by default. Zero enables collection without a positive periodic interval. Coverage profiles expression/action evaluation and helps find unused or unexpectedly expensive specification code.", "Java TLC: -coverage. Toolbox: coverage / profiling selection in Advanced Options and the model results."},
			{"-difftrace", "Print only changed values in successive states.", "The default prints complete state descriptions. This changes counterexample presentation, not the states that are explored or serialized.", "Java TLC: -difftrace. Toolbox: corresponds to viewing changes between error-trace states; the CLI flag controls textual output."},
			{"-terse", "Avoid expanding values in Print output.", "Controls value expansion in specification Print statements. Use it to keep user output compact; it is not the same as printing differences between trace states.", "Java TLC: -terse. Toolbox: additional TLC parameters."},
			{"-view", "Use the configured VIEW for state presentation.", "Enables the source VIEW-based output behavior. Define the VIEW operator in the model configuration; the flag does not take an expression argument.", "Java TLC: -view, together with the model's VIEW definition. Toolbox: model view expression; additional TLC parameters enable this CLI switch."},
			{"-dump [FORMAT] FILE", "Dump explored states as text or a DOT graph.", "Omit FORMAT for a text state dump (.dump suffix). For GraphViz use dot or dot,colorize,actionlabels,constrained,snapshot,stuttering,strict; the .dot suffix is added. colorize colors transitions by action; actionlabels names actions; constrained includes excluded states; snapshot enables snapshots; stuttering includes self transitions; strict selects strict graph output. ${metadir} in FILE expands to the run directory. The alternate class,NAME form selects a registered native state writer; it does not load JVM bytecode.", "Java TLC: -dump with the same format switches and custom IStateWriter selection. Toolbox: state-graph visualization / additional TLC parameters. A dump contains explored states, not just an error trace."},
			{"-dumpTrace FORMAT FILE", "Export a counterexample trace; repeat for formats.", "Formats accepted by the port: tla, json, tlc, tlcplain, tlcTESpec, tlcaction, dot. Dumps are produced through the corresponding postcondition modules. Paths follow the source trace module's resolution relative to the main specification. Binary tlc output can be used with -loadTrace; it is not a checkpoint.", "Java TLC: -dumpTrace FORMAT FILE, including source module-backed formats beyond its abbreviated help. Toolbox: error-trace export is the GUI counterpart."},
			{"-loadTrace FORMAT FILE", "Constrain checking to a previously exported trace.", "Accepts tlc (binary) or json. Installs the corresponding trace constraint and view. Use for model-checking trace replay, not simulation or recovery of a complete exploration. Keep the original trace's variables and model compatible.", "Java TLC: -loadTrace FORMAT FILE. Toolbox: related to trace exploration, but loading this constraint is a CLI operation rather than the full GUI trace explorer."},
			{"-postCondition MODULE!OPERATOR", "Evaluate a constant-level operator after checking.", "Repeat to add postconditions. The operator can inspect final TLC results, export artifacts, or assert final model conditions. Supply the module and operator name as one argument, such as MyChecks!CheckResult.", "Java TLC: -postCondition. Toolbox: additional TLC parameters; module definitions live in the specification/model."},
			{"-generateSpecTE [nomonolith]", "Force generation of a trace-exploration specification.", "Normally a suitable error produces a TE specification automatically, except in tool mode, continuation mode, or when checking a TE spec itself. This flag forces generation. The pinned Java TLC accepts nomonolith but does not use it to change output; Go follows that behavior. Binary companion output remains controlled separately.", "Java TLC: -generateSpecTE. Toolbox: trace exploration works with generated trace specifications and its own model settings."},
			{"-noGenerateSpecTE, -noTE", "Disable automatic trace-exploration spec generation.", "Suppress the default TE artifact. Explicit -generateSpecTE overrides this suppression. -noTE is case-insensitive in the source parser. This does not suppress the checker reporting the violation itself.", "Java TLC: the same flags. Toolbox's trace-exploration workflow is the corresponding feature."},
			{"-noGenerateSpecTEBin, -noTEBin", "Disable the TE binary companion trace.", "Keep TE specification generation settings but omit its binary trace companion. The short alias -noTEBin is case-insensitive.", "Java TLC: the same flags. Toolbox: trace-exploration artifacts; extra switches use additional TLC parameters."},
			{"-teSpecOutDir PATH", "Choose where trace-exploration artifacts are written.", "Default is the root specification's directory. A directory selects the destination; a .tla path selects the TE module filename. Relative paths are resolved by the native filesystem from the invocation's working directory.", "Java TLC: -teSpecOutDir. Toolbox normally chooses the trace-exploration model's output location."},
			{"-userFile FILE", "Write specification user output to a file.", "Creates or truncates FILE for user output such as Print. Prefer an absolute path, as in the Java usage contract. This is separate from the checker diagnostics and counterexample reports.", "Java TLC: -userFile. Toolbox: specification/user output in model results; this flag adds a file destination."},
			{"-nowarning", "Suppress all warnings.", "Warnings normally appear. Cannot be combined with -suppressMessages, which selects individual codes. Suppression changes reporting; it does not establish that the model is correct.", "Java TLC: -nowarning. Toolbox: additional TLC parameters."},
			{"-tool", "Use message-code framing for editor integration.", "Wraps output in the source tool protocol so an editor can parse message types. Plain human-readable output is the default. Tool mode disables automatic TE generation unless it is forced.", "Java TLC: -tool. Toolbox automatically uses this output protocol when launching TLC."},
			{"-debug", "Print internal checker debugging information.", "Enables developer diagnostics. It is distinct from -debugger, which attaches an interactive debug front end.", "Java TLC: -debug. Toolbox: additional TLC parameters."},
			{"-debugger [OPTIONS]", "Enable the interactive specification debugger.", "Optional comma-separated OPTIONS: nosuspend selects continuing initially; nohalt disables stopping on evaluation/runtime errors; port=N records the requested debug port (default 4712). The current Go CLI installs debugger evaluation hooks but does not launch a network DAP server, so this flag alone cannot connect a VS Code front end. Use one worker for debugging.", "Java TLC: -debugger with the same options, typically used with the TLA+ VS Code debug front end. This is separate from Java -debug diagnostics."},
			{"-DNAME=VALUE", "Set a ported TLC runtime property.", "Supply one argument, for example -Dtlc2.TLC.progressInterval=1000 (milliseconds). Properties retain their source Java names and only take effect where the Go port consumes them. JVM switches such as -Xmx and -XX flags are not accepted by tlago.", "Java: normally a JVM -DNAME=VALUE option before tlc2.TLC; the port accepts it among TLC arguments. Toolbox: additional JVM arguments/system properties for Java; here use the TLC-mode -D form."},
		}})
	}
	if command == "" || command == "repl-expr" {
		groups = append(groups, cliHelpGroup{title: "Expression evaluation (repl-expr)", options: []cliHelpOption{
			{"-spec FILE, --spec FILE", "Load a module as expression-evaluation context.", "Evaluate exactly one shell-quoted expression with this specification's operators and imports. -I and library-preference options are also supported. Without -spec, evaluation uses the built-in REPL context.", "Java: tlc2.REPL expression evaluation. This command evaluates once; it is not an interactive Toolbox model checker."},
			{"--", "Treat the remaining argument as the expression.", "Ends option processing for repl-expr. The expression is still exactly one argument; quote it to preserve spaces and shell characters. A help-like argument after -- is expression text.", "Go CLI delimiter; no dedicated TLC model-checking or Toolbox equivalent."},
		}})
	}
	if command == "" || command == "sany-xml" {
		groups = append(groups, cliHelpGroup{title: "SANY XML export (sany-xml)", options: []cliHelpOption{
			{"-o, --offline", "Skip XML schema validation.", "Skip validation against the embedded SANY XSD. By default, XML export validates through xmllint (libxml2), which must be on PATH. Validation uses no network access. Parsing and level checking still run.", "Java XMLExporter: -o. No TLC model-checking or Toolbox checking-mode equivalent."},
			{"-t, --terse", "Omit XML indentation and formatting newlines.", "Omit XML indentation and formatting newlines, preserving character data. It is unrelated to TLC -terse.", "Java XMLExporter: -t. No TLC flag equivalent."},
			{"-r, --restricted", "Accept the Java restricted XML-export option.", "Java -r exports only the root module's declarations and definitions. The current Go exporter accepts the option but does not yet apply a separate restricted-export mode.", "Java XMLExporter: -r. No TLC checking-mode equivalent."},
			{"-u, --uncomment", "Remove comment delimiters from XML pre-comments.", "Process operator pre-comments to extract their content, including boxed and single-line comment forms. Pre-comments are already exported without this flag; -u controls their normalization.", "Java XMLExporter: -u. No TLC checking-mode equivalent."},
		}})
	}
	if command == "" || command == "checkimplfile" {
		groups = append(groups, cliHelpGroup{title: "Implementation trace monitor (checkimplfile; not modelcheck)", options: []cliHelpOption{
			{"-config FILE", "Select the implementation-check configuration.", "Defaults to SPEC.cfg; the .cfg extension is optional. Defines the specification against which implementation traces are checked.", "Java: tlc2.tool.CheckImplFile -config. This monitor has no ordinary Toolbox model-editor launch equivalent."},
			{"-trace PREFIX", "Choose the incoming implementation trace-file prefix.", "Defaults to SPEC_trace. The monitor polls for implementation-produced traces and keeps running while checking them; it is not a one-shot exported counterexample reader.", "Java: CheckImplFile -trace. Different from TLC -loadTrace."},
			{"-depth N", "Set implementation-check depth; default 20.", "Sets this monitor's checking depth. It is independent of the TLC simulation -depth option.", "Java: CheckImplFile -depth."},
			{"-workers N|auto", "Set implementation-check worker count.", "Select a positive worker count or auto. This command uses its own implementation-checking workflow.", "Java: CheckImplFile -workers."},
			{"-deadlock", "Disable implementation-check deadlock checking.", "Checking is enabled by default; presence of this flag disables it, as in the Java monitor.", "Java: CheckImplFile -deadlock."},
			{"-recover DIR", "Restore implementation-check metadata.", "Use saved metadata, including the recovered unique-string table, for the monitor.", "Java: CheckImplFile -recover."},
			{"-coverage MINUTES", "Set implementation-check coverage interval.", "Nonnegative minutes for this monitor's coverage reporting. This command does not accept the full set of TLC runner flags.", "Java: CheckImplFile -coverage."},
		}})
	}
	return groups
}
