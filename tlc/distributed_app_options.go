package tlc

import (
	"fmt"
	"os"
	"strings"
)

// TLCAppOptions retains null versus empty filenames until create selects the
// packaged model or applies the source config-name default.
type TLCAppOptions struct {
	SpecFile           *string
	ConfigFile         *string
	CheckDeadlock      bool
	FPIndex            int
	FromCheckpoint     *string
	FPSetConfiguration *FPSetConfiguration
}

// ParseTLCAppOptions is the option loop in TLCApp.create, including its direct
// output, global mutations and branches that print errors but keep parsing.
func ParseTLCAppOptions(args []string) *TLCAppOptions {
	opts := &TLCAppOptions{CheckDeadlock: true, FPSetConfiguration: NewFPSetConfiguration()}
	index := 0
	for index < len(args) {
		switch args[index] {
		case "-config":
			index++
			if index >= len(args) {
				PrintTLCAppUsageError("Error: configuration file required.")
				return nil
			}
			opts.ConfigFile = javaString(strings.TrimSuffix(args[index], ".cfg"))
			index++
		case "-tool":
			index++
			Globals.Tool = true
		case "-deadlock":
			index++
			opts.CheckDeadlock = false
		case "-recover":
			index++
			if index >= len(args) {
				PrintTLCAppUsageError("Error: need to specify the metadata directory for recovery.")
				return nil
			}
			opts.FromCheckpoint = javaString(args[index] + string(os.PathSeparator))
			index++
		case "-checkpoint":
			index++
			if index >= len(args) {
				PrintTLCAppUsageError("Error: checkpoint interval required.")
				continue
			}
			value, valid := javaParseDecimalInt(args[index])
			if !valid {
				PrintTLCAppUsageError("Error: An integer for checkpoint interval is required. But encountered " + args[index])
				continue // The malformed value becomes the next option/input.
			}
			Globals.CheckpointDurationMillis = int64(value * 1000 * 60)
			if Globals.CheckpointDurationMillis < 0 {
				PrintTLCAppUsageError("Error: expect a nonnegative integer for -checkpoint option.")
			}
			index++
		case "-coverage":
			index++
			if index >= len(args) {
				PrintTLCAppUsageError("Error: coverage report interval required.")
				return nil
			}
			ToolIOPrintln("Warning: coverage reporting not supported in distributed TLC, ignoring -coverage " + args[index] + " parameter.")
			index++
		case "-terse":
			index++
			Globals.Expand = false
		case "-nowarning":
			index++
			Globals.Warn = false
		case "-maxSetSize":
			index++
			if index >= len(args) {
				PrintTLCAppUsageError("Error: maxSetSize required.")
				return nil
			}
			value, valid := javaParseDecimalInt(args[index])
			if !valid {
				PrintTLCAppUsageError("Error: An integer for maxSetSize required. But encountered " + args[index])
				return nil
			}
			if !IsValidSetSize(int(value)) {
				PrintTLCAppUsageError("Error: Value in interval [0, 2147483647] for maxSetSize required. But encountered " + args[index])
				return nil
			}
			Globals.SetBound = int(value)
			index++
		case "-fp":
			index++
			if index >= len(args) {
				PrintTLCAppUsageError("Error: expect an integer for -workers option.")
				return nil
			}
			value, valid := javaParseDecimalInt(args[index])
			if !valid {
				PrintTLCAppUsageError("Error: A number for -fp is required. But encountered " + args[index])
				return nil
			}
			if value < 0 || int(value) >= len(FP64Polys) {
				PrintTLCAppUsageError(fmt.Sprintf("Error: The number for -fp must be between 0 and %d (inclusive).", len(FP64Polys)-1))
				return nil
			}
			opts.FPIndex = int(value)
			index++
		case "-fpbits":
			index++
			if index >= len(args) {
				PrintTLCAppUsageError("Error: expect an integer for -workers option.")
				return nil
			}
			value, valid := javaParseDecimalInt(args[index])
			if !valid {
				PrintTLCAppUsageError("Error: A number for -fpbits is required. But encountered " + args[index])
				return nil
			}
			if !IsValidFPBits(int(value)) {
				PrintTLCAppUsageError("Error: Value in interval [0, 30] for fpbits required. But encountered " + args[index])
				return nil
			}
			opts.FPSetConfiguration.SetFPBits(int(value))
			index++
		case "-fpmem":
			index++
			if index >= len(args) {
				continue // The source emits no missing-value diagnostic here.
			}
			value, err := parseJavaDoubleProperty(args[index])
			if err == nil && value < 0 {
				PrintTLCAppUsageError("Error: An positive integer or a fraction for fpset memory size/percentage required. But encountered " + args[index])
				return nil
			}
			if err != nil || !setTLCAppFPMemory(opts.FPSetConfiguration, value) {
				PrintTLCAppUsageError("Error: A positive integer or a fraction for fpset memory size/percentage required. But encountered " + args[index])
				return nil
			}
			index++
		case "-metadir":
			index++
			if index >= len(args) {
				PrintTLCAppUsageError("Error: need to specify the metadata directory.")
				return nil
			}
			Globals.MetaDir = args[index] + string(os.PathSeparator)
			index++
		default:
			arg := args[index]
			if arg == "" {
				panic(NewStringIndexOutOfBoundsException(0, 0))
			}
			if arg[0] == '-' {
				PrintTLCAppUsageError("Error: unrecognized option: " + arg)
				return nil
			}
			if opts.SpecFile != nil {
				PrintTLCAppUsageError("Error: more than one input files: " + *opts.SpecFile + " and " + arg)
				return nil
			}
			opts.SpecFile = javaString(strings.TrimSuffix(arg, ".tla"))
			index++
		}
	}
	return opts
}

func setTLCAppFPMemory(config *FPSetConfiguration, value float64) (valid bool) {
	defer func() {
		if failure := recover(); failure != nil {
			if isJavaError(panicValueAsError(failure)) {
				panic(failure)
			}
			valid = false // The source catches Exception around both setters.
		}
	}()
	if value > 1 {
		ToolIOPrintln("Using -fpmem with an abolute memory value has been deprecated. Please allocate memory for the TLC process via the JVM mechanisms and use -fpmem to set the fraction to be used for fingerprint storage.")
		config.SetMemory(javaDoubleToLong(value))
		config.SetRatio(1)
	} else {
		config.SetRatio(value)
	}
	return true
}

func PrintTLCAppUsageError(message string) {
	ToolIOPrintln(message)
	ToolIOPrintln("Usage: java tlc2.tool.TLCServer [-option] inputfile")
}
