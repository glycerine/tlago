package tlc

import "time"

// fpSetLifecycle represents FPSet's inherited Object monitor and wait set.
// The memory sets use the same reentrant monitor for their synchronized methods.
type fpSetLifecycle struct {
	monitor distributedServerMonitor
	waiters []chan struct{}
}

func (f *fpSetLifecycle) fpSetLifecycleState() *fpSetLifecycle    { return f }
func (f *fpSetLifecycle) fpSetMonitor() *distributedServerMonitor { return &f.monitor }

func fpSetLifecycleAndMonitor(set FPSet) (*fpSetLifecycle, *distributedServerMonitor) {
	f, ok := set.(interface {
		fpSetLifecycleState() *fpSetLifecycle
		fpSetMonitor() *distributedServerMonitor
	})
	if !ok {
		panic(NewClassCastException())
	}
	return f.fpSetLifecycleState(), f.fpSetMonitor()
}

func (f *fpSetLifecycle) wait(monitor *distributedServerMonitor, duration time.Duration) error {
	waiter := make(chan struct{})
	f.waiters = append(f.waiters, waiter)
	depth := monitor.releaseForWait()
	timer := time.NewTimer(duration)
	select {
	case <-waiter:
	case <-timer.C:
	}
	timer.Stop()
	monitor.reacquireAfterWait(depth)
	for i, pending := range f.waiters {
		if pending == waiter {
			f.waiters = append(f.waiters[:i], f.waiters[i+1:]...)
			break
		}
	}
	return nil
}

func fpSetBaseExit(set FPSet) {
	ShutdownDistributedFPServer()
	lifecycle, monitor := fpSetLifecycleAndMonitor(set)
	monitor.Lock()
	defer monitor.Unlock()
	// Object.notify wakes one arbitrary waiter and stores no notification
	// when nobody is waiting. The order here is deterministic for Go callers.
	if len(lifecycle.waiters) != 0 {
		close(lifecycle.waiters[0])
		lifecycle.waiters = lifecycle.waiters[1:]
	}
}
