package tlc

import "testing"

func TestLiveCheckNilReceiversFailLikeJavaInvocation(t *testing.T) {
	var check *LiveCheck
	for name, call := range map[string]func(){
		"num_checker":          func() { _ = check.NumChecker() },
		"meta_dir":             func() { _ = check.GetMetaDir() },
		"tool":                 func() { _ = check.GetTool() },
		"checker":              func() { _ = check.GetChecker(0) },
		"add_init":             func() { _ = check.AddInitState(nil, nil, 0) },
		"add_next":             func() { _ = check.AddNextState(nil, nil, 0, nil) },
		"do_live_check":        func() { _ = check.DoLiveCheck() },
		"check":                func() { _, _ = check.Check(nil, false) },
		"final_check":          func() { _, _ = check.FinalCheck(nil) },
		"check_trace":          func() { _ = check.CheckTrace(nil, nil) },
		"close":                func() { _ = check.Close() },
		"begin_checkpoint":     func() { _ = check.BeginChkpt() },
		"commit_checkpoint":    func() { _ = check.CommitChkpt() },
		"flush":                func() { _ = check.FlushWritesToDiskFiles() },
		"out_degree_stats":     func() { _ = check.GetOutDegreeStatistics() },
		"calculate_in_degree":  func() { _, _ = check.CalculateInDegreeDiskGraphs(nil) },
		"calculate_out_degree": func() { _, _ = check.CalculateOutDegreeDiskGraphs(nil) },
		"recover":              func() { _ = check.Recover() },
		"reset":                func() { _ = check.Reset() },
	} {
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

func TestLiveCheckCheckTraceDereferencesSupplierAndFirstStateLikeJava(t *testing.T) {
	check := &LiveCheck{}
	for name, test := range map[string]struct {
		trace func() *StateVec
		want  any
	}{
		"null_supplier": {trace: nil, want: (*NullPointerException)(nil)},
		"null_trace":    {trace: func() *StateVec { return nil }, want: (*NullPointerException)(nil)},
		"empty_trace":   {trace: func() *StateVec { return NewStateVec(0) }, want: (*ArrayIndexOutOfBoundsException)(nil)},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				failure := recover()
				switch test.want.(type) {
				case *NullPointerException:
					if _, ok := failure.(*NullPointerException); !ok {
						t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
					}
				case *ArrayIndexOutOfBoundsException:
					if _, ok := failure.(*ArrayIndexOutOfBoundsException); !ok {
						t.Fatalf("failure = %T: %v, want ArrayIndexOutOfBoundsException", failure, failure)
					}
				}
			}()
			_ = check.CheckTrace(nil, test.trace)
		})
	}
}

func TestNoOpLiveCheckCheckTraceDoesNotInvokeSupplier(t *testing.T) {
	check := NewNoOpLiveCheck(nil, "meta")
	if err := check.CheckTrace(nil, nil); err != nil {
		t.Fatalf("no-op checkTrace = %v, want nil", err)
	}
}

func TestLiveCheckerNilReceiversFailLikeJavaInvocation(t *testing.T) {
	var checker *LiveChecker
	for name, call := range map[string]func(){
		"solution":   func() { _ = checker.GetSolution() },
		"disk_graph": func() { _ = checker.GetDiskGraph() },
		"add_init":   func() { _ = checker.AddInitState(nil, nil, 0) },
		"add_next":   func() { _ = checker.AddNextState(nil, nil, 0, nil, nil, nil) },
		"close":      func() { _ = checker.Close() },
	} {
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

func TestLiveCheckerAddNextStateRejectsNullSuccessorsLikeJava(t *testing.T) {
	checker := NewLiveChecker(&OrderOfSolution{}, 0, NewNoopLivenessStateWriter(), "", nil)
	defer func() {
		failure := recover()
		if _, ok := failure.(*NullPointerException); !ok {
			t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
		}
	}()
	_ = checker.AddNextState(nil, nil, 0, nil, nil, nil)
}
