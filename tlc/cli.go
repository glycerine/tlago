package tlc

import (
	"fmt"
	"math"
	"math/rand"
	"os"
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
	Module       string
	Operator     string
	ConstantName string
	FileName     string
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
	parseOK := false
	defer func() {
		if !parseOK && opts.UserOutput != nil {
			_ = opts.UserOutput.Close()
		}
	}()
	dump := tlcDumpOption{}
	generateTESpec := true
	generateTESpecBinaryTrace := true
	forceGenerateTESpec := false
	teSpecMonolith := true
	teSpecOut := ""
	metaDirRoot := ""
	traceFileSet := false

	tlcSuppressedCodes := NewInsMap[int, bool]()
	tlcMessagesAsErrors := NewInsMap[int, bool]()
	sanySuppressedCodes := NewInsMap[int, bool]()
	sanyMessagesAsErrors := NewInsMap[int, bool]()

	index := 0
	for index < len(args) {
		arg := args[index]
		switch {
		case arg == "-simulate" || arg == "-generate":
			if arg == "-generate" {
				opts.Probabilistic = true
				tlcSetStartupSystemProperty(toolProbabilisticProperty, "true")
			}
			opts.Mode = RunModeSimulate
			index++
			if index < len(args) && isSimulationSubargument(args[index]) {
				for _, simArg := range strings.Split(args[index], ",") {
					switch {
					case strings.HasPrefix(simArg, "num="):
						count := strings.ReplaceAll(simArg, "num=", "")
						traceNum, valid := javaParseDecimalLong(count)
						if !valid {
							return opts, NewNumberFormatException(count)
						}
						opts.TraceNum = traceNum
						opts.TraceNumSet = true
					case strings.HasPrefix(simArg, "file="):
						opts.TraceFile = strings.ReplaceAll(simArg, "file=", "")
						traceFileSet = true
					case simArg == "stats=basic":
						opts.TraceActions = "BASIC"
					case simArg == "stats=full":
						opts.TraceActions = "FULL"
					case simArg == "sched=rl":
						tlcSetStartupSystemProperty(simulatorRLProperty, "true")
						opts.SimulationSchedule = SimulationScheduleRL
					case simArg == "sched=rlaction":
						tlcSetStartupSystemProperty(simulatorRLActionProperty, "true")
						opts.SimulationSchedule = SimulationScheduleRLAction
					}
				}
				index++
				if opts.TraceNum == math.MaxInt64 && traceFileSet {
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
			index, err = parseTLCMessageCodeList(args, index, tlcSuppressedCodes, sanySuppressedCodes, "-suppressMessages", true)
			if err != nil {
				return opts, err
			}
		case arg == "-messagesAsErrors":
			var err error
			index, err = parseTLCMessageCodeList(args, index, tlcMessagesAsErrors, sanyMessagesAsErrors, "-messagesAsErrors", false)
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
		case strings.HasPrefix(arg, "-D"):
			name, value, err := parseTLCSystemPropertyOption(arg)
			if err != nil {
				return opts, err
			}
			tlcSetStartupSystemProperty(name, value)
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
			if _, valid := javaParseDecimalInt(args[index+1]); !valid {
				return opts, tlcCommandLineError("Error: An integer for -invlevel required. But encountered " + args[index+1])
			}
			opts.RuntimeParams.Invariants = append(opts.RuntimeParams.Invariants, RuntimeInvariantTemplate{
				Modules:    []string{"TLC", "Naturals"},
				Expression: fmt.Sprintf("LET _T == INSTANCE TLC _N == INSTANCE Naturals IN _N!<(_T!TLCGet(\"level\"), %s)", args[index+1]),
			})
			index += 2
		case arg == "-debugger":
			opts.DebugPort = 4712
			opts.DebugPortSet = true
			opts.DebugSuspend = defaultTLCDebugSuspend()
			opts.DebugHalt = defaultTLCDebugHalt()
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
			_, err := parseMinuteIntervalOption(args, index, "coverage report interval", "-coverage")
			if err != nil {
				return opts, err
			}
			index += 2
		case arg == "-checkpoint":
			value, err := parseMinuteIntervalOption(args, index, "checkpoint interval", "-checkpoint")
			if err != nil {
				return opts, err
			}
			opts.CheckpointDurationMillis = value
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
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: maxSetSize required.")
			}
			bound, ok := javaParseDecimalInt(args[index+1])
			if !ok {
				return opts, tlcCommandLineError("Error: An integer for maxSetSize required. But encountered " + args[index+1])
			}
			if !IsValidSetSize(int(bound)) {
				return opts, tlcCommandLineError("Error: Value in interval [0, 2147483647] for maxSetSize required. But encountered " + args[index+1])
			}
			Globals.SetBound = int(bound)
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
			metaDirRoot = cleanPathWithSeparator(args[index+1])
			Globals.MetaDir = metaDirRoot
			index += 2
		case arg == "-userFile":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: need to specify the full qualified file.")
			}
			file, err := os.Create(args[index+1])
			if err != nil {
				return opts, tlcCommandLineError("Error: Failed to create user output log file.")
			}
			if opts.UserOutput != nil {
				_ = opts.UserOutput.Close()
			}
			opts.RuntimeParams.UserFile = args[index+1]
			opts.UserOutput = file
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
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: expect a nonnegative integer for -dfid option.")
			}
			value, valid := javaParseDecimalInt(args[index+1])
			if !valid {
				return opts, tlcCommandLineError("Error: expect a nonnegative integer for -dfid option. But encountered " + args[index+1])
			}
			Globals.DFIDMax = int(value)
			if value < 0 {
				return opts, tlcCommandLineError("Error: expect a nonnegative integer for -dfid option.")
			}
			opts.DFIDMode = true
			opts.DFIDDepth = int(value)
			index += 2
		case arg == "-fp":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: expect an integer for -fp option.")
			}
			value, valid := javaParseDecimalInt(args[index+1])
			if !valid {
				return opts, tlcCommandLineError("Error: A number for -fp is required. But encountered " + args[index+1])
			}
			opts.FPIndex = int(value)
			if value < 0 || int(value) >= len(FP64Polys) {
				return opts, tlcCommandLineError(fmt.Sprintf("Error: The number for -fp must be between 0 and %d (inclusive).", len(FP64Polys)-1))
			}
			index += 2
		case arg == "-fpmem":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: fpset memory size required.")
			}
			fpMemSize, err := parseJavaDoubleProperty(args[index+1])
			if err != nil || fpMemSize < 0 || math.IsNaN(fpMemSize) {
				return opts, tlcCommandLineError("Error: An positive integer or a fraction for fpset memory size/percentage required. But encountered " + args[index+1])
			}
			if fpMemSize > 1 {
				ToolIOPrintln("Using -fpmem with an abolute memory value has been deprecated. Please allocate memory for the TLC process via the JVM mechanisms and use -fpmem to set the fraction to be used for fingerprint storage.")
				opts.FPSetConfiguration.SetMemory(javaDoubleToLong(fpMemSize))
				opts.FPSetConfiguration.SetRatio(1)
			} else {
				opts.FPSetConfiguration.SetRatio(fpMemSize)
			}
			index += 2
		case arg == "-fpbits":
			value, err := parseIntOption(args, index, "fpbits", "-fpbits")
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

	if !Globals.Warn && (tlcSuppressedCodes.Len() > 0 || sanySuppressedCodes.Len() > 0) {
		return opts, tlcCommandLineError("Error: -nowarning already suppresses all warnings, making -suppressMessages redundant. Remove -nowarning to suppress only the specified messages, or remove -suppressMessages to suppress all warnings.")
	}
	if overlap := messageCodeOverlap(tlcSuppressedCodes, tlcMessagesAsErrors); overlap != "" {
		return opts, tlcCommandLineError("Error: The following codes were set to both -suppressMessages and -messagesAsErrors: " + overlap)
	}
	if overlap := sanyMessageCodeOverlap(sanySuppressedCodes, sanyMessagesAsErrors); overlap != "" {
		return opts, tlcCommandLineError("Error: The following codes were set to both -suppressMessages and -messagesAsErrors: " + overlap)
	}
	for code := range tlcSuppressedCodes.All() {
		SuppressTLCMessage(code)
	}
	for code := range tlcMessagesAsErrors.All() {
		TreatTLCMessageAsError(code)
	}

	for code := range sanySuppressedCodes.All() {
		SuppressSANYMessage(code)
	}
	for code := range sanyMessagesAsErrors.All() {
		TreatSANYMessageAsError(code)
	}

	if opts.SpecFile == "" {
		return opts, tlcCommandLineError("Error: Missing input TLA+ module.")
	}
	if opts.ConfigFile == "" {
		opts.ConfigFile = opts.SpecFile
	}
	opts.StartTime = time.Now()
	if opts.Cleanup && opts.FromCheckpoint == "" {
		deleteDirLikeJava(MetaRoot, true)
		opts.CleanupPrecleanDone = true
	}
	specDir := ""
	if filepath.IsAbs(opts.SpecFile) {
		specDir = filepath.Dir(opts.SpecFile)
	}
	metadir, err := makeTLCMetaDir(opts.StartTime, specDir, metaDirRoot, opts.FromCheckpoint)
	if err != nil {
		return opts, tlcCommandLineError(fmt.Sprintf("Error: Could not create metadata directory: %v", err))
	}
	opts.MetaDir = metadir
	Globals.MetaDir = opts.MetaDir
	if dump.File != "" {
		writer, err := dump.newStateWriter(opts.MetaDir)
		if err != nil {
			return opts, tlcCommandLineError(fmt.Sprintf("Error: Given file name %s for dumping states invalid.", dump.File))
		}
		opts.StateWriter = writer
	} else if dump.CustomClass != "" {
		writer, err := NewCustomStateWriter(dump.CustomClass)
		if err != nil {
			return opts, tlcCommandLineError(fmt.Sprintf("Error: Could not instantiate a custom IStateWriter implementation %s.", dump.CustomClass))
		}
		opts.RuntimeParams.CustomStateWriterClass = dump.CustomClass
		opts.StateWriter = writer
	}

	opts.ForceGenerateTraceSpec = forceGenerateTESpec
	opts.GenerateTraceSpec = forceGenerateTESpec || (generateTESpec && !Globals.Tool && !Globals.Continuation && !IsTraceExplorationSpecFile(opts.SpecFile))
	opts.GenerateTraceSpecBinary = generateTESpecBinaryTrace
	opts.GenerateTraceSpecMonolith = teSpecMonolith
	opts.TraceSpecOutputDir = teSpecOut
	if opts.GenerateTraceSpec {
		opts.TraceSpecOutputDir, opts.TraceSpecModuleName = traceSpecOutputAndModule(opts.SpecFile, opts.TraceSpecOutputDir, opts.StartTime)
		if opts.GenerateTraceSpecBinary {
			opts.RuntimeParams.PostConditions = append(opts.RuntimeParams.PostConditions, RuntimePostCondition{
				Module:       "_TLCTrace",
				Operator:     "_TLCTraceSilent",
				ConstantName: "_TLCTraceFile",
				FileName:     filepath.Join(opts.TraceSpecOutputDir, opts.TraceSpecModuleName+".bin"),
			})
		}
	}
	parseOK = true
	return opts, nil
}

func (t *TLC) HandleParameters(args []string) error {
	opts, err := ParseTLCOptions(args)
	if err != nil {
		if _, commandLineError := err.(*TLCCommandLineError); t != nil && commandLineError {
			t.printWelcome()
			PrintError(ECWrongCommandlineParamsTLC, err.Error())
		}
		return err
	}
	if t == nil {
		return nil
	}
	if t.Tool != nil && opts.Tool == nil {
		opts.Tool = t.Tool
	}
	t.Options = opts
	t.printWelcome()
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
		DebugSuspend:              defaultTLCDebugSuspend(),
		DebugHalt:                 defaultTLCDebugHalt(),
		GenerateTraceSpecBinary:   true,
		GenerateTraceSpecMonolith: true,
	}
}

func parseTLCSystemPropertyOption(arg string) (string, string, error) {
	property := strings.TrimPrefix(arg, "-D")
	if property == "" {
		return "", "", tlcCommandLineError("Error: expected a property name after -D.")
	}
	name, value, hasValue := strings.Cut(property, "=")
	if name == "" {
		return "", "", tlcCommandLineError("Error: expected a property name after -D.")
	}
	if !hasValue {
		value = ""
	}
	return name, value, nil
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

func parseTLCMessageCodeList(args []string, index int, dst *InsMap[int, bool], sanyDst *InsMap[int, bool], option string, suppress bool) (int, error) {
	if index+1 >= len(args) {
		return index, tlcCommandLineError("Error: " + option + " requires a comma-separated list of message codes.")
	}
	parts := strings.Split(args[index+1], ",")
	if args[index+1] != "" {
		for len(parts) > 0 && parts[len(parts)-1] == "" {
			parts = parts[:len(parts)-1]
		}
	}
	for _, part := range parts {
		text := strings.TrimFunc(part, func(r rune) bool { return r <= ' ' })
		parsed, ok := javaParseDecimalInt(text)
		if !ok {
			return index, tlcCommandLineError("Error: " + *NewNumberFormatException(text).GetMessage())
		}
		code := int(parsed)
		if !knownJavaMessageCode(code) {
			return index, tlcCommandLineError(fmt.Sprintf("Error: unknown message code: %d", code))
		}
		if suppress && isJavaSANYErrorMessageCode(code) {
			return index, tlcCommandLineError(fmt.Sprintf("Error: code %d is an error and cannot be suppressed.", code))
		}
		if _, tlcCode := knownJavaTLCMessageCodes[code]; tlcCode {
			dst.Set(code, true)
		} else {
			sanyDst.Set(code, true)
		}
	}
	return index + 2, nil
}

func knownJavaMessageCode(code int) bool {
	if _, ok := knownJavaTLCMessageCodes[code]; ok {
		return true
	}
	_, ok := knownJavaSANYMessageCodes[code]
	return ok
}

func isJavaSANYErrorMessageCode(code int) bool {
	if _, tlcCode := knownJavaTLCMessageCodes[code]; tlcCode {
		return false
	}
	if _, ok := knownJavaSANYMessageCodes[code]; !ok {
		return false
	}
	_, warning := knownJavaSANYWarningMessageCodes[code]
	return !warning
}

func intSet(values ...int) map[int]struct{} {
	out := make(map[int]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

var knownJavaTLCMessageCodes = intSet(
	-123456, -1, 0, 1000, 1001, 1002, 1003, 1005, 1101, 1102,
	2000, 2001, 2100, 2101, 2102, 2103, 2104, 2105, 2106, 2107,
	2108, 2109, 2110, 2111, 2112, 2113, 2114, 2115, 2116, 2117,
	2118, 2119, 2120, 2121, 2122, 2123, 2124, 2125, 2126, 2127,
	2128, 2129, 2130, 2131, 2132, 2133, 2134, 2135, 2136, 2137,
	2138, 2139, 2140, 2141, 2142, 2143, 2144, 2145, 2146, 2147,
	2148, 2149, 2154, 2155, 2156, 2157, 2158, 2159, 2160, 2161,
	2162, 2163, 2164, 2165, 2166, 2167, 2168, 2169, 2170, 2171,
	2172, 2173, 2174, 2175, 2176, 2177, 2178, 2179, 2180, 2181,
	2182, 2183, 2184, 2185, 2186, 2187, 2188, 2189, 2190, 2191,
	2192, 2193, 2194, 2195, 2196, 2197, 2198, 2199, 2200, 2201,
	2202, 2203, 2204, 2205, 2206, 2207, 2208, 2209, 2210, 2211,
	2212, 2213, 2214, 2215, 2216, 2217, 2218, 2219, 2220, 2221,
	2222, 2223, 2224, 2225, 2226, 2227, 2228, 2229, 2230, 2231,
	2232, 2233, 2234, 2235, 2236, 2237, 2238, 2239, 2240, 2241,
	2242, 2243, 2244, 2245, 2246, 2247, 2248, 2249, 2250, 2251,
	2252, 2253, 2254, 2255, 2256, 2257, 2258, 2259, 2260, 2261,
	2262, 2263, 2264, 2265, 2266, 2267, 2268, 2269, 2270, 2271,
	2272, 2273, 2274, 2279, 2280, 2281, 2282, 2283, 2284, 2300,
	2301, 2400, 2401, 2402, 2403, 2404, 2405, 2501, 2502, 2772,
	2773, 2774, 2775, 2776, 2777, 2778, 2779, 2780, 3000, 3001,
	3002, 3100, 3101, 3102, 3103, 3104, 3105, 3106, 3107, 3108,
	3109, 3110, 3111, 3112, 3113, 3114, 4000, 4001, 4002, 5001,
	5002, 5003, 5004, 5005, 5006, 7000, 7001, 7002, 7003, 7004,
	7005, 7006, 7007, 7008, 7009, 7010, 20000,
)

var knownJavaSANYMessageCodes = intSet(
	1000, 4003, 4004, 4005, 4200, 4201, 4202, 4203, 4204, 4205,
	4206, 4220, 4221, 4222, 4223, 4224, 4240, 4241, 4242, 4243,
	4244, 4245, 4246, 4247, 4260, 4261, 4262, 4270, 4271, 4272,
	4273, 4274, 4275, 4290, 4291, 4292, 4293, 4294, 4310, 4311,
	4312, 4313, 4314, 4315, 4330, 4331, 4332, 4333, 4334, 4335,
	4336, 4337, 4350, 4351, 4352, 4353, 4354, 4355, 4356, 4357,
	4800, 4801, 4802, 4803, 4804, 4805,
)

var knownJavaSANYWarningMessageCodes = intSet(4800, 4801, 4802, 4803, 4804, 4805)

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
	if len(parts) == 0 {
		return ""
	}
	return "[" + strings.Join(parts, ", ") + "]"
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
		return NewCustomStateWriter(d.CustomClass)
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
			StrictPrefix: true,
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
	value, valid := javaParseDecimalInt(args[index+1])
	if !valid {
		return 0, tlcCommandLineError("Error: An integer for " + name + " required. But encountered " + args[index+1])
	}
	return int(value), nil
}

// Java multiplies the parsed int before assigning its milliseconds to a long.
// The global assignment precedes validation of that wrapped signed result.
func parseMinuteIntervalOption(args []string, index int, name, option string) (int64, error) {
	if index+1 >= len(args) {
		return 0, tlcCommandLineError("Error: " + name + " required.")
	}
	minutes, valid := javaParseDecimalInt(args[index+1])
	if !valid {
		verb := ""
		if option == "-checkpoint" {
			verb = " is"
		}
		return 0, tlcCommandLineError("Error: An integer for " + name + verb + " required. But encountered " + args[index+1])
	}
	millis := minutes * int32(60000)
	if option == "-coverage" {
		Globals.CoverageInterval = int(millis)
	} else {
		Globals.CheckpointDurationMillis = int64(millis)
	}
	if millis < 0 {
		return 0, tlcCommandLineError("Error: expect a nonnegative integer for " + option + " option.")
	}
	return int64(millis), nil
}

func parseInt64Option(args []string, index int, name string, option string) (int64, error) {
	if index+1 >= len(args) {
		return 0, tlcCommandLineError("Error: " + name + " required.")
	}
	value, valid := javaParseDecimalLong(args[index+1])
	if !valid {
		return 0, tlcCommandLineError("Error: An integer for " + name + " required. But encountered " + args[index+1])
	}
	return value, nil
}

func parseWorkerCount(text string) (int, error) {
	if strings.ToLower(strings.TrimFunc(text, func(char rune) bool { return char <= ' ' })) == "auto" {
		return runtime.NumCPU(), nil
	}
	value, valid := javaParseDecimalInt(text)
	if !valid {
		return 0, tlcCommandLineError("Error: worker number or 'auto' required. But encountered " + text)
	}
	if value < 1 {
		return 0, tlcCommandLineError("Error: at least one worker required.")
	}
	return int(value), nil
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

func traceSpecOutputAndModule(specFile string, output string, timestamp time.Time) (string, string) {
	originalModule := filepath.Base(trimTLAExtension(specFile))
	if originalModule == "" || originalModule == "." {
		originalModule = "Spec"
	}
	if output == "" {
		dir := filepath.Dir(specFile)
		if dir == "" {
			dir = "."
		}
		return dir, DeriveTESpecModuleName(originalModule, timestamp)
	}
	if strings.EqualFold(filepath.Ext(output), ".tla") {
		dir := filepath.Dir(output)
		if dir == "" {
			dir = "."
		}
		return dir, strings.TrimSuffix(filepath.Base(output), filepath.Ext(output))
	}
	return output, DeriveTESpecModuleName(originalModule, timestamp)
}

func makeTLCMetaDir(date time.Time, specDir string, metaDirRoot string, fromCheckpoint string) (string, error) {
	if fromCheckpoint != "" {
		return fromCheckpoint, nil
	}
	return CreateExclusiveDirectoryWithApproximateName(tlcMetaDirPath(date, specDir, metaDirRoot))
}

func tlcMetaDirDateLayout() string {
	if value, ok := tlcLookupSystemProperty("util.FileUtil.milliseconds"); ok && !javaBooleanProperty(value) {
		return "06-01-02-15-04-05"
	}
	return "06-01-02-15-04-05.000"
}

func IsTraceExplorationSpecFile(file string) bool {
	base := filepath.Base(trimTLAExtension(file))
	return strings.HasSuffix(base, "_TE") || strings.Contains(base, "_TE_")
}
