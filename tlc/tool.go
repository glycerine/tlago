package tlc

import "sync/atomic"

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
	GetStatesFunc           func() *SetOfStates
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

func (f *StateFunctor) GetStates() *SetOfStates {
	if f != nil && f.GetStatesFunc != nil {
		return f.GetStatesFunc()
	}
	return NewSetOfStates(0)
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
	ID int64

	Mode ToolMode

	Actions            []*Action
	InitStateSpec      []*Action
	NextStateSpec      *Action
	InitStates         []*TLCStateMut
	Invariants         []*Action
	InvariantNames     []string
	Temporals          []*Action
	TemporalNames      []string
	ImpliedTemporals   []*Action
	ImpliedTempNames   []string
	ImpliedInits       []*Action
	ImpliedInitNames   []string
	ImpliedActions     []*Action
	ImpliedActNames    []string
	ViewSpec           SemanticNode
	PostConditionSpecs []*Action
	Assumptions        []SemanticNode
	AssumptionIsAxiom  []bool
	RLReward           SemanticNode
	Periodic           SemanticNode
	ModelConstraints   []SemanticNode
	ActionConstraints  []SemanticNode
	SymmetryPerms      []*MVPerm

	RootName    string
	RootFile    string
	ConfigFile  string
	SpecDir     string
	ModelConfig *ModelConfig
	KnownStates *InsMap[uint64, *TLCStateMut]
	Definitions map[*SymbolNode]any
	DefnsByName map[*UniqueString]any
	CallStack   *CallStack

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

var nextToolID atomic.Int64

func NewTool() *Tool {
	return &Tool{
		ID:          nextToolID.Add(1),
		Mode:        ModeMC,
		RootName:    "Spec",
		ModelConfig: newModelConfig("", false),
		KnownStates: NewInsMap[uint64, *TLCStateMut](),
		Definitions: make(map[*SymbolNode]any),
		DefnsByName: make(map[*UniqueString]any),
	}
}

func (t *Tool) GetID() int64 {
	if t == nil {
		return 0
	}
	return t.ID
}

func NewToolWithModelConfig(config *ModelConfig) *Tool {
	tool := NewTool()
	tool.SetModelConfig(config)
	return tool
}

func (t *Tool) SetModelConfig(config *ModelConfig) *Tool {
	if t == nil {
		return nil
	}
	if config == nil {
		config = newModelConfig(t.ConfigFile, false)
	}
	t.ModelConfig = config
	return t
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
	return append([]*Action(nil), t.Actions...)
}

func (t *Tool) SetActions(actions []*Action) {
	if t != nil {
		t.Actions = append([]*Action(nil), actions...)
	}
}

func (t *Tool) AssignActionIDs() {
	if t == nil {
		return
	}
	id := 0
	for _, action := range t.InitStateSpec {
		if action != nil {
			action.SetID(id)
		}
		id++
	}
	for _, action := range t.Actions {
		if action != nil {
			action.SetID(id)
		}
		id++
	}
}

func (t *Tool) GetInitStates(functor *StateFunctor) error {
	if t != nil && t.GetInitStatesFunc != nil {
		return t.GetInitStatesFunc(t, functor)
	}
	if t == nil || functor == nil {
		return nil
	}
	if len(t.InitStateSpec) != 0 {
		return t.GetInitStatesImpl(functor)
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
	if t == nil {
		return NewEmptyState(), nil
	}
	return t.MakeStateImpl(pred)
}

func (t *Tool) GetNextStates(action *Action, state *TLCStateMut) (*StateVec, error) {
	if t != nil && t.GetNextStatesFunc != nil {
		return t.GetNextStatesFunc(t, action, state)
	}
	if t == nil || action == nil || action.Pred == nil {
		return NewStateVec(0), nil
	}
	return t.GetNextStatesImpl(action, state)
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

func (t *Tool) Eval(expr SemanticNode, args ...any) (value Value, err error) {
	done := t.callStackEnter(expr)
	defer func() {
		if err == nil {
			value = t.callStackToolValue(value)
		}
		done(err)
	}()
	if t != nil && t.EvalFunc != nil {
		return t.EvalFunc(t, expr, args...)
	}
	con, s0, s1, control, cm := parseEvalArgs(args...)
	return t.EvalImpl(expr, con, s0, s1, control, cm)
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
	return t.IsInModelImpl(state)
}

func (t *Tool) IsInModelForConstraint(constraint SemanticNode, state *TLCStateMut) (bool, error) {
	if t != nil && t.IsInModelForConstraintFunc != nil {
		return t.IsInModelForConstraintFunc(t, constraint, state)
	}
	return t.IsInModelForConstraintImpl(constraint, state)
}

func (t *Tool) IsInActions(s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
	if t != nil && t.IsInActionsFunc != nil {
		return t.IsInActionsFunc(t, s1, s2)
	}
	return t.IsInActionsImpl(s1, s2)
}

func (t *Tool) IsInActionsForConstraint(constraint SemanticNode, s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
	if t != nil && t.IsInActionsForConstraintFn != nil {
		return t.IsInActionsForConstraintFn(t, constraint, s1, s2)
	}
	return t.IsInActionsForConstraintImpl(constraint, s1, s2)
}

func (t *Tool) EvalReward(s1 *TLCStateMut, s2 *TLCStateMut, fallback float64) (float64, error) {
	if t != nil && t.EvalRewardFunc != nil {
		return t.EvalRewardFunc(t, s1, s2, fallback)
	}
	return t.EvalRewardImpl(s1, s2, fallback)
}

func (t *Tool) HasStateOrActionConstraints() bool {
	if t == nil {
		return false
	}
	if t.HasStateOrActionConstraintsFunc != nil {
		return t.HasStateOrActionConstraintsFunc(t)
	}
	return len(t.ModelConstraints) > 0 || len(t.ActionConstraints) > 0
}

func (t *Tool) Enabled(pred SemanticNode, con *Context, s0 *TLCStateMut, s1 *TLCStateMut) (state *TLCStateMut) {
	var err error
	done := t.callStackEnter(pred)
	defer func() { done(err) }()
	if t != nil && t.EnabledFunc != nil {
		return t.EnabledFunc(t, pred, con, s0, s1)
	}
	if t == nil {
		return s1
	}
	state, err = t.EnabledImpl(pred, EmptyActionItemList, con, s0, s1, CostModel{})
	if err != nil {
		return nil
	}
	return state
}

func (t *Tool) IsValidExpr(expr SemanticNode, ctxt *Context) (bool, error) {
	if t != nil && t.IsValidExprFunc != nil {
		return t.IsValidExprFunc(t, expr, ctxt)
	}
	return t.IsValidExprImpl(expr, ctxt)
}

func (t *Tool) IsValidTransition(action *Action, s0 *TLCStateMut, s1 *TLCStateMut) (bool, error) {
	if t != nil && t.IsValidTransitionFunc != nil {
		return t.IsValidTransitionFunc(t, action, s0, s1)
	}
	return t.IsValidTransitionImpl(action, s0, s1)
}

func (t *Tool) IsValidState(action *Action, state *TLCStateMut) (bool, error) {
	if t != nil && t.IsValidStateFunc != nil {
		return t.IsValidStateFunc(t, action, state)
	}
	return t.IsValidStateImpl(action, state)
}

func (t *Tool) IsValidAction(action *Action) (bool, error) {
	if t != nil && t.IsValidActionFunc != nil {
		return t.IsValidActionFunc(t, action)
	}
	return t.IsValidActionImpl(action)
}

func (t *Tool) GetState(fp uint64, prev ...any) (*TLCStateInfo, error) {
	if t != nil && t.GetStateFunc != nil {
		return t.GetStateFunc(t, fp, prev...)
	}
	if t == nil || t.KnownStates == nil {
		return nil, newTLCError(ECTLCFailedToRecoverInit, "state fingerprint %d is not in the state registry", fp)
	}
	state, ok := t.KnownStates.Get2(fp)
	if !ok || state == nil {
		return nil, newTLCError(ECTLCFailedToRecoverInit, "state fingerprint %d is not in the state registry", fp)
	}
	info := NewTLCStateInfo(state)
	info.FP = &fp
	if len(prev) > 0 {
		switch predecessor := prev[0].(type) {
		case *TLCStateInfo:
			if predecessor != nil {
				state.SetPredecessor(predecessor.State)
			}
		case *TLCStateMut:
			state.SetPredecessor(predecessor)
		}
	}
	return info, nil
}

func (t *Tool) RememberState(state *TLCStateMut) uint64 {
	if state == nil {
		return 0
	}
	if t.KnownStates == nil {
		t.KnownStates = NewInsMap[uint64, *TLCStateMut]()
	}
	fp := state.FingerPrint()
	t.KnownStates.Set(fp, state)
	return fp
}

func (t *Tool) SetSymmetryPermutations(perms []*MVPerm) *Tool {
	if t == nil {
		return nil
	}
	if len(perms) == 0 {
		t.SymmetryPerms = nil
		return t
	}
	t.SymmetryPerms = make([]*MVPerm, len(perms))
	copy(t.SymmetryPerms, perms)
	return t
}

func (t *Tool) GetSymmetryPerms() []*MVPerm {
	if t == nil || len(t.SymmetryPerms) == 0 {
		return nil
	}
	out := make([]*MVPerm, len(t.SymmetryPerms))
	copy(out, t.SymmetryPerms)
	return out
}

func (t *Tool) HasSymmetry() bool {
	if t == nil {
		return false
	}
	if t.HasSymmetryFunc != nil {
		return t.HasSymmetryFunc(t)
	}
	if t.ModelConfig != nil && t.ModelConfig.GetSymmetry() != "" {
		return true
	}
	return len(t.SymmetryPerms) > 0
}

func (t *Tool) GetInitStateSpec() []*Action {
	if t == nil {
		return nil
	}
	return append([]*Action(nil), t.InitStateSpec...)
}

func (t *Tool) GetSpecActions() []*Action {
	if t == nil {
		return nil
	}
	out := make([]*Action, 0, len(t.InitStateSpec)+len(t.Actions))
	out = append(out, t.InitStateSpec...)
	out = append(out, t.Actions...)
	return out
}

func (t *Tool) GetInvariants() []*Action {
	if t == nil {
		return nil
	}
	return t.Invariants
}

func (t *Tool) GetTemporals() []*Action {
	if t == nil {
		return nil
	}
	return t.Temporals
}

func (t *Tool) GetTemporalNames() []string {
	if t == nil {
		return nil
	}
	return t.TemporalNames
}

func (t *Tool) GetImpliedTemporals() []*Action {
	if t == nil {
		return nil
	}
	return t.ImpliedTemporals
}

func (t *Tool) GetImpliedTemporalNames() []string {
	if t == nil {
		return nil
	}
	return t.ImpliedTempNames
}

func (t *Tool) CheckAssumptions() int {
	if t != nil && t.CheckAssumptionsFunc != nil {
		return t.CheckAssumptionsFunc(t)
	}
	return t.CheckAssumptionsImpl()
}

func (t *Tool) CheckPostCondition() int {
	if t != nil && t.CheckPostConditionFunc != nil {
		return t.CheckPostConditionFunc(t)
	}
	return t.CheckPostConditionImpl(EmptyContext)
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

func (t *Tool) GetModelConfig() *ModelConfig {
	if t == nil {
		return nil
	}
	if t.ModelConfig == nil {
		t.ModelConfig = newModelConfig(t.ConfigFile, false)
	}
	return t.ModelConfig
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

func (t *Tool) GetNextStateSpec() *Action {
	if t == nil {
		return nil
	}
	return t.NextStateSpec
}

func (t *Tool) GetViewSpec() SemanticNode {
	if t == nil {
		return nil
	}
	return t.ViewSpec
}

func (t *Tool) GetPostConditionSpecs() []*Action {
	if t == nil {
		return nil
	}
	return append([]*Action(nil), t.PostConditionSpecs...)
}

func (t *Tool) GetModelConstraints() []SemanticNode {
	if t == nil {
		return nil
	}
	return append([]SemanticNode(nil), t.ModelConstraints...)
}

func (t *Tool) GetActionConstraints() []SemanticNode {
	if t == nil {
		return nil
	}
	return append([]SemanticNode(nil), t.ActionConstraints...)
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
