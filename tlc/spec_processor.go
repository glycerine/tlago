package tlc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const specProcessorVetoedProperty = "tlc2.tool.impl.SpecProcessor.vetoed"
const specProcessorPropertyAuxKey = "tlc.specProcessor.property"

type Defns struct {
	defnIdx int
	table   []any
	order   *InsMap[*UniqueString, struct{}]
}

func NewDefns(initialSize ...int) *Defns {
	idx := 0
	if len(initialSize) > 0 {
		idx = initialSize[0]
	}
	return &Defns{defnIdx: idx, table: make([]any, idx+32), order: NewInsMap[*UniqueString, struct{}]()}
}

func (d *Defns) Put(key any, value any) {
	if d == nil {
		return
	}
	us := defnKey(key)
	if us == nil {
		return
	}
	loc := us.DefnLoc()
	if loc == -1 {
		loc = d.defnIdx
		d.defnIdx++
		us.SetLoc(loc)
	}
	if loc >= len(d.table) {
		old := d.table
		newSize := max(2*len(old), loc+1)
		d.table = make([]any, newSize)
		copy(d.table, old)
	}
	d.table[loc] = value
	if d.order == nil {
		d.order = NewInsMap[*UniqueString, struct{}]()
	}
	d.order.Set(us, struct{}{})
}

func (d *Defns) Get(key any) any {
	if d == nil {
		return nil
	}
	us := defnKey(key)
	if us == nil {
		return nil
	}
	loc := us.DefnLoc()
	if loc < 0 || loc >= len(d.table) {
		return nil
	}
	return d.table[loc]
}

func (d *Defns) SetDefnCount(index int) {
	if d != nil {
		d.defnIdx = index
	}
}

func (d *Defns) Snapshot() *Defns {
	if d == nil {
		return NewDefns()
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
		return nil
	case *UniqueString:
		return k
	case UniqueString:
		return UniqueStringOf(k.String())
	case string:
		return UniqueStringOf(k)
	case *SymbolNode:
		if k == nil {
			return nil
		}
		return k.Name
	default:
		return UniqueStringOf(fmt.Sprint(k))
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
	spec.Processor = NewSpecProcessor(spec.RootName(), spec.Defns, config)
	spec.UnprocessedDefn = spec.Processor.UnprocessedDefns
	return spec
}

func (s *Spec) RootName() string {
	if s == nil || s.RootFile == "" {
		return "Spec"
	}
	base := filepath.Base(s.RootFile)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if base == "" {
		return "Spec"
	}
	return base
}

func (s *Spec) GetModelConfig() *ModelConfig {
	if s == nil {
		return nil
	}
	return s.Config
}

func (s *Spec) GetSpecProcessor() *SpecProcessor {
	if s == nil {
		return nil
	}
	return s.Processor
}

func (s *Spec) ApplyToTool(tool *Tool) *Tool {
	if tool == nil {
		tool = NewTool()
	}
	if s == nil || s.Processor == nil {
		return tool
	}
	s.Processor.ApplyToTool(tool)
	tool.RootName = s.RootName()
	tool.RootFile = s.RootFile
	tool.ConfigFile = s.ConfigFile
	tool.SpecDir = s.SpecDir
	tool.ModelConfig = s.Config
	tool.SpecProcessor = s.Processor
	if s.RootFile != "" && len(tool.ModuleFiles) == 0 {
		tool.ModuleFiles = []string{s.RootFile}
	}
	return tool
}

type SpecProcessor struct {
	RootFile   string
	ToolID     int32
	Defns      *Defns
	Config     *ModelConfig
	ModuleTbl  *ExternalModuleTable
	RootModule *ModuleNode

	ConstantDeclarations []*SymbolNode
	RootDefinitions      *InsMap[string, *OpDefNode]

	// Runtime targets resolve against ModuleTbl when their source phase runs.
	RuntimeParameters        RuntimeParameters
	RuntimeInvariantCompiler func(string) (*OpDefNode, error)

	Variables         []*UniqueString
	VariablesNodes    []*SymbolNode
	ProcessedDefs     *InsMap[*OpDefNode, struct{}]
	UnprocessedDefns  *Defns
	ConstantDefns     *InsMap[*ModuleNode, *InsMap[SemanticNode, any]]
	Snapshot          *Defns
	PreConstantSnap   *Defns
	InitPred          []*Action
	NextPred          *Action
	Temporals         []*Action
	TemporalNames     []string
	ImpliedTemporals  []*Action
	ImpliedTempNames  []string
	Invariants        []*Action
	InvariantNames    []string
	ImpliedInits      []*Action
	ImpliedInitNames  []string
	ImpliedActions    []*Action
	ImpliedActNames   []string
	ModelConstraints  []SemanticNode
	ActionConstraints []SemanticNode
	ConfigErrors      []*ConfigError
	PossiblePostConds []*Action
	Assumptions       []SemanticNode
	AssumptionIsAxiom []bool
	RLReward          SemanticNode
	Periodic          SemanticNode
	SymmetrySpec      string
	SpecificationName string
}

func (p *SpecProcessor) GetModuleTbl() *ExternalModuleTable { return p.ModuleTbl }
func (p *SpecProcessor) GetRootModule() *ModuleNode         { return p.RootModule }

func NewSpecProcessor(rootFile string, defns *Defns, config *ModelConfig) *SpecProcessor {
	if defns == nil {
		defns = NewDefns()
	}
	if config == nil {
		config = newModelConfig("", false)
	}
	p := &SpecProcessor{
		RootFile:         rootFile,
		ToolID:           specToolID(),
		Defns:            defns,
		Config:           config,
		ProcessedDefs:    NewInsMap[*OpDefNode, struct{}](),
		UnprocessedDefns: NewDefns(),
		ConstantDefns:    NewInsMap[*ModuleNode, *InsMap[SemanticNode, any]](),
	}
	p.PreConstantSnap = p.Defns.Snapshot()
	p.Snapshot = p.Defns.Snapshot()
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
	if p.Defns != nil {
		p.Defns.SetDefnCount(len(nodes))
		// SpecProcessor installs these predefined values before user definitions.
		p.Defns.Put("TRUE", BoolTrue)
		p.Defns.Put("FALSE", BoolFalse)
		p.Defns.Put("BOOLEAN", NewSetEnumValue([]Value{BoolFalse, BoolTrue}, true))
		stringMethod := NewMethodValue("public static tlc2.value.impl.Value tlc2.module.Strings.STRING()", 0, func(_ []Value, _ int) (Value, error) { return STRING(), nil })
		p.Defns.Put("STRING", stringMethod)
		// Java snapshots just these predefined/native entries before processing
		// source modules, ordinary definitions, native overrides or config.
		p.PreConstantSnap = p.Defns.Snapshot()
	}
}

func (p *SpecProcessor) ProcessConfig() {
	if p == nil || p.Config == nil {
		return
	}
	defer p.recoverConfigFailure()
	p.resetProcessedConfig()
	p.SpecificationName = p.Config.GetSpec()
	p.SymmetrySpec = p.Config.GetSymmetry()

	p.processConfigInvariants()
	if p.SpecificationName != "" {
		if p.Config.GetInit() != "" || p.Config.GetNext() != "" {
			p.addConfigError(ECTLCConfigNotBothSpecAndInit)
		} else {
			p.processSpecificationConfig()
		}
	} else {
		if initName := p.Config.GetInit(); initName != "" {
			if action := p.actionFromConfigName(initName, true, "initial predicate"); action != nil {
				p.InitPred = append(p.InitPred, action)
			}
		}
		if nextName := p.Config.GetNext(); nextName != "" {
			p.NextPred = p.actionFromConfigName(nextName, false, "next state action")
		}
	}
	p.processConfigProperties()
	p.processMissingInitNextConfig()
	p.processSpecPropertyTautologyWarning()
	for _, invariant := range p.RuntimeParameters.Invariants {
		if p.RuntimeInvariantCompiler == nil {
			panic(NewTLCRuntimeException(ECTLCParsingFailed2, "runtime invariant expression compiler is not configured"))
		}
		definition, err := p.RuntimeInvariantCompiler(invariant.Expression)
		if err != nil {
			panic(NewTLCRuntimeExceptionWithCause(ECTLCParsingFailed2, err))
		}
		if definition == nil {
			panic(NewNullPointerException())
		}
		action := NewActionFromOpDef(definition.Body, EmptyContext, definition, false, true)
		p.Invariants = append(p.Invariants, action)
		p.InvariantNames = append(p.InvariantNames, action.GetNameOfDefault())
	}
	for _, constraint := range p.RuntimeParameters.Constraints {
		definition := p.runtimeModuleDefinition(constraint.Module, constraint.Operator)
		p.Defns.Put(definition.Name, definition)
		p.Config.constraints = append(p.Config.constraints, definition.Name.String())
	}
	p.ModelConstraints = p.constraintNodesFromConfigNames(p.Config.GetConstraints(), "constraint", ECTLCConfigIDRequiresNoArg, ECTLCConfigSpecifiedNotDefined, ECTLCConfigIDHasValue)
	for _, constraint := range p.RuntimeParameters.ActionConstraints {
		definition := p.runtimeModuleDefinition(constraint.Module, constraint.Operator)
		p.Defns.Put(definition.Name, definition)
		p.Config.actionConstraints = append(p.Config.actionConstraints, definition.Name.String())
	}
	p.ActionConstraints = p.constraintNodesFromConfigNames(p.Config.GetActionConstraints(), "action constraint", ECTLCConfigIDRequiresNoArg, ECTLCConfigSpecifiedNotDefined, ECTLCConfigIDHasValue)
	p.RLReward = p.optionalOpBodyFromConfigName(p.Config.GetRLReward(), "rlreward", p.preConstantDefinitions())
	p.Periodic = p.optionalOpBodyFromConfigName(p.Config.GetPeriodic(), "periodic", p.preConstantDefinitions())
	p.processConfigPossible()
}

func (p *SpecProcessor) resetProcessedConfig() {
	p.ConfigErrors = nil
	p.InitPred = nil
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
	p.PossiblePostConds = nil
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
	p.ToolID = tool.ID
	tool.SpecProcessor = p
	tool.ModelConfig = p.Config
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
	p.Snapshot = p.Defns.Snapshot()
	p.ProcessConstantDefinitions(tool)
	p.ProcessConfig()
	if len(p.ConfigErrors) != 0 {
		return
	}
	p.applyDefinitionsToTool(tool)
	tool.ModelConfig = p.Config
	tool.InitStateSpec = append([]*Action(nil), p.InitPred...)
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
	if p == nil {
		return nil
	}
	return p.VariablesNodes
}

func (p *SpecProcessor) GetInitPred() []*Action {
	if p == nil {
		return nil
	}
	return append([]*Action(nil), p.InitPred...)
}

func (p *SpecProcessor) GetNextPred() *Action {
	if p == nil {
		return nil
	}
	return p.NextPred
}

func (p *SpecProcessor) GetTemporal() []*Action {
	if p == nil {
		return nil
	}
	return p.Temporals
}

func (p *SpecProcessor) GetTemporalNames() []string {
	if p == nil {
		return nil
	}
	return p.TemporalNames
}

func (p *SpecProcessor) GetImpliedTemporals() []*Action {
	if p == nil {
		return nil
	}
	return p.ImpliedTemporals
}

func (p *SpecProcessor) GetImpliedTemporalNames() []string {
	if p == nil {
		return nil
	}
	return p.ImpliedTempNames
}

func (p *SpecProcessor) GetInvariants() []*Action {
	if p == nil {
		return nil
	}
	return p.Invariants
}

func (p *SpecProcessor) GetInvariantsNames() []string {
	if p == nil {
		return nil
	}
	return p.InvariantNames
}

func (p *SpecProcessor) GetImpliedInits() []*Action {
	if p == nil {
		return nil
	}
	return p.ImpliedInits
}

func (p *SpecProcessor) GetImpliedInitNames() []string {
	if p == nil {
		return nil
	}
	return p.ImpliedInitNames
}

func (p *SpecProcessor) GetImpliedActions() []*Action {
	if p == nil {
		return nil
	}
	return p.ImpliedActions
}

func (p *SpecProcessor) GetImpliedActionNames() []string {
	if p == nil {
		return nil
	}
	return p.ImpliedActNames
}

func (p *SpecProcessor) GetModelConstraints() []SemanticNode {
	if p == nil {
		return nil
	}
	return p.ModelConstraints
}

func (p *SpecProcessor) GetActionConstraints() []SemanticNode {
	if p == nil {
		return nil
	}
	return p.ActionConstraints
}

func (p *SpecProcessor) GetRLReward() SemanticNode {
	if p == nil {
		return nil
	}
	return p.RLReward
}

func (p *SpecProcessor) GetPeriodic() SemanticNode {
	if p == nil {
		return nil
	}
	return p.Periodic
}

func (p *SpecProcessor) GetAssumptions() []SemanticNode {
	if p == nil {
		return nil
	}
	return append([]SemanticNode(nil), p.Assumptions...)
}

func (p *SpecProcessor) GetAssumptionIsAxiom() []bool {
	if p == nil {
		return nil
	}
	return append([]bool(nil), p.AssumptionIsAxiom...)
}

func (p *SpecProcessor) GetUnprocessedDefns() *Defns {
	if p == nil {
		return nil
	}
	return p.Snapshot
}

func (p *SpecProcessor) GetDefns() *Defns {
	if p == nil {
		return nil
	}
	return p.Defns
}

func (p *SpecProcessor) GetConstantDefns() *InsMap[*ModuleNode, *InsMap[SemanticNode, any]] {
	if p == nil {
		return nil
	}
	return p.ConstantDefns
}

func (p *SpecProcessor) GetPostConditionSpecs() []*Action {
	if p == nil {
		return nil
	}
	result := make([]*Action, 0, len(p.RuntimeParameters.PostConditions)+len(p.PossiblePostConds))
	for _, post := range p.RuntimeParameters.PostConditions {
		definition := p.runtimeModuleDefinition(post.Module, post.Operator)
		result = append(result, NewAction(definition.Body, EmptyContext, post.Operator))
	}
	return append(result, p.PossiblePostConds...)
}

// ParameterizedSpecObj resolves only external modules and their actual OpDefs.
// Lookup aliases, named theorems and inner modules are not substitutes for a
// definition in the requested module's source Context.
func (p *SpecProcessor) runtimeModuleDefinition(module, operator string) *OpDefNode {
	node := p.ModuleTbl.GetModuleNode(UniqueStringOf(module))
	if node == nil {
		panic(NewTLCRuntimeException(ECGeneral, "Could not find module: "+module))
	}
	definition := node.GetOpDef(UniqueStringOf(operator))
	if definition == nil {
		panic(NewTLCRuntimeException(ECGeneral, "Could not find operator: "+operator+" in module: "+module))
	}
	return definition
}

func (p *SpecProcessor) ProcessConstantDefinitions(tool *Tool) {
	if p == nil || p.Defns == nil || p.ModuleTbl == nil {
		return
	}
	if tool == nil {
		tool = p.configProcessingTool()
	} else {
		p.applyDefinitionsToTool(tool)
	}
	vetoes := specProcessorVetoedConstantOperators()
	for _, module := range p.ModuleTbl.GetModuleNodes() {
		if module.ProcessConstantDefns() {
			p.processModuleConstantDefinitions(tool, module, vetoes)
		}
	}
}

func (p *SpecProcessor) recordConstantDefinition(module *ModuleNode, node SemanticNode, value any) {
	if p.ConstantDefns == nil {
		p.ConstantDefns = NewInsMap[*ModuleNode, *InsMap[SemanticNode, any]]()
	}
	definitions := p.ConstantDefns.Get(module)
	if definitions == nil {
		definitions = NewInsMap[SemanticNode, any]()
		p.ConstantDefns.Set(module, definitions)
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
		if !ok || opDef == nil || opDef.Body == nil || tool.GetLevelBound(opDef.Body, EmptyContext) != TLCLevelConstant {
			continue
		}
		if specProcessorConstantVetoed(vetoes, opDef.Name) {
			continue
		}
		value, err := evaluateConstantOperatorDefinition(tool, opDef)
		if err != nil || value == nil {
			continue // Java catches Throwable for ordinary constant definitions.
		}
		opDef.SetToolObjectAt(p.ToolID, value)
		// Only replace the global entry when it still denotes this exact node.
		// Hidden/imported definitions retain their values on the semantic node.
		if p.Defns.Get(opDef.Name) == opDef {
			p.Defns.Put(opDef.Name, value)
			if opDef.HasSource() {
				origin = opDef.GetSource().GetOriginallyDefinedInModuleNode()
			}
			p.recordConstantDefinition(origin, opDef, value)
		}
	}
	// Java does not reapply processConstantDefns eligibility to inner modules.
	for _, inner := range module.GetInnerModules() {
		p.processModuleConstantDefinitions(tool, inner, vetoes)
	}
}

func specProcessorVetoedConstantOperators() map[string]bool {
	out := make(map[string]bool)
	if name, ok := tlcLookupSystemProperty(specProcessorVetoedProperty); ok && name != "" {
		out[name] = true
	}
	for _, name := range strings.Split(os.Getenv("TLAGO_SPEC_PROCESSOR_VETOED"), ",") {
		name = strings.TrimSpace(name)
		if name != "" {
			out[name] = true
		}
	}
	return out
}

// Java processConstantDefns evaluates replacements of declared constants even
// when the replacement is not constant-level. Only a failed evaluation of a
// genuinely nonconstant operator is a configuration error.
func (p *SpecProcessor) processDeclaredConstantDefinitions(tool *Tool, module *ModuleNode) {
	for _, declaration := range module.GetConstantDecls() {
		value := declaration.Data
		if declaration.SemanticBase != nil {
			value = declaration.GetToolObjectAt(p.ToolID)
		}
		if constant, ok := value.(Value); ok {
			InitializeValue(constant)
			p.recordConstantDefinition(module, declaration, constant)
			continue
		}
		opDef, ok := value.(*OpDefNode)
		if !ok || opDef == nil {
			continue
		}
		if opDef.Arity() != declaration.Arity {
			panic(NewTLCRuntimeException(ECTLCConfigWrongSubstitutionNumberOfArgs, declaration.GetName().String(), opDef.Name.String()))
		}
		if opDef.Arity() != 0 {
			continue
		}
		result, err := evaluateDeclaredConstantDefinition(tool, opDef)
		if err != nil {
			if tool.GetLevelBound(opDef.Body, EmptyContext) > TLCLevelConstant {
				addendum := ""
				if !isValueEvalException(err) {
					addendum = " - specifically: " + err.Error()
				}
				panic(NewTLCRuntimeException(ECTLCConfigSubstitutionNonConstant, declaration.GetName().String(), opDef.Name.String(), addendum))
			}
			continue
		}
		opDef.SetToolObjectAt(p.ToolID, result)
		if property, _ := tlcLookupSystemProperty("tlc2.tool.impl.SpecProcessor.aggressiveConstantCaching"); javaBooleanProperty(property) {
			if declaration.SemanticBase != nil {
				declaration.SetToolObjectAt(p.ToolID, result)
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
		return tool.Eval(opDef.Body, EmptyContext, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
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
		return tool.Eval(opDef.Body, EmptyContext, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
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
	if p == nil || p.Config == nil || p.Defns == nil {
		return
	}
	defer p.recoverConfigFailure()
	constants := configConstantsToDefns(p.Config.GetConstants(), p)
	for name, value := range constants.All() {
		// Java attaches a global operator constant assignment to the root
		// OpDefNode itself. Module aliases can still refer to that same node;
		// registering those aliases must not undo the configured replacement.
		if p.RootDefinitions != nil {
			if definition := p.RootDefinitions.Get(name); definition != nil {
				definition.SetToolObjectAt(p.ToolID, value)
			}
		}
		p.putConfigDefinition(name, value, tool)
	}
	p.applyConfigModuleConstants(tool)
	p.applyConfigOverrides(tool)
	p.applyConfigModuleOverrides(tool)
	p.Snapshot = p.Defns.Snapshot()
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
	if p == nil || p.Config == nil || p.Config.GetModConstants() == nil {
		return
	}
	for modName, constants := range p.Config.GetModConstants().All() {
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
	if p == nil || p.Config == nil || p.Config.GetOverrides() == nil {
		return
	}
	// Java visits root OpDefs in semantic context order, then checks leftover
	// override keys. Keep the source graph even after Defns entries are replaced.
	names := NewInsMap[string, struct{}]()
	if p.RootDefinitions != nil {
		for name := range p.RootDefinitions.All() {
			if _, ok := p.Config.GetOverrides().Get2(name); ok {
				names.Set(name, struct{}{})
			}
		}
	}
	for name := range p.Config.GetOverrides().All() {
		names.Set(name, struct{}{})
	}
	for lhs := range names.All() {
		rhs := p.Config.GetOverrides().Get(lhs)
		if lhs == "" {
			continue
		}
		if _, chained := p.Config.GetOverrides().Get2(rhs); chained {
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
			lhsDef.SetToolObjectAt(p.ToolID, rhsVal)
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
	} else if p.Defns != nil {
		for _, value := range p.Defns.All() {
			if def, ok := value.(*OpDefNode); ok {
				nodes = append(nodes, def)
			}
		}
	}
	return nodes
}

func (p *SpecProcessor) applyConfigModuleOverrides(tool *Tool) {
	if p == nil || p.Config == nil || p.Config.GetModOverrides() == nil {
		return
	}
	for modName, overrides := range p.Config.GetModOverrides().All() {
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
				SetSemanticToolObjectForToolID(p.ToolID, lhsDef.Body, rhsVal)
			}
			p.putConfigDefinition(qualified, rhsVal, tool)
		}
	}
}

func (p *SpecProcessor) putConfigDefinition(name string, value any, tool *Tool) {
	if p == nil || p.Defns == nil || name == "" || value == nil {
		return
	}
	old := p.defn(name)
	p.Defns.Put(name, value)

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
			sym.SetToolObjectAt(p.ToolID, value)
		} else {
			sym.Data = value
		}
	}
	tool.Define(sym, value)
}

func (p *SpecProcessor) hasModuleDefinitions(moduleName string) bool {
	if p == nil || p.Defns == nil {
		return false
	}
	if moduleName == p.RootFile {
		return true
	}
	prefix := moduleName + "!"
	for name := range p.Defns.All() {
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
		p.processConfigSpec(tool, spec.Body, EmptyContext, EmptyList, nil)
	case nil:
		p.addConfigError(ECTLCConfigSpecifiedNotDefined, "name", p.SpecificationName)
	default:
		p.addConfigError(ECTLCConfigIDHasValue, "value", p.SpecificationName, configValueString(spec))
	}
}

func (p *SpecProcessor) processConfigInvariants() {
	if p == nil || p.Config == nil {
		return
	}
	for _, name := range p.Config.GetInvariants() {
		if value, ok := p.defn(name).(*BoolValue); ok && value.Val {
			PrintWarning(ECTLCInvariantConstantLevel, name, value.String())
			continue
		}
		def, ok := p.configPredicateOpDef(name, "invariant", p.Defns)
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
		p.Invariants = append(p.Invariants, NewActionFromOpDef(def.Body, EmptyContext, def, false, false))
	}
}

func (p *SpecProcessor) processConfigProperties() {
	if p == nil || p.Config == nil {
		return
	}
	tool := p.configProcessingTool()
	for _, name := range p.Config.GetProperties() {
		def, ok := p.configPredicateOpDef(name, "property", p.Defns)
		if !ok {
			continue
		}
		if def.GetLevel() == TLCLevelState {
			PrintWarning(ECTLCLiveFormulaStateLevel, name)
		}
		p.processConfigProperty(tool, name, name, def.Body, EmptyContext, EmptyList)
	}
}

func (p *SpecProcessor) processConfigPossible() {
	if p == nil || p.Config == nil {
		return
	}
	possibleNames := p.Config.GetPossible()
	if len(possibleNames) == 0 {
		return
	}
	tool := p.configProcessingTool()
	for _, name := range possibleNames {
		def, ok := p.configOpDef(name, "possible", p.Defns)
		if !ok || def.Body == nil {
			continue
		}
		if def.GetLevel() >= TLCLevelTemporal {
			p.addConfigError(ECTLCConfigIDHasValue, "possible", name, "a temporal formula; only state- and action-level predicates are supported")
			continue
		}
		track := NewPossibleTrackNode(def.Body, name)
		SetSemanticToolObjectForToolID(p.ToolID, track, def)
		if tool.GetLevelBound(def.Body, EmptyContext) <= TLCLevelState {
			p.ModelConstraints = append(p.ModelConstraints, track)
		} else {
			p.ActionConstraints = append(p.ActionConstraints, track)
		}
		p.PossiblePostConds = append(p.PossiblePostConds, NewPossibleAction(NewPossibleCheckNode(name), EmptyContext, def))
	}
}

func (p *SpecProcessor) processConfigSymmetry(tool *Tool) {
	if p == nil || p.Config == nil || tool == nil {
		return
	}
	name := p.Config.GetSymmetry()
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
	if def.Body == nil {
		p.addConfigError(ECTLCConfigIDHasValue, "symmetry function", name, "a set of functions")
		tool.SetSymmetryPermutations(nil)
		return
	}
	value, err := tool.Eval(def.Body, EmptyContext, EmptyState, EmptyState, EvalConst, DoNotRecordCostModel)
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
	if offenders := p.symmetryTooSmallOffenders(def.Body, tool, setValue.Elems.Len(), nil); len(offenders) > 0 {
		printSymmetrySetTooSmallWarning(offenders)
		tool.SetSymmetryPermutations(nil)
		return
	}
	perms, err := PermutationSubgroup(value)
	if err != nil {
		p.addConfigError(ECTLCConfigIDHasValue, "symmetry function", name, err.Error())
		tool.SetSymmetryPermutations(nil)
		return
	}
	if offenders := p.symmetryTooSmallOffenders(def.Body, tool, setValue.Elems.Len(), perms); len(offenders) > 0 {
		printSymmetrySetTooSmallWarning(offenders)
	}
	tool.SetSymmetryPermutations(perms)
}

func symmetryValuesAreFunctionRecords(setValue *SetEnumValue) bool {
	if setValue == nil {
		return false
	}
	enum := setValue.Elements()
	for elem := enum.NextElement(); elem != nil; elem = enum.NextElement() {
		if _, ok := elem.(*FcnRcdValue); !ok {
			return false
		}
	}
	return enum.Err() == nil
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
		for _, mv := range perm.AllModelValues() {
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
	if p != nil && p.Config != nil {
		if override := p.Config.GetOverridenSpecNameForConfigName(name); override != "" {
			name = override
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
	if p == nil || p.Config == nil || action == nil || len(stack) == 0 {
		return
	}
	properties := p.Config.GetProperties()
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
	if p == nil || p.Config == nil || p.SpecificationName == "" || len(p.Config.GetProperties()) != 1 || len(p.Temporals) == 0 {
		return
	}
	for _, action := range p.Temporals {
		if action == nil || action.Auxiliary == nil || action.Auxiliary[specProcessorPropertyAuxKey] == nil {
			return
		}
	}
	liveName := p.Config.GetProperties()[0]
	PrintWarning(ECTLCLiveFormulaAndFairnessTautology, liveName, p.opDefLocationString(liveName), p.SpecificationName, p.opDefLocationString(p.SpecificationName))
}

func (p *SpecProcessor) opDefLocationString(name string) string {
	if op, ok := p.defn(name).(*OpDefNode); ok && op != nil {
		if loc := semanticNodeLocationString(op); loc != NullSourceLocation.String() {
			return loc
		}
		if op.Body != nil {
			if loc := semanticNodeLocationString(op.Body); loc != NullSourceLocation.String() {
				return loc
			}
			return SemanticString(op.Body)
		}
	}
	return name
}

func (p *SpecProcessor) processMissingInitNextConfig() {
	if p == nil {
		return
	}
	if len(p.InitPred) == 0 && (len(p.ImpliedInits) != 0 || len(p.ImpliedActions) != 0 || len(p.Variables) != 0 || len(p.Invariants) != 0 || len(p.ImpliedTemporals) != 0) {
		p.addConfigError(ECTLCConfigMissingInit)
	}
	if p.NextPred == nil && (len(p.ImpliedActions) != 0 || len(p.Invariants) != 0 || len(p.ImpliedTemporals) != 0) {
		p.addConfigError(ECTLCConfigMissingNext)
	}
}

func (p *SpecProcessor) configProcessingTool() *Tool {
	tool := NewTool()
	tool.ID = p.ToolID
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
		p.processConfigSpec(tool, node.Body, c, subs.Cons(node), stack)
		return
	case *APSubstInNode:
		p.processConfigSpec(tool, node.Body, c, subs, stack)
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
		p.InitPred = append(p.InitPred, NewAction(SpecsAddSubsts(pred, subs), c, ""))
		return
	}
	if level == TLCLevelTemporal {
		action := NewAction(SpecsAddSubsts(pred, subs), c, "")
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
			if tool.GetLevelBound(v.Body, c) == TLCLevelState {
				p.InitPred = append(p.InitPred, NewActionFromOpDef(SpecsAddSubsts(v.Body, subs), c, v, true, false))
				return true
			}
			p.processConfigSpec(tool, v.Body, c, subs, stack)
			return true
		case *BoolValue:
			if !v.Val {
				p.addConfigError(ECTLCConfigSpecIsTrivial, opNodeName(pred.Operator))
			}
			return true
		case *LazyValue:
			p.processConfigSpec(tool, v.Expr, v.Con, subs, stack[:len(stack)-1])
			return true
		case *LazySupplierValue:
			lv := asLazyValue(v)
			p.processConfigSpec(tool, lv.Expr, lv.Con, subs, stack[:len(stack)-1])
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

	if def, ok := val.(*OpDefNode); ok && def != nil && def.Body != nil && !def.GetInRecursive() && def.Arity() == len(args) && subs.IsEmpty() {
		if c1, err := tool.GetOpContext(def, args, c, false, DoNotRecordCostModel); err == nil {
			p.processConfigSpec(tool, def.Body, c1, subs, stack)
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
			if p.NextPred == nil {
				p.NextPred = NewAction(SpecsAddSubsts(boxArg.Args[0], subs), c, "")
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
		p.processConfigProperty(tool, name, configName, node.Body, c, subs.Cons(node))
		return
	case *APSubstInNode:
		p.processConfigProperty(tool, name, configName, node.Body, c, subs)
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
			p.processConfigProperty(tool, opName, configName, v.Body, c, subs)
			return true
		case *BoolValue:
			if !v.Val {
				p.addConfigError(ECTLCConfigSpecIsTrivial, opNodeName(opNode))
			}
			return true
		case *LazyValue:
			p.processConfigProperty(tool, name, configName, v.Expr, v.Con, subs)
			return true
		case *LazySupplierValue:
			lv := asLazyValue(v)
			p.processConfigProperty(tool, name, configName, lv.Expr, lv.Con, subs)
			return true
		case nil:
			p.addConfigError(ECTLCConfigSpecifiedNotDefined, "property", opNodeName(opNode))
			return true
		default:
			p.addConfigError(ECTLCConfigOpIsEqual, opNodeName(opNode), configValueString(v), "property")
			return true
		}
	}
	if def, ok := val.(*OpDefNode); ok && def != nil && def.Body != nil && !def.GetInRecursive() && def.Arity() == len(args) && subs.IsEmpty() {
		if c1, err := tool.GetOpContext(def, args, c, false, DoNotRecordCostModel); err == nil {
			opName := name
			if opNode.Name != nil {
				opName = opNode.Name.String()
			}
			p.processConfigProperty(tool, opName, configName, def.Body, c1, subs)
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
						p.processConfigProperty(tool, SemanticString(args[0]), configName, args[0], c1, subs)
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
	if p == nil || p.Defns == nil || tool == nil {
		return
	}
	if tool.DefnsByName == nil {
		tool.DefnsByName = make(map[*UniqueString]any)
	}
	for name, value := range p.Defns.All() {
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
	if p == nil || p.Defns == nil || name == "" {
		return nil
	}
	return p.Defns.Get(name)
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
		return p.Defns
	}
	return p.Snapshot
}

func (p *SpecProcessor) actionFromConfigName(name string, init bool, kind string) *Action {
	def, ok := p.configOpDef(name, kind, p.Defns)
	if !ok {
		return nil
	}
	return NewActionFromOpDef(def.Body, EmptyContext, def, init, false)
}

func (p *SpecProcessor) optionalOpBodyFromConfigName(name string, kind string, defns *Defns) SemanticNode {
	if name == "" {
		return nil
	}
	def, ok := p.configOpDef(name, kind, defns)
	if !ok {
		return nil
	}
	return def.Body
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
			} else if def.Body != nil {
				SetSemanticToolObjectForToolID(p.ToolID, def.Body, def)
				nodes = append(nodes, def.Body)
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
