package tlc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const specProcessorVetoedProperty = "tlc2.tool.impl.SpecProcessor.vetoed"
const specProcessorPropertyAuxKey = "tlc.specProcessor.property"

var specProcessorStatics struct {
	sync.Once
	vetoes map[string]bool
}

// InitializeSpecProcessorStatics represents a fresh Java classloader for
// SpecProcessor's property-backed static veto set. It requires an idle runtime.
func InitializeSpecProcessorStatics() {
	specProcessorStatics.Once = sync.Once{}
	specProcessorStatics.vetoes = nil
}

func initializeSpecProcessorStatics() {
	specProcessorStatics.Do(func() {
		vetoes := make(map[string]bool)
		if name, ok := tlcLookupSystemProperty(specProcessorVetoedProperty); ok && name != "" {
			vetoes[name] = true
		}
		for _, name := range strings.Split(os.Getenv("TLAGO_SPEC_PROCESSOR_VETOED"), ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				vetoes[name] = true
			}
		}
		specProcessorStatics.vetoes = vetoes
	})
}

type Defns struct {
	defnIdx int
	table   []any
	order   *InsMap[*UniqueString, struct{}]
}

func NewDefns(initialSize ...int) *Defns {
	if len(initialSize) > 1 {
		panic(NewIllegalArgumentException("NewDefns accepts at most one initial size"))
	}
	idx := 0
	if len(initialSize) > 0 {
		idx = int(int32(initialSize[0]))
	}
	size := int(int32(idx) + 32)
	if size < 0 {
		panic(NewNegativeArraySizeException(fmtInt(size)))
	}
	return &Defns{defnIdx: idx, table: make([]any, size), order: NewInsMap[*UniqueString, struct{}]()}
}

func (d *Defns) Put(key any, value any) {
	if d == nil {
		panic(NewNullPointerException())
	}
	us := defnKey(key)
	loc := us.DefnLoc()
	if loc == -1 {
		loc = d.defnIdx
		d.defnIdx = int(int32(d.defnIdx) + 1)
		us.SetLoc(loc)
	}
	if loc < 0 {
		panic(NewArrayIndexOutOfBoundsException(loc, len(d.table)))
	}
	if loc >= len(d.table) {
		old := d.table
		newSize := max(int(int32(2*int32(len(old)))), int(int32(loc)+1))
		if newSize < 0 {
			panic(NewNegativeArraySizeException(fmtInt(newSize)))
		}
		d.table = make([]any, newSize)
		copy(d.table, old)
	}
	if loc >= len(d.table) {
		panic(NewArrayIndexOutOfBoundsException(loc, len(d.table)))
	}
	d.table[loc] = value
	if d.order == nil {
		d.order = NewInsMap[*UniqueString, struct{}]()
	}
	d.order.Set(us, struct{}{})
}

func (d *Defns) Get(key any) any {
	if d == nil {
		panic(NewNullPointerException())
	}
	us := defnKey(key)
	loc := us.DefnLoc()
	if loc < 0 || loc >= len(d.table) {
		return nil
	}
	return d.table[loc]
}

func (d *Defns) SetDefnCount(index int) {
	if d == nil {
		panic(NewNullPointerException())
	}
	d.defnIdx = int(int32(index))
}

func (d *Defns) Snapshot() *Defns {
	if d == nil {
		panic(NewNullPointerException())
	}
	out := &Defns{defnIdx: d.defnIdx, table: make([]any, len(d.table)), order: NewInsMap[*UniqueString, struct{}]()}
	copy(out.table, d.table)
	if d.order != nil {
		for key := range d.order.All() {
			out.order.Set(key, struct{}{})
		}
	}
	return out
}

func (d *Defns) All() func(func(*UniqueString, any) bool) {
	return func(yield func(*UniqueString, any) bool) {
		if d == nil || d.order == nil {
			return
		}
		for key := range d.order.All() {
			if key == nil {
				continue
			}
			loc := key.DefnLoc()
			if loc < 0 || loc >= len(d.table) {
				continue
			}
			if !yield(key, d.table[loc]) {
				return
			}
		}
	}
}

func defnKey(key any) *UniqueString {
	switch k := key.(type) {
	case nil:
		panic(NewNullPointerException())
	case *UniqueString:
		if k == nil {
			panic(NewNullPointerException())
		}
		return k
	case string:
		return UniqueStringOf(k)
	case *SymbolNode:
		if k == nil || k.Name == nil {
			panic(NewNullPointerException())
		}
		return k.Name
	default:
		panic(NewIllegalArgumentException(fmt.Sprintf("Defns key must be a UniqueString, string, or internal SymbolNode, got %T", key)))
	}
}

type TLAClass struct {
	Package string
	Modules *InsMap[string, any]
}

func NewTLAClass(pkg string) *TLAClass {
	return &TLAClass{Package: pkg, Modules: NewInsMap[string, any]()}
}

type Spec struct {
	SpecDir         string
	RootFile        string
	ConfigFile      string
	ProcessedDefs   *InsMap[string, struct{}]
	Defns           *Defns
	UnprocessedDefn *Defns
	TLAClass        *TLAClass
	Config          *ModelConfig
	Processor       *SpecProcessor

	identityCaptured bool
	specDirOwner     string
	rootFileOwner    string
	configFileOwner  string
	configOwner      *ModelConfig
	processorOwner   *SpecProcessor
}

func NewSpec(rootFile string, config *ModelConfig) *Spec {
	if config == nil {
		config = newModelConfig("", false)
	}
	specDir := filepath.Dir(rootFile)
	if specDir == "." {
		specDir = ""
	}
	spec := &Spec{
		SpecDir:         specDir,
		RootFile:        rootFile,
		ConfigFile:      config.configFileName,
		ProcessedDefs:   NewInsMap[string, struct{}](),
		Defns:           NewDefns(),
		UnprocessedDefn: NewDefns(),
		TLAClass:        NewTLAClass("tlc2.module"),
		Config:          config,
	}
	spec.Processor = NewSpecProcessorWithSpecObj(spec.RootName(), spec.Defns, config, spec)
	spec.UnprocessedDefn = spec.Processor.UnprocessedDefns
	spec.captureIdentity()
	return spec
}

func (s *Spec) captureIdentity() {
	if s == nil || s.identityCaptured {
		return
	}
	s.identityCaptured = true
	s.specDirOwner = s.SpecDir
	s.rootFileOwner = s.RootFile
	s.configFileOwner = s.ConfigFile
	s.configOwner = s.Config
	s.processorOwner = s.Processor
}

func (s *Spec) ownedIdentity() (string, string, string, *ModelConfig, *SpecProcessor) {
	if s != nil && s.identityCaptured {
		return s.specDirOwner, s.rootFileOwner, s.configFileOwner, s.configOwner, s.processorOwner
	}
	if s == nil {
		return "", "", "", nil, nil
	}
	return s.SpecDir, s.RootFile, s.ConfigFile, s.Config, s.Processor
}

func (s *Spec) RootName() string {
	if s == nil {
		panic(NewNullPointerException())
	}
	_, rootFile, _, _, _ := s.ownedIdentity()
	if rootFile == "" {
		return "Spec"
	}
	base := filepath.Base(rootFile)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if base == "" {
		return "Spec"
	}
	return base
}

func (s *Spec) GetModelConfig() *ModelConfig {
	if s == nil {
		panic(NewNullPointerException())
	}
	_, _, _, config, _ := s.ownedIdentity()
	return config
}

func (s *Spec) GetSpecProcessor() *SpecProcessor {
	if s == nil {
		panic(NewNullPointerException())
	}
	_, _, _, _, processor := s.ownedIdentity()
	return processor
}

func (s *Spec) ApplyToTool(tool *Tool) *Tool {
	if tool == nil {
		tool = NewTool()
	}
	if s == nil {
		return tool
	}
	specDir, rootFile, configFile, config, processor := s.ownedIdentity()
	if processor == nil {
		return tool
	}
	processor.ApplyToTool(tool)
	tool.SetSpecIdentity(s.RootName(), rootFile, configFile, specDir)
	tool.SetModelConfig(config)
	tool.SetSpecProcessor(processor)
	if rootFile != "" && len(tool.ModuleFiles) == 0 {
		tool.ModuleFiles = []string{rootFile}
	}
	return tool
}

type SpecProcessor struct {
	RootFile         string
	rootFileCaptured bool
	rootFileOwner    string
	ToolID           int32
	toolIDCaptured   bool
	toolIDOwner      int32
	Defns            *Defns
	defnsCaptured    bool
	defnsOwner       *Defns
	Config           *ModelConfig
	configCaptured   bool
	configOwner      *ModelConfig
	ModuleTbl        *ExternalModuleTable
	RootModule       *ModuleNode
	// SpecObj retains the semantic specification object supplied to Java's
	// SpecProcessor constructor. It is intentionally untyped here because the
	// production SANY graph lives in the parent package.
	specObj any

	ConstantDeclarations []*SymbolNode
	RootDefinitions      *InsMap[string, *OpDefNode]

	// Runtime targets resolve against ModuleTbl when their source phase runs.
	RuntimeParameters         RuntimeParameters
	runtimeParametersCaptured bool
	runtimeParametersOwner    RuntimeParameters
	RuntimeInvariantCompiler  func(string) (*OpDefNode, error)

	Variables                 []*UniqueString
	VariablesNodes            []*SymbolNode
	ProcessedDefs             *InsMap[*OpDefNode, struct{}]
	UnprocessedDefns          *Defns
	ConstantDefns             *InsMap[*ModuleNode, *InsMap[SemanticNode, any]]
	constantDefnsCaptured     bool
	constantDefnsOwner        *InsMap[*ModuleNode, *InsMap[SemanticNode, any]]
	Snapshot                  *Defns
	PreConstantSnap           *Defns
	InitPred                  *Vect[*Action]
	NextPred                  *Action
	Temporals                 []*Action
	TemporalNames             []string
	ImpliedTemporals          []*Action
	ImpliedTempNames          []string
	Invariants                []*Action
	InvariantNames            []string
	ImpliedInits              []*Action
	ImpliedInitNames          []string
	ImpliedActions            []*Action
	ImpliedActNames           []string
	ModelConstraints          []SemanticNode
	ActionConstraints         []SemanticNode
	ConfigErrors              []*ConfigError
	PossiblePostConds         []*Action
	possiblePostCondsCaptured bool
	possiblePostCondsOwner    []*Action
	Assumptions               []SemanticNode
	AssumptionIsAxiom         []bool
	RLReward                  SemanticNode
	Periodic                  SemanticNode
	SymmetrySpec              string
	SpecificationName         string
}

func (p *SpecProcessor) require() *SpecProcessor {
	if p == nil {
		panic(NewNullPointerException())
	}
	return p
}

func (p *SpecProcessor) ownedToolID() int32 {
	if p != nil && p.toolIDCaptured {
		return p.toolIDOwner
	}
	if p == nil {
		return 0
	}
	return p.ToolID
}

func (p *SpecProcessor) ownedRootFile() string {
	if p != nil && p.rootFileCaptured {
		return p.rootFileOwner
	}
	if p == nil {
		return ""
	}
	return p.RootFile
}

func (p *SpecProcessor) ownedConfig() *ModelConfig {
	if p != nil && p.configCaptured {
		return p.configOwner
	}
	if p == nil {
		return nil
	}
	return p.Config
}

func (p *SpecProcessor) ownedDefns() *Defns {
	if p != nil && p.defnsCaptured {
		return p.defnsOwner
	}
	if p == nil {
		return nil
	}
	return p.Defns
}

func (p *SpecProcessor) ownedConstantDefns() *InsMap[*ModuleNode, *InsMap[SemanticNode, any]] {
	if p != nil && p.constantDefnsCaptured {
		return p.constantDefnsOwner
	}
	if p == nil {
		return nil
	}
	return p.ConstantDefns
}

func (p *SpecProcessor) ownedPossiblePostConds() []*Action {
	if p != nil && p.possiblePostCondsCaptured {
		return p.possiblePostCondsOwner
	}
	if p == nil {
		return nil
	}
	return p.PossiblePostConds
}

func (p *SpecProcessor) ownedRuntimeParameters() RuntimeParameters {
	if p != nil && p.runtimeParametersCaptured {
		return p.runtimeParametersOwner
	}
	if p == nil {
		return RuntimeParameters{}
	}
	return p.RuntimeParameters
}

func (p *SpecProcessor) SetRuntimeParameters(parameters RuntimeParameters) {
	p.require()
	parameters = parameters.captureOwners()
	p.RuntimeParameters = parameters
	p.runtimeParametersOwner = parameters
	p.runtimeParametersCaptured = true
}

func (p *SpecProcessor) setOwnedPossiblePostConds(actions []*Action) {
	if p.possiblePostCondsCaptured {
		p.possiblePostCondsOwner = actions
	}
	p.PossiblePostConds = actions
}

func (p *SpecProcessor) appendOwnedPossiblePostCond(action *Action) {
	actions := append(p.ownedPossiblePostConds(), action)
	p.setOwnedPossiblePostConds(actions)
}

func (p *SpecProcessor) GetModuleTbl() *ExternalModuleTable { return p.require().ModuleTbl }
func (p *SpecProcessor) GetRootModule() *ModuleNode         { return p.require().RootModule }
func (p *SpecProcessor) GetSpecObj() any                    { return p.require().specObj }

func NewSpecProcessor(rootFile string, defns *Defns, config *ModelConfig) *SpecProcessor {
	return newSpecProcessor(rootFile, defns, config, nil)
}

// NewSpecProcessorWithSpecObj retains the semantic specification object at the
// same construction boundary as Java's final SpecProcessor.specObj field.
func NewSpecProcessorWithSpecObj(rootFile string, defns *Defns, config *ModelConfig, specObj any) *SpecProcessor {
	return newSpecProcessor(rootFile, defns, config, specObj)
}

func newSpecProcessor(rootFile string, defns *Defns, config *ModelConfig, specObj any) *SpecProcessor {
	initializeSpecProcessorStatics()
	if defns == nil {
		defns = NewDefns()
	}
	if config == nil {
		config = newModelConfig("", false)
	}
	p := &SpecProcessor{
		RootFile:          rootFile,
		InitPred:          NewVectWithCapacity[*Action](5),
		Assumptions:       []SemanticNode{},
		AssumptionIsAxiom: []bool{},
		ToolID:            specToolID(),
		Defns:             defns,
		Config:            config,
		specObj:           specObj,
		ProcessedDefs:     NewInsMap[*OpDefNode, struct{}](),
		UnprocessedDefns:  NewDefns(),
		ConstantDefns:     newInsMapAllowNilKeys[*ModuleNode, *InsMap[SemanticNode, any]](),
	}
	p.toolIDCaptured = true
	p.toolIDOwner = p.ToolID
	p.rootFileCaptured = true
	p.rootFileOwner = p.RootFile
	p.configCaptured = true
	p.configOwner = p.Config
	p.defnsCaptured = true
	p.defnsOwner = p.Defns
	p.constantDefnsCaptured = true
	p.constantDefnsOwner = p.ConstantDefns
	p.possiblePostCondsCaptured = true
	p.possiblePostCondsOwner = p.PossiblePostConds
	p.PreConstantSnap = p.ownedDefns().Snapshot()
	p.Snapshot = p.ownedDefns().Snapshot()
	return p
}

func (p *SpecProcessor) SetVariables(names []string) {
	nodes := make([]*SymbolNode, len(names))
	for i, name := range names {
		nodes[i] = NewVariableSymbolNode(name)
	}
	p.SetVariableNodes(nodes)
}

// SetVariableNodes retains the root module's declarations, as Java's
// SpecProcessor does, including their locations and symbol identities.
func (p *SpecProcessor) SetVariableNodes(nodes []*SymbolNode) {
	p.VariablesNodes = nodes
	p.Variables = make([]*UniqueString, len(nodes))
	for i, node := range nodes {
		variable := node.Name
		variable.SetLoc(i)
		p.Variables[i] = variable
	}
	SetUniqueStringVariableCount(len(nodes))
	if p.ownedDefns() != nil {
		p.ownedDefns().SetDefnCount(len(nodes))
		// SpecProcessor installs these predefined values before user definitions.
		p.ownedDefns().Put("TRUE", BoolTrue)
		p.ownedDefns().Put("FALSE", BoolFalse)
		p.ownedDefns().Put("BOOLEAN", NewSetEnumValue([]Value{BoolFalse, BoolTrue}, true))
		stringMethod := NewMethodValue("public static tlc2.value.impl.Value tlc2.module.Strings.STRING()", 0, func(_ []Value, _ int) (Value, error) { return STRING(), nil })
		stringMethod.ParameterCount = 0
		stringMethod.captureMetadata()
		p.ownedDefns().Put("STRING", stringMethod)
		// Java snapshots just these predefined/native entries before processing
		// source modules, ordinary definitions, native overrides or config.
		p.PreConstantSnap = p.ownedDefns().Snapshot()
	}
}

func (p *SpecProcessor) ProcessConfig() {
	if p == nil || p.ownedConfig() == nil {
		return
	}
	defer p.recoverConfigFailure()
	p.resetProcessedConfig()
	p.SpecificationName = p.ownedConfig().GetSpec()
	p.SymmetrySpec = p.ownedConfig().GetSymmetry()

	p.processConfigInvariants()
	if p.SpecificationName != "" {
		if p.ownedConfig().GetInit() != "" || p.ownedConfig().GetNext() != "" {
			p.addConfigError(ECTLCConfigNotBothSpecAndInit)
		} else {
			p.processSpecificationConfig()
		}
	} else {
		if initName := p.ownedConfig().GetInit(); initName != "" {
			if action := p.actionFromConfigName(initName, true, "initial predicate"); action != nil {
				p.InitPred.AddElement(action)
			}
		}
		if nextName := p.ownedConfig().GetNext(); nextName != "" {
			p.NextPred = p.actionFromConfigName(nextName, false, "next state action")
		}
	}
	p.processConfigProperties()
	p.processMissingInitNextConfig()
	p.processSpecPropertyTautologyWarning()
	runtimeParameters := p.ownedRuntimeParameters()
	for _, invariant := range runtimeParameters.Invariants {
		if p.RuntimeInvariantCompiler == nil {
			panic(NewTLCRuntimeException(ECTLCParsingFailed2, "runtime invariant expression compiler is not configured"))
		}
		action, err := invariant.GetAction(p)
		if err != nil {
			panic(NewTLCRuntimeExceptionWithCause(ECTLCParsingFailed2, err))
		}
		p.Invariants = append(p.Invariants, action)
		p.InvariantNames = append(p.InvariantNames, action.GetNameOfDefault())
	}
	for _, constraint := range runtimeParameters.Constraints {
		carrier := constraint.ownedCarrier()
		definition := p.runtimeModuleDefinition(carrier.module, carrier.operator)
		p.ownedDefns().Put(definition.Name, definition)
		p.ownedConfig().constraints = append(p.ownedConfig().constraints, definition.Name.String())
	}
	p.ModelConstraints = p.constraintNodesFromConfigNames(p.ownedConfig().GetConstraints(), "constraint", ECTLCConfigIDRequiresNoArg, ECTLCConfigSpecifiedNotDefined, ECTLCConfigIDHasValue)
	for _, constraint := range runtimeParameters.ActionConstraints {
		carrier := constraint.ownedCarrier()
		definition := p.runtimeModuleDefinition(carrier.module, carrier.operator)
		p.ownedDefns().Put(definition.Name, definition)
		p.ownedConfig().actionConstraints = append(p.ownedConfig().actionConstraints, definition.Name.String())
	}
	p.ActionConstraints = p.constraintNodesFromConfigNames(p.ownedConfig().GetActionConstraints(), "action constraint", ECTLCConfigIDRequiresNoArg, ECTLCConfigSpecifiedNotDefined, ECTLCConfigIDHasValue)
	p.RLReward = p.optionalOpBodyFromConfigName(p.ownedConfig().GetRLReward(), "rlreward", p.preConstantDefinitions())
	p.Periodic = p.optionalOpBodyFromConfigName(p.ownedConfig().GetPeriodic(), "periodic", p.preConstantDefinitions())
	p.processConfigPossible()
}

func (p *SpecProcessor) resetProcessedConfig() {
	p.ConfigErrors = nil
	p.InitPred = NewVectWithCapacity[*Action](5)
	p.NextPred = nil
	// Java materializes non-null arrays even when a property family is empty.
	p.Temporals = []*Action{}
	p.TemporalNames = []string{}
	p.ImpliedTemporals = []*Action{}
	p.ImpliedTempNames = []string{}
	p.Invariants = []*Action{}
	p.InvariantNames = []string{}
	p.ImpliedInits = []*Action{}
	p.ImpliedInitNames = []string{}
	p.ImpliedActions = []*Action{}
	p.ImpliedActNames = []string{}
	p.ModelConstraints = nil
	p.ActionConstraints = nil
	p.setOwnedPossiblePostConds(p.ownedPossiblePostConds()[:0])
	p.RLReward = nil
	p.Periodic = nil
}

func (p *SpecProcessor) ApplyToTool(tool *Tool) {
	if p == nil || tool == nil {
		return
	}
	p.ConfigErrors = nil
	defer func() { tool.ConfigErrors = append([]*ConfigError(nil), p.ConfigErrors...) }()
	defer p.recoverConfigFailure()
	if !p.toolIDCaptured {
		p.ToolID = tool.GetID()
	}
	tool.SetSpecProcessor(p)
	tool.SetModelConfig(p.ownedConfig())
	if p.VariablesNodes != nil {
		SetStateVariableDeclarations(p.VariablesNodes)
	} else {
		names := make([]string, len(p.Variables))
		for i, variable := range p.Variables {
			names[i] = variable.String()
		}
		SetStateVariables(names)
	}
	p.applyDefinitionsToTool(tool)
	p.ProcessConfigConstantsAndOverrides(tool)
	if len(p.ConfigErrors) != 0 {
		return
	}
	// Java snapshots after processSpec installs overrides, before pre-evaluation.
	p.Snapshot = p.ownedDefns().Snapshot()
	p.ProcessConstantDefinitions(tool)
	p.ProcessConfig()
	if len(p.ConfigErrors) != 0 {
		return
	}
	p.applyDefinitionsToTool(tool)
	tool.SetModelConfig(p.ownedConfig())
	tool.InitStateSpec = p.InitPred.ToSlice()
	tool.NextStateSpec = p.NextPred
	tool.Temporals = append([]*Action(nil), p.Temporals...)
	tool.TemporalNames = append([]string(nil), p.TemporalNames...)
	tool.ImpliedTemporals = append([]*Action(nil), p.ImpliedTemporals...)
	tool.ImpliedTempNames = append([]string(nil), p.ImpliedTempNames...)
	tool.Invariants = append([]*Action(nil), p.Invariants...)
	tool.InvariantNames = append([]string(nil), p.InvariantNames...)
	tool.ImpliedInits = append([]*Action(nil), p.ImpliedInits...)
	tool.ImpliedInitNames = append([]string(nil), p.ImpliedInitNames...)
	tool.ImpliedActions = append([]*Action(nil), p.ImpliedActions...)
	tool.ImpliedActNames = append([]string(nil), p.ImpliedActNames...)
	tool.ModelConstraints = append([]SemanticNode(nil), p.ModelConstraints...)
	tool.ActionConstraints = append([]SemanticNode(nil), p.ActionConstraints...)
	tool.ConfigErrors = append([]*ConfigError(nil), p.ConfigErrors...)
	tool.Assumptions = append([]SemanticNode(nil), p.Assumptions...)
	tool.AssumptionIsAxiom = append([]bool(nil), p.AssumptionIsAxiom...)
	tool.RLReward = p.RLReward
	tool.Periodic = p.Periodic
	// Tool construction captures the view before state symmetry setup. Alias
	// lookup stays lazy inside trace rendering, where its failures are caught.
	tool.ViewSpec = tool.GetViewSpec()
	p.processConfigSymmetry(tool)
	tool.ConfigErrors = append([]*ConfigError(nil), p.ConfigErrors...)
	tool.AssignActionIDs()
}

func (p *SpecProcessor) GetVariablesNodes() []*SymbolNode {
	return p.require().VariablesNodes
}

func (p *SpecProcessor) GetInitPred() *Vect[*Action] {
	return p.require().InitPred
}

func (p *SpecProcessor) GetNextPred() *Action {
	return p.require().NextPred
}

func (p *SpecProcessor) GetTemporal() []*Action {
	return p.require().Temporals
}

func (p *SpecProcessor) GetTemporalNames() []string {
	return p.require().TemporalNames
}

func (p *SpecProcessor) GetImpliedTemporals() []*Action {
	return p.require().ImpliedTemporals
}

func (p *SpecProcessor) GetImpliedTemporalNames() []string {
	return p.require().ImpliedTempNames
}

func (p *SpecProcessor) GetInvariants() []*Action {
	return p.require().Invariants
}

func (p *SpecProcessor) GetInvariantsNames() []string {
	return p.require().InvariantNames
}

func (p *SpecProcessor) GetImpliedInits() []*Action {
	return p.require().ImpliedInits
}

func (p *SpecProcessor) GetImpliedInitNames() []string {
	return p.require().ImpliedInitNames
}

func (p *SpecProcessor) GetImpliedActions() []*Action {
	return p.require().ImpliedActions
}

func (p *SpecProcessor) GetImpliedActionNames() []string {
	return p.require().ImpliedActNames
}

func (p *SpecProcessor) GetModelConstraints() []SemanticNode {
	return p.require().ModelConstraints
}

func (p *SpecProcessor) GetActionConstraints() []SemanticNode {
	return p.require().ActionConstraints
}

func (p *SpecProcessor) GetRLReward() SemanticNode {
	return p.require().RLReward
}

func (p *SpecProcessor) GetPeriodic() SemanticNode {
	return p.require().Periodic
}

func (p *SpecProcessor) GetAssumptions() []SemanticNode {
	return p.require().Assumptions
}

func (p *SpecProcessor) GetAssumptionIsAxiom() []bool {
	return p.require().AssumptionIsAxiom
}

func (p *SpecProcessor) GetUnprocessedDefns() *Defns {
	return p.require().Snapshot
}

func (p *SpecProcessor) GetDefns() *Defns {
	return p.require().ownedDefns()
}

func (p *SpecProcessor) GetConstantDefns() *InsMap[*ModuleNode, *InsMap[SemanticNode, any]] {
	return p.require().ownedConstantDefns()
}

func (p *SpecProcessor) GetPostConditionSpecs() []*Action {
	p.require()
	possiblePostConds := p.ownedPossiblePostConds()
	runtimeParameters := p.ownedRuntimeParameters()
	result := make([]*Action, 0, len(runtimeParameters.PostConditions)+len(possiblePostConds))
	for _, post := range runtimeParameters.PostConditions {
		carrier := post.ownedCarrier()
		definition := p.runtimeModuleDefinition(carrier.module, carrier.operator)
		result = append(result, NewAction(definition.GetBody(), EmptyContext, carrier.operator))
	}
	return append(result, possiblePostConds...)
}

// ParameterizedSpecObj resolves only external modules and their actual OpDefs.
// Lookup aliases, named theorems and inner modules are not substitutes for a
// definition in the requested module's source Context.
func (p *SpecProcessor) runtimeModuleDefinition(module, operator string) *OpDefNode {
	node := p.ModuleTbl.GetModuleNode(UniqueStringOf(module))
	if node == nil {
		panic(NewTLCRuntimeException(ECGeneral, "Could not find module: "+module))
	}
	definition := node.GetOpDefByName(operator)
	if definition == nil {
		panic(NewTLCRuntimeException(ECGeneral, "Could not find operator: "+operator+" in module: "+module))
	}
	return definition
}

func (p *SpecProcessor) ProcessConstantDefinitions(tool *Tool) {
	p.require()
	modules := p.ModuleTbl.GetModuleNodes()
	if tool == nil {
		tool = p.configProcessingTool()
	} else {
		p.applyDefinitionsToTool(tool)
	}
	vetoes := specProcessorVetoedConstantOperators()
	for _, module := range modules {
		if module.ProcessConstantDefns() {
			p.processModuleConstantDefinitions(tool, module, vetoes)
		}
	}
}

func (p *SpecProcessor) recordConstantDefinition(module *ModuleNode, node SemanticNode, value any) {
	constantDefns := p.ownedConstantDefns()
	if constantDefns == nil {
		p.ConstantDefns = newInsMapAllowNilKeys[*ModuleNode, *InsMap[SemanticNode, any]]()
		constantDefns = p.ConstantDefns
	}
	definitions := constantDefns.Get(module)
	if definitions == nil {
		definitions = NewInsMap[SemanticNode, any]()
		constantDefns.Set(module, definitions)
	}
	definitions.Set(node, value)
}

func (p *SpecProcessor) processModuleConstantDefinitions(tool *Tool, module *ModuleNode, vetoes map[string]bool) {
	p.processDeclaredConstantDefinitions(tool, module)
	for _, original := range module.GetOpDefs() {
		origin := original.GetOriginallyDefinedInModuleNode()
		if origin != nil && !origin.ProcessConstantDefns() || original.Arity() != 0 {
			continue
		}
		// Java pre-evaluates only when lookup still returns an OpDefNode.
		// Native/body overrides and installed values keep their current binding.
		opDef, ok := tool.Lookup(original.Symbol, EmptyContext, EmptyState, false).(*OpDefNode)
		if !ok || opDef == nil || tool.GetLevelBound(opDef.GetBody(), EmptyContext) != TLCLevelConstant {
			continue
		}
		p.processConstantOperatorDefinition(tool, opDef, origin, vetoes)
	}
	// Java does not reapply processConstantDefns eligibility to inner modules.
	for _, inner := range module.GetInnerModules() {
		p.processModuleConstantDefinitions(tool, inner, vetoes)
	}
}

// Java's ordinary operator loop establishes its catch(Throwable) only after
// lookup and level analysis. Everything from the definition name through map
// bookkeeping is swallowed when pre-evaluation is not usable.
func (p *SpecProcessor) processConstantOperatorDefinition(tool *Tool, opDef *OpDefNode, origin *ModuleNode, vetoes map[string]bool) {
	defer func() { _ = recover() }()
	opName := opDef.Name
	if opName == nil {
		panic(NewNullPointerException())
	}
	if specProcessorConstantVetoed(vetoes, opName) {
		return
	}
	value, err := evaluateConstantOperatorDefinition(tool, opDef)
	if err != nil || value == nil {
		return
	}
	opDef.SetToolObjectAt(p.ownedToolID(), value)
	// Only replace the global entry when it still denotes this exact node.
	// Hidden/imported definitions retain their values on the semantic node.
	if p.ownedDefns().Get(opName) == opDef {
		p.ownedDefns().Put(opName, value)
		if opDef.HasSource() {
			origin = opDef.GetSource().GetOriginallyDefinedInModuleNode()
		}
		p.recordConstantDefinition(origin, opDef, value)
	}
}

func specProcessorVetoedConstantOperators() map[string]bool {
	initializeSpecProcessorStatics()
	return specProcessorStatics.vetoes
}

// Java processConstantDefns evaluates replacements of declared constants even
// when the replacement is not constant-level. Only a failed evaluation of a
// genuinely nonconstant operator is a configuration error.
func (p *SpecProcessor) processDeclaredConstantDefinitions(tool *Tool, module *ModuleNode) {
	for _, declaration := range module.GetConstantDecls() {
		value := declaration.Data
		if declaration.SemanticBase != nil {
			value = declaration.GetToolObjectAt(p.ownedToolID())
		}
		if constant, ok := value.(Value); ok && !isNil(constant) {
			InitializeValue(constant)
			p.recordConstantDefinition(module, declaration, constant)
			continue
		}
		opDef, ok := value.(*OpDefNode)
		if !ok || opDef == nil {
			continue
		}
		if opDef.Arity() != declaration.GetArity() {
			panic(NewTLCRuntimeException(ECTLCConfigWrongSubstitutionNumberOfArgs, declaration.GetName().String(), opDef.Name.String()))
		}
		if opDef.Arity() != 0 {
			continue
		}
		result, err := evaluateDeclaredConstantDefinition(tool, opDef)
		if err != nil {
			if tool.GetLevelBound(opDef.GetBody(), EmptyContext) > TLCLevelConstant {
				addendum := ""
				if !isValueEvalException(err) {
					addendum = " - specifically: " + err.Error()
				}
				panic(NewTLCRuntimeException(ECTLCConfigSubstitutionNonConstant, declaration.GetName().String(), opDef.Name.String(), addendum))
			}
			continue
		}
		opDef.SetToolObjectAt(p.ownedToolID(), result)
		if property, _ := tlcLookupSystemProperty("tlc2.tool.impl.SpecProcessor.aggressiveConstantCaching"); javaBooleanProperty(property) {
			if declaration.SemanticBase != nil {
				declaration.SetToolObjectAt(p.ownedToolID(), result)
			} else {
				declaration.Data = result
			}
		}
		p.recordConstantDefinition(module, opDef, result)
	}
}

func evaluateDeclaredConstantDefinition(tool *Tool, opDef *OpDefNode) (result any, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			failure, ok := recovered.(error)
			if !ok || !isValueEvalException(failure) && javaRuntimeException(failure) == nil {
				panic(recovered)
			}
			err = failure
		}
	}()
	result, err = DemuxWorkerValue(func() (Value, error) {
		return tool.Eval(opDef.GetBody(), EmptyContext, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
	})
	if err != nil && !isValueEvalException(err) && javaRuntimeException(err) == nil {
		panic(err)
	}
	return result, err
}

// The ordinary operator pre-evaluation loop catches Throwable, unlike the
// narrower declared-constant replacement catch above.
func evaluateConstantOperatorDefinition(tool *Tool, opDef *OpDefNode) (result any, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result, err = nil, panicValueAsError(recovered)
		}
	}()
	return DemuxWorkerValue(func() (Value, error) {
		return tool.Eval(opDef.GetBody(), EmptyContext, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
	})
}

func specProcessorConstantVetoed(vetoes map[string]bool, names ...*UniqueString) bool {
	for _, name := range names {
		if name != nil && vetoes[name.String()] {
			return true
		}
	}
	return false
}

func (p *SpecProcessor) ProcessConfigConstantsAndOverrides(tool *Tool) {
	if p == nil || p.ownedConfig() == nil || p.ownedDefns() == nil {
		return
	}
	defer p.recoverConfigFailure()
	constants := configConstantsToDefns(p.ownedConfig().GetConstants(), p)
	for name, value := range constants.All() {
		// Java attaches a global operator constant assignment to the root
		// OpDefNode itself. Module aliases can still refer to that same node;
		// registering those aliases must not undo the configured replacement.
		if p.RootDefinitions != nil {
			if definition := p.RootDefinitions.Get(name); definition != nil {
				definition.SetToolObjectAt(p.ownedToolID(), value)
			}
		}
		p.putConfigDefinition(name, value, tool)
	}
	p.applyConfigModuleConstants(tool)
	p.applyConfigOverrides(tool)
	p.applyConfigModuleOverrides(tool)
	p.Snapshot = p.ownedDefns().Snapshot()
}

func configConstantsToDefns(constants *ConfigConstants, p *SpecProcessor) *InsMap[string, any] {
	out := NewInsMap[string, any]()
	if constants == nil {
		return out
	}
	for _, line := range constants.All() {
		if line.Name == "" {
			continue
		}
		if len(line.Args) == 0 {
			out.Set(line.Name, line.Value)
			continue
		}
		existing, ok := out.Get2(line.Name)
		var opVal *OpRcdValue
		if !ok || existing == nil {
			opVal = NewOpRcdValue()
			out.Set(line.Name, opVal)
		} else if cast, ok := existing.(*OpRcdValue); ok {
			opVal = cast
		} else {
			if p != nil {
				p.addConfigError(ECTLCConfigOpArityInconsistent, line.Name)
			}
			continue
		}
		if len(opVal.Domain) != 0 && len(opVal.Domain[0]) != len(line.Args) {
			if p != nil {
				p.addConfigError(ECTLCConfigOpArityInconsistent, line.Name)
			}
			continue
		}
		values := make([]Value, 0, len(line.Args)+2)
		values = append(values, ValUndef)
		values = append(values, line.Args...)
		values = append(values, line.Value)
		opVal.AddLine(values)
	}
	return out
}

func (p *SpecProcessor) applyConfigModuleConstants(tool *Tool) {
	if p == nil || p.ownedConfig() == nil || p.ownedConfig().GetModConstants() == nil {
		return
	}
	for modName, constants := range p.ownedConfig().GetModConstants().All() {
		if modName == "" {
			continue
		}
		if !p.hasModuleDefinitions(modName) {
			p.addConfigError(ECTLCNoModules, modName)
			continue
		}
		values := configConstantsToDefns(constants, p)
		for name, value := range values.All() {
			p.putConfigDefinition(configQualifiedName(modName, name), value, tool)
		}
	}
}

func (p *SpecProcessor) applyConfigOverrides(tool *Tool) {
	if p == nil || p.ownedConfig() == nil || p.ownedConfig().GetOverrides() == nil {
		return
	}
	// Java visits root OpDefs in semantic context order, then checks leftover
	// override keys. Keep the source graph even after Defns entries are replaced.
	names := NewInsMap[string, struct{}]()
	if p.RootDefinitions != nil {
		for name := range p.RootDefinitions.All() {
			if _, ok := p.ownedConfig().GetOverrides().Get2(name); ok {
				names.Set(name, struct{}{})
			}
		}
	}
	for name := range p.ownedConfig().GetOverrides().All() {
		names.Set(name, struct{}{})
	}
	for lhs := range names.All() {
		rhs := p.ownedConfig().GetOverrides().Get(lhs)
		if lhs == "" {
			continue
		}
		if _, chained := p.ownedConfig().GetOverrides().Get2(rhs); chained {
			p.addConfigError(ECTLCConfigRHSIDAppearedAfterLHSID, rhs)
			continue
		}
		lhsVal := p.defn(lhs)
		if p.RootDefinitions != nil {
			if def := p.RootDefinitions.Get(lhs); def != nil {
				lhsVal = def
			}
		}
		if lhsVal == nil {
			for _, declaration := range p.ConstantDeclarations {
				if declaration.Name.String() == lhs {
					lhsVal = declaration
					break
				}
			}
		}
		if lhsVal == nil {
			p.addConfigError(ECTLCConfigIDDoesNotAppearInSpec, lhs)
			continue
		}
		rhsVal := p.defn(rhs)
		if rhsVal == nil {
			p.addConfigError(ECTLCConfigWrongSubstitution, lhs, rhs)
			continue
		}
		if lhsDef, ok := lhsVal.(*OpDefNode); ok && lhsDef != nil {
			if rhsDef, ok := rhsVal.(*OpDefNode); ok && rhsDef != nil {
				if lhsDef.Arity() != rhsDef.Arity() {
					p.addConfigError(ECTLCConfigWrongSubstitutionNumberOfArgs, lhs, rhs)
					continue
				}
				if enabled, _ := tlcLookupSystemProperty("tlc2.tool.impl.SpecProcessor.allowCyclicRedefinitions"); javaBooleanProperty(enabled) && SemanticIsDefinedWith(rhsDef, lhsDef) {
					SemanticSubstituteFor(p.rootSemanticNodes(), rhsDef, lhsDef)
					p.putConfigDefinition(lhs, rhsVal, nil)
					continue
				}
			}
			lhsDef.SetToolObjectAt(p.ownedToolID(), rhsVal)
		}
		p.putConfigDefinition(lhs, rhsVal, tool)
	}
}

func (p *SpecProcessor) rootSemanticNodes() []SemanticNode {
	nodes := append([]SemanticNode(nil), p.Assumptions...)
	if p.RootDefinitions != nil {
		for _, def := range p.RootDefinitions.All() {
			nodes = append(nodes, def)
		}
	} else if p.ownedDefns() != nil {
		for _, value := range p.ownedDefns().All() {
			if def, ok := value.(*OpDefNode); ok {
				nodes = append(nodes, def)
			}
		}
	}
	return nodes
}

func (p *SpecProcessor) applyConfigModuleOverrides(tool *Tool) {
	if p == nil || p.ownedConfig() == nil || p.ownedConfig().GetModOverrides() == nil {
		return
	}
	for modName, overrides := range p.ownedConfig().GetModOverrides().All() {
		if modName == "" {
			continue
		}
		if !p.hasModuleDefinitions(modName) {
			p.addConfigError(ECTLCNoModules, modName)
			continue
		}
		for lhs, rhs := range overrides.All() {
			if _, chained := overrides.Get2(rhs); chained {
				p.addConfigError(ECTLCConfigRHSIDAppearedAfterLHSID, rhs)
				continue
			}
			qualified := configQualifiedName(modName, lhs)
			lhsVal := p.defn(qualified)
			// Native overrides replaced the table entry, but Java still visits
			// the module's original OpDef and updates its shared body.
			if p.ModuleTbl != nil && p.ModuleTbl.GetModuleNode(UniqueStringOf(modName)) != nil {
				lhsVal = nil
				for _, definition := range p.ModuleTbl.GetModuleNode(UniqueStringOf(modName)).GetOpDefs() {
					if definition.Name == UniqueStringOf(lhs) {
						lhsVal = definition
						break
					}
				}
			} else if p.RootDefinitions != nil && p.RootDefinitions.Get(qualified) != nil {
				lhsVal = p.RootDefinitions.Get(qualified)
			}
			if lhsVal == nil {
				p.addConfigError(ECTLCConfigIDDoesNotAppearInSpec, lhs)
				continue
			}
			rhsVal := p.defn(rhs)
			if rhsVal == nil {
				p.addConfigError(ECTLCConfigWrongSubstitution, lhs, rhs)
				continue
			}
			if lhsDef, ok := lhsVal.(*OpDefNode); ok && lhsDef != nil {
				if rhsDef, ok := rhsVal.(*OpDefNode); ok && rhsDef != nil && lhsDef.Arity() != rhsDef.Arity() {
					p.addConfigError(ECTLCConfigWrongSubstitutionNumberOfArgs, lhs, rhs)
					continue
				}
				SetSemanticToolObjectForToolID(p.ownedToolID(), lhsDef.GetBody(), rhsVal)
			}
			p.putConfigDefinition(qualified, rhsVal, tool)
		}
	}
}

func (p *SpecProcessor) putConfigDefinition(name string, value any, tool *Tool) {
	if p == nil || p.ownedDefns() == nil || name == "" || value == nil {
		return
	}
	old := p.defn(name)
	p.ownedDefns().Put(name, value)

	if tool == nil {
		return
	}
	sym := NewSymbolNode(name)
	if oldDef, ok := old.(*OpDefNode); ok && oldDef != nil && oldDef.Symbol != nil {
		sym = oldDef.Symbol
	} else if p.RootDefinitions != nil && p.RootDefinitions.Get(name) != nil {
		sym = p.RootDefinitions.Get(name).Symbol
	} else {
		for existing := range tool.Definitions {
			if existing.Name == sym.Name {
				tool.Define(existing, value)
			}
		}
		for _, declaration := range p.ConstantDeclarations {
			if declaration.Name == sym.Name {
				sym = declaration
				break
			}
		}
	}
	if sym.Kind == SymbolConstantDecl {
		if sym.SemanticBase != nil {
			sym.SetToolObjectAt(p.ownedToolID(), value)
		} else {
			sym.Data = value
		}
	}
	tool.Define(sym, value)
}

func (p *SpecProcessor) hasModuleDefinitions(moduleName string) bool {
	if p == nil || p.ownedDefns() == nil {
		return false
	}
	if moduleName == p.ownedRootFile() {
		return true
	}
	prefix := moduleName + "!"
	for name := range p.ownedDefns().All() {
		if name != nil && strings.HasPrefix(name.String(), prefix) {
			return true
		}
	}
	return false
}

func configQualifiedName(moduleName string, name string) string {
	if moduleName == "" || strings.Contains(name, "!") {
		return name
	}
	return moduleName + "!" + name
}

func (p *SpecProcessor) processSpecificationConfig() {
	switch spec := p.defn(p.SpecificationName).(type) {
	case *OpDefNode:
		if spec == nil {
			p.addConfigError(ECTLCConfigSpecifiedNotDefined, "name", p.SpecificationName)
			return
		}
		if spec.Arity() != 0 {
			p.addConfigError(ECTLCConfigIDRequiresNoArg, p.SpecificationName)
			return
		}
		tool := p.configProcessingTool()
		p.processConfigSpec(tool, spec.GetBody(), EmptyContext, EmptyList, nil)
	case nil:
		p.addConfigError(ECTLCConfigSpecifiedNotDefined, "name", p.SpecificationName)
	default:
		p.addConfigError(ECTLCConfigIDHasValue, "value", p.SpecificationName, configValueString(spec))
	}
}

func (p *SpecProcessor) processConfigInvariants() {
	if p == nil || p.ownedConfig() == nil {
		return
	}
	for _, name := range p.ownedConfig().GetInvariants() {
		if value, ok := p.defn(name).(*BoolValue); ok && value.Val {
			PrintWarning(ECTLCInvariantConstantLevel, name, value.String())
			continue
		}
		def, ok := p.configPredicateOpDef(name, "invariant", p.ownedDefns())
		if !ok {
			continue
		}
		if def.GetLevel() >= TLCLevelAction {
			params := []string{name}
			if len(def.GetLevelParams()) != 0 {
				params = append(params, "includeWarning")
			}
			p.addConfigError(ECTLCInvariantViolatedLevel, params...)
			continue
		}
		p.InvariantNames = append(p.InvariantNames, name)
		p.Invariants = append(p.Invariants, NewActionFromOpDef(def.GetBody(), EmptyContext, def, false, false))
	}
}

func (p *SpecProcessor) processConfigProperties() {
	if p == nil || p.ownedConfig() == nil {
		return
	}
	tool := p.configProcessingTool()
	for _, name := range p.ownedConfig().GetProperties() {
		def, ok := p.configPredicateOpDef(name, "property", p.ownedDefns())
		if !ok {
			continue
		}
		if def.GetLevel() == TLCLevelState {
			PrintWarning(ECTLCLiveFormulaStateLevel, name)
		}
		p.processConfigProperty(tool, name, name, def.GetBody(), EmptyContext, EmptyList)
	}
}

func (p *SpecProcessor) processConfigPossible() {
	if p == nil || p.ownedConfig() == nil {
		return
	}
	possibleNames := p.ownedConfig().GetPossible()
	if len(possibleNames) == 0 {
		return
	}
	tool := p.configProcessingTool()
	for _, name := range possibleNames {
		def, ok := p.configOpDef(name, "possible", p.ownedDefns())
		if !ok {
			continue
		}
		body := def.GetBody()
		if body == nil {
			continue
		}
		if def.GetLevel() >= TLCLevelTemporal {
			p.addConfigError(ECTLCConfigIDHasValue, "possible", name, "a temporal formula; only state- and action-level predicates are supported")
			continue
		}
		track := NewPossibleTrackNode(body, name)
		SetSemanticToolObjectForToolID(p.ownedToolID(), track, def)
		if tool.GetLevelBound(body, EmptyContext) <= TLCLevelState {
			p.ModelConstraints = append(p.ModelConstraints, track)
		} else {
			p.ActionConstraints = append(p.ActionConstraints, track)
		}
		p.appendOwnedPossiblePostCond(NewPossibleAction(NewPossibleCheckNode(name), EmptyContext, def))
	}
}

func (p *SpecProcessor) processConfigSymmetry(tool *Tool) {
	if p == nil || p.ownedConfig() == nil || tool == nil {
		return
	}
	name := p.ownedConfig().GetSymmetry()
	if name == "" {
		tool.SetSymmetryPermutations(nil)
		return
	}
	def, ok := p.defnFrom(p.preConstantDefinitions(), name).(*OpDefNode)
	if !ok || def == nil {
		p.addConfigError(ECTLCConfigSpecifiedNotDefined, "symmetry function", name)
		tool.SetSymmetryPermutations(nil)
		return
	}
	if def.Arity() != 0 {
		p.addConfigError(ECTLCConfigIDRequiresNoArg, name)
		tool.SetSymmetryPermutations(nil)
		return
	}
	body := def.GetBody()
	if body == nil {
		p.addConfigError(ECTLCConfigIDHasValue, "symmetry function", name, "a set of functions")
		tool.SetSymmetryPermutations(nil)
		return
	}
	value, err := tool.Eval(body, EmptyContext, EmptyState, EmptyState, EvalConst, DoNotRecordCostModel)
	if err != nil {
		p.addConfigError(ECTLCConfigIDHasValue, "symmetry function", name, err.Error())
		tool.SetSymmetryPermutations(nil)
		return
	}
	setValue, ok := value.(*SetEnumValue)
	if !ok {
		p.addConfigError(ECTLCConfigIDHasValue, "symmetry function", name, "a set of functions")
		tool.SetSymmetryPermutations(nil)
		return
	}
	if !symmetryValuesAreFunctionRecords(setValue) {
		p.addConfigError(ECTLCConfigIDHasValue, "symmetry function", name, "function records")
		tool.SetSymmetryPermutations(nil)
		return
	}
	if offenders := p.symmetryTooSmallOffenders(body, tool, setValue.Elems.Len(), nil); len(offenders) > 0 {
		printSymmetrySetTooSmallWarning(offenders)
		tool.SetSymmetryPermutations(nil)
		return
	}
	perms, err := PermutationSubgroup(setValue)
	if err != nil {
		p.addConfigError(ECTLCConfigIDHasValue, "symmetry function", name, err.Error())
		tool.SetSymmetryPermutations(nil)
		return
	}
	if offenders := p.symmetryTooSmallOffenders(body, tool, setValue.Elems.Len(), perms); len(offenders) > 0 {
		printSymmetrySetTooSmallWarning(offenders)
	}
	tool.SetSymmetryPermutations(perms)
}

func symmetryValuesAreFunctionRecords(setValue *SetEnumValue) bool {
	if setValue == nil {
		return false
	}
	for _, elem := range ValueEnumerationAll(setValue.Elements()) {
		if _, ok := elem.(*FcnRcdValue); !ok {
			return false
		}
	}
	return true
}

func (p *SpecProcessor) symmetryTooSmallOffenders(body SemanticNode, tool *Tool, valueCount int, subgroup []*MVPerm) []string {
	appl, ok := body.(*OpApplNode)
	if !ok || appl == nil || len(appl.Args) == 0 {
		return nil
	}
	if subgroup == nil {
		if len(appl.Args) < valueCount {
			return nil
		}
		offenders := make([]string, 0, len(appl.Args))
		for _, arg := range appl.Args {
			offenders = append(offenders, p.symmetryArgumentDisplayName(arg))
		}
		return offenders
	}
	members := NewInsMap[*ModelValue, struct{}]()
	for _, perm := range subgroup {
		if perm == nil {
			continue
		}
		for _, mv := range perm.GetAllModelValues() {
			members.Set(mv, struct{}{})
		}
	}
	var offenders []string
	for _, arg := range appl.Args {
		set := p.symmetrySetEnumFromArgumentNode(tool, arg)
		if set == nil {
			continue
		}
		found := false
		enum := set.Elements()
		for elem := enum.NextElement(); elem != nil; elem = enum.NextElement() {
			mv, ok := elem.(*ModelValue)
			if !ok {
				continue
			}
			if _, ok := members.Get2(mv); ok {
				found = true
				break
			}
		}
		if !found {
			offenders = append(offenders, p.symmetryArgumentDisplayName(arg))
		}
	}
	return offenders
}

func (p *SpecProcessor) symmetrySetEnumFromArgumentNode(tool *Tool, node SemanticNode) *SetEnumValue {
	appl, ok := node.(*OpApplNode)
	if !ok || appl == nil || appl.Operator == nil || appl.Operator.Name == nil || appl.Operator.Name.String() != "Permutations" || len(appl.Args) != 1 {
		return nil
	}
	arg, ok := appl.Args[0].(*OpApplNode)
	if !ok || arg == nil || arg.Operator == nil {
		return nil
	}
	value := MuxWorkerValue(tool.Lookup(arg.Operator, EmptyContext, EmptyState, false), 0)
	set, _ := value.(*SetEnumValue)
	return set
}

func (p *SpecProcessor) symmetryArgumentDisplayName(node SemanticNode) string {
	name := SemanticString(node)
	// Java extracts the alias from the syntax image, not SemanticNode.toString
	// (which normally reports a source location).
	if semantic, ok := node.(interface{ GetTreeNode() any }); ok {
		if tree, ok := semantic.GetTreeNode().(interface{ GetHumanReadableImage() string }); ok {
			name = tree.GetHumanReadableImage()
		}
	}
	const permutations = "Permutations"
	if strings.HasPrefix(name, permutations) {
		name = name[len(permutations)+1 : len(name)-1]
	}
	if p != nil && p.ownedConfig() != nil {
		if override := p.ownedConfig().GetOverridenSpecNameForConfigNameNullable(javaString(name)); override != nil {
			name = *override
		}
	}
	return name
}

func printSymmetrySetTooSmallWarning(offenders []string) {
	if len(offenders) == 0 {
		return
	}
	plurality := ""
	antiPlurality := "s"
	toHave := "has"
	if len(offenders) > 1 {
		plurality = "s"
		antiPlurality = ""
		toHave = "have"
	}
	PrintWarning(ECTLCSymmetrySetTooSmall, plurality, strings.Join(offenders, ", and "), toHave, antiPlurality)
}

func (p *SpecProcessor) attachSpecPropertyOrigin(action *Action, stack []SemanticNode) {
	if p == nil || p.ownedConfig() == nil || action == nil || len(stack) == 0 {
		return
	}
	properties := p.ownedConfig().GetProperties()
	if len(properties) == 0 {
		return
	}
	for _, node := range stack {
		name := SemanticString(node)
		if semantic, ok := node.(interface{ GetTreeNode() any }); ok {
			if tree, ok := semantic.GetTreeNode().(interface{ GetHumanReadableImage() string }); ok {
				// Java compares SyntaxTreeNode.toString(), not the semantic
				// node's toString(), which normally contains its location.
				name = tree.GetHumanReadableImage()
			}
		}
		for _, property := range properties {
			if name == property {
				action.GetAuxiliary()[specProcessorPropertyAuxKey] = node
				return
			}
		}
	}
}

func (p *SpecProcessor) processSpecPropertyTautologyWarning() {
	if p == nil || p.ownedConfig() == nil || p.SpecificationName == "" || len(p.ownedConfig().GetProperties()) != 1 || len(p.Temporals) == 0 {
		return
	}
	for _, action := range p.Temporals {
		if action == nil || action.GetAuxiliary()[specProcessorPropertyAuxKey] == nil {
			return
		}
	}
	liveName := p.ownedConfig().GetProperties()[0]
	PrintWarning(ECTLCLiveFormulaAndFairnessTautology, liveName, p.opDefLocationString(liveName), p.SpecificationName, p.opDefLocationString(p.SpecificationName))
}

func (p *SpecProcessor) opDefLocationString(name string) string {
	if op, ok := p.defn(name).(*OpDefNode); ok && op != nil {
		if loc := semanticNodeLocationString(op); loc != NullSourceLocation.String() {
			return loc
		}
		if body := op.GetBody(); body != nil {
			if loc := semanticNodeLocationString(body); loc != NullSourceLocation.String() {
				return loc
			}
			return SemanticString(body)
		}
	}
	return name
}

func (p *SpecProcessor) processMissingInitNextConfig() {
	if p == nil {
		return
	}
	if p.InitPred.Size() == 0 && (len(p.ImpliedInits) != 0 || len(p.ImpliedActions) != 0 || len(p.Variables) != 0 || len(p.Invariants) != 0 || len(p.ImpliedTemporals) != 0) {
		p.addConfigError(ECTLCConfigMissingInit)
	}
	if p.NextPred == nil && (len(p.ImpliedActions) != 0 || len(p.Invariants) != 0 || len(p.ImpliedTemporals) != 0) {
		p.addConfigError(ECTLCConfigMissingNext)
	}
}

func (p *SpecProcessor) configProcessingTool() *Tool {
	tool := NewTool()
	tool.ID = p.ownedToolID()
	p.applyDefinitionsToTool(tool)
	return tool
}

func (p *SpecProcessor) processConfigSpec(tool *Tool, pred SemanticNode, c *Context, subs *List, stack []SemanticNode) {
	if p == nil || pred == nil {
		return
	}
	if c == nil {
		c = EmptyContext
	}
	if subs == nil {
		subs = EmptyList
	}
	if tool == nil {
		tool = p.configProcessingTool()
	}
	switch node := pred.(type) {
	case *SubstInNode:
		p.processConfigSpec(tool, node.GetBody(), c, subs.Cons(node), stack)
		return
	case *APSubstInNode:
		p.processConfigSpec(tool, node.GetBody(), c, subs, stack)
		return
	case *LetInNode:
		p.processConfigSpec(tool, node.Body, c, subs, stack)
		return
	case *LabelNode:
		p.processConfigSpec(tool, node.Body, c, subs, stack)
		return
	case *OpApplNode:
		stack = append(stack, node)
		if p.processConfigSpecAppl(tool, node, c, subs, stack) {
			return
		}
	}

	level := tool.GetLevelBound(pred, c)
	if level <= TLCLevelState {
		p.InitPred.AddElement(NewUnnamedAction(SpecsAddSubsts(pred, subs), c))
		return
	}
	if level == TLCLevelTemporal {
		action := NewUnnamedAction(SpecsAddSubsts(pred, subs), c)
		p.attachSpecPropertyOrigin(action, stack)
		p.Temporals = append(p.Temporals, action)
		p.TemporalNames = append(p.TemporalNames, SemanticString(pred))
		return
	}
	p.addConfigError(ECTLCCantHandleConjunct, SemanticString(pred))
}

func (p *SpecProcessor) processConfigSpecAppl(tool *Tool, pred *OpApplNode, c *Context, subs *List, stack []SemanticNode) bool {
	if pred == nil || pred.Operator == nil {
		return false
	}
	args := pred.Args
	val := tool.LookupSymbolValue(pred.Operator, c, false)
	if len(args) == 0 {
		switch v := val.(type) {
		case *OpDefNode:
			if v == nil || v.Arity() != 0 {
				p.addConfigError(ECTLCConfigIDRequiresNoArg, opNodeName(pred.Operator))
				return true
			}
			body := v.GetBody()
			if tool.GetLevelBound(body, c) == TLCLevelState {
				p.InitPred.AddElement(NewActionFromOpDef(SpecsAddSubsts(body, subs), c, v, true, false))
				return true
			}
			p.processConfigSpec(tool, body, c, subs, stack)
			return true
		case *BoolValue:
			if !v.Val {
				p.addConfigError(ECTLCConfigSpecIsTrivial, opNodeName(pred.Operator))
			}
			return true
		case *LazyValue:
			p.processConfigSpec(tool, v.ownedExpr(), v.ownedContext(), subs, stack[:len(stack)-1])
			return true
		case *LazySupplierValue:
			lv := asLazyValue(v)
			p.processConfigSpec(tool, lv.ownedExpr(), lv.ownedContext(), subs, stack[:len(stack)-1])
			return true
		case nil:
			p.addConfigError(ECTLCConfigSpecifiedNotDefined, "specification", opNodeName(pred.Operator))
			return true
		case *SymbolNode:
			p.addConfigError(ECTLCConfigOpIsEqual, opNodeName(pred.Operator), semanticNodeLocationString(v), "spec")
			return true
		default:
			p.addConfigError(ECTLCConfigOpIsEqual, opNodeName(pred.Operator), configValueString(v), "spec")
			return true
		}
	}

	if def, ok := val.(*OpDefNode); ok && def != nil && def.GetBody() != nil && !def.GetInRecursive() && def.Arity() == len(args) && subs.IsEmpty() {
		if c1, err := tool.GetOpContext(def, args, c, false, DoNotRecordCostModel); err == nil {
			p.processConfigSpec(tool, def.GetBody(), c1, subs, stack)
			return true
		}
	}

	opcode := 0
	if pred.Operator.Name != nil {
		opcode = GetOpCode(pred.Operator.Name)
	}
	switch opcode {
	case OpcodeTE, OpcodeTF:
		p.addConfigError(ECTLCSpecificationFeaturesTemporalQuantifier)
		return true
	case OpcodeCL, OpcodeLand:
		for _, arg := range args {
			p.processConfigSpec(tool, arg, c, subs, append([]SemanticNode(nil), stack...))
		}
		return true
	case OpcodeBox:
		if len(args) == 0 {
			return true
		}
		boxArg, _ := args[0].(*OpApplNode)
		if boxArg != nil && tool.GetLevelBound(boxArg, c) <= TLCLevelState {
			p.addConfigError(ECTLCLiveCannotHandleFormula, SemanticString(boxArg))
			return true
		}
		if boxArg != nil && boxArg.Operator != nil && GetOpCode(boxArg.Operator.Name) == OpcodeSA && len(boxArg.Args) > 0 {
			if len(boxArg.Args) > 1 {
				p.checkNextStateSubscript(tool, boxArg.Args[1], c, subs)
			}
			if p.NextPred == nil {
				p.NextPred = NewUnnamedAction(SpecsAddSubsts(boxArg.Args[0], subs), c)
			} else {
				p.addConfigError(ECTLCCantHandleTooManyNextStateRels)
			}
			return true
		}
		// Other boxed formulas reach the common level-based handler, which
		// also records their originating property from the complete stack.
		return false
	case OpcodeNop:
		if len(args) > 0 {
			p.processConfigSpec(tool, args[0], c, subs, stack)
			return true
		}
	}
	return false
}

func (p *SpecProcessor) processConfigProperty(tool *Tool, name string, configName string, pred SemanticNode, c *Context, subs *List) {
	if p == nil || pred == nil {
		return
	}
	if c == nil {
		c = EmptyContext
	}
	if subs == nil {
		subs = EmptyList
	}
	if tool == nil {
		tool = p.configProcessingTool()
	}
	switch node := pred.(type) {
	case *SubstInNode:
		p.processConfigProperty(tool, name, configName, node.GetBody(), c, subs.Cons(node))
		return
	case *APSubstInNode:
		p.processConfigProperty(tool, name, configName, node.GetBody(), c, subs)
		return
	case *LetInNode:
		p.processConfigProperty(tool, name, configName, node.Body, c, subs)
		return
	case *LabelNode:
		p.processConfigProperty(tool, name, configName, node.Body, c, subs)
		return
	case *OpApplNode:
		if p.processConfigPropertyAppl(tool, name, configName, node, c, subs) {
			return
		}
	}
	level := tool.GetLevelBound(pred, c)
	switch level {
	case TLCLevelConstant, TLCLevelState:
		p.ImpliedInits = append(p.ImpliedInits, NewAction(SpecsAddSubsts(pred, subs), c, configName))
		p.ImpliedInitNames = append(p.ImpliedInitNames, name)
	case TLCLevelAction:
		switch appl := pred.(type) {
		case *OpApplNode:
			opcode := 0
			if appl != nil && appl.Operator != nil && appl.Operator.Name != nil {
				opcode = GetOpCode(appl.Operator.Name)
			}
			switch opcode {
			case OpcodeSA:
				p.addConfigError(ECTLCConfigPropertyActionLevelSquareASubV, name, semanticNodeLocationString(pred))
			case OpcodeAA:
				p.addConfigError(ECTLCConfigPropertyActionLevelAngleASubV, name, semanticNodeLocationString(pred))
			default:
				p.addConfigError(ECTLCConfigPropertyActionLevel, name, semanticNodeLocationString(pred))
			}
		default:
			p.addConfigError(ECTLCConfigPropertyActionLevel, name, semanticNodeLocationString(pred))
		}
	case TLCLevelTemporal:
		p.ImpliedTemporals = append(p.ImpliedTemporals, NewAction(SpecsAddSubsts(pred, subs), c, configName))
		p.ImpliedTempNames = append(p.ImpliedTempNames, name)
	default:
		p.addConfigError(ECTLCConfigPropertyNotCorrectlyDefined, name)
	}
}

func (p *SpecProcessor) processConfigPropertyAppl(tool *Tool, name string, configName string, pred *OpApplNode, c *Context, subs *List) bool {
	if pred == nil || pred.Operator == nil {
		return false
	}
	args := pred.Args
	opNode := pred.Operator
	val := tool.LookupSymbolValue(opNode, c, false)
	if len(args) == 0 && !opNode.IsVariableDecl() {
		switch v := val.(type) {
		case *OpDefNode:
			if v == nil || v.Arity() != 0 {
				p.addConfigError(ECTLCConfigIDRequiresNoArg, opNodeName(opNode))
				return true
			}
			opName := name
			if opNode.Name != nil {
				opName = opNode.Name.String()
			}
			p.processConfigProperty(tool, opName, configName, v.GetBody(), c, subs)
			return true
		case *BoolValue:
			if !v.Val {
				p.addConfigError(ECTLCConfigSpecIsTrivial, opNodeName(opNode))
			}
			return true
		case *LazyValue:
			p.processConfigProperty(tool, name, configName, v.ownedExpr(), v.ownedContext(), subs)
			return true
		case *LazySupplierValue:
			lv := asLazyValue(v)
			p.processConfigProperty(tool, name, configName, lv.ownedExpr(), lv.ownedContext(), subs)
			return true
		case nil:
			p.addConfigError(ECTLCConfigSpecifiedNotDefined, "property", opNodeName(opNode))
			return true
		default:
			p.addConfigError(ECTLCConfigOpIsEqual, opNodeName(opNode), configValueString(v), "property")
			return true
		}
	}
	if def, ok := val.(*OpDefNode); ok && def != nil && def.GetBody() != nil && !def.GetInRecursive() && def.Arity() == len(args) && subs.IsEmpty() {
		if c1, err := tool.GetOpContext(def, args, c, false, DoNotRecordCostModel); err == nil {
			opName := name
			if opNode.Name != nil {
				opName = opNode.Name.String()
			}
			p.processConfigProperty(tool, opName, configName, def.GetBody(), c1, subs)
			return true
		}
	}
	opcode := 0
	if opNode.Name != nil {
		opcode = GetOpCode(opNode.Name)
	}
	switch opcode {
	case OpcodeBF:
		if len(args) > 0 {
			if ctxts, err := tool.Contexts(pred, c, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel); err == nil && ctxts != nil && ctxts.Err() == nil {
				if ctxts.IsDone() {
					p.addConfigError(ECTLCLiveFormulaTautology)
				} else {
					for c1 := ctxts.NextElement(); c1 != nil; c1 = ctxts.NextElement() {
						if err := ctxts.Err(); err != nil {
							panic(err)
						}
						p.processConfigProperty(tool, SemanticString(args[0]), configName, args[0], c1, subs)
					}
					if err := ctxts.Err(); err != nil {
						panic(err)
					}
				}
				return true
			}
		}
	case OpcodeCL, OpcodeLand:
		for _, arg := range args {
			p.processConfigProperty(tool, SemanticString(arg), configName, arg, c, subs)
		}
		return true
	case OpcodeBox:
		if len(args) == 0 {
			return true
		}
		boxArg := args[0]
		if boxAppl, ok := boxArg.(*OpApplNode); ok && boxAppl.Operator != nil && GetOpCode(boxAppl.Operator.Name) == OpcodeSA {
			actName := name
			if len(boxAppl.Args) == 0 && boxAppl.Operator.Name != nil {
				actName = boxAppl.Operator.Name.String()
			}
			p.ImpliedActions = append(p.ImpliedActions, NewAction(SpecsAddSubsts(boxArg, subs), c, configName))
			p.ImpliedActNames = append(p.ImpliedActNames, actName)
			return true
		}
		if tool.GetLevelBound(boxArg, c) < TLCLevelAction {
			if boxAppl, ok := boxArg.(*OpApplNode); ok && boxAppl.Operator != nil && len(boxAppl.Args) == 0 && boxAppl.Operator.Name != nil {
				name = boxAppl.Operator.Name.String()
			}
			p.Invariants = append(p.Invariants, NewAction(SpecsAddSubsts(boxArg, subs), c, configName))
			p.InvariantNames = append(p.InvariantNames, name)
			return true
		}
		p.ImpliedTemporals = append(p.ImpliedTemporals, NewAction(SpecsAddSubsts(pred, subs), c, configName))
		p.ImpliedTempNames = append(p.ImpliedTempNames, name)
		return true
	case OpcodeNop:
		if len(args) > 0 {
			p.processConfigProperty(tool, name, configName, args[0], c, subs)
			return true
		}
	}
	return false
}

func (p *SpecProcessor) applyDefinitionsToTool(tool *Tool) {
	if p == nil || p.ownedDefns() == nil || tool == nil {
		return
	}
	if tool.DefnsByName == nil {
		tool.DefnsByName = make(map[*UniqueString]any)
	}
	for name, value := range p.ownedDefns().All() {
		if name == nil || value == nil {
			continue
		}
		tool.DefnsByName[name] = value
		if opDef, ok := value.(*OpDefNode); ok && opDef != nil && opDef.Symbol != nil {
			if tool.Definitions == nil {
				tool.Definitions = make(map[*SymbolNode]any)
			}
			tool.Definitions[opDef.Symbol] = value
			// Defns retains the semantic INSTANCE node. The evaluator's name
			// cache exposes native lookup through that node's shared body.
			switch native := tool.Lookup(opDef.Symbol, EmptyContext, nil, false).(type) {
			case *MethodValue, *EvaluatingValue, *PriorityEvaluatingValue, *CallableValue:
				tool.DefnsByName[name] = native
			}
		}
	}
}

func (p *SpecProcessor) defn(name string) any {
	if p == nil || p.ownedDefns() == nil || name == "" {
		return nil
	}
	return p.ownedDefns().Get(name)
}

func (p *SpecProcessor) defnFrom(defns *Defns, name string) any {
	if name == "" {
		return nil
	}
	if defns == nil {
		return p.defn(name)
	}
	return defns.Get(name)
}

func (p *SpecProcessor) preConstantDefinitions() *Defns {
	if p == nil || p.Snapshot == nil {
		if p == nil {
			return nil
		}
		return p.ownedDefns()
	}
	return p.Snapshot
}

func (p *SpecProcessor) actionFromConfigName(name string, init bool, kind string) *Action {
	def, ok := p.configOpDef(name, kind, p.ownedDefns())
	if !ok {
		return nil
	}
	return NewActionFromOpDef(def.GetBody(), EmptyContext, def, init, false)
}

func (p *SpecProcessor) optionalOpBodyFromConfigName(name string, kind string, defns *Defns) SemanticNode {
	if name == "" {
		return nil
	}
	def, ok := p.configOpDef(name, kind, defns)
	if !ok {
		return nil
	}
	return def.GetBody()
}

func (p *SpecProcessor) configOpDef(name string, kind string, defns *Defns) (*OpDefNode, bool) {
	switch def := p.defnFrom(defns, name).(type) {
	case *OpDefNode:
		if def == nil {
			p.addConfigError(ECTLCConfigSpecifiedNotDefined, kind, name)
			return nil, false
		}
		if def.Arity() != 0 {
			p.addConfigError(ECTLCConfigIDRequiresNoArg, kind, name)
			return nil, false
		}
		return def, true
	case Value:
		p.addConfigError(ECTLCConfigIDMustNotBeConstant, kind, name)
		return nil, false
	case nil:
		p.addConfigError(ECTLCConfigSpecifiedNotDefined, kind, name)
		return nil, false
	default:
		p.addConfigError(ECTLCConfigIDMustNotBeConstant, kind, name)
		return nil, false
	}
}

func (p *SpecProcessor) configPredicateOpDef(name string, kind string, defns *Defns) (*OpDefNode, bool) {
	switch def := p.defnFrom(defns, name).(type) {
	case *OpDefNode:
		if def == nil {
			p.addConfigError(ECTLCConfigSpecifiedNotDefined, kind, name)
			return nil, false
		}
		if def.Arity() != 0 {
			p.addConfigError(ECTLCConfigIDRequiresNoArg, kind, name)
			return nil, false
		}
		return def, true
	case *BoolValue:
		if def.Val {
			return nil, false
		}
		p.addConfigError(ECTLCConfigIDHasValue, kind, name, def.String())
		return nil, false
	case Value:
		p.addConfigError(ECTLCConfigIDHasValue, kind, name, def.String())
		return nil, false
	case nil:
		p.addConfigError(ECTLCConfigSpecifiedNotDefined, kind, name)
		return nil, false
	default:
		p.addConfigError(ECTLCConfigIDHasValue, kind, name, configValueString(def))
		return nil, false
	}
}

func configValueString(value any) string {
	if value == nil {
		return "<nil>"
	}
	if s, ok := value.(interface{ String() string }); ok {
		return s.String()
	}
	return fmt.Sprint(value)
}

func opNodeName(node *SymbolNode) string {
	if node == nil || node.Name == nil {
		return "<unknown>"
	}
	return node.Name.String()
}

func semanticNodeLocationString(node SemanticNode) string {
	if loc, ok := semanticNodeSourceLocation(node); ok {
		return loc.String()
	}
	return NullSourceLocation.String()
}

func (p *SpecProcessor) constraintNodesFromConfigNames(names []string, kind string, noArgCode int, undefinedCode int, valueCode int) []SemanticNode {
	nodes := make([]SemanticNode, 0, len(names))
	for _, name := range names {
		switch def := p.defn(name).(type) {
		case *OpDefNode:
			if def == nil {
				p.addConfigError(undefinedCode, kind, name)
			} else if def.Arity() != 0 {
				p.addConfigError(noArgCode, kind, name)
			} else if body := def.GetBody(); body != nil {
				SetSemanticToolObjectForToolID(p.ownedToolID(), body, def)
				nodes = append(nodes, body)
			}
		case *BoolValue:
			if !def.Val {
				p.addConfigError(valueCode, kind, name, def.String())
			}
		case SemanticNode:
			p.addConfigError(valueCode, kind, name, SemanticString(def))
		case Value:
			p.addConfigError(valueCode, kind, name, def.String())
		default:
			p.addConfigError(undefinedCode, kind, name)
		}
	}
	return nodes
}

func (p *SpecProcessor) addConfigError(code int, params ...string) {
	if p == nil {
		return
	}
	failure := NewConfigError(code, params...)
	p.ConfigErrors = append(p.ConfigErrors, failure)
	panic(failure)
}

// Unwind exactly the configuration failure recorded by this processor. Keep
// native diagnostics at the public boundary while Java's Assert.fail ordering
// prevents later config phases, runtime compilation and mutations from running.
func (p *SpecProcessor) recoverConfigFailure() {
	if failure := recover(); failure != nil {
		last := len(p.ConfigErrors) - 1
		if last < 0 || failure != p.ConfigErrors[last] {
			panic(failure)
		}
	}
}
