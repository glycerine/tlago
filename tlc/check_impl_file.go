package tlc

import (
	"bufio"
	"fmt"
	"os"
)

const checkImplFileWaitForTraceMillis = 10000

type CheckImplFile struct {
	*CheckImpl
	States        []*TLCStateMut
	StateIndex    int
	TraceFile     string
	TraceInCount  int
	TraceOutCount int
	LoadTraceFunc func(filename string) ([]*TLCStateMut, error)
}

func NewCheckImplFile(tool *Tool, metadir string, deadlock bool, depth int, fromChkpt string, traceFile string, config ...*FPSetConfiguration) *CheckImplFile {
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
	if c == nil || c.StateIndex >= len(c.States) {
		return nil
	}
	state := c.States[c.StateIndex]
	c.StateIndex++
	return state
}

func (c *CheckImplFile) ExportTrace(trace []*TLCStateInfo) error {
	if c == nil {
		return nil
	}
	filename := fmt.Sprintf("%s_out_%d", c.TraceFile, c.TraceOutCount)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	for i, info := range trace {
		if _, err := fmt.Fprintf(writer, "STATE_%d\n", i+1); err != nil {
			_ = file.Close()
			return err
		}
		if info != nil && info.State != nil {
			if _, err := fmt.Fprintf(writer, "%s\n\n", info.State); err != nil {
				_ = file.Close()
				return err
			}
		} else if _, err := writer.WriteString("\n"); err != nil {
			_ = file.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	c.TraceOutCount++
	return nil
}

func (c *CheckImplFile) GetTrace() (bool, error) {
	if c == nil {
		return false, nil
	}
	filename := fmt.Sprintf("%s%d", c.TraceFile, c.TraceInCount)
	fmt.Fprintf(os.Stdout, "Trying to work on trace %s ...\n", filename)
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if c.LoadTraceFunc == nil {
		return false, newTLCError(ECGeneral, "CheckImplFile has no trace loader for %s", filename)
	}
	states, err := c.LoadTraceFunc(filename)
	if err != nil {
		return false, err
	}
	c.States = append(c.States[:0], states...)
	c.StateIndex = 0
	c.TraceInCount++
	return true, nil
}

type CheckImplFileOptions struct {
	ConfigFile     string
	TraceFile      string
	Deadlock       bool
	Depth          int
	FromCheckpoint string
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
			opts.FromCheckpoint = args[index+1] + string(os.PathSeparator)
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
