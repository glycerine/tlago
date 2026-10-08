package tlc

import "testing"

// TLCServerThread's source labels are consumed by external statistics tools.
// No upstream method directly tests their prefix or ASCII URI rendering.
func TestDistributedCoordinatorWorkerLabels(t *testing.T) {
	for _, test := range []struct {
		id   int
		uri  string
		want string
	}{
		{0, "tcp://worker:1234/0", "TLCWorkerThread-000-[tcp://worker:1234/0]"},
		{7, "tcp://worker:1234/7", "TLCWorkerThread-007-[tcp://worker:1234/7]"},
		{999, "tcp://worker:1234/999", "TLCWorkerThread-999-[tcp://worker:1234/999]"},
		{1000, "tcp://worker:1234/1000", "TLCWorkerThread-1000-[tcp://worker:1234/1000]"},
		{-1, "tcp://worker:1234/-1", "TLCWorkerThread--01-[tcp://worker:1234/-1]"},
		{1, "tcp://worker:1234/café", "TLCWorkerThread-001-[tcp://worker:1234/caf%C3%A9]"},
		{2, "tcp://worker:1234/cafe\u0301", "TLCWorkerThread-002-[tcp://worker:1234/caf%C3%A9]"},
		{3, "tcp://worker:1234/😀", "TLCWorkerThread-003-[tcp://worker:1234/%F0%9F%98%80]"},
		{4, "tcp://worker:1234/%ff", "TLCWorkerThread-004-[tcp://worker:1234/%ff]"},
	} {
		t.Run(test.uri, func(t *testing.T) {
			thread := &TLCServerThread{ID: test.id, URI: test.uri}
			if got := thread.Name(); got != test.want {
				t.Fatalf("thread label = %q, want %q", got, test.want)
			}
			if thread.GetURI() != test.uri {
				t.Fatal("formatting a label modified endpoint metadata")
			}
		})
	}
}
