package tlc

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// Throwable detail messages are nullable in Java. Go's error string remains a
// separate convenience, especially for fingerprint exceptions with root causes.
func javaThrowableDetailMessage(err error) *string {
	if err == nil {
		return nil
	}
	if throwable, ok := err.(interface{ GetMessage() *string }); ok {
		return throwable.GetMessage()
	}
	message := err.Error()
	return &message
}

func javaString(value string) *string { return &value }

func javaThrowableClassName(err error) string {
	switch failure := err.(type) {
	case *StackOverflowError:
		return "java.lang.StackOverflowError"
	case *RuntimeException:
		return "java.lang.RuntimeException"
	case *RemoteException:
		return "java.rmi.RemoteException"
	case *ServerException:
		return "java.rmi.ServerException"
	case *NoSuchObjectException:
		return "java.rmi.NoSuchObjectException"
	case *ConnectException:
		return "java.rmi.ConnectException"
	case *FPSetManagerException:
		return "tlc2.tool.distributed.fp.FPSetManagerException"
	case *UnmarshalException:
		return "java.rmi.UnmarshalException"
	case *ExportException:
		return "java.rmi.server.ExportException"
	case *AccessException:
		return "java.rmi.AccessException"
	case *IOException:
		return "java.io.IOException"
	case *FileNotFoundException:
		return "java.io.FileNotFoundException"
	case *UnknownHostException:
		return "java.net.UnknownHostException"
	case *InvalidPathException:
		return "java.nio.file.InvalidPathException"
	case *IndexOutOfBoundsException:
		return "java.lang.IndexOutOfBoundsException"
	case *StringIndexOutOfBoundsException:
		return "java.lang.StringIndexOutOfBoundsException"
	case *IllegalArgumentException:
		return "java.lang.IllegalArgumentException"
	case *URISyntaxException:
		return "java.net.URISyntaxException"
	case *ArithmeticException:
		return "java.lang.ArithmeticException"
	case *UnsupportedOperationException:
		return "java.lang.UnsupportedOperationException"
	case *ExecutionException:
		return "java.util.concurrent.ExecutionException"
	case *EOFException:
		return "java.io.EOFException"
	case *WrongInvocationException:
		return "util.WrongInvocationException"
	case *ArrayIndexOutOfBoundsException:
		return "java.lang.ArrayIndexOutOfBoundsException"
	case *NegativeArraySizeException:
		return "java.lang.NegativeArraySizeException"
	case *ClassCastException:
		return "java.lang.ClassCastException"
	case *OutOfMemoryError:
		return "java.lang.OutOfMemoryError"
	case *NullPointerException:
		return "java.lang.NullPointerException"
	case *NetConnectException:
		return "java.net.ConnectException"
	case *NetBindException:
		return "java.net.BindException"
	case *NotBoundException:
		return "java.rmi.NotBoundException"
	case *MalformedURLException:
		return "java.net.MalformedURLException"
	case *InterruptedException:
		return "java.lang.InterruptedException"
	case *RMIUnknownHostException:
		return "java.rmi.UnknownHostException"
	case *ConnectIOException:
		return "java.rmi.ConnectIOException"
	case *NoRouteToHostException:
		return "java.net.NoRouteToHostException"
	case *RejectedExecutionException:
		return "java.util.concurrent.RejectedExecutionException"
	case *SimulationWorkerError:
		return "tlc2.tool.SimulationWorker$SimulationWorkerError"
	case *FingerprintException:
		return "tlc2.tool.FingerprintException"
	case *EvalException:
		return "tlc2.tool.EvalException"
	case *StatefulRuntimeException:
		return "tlc2.tool.StatefulRuntimeException"
	case *WorkerException:
		return "tlc2.tool.WorkerException"
	case *ConfigFileException, *ConfigError:
		return "tlc2.tool.ConfigFileException"
	case *LiveCounterExampleException:
		return "tlc2.tool.liveness.LiveCounterExampleException"
	case *LiveException:
		return "tlc2.tool.liveness.LiveException"
	case *TLCError:
		switch javaSystemFailureCode(failure) {
		case ECSystemStackOverflow:
			return "java.lang.StackOverflowError"
		case ECSystemOutOfMemory, ECSystemOutOfMemoryLiveness:
			return "java.lang.OutOfMemoryError"
		case ECTLCBug:
			return "java.lang.AssertionError"
		}
		if isValueEvalException(failure) {
			return "tlc2.tool.EvalException"
		}
		return "util.Assert$TLCRuntimeException"
	default:
		// Foreign Go errors retain their actual type rather than inventing a
		// Java class for an exception whose provenance is not represented.
		return fmt.Sprintf("%T", err)
	}
}

func javaThrowableString(err error) string {
	name := javaThrowableClassName(err)
	if message := javaThrowableDetailMessage(err); message != nil {
		return name + ": " + *message
	}
	return name
}

// Runtime frames are captured when the ported exception is constructed, like
// Throwable.fillInStackTrace. Frame names and files describe the Go port.
type throwableTrace struct {
	pcs []uintptr
}

func captureThrowableTrace() throwableTrace {
	pcs := make([]uintptr, 32)
	for {
		n := runtime.Callers(2, pcs)
		if n < len(pcs) {
			return throwableTrace{pcs: pcs[:n]}
		}
		pcs = make([]uintptr, len(pcs)*2)
	}
}

func (trace throwableTrace) throwablePCs() []uintptr { return trace.pcs }

func javaThrowableCause(err error) error {
	if failure, ok := err.(*FingerprintException); ok {
		// Java links additional fingerprint heads via next, not getCause().
		// Their root cause belongs only to the final head.
		if failure != nil {
			return failure.Cause
		}
		return nil
	}
	if throwable, ok := err.(interface{ Unwrap() error }); ok {
		return throwable.Unwrap()
	}
	return nil
}

func javaThrowableStackTrace(err error) string {
	var output strings.Builder
	var seen []error
	var write func(error, []uintptr, bool)
	write = func(failure error, enclosing []uintptr, cause bool) {
		for _, prior := range seen {
			if reflect.TypeOf(failure).Comparable() && failure == prior {
				output.WriteString("\t[CIRCULAR REFERENCE: " + javaThrowableString(failure) + "]\n")
				return
			}
		}
		seen = append(seen, failure)
		if cause {
			output.WriteString("Caused by: ")
		}
		output.WriteString(javaThrowableString(failure) + "\n")
		var pcs []uintptr
		if trace, ok := failure.(interface{ throwablePCs() []uintptr }); ok {
			pcs = trace.throwablePCs()
		}
		common := 0
		for i, j := len(pcs)-1, len(enclosing)-1; i >= 0 && j >= 0 && pcs[i] == enclosing[j]; i, j = i-1, j-1 {
			common++
		}
		if len(pcs) > common {
			frames := runtime.CallersFrames(pcs[:len(pcs)-common])
			for {
				frame, more := frames.Next()
				fmt.Fprintf(&output, "\tat %s(%s:%d)\n", frame.Function, frame.File, frame.Line)
				if !more {
					break
				}
			}
		}
		if common > 0 {
			fmt.Fprintf(&output, "\t... %d more\n", common)
		}
		if nested := javaThrowableCause(failure); nested != nil {
			write(nested, pcs, true)
		}
	}
	if err != nil {
		write(err, nil, false)
	}
	return output.String()
}

func copyNullableMessageParameters(params []*string) []*string {
	if params == nil {
		return nil
	}
	copied := make([]*string, len(params))
	for i, param := range params {
		if param != nil {
			copied[i] = javaString(*param)
		}
	}
	return copied
}

func messageParameterStrings(params []*string) []string {
	if params == nil {
		return nil
	}
	values := make([]string, len(params))
	for i, param := range params {
		if param != nil {
			values[i] = *param
		}
	}
	return values
}

func copyMessageParameters(params []string) []string {
	if params == nil {
		return nil
	}
	copied := make([]string, len(params))
	copy(copied, params)
	return copied
}
