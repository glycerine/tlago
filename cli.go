package tlago

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func RunCLI(args []string, stdout, stderr io.Writer) int {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: tlago parse|check|modelcheck|apalache-json|sany-xml [-I DIR] FILE...")
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
		data, xmlDiags := SanyXML(spec)
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
	if code[0] == '4' {
		return "W" + code
	}
	return "E" + code
}

func (opts diagnosticCLIOptions) apply(diags Diagnostics) Diagnostics {
	if len(diags) == 0 {
		return diags
	}
	out := make(Diagnostics, 0, len(diags))
	for _, diag := range diags {
		code := strings.ToUpper(diag.Code)
		if opts.suppressed[code] {
			continue
		}
		if diag.Severity == SeverityWarning && opts.elevated[code] {
			diag.Severity = SeverityError
			diag.Message = "Warning treated as error: " + diag.Message
		}
		out = append(out, diag)
	}
	return out
}

func writeDiagnostics(w io.Writer, diags Diagnostics) {
	for _, d := range diags {
		fmt.Fprintln(w, d.String())
	}
}

func runModelCheck(args []string, stdout, stderr io.Writer) int {
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
