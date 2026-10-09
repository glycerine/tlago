package tlc

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type nativeTestBlockSelection struct {
	maximum int
	queue   StateQueue
	worker  *DistributedWorkerSmartProxy
}

func (p *nativeTestBlockSelection) GetBlocks(queue StateQueue, worker *DistributedWorkerSmartProxy) []*TLCStateMut {
	p.queue, p.worker = queue, worker
	return queue.SDequeueMany(p.maximum)
}
func (p *nativeTestBlockSelection) SetMaxTXSize(maximum int)  { p.maximum = maximum }
func (p *nativeTestBlockSelection) GetAverageBlockCnt() int64 { return 41 }

// No upstream factory tests exist. Native constructors replace reflection;
// subprocesses preserve actual process-lifetime property capture.
func TestDistributedNativeSelectorFactory(t *testing.T) {
	if scenario := os.Getenv("TLAGO_NATIVE_SELECTOR_FACTORY"); scenario != "" {
		checkDistributedNativeSelectorFactory(t, scenario)
		return
	}
	for _, scenario := range []string{"custom", "unknown", "empty-name", "constructor-error", "nil-factory", "nil-selection", "constructor-panic", "selection-panic", "base-factory"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDistributedNativeSelectorFactory$", "-test.v")
			command.Env = append(os.Environ(), "TLAGO_NATIVE_SELECTOR_FACTORY="+scenario)
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("native selector process failed: %v\n%s", err, output)
			}
			wantDiagnostics := 0
			if scenario == "unknown" || scenario == "empty-name" || scenario == "constructor-error" {
				wantDiagnostics = 2
			}
			if strings.Count(string(output), "not registered")+strings.Count(string(output), "Cannot construct") != wantDiagnostics {
				t.Fatalf("factory fallback diagnostics changed: %s", output)
			}
		})
	}
}

func checkDistributedNativeSelectorFactory(t *testing.T, scenario string) {
	name := "native-selector"
	if scenario == "empty-name" {
		name = ""
	} else if scenario == "base-factory" {
		name = "tlc2.tool.distributed.selector.BlockSelectorFactory"
	}
	tlcSetStartupSystemProperty(distributedSelectorFactoryProperty, name)
	tlcSetStartupSystemProperty(distributedSelectorStaticProperty, "true")
	tlcSetStartupSystemProperty(distributedStaticBlockSizeProperty, "7")
	server := &TLCServer{}
	failure := errors.New("native constructor/selection failure")
	constructed, selected := 0, 0
	var policies []*nativeTestBlockSelection
	if scenario != "unknown" && scenario != "empty-name" && scenario != "base-factory" {
		RegisterBlockSelectorFactory(name, func() (BlockSelectorFactory, error) {
			constructed++
			// Registration from the callback proves no registry lock spans it.
			RegisterBlockSelectorFactory("unused", nil)
			switch scenario {
			case "constructor-error":
				return nil, failure
			case "constructor-panic":
				panic(failure)
			case "nil-factory":
				return nil, nil
			}
			return func(gotServer *TLCServer) BlockSelection {
				selected++
				if gotServer != server {
					t.Fatal("factory did not receive the original coordinator")
				}
				if scenario == "selection-panic" {
					panic(failure)
				}
				if scenario == "nil-selection" {
					return nil
				}
				policy := &nativeTestBlockSelection{maximum: 3}
				policies = append(policies, policy)
				return policy
			}, nil
		})
	}
	for request := 1; request <= 2; request++ {
		// Changing the property after first initialization must not retarget it.
		if request == 2 {
			tlcSetStartupSystemProperty(distributedSelectorFactoryProperty, "different-native-selector")
			tlcSetStartupSystemProperty(distributedSelectorStaticProperty, "false")
		}
		var selection BlockSelection
		err := invokeDistributedServerOperation(func() error {
			selection = NewBlockSelectorFromProperties(server)
			return nil
		})
		if scenario == "constructor-panic" || scenario == "selection-panic" {
			wantSelected := 0
			if scenario == "selection-panic" {
				wantSelected = request
			}
			if err != failure || constructed != request || selected != wantSelected {
				t.Fatalf("factory panic swallowed or copied: %v/%d/%d", err, constructed, selected)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if scenario == "nil-selection" {
			if selection != nil || constructed != request || selected != request {
				t.Fatal("nil custom selection substituted a built-in")
			}
			continue
		}
		if scenario != "custom" {
			builtin, ok := selection.(*BlockSelector)
			if !ok || builtin.Mode != BlockSelectorStatic || builtin.StaticBlockSize != 7 {
				t.Fatalf("failed factory changed captured built-in fallback: %#v", selection)
			}
			if scenario == "constructor-error" || scenario == "nil-factory" {
				if constructed != request || selected != 0 {
					t.Fatal("constructor failure selected a custom policy or cached its instance")
				}
			}
			continue
		}
		policy, ok := selection.(*nativeTestBlockSelection)
		if !ok || constructed != request || selected != request || len(policies) != request {
			t.Fatal("custom factory did not override built-in static selection")
		}
		if request == 2 && policies[0] == policy {
			t.Fatal("custom policy instance was reused across requests")
		}
		server.BlockSelector = policy
		thread := NewTLCServerThread(&rpcTestWorker{}, "tcp://worker/primary", server, policy)
		thread.cancelKeepAlive()
		queue := &distributedSelectorQueue{result: []*TLCStateMut{{UID: 9}}}
		if got := thread.Selector.GetBlocks(queue, thread.Worker); len(got) != 1 || got[0].UID != 9 || queue.requested != 3 || policy.queue != queue || policy.worker != thread.Worker || server.GetAverageBlockCnt() != 41 {
			t.Fatal("coordinator/thread did not retain the custom policy contract")
		}
		thread.Selector.SetMaxTXSize(5)
		thread.Selector.GetBlocks(queue, thread.Worker)
		if queue.requested != 5 {
			t.Fatal("custom transfer-limit setter was bypassed")
		}
		captureFailoverToolIO(t, ToolIOTool)
		thread.Worker = NewDistributedWorkerSmartProxy(&rpcTestWorker{next: func([]*TLCStateMut) (*NextStateResult, error) {
			return nil, workerComputationFailure("native test memory failure", NewOutOfMemoryError(), true)
		}})
		thread.setStates([]*TLCStateMut{{UID: 1}, {UID: 2}})
		retryQueue := NewMemStateQueue()
		if result, retry := thread.computeBlock(retryQueue); result != nil || !retry || retryQueue.Size() != 2 || policy.maximum != 1 {
			t.Fatal("smaller-batch retry did not requeue before updating the custom policy")
		}
	}
}
