package tlc

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

type resolverFileCoordinator struct {
	*LocalServerEndpoint
	mu    sync.Mutex
	calls []string
}

func (s *resolverFileCoordinator) GetFile(name string) ([]byte, error) {
	s.mu.Lock()
	s.calls = append(s.calls, name)
	s.mu.Unlock()
	return s.LocalServerEndpoint.GetFile(name)
}

// No enabled upstream method directly tests its distributed file resolver.
// Use real coordinator file loading over TCP and require the source file-cache
// behavior: basename keys, no extension inference, and refetch after deletion.
func TestDistributedFilenameResolverCoordinatorFilesAndCache(t *testing.T) {
	directory := t.TempDir()
	write := func(name string, data []byte) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(directory, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("Spec.tla", []byte("first specification"))
	write("binary.cfg", []byte{0, 255, 13, 10})
	write("empty.cfg", []byte{})
	server := &TLCServer{Files: NewDistributedServerFiles(directory, []string{}, nil, ""), InternTable: NewInternTable(16)}
	coordinator := &resolverFileCoordinator{LocalServerEndpoint: NewLocalServerEndpoint(server)}
	_, client := startCoordinatorRPC(t, coordinator)
	// Isolate process-exit registrations; test cleanup only owns these paths.
	distributedDeleteOnExit.Lock()
	previousPaths := distributedDeleteOnExit.paths
	distributedDeleteOnExit.paths = nil
	distributedDeleteOnExit.Unlock()
	t.Cleanup(func() {
		CleanupDistributedFiles()
		distributedDeleteOnExit.Lock()
		distributedDeleteOnExit.paths = previousPaths
		distributedDeleteOnExit.Unlock()
	})
	resolver := NewDistributedFilenameToStreamResolver([]string{"unused-library"})
	resolver.SetTLCServer(client)
	first := resolver.Resolve(filepath.Join("nested", "Spec.tla"), true)
	read := func(path string, want []byte) {
		t.Helper()
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("file %s = %q/%v, want %q", path, got, err, want)
		}
	}
	if first.GetPath() != filepath.Join(resolver.tmpDir, "Spec.tla") {
		t.Fatal("resolver retained coordinator-side path")
	}
	read(first.GetPath(), []byte("first specification"))
	write("Spec.tla", []byte("changed specification"))
	if cached := resolver.Resolve(filepath.Join("other", "Spec.tla"), false); cached != first {
		t.Fatal("basename/isModule cache key changed")
	}
	read(first.GetPath(), []byte("first specification"))
	read(resolver.Resolve("binary.cfg", true).GetPath(), []byte{0, 255, 13, 10})
	read(resolver.Resolve("empty.cfg", false).GetPath(), []byte{})
	if err := os.Remove(first.GetPath()); err != nil {
		t.Fatal(err)
	}
	refetched := resolver.Resolve("Spec.tla", true)
	if refetched == first || refetched.GetPath() != first.GetPath() {
		t.Fatal("missing cached file did not produce a fresh file object at its cache path")
	}
	read(refetched.GetPath(), []byte("changed specification"))
	independent := NewDistributedFilenameToStreamResolver()
	independent.SetTLCServer(client)
	second := independent.Resolve("Spec.tla", true)
	if second.GetPath() == first.GetPath() || independent.tmpDir == resolver.tmpDir {
		t.Fatal("resolvers share temporary file ownership")
	}
	read(second.GetPath(), []byte("changed specification"))
	if resolver.IsStandardModule("Naturals") {
		t.Fatal("distributed resolver inferred a standard module")
	}
	coordinator.mu.Lock()
	calls := append([]string(nil), coordinator.calls...)
	coordinator.mu.Unlock()
	want := []string{"Spec.tla", "binary.cfg", "empty.cfg", "Spec.tla", "Spec.tla"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("coordinator fetches = %q, want %q", calls, want)
	}
}
