package tlc

import (
	"math"
	"sync"
	"sync/atomic"
)

// SimulationWorkerResultQueue preserves the FIFO, blocking take and default
// capacity of Simulator's LinkedBlockingQueue. A small bounded channel can
// prevent a worker from terminating while its consumer joins it.
type SimulationWorkerResultQueue struct {
	mu         sync.Mutex
	notEmpty   *sync.Cond
	notFull    *sync.Cond
	head, tail *simulationWorkerResultNode
	count      int
}

type simulationWorkerResultNode struct {
	result SimulationWorkerResult
	next   *simulationWorkerResultNode
}

func NewSimulationWorkerResultQueue() *SimulationWorkerResultQueue {
	q := &SimulationWorkerResultQueue{}
	q.notEmpty = sync.NewCond(&q.mu)
	q.notFull = sync.NewCond(&q.mu)
	return q
}

func (q *SimulationWorkerResultQueue) Put(result SimulationWorkerResult) {
	_ = q.putInterruptibly(result, nil)
}

func (q *SimulationWorkerResultQueue) putInterruptibly(result SimulationWorkerResult, interrupted *atomic.Bool) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	for {
		if interrupted != nil && interrupted.Load() {
			return NewInterruptedException()
		}
		// LinkedBlockingQueue's default constructor uses Integer.MAX_VALUE.
		if q.count < math.MaxInt32 {
			break
		}
		q.notFull.Wait()
	}
	q.enqueue(result)
	return nil
}

func (q *SimulationWorkerResultQueue) enqueue(result SimulationWorkerResult) {
	node := &simulationWorkerResultNode{result: result}
	if q.tail == nil {
		q.head = node
	} else {
		q.tail.next = node
	}
	q.tail = node
	q.count++
	q.notEmpty.Signal()
}

func (q *SimulationWorkerResultQueue) Offer(result SimulationWorkerResult) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.count == math.MaxInt32 {
		return false
	}
	q.enqueue(result)
	return true
}

func (q *SimulationWorkerResultQueue) signalInterruptedProducer() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.notFull.Broadcast()
}

func (q *SimulationWorkerResultQueue) Take() SimulationWorkerResult {
	q.mu.Lock()
	defer q.mu.Unlock()
	for q.head == nil {
		q.notEmpty.Wait()
	}
	node := q.head
	q.head = node.next
	if q.head == nil {
		q.tail = nil
	}
	q.count--
	q.notFull.Signal()
	return node.result
}

func (q *SimulationWorkerResultQueue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.count == 0
}
