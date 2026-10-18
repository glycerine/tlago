//go:build tlc_verify

package tlc

import (
	"fmt"
	"strings"
	"testing"
)

// Original StateQueueJPFTest workload reduced at its two monitor boundaries.
// Payload identity is constant: DummyStateQueue always returns the seeded
// pointer, and each worker enqueues that same pointer after each dequeue.
// This explores the bounded source model, not arbitrary production goroutines
// or the Java memory model. No interruptions or spurious wakeups are injected.
// Waiting-count increments remain separate from the final-worker decision,
// since the checkpoint monitor can observe that count while a worker owns
// the queue monitor. Nested monitor acquisition is also a separate step.
// The checkpoint finish read remains separate from its waiting-count recheck.
// Its count-check/wait pair can be one transition: holding that monitor
// excludes any notification until wait atomically releases it. Pure local
// loop bookkeeping and queue-protected payload operations introduce no other
// observation boundaries in this constant-pointer source workload.
type javaStateQueueSchedule struct {
	length, waiting             int
	finish, stop                bool
	queueOwner, checkpointOwner int // -1 free, 0 main, 1/2 workers
	main                        string
	workers                     [2]string
	iterations                  [2]int
}

type javaStateQueueTransition struct {
	next  javaStateQueueSchedule
	label string
}

func (s javaStateQueueSchedule) successors() []javaStateQueueTransition {
	var out []javaStateQueueTransition
	emit := func(n javaStateQueueSchedule, label string) {
		out = append(out, javaStateQueueTransition{n, label})
	}
	wakeWorkers := func(n javaStateQueueSchedule) javaStateQueueSchedule {
		for i, phase := range n.workers {
			if phase == "wait" {
				n.workers[i] = "wake"
			}
		}
		return n
	}
	n := s
	switch s.main {
	case "start":
		if s.queueOwner == -1 {
			n.queueOwner = 0
			n.main = "initial"
			emit(n, "main acquire queue")
		}
	case "initial":
		n.queueOwner = -1
		if s.finish {
			n.main = "done"
		} else {
			n.stop = true
			if s.waiting < 2 {
				n.main = "monitor"
			} else {
				n.main = "resume"
			}
		}
		emit(n, "main initial finish/count check and release queue")
	case "monitor":
		if s.checkpointOwner == -1 {
			n.checkpointOwner = 0
			n.main = "finish"
			emit(n, "main acquire checkpoint monitor")
		}
	case "finish":
		if s.finish {
			n.checkpointOwner = -1
			n.main = "done"
		} else {
			n.main = "count"
		}
		emit(n, "main finish recheck")
	case "count":
		// Preserve needsWaiting's recheck while holding mu. A worker may already
		// have notified mu between the initial count check and acquiring mu.
		const barrierRecheck = true
		n.checkpointOwner = -1
		if barrierRecheck && s.waiting >= 2 {
			n.main = "resume"
		} else {
			n.main = "wait"
		}
		emit(n, "main barrier recheck or wait releasing checkpoint monitor")
	case "wake":
		if s.checkpointOwner == -1 {
			n.checkpointOwner = 0
			n.main = "release"
			emit(n, "main reacquire checkpoint monitor after wake")
		}
	case "release":
		n.checkpointOwner = -1
		n.main = "recheck"
		emit(n, "main release checkpoint monitor")
	case "recheck":
		if s.queueOwner == -1 {
			n.queueOwner = 0
			n.main = "recheck-held"
			emit(n, "main reacquire queue")
		}
	case "recheck-held":
		n.queueOwner = -1
		if s.finish {
			n.main = "done"
		} else if s.waiting < 2 {
			n.main = "monitor"
		} else {
			n.main = "resume"
		}
		emit(n, "main finish/count recheck under queue monitor")
	case "resume":
		if s.queueOwner == -1 {
			n.queueOwner = 0
			n.main = "resume-held"
			emit(n, "main acquire queue to resume")
		}
	case "resume-held":
		n.queueOwner = -1
		n.stop = false
		n.main = "done"
		emit(wakeWorkers(n), "main resume and notify all workers")
	}
	for i, phase := range s.workers {
		n := s
		put := func(nextPhase, label string) {
			n.workers[i] = nextPhase
			emit(n, fmt.Sprintf("worker %d %s", i, label))
		}
		switch phase {
		case "start":
			if s.queueOwner == -1 {
				n.queueOwner = i + 1
				put("available", "acquire queue")
			}
		case "available":
			if s.finish {
				n.queueOwner = -1
				put("finish", "dequeue returns null")
			} else {
				put("condition", "initial finish read")
			}
		case "condition":
			if s.length == 0 || s.stop {
				n.waiting++
				put("count", "increment waiting")
			} else {
				n.length--
				n.queueOwner = -1
				put("enqueue", "dequeue seed pointer and release queue")
			}
		case "count":
			if s.waiting >= 2 {
				if s.length == 0 {
					n.waiting--
					n.queueOwner = -1
					put("finish", "no work returns null")
				} else {
					put("notify-monitor", "last worker needs checkpoint monitor")
				}
			} else {
				put("sleep", "worker needs wait")
			}
		case "notify-monitor":
			if s.checkpointOwner == -1 {
				n.checkpointOwner = i + 1
				put("notify-held", "acquire checkpoint monitor retaining queue")
			}
		case "notify-held":
			if s.main == "wait" {
				n.main = "wake"
			}
			n.checkpointOwner = -1
			put("sleep", "notify main and release checkpoint monitor")
		case "sleep":
			n.queueOwner = -1
			put("wait", "wait releases queue")
		case "wake":
			if s.queueOwner == -1 {
				n.queueOwner = i + 1
				put("awoke", "reacquire queue after wake")
			}
		case "awoke":
			n.waiting--
			if s.finish {
				n.queueOwner = -1
				put("finish", "wake finished returns null")
			} else {
				put("condition", "decrement waiting and retry")
			}
		case "enqueue":
			if s.queueOwner == -1 {
				n.queueOwner = i + 1
				put("enqueue-held", "acquire queue to enqueue")
			}
		case "enqueue-held":
			n.length++
			n.queueOwner = -1
			if s.waiting > 0 && !s.stop {
				n = wakeWorkers(n)
			}
			n.iterations[i]++
			if n.iterations[i] == 3 {
				put("finish", "enqueue seed pointer and complete three iterations")
			} else {
				put("start", "enqueue seed pointer and advance iteration")
			}
		case "finish":
			if s.queueOwner == -1 {
				n.queueOwner = i + 1
				put("finish-held", "acquire queue to finish")
			}
		case "finish-held":
			n.finish = true
			n = wakeWorkers(n)
			put("finish-monitor", "finish and notify all workers")
		case "finish-monitor":
			if s.checkpointOwner == -1 {
				n.checkpointOwner = i + 1
				put("finish-notify", "acquire checkpoint monitor retaining queue")
			}
		case "finish-notify":
			if s.main == "wait" {
				n.main = "wake"
			}
			n.queueOwner = -1
			n.checkpointOwner = -1
			put("done", "notify main and finish")
		}
	}
	return out
}

func verifyJavaStateQueueSchedules(t *testing.T) {
	t.Helper()
	// Starting main before workers does not require it to execute first. The
	// parent has no further shared mutations; joins require all actors done.
	initial := javaStateQueueSchedule{length: 1, queueOwner: -1, checkpointOwner: -1, main: "start", workers: [2]string{"start", "start"}}
	type predecessor struct {
		previous javaStateQueueSchedule
		label    string
	}
	seen := map[javaStateQueueSchedule]predecessor{initial: {}}
	pending := []javaStateQueueSchedule{initial}
	transitions, terminal := 0, 0
	for head := 0; head < len(pending); head++ {
		current := pending[head]
		next := current.successors()
		if len(next) == 0 {
			if current.main == "done" && current.workers == [2]string{"done", "done"} {
				terminal++
			} else {
				var witness []string
				for state := current; state != initial; {
					entry := seen[state]
					witness = append(witness, entry.label)
					state = entry.previous
				}
				for l, r := 0, len(witness)-1; l < r; l, r = l+1, r-1 {
					witness[l], witness[r] = witness[r], witness[l]
				}
				t.Fatalf("original queue workload deadlocks at %+v\n%s", current, strings.Join(witness, "\n"))
			}
		}
		for _, edge := range next {
			transitions++
			if _, ok := seen[edge.next]; !ok {
				seen[edge.next] = predecessor{current, edge.label}
				pending = append(pending, edge.next)
			}
		}
	}
	t.Logf("original queue monitor model: %d states, %d transitions, %d terminal states", len(seen), transitions, terminal)
}
