package tlago

import (
	"math"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Whole original tlc2.tool.simulation.SimulationWorkerTest methods. Its isolated
// classloader resets statics before each method; setUp selects BasicMultiTrace.
func javaSimulationWorkerTool(t *testing.T, config string) *tlc.Tool {
	t.Helper()
	oldUserDir := tlc.GetFilenameUserDirectory()
	oldMain, oldSimulator, oldTool := tlc.Globals.MainChecker, tlc.Globals.Simulator, tlc.Globals.Tool
	oldCoverage, oldWarn := tlc.Globals.CoverageInterval, tlc.Globals.Warn
	oldView, oldPoly := tlc.UseView(), tlc.FP64IrredPoly()
	tlc.UniqueStringInitialize()
	tlc.InitializeIntValueStatics()
	tlc.InitializeActionItemListStatics()
	tlc.InitializeFPIntSetStatics()
	tlc.SetMainChecker(nil)
	tlc.SetSimulator(nil)
	tlc.Globals.CoverageInterval = -1
	tlc.Globals.Warn = true
	tlc.SetUseView(false)
	tlc.FP64Init()
	recorder := &tlc.MemoryRecorder{}
	tlc.AddMessageRecorder(recorder)
	t.Cleanup(func() {
		tlc.RemoveMessageRecorder(recorder)
		tlc.SetMainChecker(oldMain)
		tlc.SetSimulator(oldSimulator)
		tlc.Globals.Tool = oldTool
		tlc.Globals.CoverageInterval, tlc.Globals.Warn = oldCoverage, oldWarn
		tlc.SetUseView(oldView)
		tlc.FP64InitPoly(oldPoly)
		tlc.SetFilenameUserDirectory(oldUserDir)
		tlc.SetTLCStateTool(nil)
	})
	directory, err := filepath.Abs(filepath.Join("tlc", "test_vectors", "models", "BasicMultiTrace"))
	if err != nil {
		t.Fatal(err)
	}
	tlc.SetFilenameUserDirectory(&directory)
	tool, diags, err := loadTLCAppTool("BasicMultiTrace", config, nil, tlc.RuntimeParameters{})
	requireNoErrors(t, diags)
	if err != nil {
		t.Fatal(err)
	}
	// Mode is installed before any initial/next-state evaluation, as in FastTool.
	tool.SetMode(tlc.ModeSimulation)
	tlc.SetTLCStateTool(tool)
	return tool
}

func javaSimulationWorkerStateValue(state *tlc.TLCStateMut, name string) tlc.Value {
	value, _ := state.GetVals().Get2(tlc.UniqueStringOf(name))
	return value
}

func javaSimulationWorkerStateVal(state *tlc.TLCStateMut, name string) string {
	return javaSimulationWorkerStateValue(state, name).String()
}

func javaSimulationWorkerTraceLevels(trace *tlc.StateVec) bool {
	for i := 0; i < trace.Size(); i++ {
		if trace.At(i).Level() != i+1 {
			return false
		}
	}
	return true
}

func TestJavaSimulationWorkerSuccessfulRun(t *testing.T) {
	tool := javaSimulationWorkerTool(t, "MC")
	liveCheck := tlc.NewNoOpLiveCheck(tool, "BasicMultiTrace")
	initStates := javaGeneratedInitialStates(t, tool)
	resultQueue := tlc.NewSimulationWorkerResultQueue()
	worker := tlc.NewSimulationWorker(0, tool, resultQueue, 0, 100, 1000, "", false, false, "", liveCheck, nil, nil, nil)
	defer func() { worker.Stop(); worker.Join(0) }()
	worker.Start(initStates)
	res := resultQueue.Take()
	if res.IsError() {
		t.Fatal("original assertFalse: res.IsError()")
	}
	worker.Join(0)
	if worker.IsAlive() {
		t.Fatal("original assertFalse: worker.IsAlive()")
	}
}

func TestJavaSimulationWorkerInvariantViolation(t *testing.T) {
	tool := javaSimulationWorkerTool(t, "MCInv")
	liveCheck := tlc.NewNoOpLiveCheck(tool, "BasicMultiTrace")
	initStates := javaGeneratedInitialStates(t, tool)
	resultQueue := tlc.NewSimulationWorkerResultQueue()
	maxTraceNum := int64(3)
	worker := tlc.NewSimulationWorker(0, tool, resultQueue, 0, 100, maxTraceNum, "", false, false, "", liveCheck, nil, nil, nil)
	defer func() { worker.Stop(); worker.Join(0) }()
	worker.Start(initStates)
	res := resultQueue.Take()
	if !(res.IsError()) {
		t.Fatal("original assertTrue: res.IsError()")
	}
	failure := res.Error
	if got := failure.Code; got != tlc.ECTLCInvariantViolatedBehavior {
		t.Fatalf("original assertEquals: got %v, want %v", got, tlc.ECTLCInvariantViolatedBehavior)
	}
	if got := failure.StateTrace.Size(); got != 4 {
		t.Fatalf("original assertEquals: got %v, want %v", got, 4)
	}
	if !(failure.StateTrace.At(0).IsInitial()) {
		t.Fatal("original assertTrue: failure.StateTrace.At(0).IsInitial()")
	}
	if !(javaSimulationWorkerTraceLevels(failure.StateTrace)) {
		t.Fatal("original assertTrue: javaSimulationWorkerTraceLevels(failure.StateTrace)")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "branch"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(1), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(1), "branch"); got != "6" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "6")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(2), "depth"); got != "1" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "1")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(2), "branch"); got != "6" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "6")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.Last(), "depth"); got != "2" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "2")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.Last(), "branch"); got != "6" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "6")
	}
	res = resultQueue.Take()
	if !(res.IsError()) {
		t.Fatal("original assertTrue: res.IsError()")
	}
	failure = res.Error
	if got := failure.Code; got != tlc.ECTLCInvariantViolatedBehavior {
		t.Fatalf("original assertEquals: got %v, want %v", got, tlc.ECTLCInvariantViolatedBehavior)
	}
	if got := failure.StateTrace.Size(); got != 4 {
		t.Fatalf("original assertEquals: got %v, want %v", got, 4)
	}
	if !(failure.StateTrace.At(0).IsInitial()) {
		t.Fatal("original assertTrue: failure.StateTrace.At(0).IsInitial()")
	}
	if !(javaSimulationWorkerTraceLevels(failure.StateTrace)) {
		t.Fatal("original assertTrue: javaSimulationWorkerTraceLevels(failure.StateTrace)")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "branch"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(1), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(1), "branch"); got != "2" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "2")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(2), "depth"); got != "1" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "1")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(2), "branch"); got != "2" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "2")
	}
	res = resultQueue.Take()
	if !(res.IsError()) {
		t.Fatal("original assertTrue: res.IsError()")
	}
	failure = res.Error
	if got := failure.Code; got != tlc.ECTLCInvariantViolatedBehavior {
		t.Fatalf("original assertEquals: got %v, want %v", got, tlc.ECTLCInvariantViolatedBehavior)
	}
	if got := failure.StateTrace.Size(); got != 4 {
		t.Fatalf("original assertEquals: got %v, want %v", got, 4)
	}
	if !(failure.StateTrace.At(0).IsInitial()) {
		t.Fatal("original assertTrue: failure.StateTrace.At(0).IsInitial()")
	}
	if !(javaSimulationWorkerTraceLevels(failure.StateTrace)) {
		t.Fatal("original assertTrue: javaSimulationWorkerTraceLevels(failure.StateTrace)")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "branch"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(1), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(1), "branch"); got != "5" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "5")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(2), "depth"); got != "1" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "1")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(2), "branch"); got != "5" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "5")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(3), "depth"); got != "2" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "2")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(3), "branch"); got != "5" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "5")
	}
	res = resultQueue.Take()
	if res.IsError() {
		t.Fatal("original assertFalse: res.IsError()")
	}
	worker.Join(0)
	if worker.IsAlive() {
		t.Fatal("original assertFalse: worker.IsAlive()")
	}
}

func TestJavaSimulationWorkerActionPropertyViolation(t *testing.T) {
	tool := javaSimulationWorkerTool(t, "MCActionProp")
	initStates := javaGeneratedInitialStates(t, tool)
	liveCheck := tlc.NewNoOpLiveCheck(tool, "BasicMultiTrace")
	resultQueue := tlc.NewSimulationWorkerResultQueue()
	worker := tlc.NewSimulationWorker(0, tool, resultQueue, 0, 100, 100, "", false, false, "", liveCheck, nil, nil, nil)
	defer func() { worker.Stop(); worker.Join(0) }()
	worker.Start(initStates)
	res := resultQueue.Take()
	if !(res.IsError()) {
		t.Fatal("original assertTrue: res.IsError()")
	}
	failure := res.Error
	if got := failure.Code; got != tlc.ECTLCActionPropertyViolatedBehavior {
		t.Fatalf("original assertEquals: got %v, want %v", got, tlc.ECTLCActionPropertyViolatedBehavior)
	}
	if got := failure.StateTrace.Size(); got != 3 {
		t.Fatalf("original assertEquals: got %v, want %v", got, 3)
	}
	if !(failure.StateTrace.At(0).IsInitial()) {
		t.Fatal("original assertTrue: failure.StateTrace.At(0).IsInitial()")
	}
	if !(javaSimulationWorkerTraceLevels(failure.StateTrace)) {
		t.Fatal("original assertTrue: javaSimulationWorkerTraceLevels(failure.StateTrace)")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "branch"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(1), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(1), "branch"); got != "6" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "6")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.Last(), "depth"); got != "1" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "1")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.Last(), "branch"); got != "6" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "6")
	}
	res = resultQueue.Take()
	if !(res.IsError()) {
		t.Fatal("original assertTrue: res.IsError()")
	}
	failure = res.Error
	if got := failure.Code; got != tlc.ECTLCActionPropertyViolatedBehavior {
		t.Fatalf("original assertEquals: got %v, want %v", got, tlc.ECTLCActionPropertyViolatedBehavior)
	}
	if got := failure.StateTrace.Size(); got != 3 {
		t.Fatalf("original assertEquals: got %v, want %v", got, 3)
	}
	if !(failure.StateTrace.At(0).IsInitial()) {
		t.Fatal("original assertTrue: failure.StateTrace.At(0).IsInitial()")
	}
	if !(javaSimulationWorkerTraceLevels(failure.StateTrace)) {
		t.Fatal("original assertTrue: javaSimulationWorkerTraceLevels(failure.StateTrace)")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "branch"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(1), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(1), "branch"); got != "10" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "10")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.Last(), "depth"); got != "1" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "1")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.Last(), "branch"); got != "10" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "10")
	}
	worker.Join(0)
	if worker.IsAlive() {
		t.Fatal("original assertFalse: worker.IsAlive()")
	}
}

func TestJavaSimulationWorkerInvariantBadEval(t *testing.T) {
	tool := javaSimulationWorkerTool(t, "MCBadInvNonInitState")
	initStates := javaGeneratedInitialStates(t, tool)
	liveCheck := tlc.NewNoOpLiveCheck(tool, "BasicMultiTrace")
	resultQueue := tlc.NewSimulationWorkerResultQueue()
	worker := tlc.NewSimulationWorker(0, tool, resultQueue, 0, 100, 100, "", false, false, "", liveCheck, nil, nil, nil)
	defer func() { worker.Stop(); worker.Join(0) }()
	worker.Start(initStates)
	res := resultQueue.Take()
	if !(res.IsError()) {
		t.Fatal("original assertTrue: res.IsError()")
	}
	failure := res.Error
	if got := failure.Code; got != tlc.ECTLCInvariantEvaluationFailed {
		t.Fatalf("original assertEquals: got %v, want %v", got, tlc.ECTLCInvariantEvaluationFailed)
	}
	if got := failure.StateTrace.Size(); got != 2 {
		t.Fatalf("original assertEquals: got %v, want %v", got, 2)
	}
	if !(failure.StateTrace.At(0).IsInitial()) {
		t.Fatal("original assertTrue: failure.StateTrace.At(0).IsInitial()")
	}
	if !(javaSimulationWorkerTraceLevels(failure.StateTrace)) {
		t.Fatal("original assertTrue: javaSimulationWorkerTraceLevels(failure.StateTrace)")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "branch"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.Last(), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.Last(), "branch"); got != "1" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "1")
	}
	worker.Join(0)
	if worker.IsAlive() {
		t.Fatal("original assertFalse: worker.IsAlive()")
	}
}

func TestJavaSimulationWorkerActionPropertyBadEval(t *testing.T) {
	tool := javaSimulationWorkerTool(t, "MCActionPropBadEval")
	initStates := javaGeneratedInitialStates(t, tool)
	liveCheck := tlc.NewNoOpLiveCheck(tool, "BasicMultiTrace")
	resultQueue := tlc.NewSimulationWorkerResultQueue()
	worker := tlc.NewSimulationWorker(0, tool, resultQueue, 0, 100, 100, "", false, false, "", liveCheck, nil, nil, nil)
	defer func() { worker.Stop(); worker.Join(0) }()
	worker.Start(initStates)
	res := resultQueue.Take()
	if !(res.IsError()) {
		t.Fatal("original assertTrue: res.IsError()")
	}
	failure := res.Error
	if got := failure.Code; got != tlc.ECTLCActionPropertyEvaluationFailed {
		t.Fatalf("original assertEquals: got %v, want %v", got, tlc.ECTLCActionPropertyEvaluationFailed)
	}
	worker.Join(0)
	if worker.IsAlive() {
		t.Fatal("original assertFalse: worker.IsAlive()")
	}
}

func TestJavaSimulationWorkerUnderspecifiedNext(t *testing.T) {
	tool := javaSimulationWorkerTool(t, "MCUnderspecNext")
	initStates := javaGeneratedInitialStates(t, tool)
	liveCheck := tlc.NewNoOpLiveCheck(tool, "BasicMultiTrace")
	resultQueue := tlc.NewSimulationWorkerResultQueue()
	worker := tlc.NewSimulationWorker(0, tool, resultQueue, 0, 100, 100, "", false, false, "", liveCheck, nil, nil, nil)
	defer func() { worker.Stop(); worker.Join(0) }()
	worker.Start(initStates)
	res := resultQueue.Take()
	if !(res.IsError()) {
		t.Fatal("original assertTrue: res.IsError()")
	}
	failure := res.Error
	if got := failure.Code; got != tlc.ECTLCStateNotCompletelySpecifiedNext {
		t.Fatalf("original assertEquals: got %v, want %v", got, tlc.ECTLCStateNotCompletelySpecifiedNext)
	}
	if got := failure.StateTrace.Size(); got != 2 {
		t.Fatalf("original assertEquals: got %v, want %v", got, 2)
	}
	if !(failure.StateTrace.At(0).IsInitial()) {
		t.Fatal("original assertTrue: failure.StateTrace.At(0).IsInitial()")
	}
	if !(javaSimulationWorkerTraceLevels(failure.StateTrace)) {
		t.Fatal("original assertTrue: javaSimulationWorkerTraceLevels(failure.StateTrace)")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "branch"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateValue(failure.StateTrace.Last(), "depth"); got != nil {
		t.Fatalf("original assertEquals: got %v, want %v", got, nil)
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.Last(), "branch"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	worker.Join(0)
	if worker.IsAlive() {
		t.Fatal("original assertFalse: worker.IsAlive()")
	}
}

func TestJavaSimulationWorkerDeadlock(t *testing.T) {
	tool := javaSimulationWorkerTool(t, "MC")
	initStates := javaGeneratedInitialStates(t, tool)
	liveCheck := tlc.NewNoOpLiveCheck(tool, "BasicMultiTrace")
	resultQueue := tlc.NewSimulationWorkerResultQueue()
	worker := tlc.NewSimulationWorker(0, tool, resultQueue, 0, 100, 100, "", true, false, "", liveCheck, nil, nil, nil)
	defer func() { worker.Stop(); worker.Join(0) }()
	worker.Start(initStates)
	res := resultQueue.Take()
	if !(res.IsError()) {
		t.Fatal("original assertTrue: res.IsError()")
	}
	failure := res.Error
	if got := failure.Code; got != tlc.ECTLCDeadlockReached {
		t.Fatalf("original assertEquals: got %v, want %v", got, tlc.ECTLCDeadlockReached)
	}
	if got := failure.StateTrace.Size(); got != 7 {
		t.Fatalf("original assertEquals: got %v, want %v", got, 7)
	}
	if !(failure.StateTrace.At(0).IsInitial()) {
		t.Fatal("original assertTrue: failure.StateTrace.At(0).IsInitial()")
	}
	if !(javaSimulationWorkerTraceLevels(failure.StateTrace)) {
		t.Fatal("original assertTrue: javaSimulationWorkerTraceLevels(failure.StateTrace)")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(0), "branch"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(1), "depth"); got != "0" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "0")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(1), "branch"); got != "6" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "6")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(2), "depth"); got != "1" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "1")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(2), "branch"); got != "6" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "6")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(3), "depth"); got != "2" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "2")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(3), "branch"); got != "6" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "6")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(4), "depth"); got != "3" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "3")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(4), "branch"); got != "6" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "6")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(5), "depth"); got != "4" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "4")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(5), "branch"); got != "6" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "6")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(6), "depth"); got != "5" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "5")
	}
	if got := javaSimulationWorkerStateVal(failure.StateTrace.At(6), "branch"); got != "6" {
		t.Fatalf("original assertEquals: got %v, want %v", got, "6")
	}
	worker.Join(0)
	if worker.IsAlive() {
		t.Fatal("original assertFalse: worker.IsAlive()")
	}
}

func TestJavaSimulationWorkerModelStateConstraint(t *testing.T) {
	tool := javaSimulationWorkerTool(t, "MCWithConstraint")
	initStates := javaGeneratedInitialStates(t, tool)
	liveCheck := tlc.NewNoOpLiveCheck(tool, "BasicMultiTrace")
	resultQueue := tlc.NewSimulationWorkerResultQueue()
	worker := tlc.NewSimulationWorker(0, tool, resultQueue, 0, 100, 100, "", false, false, "", liveCheck, nil, nil, nil)
	defer func() { worker.Stop(); worker.Join(0) }()
	worker.Start(initStates)
	res := resultQueue.Take()
	if res.IsError() {
		t.Fatal("original assertFalse: res.IsError()")
	}
	worker.Join(0)
	if !(resultQueue.IsEmpty()) {
		t.Fatal("original assertTrue: resultQueue.IsEmpty()")
	}
	if worker.IsAlive() {
		t.Fatal("original assertFalse: worker.IsAlive()")
	}
}

func TestJavaSimulationWorkerModelActionConstraint(t *testing.T) {
	tool := javaSimulationWorkerTool(t, "MCWithActionConstraint")
	initStates := javaGeneratedInitialStates(t, tool)
	liveCheck := tlc.NewNoOpLiveCheck(tool, "BasicMultiTrace")
	resultQueue := tlc.NewSimulationWorkerResultQueue()
	worker := tlc.NewSimulationWorker(0, tool, resultQueue, 0, 100, 100, "", false, false, "", liveCheck, nil, nil, nil)
	defer func() { worker.Stop(); worker.Join(0) }()
	worker.Start(initStates)
	res := resultQueue.Take()
	if res.IsError() {
		t.Fatal("original assertFalse: res.IsError()")
	}
	worker.Join(0)
	if !(resultQueue.IsEmpty()) {
		t.Fatal("original assertTrue: resultQueue.IsEmpty()")
	}
	if worker.IsAlive() {
		t.Fatal("original assertFalse: worker.IsAlive()")
	}
}

func TestJavaSimulationWorkerWorkerInterruption(t *testing.T) {
	tool := javaSimulationWorkerTool(t, "MCInv")
	initStates := javaGeneratedInitialStates(t, tool)
	liveCheck := tlc.NewNoOpLiveCheck(tool, "BasicMultiTrace")
	resultQueue := tlc.NewSimulationWorkerResultQueue()
	traceNum := int64(math.MaxInt64)
	worker := tlc.NewSimulationWorker(0, tool, resultQueue, 0, 100, traceNum, "", false, false, "", liveCheck, nil, nil, nil)
	defer func() { worker.Stop(); worker.Join(0) }()
	worker.Start(initStates)
	res := resultQueue.Take()
	if !(res.IsError()) {
		t.Fatal("original assertTrue: res.IsError()")
	}
	failure := res.Error
	if got := failure.Code; got != tlc.ECTLCInvariantViolatedBehavior {
		t.Fatalf("original assertEquals: got %v, want %v", got, tlc.ECTLCInvariantViolatedBehavior)
	}
	if got := failure.StateTrace.Size(); got != 4 {
		t.Fatalf("original assertEquals: got %v, want %v", got, 4)
	}
	if !(failure.StateTrace.At(0).IsInitial()) {
		t.Fatal("original assertTrue: failure.StateTrace.At(0).IsInitial()")
	}
	if !(javaSimulationWorkerTraceLevels(failure.StateTrace)) {
		t.Fatal("original assertTrue: javaSimulationWorkerTraceLevels(failure.StateTrace)")
	}
	worker.Stop()
	worker.Join(0)
	if worker.IsAlive() {
		t.Fatal("original assertFalse: worker.IsAlive()")
	}
}

func TestJavaSimulationWorkerTraceDepthObeyed(t *testing.T) {
	tool := javaSimulationWorkerTool(t, "MCInv")
	initStates := javaGeneratedInitialStates(t, tool)
	liveCheck := tlc.NewNoOpLiveCheck(tool, "BasicMultiTrace")
	resultQueue := tlc.NewSimulationWorkerResultQueue()
	traceDepth := int(1)
	worker := tlc.NewSimulationWorker(0, tool, resultQueue, 0, traceDepth, 100, "", false, false, "", liveCheck, nil, nil, nil)
	defer func() { worker.Stop(); worker.Join(0) }()
	worker.Start(initStates)
	res := resultQueue.Take()
	if res.IsError() {
		t.Fatal("original assertFalse: res.IsError()")
	}
	worker.Join(0)
	if !(resultQueue.IsEmpty()) {
		t.Fatal("original assertTrue: resultQueue.IsEmpty()")
	}
	if worker.IsAlive() {
		t.Fatal("original assertFalse: worker.IsAlive()")
	}
}

func TestJavaSimulationWorkerStateAndTraceGenerationCount(t *testing.T) {
	tool := javaSimulationWorkerTool(t, "MC")
	initStates := javaGeneratedInitialStates(t, tool)
	liveCheck := tlc.NewNoOpLiveCheck(tool, "BasicMultiTrace")
	resultQueue := tlc.NewSimulationWorkerResultQueue()
	numOfGenStates := &atomic.Int64{}
	numOfGenTraces := &atomic.Int64{}
	m2AndMean := &atomic.Int64{}
	traceDepth := int(5)
	traceNum := int64(5)
	worker := tlc.NewSimulationWorker(0, tool, resultQueue, 0, traceDepth, traceNum, "", false, false, "", liveCheck, numOfGenStates, numOfGenTraces, m2AndMean)
	defer func() { worker.Stop(); worker.Join(0) }()
	worker.Start(initStates)
	worker.Join(0)
	if worker.IsAlive() {
		t.Fatal("original assertFalse: worker.IsAlive()")
	}
	if got := numOfGenStates.Load(); got != 70 {
		t.Fatalf("original assertEquals: got %v, want %v", got, 70)
	}
	if got := numOfGenTraces.Load(); got != 5 {
		t.Fatalf("original assertEquals: got %v, want %v", got, 5)
	}
	if got := m2AndMean.Load(); got != 5 {
		t.Fatalf("original assertEquals: got %v, want %v", got, 5)
	}
}
