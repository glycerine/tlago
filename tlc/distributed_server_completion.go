package tlc

import (
	"sync/atomic"
	"time"
)

var tlcServerFinalDistinctStates atomic.Int64

func init() { tlcServerFinalDistinctStates.Store(-1) }

// TLCServerFinalNumberOfDistinctStates is Java TLCServer's static result,
// shared across server instances and left unchanged by initialization failures.
func TLCServerFinalNumberOfDistinctStates() int64 {
	return tlcServerFinalDistinctStates.Load()
}

// waitForReportLocked implements Object.wait(REPORT_INTERVAL) on the server
// monitor. Notifications are lost when no thread is waiting, as in Java.
// The caller holds monitor on entry and again on return.
func (s *TLCServer) waitForReportLocked(interval int) error {
	if interval < 0 {
		return NewIllegalArgumentException("timeout value is negative")
	}
	waiter := make(chan struct{})
	s.completionWaiter = waiter
	depth := s.monitor.releaseForWait()
	if interval == 0 {
		<-waiter
	} else {
		timer := time.NewTimer(time.Duration(interval) * time.Millisecond)
		select {
		case <-waiter:
		case <-timer.C:
		}
		timer.Stop()
	}
	s.monitor.reacquireAfterWait(depth)
	if s.completionWaiter == waiter {
		s.completionWaiter = nil
	}
	return nil
}

func (s *TLCServer) notifyCompletionLocked() {
	if s.completionWaiter != nil {
		close(s.completionWaiter)
		s.completionWaiter = nil
	}
}

func (s *TLCServer) waitForDistributedCompletion() error {
	interval := TLCServerReportIntervalMillis()
	oldGenerated, oldDistinct := int64(0), uint64(0)
	if err := func() error {
		s.monitor.Lock()
		defer s.monitor.Unlock()
		return s.waitForReportLocked(interval)
	}(); err != nil {
		return err
	}
	for {
		// This is outside the synchronized block and runs even if a worker's
		// notification has already marked the server done.
		if DoCheckPoint() {
			if err := s.Checkpoint(); err != nil {
				return err
			}
		}
		done, err := func() (bool, error) {
			s.monitor.Lock()
			defer s.monitor.Unlock()
			if !s.IsDone() {
				generated, distinct, err := s.printProgressStatsLocked(oldGenerated, oldDistinct)
				if err != nil {
					return false, err
				}
				if err := s.waitForReportLocked(interval); err != nil {
					return false, err
				}
				// Java advances the baseline after waiting, not before it.
				oldGenerated, oldDistinct = generated, distinct
			}
			return s.IsDone(), nil
		}()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
}

func (s *TLCServer) printProgressStatsLocked(oldGenerated int64, oldDistinct uint64) (int64, uint64, error) {
	generated, distinct := s.GetStatesGenerated(), s.fpSetSize()
	factor := float64(TLCServerReportIntervalMillis()) / 60000
	s.StatesPerMinute = javaDoubleToLong(float64(generated-oldGenerated) / factor)
	s.DistinctStatesPerMinute = javaDoubleToLong(float64(int64(distinct)-int64(oldDistinct)) / factor)
	level, err := s.Trace.GetLevelForReportingWithError()
	if err != nil {
		return generated, distinct, err
	}
	PrintMessage(ECTLCProgressStats, fmtInt(level),
		MessageNumberFormat(generated),
		MessageNumberFormat(int64(distinct)),
		MessageNumberFormat(s.getNewStatesLocked()),
		MessageNumberFormat(s.StatesPerMinute),
		MessageNumberFormat(s.DistinctStatesPerMinute))
	return generated, distinct, nil
}

// PrintProgressStats retains the native observation helper while using the
// distributed server's own interval and Java arithmetic/formatting.
func (s *TLCServer) PrintProgressStats(startTime time.Time, oldGenerated *int64, oldDistinct *uint64) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	var generated int64
	var distinct uint64
	if oldGenerated != nil {
		generated = *oldGenerated
	}
	if oldDistinct != nil {
		distinct = *oldDistinct
	}
	generated, distinct, err := s.printProgressStatsLocked(generated, distinct)
	if err != nil {
		return err
	}
	if oldGenerated != nil {
		*oldGenerated = generated
	}
	if oldDistinct != nil {
		*oldDistinct = distinct
	}
	return nil
}
