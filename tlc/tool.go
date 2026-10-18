package tlc

import (
	"fmt"
	"os"
	"sync"
)

const toolStateMutExtProperty = "tlc2.tool.impl.Tool.TLCStateMutExt"

type ToolMode int

const (
	ModeMC ToolMode = iota
	ModeMCDFS
	ModeSimulation
	ModeDebugger
	ModeExecutor
)

// String returns Tool.Mode's Java enum name. ModeMCDFS is a native execution
// distinction folded beneath Java's model-checking mode.
func (m ToolMode) String() string {
	switch m {
	case ModeSimulation:
		return "Simulation"
	case ModeMC:
		return "MC"
	case ModeDebugger:
		return "MC_DEBUG"
	case ModeExecutor:
		return "Executor"
	case ModeMCDFS:
		return "MC_DFS"
	default:
		return fmt.Sprintf("Tool.Mode(%d)", m)
	}
}

// Ordinal returns the source Tool.Mode ordinal. The native DFID distinction is
// not a Java enum value and returns -1.
func (m ToolMode) Ordinal() int {
	switch m {
	case ModeSimulation:
		return 0
	case ModeMC:
		return 1
	case ModeDebugger:
		return 2
	case ModeExecutor:
		return 3
	default:
		return -1
	}
}

type StateFunctor struct {
	AddElementFunc          func(*TLCStateMut) (any, error)
	SetElementFunc          func(*TLCStateMut) (any, error)
	HasStatesFunc           func() bool
	GetStatesFunc           func() *SetOfStates
	AddUnsatisfiedStateFunc func(*TLCStateMut, SemanticNode, *Context) *TLCStateMut
	debugWrapped            bool
}

func NewStateFunctor(addElement func(*TLCStateMut) (any, error)) *StateFunctor {
	return &StateFunctor{AddElementFunc: addElement}
}

func (f *StateFunctor) AddElement(state *TLCStateMut) (any, error) {
	if f == nil {
		return nil, NewNullPointerException()
	}
	if f.AddElementFunc != nil {
		return f.AddElementFunc(state)
	}
	return nil, newTLCError(ECGeneral, "IStateFunctor.AddElement is unsupported")
}

func (f *StateFunctor) SetElement(state *TLCStateMut) (any, error) {
	if f == nil {
		return nil, NewNullPointerException()
	}
	if f.SetElementFunc != nil {
		return f.SetElementFunc(state)
	}
	return nil, NewUnsupportedOperationException()
}

func (f *StateFunctor) HasStates() bool {
	if f == nil {
		panic(NewNullPointerException())
	}
	if f.HasStatesFunc != nil {
		return f.HasStatesFunc()
	}
	panic(NewUnsupportedOperationException())
}

func (f *StateFunctor) GetStates() *SetOfStates {
	if f == nil {
		panic(NewNullPointerException())
	}
	if f.GetStatesFunc != nil {
		return f.GetStatesFunc()
	}
	return NewSetOfStates(0)
}

func (f *StateFunctor) AddUnsatisfiedState(state *TLCStateMut, pred SemanticNode, con *Context) *TLCStateMut {
	if f == nil {
		panic(NewNullPointerException())
	}
	if f.AddUnsatisfiedStateFunc != nil {
		return f.AddUnsatisfiedStateFunc(state, pred, con)
	}
	return state
}

type NextStateFunctor struct {
	StateFunctor
	AddNextElementFunc         func(*TLCStateMut, *Action, *TLCStateMut) (any, error)
	IncrementStatesGeneratedFn func(int64)
	HaltFunc                   func() bool
	ShouldHaltFunc             func() bool
	AddUnsatisfiedNextStateFn  func(*TLCStateMut, *Action, *TLCStateMut, SemanticNode, *Context) *TLCStateMut
	debugWrapped               bool
}

func NewNextStateFunctor(addNextElement func(*TLCStateMut, *Action, *TLCStateMut) (any, error)) *NextStateFunctor {
	return &NextStateFunctor{AddNextElementFunc: addNextElement}
}

func (f *NextStateFunctor) AddElement(state *TLCStateMut) (any, error) {
	if f == nil {
		return nil, NewNullPointerException()
	}
	if f.StateFunctor.AddElementFunc != nil {
		return f.StateFunctor.AddElementFunc(state)
	}
	return nil, NewUnsupportedOperationException()
}

func (f *NextStateFunctor) AddNextElement(curState *TLCStateMut, action *Action, succState *TLCStateMut) (any, error) {
	if f == nil {
		return nil, NewNullPointerException()
	}
	if f.AddNextElementFunc != nil {
		return f.AddNextElementFunc(curState, action, succState)
	}
	return nil, newTLCError(ECGeneral, "INextStateFunctor.AddElement is unsupported")
}

func (f *NextStateFunctor) IncrementStatesGenerated(count int64) {
	if f == nil {
		panic(NewNullPointerException())
	}
	if f.IncrementStatesGeneratedFn != nil {
		f.IncrementStatesGeneratedFn(count)
	}
}

func (f *NextStateFunctor) Halt() bool {
	if f == nil {
		panic(NewNullPointerException())
	}
	return f.HaltFunc != nil && f.HaltFunc()
}

func (f *NextStateFunctor) ShouldHalt() bool {
	return f != nil && f.ShouldHaltFunc != nil && f.ShouldHaltFunc()
}

func (f *NextStateFunctor) AddUnsatisfiedNextState(curState *TLCStateMut, action *Action, succState *TLCStateMut, pred SemanticNode, con *Context) *TLCStateMut {
	if f == nil {
		panic(NewNullPointerException())
	}
	if f.AddUnsatisfiedNextStateFn != nil {
		return f.AddUnsatisfiedNextStateFn(curState, action, succState, pred, con)
	}
	return succState
}

type Tool struct {
	ID         int32
	idCaptured bool
	idOwner    int32

	Mode         ToolMode
	modeCaptured bool
	modeOwner    ToolMode

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

	RootName              string
	RootFile              string
	ConfigFile            string
	SpecDir               string
	specIdentityCaptured  bool
	rootNameOwner         string
	rootFileOwner         string
	configFileOwner       string
	specDirOwner          string
	ModelConfig           *ModelConfig
	modelConfigCaptured   bool
	modelConfigOwner      *ModelConfig
	SpecProcessor         *SpecProcessor
	specProcessorCaptured bool
	specProcessorOwner    *SpecProcessor
	Resolver              FilenameToStream
	resolverCaptured      bool
	resolverOwner         FilenameToStream
	ModuleFiles           []string
	DistributedFiles      *DistributedServerFiles
	CounterExampleDef     *OpDefNode
	TraceDef              *OpDefNode
	AliasSpec             SemanticNode
	Debugger              *TLCDebugger
	DebugPort             int
	DebugSuspend          bool
	DebugHalt             bool
	DebugFastTool         *Tool
	DebugEvalMode         DebugEvalMode
	debugOwnersCaptured   bool
	debuggerOwner         *TLCDebugger
	debugFastToolOwner    *Tool
	Definitions           map[*SymbolNode]any
	DefnsByName           map[*UniqueString]any
	CallStack             *CallStack

	actionsPrepared bool
	actionsCaptured bool
	actionsOwner    []*Action

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
	EnabledFunc                     func(*Tool, SemanticNode, *Context, *TLCStateMut, *TLCStateMut) (*TLCStateMut, error)
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
	ParseDebuggerExpressionFunc     func(*Tool, *ModuleNode, SourceLocation, string) (*OpDefNode, error)
	NoDebugFunc                     func(*Tool) *Tool
	IsDebuggerFunc                  func(*Tool) bool
}

func (t *Tool) ownedDebugger() *TLCDebugger {
	if t != nil && t.debugOwnersCaptured {
		return t.debuggerOwner
	}
	if t == nil {
		return nil
	}
	return t.Debugger
}

func (t *Tool) ownedDebugFastTool() *Tool {
	if t != nil && t.debugOwnersCaptured {
		return t.debugFastToolOwner
	}
	if t == nil {
		return nil
	}
	return t.DebugFastTool
}

func (t *Tool) captureDebugOwners() {
	if t == nil || t.debugOwnersCaptured {
		return
	}
	t.debugOwnersCaptured = true
	t.debuggerOwner = t.Debugger
	t.debugFastToolOwner = t.DebugFastTool
}

// Spec's static toolId is allocated once, shared by all Tool subclasses.
var specToolIdentity struct {
	sync.Once
	id int32
}

var specCoverageStatics struct {
	sync.Once
	enabled bool
}

func specToolID() int32 {
	specToolIdentity.Do(func() { specToolIdentity.id = GetSemanticToolID() })
	return specToolIdentity.id
}

func specCoverageEnabled() bool {
	specCoverageStatics.Do(func() {
		specCoverageStatics.enabled = CoverageEnabled()
	})
	return specCoverageStatics.enabled
}

// InitializeSpecStatics represents a fresh Java Spec class initialization.
// It requires an idle runtime, like the other class-static reset helpers.
func InitializeSpecStatics() {
	specCoverageStatics.Once = sync.Once{}
	specCoverageStatics.enabled = false
}

var toolStaticProperties struct {
	sync.Once
	probabilistic bool
}

// InitializeToolStatics represents a fresh Java Tool class initialization.
// Ordinary tool construction retains the first probabilistic setting.
func InitializeToolStatics() {
	toolStaticProperties.Once = sync.Once{}
	toolStaticProperties.probabilistic = false
}

// InitializeToolProperties represents Tool's class initialization, after
// command/package properties are installed and before config/tool loading.
func InitializeToolProperties() {
	toolStaticProperties.Do(func() {
		toolStaticProperties.probabilistic = Globals.Probabilistic
		if value, ok := tlcLookupSystemProperty(toolProbabilisticProperty); ok {
			toolStaticProperties.probabilistic = javaBooleanProperty(value)
		}
	})
}

func IsProbabilistic() bool {
	InitializeToolProperties()
	return toolStaticProperties.probabilistic
}

// ToolIsProbabilistic retains the earlier native spelling.
func ToolIsProbabilistic() bool { return IsProbabilistic() }

func NewTool() *Tool {
	_ = specCoverageEnabled()
	InitializeToolProperties()
	tool := &Tool{
		ID:          specToolID(),
		Mode:        ModeMC,
		RootName:    "Spec",
		ModelConfig: newModelConfig("", false),
		Definitions: make(map[*SymbolNode]any),
		DefnsByName: make(map[*UniqueString]any),
	}
	tool.idCaptured = true
	tool.idOwner = tool.ID
	tool.modeCaptured = true
	tool.modeOwner = tool.Mode
	tool.specIdentityCaptured = true
	tool.rootNameOwner = tool.RootName
	tool.rootFileOwner = tool.RootFile
	tool.configFileOwner = tool.ConfigFile
	tool.specDirOwner = tool.SpecDir
	tool.modelConfigCaptured = true
	tool.modelConfigOwner = tool.ModelConfig
	tool.specProcessorCaptured = true
	tool.specProcessorOwner = tool.SpecProcessor
	tool.resolverCaptured = true
	tool.resolverOwner = tool.Resolver
	tool.InstallStandardDefinitions()
	return tool
}

func (t *Tool) GetID() int32 {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.idCaptured {
		return t.idOwner
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

func (t *Tool) SetSpecIdentity(rootName, rootFile, configFile, specDir string) *Tool {
	if t == nil {
		return nil
	}
	t.RootName = rootName
	t.RootFile = rootFile
	t.ConfigFile = configFile
	t.SpecDir = specDir
	t.specIdentityCaptured = true
	t.rootNameOwner = rootName
	t.rootFileOwner = rootFile
	t.configFileOwner = configFile
	t.specDirOwner = specDir
	return t
}

func (t *Tool) ownedSpecIdentity() (string, string, string, string) {
	if t != nil && t.specIdentityCaptured {
		return t.rootNameOwner, t.rootFileOwner, t.configFileOwner, t.specDirOwner
	}
	if t == nil {
		return "", "", "", ""
	}
	return t.RootName, t.RootFile, t.ConfigFile, t.SpecDir
}

func (t *Tool) SetModelConfig(config *ModelConfig) *Tool {
	if t == nil {
		return nil
	}
	if config == nil {
		config = newModelConfig(t.GetConfigFile(), false)
	}
	t.ModelConfig = config
	t.modelConfigCaptured = true
	t.modelConfigOwner = config
	return t
}

func (t *Tool) ownedModelConfig() *ModelConfig {
	if t != nil && t.modelConfigCaptured {
		return t.modelConfigOwner
	}
	if t == nil {
		return nil
	}
	return t.ModelConfig
}

func (t *Tool) SetSpecProcessor(processor *SpecProcessor) *Tool {
	if t == nil {
		return nil
	}
	t.SpecProcessor = processor
	t.specProcessorCaptured = true
	t.specProcessorOwner = processor
	return t
}

func (t *Tool) ownedSpecProcessor() *SpecProcessor {
	if t != nil && t.specProcessorCaptured {
		return t.specProcessorOwner
	}
	if t == nil {
		return nil
	}
	return t.SpecProcessor
}

func (t *Tool) SetResolver(resolver FilenameToStream) *Tool {
	if t == nil {
		return nil
	}
	t.Resolver = resolver
	t.resolverCaptured = true
	t.resolverOwner = resolver
	return t
}

func (t *Tool) ownedResolver() FilenameToStream {
	if t != nil && t.resolverCaptured {
		return t.resolverOwner
	}
	if t == nil {
		return nil
	}
	return t.Resolver
}

func (t *Tool) GetMode() ToolMode {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.modeCaptured {
		return t.modeOwner
	}
	return t.Mode
}

func (t *Tool) SetMode(mode ToolMode) *Tool {
	if t == nil {
		return nil
	}
	t.Mode = mode
	t.modeCaptured = true
	t.modeOwner = mode
	if stateTool == t {
		SetTLCStateTool(t)
	}
	return t
}

func (t *Tool) usesExtendedStateMetadata() bool {
	return usesExtendedStateMetadataForMode(t.GetMode())
}

func usesExtendedStateMetadataForMode(mode ToolMode) bool {
	if value, ok := tlcLookupSystemProperty(toolStateMutExtProperty); ok && javaBooleanProperty(value) {
		return true
	}
	if javaBooleanProperty(os.Getenv("TLAGO_TLC_STATE_MUT_EXT")) {
		return true
	}
	switch mode {
	case ModeSimulation, ModeDebugger, ModeExecutor:
		return true
	default:
		return false
	}
}

func (t *Tool) GetActions() []*Action {
	if t == nil {
		panic(NewNullPointerException())
	}
	if err := t.ensureActionsPrepared(); err != nil {
		panic(err)
	}
	return t.ownedActions()
}

func (t *Tool) SetActions(actions []*Action) {
	if t != nil {
		t.Actions = append([]*Action(nil), actions...)
		t.actionsPrepared = true
		t.captureActions(t.Actions)
	}
}

func (t *Tool) ownedActions() []*Action {
	if t != nil && t.actionsCaptured {
		return t.actionsOwner
	}
	if t == nil {
		return nil
	}
	return t.Actions
}

func (t *Tool) captureActions(actions []*Action) {
	if t == nil || t.actionsCaptured {
		return
	}
	t.actionsCaptured = true
	t.actionsOwner = actions
}

func (t *Tool) SetNextStateSpec(action *Action) {
	if t == nil {
		return
	}
	t.NextStateSpec = action
	if t.ownedSpecProcessor() != nil {
		t.ownedSpecProcessor().NextPred = action
	}
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
	init := t.GetInitStateSpec()
	for i := 0; i < init.Size(); i++ {
		action := init.ElementAt(i)
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
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.GetInitStatesFunc != nil {
		return t.GetInitStatesFunc(t, functor)
	}
	if t.ownedSpecProcessor() != nil || t.GetInitStateSpec().Size() != 0 {
		return t.GetInitStatesImpl(functor)
	}
	for _, state := range t.InitStates {
		if _, err := functor.AddElement(state); err != nil {
			return err
		}
	}
	return nil
}

// GetInitStateVec is the no-argument source overload of GetInitStates. Go uses
// a distinct name because the two Java overloads have different return types.
func (t *Tool) GetInitStateVec() (*StateVec, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	states := NewStateVec(0)
	err := t.GetInitStates(NewStateFunctor(states.AddElement))
	return states, err
}

func (t *Tool) MakeState(pred SemanticNode) (*TLCStateMut, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.MakeStateFunc != nil {
		return t.MakeStateFunc(t, pred)
	}
	return t.MakeStateImpl(pred)
}

func (t *Tool) GetNextStates(action *Action, state *TLCStateMut) (*StateVec, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.GetNextStatesFunc != nil {
		return t.GetNextStatesFunc(t, action, state)
	}
	if action == nil {
		panic(NewNullPointerException())
	}
	return t.getNextStateVec(action, state)
}

func (t *Tool) GetNextStatesWithContext(action *Action, context *Context, state *TLCStateMut) (*StateVec, error) {
	if t == nil || action == nil {
		panic(NewNullPointerException())
	}
	return t.getNextStatesWithContext(action, context, state)
}

func (t *Tool) GetNextStatesWithFunctor(functor *NextStateFunctor, state *TLCStateMut) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.GetNextStatesWithFunctorFn != nil {
		return t.GetNextStatesWithFunctorFn(t, functor, state)
	}
	for _, action := range t.GetActions() {
		if _, err := t.GetNextStatesForAction(functor, state, action); err != nil {
			return true, err
		}
	}
	return false, nil
}

func (t *Tool) GetNextStatesForAction(functor *NextStateFunctor, state *TLCStateMut, action *Action) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.GetNextStatesForActionFunc != nil {
		return t.GetNextStatesForActionFunc(t, functor, state, action)
	}
	if action == nil || action.ownedPred() == nil {
		panic(NewNullPointerException())
	}
	s1 := NewEmptyState().SetPredecessor(state).SetAction(action)
	if _, err := t.GetNextStatesForPredicate(action, action.ownedPred(), EmptyActionItemList, action.ownedContext(), state, s1, functor, action.CM); err != nil {
		return true, err
	}
	return false, nil
}

func (t *Tool) Eval(expr SemanticNode, args ...any) (value Value, err error) {
	con, s0, s1, control, cm := parseEvalArgs(args...)
	if t == nil {
		panic(NewNullPointerException())
	}
	done := t.callStackEnter(expr)
	defer func() {
		if failure := recover(); failure != nil {
			switch failure := failure.(type) {
			case *TLCError, *EvalException, *FingerprintException:
				value, err = nil, failure.(error)
			default:
				done(nil)
				panic(failure)
			}
		}
		if err == nil {
			value = t.callStackToolValue(value)
		}
		done(err)
	}()
	if t != nil && t.EvalFunc != nil {
		if len(args) >= 4 {
			normalized := append([]any(nil), args...)
			normalized[3] = control
			args = normalized
		}
		return t.EvalFunc(t, expr, args...)
	}
	return t.EvalImpl(expr, con, s0, s1, control, cm)
}

func (t *Tool) IsGoodState(state *TLCStateMut) bool {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.IsGoodStateFunc != nil {
		return t.IsGoodStateFunc(t, state)
	}
	if state == nil {
		panic(NewNullPointerException())
	}
	return state.AllAssigned()
}

func (t *Tool) IsInModel(state *TLCStateMut) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.IsInModelFunc != nil {
		return t.IsInModelFunc(t, state)
	}
	return t.IsInModelImpl(state)
}

func (t *Tool) IsInModelForConstraint(constraint SemanticNode, state *TLCStateMut) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.IsInModelForConstraintFunc != nil {
		return t.IsInModelForConstraintFunc(t, constraint, state)
	}
	return t.IsInModelForConstraintImpl(constraint, state)
}

func (t *Tool) IsInActions(s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.IsInActionsFunc != nil {
		return t.IsInActionsFunc(t, s1, s2)
	}
	return t.IsInActionsImpl(s1, s2)
}

func (t *Tool) IsInActionsForConstraint(constraint SemanticNode, s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.IsInActionsForConstraintFn != nil {
		return t.IsInActionsForConstraintFn(t, constraint, s1, s2)
	}
	return t.IsInActionsForConstraintImpl(constraint, s1, s2)
}

func (t *Tool) EvalReward(s1 *TLCStateMut, s2 *TLCStateMut, fallback float64) (float64, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.EvalRewardFunc != nil {
		return t.EvalRewardFunc(t, s1, s2, fallback)
	}
	return t.EvalRewardImpl(s1, s2, fallback)
}

func (t *Tool) HasStateOrActionConstraints() bool {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.HasStateOrActionConstraintsFunc != nil {
		return t.HasStateOrActionConstraintsFunc(t)
	}
	return len(t.requireConstraintArray(t.GetModelConstraints())) > 0 || len(t.requireConstraintArray(t.GetActionConstraints())) > 0
}

func (t *Tool) Enabled(pred SemanticNode, con *Context, s0 *TLCStateMut, s1 *TLCStateMut) (state *TLCStateMut, err error) {
	done := t.callStackEnter(pred)
	defer func() { done(err) }()
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.EnabledFunc != nil {
		return t.EnabledFunc(t, pred, con, s0, s1)
	}
	return t.EnabledWithActionItems(pred, EmptyActionItemList, con, s0, s1)
}

// EnabledWithActionItem is the source overload that starts with one pending
// action item.
func (t *Tool) EnabledWithActionItem(pred SemanticNode, con *Context, s0 *TLCStateMut, s1 *TLCStateMut, item SemanticNode, kind int) (*TLCStateMut, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	acts := EmptyActionItemList.Cons(item, con, DoNotRecordCostModel, kind)
	return t.EnabledWithActionItems(pred, acts, con, s0, s1)
}

// EnabledWithActionItems combines the source action-list overloads. Omitting
// the cost model selects DO_NOT_RECORD.
func (t *Tool) EnabledWithActionItems(pred SemanticNode, acts *ActionItemList, con *Context, s0 *TLCStateMut, s1 *TLCStateMut, costModels ...CostModel) (*TLCStateMut, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if len(costModels) > 1 {
		panic(NewIllegalArgumentException("EnabledWithActionItems accepts at most one cost model"))
	}
	cm := DoNotRecordCostModel
	if len(costModels) > 0 {
		cm = costModels[0]
	}
	return t.EnabledImpl(pred, acts, con, s0, s1, cm)
}

func (t *Tool) IsValidExpr(expr SemanticNode, contexts ...*Context) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if len(contexts) > 1 {
		panic(NewIllegalArgumentException("IsValidExpr accepts at most one context"))
	}
	ctxt := EmptyContext
	if len(contexts) > 0 {
		ctxt = contexts[0]
	}
	if t != nil && t.IsValidExprFunc != nil {
		return t.IsValidExprFunc(t, expr, ctxt)
	}
	return t.IsValidExprImpl(expr, ctxt)
}

func (t *Tool) IsValidTransition(action *Action, s0 *TLCStateMut, s1 *TLCStateMut) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.IsValidTransitionFunc != nil {
		return t.IsValidTransitionFunc(t, action, s0, s1)
	}
	return t.IsValidTransitionImpl(action, s0, s1)
}

func (t *Tool) IsValidState(action *Action, state *TLCStateMut) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.IsValidStateFunc != nil {
		return t.IsValidStateFunc(t, action, state)
	}
	return t.IsValidStateImpl(action, state)
}

func (t *Tool) IsValidAction(action *Action) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.IsValidActionFunc != nil {
		return t.IsValidActionFunc(t, action)
	}
	return t.IsValidActionImpl(action)
}

func (t *Tool) GetState(fp uint64, prev ...any) (*TLCStateInfo, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if len(prev) > 1 {
		panic(NewIllegalArgumentException("GetState accepts at most one predecessor"))
	}
	if t.GetStateFunc != nil {
		return t.GetStateFunc(t, fp, prev...)
	}
	if len(prev) != 0 {
		switch predecessor := prev[0].(type) {
		case *TLCStateInfo:
			info, err := t.GetStateAfter(fp, predecessor.ownedState())
			if err == nil && info != nil {
				info.StateNumber = predecessor.StateNumber + 1
				info.FP = &fp
				return info, nil
			}
			if err != nil {
				return nil, err
			}
			return nil, NewEvalException(ECTLCFailedToRecoverNext)
		case *TLCStateMut:
			info, err := t.GetStateAfter(fp, predecessor)
			if err == nil && info != nil {
				return info, nil
			}
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
	}
	info, err := t.GetInitState(fp)
	if err == nil && info != nil {
		return info, nil
	}
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *Tool) GetInitState(fp uint64) (*TLCStateInfo, error) {
	if t == nil {
		panic(NewNullPointerException())
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
	if t == nil {
		panic(NewNullPointerException())
	}
	if predecessor == nil {
		return nil, nil
	}
	SetCurrentState(predecessor)
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
	if t == nil {
		panic(NewNullPointerException())
	}
	if successor == nil || predecessor == nil {
		return nil, nil
	}
	SetCurrentState(predecessor)
	for _, action := range t.GetActions() {
		nextStates, err := t.GetNextStates(action, predecessor)
		if err != nil {
			return nil, err
		}
		for i := 0; nextStates != nil && i < nextStates.Size(); i++ {
			state := nextStates.At(i)
			if state != nil && statesEqualForReconstruction(successor, state) {
				state.SetPredecessor(predecessor)
				return NewTLCStateInfo(state, action), nil
			}
		}
	}
	return nil, nil
}

// Tool.getState(s1, s) skips candidates whose values are incomparable to s1.
// Catch only Java's TLCRuntimeException from equality, not state generation or
// other exception types, so reconstruction can try the next successor.
func statesEqualForReconstruction(successor, candidate *TLCStateMut) (equal bool) {
	defer func() {
		if failure := recover(); failure != nil {
			if javaRuntimeException(panicValueAsError(failure)) == nil {
				panic(failure)
			}
			equal = false
		}
	}()
	return successor.Equal(candidate)
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
	if t == nil {
		panic(NewNullPointerException())
	}
	if len(t.SymmetryPerms) == 0 {
		return nil
	}
	out := make([]*MVPerm, len(t.SymmetryPerms))
	copy(out, t.SymmetryPerms)
	return out
}

func (t *Tool) HasSymmetry() bool {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.HasSymmetryFunc != nil {
		return t.HasSymmetryFunc(t)
	}
	if config := t.ownedModelConfig(); config != nil && config.GetSymmetry() != "" {
		return true
	}
	return len(t.SymmetryPerms) > 0
}

// GetInitStateSpec retains the processor's vector object so cached callers see
// its current size and backing storage after additions or removals.
func (t *Tool) GetInitStateSpec() *Vect[*Action] {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetInitPred()
	}
	return NewVectFrom(t.InitStateSpec)
}

func (t *Tool) GetSpecActions() []*Action {
	if t == nil {
		panic(NewNullPointerException())
	}
	initPred := t.GetInitStateSpec()
	actions := t.GetActions()
	return initPred.Concat(NewVectFrom(actions)).ToSlice()
}

func (t *Tool) GetInvariants() []*Action {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetInvariants()
	}
	return t.Invariants
}

func (t *Tool) GetTemporals() []*Action {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetTemporal()
	}
	return t.Temporals
}

func (t *Tool) GetTemporalNames() []string {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetTemporalNames()
	}
	return t.TemporalNames
}

func (t *Tool) GetImpliedTemporals() []*Action {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetImpliedTemporals()
	}
	return t.ImpliedTemporals
}

func (t *Tool) GetImpliedTemporalNames() []string {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetImpliedTemporalNames()
	}
	return t.ImpliedTempNames
}

func (t *Tool) CheckAssumptions() int {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.CheckAssumptionsFunc != nil {
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
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetInvariantsNames()
	}
	return t.InvariantNames
}

func (t *Tool) GetImpliedActNames() []string {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetImpliedActionNames()
	}
	return t.ImpliedActNames
}

func (t *Tool) GetRootName() string {
	if t == nil {
		panic(NewNullPointerException())
	}
	rootName, _, _, _ := t.ownedSpecIdentity()
	if rootName == "" {
		return "Spec"
	}
	return rootName
}

func (t *Tool) GetRootFile() string {
	if t == nil {
		panic(NewNullPointerException())
	}
	_, rootFile, _, _ := t.ownedSpecIdentity()
	return rootFile
}

func (t *Tool) GetConfigFile() string {
	if t == nil {
		panic(NewNullPointerException())
	}
	_, _, configFile, _ := t.ownedSpecIdentity()
	return configFile
}

func (t *Tool) GetModelConfig() *ModelConfig {
	if t == nil {
		panic(NewNullPointerException())
	}
	config := t.ownedModelConfig()
	if config == nil {
		t.SetModelConfig(nil)
		config = t.ownedModelConfig()
	}
	return config
}

func (t *Tool) GetSpecProcessor() *SpecProcessor {
	if t == nil {
		panic(NewNullPointerException())
	}
	return t.ownedSpecProcessor()
}

func (t *Tool) GetResolver() FilenameToStream {
	if t == nil {
		panic(NewNullPointerException())
	}
	return t.ownedResolver()
}

func (t *Tool) GetModuleFiles(resolver FilenameToStream) []*TLAFile {
	if t == nil {
		panic(NewNullPointerException())
	}
	files := make([]*TLAFile, 0, len(t.ModuleFiles))
	for _, filename := range t.ModuleFiles {
		if isNil(resolver) {
			panic(NewNullPointerException())
		}
		files = append(files, resolver.Resolve(filename, false))
	}
	return files
}

func (t *Tool) GetSpecDir() string {
	if t == nil {
		panic(NewNullPointerException())
	}
	_, _, _, specDir := t.ownedSpecIdentity()
	return specDir
}

func (t *Tool) GetImpliedInitNames() []string {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetImpliedInitNames()
	}
	return t.ImpliedInitNames
}

func (t *Tool) GetImpliedInits() []*Action {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetImpliedInits()
	}
	return t.ImpliedInits
}

func (t *Tool) GetImpliedActions() []*Action {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetImpliedActions()
	}
	return t.ImpliedActions
}

func (t *Tool) GetNextStateSpec() *Action {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetNextPred()
	}
	return t.NextStateSpec
}

func (t *Tool) GetViewSpec() SemanticNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	config := t.ownedModelConfig()
	if t.ownedSpecProcessor() == nil || config == nil {
		return t.ViewSpec
	}
	if name := config.GetView(); name != "" {
		return t.configGetterDefinition(name, "view function").GetBody()
	}
	view := t.ownedSpecProcessor().ownedRuntimeParameters().View
	if view == nil {
		return nil
	}
	carrier := view.ownedCarrier()
	definition := t.ownedSpecProcessor().runtimeModuleDefinition(carrier.module, carrier.operator)
	if definition.Arity() != 0 {
		panic(NewTLCRuntimeException(ECTLCConfigIDRequiresNoArg, "view function", definition.Name.String()))
	}
	return definition.GetBody()
}

func (t *Tool) HasAlias() bool {
	if t == nil {
		panic(NewNullPointerException())
	}
	config := t.ownedModelConfig()
	return config != nil && config.GetAlias() != ""
}

func (t *Tool) GetAliasSpec() SemanticNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	config := t.ownedModelConfig()
	if t.ownedSpecProcessor() == nil || config == nil {
		return t.AliasSpec
	}
	name := config.GetAlias()
	if name == "" {
		panic(NewTLCRuntimeException(ECTLCConfigNoStateType))
	}
	return t.configGetterDefinition(name, "alias").GetBody()
}

// Spec's view, alias and postcondition getters use the current definitions and
// throw coded runtime exceptions at lookup time, outside processor config phases.
func (t *Tool) configGetterDefinition(name, kind string) *OpDefNode {
	value := t.ownedSpecProcessor().defn(name)
	if value == nil {
		panic(NewTLCRuntimeException(ECTLCConfigSpecifiedNotDefined, kind, name))
	}
	definition, ok := value.(*OpDefNode)
	if !ok {
		panic(NewTLCRuntimeException(ECTLCConfigIDMustNotBeConstant, kind, name))
	}
	if definition.Arity() != 0 {
		panic(NewTLCRuntimeException(ECTLCConfigIDRequiresNoArg, kind, name))
	}
	return definition
}

func (t *Tool) GetPostConditionSpecs() []*Action {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() == nil {
		return append([]*Action(nil), t.PostConditionSpecs...)
	}
	// Spec appends freshly resolved config actions after the processor's
	// runtime and possible actions on every getter call.
	result := t.ownedSpecProcessor().GetPostConditionSpecs()
	config := t.ownedModelConfig()
	if config == nil {
		return result
	}
	for _, name := range config.GetPostConditions() {
		definition := t.configGetterDefinition(name, "post condition")
		result = append(result, NewAction(definition.GetBody(), EmptyContext, name))
	}
	return result
}

func (t *Tool) GetModelConstraints() []SemanticNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetModelConstraints()
	}
	return append([]SemanticNode(nil), t.ModelConstraints...)
}

func (t *Tool) GetActionConstraints() []SemanticNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetActionConstraints()
	}
	return append([]SemanticNode(nil), t.ActionConstraints...)
}

// Java consumers dereference the processor's constraint arrays. Standalone
// native tools also support nil slices as their empty constraint lists.
func (t *Tool) requireConstraintArray(nodes []SemanticNode) []SemanticNode {
	if t != nil && t.ownedSpecProcessor() != nil && nodes == nil {
		panic(NewNullPointerException())
	}
	return nodes
}

func (t *Tool) requireActionArray(actions []*Action) []*Action {
	if t != nil && t.ownedSpecProcessor() != nil && actions == nil {
		panic(NewNullPointerException())
	}
	return actions
}

func (t *Tool) propertyNameAt(names []string, index int) string {
	if t != nil && t.ownedSpecProcessor() != nil {
		if names == nil {
			panic(NewNullPointerException())
		}
		if index < 0 || index >= len(names) {
			panic(NewArrayIndexOutOfBoundsException(index, len(names)))
		}
	}
	return nameAt(names, index)
}

func (t *Tool) propertyActionAt(actions []*Action, index int) *Action {
	actions = t.requireActionArray(actions)
	if index < 0 || index >= len(actions) {
		panic(NewArrayIndexOutOfBoundsException(index, len(actions)))
	}
	return actions[index]
}

func (t *Tool) GetAssumptions() []SemanticNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetAssumptions()
	}
	return append([]SemanticNode(nil), t.Assumptions...)
}

func (t *Tool) GetRLReward() SemanticNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetRLReward()
	}
	return t.RLReward
}

func (t *Tool) GetPeriodic() SemanticNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetPeriodic()
	}
	return t.Periodic
}

func (t *Tool) GetAssumptionIsAxiom() []bool {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		return t.ownedSpecProcessor().GetAssumptionIsAxiom()
	}
	return append([]bool(nil), t.AssumptionIsAxiom...)
}

func (t *Tool) GetCounterExampleDef() *OpDefNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() != nil {
		// Spec.getCounterExampleDef reads the live shared definitions. A
		// debugger expression can import TLCExt after the fast tool was copied.
		value := t.ownedSpecProcessor().Defns.Get("CounterExample")
		if value == nil {
			return nil
		}
		evaluating, ok := asEvaluatingValue(value)
		if !ok {
			panic(NewTLCRuntimeException(ECGeneral))
		}
		opDef := evaluating.GetOpDef()
		if opDef == nil {
			panic(NewNullPointerException())
		}
		if opDef.Arity() != 0 {
			panic(NewTLCRuntimeException(ECGeneral))
		}
		return opDef
	}
	return t.CounterExampleDef
}

func (t *Tool) GetTraceDef() *OpDefNode {
	if t == nil {
		return nil
	}
	if t.TraceDef != nil {
		return t.TraceDef
	}
	val := t.Lookup(NewSymbolNode("Trace"), EmptyContext, EmptyState, false)
	if ev, ok := asEvaluatingValue(val); ok {
		return ev.GetOpDef()
	}
	return nil
}

func (t *Tool) GetImportedTraceDef() *OpDefNode {
	if t == nil {
		return nil
	}
	return t.TraceDef
}

func (t *Tool) LivenessIsTrue() bool {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.LivenessIsTrueFunc != nil {
		return t.LivenessIsTrueFunc(t)
	}
	return semanticGraphArrayLength(t.ownedSpecProcessor() != nil, t.GetImpliedTemporals()) == 0
}

func (t *Tool) EvalAliasInfo(current *TLCStateInfo, successor *TLCStateMut, prefix func() []*TLCStateInfo) (*TLCStateInfo, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.EvalAliasInfoFunc != nil {
		if current != nil {
			SetCurrentState(current.ownedState())
		} else {
			SetCurrentState(nil)
		}
		return t.EvalAliasInfoFunc(t, current, successor, prefix)
	}
	return t.evalAliasInfo(current, successor, prefix, true)
}

func (t *Tool) evalAliasInfo(current *TLCStateInfo, successor *TLCStateMut, prefix func() []*TLCStateInfo, withPrefix bool) (*TLCStateInfo, error) {
	if !t.HasAlias() {
		return current, nil
	}
	if current == nil {
		panic(NewNullPointerException())
	}
	ctxt := EmptyContext
	if withPrefix {
		if traceDef := t.GetImportedTraceDef(); traceDef != nil && traceDef.Symbol != nil {
			ctxt = ctxt.Cons(traceDef.Symbol, NewLazySupplierValue(traceDef, func() Value {
				if prefix == nil {
					panic(NewNullPointerException())
				}
				return traceTupleFromStateInfos(prefix())
			}))
		}
	}
	SetCurrentState(current.ownedState())
	alias, err := t.evalAliasState(current.ownedState(), successor, ctxt)
	if err != nil {
		if isJavaEvalOrRuntimeException(err) {
			return AliasTLCStateInfo(aliasEvaluationErrorState(current.ownedState(), err), current), nil
		}
		return nil, err
	}
	if alias != nil {
		return AliasTLCStateInfo(alias, current), nil
	}
	return current, nil
}

func (t *Tool) EvalAliasInfoPair(curState *TLCStateInfo, sucState *TLCStateMut) (*TLCStateInfo, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.EvalAliasInfoPairFunc != nil {
		return t.EvalAliasInfoPairFunc(t, curState, sucState)
	}
	return t.evalAliasInfo(curState, sucState, nil, false)
}

func (t *Tool) EvalAliasInfoPrefix(current *TLCStateInfo, successor *TLCStateMut, prefix []*TLCStateInfo) (*TLCStateInfo, error) {
	return t.EvalAliasInfo(current, successor, func() []*TLCStateInfo {
		if prefix == nil {
			panic(NewNullPointerException())
		}
		return append([]*TLCStateInfo(nil), prefix...)
	})
}

func (t *Tool) EvalAliasInfoPrefixSuffix(current *TLCStateInfo, successor *TLCStateMut, prefix []*TLCStateInfo, suffix ...*TLCStateInfo) (*TLCStateInfo, error) {
	return t.EvalAliasInfo(current, successor, func() []*TLCStateInfo {
		if prefix == nil {
			panic(NewNullPointerException())
		}
		out := make([]*TLCStateInfo, 0, len(prefix)+len(suffix))
		out = append(out, prefix...)
		out = append(out, suffix...)
		return out
	})
}

func (t *Tool) EvalAlias(curState *TLCStateMut, sucState *TLCStateMut) *TLCStateMut {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.EvalAliasFunc != nil {
		SetCurrentState(curState)
		return t.EvalAliasFunc(t, curState, sucState)
	}
	if !t.HasAlias() || curState == nil {
		return curState
	}
	SetCurrentState(curState)
	alias, err := t.evalAliasState(curState, sucState, EmptyContext)
	if err != nil {
		if isJavaEvalOrRuntimeException(err) {
			return aliasEvaluationErrorState(curState, err)
		}
		panic(err)
	}
	if alias != nil {
		return alias
	}
	return curState
}

func (t *Tool) evalAliasState(current *TLCStateMut, successor *TLCStateMut, ctxt *Context) (alias *TLCStateMut, err error) {
	// Value.toState may itself throw one of the two failures caught by Java
	// Tool.evalAlias. Other exceptions escape this boundary.
	defer func() {
		if failure := recover(); failure != nil {
			if cause, ok := failure.(error); ok && isJavaEvalOrRuntimeException(cause) {
				alias, err = nil, cause
				return
			}
			panic(failure)
		}
	}()
	if t == nil {
		return current, nil
	}
	aliasSpec := t.GetAliasSpec()
	if aliasSpec == nil {
		return current, nil
	}
	value, err := t.Eval(aliasSpec, ctxt, current, successor, EvalClear, CostModel{})
	if err != nil {
		return nil, err
	}
	// Only RecordValue and FcnRcdValue override Java Value.toState;
	// CounterExample inherits RecordValue's implementation.
	var record *RecordValue
	switch value := value.(type) {
	case *RecordValue:
		record = value
	case *CounterExample:
		record = value.RecordValue
	case *FcnRcdValue:
		record = value.ToRecord()
	}
	if record == nil {
		return nil, nil
	}
	return record.ToState(), nil
}

func aliasEvaluationErrorState(current *TLCStateMut, err error) *TLCStateMut {
	record := NewRecordValueFromState(current)
	message := javaThrowableDetailMessage(err)
	if message == nil {
		panic(NewNullPointerException())
	}
	return NewExtendedRecordValue(record, UniqueStringOf("_ALIASEvalError"), NewStringValue(*message)).ToState()
}

func traceTupleFromStateInfos(infos []*TLCStateInfo) Value {
	values := make([]Value, 0, len(infos))
	for _, info := range infos {
		if info == nil {
			panic(NewNullPointerException())
		}
		values = append(values, NewRecordValueFromState(info.ownedState()))
	}
	return NewTupleValue(values)
}

func (t *Tool) NoDebug() *Tool {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.NoDebugFunc != nil {
		return t.NoDebugFunc(t)
	}
	if fast := t.ownedDebugFastTool(); fast != nil {
		return fast
	}
	if t.ownedDebugger() != nil {
		copied := *t
		copied.Debugger = nil
		copied.DebugPort = -1
		copied.DebugFastTool = nil
		copied.debugOwnersCaptured = false
		copied.debuggerOwner = nil
		copied.debugFastToolOwner = nil
		if copied.GetMode() == ModeDebugger {
			copied.Mode = ModeMC
			copied.modeOwner = ModeMC
		}
		return &copied
	}
	return t
}

func (t *Tool) ParseDebuggerExpression(location SourceLocation, condition string, semanticRoot ...*ModuleNode) (*OpDefNode, error) {
	if len(semanticRoot) > 1 {
		panic(NewIllegalArgumentException("ParseDebuggerExpression accepts at most one semantic root"))
	}
	if t != nil && t.ParseDebuggerExpressionFunc != nil {
		root := t.ownedSpecProcessor().GetRootModule()
		if len(semanticRoot) > 0 {
			root = semanticRoot[0]
		}
		return t.ParseDebuggerExpressionFunc(t, root, location, condition)
	}
	return nil, fmt.Errorf("debug breakpoint expression parsing is not configured: %s", condition)
}

func (t *Tool) GetModule(name string) *ModuleNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedSpecProcessor() == nil || t.ownedSpecProcessor().ModuleTbl == nil {
		return nil
	}
	return t.ownedSpecProcessor().ModuleTbl.GetModuleNode(UniqueStringOf(name))
}

func (t *Tool) IsDebugger() bool {
	if t == nil {
		panic(NewNullPointerException())
	}
	return (t.IsDebuggerFunc != nil && t.IsDebuggerFunc(t)) || t.ownedDebugger() != nil
}
