// Java source: test-benchmark/tlc2/tool/queue/DiskQueueBenachmark.java
// (Microsoft Research, MIT license).
package tlc

import (
	"fmt"
	"sync"
	"testing"
)

var (
	javaDiskQueueBenchmarkVars    = [...]int{1, 2, 4, 8, 16, 32, 64}
	javaDiskQueueBenchmarkImpls   = [...]string{"DiskByteArrayQueue", "DiskStateQueue"}
	javaDiskQueueBenchmarkResults []*TLCStateMut
)

func prepareJavaDiskQueueBenchmark(b *testing.B, impl string, vars int) (StateQueue, *TLCStateMut) {
	b.Helper()
	directory := b.TempDir()
	var queue StateQueue
	if impl == "DiskByteArrayQueue" {
		queue = NewDiskByteArrayQueue(directory)
	} else {
		queue = NewDiskStateQueue(directory)
	}
	state := javaQueueDummyState(vars)
	b.Cleanup(func() {
		if err := queue.Delete(); err != nil {
			b.Errorf("%s.delete: %v", impl, err)
		}
		switch q := queue.(type) {
		case *DiskByteArrayQueue:
			<-q.reader.done
			<-q.writer.done
			<-q.cleaner.done
		case *DiskStateQueue:
			<-q.reader.done
			<-q.writer.done
			<-q.cleaner.done
		}
	})
	return queue, state
}

func benchmarkJavaDiskQueueGroup(b *testing.B, threadsPerSide int) {
	b.Helper()
	oldWorkers := NumWorkers()
	oldVariables, oldEmpty := stateVariables, EmptyState
	SetNumWorkers(1)
	b.Cleanup(func() {
		SetNumWorkers(oldWorkers)
		stateVariables = oldVariables
		EmptyState = oldEmpty
	})

	for _, vars := range javaDiskQueueBenchmarkVars {
		for _, impl := range javaDiskQueueBenchmarkImpls {
			name := fmt.Sprintf("vars=%d/impl=%s", vars, impl)
			b.Run(name, func(b *testing.B) {
				queue, state := prepareJavaDiskQueueBenchmark(b, impl, vars)
				results := make([]*TLCStateMut, threadsPerSide)
				start := make(chan struct{})
				var group sync.WaitGroup
				group.Add(2 * threadsPerSide)
				for producerID := 0; producerID < threadsPerSide; producerID++ {
					go func() {
						defer group.Done()
						<-start
						for i := 0; i < b.N; i++ {
							queue.SEnqueue(state)
						}
					}()
				}
				for consumerID := 0; consumerID < threadsPerSide; consumerID++ {
					id := consumerID
					go func() {
						defer group.Done()
						<-start
						var result *TLCStateMut
						for i := 0; i < b.N; i++ {
							result = queue.SDequeue()
						}
						results[id] = result
					}()
				}
				b.ResetTimer()
				close(start)
				group.Wait()
				javaDiskQueueBenchmarkResults = results
			})
		}
	}
}

func BenchmarkJavaDiskQueueG02(b *testing.B) { benchmarkJavaDiskQueueGroup(b, 1) }
func BenchmarkJavaDiskQueueG04(b *testing.B) { benchmarkJavaDiskQueueGroup(b, 2) }
func BenchmarkJavaDiskQueueG08(b *testing.B) { benchmarkJavaDiskQueueGroup(b, 4) }
func BenchmarkJavaDiskQueueG16(b *testing.B) { benchmarkJavaDiskQueueGroup(b, 8) }
