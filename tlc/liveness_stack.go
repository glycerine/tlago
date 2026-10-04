package tlc

import "unsafe"

// The reference OpenJDK runtime gives each thread a 1024 KiB execution stack.
// Bound native stack bytes, rather than the number of recursive operators, so
// resource exhaustion remains catchable as StackOverflowError in Go. Go's own
// eventual goroutine-stack overflow would terminate the entire process.
const livenessExecutionStackBytes = 1024 * 1024

//go:noinline
func checkLivenessStack(anchor *byte) {
	var here byte
	// Both pointers stay on the current goroutine's stack and are relocated by
	// Go when it grows. Convert only for this calculation, retaining no uintptr
	// across a call or stack relocation and never converting it back to a pointer.
	first, current := uintptr(unsafe.Pointer(anchor)), uintptr(unsafe.Pointer(&here))
	var used uintptr
	if first >= current {
		used = first - current
	} else {
		used = current - first
	}
	if used >= livenessExecutionStackBytes {
		// Java Error bypasses Liveness.astToLiveAppl's catch(Exception), reaching
		// TLC.process's existing StackOverflowError catch and specific diagnostic.
		panic(NewStackOverflowError())
	}
}
