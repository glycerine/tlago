package tlc

import "testing"

func TestLiveCheck1NilReceiversFailLikeJavaInvocation(t *testing.T) {
	var check *LiveCheck1
	for name, call := range map[string]func(){
		"init":                 func() { _ = check.Init(nil, nil, "") },
		"reset":                func() { _ = check.Reset() },
		"add_init":             func() { _ = check.AddInitState(nil, nil, 0) },
		"add_next":             func() { _ = check.AddNextState(nil, nil, 0, nil) },
		"add_next_pair":        func() { _ = check.AddNextStatePair(nil, nil, 0, nil, 0) },
		"set_done":             func() { check.SetDone(0) },
		"do_live_check":        func() { _ = check.DoLiveCheck() },
		"check":                func() { _, _ = check.Check(nil, false) },
		"check_trace":          func() { _ = check.CheckTrace(nil, nil) },
		"final_check":          func() { _, _ = check.FinalCheck(nil) },
		"meta_dir":             func() { _ = check.GetMetaDir() },
		"tool":                 func() { _ = check.GetTool() },
		"out_degree_stats":     func() { _ = check.GetOutDegreeStatistics() },
		"checker":              func() { _ = check.GetChecker(0) },
		"num_checker":          func() { _ = check.NumChecker() },
		"close":                func() { _ = check.Close() },
		"begin_checkpoint":     func() { _ = check.BeginChkpt() },
		"commit_checkpoint":    func() { _ = check.CommitChkpt() },
		"recover":              func() { _ = check.Recover() },
		"flush":                func() { _ = check.FlushWritesToDiskFiles() },
		"calculate_in_degree":  func() { _, _ = check.CalculateInDegreeDiskGraphs(nil) },
		"calculate_out_degree": func() { _, _ = check.CalculateOutDegreeDiskGraphs(nil) },
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

func TestLiveCheck1ExplicitConstructorRetainsNullErrorFlag(t *testing.T) {
	check, err := NewLiveCheck1WithError(&Tool{}, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if check.ErrorFound != nil {
		t.Fatalf("explicit constructor error flag = %p, want nil", check.ErrorFound)
	}
	ordinary, err := NewLiveCheck1(&Tool{})
	if err != nil {
		t.Fatal(err)
	}
	if ordinary.ErrorFound == nil {
		t.Fatal("ordinary constructor error flag = nil, want allocated AtomicBoolean")
	}
	assertLiveCheck1PanicType(t, "null_error_flag_use", func() { _ = check.claimError() }, (*NullPointerException)(nil))
	if !ordinary.claimError() || ordinary.claimError() {
		t.Fatal("ordinary constructor error flag did not preserve compare-and-set ownership")
	}
}

func TestLiveCheck1CheckTraceDereferencesLikeJava(t *testing.T) {
	check := &LiveCheck1{}
	assertLiveCheck1PanicType(t, "null_supplier", func() { _ = check.CheckTrace(nil, nil) }, (*NullPointerException)(nil))
	if err := check.CheckTrace(nil, func() *StateVec { return nil }); err != nil {
		t.Fatalf("null trace without solutions = %v, want nil", err)
	}
	if err := check.CheckTrace(nil, func() *StateVec { return NewStateVec(0) }); err != nil {
		t.Fatalf("empty trace without solutions = %v, want nil", err)
	}

	check.Solutions = []*OrderOfSolution{{}}
	assertLiveCheck1PanicType(t, "null_trace_with_solution", func() {
		_ = check.CheckTrace(nil, func() *StateVec { return nil })
	}, (*NullPointerException)(nil))
	assertLiveCheck1PanicType(t, "empty_trace_with_solution", func() {
		_ = check.CheckTrace(nil, func() *StateVec { return NewStateVec(0) })
	}, (*ArrayIndexOutOfBoundsException)(nil))
}

func TestLiveCheck1RequiresParallelBehaviorGraphsLikeJava(t *testing.T) {
	missing := func() *LiveCheck1 {
		return &LiveCheck1{Solutions: []*OrderOfSolution{{}}}
	}
	for name, call := range map[string]func(){
		"add_init": func() { _ = missing().AddInitState(nil, nil, 0) },
		"add_next_pair": func() {
			_ = missing().AddNextStatePair(nil, nil, 0, nil, 0)
		},
		"set_done": func() { missing().SetDone(0) },
		"check":    func() { _, _ = missing().Check(nil, false) },
	} {
		assertLiveCheck1PanicType(t, "missing_graph/"+name, call, (*ArrayIndexOutOfBoundsException)(nil))
	}

	nullGraph := func() *LiveCheck1 {
		return &LiveCheck1{Solutions: []*OrderOfSolution{{}}, BGraphs: []*BEGraph{nil}}
	}
	for name, call := range map[string]func(){
		"reset":         func() { _ = (&LiveCheck1{BGraphs: []*BEGraph{nil}}).Reset() },
		"add_init":      func() { _ = nullGraph().AddInitState(nil, nil, 0) },
		"add_next_pair": func() { _ = nullGraph().AddNextStatePair(nil, nil, 0, nil, 0) },
		"set_done":      func() { nullGraph().SetDone(0) },
		"check":         func() { _, _ = nullGraph().Check(nil, false) },
	} {
		assertLiveCheck1PanicType(t, "null_graph/"+name, call, (*NullPointerException)(nil))
	}
}

func TestLiveCheck1CheckDereferencesStoredNullInitialNode(t *testing.T) {
	graph := NewBEGraph("", false)
	graph.AddInitNode(nil)
	check := &LiveCheck1{
		Solutions: []*OrderOfSolution{{}},
		BGraphs:   []*BEGraph{graph},
	}
	assertLiveCheck1PanicType(t, "null_initial_node", func() { _, _ = check.Check(nil, false) }, (*NullPointerException)(nil))
}

func TestLiveCheck1TableauTraversalDereferencesMissingNodeLikeJava(t *testing.T) {
	tableau := NewEmptyTBGraph()
	tableau.AddElement(NewTBGraphNode(NewTBPar(0)))
	tableau.GetNode(0).SetIndex(1)
	tableau.SetInitCnt(1)
	solution := &OrderOfSolution{Tableau: tableau}
	state := &TLCStateMut{}
	check := &LiveCheck1{StateTrace: NewStateVecFrom([]*TLCStateMut{state})}

	assertLiveCheck1PanicType(t, "construct", func() {
		_, _ = check.ConstructBEGraph(nil, solution)
	}, (*NullPointerException)(nil))

	node := NewBTGraphNode(0, 1).BEGraphNode
	graph := NewBEGraph("", true)
	assertLiveCheck1PanicType(t, "stuttering", func() {
		_ = check.addNodesForStut(state, 0, node, nil, nil, solution, graph)
	}, (*NullPointerException)(nil))
	assertLiveCheck1PanicType(t, "done_state", func() {
		_ = check.addNextStateDone(nil, state, 0, node, solution, graph)
	}, (*NullPointerException)(nil))

	promiseSolution := &OrderOfSolution{Tableau: tableau, Promises: []*LiveExprNode{nil}}
	newPromiseCheck := func() *LiveCheck1 {
		return &LiveCheck1{
			MetaDir:    t.TempDir(),
			CurrentOOS: promiseSolution,
			CurrentPEM: &PossibleErrorModel{},
			ThirdNum:   liveCheck1MaxSecond + 1,
		}
	}
	assertLiveCheck1PanicType(t, "error_trace_promise", func() {
		_, _ = newPromiseCheck().printErrorTrace(NewBTGraphNode(0, 1).BEGraphNode)
	}, (*NullPointerException)(nil))
	assertLiveCheck1PanicType(t, "subcomponent_promise", func() {
		_ = newPromiseCheck().checkSubcomponent(NewBTGraphNode(0, 1).BEGraphNode)
	}, (*NullPointerException)(nil))
}

func TestLiveCheck1CycleBacktrackDereferencesEmptyStackLikeJava(t *testing.T) {
	node := NewBEGraphNode(0)
	node.SetCheckState([]bool{false})
	check := &LiveCheck1{
		MetaDir:    t.TempDir(),
		CurrentOOS: &OrderOfSolution{},
		CurrentPEM: &PossibleErrorModel{AEState: []int{0}},
		ThirdNum:   liveCheck1MaxSecond + 1,
	}
	assertLiveCheck1PanicType(t, "empty_backtrack", func() {
		_, _ = check.printErrorTrace(node)
	}, (*NullPointerException)(nil))
}

func assertLiveCheck1PanicType(t *testing.T, name string, call func(), want any) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		defer func() {
			failure := recover()
			switch want.(type) {
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
		call()
	})
}
