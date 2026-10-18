package tlc

import (
	"fmt"
	"sync"
	"testing"
)

var (
	javaStateQueueBenchmarkSizes  = [...]int{1, 2, 4, 8, 16, 32, 64}
	javaStateQueueBenchmarkResult [2][]*TLCStateMut
)

func javaQueueDummyState(vars int) *TLCStateMut {
	names := make([]string, vars)
	for i := range names {
		names[i] = fmt.Sprintf("v%d", i)
	}
	SetStateVariables(names)
	state := NewEmptyState()
	state.UID = 0
	for i, name := range names {
		state.Bind(UniqueStringOf(name), NewIntValue(int32(i)))
	}
	EmptyState = state
	return state
}

func prepareJavaStateQueueBenchmark(b *testing.B, size int) (*DiskStateQueue, []*TLCStateMut) {
	b.Helper()
	queue := NewDiskStateQueue(b.TempDir())
	batch := make([]*TLCStateMut, size)
	for i := range batch {
		batch[i] = javaQueueDummyState(1)
	}
	b.Cleanup(func() {
		if err := queue.Delete(); err != nil {
			b.Errorf("DiskStateQueue.delete: %v", err)
		}
		<-queue.reader.done
		<-queue.writer.done
		<-queue.cleaner.done
	})
	return queue, batch
}

func benchmarkJavaStateQueueGroup(
	b *testing.B,
	producer func(*DiskStateQueue, []*TLCStateMut),
	consumer func(*DiskStateQueue, int) []*TLCStateMut,
) {
	b.Helper()
	oldWorkers := NumWorkers()
	oldVariables, oldEmpty := stateVariables, EmptyState
	SetNumWorkers(1)
	b.Cleanup(func() {
		SetNumWorkers(oldWorkers)
		stateVariables = oldVariables
		EmptyState = oldEmpty
	})

	for _, size := range javaStateQueueBenchmarkSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			queue, batch := prepareJavaStateQueueBenchmark(b, size)
			start := make(chan struct{})
			var group sync.WaitGroup
			group.Add(4)
			for producerID := 0; producerID < 2; producerID++ {
				go func() {
					defer group.Done()
					<-start
					for i := 0; i < b.N; i++ {
						producer(queue, batch)
					}
				}()
			}
			for consumerID := 0; consumerID < 2; consumerID++ {
				id := consumerID
				go func() {
					defer group.Done()
					<-start
					var result []*TLCStateMut
					for i := 0; i < b.N; i++ {
						result = consumer(queue, size)
					}
					javaStateQueueBenchmarkResult[id] = result
				}()
			}
			b.ResetTimer()
			close(start)
			group.Wait()
		})
	}
}

func BenchmarkJavaStateQueueSingle(b *testing.B) {
	benchmarkJavaStateQueueGroup(b,
		func(queue *DiskStateQueue, batch []*TLCStateMut) {
			for _, state := range batch {
				queue.SEnqueue(state)
			}
		},
		func(queue *DiskStateQueue, size int) []*TLCStateMut {
			result := make([]*TLCStateMut, size)
			for i := range result {
				result[i] = queue.SDequeue()
			}
			return result
		})
}

func BenchmarkJavaStateQueueBatchAsymmetric(b *testing.B) {
	benchmarkJavaStateQueueGroup(b,
		func(queue *DiskStateQueue, batch []*TLCStateMut) { queue.SEnqueueAll(batch) },
		func(queue *DiskStateQueue, size int) []*TLCStateMut {
			result := make([]*TLCStateMut, size)
			for i := range result {
				result[i] = queue.SDequeue()
			}
			return result
		})
}

func BenchmarkJavaStateQueueBatchSymmetric(b *testing.B) {
	benchmarkJavaStateQueueGroup(b,
		func(queue *DiskStateQueue, batch []*TLCStateMut) { queue.SEnqueueAll(batch) },
		func(queue *DiskStateQueue, size int) []*TLCStateMut { return queue.SDequeueMany(size) })
}
