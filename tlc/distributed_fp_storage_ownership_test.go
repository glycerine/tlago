package tlc

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

type ownedFingerprintStorage struct {
	FPSet
	closes atomic.Int32
	exits  atomic.Int32
}

func (s *ownedFingerprintStorage) Close() {
	s.closes.Add(1)
	s.FPSet.Close()
}

func (s *ownedFingerprintStorage) Exit(cleanup bool) error {
	s.exits.Add(1)
	return s.FPSet.Exit(cleanup)
}

type ownershipRegistrationCoordinator struct {
	*LocalServerEndpoint
	failure error
}

func (s *ownershipRegistrationCoordinator) RegisterFPSet(set DistributedFingerprintEndpoint, hostname string) error {
	if isDistributedFPRegistrationRejected(s.failure) {
		return s.failure
	}
	if err := s.LocalServerEndpoint.RegisterFPSet(set, hostname); err != nil {
		return err
	}
	// Exercise a failure reported after the actual registration has completed.
	return s.failure
}

type ownershipBlockedFingerprint struct {
	*LocalFingerprintEndpoint
	entered chan struct{}
	release chan struct{}
}

func (s *ownershipBlockedFingerprint) Contains(fp uint64) (bool, error) {
	close(s.entered)
	<-s.release
	return s.LocalFingerprintEndpoint.Contains(fp)
}

// Native host ownership has no direct original Java method. Publication and
// command catches retain source behavior; storage release follows RPC draining.
func TestNativeFingerprintStorageOwnership(t *testing.T) {
	for _, outcome := range []string{"accepted", "rejected", "registration_failure"} {
		t.Run(outcome, func(t *testing.T) {
			manager := NewDynamicDistributedFPSetManager(2)
			registration := &distributedFPRegistration{expected: 2, remaining: 2, done: make(chan struct{})}
			coordinator := &ownershipRegistrationCoordinator{LocalServerEndpoint: NewLocalServerEndpoint(&TLCServer{FPSetManager: manager, fpRegistration: registration})}
			if outcome == "rejected" {
				coordinator.failure = NewFPSetManagerException("full")
			} else if outcome == "registration_failure" {
				coordinator.failure = distributedTestRemoteFailure("registration reply unavailable")
			}
			_, server := startCoordinatorRPC(t, coordinator)
			network, err := NewDistributedFPServerNetwork("127.0.0.1:0", "")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = network.Close() })
			borrowed := &ownedFingerprintStorage{FPSet: NewMemFPSet()}
			if err := network.Host.RegisterFingerprint("borrowed", NewLocalFingerprintEndpoint(borrowed)); err != nil {
				t.Fatal(err)
			}
			directory := t.TempDir()
			config := NewFPSetConfiguration()
			config.GetMemoryInBytesOverride = func() int64 { return 8192 }
			storage := &ownedFingerprintStorage{FPSet: NewLSBDiskFPSet(config).Init(1, directory, "owned")}
			t.Cleanup(storage.FPSet.Close)
			storage.Put(17)
			env := network.FPEnvironment(DistributedFPServerEnvironment{})
			err = env.RegisterFPSet(server, NewLocalFingerprintEndpoint(storage), "owner")
			if (err == nil) != (outcome == "accepted") {
				t.Fatalf("registration result = %v", err)
			}
			if outcome == "rejected" {
				if !isDistributedFPRegistrationRejected(err) {
					t.Fatalf("rejection category = %v", err)
				}
				unpublishDistributedFPSet(storage, false, env)
			} else if manager.NumOfServers() != 1 {
				t.Fatal("registration did not complete before its response")
			}
			if storage.closes.Load() != 0 {
				t.Fatal("registration/unpublication prematurely closed storage")
			}
			network.fingerprintsMu.Lock()
			names := append([]string(nil), network.fingerprints[storage]...)
			network.fingerprintsMu.Unlock()
			if outcome == "rejected" {
				if len(names) != 0 {
					t.Fatal("source rejection retained publication")
				}
			} else {
				if len(names) != 1 {
					t.Fatal("publication missing after accepted/ambiguous registration")
				}
				client, err := DialFingerprintEndpoint(network.Address, names[0])
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = client.CloseConnection() })
				if found, err := client.Contains(17); err != nil || !found {
					t.Fatalf("published storage unavailable: %v/%v", found, err)
				}
				if outcome == "accepted" {
					// Repeated publication of one storage object owns it once.
					if err := env.RegisterFPSet(server, NewLocalFingerprintEndpoint(storage), "owner"); err != nil {
						t.Fatal(err)
					}
					blocked := &ownershipBlockedFingerprint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(storage), entered: make(chan struct{}), release: make(chan struct{})}
					network.Host.mu.Lock()
					network.Host.fingerprints[names[0]] = blocked
					network.Host.mu.Unlock()
					var release sync.Once
					t.Cleanup(func() { release.Do(func() { close(blocked.release) }) })
					called := make(chan error, 1)
					callDone := make(chan struct{})
					go func() {
						defer close(callDone)
						found, err := client.Contains(17)
						if err == nil && !found {
							err = errors.New("accepted lookup lost stored fingerprint")
						}
						called <- err
					}()
					<-blocked.entered
					closed := make(chan error, 1)
					closeDone := make(chan struct{})
					go func() { defer close(closeDone); closed <- network.Close() }()
					t.Cleanup(func() {
						release.Do(func() { close(blocked.release) })
						<-callDone
						<-closeDone
					})
					waitForNativeRPCClosing(t, network.Host)
					if storage.closes.Load() != 0 {
						t.Fatal("host closed storage while accepted RPC was blocked")
					}
					release.Do(func() { close(blocked.release) })
					if err := <-called; err != nil {
						t.Fatal("accepted reply did not drain", err)
					}
					if err := <-closed; err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := network.Close(); err != nil {
				t.Fatal(err)
			}
			if storage.closes.Load() != 1 || storage.exits.Load() != 0 {
				t.Fatalf("native storage release = %d closes/%d exits", storage.closes.Load(), storage.exits.Load())
			}
			if borrowed.closes.Load() != 0 || borrowed.exits.Load() != 0 {
				t.Fatal("host took storage ownership from a direct publication caller")
			}
			requireNoDistributedConstructorDescriptors(t, directory)
			if _, err := os.Stat(filepath.Join(directory, "owned.fp")); err != nil {
				t.Fatal("native close deleted storage files", err)
			}
			if err := env.RegisterFPSet(server, NewLocalFingerprintEndpoint(storage), "late"); !errors.Is(err, net.ErrClosed) {
				t.Fatalf("closed host accepted new storage: %v", err)
			}
			if storage.closes.Load() != 1 {
				t.Fatal("late registration changed ownership after close")
			}
		})
	}
}
