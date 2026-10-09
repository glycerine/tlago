package tlc

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// No original factory test covers the coordinator visible to getSelector.
// Use real constructors/storage and keep process-lifetime settings isolated.
func TestDistributedSelectorFactoryCoordinatorStartup(t *testing.T) {
	if scenario := os.Getenv("TLAGO_SELECTOR_COORDINATOR_STARTUP"); scenario != "" {
		checkSelectorFactoryCoordinatorStartup(t, scenario)
		return
	}
	for _, scenario := range []string{"local", "local-panic", "distributed", "distributed-panic"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDistributedSelectorFactoryCoordinatorStartup$", "-test.v")
			command.Env = append(os.Environ(), "TLAGO_SELECTOR_COORDINATOR_STARTUP="+scenario)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("coordinator startup process failed: %v\n%s", err, output)
			}
		})
	}
}

func checkSelectorFactoryCoordinatorStartup(t *testing.T, scenario string) {
	setDistributedConstructorManagerCount(t)
	distributed := strings.HasPrefix(scenario, "distributed")
	panicFactory := strings.HasSuffix(scenario, "panic")
	directory := t.TempDir()
	app := NewTLCApp(&Tool{RootFile: "Spec"}, false)
	app.metadataSet, app.metadir, app.config = true, directory, "ConfigAlias"
	app.fpSetConfig = NewFPSetConfigurationWithRatioAndImplementation(0.25, "tlc2.tool.fp.LSBDiskFPSet")
	app.fpSetConfig.GetMemoryInBytesOverride = func() int64 { return 1 << 20 }
	tlcSetStartupSystemProperty(distributedSelectorFactoryProperty, "startup-owner")
	failure := errors.New("selector startup failure")
	var observed *TLCServer
	var sawTool *Tool
	var sawApp *TLCApp
	var sawDeadlock, sawCheckpoint, sawRegistration bool
	var storage FPSet
	policy := &nativeTestBlockSelection{maximum: 3}
	RegisterBlockSelectorFactory("startup-owner", func() (BlockSelectorFactory, error) {
		return func(server *TLCServer) BlockSelection {
			observed, sawTool, sawApp = server, server.Tool, server.app
			sawDeadlock = server.GetCheckDeadlock()
			sawCheckpoint = server.checkpointName != nil && *server.checkpointName == filepath.Base(directory)
			sawRegistration = server.fpRegistration != nil
			if server.StateQueue == nil || server.Trace == nil || server.Trace.Tool != app.Tool || server.FPSetManager == nil || server.GetSpecFileName() != "Spec" || server.GetConfigFileName() != "ConfigAlias" {
				t.Fatal("factory ran before queue, trace, manager or names were initialized")
			}
			if distributed {
				if server.FPSetManager.NonDistributed || server.FPSetManager.ExpectedNumServers != 1 {
					t.Fatal("factory did not receive the base constructor's dynamic manager")
				}
			} else {
				if !server.FPSetManager.NonDistributed {
					t.Fatal("factory did not receive local fingerprint storage")
				}
				storage = server.FPSetManager.entry(0).set.(*LocalFingerprintEndpoint).Set
			}
			if panicFactory {
				panic(failure)
			}
			return policy
		}, nil
	})
	before := javaPoolWriterGoroutines()
	var server *TLCServer
	var constructorErr error
	var panicValue any
	func() {
		defer func() { panicValue = recover() }()
		if distributed {
			server, constructorErr = NewDistributedFPSetTLCServer(app, 2)
		} else {
			server, constructorErr = NewTLCServerFromApp(app)
		}
	}()
	var cleanup sync.Once
	release := func() {
		cleanup.Do(func() {
			if server == nil {
				return
			}
			server.StateQueue.(*DiskStateQueue).finishConstructionAndWait()
			_ = server.Trace.Close()
			if storage != nil {
				storage.Close()
			}
		})
	}
	t.Cleanup(release)
	if constructorErr != nil || observed == nil || sawTool != app.Tool || sawApp != app || sawDeadlock || !sawCheckpoint || sawRegistration {
		t.Fatalf("factory saw incomplete application: error=%v, tool=%p/%p, app=%p/%p, deadlock=%v, checkpoint=%v, registration=%v", constructorErr, sawTool, app.Tool, sawApp, app, sawDeadlock, sawCheckpoint, sawRegistration)
	}
	if panicFactory {
		if server != nil || panicValue != failure {
			t.Fatalf("factory panic replaced or swallowed: %p/%v", server, panicValue)
		}
	} else {
		if panicValue != nil || server != observed || server.BlockSelector != policy || server.app != app {
			t.Fatal("successful constructor replaced the observed coordinator or policy")
		}
		if distributed && (server.fpRegistration == nil || server.fpRegistration.expected != 2) {
			t.Fatal("subclass registration state was not assigned after base factory creation")
		}
		release()
	}
	awaitDistributedConstructorWorkers(t, before)
	requireNoDistributedConstructorDescriptors(t, directory)
	if _, err := os.Stat(filepath.Join(directory, "Spec.st")); err != nil {
		t.Fatalf("native cleanup deleted the initialized trace: %v", err)
	}
	if !distributed {
		if _, err := os.Stat(filepath.Join(directory, "Spec_0.fp")); err != nil {
			t.Fatalf("native cleanup deleted initialized storage: %v", err)
		}
	}
}
