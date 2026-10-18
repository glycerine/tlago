// Complete translation of tlc2.tool.queue.DiskPoolWriterTest. Java observes
// Thread.State and joins its writer thread. Go observes the actual writer
// goroutine's runtime state and waits for its termination with the same timeout;
// notification flags or finished flags cannot substitute for those assertions.
package tlc

import (
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

const javaPoolWriterTimeout = 5000 * time.Millisecond

// Runtime IDs identify the actual native goroutine even before its startup
// wrapper enters run. Snapshotting before construction excludes other queues;
// each original test constructs exactly one new writer of the requested type.
type javaPoolWriterThread struct {
	id   string
	mu   *sync.Mutex
	cond *sync.Cond
}

func javaPoolWriterStacks() []string {
	buffer := make([]byte, 64*1024)
	for {
		n := runtime.Stack(buffer, true)
		if n < len(buffer) {
			return strings.Split(string(buffer[:n]), "\n\n")
		}
		buffer = make([]byte, len(buffer)*2)
	}
}
func javaPoolWriterGoroutines() map[string]bool {
	ids := make(map[string]bool)
	for _, stack := range javaPoolWriterStacks() {
		fields := strings.Fields(strings.SplitN(stack, "\n", 2)[0])
		if len(fields) >= 2 {
			ids[fields[1]] = true
		}
	}
	return ids
}
func javaNewPoolWriterThread(t *testing.T, before map[string]bool, kind string, mu *sync.Mutex, cond *sync.Cond) javaPoolWriterThread {
	t.Helper()
	id := ""
	for _, stack := range javaPoolWriterStacks() {
		fields := strings.Fields(strings.SplitN(stack, "\n", 2)[0])
		if len(fields) < 2 || before[fields[1]] || !strings.Contains(stack, "(*"+kind+").Start") {
			continue
		}
		if id != "" {
			t.Fatal("multiple new writer goroutines: cannot identify the source thread")
		}
		id = fields[1]
	}
	if id == "" {
		t.Fatal("queue constructor did not start its writer goroutine")
	}
	return javaPoolWriterThread{id: id, mu: mu, cond: cond}
}
func (thread javaPoolWriterThread) state() (state string, alive bool) {
	for _, stack := range javaPoolWriterStacks() {
		if !strings.HasPrefix(stack, "goroutine "+thread.id+" [") {
			continue
		}
		line := strings.SplitN(stack, "\n", 2)[0]
		start, end := strings.IndexByte(line, '['), strings.LastIndexByte(line, ']')
		if start < 0 || end < start {
			panic("writer goroutine has no runtime state: " + line)
		}
		return line[start+1 : end], true
	}
	return "terminated", false
}

func javaAwaitPoolWriterWaiting(t *testing.T, thread javaPoolWriterThread) {
	t.Helper()
	deadline := time.Now().Add(javaPoolWriterTimeout)
	state, alive := thread.state()
	for alive && state != "sync.Cond.Wait" && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		state, alive = thread.state()
	}
	if !alive {
		buffer := make([]byte, 64*1024)
		n := runtime.Stack(buffer, true)
		t.Fatalf("writer exited while waiting for state WAITING; wanted %s; goroutines:\n%s", thread.id, buffer[:n])
	}
	if state != "sync.Cond.Wait" {
		t.Fatalf("writer did not enter state WAITING: %s", state)
	}
}

func javaAssertPoolWriterIgnoresEmptyWake(t *testing.T, thread javaPoolWriterThread) {
	t.Helper()
	javaAwaitPoolWriterWaiting(t, thread)
	func() {
		thread.mu.Lock()
		defer thread.mu.Unlock()
		thread.cond.Broadcast()
		deadline := time.Now().Add(javaPoolWriterTimeout)
		state, _ := thread.state()
		for state == "sync.Cond.Wait" && time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond)
			state, _ = thread.state()
		}
		if state == "sync.Cond.Wait" {
			t.Fatal("writer did not observe the notification")
		}
	}()
	javaAwaitPoolWriterWaiting(t, thread)
	if _, alive := thread.state(); !alive {
		t.Fatal("writer exited after an empty wake")
	}
}

func javaAssertPoolWriterStopsOnFinish(t *testing.T, thread javaPoolWriterThread, finish func()) {
	t.Helper()
	finish()
	deadline := time.Now().Add(javaPoolWriterTimeout)
	_, alive := thread.state()
	for alive && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		_, alive = thread.state()
	}
	if alive {
		t.Fatal("writer did not stop when the queue finished")
	}
}

func TestJavaDiskPoolWriter(t *testing.T) {
	t.Run("testStatePoolWriterIgnoresEmptyWakeAndStopsOnFinish", func(t *testing.T) {
		before := javaPoolWriterGoroutines()
		queue := NewDiskStateQueue(t.TempDir())
		t.Cleanup(queue.FinishAll)
		thread := javaNewPoolWriterThread(t, before, "StatePoolWriter", &queue.writer.mu, queue.writer.cond)
		javaAssertPoolWriterIgnoresEmptyWake(t, thread)
		javaAssertPoolWriterStopsOnFinish(t, thread, queue.FinishAll)
	})
	t.Run("testByteArrayPoolWriterIgnoresEmptyWakeAndStopsOnFinish", func(t *testing.T) {
		before := javaPoolWriterGoroutines()
		queue := NewDiskByteArrayQueue(t.TempDir())
		t.Cleanup(queue.FinishAll)
		thread := javaNewPoolWriterThread(t, before, "ByteArrayPoolWriter", &queue.writer.mu, queue.writer.cond)
		javaAssertPoolWriterIgnoresEmptyWake(t, thread)
		javaAssertPoolWriterStopsOnFinish(t, thread, queue.FinishAll)
	})
}
