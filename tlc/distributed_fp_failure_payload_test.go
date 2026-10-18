package tlc

import (
	"fmt"
	"strings"
	"testing"
)

type fingerprintGraphFailureEndpoint struct {
	*LocalFingerprintEndpoint
	failure error
	panics  bool
}

func (e *fingerprintGraphFailureEndpoint) Put(uint64) (bool, error) {
	if e.panics {
		panic(e.failure)
	}
	return false, e.failure
}

// No enabled upstream test directly checks remote FP failure graphs. Source
// remote calls preserve throwable diagnostics and local/remote catch behavior;
// the Go boundary carries those contracts without JVM exception reconstruction.
func TestFingerprintRPCPreservesFailureGraph(t *testing.T) {
	for _, family := range []string{"io", "null", "fatal"} {
		for _, panics := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/panic=%v", family, panics), func(t *testing.T) {
				cause := NewStatefulRuntimeException("storage cause")
				var original error
				switch family {
				case "io":
					failure := NewIOException()
					failure.Cause = cause
					failure.addSuppressedError(cause)
					original = failure
				case "null":
					failure := NewNullPointerException()
					failure.Cause = cause
					failure.addSuppressedError(cause)
					original = failure
				case "fatal":
					failure := NewAssertionError("fatal storage")
					failure.Cause = cause
					failure.addSuppressedError(cause)
					original = failure
				}
				endpoint := &fingerprintGraphFailureEndpoint{NewLocalFingerprintEndpoint(NewMemFPSet()), original, panics}
				_, client := startFingerprintRPC(t, endpoint)
				_, err := client.Put(1)
				failure, ok := err.(*DistributedOperationError)
				if !ok {
					t.Fatalf("failure graph was flattened: %T %v", err, err)
				}
				if isJavaIOException(failure) != (family != "null") || isDistributedNullFailure(failure) != (family == "null") || isDistributedRemoteFailure(failure) != (family == "fatal") {
					t.Fatalf("remote failure changed source catch categories: %#v", failure)
				}
				// A fatal remote call adds an I/O wrapper. The storage graph
				// remains its cause, rather than becoming a local fatal panic.
				storage := failure
				if family == "fatal" {
					storage, ok = failure.Cause.(*DistributedOperationError)
					if !ok {
						t.Fatal("fatal remote failure lost its storage cause")
					}
				}
				if storage.Class != javaThrowableClassName(original) || storage.Cause == nil || storage.Cause == cause || len(storage.Suppressed) != 1 || storage.Suppressed[0] != storage.Cause {
					t.Fatalf("failure class, sharing or receiver ownership changed: %#v", storage)
				}
				if family != "fatal" && storage.GetMessage() != nil {
					t.Fatal("null storage message became a diagnostic string")
				}
				if detail := javaThrowableDetailMessage(storage.Cause); detail == nil || *detail != "storage cause" || !strings.Contains(storage.Stack, "distributed_fp_failure_payload_test.go") {
					t.Fatal("sender cause detail or actual Go diagnostic stack was lost")
				}
				if count, err := client.Size(); err != nil || count != 0 {
					t.Fatalf("FP host failed after returning its storage failure: %d, %v", count, err)
				}
			})
		}
	}
}
