package tlc

import (
	"fmt"
	"strings"
	"testing"
)

type fatalRPCWorker struct {
	*rpcTestWorker
	failure error
	panics  bool
}

func (w *fatalRPCWorker) fail() error {
	if w.panics {
		panic(w.failure)
	}
	return w.failure
}
func (w *fatalRPCWorker) GetNextStates([]*TLCStateMut) (*NextStateResult, error) {
	return nil, w.fail()
}
func (w *fatalRPCWorker) IsAlive() (bool, error)              { return false, w.fail() }
func (w *fatalRPCWorker) GetCacheRateRatio() (float64, error) { return 0, w.fail() }
func (w *fatalRPCWorker) Exit() error                         { return w.fail() }

type fatalRPCCoordinator struct {
	*LocalServerEndpoint
	failure error
	panics  bool
}

func (s *fatalRPCCoordinator) GetIrredPolyForFP() (uint64, error) {
	if s.panics {
		panic(s.failure)
	}
	return 0, s.failure
}

// No enabled Java test directly covers Go endpoint return forms. Fatal remote
// failures must retain the I/O boundary regardless of returns versus panics.
func TestDistributedRPCFatalFailuresUseRemoteBoundary(t *testing.T) {
	for _, operation := range []string{"next", "alive", "cache", "exit", "coordinator"} {
		for _, panics := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/panic=%v", operation, panics), func(t *testing.T) {
				fatal := NewAssertionError("fatal endpoint")
				cause := NewRuntimeException("original cause")
				fatal.Cause = cause
				fatal.addSuppressedError(cause)
				var err error
				if operation == "coordinator" {
					_, client := startCoordinatorRPC(t, &fatalRPCCoordinator{NewLocalServerEndpoint(&TLCServer{FileName: "Spec"}), fatal, panics})
					_, err = client.GetIrredPolyForFP()
					if spec, failure := client.GetSpecFileName(); failure != nil || spec != "Spec" {
						t.Fatalf("coordinator failed after remote failure: %q, %v", spec, failure)
					}
				} else {
					_, client := startWorkerRPC(t, &fatalRPCWorker{&rpcTestWorker{}, fatal, panics})
					switch operation {
					case "next":
						_, err = client.GetNextStates([]*TLCStateMut{})
					case "alive":
						_, err = client.IsAlive()
					case "cache":
						_, err = client.GetCacheRateRatio()
					case "exit":
						err = client.Exit()
					}
					if uri, failure := client.GetURI(); failure != nil || uri != "tcp://worker:1234/primary" {
						t.Fatalf("worker failed or unpublished after remote failure: %q, %v", uri, failure)
					}
				}
				remote, ok := err.(*DistributedOperationError)
				if !ok || !isDistributedRemoteFailure(err) || !isJavaIOException(err) || isRecoverableDistributedError(err) {
					t.Fatalf("fatal remote failure lost boundary: %T %v", err, err)
				}
				storage, ok := remote.Cause.(*DistributedOperationError)
				if !ok || storage.Class != javaThrowableClassName(fatal) || storage.Error() != "fatal endpoint" || !strings.Contains(storage.Stack, "distributed_rpc_fatal_failure_test.go") {
					t.Fatalf("original fatal diagnostic graph lost: %#v", remote.Cause)
				}
				if len(storage.Suppressed) != 1 || storage.Suppressed[0] != storage.Cause || storage.Cause == cause || storage.Cause.Error() != "original cause" {
					t.Fatal("cause/suppressed sharing or receiver ownership changed")
				}
			})
		}
	}
}
