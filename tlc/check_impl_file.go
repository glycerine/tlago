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
	for index := 0; index < len(args); {
		switch args[index] {
		case "-config":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: expect a file name for -config option.")
			}
			opts.ConfigFile = trimConfigExtension(args[index+1])
			index += 2
		case "-deadlock":
			opts.Deadlock = false
			index++
		case "-recover":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: need to specify the metadata directory for recovery.")
			}
			opts.FromCheckpoint = cleanPathWithSeparator(args[index+1])
			index += 2
		case "-workers":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: expect an integer for -workers option.")
			}
			workers, err := parseWorkerCount(args[index+1])
			if err != nil {
				return opts, err
			}
			opts.Workers = workers
			SetNumWorkers(workers)
			index += 2
		case "-depth":
			depth, err := parseIntOption(args, index, "depth", "-depth")
			if err != nil {
				return opts, err
			}
			opts.Depth = depth
			index += 2
		case "-trace":
			if index+1 >= len(args) {
				return opts, tlcCommandLineError("Error: trace file prefix required.")
			}
			opts.TraceFile = args[index+1]
			index += 2
		case "-coverage":
			coverage, err := parseNonnegativeIntOption(args, index, "coverage", "-coverage")
			if err != nil {
				return opts, err
			}
			opts.CoverageMillis = coverage * 60 * 1000
			Globals.CoverageInterval = opts.CoverageMillis
			index += 2
		default:
			if len(args[index]) > 0 && args[index][0] == '-' {
				return opts, tlcCommandLineError("Error: unrecognized option: " + args[index])
			}
			if opts.MainFile != "" {
				return opts, tlcCommandLineError("Error: more than one input files: " + opts.MainFile + " and " + args[index])
			}
			opts.MainFile = trimTLAExtension(args[index])
			index++
		}
	}
	if opts.MainFile == "" {
		return opts, tlcCommandLineError("Error: Missing input TLA+ module.")
	}
	if opts.ConfigFile == "" {
		opts.ConfigFile = opts.MainFile
	}
	if opts.TraceFile == "" {
		opts.TraceFile = opts.MainFile + "_trace"
	}
	return opts, nil
}
