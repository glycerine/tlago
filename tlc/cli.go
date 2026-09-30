package tlc

import (
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type RuntimeParameters struct {
	Invariants             []RuntimeInvariantTemplate
	Constraints            []RuntimeConstraint
	ActionConstraints      []RuntimeConstraint
	PostConditions         []RuntimePostCondition
	View                   *RuntimeView
	CustomStateWriterClass string
	UserFile               string
}

type RuntimeInvariantTemplate struct {
	Modules    []string
	Expression string
}

type RuntimeConstraint struct {
	Module       string
	Operator     string
	ConstantName string
	FileName     string
}

type RuntimePostCondition struct {
	Module       string
	Operator     string
	ConstantName string
	FileName     string
}

type RuntimeView struct {
	Module   string
	Operator string
}

type TLCCommandLineError struct {
	Message string
	Usage   bool
}

func (e *TLCCommandLineError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func ParseTLCOptions(args []string) (Options, error) {
	opts := defaultTLCCommandLineOptions()
	dump := tlcDumpOption{}
	generateTESpec := true
	generateTESpecBinaryTrace := true
	forceGenerateTESpec := false
	teSpecMonolith := true
	teSpecOut := ""

	tlcSuppressedCodes := NewInsMap[int, bool]()
	tlcMessagesAsErrors := NewInsMap[int, bool]()

	index := 0
	for index < len(args) {
		arg := args[index]
		switch {
		case arg == "-simulate" || arg == "-generate":
			if arg == "-generate" {
				opts.Probabilistic = true
			}
			opts.Mode = RunModeSimulate
			index++
			if index < len(args) && isSimulationSubargument(args[index]) {
				for _, simArg := range strings.Split(args[index], ",") {
					switch {
					case strings.HasPrefix(simArg, "num="):
						traceNum, err := strconv.ParseInt(strings.TrimPrefix(simArg, "num="), 10, 64)
						if err != nil {
							return opts, tlcCommandLineError("Error: An integer for simulation trace count required. But encountered " + simArg)
						}
						opts.TraceNum = traceNum
					case strings.HasPrefix(simArg, "file="):
						opts.TraceFile = strings.TrimPrefix(simArg, "file=")
					case simArg == "stats=basic":
						opts.TraceActions = "BASIC"
					case simArg == "stats=full":
						opts.TraceActions = "FULL"
					case simArg == "sched=rl":
						opts.SimulationSchedule = SimulationScheduleRL
					case simArg == "sched=rlaction":
						opts.SimulationSchedule = SimulationScheduleRLAction
					}
				}
				index++
				if opts.TraceNum == math.MaxInt64 && opts.TraceFile != "" {
					return opts, tlcCommandLineError("Error: You need to specify a 'num' argument when using the 'file' argument, for example: '\"num=5,file=test.txt\"'.")
				}
			}
		case arg == "-modelcheck":
			index++
		case arg == "-difftrace":
			Globals.PrintDiffsOnly = true
			index++
		case arg == "-deadlock":
			opts.NoDeadlock = true
			opts.Deadlock = false
			index++
		case arg == "-cleanup":
			opts.Cleanup = true
			index++
		case arg == "-nowarning":
			Globals.Warn = false
			index++
		case arg == "-suppressMessages":
			var err error
			index, err = parseTLCMessageCodeList(args, index, tlcSuppressedCodes, "-suppressMessages")
			if err != nil {
				return opts, err
			}
		case arg == "-messagesAsErrors":
			var err error
			index, err = parseTLCMessageCodeList(args, index, tlcMessagesAsErrors, "-messagesAsErrors")
			if err != nil {
				return opts, err
			}
		case arg == "-gzip":
			Globals.UseGZIP = true
			index++
		case arg == "-terse":
			Globals.Expand = false
			index++
		case arg == "-continue":
			Globals.Continuation = true
			index++
		case arg == "-view":
			Globals.UseView = true
			index++
		case arg == "-debug":
			Globals.Debug = true
			index++
		case arg == "-inv":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: invariant expression required.")
			}
			expr := args[index+1]
			opts.RuntimeParams.Invariants = append(opts.RuntimeParams.Invariants, RuntimeInvariantTemplate{
				Modules:    runtimeInvariantModules(expr),
				Expression: expr,
			})
			index += 2
		case arg == "-invlevel":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: An integer for -invlevel required.")
			}
			if _, err := strconv.Atoi(args[index+1]); err != nil {
				return opts, tlcCommandLineError("Error: An integer for -invlevel required. But encountered " + args[index+1])
			}
			opts.RuntimeParams.Invariants = append(opts.RuntimeParams.Invariants, RuntimeInvariantTemplate{
				Modules:    []string{"TLC", "Naturals"},
				Expression: fmt.Sprintf("LET _T == INSTANCE TLC _N == INSTANCE Naturals IN _N!<(_T!TLCGet(\"level\"), %s)", args[index+1]),
			})
			index += 2
		case arg == "-debugger":
			opts.DebugPort = 4712
			opts.DebugSuspend = true
			opts.DebugHalt = true
			index++
			if index < len(args) && isDebuggerSubargument(args[index]) {
				sub := strings.ToLower(args[index])
				opts.DebugSuspend = !strings.Contains(sub, "nosuspend")
				opts.DebugHalt = !strings.Contains(sub, "nohalt")
				if match := regexp.MustCompile(`.*port=([0-9]{1,5}).*`).FindStringSubmatch(args[index]); len(match) == 2 {
					port, err := strconv.Atoi(match[1])
					if err != nil {
						return opts, tlcCommandLineError("Error: debugger port must be an integer. But encountered " + match[1])
					}
					opts.DebugPort = port
				}
				index++
			}
		case arg == "-tool":
			opts.ToolMode = true
			Globals.Tool = true
			index++
		case arg == "-generateSpecTE":
			forceGenerateTESpec = true
			opts.ToolMode = true
			Globals.Tool = true
			index++
			if index < len(args) && args[index] == "nomonolith" {
				teSpecMonolith = false
				index++
			}
		case arg == "-noGenerateSpecTE" || strings.EqualFold(arg, "-noTE"):
			generateTESpec = false
			index++
		case arg == "-noGenerateSpecTEBin" || strings.EqualFold(arg, "-noTEBin"):
			generateTESpecBinaryTrace = false
			index++
		case arg == "-teSpecOutDir":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: expected a path for -teSpecOutDir option.")
			}
			teSpecOut = args[index+1]
			index += 2
		case arg == "-help" || arg == "-h":
			return opts, &TLCCommandLineError{Message: "usage requested", Usage: true}
		case arg == "-lncheck":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: expect a strategy such as final for -lncheck option.")
			}
			Globals.LNCheck = strings.ToLower(args[index+1])
			index += 2
		case arg == "-config":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: expect a file name for -config option.")
			}
			opts.ConfigFile = trimConfigExtension(args[index+1])
			index += 2
		case arg == "-dump":
			next, parsed, err := parseTLCDumpOption(args, index)
			if err != nil {
				return opts, err
			}
			dump = parsed
			index = next
		case strings.EqualFold(arg, "-loadTrace"):
			if index+2 >= len(args) {
				return opts, tlcCommandLineError("Error: A format and a file name for loading traces required.")
			}
			format := args[index+1]
			fileName := args[index+2]
			switch {
			case strings.EqualFold(format, "tlc"):
				opts.RuntimeParams.Constraints = append(opts.RuntimeParams.Constraints, RuntimeConstraint{Module: "_TLCTrace", Operator: "_TLCTraceConstraint", ConstantName: "_TLCTraceInputFile", FileName: fileName})
				opts.RuntimeParams.View = &RuntimeView{Module: "_TLCTrace", Operator: "_TLCTraceView"}
			case strings.EqualFold(format, "json"):
				opts.RuntimeParams.Constraints = append(opts.RuntimeParams.Constraints, RuntimeConstraint{Module: "_JsonTrace", Operator: "_JsonTraceConstraint", ConstantName: "_JsonTraceInputFile", FileName: fileName})
				opts.RuntimeParams.View = &RuntimeView{Module: "_JsonTrace", Operator: "_JsonTraceView"}
			default:
				return opts, tlcCommandLineError("Error: Unknown format " + format + " given to -loadTrace.")
			}
			index += 3
		case strings.EqualFold(arg, "-dumpTrace"):
			if index+2 >= len(args) {
				return opts, tlcCommandLineError("Error: A format and a file name for dumping traces required.")
			}
			post, err := runtimePostConditionForTraceFormat(args[index+1], args[index+2])
			if err != nil {
				return opts, err
			}
			opts.RuntimeParams.PostConditions = append(opts.RuntimeParams.PostConditions, post)
			index += 3
		case strings.EqualFold(arg, "-postCondition"):
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: Module!Operator for postCondition required.")
			}
			post, err := runtimePostConditionFromModuleBangOperator(args[index+1])
			if err != nil {
				return opts, err
			}
			opts.RuntimeParams.PostConditions = append(opts.RuntimeParams.PostConditions, post)
			index += 2
		case arg == "-coverage":
			value, err := parseNonnegativeIntOption(args, index, "coverage report interval", "-coverage")
			if err != nil {
				return opts, err
			}
			Globals.CoverageInterval = value * 60 * 1000
			index += 2
		case arg == "-checkpoint":
			value, err := parseNonnegativeIntOption(args, index, "checkpoint interval", "-checkpoint")
			if err != nil {
				return opts, err
			}
			opts.CheckpointDurationMillis = int64(value) * 60 * 1000
			Globals.CheckpointDurationMillis = opts.CheckpointDurationMillis
			index += 2
		case arg == "-depth":
			value, err := parseIntOption(args, index, "trace length", "-depth")
			if err != nil {
				return opts, err
			}
			opts.TraceDepth = value
			index += 2
		case arg == "-seed":
			value, err := parseInt64Option(args, index, "seed", "-seed")
			if err != nil {
				return opts, err
			}
			opts.Seed = value
			opts.NoSeed = false
			index += 2
		case arg == "-aril":
			value, err := parseInt64Option(args, index, "aril", "-aril")
			if err != nil {
				return opts, err
			}
			opts.Aril = value
			index += 2
		case arg == "-maxSetSize":
			value, err := parseNonnegativeIntOption(args, index, "maxSetSize", "-maxSetSize")
			if err != nil {
				return opts, err
			}
			if !IsValidSetSize(value) {
				return opts, tlcCommandLineError("Error: Value in interval [0, 2147483647] for maxSetSize required. But encountered " + args[index+1])
			}
			Globals.SetBound = value
			index += 2
		case arg == "-recover":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: need to specify the metadata directory for recovery.")
			}
			opts.FromCheckpoint = cleanPathWithSeparator(args[index+1])
			index += 2
		case arg == "-metadir":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: need to specify the metadata directory.")
			}
			opts.MetaDir = cleanPathWithSeparator(args[index+1])
			Globals.MetaDir = opts.MetaDir
			index += 2
		case arg == "-userFile":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: need to specify the full qualified file.")
			}
			opts.RuntimeParams.UserFile = args[index+1]
			index += 2
		case arg == "-workers":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: expect an integer or 'auto' for -workers option.")
			}
			workers, err := parseWorkerCount(args[index+1])
			if err != nil {
				return opts, err
			}
			opts.Workers = workers
			SetNumWorkers(workers)
			index += 2
		case arg == "-dfid":
			value, err := parseNonnegativeIntOption(args, index, "dfid", "-dfid")
			if err != nil {
				return opts, err
			}
			opts.DFIDDepth = value
			Globals.DFIDMax = value
			index += 2
		case arg == "-fp":
			value, err := parseNonnegativeIntOption(args, index, "fp", "-fp")
			if err != nil {
				return opts, err
			}
			if value >= len(FP64Polys) {
				return opts, tlcCommandLineError(fmt.Sprintf("Error: The number for -fp must be between 0 and %d (inclusive).", len(FP64Polys)-1))
			}
			opts.FPIndex = value
			index += 2
		case arg == "-fpmem":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: fpset memory size required.")
			}
			fpMemSize, err := strconv.ParseFloat(args[index+1], 64)
			if err != nil || fpMemSize < 0 {
				return opts, tlcCommandLineError("Error: An positive integer or a fraction for fpset memory size/percentage required. But encountered " + args[index+1])
			}
			if fpMemSize > 1 {
				opts.FPSetConfiguration.SetMemory(int64(fpMemSize))
				opts.FPSetConfiguration.SetRatio(1)
			} else {
				opts.FPSetConfiguration.SetRatio(fpMemSize)
			}
			index += 2
		case arg == "-fpbits":
			value, err := parseNonnegativeIntOption(args, index, "fpbits", "-fpbits")
			if err != nil {
				return opts, err
			}
			if !IsValidFPBits(value) {
				return opts, tlcCommandLineError("Error: Value in interval [0, 30] for fpbits required. But encountered " + args[index+1])
			}
			opts.FPSetConfiguration.SetFPBits(value)
			index += 2
		default:
			if strings.HasPrefix(arg, "-") {
				return opts, tlcCommandLineError("Error: unrecognized option: " + arg)
			}
			if opts.SpecFile != "" {
				return opts, tlcCommandLineError("Error: more than one input files: " + opts.SpecFile + " and " + arg)
			}
			opts.SpecFile = trimTLAExtension(arg)
			index++
		}
	}

	if !Globals.Warn && tlcSuppressedCodes.Len() > 0 {
		return opts, tlcCommandLineError("Error: -nowarning already suppresses all warnings, making -suppressMessages redundant. Remove -nowarning to suppress only the specified messages, or remove -suppressMessages to suppress all warnings.")
	}
	if overlap := messageCodeOverlap(tlcSuppressedCodes, tlcMessagesAsErrors); overlap != "" {
		return opts, tlcCommandLineError("Error: The following codes were set to both -suppressMessages and -messagesAsErrors: " + overlap)
	}
	for code := range tlcSuppressedCodes.All() {
		SuppressTLCMessage(code)
	}
	for code := range tlcMessagesAsErrors.All() {
		TreatTLCMessageAsError(code)
	}

	if opts.SpecFile == "" {
		return opts, tlcCommandLineError("Error: Missing input TLA+ module.")
	}
	if opts.ConfigFile == "" {
		opts.ConfigFile = opts.SpecFile
	}
	if opts.MetaDir == "" {
		opts.MetaDir = "states"
	}
	if dump.File != "" {
		writer, err := dump.newStateWriter(opts.MetaDir)
		if err != nil {
			return opts, tlcCommandLineError(fmt.Sprintf("Error: Given file name %s for dumping states invalid.", dump.File))
		}
		opts.StateWriter = writer
	} else if dump.CustomClass != "" {
		opts.RuntimeParams.CustomStateWriterClass = dump.CustomClass
		opts.StateWriter = NewNoopStateWriter()
	}

	opts.ForceGenerateTraceSpec = forceGenerateTESpec
	opts.GenerateTraceSpec = forceGenerateTESpec || (generateTESpec && !Globals.Tool && !Globals.Continuation && !IsTraceExplorationSpecFile(opts.SpecFile))
	opts.GenerateTraceSpecBinary = generateTESpecBinaryTrace
	opts.GenerateTraceSpecMonolith = teSpecMonolith
	opts.TraceSpecOutputDir = teSpecOut
	opts.StartTime = time.Now()
	return opts, nil
}

func (t *TLC) HandleParameters(args []string) error {
	opts, err := ParseTLCOptions(args)
	if err != nil {
		return err
	}
	if t == nil {
		return nil
	}
	if t.Tool != nil && opts.Tool == nil {
		opts.Tool = t.Tool
	}
	t.Options = opts
	return nil
}

func defaultTLCCommandLineOptions() Options {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return Options{
		Mode:                      RunModeModelCheck,
		Deadlock:                  true,
		NoSeed:                    true,
		TraceNum:                  math.MaxInt64,
		TraceDepth:                100,
		Workers:                   1,
		FPIndex:                   rng.Intn(len(FP64Polys)),
		FPSetConfiguration:        NewFPSetConfiguration(),
		DebugPort:                 -1,
		DebugSuspend:              true,
		DebugHalt:                 true,
		GenerateTraceSpecBinary:   true,
		GenerateTraceSpecMonolith: true,
	}
}

func tlcCommandLineError(msg string) error {
	return &TLCCommandLineError{Message: msg}
}

func isSimulationSubargument(arg string) bool {
	return strings.Contains(arg, "stats=") || strings.Contains(arg, "file=") || strings.Contains(arg, "num=") || strings.Contains(arg, "sched")
}

func isDebuggerSubargument(arg string) bool {
	lower := strings.ToLower(arg)
	return strings.Contains(lower, "port=") || strings.Contains(lower, "nosuspend") || strings.Contains(lower, "nohalt") || strings.Contains(lower, "suspend") || strings.Contains(lower, "halt")
}

func parseTLCMessageCodeList(args []string, index int, dst *InsMap[int, bool], option string) (int, error) {
	if index+1 >= len(args) {
		return index, tlcCommandLineError("Error: " + option + " requires a comma-separated list of message codes.")
	}
	for _, part := range strings.Split(args[index+1], ",") {
		code, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return index, tlcCommandLineError("Error: unknown message code: " + strings.TrimSpace(part))
		}
		dst.Set(code, true)
	}
	return index + 2, nil
}

func messageCodeOverlap(a *InsMap[int, bool], b *InsMap[int, bool]) string {
	if a == nil || b == nil {
		return ""
	}
	var parts []string
	for code := range a.All() {
		if _, ok := b.Get2(code); ok {
			parts = append(parts, strconv.Itoa(code))
		}
	}
	return strings.Join(parts, ", ")
}

func runtimeInvariantModules(expr string) []string {
	re := regexp.MustCompile(`INSTANCE\s+(\w+)`)
	seen := NewInsMap[string, bool]()
	for _, match := range re.FindAllStringSubmatch(expr, -1) {
		if len(match) == 2 {
			seen.Set(match[1], true)
		}
	}
	modules := make([]string, 0, seen.Len())
	for module := range seen.All() {
		modules = append(modules, module)
	}
	return modules
}

type tlcDumpOption struct {
	File         string
	AsDot        bool
	Colorize     bool
	ActionLabels bool
	Snapshot     bool
	Constrained  bool
	Stuttering   bool
	Strict       bool
	CustomClass  string
}

func parseTLCDumpOption(args []string, index int) (int, tlcDumpOption, error) {
	index++
	if index >= len(args) {
		return index, tlcDumpOption{}, tlcCommandLineError("Error: A file name for dumping states required.")
	}
	if strings.HasPrefix(args[index], "class,") {
		return index + 1, tlcDumpOption{CustomClass: strings.TrimPrefix(args[index], "class,")}, nil
	}
	if index+1 < len(args) && strings.HasPrefix(args[index], "dot") {
		dotArgs := strings.ToLower(args[index])
		return index + 2, tlcDumpOption{
			File:         requireDumpSuffix(args[index+1], ".dot"),
			AsDot:        true,
			Colorize:     strings.Contains(dotArgs, "colorize"),
			ActionLabels: strings.Contains(dotArgs, "actionlabels"),
			Snapshot:     strings.Contains(dotArgs, "snapshot"),
			Constrained:  strings.Contains(dotArgs, "constrained"),
			Stuttering:   strings.Contains(dotArgs, "stuttering"),
			Strict:       strings.Contains(dotArgs, "strict"),
		}, nil
	}
	return index + 1, tlcDumpOption{File: requireDumpSuffix(args[index], ".dump")}, nil
}

func (d tlcDumpOption) newStateWriter(metadir string) (*StateWriter, error) {
	if d.CustomClass != "" {
		return &StateWriter{Noop: true}, nil
	}
	file := strings.Replace(d.File, "${metadir}", metadir, 1)
	if d.AsDot {
		return NewDotStateWriter(file, DotStateWriterOptions{
			Colorize:     d.Colorize,
			ActionLabels: d.ActionLabels,
			Snapshot:     d.Snapshot,
			Constrained:  d.Constrained,
			Stuttering:   d.Stuttering,
			Strict:       d.Strict,
		})
	}
	return NewStateDumpWriter(file)
}

func runtimePostConditionForTraceFormat(format string, fileName string) (RuntimePostCondition, error) {
	switch {
	case strings.EqualFold(format, "json"):
		return RuntimePostCondition{Module: "_JsonTrace", Operator: "_JsonTrace", ConstantName: "_JsonTraceFile", FileName: fileName}, nil
	case strings.EqualFold(format, "tla"):
		return RuntimePostCondition{Module: "_TLAPlusCounterExample", Operator: "_TLAPlusCounterExample", ConstantName: "_TLAPlusCounterExampleFile", FileName: fileName}, nil
	case strings.EqualFold(format, "tlc"):
		return RuntimePostCondition{Module: "_TLCTrace", Operator: "_TLCTrace", ConstantName: "_TLCTraceFile", FileName: fileName}, nil
	case strings.EqualFold(format, "tlcplain"):
		return RuntimePostCondition{Module: "_TLCTracePlain", Operator: "_TLCTrace", ConstantName: "_TLCTraceFile", FileName: fileName}, nil
	case strings.EqualFold(format, "tlcTESpec"):
		return RuntimePostCondition{Module: "_TLCTESpec", Operator: "_TLCTrace", ConstantName: "_TLCTraceFile", FileName: fileName}, nil
	case strings.EqualFold(format, "tlcaction"):
		return RuntimePostCondition{Module: "_TLCActionTrace", Operator: "_TLCTrace", ConstantName: "_TLCTraceFile", FileName: fileName}, nil
	case strings.EqualFold(format, "dot"):
		return RuntimePostCondition{Module: "_DotTrace", Operator: "_DotTrace", ConstantName: "_DotTraceFile", FileName: fileName}, nil
	default:
		return RuntimePostCondition{}, tlcCommandLineError("Error: Unknown format " + format + " given to -dumpTrace.")
	}
}

func runtimePostConditionFromModuleBangOperator(moduleBangOp string) (RuntimePostCondition, error) {
	parts := strings.Split(moduleBangOp, "!")
	if len(parts) != 2 {
		return RuntimePostCondition{}, tlcCommandLineError("Module!Operator for postCondition must be of the form module!operator. Encountered: " + moduleBangOp)
	}
	module := parts[0]
	def := parts[1]
	if module == "" {
		return RuntimePostCondition{}, tlcCommandLineError("Module name for postCondition required. Encountered: " + moduleBangOp)
	}
	if strings.Contains(module, ".") {
		return RuntimePostCondition{}, tlcCommandLineError("Module name for postCondition must not contain a dot. Encountered: " + module)
	}
	if def == "" {
		return RuntimePostCondition{}, tlcCommandLineError("Operator name for postCondition required. Encountered: " + moduleBangOp)
	}
	if strings.Contains(def, ".") {
		return RuntimePostCondition{}, tlcCommandLineError("Operator name for postCondition must not contain a dot. Encountered: " + def)
	}
	return RuntimePostCondition{Module: module, Operator: def}, nil
}

func parseIntOption(args []string, index int, name string, option string) (int, error) {
	if index+1 >= len(args) {
		return 0, tlcCommandLineError("Error: " + name + " required.")
	}
	value, err := strconv.Atoi(args[index+1])
	if err != nil {
		return 0, tlcCommandLineError("Error: An integer for " + name + " required. But encountered " + args[index+1])
	}
	return value, nil
}

func parseNonnegativeIntOption(args []string, index int, name string, option string) (int, error) {
	value, err := parseIntOption(args, index, name, option)
	if err != nil {
		return 0, err
	}
	if value < 0 {
		return 0, tlcCommandLineError("Error: expect a nonnegative integer for " + option + " option.")
	}
	return value, nil
}

func parseInt64Option(args []string, index int, name string, option string) (int64, error) {
	if index+1 >= len(args) {
		return 0, tlcCommandLineError("Error: " + name + " required.")
	}
	value, err := strconv.ParseInt(args[index+1], 10, 64)
	if err != nil {
		return 0, tlcCommandLineError("Error: An integer for " + name + " required. But encountered " + args[index+1])
	}
	return value, nil
}

func parseWorkerCount(text string) (int, error) {
	if strings.EqualFold(strings.TrimSpace(text), "auto") {
		return runtime.NumCPU(), nil
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return 0, tlcCommandLineError("Error: worker number or 'auto' required. But encountered " + text)
	}
	if value < 1 {
		return 0, tlcCommandLineError("Error: at least one worker required.")
	}
	return value, nil
}

func requireDumpSuffix(file string, suffix string) string {
	if strings.HasSuffix(file, suffix) {
		return file
	}
	return file + suffix
}

func trimTLAExtension(file string) string {
	if strings.EqualFold(filepath.Ext(file), ".tla") {
		return strings.TrimSuffix(file, filepath.Ext(file))
	}
	return file
}

func trimConfigExtension(file string) string {
	if strings.EqualFold(filepath.Ext(file), ".cfg") {
		return strings.TrimSuffix(file, filepath.Ext(file))
	}
	return file
}

func cleanPathWithSeparator(path string) string {
	if path == "" {
		return path
	}
	clean := filepath.Clean(path)
	if strings.HasSuffix(path, string(filepath.Separator)) {
		return clean + string(filepath.Separator)
	}
	return clean + string(filepath.Separator)
}

func IsTraceExplorationSpecFile(file string) bool {
	base := filepath.Base(trimTLAExtension(file))
	return strings.HasSuffix(base, "_TE") || strings.Contains(base, "_TE_")
}
