package tlc

import (
	"fmt"
	"testing"
	"time"
)

type nestedInitializationChild struct {
	*MemFPSet
	entered chan string
	release <-chan struct{}
	failure error
}

func (s *nestedInitializationChild) Init(threads int, directory, filename string) FPSet {
	s.entered <- fmt.Sprintf("%d:%s:%s", threads, directory, filename)
	<-s.release
	if s.failure != nil {
		panic(s.failure)
	}
	// Source ignores the returned replacement, retaining the original child.
	return NewMemFPSet()
}

// No enabled original method tests this boundary. Native goroutines join child
// initialization and retain errors without Java ForkJoin exception copying.
func TestDistributedNestedInitializationFailureOwnership(t *testing.T) {
	for _, failedChild := range []int{0, 1} {
		for _, family := range []string{"io", "unchecked", "fatal"} {
			t.Run(fmt.Sprintf("child%d/%s", failedChild, family), func(t *testing.T) {
				var failure error
				switch family {
				case "io":
					failure = NewIOException("initialization storage failure")
				case "unchecked":
					failure = NewRuntimeException("unchecked initialization failure")
				case "fatal":
					failure = NewAssertionError("fatal initialization failure")
				}
				entered := make(chan string, 2)
				held := make(chan struct{})
				defer close(held)
				ready := make(chan struct{})
				close(ready)
				children := []*nestedInitializationChild{
					{MemFPSet: NewMemFPSet(), entered: entered, release: held},
					{MemFPSet: NewMemFPSet(), entered: entered, release: held},
				}
				children[failedChild].release = ready
				children[failedChild].failure = failure
				set := &MultiFPSet{Sets: []FPSet{children[0], children[1]}}
				done := make(chan any, 1)
				go func() {
					defer func() { done <- recover() }()
					set.Init(3, "metadata", "job")
				}()
				seen := map[string]bool{}
				for range 2 {
					select {
					case name := <-entered:
						seen[name] = true
					case <-time.After(5 * time.Second):
						t.Fatal("child initialization did not run independently")
					}
				}
				if !seen["3:metadata:job_0"] || !seen["3:metadata:job_1"] {
					t.Fatalf("initialization arguments = %v", seen)
				}
				select {
				case value := <-done:
					t.Fatalf("initialization returned before held child joined: %v", value)
				default:
				}
				held <- struct{}{}
				select {
				case value := <-done:
					if family == "io" {
						wrapped, ok := value.(*RuntimeException)
						if !ok || wrapped.Cause != failure || isJavaIOException(wrapped) {
							t.Fatalf("checked I/O lost its single operation boundary: %T/%v", value, value)
						}
					} else if value != failure {
						t.Fatalf("native initialization copied its original failure: %T/%v", value, value)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("initialization did not join after releasing its child")
				}
				if set.Sets[0] != children[0] || set.Sets[1] != children[1] {
					t.Fatal("initialization published a returned replacement child")
				}
			})
		}
	}
}
