// Java source: util/FileUtilTest.java.
// Copyright (c) 2023, Oracle and/or its affiliates.
package tlc

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func javaFileUtilSetup(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	old := Globals.MetaDir
	t.Cleanup(func() { Globals.MetaDir = old })
	return directory
}

func TestJavaReturnFromCheckpoint(t *testing.T) {
	directory := javaFileUtilSetup(t)
	checkpoint := "abc"
	if path := MakeMetaDir(time.Now(), directory, &checkpoint); path != checkpoint {
		t.Fatalf("makeMetaDir = %q, want %q", path, checkpoint)
	}
}

func TestJavaDuplicateStateDirCreation(t *testing.T) {
	directory := javaFileUtilSetup(t)
	now := time.Now()
	path1 := MakeMetaDir(now, directory, nil)
	path2 := MakeMetaDir(now, directory, nil)
	if path1 == path2 {
		t.Fatalf("duplicate metadata directories: %q", path1)
	}
	parent := filepath.Join(directory, MetaRoot)
	for _, path := range []string{path1, path2} {
		if filepath.Dir(path) != parent {
			t.Fatalf("parent of %q = %q, want %q", path, filepath.Dir(path), parent)
		}
	}
}

func TestJavaUseDifferentMetaDir(t *testing.T) {
	directory := javaFileUtilSetup(t)
	Globals.MetaDir = filepath.Join(directory, "fizz", "buzz")
	path := MakeMetaDir(time.Now(), directory, nil)
	if filepath.Dir(path) != Globals.MetaDir {
		t.Fatalf("parent of %q = %q, want %q", path, filepath.Dir(path), Globals.MetaDir)
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		t.Fatalf("metadata path is not a directory: %q, %v", path, err)
	}
}
