package tlc

import "testing"

func TestPoolHelpersRejectSecondStart(t *testing.T) {
	type thread struct {
		start  func()
		finish func()
		done   <-chan struct{}
	}

	stateReader := NewStatePoolReader(0, "")
	byteReader := NewByteArrayPoolReader(0, "")
	stateWriter := NewStatePoolWriter(0, nil)
	byteWriter := NewByteArrayPoolWriter(0, nil)
	stateCleaner := NewStatePoolCleaner(nil)
	byteCleaner := NewByteArrayPoolCleaner(nil)
	threads := map[string]thread{
		"state_reader":  {stateReader.Start, stateReader.SetFinished, stateReader.done},
		"byte_reader":   {byteReader.Start, byteReader.SetFinished, byteReader.done},
		"state_writer":  {stateWriter.Start, stateWriter.SetFinished, stateWriter.done},
		"byte_writer":   {byteWriter.Start, byteWriter.SetFinished, byteWriter.done},
		"state_cleaner": {stateCleaner.Start, stateCleaner.SetFinished, stateCleaner.done},
		"byte_cleaner":  {byteCleaner.Start, byteCleaner.SetFinished, byteCleaner.done},
	}

	for name, thread := range threads {
		t.Run(name, func(t *testing.T) {
			thread.start()
			defer func() {
				thread.finish()
				<-thread.done
			}()
			defer func() {
				failure := recover()
				if _, ok := failure.(*IllegalThreadStateException); !ok {
					t.Fatalf("failure = %T: %v, want IllegalThreadStateException", failure, failure)
				}
			}()
			thread.start()
		})
	}
}

func TestPoolHelpersRejectNullReceivers(t *testing.T) {
	var stateReader *StatePoolReader
	var byteReader *ByteArrayPoolReader
	var stateWriter *StatePoolWriter
	var byteWriter *ByteArrayPoolWriter
	var stateCleaner *StatePoolCleaner
	var byteCleaner *ByteArrayPoolCleaner

	calls := map[string]func(){
		"state_reader/start":         stateReader.Start,
		"state_reader/wakeup":        stateReader.Wakeup,
		"state_reader/restart":       func() { stateReader.Restart("", false) },
		"state_reader/do_work":       func() { _, _ = stateReader.DoWork(nil, "") },
		"state_reader/get_cache":     func() { _, _ = stateReader.GetCache(nil, "") },
		"state_reader/set_finished":  stateReader.SetFinished,
		"byte_reader/start":          byteReader.Start,
		"byte_reader/wakeup":         byteReader.Wakeup,
		"byte_reader/restart":        func() { byteReader.Restart("", false) },
		"byte_reader/do_work":        func() { _, _ = byteReader.DoWork(nil, "") },
		"byte_reader/get_cache":      func() { _, _ = byteReader.GetCache(nil, "") },
		"byte_reader/set_finished":   byteReader.SetFinished,
		"state_writer/start":         stateWriter.Start,
		"state_writer/do_work":       func() { _, _ = stateWriter.DoWork(nil, "") },
		"state_writer/ensure":        func() { _ = stateWriter.EnsureWritten() },
		"state_writer/set_finished":  stateWriter.SetFinished,
		"byte_writer/start":          byteWriter.Start,
		"byte_writer/do_work":        func() { _, _ = byteWriter.DoWork(nil, "") },
		"byte_writer/ensure":         func() { _ = byteWriter.EnsureWritten() },
		"byte_writer/set_finished":   byteWriter.SetFinished,
		"state_cleaner/start":        stateCleaner.Start,
		"state_cleaner/delete_up_to": func() { stateCleaner.DeleteUpTo(1) },
		"state_cleaner/set_finished": stateCleaner.SetFinished,
		"state_cleaner/finish_wait":  stateCleaner.FinishAndWait,
		"byte_cleaner/start":         byteCleaner.Start,
		"byte_cleaner/delete_up_to":  func() { byteCleaner.DeleteUpTo(1) },
		"byte_cleaner/set_finished":  byteCleaner.SetFinished,
		"byte_cleaner/finish_wait":   byteCleaner.FinishAndWait,
	}

	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
				}
			}()
			call()
		})
	}
}
