package tlc

import (
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

type orderedExitWorker struct {
	*rpcTestWorker
	failure   error
	panics    bool
	exitCalls int
}

func (w *orderedExitWorker) Exit() error {
	w.exitCalls++
	if w.panics {
		panic(w.failure)
	}
	return w.failure
}

func workerExitFailure(family string) error {
	switch family {
	case "connect":
		return workerConnectionFailure(net.ErrClosed)
	case "missing":
		return workerEndpointRemovedFailure("worker unavailable")
	case "server":
		failure := distributedTestRemoteFailure("worker server failure")
		failure.ExitIgnorable = true
		return failure
	case "remote":
		return distributedTestRemoteFailure("worker remote failure")
	case "io":
		return NewIOException("worker I/O failure")
	case "runtime":
		return NewRuntimeException("worker runtime failure")
	case "fatal":
		return NewAssertionError("worker fatal failure")
	}
	panic("unknown worker failure family")
}

func registerCompletedExitThread(server *TLCServer, worker DistributedWorkerEndpoint, uri string) {
	done := make(chan struct{})
	close(done)
	thread := &TLCServerThread{Server: server, Worker: NewDistributedWorkerSmartProxy(worker), URI: uri, CacheRateHitRatio: -1, runDone: done}
	thread.setStates([]*TLCStateMut{})
	server.RegisterTLCServerThread(thread)
}

// The source has no direct worker-exit failure-order methods. The shutdown hook
// has narrower silent catches than normal completion, including over Go payloads.
func TestDistributedShutdownWorkerExitCatches(t *testing.T) {
	for _, family := range []string{"connect", "missing", "server", "remote", "io", "runtime", "fatal"} {
		for _, encoded := range []bool{false, true} {
			for _, panics := range []bool{false, true} {
				t.Run(family+"/encoded="+strconv.FormatBool(encoded)+"/panic="+strconv.FormatBool(panics), func(t *testing.T) {
					captureFailoverToolIO(t, ToolIOTool)
					failure := workerExitFailure(family)
					if encoded {
						payload, err := EncodeDistributedFailure(failure)
						if err != nil {
							t.Fatal(err)
						}
						failure, err = DecodeDistributedFailure(payload)
						if err != nil {
							t.Fatal(err)
						}
					}
					first := &orderedExitWorker{rpcTestWorker: &rpcTestWorker{}, failure: failure, panics: panics}
					second := &orderedExitWorker{rpcTestWorker: &rpcTestWorker{}}
					server := &TLCServer{}
					registerCompletedExitThread(server, first, "tcp://first:1/worker")
					registerCompletedExitThread(server, second, "tcp://second:1/worker")
					server.ConfigurePublication(TLCServerPublication{GetRegistry: func(int) (*TLCServerRegistry, error) {
						return &TLCServerRegistry{Lookup: func(string) (*TLCServer, error) { return server, nil }}, nil
					}})
					recorder := &MemoryRecorder{}
					AddMessageRecorder(recorder)
					defer RemoveMessageRecorder(recorder)
					err := server.RunWorkerShutdownHook()
					unchecked := family == "runtime" || family == "fatal"
					if unchecked && err != failure || !unchecked && err != nil {
						t.Fatalf("shutdown catch = %T/%v", err, err)
					}
					wantSecond := 1
					if unchecked {
						wantSecond = 0
					}
					wantErrors := 1
					if unchecked || family == "connect" || family == "missing" {
						wantErrors = 0
					}
					if first.exitCalls != 1 || second.exitCalls != wantSecond || server.GetWorkerCount() != 2 || len(recorder.Records(ECGeneral)) != wantErrors {
						t.Fatal("shutdown changed catch scope, iteration, diagnostics or registration ownership")
					}
					if wantErrors == 1 {
						// MP's GENERAL throwable overload formats a string parameter;
						// it does not record the Throwable object itself.
						message := recorder.Records(ECGeneral)[0]
						if message.Throwable != nil || len(message.Params) != 1 || !strings.Contains(message.Params[0], failure.Error()) {
							t.Fatal("shutdown changed the source GENERAL diagnostic event")
						}
					}
				})
			}
		}
	}
}

func TestDistributedFinalWorkerExitOrderProcess(t *testing.T) {
	if scenario := os.Getenv("TLAGO_FINAL_WORKER_EXIT"); scenario != "" {
		checkFinalWorkerExitOrder(t, scenario)
		return
	}
	for _, family := range []string{"connect", "missing", "server", "remote", "runtime", "fatal"} {
		for _, panics := range []string{"false", "true"} {
			scenario := family + "/" + panics
			t.Run(scenario, func(t *testing.T) {
				command := exec.Command(os.Args[0], "-test.run=^TestDistributedFinalWorkerExitOrderProcess$")
				command.Env = append(os.Environ(), "TLAGO_FINAL_WORKER_EXIT="+scenario, tlcServerPropertyPrefix+".report=1")
				if output, err := command.CombinedOutput(); err != nil {
					t.Fatalf("native exit process failed: %v\n%s", err, output)
				}
			})
		}
	}
}

func TestDistributedShutdownUnavailableWorkersOverTCP(t *testing.T) {
	for _, phase := range []string{"unpublished", "closed_connection"} {
		t.Run(phase, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			_, client := startWorkerRPC(t, &rpcTestWorker{})
			if phase == "unpublished" {
				if err := client.Exit(); err != nil {
					t.Fatal(err)
				}
			} else if err := client.CloseConnection(); err != nil {
				t.Fatal(err)
			}
			failure := client.Exit()
			if !isDistributedWorkerUnavailable(failure) || !isIgnorableDistributedWorkerExit(failure) {
				t.Fatalf("native unavailable worker lost exit traits: %T/%v", failure, failure)
			}
			if isDistributedWorkerEndpointRemoved(failure) != (phase == "unpublished") {
				t.Fatal("endpoint removal became indistinguishable from connection closure")
			}
			server := &TLCServer{}
			second := &orderedExitWorker{rpcTestWorker: &rpcTestWorker{}}
			registerCompletedExitThread(server, client, "tcp://unavailable:1/worker")
			registerCompletedExitThread(server, second, "tcp://second:1/worker")
			server.ConfigurePublication(TLCServerPublication{GetRegistry: func(int) (*TLCServerRegistry, error) {
				return &TLCServerRegistry{Lookup: func(string) (*TLCServer, error) { return server, nil }}, nil
			}})
			recorder := &MemoryRecorder{}
			AddMessageRecorder(recorder)
			defer RemoveMessageRecorder(recorder)
			if err := server.RunWorkerShutdownHook(); err != nil || second.exitCalls != 1 || server.GetWorkerCount() != 2 || recorder.Recorded(ECGeneral) {
				t.Fatalf("native unavailable shutdown catch changed: %v", err)
			}
		})
	}
}

func checkFinalWorkerExitOrder(t *testing.T, scenario string) {
	captureFailoverToolIO(t, ToolIOTool)
	family, panics, _ := strings.Cut(scenario, "/")
	failure := workerExitFailure(family)
	first := &orderedExitWorker{rpcTestWorker: &rpcTestWorker{}, failure: failure, panics: panics == "true"}
	second := &orderedExitWorker{rpcTestWorker: &rpcTestWorker{}}
	directory := t.TempDir()
	trace := NewTLCTrace(directory, "Spec")
	defer trace.Close()
	store := NewMemFPSet()
	store.Init(1, directory, "Spec")
	server := NewTLCServer("Spec", "Spec", directory, NewNonDistributedFPSetManager(store, "local", trace), NewMemStateQueue(directory), trace)
	server.SetTool(NewTool())
	registerCompletedExitThread(server, first, "tcp://first:1/worker")
	registerCompletedExitThread(server, second, "tcp://second:1/worker")
	server.ConfigurePublication(TLCServerPublication{
		LocalHostName: func() (string, error) { return "local", nil },
		CreateRegistry: func(int) (*TLCServerRegistry, error) {
			return &TLCServerRegistry{Rebind: func(name string, s *TLCServer) error {
				if name == TLCServerWorkerName {
					s.SetDone()
				}
				return nil
			}, Unbind: func(string) error { return nil }}, nil
		},
		Unexport: func(*TLCServer, bool) (bool, error) { return true, nil },
	})
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	code, err := server.ModelCheck()
	ignored := family == "connect" || family == "missing" || family == "server"
	if ignored {
		if code != NoError || err != nil || first.exitCalls != 1 || second.exitCalls != 1 || server.GetWorkerCount() != 0 || !server.executor.IsShutdown() || TLCServerFinalNumberOfDistinctStates() != 0 {
			t.Fatalf("ignored exit interrupted completion: %d/%v", code, err)
		}
		if len(recorder.Records(ECGeneral)) != 1 || len(recorder.Records(ECTLCDistributedWorkerStats)) != 2 || !recorder.Recorded(ECTLCFinished) {
			t.Fatal("ignored exit changed warnings or final reporting")
		}
	} else {
		if code != ECGeneral || err != failure || first.exitCalls != 1 || second.exitCalls != 0 || server.GetWorkerCount() != 1 || server.executor.IsShutdown() || TLCServerFinalNumberOfDistinctStates() != -1 {
			t.Fatalf("unhandled exit continued or lost removal: %d/%v", code, err)
		}
		if recorder.Recorded(ECGeneral) || len(recorder.Records(ECTLCDistributedWorkerStats)) != 1 || recorder.Recorded(ECTLCFinished) {
			t.Fatal("unhandled exit entered a catch or later reporting")
		}
	}
	if recorder.Recorded(ECTLCDistributedWorkerDeregistered) {
		t.Fatal("final removal emitted worker-loss deregistration")
	}
}
