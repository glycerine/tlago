//go:build tlc_verify

package tlc

import (
	"os"
	"os/exec"
	"sync"
	"testing"
)

// Original DiskStateQueueJPFTest.testDeadlockFreedom workload and writer join.
// Source pool models feed the actual native queue, pool workers and shared
// state serialization path. The bounded explorer supplements their execution.
// TLAGO_DISK_QUEUE_VERIFY_BYTES=1 retains the former real-file diagnostic,
// whose byte encoding differs from the source JPF raw-long stream contract.
func TestJavaDiskStateQueueVerification(t *testing.T) {
	const child = "TLAGO_DISK_QUEUE_VERIFY_CHILD"
	if os.Getenv(child) != "1" {
		verifyJavaDiskQueueSchedules(t)
		// Source sets BufSize before class initialization. A fresh process keeps
		// the native startup snapshot independent of other queue tests.
		command := exec.Command(os.Args[0], "-test.run=^TestJavaDiskStateQueueVerification$", "-test.timeout=30s", "-test.v")
		command.Env = append(os.Environ(), child+"=1")
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("original disk queue workload: %v\n%s", err, output)
		}
		t.Logf("%s", output)
		return
	}

	const bufferSize = 1
	const workers = 1
	const enqueues = bufferSize + 1
	tlcSetSystemProperty(diskStateQueueBufferSizeProperty, "1")
	SetNumWorkers(workers)
	// DummyTLCState writes only the base header. Its distinct createEmpty
	// override corresponds to a fresh native empty-state allocation per read.
	stateVariables = nil
	SetUniqueStringVariableCount(0)
	EmptyState = newEmptyStateOfKind(false)
	state := newEmptyStateOfKind(false)
	state.UID = 0 // Only the source no-argument Dummy constructor sets uid=0.
	var streams statePoolStreamFactory
	var pool *javaQueuePoolFiles
	if os.Getenv("TLAGO_DISK_QUEUE_VERIFY_BYTES") != "1" {
		pool = &javaQueuePoolFiles{}
		streams = pool
	}
	queue := newDiskStateQueueWithStreams(t.TempDir(), streams)
	// Join native owned pool workers before fixture removal at child exit.
	defer func() {
		queue.finishConstructionAndWait()
		if pool != nil {
			pool.mu.Lock()
			defer pool.mu.Unlock()
			t.Logf("source pool model: records=%v, completed native prefetches=%d", pool.values, pool.completedReads)
		}
	}()
	queue.Enqueue(state)
	states := make([]*TLCStateMut, enqueues)
	for i := range states {
		// Dummy(fp=i+1) retains INIT_UID; fp itself is neither serialized nor
		// consulted by this queue workload. Keep each allocation distinct.
		states[i] = newEmptyStateOfKind(false)
	}
	var threads sync.WaitGroup
	for i := 0; i < workers; i++ {
		threads.Add(1)
		go func() {
			defer threads.Done()
			// Original WorkerTask(1, queue, states).
			state := queue.SDequeue()
			if state == nil {
				queue.FinishAll()
				return
			}
			queue.SEnqueueAll(states)
			queue.FinishAll()
		}()
	}
	threads.Add(1)
	go func() {
		defer threads.Done()
		// Original MainTask, started after the worker.
		if queue.SuspendAll() {
			queue.ResumeAll()
		}
	}()
	threads.Wait()
	// A receive on done joins the actual writer and establishes termination,
	// corresponding to source writer.join(); assert !writer.isAlive().
	<-queue.writer.done
}
