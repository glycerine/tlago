package tlc

import (
	"os"
	"strings"
	"testing"
)

type startupOwnerWorker struct {
	*rpcTestWorker
	cacheReads int
}

func (w *startupOwnerWorker) GetCacheRateRatio() (float64, error) {
	w.cacheReads++
	return 0, nil
}

// Java has no direct test of the required coordinator at run entry. Queue
// capture follows worker increment and precedes the run catch/finally region.
func TestDistributedThreadRunRequiresCoordinatorBeforeHandlers(t *testing.T) {
	oldWorkers := NumWorkers()
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })
	t.Run("missing_thread", func(t *testing.T) {
		SetNumWorkers(7)
		err := invokeDistributedServerOperation(func() error { (*TLCServerThread)(nil).Run(); return nil })
		if _, ok := err.(*NullPointerException); !ok || NumWorkers() != 7 {
			t.Fatalf("missing thread failure %T/%v, workers %d", err, err, NumWorkers())
		}
	})
	for _, asynchronous := range []bool{false, true} {
		name := "direct"
		if asynchronous {
			name = "owned_goroutine"
		}
		t.Run(name, func(t *testing.T) {
			SetNumWorkers(7)
			worker := &startupOwnerWorker{rpcTestWorker: &rpcTestWorker{}}
			thread := &TLCServerThread{Worker: NewDistributedWorkerSmartProxy(worker),
				CacheRateHitRatio: -1, keepAliveDone: make(chan struct{}), runDone: make(chan struct{})}
			state := &TLCStateMut{UID: 31, level: 4}
			thread.setStates([]*TLCStateMut{state})
			thread.cleanupGlobals.Store(true)
			t.Cleanup(thread.cancelKeepAlive)
			if asynchronous {
				output, err := os.CreateTemp(t.TempDir(), "startup-stderr-")
				if err != nil {
					t.Fatal(err)
				}
				previous := os.Stderr
				os.Stderr = output
				t.Cleanup(func() { os.Stderr = previous; _ = output.Close() })
				thread.Start()
				thread.Join()
				os.Stderr = previous
				data, err := os.ReadFile(output.Name())
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(data), "NullPointerException") {
					t.Fatalf("uncaught startup failure was not reported: %s", data)
				}
			} else {
				err := invokeDistributedServerOperation(func() error { thread.Run(); return nil })
				if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("missing coordinator failure %T/%v", err, err)
				}
			}
			if NumWorkers() != 8 || worker.cacheReads != 0 || thread.CacheRateHitRatio != -1 || !thread.cleanupGlobals.Load() || thread.keepAliveStopped.Load() || thread.GetCurrentSize() != 1 || thread.currentStates()[0] != state {
				t.Fatal("startup failure skipped worker increment or entered run catch/finally")
			}
		})
	}
}
