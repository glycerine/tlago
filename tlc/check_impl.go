package tlc

import (
	"fmt"
	"time"
)

const checkImplTraceDurationMillis int64 = 30000

type CheckImpl struct {
	*ModelChecker
	Depth         int
	CoverSet      FPSet
	CurState      *TLCStateMut
	StateEnum     *TLCTraceEnumerator
	LastTraceTime int64
	GetStateFunc  func() *TLCStateMut
	ExportTraceFn func([]*TLCStateInfo) error
}

func NewCheckImpl(tool *Tool, metadir string, deadlock bool, depth int, fromChkpt *string, config ...*FPSetConfiguration) *CheckImpl {
	if len(config) > 1 {
		panic(NewIllegalArgumentException("NewCheckImpl accepts at most one fingerprint-set configuration"))
	}
	fpConfig := NewFPSetConfiguration()
	if len(config) > 0 {
		if config[0] == nil {
			panic(NewNullPointerException())
		}
		fpConfig = config[0]
	}
	startTime := time.Now()
	mc := NewModelChecker(tool, metadir, deadlock,
		WithModelCheckerFromCheckpoint(fromChkpt),
		WithModelCheckerFPSetConfiguration(fpConfig),
		WithModelCheckerStartTime(startTime),
	)
	rootName := "Spec"
	if tool != nil {
		rootName = tool.GetRootName()
	}
	coverSet := NewFPSet(NewFPSetConfiguration()).Init(NumWorkers(), metadir, rootName+"_cs")
	return &CheckImpl{
		ModelChecker: mc,
		Depth:        checkImplDepth(depth),
		CoverSet:     coverSet,
	}
}

func (c *CheckImpl) Init() (int, error) {
	if c == nil || c.ModelChecker == nil {
		panic(NewNullPointerException())
	}
	recovered, err := c.Recover()
	if err != nil {
		return ECSystemCheckpointRecoveryCorrupt, err
	}
	if !recovered {
		_ = c.CheckAssumptions()
		if _, err := c.DoInit(false); err != nil {
			return ECGeneral, err
		}
	}
	ToolIOPrintln(fmt.Sprintf("Creating a partial state space of depth %d ... ", c.Depth))
	result, err := c.RunTLC(c.Depth)
	if err != nil {
		return result, err
	}
	if result != NoError {
		ToolIOPrintln("\nExit: failed to create the partial state space.")
		ExitTLCProcess(result)
		return result, err
	}
	ToolIOPrintln("completed.")
	c.LastTraceTime = time.Now().UnixMilli()
	if c.ownedTrace() == nil {
		panic(NewNullPointerException())
	}
	c.StateEnum, err = c.ownedTrace().Elements()
	if err != nil {
		return ECGeneral, err
	}
	return NoError, nil
}

func (c *CheckImpl) Reset() error {
	if c == nil {
		panic(NewNullPointerException())
	}
	c.CurState = nil
	if c.StateEnum == nil {
		panic(NewNullPointerException())
	}
	return c.StateEnum.Reset(-1)
}

func (c *CheckImpl) MakeStateSpace(state *TLCStateMut, depth int) (int, error) {
	if c == nil || c.ModelChecker == nil || c.ownedTrace() == nil || state == nil {
		panic(NewNullPointerException())
	}
	depth1 := checkImplStateSpaceDepth(c.ownedTrace().GetLevel(state.UID), depth)
	c.StateQueue = NewDiskStateQueue(c.ownedMetadir())
	c.StateQueue.Enqueue(state)
	result, err := c.RunTLC(depth1)
	if err != nil {
		return result, err
	}
	if result != NoError {
		ExitTLCProcess(result)
		return result, nil
	}
	return NoError, nil
}

func checkImplStateSpaceDepth(level int, depth int) int {
	return int(int32(level) + int32(depth))
}

func checkImplDepth(depth int) int {
	return int(int32(depth))
}

func (c *CheckImpl) GetState() *TLCStateMut {
	if c == nil {
		panic(NewNullPointerException())
	}
	if c.GetStateFunc == nil {
		return nil
	}
	return c.GetStateFunc()
}

func (c *CheckImpl) ExportTrace(trace []*TLCStateInfo) error {
	if c == nil {
		panic(NewNullPointerException())
	}
	if c.ExportTraceFn == nil {
		return nil
	}
	return c.ExportTraceFn(trace)
}

func (c *CheckImpl) CheckReachability(s0 *TLCStateMut, s1 *TLCStateMut) (bool, error) {
	if c == nil || c.ownedTool() == nil {
		panic(NewNullPointerException())
	}
	ok, err := c.ownedTool().IsValidTransition(c.ownedTool().GetNextStateSpec(), s0, s1)
	if err != nil {
		return false, err
	}
	if !ok {
		ToolIOPrintln("The following transition is illegal: ")
		PrintStandaloneErrorState(s0)
		PrintStandaloneErrorState(s1)
		return false, nil
	}
	count := checkImplActionCount(c.ownedTool().GetImpliedActions())
	for i := 0; i < count; i++ {
		action := checkImplActionAt(c.ownedTool().GetImpliedActions(), i)
		ok, err := c.ownedTool().IsValidTransition(action, s0, s1)
		if err != nil {
			return false, err
		}
		if !ok {
			ToolIOPrintln(fmt.Sprintf("Error: Action property %s is violated.", checkImplNameAt(c.ownedTool().GetImpliedActNames(), i)))
			PrintStandaloneErrorState(s0)
			PrintStandaloneErrorState(s1)
			return false, nil
		}
	}
	return true, nil
}

func (c *CheckImpl) CheckState(state *TLCStateMut) (bool, error) {
	if c == nil || state == nil {
		panic(NewNullPointerException())
	}
	fp := state.FingerPrint()
	if isNil(c.CoverSet) {
		panic(NewNullPointerException())
	}
	seen := c.CoverSet.Put(fp)
	if seen {
		return true, nil
	}
	if isNil(c.FPSet) {
		panic(NewNullPointerException())
	}
	if !c.FPSet.Contains(fp) {
		if c.ownedTrace() == nil {
			panic(NewNullPointerException())
		}
		_, err := c.ownedTrace().WriteStateRecord(c.CurState, fp, state)
		if err != nil {
			return false, err
		}
		if c.ownedTool() == nil {
			panic(NewNullPointerException())
		}
		count := checkImplActionCount(c.ownedTool().GetInvariants())
		for i := 0; i < count; i++ {
			invariant := checkImplActionAt(c.ownedTool().GetInvariants(), i)
			ok, err := c.ownedTool().IsValidState(invariant, state)
			if err != nil {
				return false, err
			}
			if !ok {
				ToolIOPrintln(fmt.Sprintf("Error: Invariant %s is violated. The behavior up to this point is:", checkImplNameAt(c.ownedTool().GetInvNames(), i)))
				return false, nil
			}
		}
	}
	return true, nil
}

func (c *CheckImpl) GenerateNewTrace() ([]*TLCStateInfo, error) {
	if c == nil || c.StateEnum == nil {
		panic(NewNullPointerException())
	}
	for {
		pos, err := c.StateEnum.NextPos()
		if err != nil {
			return nil, err
		}
		if pos == -1 {
			break
		}
		fp, err := c.StateEnum.NextFP()
		if err != nil {
			return nil, err
		}
		if isNil(c.CoverSet) {
			panic(NewNullPointerException())
		}
		if !c.CoverSet.Contains(fp) {
			if c.ownedTrace() == nil {
				panic(NewNullPointerException())
			}
			return c.ownedTrace().GetTraceAt(pos, true), nil
		}
	}
	return nil, nil
}

func (c *CheckImpl) CheckTrace() error {
	if c == nil {
		panic(NewNullPointerException())
	}
	c.CurState = c.GetState()
	if c.CurState == nil {
		return nil
	}
	if _, err := c.CheckState(c.CurState); err != nil {
		return err
	}
	for {
		state := c.GetState()
		if state == nil {
			return nil
		}
		if _, err := c.CheckState(state); err != nil {
			return err
		}
		if _, err := c.CheckReachability(c.CurState, state); err != nil {
			return err
		}
		c.CurState = state
	}
}

func (c *CheckImpl) Export() error {
	if c == nil {
		panic(NewNullPointerException())
	}
	curTime := time.Now().UnixMilli()
	if curTime-c.LastTraceTime <= checkImplTraceDurationMillis {
		return nil
	}
	trace, err := c.GenerateNewTrace()
	if err != nil {
		return err
	}
	if trace != nil {
		if err := c.ExportTrace(trace); err != nil {
			return err
		}
	}
	c.LastTraceTime = curTime
	return nil
}

func checkImplNameAt(names []string, index int) string {
	if names == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(names) {
		panic(NewArrayIndexOutOfBoundsException(index, len(names)))
	}
	return names[index]
}

func checkImplActionCount(actions []*Action) int {
	if actions == nil {
		panic(NewNullPointerException())
	}
	return len(actions)
}

func checkImplActionAt(actions []*Action, index int) *Action {
	if actions == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(actions) {
		panic(NewArrayIndexOutOfBoundsException(index, len(actions)))
	}
	return actions[index]
}
