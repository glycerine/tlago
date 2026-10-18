package tlc

import (
	"math"
	"runtime"
	"sync"
	"testing"
)

type threadStatisticsQueue struct {
	StateQueue
	remaining int
	state     *TLCStateMut
}

func (q *threadStatisticsQueue) SDequeueMany(int) []*TLCStateMut {
	if q.remaining == 0 {
		return nil
	}
	q.remaining--
	return []*TLCStateMut{q.state}
}

// No original Java method checks concurrent observation of server-thread
// statistics. Exercise the actual run/update/finalizer paths, keeping this
// native race check short and retaining the source signed-int overflow.
func TestDistributedThreadConcurrentStatistics(t *testing.T) {
	oldWorkers := NumWorkers()
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })
	const batches = 1000
	storage := NewMemFPSet()
	storage.Put(41)
	queue := &threadStatisticsQueue{StateQueue: NewMemStateQueue(), remaining: batches, state: &TLCStateMut{UID: 7}}
	server := &TLCServer{StateQueue: queue, FPSetManager: NewDistributedFPSetManager(NewLocalFingerprintEndpoint(storage))}
	thread := &TLCServerThread{
		Server: server, Selector: NewStaticBlockSelector(server, 1),
		SentStates: math.MaxInt32 - 3, ReceivedStates: math.MaxInt32 - 3,
		CacheRateHitRatio: -1, keepAliveDone: make(chan struct{}),
		Worker: NewDistributedWorkerSmartProxy(&rpcTestWorker{next: func([]*TLCStateMut) (*NextStateResult, error) {
			return NewNextStateResult([]*StateVec{NewStateVecFrom([]*TLCStateMut{{UID: 7}})}, []*LongVec{NewLongVecFrom([]int64{41})}, 1, 2), nil
		}}),
	}
	thread.TimerTask = &TLCTimerTask{Thread: thread}
	thread.setStates([]*TLCStateMut{})
	start, done := make(chan struct{}), make(chan struct{})
	var readers sync.WaitGroup
	for i := 0; i < 4; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			<-start
			for {
				thread.GetSentStates()
				thread.GetReceivedStates()
				ratio := thread.GetCacheRateRatio()
				if ratio != -1 && !math.IsNaN(ratio) {
					t.Errorf("unexpected cache statistic: %v", ratio)
				}
				select {
				case <-done:
					return
				default:
					runtime.Gosched()
				}
			}
		}()
	}
	close(start)
	thread.Run()
	close(done)
	readers.Wait()
	want := int(math.MinInt32 + batches - 4)
	if thread.GetSentStates() != want || thread.GetReceivedStates() != want || !math.IsNaN(thread.GetCacheRateRatio()) {
		t.Fatalf("final statistics = %d/%d/%v, want %d/%d/NaN", thread.GetSentStates(), thread.GetReceivedStates(), thread.GetCacheRateRatio(), want, want)
	}
	if !server.IsDone() || server.LastError != nil || server.WorkerStatesGenerated.Load() != batches || storage.Size() != 1 || queue.Size() != 0 {
		t.Fatal("statistics observation changed completion, generated-state accounting or fingerprint publication")
	}
}
