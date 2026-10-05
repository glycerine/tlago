package tlc

import (
	"math"
	"testing"
)

// Mechanical port of tlc2.tool.SimulationWorkerTest and its equality dummy.
type javaSimulationEqualityDummyTLCState struct {
	*javaDummyTLCState
	level       int
	id          int32
	predecessor TLCPredecessorState
}

func newJavaSimulationEqualityDummyTLCState(fp uint64, id int32, predecessor ...TLCPredecessorState) *javaSimulationEqualityDummyTLCState {
	s := &javaSimulationEqualityDummyTLCState{javaDummyTLCState: newJavaDummyTLCState(int64(fp)), level: TLCStateInitLevel, id: id}
	if len(predecessor) > 0 {
		s.SetTracePredecessor(predecessor[0])
	}
	return s
}
func (s *javaSimulationEqualityDummyTLCState) FingerPrint() uint64              { return s.fp }
func (s *javaSimulationEqualityDummyTLCState) FingerPrintWithTool(*Tool) uint64 { return s.fp }
func (s *javaSimulationEqualityDummyTLCState) HashCode() int32                  { return 31*(31+s.id) + int32(s.fp) }
func (s *javaSimulationEqualityDummyTLCState) Equal(obj TLCState) bool {
	if s == obj {
		return true
	}
	other, ok := obj.(*javaSimulationEqualityDummyTLCState)
	return ok && other != nil && s.fp == other.fp && s.id == other.id
}
func (s *javaSimulationEqualityDummyTLCState) TracePredecessor() TLCPredecessorState {
	return s.predecessor
}
func (s *javaSimulationEqualityDummyTLCState) SetTracePredecessor(pred TLCPredecessorState) {
	s.predecessor = pred
	if pred.Level() == math.MaxInt32 {
		panic(newTLCError(ECTLCTraceTooLong, "%s", s.String()))
	}
	s.level = pred.Level() + 1
}

func (s *javaSimulationEqualityDummyTLCState) Level() int      { return s.level }
func (s *javaSimulationEqualityDummyTLCState) IsInitial() bool { return s.level == TLCStateInitLevel }

func TestJavaSimulationWorkerGetTraceTLCState0(t *testing.T) {
	sw := NewSimulationWorker(0, nil, nil, 0, 0, 0, "", false, false, "", nil, nil, nil, nil)
	trace := sw.GetTrace(nil)
	if trace.Size() != 0 {
		t.Fatalf("original assertEquals failed: 0, trace.Size()")
	}
}

func TestJavaSimulationWorkerGetTraceTLCState1(t *testing.T) {
	sw := NewSimulationWorker(0, nil, nil, 0, 0, 0, "", false, false, "", nil, nil, nil, nil)
	a := newJavaSimulationEqualityDummyTLCState(0, 0)
	trace := sw.GetTrace(a)
	if trace.Size() != 1 {
		t.Fatalf("original assertEquals failed: 1, trace.Size()")
	}
	if !(trace.ElementAt(0).(TLCPredecessorState).IsInitial()) {
		t.Fatalf("original assertTrue failed: trace.ElementAt(0).(TLCPredecessorState).IsInitial()")
	}
	if trace.ElementAt(0).(TLCPredecessorState).TracePredecessor() != nil {
		t.Fatalf("original assertNull failed: trace.ElementAt(0).(TLCPredecessorState).TracePredecessor()")
	}
	if trace.ElementAt(0).(TLCPredecessorState).FingerPrint() != 0 {
		t.Fatalf("original assertEquals failed: 0, trace.ElementAt(0).(TLCPredecessorState).FingerPrint()")
	}
}

func TestJavaSimulationWorkerGetTraceTLCState2(t *testing.T) {
	sw := NewSimulationWorker(0, nil, nil, 0, 0, 0, "", false, false, "", nil, nil, nil, nil)
	a := newJavaSimulationEqualityDummyTLCState(0, 0)
	b := newJavaSimulationEqualityDummyTLCState(0, 0, a)
	trace := sw.GetTrace(b)
	if trace.Size() != 1 {
		t.Fatalf("original assertEquals failed: 1, trace.Size()")
	}
	if !(trace.ElementAt(0).(TLCPredecessorState).IsInitial()) {
		t.Fatalf("original assertTrue failed: trace.ElementAt(0).(TLCPredecessorState).IsInitial()")
	}
	if trace.ElementAt(0).(TLCPredecessorState).TracePredecessor() != nil {
		t.Fatalf("original assertNull failed: trace.ElementAt(0).(TLCPredecessorState).TracePredecessor()")
	}
	if trace.ElementAt(0).(TLCPredecessorState).FingerPrint() != 0 {
		t.Fatalf("original assertEquals failed: 0, trace.ElementAt(0).(TLCPredecessorState).FingerPrint()")
	}
}

func TestJavaSimulationWorkerGetTraceTLCState3(t *testing.T) {
	sw := NewSimulationWorker(0, nil, nil, 0, 0, 0, "", false, false, "", nil, nil, nil, nil)
	a := newJavaSimulationEqualityDummyTLCState(0, 0)
	b := newJavaSimulationEqualityDummyTLCState(0, 0, a)
	c := newJavaSimulationEqualityDummyTLCState(0, 0, b)
	trace := sw.GetTrace(c)
	if trace.Size() != 1 {
		t.Fatalf("original assertEquals failed: 1, trace.Size()")
	}
	if !(trace.ElementAt(0).(TLCPredecessorState).IsInitial()) {
		t.Fatalf("original assertTrue failed: trace.ElementAt(0).(TLCPredecessorState).IsInitial()")
	}
	if trace.ElementAt(0).(TLCPredecessorState).TracePredecessor() != nil {
		t.Fatalf("original assertNull failed: trace.ElementAt(0).(TLCPredecessorState).TracePredecessor()")
	}
	if trace.ElementAt(0).(TLCPredecessorState).FingerPrint() != 0 {
		t.Fatalf("original assertEquals failed: 0, trace.ElementAt(0).(TLCPredecessorState).FingerPrint()")
	}
}

func TestJavaSimulationWorkerGetTraceTLCState4(t *testing.T) {
	sw := NewSimulationWorker(0, nil, nil, 0, 0, 0, "", false, false, "", nil, nil, nil, nil)
	a := newJavaSimulationEqualityDummyTLCState(0, 0)
	b := newJavaSimulationEqualityDummyTLCState(0, 0, a)
	c := newJavaSimulationEqualityDummyTLCState(0, 0, b)
	d := newJavaSimulationEqualityDummyTLCState(1, 1, c)
	trace := sw.GetTrace(d)
	if trace.Size() != 2 {
		t.Fatalf("original assertEquals failed: 2, trace.Size()")
	}
	if !(trace.ElementAt(0).(TLCPredecessorState).IsInitial()) {
		t.Fatalf("original assertTrue failed: trace.ElementAt(0).(TLCPredecessorState).IsInitial()")
	}
	if trace.ElementAt(0).(TLCPredecessorState).FingerPrint() != 0 {
		t.Fatalf("original assertEquals failed: 0, trace.ElementAt(0).(TLCPredecessorState).FingerPrint()")
	}
	if trace.ElementAt(0).(TLCPredecessorState).TracePredecessor() != nil {
		t.Fatalf("original assertNull failed: trace.ElementAt(0).(TLCPredecessorState).TracePredecessor()")
	}
	if trace.ElementAt(1).(TLCPredecessorState).IsInitial() {
		t.Fatalf("original assertFalse failed: trace.ElementAt(1).(TLCPredecessorState).IsInitial()")
	}
	if trace.ElementAt(1).(TLCPredecessorState).FingerPrint() != 1 {
		t.Fatalf("original assertEquals failed: 1, trace.ElementAt(1).(TLCPredecessorState).FingerPrint()")
	}
	if trace.ElementAt(1).(TLCPredecessorState).TracePredecessor() != a {
		t.Fatalf("original assertSame failed: a, trace.ElementAt(1).(TLCPredecessorState).TracePredecessor()")
	}
}

func TestJavaSimulationWorkerGetTraceTLCState5(t *testing.T) {
	sw := NewSimulationWorker(0, nil, nil, 0, 0, 0, "", false, false, "", nil, nil, nil, nil)
	a := newJavaSimulationEqualityDummyTLCState(0, 0)
	b := newJavaSimulationEqualityDummyTLCState(0, 0, a)
	c := newJavaSimulationEqualityDummyTLCState(0, 0, b)
	d := newJavaSimulationEqualityDummyTLCState(1, 1, c)
	e := newJavaSimulationEqualityDummyTLCState(2, 2, d)
	trace := sw.GetTrace(e)
	if trace.Size() != 3 {
		t.Fatalf("original assertEquals failed: 3, trace.Size()")
	}
	if !(trace.ElementAt(0).(TLCPredecessorState).IsInitial()) {
		t.Fatalf("original assertTrue failed: trace.ElementAt(0).(TLCPredecessorState).IsInitial()")
	}
	if trace.ElementAt(0).(TLCPredecessorState).FingerPrint() != 0 {
		t.Fatalf("original assertEquals failed: 0, trace.ElementAt(0).(TLCPredecessorState).FingerPrint()")
	}
	if trace.ElementAt(0).(TLCPredecessorState).TracePredecessor() != nil {
		t.Fatalf("original assertNull failed: trace.ElementAt(0).(TLCPredecessorState).TracePredecessor()")
	}
	if trace.ElementAt(1).(TLCPredecessorState).IsInitial() {
		t.Fatalf("original assertFalse failed: trace.ElementAt(1).(TLCPredecessorState).IsInitial()")
	}
	if trace.ElementAt(1).(TLCPredecessorState).FingerPrint() != 1 {
		t.Fatalf("original assertEquals failed: 1, trace.ElementAt(1).(TLCPredecessorState).FingerPrint()")
	}
	if trace.ElementAt(1).(TLCPredecessorState).TracePredecessor() != a {
		t.Fatalf("original assertSame failed: a, trace.ElementAt(1).(TLCPredecessorState).TracePredecessor()")
	}
	if trace.ElementAt(2).(TLCPredecessorState).IsInitial() {
		t.Fatalf("original assertFalse failed: trace.ElementAt(2).(TLCPredecessorState).IsInitial()")
	}
	if trace.ElementAt(2).(TLCPredecessorState).FingerPrint() != 2 {
		t.Fatalf("original assertEquals failed: 2, trace.ElementAt(2).(TLCPredecessorState).FingerPrint()")
	}
	if trace.ElementAt(2).(TLCPredecessorState).TracePredecessor() != d {
		t.Fatalf("original assertSame failed: d, trace.ElementAt(2).(TLCPredecessorState).TracePredecessor()")
	}
}

func TestJavaSimulationWorkerGetTraceTLCState6(t *testing.T) {
	sw := NewSimulationWorker(0, nil, nil, 0, 0, 0, "", false, false, "", nil, nil, nil, nil)
	a := newJavaSimulationEqualityDummyTLCState(0, 0)
	b := newJavaSimulationEqualityDummyTLCState(0, 0, a)
	c := newJavaSimulationEqualityDummyTLCState(0, 0, b)
	d := newJavaSimulationEqualityDummyTLCState(1, 1, c)
	e := newJavaSimulationEqualityDummyTLCState(1, 1, d)
	trace := sw.GetTrace(e)
	if trace.Size() != 2 {
		t.Fatalf("original assertEquals failed: 2, trace.Size()")
	}
	if !(trace.ElementAt(0).(TLCPredecessorState).IsInitial()) {
		t.Fatalf("original assertTrue failed: trace.ElementAt(0).(TLCPredecessorState).IsInitial()")
	}
	if trace.ElementAt(0).(TLCPredecessorState).TracePredecessor() != nil {
		t.Fatalf("original assertNull failed: trace.ElementAt(0).(TLCPredecessorState).TracePredecessor()")
	}
	if trace.ElementAt(0).(TLCPredecessorState).FingerPrint() != 0 {
		t.Fatalf("original assertEquals failed: 0, trace.ElementAt(0).(TLCPredecessorState).FingerPrint()")
	}
	if trace.ElementAt(1).(TLCPredecessorState).IsInitial() {
		t.Fatalf("original assertFalse failed: trace.ElementAt(1).(TLCPredecessorState).IsInitial()")
	}
	if trace.ElementAt(1).(TLCPredecessorState).FingerPrint() != 1 {
		t.Fatalf("original assertEquals failed: 1, trace.ElementAt(1).(TLCPredecessorState).FingerPrint()")
	}
	if trace.ElementAt(1).(TLCPredecessorState).TracePredecessor() != a {
		t.Fatalf("original assertSame failed: a, trace.ElementAt(1).(TLCPredecessorState).TracePredecessor()")
	}
}
