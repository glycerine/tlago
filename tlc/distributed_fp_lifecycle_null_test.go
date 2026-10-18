package tlc

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

// No direct upstream method covers lifecycle calls on nullable endpoints.
// Checkpoint.run catches only I/O; close catches ordinary failures and continues.
func TestDistributedFingerprintCheckpointNullEndpoint(t *testing.T) {
	for _, phase := range []string{"begin", "commit", "recover"} {
		t.Run(phase, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			var calls []string
			first := traversalRegistration("first", &calls)
			healthy := traversalRegistration("healthy", &calls)
			manager := NewDistributedFPSetManager()
			manager.fpSets = []*distributedFPSets{first, healthy}
			if phase == "commit" {
				first.set.(*traversalFingerprintEndpoint).callback = func(string) error {
					manager.mu.Lock()
					manager.fpSets[0] = &distributedFPSets{hostname: "replacement", available: true}
					manager.mu.Unlock()
					return nil
				}
			} else {
				first.set = nil
			}
			var err error
			if phase == "recover" {
				err = manager.Recover("job")
			} else {
				err = manager.Checkpoint("job")
			}
			if _, ok := err.(*NullPointerException); !ok {
				t.Fatalf("%s failure %T/%v, want unchecked null failure", phase, err, err)
			}
			want := []string(nil)
			if phase == "commit" {
				want = []string{"first.begin"}
			}
			if !reflect.DeepEqual(calls, want) || len(ToolIOGetAllMessages()) != 0 || !healthy.available || !first.available {
				t.Fatal("null lifecycle call entered I/O catch, continued, or reassigned endpoints")
			}
		})
	}
}

func TestDistributedFingerprintCloseNullEndpointContinues(t *testing.T) {
	var calls []string
	healthy := traversalRegistration("healthy", &calls)
	manager := NewDistributedFPSetManager()
	manager.fpSets = []*distributedFPSets{{hostname: "null", available: true}, healthy}
	output, err := os.CreateTemp(t.TempDir(), "close-stderr-")
	if err != nil {
		t.Fatal(err)
	}
	previousStderr := os.Stderr
	os.Stderr = output
	defer func() { os.Stderr = previousStderr; _ = output.Close() }()
	if err := manager.Close(true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "java.lang.NullPointerException") || strings.Contains(string(data), "runtime.errorString") || !reflect.DeepEqual(calls, []string{"healthy.exit"}) || !manager.entry(0).available {
		t.Fatalf("null close diagnostic/continuation changed: %s, calls %v", data, calls)
	}
}

func TestNonDistributedFingerprintLifecycleNullEndpoint(t *testing.T) {
	for _, operation := range []string{"checkpoint", "commit", "recover", "close"} {
		t.Run(operation, func(t *testing.T) {
			manager := NewNonDistributedFPSetManager(nil, "null", nil)
			err := invokeDistributedServerOperation(func() error {
				switch operation {
				case "checkpoint":
					return manager.Checkpoint("job")
				case "commit":
					return manager.CommitCheckpoint()
				case "recover":
					return manager.Recover("job")
				default:
					return manager.Close(true)
				}
			})
			if javaThrowableClassName(err) != "java.lang.NullPointerException" || !manager.entry(0).available || manager.managerIsBroken {
				t.Fatalf("local lifecycle failure %T/%v changed category or ownership", err, err)
			}
		})
	}
}
