package tlc

import (
	"testing"
	"time"
)

func TestDFIDPostLevelChecksEveryWorker(t *testing.T) {
	first := &DFIDWorker{}
	second := &DFIDWorker{MoreLevel: true}
	checker := &DFIDModelChecker{DFIDWorkers: []*DFIDWorker{first, second}}
	if checker.shouldTerminateAfterLevel() {
		t.Fatal("second worker has another level; DFID stopped after worker zero")
	}

	second.SetStop(2)
	if !checker.shouldTerminateAfterLevel() {
		t.Fatal("second worker terminated; DFID did not inspect it")
	}

	second.SetStop(0)
	second.MoreLevel = false
	if !checker.shouldTerminateAfterLevel() {
		t.Fatal("no worker has another level; DFID did not terminate")
	}
}

func TestDFIDLevelCompletionJoinsEveryWorker(t *testing.T) {
	firstDone := make(chan struct{})
	secondDone := make(chan struct{})
	close(firstDone)
	checker := &DFIDModelChecker{
		DFIDWorkers:    []*DFIDWorker{{}, {}},
		dfidWorkerDone: []chan struct{}{firstDone, secondDone},
	}
	joined := make(chan struct{})
	go func() {
		checker.joinDFIDWorkers()
		close(joined)
	}()
	select {
	case <-joined:
		t.Fatal("DFID returned after worker zero while worker one was still active")
	case <-time.After(50 * time.Millisecond):
	}
	close(secondDone)
	select {
	case <-joined:
	case <-time.After(time.Second):
		t.Fatal("DFID did not return after every worker completed")
	}
}

func TestDFIDCompletionWakesCoordinator(t *testing.T) {
	early := &DFIDModelChecker{AbstractChecker: &AbstractChecker{}}
	early.SetDone()
	earlyWakeup := early.beginCoordinatorWait()
	select {
	case <-earlyWakeup:
	default:
		t.Fatal("DFID lost completion before the coordinator began waiting")
	}
	early.endCoordinatorWait()

	checker := &DFIDModelChecker{AbstractChecker: &AbstractChecker{}}
	wakeup := checker.beginCoordinatorWait()
	defer checker.endCoordinatorWait()
	go checker.SetDone()
	select {
	case <-wakeup:
	case <-time.After(time.Second):
		t.Fatal("DFID coordinator did not wake when a worker set done")
	}
	if !checker.Done.Load() {
		t.Fatal("DFID wake preceded the done flag")
	}
}

func TestDFIDCompletedLevelSkipsProgress(t *testing.T) {
	checker := &DFIDModelChecker{
		AbstractChecker: &AbstractChecker{},
		FPSet:           NewMemFPIntSet(),
	}
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)

	checker.SetDone()
	if checker.reportDFIDProgressWhileActive(1, 2) {
		t.Fatal("completed level continued periodic reporting")
	}
	if recorder.Recorded(ECTLCProgressStatsDFID) {
		t.Fatal("DFID reported progress after completion")
	}
	checker.Done.Store(false)
	if !checker.reportDFIDProgressWhileActive(1, 2) {
		t.Fatal("active level skipped periodic reporting")
	}
	if got := len(recorder.Records(ECTLCProgressStatsDFID)); got != 1 {
		t.Fatalf("active level recorded %d progress events, want 1", got)
	}
}

func TestDFIDProgressRecorderCanCompleteLevel(t *testing.T) {
	checker := &DFIDModelChecker{
		AbstractChecker: &AbstractChecker{},
		FPSet:           NewMemFPIntSet(),
	}
	recorder := RecorderFunc(func(message Message) {
		if message.Code == ECTLCProgressStatsDFID {
			checker.SetDone()
		}
	})
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)

	finished := make(chan bool, 1)
	go func() { finished <- checker.reportDFIDProgressWhileActive(1, 2) }()
	select {
	case active := <-finished:
		if !active || !checker.Done.Load() {
			t.Fatal("progress recorder did not complete the active level")
		}
	case <-time.After(time.Second):
		t.Fatal("progress recorder deadlocked on the checker monitor")
	}
}
