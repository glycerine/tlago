package tlc

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// TLCServer.getFile creates InJarFilenameToStream for every request. Its
// defaults are live; only the worker resolver caches files. No direct original
// Java method covers changing defaults between coordinator file requests.
func TestDistributedServerFilesRefreshDefaultSearchPaths(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	const name = "RemoteLiveLookup.tla"
	for directory, contents := range map[string]string{first: "first", second: "second"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	// Keep this fixture's default properties independent of other command tests.
	tlcSystemProperties.Lock()
	oldLibrary, hadLibrary := tlcSystemProperties.values[TLALibraryProperty]
	delete(tlcSystemProperties.values, TLALibraryProperty)
	tlcSystemProperties.Unlock()
	t.Cleanup(func() {
		tlcSystemProperties.Lock()
		defer tlcSystemProperties.Unlock()
		if hadLibrary {
			tlcSystemProperties.values[TLALibraryProperty] = oldLibrary
		} else {
			delete(tlcSystemProperties.values, TLALibraryProperty)
		}
	})
	oldUser := GetFilenameUserDirectory()
	t.Cleanup(func() { SetFilenameUserDirectory(oldUser) })
	// All new resolver directories and resource copies are owned by this test.
	t.Setenv("TMPDIR", t.TempDir())
	t.Run("library", func(t *testing.T) {
		t.Setenv(TLALibraryProperty, first)
		files := NewDistributedServerFiles(t.TempDir(), nil, nil, "")
		server := &TLCServer{Files: files, InternTable: NewInternTable(16)}
		_, client := startCoordinatorRPC(t, NewLocalServerEndpoint(server))
		for _, row := range []struct{ directory, contents string }{{first, "first"}, {second, "second"}} {
			t.Setenv(TLALibraryProperty, row.directory)
			data, err := client.GetFile(name)
			if err != nil || string(data) != row.contents {
				t.Fatalf("live library file = %q/%v", data, err)
			}
		}
		// Explicit native library overrides remain captured and independent.
		paths := []string{first}
		explicit := NewDistributedServerFiles(t.TempDir(), paths, nil, "")
		paths[0] = second
		data, err := (&TLCServer{Files: explicit}).GetFile(name)
		if err != nil || string(data) != "first" {
			t.Fatalf("explicit library override changed: %q/%v", data, err)
		}
	})
	t.Run("user_directory", func(t *testing.T) {
		SetFilenameUserDirectory(&first)
		files := NewDistributedServerFiles("", []string{}, nil, "")
		server := &TLCServer{Files: files, InternTable: NewInternTable(16)}
		_, client := startCoordinatorRPC(t, NewLocalServerEndpoint(server))
		for _, row := range []struct{ directory, contents string }{{first, "first"}, {second, "second"}} {
			SetFilenameUserDirectory(&row.directory)
			data, err := client.GetFile(name)
			if err != nil || string(data) != row.contents {
				t.Fatalf("live user directory file = %q/%v", data, err)
			}
		}
	})
}

func TestDistributedServerFilesFreshResourceOwnership(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
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
	resources := fstest.MapFS{"model/Spec.tla": &fstest.MapFile{Data: []byte("first")}}
	files := NewDistributedServerFiles(t.TempDir(), []string{}, resources, "")
	first := files.resolve("Spec.tla")
	resources["model/Spec.tla"] = &fstest.MapFile{Data: []byte("second")}
	second := files.resolve("Spec.tla")
	if filepath.Dir(first) == filepath.Dir(second) {
		t.Fatal("coordinator requests reused resource temporary ownership")
	}
	for path, want := range map[string][]byte{first: []byte("first"), second: []byte("second")} {
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("resource copy %s = %q/%v", path, got, err)
		}
	}
	CleanupDistributedFiles()
	for _, path := range []string{filepath.Dir(first), filepath.Dir(second)} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("request resource directory was not released at process-exit boundary: %s/%v", path, err)
		}
	}
}
