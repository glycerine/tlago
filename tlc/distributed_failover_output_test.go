package tlc

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type failoverWarningEndpoint struct {
	*LocalFingerprintEndpoint
	failure error
}

func (e *failoverWarningEndpoint) Put(uint64) (bool, error)      { return false, e.failure }
func (e *failoverWarningEndpoint) Contains(uint64) (bool, error) { return false, e.failure }
func (e *failoverWarningEndpoint) PutBlock(*LongVec) (*BitVector, error) {
	return nil, e.failure
}
func (e *failoverWarningEndpoint) ContainsBlock(*LongVec) (*BitVector, error) {
	return nil, e.failure
}
func (e *failoverWarningEndpoint) Size() (uint64, error)          { return 0, e.failure }
func (e *failoverWarningEndpoint) GetStatesSeen() (uint64, error) { return 0, e.failure }

func captureFailoverToolIO(t *testing.T, mode int) {
	t.Helper()
	// Retain pending messages and native stream assignments from other tests.
	toolIO.Lock()
	oldMode, oldOut, oldErr := toolIO.mode, toolIO.out, toolIO.err
	oldOutHigh, oldErrHigh := toolIO.outHigh, toolIO.errHigh
	oldSystemOutHigh, oldSystemErrHigh := toolIOSystemOutHigh, toolIOSystemErrHigh
	oldCaptureOut, oldCaptureErr := toolIO.captureOut, toolIO.captureErr
	oldMessages, oldNext := toolIO.messages, toolIO.nextMessage
	toolIO.Unlock()
	t.Cleanup(func() {
		toolIO.Lock()
		defer toolIO.Unlock()
		toolIO.mode, toolIO.out, toolIO.err = oldMode, oldOut, oldErr
		toolIO.outHigh, toolIO.errHigh = oldOutHigh, oldErrHigh
		toolIOSystemOutHigh, toolIOSystemErrHigh = oldSystemOutHigh, oldSystemErrHigh
		toolIO.captureOut, toolIO.captureErr = oldCaptureOut, oldCaptureErr
		toolIO.messages, toolIO.nextMessage = oldMessages, oldNext
	})
	ToolIOSetMode(mode)
	ToolIOReset()
}

// Original manager tests exercise failover without asserting ToolIO routing.
// Preserve the source two println calls, including the embedded message newline.
func TestDistributedFailoverWarningsUseToolIO(t *testing.T) {
	for _, operation := range []struct {
		name string
		call func(*DistributedFPSetManager)
	}{
		{"put", func(m *DistributedFPSetManager) { m.Put(1) }},
		{"contains", func(m *DistributedFPSetManager) { m.Contains(1) }},
		{"put_block", func(m *DistributedFPSetManager) { m.PutBlock([]*LongVec{NewLongVecFrom([]int64{1})}) }},
		{"contains_block", func(m *DistributedFPSetManager) { m.ContainsBlock([]*LongVec{NewLongVecFrom([]int64{1})}) }},
		{"size", func(m *DistributedFPSetManager) { m.Size() }},
		{"states_seen", func(m *DistributedFPSetManager) { m.GetStatesSeen() }},
	} {
		for _, mode := range []int{ToolIOSystem, ToolIOTool} {
			for _, message := range []*string{nil, javaString("offline")} {
				t.Run(fmt.Sprintf("%s/mode=%d/null_message=%v", operation.name, mode, message == nil), func(t *testing.T) {
					captureFailoverToolIO(t, mode)
					var output, errors bytes.Buffer
					if mode == ToolIOSystem {
						ToolIOSetSystemStreams(&output, &errors)
					}
					failure := NewIOException()
					if message != nil {
						failure = NewIOException(*message)
					}
					endpoint := &failoverWarningEndpoint{NewLocalFingerprintEndpoint(NewMemFPSet()), failure}
					manager := NewDistributedFPSetManager(endpoint)
					manager.fpSets[0].hostname = "failed-fp"
					operation.call(manager)
					want := []string{
						fmt.Sprintf("Warning: Failed to connect from %s to the fp server at failed-fp.\n%s", manager.GetHostName(), javaNullableString(message)),
						"Warning: there is no fp server available.",
					}
					if mode == ToolIOTool {
						if got := ToolIOGetAllMessages(); !reflect.DeepEqual(got, want) {
							t.Fatalf("recorded warnings = %q, want %q", got, want)
						}
					} else if got := output.String(); got != want[0]+"\n"+want[1]+"\n" || errors.Len() != 0 {
						t.Fatalf("warning streams: out %q, err %q", got, errors.String())
					}
					if manager.NumOfAliveServers() != 0 {
						t.Fatal("warning routing changed failed-server reassignment")
					}
				})
			}
		}
	}
}

func TestFingerprintRPCFailoverWarningUsesToolIO(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	lostStorage, survivingStorage := NewMemFPSet(), NewMemFPSet()
	lostHost, lostEndpoint := startFingerprintRPC(t, NewLocalFingerprintEndpoint(lostStorage))
	_, survivingEndpoint := startFingerprintRPC(t, NewLocalFingerprintEndpoint(survivingStorage))
	manager := NewDistributedFPSetManager(lostEndpoint, survivingEndpoint)
	manager.fpSets[0].hostname = "lost-fp"
	manager.fpSets[1].hostname = "surviving-fp"
	// Close before any insertion is issued: this is an unambiguous endpoint
	// loss, not a retry of an ambiguously completed insertion.
	if err := lostHost.Close(); err != nil {
		t.Fatal(err)
	}
	if manager.Put(2) || !survivingStorage.Contains(2) || lostStorage.Size() != 0 || manager.NumOfAliveServers() != 1 {
		t.Fatal("native endpoint loss did not reassign the insertion to the surviving FP server")
	}
	messages := ToolIOGetAllMessages()
	wantPrefix := fmt.Sprintf("Warning: Failed to connect from %s to the fp server at lost-fp.\n", manager.GetHostName())
	if len(messages) != 1 || !strings.HasPrefix(messages[0], wantPrefix) || len(messages[0]) == len(wantPrefix) {
		t.Fatalf("native failover warning was not captured with its failure detail: %q", messages)
	}
}
