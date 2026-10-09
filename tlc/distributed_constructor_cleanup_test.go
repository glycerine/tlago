package tlc

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// No original method tests failed-constructor ownership. These native checks
// preserve construction/error ordering while requiring owned workers to stop.
func TestDistributedConstructorRollback(t *testing.T) {
	setDistributedConstructorManagerCount(t)
	for _, phase := range []string{"trace_open", "null_configuration", "fingerprint_init", "invalid_manager_count", "negative_registration_count"} {
		t.Run(phase, func(t *testing.T) {
			directory := t.TempDir()
			app := NewTLCApp(&Tool{RootFile: "Spec"}, true)
			app.metadataSet, app.metadir = true, directory
			if phase == "trace_open" {
				app.metadir = filepath.Join(directory, "missing")
			}
			if phase == "invalid_manager_count" {
				tlcServerProperties.expectedFPSetCount = 0
				t.Cleanup(func() { tlcServerProperties.expectedFPSetCount = 1 })
			}
			if phase == "fingerprint_init" {
				app.fpSetConfig = NewFPSetConfigurationWithRatioAndImplementation(0.25, "tlc2.tool.fp.LSBDiskFPSet")
				app.fpSetConfig.GetMemoryInBytesOverride = func() int64 { return 1 << 20 }
				// One nested child opens its actual readers; the other fails to
				// initialize against a directory. Initialization joins both.
				if err := os.Mkdir(filepath.Join(directory, "Spec_1.fp"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			before := javaPoolWriterGoroutines()
			for i := 0; i < 3; i++ {
				var failure any
				var err error
				func() {
					defer func() { failure = recover() }()
					var server *TLCServer
					if phase == "negative_registration_count" || phase == "invalid_manager_count" {
						server, err = NewDistributedFPSetTLCServer(app, -1)
					} else {
						server, err = NewTLCServerFromApp(app)
					}
					if server != nil {
						t.Fatal("failed constructor returned a server")
					}
				}()
				switch phase {
				case "trace_open":
					if err == nil || failure != nil {
						t.Fatalf("trace-open failure = %v/%v", err, failure)
					}
				case "null_configuration":
					if _, ok := failure.(*NullPointerException); !ok || err != nil {
						t.Fatalf("null configuration failure = %T/%v/%v", failure, failure, err)
					}
				case "negative_registration_count", "invalid_manager_count":
					if _, ok := failure.(*IllegalArgumentException); !ok || err != nil {
						t.Fatalf("negative count failure = %T/%v/%v", failure, failure, err)
					}
				case "fingerprint_init":
					if _, ok := failure.(*RuntimeException); !ok || err != nil {
						t.Fatalf("nested fingerprint initialization failure = %T/%v/%v", failure, failure, err)
					}
				}
			}
			awaitDistributedConstructorWorkers(t, before)
			if phase == "trace_open" {
				if _, err := os.Stat(app.metadir); !os.IsNotExist(err) {
					t.Fatalf("failed trace opening created metadata: %v", err)
				}
			} else if _, err := os.Stat(filepath.Join(directory, "Spec.st")); err != nil {
				t.Fatalf("rollback deleted the already-created trace: %v", err)
			}
			requireNoDistributedConstructorDescriptors(t, directory)
			if phase == "fingerprint_init" {
				if info, err := os.Stat(filepath.Join(directory, "Spec_0.fp")); err != nil || !info.Mode().IsRegular() {
					t.Fatalf("successful sibling initialization was skipped or removed: %v/%v", info, err)
				}
			}
		})
	}

	// Negative subclass validation must not overtake failure in base construction.
	app := NewTLCApp(&Tool{RootFile: "Spec"}, true)
	app.metadataSet, app.metadir = true, filepath.Join(t.TempDir(), "missing")
	if server, err := NewDistributedFPSetTLCServer(app, -1); server != nil || err == nil {
		t.Fatalf("base failure precedence = %v/%v", server, err)
	}
}

func awaitDistributedConstructorWorkers(t *testing.T, before map[string]bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		var live []string
		for _, stack := range javaPoolWriterStacks() {
			fields := strings.Fields(strings.SplitN(stack, "\n", 2)[0])
			if len(fields) < 2 || before[fields[1]] {
				continue
			}
			for _, kind := range []string{"StatePoolReader", "StatePoolWriter", "StatePoolCleaner"} {
				if strings.Contains(stack, "(*"+kind+").Start") {
					live = append(live, stack)
				}
			}
		}
		if len(live) == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("failed constructor retained queue workers:\n%s", strings.Join(live, "\n"))
		}
		time.Sleep(time.Millisecond)
	}
}

func requireNoDistributedConstructorDescriptors(t *testing.T, directory string) {
	t.Helper()
	if runtime.GOOS != "linux" {
		return // Descriptor enumeration is supplementary to portable worker checks.
	}
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		path, err := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
		if err == nil && strings.HasPrefix(path, directory+string(os.PathSeparator)) {
			t.Fatalf("constructor retained an open metadata descriptor: %s", path)
		}
	}
}

func TestDistributedConstructorTransfersResources(t *testing.T) {
	setDistributedConstructorManagerCount(t)
	app := NewTLCApp(&Tool{RootFile: "Spec"}, true)
	app.metadataSet, app.metadir = true, t.TempDir()
	before := javaPoolWriterGoroutines()
	server, err := NewDistributedFPSetTLCServer(app, 0)
	if err != nil {
		t.Fatal(err)
	}
	queue := server.StateQueue.(*DiskStateQueue)
	t.Cleanup(func() {
		queue.finishConstructionAndWait()
		_ = server.Trace.Close()
	})
	for _, done := range []chan struct{}{queue.reader.done, queue.writer.done, queue.cleaner.done} {
		select {
		case <-done:
			t.Fatal("successful constructor stopped an owned queue worker")
		default:
		}
	}
	if server.Trace.closed || server.Trace.raf.closed {
		t.Fatal("successful constructor closed its trace")
	}
	queue.finishConstructionAndWait()
	if err := server.Trace.Close(); err != nil {
		t.Fatal(err)
	}
	awaitDistributedConstructorWorkers(t, before)
	requireNoDistributedConstructorDescriptors(t, app.metadir)
}

func setDistributedConstructorManagerCount(t *testing.T) {
	t.Helper()
	InitializeTLCServerProperties()
	previous := tlcServerProperties.expectedFPSetCount
	tlcServerProperties.expectedFPSetCount = 1
	t.Cleanup(func() { tlcServerProperties.expectedFPSetCount = previous })
}
