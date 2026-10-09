package tlc

import (
	"bytes"
	"encoding/gob"
	"io"
	"net"
	"net/rpc"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Observe the next socket read after the complete encoded header has been
// consumed. net/rpc has accepted that header and is now decoding its body.
type distributedPartialRequestListener struct {
	net.Listener
	headerBytes int
	bodyRead    chan struct{}
}

func (l *distributedPartialRequestListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &distributedPartialRequestConn{Conn: conn, remaining: l.headerBytes, bodyRead: l.bodyRead}, nil
}

type distributedPartialRequestConn struct {
	net.Conn
	remaining int
	bodyRead  chan struct{}
	once      sync.Once
}

func (c *distributedPartialRequestConn) Read(data []byte) (int, error) {
	if c.remaining == 0 {
		c.once.Do(func() { close(c.bodyRead) })
	}
	n, err := c.Conn.Read(data)
	c.remaining = max(0, c.remaining-n)
	return n, err
}

type distributedPartialRequestEndpoint struct {
	*LocalFingerprintEndpoint
	calls   atomic.Int32
	entered chan struct{}
	release chan struct{}
}

func (e *distributedPartialRequestEndpoint) Size() (uint64, error) {
	e.calls.Add(1)
	return e.LocalFingerprintEndpoint.Size()
}

func (e *distributedPartialRequestEndpoint) Exit(bool) error {
	close(e.entered)
	<-e.release
	return nil
}

// This is a native transport ownership check; no original Java method directly
// tests a Go RPC header whose sender never supplies the corresponding body.
func TestFingerprintRPCGracefulCloseIncompleteRequest(t *testing.T) {
	for _, accepted := range []bool{false, true} {
		name := "incomplete_only"
		if accepted {
			name = "accepted_exit_and_incomplete"
		}
		t.Run(name, func(t *testing.T) { checkFingerprintRPCGracefulCloseIncompleteRequest(t, accepted) })
	}
}

func checkFingerprintRPCGracefulCloseIncompleteRequest(t *testing.T, accepted bool) {
	t.Helper()
	var header bytes.Buffer
	encoder := gob.NewEncoder(&header)
	if accepted {
		if err := encoder.Encode(&rpc.Request{ServiceMethod: "Fingerprint.Call", Seq: 1}); err != nil {
			t.Fatal(err)
		}
		if err := encoder.Encode(DistributedFingerprintRequest{Object: "primary", Operation: "exit"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := encoder.Encode(&rpc.Request{ServiceMethod: "Fingerprint.Call", Seq: 2}); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	observed := &distributedPartialRequestListener{Listener: listener, headerBytes: header.Len(), bodyRead: make(chan struct{})}
	server := NewDistributedRPCServer()
	endpoint := &distributedPartialRequestEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(NewMemFPSet()), entered: make(chan struct{}), release: make(chan struct{})}
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(endpoint.release) }) }
	t.Cleanup(release)
	if err := server.RegisterFingerprint("primary", endpoint); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(observed) }()
	t.Cleanup(func() {
		_ = server.Close()
		if err := <-served; err != nil {
			t.Error(err)
		}
	})
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Write(header.Bytes()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-observed.bodyRead:
	case <-time.After(5 * time.Second):
		t.Fatal("server did not reach the missing request body")
	}
	if accepted {
		select {
		case <-endpoint.entered:
		case <-time.After(5 * time.Second):
			t.Fatal("complete Exit request did not enter its handler")
		}
	}
	closed := make(chan error, 1)
	go func() { closed <- server.CloseGracefully() }()
	if accepted {
		waitForNativeRPCClosing(t, server)
		select {
		case err := <-closed:
			t.Fatalf("graceful close abandoned accepted Exit: %v", err)
		default:
		}
		release()
	}
	select {
	case err := <-closed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		// Join the graceful owner even on failure; forced close releases the
		// blocked body read so a red test cannot leave background work behind.
		_ = server.Close()
		<-closed
		t.Fatal("incomplete request body prevented graceful shutdown")
	}
	if endpoint.calls.Load() != 0 {
		t.Fatal("incomplete request invoked fingerprint storage")
	}
	if accepted {
		if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
			t.Fatal(err)
		}
		decoder := gob.NewDecoder(conn)
		var gotExit bool
		for {
			var response rpc.Response
			if err := decoder.Decode(&response); err != nil {
				if err != io.EOF {
					t.Fatal(err)
				}
				break
			}
			var body DistributedFingerprintReply
			if response.Error != "" {
				if err := decoder.Decode(nil); err != nil {
					t.Fatal(err)
				}
			} else if err := decoder.Decode(&body); err != nil {
				t.Fatal(err)
			}
			if response.Seq == 1 {
				if gotExit || response.Error != "" || body.Failure != nil {
					t.Fatal("accepted Exit reply was duplicated or failed")
				}
				gotExit = true
			}
		}
		if !gotExit {
			t.Fatal("graceful shutdown dropped accepted Exit reply")
		}
	}
}
