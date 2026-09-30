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
	return nil, newTLCError(ECGeneral, "IStateFunctor.AddElement is unsupported")
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

func (f *NextStateFunctor) AddElement(state *TLCStateMut) (any, error) {
	return nil, newTLCError(ECGeneral, "INextStateFunctor.AddElement is unsupported")
}

func (f *NextStateFunctor) AddNextElement(curState *TLCStateMut, action *Action, succState *TLCStateMut) (any, error) {
	if f != nil && f.AddNextElementFunc != nil {
		return f.AddNextElementFunc(curState, action, succState)
	}
	return nil, newTLCError(ECGeneral, "INextStateFunctor.AddElement is unsupported")
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
	ConfigErrors       []*ConfigError
	SymmetryPerms      []*MVPerm

	RootName          string
	RootFile          string
	ConfigFile        string
	SpecDir           string
	ModelConfig       *ModelConfig
	SpecProcessor     *SpecProcessor
	ModuleFiles       []string
	CounterExampleDef *OpDefNode
	AliasSpec         SemanticNode
	KnownStates       *InsMap[uint64, *TLCStateMut]
	Definitions       map[*SymbolNode]any
	DefnsByName       map[*UniqueString]any
	CallStack         *CallStack

	actionsPrepared bool

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
	tool := &Tool{
		ID:          nextToolID.Add(1),
		Mode:        ModeMC,
		RootName:    "Spec",
		ModelConfig: newModelConfig("", false),
		KnownStates: NewInsMap[uint64, *TLCStateMut](),
		Definitions: make(map[*SymbolNode]any),
		DefnsByName: make(map[*UniqueString]any),
	}
	tool.InstallStandardDefinitions()
	return tool
}

func (t *Tool) GetID() int64 {
	if t == nil {
		return 0
	}
	return t.ID
}

func (t *Tool) GetId() int {
	return int(t.GetID())
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
	if err := t.ensureActionsPrepared(); err != nil {
		panic(err)
	}
	return append([]*Action(nil), t.Actions...)
}

func (t *Tool) SetActions(actions []*Action) {
	if t != nil {
		t.Actions = append([]*Action(nil), actions...)
		t.actionsPrepared = true
	}
}

func (t *Tool) SetNextStateSpec(action *Action) {
	if t == nil {
		return
	}
	t.NextStateSpec = action
	t.Actions = nil
	t.actionsPrepared = false
}

func (t *Tool) AssignActionIDs() {
	if t == nil {
		return
	}
	t.assignActionIDs(t.GetActions())
}

func (t *Tool) assignActionIDs(actions []*Action) {
	id := 0
	for _, action := range t.InitStateSpec {
		if action != nil {
			action.SetID(id)
		}
		id++
	}
	for _, action := range actions {
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
	if t == nil {
		return nil, newTLCError(ECTLCFailedToRecoverInit, "state fingerprint %d cannot be recovered without a tool", fp)
	}
	if len(prev) != 0 {
		switch predecessor := prev[0].(type) {
		case *TLCStateInfo:
			if fallback := t.getKnownState(fp, predecessor); fallback != nil {
				fallback.StateNumber = predecessor.StateNumber + 1
				return fallback, nil
			}
			info, err := t.GetStateAfter(fp, predecessor.State)
			if err == nil && info != nil {
				info.StateNumber = predecessor.StateNumber + 1
				info.FP = &fp
				return info, nil
			}
			if err != nil {
				return nil, err
			}
			return nil, newTLCError(ECTLCFailedToRecoverNext, "successor fingerprint %d could not be regenerated", fp)
		case *TLCStateMut:
			if fallback := t.getKnownState(fp, predecessor); fallback != nil {
				return fallback, nil
			}
			info, err := t.GetStateAfter(fp, predecessor)
			if err == nil && info != nil {
				return info, nil
			}
			if err != nil {
				return nil, err
			}
			return nil, newTLCError(ECTLCFailedToRecoverNext, "successor fingerprint %d could not be regenerated", fp)
		}
	}
	info, err := t.GetInitState(fp)
	if err == nil && info != nil {
		return info, nil
	}
	if fallback := t.getKnownState(fp); fallback != nil {
		return fallback, nil
	}
	if err != nil {
		return nil, err
	}
	return nil, newTLCError(ECTLCFailedToRecoverInit, "initial state fingerprint %d could not be regenerated", fp)
}

func (t *Tool) GetInitState(fp uint64) (*TLCStateInfo, error) {
	if t == nil {
		return nil, nil
	}
	var found *TLCStateMut
	functor := NewStateFunctor(func(state *TLCStateMut) (any, error) {
		if state != nil && found == nil && state.FingerPrint() == fp {
			found = state
		}
		return nil, nil
	})
	if err := t.GetInitStates(functor); err != nil {
		return nil, err
	}
	if found == nil {
		return nil, nil
	}
	info := NewTLCStateInfo(found)
	info.FP = &fp
	return info, nil
}

func (t *Tool) GetStateAfter(fp uint64, predecessor *TLCStateMut) (*TLCStateInfo, error) {
	if t == nil || predecessor == nil {
		return nil, nil
	}
	restoreRandomState := PushRandomEnumerableState(predecessor)
	defer restoreRandomState()
	restoreCurrentState := PushCurrentState(predecessor)
	defer restoreCurrentState()
	for _, action := range t.GetActions() {
		nextStates, err := t.GetNextStates(action, predecessor)
		if err != nil {
			return nil, err
		}
		for i := 0; nextStates != nil && i < nextStates.Size(); i++ {
			state := nextStates.At(i)
			if state != nil && state.FingerPrint() == fp {
				state.SetPredecessor(predecessor)
				return NewTLCStateInfo(state, action), nil
			}
		}
	}
	return nil, nil
}

func (t *Tool) GetStateForTransition(successor *TLCStateMut, predecessor *TLCStateMut) (*TLCStateInfo, error) {
	if t == nil || successor == nil || predecessor == nil {
		return nil, nil
	}
	restoreRandomState := PushRandomEnumerableState(predecessor)
	defer restoreRandomState()
	restoreCurrentState := PushCurrentState(predecessor)
	defer restoreCurrentState()
	for _, action := range t.GetActions() {
		nextStates, err := t.GetNextStates(action, predecessor)
		if err != nil {
			return nil, err
		}
		for i := 0; nextStates != nil && i < nextStates.Size(); i++ {
			state := nextStates.At(i)
			if state != nil && successor.Equal(state) {
				state.SetPredecessor(predecessor)
				return NewTLCStateInfo(state, action), nil
			}
		}
	}
	return nil, nil
}

func (t *Tool) getKnownState(fp uint64, prev ...any) *TLCStateInfo {
	if t == nil || t.KnownStates == nil {
		return nil
	}
	state, ok := t.KnownStates.Get2(fp)
	if !ok || state == nil {
		return nil
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
	return info
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
	actions := t.GetActions()
	out := make([]*Action, 0, len(t.InitStateSpec)+len(actions))
	out = append(out, t.InitStateSpec...)
	out = append(out, actions...)
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
	def := t.GetCounterExampleDef()
	if t == nil || def == nil || def.Symbol == nil {
		return t.CheckPostCondition()
	}
	if value == nil {
		value = NewEmptyCounterExample()
	}
	return t.CheckPostConditionImpl(EmptyContext.Cons(def.Symbol, value))
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

func (t *Tool) GetSpecProcessor() *SpecProcessor {
	if t == nil {
		return nil
	}
	return t.SpecProcessor
}

func (t *Tool) GetModuleFiles() []string {
	if t == nil {
		return nil
	}
	return append([]string(nil), t.ModuleFiles...)
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

func (t *Tool) HasAlias() bool {
	return t != nil && t.AliasSpec != nil
}

func (t *Tool) GetAliasSpec() SemanticNode {
	if t == nil {
		return nil
	}
	return t.AliasSpec
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

func (t *Tool) GetAssumptions() []SemanticNode {
	if t == nil {
		return nil
	}
	return append([]SemanticNode(nil), t.Assumptions...)
}

func (t *Tool) GetAssumptionIsAxiom() []bool {
	if t == nil {
		return nil
	}
	return append([]bool(nil), t.AssumptionIsAxiom...)
}

func (t *Tool) GetCounterExampleDef() *OpDefNode {
	if t == nil {
		return nil
	}
	if t.CounterExampleDef != nil {
		return t.CounterExampleDef
	}
	val := t.Lookup(NewSymbolNode("CounterExample"), EmptyContext, EmptyState, false)
	if ev, ok := val.(*EvaluatingValue); ok {
		return ev.OpDef
	}
	return nil
}

func (t *Tool) GetTraceDef() *OpDefNode {
	if t == nil {
		return nil
	}
	val := t.Lookup(NewSymbolNode("Trace"), EmptyContext, EmptyState, false)
	if ev, ok := val.(*EvaluatingValue); ok {
		return ev.OpDef
	}
	return nil
}

func (t *Tool) LivenessIsTrue() bool {
	if t != nil && t.LivenessIsTrueFunc != nil {
		return t.LivenessIsTrueFunc(t)
	}
	return t == nil || len(t.GetImpliedTemporals()) == 0
}

func (t *Tool) EvalAliasInfo(current *TLCStateInfo, successor *TLCStateMut, prefix func() []*TLCStateInfo) (*TLCStateInfo, error) {
	if t != nil && t.EvalAliasInfoFunc != nil {
		var restore func()
		if current != nil {
			restore = PushCurrentState(current.State)
		} else {
			restore = PushCurrentState(nil)
		}
		defer restore()
		return t.EvalAliasInfoFunc(t, current, successor, prefix)
	}
	if t == nil || !t.HasAlias() || current == nil || current.State == nil {
		return current, nil
	}
	ctxt := EmptyContext
	if prefix != nil {
		if traceDef := t.GetTraceDef(); traceDef != nil && traceDef.Symbol != nil {
			ctxt = ctxt.Cons(traceDef.Symbol, traceTupleFromStateInfos(prefix()))
		}
	}
	restore := PushCurrentState(current.State)
	defer restore()
	alias, err := t.evalAliasState(current.State, successor, ctxt)
	if err != nil {
		return AliasTLCStateInfo(aliasEvaluationErrorState(current.State, err), current), nil
	}
	if alias != nil {
		return AliasTLCStateInfo(alias, current), nil
	}
	return current, nil
}

func (t *Tool) EvalAliasInfoPair(curState *TLCStateInfo, sucState *TLCStateMut) (*TLCStateInfo, error) {
	if t != nil && t.EvalAliasInfoPairFunc != nil {
		return t.EvalAliasInfoPairFunc(t, curState, sucState)
	}
	return t.EvalAliasInfo(curState, sucState, nil)
}

func (t *Tool) EvalAliasInfoPrefix(current *TLCStateInfo, successor *TLCStateMut, prefix []*TLCStateInfo) (*TLCStateInfo, error) {
	return t.EvalAliasInfo(current, successor, func() []*TLCStateInfo {
		return append([]*TLCStateInfo(nil), prefix...)
	})
}

func (t *Tool) EvalAliasInfoPrefixSuffix(current *TLCStateInfo, successor *TLCStateMut, prefix []*TLCStateInfo, suffix ...*TLCStateInfo) (*TLCStateInfo, error) {
	return t.EvalAliasInfo(current, successor, func() []*TLCStateInfo {
		out := make([]*TLCStateInfo, 0, len(prefix)+len(suffix))
		out = append(out, prefix...)
		out = append(out, suffix...)
		return out
	})
}

func (t *Tool) EvalAlias(curState *TLCStateMut, sucState *TLCStateMut) *TLCStateMut {
	if t != nil && t.EvalAliasFunc != nil {
		restore := PushCurrentState(curState)
		defer restore()
		return t.EvalAliasFunc(t, curState, sucState)
	}
	if t == nil || !t.HasAlias() || curState == nil {
		return curState
	}
	restore := PushCurrentState(curState)
	defer restore()
	alias, err := t.evalAliasState(curState, sucState, EmptyContext)
	if err != nil {
		return aliasEvaluationErrorState(curState, err)
	}
	if alias != nil {
		return alias
	}
	return curState
}

func (t *Tool) evalAliasState(current *TLCStateMut, successor *TLCStateMut, ctxt *Context) (*TLCStateMut, error) {
	if t == nil || t.AliasSpec == nil {
		return current, nil
	}
	if successor == nil {
		successor = EmptyState
	}
	value, err := t.Eval(t.AliasSpec, ctxt, current, successor, EvalClear, CostModel{})
	if err != nil {
		return nil, err
	}
	rcd := asRecordValue(value)
	if rcd == nil {
		return nil, nil
	}
	alias := rcd.ToState()
	if alias != nil && current != nil {
		alias.level = current.level
		alias.pred = current.pred
		alias.action = current.action
	}
	return alias, nil
}

func aliasEvaluationErrorState(current *TLCStateMut, err error) *TLCStateMut {
	record := EmptyRecord
	if current != nil {
		record = NewRecordValueFromInsMap(current.Values())
	}
	names := append([]*UniqueString(nil), record.Names...)
	values := append([]Value(nil), record.Values...)
	names = append(names, UniqueStringOf("_ALIASEvalError"))
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	values = append(values, NewStringValue(msg))
	state := NewRecordValue(names, values, false).ToState()
	if state != nil && current != nil {
		state.level = current.level
		state.pred = current.pred
		state.action = current.action
	}
	return state
}

func traceTupleFromStateInfos(infos []*TLCStateInfo) Value {
	values := make([]Value, 0, len(infos))
	for _, info := range infos {
		if info == nil || info.State == nil {
			continue
		}
		values = append(values, NewRecordValueFromInsMap(info.State.Values()))
	}
	return NewTupleValue(values)
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
