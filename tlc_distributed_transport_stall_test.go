package tlago

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glycerine/tlago/tlc"
)

// A test-owned byte relay holds selected traffic directions without closing
// either TCP connection or decoding/replacing TLC requests and answers.
type nativeFingerprintTrafficGate struct {
	path      string
	direction string
	once      [2]sync.Once
}

type nativeGatedTCPWriter struct {
	net.Conn
	gate      *nativeFingerprintTrafficGate
	direction string
}

func (w nativeGatedTCPWriter) Write(data []byte) (int, error) {
	if w.gate.direction != "" && w.gate.direction != w.direction {
		return w.Conn.Write(data)
	}
	if _, err := os.Stat(w.gate.path + ".block"); err == nil {
		if _, err := os.Stat(w.gate.path); errors.Is(err, os.ErrNotExist) {
			index := 0
			if w.direction == "reply" {
				index = 1
			}
			w.gate.once[index].Do(func() { fmt.Println("NATIVE_FP_TCP_TRAFFIC_BLOCKED=" + w.direction) })
			if err := nativeDistributedWaitForFile(w.gate.path); err != nil {
				return 0, err
			}
		} else if err != nil {
			return 0, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return 0, err
	}
	return w.Conn.Write(data)
}

func nativeFingerprintTransportStallHost(args []string, releasePath string) error {
	if releasePath == "" {
		return fmt.Errorf("TCP gate path missing")
	}
	direction := os.Getenv("TLAGO_NATIVE_FP_STALL_DIRECTION")
	if direction != "" && direction != "request" && direction != "reply" {
		return fmt.Errorf("unknown TCP gate direction %q", direction)
	}
	args, err := tlc.ExtractDistributedStartupProperties(args)
	if err != nil {
		return err
	}
	backend, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	host := tlc.NewDistributedRPCServer()
	served := make(chan error, 1)
	go func() { served <- host.Serve(backend) }()
	joined := false
	defer func() {
		if !joined {
			_ = host.CloseGracefully()
			<-served
		}
	}()
	relay, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	gate := &nativeFingerprintTrafficGate{path: releasePath, direction: direction}
	var copies sync.WaitGroup
	accepted := make(chan struct{})
	go func() {
		defer close(accepted)
		for {
			client, err := relay.Accept()
			if err != nil {
				return
			}
			server, err := net.Dial("tcp", backend.Addr().String())
			if err != nil {
				_ = client.Close()
				continue
			}
			copies.Add(1)
			go func() {
				defer copies.Done()
				defer func() {
					_ = client.Close()
					_ = server.Close()
				}()
				returned := make(chan struct{})
				go func() {
					_, _ = io.Copy(nativeGatedTCPWriter{Conn: server, gate: gate, direction: "request"}, client)
					close(returned)
				}()
				_, _ = io.Copy(nativeGatedTCPWriter{Conn: client, gate: gate, direction: "reply"}, server)
				_ = client.Close()
				_ = server.Close()
				<-returned
			}()
		}
	}()
	defer func() {
		_ = relay.Close()
		<-accepted
		// Drain accepted backend replies before its connections close. The
		// relay forwards those bytes through EOF before closing client sockets.
		_ = host.CloseGracefully()
		<-served
		joined = true
		copies.Wait()
	}()
	discovery := tlc.NewDistributedNetworkDiscovery()
	env := tlc.DistributedFPServerEnvironment{Lookup: discovery.Lookup, ToolOut: os.Stdout, SystemOut: os.Stdout, SystemErr: os.Stderr,
		LocalHostName: func() (string, error) { return "127.0.0.1", nil },
		RegisterFPSet: func(server tlc.DistributedServerEndpoint, endpoint tlc.DistributedFingerprintEndpoint, hostname string) error {
			coordinator, ok := server.(*tlc.NetworkServerEndpoint)
			if !ok {
				return fmt.Errorf("TCP relay requires native coordinator")
			}
			if err := host.RegisterFingerprint("relayed", endpoint); err != nil {
				return err
			}
			return coordinator.RegisterFPSetReference(tlc.DistributedEndpointReference{Address: relay.Addr().String(), Object: "relayed"}, hostname)
		}, UnpublishFPSet: func(tlc.FPSet, bool) { host.UnregisterFingerprint("relayed") },
	}
	tlc.RunDistributedFPServer(args, env)
	return nil
}

func checkNativeFingerprintTransportStall(t *testing.T, ctx context.Context, coordinator, fingerprint *nativeDistributedTestProcess, port int, releasePath, direction string) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	marker := "NATIVE_FP_TCP_TRAFFIC_BLOCKED=" + direction
	for !strings.Contains(fingerprint.output.String(), marker) || len(nativeDistributedMessages(coordinator.output.String(), tlc.ECTLCDistributedWorkerRegistered)) != 1 {
		select {
		case err := <-fingerprint.done:
			fingerprint.joined = true
			t.Fatalf("relay exited before traffic barrier: %v", err)
		case err := <-coordinator.done:
			coordinator.joined = true
			t.Fatalf("coordinator exited before traffic barrier: %v", err)
		case <-ctx.Done():
			t.Fatal("TCP stall watchdog expired")
		case <-ticker.C:
		}
	}
	if direction != "" {
		opposite := "reply"
		if direction == "reply" {
			opposite = "request"
		}
		if strings.Contains(fingerprint.output.String(), "NATIVE_FP_TCP_TRAFFIC_BLOCKED="+opposite) {
			t.Fatal("relay blocked the unselected traffic direction")
		}
	}
	// These deadlines bound test control probes only; production connections
	// keep their source timeout policy and remain stalled until relay release.
	probe := func(address string) *rpc.Client {
		connection, err := net.DialTimeout("tcp", address, 5*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if err := connection.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
			_ = connection.Close()
			t.Fatal(err)
		}
		client := rpc.NewClient(connection)
		t.Cleanup(func() { _ = client.Close() })
		return client
	}
	server := probe(net.JoinHostPort("127.0.0.1", fmt.Sprint(port)))
	var status tlc.DistributedServerReply
	if err := server.Call("Coordinator.Call", tlc.DistributedServerRequest{Object: tlc.TLCServerName, Operation: "done"}, &status); err != nil || status.Failure != nil || status.Bool {
		t.Fatalf("coordinator not available with unfinished work: %v/%v/%v", err, status.Failure, status.Bool)
	}
	var snapshot tlc.DistributedServerReply
	if err := server.Call("Coordinator.Call", tlc.DistributedServerRequest{Object: tlc.TLCServerName, Operation: "manager"}, &snapshot); err != nil || snapshot.Failure != nil || snapshot.Manager == nil || len(snapshot.Manager.Partitions) != 2 || snapshot.Manager.Partitions[0] == snapshot.Manager.Partitions[1] {
		t.Fatalf("TCP stall unexpectedly changed fingerprint routing: %v/%v", err, snapshot.Failure)
	}
	registered := nativeDistributedMessages(coordinator.output.String(), tlc.ECTLCDistributedWorkerRegistered)
	uri, err := url.Parse(regexp.MustCompile(`tcp://[^\s]+`).FindString(registered[0]))
	if err != nil {
		t.Fatal(err)
	}
	worker := probe(uri.Host)
	for _, operation := range []string{"alive", "cache"} {
		var reply tlc.DistributedWorkerReply
		if err := worker.Call("Worker.Call", tlc.DistributedWorkerRequest{Object: strings.TrimPrefix(uri.Path, "/"), Operation: operation}, &reply); err != nil || reply.Failure != nil || operation == "alive" && !reply.Alive {
			t.Fatalf("worker control %s blocked during fingerprint TCP stall: %v/%v", operation, err, reply.Failure)
		}
	}
	if len(nativeDistributedMessages(coordinator.output.String(), tlc.ECTLCFinished)) != 0 {
		t.Fatal("stalled model falsely completed")
	}
	if err := os.WriteFile(releasePath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Logf("fingerprint TCP traffic held (%q, empty means both); coordinator and worker control calls respond; relay released", direction)
}
