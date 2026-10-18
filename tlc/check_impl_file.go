package tlc

import (
	"fmt"
	"os"
)

const checkImplFileWaitForTraceMillis = 10000

type CheckImplFile struct {
	*CheckImpl
	States            []*TLCStateMut
	StateIndex        int
	TraceFile         string
	TraceInCount      int
	TraceOutCount     int
	LoadTraceFunc     func(filename string) ([]*TLCStateMut, error)
	LoadTraceIntoFunc func(filename string, publish func([]*TLCStateMut)) error
}

func NewCheckImplFile(tool *Tool, metadir string, deadlock bool, depth int, fromChkpt *string, traceFile string, config ...*FPSetConfiguration) *CheckImplFile {
	checker := &CheckImplFile{
		CheckImpl:     NewCheckImpl(tool, metadir, deadlock, depth, fromChkpt, config...),
		TraceFile:     traceFile,
		TraceInCount:  1,
		TraceOutCount: 1,
	}
	checker.CheckImpl.GetStateFunc = checker.GetState
	checker.CheckImpl.ExportTraceFn = checker.ExportTrace
	return checker
}

func (c *CheckImplFile) GetState() *TLCStateMut {
	if c == nil || c.States == nil {
		panic(NewNullPointerException())
	}
	if c.StateIndex >= len(c.States) {
		return nil
	}
	index := c.StateIndex
	c.StateIndex = checkImplFileIncrement(c.StateIndex)
	if index < 0 {
		panic(NewArrayIndexOutOfBoundsException(index, len(c.States)))
	}
	return c.States[index]
}

func (c *CheckImplFile) ExportTrace(trace []*TLCStateInfo) error {
	if c == nil {
		panic(NewNullPointerException())
	}
	filename := fmt.Sprintf("%s_out_%d", c.TraceFile, c.TraceOutCount)
	file, err := openFileOutputStream(filename)
	if err != nil {
		return err
	}
	writer := &statePrintWriter{buffer: NewJavaBufferedWriter(file)}
	lineSeparator := javaFormatLineSeparator()
	if trace == nil {
		panic(NewNullPointerException())
	}
	for i, info := range trace {
		_, _ = fmt.Fprintf(writer, "STATE_%d%s", i+1, lineSeparator)
		if info == nil {
			panic(NewNullPointerException())
		}
		if state := info.ownedState(); state != nil {
			_, _ = fmt.Fprintf(writer, "%s\n%s", state, lineSeparator)
		} else {
			_, _ = writer.WriteString("null\n" + lineSeparator)
		}
	}
	// PrintWriter records output failures internally; CheckImplFile never calls
	// checkError, so only opening the FileOutputStream can fail this method.
	writer.Flush()
	_ = file.Close()
	c.TraceOutCount = checkImplFileIncrement(c.TraceOutCount)
	return nil
}

func (c *CheckImplFile) GetTrace() (bool, error) {
	if c == nil {
		panic(NewNullPointerException())
	}
	filename := fmt.Sprintf("%s%d", c.TraceFile, c.TraceInCount)
	ToolIOPrintln(fmt.Sprintf("Trying to work on trace %s ...", filename))
	if _, err := os.Stat(filename); err != nil {
		return false, nil
	}
	if c.LoadTraceIntoFunc != nil {
		err := c.LoadTraceIntoFunc(filename, func(states []*TLCStateMut) {
			c.States = states
		})
		if err != nil {
			return false, err
		}
	} else if c.LoadTraceFunc == nil {
		return false, newTLCError(ECGeneral, "CheckImplFile has no trace loader for %s", filename)
	} else {
		states, err := c.LoadTraceFunc(filename)
		if err != nil {
			if states != nil {
				c.States = make([]*TLCStateMut, len(states))
				copy(c.States, states)
			}
			return false, err
		}
		c.States = make([]*TLCStateMut, len(states))
		copy(c.States, states)
	}
	c.StateIndex = 0
	c.TraceInCount = checkImplFileIncrement(c.TraceInCount)
	return true, nil
}

func checkImplFileIncrement(value int) int {
	return int(int32(value) + 1)
}

type CheckImplFileOptions struct {
	ConfigFile     string
	TraceFile      string
	Deadlock       bool
	Depth          int
	FromCheckpoint *string
	Workers        int
	CoverageMillis int
	MainFile       string
}

func ParseCheckImplFileOptions(args []string) (CheckImplFileOptions, error) {
	opts := CheckImplFileOptions{Deadlock: true, Depth: 20}
	mainFileSet, configFileSet, traceFileSet := false, false, false
	for index := 0; index < len(args); {
		switch args[index] {
		case "-config":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError(GetMessage(ECCheckParamExpectConfigFilename))
			}
			opts.ConfigFile = trimConfigExtension(args[index+1])
			configFileSet = true
			index += 2
		case "-deadlock":
			opts.Deadlock = false
			index++
		case "-recover":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError(GetMessage(ECCheckParamNeedToSpecifyConfigDir))
			}
			fromCheckpoint := args[index+1] + string(os.PathSeparator)
			opts.FromCheckpoint = &fromCheckpoint
			index += 2
		case "-workers":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError(GetMessage(ECCheckParamWorkerNumberRequired2))
			}
			workers, valid := javaParseDecimalInt(args[index+1])
			if !valid {
				return opts, tlcCommandLineError(GetMessage(ECCheckParamWorkerNumberRequired, args[index+1]))
			}
			opts.Workers = int(workers)
			SetNumWorkers(int(workers))
			if NumWorkers() < 1 {
				return opts, tlcCommandLineError(GetMessage(ECCheckParamWorkerNumberTooSmall))
			}
			index += 2
		case "-depth":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError(GetMessage(ECCheckParamDepthRequired2))
			}
			depth, valid := javaParseDecimalInt(args[index+1])
			if !valid {
				return opts, tlcCommandLineError(GetMessage(ECCheckParamDepthRequired, args[index+1]))
			}
			opts.Depth = int(depth)
			index += 2
		case "-trace":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError(GetMessage(ECCheckParamTraceRequired))
			}
			opts.TraceFile = args[index+1]
			traceFileSet = true
			index += 2
		case "-coverage":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError(GetError(ECCheckParamCovreageRequired))
			}
			coverage, valid := javaParseDecimalInt(args[index+1])
			if !valid {
				return opts, tlcCommandLineError(GetError(ECCheckParamCovreageRequired, args[index+1]))
			}
			opts.CoverageMillis = int(coverage * int32(60000))
			Globals.CoverageInterval = opts.CoverageMillis
			if Globals.CoverageInterval < 0 {
				return opts, tlcCommandLineError(GetMessage(ECCheckParamCovreageTooSmall))
			}
			index += 2
		default:
			if args[index] == "" {
				return opts, NewStringIndexOutOfBoundsException(0, 0)
			}
			if args[index][0] == '-' {
				return opts, tlcCommandLineError(GetError(ECCheckParamUnrecognized, args[index]))
			}
			if mainFileSet {
				return opts, tlcCommandLineError(GetError(ECCheckParamUnrecognized, args[index], opts.MainFile))
			}
			opts.MainFile = trimTLAExtension(args[index])
			mainFileSet = true
			index++
		}
	}
	if !mainFileSet {
		return opts, tlcCommandLineError(GetMessage(ECCheckParamMissingTLAModule))
	}
	if !configFileSet {
		opts.ConfigFile = opts.MainFile
	}
	if !traceFileSet {
		opts.TraceFile = opts.MainFile + "_trace"
	}
	return opts, nil
}
