package tlc

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Native host ownership has no direct Java test. A second close must join
// callback cleanup even after accepted replies have finished, and retain its
// failure. Forced close must still be able to interrupt a graceful reply drain.
func TestDistributedRPCCloseJoinsCallbackCleanup(t *testing.T) {
	for _, firstGraceful := range []bool{false, true} {
		for _, secondGraceful := range []bool{false, true} {
			name := map[bool]string{false: "forced", true: "graceful"}
			t.Run(name[firstGraceful]+"/"+name[secondGraceful], func(t *testing.T) {
				server := NewDistributedRPCServer()
				entered, release := make(chan struct{}), make(chan struct{})
				var calls atomic.Int32
				var releaseOnce sync.Once
				var finished []<-chan struct{}
				failure := errors.New("callback cleanup failed")
				if err := server.outbound.add(distributedConnectionCloser(func() error {
					calls.Add(1)
					close(entered)
					<-release
					return failure
				})); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					releaseOnce.Do(func() { close(release) })
					for _, done := range finished {
						select {
						case <-done:
						case <-time.After(2 * time.Second):
							t.Error("host close did not finish after releasing callback")
						}
					}
				})
				startClose := func(graceful bool) <-chan error {
					result := make(chan error, 1)
					done := make(chan struct{})
					finished = append(finished, done)
					go func() {
						defer close(done)
						if graceful {
							result <- server.CloseGracefully()
						} else {
							result <- server.Close()
						}
					}()
					return result
				}
				first := startClose(firstGraceful)
				select {
				case <-entered:
				case <-time.After(2 * time.Second):
					t.Fatal("host did not begin callback cleanup")
				}
				second := startClose(secondGraceful)
				select {
				case err := <-second:
					t.Fatalf("concurrent close returned before callback cleanup: %v", err)
				case <-time.After(50 * time.Millisecond):
				}
				releaseOnce.Do(func() { close(release) })
				for _, result := range []<-chan error{first, second} {
					select {
					case err := <-result:
						if !errors.Is(err, failure) {
							t.Fatalf("close lost callback cleanup failure: %v", err)
						}
					case <-time.After(2 * time.Second):
						t.Fatal("close did not join callback cleanup")
					}
				}
				if err := server.CloseGracefully(); !errors.Is(err, failure) || calls.Load() != 1 {
					t.Fatalf("repeated close changed cleanup result or repeated it: %v, calls %d", err, calls.Load())
				}
			})
		}
	}
}
