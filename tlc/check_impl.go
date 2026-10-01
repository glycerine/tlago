package tlc

import (
	"fmt"
	"os"
	"time"
)

const checkImplTraceDuration = 30 * time.Second

type CheckImpl struct {
	*ModelChecker
	Depth         int
	CoverSet      FPSet
	CurState      *TLCStateMut
	StateEnum     *TLCTraceEnumerator
	LastTraceTime time.Time
	GetStateFunc  func() *TLCStateMut
	ExportTraceFn func([]*TLCStateInfo) error
}

func NewCheckImpl(tool *Tool, metadir string, deadlock bool, depth int, fromChkpt string, config ...*FPSetConfiguration) *CheckImpl {
	fpConfig := NewFPSetConfiguration()
	if len(config) > 0 && config[0] != nil {
		fpConfig = config[0]
	}
	mc := NewModelChecker(tool, metadir, deadlock,
		WithModelCheckerFromCheckpoint(fromChkpt),
		WithModelCheckerFPSet(NewFPSet(fpConfig)),
	)
	rootName := "Spec"
	if tool != nil {
		rootName = tool.GetRootName()
	}
	coverSet := NewFPSet(NewFPSetConfiguration()).Init(NumWorkers(), metadir, rootName+"_cs")
	return &CheckImpl{
		ModelChecker: mc,
		Depth:        depth,
		CoverSet:     coverSet,
	}
}

func (c *CheckImpl) Init() (int, error) {
	if c == nil || c.ModelChecker == nil {
		return ECGeneral, newTLCError(ECGeneral, "CheckImpl has no model checker")
	}
	recovered, err := c.Recover()
	if err != nil {
		return ECSystemCheckpointRecoveryCorrupt, err
	}
	if !recovered {
		if result := c.CheckAssumptions(); result != NoError {
			return result, nil
		}
		result, err := c.DoInit(false)
		if err != nil || result != NoError {
			return result, err
		}
	}
	fmt.Fprintf(os.Stdout, "Creating a partial state space of depth %d ... \n", c.Depth)
	result, err := c.RunTLC(c.Depth)
	if err != nil || result != NoError {
		fmt.Fprintln(os.Stdout, "\nExit: failed to create the partial state space.")
		return result, err
	}
	fmt.Fprintln(os.Stdout, "completed.")
	c.LastTraceTime = time.Now()
	if c.Trace != nil {
		c.StateEnum = c.Trace.Elements()
	}
	return NoError, nil
}

func (c *CheckImpl) Reset() {
	if c == nil {
		return
	}
	c.CurState = nil
	if c.StateEnum != nil {
		c.StateEnum.Reset(-1)
	} else if c.Trace != nil {
		c.StateEnum = c.Trace.Elements()
	}
}

func (c *CheckImpl) MakeStateSpace(state *TLCStateMut, depth int) (int, error) {
	if c == nil || c.ModelChecker == nil {
		return ECGeneral, newTLCError(ECGeneral, "CheckImpl has no model checker")
	}
	depth1 := depth
	if state != nil {
		if c.Trace != nil {
			traceLevel := c.Trace.GetLevelForState(state)
			if traceLevel > 0 {
				depth1 += traceLevel
			} else {
				depth1 += state.Level()
			}
		} else {
			depth1 += state.Level()
		}
	}
	c.StateQueue = NewDiskStateQueue(c.Metadir)
	if state != nil {
		c.StateQueue.Enqueue(state)
	}
	result, err := c.RunTLC(depth1)
	if err != nil || result != NoError {
		return result, err
	}
	return NoError, nil
}

func (c *CheckImpl) GetState() *TLCStateMut {
	if c == nil || c.GetStateFunc == nil {
		return nil
	}
	return c.GetStateFunc()
}

func (c *CheckImpl) ExportTrace(trace []*TLCStateInfo) error {
	if c == nil || c.ExportTraceFn == nil {
		return nil
	}
	return c.ExportTraceFn(trace)
}

func (c *CheckImpl) CheckReachability(s0 *TLCStateMut, s1 *TLCStateMut) (bool, error) {
	if c == nil || c.Tool == nil {
		return false, newTLCError(ECGeneral, "CheckImpl has no tool")
	}
	if c.Tool.NextStateSpec != nil {
		ok, err := c.Tool.IsValidTransition(c.Tool.NextStateSpec, s0, s1)
		if err != nil {
			return false, err
		}
		if !ok {
			fmt.Fprintln(os.Stdout, "The following transition is illegal: ")
			PrintStandaloneErrorState(s0)
			PrintStandaloneErrorState(s1)
			return false, nil
		}
	}
	actions := c.Tool.GetImpliedActions()
	for i, action := range actions {
		ok, err := c.Tool.IsValidTransition(action, s0, s1)
		if err != nil {
			return false, err
		}
		if !ok {
			if i < len(c.Tool.ImpliedActNames) {
				PrintError(ECTLCActionPropertyViolatedBehavior, c.Tool.ImpliedActNames[i])
			}
			PrintStandaloneErrorState(s0)
			PrintStandaloneErrorState(s1)
			return false, nil
		}
	}
	return true, nil
}

func (c *CheckImpl) CheckState(state *TLCStateMut) (bool, error) {
	if c == nil || state == nil {
		return true, nil
	}
	fp := state.FingerPrint()
	seen := false
	if c.CoverSet != nil {
		seen = c.CoverSet.Put(fp)
	}
	if seen {
		return true, nil
	}
	if c.FPSet == nil || !c.FPSet.Contains(fp) {
		if c.Trace != nil {
			var err error
			if c.CurState == nil {
				err = c.Trace.WriteInitState(state, fp)
			} else {
				err = c.Trace.WriteNextState(c.CurState, fp, state, nil)
			}
			if err != nil {
				return false, err
			}
		}
		for i, invariant := range c.Tool.GetInvariants() {
			ok, err := c.Tool.IsValidState(invariant, state)
			if err != nil {
				return false, err
			}
			if !ok {
				if i < len(c.Tool.InvariantNames) {
					PrintError(ECTLCInvariantViolatedBehavior, c.Tool.InvariantNames[i])
				}
				fmt.Fprintln(os.Stdout, "The behavior up to this point is:")
				return false, nil
			}
		}
	}
	return true, nil
}

func (c *CheckImpl) GenerateNewTrace() ([]*TLCStateInfo, error) {
	if c == nil || c.StateEnum == nil || c.Trace == nil {
		return nil, nil
	}
	for pos := c.StateEnum.NextPos(); pos != -1; pos = c.StateEnum.NextPos() {
		fp := c.StateEnum.NextFP()
		if c.CoverSet == nil || !c.CoverSet.Contains(fp) {
			return c.Trace.GetTraceAt(pos, true), nil
		}
	}
	return nil, nil
}

func (c *CheckImpl) CheckTrace() error {
	if c == nil {
		return nil
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
		return nil
	}
	if time.Since(c.LastTraceTime) <= checkImplTraceDuration {
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
	c.LastTraceTime = time.Now()
	return nil
}
