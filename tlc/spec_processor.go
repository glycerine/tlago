package tlc

import (
	"fmt"
	"path/filepath"
	"strings"
)

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
	return tool
}

type SpecProcessor struct {
	RootFile string
	ToolID   int64
	Defns    *Defns
	Config   *ModelConfig

	Variables         []*UniqueString
	ProcessedDefs     *InsMap[string, struct{}]
	UnprocessedDefns  *Defns
	ConstantDefns     *InsMap[string, Value]
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
	PossiblePostConds []*Action
	Assumptions       []SemanticNode
	AssumptionIsAxiom []bool
	RLReward          SemanticNode
	Periodic          SemanticNode
	ViewSpec          SemanticNode
	SymmetrySpec      string
	AliasSpec         string
	SpecificationName string
}

func NewSpecProcessor(rootFile string, defns *Defns, config *ModelConfig) *SpecProcessor {
	if defns == nil {
		defns = NewDefns()
	}
	if config == nil {
		config = newModelConfig("", false)
	}
	p := &SpecProcessor{
		RootFile:         rootFile,
		ToolID:           nextToolID.Add(1),
		Defns:            defns,
		Config:           config,
		ProcessedDefs:    NewInsMap[string, struct{}](),
		UnprocessedDefns: NewDefns(),
		ConstantDefns:    NewInsMap[string, Value](),
	}
	p.ProcessConfig()
	return p
}

func (p *SpecProcessor) SetVariables(names []string) {
	p.Variables = make([]*UniqueString, len(names))
	for i, name := range names {
		variable := UniqueStringOf(name)
		variable.SetLoc(i)
		p.Variables[i] = variable
	}
	SetUniqueStringVariableCount(len(names))
	if p.Defns != nil {
		p.Defns.SetDefnCount(len(names))
	}
}

func (p *SpecProcessor) ProcessConfig() {
	if p == nil || p.Config == nil {
		return
	}
	p.SpecificationName = p.Config.GetSpec()
	p.SymmetrySpec = p.Config.GetSymmetry()
	p.AliasSpec = p.Config.GetAlias()
	p.ViewSpec = p.optionalSemanticFromConfigName(p.Config.GetView())
	p.RLReward = p.optionalSemanticFromConfigName(p.Config.GetRLReward())
	p.Periodic = p.optionalSemanticFromConfigName(p.Config.GetPeriodic())

	p.InitPred = nil
	p.NextPred = nil
	if p.SpecificationName != "" {
		p.processSpecificationConfig()
	} else {
		if initName := p.Config.GetInit(); initName != "" {
			p.InitPred = append(p.InitPred, p.actionFromConfigName(initName, true))
		}
		if nextName := p.Config.GetNext(); nextName != "" {
			p.NextPred = p.actionFromConfigName(nextName, false)
		}
	}
	p.Invariants, p.InvariantNames = p.actionsFromConfigNames(p.Config.GetInvariants(), false)
	p.processConfigProperties()
	p.PossiblePostConds, _ = p.actionsFromConfigNames(p.Config.GetPostConditions(), false)
	p.ModelConstraints = p.constraintNodesFromConfigNames(p.Config.GetConstraints())
	p.ActionConstraints = p.constraintNodesFromConfigNames(p.Config.GetActionConstraints())
}

func (p *SpecProcessor) ApplyToTool(tool *Tool) {
	if p == nil || tool == nil {
		return
	}
	names := make([]string, len(p.Variables))
	for i, variable := range p.Variables {
		names[i] = variable.String()
	}
	if len(names) != 0 {
		SetStateVariables(names)
		if p.Defns != nil {
			p.Defns.SetDefnCount(len(names))
		}
	}
	p.ProcessConfig()
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
	tool.PostConditionSpecs = append([]*Action(nil), p.PossiblePostConds...)
	tool.Assumptions = append([]SemanticNode(nil), p.Assumptions...)
	tool.AssumptionIsAxiom = append([]bool(nil), p.AssumptionIsAxiom...)
	tool.RLReward = p.RLReward
	tool.Periodic = p.Periodic
	tool.ViewSpec = p.ViewSpec
	tool.AssignActionIDs()
}

func (p *SpecProcessor) processSpecificationConfig() {
	def, ok := p.defn(p.SpecificationName).(*OpDefNode)
	if !ok || def == nil || def.Arity() != 0 {
		return
	}
	tool := p.configProcessingTool()
	p.processConfigSpec(tool, def.Body, EmptyContext, EmptyList, nil)
}

func (p *SpecProcessor) processConfigProperties() {
	if p == nil || p.Config == nil {
		return
	}
	tool := p.configProcessingTool()
	for _, name := range p.Config.GetProperties() {
		switch prop := p.defn(name).(type) {
		case *OpDefNode:
			if prop != nil && prop.Arity() == 0 {
				p.processConfigProperty(tool, name, name, prop.Body, EmptyContext, EmptyList)
			}
		case *BoolValue:
			if !prop.Val {
				p.addPlaceholderImpliedTemporal(name)
			}
		case nil:
			p.addPlaceholderImpliedTemporal(name)
		default:
			p.addPlaceholderImpliedTemporal(name)
		}
	}
}

func (p *SpecProcessor) addPlaceholderImpliedTemporal(name string) {
	if name == "" {
		return
	}
	p.ImpliedTemporals = append(p.ImpliedTemporals, &Action{Name: name, Con: EmptyContext})
	p.ImpliedTempNames = append(p.ImpliedTempNames, name)
}

func (p *SpecProcessor) configProcessingTool() *Tool {
	tool := NewTool()
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
		if p.processConfigSpecAppl(tool, node, c, subs, stack) {
			return
		}
	}

	level := tool.GetLevelBound(pred, c)
	if level <= TLCLevelState {
		p.InitPred = append(p.InitPred, NewAction(SpecsAddSubsts(pred, subs), c, SemanticString(pred)))
		return
	}
	if level == TLCLevelTemporal {
		action := NewAction(SpecsAddSubsts(pred, subs), c, SemanticString(pred))
		p.Temporals = append(p.Temporals, action)
		p.TemporalNames = append(p.TemporalNames, SemanticString(pred))
	}
}

func (p *SpecProcessor) processConfigSpecAppl(tool *Tool, pred *OpApplNode, c *Context, subs *List, stack []SemanticNode) bool {
	if pred == nil || pred.Operator == nil {
		return false
	}
	stack = append(stack, pred)
	args := pred.Args
	val := tool.Lookup(pred.Operator, c, EmptyState, false)
	if len(args) == 0 {
		switch v := val.(type) {
		case *OpDefNode:
			if v == nil || v.Arity() != 0 {
				return true
			}
			if tool.GetLevelBound(v.Body, c) == TLCLevelState {
				p.InitPred = append(p.InitPred, NewActionFromOpDef(SpecsAddSubsts(v.Body, subs), c, v, true, false))
				return true
			}
			p.processConfigSpec(tool, v.Body, c, subs, stack)
			return true
		case *BoolValue:
			return true
		case *LazyValue:
			p.processConfigSpec(tool, v.Expr, v.Con, subs, stack[:len(stack)-1])
			return true
		}
	}

	if def, ok := val.(*OpDefNode); ok && def != nil && def.Body != nil && def.Arity() == len(args) && subs.IsEmpty() {
		if c1, err := tool.GetOpContext(def, args, c, true, DoNotRecordCostModel); err == nil {
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
		if boxArg != nil && boxArg.Operator != nil && GetOpCode(boxArg.Operator.Name) == OpcodeSA && len(boxArg.Args) > 0 {
			if p.NextPred == nil {
				p.NextPred = NewAction(SpecsAddSubsts(boxArg.Args[0], subs), c, SemanticString(boxArg.Args[0]))
			}
			return true
		}
		action := NewAction(SpecsAddSubsts(pred, subs), c, SemanticString(pred))
		p.Temporals = append(p.Temporals, action)
		p.TemporalNames = append(p.TemporalNames, SemanticString(pred))
		return true
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
		p.ImpliedActions = append(p.ImpliedActions, NewAction(SpecsAddSubsts(pred, subs), c, configName))
		p.ImpliedActNames = append(p.ImpliedActNames, name)
	case TLCLevelTemporal:
		p.ImpliedTemporals = append(p.ImpliedTemporals, NewAction(SpecsAddSubsts(pred, subs), c, configName))
		p.ImpliedTempNames = append(p.ImpliedTempNames, name)
	}
}

func (p *SpecProcessor) processConfigPropertyAppl(tool *Tool, name string, configName string, pred *OpApplNode, c *Context, subs *List) bool {
	if pred == nil || pred.Operator == nil {
		return false
	}
	args := pred.Args
	opNode := pred.Operator
	val := tool.Lookup(opNode, c, EmptyState, false)
	if len(args) == 0 {
		switch v := val.(type) {
		case *OpDefNode:
			if v == nil || v.Arity() != 0 {
				return true
			}
			opName := name
			if opNode.Name != nil {
				opName = opNode.Name.String()
			}
			p.processConfigProperty(tool, opName, configName, v.Body, c, subs)
			return true
		case *BoolValue:
			return true
		case *LazyValue:
			p.processConfigProperty(tool, name, configName, v.Expr, v.Con, subs)
			return true
		}
	}
	if def, ok := val.(*OpDefNode); ok && def != nil && def.Body != nil && def.Arity() == len(args) && subs.IsEmpty() {
		if c1, err := tool.GetOpContext(def, args, c, true, DoNotRecordCostModel); err == nil {
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
			if ctxts, err := tool.Contexts(pred, c, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel); err == nil && ctxts != nil && ctxts.Err() == nil && !ctxts.IsDone() {
				for c1 := ctxts.NextElement(); c1 != nil; c1 = ctxts.NextElement() {
					p.processConfigProperty(tool, SemanticString(args[0]), configName, args[0], c1, subs)
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
		}
	}
}

func (p *SpecProcessor) defn(name string) any {
	if p == nil || p.Defns == nil || name == "" {
		return nil
	}
	return p.Defns.Get(name)
}

func (p *SpecProcessor) actionFromConfigName(name string, init bool) *Action {
	if def, ok := p.defn(name).(*OpDefNode); ok && def != nil && def.Arity() == 0 {
		return NewActionFromOpDef(def.Body, EmptyContext, def, init, false)
	}
	return &Action{Name: name, IsInitPred: init, Con: EmptyContext}
}

func (p *SpecProcessor) actionsFromConfigNames(names []string, init bool) ([]*Action, []string) {
	actions := make([]*Action, 0, len(names))
	labels := make([]string, 0, len(names))
	for _, name := range names {
		action := p.actionFromConfigName(name, init)
		actions = append(actions, action)
		labels = append(labels, name)
	}
	return actions, labels
}

func (p *SpecProcessor) optionalSemanticFromConfigName(name string) SemanticNode {
	if name == "" {
		return nil
	}
	return p.semanticNodeFromConfigName(name)
}

func (p *SpecProcessor) semanticNodeFromConfigName(name string) SemanticNode {
	switch def := p.defn(name).(type) {
	case *OpDefNode:
		if def != nil && def.Arity() == 0 {
			return def.Body
		}
	case SemanticNode:
		return def
	case Value:
		return def
	}
	return name
}

func (p *SpecProcessor) semanticNodesFromConfigNames(names []string) []SemanticNode {
	nodes := make([]SemanticNode, len(names))
	for i, name := range names {
		nodes[i] = p.semanticNodeFromConfigName(name)
	}
	return nodes
}

func (p *SpecProcessor) constraintNodesFromConfigNames(names []string) []SemanticNode {
	nodes := make([]SemanticNode, 0, len(names))
	for _, name := range names {
		switch def := p.defn(name).(type) {
		case *OpDefNode:
			if def != nil && def.Arity() == 0 && def.Body != nil {
				setSemanticToolObject(def.Body, def)
				nodes = append(nodes, def.Body)
			}
		case *BoolValue:
			if !def.Val {
				nodes = append(nodes, def)
			}
		case SemanticNode:
			nodes = append(nodes, def)
		default:
			nodes = append(nodes, name)
		}
	}
	return nodes
}
