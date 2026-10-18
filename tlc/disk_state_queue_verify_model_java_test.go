//go:build tlc_verify

// Copyright (c) 2026 NVIDIA Corp. All rights reserved.
// Copyright (c) 2003 Compaq Corporation. All rights reserved.
package tlc

import (
	"strings"
	"testing"
)

// The original DiskStateQueueJPFTest has a one-entry buffer and one worker
// iteration. Only the seed is dequeued; appending its two distinct states
// spills exactly one file, pool/0. PoolFiles has exactly one name slot.
// The source streams record signed longs, including INIT_UID=-1, directly.
// This is the original test's storage model, not native byte serialization.
// The only initial fill takes the no-file/getCache-miss branch. No later
// worker dequeue exists, so doWork's late-file and the queue's later-pool
// branches are unreachable. The cleaner cannot delete: loPool stays 1 and
// its threshold is 100. Its independent leaf monitor/daemon wait therefore
// cannot form a cycle and is collapsed into finish-reader-held.
// Keep the writer monitor across recording, commit and nested reader wakeup;
// keep the reader monitor across prefetch. PoolFiles' own synchronized atomic
// operations are separate transitions, including open, commit and close.
// Parent join/termination is explicit: daemon reader operations remain possible
// after writer termination until the parent finishes its final assertion.
// Fixed three-record storage is the exact reachable prefix of the source
// eight-slot output buffer: workerId=Short.MAX_VALUE, uid, level=1. No value
// serialization, extra pool names, interrupts or spurious wakes occur here.
type javaDiskQueueSchedule struct {
	length, waiting                         int
	finish, stop                            bool
	q, mu, writerLock, readerLock           int // -1 free; main=0,worker=1,writer=2,reader=3
	main, worker, writer, reader, parent    string
	writerFinished, readerFinished, pending bool
	readerFile, readerFull, canRead         bool
	buffers                                 [4]uint8 // deq/enq/writer/reader allocation slots: 0=nil,1=seed,2/3=distinct appends,4=decoded
	enqBuffer, writerBuffer                 uint8
	enqIndex, hiPool                        uint8
	poolExists, poolWriting                 bool
	poolReaders                             int
	output                                  [3]int64
	outputCount                             uint8
	committed                               [3]int64
	input                                   [3]int64
	inputCount                              uint8
	decoded                                 [3]int64
	violation                               string
}

type javaDiskQueueEdge struct {
	next  javaDiskQueueSchedule
	label string
}

func (s *javaDiskQueueSchedule) notifyMain() {
	if s.main == "wait" {
		s.main = "wake"
	}
}
func (s *javaDiskQueueSchedule) notifyWriter() {
	if s.writer == "wait" {
		s.writer = "wake"
	}
}
func (s *javaDiskQueueSchedule) notifyReader() {
	if s.reader == "wait" {
		s.reader = "wake"
	}
}
func (s *javaDiskQueueSchedule) openPoolWriter() {
	if s.poolWriting {
		s.violation = "PoolFiles: two writers open the same pool"
		return
	}
	if s.poolReaders > 0 {
		s.violation = "PoolFiles: write while pool has readers"
		return
	}
	s.poolWriting = true
	s.outputCount = 0
	s.output = [3]int64{}
}
func (s *javaDiskQueueSchedule) recordPoolValue() {
	state := s.buffers[s.writerBuffer]
	if state == 0 {
		s.violation = "null state in writer buffer"
		return
	}
	values := [3]int64{32767, -1, 1}
	if state == 1 {
		values[1] = 0
	}
	s.output[s.outputCount] = values[s.outputCount]
	s.outputCount++
}
func (s *javaDiskQueueSchedule) commitPool() {
	s.committed = s.output
	s.poolExists = true
	s.poolWriting = false
}
func (s *javaDiskQueueSchedule) openPoolReader() {
	if s.poolWriting {
		s.violation = "PoolFiles: read while pool is open for writing"
		return
	}
	if !s.poolExists {
		s.violation = "PoolFiles: read a pool never written"
		return
	}
	s.poolReaders++
	s.input = s.committed
	s.inputCount = 0
}
func (s *javaDiskQueueSchedule) readPoolValue() {
	if s.inputCount == 3 {
		s.violation = "PoolFiles: read beyond written values"
		return
	}
	value := s.input[s.inputCount]
	if s.inputCount != 1 {
		value = int64(int16(value))
	} // source shortNat casts
	s.decoded[s.inputCount] = value
	s.inputCount++
}
func (s *javaDiskQueueSchedule) closePoolReader() {
	s.poolReaders--
	if s.inputCount != 3 {
		s.violation = "PoolFiles: close with unread values"
		return
	}
	if s.decoded[2] < 0 {
		s.violation = "TLCState.read: negative level"
	}
}

func (s javaDiskQueueSchedule) successors() []javaDiskQueueEdge {
	if s.parent == "done" || s.violation != "" {
		return nil
	}
	var out []javaDiskQueueEdge
	emit := func(n javaDiskQueueSchedule, label string) { out = append(out, javaDiskQueueEdge{n, label}) }
	n := s
	switch s.main {
	case "start":
		if s.q == -1 {
			n.q = 0
			n.main = "initial"
			emit(n, "main acquire queue")
		}
	case "initial":
		n.q = -1
		if s.finish {
			n.main = "done"
		} else {
			n.stop = true
			if s.waiting < 1 {
				n.main = "monitor"
			} else {
				n.main = "resume"
			}
		}
		emit(n, "main publish stop/count under queue")
	case "monitor":
		if s.mu == -1 {
			n.mu = 0
			n.main = "finish"
			emit(n, "main acquire checkpoint monitor")
		}
	case "finish":
		if s.finish {
			n.mu = -1
			n.main = "done"
		} else {
			n.main = "count"
		}
		emit(n, "main finish recheck")
	case "count":
		n.mu = -1
		if s.waiting >= 1 {
			n.main = "resume"
		} else {
			n.main = "wait"
		}
		emit(n, "main barrier recheck or wait")
	case "wake":
		if s.mu == -1 {
			n.mu = 0
			n.main = "release"
			emit(n, "main reacquire checkpoint monitor")
		}
	case "release":
		n.mu = -1
		n.main = "recheck"
		emit(n, "main release checkpoint monitor")
	case "recheck":
		if s.q == -1 {
			n.q = 0
			n.main = "recheck-held"
			emit(n, "main reacquire queue")
		}
	case "recheck-held":
		n.q = -1
		if s.finish {
			n.main = "done"
		} else if s.waiting < 1 {
			n.main = "monitor"
		} else {
			n.main = "resume"
		}
		emit(n, "main recheck under queue")
	case "resume":
		if s.q == -1 {
			n.q = 0
			n.main = "resume-held"
			emit(n, "main acquire queue to resume")
		}
	case "resume-held":
		n.q = -1
		n.stop = false
		n.main = "done"
		if s.worker == "wait" {
			n.worker = "wake"
		}
		emit(n, "main resume and notify worker")
	}
	n = s
	w := func(phase, label string) { n.worker = phase; emit(n, "worker "+label) }
	switch s.worker {
	case "start":
		if s.q == -1 {
			n.q = 1
			w("available", "acquire queue")
		}
	case "available":
		if s.finish {
			n.q = -1
			w("finish", "dequeue returns null")
		} else {
			w("condition", "initial finish read")
		}
	case "condition":
		if s.length == 0 || s.stop {
			n.waiting++
			w("count", "increment waiting")
		} else {
			w("fill-writer", "fill initial dequeue buffer")
		}
	case "count":
		if s.waiting >= 1 {
			if s.length == 0 {
				n.waiting--
				n.q = -1
				w("finish", "no work returns null")
			} else {
				w("notify-monitor", "last worker needs checkpoint monitor")
			}
		} else {
			w("sleep", "prepare queue wait")
		}
	case "notify-monitor":
		if s.mu == -1 {
			n.mu = 1
			w("notify-held", "acquire checkpoint retaining queue")
		}
	case "notify-held":
		n.notifyMain()
		n.mu = -1
		w("sleep", "notify checkpoint and release monitor")
	case "sleep":
		n.q = -1
		w("wait", "wait releases queue")
	case "wake":
		if s.q == -1 {
			n.q = 1
			w("awoke", "reacquire queue")
		}
	case "awoke":
		n.waiting--
		if s.finish {
			n.q = -1
			w("finish", "wake finished returns null")
		} else {
			w("condition", "decrement waiting and retry")
		}
	case "fill-writer":
		if s.writerLock == -1 {
			n.writerLock = 1
			w("fill-writer-held", "acquire writer for ensureWritten")
		}
	case "fill-writer-held":
		// No spill precedes this sole dequeue. Preserve the monitor boundary;
		// no pending writer file is possible before the first append.
		if s.pending {
			n.violation = "unexpected pre-seed spill"
		}
		n.writerLock = -1
		w("fill-reader", "ensureWritten returns and releases writer")
	case "fill-reader":
		if s.readerLock == -1 {
			n.readerLock = 1
			w("fill-reader-held", "acquire reader for getCache")
		}
	case "fill-reader-held":
		if s.readerFull || s.canRead {
			n.violation = "unexpected pre-seed reader cache"
		}
		n.readerLock = -1
		w("copy-seed", "getCache misses and releases reader")
	case "copy-seed":
		n.buffers[0] = s.buffers[s.enqBuffer]
		n.enqIndex = 0
		w("dequeue", "copy seed from enqueue buffer")
	case "dequeue":
		if s.buffers[0] == 0 {
			n.violation = "Null state found on queue"
		}
		n.length--
		n.q = -1
		w("enqueue", "dequeue seed and release queue")
	case "enqueue":
		if s.q == -1 {
			n.q = 1
			w("append-first", "acquire queue for two-state enqueue")
		}
	case "append-first":
		n.buffers[s.enqBuffer] = 2
		n.enqIndex = 1
		w("spill", "append first distinct state")
	case "spill":
		if s.writerLock == -1 {
			n.writerLock = 1
			w("spill-held", "acquire writer retaining queue")
		}
	case "spill-held":
		if s.pending || s.hiPool != 0 {
			n.violation = "unexpected second spill in original workload"
		}
		n.enqBuffer = s.writerBuffer
		n.writerBuffer = s.enqBuffer
		n.hiPool++
		n.enqIndex = 0
		n.pending = true
		n.notifyWriter()
		n.writerLock = -1
		w("append-second", "handoff full buffer, notify writer and release writer")
	case "append-second":
		n.enqIndex = 1
		n.buffers[s.enqBuffer] = 3
		w("enqueue-end", "append second distinct state")
	case "enqueue-end":
		n.length += 2
		n.q = -1
		w("finish", "publish enqueue count and release queue")
	case "finish":
		if s.q == -1 {
			n.q = 1
			w("finish-held", "acquire queue to finish")
		}
	case "finish-held":
		n.finish = true
		w("finish-monitor", "publish finish and notify queue")
	case "finish-monitor":
		if s.mu == -1 {
			n.mu = 1
			w("finish-notify", "acquire finish checkpoint monitor")
		}
	case "finish-notify":
		n.notifyMain()
		n.mu = -1
		n.q = -1
		w("finish-writer", "notify checkpoint, release both monitors")
	case "finish-writer":
		if s.writerLock == -1 {
			n.writerLock = 1
			w("finish-writer-held", "acquire writer to finish")
		}
	case "finish-writer-held":
		n.writerFinished = true
		n.notifyWriter()
		n.writerLock = -1
		w("finish-reader", "finish writer and notifyAll")
	case "finish-reader":
		if s.readerLock == -1 {
			n.readerLock = 1
			w("finish-reader-held", "acquire reader to finish")
		}
	case "finish-reader-held":
		n.readerFinished = true
		n.notifyReader()
		n.readerLock = -1
		w("done", "finish reader and passive cleaner")
	}
	n = s
	writer := func(phase, label string) { n.writer = phase; emit(n, "writer "+label) }
	switch s.writer {
	case "start", "wake":
		if s.writerLock == -1 {
			n.writerLock = 2
			writer("condition", "acquire writer monitor")
		}
	case "condition":
		if s.pending {
			writer("open", "pending file before finish check")
		} else if s.writerFinished {
			n.writerLock = -1
			writer("done", "finished without pending file")
		} else {
			n.writerLock = -1
			writer("wait", "wait releases writer")
		}
	case "open":
		n.openPoolWriter()
		writer("write", "PoolFiles.openForWriting")
	case "write":
		n.recordPoolValue()
		if n.outputCount == 3 {
			writer("commit", "record last signed header field")
		} else {
			writer("write", "record signed header field")
		}
	case "commit":
		n.commitPool()
		writer("clear", "ValueOutputStream.close commits exact prefix")
	case "clear":
		n.pending = false
		writer("reader", "clear pending file and notify writer waiters")
	case "reader":
		if s.readerLock == -1 {
			n.readerLock = 2
			writer("reader-held", "acquire reader retaining writer")
		}
	case "reader-held":
		n.canRead = true
		n.notifyReader()
		n.readerLock = -1
		writer("condition", "wakeup reader and release reader monitor")
	}
	n = s
	reader := func(phase, label string) { n.reader = phase; emit(n, "reader "+label) }
	switch s.reader {
	case "start", "wake":
		if s.readerLock == -1 {
			n.readerLock = 3
			if s.reader == "wake" {
				reader("awoke", "reacquire reader after wait")
			} else {
				reader("condition", "acquire reader monitor")
			}
		}
	case "awoke":
		if s.readerFinished {
			n.readerLock = -1
			reader("done", "finish checked only after wait")
		} else {
			reader("condition", "retry prefetch condition")
		}
	case "condition":
		if !s.readerFile || s.readerFull || !s.canRead {
			n.readerLock = -1
			reader("wait", "wait releases reader")
		} else {
			reader("open", "prefetch ready even if already finished")
		}
	case "open":
		n.openPoolReader()
		reader("read", "PoolFiles.openForReading clones committed values")
	case "read":
		n.readPoolValue()
		if n.inputCount == 3 {
			reader("close", "read final signed header field")
		} else {
			reader("read", "read signed header field")
		}
	case "close":
		n.closePoolReader()
		reader("publish", "ValueInputStream.close checks exact consumption")
	case "publish":
		n.buffers[3] = 4
		n.readerFile = false
		n.readerFull = true
		reader("condition", "publish decoded prefetch buffer")
	}
	n = s
	switch s.parent {
	case "worker":
		if s.worker == "done" {
			n.parent = "main"
			emit(n, "parent joins worker")
		}
	case "main":
		if s.main == "done" {
			n.parent = "writer"
			emit(n, "parent joins main")
		}
	case "writer":
		if s.writer == "done" {
			n.parent = "assert"
			emit(n, "parent joins actual modeled writer")
		}
	case "assert":
		if s.writer != "done" {
			n.violation = "original assertion: writer still alive"
		}
		n.parent = "done"
		emit(n, "parent asserts writer terminated; daemon reader need not join")
	}
	return out
}

func verifyJavaDiskQueueSchedules(t *testing.T) {
	t.Helper()
	initial := javaDiskQueueSchedule{length: 1, q: -1, mu: -1, writerLock: -1, readerLock: -1, main: "start", worker: "start", writer: "start", reader: "start", parent: "worker", readerFile: true, enqBuffer: 1, writerBuffer: 2, buffers: [4]uint8{0, 1, 0, 0}, enqIndex: 1}
	type predecessor struct {
		previous javaDiskQueueSchedule
		label    string
	}
	seen := map[javaDiskQueueSchedule]predecessor{initial: {}}
	pending := []javaDiskQueueSchedule{initial}
	transitions, terminal := 0, 0
	witness := func(current javaDiskQueueSchedule) string {
		var lines []string
		for current != initial {
			p := seen[current]
			lines = append(lines, p.label)
			current = p.previous
		}
		for l, r := 0, len(lines)-1; l < r; l, r = l+1, r-1 {
			lines[l], lines[r] = lines[r], lines[l]
		}
		return strings.Join(lines, "\n")
	}
	for head := 0; head < len(pending); head++ {
		current := pending[head]
		if current.violation != "" {
			t.Fatalf("original pool model violation: %s\n%s", current.violation, witness(current))
		}
		if current.parent == "done" {
			terminal++
			continue
		}
		next := current.successors()
		if len(next) == 0 {
			t.Fatalf("original disk queue workload deadlocks at %+v\n%s", current, witness(current))
		}
		for _, edge := range next {
			transitions++
			if _, ok := seen[edge.next]; !ok {
				seen[edge.next] = predecessor{current, edge.label}
				pending = append(pending, edge.next)
			}
		}
	}
	t.Logf("original disk queue model: %d states, %d transitions, %d completed parent states", len(seen), transitions, terminal)
}
