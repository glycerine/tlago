package tlc

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

// Java TLC has no dedicated worker-URI test. These expectations were captured
// from URI.create/getHost/toASCIIString on the local OpenJDK 21.0.12.1, using
// TLCWorker's authority/path construction. The user requires native TCP rather
// than RMI: only the three-character scheme is changed in expected diagnostics.
// All authority, IPv4/IPv6, scope, UTF-16 index and NFC checks remain exact.
func TestDistributedWorkerURIMatchesJava(t *testing.T) {
	var fixtures []struct {
		Hostname string  `json:"hostname"`
		Port     int     `json:"port"`
		ThreadID int     `json:"threadID"`
		Host     *string `json:"host"`
		ASCII    string  `json:"ascii"`
		Error    string  `json:"error"`
		Index    int     `json:"index"`
		Reason   string  `json:"reason"`
	}
	data, err := os.ReadFile("test_vectors/worker_uri_java21.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	for i, fixture := range fixtures {
		fixture.Error = strings.ReplaceAll(fixture.Error, "rmi://", "tcp://")
		fixture.ASCII = strings.ReplaceAll(fixture.ASCII, "rmi://", "tcp://")
		t.Run(fmt.Sprintf("%03d_%q", i, fixture.Hostname), func(t *testing.T) {
			var uri *distributedWorkerURIValue
			var failure any
			func() {
				defer func() { failure = recover() }()
				uri = newDistributedWorkerURI(DistributedWorkerAddress{Hostname: fixture.Hostname, Port: fixture.Port}, fixture.ThreadID)
			}()
			if fixture.Error != "" {
				argument, ok := failure.(*IllegalArgumentException)
				if !ok {
					t.Fatalf("failure = %T, want IllegalArgumentException", failure)
				}
				if got := argument.GetMessage(); got == nil || *got != fixture.Error {
					t.Fatalf("message = %v, want %q", got, fixture.Error)
				}
				syntax, ok := argument.GetCause().(*URISyntaxException)
				if !ok {
					t.Fatalf("cause = %T, want URISyntaxException", argument.GetCause())
				}
				if syntax.GetIndex() != fixture.Index || syntax.GetReason() != fixture.Reason {
					t.Fatalf("reason/index = %q/%d, want %q/%d", syntax.GetReason(), syntax.GetIndex(), fixture.Reason, fixture.Index)
				}
				return
			}
			if failure != nil {
				t.Fatalf("unexpected failure: %v", failure)
			}
			if (uri.host == nil) != (fixture.Host == nil) || uri.host != nil && *uri.host != *fixture.Host {
				t.Fatalf("host = %v, want %v", uri.host, fixture.Host)
			}
			if got := uri.asciiString(); got != fixture.ASCII {
				t.Fatalf("ASCII = %q, want %q", got, fixture.ASCII)
			}
		})
	}
}

func TestDistributedWorkerURIDiagnosticsAndRegistration(t *testing.T) {
	manager := NewDistributedFPSetManager(NewLocalFingerprintEndpoint(NewMemFPSet()))
	worker := NewDistributedWorker(7, nil, manager, DistributedWorkerAddress{Hostname: "e\u0301.example", Port: 10997})
	if got, want := worker.GetURI(), "tcp://e\u0301.example:10997/7"; got != want {
		t.Fatalf("URI = %q, want %q", got, want)
	}
	if got, want := distributedWorkerURI(worker), worker.GetURI(); got != want {
		t.Fatalf("registration URI = %q, want %q", got, want)
	}
	worker.Runtime.executor.Shutdown()
	proxy := NewDistributedWorkerSmartProxy(NewLocalWorkerEndpoint(worker))
	_, err := proxy.GetNextStates([]*TLCStateMut{})
	remote, ok := err.(*DistributedOperationError)
	if !ok {
		t.Fatalf("proxy error = %T, want DistributedOperationError", err)
	}
	if got, want := *remote.Message, "Executor rejected task at worker: tcp://%C3%A9.example:10997/7"; got != want {
		t.Fatalf("worker message = %q, want %q", got, want)
	}
	if _, ok := remote.Cause.(*RejectedExecutionException); !ok {
		t.Fatalf("cause = %T", remote.Cause)
	}
}
