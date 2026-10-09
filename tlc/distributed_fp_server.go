package tlc

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

var distributedFPServerRunning atomic.Bool

func init() { distributedFPServerRunning.Store(true) }

// ShutdownDistributedFPServer is DistributedFPSet.shutdown. It does not wake
// Object.wait and main does not reset this static flag on a later invocation.
func ShutdownDistributedFPServer() { distributedFPServerRunning.Store(false) }

// DistributedFPServerEnvironment supplies native boundaries used by the Java
// command. Wait replaces Object.wait while the calling thread owns the set's
// monitor; a replacement must return with that monitor still owned.
type DistributedFPServerEnvironment struct {
	Lookup            TLCServerLookup
	Sleep             DistributedLookupSleep
	ToolOut           io.Writer
	SystemOut         io.Writer
	SystemErr         io.Writer
	CurrentTimeMillis func() int64
	LocalHostName     func() (string, error)
	RegisterFPSet     func(DistributedServerEndpoint, DistributedFingerprintEndpoint, string) error
	UnpublishFPSet    func(FPSet, bool)
	Wait              func(FPSet, time.Duration) error
}

// RunDistributedFPServer ports DistributedFPSet.main, including its early
// returns, catch(Throwable), reporting monitor and output flush placement.
func RunDistributedFPServer(args []string, env DistributedFPServerEnvironment) {
	if env.ToolOut == nil {
		env.ToolOut = os.Stdout
	}
	if env.SystemOut == nil {
		env.SystemOut = os.Stdout
	}
	if env.SystemErr == nil {
		env.SystemErr = os.Stderr
	}
	if env.CurrentTimeMillis == nil {
		env.CurrentTimeMillis = func() int64 { return time.Now().UnixMilli() }
	}
	if env.LocalHostName == nil {
		env.LocalHostName = distributedLocalHostName
	}
	fmt.Fprintln(env.ToolOut, "TLC Distributed FP Server "+TLCVersion())
	if len(args) != 1 {
		fmt.Fprintln(env.ToolOut, "Error: Missing hostname of the TLC server to be contacted.")
		fmt.Fprintln(env.ToolOut, "Usage: tlago fpserver host")
		return
	}
	flush, err := runDistributedFPServer(args[0], env)
	if err != nil {
		PrintError(ECGeneral, javaGeneralErrorMessage("", err))
		fmt.Fprintln(env.ToolOut, "Error: Failed to start FPSet  for server "+args[0]+".\n"+javaNullableString(javaThrowableDetailMessage(err)))
	}
	if flush {
		// PrintStream.flush reports checked I/O through its internal error
		// flag; callers do not receive that failure as another Throwable.
		if stream, ok := env.ToolOut.(interface{ Flush() error }); ok {
			_ = stream.Flush()
		}
	}
}

func runDistributedFPServer(serverName string, env DistributedFPServerEnvironment) (flush bool, err error) {
	flush = true
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	server, err := LookupDistributedFPServer(serverName, env.Lookup, env.Sleep, env.ToolOut)
	if err != nil {
		return true, err
	}
	tmpdir := os.TempDir()
	if configured, ok := tlcLookupSystemProperty("java.io.tmpdir"); ok {
		tmpdir = configured
	}
	metadir := tmpdir + string(os.PathSeparator) + "FPSet" + fmtInt64(env.CurrentTimeMillis())
	metadirFile := filenameNormalizeFile(metadir)
	if _, err := os.Stat(metadirFile); err != nil {
		filenameFileMkdirs(metadirFile)
	}
	config := NewFPSetConfigurationWithRatio(1)
	config.SetFPBits(1)
	set := NewFPSet(config)
	registrationAttempted := false
	defer func() {
		// No remote owner can refer to this store before registration starts.
		// Preserve files and the startup failure while releasing native handles.
		// After registration starts, its outcome may be ambiguous; the native
		// network owner must drain calls before closing published storage.
		if !registrationAttempted && set != nil {
			set.Close()
		}
	}()
	filename := "FPSet" + fmtInt64(env.CurrentTimeMillis())
	if set == nil {
		return true, NewNullPointerException()
	}
	set.Init(1, metadir, filename)
	fmt.Fprintln(env.SystemErr, "FPSet instance type is: "+fpSetClassName(set))
	if multi, ok := set.(*MultiFPSet); ok {
		for _, nested := range multi.Sets {
			fmt.Fprintln(env.SystemErr, "...with nested instance type: "+fpSetClassName(nested))
		}
	}
	hostname, err := env.LocalHostName()
	if err != nil {
		return true, err
	}
	if server == nil {
		return true, NewNullPointerException()
	}
	registrationAttempted = true
	if err := invokeDistributedFPRegistration(env.RegisterFPSet, server, set, hostname); err != nil {
		if isDistributedFPRegistrationRejected(err) {
			unpublishDistributedFPSet(set, false, env)
			fmt.Fprintln(env.ToolOut, javaNullableString(javaThrowableDetailMessage(err)))
			return false, nil
		}
		return true, err
	}
	fmt.Fprintln(env.SystemOut, "Fingerprint set server at "+hostname+" is ready.")
	return true, reportDistributedFPServer(set, hostname, env)
}

func invokeDistributedFPRegistration(register func(DistributedServerEndpoint, DistributedFingerprintEndpoint, string) error, server DistributedServerEndpoint, set FPSet, hostname string) (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	if register == nil {
		return server.RegisterFPSet(NewLocalFingerprintEndpoint(set), hostname)
	}
	return register(server, NewLocalFingerprintEndpoint(set), hostname)
}

func reportDistributedFPServer(set FPSet, hostname string, env DistributedFPServerEnvironment) error {
	lifecycle, monitor := fpSetLifecycleAndMonitor(set)
	monitor.Lock()
	defer monitor.Unlock()
	for distributedFPServerRunning.Load() {
		fmt.Fprintln(env.ToolOut, "Progress: The number of fingerprints stored at "+hostname+" is "+fmtInt64(int64(set.Size()))+".")
		var err error
		if env.Wait != nil {
			err = env.Wait(set, 300000*time.Millisecond)
		} else {
			err = lifecycle.wait(monitor, 300000*time.Millisecond)
		}
		if err != nil {
			return err
		}
	}
	unpublishDistributedFPSet(set, false, env)
	fmt.Fprintln(env.ToolOut, "Exiting TLC Distributed FP Server")
	return nil
}

func unpublishDistributedFPSet(set FPSet, force bool, env DistributedFPServerEnvironment) {
	set.UnexportObject(force)
	if env.UnpublishFPSet != nil {
		env.UnpublishFPSet(set, force)
	}
}

func fpSetClassName(set FPSet) string {
	switch set.(type) {
	case *MultiFPSet:
		return "tlc2.tool.fp.MultiFPSet"
	case *MSBDiskFPSet:
		return "tlc2.tool.fp.MSBDiskFPSet"
	case *LSBDiskFPSet:
		return "tlc2.tool.fp.LSBDiskFPSet"
	case *OffHeapDiskFPSet:
		return "tlc2.tool.fp.OffHeapDiskFPSet"
	case *HeapBasedDiskFPSet:
		return "tlc2.tool.fp.HeapBasedDiskFPSet"
	case *NonCheckpointableDiskFPSet:
		return "tlc2.tool.fp.NonCheckpointableDiskFPSet"
	case *DiskFPSet:
		return "tlc2.tool.fp.DiskFPSet"
	case *MemFPSet:
		return "tlc2.tool.fp.MemFPSet"
	case *MemFPSet1:
		return "tlc2.tool.fp.MemFPSet1"
	case *MemFPSet2:
		return "tlc2.tool.fp.MemFPSet2"
	case *NoopFPSet:
		return "tlc2.tool.fp.NoopFPSet"
	default:
		panic(NewClassCastException())
	}
}

func javaNullableString(value *string) string {
	if value == nil {
		return "null"
	}
	return *value
}
