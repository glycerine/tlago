package tlago

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tlcruntime "github.com/glycerine/tlago/tlc"
)

func RunCLI(args []string, stdout, stderr io.Writer) int {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if len(args) == 0 {
		if tlcruntime.NewModelInJar().HasModel() {
			return runModelCheck(args, stdout, stderr)
		}
		fmt.Fprintln(stderr, "usage: tlago [TLC FLAGS] SPEC | parse|check|modelcheck|checkimplfile|repl-expr|apalache-json|sany-xml [OPTIONS] FILE...")
		fmt.Fprintln(stderr, "Run tlago -help for commands, flags, examples, and Java/Toolbox equivalents.")
		return ExitToolFailure
	}
	if isCLIHelpFlag(args[0]) {
		printCLIHelp(stdout, "")
		return ExitOK
	}
	if args[0] == "help" {
		if len(args) == 1 {
			printCLIHelp(stdout, "")
			return ExitOK
		}
		if len(args) == 2 && canonicalCLICommand(args[1]) != "" {
			printCLIHelp(stdout, canonicalCLICommand(args[1]))
			return ExitOK
		}
		fmt.Fprintln(stderr, "usage: tlago help [COMMAND]")
		return ExitToolFailure
	}
	cmd := args[0]
	if canonicalCLICommand(cmd) == "" {
		if cliArgsContainHelp(args) {
			printCLIHelp(stdout, "modelcheck")
			return ExitOK
		}
		return runModelCheck(args, stdout, stderr)
	}
	files := args[1:]
	if command := canonicalCLICommand(cmd); command != "" && cliArgsContainHelp(files) {
		printCLIHelp(stdout, command)
		return ExitOK
	}
	if len(files) == 0 && cmd != "modelcheck" && cmd != "mc" {
		fmt.Fprintln(stderr, "at least one file is required")
		return ExitToolFailure
	}
	switch cmd {
	case "parse":
		return runParse(files, stdout, stderr)
	case "check":
		return runCheck(files, stdout, stderr)
	case "modelcheck", "mc":
		return runModelCheck(files, stdout, stderr)
	case "checkimplfile", "check-impl-file":
		return runCheckImplFile(files, stdout, stderr)
	case "repl-expr", "repl":
		return runREPLExpression(files, stdout, stderr)
	case "apalache-json":
		return runApalacheJSON(files, stdout, stderr)
	case "sany-xml":
		return runSanyXML(files, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", cmd)
		return ExitToolFailure
	}
}

func runSanyXML(args []string, stdout, stderr io.Writer) int {
	if cliArgsContainHelp(args) {
		printSanyXMLUsage(stdout)
		return ExitOK
	}
	opts, xmlOpts, err := parseSanyXMLCLIOptions(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitToolFailure
	}
	if len(opts.files) == 0 {
		fmt.Fprintln(stderr, "at least one file is required")
		return ExitToolFailure
	}
	exit := ExitOK
	for i, file := range opts.files {
		spec, diags := LoadSanySpec(file, opts.load)
		if diags.HasErrors() {
			writeDiagnostics(stderr, diags)
			exit = ExitSyntaxFailure
			continue
		}
		sem := CheckSpec(spec)
		writeDiagnostics(stderr, sem)
		if sem.HasErrors() {
			exit = ExitSemanticFailure
			continue
		}
		data, xmlDiags := SanyXMLWithOptions(spec, xmlOpts)
		writeDiagnostics(stderr, xmlDiags)
		if xmlDiags.HasErrors() {
			exit = ExitSemanticFailure
			continue
		}
		if i > 0 {
			fmt.Fprintln(stdout)
		}
		_, _ = stdout.Write(data)
		if len(data) == 0 || data[len(data)-1] != '\n' {
			fmt.Fprintln(stdout)
		}
	}
	return exit
}

func runREPLExpression(args []string, stdout, stderr io.Writer) int {
	opts := REPLEvalOptions{PrintOutput: stdout}
	var exprs []string
	for i := 0; i < len(args); i++ {
		loadOpts := LoadOptions{
			LibraryPaths:         opts.LibraryPaths,
			PreferLibraryModules: opts.PreferLibraryModules,
		}
		if ok, next, err := consumeLoadCLIOption(args, i, &loadOpts); ok || err != nil {
			if err != nil {
				fmt.Fprintln(stderr, err)
				return ExitToolFailure
			}
			opts.LibraryPaths = loadOpts.LibraryPaths
			opts.PreferLibraryModules = loadOpts.PreferLibraryModules
			i = next
			continue
		}
		switch args[i] {
		case "-spec", "--spec":
			i++
			if i >= len(args) {
				fmt.Fprintln(stderr, "-spec requires a TLA+ module file")
				return ExitToolFailure
			}
			opts.SpecFile = args[i]
		case "--":
			exprs = append(exprs, args[i+1:]...)
			i = len(args)
		default:
			exprs = append(exprs, args[i])
		}
	}
	if len(exprs) != 1 {
		fmt.Fprintln(stderr, "repl-expr requires exactly one expression argument")
		return ExitToolFailure
	}
	value, diags, err := EvaluateREPLExpression(exprs[0], opts)
	writeDiagnostics(stderr, diags)
	if diags.HasErrors() {
		return ExitSemanticFailure
	}
	if err != nil {
		fmt.Fprintf(stderr, "Error evaluating expression: '%s'\n%s\n", exprs[0], sanitizeREPLError(err.Error()))
		return ExitSemanticFailure
	}
	if value != "" {
		fmt.Fprintln(stdout, value)
	}
	return ExitOK
}

func sanitizeREPLError(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n", " ")
	return strings.TrimSpace(text)
}

func cliArgsContainHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			return false
		}
		if isCLIHelpFlag(arg) {
			return true
		}
	}
	return false
}

func printSanyXMLUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: tlago sany-xml [-o] [-t] [-r] [-u] [-I DIR] FILE...")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Parse, check, and export TLA+ modules as SANY XML.")
}

func parseSanyXMLCLIOptions(args []string) (commonCLIOptions, SanyXMLOptions, error) {
	var xmlOpts SanyXMLOptions
	commonArgs := make([]string, 0, len(args))
	for _, arg := range args {
		switch arg {
		case "-o", "--offline":
			xmlOpts.Offline = true
		case "-t", "--terse":
			xmlOpts.Terse = true
		case "-r", "--restricted":
			xmlOpts.Restricted = true
		case "-u", "--uncomment":
			xmlOpts.UncommentPreComments = true
		default:
			commonArgs = append(commonArgs, arg)
		}
	}
	opts, err := parseCommonCLIOptions(commonArgs, false)
	return opts, xmlOpts, err
}

func runApalacheJSON(args []string, stdout, stderr io.Writer) int {
	opts, err := parseCommonCLIOptions(args, false)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitToolFailure
	}
	if len(opts.files) == 0 {
		fmt.Fprintln(stderr, "at least one file is required")
		return ExitToolFailure
	}
	exit := ExitOK
	for i, file := range opts.files {
		spec, diags := LoadSanySpec(file, opts.load)
		if diags.HasErrors() {
			writeDiagnostics(stderr, diags)
			exit = ExitSyntaxFailure
			continue
		}
		sem := CheckSpec(spec)
		writeDiagnostics(stderr, sem)
		if sem.HasErrors() {
			exit = ExitSemanticFailure
			continue
		}
		data, irDiags := ApalacheIRJSON(spec, ApalacheIROptions{})
		writeDiagnostics(stderr, irDiags)
		if irDiags.HasErrors() {
			exit = ExitSemanticFailure
			continue
		}
		if i > 0 {
			fmt.Fprintln(stdout)
		}
		fmt.Fprintln(stdout, string(data))
	}
	return exit
}

func runParse(args []string, stdout, stderr io.Writer) int {
	opts, err := parseCommonCLIOptions(args, false)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitToolFailure
	}
	if len(opts.files) == 0 {
		fmt.Fprintln(stderr, "at least one file is required")
		return ExitToolFailure
	}
	exit := ExitOK
	for _, file := range opts.files {
		spec, diags := LoadSanySpec(file, opts.load)
		writeDiagnostics(stderr, diags)
		if diags.HasErrors() {
			exit = ExitSyntaxFailure
			continue
		}
		if spec.Root != nil {
			fmt.Fprintf(stdout, "Parsed module %s\n", spec.Root.Name)
		}
	}
	return exit
}

func runCheck(args []string, stdout, stderr io.Writer) int {
	opts, err := parseCommonCLIOptions(args, true)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitToolFailure
	}
	if len(opts.files) == 0 {
		fmt.Fprintln(stderr, "at least one file is required")
		return ExitToolFailure
	}
	exit := ExitOK
	for _, file := range opts.files {
		settings := defaultSanyDriverSettings()
		settings.messages = opts.diag
		spec, status := parseSanyWithSettings(file, opts.load, settings, func(_ Severity, message string) { fmt.Fprintln(stderr, message) })
		if status != 0 {
			if status == -1 {
				status = sanyExitCode(ExitToolFailure)
			}
			exit = int(status)
			continue
		}
		if spec != nil && spec.Root != nil {
			fmt.Fprintf(stdout, "Checked module %s\n", spec.Root.Name)
		}
	}
	return exit
}

type commonCLIOptions struct {
	load  LoadOptions
	diag  diagnosticCLIOptions
	files []string
}

type diagnosticCLIOptions struct {
	suppressed map[string]bool
	elevated   map[string]bool
}

func parseCommonCLIOptions(args []string, diagnostics bool) (commonCLIOptions, error) {
	opts := commonCLIOptions{}
	for i := 0; i < len(args); i++ {
		if ok, next, err := consumeLoadCLIOption(args, i, &opts.load); ok || err != nil {
			if err != nil {
				return opts, err
			}
			i = next
			continue
		}
		switch args[i] {
		case "-error-codes":
			if !diagnostics {
				opts.files = append(opts.files, args[i])
			}
			// check already exposes descriptive syntax/semantic exit statuses.
		case "-suppressMessages", "--suppressMessages", "-suppress-messages", "--suppress-messages":
			if !diagnostics {
				opts.files = append(opts.files, args[i])
				continue
			}
			i++
			if i >= len(args) {
				return opts, fmt.Errorf("%s requires a comma-separated diagnostic code list", args[i-1])
			}
			if err := addDiagnosticCodes(&opts.diag.suppressed, args[i]); err != nil {
				return opts, err
			}
		case "-messagesAsErrors", "--messagesAsErrors", "-messages-as-errors", "--messages-as-errors":
			if !diagnostics {
				opts.files = append(opts.files, args[i])
				continue
			}
			i++
			if i >= len(args) {
				return opts, fmt.Errorf("%s requires a comma-separated diagnostic code list", args[i-1])
			}
			if err := addDiagnosticCodes(&opts.diag.elevated, args[i]); err != nil {
				return opts, err
			}
		default:
			opts.files = append(opts.files, args[i])
		}
	}
	if diagnostics {
		if err := opts.diag.validate(); err != nil {
			return opts, err
		}
	}
	return opts, nil
}

func consumeLoadCLIOption(args []string, i int, opts *LoadOptions) (bool, int, error) {
	arg := args[i]
	switch {
	case arg == "-I" || arg == "--include" || arg == "-include" || arg == "--library-path":
		if i+1 >= len(args) {
			return true, i, fmt.Errorf("%s requires a directory", arg)
		}
		opts.LibraryPaths = append(opts.LibraryPaths, args[i+1])
		return true, i + 1, nil
	case strings.HasPrefix(arg, "-I="):
		dir := strings.TrimPrefix(arg, "-I=")
		if dir == "" {
			return true, i, fmt.Errorf("%s requires a directory", arg)
		}
		opts.LibraryPaths = append(opts.LibraryPaths, dir)
		return true, i, nil
	case strings.HasPrefix(arg, "--include="):
		dir := strings.TrimPrefix(arg, "--include=")
		if dir == "" {
			return true, i, fmt.Errorf("%s requires a directory", arg)
		}
		opts.LibraryPaths = append(opts.LibraryPaths, dir)
		return true, i, nil
	case strings.HasPrefix(arg, "--library-path="):
		dir := strings.TrimPrefix(arg, "--library-path=")
		if dir == "" {
			return true, i, fmt.Errorf("%s requires a directory", arg)
		}
		opts.LibraryPaths = append(opts.LibraryPaths, dir)
		return true, i, nil
	case strings.HasPrefix(arg, "-I") && len(arg) > len("-I"):
		opts.LibraryPaths = append(opts.LibraryPaths, strings.TrimPrefix(arg, "-I"))
		return true, i, nil
	case arg == "--prefer-library-modules" || arg == "-preferLibraryModules" || arg == "--preferLibraryModules":
		opts.PreferLibraryModules = true
		return true, i, nil
	default:
		return false, i, nil
	}
}

func addDiagnosticCodes(dst *map[string]bool, text string) error {
	if *dst == nil {
		*dst = map[string]bool{}
	}
	for _, part := range strings.Split(text, ",") {
		code := normalizeDiagnosticCode(part)
		if code == "" {
			return fmt.Errorf("empty diagnostic code in %q", text)
		}
		(*dst)[code] = true
	}
	return nil
}

func normalizeDiagnosticCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return ""
	}
	if strings.HasPrefix(code, "E") || strings.HasPrefix(code, "W") {
		return code
	}
	if _, ok := diagnosticCodeInfos["W"+code]; ok {
		return "W" + code
	}
	if _, ok := diagnosticCodeInfos["E"+code]; ok {
		return "E" + code
	}
	return "E" + code
}

func (opts diagnosticCLIOptions) validate() error {
	for code := range opts.suppressed {
		normalized, info, ok := lookupDiagnosticCode(code)
		if !ok {
			return fmt.Errorf("unknown message code %s", code)
		}
		if info.Severity != SeverityWarning {
			return fmt.Errorf("message code %s cannot be suppressed", normalized)
		}
	}
	for code := range opts.elevated {
		normalized, info, ok := lookupDiagnosticCode(code)
		if !ok {
			return fmt.Errorf("unknown message code %s", code)
		}
		if info.Severity != SeverityWarning {
			return fmt.Errorf("message code %s cannot be elevated with -messagesAsErrors", normalized)
		}
	}
	for code := range opts.suppressed {
		normalized := normalizeDiagnosticCode(code)
		if diagnosticCodeSetContains(opts.elevated, normalized) {
			return fmt.Errorf("codes were set to both -suppressMessages and -messagesAsErrors: %s", normalized)
		}
	}
	return nil
}

func (opts diagnosticCLIOptions) withTLCMessageControls() diagnosticCLIOptions {
	suppressed, elevated := tlcruntime.SANYMessageControls()
	for _, code := range suppressed {
		if opts.suppressed == nil {
			opts.suppressed = map[string]bool{}
		}
		opts.suppressed[normalizeDiagnosticCode(strconv.Itoa(code))] = true
	}
	for _, code := range elevated {
		if opts.elevated == nil {
			opts.elevated = map[string]bool{}
		}
		opts.elevated[normalizeDiagnosticCode(strconv.Itoa(code))] = true
	}
	return opts
}

func (opts diagnosticCLIOptions) apply(diags Diagnostics) Diagnostics {
	return DiagnosticOptions{
		SuppressedCodes: opts.suppressed,
		ElevatedCodes:   opts.elevated,
	}.Apply(diags)
}

func writeDiagnostics(w io.Writer, diags Diagnostics) {
	for _, d := range diags {
		fmt.Fprintln(w, d.String())
	}
}

func runModelCheck(args []string, stdout, stderr io.Writer) int {
	restoreStreams := tlcruntime.ToolIOSetSystemStreams(stdout, stderr)
	defer restoreStreams()
	tlcArgs, loadOpts, err := extractTLCLoadOptions(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitToolFailure
	}
	checker := tlcruntime.NewTLC(tlcruntime.Options{})
	if err := checker.HandleParameters(tlcArgs); err != nil {
		// HandleParameters reports Java's coded command-line diagnostic itself.
		if _, reported := err.(*tlcruntime.TLCCommandLineError); !reported {
			fmt.Fprintln(stderr, err)
		}
		return ExitToolFailure
	}
	opts := checker.Options
	var resolver tlcruntime.FilenameToStream
	if opts.PackagedModel != nil {
		classpath, err := tlcApplicationClasspath(opts.PackagedModel.Classpath())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return ExitToolFailure
		}
		resolver = tlcruntime.NewInJarFilenameToStream(tlcruntime.ModelInJarPath,
			tlcruntime.FilenameResolverOptions{Classpath: classpath})
	} else {
		classpath, err := tlcApplicationClasspath(nil)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return ExitToolFailure
		}
		var userDirectory *string
		if filepath.IsAbs(opts.SpecFile) {
			directory := filepath.Dir(opts.SpecFile)
			userDirectory = &directory
		}
		resolver = tlcruntime.NewSimpleFilenameToStream(loadOpts.LibraryPaths,
			tlcruntime.FilenameResolverOptions{Classpath: classpath, UserDirectory: userDirectory})
	}
	// FastTool/SpecProcessor construction belongs inside TLC.process, after
	// intern recovery and startup reporting. Ordinary and packaged models share
	// the same config-before-SANY loading and source exception boundaries.
	opts.LoadTool = func() (*tlcruntime.Tool, error) {
		tool, diags, err := loadTLCAppTool(opts.SpecFile, opts.ConfigFile, resolver, opts.RuntimeParams)
		if err == nil && diags.HasErrors() {
			return nil, diags
		}
		return tool, err
	}
	return runParsedTLCModelCheck(opts, stdout, stderr)
}

func runParsedTLCModelCheck(opts tlcruntime.Options, stdout, stderr io.Writer) int {
	result, err := tlcruntime.NewTLC(opts).Process(context.Background())
	if result == nil {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		return ExitToolFailure
	}
	// TLC.process has already reported checker and handled exception outcomes.
	// Java main maps that result directly to its exit status without another
	// error line or a second completion summary.
	return result.ExitStatus
}

func runCheckImplFile(args []string, stdout, stderr io.Writer) int {
	opts, err := tlcruntime.ParseCheckImplFileOptions(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitToolFailure
	}
	if opts.FromCheckpoint != "" {
		if err := tlcruntime.RecoverUniqueStrings(opts.FromCheckpoint); err != nil {
			fmt.Fprintln(stderr, err)
			return ExitToolFailure
		}
	}
	tlcruntime.FP64Init()
	spec, diags := LoadSanySpec(opts.MainFile, LoadOptions{})
	if diags.HasErrors() {
		writeDiagnostics(stderr, diags)
		return ExitSyntaxFailure
	}
	sem := CheckSpec(spec)
	if sem.HasErrors() {
		writeDiagnostics(stderr, sem)
		return ExitSemanticFailure
	}
	cfg, err := tlcruntime.ParseModelConfigFile(opts.ConfigFile)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitSyntaxFailure
	}
	tool, toolDiags := BuildTLCTool(spec, cfg, tlcruntime.RuntimeParameters{})
	if toolDiags.HasErrors() {
		writeDiagnostics(stderr, toolDiags)
		return ExitSemanticFailure
	}
	metadir, err := makeCheckImplFileMetaDir(opts.MainFile, opts.FromCheckpoint)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitToolFailure
	}
	fmt.Fprintln(stdout, "TLC CheckImpl"+tlcruntime.TLCVersion())
	checker := tlcruntime.NewCheckImplFile(tool, metadir, opts.Deadlock, opts.Depth, opts.FromCheckpoint, opts.TraceFile)
	checker.LoadTraceFunc = NewCheckImplFileTraceLoader(tool, LoadOptions{})
	if result, err := checker.Init(); err != nil || result != tlcruntime.NoError {
		if err != nil {
			fmt.Fprintln(stderr, err)
		} else {
			fmt.Fprintf(stderr, "CheckImplFile failed with error code %d\n", result)
		}
		return ExitSemanticFailure
	}
	for {
		if err := checker.Export(); err != nil {
			fmt.Fprintln(stderr, err)
			return ExitToolFailure
		}
		ok, err := checker.GetTrace()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return ExitSemanticFailure
		}
		if ok {
			if err := checker.CheckTrace(); err != nil {
				fmt.Fprintln(stderr, err)
				return ExitSemanticFailure
			}
			continue
		}
		time.Sleep(10 * time.Second)
	}
}

func makeCheckImplFileMetaDir(mainFile string, fromCheckpoint string) (string, error) {
	if fromCheckpoint != "" {
		return fromCheckpoint, nil
	}
	specDir := ""
	if filepath.IsAbs(mainFile) {
		specDir = filepath.Dir(mainFile)
	}
	root := filepath.Join(specDir, tlcruntime.MetaRoot)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	name := time.Now().Format("06-01-02-15-04-05.000")
	path := filepath.Join(root, name)
	if err := os.Mkdir(path, 0o755); err == nil {
		return path, nil
	} else if !os.IsExist(err) {
		return "", err
	}
	return os.MkdirTemp(root, name)
}

func tlcRuntimeParameterModules(params tlcruntime.RuntimeParameters) []string {
	return params.ExtendeeModules()
}

func extractTLCLoadOptions(args []string) ([]string, LoadOptions, error) {
	loadOpts := LoadOptions{}
	tlcArgs := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if ok, next, err := consumeLoadCLIOption(args, i, &loadOpts); ok || err != nil {
			if err != nil {
				return nil, loadOpts, err
			}
			i = next
			continue
		}
		// TLC owns validation of both TLC and SANY message codes.
		tlcArgs = append(tlcArgs, args[i])
	}
	return tlcArgs, loadOpts, nil
}
