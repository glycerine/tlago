package tlc

import (
	"bytes"
	"reflect"
	"testing"
)

// The upstream manager suite has no checkpoint/close registration-change tests.
// These focused checks exercise the live lookup order in FPSetManager itself.
type traversalFingerprintEndpoint struct {
	*LocalFingerprintEndpoint
	name     string
	calls    *[]string
	callback func(string) error
}

func (e *traversalFingerprintEndpoint) call(phase string) error {
	*e.calls = append(*e.calls, e.name+"."+phase)
	if e.callback != nil {
		return e.callback(phase)
	}
	return nil
}
func (e *traversalFingerprintEndpoint) BeginChkptFile(string) error  { return e.call("begin") }
func (e *traversalFingerprintEndpoint) CommitChkptFile(string) error { return e.call("commit") }
func (e *traversalFingerprintEndpoint) RecoverFile(string) error     { return e.call("recover") }
func (e *traversalFingerprintEndpoint) Exit(bool) error              { return e.call("exit") }

func traversalRegistration(name string, calls *[]string) *distributedFPSets {
	return &distributedFPSets{set: &traversalFingerprintEndpoint{
		LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(NewMemFPSet()), name: name, calls: calls,
	}, hostname: name, available: true}
}

func TestDistributedCheckpointReadsLiveRegistrations(t *testing.T) {
	for _, checkpoint := range []bool{true, false} {
		t.Run(map[bool]string{true: "checkpoint", false: "recover"}[checkpoint], func(t *testing.T) {
			var calls []string
			a, b, c := traversalRegistration("a", &calls), traversalRegistration("b", &calls), traversalRegistration("c", &calls)
			d, extra := traversalRegistration("d", &calls), traversalRegistration("extra", &calls)
			manager := &DistributedFPSetManager{fpSets: []*distributedFPSets{nil, a, b, a}}
			a.set.(*traversalFingerprintEndpoint).callback = func(phase string) error {
				if phase == "begin" || phase == "recover" {
					manager.mu.Lock()
					manager.fpSets[1] = d                          // Commit must resolve this replacement.
					manager.fpSets[2] = c                          // Later traversal must use the live slot.
					manager.fpSets[3] = b                          // Tail selection occurs after the first operation.
					manager.fpSets = append(manager.fpSets, extra) // Initial length stays fixed.
					manager.mu.Unlock()
				}
				return nil
			}
			if err := manager.checkpointInner("job", checkpoint); err != nil {
				t.Fatal(err)
			}
			want := []string{"a.recover", "c.recover", "b.recover"}
			if checkpoint {
				want = []string{"a.begin", "d.commit", "c.begin", "c.commit", "b.begin", "b.commit"}
			}
			if !reflect.DeepEqual(calls, want) {
				t.Fatalf("calls %v, want %v", calls, want)
			}
		})
	}
}

func TestDistributedCheckpointReportsLiveFailureHostname(t *testing.T) {
	for _, panics := range []bool{false, true} {
		t.Run(map[bool]string{false: "return", true: "panic"}[panics], func(t *testing.T) {
			var calls []string
			var output bytes.Buffer
			oldMode := ToolIOGetMode()
			ToolIOSetMode(ToolIOSystem)
			defer ToolIOSetMode(oldMode)
			defer ToolIOSetSystemStreams(&output, &output)()
			a, b := traversalRegistration("a", &calls), traversalRegistration("b", &calls)
			manager := &DistributedFPSetManager{fpSets: []*distributedFPSets{a, a}}
			a.set.(*traversalFingerprintEndpoint).callback = func(string) error {
				manager.mu.Lock()
				manager.fpSets[0], manager.fpSets[1] = b, b
				manager.mu.Unlock()
				failure := NewIOException("unavailable")
				if panics {
					panic(failure)
				}
				return failure
			}
			if err := manager.Checkpoint("job"); err != nil {
				t.Fatal(err)
			}
			if got, want := output.String(), "Error: Failed to checkpoint the fingerprint server at b. This server might be down.\n"; got != want {
				t.Fatalf("output %q, want %q", got, want)
			}
			want := []string{"a.begin", "b.begin", "b.commit"}
			if !reflect.DeepEqual(calls, want) {
				t.Fatalf("calls %v, want %v", calls, want)
			}
		})
	}
}

func TestDistributedCloseCapturesNextBeforeExit(t *testing.T) {
	var calls []string
	a, b, c, d := traversalRegistration("a", &calls), traversalRegistration("b", &calls), traversalRegistration("c", &calls), traversalRegistration("d", &calls)
	manager := &DistributedFPSetManager{fpSets: []*distributedFPSets{nil, a, b, c, a}}
	a.set.(*traversalFingerprintEndpoint).callback = func(string) error {
		manager.mu.Lock()
		manager.fpSets[2], manager.fpSets[3], manager.fpSets[4] = d, d, d
		manager.fpSets = append(manager.fpSets, c)
		manager.mu.Unlock()
		return nil
	}
	if err := manager.Close(true); err != nil {
		t.Fatal(err)
	}
	// The captured b survives replacement. Slot 3 is read live. Slot 4 was
	// trimmed before exit, and the appended slot lies outside the initial length.
	want := []string{"a.exit", "b.exit", "d.exit"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls %v, want %v", calls, want)
	}
}

func TestDistributedRegistrationTraversalUsesWrapperIdentity(t *testing.T) {
	for _, operation := range []string{"checkpoint", "recover", "close"} {
		t.Run(operation, func(t *testing.T) {
			var calls []string
			a, b := traversalRegistration("a", &calls), traversalRegistration("b", &calls)
			anotherA := &distributedFPSets{set: a.set, hostname: a.hostname, available: true}
			manager := &DistributedFPSetManager{fpSets: []*distributedFPSets{nil, a, a, nil, b, anotherA, a, nil}}
			var err error
			var want []string
			switch operation {
			case "checkpoint":
				err = manager.Checkpoint("job")
				want = []string{"a.begin", "a.commit", "b.begin", "b.commit", "a.begin", "a.commit"}
			case "recover":
				err = manager.Recover("job")
				want = []string{"a.recover", "b.recover", "a.recover"}
			case "close":
				err = manager.Close(true)
				want = []string{"a.exit", "b.exit", "a.exit"}
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(calls, want) {
				t.Fatalf("calls %v, want %v", calls, want)
			}
		})
	}
}
