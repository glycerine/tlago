package tlc

import (
	"strconv"
	"testing"
	"time"
)

// Upstream has no direct TLCServerMXWrapper tests. These native checks follow
// its synchronized control methods and notification/failure ordering.
type managementControlQueue struct {
	*MemStateQueue
	onControl func(string)
}

func (q *managementControlQueue) FinishAll()       { q.onControl("stop") }
func (q *managementControlQueue) SuspendAll() bool { q.onControl("suspend"); return false }
func (q *managementControlQueue) ResumeAll()       { q.onControl("resume") }

func TestDistributedManagementControlOrder(t *testing.T) {
	for _, operation := range []string{"stop", "suspend", "resume"} {
		for _, fails := range []bool{false, true} {
			t.Run(operation+"/failure="+strconv.FormatBool(fails), func(t *testing.T) {
				server := &TLCServer{}
				waiter := make(chan struct{})
				server.completionWaiter = waiter
				failure := NewRuntimeException("queue control failed")
				calls := 0
				server.StateQueue = &managementControlQueue{onControl: func(got string) {
					calls++
					if got != operation || server.monitor.owner.Load() != currentGoroutineID() {
						t.Error("queue control ran without owning the coordinator monitor")
					}
					if server.IsDone() != (operation == "stop") {
						t.Error("done flag changed at the wrong point")
					}
					select {
					case <-waiter:
						t.Error("coordinator notified before queue control completed")
					default:
					}
					if fails {
						panic(failure)
					}
				}}
				wrapper := &TLCServerMXWrapper{Server: server}
				var recovered any
				func() {
					defer func() { recovered = recover() }()
					switch operation {
					case "stop":
						wrapper.Stop()
					case "suspend":
						wrapper.Suspend()
					case "resume":
						wrapper.Resume()
					}
				}()
				if calls != 1 || (fails && recovered != failure) || (!fails && recovered != nil) {
					t.Fatalf("control calls/failure = %d/%v", calls, recovered)
				}
				if server.monitor.owner.Load() != 0 {
					t.Fatal("control retained the monitor after returning or failing")
				}
				select {
				case <-waiter:
					if operation != "stop" || fails || server.completionWaiter != nil {
						t.Fatal("unexpected completion notification")
					}
				default:
					if operation == "stop" && !fails {
						t.Fatal("successful stop did not notify the coordinator")
					}
				}
			})
		}
	}
}

func TestDistributedManagementStopWakesReportingWait(t *testing.T) {
	server := &TLCServer{StateQueue: NewMemStateQueue()}
	wrapper := &TLCServerMXWrapper{Server: server}
	started := make(chan struct{})
	finished := make(chan error, 1)
	joined := make(chan struct{})
	go func() {
		defer close(joined)
		server.monitor.Lock()
		defer server.monitor.Unlock()
		close(started)
		finished <- server.waitForReportLocked(0)
	}()
	defer func() { <-joined }()
	<-started
	// Acquiring the monitor proves the waiter has published its channel and
	// released its monitor. The source allows reentrant management callbacks.
	server.monitor.Lock()
	wrapper.Stop()
	server.monitor.Unlock()
	select {
	case err := <-finished:
		if err != nil || !server.IsDone() {
			t.Fatalf("stop completion = %v, done %v", err, server.IsDone())
		}
	case <-time.After(time.Second):
		// Release and join even when testing the broken implementation.
		server.monitor.Lock()
		server.notifyCompletionLocked()
		server.monitor.Unlock()
		<-finished
		t.Fatal("stop left the reporting wait asleep")
	}
}

func TestDistributedManagementControlMissingOwner(t *testing.T) {
	for _, operation := range []string{"stop", "suspend", "resume"} {
		for _, missing := range []string{"wrapper", "server", "queue"} {
			t.Run(operation+"/"+missing, func(t *testing.T) {
				server := &TLCServer{}
				waiter := make(chan struct{})
				server.completionWaiter = waiter
				wrapper := &TLCServerMXWrapper{Server: server}
				if missing == "wrapper" {
					wrapper = nil
				} else if missing == "server" {
					wrapper.Server = nil
				}
				requireCoordinatorNullFailure(t, func() error {
					switch operation {
					case "stop":
						wrapper.Stop()
					case "suspend":
						wrapper.Suspend()
					case "resume":
						wrapper.Resume()
					}
					return nil
				})
				if server.IsDone() != (operation == "stop" && missing == "queue") || server.monitor.owner.Load() != 0 {
					t.Fatal("missing owner changed done ordering or retained the monitor")
				}
				select {
				case <-waiter:
					t.Fatal("failed control notified completion")
				default:
				}
			})
		}
	}
}
