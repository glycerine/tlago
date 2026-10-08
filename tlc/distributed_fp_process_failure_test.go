package tlc

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

// Native process coverage supplements the original manager tests. Unlike
// closing a hosted connection, killing this child also destroys its live store.
func TestFingerprintRPCProcessCrashRecovery(t *testing.T) {
	if directory := os.Getenv("TLAGO_FP_CRASH_DIRECTORY"); directory != "" {
		storage := NewMemFPSet()
		storage.Init(1, directory, "primary")
		host := NewDistributedRPCServer()
		if err := host.RegisterFingerprint("primary", NewLocalFingerprintEndpoint(storage)); err != nil {
			t.Fatal(err)
		}
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("READY " + listener.Addr().String())
		if err := host.Serve(listener); err != nil {
			t.Fatal(err)
		}
		return
	}
	captureFailoverToolIO(t, ToolIOTool)
	directory := t.TempDir()
	start := func() (*NetworkFingerprintEndpoint, func()) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestFingerprintRPCProcessCrashRecovery$")
		command.Env = append(os.Environ(), "TLAGO_FP_CRASH_DIRECTORY="+directory)
		var stderr bytes.Buffer
		command.Stderr = &stderr
		stdout, err := command.StdoutPipe()
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		if err := command.Start(); err != nil {
			cancel()
			t.Fatal(err)
		}
		var stopOnce sync.Once
		stop := func() {
			t.Helper()
			stopOnce.Do(func() {
				killErr := command.Process.Kill()
				waitErr := command.Wait()
				cancel()
				var exit *exec.ExitError
				if killErr != nil || !errors.As(waitErr, &exit) || exit.Success() {
					t.Errorf("fingerprint child did not terminate after kill: %v/%v\n%s", killErr, waitErr, stderr.String())
				}
			})
		}
		t.Cleanup(stop)
		ready := make(chan string, 1)
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				if strings.HasPrefix(scanner.Text(), "READY ") {
					ready <- strings.TrimPrefix(scanner.Text(), "READY ")
					return
				}
			}
			ready <- ""
		}()
		var address string
		select {
		case address = <-ready:
		case <-time.After(5 * time.Second):
			t.Fatal("fingerprint child readiness watchdog expired")
		}
		if address == "" {
			t.Fatal("fingerprint child exited before readiness")
		}
		client, err := DialFingerprintEndpoint(address, "primary")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = client.CloseConnection() })
		return client, stop
	}

	t.Log("starting fingerprint process and committing checkpoint")
	primary, kill := start()
	if bits, err := primary.PutBlock(NewLongVecFrom([]int64{2, 4})); err != nil || bits == nil || bits.TrueCount() != 2 {
		t.Fatalf("initial insertion: %v/%v", bits, err)
	}
	if err := primary.BeginChkptFile("job"); err != nil {
		t.Fatal(err)
	}
	if err := primary.CommitChkptFile("job"); err != nil {
		t.Fatal(err)
	}
	if known, err := primary.Put(6); err != nil || known {
		t.Fatalf("post-checkpoint insertion: %v/%v", known, err)
	}
	survivorStore := NewMemFPSet()
	_, survivor := startFingerprintRPC(t, NewLocalFingerprintEndpoint(survivorStore))
	manager := NewDistributedFPSetManager(primary, survivor)
	manager.fpSets[0].hostname = "killed-primary"
	t.Log("killing fingerprint process and checking manager failover")
	kill()
	partitions := []*LongVec{NewLongVecFrom([]int64{2, 4, 6}), NewLongVec()}
	answers := manager.ContainsBlock(partitions)
	requireJavaBitCounts(t, answers, 3, 0)
	if manager.NumOfAliveServers() != 1 || manager.entry(0) != manager.entry(1) || manager.entry(0).set != survivor || survivorStore.Size() != 0 {
		t.Fatal("process loss did not reassign without copying lost fingerprints")
	}
	requireJavaBitCounts(t, manager.PutBlock(partitions), 3, 0)
	requireJavaBitCounts(t, manager.PutBlock(partitions), 0, 0)
	if survivorStore.Size() != 3 {
		t.Fatal("surviving store did not retain the reassigned block")
	}
	messages := ToolIOGetAllMessages()
	if len(messages) != 1 || !strings.Contains(messages[0], "to the fp server at killed-primary.\n") {
		t.Fatalf("process loss warning: %q", messages)
	}

	t.Log("starting fresh fingerprint process and recovering committed checkpoint")
	recovered, _ := start()
	if size, err := recovered.Size(); err != nil || size != 0 {
		t.Fatalf("fresh process inherited volatile fingerprints: %d/%v", size, err)
	}
	if err := recovered.RecoverFile("job"); err != nil {
		t.Fatal(err)
	}
	if bits, err := recovered.ContainsBlock(partitions[0]); err != nil || bits == nil || bits.Get(0) || bits.Get(1) || !bits.Get(2) || bits.Get(3) {
		t.Fatalf("recovery did not preserve committed membership only: %v/%v", bits, err)
	}
	if size, err := recovered.Size(); err != nil || size != 2 {
		t.Fatalf("recovered membership count: %d/%v", size, err)
	}
	if _, err := primary.Put(8); !errors.Is(err, rpc.ErrShutdown) {
		t.Fatalf("dead client redialed or lost its native shutdown cause: %v", err)
	}
	if manager.entry(0).set != survivor {
		t.Fatal("recovery silently replaced manager registrations")
	}
}
