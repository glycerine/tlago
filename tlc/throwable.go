package tlc

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
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
	if diagnostic, ok := err.(interface{ diagnosticClassName() string }); ok {
		return diagnostic.diagnosticClassName()
	}
	switch failure := err.(type) {
	case *FrontEndException:
		return "tla2sany.drivers.FrontEndException"
	case *NamingException:
		return "javax.naming.NamingException"
	case *UnsupportedEncodingException:
		return "java.io.UnsupportedEncodingException"
	case *IllegalStateException:
		return "java.lang.IllegalStateException"
	case *NumberFormatException:
		return "java.lang.NumberFormatException"
	case *NoSuchElementException:
		return "java.util.NoSuchElementException"
	case *ExceptionInInitializerError:
		return "java.lang.ExceptionInInitializerError"
	case *NoClassDefFoundError:
		return "java.lang.NoClassDefFoundError"
	case *ClassNotFoundException:
		return "java.lang.ClassNotFoundException"
	case *IllegalAccessException:
		return "java.lang.IllegalAccessException"
	case *InstantiationException:
		return "java.lang.InstantiationException"
	case *NoSuchMethodException:
		return "java.lang.NoSuchMethodException"
	case *InvocationTargetException:
		return "java.lang.reflect.InvocationTargetException"
	case *AssertionError:
		return "java.lang.AssertionError"
	case *JavaError:
		return "java.lang.Error"
	case *SecurityException:
		return "java.lang.SecurityException"
	case *ConcurrentModificationException:
		return "java.util.ConcurrentModificationException"
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
	case *CharacterCodingException:
		return "java.nio.charset.CharacterCodingException"
	case *MalformedInputException:
		return "java.nio.charset.MalformedInputException"
	case *UnmappableCharacterException:
		return "java.nio.charset.UnmappableCharacterException"
	case *UnsupportedCharsetException:
		return "java.nio.charset.UnsupportedCharsetException"
	case *IllegalCharsetNameException:
		return "java.nio.charset.IllegalCharsetNameException"
	case *IOException:
		return "java.io.IOException"
	case *FileNotFoundException:
		return "java.io.FileNotFoundException"
	case *UnknownHostException:
		return "java.net.UnknownHostException"
	case *InvalidPathException:
		return "java.nio.file.InvalidPathException"
	case *FileSystemException:
		return "java.nio.file.FileSystemException"
	case *NoSuchFileException:
		return "java.nio.file.NoSuchFileException"
	case *AccessDeniedException:
		return "java.nio.file.AccessDeniedException"
	case *FileAlreadyExistsException:
		return "java.nio.file.FileAlreadyExistsException"
	case *IndexOutOfBoundsException:
		return "java.lang.IndexOutOfBoundsException"
	case *StringIndexOutOfBoundsException:
		return "java.lang.StringIndexOutOfBoundsException"
	case *IllegalFormatException:
		return "java.util.IllegalFormatException"
	case *DuplicateFormatFlagsException:
		return "java.util.DuplicateFormatFlagsException"
	case *FormatFlagsConversionMismatchException:
		return "java.util.FormatFlagsConversionMismatchException"
	case *IllegalFormatConversionException:
		return "java.util.IllegalFormatConversionException"
	case *IllegalFormatFlagsException:
		return "java.util.IllegalFormatFlagsException"
	case *IllegalFormatPrecisionException:
		return "java.util.IllegalFormatPrecisionException"
	case *IllegalFormatWidthException:
		return "java.util.IllegalFormatWidthException"
	case *MissingFormatArgumentException:
		return "java.util.MissingFormatArgumentException"
	case *MissingFormatWidthException:
		return "java.util.MissingFormatWidthException"
	case *UnknownFormatConversionException:
		return "java.util.UnknownFormatConversionException"
	case *IllegalFormatArgumentIndexException:
		return "java.util.IllegalFormatArgumentIndexException"
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
	case *InvariantViolatedException:
		return "tlc2.tool.INextStateFunctor$InvariantViolatedException"
	case *DoInitInvariantViolatedException:
		return "tlc2.tool.ModelChecker$DoInitFunctor$InvariantViolatedException"
	case *WorkerWrappingRuntimeException:
		return "tlc2.tool.Worker$WrappingRuntimeException"
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
		if failure.Expr != nil {
			return "util.Assert$TLCDetailedRuntimeException"
		}
		return "util.Assert$TLCRuntimeException"
	default:
		// Foreign Go errors retain their actual type rather than inventing a
		// Java class for an exception whose provenance is not represented.
		return fmt.Sprintf("%T", err)
	}
}

func javaThrowableString(err error) string {
	if override, ok := err.(interface{ JavaThrowableString() string }); ok {
		return override.JavaThrowableString()
	}
	return javaBasicThrowableString(err)
}

func javaBasicThrowableString(err error) string {
	name := javaThrowableClassName(err)
	if message := javaThrowableDetailMessage(err); message != nil {
		return name + ": " + *message
	}
	return name
}

// Runtime frames are captured when the ported exception is constructed, like
// Throwable.fillInStackTrace. Frame names and files describe the Go port.
type throwableTrace struct {
	pcs         []uintptr
	suppression *throwableSuppression
	remoteStack string
}

func (trace throwableTrace) remoteThrowableStack() string { return trace.remoteStack }

type throwableSuppression struct {
	mu     sync.Mutex
	errors []error
}

func captureThrowableTrace() throwableTrace {
	pcs := make([]uintptr, 32)
	for {
		n := runtime.Callers(2, pcs)
		if n < len(pcs) {
			return throwableTrace{pcs: pcs[:n], suppression: &throwableSuppression{}}
		}
		pcs = make([]uintptr, len(pcs)*2)
	}
}

func (trace throwableTrace) throwablePCs() []uintptr { return trace.pcs }

func (trace throwableTrace) GetSuppressed() []error {
	if trace.suppression == nil {
		return []error{}
	}
	trace.suppression.mu.Lock()
	defer trace.suppression.mu.Unlock()
	return append([]error{}, trace.suppression.errors...)
}

func (trace throwableTrace) addSuppressedError(err error) {
	if trace.suppression != nil {
		trace.suppression.mu.Lock()
		defer trace.suppression.mu.Unlock()
		trace.suppression.errors = append(trace.suppression.errors, err)
	}
}

// Used by source try-with-resources paths after both operations have failed.
func javaSuppressCloseError(primary, secondary error) {
	if primary != nil && secondary != nil {
		if throwable, ok := primary.(interface{ addSuppressedError(error) }); ok {
			throwable.addSuppressedError(secondary)
		}
	}
}

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
	if remote, ok := err.(interface{ remoteThrowableStack() string }); ok && remote.remoteThrowableStack() != "" {
		return remote.remoteThrowableStack()
	}
	var output strings.Builder
	var seen []error
	var write func(error, []uintptr, string, string)
	write = func(failure error, enclosing []uintptr, caption, indent string) {
		for _, prior := range seen {
			if reflect.TypeOf(failure).Comparable() && failure == prior {
				output.WriteString(indent + caption + "[CIRCULAR REFERENCE: " + javaThrowableString(failure) + "]\n")
				return
			}
		}
		seen = append(seen, failure)
		output.WriteString(indent + caption + javaThrowableString(failure) + "\n")
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
				fmt.Fprintf(&output, "%s\tat %s(%s:%d)\n", indent, frame.Function, frame.File, frame.Line)
				if !more {
					break
				}
			}
		}
		if common > 0 {
			fmt.Fprintf(&output, "%s\t... %d more\n", indent, common)
		}
		if throwable, ok := failure.(interface{ GetSuppressed() []error }); ok {
			for _, suppressed := range throwable.GetSuppressed() {
				write(suppressed, pcs, "Suppressed: ", indent+"\t")
			}
		}
		if nested := javaThrowableCause(failure); nested != nil {
			write(nested, pcs, "Caused by: ", indent)
		}
	}
	if err != nil {
		write(err, nil, "", "")
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

// JavaThrowableString exposes Throwable.toString to the parser integration.
func JavaThrowableString(err error) string { return javaThrowableString(err) }

// JavaThrowableStackTrace exposes the captured port frames at parser boundaries.
func JavaThrowableStackTrace(err error) string { return javaThrowableStackTrace(err) }
