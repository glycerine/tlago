package tlc

import (
	"bytes"
	"testing"
)

// The original distributed tests have no direct coordinator-call failure
// cases. These focused tests cover the Go endpoint boundary added for transport.
func TestDistributedServerStatusEndpointFailure(t *testing.T) {
	failure := distributedTestRemoteFailure("status unavailable")
	endpoint := &failingServerEndpoint{statusFailure: failure}
	lookupCalls := 0
	lookup := NewTLCServerStatusLookup(func(url string) (DistributedServerEndpoint, error) {
		lookupCalls++
		if url != "//coordinator:10997/TLCServerWORKER" {
			t.Fatalf("lookup URL = %q", url)
		}
		return endpoint, nil
	})
	done, err := lookup("//coordinator:10997/TLCServerWORKER")
	if done || err != failure || lookupCalls != 1 || endpoint.statusCalls != 1 {
		t.Fatalf("done/error/lookup/status = %v/%v/%d/%d", done, err, lookupCalls, endpoint.statusCalls)
	}
}

func TestDistributedWorkerBootstrapPolynomialFailure(t *testing.T) {
	failure := distributedTestRemoteFailure("polynomial unavailable")
	endpoint := &failingServerEndpoint{polyFailure: failure}
	process := NewDistributedWorkerProcess()
	var output bytes.Buffer
	loaded := false
	oldPoly, oldIntern := FP64IrredPoly(), internTable
	err := process.start("coordinator", 1, DistributedWorkerEnvironment{
		ToolOut: &output,
		Lookup: func(string) (DistributedServerEndpoint, error) {
			return endpoint, nil
		},
		LoadApp: func(DistributedServerEndpoint, *DistributedFilenameToStreamResolver) (*TLCApp, error) {
			loaded = true
			return nil, nil
		},
	})
	if err != failure || endpoint.polyCalls != 1 || loaded {
		t.Fatalf("error/poly calls/load = %v/%d/%v", err, endpoint.polyCalls, loaded)
	}
	if FP64IrredPoly() != oldPoly || internTable != oldIntern || process.Group != nil || process.Resolver != nil {
		t.Fatal("failed polynomial call changed worker initialization")
	}
}

func TestLocalServerEndpointInternMetadataOwnership(t *testing.T) {
	server := &TLCServer{InternTable: NewInternTable(16)}
	coordinator := server.InternTable.Put("variable")
	coordinator.loc = 3
	endpoint := NewLocalServerEndpoint(server)
	worker, err := endpoint.Intern("variable")
	if err != nil {
		t.Fatal(err)
	}
	if worker == coordinator || worker.s != coordinator.s || worker.tok != coordinator.tok || worker.loc != 3 {
		t.Fatal("worker interning lost coordinator identity or shares mutable metadata")
	}
	worker.loc = 8
	if coordinator.loc != 3 {
		t.Fatal("worker location changed coordinator metadata")
	}
}

type failingServerEndpoint struct {
	DistributedServerEndpoint
	statusFailure, polyFailure error
	statusCalls, polyCalls     int
}

func (e *failingServerEndpoint) IsDone() (bool, error) {
	e.statusCalls++
	return false, e.statusFailure
}

func (e *failingServerEndpoint) GetIrredPolyForFP() (uint64, error) {
	e.polyCalls++
	return 0, e.polyFailure
}
