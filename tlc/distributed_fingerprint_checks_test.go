package tlc

import (
	"math"
	"os"
	"strings"
	"testing"
)

type distributedCheckFailureEndpoint struct {
	*LocalFingerprintEndpoint
	failure error
	panics  bool
}

func TestDistributedFingerprintChecksUncheckedCompletion(t *testing.T) {
	for _, failure := range []error{NewRuntimeException("unchecked fingerprint check"), NewAssertionError("fatal fingerprint check")} {
		for _, panics := range []bool{false, true} {
			func() {
				captureFailoverToolIO(t, ToolIOTool)
				output, err := os.CreateTemp(t.TempDir(), "check-stderr-")
				if err != nil {
					t.Fatal(err)
				}
				previousStderr := os.Stderr
				os.Stderr = output
				defer func() { os.Stderr = previousStderr; _ = output.Close() }()
				endpoint := &distributedCheckFailureEndpoint{NewLocalFingerprintEndpoint(NewMemFPSet()), failure, panics}
				manager := NewDistributedFPSetManager(endpoint)
				if got := manager.CheckFPs(); got != math.MaxInt64 {
					t.Fatalf("failed completion = %d, want MaxInt64", got)
				}
				// Source manager prints failed task completions and continues;
				// these unchecked failures are not callable I/O false results.
				if !manager.CheckInvariant() {
					t.Fatal("unchecked failure entered callable I/O catch")
				}
				if messages := ToolIOGetAllMessages(); len(messages) != 0 {
					t.Fatalf("unchecked failure printed GENERAL: %q", messages)
				}
				data, err := os.ReadFile(output.Name())
				if err != nil {
					t.Fatal(err)
				}
				if strings.Count(string(data), "java.util.concurrent.ExecutionException:") != 2 || !strings.Contains(string(data), failure.Error()) {
					t.Fatalf("failed completion diagnostic = %s", data)
				}
			}()
		}
	}
}

func TestDistributedFingerprintChecksIOOverTCP(t *testing.T) {
	for _, panics := range []bool{false, true} {
		t.Run(map[bool]string{false: "returned", true: "panicked"}[panics], func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			endpoint := &distributedCheckFailureEndpoint{NewLocalFingerprintEndpoint(NewMemFPSet()), NewIOException("remote check I/O failure"), panics}
			_, client := startFingerprintRPC(t, endpoint)
			manager := NewDistributedFPSetManager(client)
			if got := manager.CheckFPs(); got != math.MaxInt64 {
				t.Fatalf("remote check = %d, want MaxInt64", got)
			}
			if manager.CheckInvariant() {
				t.Fatal("remote I/O failure incorrectly passed invariant check")
			}
			messages := ToolIOGetAllMessages()
			if len(messages) != 2 || !strings.Contains(messages[0], "remote check I/O failure") || !strings.Contains(messages[1], "remote check I/O failure") {
				t.Fatalf("remote check diagnostics = %q", messages)
			}
			if !manager.entry(0).available || manager.entry(0).set != client {
				t.Fatal("remote check failure changed partition ownership")
			}
			if size, err := client.Size(); err != nil || size != 0 {
				t.Fatalf("remote check failure stopped host: %d/%v", size, err)
			}
		})
	}
}

func (e *distributedCheckFailureEndpoint) checkFailure() error {
	if e.panics {
		panic(e.failure)
	}
	return e.failure
}
func (e *distributedCheckFailureEndpoint) CheckFPs() (uint64, error) {
	return 0, e.checkFailure()
}
func (e *distributedCheckFailureEndpoint) CheckInvariant(...uint64) (bool, error) {
	return true, e.checkFailure()
}

// No original method covers CheckFPsCallable/CheckInvariantCallable failures.
// Endpoint errors and panics must reach their source IOException-only catches.
func TestDistributedFingerprintChecksCatchIO(t *testing.T) {
	for _, panics := range []bool{false, true} {
		for _, operation := range []string{"fingerprints", "invariant"} {
			t.Run(operation+map[bool]string{false: "/returned", true: "/panicked"}[panics], func(t *testing.T) {
				captureFailoverToolIO(t, ToolIOTool)
				endpoint := &distributedCheckFailureEndpoint{NewLocalFingerprintEndpoint(NewMemFPSet()), NewIOException("fingerprint check I/O failure"), panics}
				manager := NewDistributedFPSetManager(endpoint)
				if operation == "fingerprints" {
					if got := manager.CheckFPs(); got != math.MaxInt64 {
						t.Fatalf("failed check = %d, want MaxInt64", got)
					}
				} else if manager.CheckInvariant() {
					t.Fatal("I/O failure incorrectly passed invariant check")
				}
				messages := ToolIOGetAllMessages()
				if len(messages) != 1 || !strings.Contains(messages[0], "fingerprint check I/O failure") {
					t.Fatalf("check diagnostic = %q", messages)
				}
				if !manager.entry(0).available || manager.entry(0).set != endpoint {
					t.Fatal("check failure reassigned or disabled fingerprint storage")
				}
			})
		}
	}
}

// A null endpoint is captured by a callable; a null registration fails during
// submission instead. Source catches failed completions and retains other results.
func TestDistributedFingerprintChecksNullEndpointCompletion(t *testing.T) {
	for _, operation := range []string{"fingerprints", "invariant"} {
		t.Run(operation, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			output, err := os.CreateTemp(t.TempDir(), "null-check-stderr-")
			if err != nil {
				t.Fatal(err)
			}
			previousStderr := os.Stderr
			os.Stderr = output
			defer func() { os.Stderr = previousStderr; _ = output.Close() }()
			healthy := NewMemFPSet()
			healthy.Put(11)
			healthy.Put(18)
			manager := NewDistributedFPSetManager(NewLocalFingerprintEndpoint(healthy))
			if err := manager.RegisterFPSet(nil, "null-endpoint"); err != nil {
				t.Fatal(err)
			}
			first, second := manager.entry(0), manager.entry(1)
			if operation == "fingerprints" {
				if got, want := manager.CheckFPs(), healthy.CheckFPs(); got != want {
					t.Fatalf("check result %d, want %d", got, want)
				}
			} else if !manager.CheckInvariant() {
				t.Fatal("null completion entered callable I/O catch")
			}
			data, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(data), "java.util.concurrent.ExecutionException:") != 1 || !strings.Contains(string(data), "java.lang.NullPointerException") {
				t.Fatalf("completion diagnostic %s", data)
			}
			if len(ToolIOGetAllMessages()) != 0 || manager.entry(0) != first || manager.entry(1) != second || !first.available || !second.available {
				t.Fatal("failed completion changed registrations or entered I/O catch")
			}
		})
	}
}

func TestDistributedFingerprintChecksNullRegistrationEscapes(t *testing.T) {
	for _, operation := range []string{"fingerprints", "invariant"} {
		t.Run(operation, func(t *testing.T) {
			manager := NewDistributedFPSetManager()
			manager.fpSets = []*distributedFPSets{nil}
			defer func() {
				if failure := recover(); failure == nil {
					t.Fatal("null registration was treated as a failed task completion")
				} else if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("failure %T, want null registration failure", failure)
				}
			}()
			if operation == "fingerprints" {
				manager.CheckFPs()
			} else {
				manager.CheckInvariant()
			}
		})
	}
}
