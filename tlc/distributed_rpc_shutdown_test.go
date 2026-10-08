package tlc

import (
	"testing"
	"time"
)

type gatedFingerprintExit struct {
	*LocalFingerprintEndpoint
	entered chan struct{}
	release chan struct{}
}

func (endpoint *gatedFingerprintExit) Exit(bool) error {
	close(endpoint.entered)
	<-endpoint.release
	return nil
}

// No direct enabled Java test covers this native reply/process boundary.
// The command's reporting loop can return while Exit is still in its handler.
func TestFingerprintRPCGracefulClosePreservesExitReply(t *testing.T) {
	endpoint := &gatedFingerprintExit{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(NewMemFPSet()), entered: make(chan struct{}), release: make(chan struct{})}
	server, client := startFingerprintRPC(t, endpoint)
	exit := make(chan error, 1)
	go func() { exit <- client.Exit(false) }()
	<-endpoint.entered
	closed := make(chan error, 1)
	go func() { closed <- server.CloseGracefully() }()
	waitForNativeRPCClosing(t, server)
	otherClose := make(chan error, 1)
	go func() { otherClose <- server.CloseGracefully() }()
	select {
	case err := <-closed:
		t.Fatalf("close returned before accepted reply: %v", err)
	default:
	}
	close(endpoint.release)
	if err := <-exit; err != nil {
		t.Fatalf("Exit reply was cut off by command shutdown: %v", err)
	}
	if err := <-closed; err != nil {
		t.Fatal(err)
	}
	if err := <-otherClose; err != nil {
		t.Fatal(err)
	}
	if _, err := client.Size(); err == nil {
		t.Fatal("closed process accepted another call")
	}
}

func TestFingerprintRPCForcedCloseCanInterruptDrain(t *testing.T) {
	endpoint := &gatedFingerprintExit{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(NewMemFPSet()), entered: make(chan struct{}), release: make(chan struct{})}
	server, client := startFingerprintRPC(t, endpoint)
	exit := make(chan error, 1)
	go func() { exit <- client.Exit(false) }()
	<-endpoint.entered
	closed := make(chan error, 1)
	go func() { closed <- server.CloseGracefully() }()
	waitForNativeRPCClosing(t, server)
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-exit; err == nil {
		t.Fatal("forced transport closure acknowledged an unfinished call")
	}
	close(endpoint.release)
	if err := <-closed; err != nil {
		t.Fatal(err)
	}
}

func TestFingerprintRPCGracefulCloseIdleConnection(t *testing.T) {
	server, _ := startFingerprintRPC(t, NewLocalFingerprintEndpoint(NewMemFPSet()))
	if err := server.CloseGracefully(); err != nil {
		t.Fatal(err)
	}
}

func waitForNativeRPCClosing(t *testing.T, server *DistributedRPCServer) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		server.mu.Lock()
		closed := server.closed
		server.mu.Unlock()
		if closed {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("transport did not start closing")
		}
		time.Sleep(time.Millisecond)
	}
}
