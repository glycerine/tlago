package tlago

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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
		fmt.Fprintln(stderr, "usage: tlago parse|check|modelcheck|checkimplfile|repl-expr|apalache-json|sany-xml [-I DIR] FILE...")
		return ExitToolFailure
	}
	cmd := args[0]
	files := args[1:]
	if len(files) == 0 {
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
		switch arg {
		case "-help", "--help", "-h":
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
		spec, diags := LoadSanySpec(file, opts.load)
		diags = opts.diag.apply(diags)
		if diags.HasErrors() {
			writeDiagnostics(stderr, diags)
			exit = ExitSyntaxFailure
			continue
		}
		sem := CheckSpec(spec)
		sem = opts.diag.apply(sem)
		writeDiagnostics(stderr, sem)
		if sem.HasErrors() {
			exit = ExitSemanticFailure
			continue
		}
		if spec.Root != nil {
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
			return fmt.Errorf("message code %s cannot be configured in both -suppressMessages and -messagesAsErrors", normalized)
		}
	}
	return nil
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
	if tlcArgs, ok := stripTLCModelCheckFlag(args); ok {
		return runTLCModelCheck(tlcArgs, stdout, stderr)
	}
	cfgPath := ""
	opts := ModelCheckOptions{}
	loadOpts := LoadOptions{}
	var files []string
	for i := 0; i < len(args); i++ {
		if ok, next, err := consumeLoadCLIOption(args, i, &loadOpts); ok || err != nil {
			if err != nil {
				fmt.Fprintln(stderr, err)
				return ExitToolFailure
			}
			i = next
			continue
		}
		switch args[i] {
		case "-config", "--config":
			i++
			if i >= len(args) {
				fmt.Fprintln(stderr, "-config requires a file")
				return ExitToolFailure
			}
			cfgPath = args[i]
		case "-maxStates", "--maxStates", "-max-states", "--max-states":
			i++
			if i >= len(args) {
				fmt.Fprintln(stderr, "-maxStates requires a positive integer")
				return ExitToolFailure
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n <= 0 {
				fmt.Fprintf(stderr, "-maxStates requires a positive integer, got %q\n", args[i])
				return ExitToolFailure
			}
			opts.MaxStates = n
		default:
			files = append(files, args[i])
		}
	}
	if len(files) != 1 {
		fmt.Fprintln(stderr, "modelcheck requires exactly one spec file")
		return ExitToolFailure
	}
	specPath := files[0]
	if cfgPath == "" {
		cfgPath = strings.TrimSuffix(specPath, filepath.Ext(specPath)) + ".cfg"
	}
	spec, diags := LoadSanySpec(specPath, loadOpts)
	if diags.HasErrors() {
		writeDiagnostics(stderr, diags)
		return ExitSyntaxFailure
	}
	sem := CheckSpec(spec)
	if sem.HasErrors() {
		writeDiagnostics(stderr, sem)
		return ExitSemanticFailure
	}
	cfgData, err := os.ReadFile(cfgPath)
	if err != nil {
		fmt.Fprintf(stderr, "cannot read config %s: %v\n", cfgPath, err)
		return ExitSyntaxFailure
	}
	cfg, cfgDiags := ParseConfigSource(cfgPath, string(cfgData))
	if cfgDiags.HasErrors() {
		writeDiagnostics(stderr, cfgDiags)
		return ExitSyntaxFailure
	}
	result, runDiags := ModelCheck(spec, cfg, opts)
	if runDiags.HasErrors() {
		writeDiagnostics(stderr, runDiags)
		return ExitSemanticFailure
	}
	if !result.OK {
		fmt.Fprintf(stderr, "%s\n", result.Error)
		for i, st := range result.Trace {
			fmt.Fprintf(stderr, "%d: %s\n", i+1, formatState(st))
		}
		return ExitSemanticFailure
	}
	fmt.Fprintf(stdout, "Model checking completed: %d states explored\n", result.StatesExplored)
	return ExitOK
}

func stripTLCModelCheckFlag(args []string) ([]string, bool) {
	out := make([]string, 0, len(args))
	found := false
	for _, arg := range args {
		switch arg {
		case "-tlc", "--tlc", "-go-tlc", "--go-tlc":
			found = true
		default:
			out = append(out, arg)
		}
	}
	return out, found
}

func runTLCModelCheck(args []string, stdout, stderr io.Writer) int {
	tlcArgs, loadOpts, diagOpts, err := extractTLCLoadOptions(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitToolFailure
	}
	opts, err := tlcruntime.ParseTLCOptions(tlcArgs)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitToolFailure
	}
	loadOpts.ExtraModules = appendModuleNames(loadOpts.ExtraModules, tlcRuntimeParameterModules(opts.RuntimeParams)...)
	spec, diags := LoadSanySpec(opts.SpecFile, loadOpts)
	diags = diagOpts.apply(diags)
	if diags.HasErrors() {
		writeDiagnostics(stderr, diags)
		return ExitSyntaxFailure
	}
	sem := CheckSpec(spec)
	sem = diagOpts.apply(sem)
	if sem.HasErrors() {
		writeDiagnostics(stderr, sem)
		return ExitSemanticFailure
	}
	cfg, err := tlcruntime.ParseModelConfigFile(opts.ConfigFile)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitSyntaxFailure
	}
	tool, toolDiags := BuildTLCTool(spec, cfg, opts.RuntimeParams)
	toolDiags = diagOpts.apply(toolDiags)
	if toolDiags.HasErrors() {
		writeDiagnostics(stderr, toolDiags)
		return ExitSemanticFailure
	}
	opts.Tool = tool
	if cfg.GetCheckDeadlock() {
		opts.Deadlock = true
	} else {
		opts.NoDeadlock = true
		opts.Deadlock = false
	}
	result, err := tlcruntime.NewTLC(opts).Process(context.Background())
	if err != nil {
		fmt.Fprintln(stderr, err)
	}
	if result == nil {
		return ExitToolFailure
	}
	if result.ErrorCode != tlcruntime.NoError {
		if err == nil {
			fmt.Fprintf(stderr, "TLC failed with error code %d\n", result.ErrorCode)
		}
		return ExitSemanticFailure
	}
	fmt.Fprintf(stdout, "TLC model checking completed: %d states generated, %d distinct states\n", result.StatesGenerated, result.DistinctStates)
	return ExitOK
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

func extractTLCLoadOptions(args []string) ([]string, LoadOptions, diagnosticCLIOptions, error) {
	loadOpts := LoadOptions{}
	diagOpts := diagnosticCLIOptions{}
	tlcArgs := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if ok, next, err := consumeLoadCLIOption(args, i, &loadOpts); ok || err != nil {
			if err != nil {
				return nil, loadOpts, diagOpts, err
			}
			i = next
			continue
		}
		switch args[i] {
		case "-suppressMessages":
			tlcArgs = append(tlcArgs, args[i])
			i++
			if i >= len(args) {
				return nil, loadOpts, diagOpts, fmt.Errorf("%s requires a comma-separated diagnostic code list", args[i-1])
			}
			tlcArgs = append(tlcArgs, args[i])
			if err := addKnownSANYDiagnosticCodes(&diagOpts.suppressed, args[i], true); err != nil {
				return nil, loadOpts, diagOpts, err
			}
			continue
		case "-messagesAsErrors":
			tlcArgs = append(tlcArgs, args[i])
			i++
			if i >= len(args) {
				return nil, loadOpts, diagOpts, fmt.Errorf("%s requires a comma-separated diagnostic code list", args[i-1])
			}
			tlcArgs = append(tlcArgs, args[i])
			if err := addKnownSANYDiagnosticCodes(&diagOpts.elevated, args[i], false); err != nil {
				return nil, loadOpts, diagOpts, err
			}
			continue
		}
		tlcArgs = append(tlcArgs, args[i])
	}
	if err := validateKnownSANYDiagnosticOverlap(diagOpts); err != nil {
		return nil, loadOpts, diagOpts, err
	}
	return tlcArgs, loadOpts, diagOpts, nil
}

func addKnownSANYDiagnosticCodes(dst *map[string]bool, text string, suppress bool) error {
	for _, part := range strings.Split(text, ",") {
		code := normalizeDiagnosticCode(part)
		if code == "" {
			return fmt.Errorf("empty diagnostic code in %q", text)
		}
		normalized, info, ok := lookupDiagnosticCode(code)
		if !ok {
			continue
		}
		if suppress && info.Severity != SeverityWarning {
			return fmt.Errorf("message code %s cannot be suppressed", normalized)
		}
		if *dst == nil {
			*dst = map[string]bool{}
		}
		(*dst)[normalized] = true
	}
	return nil
}

func validateKnownSANYDiagnosticOverlap(opts diagnosticCLIOptions) error {
	for code := range opts.suppressed {
		if diagnosticCodeSetContains(opts.elevated, code) {
			return fmt.Errorf("message code %s cannot be configured in both -suppressMessages and -messagesAsErrors", code)
		}
	}
	return nil
}

func formatState(st State) string {
	keys := make([]string, 0, len(st))
	for key := range st {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s = %d", key, st[key]))
	}
	return strings.Join(parts, ", ")
}
