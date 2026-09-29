package tlc

type ToolMode int

const (
	ModeMC ToolMode = iota
	ModeMCDFS
	ModeSimulation
	ModeDebugger
)

type StateFunctor struct {
	AddElementFunc          func(*TLCStateMut) (any, error)
	SetElementFunc          func(*TLCStateMut) (any, error)
	HasStatesFunc           func() bool
	GetStatesFunc           func() *StateVec
	AddUnsatisfiedStateFunc func(*TLCStateMut, SemanticNode, *Context) *TLCStateMut
}

func NewStateFunctor(addElement func(*TLCStateMut) (any, error)) *StateFunctor {
	return &StateFunctor{AddElementFunc: addElement}
}

func (f *StateFunctor) AddElement(state *TLCStateMut) (any, error) {
	if f != nil && f.AddElementFunc != nil {
		return f.AddElementFunc(state)
	}
	return nil, nil
}

func (f *StateFunctor) SetElement(state *TLCStateMut) (any, error) {
	if f != nil && f.SetElementFunc != nil {
		return f.SetElementFunc(state)
	}
	return nil, newTLCError(ECGeneral, "StateFunctor.SetElement is unsupported")
}

func (f *StateFunctor) HasStates() bool {
	return f != nil && f.HasStatesFunc != nil && f.HasStatesFunc()
}

func (f *StateFunctor) GetStates() *StateVec {
	if f != nil && f.GetStatesFunc != nil {
		return f.GetStatesFunc()
	}
	return NewStateVec(0)
}

func (f *StateFunctor) AddUnsatisfiedState(state *TLCStateMut, pred SemanticNode, con *Context) *TLCStateMut {
	if f != nil && f.AddUnsatisfiedStateFunc != nil {
		return f.AddUnsatisfiedStateFunc(state, pred, con)
	}
	return state
}

type NextStateFunctor struct {
	StateFunctor
	AddNextElementFunc         func(*TLCStateMut, *Action, *TLCStateMut) (any, error)
	IncrementStatesGeneratedFn func(int64)
	HaltFunc                   func() bool
	AddUnsatisfiedNextStateFn  func(*TLCStateMut, *Action, *TLCStateMut, SemanticNode, *Context) *TLCStateMut
}

func NewNextStateFunctor(addNextElement func(*TLCStateMut, *Action, *TLCStateMut) (any, error)) *NextStateFunctor {
	return &NextStateFunctor{AddNextElementFunc: addNextElement}
}

func (f *NextStateFunctor) AddNextElement(curState *TLCStateMut, action *Action, succState *TLCStateMut) (any, error) {
	if f != nil && f.AddNextElementFunc != nil {
		return f.AddNextElementFunc(curState, action, succState)
	}
	return nil, nil
}

func (f *NextStateFunctor) IncrementStatesGenerated(count int64) {
	if f != nil && f.IncrementStatesGeneratedFn != nil {
		f.IncrementStatesGeneratedFn(count)
	}
}

func (f *NextStateFunctor) Halt() bool {
	return f != nil && f.HaltFunc != nil && f.HaltFunc()
}

func (f *NextStateFunctor) AddUnsatisfiedNextState(curState *TLCStateMut, action *Action, succState *TLCStateMut, pred SemanticNode, con *Context) *TLCStateMut {
	if f != nil && f.AddUnsatisfiedNextStateFn != nil {
		return f.AddUnsatisfiedNextStateFn(curState, action, succState, pred, con)
	}
	return succState
}

type Tool struct {
	Mode ToolMode

	Actions          []*Action
	InitStates       []*TLCStateMut
	Invariants       []*Action
	InvariantNames   []string
	ImpliedInits     []*Action
	ImpliedInitNames []string
	ImpliedActions   []*Action
	ImpliedActNames  []string

	RootName   string
	RootFile   string
	ConfigFile string
	SpecDir    string

	GetInitStatesFunc               func(*Tool, *StateFunctor) error
	GetNextStatesFunc               func(*Tool, *Action, *TLCStateMut) (*StateVec, error)
	GetNextStatesWithFunctorFn      func(*Tool, *NextStateFunctor, *TLCStateMut) (bool, error)
	GetNextStatesForActionFunc      func(*Tool, *NextStateFunctor, *TLCStateMut, *Action) (bool, error)
	MakeStateFunc                   func(*Tool, SemanticNode) (*TLCStateMut, error)
	EvalFunc                        func(*Tool, SemanticNode, ...any) (Value, error)
	IsGoodStateFunc                 func(*Tool, *TLCStateMut) bool
	IsInModelFunc                   func(*Tool, *TLCStateMut) (bool, error)
	IsInModelForConstraintFunc      func(*Tool, SemanticNode, *TLCStateMut) (bool, error)
	IsInActionsFunc                 func(*Tool, *TLCStateMut, *TLCStateMut) (bool, error)
	IsInActionsForConstraintFn      func(*Tool, SemanticNode, *TLCStateMut, *TLCStateMut) (bool, error)
	EvalRewardFunc                  func(*Tool, *TLCStateMut, *TLCStateMut, float64) (float64, error)
	HasStateOrActionConstraintsFunc func(*Tool) bool
	EnabledFunc                     func(*Tool, SemanticNode, *Context, *TLCStateMut, *TLCStateMut) *TLCStateMut
	IsValidExprFunc                 func(*Tool, SemanticNode, *Context) (bool, error)
	IsValidTransitionFunc           func(*Tool, *Action, *TLCStateMut, *TLCStateMut) (bool, error)
	IsValidStateFunc                func(*Tool, *Action, *TLCStateMut) (bool, error)
	IsValidActionFunc               func(*Tool, *Action) (bool, error)
	GetStateFunc                    func(*Tool, uint64, ...any) (*TLCStateInfo, error)
	HasSymmetryFunc                 func(*Tool) bool
	CheckAssumptionsFunc            func(*Tool) int
	CheckPostConditionFunc          func(*Tool) int
	CheckPostConditionCEFunc        func(*Tool, Value) int
	LivenessIsTrueFunc              func(*Tool) bool
	EvalAliasInfoFunc               func(*Tool, *TLCStateInfo, *TLCStateMut, func() []*TLCStateInfo) (*TLCStateInfo, error)
	EvalAliasInfoPairFunc           func(*Tool, *TLCStateInfo, *TLCStateMut) (*TLCStateInfo, error)
	EvalAliasFunc                   func(*Tool, *TLCStateMut, *TLCStateMut) *TLCStateMut
	NoDebugFunc                     func(*Tool) *Tool
	IsDebuggerFunc                  func(*Tool) bool
}

func NewTool() *Tool {
	return &Tool{Mode: ModeMC, RootName: "Spec"}
}

func (t *Tool) GetMode() ToolMode {
	if t == nil {
		return ModeMC
	}
	return t.Mode
}

func (t *Tool) GetActions() []*Action {
	if t == nil {
		return nil
	}
	return t.Actions
}

func (t *Tool) GetInitStates(functor *StateFunctor) error {
	if t != nil && t.GetInitStatesFunc != nil {
		return t.GetInitStatesFunc(t, functor)
	}
	if t == nil || functor == nil {
		return nil
	}
	for _, state := range t.InitStates {
		if _, err := functor.AddElement(state); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tool) MakeState(pred SemanticNode) (*TLCStateMut, error) {
	if t != nil && t.MakeStateFunc != nil {
		return t.MakeStateFunc(t, pred)
	}
	return NewEmptyState(), nil
}

func (t *Tool) GetNextStates(action *Action, state *TLCStateMut) (*StateVec, error) {
	if t != nil && t.GetNextStatesFunc != nil {
		return t.GetNextStatesFunc(t, action, state)
	}
	return NewStateVec(0), nil
}

func (t *Tool) GetNextStatesWithFunctor(functor *NextStateFunctor, state *TLCStateMut) (bool, error) {
	if t != nil && t.GetNextStatesWithFunctorFn != nil {
		return t.GetNextStatesWithFunctorFn(t, functor, state)
	}
	for _, action := range t.GetActions() {
		halt, err := t.GetNextStatesForAction(functor, state, action)
		if halt || err != nil {
			return halt, err
		}
	}
	return false, nil
}

func (t *Tool) GetNextStatesForAction(functor *NextStateFunctor, state *TLCStateMut, action *Action) (bool, error) {
	if t != nil && t.GetNextStatesForActionFunc != nil {
		return t.GetNextStatesForActionFunc(t, functor, state, action)
	}
	next, err := t.GetNextStates(action, state)
	if err != nil {
		return true, err
	}
	if next == nil {
		return false, nil
	}
	for i := 0; i < next.Size(); i++ {
		if _, err := functor.AddNextElement(state, action, next.At(i)); err != nil {
			return true, err
		}
		if functor.Halt() {
			return true, nil
		}
	}
	return false, nil
}

func (t *Tool) Eval(expr SemanticNode, args ...any) (Value, error) {
	if t != nil && t.EvalFunc != nil {
		return t.EvalFunc(t, expr, args...)
	}
	return ValUndef, nil
}

func (t *Tool) IsGoodState(state *TLCStateMut) bool {
	if t != nil && t.IsGoodStateFunc != nil {
		return t.IsGoodStateFunc(t, state)
	}
	return state != nil && state.AllAssigned()
}

func (t *Tool) IsInModel(state *TLCStateMut) (bool, error) {
	if t != nil && t.IsInModelFunc != nil {
		return t.IsInModelFunc(t, state)
	}
	return true, nil
}

func (t *Tool) IsInModelForConstraint(constraint SemanticNode, state *TLCStateMut) (bool, error) {
	if t != nil && t.IsInModelForConstraintFunc != nil {
		return t.IsInModelForConstraintFunc(t, constraint, state)
	}
	return t.IsInModel(state)
}

func (t *Tool) IsInActions(s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
	if t != nil && t.IsInActionsFunc != nil {
		return t.IsInActionsFunc(t, s1, s2)
	}
	return true, nil
}

func (t *Tool) IsInActionsForConstraint(constraint SemanticNode, s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
	if t != nil && t.IsInActionsForConstraintFn != nil {
		return t.IsInActionsForConstraintFn(t, constraint, s1, s2)
	}
	return t.IsInActions(s1, s2)
}

func (t *Tool) EvalReward(s1 *TLCStateMut, s2 *TLCStateMut, fallback float64) (float64, error) {
	if t != nil && t.EvalRewardFunc != nil {
		return t.EvalRewardFunc(t, s1, s2, fallback)
	}
	return fallback, nil
}

func (t *Tool) HasStateOrActionConstraints() bool {
	return t != nil && t.HasStateOrActionConstraintsFunc != nil && t.HasStateOrActionConstraintsFunc(t)
}

func (t *Tool) Enabled(pred SemanticNode, con *Context, s0 *TLCStateMut, s1 *TLCStateMut) *TLCStateMut {
	if t != nil && t.EnabledFunc != nil {
		return t.EnabledFunc(t, pred, con, s0, s1)
	}
	return s1
}

func (t *Tool) IsValidExpr(expr SemanticNode, ctxt *Context) (bool, error) {
	if t != nil && t.IsValidExprFunc != nil {
		return t.IsValidExprFunc(t, expr, ctxt)
	}
	return true, nil
}

func (t *Tool) IsValidTransition(action *Action, s0 *TLCStateMut, s1 *TLCStateMut) (bool, error) {
	if t != nil && t.IsValidTransitionFunc != nil {
		return t.IsValidTransitionFunc(t, action, s0, s1)
	}
	return true, nil
}

func (t *Tool) IsValidState(action *Action, state *TLCStateMut) (bool, error) {
	if t != nil && t.IsValidStateFunc != nil {
		return t.IsValidStateFunc(t, action, state)
	}
	return true, nil
}

func (t *Tool) IsValidAction(action *Action) (bool, error) {
	if t != nil && t.IsValidActionFunc != nil {
		return t.IsValidActionFunc(t, action)
	}
	return true, nil
}

func (t *Tool) GetState(fp uint64, prev ...any) (*TLCStateInfo, error) {
	if t != nil && t.GetStateFunc != nil {
		return t.GetStateFunc(t, fp, prev...)
	}
	return nil, newTLCError(ECGeneral, "state reconstruction is not implemented")
}

func (t *Tool) HasSymmetry() bool {
	return t != nil && t.HasSymmetryFunc != nil && t.HasSymmetryFunc(t)
}

func (t *Tool) GetInitStateSpec() []*Action {
	return nil
}

func (t *Tool) GetInvariants() []*Action {
	if t == nil {
		return nil
	}
	return t.Invariants
}

func (t *Tool) CheckAssumptions() int {
	if t != nil && t.CheckAssumptionsFunc != nil {
		return t.CheckAssumptionsFunc(t)
	}
	return NoError
}

func (t *Tool) CheckPostCondition() int {
	if t != nil && t.CheckPostConditionFunc != nil {
		return t.CheckPostConditionFunc(t)
	}
	return NoError
}

func (t *Tool) CheckPostConditionWithCounterExample(value Value) int {
	if t != nil && t.CheckPostConditionCEFunc != nil {
		return t.CheckPostConditionCEFunc(t, value)
	}
	return t.CheckPostCondition()
}

func (t *Tool) GetInvNames() []string {
	if t == nil {
		return nil
	}
	return t.InvariantNames
}

func (t *Tool) GetImpliedActNames() []string {
	if t == nil {
		return nil
	}
	return t.ImpliedActNames
}

func (t *Tool) GetRootName() string {
	if t == nil || t.RootName == "" {
		return "Spec"
	}
	return t.RootName
}

func (t *Tool) GetRootFile() string {
	if t == nil {
		return ""
	}
	return t.RootFile
}

func (t *Tool) GetConfigFile() string {
	if t == nil {
		return ""
	}
	return t.ConfigFile
}

func (t *Tool) GetSpecDir() string {
	if t == nil {
		return ""
	}
	return t.SpecDir
}

func (t *Tool) GetImpliedInitNames() []string {
	if t == nil {
		return nil
	}
	return t.ImpliedInitNames
}

func (t *Tool) GetImpliedInits() []*Action {
	if t == nil {
		return nil
	}
	return t.ImpliedInits
}

func (t *Tool) GetImpliedActions() []*Action {
	if t == nil {
		return nil
	}
	return t.ImpliedActions
}

func (t *Tool) LivenessIsTrue() bool {
	if t != nil && t.LivenessIsTrueFunc != nil {
		return t.LivenessIsTrueFunc(t)
	}
	return true
}

func (t *Tool) EvalAliasInfo(current *TLCStateInfo, successor *TLCStateMut, prefix func() []*TLCStateInfo) (*TLCStateInfo, error) {
	if t != nil && t.EvalAliasInfoFunc != nil {
		return t.EvalAliasInfoFunc(t, current, successor, prefix)
	}
	return NewTLCStateInfo(successor), nil
}

func (t *Tool) EvalAliasInfoPair(curState *TLCStateInfo, sucState *TLCStateMut) (*TLCStateInfo, error) {
	if t != nil && t.EvalAliasInfoPairFunc != nil {
		return t.EvalAliasInfoPairFunc(t, curState, sucState)
	}
	return NewTLCStateInfo(sucState), nil
}

func (t *Tool) EvalAlias(curState *TLCStateMut, sucState *TLCStateMut) *TLCStateMut {
	if t != nil && t.EvalAliasFunc != nil {
		return t.EvalAliasFunc(t, curState, sucState)
	}
	return sucState
}

func (t *Tool) NoDebug() *Tool {
	if t != nil && t.NoDebugFunc != nil {
		return t.NoDebugFunc(t)
	}
	return t
}

func (t *Tool) IsDebugger() bool {
	return t != nil && t.IsDebuggerFunc != nil && t.IsDebuggerFunc(t)
}
